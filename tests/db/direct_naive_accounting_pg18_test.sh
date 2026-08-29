#!/usr/bin/env bash
set -Eeuo pipefail
umask 077

repo_root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd -P)"
suffix="${GITHUB_RUN_ID:-local}_${GITHUB_RUN_ATTEMPT:-1}_${BASHPID}"
suffix="${suffix//[^a-zA-Z0-9_]/_}"
test_db="pvnaive_ws1_accounting_${suffix,,}"
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
[[ "${schema}" =~ ^[0-9]+$ && "${schema}" -ge 10 ]] || { echo "ERROR: expected schema >=10, got ${schema}" >&2; exit 1; }

contract="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command "
SELECT concat_ws('|',
 (to_regclass('pvnaive.direct_naive_accounting_events') IS NOT NULL)::text,
 (to_regclass('pvnaive.direct_naive_accounting_sessions') IS NOT NULL)::text,
 (to_regclass('pvnaive.direct_naive_accounting_terms') IS NOT NULL)::text,
 (to_regclass('pvnaive.direct_naive_accounting_claims') IS NOT NULL)::text,
 (to_regprocedure('pvnaive.direct_naive_accounting_authorize(uuid,timestamptz)') IS NOT NULL)::text,
 (to_regprocedure('pvnaive.direct_naive_accounting_claim(uuid,text,uuid,uuid,bigint,text,bigint,timestamptz)') IS NOT NULL)::text,
 (to_regprocedure('pvnaive.direct_naive_accounting_ingest(uuid,text,text,uuid,uuid,bigint,timestamptz,boolean,bigint,bigint,boolean)') IS NOT NULL)::text,
 (to_regprocedure('pvnaive.direct_naive_accounting_read(uuid,timestamptz,bigint)') IS NOT NULL)::text
);")"
[[ "${contract}" == 'true|true|true|true|true|true|true|true' ]] || {
  echo "ERROR: accounting schema contract failed: ${contract}" >&2; exit 1;
}

psql_admin --dbname "${test_db}" <<'SQL' >/dev/null
INSERT INTO pvnaive.actors (id, tenant_id, actor_role, email, display_name, status)
VALUES ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', NULL, 'owner', 'owner@ws1.invalid', 'WS1 Owner', 'active');

INSERT INTO pvnaive.users (id, tenant_id, username, display_name, status, created_by_actor_id)
SELECT 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', id, 'alice', 'Alice', 'active', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa'
FROM pvnaive.tenants WHERE slug='direct' AND tenant_type='system';

INSERT INTO pvnaive.naive_runtime_credentials (
  id, username, secret_hash, secret_ciphertext, secret_nonce, encryption_key_id,
  status, origin, created_by_actor_id, updated_by_actor_id
) VALUES (
  '11111111-1111-1111-1111-111111111111', 'alice', decode(repeat('01',32),'hex'),
  decode(repeat('02',32),'hex'), decode(repeat('03',12),'hex'), 'runtime-v1',
  'active', 'panel', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa'
);

INSERT INTO pvnaive.service_terms (
  id, tenant_id, user_id, quota_bytes, duration_seconds, start_policy, purchased_at, state
)
SELECT 'cccccccc-cccc-cccc-cccc-cccccccccccc', tenant_id, id, 100, 3600,
       'on_first_successful_connection', '2026-08-29T17:00:00Z', 'pending'
FROM pvnaive.users WHERE id='bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb';

INSERT INTO pvnaive.user_runtime_credentials (
  id, tenant_id, user_id, service_term_id, runtime_credential_id, role, bound_at
)
SELECT 'dddddddd-dddd-dddd-dddd-dddddddddddd', tenant_id, id,
       'cccccccc-cccc-cccc-cccc-cccccccccccc', '11111111-1111-1111-1111-111111111111',
       'primary', '2026-08-29T17:00:00Z'
FROM pvnaive.users WHERE id='bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb';
SQL

# Authorization is read/check only: it must never activate first-use.
auth="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command "SET ROLE pvnaive_app; SELECT concat_ws('|',tracked::text,allowed::text,remaining_bytes::text) FROM pvnaive.direct_naive_accounting_authorize('11111111-1111-1111-1111-111111111111','2026-08-29T18:00:00Z');")"
[[ "${auth##*$'\n'}" == 'true|true|100' ]] || { echo "ERROR: authorize: ${auth}" >&2; exit 1; }
not_started="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command "SELECT first_connected_at IS NULL AND starts_at IS NULL FROM pvnaive.service_terms WHERE id='cccccccc-cccc-cccc-cccc-cccccccccccc';")"
[[ "${not_started}" == t ]] || { echo 'ERROR: authorize started first-use' >&2; exit 1; }

