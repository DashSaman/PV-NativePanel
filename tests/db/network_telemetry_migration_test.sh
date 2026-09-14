#!/usr/bin/env bash
set -Eeuo pipefail
umask 077

# R1 / STEER-001 migration gate: proves the network telemetry schema contract,
# idempotent ingest (restart/replay never double-counts), partition auto-ensure
# at a future month boundary, monotonic aggregate upserts, and the trusted
# boundary (pvnaive_app has no direct table access; unknown credentials are
# honestly reported untracked, never guessed).

repo_root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd -P)"
suffix="${GITHUB_RUN_ID:-local}_${GITHUB_RUN_ATTEMPT:-1}_${BASHPID}"
suffix="${suffix//[^a-zA-Z0-9_]/_}"
test_db="pvnaive_r1_network_${suffix,,}"
: "${PVNAIVE_DB_HOST:=127.0.0.1}"
: "${PVNAIVE_DB_PORT:=5432}"
: "${PVNAIVE_DB_USER:=postgres}"
export PVNAIVE_DB_HOST PVNAIVE_DB_PORT PVNAIVE_DB_USER

psql_admin() {
  psql --no-psqlrc --set ON_ERROR_STOP=1 --host "${PVNAIVE_DB_HOST}" \
    --port "${PVNAIVE_DB_PORT}" --username "${PVNAIVE_DB_USER}" "$@"
}
cleanup() {
  psql_admin --dbname postgres --command \
    "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname='${test_db}' AND pid <> pg_backend_pid()" \
    >/dev/null 2>&1 || true
  dropdb --if-exists --host "${PVNAIVE_DB_HOST}" --port "${PVNAIVE_DB_PORT}" \
    --username "${PVNAIVE_DB_USER}" "${test_db}" >/dev/null 2>&1 || true
  psql_admin --dbname postgres --command 'DROP ROLE IF EXISTS pvnaive_app; DROP ROLE IF EXISTS pvnaive_owner;' >/dev/null 2>&1 || true
}
trap cleanup EXIT
cleanup

psql_admin --dbname postgres <<'SQL' >/dev/null
CREATE ROLE pvnaive_owner NOLOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOREPLICATION NOBYPASSRLS;
CREATE ROLE pvnaive_app NOLOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOREPLICATION NOBYPASSRLS;
SQL
createdb --host "${PVNAIVE_DB_HOST}" --port "${PVNAIVE_DB_PORT}" --username "${PVNAIVE_DB_USER}" \
  --owner pvnaive_owner --encoding UTF8 --template template0 "${test_db}"
PVNAIVE_DB_NAME="${test_db}" "${repo_root}/scripts/db/migrate.sh" >/dev/null

schema="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command 'SELECT COALESCE(MAX(version),0) FROM pvnaive.schema_migrations')"
[[ "${schema}" =~ ^[0-9]+$ && "${schema}" -ge 31 ]] || { echo "ERROR: expected schema >=31, got ${schema}" >&2; exit 1; }

# 1) Schema contract.
contract="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command "
SELECT concat_ws('|',
 (to_regclass('pvnaive.session_network_samples') IS NOT NULL)::text,
 (to_regclass('pvnaive.user_node_network_agg') IS NOT NULL)::text,
 (to_regprocedure('pvnaive.network_sample_ingest(uuid,text,uuid,uuid,bigint,timestamptz,text,integer,integer,bigint,bigint,bigint,bigint,bigint)') IS NOT NULL)::text,
 (to_regprocedure('pvnaive.network_agg_upsert(uuid,text,text,bigint,bigint,real,bigint,real,bigint,timestamptz)') IS NOT NULL)::text,
 (to_regprocedure('pvnaive.network_agg_read(bigint)') IS NOT NULL)::text,
 (to_regproc('pvnaive.network_ensure_partition') IS NOT NULL)::text
);")"
[[ "${contract}" == 'true|true|true|true|true|true' ]] || {
  echo "ERROR: network telemetry schema contract failed: ${contract}" >&2; exit 1;
}

