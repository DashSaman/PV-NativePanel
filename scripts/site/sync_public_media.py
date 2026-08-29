#!/usr/bin/env python3
from __future__ import annotations

import argparse
import hashlib
import json
import os
import re
import sys
import tempfile
import time
import urllib.error
import urllib.request
from pathlib import Path
from urllib.parse import urlparse

DEFAULT_MAX_ITEM_BYTES = 512 * 1024 * 1024
DEFAULT_MAX_TOTAL_BYTES = 2 * 1024 * 1024 * 1024
DEFAULT_HTTP_ATTEMPTS = 6
DEFAULT_RETRY_SECONDS = 15
ALLOWED_MIME = {
    "video/mp4",
    "video/webm",
    "audio/mpeg",
    "audio/mp4",
    "audio/ogg",
    "application/pdf",
}


def fail(message: str) -> "NoReturn":
    raise ValueError(message)


def validate_local_path(value: str) -> Path:
    if not value or value.startswith("/") or "\\" in value:
        fail(f"invalid local path: {value}")
    parts = Path(value).parts
    if ".." in parts or "." in parts:
        fail(f"invalid local path: {value}")
    if not re.fullmatch(r"[A-Za-z0-9._/-]+", value):
        fail(f"unsafe local path: {value}")
    return Path(value)


def validate_source_url(value: str, allow_http_fixtures: bool, root: Path) -> str:
    parsed = urlparse(value)
    if parsed.scheme == "https" and parsed.netloc:
        return value
    if parsed.scheme == "http" and parsed.netloc and allow_http_fixtures:
        resolved_root = root.resolve()
        tmp_root = Path(tempfile.gettempdir()).resolve()
        try:
            resolved_root.relative_to(tmp_root)
        except ValueError as exc:
            raise ValueError("HTTP fixture override is allowed only under the temporary test root") from exc
        return value
    fail(f"media source must use HTTPS: {value}")


def file_sha256(path: Path) -> str:
    digest = hashlib.sha256()
    with path.open("rb") as handle:
        for chunk in iter(lambda: handle.read(1024 * 1024), b""):
            digest.update(chunk)
    return digest.hexdigest()


def existing_is_verified(target: Path, expected_bytes: int, expected_sha256: str | None) -> bool:
    if not target.is_file() or target.stat().st_size != expected_bytes:
        return False
    if expected_sha256:
        return file_sha256(target) == expected_sha256
    # A previous successful sync already validated HTTPS delivery, MIME and
    # exact byte count before atomically promoting this file. When the
    # upstream manifest has no checksum, exact size is the strongest stable
    # resumability signal available and prevents re-downloading good media.
    return True


def retry_delay_seconds(exc: urllib.error.HTTPError, attempt: int) -> int:
    retry_after = exc.headers.get("Retry-After") if exc.headers else None
    if retry_after is not None:
        try:
            return max(0, min(int(retry_after), 300))
        except (TypeError, ValueError):
            pass
    return min(DEFAULT_RETRY_SECONDS * (2 ** max(0, attempt - 1)), 120)


def open_with_rate_limit_retry(request: urllib.request.Request, item_id: str):
    for attempt in range(1, DEFAULT_HTTP_ATTEMPTS + 1):
        try:
            return urllib.request.urlopen(request, timeout=45)
        except urllib.error.HTTPError as exc:
            if exc.code != 429 or attempt >= DEFAULT_HTTP_ATTEMPTS:
                raise
            delay = retry_delay_seconds(exc, attempt)
            print(
                f"RETRY rate-limit {item_id} attempt={attempt}/{DEFAULT_HTTP_ATTEMPTS} wait={delay}s",
                file=sys.stderr,
            )
            if delay:
                time.sleep(delay)
    raise AssertionError("unreachable")


def iter_mirror_entries(payload: dict):
    items = payload.get("items")
    if not isinstance(items, list):
        fail("media manifest items must be a list")
    for item in items:
        if not item.get("mirror_allowed"):
            continue
        item_id = str(item.get("id") or "").strip()
        rights_note = str(item.get("rights_note") or "").strip()
        attribution = str(item.get("attribution") or "").strip()
        if not item_id:
            fail("mirror-approved media is missing id")
        if not rights_note:
            fail(f"mirror-approved media {item_id} is missing rights note")
        if not attribution:
            fail(f"mirror-approved media {item_id} is missing attribution")
        qualities = item.get("qualities")
        if not isinstance(qualities, list) or not qualities:
            fail(f"mirror-approved media {item_id} has no qualities")
        for quality in qualities:
            local_path = quality.get("local_path")
            if not local_path:
                fail(f"mirror-approved media {item_id} quality is missing local path")
            expected_bytes = quality.get("bytes")
            if not isinstance(expected_bytes, int) or expected_bytes <= 0:
                fail(f"mirror-approved media {item_id} requires a verified positive size")
            mime = quality.get("mime")
            if mime not in ALLOWED_MIME:
                fail(f"mirror-approved media {item_id} has unsupported MIME {mime}")
            checksum = quality.get("sha256")
            if checksum is not None and not re.fullmatch(r"[0-9a-f]{64}", checksum):
                fail(f"mirror-approved media {item_id} has invalid checksum")
            yield {
                "id": item_id,
                "url": quality.get("url") or "",
                "local_path": str(local_path),
                "bytes": expected_bytes,
                "mime": mime,
                "sha256": checksum,
            }


