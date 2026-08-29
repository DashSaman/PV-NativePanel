#!/usr/bin/env bash
set -Eeuo pipefail
umask 077

repo_root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd -P)"
: "${PVNAIVE_DB_HOST:=127.0.0.1}"
: "${PVNAIVE_DB_PORT:=5432}"
: "${PVNAIVE_DB_USER:=postgres}"
export PVNAIVE_DB_HOST PVNAIVE_DB_PORT PVNAIVE_DB_USER

suffix="${GITHUB_RUN_ID:-local}_${GITHUB_RUN_ATTEMPT:-1}_${BASHPID}"
suffix="${suffix//[^a-zA-Z0-9_]/_}"
test_db="pvnaive_migration_test_customer_lifecycle_${suffix,,}"
app_password='pvnaive-customer-ci-only'

psql_admin() {
  psql --no-psqlrc --set ON_ERROR_STOP=1 --host "${PVNAIVE_DB_HOST}" --port "${PVNAIVE_DB_PORT}" --username "${PVNAIVE_DB_USER}" "$@"
}

cleanup() {
  psql_admin --dbname postgres --command "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname='${test_db}' AND pid <> pg_backend_pid()" >/dev/null 2>&1 || true
  dropdb --if-exists --host "${PVNAIVE_DB_HOST}" --port "${PVNAIVE_DB_PORT}" --username "${PVNAIVE_DB_USER}" "${test_db}" >/dev/null 2>&1 || true
  psql_admin --dbname postgres --command 'DROP ROLE IF EXISTS pvnaive_app; DROP ROLE IF EXISTS pvnaive_owner;' >/dev/null 2>&1 || true
}
trap cleanup EXIT HUP INT TERM
cleanup

for file in \
  db/migrations/0004_customer_lifecycle_foundation.up.sql \
  db/migrations/0004_customer_lifecycle_foundation.down.sql; do
  [[ -f "${repo_root}/${file}" ]] || { echo "ERROR: missing ${file}" >&2; exit 1; }
done

grep -Fqx -- '-- pvnaive:migration-version 0004' "${repo_root}/db/migrations/0004_customer_lifecycle_foundation.up.sql"
grep -Fqx -- '-- pvnaive:destructive true' "${repo_root}/db/migrations/0004_customer_lifecycle_foundation.down.sql"

psql_admin --dbname postgres <<SQL >/dev/null
CREATE ROLE pvnaive_owner NOLOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOREPLICATION NOBYPASSRLS;
CREATE ROLE pvnaive_app LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOREPLICATION NOBYPASSRLS PASSWORD '${app_password}';
ALTER ROLE pvnaive_app SET row_security = on;
SQL
createdb --host "${PVNAIVE_DB_HOST}" --port "${PVNAIVE_DB_PORT}" --username "${PVNAIVE_DB_USER}" --owner pvnaive_owner --encoding UTF8 --template template0 "${test_db}"
export PVNAIVE_DB_NAME="${test_db}"
"${repo_root}/scripts/db/migrate.sh" >/dev/null

version="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command 'SELECT COALESCE(MAX(version),0) FROM pvnaive.schema_migrations')"
[[ "${version}" == "9" ]] || { echo "ERROR: schema version=${version}, want=9" >&2; exit 1; }

contract="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command "
SELECT
  (to_regclass('pvnaive.service_terms') IS NOT NULL) || '|' ||
  (to_regclass('pvnaive.user_runtime_credentials') IS NOT NULL) || '|' ||
  ((SELECT COUNT(*) FROM pvnaive.tenants WHERE tenant_type='system' AND slug='direct' AND status='active')=1) || '|' ||
  (SELECT relforcerowsecurity FROM pg_class WHERE oid='pvnaive.service_terms'::regclass) || '|' ||
  (SELECT relforcerowsecurity FROM pg_class WHERE oid='pvnaive.user_runtime_credentials'::regclass);")"
[[ "${contract}" == "true|true|true|true|true" || "${contract}" == "t|t|t|t|t" ]] || { echo "ERROR: lifecycle schema contract failed: ${contract}" >&2; exit 1; }

psql_admin --dbname "${test_db}" >/dev/null <<'SQL'
INSERT INTO pvnaive.actors (id, tenant_id, actor_role, email, display_name, password_hash, status)
VALUES (
  'a4100000-0000-0000-0000-000000000001', NULL, 'owner',
  'customer-owner@example.invalid', 'Customer Owner',
  '$argon2id$v=19$m=19456,t=2,p=1$AAAAAAAAAAAAAAAAAAAAAA$AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA',
  'active'
);
INSERT INTO pvnaive.auth_sessions (
  id, tenant_id, actor_id, token_hash, refresh_family_id, user_agent_hash,
  expires_at, absolute_expires_at, csrf_token_hash
) VALUES (
  'b4100000-0000-0000-0000-000000000001', NULL,
  'a4100000-0000-0000-0000-000000000001', decode(repeat('14',32),'hex'),
  'c4100000-0000-0000-0000-000000000001', decode(repeat('24',32),'hex'),
  clock_timestamp()+interval '1 hour', clock_timestamp()+interval '12 hours', decode(repeat('34',32),'hex')
);
SQL

PGPASSWORD="${app_password}" psql --no-psqlrc --set ON_ERROR_STOP=1 \
  --host "${PVNAIVE_DB_HOST}" --port "${PVNAIVE_DB_PORT}" --username pvnaive_app --dbname "${test_db}" >/dev/null <<'SQL'
