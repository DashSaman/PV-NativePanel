#!/usr/bin/env bash
set -Eeuo pipefail
umask 077

repo_root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd -P)"
up="${repo_root}/db/migrations/0002_auth_foundation.up.sql"
down="${repo_root}/db/migrations/0002_auth_foundation.down.sql"

[[ -f "${up}" ]] || { echo "ERROR: missing 0002_auth_foundation.up.sql" >&2; exit 1; }
[[ -f "${down}" ]] || { echo "ERROR: missing 0002_auth_foundation.down.sql" >&2; exit 1; }

grep -Fqx -- '-- pvnaive:migration-version 0002' "${up}"
grep -Fqx -- '-- pvnaive:transactional true' "${up}"
grep -Fqx -- '-- pvnaive:destructive false' "${up}"
grep -Fqx -- '-- pvnaive:migration-version 0002' "${down}"
grep -Fqx -- '-- pvnaive:transactional true' "${down}"
grep -Fqx -- '-- pvnaive:destructive true' "${down}"

: "${PVNAIVE_DB_HOST:=127.0.0.1}"
: "${PVNAIVE_DB_PORT:=5432}"
: "${PVNAIVE_DB_USER:=postgres}"
export PVNAIVE_DB_HOST PVNAIVE_DB_PORT PVNAIVE_DB_USER

test_suffix="${GITHUB_RUN_ID:-local}_${GITHUB_RUN_ATTEMPT:-1}_${BASHPID}"
test_suffix="${test_suffix//[^a-zA-Z0-9_]/_}"
test_db="pvnaive_migration_test_auth_${test_suffix,,}"

psql_admin() {
  psql --no-psqlrc --set ON_ERROR_STOP=1 \
    --host "${PVNAIVE_DB_HOST}" --port "${PVNAIVE_DB_PORT}" \
    --username "${PVNAIVE_DB_USER}" "$@"
}

cleanup() {
  psql_admin --dbname postgres --command \
    "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname='${test_db}' AND pid <> pg_backend_pid()" \
    >/dev/null 2>&1 || true
  dropdb --if-exists --host "${PVNAIVE_DB_HOST}" --port "${PVNAIVE_DB_PORT}" \
    --username "${PVNAIVE_DB_USER}" "${test_db}" >/dev/null 2>&1 || true
  psql_admin --dbname postgres --command \
    'DROP ROLE IF EXISTS pvnaive_app; DROP ROLE IF EXISTS pvnaive_owner;' >/dev/null 2>&1 || true
}
trap cleanup EXIT
cleanup

psql_admin --dbname postgres <<'SQL' >/dev/null
CREATE ROLE pvnaive_owner NOLOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOREPLICATION NOBYPASSRLS;
CREATE ROLE pvnaive_app LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOREPLICATION NOBYPASSRLS PASSWORD 'pvnaive-auth-ci-only';
ALTER ROLE pvnaive_app SET row_security = on;
SQL
createdb --host "${PVNAIVE_DB_HOST}" --port "${PVNAIVE_DB_PORT}" \
  --username "${PVNAIVE_DB_USER}" --owner pvnaive_owner --encoding UTF8 --template template0 "${test_db}"

export PVNAIVE_DB_NAME="${test_db}"
"${repo_root}/scripts/db/migrate.sh" >/dev/null

schema_version="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command \
  'SELECT COALESCE(MAX(version),0) FROM pvnaive.schema_migrations')"
[[ "${schema_version}" == "2" ]] || { echo "ERROR: schema version=${schema_version}, want=2" >&2; exit 1; }

contract="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command "
SELECT
  EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema='pvnaive' AND table_name='actors' AND column_name='failed_login_attempts') || '|' ||
  EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema='pvnaive' AND table_name='actors' AND column_name='locked_until') || '|' ||
  EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema='pvnaive' AND table_name='auth_sessions' AND column_name='csrf_token_hash') || '|' ||
  EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema='pvnaive' AND table_name='auth_sessions' AND column_name='absolute_expires_at') || '|' ||
  (to_regclass('pvnaive.actor_totp_factors') IS NOT NULL) || '|' ||
  (to_regclass('pvnaive.actor_mfa_recovery_codes') IS NOT NULL);")"
