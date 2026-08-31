#!/usr/bin/env bash
set -Eeuo pipefail
umask 077

repo_root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd -P)"
api_binary="${PVNAIVE_REHEARSAL_API_BINARY:-${repo_root}/dist/rehearsal/pvnaive}"
password_binary="${PVNAIVE_REHEARSAL_PASSWORD_BINARY:-${repo_root}/dist/rehearsal/pvnaive-password}"
agent_binary="${PVNAIVE_REHEARSAL_RUNTIME_AGENT_BINARY:-${repo_root}/dist/rehearsal/pvnaive-runtime-agent-rehearsal}"
: "${PVNAIVE_DB_HOST:=127.0.0.1}"
: "${PVNAIVE_DB_PORT:=5432}"
: "${PVNAIVE_DB_USER:=postgres}"
export PVNAIVE_DB_HOST PVNAIVE_DB_PORT PVNAIVE_DB_USER

for required in psql createdb dropdb curl jq sha256sum; do
  command -v "${required}" >/dev/null 2>&1 || { echo "ERROR: missing ${required}" >&2; exit 1; }
done
for binary in "${api_binary}" "${password_binary}" "${agent_binary}"; do
  [[ -x "${binary}" ]] || { echo "ERROR: rehearsal binary missing: ${binary}" >&2; exit 1; }
done

test_db="pvnaive"
api_port="18081"
api_pid=""
agent_pid=""
tmpdir=""
password='S04R-Rehearsal-Owner-Password'
owner_email='owner-s04r@example.invalid'

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
  if [[ -n "${agent_pid}" ]]; then
    kill "${agent_pid}" >/dev/null 2>&1 || true
    wait "${agent_pid}" >/dev/null 2>&1 || true
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
CREATE ROLE pvnaive_app LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOREPLICATION NOBYPASSRLS PASSWORD 'pvnaive-s04r-ci-only';
ALTER ROLE pvnaive_app SET row_security = on;
SQL
createdb --host "${PVNAIVE_DB_HOST}" --port "${PVNAIVE_DB_PORT}" --username "${PVNAIVE_DB_USER}" \
  --owner pvnaive_owner --encoding UTF8 --template template0 "${test_db}"

PVNAIVE_DB_NAME="${test_db}" PVNAIVE_MIGRATIONS_DIR="${repo_root}/db/migrations" \
  "${repo_root}/scripts/db/migrate.sh" >/dev/null
schema_version="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command 'SELECT COALESCE(MAX(version),0) FROM pvnaive.schema_migrations')"
[[ "${schema_version}" =~ ^[0-9]+$ && "${schema_version}" -ge 8 ]] || { echo "ERROR: S04R compatibility rehearsal schema version=${schema_version}, minimum=8" >&2; exit 1; }

phc="$(printf '%s\n' "${password}" | "${password_binary}")"
[[ "${phc}" == '$argon2id$'* ]] || { echo 'ERROR: password helper returned invalid PHC' >&2; exit 1; }
psql_admin --dbname "${test_db}" --set=owner_email="${owner_email}" --set=owner_hash="${phc}" <<'SQL' >/dev/null
SET ROLE pvnaive_owner;
INSERT INTO pvnaive.actors (tenant_id, actor_role, email, display_name, password_hash, mfa_required, status, password_changed_at)
VALUES (NULL, 'owner', :'owner_email', 'S04R Rehearsal Owner', :'owner_hash', false, 'active', clock_timestamp());
SQL
unset phc

dd if=/dev/urandom of="${tmpdir}/auth.key" bs=32 count=1 status=none
dd if=/dev/urandom of="${tmpdir}/runtime.key" bs=32 count=1 status=none
chmod 0600 "${tmpdir}/auth.key" "${tmpdir}/runtime.key"
socket="${tmpdir}/runtime-agent.sock"

PVNAIVE_RUNTIME_AGENT_SOCKET="${socket}" "${agent_binary}" >"${tmpdir}/agent.log" 2>&1 &
agent_pid=$!
for _ in $(seq 1 30); do
  if [[ -S "${socket}" ]] && curl --fail --silent --unix-socket "${socket}" http://unix/v1/health | grep -q '"status":"ok"'; then
    break
  fi
  sleep 1
done
kill -0 "${agent_pid}" >/dev/null 2>&1 || { cat "${tmpdir}/agent.log" >&2; echo 'ERROR: rehearsal runtime agent exited' >&2; exit 1; }
[[ -S "${socket}" ]] || { echo 'ERROR: rehearsal runtime socket missing' >&2; exit 1; }

initial_sha="$(curl --fail --silent --unix-socket "${socket}" http://unix/v1/inspect | jq -r '.caddy_sha256')"
[[ "${#initial_sha}" == 64 ]] || { echo 'ERROR: invalid initial runtime SHA' >&2; exit 1; }

PVNAIVE_DB_HOST="${PVNAIVE_DB_HOST}" \
PVNAIVE_DB_PORT="${PVNAIVE_DB_PORT}" \
PVNAIVE_DB_NAME="${test_db}" \
PVNAIVE_DB_USER=pvnaive_app \
PVNAIVE_DB_CONNECT_TIMEOUT=5 \
PVNAIVE_EXPECTED_SCHEMA_VERSION="${schema_version}" \
PGPASSWORD=pvnaive-s04r-ci-only \
PGSSLMODE=disable \
PVNAIVE_AUTH_KEY_FILE="${tmpdir}/auth.key" \
PVNAIVE_RUNTIME_KEY_FILE="${tmpdir}/runtime.key" \
PVNAIVE_RUNTIME_KEY_ID=runtime-rehearsal-v1 \
PVNAIVE_RUNTIME_AGENT_SOCKET="${socket}" \
PVNAIVE_NAIVE_PUBLIC_HOST="naive-rehearsal.example.invalid:443" \
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

echo 'S04R_FULL_REHEARSAL=PASSED'
