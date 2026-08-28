#!/usr/bin/env bash
set -Eeuo pipefail
umask 077

repo_root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd -P)"
api_binary="${PVNAIVE_REHEARSAL_API_BINARY:-${repo_root}/dist/rehearsal/pvnaive}"
password_binary="${PVNAIVE_REHEARSAL_PASSWORD_BINARY:-${repo_root}/dist/rehearsal/pvnaive-password}"
: "${PVNAIVE_DB_HOST:=127.0.0.1}"
: "${PVNAIVE_DB_PORT:=5432}"
: "${PVNAIVE_DB_USER:=postgres}"
export PVNAIVE_DB_HOST PVNAIVE_DB_PORT PVNAIVE_DB_USER

for required in psql createdb dropdb curl sha256sum; do
  command -v "${required}" >/dev/null 2>&1 || { echo "ERROR: missing ${required}" >&2; exit 1; }
done
[[ -x "${api_binary}" ]] || { echo "ERROR: rehearsal API binary is missing" >&2; exit 1; }
[[ -x "${password_binary}" ]] || { echo "ERROR: rehearsal password helper is missing" >&2; exit 1; }

# This job owns an isolated PostgreSQL container, so use the exact production
# database name. That keeps the rehearsal contract identical to systemd and
# lets the API fail closed on unexpected database names.
test_db="pvnaive"
api_port="18080"
api_pid=""
tmpdir=""
password='S04-Rehearsal-Password-Only'
owner_email='owner-s04@example.invalid'

psql_admin() {
  psql --no-psqlrc --set ON_ERROR_STOP=1 \
    --host "${PVNAIVE_DB_HOST}" --port "${PVNAIVE_DB_PORT}" \
    --username "${PVNAIVE_DB_USER}" "$@"
}

cleanup() {
  if [[ -n "${api_pid}" ]]; then
    kill "${api_pid}" >/dev/null 2>&1 || true
    wait "${api_pid}" >/dev/null 2>&1 || true
  fi
  psql_admin --dbname postgres --command \
    "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname='${test_db}' AND pid <> pg_backend_pid()" >/dev/null 2>&1 || true
  dropdb --if-exists --host "${PVNAIVE_DB_HOST}" --port "${PVNAIVE_DB_PORT}" --username "${PVNAIVE_DB_USER}" "${test_db}" >/dev/null 2>&1 || true
  psql_admin --dbname postgres --command 'DROP ROLE IF EXISTS pvnaive_app; DROP ROLE IF EXISTS pvnaive_owner;' >/dev/null 2>&1 || true
  if [[ -n "${tmpdir}" ]]; then
    rm -rf -- "${tmpdir}" || true
  fi
}
cleanup
tmpdir="$(mktemp -d)"
trap cleanup EXIT HUP INT TERM

psql_admin --dbname postgres <<'SQL' >/dev/null
CREATE ROLE pvnaive_owner NOLOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOREPLICATION NOBYPASSRLS;
CREATE ROLE pvnaive_app LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOREPLICATION NOBYPASSRLS PASSWORD 'pvnaive-s04-ci-only';
ALTER ROLE pvnaive_app SET row_security = on;
SQL
createdb --host "${PVNAIVE_DB_HOST}" --port "${PVNAIVE_DB_PORT}" --username "${PVNAIVE_DB_USER}" \
  --owner pvnaive_owner --encoding UTF8 --template template0 "${test_db}"

PVNAIVE_DB_NAME="${test_db}" "${repo_root}/scripts/db/migrate.sh" >/dev/null
schema_version="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command 'SELECT COALESCE(MAX(version),0) FROM pvnaive.schema_migrations')"
[[ "${schema_version}" == "2" ]] || { echo "ERROR: rehearsal schema version=${schema_version}" >&2; exit 1; }

phc="$(printf '%s\n' "${password}" | "${password_binary}")"
[[ "${phc}" == '$argon2id$'* ]] || { echo 'ERROR: rehearsal password helper returned invalid PHC' >&2; exit 1; }
psql_admin --dbname "${test_db}" --set=owner_email="${owner_email}" --set=owner_hash="${phc}" <<'SQL' >/dev/null
SET ROLE pvnaive_owner;
INSERT INTO pvnaive.actors (tenant_id, actor_role, email, display_name, password_hash, mfa_required, status, password_changed_at)
VALUES (NULL, 'owner', :'owner_email', 'Rehearsal Owner', :'owner_hash', false, 'active', clock_timestamp());
SQL
unset phc

