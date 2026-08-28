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
  'ROLLBACK_DATABASE=PASS' \
  'ROLLBACK_DATABASE=FAIL' \
  'S04R_RESULT=PASSED'; do
  grep -Fq -- "${required}" "${target}" || { echo "ERROR: upgrade missing contract token: ${required}" >&2; exit 1; }
done

# After the ERR trap is installed, fail() must return a non-zero status instead
# of calling exit directly. A direct exit bypasses ERR and leaves a partial
# production upgrade without executing rollback_on_error.
if grep -Eq '^fail\(\)[[:space:]]*\{.*exit[[:space:]]+1' "${target}"; then
  echo 'ERROR: upgrade fail() must not bypass ERR rollback with exit 1' >&2
  exit 1
fi
grep -Eq '^fail\(\)[[:space:]]*\{.*return[[:space:]]+1' "${target}" || {
  echo 'ERROR: upgrade fail() must return 1 so ERR trap can execute rollback' >&2
  exit 1
}

if grep -Eq 'systemctl[[:space:]]+(restart|reload)[[:space:]]+caddy-naive\.service' "${target}"; then
  echo 'ERROR: upgrade must never restart or reload Caddy' >&2
  exit 1
fi
if grep -Eq '(cat|xxd|base64|od)[[:space:]].*/etc/pvnaive/runtime\.key' "${target}"; then
  echo 'ERROR: upgrade may not print runtime key material' >&2
  exit 1
fi

# Replacing /opt/pvnaive/bin/pvnaive-runtime-agent while the service is already
# active is not enough: enable --now leaves the old process running. Every
# successful S04R upgrade must explicitly restart only the runtime agent so the
# newly installed binary is the process serving the Unix socket.
grep -Fq 'systemctl restart pvnaive-runtime-agent.service' "${target}" || {
  echo 'ERROR: upgrade must restart pvnaive-runtime-agent.service after installing its binary' >&2
  exit 1
}
runtime_binary_line="$(grep -n -F '"${bundle_root}/bin/pvnaive-runtime-agent" /opt/pvnaive/bin/pvnaive-runtime-agent' "${target}" | head -n1 | cut -d: -f1)"
runtime_restart_line="$(grep -n -F 'systemctl restart pvnaive-runtime-agent.service' "${target}" | head -n1 | cut -d: -f1)"
[[ -n "${runtime_binary_line}" && -n "${runtime_restart_line}" && "${runtime_binary_line}" -lt "${runtime_restart_line}" ]] || {
  echo 'ERROR: runtime agent binary installation must precede runtime agent restart' >&2
  exit 1
}

backup_line="$(grep -n -F 'bash "${bundle_root}/scripts/db/backup.sh"' "${target}" | head -n1 | cut -d: -f1)"
migrate_line="$(grep -n -F 'bash "${bundle_root}/scripts/db/migrate.sh"' "${target}" | head -n1 | cut -d: -f1)"
[[ -n "${backup_line}" && -n "${migrate_line}" && "${backup_line}" -lt "${migrate_line}" ]] || {
  echo 'ERROR: database backup must precede migration' >&2
  exit 1
}

bash tests/stages/S04R_module_pipefail_regression_test.sh

echo 'S04R_UPGRADE_CONTRACT=PASSED'
