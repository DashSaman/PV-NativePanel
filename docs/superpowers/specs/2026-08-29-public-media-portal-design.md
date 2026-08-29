# Public Media Portal Design

Date: 2026-08-29
Branch: `s05-user-quota-design`
Status: proposed, awaiting user review before implementation

## Goal

Expand the existing Persian RTL public homepage into a professional, multi-page public media portal that looks and behaves like a real independent news/reference site while keeping `/panel/` fully separate and undiscoverable from the public UI.

The portal will provide richer internal navigation, article/detail pages, video/audio playback, downloadable media, photo galleries, source directories, and archive pages. Political content remains source-attributed and informational; the portal must not impersonate an official government or leadership website.

## Product Boundary

Public surface:

- `/` — homepage
- `/news/` — latest-news index
- `/news/<slug>.html` — individual article/reference page
- `/videos/` — video archive
- `/videos/<slug>.html` — video detail with HTML5 player and download options
- `/audio/` — audio/speech archive
- `/audio/<slug>.html` — audio detail with HTML5 player and download options
- `/gallery/` — photo gallery index
- `/gallery/<slug>.html` — gallery detail
- `/downloads/` — downloadable-file library
- `/sources/` — official/internal source directory
- `/about/` — independent-site disclosure and sourcing policy

Protected management surface remains:

- `/panel/` — unchanged, not linked or advertised anywhere in `site/`

## Visual Direction

Use the already-approved visual system as the base and upgrade it into a denser magazine/news portal:

- dark navy masthead and editorial accents
- warm off-white page background
- restrained red accent
- larger desktop content width
- high-density card grids without clutter
- consistent image ratios
- strong Persian typography hierarchy
- mobile-first responsive collapse
- clear play/download affordances
- internal breadcrumbs and related-content blocks

The site should appear established and content-rich, but should clearly brand itself as an independent information/reference portal rather than copying official government branding.

## Content Model

Replace the single flat `articles.json` assumption with a versioned portal dataset while preserving static-first fallback.

Proposed files:

- `site/data/portal.json` — top-level source registry, homepage selections, navigation labels
- `site/data/articles.json` — article/reference entries
- `site/data/media.json` — video/audio/download metadata
- `site/data/galleries.json` — gallery/photo metadata

Every entry must include a stable `id`, slug, title, short summary, publication label/date when known, source name/domain, canonical source URL, and content type.

Media entries additionally contain:

- `kind`: `video`, `audio`, `pdf`, `document`
- `poster` or artwork
- `duration_seconds` when known
- `qualities[]` with label, MIME type, byte size, and URL
- `local_path` only when a local copy is explicitly permitted and actually present
- `source_url`
- `rights_note`
- `attribution`
- `sha256` for locally mirrored files

Unknown byte sizes must not be fabricated. UI displays size only when verified.

## Media Hosting Policy

The user wants speeches, messages, audio, and videos available for online playback and download from the site.

Use a hybrid ingestion policy:

1. When the source explicitly offers downloadable media and reuse/local mirroring is permitted, download a local copy to the public media store and serve it from the site.
2. When reuse permission is unclear, keep the media page local but stream/download from the original HTTPS media/CDN URL rather than silently copying the binary.
3. Preserve attribution and canonical source URL on every media detail page.
4. Never remove watermarks, attribution, metadata, or source identity from mirrored media.
5. Never fabricate a local file, size, duration, checksum, or source.

Current research confirms that KHAMENEI.IR-origin media is distributed in multiple downloadable MP4/audio qualities through linked archives/CDN endpoints. For example, the public archive page for the video “شما اکنون در طرف درست تاریخ ایستاده‌اید” exposes high and medium MP4 downloads from `idc0-cdn5.khamenei.ir` with displayed sizes. The implementation must verify each chosen direct media URL before adding it to the manifest.

## Local Media Store

Repository keeps metadata and scripts, not large binary media by default.

Server layout after deployment:

- `/var/www/naive/media/video/`
- `/var/www/naive/media/audio/`
- `/var/www/naive/media/docs/`
- `/var/www/naive/media/posters/`

A new idempotent sync script will fetch only manifest entries marked `mirror_allowed=true`, validate HTTPS source, expected MIME type, declared byte size tolerance, and optional SHA-256, then atomically promote the file.

A failed download must leave the previously verified local file untouched.

## Playback

Video pages use native HTML5 `<video controls preload="metadata">` with poster images and multiple `<source>` elements where local/direct MP4 variants are available.

Audio pages use native `<audio controls preload="metadata">`.

Playback requirements:

- no autoplay
- keyboard-accessible controls
- no third-party JavaScript player dependency
- direct download button for each verified quality
- visible format and verified size
- source attribution beside player
- graceful fallback link when inline playback is unsupported

