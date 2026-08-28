#!/usr/bin/env bash
set -Eeuo pipefail
umask 077

stage_id="S04R-UPGRADE"
script_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd -P)"
bundle_root="$(cd -- "${script_dir}/../.." && pwd -P)"
expected_caddy_sha="${PVNAIVE_EXPECTED_CADDY_SHA256:-}"
caddy_file="/etc/caddy/Caddyfile"
caddy_bin="/usr/local/bin/caddy"
runtime_key="/etc/pvnaive/runtime.key"
runtime_socket="/run/pvnaive/runtime-agent.sock"
db_env="/etc/pvnaive/db.env"
stamp="$(date -u +%Y%m%dT%H%M%SZ)"
backup_dir="/var/backups/pvnaive/s04r/${stamp}"
web_release_dir=""
web_current_before=""
db_release_before=""
db_backup_path=""
schema_before=""
migration_applied=0
runtime_key_created=0
runtime_unit_existed=0
api_unit_existed=0
api_was_active=0
runtime_was_active=0
rollback_started=0

fail() { echo "ERROR: $*" >&2; exit 1; }

[[ ${EUID} -eq 0 ]] || fail 'run as root'
[[ "${expected_caddy_sha}" =~ ^[0-9a-f]{64}$ ]] || fail 'PVNAIVE_EXPECTED_CADDY_SHA256 must be the 64-character SHA from the read-only preflight'

for command_name in sha256sum systemctl install cp curl psql tar stat file; do
  command -v "${command_name}" >/dev/null 2>&1 || fail "missing ${command_name}"
done
[[ -x "${caddy_bin}" && -f "${caddy_file}" ]] || fail 'Caddy baseline is missing'
[[ -r "${db_env}" ]] || fail 'database environment is missing'
[[ -f "${bundle_root}/SHA256SUMS" ]] || fail 'bundle SHA256SUMS is missing'

(
  cd "${bundle_root}"
  sha256sum --check --strict SHA256SUMS >/dev/null
) || fail 'bundle checksum verification failed'

for required in \
  bin/pvnaive bin/pvnaive-password bin/pvnaive-runtime-agent web/index.html \
  db/migrations/0003_naive_runtime_credentials.up.sql db/migrations/0003_naive_runtime_credentials.down.sql \
  scripts/db/backup.sh scripts/db/migrate.sh scripts/db/rollback.sh scripts/db/promote-release.sh scripts/db/set-expected-schema-version.sh \
  systemd/pvnaive-api.service systemd/pvnaive-runtime-agent.service; do
  [[ -f "${bundle_root}/${required}" ]] || fail "bundle file missing: ${required}"
done

current_caddy_sha="$(sha256sum /etc/caddy/Caddyfile | awk '{print $1}')"
[[ "${current_caddy_sha}" == "${expected_caddy_sha}" ]] || fail "Caddyfile SHA changed since preflight: ${current_caddy_sha}"
/usr/local/bin/caddy validate --config /etc/caddy/Caddyfile --adapter caddyfile >/dev/null || fail 'Caddy validation failed'
"${caddy_bin}" list-modules | grep -Fx 'http.handlers.forward_proxy' >/dev/null || fail 'Caddy forward_proxy module is missing'
caddy_version="$(${caddy_bin} version 2>&1 | head -n1)"
[[ "${caddy_version}" == v2.11.2* ]] || fail "unexpected Caddy version: ${caddy_version}"
systemctl is-active --quiet caddy-naive.service || fail 'caddy-naive.service is not active'
caddy_mainpid_before="$(systemctl show caddy-naive.service --property=MainPID --value)"
caddy_nrestarts_before="$(systemctl show caddy-naive.service --property=NRestarts --value)"
[[ "${caddy_mainpid_before}" =~ ^[1-9][0-9]*$ ]] || fail 'invalid Caddy MainPID'
[[ "${caddy_nrestarts_before}" =~ ^[0-9]+$ ]] || fail 'invalid Caddy NRestarts'
echo "CADDY_SHA256_BEFORE=${current_caddy_sha}"
echo "CADDY_VERSION=${caddy_version}"
echo "CADDY_MainPID_BEFORE=${caddy_mainpid_before}"
echo "CADDY_NRestarts_BEFORE=${caddy_nrestarts_before}"

