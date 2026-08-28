#!/usr/bin/env python3
from __future__ import annotations

import copy
import importlib.util
import json
import tempfile
import unittest
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parents[2]
GENERATOR = REPO_ROOT / "scripts" / "site" / "build_public_portal.py"
spec = importlib.util.spec_from_file_location("public_portal", GENERATOR)
portal = importlib.util.module_from_spec(spec)
assert spec.loader is not None
spec.loader.exec_module(portal)

BASE_ITEM = {
    "id": "fixture-video",
    "slug": "fixture-video",
    "kind": "video",
    "title": "Fixture",
    "summary": "Fixture media",
    "published_label": "test",
    "source_name": "Example Source",
    "source_domain": "example.com",
    "source_url": "https://example.com/item",
    "poster": "/assets/news-fallback.svg",
    "duration_seconds": 10,
    "mirror_allowed": True,
    "rights_note": "CC BY 4.0",
    "attribution": "Example Source — CC BY 4.0",
    "qualities": [
        {
            "label": "Original",
            "mime": "video/webm",
            "bytes": 1234,
            "url": "https://example.com/media.webm",
            "local_path": "video/fixture.webm",
            "sha256": None,
        }
    ],
}


def validate(items):
    with tempfile.TemporaryDirectory() as tmp:
        path = Path(tmp) / "media.json"
        path.write_text(json.dumps({"version": 1, "items": items}), encoding="utf-8")
        return portal.load_and_validate_media(path)


class PublicMediaManifestTest(unittest.TestCase):
    def test_valid_mirrored_item_passes(self):
        payload = validate([copy.deepcopy(BASE_ITEM)])
        self.assertEqual(payload["items"][0]["id"], "fixture-video")

    def test_mirrored_item_requires_attribution(self):
        item = copy.deepcopy(BASE_ITEM)
        item["attribution"] = ""
        with self.assertRaisesRegex(ValueError, "attribution"):
            validate([item])

    def test_mirrored_item_requires_rights_note(self):
        item = copy.deepcopy(BASE_ITEM)
        item["rights_note"] = ""
        with self.assertRaisesRegex(ValueError, "rights_note"):
            validate([item])

    def test_http_source_is_rejected(self):
        item = copy.deepcopy(BASE_ITEM)
        item["source_url"] = "http://example.com/item"
        with self.assertRaisesRegex(ValueError, "HTTPS"):
            validate([item])

    def test_duplicate_id_is_rejected(self):
        first = copy.deepcopy(BASE_ITEM)
        second = copy.deepcopy(BASE_ITEM)
        second["slug"] = "fixture-video-two"
        with self.assertRaisesRegex(ValueError, "duplicate"):
            validate([first, second])

    def test_mirrored_quality_requires_positive_bytes(self):
        for invalid in (None, 0, -1):
            item = copy.deepcopy(BASE_ITEM)
            item["qualities"][0]["bytes"] = invalid
            with self.subTest(invalid=invalid), self.assertRaisesRegex(ValueError, "byte size"):
                validate([item])

    def test_mirrored_quality_requires_local_path(self):
        item = copy.deepcopy(BASE_ITEM)
        item["qualities"][0]["local_path"] = None
        with self.assertRaisesRegex(ValueError, "local_path"):
            validate([item])

    def test_path_traversal_is_rejected(self):
        item = copy.deepcopy(BASE_ITEM)
        item["qualities"][0]["local_path"] = "../escape.webm"
        with self.assertRaisesRegex(ValueError, "local path"):
            validate([item])

    def test_unsupported_mime_is_rejected(self):
        item = copy.deepcopy(BASE_ITEM)
        item["qualities"][0]["mime"] = "application/octet-stream"
        with self.assertRaisesRegex(ValueError, "MIME"):
            validate([item])

    def test_remote_only_quality_may_omit_size(self):
        item = copy.deepcopy(BASE_ITEM)
        item["mirror_allowed"] = False
        item["qualities"][0]["bytes"] = None
        item["qualities"][0]["local_path"] = None
        validate([item])


if __name__ == "__main__":
    unittest.main(verbosity=2)
