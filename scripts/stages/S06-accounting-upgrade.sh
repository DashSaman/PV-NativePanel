#!/usr/bin/env bash
set -Eeuo pipefail
umask 077

stage_id="S06-ACCOUNTING-UPGRADE"
script_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd -P)"
bundle_root="$(cd -- "${script_dir}/../.." && pwd -P)"
expected_caddy_sha="${PVNAIVE_EXPECTED_CADDY_SHA256:-}"
accounting_caddy_sha="${PVNAIVE_ACCOUNTING_CADDY_SHA256:-}"
caddy_file="/etc/caddy/Caddyfile"
caddy_bin="/usr/local/bin/caddy"
caddy_unit="caddy-naive.service"
db_env="/etc/pvnaive/db.env"
api_env="/etc/pvnaive/api.env"
runtime_key="/etc/pvnaive/runtime.key"
runtime_socket="/run/pvnaive/runtime-agent.sock"
telemetry_socket="/run/pvnaive/accounting.sock"

S06_ACCOUNTING_BASE_SCHEMA=16
S06_ACCOUNTING_TARGET_SCHEMA=17
PVNAIVE_DB_RELEASE_SCHEMA_VERSION=17
PVNAIVE_DB_RELEASE_MIGRATION_FILE=0017_operator_session_peers.up.sql
PVNAIVE_EXPECTED_SCHEMA_VERSION=17
PVNAIVE_ROLLBACK_CHAIN_TARGET_SCHEMA=16

stamp="$(date -u +%Y%m%dT%H%M%SZ)"
backup_dir="/var/backups/pvnaive/s06-accounting/${stamp}"

caddy_sha_before=""
caddy_bin_sha_before=""
caddy_pid_before=""
caddy_restarts_before=""
db_backup_path=""
schema_before=""
rollback_started=0
caddy_activated=0

fail() { echo "ERROR: $*" >&2; return 1; }
query_schema() {
  runuser -u postgres -- psql --no-psqlrc --tuples-only --no-align --set ON_ERROR_STOP=1 \
    --host /var/run/postgresql --port "${PVNAIVE_DB_PORT}" --username postgres --dbname pvnaive \
    --command 'SELECT COALESCE(MAX(version),0) FROM pvnaive.schema_migrations'
}

[[ ${EUID} -eq 0 ]] || fail 'run as root'
[[ "${expected_caddy_sha}" =~ ^[0-9a-f]{64}$ ]] || fail 'PVNAIVE_EXPECTED_CADDY_SHA256 must come from S06 accounting preflight'
[[ "${accounting_caddy_sha}" =~ ^[0-9a-f]{64}$ ]] || fail 'PVNAIVE_ACCOUNTING_CADDY_SHA256 must come from the accounting Caddy candidate provenance'
for cmd in sha256sum systemctl install cp curl psql tar stat file runuser awk sed grep find readlink ln mv ss; do
  command -v "${cmd}" >/dev/null 2>&1 || fail "missing ${cmd}"
done
[[ -x "${caddy_bin}" && -f "${caddy_file}" ]] || fail 'Caddy baseline is missing'
[[ -r "${db_env}" && -r "${api_env}" ]] || fail 'PVNaive environment baseline is missing'
[[ -f "${runtime_key}" && "$(stat -c '%s' "${runtime_key}")" == 32 ]] || fail 'runtime key baseline is invalid'
[[ -f "${bundle_root}/SHA256SUMS" && -f "${bundle_root}/RELEASE.json" ]] || fail 'S06 accounting release metadata is missing'
(
  cd "${bundle_root}"
  sha256sum --check --strict SHA256SUMS >/dev/null
) || fail 'bundle checksum verification failed'

grep -Fq '"stage": "S06-ACCOUNTING-RELEASE"' "${bundle_root}/RELEASE.json" || fail 'bundle stage mismatch'
grep -Fq '"base_schema_version": 16' "${bundle_root}/RELEASE.json" || fail 'bundle base schema mismatch'
grep -Fq '"schema_version": 17' "${bundle_root}/RELEASE.json" || fail 'bundle target schema mismatch'
grep -Fq '"usage_accounting_proven": false' "${bundle_root}/RELEASE.json" || fail 'bundle must not claim accounting proof'
"${bundle_root}/scripts/stages/S06-accounting-preflight.sh" >/dev/null || fail 'S06 accounting preflight did not pass'

