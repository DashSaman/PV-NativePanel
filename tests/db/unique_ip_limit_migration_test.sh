#!/usr/bin/env bash
set -Eeuo pipefail
umask 077
repo_root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd -P)"
up="$repo_root/db/migrations/0020_unique_ip_limit.up.sql"
down="$repo_root/db/migrations/0020_unique_ip_limit.down.sql"
[[ -f "$up" ]] || { echo 'RED: missing schema20 unique IP limit migration' >&2; exit 1; }
[[ -f "$down" ]] || { echo 'RED: missing schema20 unique IP limit rollback' >&2; exit 1; }
grep -Fqx -- '-- pvnaive:migration-version 0020' "$up"
grep -Fqx -- '-- pvnaive:transactional true' "$up"
grep -Fqx -- '-- pvnaive:destructive false' "$up"
grep -Fqx -- '-- pvnaive:migration-version 0020' "$down"
grep -Fqx -- '-- pvnaive:destructive true' "$down"

: "${PVNAIVE_DB_HOST:=127.0.0.1}"
: "${PVNAIVE_DB_PORT:=5432}"
: "${PVNAIVE_DB_USER:=postgres}"
export PVNAIVE_DB_HOST PVNAIVE_DB_PORT PVNAIVE_DB_USER
suffix="${GITHUB_RUN_ID:-local}_${GITHUB_RUN_ATTEMPT:-1}_${BASHPID}"; suffix="${suffix//[^a-zA-Z0-9_]/_}"
test_db="pvnaive_migration_test_uniqueip_${suffix,,}"
app_password="pvnaive-uniqueip-ci"
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
for version in $(seq 1 19); do prefix="$(printf '%04d' "$version")"; cp "$repo_root/db/migrations/${prefix}_"*.sql "$tmp/migrations/"; done
( cd "$tmp/migrations" && sha256sum *.sql > SHA256SUMS )
PVNAIVE_MIGRATIONS_DIR="$tmp/migrations" "$repo_root/scripts/db/migrate.sh" >/dev/null
[[ "$(psql_admin -d "$test_db" -Atc 'select max(version) from pvnaive.schema_migrations')" == 19 ]]

cp "$up" "$down" "$tmp/migrations/"
( cd "$tmp/migrations" && sha256sum *.sql > SHA256SUMS )
PVNAIVE_MIGRATIONS_DIR="$tmp/migrations" "$repo_root/scripts/db/migrate.sh" >/dev/null
[[ "$(psql_admin -d "$test_db" -Atc 'select max(version) from pvnaive.schema_migrations')" == 20 ]]
[[ "$(psql_admin -d "$test_db" -Atc "select exists(select 1 from information_schema.columns where table_schema='pvnaive' and table_name='service_terms' and column_name='unique_ip_limit')")" =~ ^(t|true)$ ]]
[[ "$(psql_admin -d "$test_db" -Atc "select exists(select 1 from information_schema.columns where table_schema='pvnaive' and table_name='plans' and column_name='unique_ip_limit')")" =~ ^(t|true)$ ]]