[[ "${contract}" == "true|true|true|true|true|true" || "${contract}" == "t|t|t|t|t|t" ]] || {
  echo "ERROR: auth schema contract failed: ${contract}" >&2; exit 1;
}

psql_admin --dbname "${test_db}" --command "
INSERT INTO pvnaive.actors (id, tenant_id, actor_role, email, display_name, password_hash, status)
VALUES ('99000000-0000-0000-0000-000000000099', NULL, 'owner', 'owner-auth-test@example.invalid', 'Owner', '\$argon2id\$v=19\$m=19456,t=2,p=1\$AAAAAAAAAAAAAAAAAAAAAA\$AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA', 'active');" >/dev/null

lookup="$(PGPASSWORD='pvnaive-auth-ci-only' psql --no-psqlrc --set ON_ERROR_STOP=1 \
  --host "${PVNAIVE_DB_HOST}" --port "${PVNAIVE_DB_PORT}" --username pvnaive_app --dbname "${test_db}" \
  --tuples-only --no-align --command \
  "SELECT actor_id || '|' || actor_role || '|' || status FROM pvnaive.auth_lookup_actor('OWNER-AUTH-TEST@example.invalid')")"
[[ "${lookup}" == "99000000-0000-0000-0000-000000000099|owner|active" ]] || {
  echo "ERROR: pre-auth lookup failed: ${lookup}" >&2; exit 1;
}

for _ in 1 2 3 4 5; do
  PGPASSWORD='pvnaive-auth-ci-only' psql --no-psqlrc --set ON_ERROR_STOP=1 \
    --host "${PVNAIVE_DB_HOST}" --port "${PVNAIVE_DB_PORT}" --username pvnaive_app --dbname "${test_db}" \
    --command "SELECT pvnaive.auth_record_login_failure('99000000-0000-0000-0000-000000000099')" >/dev/null
done
lock_state="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command \
  "SELECT failed_login_attempts || '|' || (locked_until > clock_timestamp()) FROM pvnaive.actors WHERE id='99000000-0000-0000-0000-000000000099'")"
[[ "${lock_state}" == "5|true" || "${lock_state}" == "5|t" ]] || {
  echo "ERROR: lockout contract failed: ${lock_state}" >&2; exit 1;
}

PGPASSWORD='pvnaive-auth-ci-only' psql --no-psqlrc --set ON_ERROR_STOP=1 \
  --host "${PVNAIVE_DB_HOST}" --port "${PVNAIVE_DB_PORT}" --username pvnaive_app --dbname "${test_db}" \
  --command "SELECT pvnaive.auth_record_login_success('99000000-0000-0000-0000-000000000099')" >/dev/null
reset_state="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command \
  "SELECT failed_login_attempts || '|' || (locked_until IS NULL) FROM pvnaive.actors WHERE id='99000000-0000-0000-0000-000000000099'")"
[[ "${reset_state}" == "0|true" || "${reset_state}" == "0|t" ]] || {
  echo "ERROR: login-success reset failed: ${reset_state}" >&2; exit 1;
}

for table in actor_totp_factors actor_mfa_recovery_codes; do
  direct="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command \
    "SELECT has_table_privilege('pvnaive_app','pvnaive.${table}','SELECT')")"
  [[ "${direct}" == "false" || "${direct}" == "f" ]] || {
    echo "ERROR: pvnaive_app has direct SELECT on ${table}" >&2; exit 1;
  }
done

PVNAIVE_DISPOSABLE_DB=1 PVNAIVE_ALLOW_DESTRUCTIVE_ROLLBACK=ROLLBACK_ONE_MIGRATION \
  "${repo_root}/scripts/db/rollback.sh" >/dev/null

remaining="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command \
  "SELECT (to_regnamespace('pvnaive') IS NOT NULL) || '|' || (SELECT COALESCE(MAX(version),0) FROM pvnaive.schema_migrations)")"
[[ "${remaining}" == "true|1" || "${remaining}" == "t|1" ]] || {
  echo "ERROR: v2 rollback did not preserve v1: ${remaining}" >&2; exit 1;
}

echo "PVNAIVE_AUTH_MIGRATION_TEST=PASSED"
