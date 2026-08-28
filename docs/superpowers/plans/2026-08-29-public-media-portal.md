# Public Media Portal Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Turn the current single-page Persian public newsroom into a static multi-page media portal with internal navigation, article/video/audio/gallery detail pages, a truthful downloads library, and a safe optional media mirroring workflow while keeping `/panel/` untouched and undiscoverable.

**Architecture:** Keep the public surface fully static and independent from the Go management API. Curated JSON manifests drive a deterministic shell/Python site generator that emits route directories and detail pages; media metadata separates locally mirrored files from remote official-source playback/download URLs. A separate hardened sync script mirrors only entries explicitly marked `mirror_allowed=true`, validates size/MIME/checksum, and atomically promotes files under the public media tree.

**Tech Stack:** Static HTML/CSS/vanilla JS, JSON manifests, Bash/Python 3 build tooling, existing Caddy static file serving, GitHub Actions CI, existing release/publisher scripts.

**Spec:** `docs/superpowers/specs/2026-08-29-public-media-portal-design.md`

## Global Constraints

- Public UI must never advertise or link `/panel/`.
- Do not change Caddy routing, restart Caddy, or rewrite `/etc/caddy/Caddyfile`.
- Browser code must not fetch live third-party news APIs.
- Political content must remain source-attributed and informational; do not impersonate an official government/leadership website.
- Do not fabricate titles, quotations, dates, durations, sizes, MIME types, checksums, or download URLs.
- Mirror binaries locally only when `mirror_allowed=true` and the rights note supports local mirroring.
- Large media binaries are not committed to Git.
- All new behaviors are introduced test-first.

---

### Task 1: Multi-page route contract and deterministic static generator

**Files:**
- Create: `tests/site/public_portal_routes_test.sh`
- Create: `scripts/site/build_public_portal.py`
- Create: `site/data/portal.json`
- Modify: `tests/site/public_homepage_test.sh`
- Modify: `.github/workflows/ci.yml`

**Interfaces:**
- Consumes: existing `site/index.html`, `site/assets/site.css`, `site/assets/site.js`, `site/data/articles.json`.
- Produces: `python3 scripts/site/build_public_portal.py --check` and generated public routes `/news/`, `/videos/`, `/audio/`, `/gallery/`, `/downloads/`, `/sources/`, `/about/` plus detail pages from manifests.

- [ ] **Step 1: Write the failing route contract**

Create `tests/site/public_portal_routes_test.sh` that fails unless all route index files and at least one detail page per content type exist, every generated page is Persian RTL, internal navigation contains local links, no public file contains `href="/panel`, and all external `href`/`src` values are HTTPS.

- [ ] **Step 2: Run RED**

Run:
```bash
bash tests/site/public_portal_routes_test.sh
```
Expected: FAIL because `/news/index.html`, `/videos/index.html`, `/audio/index.html`, `/gallery/index.html`, `/downloads/index.html`, `/sources/index.html`, and `/about/index.html` do not exist.

- [ ] **Step 3: Implement minimal deterministic generator**

Create `scripts/site/build_public_portal.py` using only Python stdlib. It must:
- load JSON manifests,
- escape all text with `html.escape`,
- reject non-HTTPS external URLs,
- reject local paths containing `..`, backslashes, or absolute paths,
- render shared header/footer and breadcrumbs,
- write deterministic UTF-8 output,
- support `--check` to validate manifests without writing,
- support `--output-root <path>` for CI fixture generation,
- default to writing under `site/`.

- [ ] **Step 4: Add navigation manifest**

Create `site/data/portal.json` with portal name, disclosure, sections, and the route map. Keep independent/reference-site wording and no official-site impersonation.

- [ ] **Step 5: Wire CI**

Add `bash tests/site/public_portal_routes_test.sh` and `python3 scripts/site/build_public_portal.py --check` to the existing database/site gate before release packaging.

- [ ] **Step 6: Run GREEN**