psql_admin -d "$test_db" <<'SQL' >/dev/null
INSERT INTO pvnaive.actors(id,tenant_id,actor_role,email,display_name,status)
VALUES('18180000-0000-0000-0000-000000000011',NULL,'owner','task15-owner@example.invalid','Task15 Owner','active');
INSERT INTO pvnaive.users(id,tenant_id,username,display_name,status,created_by_actor_id)
SELECT '18180000-0000-0000-0000-000000000021',id,'task15-ip-limited','Task15 IP Limited','active','18180000-0000-0000-0000-000000000011'
FROM pvnaive.tenants WHERE tenant_type='system' AND slug='direct';
INSERT INTO pvnaive.naive_runtime_credentials(id,username,secret_hash,secret_ciphertext,secret_nonce,encryption_key_id,status,origin,created_by_actor_id,updated_by_actor_id)
VALUES('18180000-0000-0000-0000-000000000031','task15-ip-limited',decode(repeat('31',32),'hex'),decode(repeat('41',16),'hex'),decode(repeat('51',12),'hex'),'runtime-v1','active','panel','18180000-0000-0000-0000-000000000011','18180000-0000-0000-0000-000000000011');
BEGIN;
SELECT pvnaive.direct_naive_accounting_enter_context();
INSERT INTO pvnaive.service_terms(id,tenant_id,user_id,quota_bytes,duration_seconds,start_policy,purchased_at,starts_at,expires_at,state,concurrency_limit,unique_ip_limit,accounting_baseline_state,accounting_baseline_source,accounting_baseline_cutoff_at,accounting_baseline_upload_bytes,accounting_baseline_download_bytes)
-- unique_ip_limit=2: two unique IPs allowed; third distinct IP must be rejected
SELECT '18180000-0000-0000-0000-000000000041',tenant_id,id,1000000,3600,'on_creation','2026-08-31 00:00:00+00','2026-08-31 00:00:00+00','2026-09-01 00:00:00+00','active',NULL,2,'known','fresh_managed_term','2026-08-31 00:00:00+00',0,0
FROM pvnaive.users WHERE id='18180000-0000-0000-000000000021';
SELECT pvnaive.direct_naive_accounting_leave_context();
COMMIT;
INSERT INTO pvnaive.user_runtime_credentials(id,tenant_id,user_id,service_term_id,runtime_credential_id,role,bound_at)
SELECT '18180000-0000-0000-0000-000000000051',tenant_id,id,'18180000-0000-0000-0000-000000000041','18180000-0000-0000-0000-000000000031','primary','2026-08-31 00:00:00+00'
FROM pvnaive.users WHERE id='18180000-0000-0000-0000-000000000021';
SQL

# Two sessions from the SAME IP should both be accepted (same IP counts once)
same_ip_a="$(psql_app -c "SELECT concat_ws('|',accepted::text,reason) FROM pvnaive.direct_naive_accounting_ingest('18180000-0000-0000-0000-000000000031','task15-ip-limited','node-task15','18180000-0000-0000-000000000061','18180000-0000-0000-000000000071',1,'2026-08-31 01:00:00+00',true,0,0,false,'10.0.0.1')")"
[[ "$same_ip_a" =~ ^(t|true)\|accepted$ ]] || { echo "ERROR: first session from 10.0.0.1 rejected: $same_ip_a" >&2; exit 1; }
same_ip_b="$(psql_app -c "SELECT concat_ws('|',accepted::text,reason) FROM pvnaive.direct_naive_accounting_ingest('18180000-0000-0000-000000000031','task15-ip-limited','node-task15','18180000-0000-0000-000000000061','18180000-0000-0000-000000000072',1,'2026-08-31 01:00:01+00',true,0,0,false,'10.0.0.1')")"
[[ "$same_ip_b" =~ ^(t|true)\|accepted$ ]] || { echo "ERROR: second session from same IP 10.0.0.1 rejected: $same_ip_b" >&2; exit 1; }

# A session from a NEW IP should be accepted as unique #2 (limit=2 allows 2 unique IPs)
new_ip2="$(psql_app -c "SELECT concat_ws('|',accepted::text,reason) FROM pvnaive.direct_naive_accounting_ingest('18180000-0000-0000-0000-000000000031','task15-ip-limited','node-task15','18180000-0000-0000-000000000061','18180000-0000-0000-000000000073',1,'2026-08-31 01:00:02+00',true,0,0,false,'10.0.0.2')")"
[[ "$new_ip2" =~ ^(t|true)\|accepted$ ]] || { echo "ERROR: third session from new IP 10.0.0.2 rejected (should be accepted as unique #2): $new_ip2" >&2; exit 1; }

# A session from a THIRD distinct IP should be rejected (unique_ip_limit=2, 3 distinct IPs exceed limit)
diff_ip="$(psql_app -c "SELECT concat_ws('|',accepted::text,reason) FROM pvnaive.direct_naive_accounting_ingest('18180000-0000-0000-000000000031','task15-ip-limited','node-task15','18180000-0000-0000-000000000061','18180000-0000-0000-000000000074',1,'2026-08-31 01:00:03+00',true,0,0,false,'10.0.0.3')")"
[[ "$diff_ip" =~ ^(f|false)\|unique_ip_limit$ ]] || { echo "ERROR: fourth session from different IP 10.0.0.3 not rejected (should be rejected, exceeds limit=2): $diff_ip" >&2; exit 1; }