BEGIN;
SELECT pvnaive.set_request_context(decode(repeat('14',32),'hex'));

INSERT INTO pvnaive.users (id,tenant_id,username,display_name,status,created_by_actor_id)
SELECT 'd4100000-0000-0000-0000-000000000001',id,'test1','Test One','active','a4100000-0000-0000-0000-000000000001'
FROM pvnaive.tenants WHERE tenant_type='system' AND slug='direct';

INSERT INTO pvnaive.naive_runtime_credentials (
  id,username,secret_hash,secret_ciphertext,secret_nonce,encryption_key_id,status,origin,created_by_actor_id,updated_by_actor_id
) VALUES
('f4100000-0000-0000-0000-000000000001','test1-runtime-a',decode(repeat('44',32),'hex'),decode(repeat('54',16),'hex'),decode(repeat('64',12),'hex'),'runtime-v1','active','panel','a4100000-0000-0000-0000-000000000001','a4100000-0000-0000-0000-000000000001'),
('f4100000-0000-0000-0000-000000000002','test1-runtime-b',decode(repeat('45',32),'hex'),decode(repeat('55',16),'hex'),decode(repeat('65',12),'hex'),'runtime-v1','active','panel','a4100000-0000-0000-0000-000000000001','a4100000-0000-0000-0000-000000000001');

INSERT INTO pvnaive.service_terms (
  id,tenant_id,user_id,quota_bytes,duration_seconds,start_policy,purchased_at,state
)
SELECT '14100000-0000-0000-0000-000000000001',tenant_id,id,50000000000,2592000,'on_first_successful_connection',clock_timestamp(),'pending'
FROM pvnaive.users WHERE username='test1';

INSERT INTO pvnaive.user_runtime_credentials (tenant_id,user_id,service_term_id,runtime_credential_id,role)
SELECT tenant_id,id,'14100000-0000-0000-0000-000000000001','f4100000-0000-0000-0000-000000000001','primary'
FROM pvnaive.users WHERE username='test1';
COMMIT;
SQL

snapshot="$(PGPASSWORD="${app_password}" psql --no-psqlrc --set ON_ERROR_STOP=1 \
  --host "${PVNAIVE_DB_HOST}" --port "${PVNAIVE_DB_PORT}" --username pvnaive_app --dbname "${test_db}" --tuples-only --no-align <<'SQL'
BEGIN;
SELECT pvnaive.set_request_context(decode(repeat('14',32),'hex'));
SELECT quota_bytes || '|' || duration_seconds || '|' || start_policy || '|' || state || '|' || (starts_at IS NULL) || '|' || (expires_at IS NULL)
FROM pvnaive.service_terms WHERE id='14100000-0000-0000-0000-000000000001';
ROLLBACK;
SQL
)"
snapshot="$(printf '%s\n' "${snapshot}" | grep -E '^50000000000\|' | tail -n1)"
[[ "${snapshot}" == "50000000000|2592000|on_first_successful_connection|pending|true|true" || "${snapshot}" == "50000000000|2592000|on_first_successful_connection|pending|t|t" ]] || { echo "ERROR: term snapshot contract failed: ${snapshot}" >&2; exit 1; }

set +e
PGPASSWORD="${app_password}" psql --no-psqlrc --set ON_ERROR_STOP=1 \
  --host "${PVNAIVE_DB_HOST}" --port "${PVNAIVE_DB_PORT}" --username pvnaive_app --dbname "${test_db}" >/dev/null 2>&1 <<'SQL'
BEGIN;
SELECT pvnaive.set_request_context(decode(repeat('14',32),'hex'));
INSERT INTO pvnaive.user_runtime_credentials (tenant_id,user_id,service_term_id,runtime_credential_id,role)
SELECT tenant_id,id,'14100000-0000-0000-0000-000000000001','f4100000-0000-0000-0000-000000000002','primary'
FROM pvnaive.users WHERE username='test1';
COMMIT;
SQL
second_primary_rc=$?
set -e
[[ "${second_primary_rc}" -ne 0 ]] || { echo 'ERROR: second active primary binding was accepted' >&2; exit 1; }

# Exercise each destructive rollback in order: v9 -> v8 -> v7 -> v6 -> v5 -> v4 -> v3.
for want in 8 7 6 5 4 3; do
  PVNAIVE_DISPOSABLE_DB=1 PVNAIVE_ALLOW_DESTRUCTIVE_ROLLBACK=ROLLBACK_ONE_MIGRATION "${repo_root}/scripts/db/rollback.sh" >/dev/null
  got="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command 'SELECT COALESCE(MAX(version),0) FROM pvnaive.schema_migrations')"
  [[ "${got}" == "${want}" ]] || { echo "ERROR: rollback schema=${got}, want=${want}" >&2; exit 1; }
done
remaining="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command "SELECT (to_regclass('pvnaive.naive_runtime_credentials') IS NOT NULL) || '|' || (to_regclass('pvnaive.service_terms') IS NULL)")"
[[ "${remaining}" == "true|true" || "${remaining}" == "t|t" ]] || { echo "ERROR: v4 rollback did not preserve v3 cleanly: ${remaining}" >&2; exit 1; }

echo 'PVNAIVE_CUSTOMER_LIFECYCLE_MIGRATION_TEST=PASSED'