## Downloads Library

`/downloads/` will list verified files with:

- title
- type (MP4, MP3, PDF, etc.)
- verified size
- duration for media when known
- source
- publication date/label
- direct download button
- detail-page link

Download responses rely on normal static-file serving. No custom application backend is introduced for public media.

## Internal Linking

Natural internal navigation is a primary requirement.

Each detail page should link to:

- category/index page
- source page
- 3–6 related entries
- previous/next item where meaningful
- downloads page when media exists
- gallery page when related photos exist

Homepage sections should link inward rather than sending every click directly off-site. Canonical source links remain visible on detail pages.

## Static Generation Approach

Keep the public site static and independent from the Go management API.

Add a small deterministic build script under `scripts/site/` that reads curated JSON manifests and emits static HTML files into `site/` or a generated staging directory before release packaging.

Do not add a CMS, database, server-side rendering framework, or runtime dependency.

Benefits:

- public root remains available even if upstream sources are down
- no third-party API dependency in browser JavaScript
- easy backup/rollback
- fast Caddy static serving
- no change to `/panel/` application architecture

## Source and Provenance Rules

Allowed source families include official or established domestic sources already used by the portal, such as:

- `farsi.khamenei.ir`
- `khamenei.ir`
- `leader.ir`
- `rahbar.ir`
- `president.ir`
- `irna.ir`
- `isna.ir`
- `tasnimnews.com` / `tasnimnews.ir`
- `mehrnews.com`
- other explicitly reviewed sources added to the registry

Every factual headline/summary must map to a canonical source URL. Short quotations must remain short and attributed. Full copyrighted articles are not copied into local pages.

## Security and Isolation

The public portal must not reveal management/data-plane vocabulary already prohibited by `tests/site/public_homepage_test.sh`.

Requirements:

- no `/panel/` links in public HTML/JSON
- no public JS requests to management APIs
- no Caddyfile rewrite, reload, or restart for content publication
- only HTTPS external URLs
- sanitize generated href/src values against the curated allowlist
- local media paths cannot contain `..`, shell metacharacters, or absolute paths
- media sync uses temp files + atomic rename
- file size cap per media item and global sync cap
- `Content-Type` is verified before promotion

## Publication and Rollback

Reuse `scripts/release/publish-public-site.sh` as the final atomic site publisher.

Extend release flow in two independent stages:

1. Build/verify static portal tree.
2. Optionally sync approved media binaries, then publish the static tree.

The publisher must continue to:

- back up `/var/www/naive`
- preserve `/etc/caddy/Caddyfile` byte-for-byte
- preserve Caddy MainPID and NRestarts
- probe `/`
- probe `/panel/`
- rollback on failure

Media sync gets its own backup/rollback evidence and must not delete unrelated files.

## TDD / CI Contracts

Add RED tests first for:

- required public routes/files exist
- at least one video detail page and one audio detail page contain native media players
- `/downloads/` contains verified-size metadata and download links
- internal links point to local pages
- no public page advertises `/panel/`
- all curated external links are HTTPS
- media manifest rejects missing/zero/fabricated size fields for mirrored files
- sync script rejects HTTP, wrong MIME, oversized files, checksum mismatch, and path traversal
- sync script is idempotent
- public publisher still preserves Caddy SHA/PID/NRestarts
- release bundle includes all generated pages/manifests/scripts but not large binary media artifacts by default

CI remains fully offline from production. Media-download integration tests use local fixture HTTP servers/files, never live upstream hosts.

## Initial Content Target

First implementation target:

- homepage with 15–25 curated cards/links
- 6–10 article/detail pages
- 3–5 video detail pages
- 3–5 audio/speech detail pages
- 2 gallery detail pages
- 1 downloads library page
- 1 sources directory page
- 1 about/sourcing-policy page

For media, start with entries whose source pages expose explicit downloadable qualities and verified sizes. The build must support adding more entries later only by extending JSON data.

## Non-Goals

- no public user accounts
- no comments
- no CMS/admin editor in this phase
- no live scraping from the browser
- no impersonation of an official government website
- no analytics/tracking dependency
- no Caddy routing redesign
- no changes to the protected panel architecture

## Acceptance Criteria

The feature is complete only when:

1. the multi-page static portal builds deterministically;
2. internal navigation works across all public pages;
3. video/audio pages play verified media online;
4. verified downloadable qualities expose truthful size/format metadata;
5. locally mirrored media is only created through the approved sync policy;
6. upstream failure does not break the static HTML shell;
7. `/panel/` remains reachable but absent from public navigation/source content;
8. full CI, PostgreSQL gates, auth rehearsal, runtime rehearsal, release bundle, and public-site publication contracts are green;
9. production publication is a separate explicit operator action after CI success.