# Replay of an exact session identity is still idempotent (duplicate semantics from schema19)
# ORDER BY session_id makes the winner deterministic: ...0071 < ...0072
winner="$(psql_admin -d "$test_db" -Atc "select session_id from pvnaive.direct_naive_accounting_sessions where service_term_id='18180000-0000-0000-0000-000000000041' and not final order by session_id limit 1")"
[[ "$winner" == "18180000-0000-0000-0000-000000000071" ]] || { echo "ERROR: deterministic winner must be ...0071, got: $winner" >&2; exit 1; }
replay="$(psql_app -c "SELECT concat_ws('|',accepted::text,duplicate::text,reason) FROM pvnaive.direct_naive_accounting_ingest('18180000-0000-0000-0000-000000000031','task15-ip-limited','node-task15','18180000-0000-0000-0000-000000000061','$winner',1,'2026-08-31 01:00:00+00',true,0,0,false,'10.0.0.1')")"
[[ "$replay" =~ ^(t|true)\|(t|true)\|duplicate$ ]] || { echo "ERROR: same-session replay lost idempotency: $replay" >&2; exit 1; }

# After closing both sessions, a third different IP should be accepted
psql_app -c "SELECT pvnaive.direct_naive_accounting_ingest('18180000-0000-0000-0000-000000000031','task15-ip-limited','node-task15','18180000-0000-0000-0000-000000000061','$winner',2,'2026-08-31 01:00:10+00',true,0,0,true,'10.0.0.1')" >/dev/null
other="$(psql_admin -d "$test_db" -Atc "select session_id from pvnaive.direct_naive_accounting_sessions where service_term_id='18180000-0000-0000-0000-000000000041' and not final and session_id<>'$winner' limit 1")"
psql_app -c "SELECT pvnaive.direct_naive_accounting_ingest('18180000-0000-0000-0000-000000000031','task15-ip-limited','node-task15','18180000-0000-0000-0000-000000000061','$other',2,'2026-08-31 01:00:11+00',true,0,0,true,'10.0.0.1')" >/dev/null
after_close="$(psql_app -c "SELECT concat_ws('|',accepted::text,reason) FROM pvnaive.direct_naive_accounting_ingest('18180000-0000-0000-0000-000000000031','task15-ip-limited','node-task15','18180000-0000-0000-0000-000000000061','18180000-0000-0000-0000-000000000074',1,'2026-08-31 01:01:00+00',true,0,0,false,'10.0.0.3')")"
[[ "$after_close" =~ ^(t|true)\|accepted$ ]] || { echo "ERROR: new IP after close was rejected: $after_close" >&2; exit 1; }

# Test NULL unlimited term accepts any IP
psql_admin -d "$test_db" <<'SQL' >/dev/null
INSERT INTO pvnaive.users(id,tenant_id,username,display_name,status,created_by_actor_id)
SELECT '18180000-0000-0000-0000-000000000121',id,'task15-ip-unlimited','Task15 IP Unlimited','active','18180000-0000-0000-0000-000000000011'
FROM pvnaive.tenants WHERE tenant_type='system' AND slug='direct';
INSERT INTO pvnaive.naive_runtime_credentials(id,username,secret_hash,secret_ciphertext,secret_nonce,encryption_key_id,status,origin,created_by_actor_id,updated_by_actor_id)
VALUES('18180000-0000-0000-0000-000000000131','task15-ip-unlimited',decode(repeat('32',32),'hex'),decode(repeat('42',16),'hex'),decode(repeat('52',12),'hex'),'runtime-v1','active','panel','18180000-0000-0000-0000-000000000011','18180000-0000-0000-0000-000000000011');
BEGIN;
SELECT pvnaive.direct_naive_accounting_enter_context();
INSERT INTO pvnaive.service_terms(id,tenant_id,user_id,quota_bytes,duration_seconds,start_policy,purchased_at,starts_at,expires_at,state,concurrency_limit,unique_ip_limit,accounting_baseline_state,accounting_baseline_source,accounting_baseline_cutoff_at,accounting_baseline_upload_bytes,accounting_baseline_download_bytes)
SELECT '18180000-0000-0000-0000-000000000141',tenant_id,id,1000000,3600,'on_creation','2026-08-31 00:00:00+00','2026-08-31 00:00:00+00','2026-09-01 00:00:00+00','active',NULL,NULL,'known','fresh_managed_term','2026-08-31 00:00:00+00',0,0
FROM pvnaive.users WHERE id='18180000-0000-0000-0000-000000000121';
SELECT pvnaive.direct_naive_accounting_leave_context();
COMMIT;
INSERT INTO pvnaive.user_runtime_credentials(id,tenant_id,user_id,service_term_id,runtime_credential_id,role,bound_at)
SELECT '18180000-0000-0000-0000-000000000151',tenant_id,id,'18180000-0000-0000-0000-000000000141','18180000-0000-0000-0000-000000000131','primary','2026-08-31 00:00:00+00'
FROM pvnaive.users WHERE id='18180000-0000-0000-0000-000000000121';
SQL
for sid in 171 172 173; do
  out="$(psql_app -c "SELECT concat_ws('|',accepted::text,reason) FROM pvnaive.direct_naive_accounting_ingest('18180000-0000-0000-0000-000000000131','task15-ip-unlimited','node-task15','18180000-0000-0000-0000-000000000161','18180000-0000-0000-0000-000000000${sid}',1,'2026-08-31 02:00:00+00',true,0,0,false,'10.0.${sid##1}.1')")"
  [[ "$out" =~ ^(t|true)\|accepted$ ]] || { echo "ERROR: unlimited term rejected: $out" >&2; exit 1; }