for required in \
  bin/pvnaive bin/pvnaive-password bin/pvnaive-runtime-agent bin/pvnaive-telemetry-agent web/index.html \
  caddy/caddy-pvnaive-accounting caddy/caddy-pvnaive-accounting.sha256 \
  PROVENANCE.txt \
  db/migrations/0017_operator_session_peers.up.sql db/migrations/0017_operator_session_peers.down.sql \
  db/migrations/SHA256SUMS scripts/db/backup.sh scripts/db/migrate.sh scripts/db/rollback.sh \
  scripts/db/promote-release.sh scripts/db/set-expected-schema-version.sh \
  systemd/pvnaive-api.service systemd/pvnaive-runtime-agent.service systemd/pvnaive-telemetry-agent.service; do
  [[ -f "${bundle_root}/${required}" ]] || fail "bundle file missing: ${required}"
done

candidate_caddy_sha="$(sha256sum "${bundle_root}/caddy/caddy-pvnaive-accounting" | awk '{print $1}')"
[[ "${candidate_caddy_sha}" == "${accounting_caddy_sha}" ]] || fail "bundle candidate SHA ${candidate_caddy_sha} does not match manifest ${accounting_caddy_sha}"
grep -Fq "${accounting_caddy_sha}  caddy-pvnaive-accounting" "${bundle_root}/caddy/caddy-pvnaive-accounting.sha256" || fail 'bundle candidate checksum manifest mismatch'
grep -Fq "binary_sha256=${accounting_caddy_sha}" "${bundle_root}/PROVENANCE.txt" || fail 'bundle candidate provenance does not match manifest SHA'
grep -Fq 'reproducibility_verified=true' "${bundle_root}/PROVENANCE.txt" || fail 'bundle candidate provenance lacks reproducibility proof'
"${bundle_root}/caddy/caddy-pvnaive-accounting" validate --config "${caddy_file}" --adapter caddyfile >/dev/null || fail 'bundle candidate failed config validation'
"${bundle_root}/caddy/caddy-pvnaive-accounting" list-modules | grep -Fx 'http.handlers.forward_proxy' >/dev/null || fail 'bundle candidate forward_proxy module missing'
"${bundle_root}/caddy/caddy-pvnaive-accounting" list-modules | grep -Fq 'pvnaive_accounting' >/dev/null || fail 'bundle candidate accounting module missing'

caddy_sha_before="$(sha256sum "${caddy_file}" | awk '{print $1}')"
[[ "${caddy_sha_before}" == "${expected_caddy_sha}" ]] || fail "Caddyfile changed since preflight: ${caddy_sha_before}"
"${caddy_bin}" validate --config "${caddy_file}" --adapter caddyfile >/dev/null || fail 'Caddy validation failed'
systemctl is-active --quiet "${caddy_unit}" || fail 'caddy-naive.service is not active'
caddy_bin_sha_before="$(sha256sum "${caddy_bin}" | awk '{print $1}')"
caddy_pid_before="$(systemctl show "${caddy_unit}" --property=MainPID --value)"
caddy_restarts_before="$(systemctl show "${caddy_unit}" --property=NRestarts --value)"
[[ "${caddy_bin_sha_before}" =~ ^[0-9a-f]{64}$ ]] || fail 'invalid Caddy binary SHA before upgrade'
[[ "${caddy_pid_before}" =~ ^[1-9][0-9]*$ && "${caddy_restarts_before}" =~ ^[0-9]+$ ]] || fail 'invalid Caddy service counters'

systemctl is-active --quiet pvnaive-api.service || fail 'pvnaive-api.service is not active'
systemctl is-active --quiet pvnaive-runtime-agent.service || fail 'pvnaive-runtime-agent.service is not active'
systemctl is-active --quiet pvnaive-telemetry-agent.service || fail 'pvnaive-telemetry-agent.service is not active'
curl --fail --silent --show-error http://127.0.0.1:8080/api/v1/health/ready | grep -q '"ready":true' || fail 'API is not ready before upgrade'
[[ -S "${runtime_socket}" ]] || fail 'runtime agent socket missing before upgrade'
curl --fail --silent --show-error --unix-socket "${runtime_socket}" http://unix/v1/health | grep -q '"status":"ok"' || fail 'runtime agent unhealthy before upgrade'
[[ -S "${telemetry_socket}" ]] || fail 'telemetry accounting socket missing before upgrade'
curl --fail --silent --show-error --unix-socket "${telemetry_socket}" http://unix/v1/accounting/health | grep -q '"status":"ok"' || fail 'telemetry agent unhealthy before upgrade'