Run locally:
```bash
python3 scripts/site/build_public_portal.py
bash tests/site/public_portal_routes_test.sh
bash tests/site/public_homepage_test.sh
```
Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add tests/site scripts/site site/data/portal.json site/news site/videos site/audio site/gallery site/downloads site/sources site/about .github/workflows/ci.yml
git commit -m "feat(site): generate multi-page public portal"
```

---

### Task 2: Curated article, video, audio and gallery manifests

**Files:**
- Create: `tests/site/public_media_manifest_test.py`
- Create: `site/data/media.json`
- Create: `site/data/galleries.json`
- Modify: `site/data/articles.json`
- Modify: `scripts/site/build_public_portal.py`

**Interfaces:**
- Produces media records with: `id`, `slug`, `kind`, `title`, `summary`, `published_label`, `source_name`, `source_domain`, `source_url`, `poster`, `duration_seconds|null`, `qualities[]`, `mirror_allowed`, `rights_note`, `attribution`.
- Each `qualities[]` item: `label`, `mime`, `bytes|null`, `url`, `local_path|null`, `sha256|null`.

- [ ] **Step 1: Write failing manifest validation tests**

`tests/site/public_media_manifest_test.py` must reject:
- missing source URL/domain,
- HTTP URLs,
- duplicate IDs/slugs,
- `mirror_allowed=true` without non-empty `rights_note`,
- mirrored quality with null/zero byte size,
- mirrored quality without `local_path`,
- path traversal,
- unsupported media MIME.

It must allow remote-only qualities to have unknown size, but UI must then omit size rather than invent one.

- [ ] **Step 2: Run RED**

```bash
python3 -m unittest tests.site.public_media_manifest_test -v
```
Expected: FAIL because media manifest/validator does not exist.

- [ ] **Step 3: Implement validator inside generator module**

Expose `load_and_validate_media(path)` and `validate_local_path(value)` in `scripts/site/build_public_portal.py` so tests execute real production validation.

- [ ] **Step 4: Curate initial manifests**

Add 15–25 homepage/article references, 3–5 video entries, 3–5 audio/speech entries, and 2 galleries. Each entry must have a canonical official/established source URL. Only mark `mirror_allowed=true` where reuse/local mirroring is explicitly supported; otherwise use remote official media URLs.

- [ ] **Step 5: Rebuild routes and run GREEN**

```bash
python3 scripts/site/build_public_portal.py
python3 -m unittest tests.site.public_media_manifest_test -v
bash tests/site/public_portal_routes_test.sh
```
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add tests/site/public_media_manifest_test.py site/data scripts/site/build_public_portal.py site
git commit -m "feat(site): add curated public media catalog"
```

---

### Task 3: Native online playback and truthful download library

**Files:**
- Create: `tests/site/public_media_pages_test.sh`
- Modify: `scripts/site/build_public_portal.py`
- Modify: `site/assets/site.css`
- Modify: generated `site/videos/**`, `site/audio/**`, `site/downloads/index.html`

**Interfaces:**
- Video detail pages contain `<video controls preload="metadata">` and one or more `<source>` elements.
- Audio detail pages contain `<audio controls preload="metadata">`.
- Download links use `download` only for local files; remote official URLs use ordinary external links.

- [ ] **Step 1: Write failing media-page contract**

Require at least one video and one audio detail page, native players, format labels, source attribution, a visible canonical-source link, and download rows whose displayed byte size matches the manifest when `bytes` is known.

- [ ] **Step 2: Run RED**

```bash
bash tests/site/public_media_pages_test.sh
```
Expected: FAIL because generated detail pages do not yet contain native media players/download rows.

- [ ] **Step 3: Implement player/download templates**

Render poster, title, metadata, `<video>`/`<audio>`, quality table, format, verified size via a deterministic human-size helper, source attribution, rights note, and related local links.

- [ ] **Step 4: Improve portal density**

Extend `site/assets/site.css` for media hero, player shell, quality table, download cards, gallery grids, related-content rows, breadcrumb bar, archive pagination placeholders, and mobile layouts.

- [ ] **Step 5: Run GREEN**

