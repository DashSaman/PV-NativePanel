#!/usr/bin/env bash
set -Eeuo pipefail
umask 077

repo_root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd -P)"
test_suffix="${GITHUB_RUN_ID:-local}_${GITHUB_RUN_ATTEMPT:-1}_${BASHPID}"
test_suffix="${test_suffix//[^a-zA-Z0-9_]/_}"
test_db="pvnaive_health_test_${test_suffix,,}"
temp_root="$(mktemp -d)"
app_password="pvnaive-health-ci-only-${BASHPID}"

: "${PVNAIVE_DB_HOST:=127.0.0.1}"
: "${PVNAIVE_DB_PORT:=5432}"
: "${PVNAIVE_DB_USER:=postgres}"
admin_user="${PVNAIVE_DB_USER}"
export PVNAIVE_DB_HOST PVNAIVE_DB_PORT

psql_admin() {
  psql --no-psqlrc --set ON_ERROR_STOP=1 \
    --host "${PVNAIVE_DB_HOST}" \
    --port "${PVNAIVE_DB_PORT}" \
    --username "${admin_user}" \
    "$@"
}

cleanup() {
  psql_admin --dbname postgres --command \
    "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = '${test_db}' AND pid <> pg_backend_pid()" \
    >/dev/null 2>&1 || true
  dropdb --if-exists \
    --host "${PVNAIVE_DB_HOST}" \
    --port "${PVNAIVE_DB_PORT}" \
    --username "${admin_user}" \
    "${test_db}" >/dev/null 2>&1 || true
  psql_admin --dbname postgres --command \
    'DROP ROLE IF EXISTS pvnaive_app; DROP ROLE IF EXISTS pvnaive_owner;' \
    >/dev/null 2>&1 || true
  rm -rf -- "${temp_root}"
}
trap cleanup EXIT
cleanup
install -d -m 0700 "${temp_root}"

psql_admin --dbname postgres <<SQL >/dev/null
CREATE ROLE pvnaive_owner NOLOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOREPLICATION NOBYPASSRLS;
CREATE ROLE pvnaive_app LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOREPLICATION NOBYPASSRLS PASSWORD '${app_password}';
ALTER ROLE pvnaive_app SET row_security = on;
SQL

createdb \
  --host "${PVNAIVE_DB_HOST}" \
  --port "${PVNAIVE_DB_PORT}" \
  --username "${admin_user}" \
  --owner pvnaive_owner \
  --encoding UTF8 \
  --template template0 \
  "${test_db}"

psql_admin --dbname postgres --command \
  "REVOKE ALL ON DATABASE \"${test_db}\" FROM PUBLIC; GRANT CONNECT ON DATABASE \"${test_db}\" TO pvnaive_app; REVOKE TEMPORARY ON DATABASE \"${test_db}\" FROM pvnaive_app;" \
  >/dev/null

PVNAIVE_DB_NAME="${test_db}" \
PVNAIVE_DB_USER="${admin_user}" \
  "${repo_root}/scripts/db/migrate.sh" >/dev/null

pgpass="${temp_root}/app.pgpass"
printf '%s:%s:%s:%s:%s\n' \
  "${PVNAIVE_DB_HOST}" "${PVNAIVE_DB_PORT}" "${test_db}" "pvnaive_app" "${app_password}" > "${pgpass}"
chmod 0600 "${pgpass}"

raw_server_addr="$(
  PGPASSWORD="${app_password}" psql --no-psqlrc --set ON_ERROR_STOP=1 \
    --host "${PVNAIVE_DB_HOST}" --port "${PVNAIVE_DB_PORT}" \
    --username pvnaive_app --dbname "${test_db}" \
    --tuples-only --no-align --command 'SELECT inet_server_addr()::text'
)"
normalized_server_addr="$(
  PGPASSWORD="${app_password}" psql --no-psqlrc --set ON_ERROR_STOP=1 \
    --host "${PVNAIVE_DB_HOST}" --port "${PVNAIVE_DB_PORT}" \
    --username pvnaive_app --dbname "${test_db}" \
    --tuples-only --no-align --command 'SELECT host(inet_server_addr())'
)"
[[ "${normalized_server_addr}" == "127.0.0.1" ]] || {
  echo "ERROR: PostgreSQL server address did not normalize to IPv4 loopback: raw=${raw_server_addr} normalized=${normalized_server_addr}" >&2
  exit 1
}

echo "HEALTH_TEST_RAW_SERVER_ADDR=${raw_server_addr}"
echo "HEALTH_TEST_NORMALIZED_SERVER_ADDR=${normalized_server_addr}"

health_output="$(
  env -u PGPASSWORD \
    PVNAIVE_DB_HOST="${PVNAIVE_DB_HOST}" \
    PVNAIVE_DB_PORT="${PVNAIVE_DB_PORT}" \
    PVNAIVE_DB_NAME="${test_db}" \
    PVNAIVE_DB_USER=pvnaive_app \
    PVNAIVE_DB_CONNECT_TIMEOUT=5 \
    PVNAIVE_EXPECTED_SCHEMA_VERSION=3 \
    PVNAIVE_EXPECTED_DB_USER=pvnaive_app \
    PGPASSFILE="${pgpass}" \
    "${repo_root}/scripts/db/health.sh"
)"

grep -Fqx 'PVNAIVE_DB_HEALTH=OK' <<< "${health_output}"
grep -Fqx 'PVNAIVE_SCHEMA_VERSION=3' <<< "${health_output}"
grep -Fqx 'PVNAIVE_DB_USER=pvnaive_app' <<< "${health_output}"
grep -Fqx 'PVNAIVE_DB_SERVER_ADDRESS=127.0.0.1' <<< "${health_output}"
grep -Fqx "PVNAIVE_DB_SERVER_PORT=${PVNAIVE_DB_PORT}" <<< "${health_output}"
grep -Fqx 'PVNAIVE_DB_CLIENT_ADDRESS=127.0.0.1' <<< "${health_output}"
grep -Fqx 'PVNAIVE_SECRET_DIRECT_SELECT=DENIED' <<< "${health_output}"
grep -Fqx 'PVNAIVE_MFA_DIRECT_SELECT=DENIED' <<< "${health_output}"

if env -u PGPASSWORD \
  PVNAIVE_DB_HOST="${PVNAIVE_DB_HOST}" \
  PVNAIVE_DB_PORT="${PVNAIVE_DB_PORT}" \
  PVNAIVE_DB_NAME="${test_db}" \
  PVNAIVE_DB_USER=pvnaive_app \
  PVNAIVE_RUN_AS_OS_USER=postgres \
  PGPASSFILE="${pgpass}" \
  "${repo_root}/scripts/db/health.sh" >/dev/null 2>&1; then
  echo "ERROR: health accepted PVNAIVE_RUN_AS_OS_USER override" >&2
  exit 1
fi

if env -u PGPASSWORD \
  PVNAIVE_DB_HOST=localhost \
  PVNAIVE_DB_PORT="${PVNAIVE_DB_PORT}" \
  PVNAIVE_DB_NAME="${test_db}" \
  PVNAIVE_DB_USER=pvnaive_app \
  PGPASSFILE="${pgpass}" \
  "${repo_root}/scripts/db/health.sh" >/dev/null 2>&1; then
  echo "ERROR: health accepted a non-explicit loopback hostname" >&2
  exit 1
fi

echo "PVNAIVE_DB_HEALTH_TEST=PASSED"