listener_before="$(ss -H -lntp 2>/dev/null || true)"
for port in 22 80 443; do
  awk -v p=":${port}" '$4 ~ p"$" {found=1} END {exit !found}' <<<"${listener_before}" || fail "required TCP port ${port} is not listening"
done

set -a
# shellcheck disable=SC1091
source "${db_env}"
set +a
[[ "${PVNAIVE_DB_NAME:-}" == "pvnaive" ]] || fail 'unexpected database name'
[[ "${PVNAIVE_DB_PORT:-}" =~ ^[0-9]+$ ]] || fail 'invalid database port'
schema_before="$(runuser -u postgres -- psql --no-psqlrc --tuples-only --no-align --set ON_ERROR_STOP=1 \
  --host /var/run/postgresql --port "${PVNAIVE_DB_PORT}" --username postgres --dbname pvnaive \
  --command 'SELECT COALESCE(MAX(version),0) FROM pvnaive.schema_migrations')"
[[ "${schema_before}" == "2" || "${schema_before}" == "3" ]] || fail "unsupported schema before S04R: ${schema_before}"

api_unit_existed=0
runtime_unit_existed=0
[[ -f /etc/systemd/system/pvnaive-api.service ]] && api_unit_existed=1
[[ -f /etc/systemd/system/pvnaive-runtime-agent.service ]] && runtime_unit_existed=1
if systemctl is-active --quiet pvnaive-api.service; then api_was_active=1; fi
if systemctl is-active --quiet pvnaive-runtime-agent.service; then runtime_was_active=1; fi
web_current_before="$(readlink -f /opt/pvnaive/web/current 2>/dev/null || true)"
db_release_before="$(readlink -f /opt/pvnaive/db/current 2>/dev/null || true)"

install -d -o root -g root -m 0700 "${backup_dir}"
cp -a -- "${caddy_file}" "${backup_dir}/Caddyfile.before"
[[ ! -f /opt/pvnaive/bin/pvnaive ]] || cp -a -- /opt/pvnaive/bin/pvnaive "${backup_dir}/pvnaive.before"
[[ ! -f /opt/pvnaive/bin/pvnaive-password ]] || cp -a -- /opt/pvnaive/bin/pvnaive-password "${backup_dir}/pvnaive-password.before"
[[ ! -f /opt/pvnaive/bin/pvnaive-runtime-agent ]] || cp -a -- /opt/pvnaive/bin/pvnaive-runtime-agent "${backup_dir}/pvnaive-runtime-agent.before"
[[ "${api_unit_existed}" == "0" ]] || cp -a -- /etc/systemd/system/pvnaive-api.service "${backup_dir}/pvnaive-api.service.before"
[[ "${runtime_unit_existed}" == "0" ]] || cp -a -- /etc/systemd/system/pvnaive-runtime-agent.service "${backup_dir}/pvnaive-runtime-agent.service.before"
printf '%s\n' "${web_current_before}" >"${backup_dir}/web-current.before"
printf '%s\n' "${db_release_before}" >"${backup_dir}/db-current.before"
printf '%s\n' "${current_caddy_sha}" >"${backup_dir}/caddy-sha256.before"
printf '%s\n' "${caddy_mainpid_before}" >"${backup_dir}/caddy-mainpid.before"
printf '%s\n' "${caddy_nrestarts_before}" >"${backup_dir}/caddy-nrestarts.before"
echo "BACKUP_DIR=${backup_dir}"

