#!/usr/bin/env bash
# Disposable end-to-end database rehearsal for the real S03 target.
# This intentionally exercises the exact production migration, health,
# encrypted backup, restore and rollback scripts before S03 creates pvnaive.
set -Eeuo pipefail
umask 077

stage_id="S03-POSTGRES18-REHEARSAL"
expected_host="testAmir5-3"
expected_caddy_sha256="101884de2dd11cb9d276df8e72cd068bed50e4ec6eb4ebb477184dda7a86e8b1"
script_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd -P)"
repo_root="$(cd -- "${script_dir}/../.." && pwd -P)"
stamp="$(date -u +%Y%m%dT%H%M%SZ)"
suffix="${stamp,,}_${BASHPID}"
source_db="pvnaive_migration_test_s03_${suffix}"
restore_db="pvnaive_restore_test_s03_${suffix}"
temp_root="/var/lib/pvnaive/.s03-rehearsal-${suffix}"
backup_root="/var/backups/pvnaive/rehearsal/${suffix}"
pgpass_file="${temp_root}/db.pgpass"
age_identity="${temp_root}/backup.agekey"
age_recipient="${temp_root}/backup.recipient"
cluster_version=""
cluster_name=""
cluster_port=""
roles_created=0
source_db_created=0
restore_db_created=0
keep_failure_artifacts=0
caddy_before=""

fail() {
  echo "S03_REHEARSAL_RESULT=FAILED" >&2
  echo "ERROR=$*" >&2
  return 1
}

postgres_psql() {
  runuser -u postgres -- psql \
    --no-psqlrc \
    --set ON_ERROR_STOP=1 \
    --host /var/run/postgresql \
    --port "${cluster_port}" \
    --username postgres \
    "$@"
}

cleanup() {
  local rc="$1"
  trap - EXIT HUP INT TERM
  set +e

  if [[ -n "${cluster_port}" ]]; then
    for database_name in "${restore_db}" "${source_db}"; do
      postgres_psql --dbname postgres --command \
        "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname='${database_name}' AND pid <> pg_backend_pid()" \
        >/dev/null 2>&1 || true
      runuser -u postgres -- dropdb --if-exists \
        --host /var/run/postgresql --port "${cluster_port}" "${database_name}" \
        >/dev/null 2>&1 || true
    done
    postgres_psql --dbname postgres --command \
      'DROP ROLE IF EXISTS pvnaive_app; DROP ROLE IF EXISTS pvnaive_owner;' \
      >/dev/null 2>&1 || true
  fi

  rm -rf -- "${temp_root}" 2>/dev/null || true
  if [[ "${rc}" == "0" ]]; then
    rm -rf -- "${backup_root}" 2>/dev/null || true
  else
    keep_failure_artifacts=1
    echo "REHEARSAL_FAILURE_BACKUP_ROOT=${backup_root}" >&2
  fi

  if [[ -n "${caddy_before}" && -f /etc/caddy/Caddyfile ]]; then
    caddy_after="$(sha256sum /etc/caddy/Caddyfile | awk '{print $1}')"
    echo "CADDY_BEFORE=${caddy_before}"
    echo "CADDY_AFTER=${caddy_after}"
    if [[ "${caddy_after}" != "${caddy_before}" ]]; then
      echo "CRITICAL=CADDYFILE_CHANGED_DURING_REHEARSAL" >&2
      rc=1
    fi
  fi

  if [[ "${rc}" != "0" ]]; then
    echo "S03_REHEARSAL_CLEANUP=COMPLETED"
  fi
  exit "${rc}"
}
trap 'cleanup "$?"' EXIT
trap 'exit 129' HUP
trap 'exit 130' INT
trap 'exit 143' TERM

[[ ${EUID} -eq 0 ]] || fail "run as root"
[[ "$(hostname)" == "${expected_host}" ]] || fail "unexpected host"
[[ -f /opt/pvnaive/FOUNDATION.json ]] || fail "S02 foundation marker is missing"
[[ ! -f /opt/pvnaive/S03_DATABASE.json ]] || fail "S03 success marker already exists"
[[ -f /var/lib/pvnaive/S03_POSTGRES_CLUSTER_OWNER ]] || fail "S03 PostgreSQL provenance marker is missing"
grep -Fqx 'stage=S03-DATABASE' /var/lib/pvnaive/S03_POSTGRES_CLUSTER_OWNER || fail "invalid cluster provenance stage"
grep -Fqx 'host=testAmir5-3' /var/lib/pvnaive/S03_POSTGRES_CLUSTER_OWNER || fail "invalid cluster provenance host"
grep -Fqx 'purpose=pvnaive-dedicated-postgresql' /var/lib/pvnaive/S03_POSTGRES_CLUSTER_OWNER || fail "invalid cluster provenance purpose"
[[ ! -f /var/run/reboot-required ]] || fail "reboot is required before rehearsal"

