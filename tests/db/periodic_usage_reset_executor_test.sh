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
test_db="pvnaive_migration_test_periodic_executor_${suffix,,}"
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
cp "${repo_root}/db/migrations/"[0-9][0-9][0-9][0-9]_*.sql "$tmp/migrations/"
( cd "$tmp/migrations" && sha256sum *.sql > SHA256SUMS )
PVNAIVE_MIGRATIONS_DIR="$tmp/migrations" "${repo_root}/scripts/db/migrate.sh" >/dev/null
[[ "$(psql_admin -d "$test_db" -Atc 'select max(version) from pvnaive.schema_migrations')" == 16 ]]

psql_admin -d "$test_db" <<'SQL' >/dev/null
INSERT INTO pvnaive.actors(id,tenant_id,actor_role,email,display_name,status)
VALUES('a9260000-0000-0000-0000-000000000001',NULL,'owner','task9-executor-owner@example.invalid','Task9 Executor Owner','active');
INSERT INTO pvnaive.plans(
  id,tenant_id,code,name,status,protocol_id,quota_bytes,duration_seconds,
  base_price_minor,currency,start_policy,no_expiry,reset_strategy,enabled
)
SELECT 'b9260000-0000-0000-0000-000000000001',id,'daily_exec_task9','Daily Executor Task9','active','naive',1000000,2592000,
       0,'EUR','on_creation',false,'daily',true
FROM pvnaive.tenants WHERE tenant_type='system' AND slug='direct';
INSERT INTO pvnaive.users(id,tenant_id,username,display_name,status,created_by_actor_id)
SELECT 'c9260000-0000-0000-0000-000000000001',id,'task9-success','Task9 Success','active','a9260000-0000-0000-0000-000000000001'
FROM pvnaive.tenants WHERE tenant_type='system' AND slug='direct';
INSERT INTO pvnaive.users(id,tenant_id,username,display_name,status,created_by_actor_id)
SELECT 'c9260000-0000-0000-0000-000000000002',id,'task9-defer','Task9 Defer','active','a9260000-0000-0000-0000-000000000001'
FROM pvnaive.tenants WHERE tenant_type='system' AND slug='direct';
BEGIN;
SELECT pvnaive.direct_naive_accounting_enter_context();
INSERT INTO pvnaive.service_terms(
 id,tenant_id,user_id,plan_id,quota_bytes,duration_seconds,start_policy,purchased_at,starts_at,expires_at,state,
 accounting_baseline_state,accounting_baseline_source,accounting_baseline_cutoff_at,
 accounting_baseline_upload_bytes,accounting_baseline_download_bytes
)
SELECT 'd9260000-0000-0000-0000-000000000001',tenant_id,id,'b9260000-0000-0000-0000-000000000001',1000000,2592000,
       'on_creation',clock_timestamp()-interval '2 days',clock_timestamp()-interval '2 days',clock_timestamp()+interval '28 days','active',
       'known','fresh_managed_term',clock_timestamp()-interval '2 days',0,0
FROM pvnaive.users WHERE id='c9260000-0000-0000-0000-000000000001';
INSERT INTO pvnaive.service_terms(
 id,tenant_id,user_id,plan_id,quota_bytes,duration_seconds,start_policy,purchased_at,starts_at,expires_at,state,
 accounting_baseline_state,accounting_baseline_source,accounting_baseline_cutoff_at,
 accounting_baseline_upload_bytes,accounting_baseline_download_bytes
)
SELECT 'd9260000-0000-0000-0000-000000000002',tenant_id,id,'b9260000-0000-0000-0000-000000000001',1000000,2592000,
       'on_creation',clock_timestamp()-interval '2 days',clock_timestamp()-interval '2 days',clock_timestamp()+interval '28 days','active',
       'known','fresh_managed_term',clock_timestamp()-interval '2 days',0,0
FROM pvnaive.users WHERE id='c9260000-0000-0000-0000-000000000002';
SELECT pvnaive.direct_naive_accounting_leave_context();
COMMIT;
INSERT INTO pvnaive.direct_naive_accounting_terms(service_term_id,upload_bytes,download_bytes,reserved_bytes,accounting_complete)
VALUES
 ('d9260000-0000-0000-0000-000000000001',120,80,0,true),
 ('d9260000-0000-0000-0000-000000000002',10,20,5,true);
SQL

# Both schedules are already overdue: one succeeds, one defers on reservation.
first="$(psql_admin -d "$test_db" -At -F'|' <<'SQL'
SET ROLE pvnaive_app;
SELECT processed,succeeded,deferred,skipped FROM pvnaive.execute_due_scheduled_usage_resets(10);
RESET ROLE;
SQL
)"
first="$(printf '%s\n' "$first" | grep -E '^[0-9]+\|' | tail -n1)"
[[ "$first" == '2|1|1|0' ]] || { echo "ERROR: first scheduler batch=$first" >&2; exit 1; }

