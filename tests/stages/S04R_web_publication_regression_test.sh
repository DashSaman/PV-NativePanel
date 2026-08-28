#!/usr/bin/env bash
set -Eeuo pipefail
umask 077

repo_root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd -P)"
upgrade="${repo_root}/scripts/stages/S04R-upgrade.sh"
web_package="${repo_root}/web/package.json"

[[ -f "${upgrade}" ]] || { echo "ERROR: missing ${upgrade}" >&2; exit 1; }
[[ -f "${web_package}" ]] || { echo "ERROR: missing ${web_package}" >&2; exit 1; }

# Production serves /panel/* from /var/www/pvnaive-preview/current. Vite's
# default absolute /assets/* URLs bypass that handle and are restricted by the
# legacy exact-name Caddy matcher, so a new content-hashed build returns 404.
# The web build must therefore emit /panel/assets/* URLs.
grep -Fq 'vite build --base=/panel/' "${web_package}" || {
  echo 'ERROR: S04R web build must use Vite base=/panel/ so hashed assets stay under the /panel/* route' >&2
  exit 1
}

# S04R must publish the web build to the Caddy-served preview release tree,
# not only to /opt/pvnaive/web/current. Keep the existing root:caddy 0750/0640
# permission model and atomically promote the preview current symlink.
for required in \
  '/var/www/pvnaive-preview/releases' \
  '/var/www/pvnaive-preview/current' \
  'root -g caddy -m 0750' \
  'find "${preview_release_dir}" -type d -exec chmod 0750 {} +' \
  'find "${preview_release_dir}" -type f -exec chmod 0640 {} +' \
  'mv -Tf -- "${preview_current}.new" "${preview_current}"'; do
  grep -Fq -- "${required}" "${upgrade}" || {
    echo "ERROR: S04R upgrade missing preview publication contract: ${required}" >&2
    exit 1
  }
done

# A failed upgrade after preview promotion must restore the previous serving
# target. This is required independently of /opt/pvnaive/web/current rollback.
grep -Fq 'preview_current_before=' "${upgrade}" || {
  echo 'ERROR: S04R upgrade does not capture the previous preview current target' >&2
  exit 1
}
grep -Fq '"${preview_current_before}" "${preview_current}.rollback"' "${upgrade}" || {
  echo 'ERROR: S04R rollback does not prepare restoration of the previous preview target' >&2
  exit 1
}
grep -Fq 'mv -Tf -- "${preview_current}.rollback" "${preview_current}"' "${upgrade}" || {
  echo 'ERROR: S04R rollback does not atomically restore preview current' >&2
  exit 1
}

# Publication must not require a Caddy restart/reload or a Caddyfile rewrite.
if grep -Eq 'systemctl[[:space:]]+(restart|reload)[[:space:]]+caddy-naive\.service' "${upgrade}"; then
  echo 'ERROR: web publication regression fix must not restart or reload Caddy' >&2
  exit 1
fi
if grep -Eq '(^|[[:space:]])(cp|mv|install|sed|perl|python[^[:space:]]*)[^\n]*/etc/caddy/Caddyfile' "${upgrade}"; then
  echo 'ERROR: web publication regression fix must not rewrite the production Caddyfile' >&2
  exit 1
fi

echo 'S04R_WEB_PUBLICATION_REGRESSION=PASSED'