done

# ---------------------------------------------------------------------------
# PostgreSQL concurrency race: two simultaneous first-open calls from
# different trusted IPs with distinct session IDs must produce exactly one
# accepted and one unique_ip_limit rejection.  The ServiceTerm row lock
# (FOR UPDATE OF st) and same-transaction peer insertion guarantee that the
# loser creates NO direct_naive_accounting_sessions row and NO peer row.
# ---------------------------------------------------------------------------
psql_admin -d "$test_db" <<'SQL' >/dev/null
INSERT INTO pvnaive.users(id,tenant_id,username,display_name,status,created_by_actor_id)
SELECT '18180000-0000-0000-0000-000000000221',id,'task15-race-limited','Task15 Race Limited','active','18180000-0000-0000-0000-000000000011'
FROM pvnaive.tenants WHERE tenant_type='system' AND slug='direct';
INSERT INTO pvnaive.naive_runtime_credentials(id,username,secret_hash,secret_ciphertext,secret_nonce,encryption_key_id,status,origin,created_by_actor_id,updated_by_actor_id)
VALUES('18180000-0000-0000-0000-000000000231','task15-race-limited',decode(repeat('33',32),'hex'),decode(repeat('43',16),'hex'),decode(repeat('53',12),'hex'),'runtime-v1','active','panel','18180000-0000-0000-0000-000000000011','18180000-0000-0000-0000-000000000011');
BEGIN;
SELECT pvnaive.direct_naive_accounting_enter_context();
INSERT INTO pvnaive.service_terms(id,tenant_id,user_id,quota_bytes,duration_seconds,start_policy,purchased_at,starts_at,expires_at,state,concurrency_limit,unique_ip_limit,accounting_baseline_state,accounting_baseline_source,accounting_baseline_cutoff_at,accounting_baseline_upload_bytes,accounting_baseline_download_bytes)
SELECT '18180000-0000-0000-0000-000000000241',tenant_id,id,1000000,3600,'on_creation','2026-08-31 00:00:00+00','2026-08-31 00:00:00+00','2026-09-01 00:00:00+00','active',NULL,1,'known','fresh_managed_term','2026-08-31 00:00:00+00',0,0
FROM pvnaive.users WHERE id='18180000-0000-0000-0000-000000000221';
SELECT pvnaive.direct_naive_accounting_leave_context();
COMMIT;
INSERT INTO pvnaive.user_runtime_credentials(id,tenant_id,user_id,service_term_id,runtime_credential_id,role,bound_at)
SELECT '18180000-0000-0000-0000-000000000251',tenant_id,id,'18180000-0000-0000-0000-000000000241','18180000-0000-0000-0000-000000000231','primary','2026-08-31 00:00:00+00'
FROM pvnaive.users WHERE id='18180000-0000-0000-0000-000000000221';
SQL

