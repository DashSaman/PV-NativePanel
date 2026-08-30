#!/usr/bin/env bash
set -Eeuo pipefail
umask 077
repo_root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd -P)"
up="${repo_root}/db/migrations/0013_subscription_account_projection.up.sql"
down="${repo_root}/db/migrations/0013_subscription_account_projection.down.sql"
[[ -f "${up}" ]] || { echo 'ERROR: missing schema13 subscription account projection migration' >&2; exit 1; }
[[ -f "${down}" ]] || { echo 'ERROR: missing schema13 subscription account projection rollback' >&2; exit 1; }
grep -Fqx -- '-- pvnaive:migration-version 0013' "${up}"
grep -Fqx -- '-- pvnaive:transactional true' "${up}"
grep -Fqx -- '-- pvnaive:destructive false' "${up}"
grep -Fqx -- '-- pvnaive:migration-version 0013' "${down}"
grep -Fqx -- '-- pvnaive:transactional true' "${down}"
grep -Fqx -- '-- pvnaive:destructive true' "${down}"

: "${PVNAIVE_DB_HOST:=127.0.0.1}"
: "${PVNAIVE_DB_PORT:=5432}"
: "${PVNAIVE_DB_USER:=postgres}"
export PVNAIVE_DB_HOST PVNAIVE_DB_PORT PVNAIVE_DB_USER
suffix="${GITHUB_RUN_ID:-local}_${GITHUB_RUN_ATTEMPT:-1}_${BASHPID}"
suffix="${suffix//[^a-zA-Z0-9_]/_}"
test_db="pvnaive_migration_test_subscription_account_${suffix,,}"
psql_admin(){ psql --no-psqlrc --set ON_ERROR_STOP=1 --host "$PVNAIVE_DB_HOST" --port "$PVNAIVE_DB_PORT" --username "$PVNAIVE_DB_USER" "$@"; }
cleanup(){
  psql_admin --dbname postgres --command "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname='${test_db}' AND pid<>pg_backend_pid()" >/dev/null 2>&1 || true
  dropdb --if-exists --host "$PVNAIVE_DB_HOST" --port "$PVNAIVE_DB_PORT" --username "$PVNAIVE_DB_USER" "$test_db" >/dev/null 2>&1 || true
  psql_admin --dbname postgres --command 'DROP ROLE IF EXISTS pvnaive_app; DROP ROLE IF EXISTS pvnaive_owner;' >/dev/null 2>&1 || true
}
trap cleanup EXIT HUP INT TERM
cleanup
psql_admin --dbname postgres <<'SQL' >/dev/null
CREATE ROLE pvnaive_owner NOLOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOREPLICATION NOBYPASSRLS;
CREATE ROLE pvnaive_app LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOREPLICATION NOBYPASSRLS PASSWORD 'subscription-account-ci-only';
ALTER ROLE pvnaive_app SET row_security=on;
SQL
createdb --host "$PVNAIVE_DB_HOST" --port "$PVNAIVE_DB_PORT" --username "$PVNAIVE_DB_USER" --owner pvnaive_owner --encoding UTF8 --template template0 "$test_db"
export PVNAIVE_DB_NAME="$test_db"

v12="$(mktemp -d)"; trap 'rm -rf -- "$v12"; cleanup' EXIT HUP INT TERM
for version in $(seq -w 1 12); do cp "${repo_root}/db/migrations/00${version}_"* "$v12/"; done
( cd "$v12"; sha256sum *.sql > SHA256SUMS )
PVNAIVE_MIGRATIONS_DIR="$v12" "${repo_root}/scripts/db/migrate.sh" >/dev/null
[[ "$(psql_admin --dbname "$test_db" --tuples-only --no-align --command 'SELECT max(version) FROM pvnaive.schema_migrations')" == 12 ]]