backup_output="$(
  PVNAIVE_DB_HOST=/var/run/postgresql \
  PVNAIVE_DB_PORT="${PVNAIVE_DB_PORT}" \
  PVNAIVE_DB_NAME=pvnaive \
  PVNAIVE_DB_USER=postgres \
  PVNAIVE_RUN_AS_OS_USER=postgres \
    bash "${bundle_root}/scripts/db/backup.sh"
)"
db_backup_path="$(awk -F= '$1=="PVNAIVE_BACKUP_PATH" {print $2}' <<<"${backup_output}")"
[[ -n "${db_backup_path}" && -f "${db_backup_path}" ]] || fail 'encrypted pre-upgrade database backup failed'
echo 'DB_BACKUP=VERIFIED_ENCRYPTED'

rollback_on_error() {
  local line="$1" code="$2"
  [[ "${rollback_started}" == "0" ]] || exit "${code}"
  rollback_started=1
  trap - ERR HUP INT TERM
  set +e
  echo "S04R_RESULT=FAILED"
  echo "FAILED_LINE=${line}"
  echo "ROLLBACK=STARTED"

  systemctl stop pvnaive-runtime-agent.service >/dev/null 2>&1 || true
  if [[ "${runtime_unit_existed}" == "1" && -f "${backup_dir}/pvnaive-runtime-agent.service.before" ]]; then
    cp -a -- "${backup_dir}/pvnaive-runtime-agent.service.before" /etc/systemd/system/pvnaive-runtime-agent.service
  else
    rm -f -- /etc/systemd/system/pvnaive-runtime-agent.service
  fi
  if [[ "${api_unit_existed}" == "1" && -f "${backup_dir}/pvnaive-api.service.before" ]]; then
    cp -a -- "${backup_dir}/pvnaive-api.service.before" /etc/systemd/system/pvnaive-api.service
  fi

  for name in pvnaive pvnaive-password pvnaive-runtime-agent; do
    if [[ -f "${backup_dir}/${name}.before" ]]; then
      cp -a -- "${backup_dir}/${name}.before" "/opt/pvnaive/bin/${name}"
    elif [[ "${name}" == "pvnaive-runtime-agent" ]]; then
      rm -f -- "/opt/pvnaive/bin/${name}"
    fi
  done

  if [[ -n "${web_current_before}" && -d "${web_current_before}" ]]; then
    ln -sfn -- "${web_current_before}" /opt/pvnaive/web/current
  fi
  [[ -z "${web_release_dir}" ]] || rm -rf -- "${web_release_dir}"

  if [[ "${migration_applied}" == "1" && -n "${db_backup_path}" ]]; then
    PVNAIVE_DB_HOST=/var/run/postgresql \
    PVNAIVE_DB_PORT="${PVNAIVE_DB_PORT}" \
    PVNAIVE_DB_NAME=pvnaive \
    PVNAIVE_DB_USER=postgres \
    PVNAIVE_RUN_AS_OS_USER=postgres \
    PVNAIVE_MIGRATIONS_DIR="${bundle_root}/db/migrations" \
    PVNAIVE_ALLOW_DESTRUCTIVE_ROLLBACK=ROLLBACK_ONE_MIGRATION \
    PVNAIVE_CONFIRMED_BACKUP="${db_backup_path}" \
      bash "${bundle_root}/scripts/db/rollback.sh" >/dev/null 2>&1 || true
    PVNAIVE_DB_ENV_FILE="${db_env}" bash "${bundle_root}/scripts/db/set-expected-schema-version.sh" 2 >/dev/null 2>&1 || true
  fi
  if [[ -n "${db_release_before}" && -d "${db_release_before}" ]]; then
    ln -sfn -- "${db_release_before}" /opt/pvnaive/db/current
  fi

  if [[ "${runtime_key_created}" == "1" ]]; then
    rm -f -- "${runtime_key}"
  fi
  systemctl daemon-reload >/dev/null 2>&1 || true
  [[ "${runtime_was_active}" == "0" ]] || systemctl start pvnaive-runtime-agent.service >/dev/null 2>&1 || true
  [[ "${api_was_active}" == "0" ]] || systemctl restart pvnaive-api.service >/dev/null 2>&1 || true

  final_caddy_sha="$(sha256sum /etc/caddy/Caddyfile 2>/dev/null | awk '{print $1}')"
  final_caddy_pid="$(systemctl show caddy-naive.service --property=MainPID --value 2>/dev/null || true)"
  final_caddy_restarts="$(systemctl show caddy-naive.service --property=NRestarts --value 2>/dev/null || true)"
  if [[ "${final_caddy_sha}" == "${current_caddy_sha}" && "${final_caddy_pid}" == "${caddy_mainpid_before}" && "${final_caddy_restarts}" == "${caddy_nrestarts_before}" ]]; then
    echo 'ROLLBACK_CADDY_INVARIANTS=PASS'
  else
    echo 'ROLLBACK_CADDY_INVARIANTS=FAIL'
  fi
  echo 'ROLLBACK=COMPLETED_BEST_EFFORT'
  exit "${code}"
}
trap 'rollback_on_error "${LINENO}" "$?"' ERR
trap 'rollback_on_error "${LINENO}" 129' HUP
trap 'rollback_on_error "${LINENO}" 130' INT
trap 'rollback_on_error "${LINENO}" 143' TERM

