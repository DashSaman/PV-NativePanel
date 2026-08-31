#!/usr/bin/env bash
set -Eeuo pipefail
umask 077
repo_root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd -P)"
up="$repo_root/db/migrations/0019_concurrent_session_limit.up.sql"
down="$repo_root/db/migrations/0019_concurrent_session_limit.down.sql"
[[ -f "$up" ]] || { echo 'RED: missing schema19 concurrent session limit migration' >&2; exit 1; }
[[ -f "$down" ]] || { echo 'RED: missing schema19 concurrent session limit rollback' >&2; exit 1; }
grep -Fqx -- '-- pvnaive:migration-version 0019' "$up"
grep -Fqx -- '-- pvnaive:transactional true' "$up"
grep -Fqx -- '-- pvnaive:destructive false' "$up"
grep -Fqx -- '-- pvnaive:migration-version 0019' "$down"
grep -Fqx -- '-- pvnaive:destructive true' "$down"

: "${PVNAIVE_DB_HOST:=127.0.0.1}"
: "${PVNAIVE_DB_PORT:=5432}"
: "${PVNAIVE_DB_USER:=postgres}"
export PVNAIVE_DB_HOST PVNAIVE_DB_PORT PVNAIVE_DB_USER
suffix="${GITHUB_RUN_ID:-local}_${GITHUB_RUN_ATTEMPT:-1}_${BASHPID}"; suffix="${suffix//[^a-zA-Z0-9_]/_}"
test_db="pvnaive_migration_test_concurrency_${suffix,,}"
app_password="pvnaive-concurrency-ci"
tmp="$(mktemp -d)"
psql_admin(){ psql --no-psqlrc --set ON_ERROR_STOP=1 -h "$PVNAIVE_DB_HOST" -p "$PVNAIVE_DB_PORT" -U "$PVNAIVE_DB_USER" "$@"; }
psql_app(){ PGPASSWORD="$app_password" psql --no-psqlrc --set ON_ERROR_STOP=1 -qAt -h "$PVNAIVE_DB_HOST" -p "$PVNAIVE_DB_PORT" -U pvnaive_app -d "$test_db" "$@"; }
cleanup(){
  rm -rf "$tmp" 2>/dev/null || true
  psql_admin -d postgres -c "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname='${test_db}' AND pid<>pg_backend_pid()" >/dev/null 2>&1 || true
  dropdb --if-exists -h "$PVNAIVE_DB_HOST" -p "$PVNAIVE_DB_PORT" -U "$PVNAIVE_DB_USER" "$test_db" >/dev/null 2>&1 || true
  psql_admin -d postgres -c 'DROP ROLE IF EXISTS pvnaive_app; DROP ROLE IF EXISTS pvnaive_owner;' >/dev/null 2>&1 || true
}
trap cleanup EXIT HUP INT TERM
cleanup
psql_admin -d postgres <<SQL >/dev/null
CREATE ROLE pvnaive_owner NOLOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOREPLICATION NOBYPASSRLS;
CREATE ROLE pvnaive_app LOGIN PASSWORD '${app_password}' NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOREPLICATION NOBYPASSRLS;
SQL
createdb -h "$PVNAIVE_DB_HOST" -p "$PVNAIVE_DB_PORT" -U "$PVNAIVE_DB_USER" --owner pvnaive_owner --encoding UTF8 --template template0 "$test_db"
export PVNAIVE_DB_NAME="$test_db"
mkdir -p "$tmp/migrations"
for version in $(seq 1 18); do prefix="$(printf '%04d' "$version")"; cp "$repo_root/db/migrations/${prefix}_"*.sql "$tmp/migrations/"; done
( cd "$tmp/migrations" && sha256sum *.sql > SHA256SUMS )
PVNAIVE_MIGRATIONS_DIR="$tmp/migrations" "$repo_root/scripts/db/migrate.sh" >/dev/null
[[ "$(psql_admin -d "$test_db" -Atc 'select max(version) from pvnaive.schema_migrations')" == 18 ]]