dd if=/dev/urandom of="${tmpdir}/auth.key" bs=32 count=1 status=none
chmod 0600 "${tmpdir}/auth.key"

PVNAIVE_DB_HOST="${PVNAIVE_DB_HOST}" \
PVNAIVE_DB_PORT="${PVNAIVE_DB_PORT}" \
PVNAIVE_DB_NAME="${test_db}" \
PVNAIVE_DB_USER=pvnaive_app \
PVNAIVE_DB_CONNECT_TIMEOUT=5 \
PGPASSWORD=pvnaive-s04-ci-only \
PGSSLMODE=disable \
PVNAIVE_AUTH_KEY_FILE="${tmpdir}/auth.key" \
PVNAIVE_LISTEN="127.0.0.1:${api_port}" \
  "${api_binary}" >"${tmpdir}/api.log" 2>&1 &
api_pid=$!

for _ in $(seq 1 30); do
  if curl --fail --silent "http://127.0.0.1:${api_port}/api/v1/health/ready" | grep -q '"ready":true'; then
    break
  fi
  sleep 1
done
kill -0 "${api_pid}" >/dev/null 2>&1 || { cat "${tmpdir}/api.log" >&2; echo 'ERROR: rehearsal API exited' >&2; exit 1; }
curl --fail --silent "http://127.0.0.1:${api_port}/api/v1/health/live" | grep -q '"status":"ok"'
curl --fail --silent "http://127.0.0.1:${api_port}/api/v1/health/ready" | grep -q '"ready":true'

login_body="$(printf '{"email":"%s","password":"%s","totp_code":""}' "${owner_email}" "${password}")"
curl --fail --silent --show-error \
  --cookie-jar "${tmpdir}/cookies.txt" \
  --header 'Content-Type: application/json' \
  --data "${login_body}" \
  "http://127.0.0.1:${api_port}/api/v1/auth/login" >"${tmpdir}/login.json"
grep -q '"status":"authenticated"' "${tmpdir}/login.json"
grep -q '"role":"owner"' "${tmpdir}/login.json"

curl --fail --silent --show-error --cookie "${tmpdir}/cookies.txt" \
  "http://127.0.0.1:${api_port}/api/v1/me" >"${tmpdir}/me.json"
grep -q '"role":"owner"' "${tmpdir}/me.json"

audit_count="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --set=owner_email="${owner_email}" --command "SELECT COUNT(*) FROM pvnaive.audit_events WHERE actor_id=(SELECT id FROM pvnaive.actors WHERE lower(email)=lower('owner-s04@example.invalid')) AND action='auth.login' AND outcome='success'")"
[[ "${audit_count}" == "1" ]] || { echo "ERROR: login success audit count=${audit_count}" >&2; exit 1; }

csrf="$(awk '$6 == "__Host-pvnaive_csrf" {print $7}' "${tmpdir}/cookies.txt" | tail -n1)"
[[ -n "${csrf}" ]] || { echo 'ERROR: CSRF cookie missing' >&2; exit 1; }
curl --fail --silent --show-error --cookie "${tmpdir}/cookies.txt" \
  --header "X-CSRF-Token: ${csrf}" --request POST \
  "http://127.0.0.1:${api_port}/api/v1/auth/logout" >"${tmpdir}/logout.json"
grep -q '"status":"logged_out"' "${tmpdir}/logout.json"

status="$(curl --silent --output /dev/null --write-out '%{http_code}' --cookie "${tmpdir}/cookies.txt" \
  "http://127.0.0.1:${api_port}/api/v1/me")"
[[ "${status}" == "401" ]] || { echo "ERROR: revoked session returned HTTP ${status}" >&2; exit 1; }

session_contract="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command \
  'SELECT COUNT(*) || '"'|'"' || COALESCE(MIN(octet_length(token_hash)),0) || '"'|'"' || COUNT(*) FILTER (WHERE revoked_at IS NOT NULL) FROM pvnaive.auth_sessions')"
[[ "${session_contract}" == "1|32|1" ]] || { echo "ERROR: session persistence contract failed: ${session_contract}" >&2; exit 1; }

echo 'S04_AUTH_REHEARSAL=PASSED'
