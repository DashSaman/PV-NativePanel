#!/usr/bin/env bash
set -Eeuo pipefail
umask 077

repo_root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd -P)"
: "${PVNAIVE_DB_HOST:=127.0.0.1}"
: "${PVNAIVE_DB_PORT:=5432}"
: "${PVNAIVE_DB_USER:=postgres}"
export PVNAIVE_DB_HOST PVNAIVE_DB_PORT PVNAIVE_DB_USER

test_suffix="${GITHUB_RUN_ID:-local}_${GITHUB_RUN_ATTEMPT:-1}_${BASHPID}"
test_suffix="${test_suffix//[^a-zA-Z0-9_]/_}"
test_db="pvnaive_auth_actor_mutation_${test_suffix,,}"
app_password="pvnaive-auth-actor-ci-only"
actor_a="91000000-0000-0000-0000-000000000001"
actor_b="91000000-0000-0000-0000-000000000002"

psql_admin() {
  psql --no-psqlrc --set ON_ERROR_STOP=1 \
    --host "${PVNAIVE_DB_HOST}" --port "${PVNAIVE_DB_PORT}" \
    --username "${PVNAIVE_DB_USER}" "$@"
}

psql_app() {
  PGPASSWORD="${app_password}" psql --no-psqlrc --set ON_ERROR_STOP=1 \
    --host "${PVNAIVE_DB_HOST}" --port "${PVNAIVE_DB_PORT}" \
    --username pvnaive_app --dbname "${test_db}" "$@"
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

psql_admin --dbname postgres <<SQL >/dev/null
CREATE ROLE pvnaive_owner NOLOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOREPLICATION NOBYPASSRLS;
CREATE ROLE pvnaive_app LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOREPLICATION NOBYPASSRLS PASSWORD '${app_password}';
ALTER ROLE pvnaive_app SET row_security = on;
SQL
createdb --host "${PVNAIVE_DB_HOST}" --port "${PVNAIVE_DB_PORT}" \
  --username "${PVNAIVE_DB_USER}" --owner pvnaive_owner --encoding UTF8 --template template0 "${test_db}"

export PVNAIVE_DB_NAME="${test_db}"
"${repo_root}/scripts/db/migrate.sh" >/dev/null

schema_version="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command \
  'SELECT COALESCE(MAX(version),0) FROM pvnaive.schema_migrations')"
[[ "${schema_version}" == "22" ]] || { echo "ERROR: schema version=${schema_version}, want=22" >&2; exit 1; }

psql_admin --dbname "${test_db}" <<SQL >/dev/null
INSERT INTO pvnaive.actors (id, tenant_id, actor_role, email, display_name, password_hash, status)
VALUES
  ('${actor_a}', NULL, 'owner', 'actor-a@example.invalid', 'Actor A', '\$argon2id\$v=19\$m=19456,t=2,p=1\$AAAAAAAAAAAAAAAAAAAAAA\$AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA', 'active'),
  ('${actor_b}', NULL, 'owner', 'actor-b@example.invalid', 'Actor B', '\$argon2id\$v=19\$m=19456,t=2,p=1\$BBBBBBBBBBBBBBBBBBBBBB\$BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB', 'active');
SQL

psql_app --command "SELECT pvnaive.auth_create_session(
  '${actor_a}'::uuid,
  decode(repeat('11',32),'hex'),
  decode(repeat('22',32),'hex'),
  '91000000-0000-0000-0000-000000000011'::uuid,
  NULL,
  clock_timestamp() + interval '1 hour',
  clock_timestamp() + interval '4 hours')" >/dev/null

set +e
foreign_password_output="$(psql_app 2>&1 <<SQL
BEGIN;
SELECT * FROM pvnaive.set_request_context(decode(repeat('11',32),'hex'));
SELECT pvnaive.auth_update_actor_password(
  '${actor_b}'::uuid,
  '\$argon2id\$v=19\$m=19456,t=2,p=1\$CCCCCCCCCCCCCCCCCCCCCC\$CCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCC');
ROLLBACK;
SQL
)"
foreign_password_rc=$?
set -e
[[ "${foreign_password_rc}" -ne 0 ]] || {
  echo 'ERROR: actor A could mutate actor B password through SECURITY DEFINER function' >&2
  exit 1
}
grep -Fq 'authentication context required' <<<"${foreign_password_output}" || {
  echo "ERROR: cross-actor password mutation was rejected for the wrong reason: ${foreign_password_output}" >&2
  exit 1
}

set +e
foreign_profile_output="$(psql_app 2>&1 <<SQL
BEGIN;
SELECT * FROM pvnaive.set_request_context(decode(repeat('11',32),'hex'));
SELECT * FROM pvnaive.auth_update_actor_profile('${actor_b}'::uuid, 'stolen@example.invalid', 'Stolen');
ROLLBACK;
SQL
)"
foreign_profile_rc=$?
set -e
[[ "${foreign_profile_rc}" -ne 0 ]] || {
  echo 'ERROR: actor A could mutate actor B profile through SECURITY DEFINER function' >&2
  exit 1
}
grep -Fq 'authentication context required' <<<"${foreign_profile_output}" || {
  echo "ERROR: cross-actor profile mutation was rejected for the wrong reason: ${foreign_profile_output}" >&2
  exit 1
}

psql_app >/dev/null <<SQL
BEGIN;
SELECT * FROM pvnaive.set_request_context(decode(repeat('11',32),'hex'));
SELECT * FROM pvnaive.auth_update_actor_profile('${actor_a}'::uuid, 'actor-a2@example.invalid', 'Actor A2');
COMMIT;
SQL

self_state="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command \
  "SELECT email || '|' || display_name FROM pvnaive.actors WHERE id='${actor_a}'")"
[[ "${self_state}" == "actor-a2@example.invalid|Actor A2" ]] || {
  echo "ERROR: same-actor profile update failed: ${self_state}" >&2
  exit 1
}

echo 'PVNAIVE_AUTH_ACTOR_CREDENTIAL_MUTATIONS_MIGRATION_TEST=PASSED'