cp "$up" "$down" "$tmp/migrations/"
( cd "$tmp/migrations" && sha256sum *.sql > SHA256SUMS )
PVNAIVE_MIGRATIONS_DIR="$tmp/migrations" "$repo_root/scripts/db/migrate.sh" >/dev/null
[[ "$(psql_admin -d "$test_db" -Atc 'select max(version) from pvnaive.schema_migrations')" == 19 ]]
[[ "$(psql_admin -d "$test_db" -Atc "select exists(select 1 from information_schema.columns where table_schema='pvnaive' and table_name='service_terms' and column_name='concurrency_limit')")" =~ ^(t|true)$ ]]

psql_admin -d "$test_db" <<'SQL' >/dev/null
INSERT INTO pvnaive.actors(id,tenant_id,actor_role,email,display_name,status)
VALUES('18180000-0000-0000-0000-000000000011',NULL,'owner','task14-owner@example.invalid','Task14 Owner','active');
INSERT INTO pvnaive.users(id,tenant_id,username,display_name,status,created_by_actor_id)
SELECT '18180000-0000-0000-0000-000000000021',id,'task14-limited','Task14 Limited','active','18180000-0000-0000-0000-000000000011'
FROM pvnaive.tenants WHERE tenant_type='system' AND slug='direct';
INSERT INTO pvnaive.naive_runtime_credentials(id,username,secret_hash,secret_ciphertext,secret_nonce,encryption_key_id,status,origin,created_by_actor_id,updated_by_actor_id)
VALUES('18180000-0000-0000-0000-000000000031','task14-limited',decode(repeat('31',32),'hex'),decode(repeat('41',16),'hex'),decode(repeat('51',12),'hex'),'runtime-v1','active','panel','18180000-0000-0000-0000-000000000011','18180000-0000-0000-0000-000000000011');
BEGIN;
SELECT pvnaive.direct_naive_accounting_enter_context();
INSERT INTO pvnaive.service_terms(id,tenant_id,user_id,quota_bytes,duration_seconds,start_policy,purchased_at,starts_at,expires_at,state,concurrency_limit,accounting_baseline_state,accounting_baseline_source,accounting_baseline_cutoff_at,accounting_baseline_upload_bytes,accounting_baseline_download_bytes)
SELECT '18180000-0000-0000-0000-000000000041',tenant_id,id,1000000,3600,'on_creation','2026-08-31 00:00:00+00','2026-08-31 00:00:00+00','2026-09-01 00:00:00+00','active',1,'known','fresh_managed_term','2026-08-31 00:00:00+00',0,0
FROM pvnaive.users WHERE id='18180000-0000-0000-0000-000000000021';
SELECT pvnaive.direct_naive_accounting_leave_context();
COMMIT;
INSERT INTO pvnaive.user_runtime_credentials(id,tenant_id,user_id,service_term_id,runtime_credential_id,role,bound_at)
SELECT '18180000-0000-0000-0000-000000000051',tenant_id,id,'18180000-0000-0000-0000-000000000041','18180000-0000-0000-0000-000000000031','primary','2026-08-31 00:00:00+00'
FROM pvnaive.users WHERE id='18180000-0000-0000-0000-000000000021';
SQL

open_session(){
  local session_id="$1" out="$2"
  psql_app -c "SELECT concat_ws('|',accepted::text,reason) FROM pvnaive.direct_naive_accounting_ingest('18180000-0000-0000-0000-000000000031','task14-limited','node-task14','18180000-0000-0000-0000-000000000061','$session_id',1,'2026-08-31 01:00:00+00',true,0,0,false)" >"$out"
}
open_session '18180000-0000-0000-0000-000000000071' "$tmp/a.out" & a_pid=$!
open_session '18180000-0000-0000-0000-000000000072' "$tmp/b.out" & b_pid=$!
wait "$a_pid"; wait "$b_pid"
cat "$tmp/a.out" "$tmp/b.out" >"$tmp/opens.out"
accepted_count="$(grep -Ec '^(t|true)\|accepted$' "$tmp/opens.out" || true)"
limited_count="$(grep -Ec '^(f|false)\|concurrent_session_limit$' "$tmp/opens.out" || true)"
[[ "$accepted_count" == 1 ]] || { echo "ERROR: expected one accepted race winner" >&2; cat "$tmp/opens.out" >&2; exit 1; }
[[ "$limited_count" == 1 ]] || { echo "ERROR: expected one concurrency-limited race loser" >&2; cat "$tmp/opens.out" >&2; exit 1; }
[[ "$(psql_admin -d "$test_db" -Atc "select count(*) from pvnaive.direct_naive_accounting_sessions where service_term_id='18180000-0000-0000-0000-000000000041' and not final")" == 1 ]]