race_open(){
  local session_id="$1" client_ip="$2" out="$3"
  psql_app -c "SELECT concat_ws('|',accepted::text,reason) FROM pvnaive.direct_naive_accounting_ingest('18180000-0000-0000-0000-000000000231','task15-race-limited','node-task15','18180000-0000-0000-0000-000000000261','$session_id',1,'2026-08-31 03:00:00+00',true,0,0,false,'$client_ip')" >"$out"
}

# Launch two first-open calls simultaneously from two different trusted IPs.
race_open '18180000-0000-0000-0000-000000000271' '10.0.10.1' "$tmp/race_a.out" &
race_a_pid=$!
race_open '18180000-0000-0000-0000-000000000272' '10.0.10.2' "$tmp/race_b.out" &
race_b_pid=$!
wait "$race_a_pid"; wait "$race_b_pid"
cat "$tmp/race_a.out" "$tmp/race_b.out" >"$tmp/race.out"

race_accepted="$(grep -Ec '^(t|true)\|accepted$' "$tmp/race.out" || true)"
race_limited="$(grep -Ec '^(f|false)\|unique_ip_limit$' "$tmp/race.out" || true)"
[[ "$race_accepted" == 1 ]] || { echo "ERROR: race: expected exactly one accepted, got $race_accepted" >&2; cat "$tmp/race.out" >&2; exit 1; }
[[ "$race_limited" == 1 ]] || { echo "ERROR: race: expected exactly one unique_ip_limit, got $race_limited" >&2; cat "$tmp/race.out" >&2; exit 1; }

# Exactly one session row must exist — the loser must NOT have created one.
race_session_count="$(psql_admin -d "$test_db" -Atc "select count(*) from pvnaive.direct_naive_accounting_sessions where service_term_id='18180000-0000-0000-0000-000000000241' and not final")"
[[ "$race_session_count" == 1 ]] || { echo "ERROR: race: expected 1 active session, got $race_session_count" >&2; exit 1; }

# Lock acquisition order is intentionally not assumed.  The persisted winner
# must be exactly the session whose concurrent call returned accepted.
race_a_result="$(tr -d '\r\n' <"$tmp/race_a.out")"
race_b_result="$(tr -d '\r\n' <"$tmp/race_b.out")"
if [[ "$race_a_result" =~ ^(t|true)\|accepted$ ]]; then
  expected_race_winner='18180000-0000-0000-0000-000000000271'
elif [[ "$race_b_result" =~ ^(t|true)\|accepted$ ]]; then
  expected_race_winner='18180000-0000-0000-0000-000000000272'
else
  echo "ERROR: race: neither concurrent call was accepted" >&2
  cat "$tmp/race.out" >&2
  exit 1
fi
race_winner="$(psql_admin -d "$test_db" -Atc "select session_id from pvnaive.direct_naive_accounting_sessions where service_term_id='18180000-0000-0000-0000-000000000241' and not final")"
[[ "$race_winner" == "$expected_race_winner" ]] || { echo "ERROR: race: persisted winner $race_winner does not match accepted call $expected_race_winner" >&2; exit 1; }

# Loser must have NO peer row (same-txn insertion means no leak).
race_peer_count="$(psql_admin -d "$test_db" -Atc "select count(*) from pvnaive.direct_naive_accounting_session_peers where service_term_id='18180000-0000-0000-0000-000000000241'")"
[[ "$race_peer_count" == 1 ]] || { echo "ERROR: race: expected exactly 1 peer row, got $race_peer_count" >&2; exit 1; }

# first_connected_at must reflect only the winner's timestamp.
race_first="$(psql_admin -d "$test_db" -Atc "select first_connected_at from pvnaive.service_terms where id='18180000-0000-0000-0000-000000000241'")"
[[ "$race_first" == "2026-08-31 03:00:00+00" ]] || { echo "ERROR: race: first_connected_at=$race_first, expected 2026-08-31 03:00:00+00" >&2; exit 1; }

# Usage must reflect only the winner (zero bytes for a fresh open).
race_usage="$(psql_admin -d "$test_db" -Atc "select coalesce(upload_bytes,0) + coalesce(download_bytes,0) from pvnaive.direct_naive_accounting_terms where service_term_id='18180000-0000-0000-0000-000000000241'")"
[[ "$race_usage" == 0 ]] || { echo "ERROR: race: usage=$race_usage, expected 0" >&2; exit 1; }