def download_and_promote(entry: dict, root: Path, allow_http_fixtures: bool, max_item_bytes: int) -> int:
    item_id = entry["id"]
    expected_bytes = int(entry["bytes"])
    if expected_bytes > max_item_bytes:
        fail(f"media {item_id} exceeds configured item size limit")

    relative_path = validate_local_path(entry["local_path"])
    target = (root / relative_path).resolve()
    resolved_root = root.resolve()
    try:
        target.relative_to(resolved_root)
    except ValueError as exc:
        raise ValueError(f"invalid local path escaped media root: {entry['local_path']}") from exc

    url = validate_source_url(entry["url"], allow_http_fixtures, root)
    target.parent.mkdir(parents=True, exist_ok=True)

    if existing_is_verified(target, expected_bytes, entry.get("sha256")):
        print(f"SKIP verified {item_id} bytes={expected_bytes} path={relative_path.as_posix()}")
        return 0

    request = urllib.request.Request(url, headers={"User-Agent": "PVNaive-public-media-sync/2"})
    temp_path: Path | None = None
    try:
        with open_with_rate_limit_retry(request, item_id) as response:
            content_type = response.headers.get_content_type()
            if content_type != entry["mime"]:
                fail(f"MIME mismatch for {item_id}: expected {entry['mime']}, got {content_type}")

            header_length = response.headers.get("Content-Length")
            if header_length is not None:
                try:
                    declared = int(header_length)
                except ValueError as exc:
                    raise ValueError(f"invalid Content-Length for {item_id}") from exc
                if declared > max_item_bytes:
                    fail(f"media {item_id} exceeds configured item size limit")
                if declared != expected_bytes:
                    fail(f"size mismatch for {item_id}: expected {expected_bytes}, got {declared}")

            digest = hashlib.sha256()
            written = 0
            fd, name = tempfile.mkstemp(prefix=f".{target.name}.", suffix=".part", dir=str(target.parent))
            os.close(fd)
            temp_path = Path(name)
            with temp_path.open("wb") as output:
                while True:
                    chunk = response.read(1024 * 1024)
                    if not chunk:
                        break
                    written += len(chunk)
                    if written > max_item_bytes:
                        fail(f"media {item_id} exceeds configured item size limit")
                    if written > expected_bytes:
                        fail(f"size mismatch for {item_id}: expected {expected_bytes}, received more")
                    output.write(chunk)
                    digest.update(chunk)
                output.flush()
                os.fsync(output.fileno())

            if written != expected_bytes:
                fail(f"size mismatch for {item_id}: expected {expected_bytes}, got {written}")
            expected_sha = entry.get("sha256")
            actual_sha = digest.hexdigest()
            if expected_sha and actual_sha != expected_sha:
                fail(f"checksum mismatch for {item_id}: expected {expected_sha}, got {actual_sha}")

            os.chmod(temp_path, 0o644)
            os.replace(temp_path, target)
            temp_path = None
            print(f"SYNCED {item_id} bytes={written} path={relative_path.as_posix()}")
            return written
    finally:
        if temp_path is not None:
            try:
                temp_path.unlink()
            except FileNotFoundError:
                pass


def plan_entries(entries: list[dict], root: Path, allow_http_fixtures: bool, max_item_bytes: int, max_total_bytes: int) -> int:
    total = 0
    for entry in entries:
        validate_local_path(entry["local_path"])
        validate_source_url(entry["url"], allow_http_fixtures, root)
        expected_bytes = int(entry["bytes"])
        if expected_bytes > max_item_bytes:
            fail(f"media {entry['id']} exceeds configured item size limit")
        total += expected_bytes
        if total > max_total_bytes:
            fail("planned media exceeds configured total size limit")
        print(f"PLAN mirror {entry['id']} bytes={expected_bytes} path={entry['local_path']}")
    return total


def main() -> int:
    parser = argparse.ArgumentParser(description="Safely mirror explicitly approved public media files")
    parser.add_argument("--manifest", type=Path, default=Path("site/data/media.json"))
    parser.add_argument("--root", type=Path, default=Path("/var/www/naive/media"))
    parser.add_argument("--apply", action="store_true", help="download and atomically promote approved media")
    parser.add_argument("--allow-http-fixtures", action="store_true", help="test-only: allow HTTP when root is under the OS temp directory")
    parser.add_argument("--max-item-bytes", type=int, default=DEFAULT_MAX_ITEM_BYTES)
    parser.add_argument("--max-total-bytes", type=int, default=DEFAULT_MAX_TOTAL_BYTES)
    args = parser.parse_args()

    if args.max_item_bytes <= 0 or args.max_total_bytes <= 0:
        fail("size limit values must be positive")

    payload = json.loads(args.manifest.read_text(encoding="utf-8"))
    entries = list(iter_mirror_entries(payload))
    planned = plan_entries(entries, args.root, args.allow_http_fixtures, args.max_item_bytes, args.max_total_bytes)

    if not args.apply:
        print(f"PUBLIC_MEDIA_SYNC_DRY_RUN=PASSED items={len(entries)} bytes={planned}")
        return 0

    args.root.mkdir(parents=True, exist_ok=True)
    synced_bytes = 0
    for entry in entries:
        synced_bytes += download_and_promote(entry, args.root, args.allow_http_fixtures, args.max_item_bytes)
        if synced_bytes > args.max_total_bytes:
            fail("downloaded media exceeds configured total size limit")
    print(f"PUBLIC_MEDIA_SYNC_RESULT=PASSED items={len(entries)} new_bytes={synced_bytes} planned_bytes={planned}")
    return 0


if __name__ == "__main__":
    try:
        raise SystemExit(main())
    except Exception as exc:
        print(f"ERROR: {exc}", file=sys.stderr)
        raise SystemExit(1)
