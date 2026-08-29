#!/usr/bin/env bash
set -Eeuo pipefail
umask 077

stage_id="S06-OWNER-UPGRADE"
script_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd -P)"
bundle_root="$(cd -- "${script_dir}/../.." && pwd -P)"
expected_caddy_sha="${PVNAIVE_EXPECTED_CADDY_SHA256:-}"
caddy_file="/etc/caddy/Caddyfile"
caddy_bin="/usr/local/bin/caddy"
db_env="/etc/pvnaive/db.env"
api_env="/etc/pvnaive/api.env"
runtime_key="/etc/pvnaive/runtime.key"
runtime_socket="/run/pvnaive/runtime-agent.sock"
preview_root="/var/www/pvnaive-preview/releases"
preview_current="/var/www/pvnaive-preview/current"
stamp="$(date -u +%Y%m%dT%H%M%SZ)"
backup_dir="/var/backups/pvnaive/s06-owner/${stamp}"
web_release_dir=""
preview_release_dir=""
web_before=""
preview_before=""
db_release_before=""
db_backup_path=""
caddy_sha_before=""
caddy_pid_before=""
caddy_restarts_before=""
rollback_started=0

fail() { echo "ERROR: $*" >&2; return 1; }
query_schema() {
  runuser -u postgres -- psql --no-psqlrc --tuples-only --no-align --set ON_ERROR_STOP=1 \
    --host /var/run/postgresql --port "${PVNAIVE_DB_PORT}" --username postgres --dbname pvnaive \
    --command 'SELECT COALESCE(MAX(version),0) FROM pvnaive.schema_migrations'
}

[[ ${EUID} -eq 0 ]] || fail 'run as root'
[[ "${expected_caddy_sha}" =~ ^[0-9a-f]{64}$ ]] || fail 'PVNAIVE_EXPECTED_CADDY_SHA256 must come from S06 preflight'
for cmd in sha256sum systemctl install cp curl psql tar stat file runuser awk sed grep find readlink ln mv ss; do
  command -v "${cmd}" >/dev/null 2>&1 || fail "missing ${cmd}"
done
[[ -x "${caddy_bin}" && -f "${caddy_file}" ]] || fail 'Caddy baseline is missing'
[[ -r "${db_env}" && -r "${api_env}" ]] || fail 'PVNaive environment baseline is missing'
[[ -f "${runtime_key}" && "$(stat -c '%s' "${runtime_key}")" == 32 ]] || fail 'runtime key baseline is invalid'
[[ -f "${bundle_root}/SHA256SUMS" && -f "${bundle_root}/RELEASE.json" ]] || fail 'S06 release metadata is missing'
(
  cd "${bundle_root}"
  sha256sum --check --strict SHA256SUMS >/dev/null
) || fail 'bundle checksum verification failed'
grep -Fq '"stage": "S06-OWNER-CUSTOMER-OPS"' "${bundle_root}/RELEASE.json" || fail 'bundle stage mismatch'
grep -Fq '"base_schema_version": 6' "${bundle_root}/RELEASE.json" || fail 'bundle base schema mismatch'
grep -Fq '"schema_version": 7' "${bundle_root}/RELEASE.json" || fail 'bundle target schema mismatch'
grep -Fq '"usage_accounting_proven": false' "${bundle_root}/RELEASE.json" || fail 'bundle must not claim accounting proof'

for required in \
  bin/pvnaive bin/pvnaive-password bin/pvnaive-runtime-agent web/index.html \
  db/migrations/0007_subscription_token_recovery.up.sql db/migrations/0007_subscription_token_recovery.down.sql \
  db/migrations/SHA256SUMS scripts/db/backup.sh scripts/db/migrate.sh scripts/db/rollback.sh \
  scripts/db/promote-release.sh scripts/db/set-expected-schema-version.sh \
  systemd/pvnaive-api.service systemd/pvnaive-runtime-agent.service; do
  [[ -f "${bundle_root}/${required}" ]] || fail "bundle file missing: ${required}"
done

caddy_sha_before="$(sha256sum "${caddy_file}" | awk '{print $1}')"
[[ "${caddy_sha_before}" == "${expected_caddy_sha}" ]] || fail "Caddyfile changed since preflight: ${caddy_sha_before}"
"${caddy_bin}" validate --config "${caddy_file}" --adapter caddyfile >/dev/null || fail 'Caddy validation failed'
"${caddy_bin}" list-modules | grep -Fx 'http.handlers.forward_proxy' >/dev/null || fail 'forward_proxy module missing'
systemctl is-active --quiet caddy-naive.service || fail 'caddy-naive.service is not active'
caddy_pid_before="$(systemctl show caddy-naive.service --property=MainPID --value)"
caddy_restarts_before="$(systemctl show caddy-naive.service --property=NRestarts --value)"
[[ "${caddy_pid_before}" =~ ^[1-9][0-9]*$ && "${caddy_restarts_before}" =~ ^[0-9]+$ ]] || fail 'invalid Caddy service counters'

