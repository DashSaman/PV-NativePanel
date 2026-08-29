#!/usr/bin/env bash
set -Eeuo pipefail
umask 077

repo_root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd -P)"
source_dir="${1:-${PVNAIVE_PUBLIC_SITE_SOURCE:-${repo_root}/site}}"
public_host="${PVNAIVE_PUBLIC_HOST:-namir.softarg.ir}"
expected_host="${PVNAIVE_EXPECTED_HOST:-testAmir5-3}"
live_root="/var/www/naive"
backup_root="/var/backups/pvnaive/public-site"
caddy_file="/etc/caddy/Caddyfile"
caddy_service="caddy-naive.service"
stamp="$(date -u +%Y%m%dT%H%M%SZ)"
backup_dir="${backup_root}/${stamp}"
stage_root="/var/www/.naive.stage.${stamp}"
previous_root="/var/www/.naive.previous.${stamp}"
failed_root="/var/www/.naive.failed.${stamp}"
root_body=""
published=0

fail() {
  echo "PUBLIC_SITE_RESULT=FAILED" >&2
  echo "ERROR=$*" >&2
  return 1
}

cleanup_unpublished() {
  rm -rf -- "${stage_root}" >/dev/null 2>&1 || true
  if [[ -n "${root_body}" ]]; then
    rm -f -- "${root_body}" >/dev/null 2>&1 || true
  fi
}

rollback_on_error() {
  local rc="${1:-1}"
  trap - ERR EXIT HUP INT TERM
  if ((published == 1)); then
    if [[ -d "${live_root}" && -d "${previous_root}" ]]; then
      rm -rf -- "${failed_root}" >/dev/null 2>&1 || true
      mv -- "${live_root}" "${failed_root}" || true
      mv -- "${previous_root}" "${live_root}" || true
    fi
  fi
  cleanup_unpublished
  echo "PUBLIC_SITE_ROLLBACK=ATTEMPTED" >&2
  exit "${rc}"
}
trap 'rollback_on_error $?' ERR
trap 'rollback_on_error 129' HUP
trap 'rollback_on_error 130' INT
trap 'rollback_on_error 143' TERM
trap cleanup_unpublished EXIT

[[ ${EUID} -eq 0 ]] || fail 'run as root'
[[ "$(hostname)" == "${expected_host}" ]] || fail "unexpected host: $(hostname)"

for command_name in sha256sum systemctl tar curl grep find cp mv install stat rm; do
  command -v "${command_name}" >/dev/null 2>&1 || fail "missing command: ${command_name}"
done
[[ -x /usr/local/bin/caddy ]] || fail 'Caddy binary is missing'

[[ -d "${source_dir}" ]] || fail "public source directory is missing: ${source_dir}"
for required in index.html assets/site.css assets/site.js assets/news-mark.svg assets/news-fallback.svg data/articles.json; do
  [[ -f "${source_dir}/${required}" ]] || fail "public source missing ${required}"
done

# Static safety checks mirror the repository contract and deliberately avoid
# executing third-party content. The protected management path stays undisclosed.
grep -Eq '<html[^>]+lang="fa"[^>]+dir="rtl"|<html[^>]+dir="rtl"[^>]+lang="fa"' "${source_dir}/index.html" || \
  fail 'public index is not Persian RTL'
if grep -Fq 'href="/panel' "${source_dir}/index.html" || grep -Fq "href='/panel" "${source_dir}/index.html"; then
  fail 'public index advertises the protected panel'
fi
for marker in breaking-news hero-story latest-news leader-messages politics economy international; do
  grep -Fq "data-section=\"${marker}\"" "${source_dir}/index.html" || fail "missing public section ${marker}"
done

[[ -f "${caddy_file}" ]] || fail 'Caddyfile is missing'
[[ -d "${live_root}" ]] || fail "legacy public root is missing: ${live_root}"
grep -Fq '/var/www/naive' "${caddy_file}" || fail 'Caddyfile does not reference the expected legacy public root'
systemctl is-active --quiet "${caddy_service}" || fail 'caddy-naive.service is not active'
/usr/local/bin/caddy validate --config "${caddy_file}" --adapter caddyfile >/dev/null || fail 'Caddy validation failed'