set -a
# shellcheck disable=SC1091
source "${db_env}"
set +a
[[ "${PVNAIVE_DB_NAME:-}" == pvnaive ]] || fail 'unexpected database name'
[[ "${PVNAIVE_EXPECTED_SCHEMA_VERSION:-}" == "${S06_ACCOUNTING_BASE_SCHEMA}" ]] || fail 'db.env must expect schema 16 before S06 accounting upgrade'
schema_before="$(query_schema)"
[[ "${schema_before}" == "${S06_ACCOUNTING_BASE_SCHEMA}" ]] || fail "S06 accounting upgrade requires live schema 16, got ${schema_before}"

install -d -o root -g root -m 0700 "${backup_dir}"
cp -a -- "${caddy_file}" "${backup_dir}/Caddyfile.before"
cp -a -- "${caddy_bin}" "${backup_dir}/caddy-before"
cp -a -- "${db_env}" "${backup_dir}/db.env.before"
cp -a -- "${api_env}" "${backup_dir}/api.env.before"
cp -a -- /opt/pvnaive/bin/pvnaive "${backup_dir}/pvnaive.before"
cp -a -- /opt/pvnaive/bin/pvnaive-password "${backup_dir}/pvnaive-password.before"
cp -a -- /opt/pvnaive/bin/pvnaive-runtime-agent "${backup_dir}/pvnaive-runtime-agent.before"
cp -a -- /opt/pvnaive/bin/pvnaive-telemetry-agent "${backup_dir}/pvnaive-telemetry-agent.before"
cp -a -- /etc/systemd/system/pvnaive-api.service "${backup_dir}/pvnaive-api.service.before"
cp -a -- /etc/systemd/system/pvnaive-runtime-agent.service "${backup_dir}/pvnaive-runtime-agent.service.before"
cp -a -- /etc/systemd/system/pvnaive-telemetry-agent.service "${backup_dir}/pvnaive-telemetry-agent.service.before"
web_current_before="$(readlink -f /opt/pvnaive/web/current)"
db_current_before="$(readlink -f /opt/pvnaive/db/current)"
printf '%s\n' "${web_current_before}" >"${backup_dir}/web-current.before"
printf '%s\n' "${db_current_before}" >"${backup_dir}/db-current.before"
printf '%s\n' "${caddy_sha_before}" >"${backup_dir}/caddy-sha.before"
printf '%s\n' "${caddy_bin_sha_before}" >"${backup_dir}/caddy-binary-sha.before"
printf '%s\n' "${caddy_pid_before}" >"${backup_dir}/caddy-pid.before"
printf '%s\n' "${caddy_restarts_before}" >"${backup_dir}/caddy-restarts.before"
printf '%s\n' "${schema_before}" >"${backup_dir}/schema-before"

echo "BACKUP_DIR=${backup_dir}"
echo 'CADDY_UNIT=caddy-naive.service'
echo "CADDY_BINARY_SHA256_BEFORE=${caddy_bin_sha_before}"
echo 'CADDY_BINARY_BACKUP=caddy-before'
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
grep -Fq "\"schema_version\": ${S06_ACCOUNTING_BASE_SCHEMA}," "$(dirname -- "${db_backup_path}")/metadata.json" || fail 'database backup is not schema 16'
echo "DB_BACKUP_PATH=${db_backup_path}"
echo 'DB_BACKUP=VERIFIED_ENCRYPTED'