systemctl is-active --quiet pvnaive-api.service || fail 'pvnaive-api.service is not active'
systemctl is-active --quiet pvnaive-runtime-agent.service || fail 'pvnaive-runtime-agent.service is not active'
curl --fail --silent --show-error http://127.0.0.1:8080/api/v1/health/ready | grep -q '"ready":true' || fail 'API is not ready before upgrade'
[[ -S "${runtime_socket}" ]] || fail 'runtime agent socket missing before upgrade'
curl --fail --silent --show-error --unix-socket "${runtime_socket}" http://unix/v1/health | grep -q '"status":"ok"' || fail 'runtime agent unhealthy before upgrade'

set -a
# shellcheck disable=SC1091
source "${db_env}"
set +a
[[ "${PVNAIVE_DB_NAME:-}" == pvnaive ]] || fail 'unexpected database name'
[[ "${PVNAIVE_EXPECTED_SCHEMA_VERSION:-}" == 6 ]] || fail 'db.env must expect schema 6 before S06'
schema_before="$(query_schema)"
[[ "${schema_before}" == 6 ]] || fail "S06 requires live schema 6, got ${schema_before}"

web_before="$(readlink -f /opt/pvnaive/web/current 2>/dev/null || true)"
preview_before="$(readlink -f "${preview_current}" 2>/dev/null || true)"
db_release_before="$(readlink -f /opt/pvnaive/db/current 2>/dev/null || true)"
[[ -n "${web_before}" && -d "${web_before}" ]] || fail 'web current baseline invalid'
[[ -n "${preview_before}" && -d "${preview_before}" ]] || fail 'preview current baseline invalid'

install -d -o root -g root -m 0700 "${backup_dir}"
cp -a -- "${caddy_file}" "${backup_dir}/Caddyfile.before"
cp -a -- "${db_env}" "${backup_dir}/db.env.before"
cp -a -- "${api_env}" "${backup_dir}/api.env.before"
cp -a -- /opt/pvnaive/bin/pvnaive "${backup_dir}/pvnaive.before"
cp -a -- /opt/pvnaive/bin/pvnaive-password "${backup_dir}/pvnaive-password.before"
cp -a -- /opt/pvnaive/bin/pvnaive-runtime-agent "${backup_dir}/pvnaive-runtime-agent.before"
cp -a -- /etc/systemd/system/pvnaive-api.service "${backup_dir}/pvnaive-api.service.before"
cp -a -- /etc/systemd/system/pvnaive-runtime-agent.service "${backup_dir}/pvnaive-runtime-agent.service.before"
printf '%s\n' "${web_before}" >"${backup_dir}/web-current.before"
printf '%s\n' "${preview_before}" >"${backup_dir}/preview-current.before"
printf '%s\n' "${db_release_before}" >"${backup_dir}/db-current.before"
printf '%s\n' "${caddy_sha_before}" >"${backup_dir}/caddy-sha.before"
printf '%s\n' "${caddy_pid_before}" >"${backup_dir}/caddy-pid.before"
printf '%s\n' "${caddy_restarts_before}" >"${backup_dir}/caddy-restarts.before"

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
[[ -n "${db_backup_path}" && -f "${db_backup_path}" ]] || fail 'encrypted database backup failed'
grep -Fq '"schema_version": 6,' "$(dirname -- "${db_backup_path}")/metadata.json" || fail 'database backup is not schema 6'
echo "DB_BACKUP_PATH=${db_backup_path}"
echo 'DB_BACKUP=VERIFIED_ENCRYPTED'