# ---------------------------------------------------------------------------
# Parallel same-IP two sessions must both be accepted while distinct-IP
# count stays at 1 (same IP counts as one unique IP).
# ---------------------------------------------------------------------------
race_open '18180000-0000-0000-0000-000000000273' '10.0.10.1' "$tmp/same_a.out" &
same_a_pid=$!
race_open '18180000-0000-0000-0000-000000000274' '10.0.10.1' "$tmp/same_b.out" &
same_b_pid=$!
wait "$same_a_pid"; wait "$same_b_pid"
cat "$tmp/same_a.out" "$tmp/same_b.out" >"$tmp/same.out"

same_accepted="$(grep -Ec '^(t|true)\|accepted$' "$tmp/same.out" || true)"
[[ "$same_accepted" == 2 ]] || { echo "ERROR: same-IP race: expected 2 accepted, got $same_accepted" >&2; cat "$tmp/same.out" >&2; exit 1; }

# Distinct-IP count must still be 1 (only 10.0.10.1 as active peer IP).
same_ip_count="$(psql_admin -d "$test_db" -Atc "select count(distinct host(client_ip)) from pvnaive.direct_naive_accounting_session_peers where service_term_id='18180000-0000-0000-0000-000000000241'")"
[[ "$same_ip_count" == 1 ]] || { echo "ERROR: same-IP race: expected 1 distinct IP, got $same_ip_count" >&2; exit 1; }

# Total active session count must be 3 (winner + two same-IP).
same_session_count="$(psql_admin -d "$test_db" -Atc "select count(*) from pvnaive.direct_naive_accounting_sessions where service_term_id='18180000-0000-0000-0000-000000000241' and not final")"
[[ "$same_session_count" == 3 ]] || { echo "ERROR: same-IP race: expected 3 active sessions, got $same_session_count" >&2; exit 1; }

# ---------------------------------------------------------------------------
# BLOCKER regression: concurrency_limit rejects while unique_ip_limit would
# allow. The schema20 wrapper must capture the schema19 result, return its
# rejection unchanged, and insert NO peer row.  This proves the RETURN QUERY
# does not leak into an unconditional peer INSERT after a rejection.
# ---------------------------------------------------------------------------
psql_admin -d "$test_db" <<'SQL' >/dev/null
INSERT INTO pvnaive.users(id,tenant_id,username,display_name,status,created_by_actor_id)
SELECT '18180000-0000-0000-0000-000000000321',id,'task15-concurrency-reject','Task15 Concurrency Reject','active','18180000-0000-0000-0000-000000000011'
FROM pvnaive.tenants WHERE tenant_type='system' AND slug='direct';
INSERT INTO pvnaive.naive_runtime_credentials(id,username,secret_hash,secret_ciphertext,secret_nonce,encryption_key_id,status,origin,created_by_actor_id,updated_by_actor_id)
VALUES('18180000-0000-0000-0000-000000000331','task15-concurrency-reject',decode(repeat('34',32),'hex'),decode(repeat('44',16),'hex'),decode(repeat('54',12),'hex'),'runtime-v1','active','panel','18180000-0000-0000-0000-000000000011','18180000-0000-0000-0000-000000000011');
BEGIN;
SELECT pvnaive.direct_naive_accounting_enter_context();
INSERT INTO pvnaive.service_terms(id,tenant_id,user_id,quota_bytes,duration_seconds,start_policy,purchased_at,starts_at,expires_at,state,concurrency_limit,unique_ip_limit,accounting_baseline_state,accounting_baseline_source,accounting_baseline_cutoff_at,accounting_baseline_upload_bytes,accounting_baseline_download_bytes)
SELECT '18180000-0000-0000-0000-000000000341',tenant_id,id,1000000,3600,'on_creation','2026-08-31 00:00:00+00','2026-08-31 00:00:00+00','2026-09-01 00:00:00+00','active',1,10,'known','fresh_managed_term','2026-08-31 00:00:00+00',0,0
FROM pvnaive.users WHERE id='18180000-0000-0000-0000-000000000321';
SELECT pvnaive.direct_naive_accounting_leave_context();
COMMIT;
INSERT INTO pvnaive.user_runtime_credentials(id,tenant_id,user_id,service_term_id,runtime_credential_id,role,bound_at)
SELECT '18180000-0000-0000-0000-000000000351',tenant_id,id,'18180000-0000-0000-0000-000000000341','18180000-0000-0000-0000-000000000331','primary','2026-08-31 00:00:00+00'
FROM pvnaive.users WHERE id='18180000-0000-0000-0000-000000000321';
SQL

