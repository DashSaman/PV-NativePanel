#!/usr/bin/env bash
set -Eeuo pipefail
umask 077
repo_root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd -P)"
up="${repo_root}/db/migrations/0015_bulk_usage_reset.up.sql"
down="${repo_root}/db/migrations/0015_bulk_usage_reset.down.sql"
[[ -f "$up" ]] || { echo 'ERROR: missing schema15 bulk usage reset migration' >&2; exit 1; }
[[ -f "$down" ]] || { echo 'ERROR: missing schema15 bulk usage reset rollback' >&2; exit 1; }
grep -Fqx -- '-- pvnaive:migration-version 0015' "$up"
grep -Fqx -- '-- pvnaive:transactional true' "$up"
grep -Fqx -- '-- pvnaive:destructive false' "$up"
grep -Fqx -- '-- pvnaive:migration-version 0015' "$down"
grep -Fqx -- '-- pvnaive:transactional true' "$down"
grep -Fqx -- '-- pvnaive:destructive true' "$down"
: "${PVNAIVE_DB_HOST:=127.0.0.1}"
: "${PVNAIVE_DB_PORT:=5432}"
: "${PVNAIVE_DB_USER:=postgres}"
export PVNAIVE_DB_HOST PVNAIVE_DB_PORT PVNAIVE_DB_USER
suffix="${GITHUB_RUN_ID:-local}_${GITHUB_RUN_ATTEMPT:-1}_${BASHPID}"
suffix="${suffix//[^a-zA-Z0-9_]/_}"
test_db="pvnaive_migration_test_bulk_reset_${suffix,,}"
fixture="$(mktemp -d)"
psql_admin(){ psql --no-psqlrc --set ON_ERROR_STOP=1 -h "$PVNAIVE_DB_HOST" -p "$PVNAIVE_DB_PORT" -U "$PVNAIVE_DB_USER" "$@"; }
cleanup(){
  rm -rf -- "$fixture" 2>/dev/null || true
  psql_admin -d postgres -c "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname='${test_db}' AND pid<>pg_backend_pid()" >/dev/null 2>&1 || true
  dropdb --if-exists -h "$PVNAIVE_DB_HOST" -p "$PVNAIVE_DB_PORT" -U "$PVNAIVE_DB_USER" "$test_db" >/dev/null 2>&1 || true
  psql_admin -d postgres -c 'DROP ROLE IF EXISTS pvnaive_app; DROP ROLE IF EXISTS pvnaive_owner;' >/dev/null 2>&1 || true
}
trap cleanup EXIT HUP INT TERM
cleanup
psql_admin -d postgres <<'SQL' >/dev/null
CREATE ROLE pvnaive_owner NOLOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOREPLICATION NOBYPASSRLS;
CREATE ROLE pvnaive_app NOLOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOREPLICATION NOBYPASSRLS;
SQL
createdb -h "$PVNAIVE_DB_HOST" -p "$PVNAIVE_DB_PORT" -U "$PVNAIVE_DB_USER" --owner pvnaive_owner --encoding UTF8 --template template0 "$test_db"
export PVNAIVE_DB_NAME="$test_db"
mkdir -p "$fixture/migrations"
for version in $(seq 1 14); do prefix="$(printf '%04d' "$version")"; cp "${repo_root}/db/migrations/${prefix}_"*.sql "$fixture/migrations/"; done
( cd "$fixture/migrations" && sha256sum *.sql > SHA256SUMS )
PVNAIVE_MIGRATIONS_DIR="$fixture/migrations" "${repo_root}/scripts/db/migrate.sh" >/dev/null
[[ "$(psql_admin -d "$test_db" -Atc 'SELECT max(version) FROM pvnaive.schema_migrations')" == 14 ]]
psql_admin -d "$test_db" <<'SQL' >/dev/null
INSERT INTO pvnaive.actors(id,tenant_id,actor_role,email,display_name,status)
VALUES('a8150000-0000-0000-0000-000000000001',NULL,'owner','bulk-reset@example.invalid','Bulk Reset Test','active');
INSERT INTO pvnaive.customer_bulk_operations(id,tenant_id,actor_id,idempotency_key,request_hash,action,request,status,preview)
SELECT 'c8150000-0000-0000-0000-000000000001',t.id,'a8150000-0000-0000-0000-000000000001',
       'legacy-bulk-key-15',decode(repeat('80',32),'hex'),'suspend',
       '{"action":"suspend","customer_ids":["example"]}'::jsonb,'previewed',
       '{"requested":1,"affected":1,"changes":["suspend"],"conflicts":[],"skipped":[],"invalid":[]}'::jsonb