rollback() {
  local line="$1" code="$2" current_schema=""
  [[ "${rollback_started}" == 0 ]] || exit "${code}"
  rollback_started=1
  trap - ERR HUP INT TERM
  set +e
  echo 'S06_RESULT=FAILED'
  echo "FAILED_LINE=${line}"
  echo 'ROLLBACK=STARTED'
  systemctl stop pvnaive-api.service >/dev/null 2>&1 || true
  systemctl stop pvnaive-runtime-agent.service >/dev/null 2>&1 || true

  cp -a -- "${backup_dir}/pvnaive.before" /opt/pvnaive/bin/pvnaive
  cp -a -- "${backup_dir}/pvnaive-password.before" /opt/pvnaive/bin/pvnaive-password
  cp -a -- "${backup_dir}/pvnaive-runtime-agent.before" /opt/pvnaive/bin/pvnaive-runtime-agent
  cp -a -- "${backup_dir}/pvnaive-api.service.before" /etc/systemd/system/pvnaive-api.service
  cp -a -- "${backup_dir}/pvnaive-runtime-agent.service.before" /etc/systemd/system/pvnaive-runtime-agent.service
  cp -a -- "${backup_dir}/api.env.before" "${api_env}"
  cp -a -- "${backup_dir}/db.env.before" "${db_env}"

  [[ -z "${web_before}" || ! -d "${web_before}" ]] || ln -sfn -- "${web_before}" /opt/pvnaive/web/current
  if [[ -n "${preview_before}" && -d "${preview_before}" ]]; then
    rm -f -- "${preview_current}.rollback"
    ln -s -- "${preview_before}" "${preview_current}.rollback"
    mv -Tf -- "${preview_current}.rollback" "${preview_current}"
  fi
  [[ -z "${web_release_dir}" ]] || rm -rf -- "${web_release_dir}"
  [[ -z "${preview_release_dir}" ]] || rm -rf -- "${preview_release_dir}"
  [[ -z "${db_release_before}" || ! -d "${db_release_before}" ]] || ln -sfn -- "${db_release_before}" /opt/pvnaive/db/current

  current_schema="$(query_schema 2>/dev/null || true)"
  if [[ "${current_schema}" == 7 && -n "${db_backup_path}" ]]; then
    if PVNAIVE_DB_HOST=/var/run/postgresql \
      PVNAIVE_DB_PORT="${PVNAIVE_DB_PORT}" \
      PVNAIVE_DB_NAME=pvnaive \
      PVNAIVE_DB_USER=postgres \
      PVNAIVE_RUN_AS_OS_USER=postgres \
      PVNAIVE_MIGRATIONS_DIR="${bundle_root}/db/migrations" \
      PVNAIVE_ALLOW_DESTRUCTIVE_ROLLBACK=ROLLBACK_MIGRATION_CHAIN \
      PVNAIVE_ROLLBACK_CHAIN_TARGET_SCHEMA=6 \
      PVNAIVE_CONFIRMED_BACKUP="${db_backup_path}" \
        bash "${bundle_root}/scripts/db/rollback.sh" >/dev/null 2>&1; then
      [[ "$(query_schema 2>/dev/null || true)" == 6 ]] && echo 'ROLLBACK_DATABASE=PASS' || echo 'ROLLBACK_DATABASE=FAIL'
    else
      echo 'ROLLBACK_DATABASE=FAIL'
    fi
  else
    echo 'ROLLBACK_DATABASE=NOT_REQUIRED'
  fi

  systemctl daemon-reload >/dev/null 2>&1 || true
  systemctl restart pvnaive-runtime-agent.service >/dev/null 2>&1 || true
  systemctl restart pvnaive-api.service >/dev/null 2>&1 || true
  final_sha="$(sha256sum "${caddy_file}" 2>/dev/null | awk '{print $1}')"
  final_pid="$(systemctl show caddy-naive.service --property=MainPID --value 2>/dev/null || true)"
  final_restarts="$(systemctl show caddy-naive.service --property=NRestarts --value 2>/dev/null || true)"
  if [[ "${final_sha}" == "${caddy_sha_before}" && "${final_pid}" == "${caddy_pid_before}" && "${final_restarts}" == "${caddy_restarts_before}" ]]; then
    echo 'ROLLBACK_CADDY_INVARIANTS=PASS'
  else
    echo 'ROLLBACK_CADDY_INVARIANTS=FAIL'
  fi
  echo 'ROLLBACK=COMPLETED_BEST_EFFORT'
  exit "${code}"
}
trap 'rollback "${LINENO}" "$?"' ERR
trap 'rollback "${LINENO}" 129' HUP
trap 'rollback "${LINENO}" 130' INT
trap 'rollback "${LINENO}" 143' TERM