# Schema 10 truthfulness: any pending reservation is immediately incomplete.
psql_admin --dbname "${test_db}" --command "UPDATE pvnaive.direct_naive_accounting_terms SET reserved_bytes=1 WHERE service_term_id='cccccccc-cccc-cccc-cccc-cccccccccccc';" >/dev/null
pending_auth="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command "SET ROLE pvnaive_app; SELECT concat_ws('|',remaining_bytes::text,accounting_complete::text) FROM pvnaive.direct_naive_accounting_authorize('11111111-1111-1111-1111-111111111111','2026-08-29T18:00:00Z');")"
[[ "${pending_auth##*$'\n'}" == '99|false' ]] || { echo "ERROR: schema10 pending authorize: ${pending_auth}" >&2; exit 1; }
pending_read="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command "SET ROLE pvnaive_app; SELECT concat_ws('|',remaining_bytes::text,accounting_complete::text) FROM pvnaive.direct_naive_accounting_read('cccccccc-cccc-cccc-cccc-cccccccccccc','2026-08-29T18:00:00Z',90);")"
[[ "${pending_read##*$'\n'}" == '99|false' ]] || { echo "ERROR: schema10 pending read: ${pending_read}" >&2; exit 1; }
psql_admin --dbname "${test_db}" --command "UPDATE pvnaive.direct_naive_accounting_terms SET reserved_bytes=0 WHERE service_term_id='cccccccc-cccc-cccc-cccc-cccccccccccc';" >/dev/null
complete_again="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command "SET ROLE pvnaive_app; SELECT accounting_complete FROM pvnaive.direct_naive_accounting_authorize('11111111-1111-1111-1111-111111111111','2026-08-29T18:00:00Z');")"
[[ "${complete_again##*$'\n'}" == t ]] || { echo "ERROR: schema10 settled completeness: ${complete_again}" >&2; exit 1; }

# Sequence 1 is emitted only after authenticated CONNECT + successful target dial.
open1="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command "SET ROLE pvnaive_app; SELECT concat_ws('|',accepted::text,duplicate::text,reason) FROM pvnaive.direct_naive_accounting_ingest('11111111-1111-1111-1111-111111111111','alice','direct-1','22222222-2222-2222-2222-222222222222','33333333-3333-3333-3333-333333333333',1,'2026-08-29T18:00:01Z',true,0,0,false);")"
[[ "${open1##*$'\n'}" == 'true|false|accepted' ]] || { echo "ERROR: open event: ${open1}" >&2; exit 1; }
started="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command "SELECT first_connected_at='2026-08-29T18:00:01Z'::timestamptz AND starts_at=first_connected_at AND expires_at=first_connected_at+interval '1 hour' FROM pvnaive.service_terms WHERE id='cccccccc-cccc-cccc-cccc-cccccccccccc';")"
[[ "${started}" == t ]] || { echo 'ERROR: first successful CONNECT activation mismatch' >&2; exit 1; }

# Duplicate is idempotent, same sequence with changed payload is a conflict.
dup="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command "SET ROLE pvnaive_app; SELECT concat_ws('|',accepted::text,duplicate::text,reason) FROM pvnaive.direct_naive_accounting_ingest('11111111-1111-1111-1111-111111111111','alice','direct-1','22222222-2222-2222-2222-222222222222','33333333-3333-3333-3333-333333333333',1,'2026-08-29T18:00:01Z',true,0,0,false);")"
[[ "${dup##*$'\n'}" == 'true|true|duplicate' ]] || { echo "ERROR: duplicate: ${dup}" >&2; exit 1; }
conflict="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command "SET ROLE pvnaive_app; SELECT concat_ws('|',accepted::text,reason) FROM pvnaive.direct_naive_accounting_ingest('11111111-1111-1111-1111-111111111111','alice','direct-1','22222222-2222-2222-2222-222222222222','33333333-3333-3333-3333-333333333333',1,'2026-08-29T18:00:01Z',true,1,0,false);")"
[[ "${conflict##*$'\n'}" == 'false|sequence_conflict' ]] || { echo "ERROR: sequence conflict: ${conflict}" >&2; exit 1; }

# Open a second simultaneous session against the same ServiceTerm budget.
open2="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command "SET ROLE pvnaive_app; SELECT accepted FROM pvnaive.direct_naive_accounting_ingest('11111111-1111-1111-1111-111111111111','alice','direct-1','22222222-2222-2222-2222-222222222222','44444444-4444-4444-4444-444444444444',1,'2026-08-29T18:00:02Z',true,0,0,false);")"
[[ "${open2##*$'\n'}" == t ]] || { echo "ERROR: second session open: ${open2}" >&2; exit 1; }

