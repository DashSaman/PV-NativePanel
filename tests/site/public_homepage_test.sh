#!/usr/bin/env bash
set -Eeuo pipefail

repo_root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd -P)"
public_root="${repo_root}/site"
index="${public_root}/index.html"
css="${public_root}/assets/site.css"
script="${public_root}/assets/site.js"
articles="${public_root}/data/articles.json"

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
  'data-section="hero-story"' \
  'data-section="latest-news"' \
  'data-section="leader-messages"' \
  'data-section="politics"' \
  'data-section="economy"' \
  'data-section="international"'; do
  grep -Fq "${marker}" "${index}" || {
    echo "ERROR: public homepage missing newsroom section ${marker}" >&2
    exit 1
  }
done

for source_domain in khamenei.ir president.ir irna.ir isna.ir tasnimnews.com farsnews.ir mehrnews.com; do
  grep -Fq "${source_domain}" "${articles}" || {
    echo "ERROR: curated source registry missing ${source_domain}" >&2
    exit 1
  }
done

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

echo 'PUBLIC_HOMEPAGE_CONTRACT=PASSED'