caddy_sha_before="$(sha256sum /etc/caddy/Caddyfile | awk '{print $1}')"
caddy_pid_before="$(systemctl show "${caddy_service}" --property=MainPID --value)"
caddy_restarts_before="$(systemctl show "${caddy_service}" --property=NRestarts --value)"
[[ "${caddy_pid_before}" =~ ^[1-9][0-9]*$ ]] || fail 'Caddy MainPID is invalid'
[[ "${caddy_restarts_before}" =~ ^[0-9]+$ ]] || fail 'Caddy NRestarts is invalid'

panel_code_before="$(curl --silent --show-error --max-time 8 \
  --resolve "${public_host}:443:127.0.0.1" \
  --output /dev/null --write-out '%{http_code}' \
  "https://${public_host}/panel/" || true)"
case "${panel_code_before}" in
  200|301|302|401|403) ;;
  *) fail "protected panel preflight returned HTTP ${panel_code_before:-unavailable}" ;;
esac

install -d -m 0700 "${backup_root}" "${backup_dir}"
tar --xattrs --acls -C /var/www -czf "${backup_dir}/public-site-before.tar.gz" naive
(
  cd "${backup_dir}"
  sha256sum public-site-before.tar.gz > SHA256SUMS
  sha256sum --check --strict SHA256SUMS >/dev/null
)
chmod -R go-rwx "${backup_dir}"

rm -rf -- "${stage_root}" "${previous_root}" "${failed_root}"
install -d -o root -g caddy -m 0750 "${stage_root}"
cp -a "${source_dir}/." "${stage_root}/"
find "${stage_root}" -type d -exec chmod 0750 {} +
find "${stage_root}" -type f -exec chmod 0640 {} +
chown -R root:caddy "${stage_root}"

# Promote by two same-filesystem directory renames. Caddy configuration is not
# rewritten or reloaded; the legacy pathname remains the serving contract.
mv -- "${live_root}" "${previous_root}"
mv -- "${stage_root}" "${live_root}"
published=1

root_body="$(mktemp)"
root_code="$(curl --silent --show-error --max-time 8 \
  --resolve "${public_host}:443:127.0.0.1" \
  --output "${root_body}" --write-out '%{http_code}' \
  "https://${public_host}/")"
[[ "${root_code}" == "200" ]] || fail "public root returned HTTP ${root_code}"
grep -Fq 'درگاه ایران' "${root_body}" || fail 'public root did not return the promoted page'
rm -f -- "${root_body}"
root_body=""

panel_code_after="$(curl --silent --show-error --max-time 8 \
  --resolve "${public_host}:443:127.0.0.1" \
  --output /dev/null --write-out '%{http_code}' \
  "https://${public_host}/panel/" || true)"
case "${panel_code_after}" in
  200|301|302|401|403) ;;
  *) fail "protected panel post-publish returned HTTP ${panel_code_after:-unavailable}" ;;
esac

caddy_sha_after="$(sha256sum /etc/caddy/Caddyfile | awk '{print $1}')"
caddy_pid_after="$(systemctl show "${caddy_service}" --property=MainPID --value)"
caddy_restarts_after="$(systemctl show "${caddy_service}" --property=NRestarts --value)"
[[ "${caddy_sha_after}" == "${caddy_sha_before}" ]] || fail 'Caddyfile changed during public-site publication'
[[ "${caddy_pid_after}" == "${caddy_pid_before}" ]] || fail 'Caddy MainPID changed during public-site publication'
[[ "${caddy_restarts_after}" == "${caddy_restarts_before}" ]] || fail 'Caddy NRestarts changed during public-site publication'

rm -rf -- "${previous_root}" "${failed_root}"
published=0
trap - ERR HUP INT TERM EXIT

printf '%s\n' \
  'PUBLIC_SITE_RESULT=PASSED' \
  "PUBLIC_SITE_HOST=${public_host}" \
  "PUBLIC_SITE_ROOT=${live_root}" \
  "PUBLIC_SITE_BACKUP=${backup_dir}/public-site-before.tar.gz" \
  "PUBLIC_SITE_ROOT_HTTP=${root_code}" \
  "PUBLIC_SITE_PANEL_HTTP=${panel_code_after}" \
  "CADDYFILE_SHA256=${caddy_sha_after}" \
  "CADDY_MAINPID=${caddy_pid_after}" \
  "CADDY_NRESTARTS=${caddy_restarts_after}"