rollback() {
  local line="$1" code="$2" current_schema=""
  [[ "${rollback_started}" == 0 ]] || exit "${code}"
  rollback_started=1
  trap - ERR HUP INT TERM
  set +e
  echo 'S06_ACCOUNTING_RESULT=FAILED'
  echo "FAILED_LINE=${line}"
  echo 'ROLLBACK=STARTED'

  # Once the new Caddy has been activated it may already have created immutable
  # schema17 trusted-peer evidence. First remove the producer by restoring the
  # prior Caddy binary, then inspect evidence. If evidence exists, preserve the
  # additive schema17 + new control-plane binaries instead of attempting the
  # intentionally forbidden destructive 17->16 rollback.
  trusted_peer_evidence=0
  if [[ "${caddy_activated}" == 1 ]]; then
    cp -a -- "${backup_dir}/caddy-before" "${caddy_bin}"
    chmod 0755 "${caddy_bin}"
    systemctl restart "${caddy_unit}" >/dev/null 2>&1 || true
    caddy_activated=0
    current_schema="$(query_schema 2>/dev/null || true)"
    if [[ "${current_schema}" == "${S06_ACCOUNTING_TARGET_SCHEMA}" ]]; then
      trusted_peer_evidence="$(runuser -u postgres -- psql --no-psqlrc --tuples-only --no-align --set ON_ERROR_STOP=1 \
        --host /var/run/postgresql --port "${PVNAIVE_DB_PORT}" --username postgres --dbname pvnaive \
        --command "SELECT CASE WHEN to_regclass('pvnaive.direct_naive_accounting_session_peers') IS NULL THEN 0 ELSE (SELECT count(*) FROM pvnaive.direct_naive_accounting_session_peers) END" 2>/dev/null || echo 1)"
      trusted_peer_evidence="${trusted_peer_evidence//[[:space:]]/}"
    fi
    if [[ "${trusted_peer_evidence}" =~ ^[1-9][0-9]*$ ]]; then
      echo 'ROLLBACK_CADDY_BINARY=RESTORED'
      echo 'ROLLBACK_DATABASE=PRESERVED_TRUSTED_PEER_EVIDENCE'
      echo "TRUSTED_PEER_EVIDENCE_ROWS=${trusted_peer_evidence}"
      echo 'ROLLBACK_MODE=POST_ACTIVATION_EVIDENCE_PRESERVING'
      systemctl restart pvnaive-telemetry-agent.service >/dev/null 2>&1 || true
      systemctl restart pvnaive-runtime-agent.service >/dev/null 2>&1 || true
      systemctl restart pvnaive-api.service >/dev/null 2>&1 || true
      echo 'ROLLBACK=COMPLETED_EVIDENCE_PRESERVING'
      exit "${code}"
    fi
  fi

  systemctl stop pvnaive-api.service >/dev/null 2>&1 || true
  systemctl stop pvnaive-runtime-agent.service >/dev/null 2>&1 || true
  systemctl stop pvnaive-telemetry-agent.service >/dev/null 2>&1 || true

  cp -a -- "${backup_dir}/pvnaive.before" /opt/pvnaive/bin/pvnaive
  cp -a -- "${backup_dir}/pvnaive-password.before" /opt/pvnaive/bin/pvnaive-password
  cp -a -- "${backup_dir}/pvnaive-runtime-agent.before" /opt/pvnaive/bin/pvnaive-runtime-agent
  cp -a -- "${backup_dir}/pvnaive-telemetry-agent.before" /opt/pvnaive/bin/pvnaive-telemetry-agent
  cp -a -- "${backup_dir}/pvnaive-api.service.before" /etc/systemd/system/pvnaive-api.service
  cp -a -- "${backup_dir}/pvnaive-runtime-agent.service.before" /etc/systemd/system/pvnaive-runtime-agent.service
  cp -a -- "${backup_dir}/pvnaive-telemetry-agent.service.before" /etc/systemd/system/pvnaive-telemetry-agent.service
  web_current_before="$(cat "${backup_dir}/web-current.before")"
  db_current_before="$(cat "${backup_dir}/db-current.before")"
  [[ -d "${web_current_before}" ]] && ln -sfn -- "${web_current_before}" /opt/pvnaive/web/current
  [[ -d "${db_current_before}" ]] && ln -sfn -- "${db_current_before}" /opt/pvnaive/db/current
  cp -a -- "${backup_dir}/api.env.before" "${api_env}"
  cp -a -- "${backup_dir}/db.env.before" "${db_env}"
  if [[ -f "${backup_dir}/caddy-before" ]]; then
    cp -a -- "${backup_dir}/caddy-before" "${caddy_bin}"
    chmod 0755 "${caddy_bin}"
    systemctl restart "${caddy_unit}" >/dev/null 2>&1 || true
    echo 'ROLLBACK_CADDY_BINARY=RESTORED'
  else
    echo 'ROLLBACK_CADDY_BINARY=NOT_REQUIRED'
  fi

  current_schema="$(query_schema 2>/dev/null || true)"
  if [[ "${current_schema}" == "${S06_ACCOUNTING_TARGET_SCHEMA}" && -n "${db_backup_path}" ]]; then
    if PVNAIVE_DB_HOST=/var/run/postgresql \
      PVNAIVE_DB_PORT="${PVNAIVE_DB_PORT}" \
      PVNAIVE_DB_NAME=pvnaive \
      PVNAIVE_DB_USER=postgres \
      PVNAIVE_RUN_AS_OS_USER=postgres \
      PVNAIVE_MIGRATIONS_DIR="${bundle_root}/db/migrations" \
      PVNAIVE_ALLOW_DESTRUCTIVE_ROLLBACK=ROLLBACK_MIGRATION_CHAIN \
      PVNAIVE_ROLLBACK_CHAIN_TARGET_SCHEMA=16 \
      PVNAIVE_CONFIRMED_BACKUP="${db_backup_path}" \
        bash "${bundle_root}/scripts/db/rollback.sh" >/dev/null 2>&1; then
      [[ "$(query_schema 2>/dev/null || true)" == "${S06_ACCOUNTING_BASE_SCHEMA}" ]] && echo 'ROLLBACK_DATABASE=PASS' || echo 'ROLLBACK_DATABASE=FAIL'
    else
      echo 'ROLLBACK_DATABASE=FAIL'
    fi
  else
    echo 'ROLLBACK_DATABASE=NOT_REQUIRED'
  fi

  systemctl daemon-reload >/dev/null 2>&1 || true
  systemctl restart pvnaive-telemetry-agent.service >/dev/null 2>&1 || true
  systemctl restart pvnaive-runtime-agent.service >/dev/null 2>&1 || true
  systemctl restart pvnaive-api.service >/dev/null 2>&1 || true
  final_bin_sha="$(sha256sum "${caddy_bin}" 2>/dev/null | awk '{print $1}')"
  if [[ "${final_bin_sha}" == "${caddy_bin_sha_before}" ]]; then
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
[[ ! -e "${web_release_dir}" ]] || fail 'release destination already exists'

install -d -o root -g pvnaive -m 0750 /opt/pvnaive/bin /opt/pvnaive/web/releases
install -d -o root -g pvnaive -m 0750 "${web_release_dir}"
cp -a -- "${bundle_root}/web/." "${web_release_dir}/"
chown -R root:pvnaive "${web_release_dir}"
ln -sfn -- "${web_release_dir}" /opt/pvnaive/web/current

install -o root -g pvnaive -m 0750 "${bundle_root}/bin/pvnaive" /opt/pvnaive/bin/pvnaive
install -o root -g pvnaive -m 0750 "${bundle_root}/bin/pvnaive-password" /opt/pvnaive/bin/pvnaive-password
install -o root -g pvnaive -m 0750 "${bundle_root}/bin/pvnaive-runtime-agent" /opt/pvnaive/bin/pvnaive-runtime-agent
install -o root -g pvnaive -m 0750 "${bundle_root}/bin/pvnaive-telemetry-agent" /opt/pvnaive/bin/pvnaive-telemetry-agent
install -o root -g root -m 0644 "${bundle_root}/systemd/pvnaive-api.service" /etc/systemd/system/pvnaive-api.service
install -o root -g root -m 0644 "${bundle_root}/systemd/pvnaive-runtime-agent.service" /etc/systemd/system/pvnaive-runtime-agent.service
install -o root -g root -m 0644 "${bundle_root}/systemd/pvnaive-telemetry-agent.service" /etc/systemd/system/pvnaive-telemetry-agent.service

PVNAIVE_DB_HOST=/var/run/postgresql \
PVNAIVE_DB_PORT="${PVNAIVE_DB_PORT}" \
PVNAIVE_DB_NAME=pvnaive \
PVNAIVE_DB_USER=postgres \
PVNAIVE_RUN_AS_OS_USER=postgres \
PVNAIVE_MIGRATIONS_DIR="${bundle_root}/db/migrations" \
  bash "${bundle_root}/scripts/db/migrate.sh" >/dev/null
schema_after="$(query_schema)"
[[ "${schema_after}" == "${S06_ACCOUNTING_TARGET_SCHEMA}" ]] || fail "schema did not reach version 17: ${schema_after}"
PVNAIVE_DB_ENV_FILE="${db_env}" bash "${bundle_root}/scripts/db/set-expected-schema-version.sh" 17 >/dev/null

PVNAIVE_DB_RELEASE_SOURCE_ROOT="${bundle_root}" \
PVNAIVE_DB_RELEASE_ROOT=/opt/pvnaive/db/releases \
PVNAIVE_DB_CURRENT_LINK=/opt/pvnaive/db/current \
PVNAIVE_DB_RELEASE_SCHEMA_VERSION=17 \
PVNAIVE_DB_RELEASE_MIGRATION_FILE=0017_operator_session_peers.up.sql \
PVNAIVE_DB_RELEASE_OWNER_USER=root \
PVNAIVE_DB_RELEASE_OWNER_GROUP=pvnaive \
  bash "${bundle_root}/scripts/db/promote-release.sh" >/dev/null

systemctl daemon-reload
install -o root -g root -m 0755 "${bundle_root}/caddy/caddy-pvnaive-accounting" "${caddy_bin}"
echo "ACCOUNTING_CADDY_INSTALLED_SHA256=$(sha256sum "${caddy_bin}" | awk '{print $1}')"
systemctl restart pvnaive-telemetry-agent.service
for _ in $(seq 1 20); do
  [[ -S "${telemetry_socket}" ]] && curl --fail --silent --show-error --unix-socket "${telemetry_socket}" http://unix/v1/accounting/health 2>/dev/null | grep -q '"status":"ok"' && break
  sleep 1
done
curl --fail --silent --show-error --unix-socket "${telemetry_socket}" http://unix/v1/accounting/health | grep -q '"status":"ok"' || fail 'telemetry agent health failed before Caddy activation'
caddy_activated=1
systemctl restart caddy-naive.service
systemctl restart pvnaive-runtime-agent.service
systemctl restart pvnaive-api.service

for _ in $(seq 1 30); do
  if [[ -S "${runtime_socket}" ]] && curl --fail --silent --show-error --unix-socket "${runtime_socket}" http://unix/v1/health 2>/dev/null | grep -q '"status":"ok"' && curl --fail --silent --show-error http://127.0.0.1:8080/api/v1/health/ready 2>/dev/null | grep -q '"ready":true'; then
    break
  fi
  sleep 1
done
curl --fail --silent --show-error --unix-socket "${runtime_socket}" http://unix/v1/health | grep -q '"status":"ok"' || fail 'runtime agent health failed after upgrade'
curl --fail --silent --show-error --unix-socket "${telemetry_socket}" http://unix/v1/accounting/health | grep -q '"status":"ok"' || fail 'telemetry agent health failed after upgrade'
curl --fail --silent --show-error http://127.0.0.1:8080/api/v1/health/ready | grep -q '"ready":true' || fail 'API readiness failed after upgrade'
[[ "$(query_schema)" == "${S06_ACCOUNTING_TARGET_SCHEMA}" ]] || fail 'schema verification failed after service start'

grep -Fxq 'PVNAIVE_EXPECTED_SCHEMA_VERSION=17' "${db_env}" || fail 'db.env schema expectation was not promoted to 17'
final_bin_sha="$(sha256sum "${caddy_bin}" | awk '{print $1}')"
[[ "${final_bin_sha}" == "${accounting_caddy_sha}" ]] || fail "accounting Caddy binary SHA mismatch after install: ${final_bin_sha}"
"${caddy_bin}" list-modules | grep -Fq 'pvnaive_accounting' >/dev/null || fail 'accounting module missing after install'
final_pid="$(systemctl show "${caddy_unit}" --property=MainPID --value)"
final_restarts="$(systemctl show "${caddy_unit}" --property=NRestarts --value)"
[[ "${final_pid}" =~ ^[1-9][0-9]*$ ]] || fail 'invalid Caddy MainPID after restart'
[[ "${final_restarts}" =~ ^[0-9]+$ ]] || fail 'invalid Caddy NRestarts after restart'

trap - ERR HUP INT TERM
echo 'S06_ACCOUNTING_RESULT=PASSED'
echo "SOURCE_COMMIT=${source_commit}"
echo "SCHEMA_VERSION=${S06_ACCOUNTING_TARGET_SCHEMA}"
echo "DB_BACKUP_PATH=${db_backup_path}"
echo "ACCOUNTING_CADDY_SHA256=${final_bin_sha}"
echo "CADDY_MainPID_AFTER=${final_pid}"
echo "CADDY_NRestarts_AFTER=${final_restarts}"
echo 'ONE_CONTROLLED_CADDY_RESTART=CONFIRMED'
echo 'POSTFLIGHT_CADDY_RESTART_COUNT=1'
echo 'USAGE_ACCOUNTING=NOT_PROVEN'
echo 'FIRST_SUCCESS_CONNECT_PRODUCER=NOT_PROVEN'