```bash
python3 scripts/site/build_public_portal.py
bash tests/site/public_media_pages_test.sh
bash tests/site/public_portal_routes_test.sh
```
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add tests/site/public_media_pages_test.sh scripts/site/build_public_portal.py site/assets/site.css site/videos site/audio site/downloads site/gallery site/news site/sources site/about
git commit -m "feat(site): add native media playback and downloads"
```

---

### Task 4: Hardened optional media mirroring

**Files:**
- Create: `tests/site/media_sync_test.sh`
- Create: `scripts/site/sync_public_media.py`
- Modify: `site/data/media.json`

**Interfaces:**
- Command: `python3 scripts/site/sync_public_media.py --manifest site/data/media.json --root /var/www/naive/media`
- Default dry-run unless `--apply` is supplied.
- Mirrors only qualities with `mirror_allowed=true` and `local_path`.

- [ ] **Step 1: Write RED sync tests using a local fixture HTTP server**

Test rejection of HTTP source URLs, path traversal, wrong MIME, zero/oversize bodies, expected-size mismatch, checksum mismatch, and unknown `mirror_allowed` rights. Test idempotency and atomic replacement using temporary directories.

- [ ] **Step 2: Run RED**

```bash
bash tests/site/media_sync_test.sh
```
Expected: FAIL because sync tool does not exist.

- [ ] **Step 3: Implement sync tool**

Use Python stdlib `urllib.request`, `tempfile`, `hashlib`, `os.replace`. Enforce:
- HTTPS in production manifests,
- configurable `--max-item-bytes` default 512 MiB,
- configurable `--max-total-bytes` default 2 GiB,
- MIME allowlist `video/mp4`, `audio/mpeg`, `audio/mp4`, `application/pdf`,
- exact expected byte count when declared,
- SHA-256 validation when declared,
- temp file in target filesystem then `os.replace`,
- no deletion of unrelated files,
- dry-run report.

For tests, add `--allow-http-fixtures` accepted only with destination under a temporary/test root.

- [ ] **Step 4: Run GREEN**

```bash
bash tests/site/media_sync_test.sh
```
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add tests/site/media_sync_test.sh scripts/site/sync_public_media.py site/data/media.json
git commit -m "feat(site): add safe public media mirror sync"
```

---

### Task 5: Release bundle and publisher regression

**Files:**
- Modify: `tests/release/S04R_bundle_contract_test.sh`
- Modify: `scripts/release/build-s04-bundle.sh`
- Modify: `tests/site/public_homepage_test.sh`
- Modify: `scripts/release/publish-public-site.sh` only if needed to preserve generated route tree; do not change Caddy behavior.

**Interfaces:**
- Production bundle contains generated portal HTML, manifests, and sync/build tooling.
- Production bundle does not contain large mirrored media binaries.

- [ ] **Step 1: Write RED bundle assertions**

Require `public-site/news/index.html`, `public-site/videos/index.html`, `public-site/audio/index.html`, `public-site/downloads/index.html`, `public-site/data/media.json`, and `scripts/site/sync_public_media.py`; reject media binaries larger than 1 MiB in the Git/release portal tree.

- [ ] **Step 2: Run RED bundle contract**

```bash
bash tests/release/S04R_bundle_contract_test.sh
```
Expected: FAIL until build script carries new generated routes/tooling correctly.

- [ ] **Step 3: Update bundle builder minimally**

Package complete `site/` tree and `scripts/site/` metadata/tooling while preserving current checksum behavior.

- [ ] **Step 4: Run GREEN regression suite**

```bash
bash tests/site/public_homepage_test.sh
bash tests/site/public_portal_routes_test.sh
bash tests/site/public_media_pages_test.sh
bash tests/site/media_sync_test.sh
bash tests/release/S04R_bundle_contract_test.sh
```
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add tests/release tests/site scripts/release scripts/site
git commit -m "test(release): carry multi-page public media portal"
```

---

### Task 6: Full CI, evidence, and production operator commands

**Files:**
- Modify: `WORKLOG.md`
- Modify: `PROJECT_STATUS.md`
- Modify: `AGENT_HANDOFF.md`
- PR #5 conversation evidence

**Interfaces:**
- No production mutation during implementation.
- Production operator receives two explicit actions after GREEN: publish static site, then optionally sync approved mirrored media.

- [ ] **Step 1: Run full repository CI locally where feasible**

```bash
go test ./...
cd web && npm test && npm run build
```
Then return to repo root and run all site/release tests above.

- [ ] **Step 2: Push and verify GitHub Actions**

Require Go, Web, PostgreSQL/database, S04 auth rehearsal, S04R runtime rehearsal, and bundle jobs all `success` on the exact HEAD commit.

- [ ] **Step 3: Update PM docs**

Record routes, media policy, exact CI run, commit SHA, known remote-only vs mirror-approved entries, and production commands.

- [ ] **Step 4: Add PR evidence comment**

Document RED/GREEN cycle and explicitly state no production Caddy mutation/restart occurred.

- [ ] **Step 5: Prepare operator commands**

Static portal publish remains:
```bash
bash scripts/release/publish-public-site.sh site
```

Optional approved-media sync:
```bash
python3 scripts/site/sync_public_media.py --manifest site/data/media.json --root /var/www/naive/media --apply
```

The sync command must print every mirrored/skipped entry and total bytes before completion.

- [ ] **Step 6: Final verification**

Re-check exact HEAD CI and only then report completion.
