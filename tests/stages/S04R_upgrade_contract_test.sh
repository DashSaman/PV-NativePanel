#!/usr/bin/env bash
set -Eeuo pipefail

target="scripts/stages/S04R-upgrade.sh"
[[ -f "${target}" ]] || { echo "ERROR: missing ${target}" >&2; exit 1; }
bash -n "${target}"

for required in \
  'PVNAIVE_EXPECTED_CADDY_SHA256' \
  'sha256sum /etc/caddy/Caddyfile' \
  '/usr/local/bin/caddy validate' \
  'systemctl show caddy-naive.service' \
  'MainPID' \
  'NRestarts' \
  'scripts/db/backup.sh' \
  '/etc/pvnaive/runtime.key' \
  'pvnaive-runtime-agent.service' \
  'pvnaive-api.service' \
  'runtime-agent.sock' \
  'S04R_RESULT=PASSED'; do
  grep -Fq -- "${required}" "${target}" || { echo "ERROR: upgrade missing contract token: ${required}" >&2; exit 1; }
done

if grep -Eq 'systemctl[[:space:]]+(restart|reload)[[:space:]]+caddy-naive\.service' "${target}"; then
  echo 'ERROR: upgrade must never restart or reload Caddy' >&2
  exit 1
fi
if grep -Eq '(cat|xxd|base64|od)[[:space:]].*/etc/pvnaive/runtime\.key' "${target}"; then
  echo 'ERROR: upgrade may not print runtime key material' >&2
  exit 1
fi

backup_line="$(grep -n -F 'bash "${bundle_root}/scripts/db/backup.sh"' "${target}" | head -n1 | cut -d: -f1)"
migrate_line="$(grep -n -F 'bash "${bundle_root}/scripts/db/migrate.sh"' "${target}" | head -n1 | cut -d: -f1)"
[[ -n "${backup_line}" && -n "${migrate_line}" && "${backup_line}" -lt "${migrate_line}" ]] || {
  echo 'ERROR: database backup must precede migration' >&2
  exit 1
}

echo 'S04R_UPGRADE_CONTRACT=PASSED'