source_commit="$(awk -F'"' '$2=="source_commit" {print $4}' "${bundle_root}/RELEASE.json" | head -n1)"
[[ "${source_commit}" =~ ^[0-9a-f]{40}$ ]] || fail 'invalid source commit'
release_id="${source_commit:0:12}"
web_release_dir="/opt/pvnaive/web/releases/${stamp}-${release_id}"
preview_release_dir="${preview_root}/${stamp}-${release_id}"
[[ ! -e "${web_release_dir}" && ! -e "${preview_release_dir}" ]] || fail 'release destination already exists'

install -d -o root -g pvnaive -m 0750 /opt/pvnaive/bin /opt/pvnaive/web/releases
install -d -o root -g pvnaive -m 0750 "${web_release_dir}"
cp -a -- "${bundle_root}/web/." "${web_release_dir}/"
chown -R root:pvnaive "${web_release_dir}"
ln -sfn -- "${web_release_dir}" /opt/pvnaive/web/current

install -d -o root -g caddy -m 0750 "${preview_release_dir}"
cp -a -- "${bundle_root}/web/." "${preview_release_dir}/"
chown -R root:caddy "${preview_release_dir}"
find "${preview_release_dir}" -type d -exec chmod 0750 {} +
find "${preview_release_dir}" -type f -exec chmod 0640 {} +
runuser -u caddy -- test -r "${preview_release_dir}/index.html" || fail 'Caddy cannot read staged panel index'

