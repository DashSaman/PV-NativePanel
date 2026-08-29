#!/usr/bin/env bash
set -Eeuo pipefail

repo_root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd -P)"
public_root="${repo_root}/site"
index="${public_root}/index.html"
css="${public_root}/assets/site.css"
script="${public_root}/assets/site.js"
articles="${public_root}/data/articles.json"
publisher="${repo_root}/scripts/release/publish-public-site.sh"

for required_file in "${index}" "${css}" "${script}" "${articles}"; do
  [[ -f "${required_file}" ]] || {
    echo "ERROR: public newsroom missing ${required_file#${repo_root}/}" >&2
    exit 1
  }
done

grep -Eq '<html[^>]+lang="fa"[^>]+dir="rtl"|<html[^>]+dir="rtl"[^>]+lang="fa"' "${index}" || {
  echo 'ERROR: public homepage must be Persian RTL' >&2
  exit 1
}

for marker in \
  'data-section="breaking-news"' \
  'data-section="featured-grid"' \
  'data-section="hero-story"' \
  'data-section="latest-news"' \
  'data-section="leader-messages"' \
  'data-section="multimedia"' \
  'data-section="photo-gallery"' \
  'data-section="source-desk"' \
  'data-section="politics"' \
  'data-section="economy"' \
  'data-section="international"'; do
  grep -Fq "${marker}" "${index}" || {
    echo "ERROR: public homepage missing newsroom section ${marker}" >&2
    exit 1
  }
done

for source_domain in khamenei.ir leader.ir president.ir media.president.ir irna.ir isna.ir tasnimnews.com farsnews.ir mehrnews.com; do
  grep -Fq "${source_domain}" "${articles}" || {
    echo "ERROR: curated source registry missing ${source_domain}" >&2
    exit 1
  }
done

# The richer homepage must carry real media semantics in its local cache and a
# usable static fallback. Video is opt-in (no autoplay), source-linked and
# remains non-critical to rendering the page if the remote media host is down.
grep -Fq '"media_type":"video"' "${articles}" || {
  echo 'ERROR: local article cache needs at least one sourced video item' >&2
  exit 1
}
grep -Fq '"media_type":"photo"' "${articles}" || {
  echo 'ERROR: local article cache needs sourced photo-gallery items' >&2
  exit 1
}
grep -Fq '"media_url":"https://media.president.ir/' "${articles}" || {
  echo 'ERROR: video item must reference the official presidency media host' >&2
  exit 1
}
grep -Eq '<video[^>]+controls[^>]+preload="metadata"|<video[^>]+preload="metadata"[^>]+controls' "${index}" || {
  echo 'ERROR: static homepage needs an accessible opt-in video player' >&2
  exit 1
}
if grep -Eiq '<video[^>]+autoplay' "${index}"; then
  echo 'ERROR: public political/news video must never autoplay' >&2
  exit 1
fi
static_story_count="$(grep -Eoc '<article class="(news-card|feed-item|side-story|gallery-card|media-card)' "${index}" || true)"
(( static_story_count >= 10 )) || {
  echo "ERROR: static fallback is too sparse for a professional newsroom (${static_story_count} stories)" >&2
  exit 1
}

# Everything under site/ is potentially file-server reachable. Both file names
# and textual source must remain neutral; internal product/data-plane vocabulary
# belongs only behind the management surface.
forbidden='PVNaive|PVNETWORK|NaiveProxy|VPN|Proxy|Tunnel|Credential|Subscription Management'
if find "${public_root}" -type f -printf '%P\n' | grep -Eiq "${forbidden}"; then
  echo 'ERROR: public tree contains a product-revealing file name' >&2
  exit 1
fi
if grep -RIiEq --binary-files=without-match "${forbidden}" "${public_root}"; then
  echo 'ERROR: public tree leaks management/data-plane product vocabulary' >&2
  exit 1
fi

# The newsroom renders from a local curated cache. A source outage must never be
# able to take the root page down, and browser code must not fetch news directly
# from third-party origins.
grep -Fq '/data/articles.json' "${script}" || {
  echo 'ERROR: newsroom must render from the local curated article cache' >&2
  exit 1
}
grep -Fq 'renderFallback' "${script}" || {
  echo 'ERROR: newsroom must provide a local render fallback' >&2
  exit 1
}
if grep -Eiq 'fetch\([^)]*https?://' "${script}"; then
  echo 'ERROR: browser must not depend directly on third-party news APIs' >&2
  exit 1
fi

# Keep the protected application intentionally undiscoverable from public UI.
if grep -Fq 'href="/panel' "${index}" || grep -Fq "href='/panel" "${index}"; then
  echo 'ERROR: public homepage must not advertise the management panel' >&2
  exit 1
fi

# Responsive and accessibility floor for a production public homepage.
grep -Fq '@media' "${css}" || { echo 'ERROR: public homepage lacks responsive CSS' >&2; exit 1; }
grep -Fq 'aria-label=' "${index}" || { echo 'ERROR: public navigation needs accessible labels' >&2; exit 1; }
grep -Fq 'loading="lazy"' "${index}" || { echo 'ERROR: news imagery must lazy-load below the fold' >&2; exit 1; }

# Live-root publication must replace only the legacy camouflage root that S02
# backed up at /var/www/naive. It must preserve Caddy byte-for-byte and prove
# that the already-separated /panel/ route remains available after promotion.
[[ -f "${publisher}" ]] || {
  echo 'ERROR: missing safe public-site publisher' >&2
  exit 1
}
bash -n "${publisher}"
for required in \
  '/var/www/naive' \
  '/var/backups/pvnaive/public-site' \
  'sha256sum /etc/caddy/Caddyfile' \
  'MainPID' \
  'NRestarts' \
  'public-site-before.tar.gz' \
  'https://${public_host}/' \
  'https://${public_host}/panel/' \
  'PUBLIC_SITE_RESULT=PASSED'; do
  grep -Fq -- "${required}" "${publisher}" || {
    echo "ERROR: public-site publisher missing contract token: ${required}" >&2
    exit 1
  }
done
if grep -Eq 'systemctl[[:space:]]+(restart|reload)[[:space:]]+caddy-naive\.service' "${publisher}"; then
  echo 'ERROR: public-site publication must not restart or reload Caddy' >&2
  exit 1
fi
if grep -Eq '(^|[[:space:]])(cp|mv|install|sed|perl|python[^[:space:]]*)[^\n]*/etc/caddy/Caddyfile' "${publisher}"; then
  echo 'ERROR: public-site publication must not rewrite the Caddyfile' >&2
  exit 1
fi

echo 'PUBLIC_HOMEPAGE_CONTRACT=PASSED'