# Open one session to fill the concurrency slot (concurrency_limit=1).
cr_first="$(psql_app -c "SELECT concat_ws('|',accepted::text,reason) FROM pvnaive.direct_naive_accounting_ingest('18180000-0000-0000-0000-000000000331','task15-concurrency-reject','node-task15','18180000-0000-0000-0000-000000000361','18180000-0000-0000-0000-000000000371',1,'2026-08-31 04:00:00+00',true,0,0,false,'10.0.50.1')")"
[[ "$cr_first" =~ ^(t|true)\|accepted$ ]] || { echo "ERROR: concurrency-reject: first session rejected: $cr_first" >&2; exit 1; }

# Verify the first session has a peer row.
cr_first_peer="$(psql_admin -d "$test_db" -Atc "select count(*) from pvnaive.direct_naive_accounting_session_peers where service_term_id='18180000-0000-0000-0000-000000000341'")"
[[ "$cr_first_peer" == 1 ]] || { echo "ERROR: concurrency-reject: expected 1 peer after first open, got $cr_first_peer" >&2; exit 1; }

# Attempt a second session from a DIFFERENT IP. unique_ip_limit=10 would allow
# it, but concurrency_limit=1 must block it. The schema20 wrapper must return
# the schema19 rejection without creating any peer row.
cr_second="$(psql_app -c "SELECT concat_ws('|',accepted::text,reason) FROM pvnaive.direct_naive_accounting_ingest('18180000-0000-0000-0000-000000000331','task15-concurrency-reject','node-task15','18180000-0000-0000-0000-000000000361','18180000-0000-0000-0000-000000000372',1,'2026-08-31 04:00:01+00',true,0,0,false,'10.0.50.2')")"
[[ "$cr_second" =~ ^(f|false)\|concurrent_session_limit$ ]] || { echo "ERROR: concurrency-reject: expected concurrent_session_limit, got: $cr_second" >&2; exit 1; }

# The loser must have created NO session row.
cr_session_count="$(psql_admin -d "$test_db" -Atc "select count(*) from pvnaive.direct_naive_accounting_sessions where service_term_id='18180000-0000-0000-0000-000000000341' and not final")"
[[ "$cr_session_count" == 1 ]] || { echo "ERROR: concurrency-reject: expected 1 active session, got $cr_session_count" >&2; exit 1; }

# The loser must have created NO peer row. This is the core blocker regression:
# the schema20 wrapper must not INSERT into session_peers after a schema19
# rejection (RETURN QUERY does not terminate PL/pgSQL).
cr_peer_count="$(psql_admin -d "$test_db" -Atc "select count(*) from pvnaive.direct_naive_accounting_session_peers where service_term_id='18180000-0000-0000-0000-000000000341'")"
[[ "$cr_peer_count" == 1 ]] || { echo "ERROR: concurrency-reject: expected 1 peer (no leak), got $cr_peer_count" >&2; exit 1; }