panel_script="$(sed -nE 's@.*<script[^>]+src="([^"]+\.js)"[^>]*>.*@\1@p' "${preview_release_dir}/index.html" | head -n1)"
panel_css="$(sed -nE 's@.*<link[^>]+href="([^"]+\.css)"[^>]*>.*@\1@p' "${preview_release_dir}/index.html" | head -n1)"
[[ "${panel_script}" == /panel/assets/*.js && "${panel_css}" == /panel/assets/*.css ]] || fail 'panel assets are not under /panel/assets'
runuser -u caddy -- test -r "${preview_release_dir}/${panel_script#/panel/}" || fail 'staged panel JS unreadable'
runuser -u caddy -- test -r "${preview_release_dir}/${panel_css#/panel/}" || fail 'staged panel CSS unreadable'

install -o root -g pvnaive -m 0750 "${bundle_root}/bin/pvnaive" /opt/pvnaive/bin/pvnaive
install -o root -g pvnaive -m 0750 "${bundle_root}/bin/pvnaive-password" /opt/pvnaive/bin/pvnaive-password
install -o root -g pvnaive -m 0750 "${bundle_root}/bin/pvnaive-runtime-agent" /opt/pvnaive/bin/pvnaive-runtime-agent
install -o root -g root -m 0644 "${bundle_root}/systemd/pvnaive-api.service" /etc/systemd/system/pvnaive-api.service
install -o root -g root -m 0644 "${bundle_root}/systemd/pvnaive-runtime-agent.service" /etc/systemd/system/pvnaive-runtime-agent.service

PVNAIVE_DB_HOST=/var/run/postgresql \
PVNAIVE_DB_PORT="${PVNAIVE_DB_PORT}" \
PVNAIVE_DB_NAME=pvnaive \
PVNAIVE_DB_USER=postgres \
PVNAIVE_RUN_AS_OS_USER=postgres \
PVNAIVE_MIGRATIONS_DIR="${bundle_root}/db/migrations" \
  bash "${bundle_root}/scripts/db/migrate.sh" >/dev/null
schema_after="$(query_schema)"
[[ "${schema_after}" == 7 ]] || fail "schema did not reach version 7: ${schema_after}"
PVNAIVE_DB_ENV_FILE="${db_env}" bash "${bundle_root}/scripts/db/set-expected-schema-version.sh" 7 >/dev/null

PVNAIVE_DB_RELEASE_SOURCE_ROOT="${bundle_root}" \
PVNAIVE_DB_RELEASE_ROOT=/opt/pvnaive/db/releases \
PVNAIVE_DB_CURRENT_LINK=/opt/pvnaive/db/current \
PVNAIVE_DB_RELEASE_SCHEMA_VERSION=7 \
PVNAIVE_DB_RELEASE_MIGRATION_FILE=0007_subscription_token_recovery.up.sql \
PVNAIVE_DB_RELEASE_OWNER_USER=root \
PVNAIVE_DB_RELEASE_OWNER_GROUP=pvnaive \
  bash "${bundle_root}/scripts/db/promote-release.sh" >/dev/null

systemctl daemon-reload
systemctl restart pvnaive-runtime-agent.service
systemctl restart pvnaive-api.service
for _ in $(seq 1 30); do
  if [[ -S "${runtime_socket}" ]] && curl --fail --silent --show-error --unix-socket "${runtime_socket}" http://unix/v1/health 2>/dev/null | grep -q '"status":"ok"' && curl --fail --silent --show-error http://127.0.0.1:8080/api/v1/health/ready 2>/dev/null | grep -q '"ready":true'; then
    break
  fi
  sleep 1
done
curl --fail --silent --show-error --unix-socket "${runtime_socket}" http://unix/v1/health | grep -q '"status":"ok"' || fail 'runtime agent health failed after upgrade'
curl --fail --silent --show-error http://127.0.0.1:8080/api/v1/health/ready | grep -q '"ready":true' || fail 'API readiness failed after upgrade'
[[ "$(query_schema)" == 7 ]] || fail 'schema verification failed after service restart'

grep -Fxq 'PVNAIVE_EXPECTED_SCHEMA_VERSION=7' "${db_env}" || fail 'db.env schema expectation was not promoted to 7'
final_sha="$(sha256sum "${caddy_file}" | awk '{print $1}')"
final_pid="$(systemctl show caddy-naive.service --property=MainPID --value)"
final_restarts="$(systemctl show caddy-naive.service --property=NRestarts --value)"
[[ "${final_sha}" == "${caddy_sha_before}" ]] || fail 'Caddyfile changed during S06'
[[ "${final_pid}" == "${caddy_pid_before}" ]] || fail 'Caddy MainPID changed during S06'
[[ "${final_restarts}" == "${caddy_restarts_before}" ]] || fail 'Caddy NRestarts changed during S06'

rm -f -- "${preview_current}.new"
ln -s -- "${preview_release_dir}" "${preview_current}.new"
mv -Tf -- "${preview_current}.new" "${preview_current}"
[[ "$(readlink -f "${preview_current}")" == "${preview_release_dir}" ]] || fail 'panel preview promotion failed'

panel_index="${backup_dir}/panel-index.after"
panel_js="${backup_dir}/panel-js.after"
panel_css_body="${backup_dir}/panel-css.after"
curl --fail --silent --show-error --resolve 'namir.softarg.ir:443:127.0.0.1' 'https://namir.softarg.ir/panel/' -o "${panel_index}"
curl --fail --silent --show-error --resolve 'namir.softarg.ir:443:127.0.0.1' "https://namir.softarg.ir${panel_script}" -o "${panel_js}"
curl --fail --silent --show-error --resolve 'namir.softarg.ir:443:127.0.0.1' "https://namir.softarg.ir${panel_css}" -o "${panel_css_body}"
[[ "$(sha256sum "${panel_index}" | awk '{print $1}')" == "$(sha256sum "${preview_release_dir}/index.html" | awk '{print $1}')" ]] || fail 'served panel index mismatch'
[[ "$(sha256sum "${panel_js}" | awk '{print $1}')" == "$(sha256sum "${preview_release_dir}/${panel_script#/panel/}" | awk '{print $1}')" ]] || fail 'served panel JS mismatch'
[[ "$(sha256sum "${panel_css_body}" | awk '{print $1}')" == "$(sha256sum "${preview_release_dir}/${panel_css#/panel/}" | awk '{print $1}')" ]] || fail 'served panel CSS mismatch'

final_sha="$(sha256sum "${caddy_file}" | awk '{print $1}')"
final_pid="$(systemctl show caddy-naive.service --property=MainPID --value)"
final_restarts="$(systemctl show caddy-naive.service --property=NRestarts --value)"
[[ "${final_sha}" == "${caddy_sha_before}" && "${final_pid}" == "${caddy_pid_before}" && "${final_restarts}" == "${caddy_restarts_before}" ]] || fail 'Caddy invariants changed during panel publication'

trap - ERR HUP INT TERM
echo 'S06_RESULT=PASSED'
echo "SOURCE_COMMIT=${source_commit}"
echo 'SCHEMA_VERSION=7'
echo "DB_BACKUP_PATH=${db_backup_path}"
echo "WEB_RELEASE=${web_release_dir}"
echo "WEB_PREVIEW_RELEASE=${preview_release_dir}"
echo "CADDY_SHA256_AFTER=${final_sha}"
echo "CADDY_MainPID_AFTER=${final_pid}"
echo "CADDY_NRestarts_AFTER=${final_restarts}"
echo 'CADDY_ACTION=none'
echo 'USAGE_ACCOUNTING=NOT_PROVEN'
echo 'FIRST_SUCCESS_CONNECT_PRODUCER=NOT_PROVEN'