install -d -o root -g pvnaive -m 0750 /opt/pvnaive/bin /opt/pvnaive/runtime/releases /opt/pvnaive/web/releases
source_commit="$(awk -F'"' '$2=="source_commit" {print $4}' "${bundle_root}/RELEASE.json" | head -n1)"
[[ "${source_commit}" =~ ^[0-9a-f]{40}$ ]] || fail 'invalid bundle source commit'
release_id="${source_commit:0:12}"
web_release_dir="/opt/pvnaive/web/releases/${stamp}-${release_id}"
install -d -o root -g pvnaive -m 0750 "${web_release_dir}"
cp -a -- "${bundle_root}/web/." "${web_release_dir}/"
chown -R root:pvnaive "${web_release_dir}"
ln -sfn -- "${web_release_dir}" /opt/pvnaive/web/current

install -o root -g pvnaive -m 0750 "${bundle_root}/bin/pvnaive" /opt/pvnaive/bin/pvnaive
install -o root -g pvnaive -m 0750 "${bundle_root}/bin/pvnaive-password" /opt/pvnaive/bin/pvnaive-password
install -o root -g pvnaive -m 0750 "${bundle_root}/bin/pvnaive-runtime-agent" /opt/pvnaive/bin/pvnaive-runtime-agent

if [[ -e "${runtime_key}" ]]; then
  [[ "$(stat -c '%s' "${runtime_key}")" == "32" ]] || fail 'existing runtime key must be exactly 32 bytes'
  [[ "$(stat -c '%U:%G:%a' "${runtime_key}")" == "root:pvnaive:640" ]] || fail 'existing runtime key ownership/mode must be root:pvnaive 0640'
  echo 'RUNTIME_KEY=preserved'
else
  key_temp="${backup_dir}/runtime.key.new"
  head -c 32 /dev/urandom >"${key_temp}"
  install -o root -g pvnaive -m 0640 "${key_temp}" "${runtime_key}"
  rm -f -- "${key_temp}"
  runtime_key_created=1
  echo 'RUNTIME_KEY=created'
fi

if [[ "${schema_before}" == "2" ]]; then
  PVNAIVE_DB_HOST=/var/run/postgresql \
  PVNAIVE_DB_PORT="${PVNAIVE_DB_PORT}" \
  PVNAIVE_DB_NAME=pvnaive \
  PVNAIVE_DB_USER=postgres \
  PVNAIVE_RUN_AS_OS_USER=postgres \
  PVNAIVE_MIGRATIONS_DIR="${bundle_root}/db/migrations" \
    bash "${bundle_root}/scripts/db/migrate.sh" >/dev/null
  migration_applied=1