# Fixture: owner, user, active runtime credential (same shape as accounting gate).
psql_admin --dbname "${test_db}" <<'SQL' >/dev/null
INSERT INTO pvnaive.actors (id, tenant_id, actor_role, email, display_name, status)
VALUES ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', NULL, 'owner', 'owner@r1.invalid', 'R1 Owner', 'active');

INSERT INTO pvnaive.users (id, tenant_id, username, display_name, status, created_by_actor_id)
SELECT 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', id, 'r1alice', 'R1 Alice', 'active', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa'
FROM pvnaive.tenants WHERE slug='direct' AND tenant_type='system';

INSERT INTO pvnaive.naive_runtime_credentials (
  id, username, secret_hash, secret_ciphertext, secret_nonce, encryption_key_id,
  status, origin, created_by_actor_id, updated_by_actor_id
) VALUES (
  '11111111-1111-1111-1111-111111111111', 'r1alice', decode(repeat('11',32),'hex'),
  decode(repeat('12',32),'hex'), decode(repeat('13',12),'hex'), 'runtime-v1',
  'active', 'panel', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa'
);
SQL

ingest() {
  psql_admin --dbname "${test_db}" --tuples-only --no-align --command "SET ROLE pvnaive_app; SELECT concat_ws('|',tracked::text,accepted::text,duplicate::text,reason,COALESCE(user_id::text,'-')) FROM pvnaive.network_sample_ingest('11111111-1111-1111-1111-111111111111','direct-1','22222222-2222-2222-2222-222222222222','33333333-3333-3333-3333-333333333333',${1},'2026-09-14T10:00:0${2}Z','${3}',${4},${5},${6},${7},${8},${9},${10});"
}

# 2) First ingest: accepted, user resolved through the accounting join.
first="$(ingest 1 1 upstream 42000 8000 1200 4 100000 200000 60000000)"
[[ "${first##*$'\n'}" == "true|false|false|accepted|bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb" ]] || {
  echo "ERROR: first ingest: ${first}" >&2; exit 1;
}

# 3) Identical replay (restart/reload) is a duplicate and adds NO row (STEER-001).
replay="$(ingest 1 1 upstream 42000 8000 1200 4 100000 200000 60000000)"
[[ "${replay##*$'\n'}" == "true|false|true|duplicate|bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb" ]] || {
  echo "ERROR: replay dedup: ${replay}" >&2; exit 1;
}
count="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command "SELECT count(*) FROM pvnaive.session_network_samples;")"
[[ "${count}" == 1 ]] || { echo "ERROR: replay double-counted: count=${count}" >&2; exit 1; }

# 4) New sample_seq in the same session is accepted.
second="$(ingest 2 2 upstream 45000 9000 1500 6 200000 400000 120000000)"
[[ "${second##*$'\n'}" == "true|false|false|accepted|bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb" ]] || {
  echo "ERROR: second ingest: ${second}" >&2; exit 1;
}

# 5) Client-path sample is stored under the same identity.
third="$(ingest 3 3 client 18000 3000 900 2 50000 80000 30000000)"
[[ "${third##*$'\n'}" == "true|false|false|accepted|bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb" ]] || {
  echo "ERROR: client-path ingest: ${third}" >&2; exit 1;
}

# 6) Unknown credential is honestly untracked (fail-closed, never guessed).
untracked="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command "SET ROLE pvnaive_app; SELECT concat_ws('|',tracked::text,accepted::text,reason) FROM pvnaive.network_sample_ingest('99999999-9999-9999-9999-999999999999','direct-1','22222222-2222-2222-2222-222222222222','33333333-3333-3333-3333-333333333333',1,'2026-09-14T10:00:05Z','upstream',10000,100,10,0,1000,1000,1000000);")"
[[ "${untracked##*$'\n'}" == 'false|false|untracked' ]] || {
  echo "ERROR: untracked credential: ${untracked}" >&2; exit 1;
}