success_projection="$(psql_admin -d "$test_db" -At -F'|' -c "SELECT upload_bytes,download_bytes,reserved_bytes,(last_reset_at IS NOT NULL)::text FROM pvnaive.direct_naive_accounting_terms WHERE service_term_id='d9260000-0000-0000-0000-000000000001'")"
[[ "$success_projection" == '0|0|0|true' || "$success_projection" == '0|0|0|t' ]]
[[ "$(psql_admin -d "$test_db" -Atc "SELECT count(*) FROM pvnaive.direct_naive_accounting_reset_events WHERE service_term_id='d9260000-0000-0000-0000-000000000001' AND reason='scheduled' AND actor_id='00000000-0000-0000-0000-000000000016'")" == 1 ]]
[[ "$(psql_admin -d "$test_db" -Atc "SELECT count(*) FROM pvnaive.audit_events WHERE object_id='d9260000-0000-0000-0000-000000000001' AND action='customer.usage.reset' AND outcome='success'")" == 1 ]]
[[ "$(psql_admin -d "$test_db" -Atc "SELECT count(*) FROM pvnaive.scheduled_usage_reset_attempts WHERE service_term_id='d9260000-0000-0000-0000-000000000001' AND outcome='success' AND reset_event_id IS NOT NULL")" == 1 ]]
[[ "$(psql_admin -d "$test_db" -Atc "SELECT (next_due_at-anchor_at)=interval '1 day' FROM pvnaive.service_term_reset_schedules WHERE service_term_id='d9260000-0000-0000-0000-000000000001'")" =~ ^(t|true)$ ]]

# Deferred target is untouched and receives bounded retry state/history.
defer_projection="$(psql_admin -d "$test_db" -At -F'|' -c "SELECT upload_bytes,download_bytes,reserved_bytes FROM pvnaive.direct_naive_accounting_terms WHERE service_term_id='d9260000-0000-0000-0000-000000000002'")"
[[ "$defer_projection" == '10|20|5' ]]
[[ "$(psql_admin -d "$test_db" -Atc "SELECT count(*) FROM pvnaive.direct_naive_accounting_reset_events WHERE service_term_id='d9260000-0000-0000-0000-000000000002' AND reason='scheduled'")" == 0 ]]
[[ "$(psql_admin -d "$test_db" -Atc "SELECT count(*) FROM pvnaive.scheduled_usage_reset_attempts WHERE service_term_id='d9260000-0000-0000-0000-000000000002' AND outcome='deferred' AND reason='reservation_pending'")" == 1 ]]
[[ "$(psql_admin -d "$test_db" -Atc "SELECT retry_after_at > last_attempt_at AND consecutive_failures=1 FROM pvnaive.service_term_reset_schedules WHERE service_term_id='d9260000-0000-0000-0000-000000000002'")" =~ ^(t|true)$ ]]

# Immediate restart/re-run cannot duplicate success or bypass retry_after.
second="$(psql_admin -d "$test_db" -At -F'|' <<'SQL'
SET ROLE pvnaive_app;
SELECT processed,succeeded,deferred,skipped FROM pvnaive.execute_due_scheduled_usage_resets(10);
RESET ROLE;
SQL
)"
second="$(printf '%s\n' "$second" | grep -E '^[0-9]+\|' | tail -n1)"
[[ "$second" == '0|0|0|0' ]] || { echo "ERROR: replay scheduler batch=$second" >&2; exit 1; }
[[ "$(psql_admin -d "$test_db" -Atc "SELECT count(*) FROM pvnaive.direct_naive_accounting_reset_events WHERE reason='scheduled'")" == 1 ]]

# A proven explicit reset newer than the missed due boundary satisfies that
# occurrence without a second reset. The next cadence starts from that epoch.
psql_admin -d "$test_db" <<'SQL' >/dev/null
UPDATE pvnaive.direct_naive_accounting_terms
   SET upload_bytes=0,download_bytes=0,reserved_bytes=0,last_reset_at=clock_timestamp(),updated_at=clock_timestamp()
 WHERE service_term_id='d9260000-0000-0000-0000-000000000002';
UPDATE pvnaive.service_term_reset_schedules
   SET retry_after_at=NULL
 WHERE service_term_id='d9260000-0000-0000-0000-000000000002';
SQL
third="$(psql_admin -d "$test_db" -At -F'|' <<'SQL'
SET ROLE pvnaive_app;
SELECT processed,succeeded,deferred,skipped FROM pvnaive.execute_due_scheduled_usage_resets(10);
RESET ROLE;
SQL
)"
third="$(printf '%s\n' "$third" | grep -E '^[0-9]+\|' | tail -n1)"
[[ "$third" == '1|0|0|1' ]] || { echo "ERROR: explicit-reset scheduler batch=$third" >&2; exit 1; }
[[ "$(psql_admin -d "$test_db" -Atc "SELECT count(*) FROM pvnaive.scheduled_usage_reset_attempts WHERE service_term_id='d9260000-0000-0000-0000-000000000002' AND outcome='skipped' AND reason='explicit_reset_satisfied_period'")" == 1 ]]
[[ "$(psql_admin -d "$test_db" -Atc "SELECT next_due_at-anchor_at=interval '1 day' AND retry_after_at IS NULL AND consecutive_failures=0 FROM pvnaive.service_term_reset_schedules WHERE service_term_id='d9260000-0000-0000-0000-000000000002'")" =~ ^(t|true)$ ]]
[[ "$(psql_admin -d "$test_db" -Atc "SELECT count(*) FROM pvnaive.direct_naive_accounting_reset_events WHERE reason='scheduled'")" == 1 ]]

# Once immutable scheduler history exists, destructive rollback must fail closed.
set +e
PVNAIVE_DISPOSABLE_DB=1 PVNAIVE_ALLOW_DESTRUCTIVE_ROLLBACK=ROLLBACK_ONE_MIGRATION PVNAIVE_MIGRATIONS_DIR="$tmp/migrations" "${repo_root}/scripts/db/rollback.sh" >/dev/null 2>&1
rollback_rc=$?
set -e
(( rollback_rc != 0 )) || { echo 'ERROR: schema16 rollback did not fail closed with scheduler history' >&2; exit 1; }
[[ "$(psql_admin -d "$test_db" -Atc 'select max(version) from pvnaive.schema_migrations')" == 16 ]]

echo 'PVNAIVE_PERIODIC_USAGE_RESET_EXECUTOR_TEST=PASSED'