# Create a fresh managed term at schema12 so its known-zero provenance must survive
# the public token projection added by schema13.
psql_admin --dbname "$test_db" <<'SQL' >/dev/null
INSERT INTO pvnaive.actors(id,tenant_id,actor_role,email,display_name,status)
VALUES('a6130000-0000-0000-0000-000000000001',NULL,'owner','task6-owner@example.invalid','Task6 Owner','active');
INSERT INTO pvnaive.users(id,tenant_id,username,display_name,status,created_by_actor_id)
SELECT 'b6130000-0000-0000-0000-000000000001',id,'task6-user','Task6 User','active','a6130000-0000-0000-0000-000000000001'
FROM pvnaive.tenants WHERE tenant_type='system' AND slug='direct';
INSERT INTO pvnaive.naive_runtime_credentials(id,username,secret_hash,secret_ciphertext,secret_nonce,encryption_key_id,status,origin,created_by_actor_id,updated_by_actor_id)
VALUES('c6130000-0000-0000-0000-000000000001','task6-user',decode(repeat('11',32),'hex'),decode(repeat('22',32),'hex'),decode(repeat('33',12),'hex'),'runtime-v1','active','panel','a6130000-0000-0000-0000-000000000001','a6130000-0000-0000-0000-000000000001');
INSERT INTO pvnaive.service_terms(id,tenant_id,user_id,quota_bytes,duration_seconds,start_policy,purchased_at,state,accounting_baseline_state,accounting_baseline_source,accounting_baseline_cutoff_at,accounting_baseline_upload_bytes,accounting_baseline_download_bytes)
SELECT 'd6130000-0000-0000-0000-000000000001',tenant_id,id,1000,3600,'on_creation','2026-08-30T12:00:00Z','active','known','fresh_managed_term','2026-08-30T12:00:00Z',0,0
FROM pvnaive.users WHERE id='b6130000-0000-0000-0000-000000000001';
INSERT INTO pvnaive.user_runtime_credentials(id,tenant_id,user_id,service_term_id,runtime_credential_id,role,bound_at)
SELECT 'e6130000-0000-0000-0000-000000000001',tenant_id,id,'d6130000-0000-0000-0000-000000000001','c6130000-0000-0000-0000-000000000001','primary','2026-08-30T12:00:00Z'
FROM pvnaive.users WHERE id='b6130000-0000-0000-0000-000000000001';
INSERT INTO pvnaive.direct_subscription_tokens(tenant_id,user_id,service_term_id,runtime_credential_id,token_hash,token_prefix,status,user_state,service_state,runtime_username,secret_ciphertext,secret_nonce,encryption_key_id,expires_at,quota_bytes,duration_seconds,start_policy,starts_at,first_connected_at)
SELECT u.tenant_id,u.id,st.id,rc.id,decode(repeat('44',32),'hex'),'task6tok','active',u.status,st.state,rc.username,rc.secret_ciphertext,rc.secret_nonce,rc.encryption_key_id,NULL,st.quota_bytes,st.duration_seconds,st.start_policy,st.starts_at,st.first_connected_at
FROM pvnaive.users u JOIN pvnaive.service_terms st ON st.user_id=u.id JOIN pvnaive.naive_runtime_credentials rc ON rc.id='c6130000-0000-0000-0000-000000000001'
WHERE u.id='b6130000-0000-0000-0000-000000000001';
SQL

PVNAIVE_MIGRATIONS_DIR="${repo_root}/db/migrations" "${repo_root}/scripts/db/migrate.sh" >/dev/null
[[ "$(psql_admin --dbname "$test_db" --tuples-only --no-align --command 'SELECT max(version) FROM pvnaive.schema_migrations')" == 13 ]]
contract="$(PGPASSWORD='subscription-account-ci-only' psql --no-psqlrc --set ON_ERROR_STOP=1 --host "$PVNAIVE_DB_HOST" --port "$PVNAIVE_DB_PORT" --username pvnaive_app --dbname "$test_db" --tuples-only --no-align --command "SELECT service_term_id::text||'|'||accounting_baseline_state||'|'||accounting_baseline_source||'|'||accounting_baseline_upload_bytes||'|'||accounting_baseline_download_bytes FROM pvnaive.resolve_direct_subscription_account_profile(decode(repeat('44',32),'hex'))")"
[[ "$contract" == 'd6130000-0000-0000-0000-000000000001|known|fresh_managed_term|0|0' ]] || { echo "ERROR: schema13 resolver contract=$contract" >&2; exit 1; }

PVNAIVE_DISPOSABLE_DB=1 PVNAIVE_ALLOW_DESTRUCTIVE_ROLLBACK=ROLLBACK_ONE_MIGRATION PVNAIVE_MIGRATIONS_DIR="${repo_root}/db/migrations" "${repo_root}/scripts/db/rollback.sh" >/dev/null
[[ "$(psql_admin --dbname "$test_db" --tuples-only --no-align --command 'SELECT max(version) FROM pvnaive.schema_migrations')" == 12 ]]
columns_after="$(psql_admin --dbname "$test_db" --tuples-only --no-align --command "SELECT count(*) FROM information_schema.columns WHERE table_schema='pvnaive' AND table_name='direct_subscription_tokens' AND column_name LIKE 'accounting_baseline_%'")"
[[ "$columns_after" == 0 ]] || { echo "ERROR: schema13 columns survived rollback=$columns_after" >&2; exit 1; }
echo 'SUBSCRIPTION_ACCOUNT_PROJECTION_MIGRATION_TEST=PASSED'