claim1="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command "SET ROLE pvnaive_app; SELECT granted_bytes FROM pvnaive.direct_naive_accounting_claim('11111111-1111-1111-1111-111111111111','direct-1','22222222-2222-2222-2222-222222222222','33333333-3333-3333-3333-333333333333',2,'upload',80,'2026-08-29T18:00:03Z');")"
claim2="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command "SET ROLE pvnaive_app; SELECT granted_bytes FROM pvnaive.direct_naive_accounting_claim('11111111-1111-1111-1111-111111111111','direct-1','22222222-2222-2222-2222-222222222222','44444444-4444-4444-4444-444444444444',2,'upload',80,'2026-08-29T18:00:03Z');")"
[[ "${claim1##*$'\n'}" == 80 && "${claim2##*$'\n'}" == 20 ]] || { echo "ERROR: shared reservation ${claim1}/${claim2}" >&2; exit 1; }

settle1="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command "SET ROLE pvnaive_app; SELECT concat_ws('|',accepted::text,upload_delta::text,remaining_bytes::text) FROM pvnaive.direct_naive_accounting_ingest('11111111-1111-1111-1111-111111111111','alice','direct-1','22222222-2222-2222-2222-222222222222','33333333-3333-3333-3333-333333333333',2,'2026-08-29T18:00:04Z',true,80,0,false);")"
[[ "${settle1##*$'\n'}" == 'true|80|0' ]] || { echo "ERROR: settle1: ${settle1}" >&2; exit 1; }
settle2="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command "SET ROLE pvnaive_app; SELECT concat_ws('|',accepted::text,upload_delta::text,quota_depleted::text) FROM pvnaive.direct_naive_accounting_ingest('11111111-1111-1111-1111-111111111111','alice','direct-1','22222222-2222-2222-2222-222222222222','44444444-4444-4444-4444-444444444444',2,'2026-08-29T18:00:05Z',true,20,0,false);")"
[[ "${settle2##*$'\n'}" == 'true|20|true' ]] || { echo "ERROR: settle2: ${settle2}" >&2; exit 1; }

usage="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command "SELECT concat_ws('|',upload_bytes::text,download_bytes::text,reserved_bytes::text) FROM pvnaive.direct_naive_accounting_terms WHERE service_term_id='cccccccc-cccc-cccc-cccc-cccccccccccc';")"
[[ "${usage}" == '100|0|0' ]] || { echo "ERROR: exact usage: ${usage}" >&2; exit 1; }
state="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command "SELECT state FROM pvnaive.service_terms WHERE id='cccccccc-cccc-cccc-cccc-cccccccccccc';")"
[[ "${state}" == quota_depleted ]] || { echo "ERROR: quota state: ${state}" >&2; exit 1; }

# Read model: two fresh sessions are online; a later stale read is offline/incomplete.
read_now="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command "SET ROLE pvnaive_app; SELECT concat_ws('|',upload_bytes::text,used_bytes::text,remaining_bytes::text,online::text,session_count::text,accounting_complete::text) FROM pvnaive.direct_naive_accounting_read('cccccccc-cccc-cccc-cccc-cccccccccccc','2026-08-29T18:00:06Z',90);")"
[[ "${read_now##*$'\n'}" == '100|100|0|true|2|false' || "${read_now##*$'\n'}" == '100|100|0|true|2|true' ]] || { echo "ERROR: fresh read: ${read_now}" >&2; exit 1; }
read_stale="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command "SET ROLE pvnaive_app; SELECT concat_ws('|',online::text,session_count::text,accounting_complete::text) FROM pvnaive.direct_naive_accounting_read('cccccccc-cccc-cccc-cccc-cccccccccccc','2026-08-29T18:05:00Z',90);")"
[[ "${read_stale##*$'\n'}" == 'false|0|false' ]] || { echo "ERROR: stale read: ${read_stale}" >&2; exit 1; }

# Append-only event ledger must reject mutation.
set +e
psql_admin --dbname "${test_db}" --command "UPDATE pvnaive.direct_naive_accounting_events SET upload_delta=999 WHERE source_sequence=2" >/dev/null 2>&1
mutation_rc=$?
set -e
(( mutation_rc != 0 )) || { echo 'ERROR: event ledger is mutable' >&2; exit 1; }

echo DIRECT_NAIVE_ACCOUNTING_PG18_TEST=PASSED