for command_name in \
  age age-keygen bash grep openssl pg_dump pg_isready pg_lsclusters pg_restore \
  psql runuser sha256sum systemctl systemd-analyze; do
  command -v "${command_name}" >/dev/null 2>&1 || fail "required command missing: ${command_name}"
done

caddy_before="$(sha256sum /etc/caddy/Caddyfile | awk '{print $1}')"
[[ "${caddy_before}" == "${expected_caddy_sha256}" ]] || fail "Caddyfile checksum mismatch"
systemctl is-active --quiet caddy-naive.service || fail "caddy-naive.service is not active"

mapfile -t clusters < <(pg_lsclusters --no-header | awk '{print $1 "|" $2 "|" $3 "|" $4}')
((${#clusters[@]} == 1)) || fail "expected exactly one PostgreSQL cluster, found ${#clusters[@]}"
IFS='|' read -r cluster_version cluster_name cluster_port cluster_status <<< "${clusters[0]}"
[[ "${cluster_version}" == "18" ]] || fail "unexpected PostgreSQL major: ${cluster_version}"
[[ "${cluster_name}" == "main" ]] || fail "unexpected PostgreSQL cluster: ${cluster_name}"
[[ "${cluster_port}" == "5432" ]] || fail "unexpected PostgreSQL port: ${cluster_port}"
[[ "${cluster_status}" == "online" ]] || fail "PostgreSQL cluster is not online"

preexisting="$(postgres_psql --dbname postgres --tuples-only --no-align --command \
  "SELECT (SELECT COUNT(*) FROM pg_database WHERE datname='pvnaive') || '|' || (SELECT COUNT(*) FROM pg_roles WHERE rolname IN ('pvnaive_owner','pvnaive_app'))")"
[[ "${preexisting}" == "0|0" ]] || fail "real pvnaive database/roles are not clean before rehearsal: ${preexisting}"

for path in \
  /etc/pvnaive/db.env /etc/pvnaive/db.pgpass \
  /etc/pvnaive/backup.agekey /etc/pvnaive/backup.recipient \
  /etc/systemd/system/pvnaive-db-health.service \
  /etc/systemd/system/pvnaive-db-health.timer \
  /opt/pvnaive/db/current; do
  [[ ! -e "${path}" && ! -L "${path}" ]] || fail "unexpected real S03 artifact before rehearsal: ${path}"
done

bash -n "${repo_root}"/scripts/db/*.sh "${repo_root}"/scripts/stages/*.sh "${repo_root}"/tests/db/*.sh "${repo_root}"/tests/stages/*.sh
(
  cd "${repo_root}/db/migrations"
  sha256sum --check --strict SHA256SUMS
)
bash "${repo_root}/tests/stages/S03_preflight_test.sh"
bash "${repo_root}/tests/stages/S03_apt_candidate_pipefail_test.sh"
bash "${repo_root}/tests/stages/S03_ubuntu2604_contract_test.sh"
echo "STATIC_GATES=PASSED"

systemd-analyze verify \
  "${repo_root}/ops/systemd/pvnaive-db-health.service" \
  "${repo_root}/ops/systemd/pvnaive-db-health.timer"
echo "SYSTEMD_UNIT_VERIFY=PASSED"

install -d -o root -g root -m 0700 "${temp_root}"
install -d -o root -g root -m 0700 "${backup_root}"

app_password="$(openssl rand -hex 32)"
roles_created=1
postgres_psql --dbname postgres <<SQL >/dev/null
CREATE ROLE pvnaive_owner NOLOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOREPLICATION NOBYPASSRLS;
CREATE ROLE pvnaive_app LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOREPLICATION NOBYPASSRLS CONNECTION LIMIT 40 PASSWORD '${app_password}';
ALTER ROLE pvnaive_app SET statement_timeout = '30s';
ALTER ROLE pvnaive_app SET lock_timeout = '5s';
ALTER ROLE pvnaive_app SET idle_in_transaction_session_timeout = '30s';
ALTER ROLE pvnaive_app SET row_security = on;
SQL

role_flags="$(postgres_psql --dbname postgres --tuples-only --no-align --command \
  "SELECT rolsuper || '|' || rolcreatedb || '|' || rolcreaterole || '|' || rolinherit || '|' || rolreplication || '|' || rolbypassrls FROM pg_roles WHERE rolname='pvnaive_app'")"
[[ "${role_flags}" == "f|f|f|f|f|f" ]] || fail "pvnaive_app role flags are unsafe: ${role_flags}"
echo "APP_ROLE_FLAGS=PASSED"

source_db_created=1
runuser -u postgres -- createdb \
  --host /var/run/postgresql --port "${cluster_port}" \
  --owner pvnaive_owner --encoding UTF8 --template template0 "${source_db}"
postgres_psql --dbname "${source_db}" <<SQL >/dev/null
REVOKE ALL ON DATABASE "${source_db}" FROM PUBLIC;
GRANT CONNECT ON DATABASE "${source_db}" TO pvnaive_app;
REVOKE TEMPORARY ON DATABASE "${source_db}" FROM pvnaive_app;
SQL

PVNAIVE_DB_HOST=/var/run/postgresql \
PVNAIVE_DB_PORT="${cluster_port}" \
PVNAIVE_DB_NAME="${source_db}" \
PVNAIVE_DB_USER=postgres \
PVNAIVE_RUN_AS_OS_USER=postgres \
PVNAIVE_MIGRATIONS_DIR="${repo_root}/db/migrations" \
  "${repo_root}/scripts/db/migrate.sh"

idempotent_output="$(
  PVNAIVE_DB_HOST=/var/run/postgresql \
  PVNAIVE_DB_PORT="${cluster_port}" \
  PVNAIVE_DB_NAME="${source_db}" \
  PVNAIVE_DB_USER=postgres \
  PVNAIVE_RUN_AS_OS_USER=postgres \
  PVNAIVE_MIGRATIONS_DIR="${repo_root}/db/migrations" \
    "${repo_root}/scripts/db/migrate.sh"
)"
grep -Fq 'MIGRATION 0001=ALREADY_APPLIED' <<< "${idempotent_output}" || fail "migration idempotency gate failed"
echo "MIGRATION_REAPPLY=PASSED"

install -o pvnaive -g pvnaive -m 0600 /dev/null "${pgpass_file}"
printf '127.0.0.1:%s:%s:pvnaive_app:%s\n' \
  "${cluster_port}" "${source_db}" "${app_password}" > "${pgpass_file}"
chown pvnaive:pvnaive "${pgpass_file}"
chmod 0600 "${pgpass_file}"
unset app_password

health_output="$(runuser -u pvnaive -- env -i \
  PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin \
  PVNAIVE_DB_HOST=127.0.0.1 \
  PVNAIVE_DB_PORT="${cluster_port}" \
  PVNAIVE_DB_NAME="${source_db}" \
  PVNAIVE_DB_USER=pvnaive_app \
  PVNAIVE_DB_CONNECT_TIMEOUT=5 \
  PVNAIVE_EXPECTED_SCHEMA_VERSION=1 \
  PVNAIVE_EXPECTED_DB_USER=pvnaive_app \
  PGPASSFILE="${pgpass_file}" \
  "${repo_root}/scripts/db/health.sh")"
printf '%s\n' "${health_output}"
grep -Fqx 'PVNAIVE_DB_HEALTH=OK' <<< "${health_output}" || fail "application health did not pass"
grep -Fqx 'PVNAIVE_DB_USER=pvnaive_app' <<< "${health_output}" || fail "application health user mismatch"
grep -Fqx 'PVNAIVE_DB_SERVER_ADDRESS=127.0.0.1' <<< "${health_output}" || fail "application health server endpoint mismatch"
grep -Fqx "PVNAIVE_DB_SERVER_PORT=${cluster_port}" <<< "${health_output}" || fail "application health server port mismatch"
grep -Fqx 'PVNAIVE_SECRET_DIRECT_SELECT=DENIED' <<< "${health_output}" || fail "application secret boundary did not pass"
echo "APPLICATION_HEALTH=PASSED"

age-keygen -o "${age_identity}" >/dev/null 2>&1
age-keygen -y "${age_identity}" > "${age_recipient}"
chmod 0600 "${age_identity}" "${age_recipient}"

backup_output="$(
  PVNAIVE_DB_HOST=/var/run/postgresql \
  PVNAIVE_DB_PORT="${cluster_port}" \
  PVNAIVE_DB_NAME="${source_db}" \
  PVNAIVE_DB_USER=postgres \
  PVNAIVE_RUN_AS_OS_USER=postgres \
  PVNAIVE_BACKUP_ROOT="${backup_root}" \
  PVNAIVE_BACKUP_RECIPIENT_FILE="${age_recipient}" \
  PVNAIVE_BACKUP_IDENTITY_FILE="${age_identity}" \
    "${repo_root}/scripts/db/backup.sh"
)"
printf '%s\n' "${backup_output}"
grep -Fqx 'PVNAIVE_BACKUP_RESULT=PASSED' <<< "${backup_output}" || fail "encrypted backup did not pass"
backup_file="$(awk -F= '/^PVNAIVE_BACKUP_PATH=/ {value=$2} END {print value}' <<< "${backup_output}")"
[[ -f "${backup_file}" ]] || fail "backup archive path was not created"
[[ ! -e "$(dirname -- "${backup_file}")/pvnaive.dump" ]] || fail "plaintext database dump exists"
(cd "$(dirname -- "${backup_file}")" && sha256sum --check --strict SHA256SUMS)
echo "ENCRYPTED_BACKUP=PASSED"

restore_db_created=1
restore_output="$(
  PVNAIVE_DB_HOST=/var/run/postgresql \
  PVNAIVE_DB_PORT="${cluster_port}" \
  PVNAIVE_DB_NAME="${source_db}" \
  PVNAIVE_DB_USER=postgres \
  PVNAIVE_RUN_AS_OS_USER=postgres \
  PVNAIVE_RESTORE_BACKUP="${backup_file}" \
  PVNAIVE_RESTORE_TARGET_DB="${restore_db}" \
  PVNAIVE_BACKUP_IDENTITY_FILE="${age_identity}" \
    "${repo_root}/scripts/db/restore.sh"
)"
printf '%s\n' "${restore_output}"
for required_line in \
  PVNAIVE_RESTORE_RESULT=PASSED \
  PVNAIVE_RESTORE_SCHEMA_VERSION=1 \
  PVNAIVE_RESTORE_OWNERSHIP=PASSED \
  PVNAIVE_RESTORE_ACLS=PASSED \
  PVNAIVE_RESTORE_SIGNING_KEY=PASSED; do
  grep -Fqx "${required_line}" <<< "${restore_output}" || fail "restore gate missing: ${required_line}"
done
echo "RESTORE_DRILL=PASSED"

postgres_psql --dbname postgres --command \
  "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname='${restore_db}' AND pid <> pg_backend_pid()" \
  >/dev/null
runuser -u postgres -- dropdb \
  --host /var/run/postgresql --port "${cluster_port}" "${restore_db}"
restore_db_created=0

PVNAIVE_DB_HOST=/var/run/postgresql \
PVNAIVE_DB_PORT="${cluster_port}" \
PVNAIVE_DB_NAME="${source_db}" \
PVNAIVE_DB_USER=postgres \
PVNAIVE_RUN_AS_OS_USER=postgres \
PVNAIVE_MIGRATIONS_DIR="${repo_root}/db/migrations" \
PVNAIVE_DISPOSABLE_DB=1 \
  "${repo_root}/scripts/db/rollback.sh"

schema_exists="$(postgres_psql --dbname "${source_db}" --tuples-only --no-align --command "SELECT to_regnamespace('pvnaive') IS NOT NULL")"
[[ "${schema_exists}" == "f" ]] || fail "disposable rollback left pvnaive schema behind"
echo "ROLLBACK_DRILL=PASSED"

caddy_after="$(sha256sum /etc/caddy/Caddyfile | awk '{print $1}')"
[[ "${caddy_after}" == "${caddy_before}" ]] || fail "Caddyfile changed during rehearsal"
systemctl is-active --quiet caddy-naive.service || fail "Caddy became inactive during rehearsal"
ss -lntp | grep -E ':(22|80|443|5432)([[:space:]]|$)' || true

echo "S03_REHEARSAL_RESULT=PASSED"
echo "POSTGRES_CLUSTER=${cluster_version}/${cluster_name}"
echo "POSTGRES_PORT=${cluster_port}"
echo "CADDY_CHANGED=false"
echo "SSH_CHANGED=false"
echo "FIREWALL_CHANGED=false"
