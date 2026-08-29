#!/usr/bin/env python3
import json
from pathlib import Path

repo_root = Path(__file__).resolve().parents[2]
manifest = json.loads((repo_root / "site/data/media.json").read_text(encoding="utf-8"))
items = manifest.get("items", [])

video_items = [item for item in items if item.get("kind") == "video"]
mirrored = [item for item in items if item.get("mirror_allowed")]
local_qualities = [
    quality
    for item in mirrored
    for quality in item.get("qualities", [])
    if quality.get("local_path")
]
local_bytes = sum(int(quality.get("bytes") or 0) for quality in local_qualities)

assert len(items) >= 9, f"expected at least 9 media entries, found {len(items)}"
assert len(video_items) >= 6, f"expected at least 6 video entries, found {len(video_items)}"
assert len(mirrored) >= 4, f"expected at least 4 locally mirrored entries, found {len(mirrored)}"
assert len(local_qualities) >= 4, f"expected at least 4 local downloadable qualities, found {len(local_qualities)}"
assert local_bytes >= 80_000_000, f"expected at least 80 MB of verified local media, found {local_bytes} bytes"

for item in mirrored:
    assert item.get("rights_note"), f"mirrored item {item.get('id')} is missing rights note"
    assert item.get("attribution"), f"mirrored item {item.get('id')} is missing attribution"
    for quality in item.get("qualities", []):
        assert quality.get("url", "").startswith("https://"), f"mirrored item {item.get('id')} must use HTTPS"
        assert quality.get("bytes", 0) > 0, f"mirrored item {item.get('id')} must have verified size"
        assert quality.get("local_path"), f"mirrored item {item.get('id')} must have local path"

print(
    "PUBLIC_MEDIA_DENSITY_CONTRACT=PASSED "
    f"items={len(items)} videos={len(video_items)} mirrored={len(mirrored)} local_bytes={local_bytes}"
)
