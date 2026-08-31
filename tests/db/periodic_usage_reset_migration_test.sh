#!/usr/bin/env bash
set -Eeuo pipefail
umask 077
repo_root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd -P)"
up="${repo_root}/db/migrations/0016_periodic_usage_reset.up.sql"
down="${repo_root}/db/migrations/0016_periodic_usage_reset.down.sql"
[[ -f "$up" ]] || { echo 'ERROR: missing schema16 periodic usage reset migration' >&2; exit 1; }
[[ -f "$down" ]] || { echo 'ERROR: missing schema16 periodic usage reset rollback' >&2; exit 1; }
grep -Fqx -- '-- pvnaive:migration-version 0016' "$up"
grep -Fqx -- '-- pvnaive:transactional true' "$up"
grep -Fqx -- '-- pvnaive:destructive false' "$up"
grep -Fqx -- '-- pvnaive:migration-version 0016' "$down"
grep -Fqx -- '-- pvnaive:destructive true' "$down"

: "${PVNAIVE_DB_HOST:=127.0.0.1}"
: "${PVNAIVE_DB_PORT:=5432}"
: "${PVNAIVE_DB_USER:=postgres}"
export PVNAIVE_DB_HOST PVNAIVE_DB_PORT PVNAIVE_DB_USER
suffix="${GITHUB_RUN_ID:-local}_${GITHUB_RUN_ATTEMPT:-1}_${BASHPID}"
suffix="${suffix//[^a-zA-Z0-9_]/_}"
test_db="pvnaive_migration_test_periodic_reset_${suffix,,}"
tmp="$(mktemp -d)"
psql_admin(){ psql --no-psqlrc --set ON_ERROR_STOP=1 -h "$PVNAIVE_DB_HOST" -p "$PVNAIVE_DB_PORT" -U "$PVNAIVE_DB_USER" "$@"; }
cleanup(){
  rm -rf "$tmp" 2>/dev/null || true
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
mkdir -p "$tmp/migrations"
for version in $(seq 1 15); do prefix="$(printf '%04d' "$version")"; cp "${repo_root}/db/migrations/${prefix}_"*.sql "$tmp/migrations/"; done
( cd "$tmp/migrations" && sha256sum *.sql > SHA256SUMS )
PVNAIVE_MIGRATIONS_DIR="$tmp/migrations" "${repo_root}/scripts/db/migrate.sh" >/dev/null
[[ "$(psql_admin -d "$test_db" -Atc 'select max(version) from pvnaive.schema_migrations')" == 15 ]]

psql_admin -d "$test_db" <<'SQL' >/dev/null
INSERT INTO pvnaive.actors(id,tenant_id,actor_role,email,display_name,status)
VALUES('a9160000-0000-0000-0000-000000000001',NULL,'owner','task9-owner@example.invalid','Task9 Owner','active');
INSERT INTO pvnaive.plans(
  id,tenant_id,code,name,status,protocol_id,quota_bytes,duration_seconds,
  base_price_minor,currency,start_policy,no_expiry,reset_strategy,enabled
)
SELECT 'b9160000-0000-0000-0000-000000000001',id,'daily_task9','Daily Task9','active','naive',1000000,2592000,
       0,'EUR','on_creation',false,'daily',true
FROM pvnaive.tenants WHERE tenant_type='system' AND slug='direct';
INSERT INTO pvnaive.users(id,tenant_id,username,display_name,status,created_by_actor_id)
SELECT 'c9160000-0000-0000-0000-000000000001',id,'task9-active','Task9 Active','active','a9160000-0000-0000-0000-000000000001'
FROM pvnaive.tenants WHERE tenant_type='system' AND slug='direct';
INSERT INTO pvnaive.users(id,tenant_id,username,display_name,status,created_by_actor_id)
SELECT 'c9160000-0000-0000-0000-000000000002',id,'task9-pending','Task9 Pending','active','a9160000-0000-0000-0000-000000000001'
FROM pvnaive.tenants WHERE tenant_type='system' AND slug='direct';
BEGIN;
SELECT pvnaive.direct_naive_accounting_enter_context();
INSERT INTO pvnaive.service_terms(
 id,tenant_id,user_id,plan_id,quota_bytes,duration_seconds,start_policy,purchased_at,starts_at,expires_at,state,
 accounting_baseline_state,accounting_baseline_source,accounting_baseline_cutoff_at,
 accounting_baseline_upload_bytes,accounting_baseline_download_bytes
)
SELECT 'd9160000-0000-0000-0000-000000000001',tenant_id,id,'b9160000-0000-0000-0000-000000000001',1000000,2592000,
       'on_creation','2026-08-31 00:00:00+00','2026-08-31 00:00:00+00','2026-09-30 00:00:00+00','active',
       'known','fresh_managed_term','2026-08-31 00:00:00+00',0,0
FROM pvnaive.users WHERE id='c9160000-0000-0000-0000-000000000001';
INSERT INTO pvnaive.service_terms(
 id,tenant_id,user_id,plan_id,quota_bytes,duration_seconds,start_policy,purchased_at,state,
 accounting_baseline_state,accounting_baseline_source,accounting_baseline_cutoff_at,
 accounting_baseline_upload_bytes,accounting_baseline_download_bytes
)
SELECT 'd9160000-0000-0000-0000-000000000002',tenant_id,id,'b9160000-0000-0000-0000-000000000001',1000000,2592000,
       'on_first_successful_connection','2026-08-31 00:00:00+00','pending',
       'known','fresh_managed_term','2026-08-31 00:00:00+00',0,0
FROM pvnaive.users WHERE id='c9160000-0000-0000-0000-000000000002';
SELECT pvnaive.direct_naive_accounting_leave_context();
COMMIT;
SQL

cp "${repo_root}/db/migrations/0016_"*.sql "$tmp/migrations/"
( cd "$tmp/migrations" && sha256sum *.sql > SHA256SUMS )
PVNAIVE_MIGRATIONS_DIR="$tmp/migrations" "${repo_root}/scripts/db/migrate.sh" >/dev/null
[[ "$(psql_admin -d "$test_db" -Atc 'select max(version) from pvnaive.schema_migrations')" == 16 ]]

flags="$(psql_admin -d "$test_db" -Atc "select relname||':'||relrowsecurity::text||':'||relforcerowsecurity::text from pg_class where oid in ('pvnaive.service_term_reset_schedules'::regclass,'pvnaive.scheduled_usage_reset_attempts'::regclass) order by relname")"
grep -Eq '^scheduled_usage_reset_attempts:(t|true):(t|true)$' <<<"$flags"
grep -Eq '^service_term_reset_schedules:(t|true):(t|true)$' <<<"$flags"

math="$(psql_admin -d "$test_db" -Atc "
SELECT pvnaive.next_usage_reset_due('2026-08-31 00:00:00+00','daily',NULL) AT TIME ZONE 'UTC';
SELECT pvnaive.next_usage_reset_due('2026-08-31 00:00:00+00','weekly',NULL) AT TIME ZONE 'UTC';
SELECT pvnaive.next_usage_reset_due('2026-01-31 12:00:00+00','monthly',NULL) AT TIME ZONE 'UTC';
SELECT pvnaive.next_usage_reset_due('2024-02-29 12:00:00+00','yearly',NULL) AT TIME ZONE 'UTC';
SELECT pvnaive.next_usage_reset_due('2026-08-31 00:00:00+00','custom',3) AT TIME ZONE 'UTC';
SELECT pvnaive.next_usage_reset_due('2026-08-31 00:00:00+00','none',NULL) IS NULL;")"
expected=$'2026-09-01 00:00:00\n2026-09-07 00:00:00\n2026-02-28 12:00:00\n2025-02-28 12:00:00\n2026-09-03 00:00:00\nt'
[[ "$math" == "$expected" || "$math" == "${expected%$'t'}true" ]]

active="$(psql_admin -d "$test_db" -Atc "select strategy||'|'||timezone_name||'|'||(anchor_at at time zone 'UTC')||'|'||(next_due_at at time zone 'UTC') from pvnaive.service_term_reset_schedules where service_term_id='d9160000-0000-0000-0000-000000000001'")"
[[ "$active" == 'daily|UTC|2026-08-31 00:00:00|2026-09-01 00:00:00' ]]
pending="$(psql_admin -d "$test_db" -Atc "select strategy||'|'||(anchor_at is null)::text||'|'||(next_due_at is null)::text from pvnaive.service_term_reset_schedules where service_term_id='d9160000-0000-0000-0000-000000000002'")"
[[ "$pending" == 'daily|true|true' || "$pending" == 'daily|t|t' ]]

# First successful connection path uses the existing signed synthetic accounting context.
psql_admin -d "$test_db" <<'SQL' >/dev/null
BEGIN;
SELECT pvnaive.direct_naive_accounting_enter_context();
UPDATE pvnaive.service_terms
   SET starts_at='2026-08-31 05:00:00+00', first_connected_at='2026-08-31 05:00:00+00',
       expires_at='2026-09-30 05:00:00+00', state='active'
 WHERE id='d9160000-0000-0000-0000-000000000002';
SELECT pvnaive.direct_naive_accounting_leave_context();
COMMIT;
SQL
activated="$(psql_admin -d "$test_db" -Atc "select (anchor_at at time zone 'UTC')||'|'||(next_due_at at time zone 'UTC') from pvnaive.service_term_reset_schedules where service_term_id='d9160000-0000-0000-0000-000000000002'")"
[[ "$activated" == '2026-08-31 05:00:00|2026-09-01 05:00:00' ]]

# A later Plan edit cannot rewrite an already-adopted ServiceTerm schedule.
psql_admin -d "$test_db" -c "UPDATE pvnaive.plans SET reset_strategy='monthly' WHERE id='b9160000-0000-0000-0000-000000000001'" >/dev/null
[[ "$(psql_admin -d "$test_db" -Atc "select strategy from pvnaive.service_term_reset_schedules where service_term_id='d9160000-0000-0000-0000-000000000001'")" == daily ]]

# renew_current copies the frozen ServiceTerm policy, not the now-edited Plan.
psql_admin -d "$test_db" <<'SQL' >/dev/null
BEGIN;
SELECT pvnaive.direct_naive_accounting_enter_context();
INSERT INTO pvnaive.service_terms(
 id,tenant_id,user_id,plan_id,quota_bytes,duration_seconds,start_policy,purchased_at,starts_at,expires_at,state,
 renewal_kind,renewed_from_term_id,
 accounting_baseline_state,accounting_baseline_source,accounting_baseline_cutoff_at,
 accounting_baseline_upload_bytes,accounting_baseline_download_bytes
)
SELECT 'd9160000-0000-0000-0000-000000000003',tenant_id,id,'b9160000-0000-0000-0000-000000000001',1000000,2592000,
       'on_creation','2026-09-01 00:00:00+00','2026-09-01 00:00:00+00','2026-10-01 00:00:00+00','active',
       'renew_current','d9160000-0000-0000-0000-000000000001',
       'known','fresh_managed_term','2026-09-01 00:00:00+00',0,0
FROM pvnaive.users WHERE id='c9160000-0000-0000-0000-000000000001';
SELECT pvnaive.direct_naive_accounting_leave_context();
COMMIT;
SQL
[[ "$(psql_admin -d "$test_db" -Atc "select strategy from pvnaive.service_term_reset_schedules where service_term_id='d9160000-0000-0000-0000-000000000003'")" == daily ]]

actor="$(psql_admin -d "$test_db" -Atc "select actor_role||'|'||status||'|'||(password_hash is null)::text from pvnaive.actors where id='00000000-0000-0000-0000-000000000016'")"
[[ "$actor" == 'owner|disabled|true' || "$actor" == 'owner|disabled|t' ]]

PVNAIVE_DISPOSABLE_DB=1 PVNAIVE_MIGRATIONS_DIR="$tmp/migrations" "${repo_root}/scripts/db/rollback.sh" >/dev/null
[[ "$(psql_admin -d "$test_db" -Atc 'select max(version) from pvnaive.schema_migrations')" == 15 ]]
[[ "$(psql_admin -d "$test_db" -Atc "select to_regclass('pvnaive.service_term_reset_schedules') is null")" =~ ^(t|true)$ ]]
[[ "$(psql_admin -d "$test_db" -Atc "select count(*) from pvnaive.actors where id='00000000-0000-0000-0000-000000000016'")" == 0 ]]
echo 'PVNAIVE_PERIODIC_USAGE_RESET_MIGRATION_TEST=PASSED'
