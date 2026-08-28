#!/usr/bin/env bash
set -Eeuo pipefail

repo_root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd -P)"
site_root="${repo_root}/site"

required_routes=(
  news/index.html
  videos/index.html
  audio/index.html
  gallery/index.html
  downloads/index.html
  sources/index.html
  about/index.html
)

for route in "${required_routes[@]}"; do
  path="${site_root}/${route}"
  [[ -f "${path}" ]] || {
    echo "ERROR: missing public portal route ${route}" >&2
    exit 1
  }
  grep -Eq '<html[^>]+lang="fa"[^>]+dir="rtl"|<html[^>]+dir="rtl"[^>]+lang="fa"' "${path}" || {
    echo "ERROR: ${route} must be Persian RTL" >&2
    exit 1
  }
done

find_detail() {
  local dir="$1"
  find "${site_root}/${dir}" -mindepth 1 -maxdepth 1 -type f -name '*.html' ! -name 'index.html' -print -quit
}

for section in news videos audio gallery; do
  detail="$(find_detail "${section}")"
  [[ -n "${detail}" ]] || {
    echo "ERROR: ${section} must contain at least one detail page" >&2
    exit 1
  }
done

# The site must navigate internally instead of sending every interaction off-site.
internal_link_count="$(grep -RhoE 'href="/(news|videos|audio|gallery|downloads|sources|about)/[^"#]*"' "${site_root}" --include='*.html' | wc -l | tr -d ' ')"
[[ "${internal_link_count}" -ge 12 ]] || {
  echo "ERROR: public portal needs richer internal navigation (found ${internal_link_count})" >&2
  exit 1
}

# The management surface stays intentionally undiscoverable from all public files.
if grep -RIq --include='*.html' --include='*.json' 'href="/panel\|href='"'"'/panel\|/panel/' "${site_root}"; then
  echo 'ERROR: public portal advertises the management panel' >&2
  exit 1
fi

# External href/src values must be HTTPS; local absolute paths are allowed.
if grep -RhoE '(href|src)="http://[^" ]+"' "${site_root}" --include='*.html' --include='*.js' | grep -q .; then
  echo 'ERROR: public portal contains non-HTTPS external asset/link' >&2
  exit 1
fi

python3 "${repo_root}/tests/site/public_media_manifest_test.py"

echo 'PUBLIC_PORTAL_ROUTES_CONTRACT=PASSED'