winner="$(psql_admin -d "$test_db" -Atc "select session_id from pvnaive.direct_naive_accounting_sessions where service_term_id='18180000-0000-0000-0000-000000000041' and not final limit 1")"
replay="$(psql_app -c "SELECT concat_ws('|',accepted::text,duplicate::text,reason) FROM pvnaive.direct_naive_accounting_ingest('18180000-0000-0000-0000-000000000031','task14-limited','node-task14','18180000-0000-0000-0000-000000000061','$winner',1,'2026-08-31 01:00:00+00',true,0,0,false)")"
[[ "$replay" =~ ^(t|true)\|(t|true)\|duplicate$ ]] || { echo "ERROR: same-session replay lost idempotency: $replay" >&2; exit 1; }

closed="$(psql_app -c "SELECT concat_ws('|',accepted::text,reason) FROM pvnaive.direct_naive_accounting_ingest('18180000-0000-0000-0000-000000000031','task14-limited','node-task14','18180000-0000-0000-0000-000000000061','$winner',2,'2026-08-31 01:00:02+00',true,0,0,true)")"
[[ "$closed" =~ ^(t|true)\|accepted$ ]]
third="$(psql_app -c "SELECT concat_ws('|',accepted::text,reason) FROM pvnaive.direct_naive_accounting_ingest('18180000-0000-0000-0000-000000000031','task14-limited','node-task14','18180000-0000-0000-0000-000000000061','18180000-0000-0000-0000-000000000073',1,'2026-08-31 01:00:03+00',true,0,0,false)")"
[[ "$third" =~ ^(t|true)\|accepted$ ]] || { echo "ERROR: reconnect after final close was blocked: $third" >&2; exit 1; }

# A crashed/non-final session must not consume a slot forever. At >90s since
# its last observation, a new exact session is eligible for admission.
stale_replacement="$(psql_app -c "SELECT concat_ws('|',accepted::text,reason) FROM pvnaive.direct_naive_accounting_ingest('18180000-0000-0000-0000-000000000031','task14-limited','node-task14','18180000-0000-0000-0000-000000000061','18180000-0000-0000-0000-000000000074',1,'2026-08-31 01:01:34+00',true,0,0,false)")"
[[ "$stale_replacement" =~ ^(t|true)\|accepted$ ]] || { echo "ERROR: stale non-final session permanently consumed slot: $stale_replacement" >&2; exit 1; }

# An accounting-incomplete session is not a trustworthy live-session slot.
incomplete="$(psql_app -c "SELECT concat_ws('|',accepted::text,reason) FROM pvnaive.direct_naive_accounting_ingest('18180000-0000-0000-0000-000000000031','task14-limited','node-task14','18180000-0000-0000-0000-000000000061','18180000-0000-0000-0000-000000000074',3,'2026-08-31 01:01:35+00',true,0,0,false)")"
[[ "$incomplete" =~ ^(f|false)\|sequence_gap$ ]] || { echo "ERROR: expected sequence gap to mark session incomplete: $incomplete" >&2; exit 1; }
incomplete_replacement="$(psql_app -c "SELECT concat_ws('|',accepted::text,reason) FROM pvnaive.direct_naive_accounting_ingest('18180000-0000-0000-0000-000000000031','task14-limited','node-task14','18180000-0000-0000-0000-000000000061','18180000-0000-0000-0000-000000000075',1,'2026-08-31 01:01:36+00',true,0,0,false)")"
[[ "$incomplete_replacement" =~ ^(t|true)\|accepted$ ]] || { echo "ERROR: accounting-incomplete session consumed slot: $incomplete_replacement" >&2; exit 1; }

