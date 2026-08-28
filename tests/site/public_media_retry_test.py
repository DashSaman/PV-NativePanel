#!/usr/bin/env python3
from __future__ import annotations

import email.message
import importlib.util
import tempfile
import unittest
import urllib.error
from pathlib import Path
from unittest import mock

REPO_ROOT = Path(__file__).resolve().parents[2]
SYNC_PATH = REPO_ROOT / "scripts" / "site" / "sync_public_media.py"
spec = importlib.util.spec_from_file_location("sync_public_media", SYNC_PATH)
assert spec and spec.loader
sync = importlib.util.module_from_spec(spec)
spec.loader.exec_module(sync)


class FakeHeaders:
    def __init__(self, content_type: str, content_length: int):
        self.content_type = content_type
        self.content_length = content_length

    def get_content_type(self):
        return self.content_type

    def get(self, name: str):
        if name.lower() == "content-length":
            return str(self.content_length)
        return None


class FakeResponse:
    def __init__(self, payload: bytes, content_type: str = "video/webm"):
        self.payload = payload
        self.offset = 0
        self.headers = FakeHeaders(content_type, len(payload))

    def __enter__(self):
        return self

    def __exit__(self, exc_type, exc, tb):
        return False

    def read(self, size: int):
        if self.offset >= len(self.payload):
            return b""
        chunk = self.payload[self.offset:self.offset + size]
        self.offset += len(chunk)
        return chunk


class PublicMediaRetryTest(unittest.TestCase):
    def test_existing_file_with_verified_size_skips_when_checksum_is_absent(self):
        with tempfile.TemporaryDirectory() as tmp:
            target = Path(tmp) / "fixture.webm"
            payload = b"already-verified-over-https"
            target.write_bytes(payload)
            self.assertTrue(sync.existing_is_verified(target, len(payload), None))

    def test_http_429_is_retried_and_then_download_succeeds(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            payload = b"rate-limit-retry-fixture"
            entry = {
                "id": "retry-fixture",
                "url": "https://example.com/retry.webm",
                "local_path": "video/retry.webm",
                "bytes": len(payload),
                "mime": "video/webm",
                "sha256": None,
            }
            headers = email.message.Message()
            headers["Retry-After"] = "0"
            rate_limited = urllib.error.HTTPError(
                entry["url"], 429, "Too Many Requests", headers, None
            )
            responses = [rate_limited, FakeResponse(payload)]

            def fake_urlopen(*_args, **_kwargs):
                next_value = responses.pop(0)
                if isinstance(next_value, BaseException):
                    raise next_value
                return next_value

            with mock.patch.object(sync.urllib.request, "urlopen", side_effect=fake_urlopen):
                written = sync.download_and_promote(
                    entry,
                    root,
                    allow_http_fixtures=False,
                    max_item_bytes=1024,
                )

            self.assertEqual(written, len(payload))
            self.assertEqual((root / "video/retry.webm").read_bytes(), payload)
            self.assertEqual(responses, [])


if __name__ == "__main__":
    unittest.main(verbosity=2)