# 7) Malformed sample is rejected (retrans > segs_out impossible).
set +e
malformed_rc="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command "SET ROLE pvnaive_app; SELECT pvnaive.network_sample_ingest('11111111-1111-1111-1111-111111111111','direct-1','22222222-2222-2222-2222-222222222222','33333333-3333-3333-3333-333333333333',4,'2026-09-14T10:00:06Z','upstream',10000,100,10,50,1000,1000,1000000);" >/dev/null 2>&1; echo $?)"
set -e
[[ "${malformed_rc}" != 0 ]] || { echo 'ERROR: malformed sample accepted' >&2; exit 1; }

# 8) Partition auto-ensure at a future month boundary (no 2027-03 partition exists).
future="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command "SET ROLE pvnaive_app; SELECT accepted::text FROM pvnaive.network_sample_ingest('11111111-1111-1111-1111-111111111111','direct-1','22222222-2222-2222-2222-222222222222','33333333-3333-3333-3333-333333333333',5,'2027-03-15T10:00:00Z','upstream',20000,100,20,1,1000,2000,1000000);")"
[[ "${future##*$'\n'}" == t ]] || { echo "ERROR: future partition ingest: ${future}" >&2; exit 1; }
auto_partition="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command "SELECT to_regclass('pvnaive.session_network_samples_2027_03') IS NOT NULL;")"
[[ "${auto_partition}" == t ]] || { echo 'ERROR: partition auto-ensure failed' >&2; exit 1; }

# 9) Aggregate upsert: monotonic guard — newer wins, older is ignored.
agg_new="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command "SET ROLE pvnaive_app; SELECT pvnaive.network_agg_upsert('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb','direct-1','client',19000,3200,0.02,700000,1.0,8,'2026-09-14T10:01:00Z');")"
[[ "${agg_new##*$'\n'}" == t ]] || { echo "ERROR: agg upsert new: ${agg_new}" >&2; exit 1; }
agg_old="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command "SET ROLE pvnaive_app; SELECT pvnaive.network_agg_upsert('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb','direct-1','client',99000,99000,0.99,1,0.1,3,'2026-09-14T09:00:00Z');")"
[[ "${agg_old##*$'\n'}" == t ]] || { echo "ERROR: agg upsert old call: ${agg_old}" >&2; exit 1; }
agg_check="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command "SELECT concat_ws('|',rtt_ewma_micros::text,sample_count::text) FROM pvnaive.user_node_network_agg WHERE user_id='bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb' AND node_id='direct-1' AND path='client';")"
[[ "${agg_check}" == '19000|8' ]] || { echo "ERROR: older aggregate overwrote newer: ${agg_check}" >&2; exit 1; }

# 10) network_agg_read honors staleness (fresh row returned, aged-out omitted).
fresh_read="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command "SET ROLE pvnaive_app; SELECT count(*) FROM pvnaive.network_agg_read(3600);")"
[[ "${fresh_read}" == 1 ]] || { echo "ERROR: fresh read count: ${fresh_read}" >&2; exit 1; }
aged_read="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command "SET ROLE pvnaive_app; SELECT count(*) FROM pvnaive.network_agg_read(1);")"
[[ "${aged_read}" == 0 ]] || { echo "ERROR: stale read should be empty: ${aged_read}" >&2; exit 1; }

# 11) Trusted boundary: pvnaive_app cannot touch the tables directly.
set +e
psql_admin --dbname "${test_db}" --command "SET ROLE pvnaive_app; SELECT count(*) FROM pvnaive.session_network_samples;" >/dev/null 2>&1
direct_rc=$?
set -e
[[ "${direct_rc}" != 0 ]] || { echo 'ERROR: pvnaive_app read telemetry table directly' >&2; exit 1; }

echo "NETWORK_TELEMETRY_MIGRATION_GATE=PASSED"