psql_admin -d "$test_db" <<'SQL' >/dev/null
INSERT INTO pvnaive.users(id,tenant_id,username,display_name,status,created_by_actor_id)
SELECT '18180000-0000-0000-0000-000000000121',id,'task14-unlimited','Task14 Unlimited','active','18180000-0000-0000-0000-000000000011'
FROM pvnaive.tenants WHERE tenant_type='system' AND slug='direct';
INSERT INTO pvnaive.naive_runtime_credentials(id,username,secret_hash,secret_ciphertext,secret_nonce,encryption_key_id,status,origin,created_by_actor_id,updated_by_actor_id)
VALUES('18180000-0000-0000-0000-000000000131','task14-unlimited',decode(repeat('32',32),'hex'),decode(repeat('42',16),'hex'),decode(repeat('52',12),'hex'),'runtime-v1','active','panel','18180000-0000-0000-0000-000000000011','18180000-0000-0000-0000-000000000011');
BEGIN;
SELECT pvnaive.direct_naive_accounting_enter_context();
INSERT INTO pvnaive.service_terms(id,tenant_id,user_id,quota_bytes,duration_seconds,start_policy,purchased_at,starts_at,expires_at,state,concurrency_limit,accounting_baseline_state,accounting_baseline_source,accounting_baseline_cutoff_at,accounting_baseline_upload_bytes,accounting_baseline_download_bytes)
SELECT '18180000-0000-0000-0000-000000000141',tenant_id,id,1000000,3600,'on_creation','2026-08-31 00:00:00+00','2026-08-31 00:00:00+00','2026-09-01 00:00:00+00','active',NULL,'known','fresh_managed_term','2026-08-31 00:00:00+00',0,0
FROM pvnaive.users WHERE id='18180000-0000-0000-0000-000000000121';
SELECT pvnaive.direct_naive_accounting_leave_context();
COMMIT;
INSERT INTO pvnaive.user_runtime_credentials(id,tenant_id,user_id,service_term_id,runtime_credential_id,role,bound_at)
SELECT '18180000-0000-0000-0000-000000000151',tenant_id,id,'18180000-0000-0000-0000-000000000141','18180000-0000-0000-0000-000000000131','primary','2026-08-31 00:00:00+00'
FROM pvnaive.users WHERE id='18180000-0000-0000-0000-000000000121';
SQL
for sid in 171 172; do
  out="$(psql_app -c "SELECT concat_ws('|',accepted::text,reason) FROM pvnaive.direct_naive_accounting_ingest('18180000-0000-0000-0000-000000000131','task14-unlimited','node-task14','18180000-0000-0000-0000-000000000161','18180000-0000-0000-0000-000000000${sid}',1,'2026-08-31 02:00:00+00',true,0,0,false)")"
  [[ "$out" =~ ^(t|true)\|accepted$ ]] || { echo "ERROR: unlimited term rejected: $out" >&2; exit 1; }
done

PVNAIVE_DISPOSABLE_DB=1 PVNAIVE_MIGRATIONS_DIR="$tmp/migrations" "$repo_root/scripts/db/rollback.sh" >/dev/null
[[ "$(psql_admin -d "$test_db" -Atc 'select max(version) from pvnaive.schema_migrations')" == 18 ]]
[[ "$(psql_admin -d "$test_db" -Atc "select exists(select 1 from information_schema.columns where table_schema='pvnaive' and table_name='service_terms' and column_name='concurrency_limit')")" =~ ^(f|false)$ ]]
echo 'PVNAIVE_CONCURRENT_SESSION_LIMIT_MIGRATION_TEST=PASSED'