FROM pvnaive.tenants t WHERE t.tenant_type='system' AND t.slug='direct';
SQL
cp "${repo_root}/db/migrations/0015_"*.sql "$fixture/migrations/"
( cd "$fixture/migrations" && sha256sum *.sql > SHA256SUMS )
PVNAIVE_MIGRATIONS_DIR="$fixture/migrations" "${repo_root}/scripts/db/migrate.sh" >/dev/null
[[ "$(psql_admin -d "$test_db" -Atc 'SELECT max(version) FROM pvnaive.schema_migrations')" == 15 ]]
[[ "$(psql_admin -d "$test_db" -Atc "SELECT to_regclass('pvnaive.customer_bulk_reset_operations') IS NOT NULL")" =~ ^(t|true)$ ]]
constraint="$(psql_admin -d "$test_db" -Atc "SELECT pg_get_constraintdef(oid) FROM pg_constraint WHERE conrelid='pvnaive.customer_bulk_reset_operations'::regclass AND contype='c' AND pg_get_constraintdef(oid) LIKE '%reset_usage%' LIMIT 1")"
grep -Fq "reset_usage" <<<"$constraint"
[[ "$(psql_admin -d "$test_db" -Atc "SELECT to_regclass('pvnaive.customer_bulk_operation_keys') IS NOT NULL")" =~ ^(t|true)$ ]]
[[ "$(psql_admin -d "$test_db" -Atc "SELECT action FROM pvnaive.customer_bulk_operation_keys WHERE idempotency_key='legacy-bulk-key-15'")" == suspend ]]
[[ "$(psql_admin -d "$test_db" -Atc "SELECT relforcerowsecurity FROM pg_class WHERE oid='pvnaive.customer_bulk_operations'::regclass")" =~ ^(t|true)$ ]]
set +e
psql_admin -d "$test_db" -c "INSERT INTO pvnaive.customer_bulk_operation_keys(tenant_id,actor_id,idempotency_key,request_hash,action) SELECT t.id,'a8150000-0000-0000-0000-000000000001','legacy-bulk-key-15',decode(repeat('82',32),'hex'),'reset_usage' FROM pvnaive.tenants t WHERE t.tenant_type='system' AND t.slug='direct'" >/dev/null 2>&1
duplicate_key_rc=$?
set -e
[[ "$duplicate_key_rc" -ne 0 ]] || { echo 'ERROR: shared bulk idempotency namespace accepted duplicate key' >&2; exit 1; }
psql_admin -d "$test_db" <<'SQL' >/dev/null
INSERT INTO pvnaive.customer_bulk_operation_keys(tenant_id,actor_id,idempotency_key,request_hash,action)
SELECT t.id,'a8150000-0000-0000-0000-000000000001','bulk-reset-schema15',decode(repeat('81',32),'hex'),'reset_usage'
FROM pvnaive.tenants t WHERE t.tenant_type='system' AND t.slug='direct';
INSERT INTO pvnaive.customer_bulk_reset_operations(id,tenant_id,actor_id,idempotency_key,request_hash,action,request,status,preview)
SELECT 'b8150000-0000-0000-0000-000000000001',t.id,'a8150000-0000-0000-0000-000000000001',
       'bulk-reset-schema15',decode(repeat('81',32),'hex'),'reset_usage',
       '{"action":"reset_usage","customer_ids":["example"]}'::jsonb,'previewed',
       '{"requested":1,"affected":1,"changes":["reset_usage"],"conflicts":[],"skipped":[],"invalid":[]}'::jsonb
FROM pvnaive.tenants t WHERE t.tenant_type='system' AND t.slug='direct';
SQL
set +e
PVNAIVE_DISPOSABLE_DB=1 PVNAIVE_MIGRATIONS_DIR="$fixture/migrations" "${repo_root}/scripts/db/rollback.sh" >/dev/null 2>&1
blocked_rc=$?
set -e
[[ "$blocked_rc" -ne 0 ]] || { echo 'ERROR: schema15 rollback discarded reset_usage bulk history' >&2; exit 1; }
[[ "$(psql_admin -d "$test_db" -Atc 'SELECT max(version) FROM pvnaive.schema_migrations')" == 15 ]]
psql_admin -d "$test_db" -c "DELETE FROM pvnaive.customer_bulk_reset_operations WHERE id='b8150000-0000-0000-0000-000000000001'" >/dev/null
PVNAIVE_DISPOSABLE_DB=1 PVNAIVE_MIGRATIONS_DIR="$fixture/migrations" "${repo_root}/scripts/db/rollback.sh" >/dev/null
[[ "$(psql_admin -d "$test_db" -Atc 'SELECT max(version) FROM pvnaive.schema_migrations')" == 14 ]]
for table in customer_bulk_reset_operations customer_bulk_operation_keys; do
  [[ "$(psql_admin -d "$test_db" -Atc "SELECT to_regclass('pvnaive.${table}') IS NULL")" =~ ^(t|true)$ ]] || {
    echo "ERROR: schema14 rollback left ${table} behind" >&2
    exit 1
  }
done
[[ "$(psql_admin -d "$test_db" -Atc "SELECT action FROM pvnaive.customer_bulk_operations WHERE idempotency_key='legacy-bulk-key-15'")" == suspend ]]
echo 'PVNAIVE_BULK_USAGE_RESET_MIGRATION_TEST=PASSED'