fi
schema_after="$(runuser -u postgres -- psql --no-psqlrc --tuples-only --no-align --set ON_ERROR_STOP=1 \
  --host /var/run/postgresql --port "${PVNAIVE_DB_PORT}" --username postgres --dbname pvnaive \
  --command 'SELECT COALESCE(MAX(version),0) FROM pvnaive.schema_migrations')"
[[ "${schema_after}" == "3" ]] || fail "schema did not reach version 3: ${schema_after}"
PVNAIVE_DB_ENV_FILE="${db_env}" bash "${bundle_root}/scripts/db/set-expected-schema-version.sh" 3 >/dev/null

PVNAIVE_DB_RELEASE_SOURCE_ROOT="${bundle_root}" \
PVNAIVE_DB_RELEASE_ROOT=/opt/pvnaive/db/releases \
PVNAIVE_DB_CURRENT_LINK=/opt/pvnaive/db/current \
PVNAIVE_DB_RELEASE_SCHEMA_VERSION=3 \
PVNAIVE_DB_RELEASE_MIGRATION_FILE=0003_naive_runtime_credentials.up.sql \
PVNAIVE_DB_RELEASE_OWNER_USER=root \
PVNAIVE_DB_RELEASE_OWNER_GROUP=pvnaive \
  bash "${bundle_root}/scripts/db/promote-release.sh" >/dev/null

install -o root -g root -m 0644 "${bundle_root}/systemd/pvnaive-runtime-agent.service" /etc/systemd/system/pvnaive-runtime-agent.service
install -o root -g root -m 0644 "${bundle_root}/systemd/pvnaive-api.service" /etc/systemd/system/pvnaive-api.service
systemctl daemon-reload
systemctl enable --now pvnaive-runtime-agent.service
systemctl restart pvnaive-api.service

for _ in $(seq 1 30); do
  [[ -S "${runtime_socket}" ]] && curl --fail --silent --show-error --unix-socket "${runtime_socket}" http://unix/v1/health 2>/dev/null | grep -q '"status":"ok"' && break
  sleep 1
done
[[ -S "${runtime_socket}" ]] || fail 'runtime-agent.sock was not created'
curl --fail --silent --show-error --unix-socket "${runtime_socket}" http://unix/v1/health | grep -q '"status":"ok"' || fail 'runtime agent health failed'
curl --fail --silent --show-error http://127.0.0.1:8080/api/v1/health/ready | grep -q '"ready":true' || fail 'API readiness failed after S04R upgrade'

final_caddy_sha="$(sha256sum /etc/caddy/Caddyfile | awk '{print $1}')"
final_caddy_pid="$(systemctl show caddy-naive.service --property=MainPID --value)"
final_caddy_restarts="$(systemctl show caddy-naive.service --property=NRestarts --value)"
[[ "${final_caddy_sha}" == "${current_caddy_sha}" ]] || fail 'Caddyfile changed during S04R upgrade'
[[ "${final_caddy_pid}" == "${caddy_mainpid_before}" ]] || fail 'Caddy MainPID changed during S04R upgrade'
[[ "${final_caddy_restarts}" == "${caddy_nrestarts_before}" ]] || fail 'Caddy NRestarts changed during S04R upgrade'
/usr/local/bin/caddy validate --config /etc/caddy/Caddyfile --adapter caddyfile >/dev/null || fail 'Caddy validation failed after S04R upgrade'

trap - ERR HUP INT TERM
echo "S04R_RESULT=PASSED"
echo "SOURCE_COMMIT=${source_commit}"
echo "SCHEMA_VERSION=${schema_after}"
echo "CADDY_SHA256_AFTER=${final_caddy_sha}"
echo "CADDY_MainPID_AFTER=${final_caddy_pid}"
echo "CADDY_NRestarts_AFTER=${final_caddy_restarts}"
echo 'CADDY_ACTION=none'
echo 'NEXT_ACTION=Owner login and guarded live credential import; do not mutate credentials until import equivalence passes.'