# ---------------------------------------------------------------------------
# Invalid client_ip must be validated before ::inet cast. Malformed input
# must not become a generic DB error.  Use a NULL concurrency_limit term so
# the concurrency gate does not block us before IP validation.
# ---------------------------------------------------------------------------
psql_admin -d "$test_db" <<'SQL' >/dev/null
INSERT INTO pvnaive.users(id,tenant_id,username,display_name,status,created_by_actor_id)
SELECT '18180000-0000-0000-0000-000000000421',id,'task15-invalid-ip','Task15 Invalid IP','active','18180000-0000-0000-0000-000000000011'
FROM pvnaive.tenants WHERE tenant_type='system' AND slug='direct';
INSERT INTO pvnaive.naive_runtime_credentials(id,username,secret_hash,secret_ciphertext,secret_nonce,encryption_key_id,status,origin,created_by_actor_id,updated_by_actor_id)
VALUES('18180000-0000-0000-0000-000000000431','task15-invalid-ip',decode(repeat('35',32),'hex'),decode(repeat('45',16),'hex'),decode(repeat('55',12),'hex'),'runtime-v1','active','panel','18180000-0000-0000-0000-000000000011','18180000-0000-0000-0000-000000000011');
BEGIN;
SELECT pvnaive.direct_naive_accounting_enter_context();
INSERT INTO pvnaive.service_terms(id,tenant_id,user_id,quota_bytes,duration_seconds,start_policy,purchased_at,starts_at,expires_at,state,concurrency_limit,unique_ip_limit,accounting_baseline_state,accounting_baseline_source,accounting_baseline_cutoff_at,accounting_baseline_upload_bytes,accounting_baseline_download_bytes)
SELECT '18180000-0000-0000-0000-000000000441',tenant_id,id,1000000,3600,'on_creation','2026-08-31 00:00:00+00','2026-08-31 00:00:00+00','2026-09-01 00:00:00+00','active',NULL,1,'known','fresh_managed_term','2026-08-31 00:00:00+00',0,0
FROM pvnaive.users WHERE id='18180000-0000-0000-0000-000000000421';
SELECT pvnaive.direct_naive_accounting_leave_context();
COMMIT;
INSERT INTO pvnaive.user_runtime_credentials(id,tenant_id,user_id,service_term_id,runtime_credential_id,role,bound_at)
SELECT '18180000-0000-0000-0000-000000000451',tenant_id,id,'18180000-0000-0000-0000-000000000441','18180000-0000-0000-0000-000000000431','primary','2026-08-31 00:00:00+00'
FROM pvnaive.users WHERE id='18180000-0000-0000-0000-000000000421';
SQL

# Malformed client_ip must produce a clean rejection, not a raw PG error.
invalid_ip="$(psql_app -c "SELECT concat_ws('|',accepted::text,reason) FROM pvnaive.direct_naive_accounting_ingest('18180000-0000-0000-0000-000000000431','task15-invalid-ip','node-task15','18180000-0000-0000-0000-000000000461','18180000-0000-0000-0000-000000000471',1,'2026-08-31 05:00:00+00',true,0,0,false,'not-an-ip-address')" 2>&1 || true)"
if echo "$invalid_ip" | grep -qiE 'invalid input syntax|inet|type.*error'; then
    echo "ERROR: invalid client_ip leaked a raw DB error: $invalid_ip" >&2; exit 1
fi
[[ "$invalid_ip" =~ ^(f|false)\|invalid_client_ip$ ]] || { echo "ERROR: invalid client_ip: expected invalid_client_ip, got: $invalid_ip" >&2; exit 1; }

# No session or peer row must be created for the invalid-IP attempt.
invalid_session_count="$(psql_admin -d "$test_db" -Atc "select count(*) from pvnaive.direct_naive_accounting_sessions where service_term_id='18180000-0000-0000-0000-000000000441' and not final")"
[[ "$invalid_session_count" == 0 ]] || { echo "ERROR: invalid client_ip: expected 0 sessions, got $invalid_session_count" >&2; exit 1; }
invalid_peer_count="$(psql_admin -d "$test_db" -Atc "select count(*) from pvnaive.direct_naive_accounting_session_peers where service_term_id='18180000-0000-0000-0000-000000000441'")"
[[ "$invalid_peer_count" == 0 ]] || { echo "ERROR: invalid client_ip: expected 0 peers, got $invalid_peer_count" >&2; exit 1; }

PVNAIVE_DISPOSABLE_DB=1 PVNAIVE_MIGRATIONS_DIR="$tmp/migrations" "$repo_root/scripts/db/rollback.sh" >/dev/null
[[ "$(psql_admin -d "$test_db" -Atc 'select max(version) from pvnaive.schema_migrations')" == 19 ]]
[[ "$(psql_admin -d "$test_db" -Atc "select exists(select 1 from information_schema.columns where table_schema='pvnaive' and table_name='service_terms' and column_name='unique_ip_limit')")" =~ ^(f|false)$ ]]
echo 'PVNAIVE_UNIQUE_IP_LIMIT_MIGRATION_TEST=PASSED'
