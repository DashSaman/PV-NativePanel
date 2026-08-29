#!/usr/bin/env bash
set -Eeuo pipefail
umask 077

repo_root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd -P)"
test_suffix="${GITHUB_RUN_ID:-local}_${GITHUB_RUN_ATTEMPT:-1}_${BASHPID}"
test_suffix="${test_suffix//[^a-zA-Z0-9_]/_}"
test_db="pvnaive_ws1_accounting_${test_suffix,,}"
: "${PVNAIVE_DB_HOST:=127.0.0.1}"
: "${PVNAIVE_DB_PORT:=5432}"
: "${PVNAIVE_DB_USER:=postgres}"
export PVNAIVE_DB_HOST PVNAIVE_DB_PORT PVNAIVE_DB_USER

psql_admin() {
  psql --no-psqlrc --set ON_ERROR_STOP=1 --host "${PVNAIVE_DB_HOST}" --port "${PVNAIVE_DB_PORT}" --username "${PVNAIVE_DB_USER}" "$@"
}
cleanup() {
  psql_admin --dbname postgres --command "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname='${test_db}' AND pid <> pg_backend_pid()" >/dev/null 2>&1 || true
  dropdb --if-exists --host "${PVNAIVE_DB_HOST}" --port "${PVNAIVE_DB_PORT}" --username "${PVNAIVE_DB_USER}" "${test_db}" >/dev/null 2>&1 || true
  psql_admin --dbname postgres --command 'DROP ROLE IF EXISTS pvnaive_app; DROP ROLE IF EXISTS pvnaive_owner;' >/dev/null 2>&1 || true
}
trap cleanup EXIT
cleanup

psql_admin --dbname postgres <<'SQL' >/dev/null
CREATE ROLE pvnaive_owner NOLOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOREPLICATION NOBYPASSRLS;
CREATE ROLE pvnaive_app NOLOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOREPLICATION NOBYPASSRLS;
SQL
createdb --host "${PVNAIVE_DB_HOST}" --port "${PVNAIVE_DB_PORT}" --username "${PVNAIVE_DB_USER}" --owner pvnaive_owner --encoding UTF8 --template template0 "${test_db}"
PVNAIVE_DB_NAME="${test_db}" "${repo_root}/scripts/db/migrate.sh" >/dev/null

schema="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command 'SELECT COALESCE(MAX(version),0) FROM pvnaive.schema_migrations')"
[[ "${schema}" == 9 ]] || { echo "ERROR: expected schema 9, got ${schema}" >&2; exit 1; }

contract="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command "
SELECT
  to_regclass('pvnaive.direct_naive_accounting_events') IS NOT NULL || '|' ||
  to_regclass('pvnaive.direct_naive_accounting_sessions') IS NOT NULL || '|' ||
  to_regclass('pvnaive.direct_naive_accounting_terms') IS NOT NULL || '|' ||
  to_regclass('pvnaive.direct_naive_accounting_claims') IS NOT NULL || '|' ||
  to_regprocedure('pvnaive.direct_naive_accounting_authorize(uuid,timestamptz)') IS NOT NULL || '|' ||
  to_regprocedure('pvnaive.direct_naive_accounting_claim(uuid,text,uuid,uuid,bigint,text,bigint,timestamptz)') IS NOT NULL || '|' ||
  to_regprocedure('pvnaive.direct_naive_accounting_ingest(uuid,text,text,uuid,uuid,bigint,timestamptz,boolean,bigint,bigint,boolean)') IS NOT NULL;")"
[[ "${contract}" == "true|true|true|true|true|true|true" || "${contract}" == "t|t|t|t|t|t|t" ]] || {
  echo "ERROR: accounting schema contract failed: ${contract}" >&2
  exit 1
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
  '11111111-1111-1111-1111-111111111111', 'alice', repeat(E'\\001',32)::bytea,
  repeat(E'\\002',16)::bytea, repeat(E'\\003',12)::bytea, 'runtime-v1',
  'active', 'panel', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa'
);

INSERT INTO pvnaive.service_terms (
  id, tenant_id, user_id, quota_bytes, duration_seconds, start_policy,
  purchased_at, state
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

# Authorization alone must never start first-use.
auth="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command "SET ROLE pvnaive_app; SELECT tracked||'|'||allowed||'|'||COALESCE(remaining_bytes::text,'null') FROM pvnaive.direct_naive_accounting_authorize('11111111-1111-1111-1111-111111111111','2026-08-29T18:00:00Z');")"
[[ "${auth##*$'\n'}" == "t|t|100" || "${auth##*$'\n'}" == "true|true|100" ]] || { echo "ERROR: authorize failed: ${auth}" >&2; exit 1; }
first_before="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command "SELECT first_connected_at IS NULL FROM pvnaive.service_terms WHERE id='cccccccc-cccc-cccc-cccc-cccccccccccc';")"
[[ "${first_before}" == t || "${first_before}" == true ]] || { echo 'ERROR: authorize incorrectly activated first-use' >&2; exit 1; }

# seq=1 / cumulative zero is emitted only after authenticated CONNECT + target dial succeeds.
open1="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command "SET ROLE pvnaive_app; SELECT accepted||'|'||duplicate||'|'||reason FROM pvnaive.direct_naive_accounting_ingest('11111111-1111-1111-1111-111111111111','alice','direct-1','22222222-2222-2222-2222-222222222222','33333333-3333-3333-3333-333333333333',1,'2026-08-29T18:00:01Z',true,0,0,false);")"
[[ "${open1##*$'\n'}" == "t|f|accepted" || "${open1##*$'\n'}" == "true|false|accepted" ]] || { echo "ERROR: first CONNECT event failed: ${open1}" >&2; exit 1; }
first_after="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command "SELECT first_connected_at='2026-08-29T18:00:01Z'::timestamptz AND starts_at=first_connected_at AND expires_at=first_connected_at+interval '1 hour' FROM pvnaive.service_terms WHERE id='cccccccc-cccc-cccc-cccc-cccccccccccc';")"
[[ "${first_after}" == t || "${first_after}" == true ]] || { echo 'ERROR: first CONNECT did not activate term exactly once' >&2; exit 1; }

# Exact duplicate is idempotent; same-sequence conflict is detected.
dup="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command "SET ROLE pvnaive_app; SELECT accepted||'|'||duplicate||'|'||reason FROM pvnaive.direct_naive_accounting_ingest('11111111-1111-1111-1111-111111111111','alice','direct-1','22222222-2222-2222-2222-222222222222','33333333-3333-3333-3333-333333333333',1,'2026-08-29T18:00:01Z',true,0,0,false);")"
[[ "${dup##*$'\n'}" == "t|t|duplicate" || "${dup##*$'\n'}" == "true|true|duplicate" ]] || { echo "ERROR: duplicate not idempotent: ${dup}" >&2; exit 1; }
conflict="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command "SET ROLE pvnaive_app; SELECT accepted||'|'||reason FROM pvnaive.direct_naive_accounting_ingest('11111111-1111-1111-1111-111111111111','alice','direct-1','22222222-2222-2222-2222-222222222222','33333333-3333-3333-3333-333333333333',1,'2026-08-29T18:00:01Z',true,1,0,false);")"
[[ "${conflict##*$'\n'}" == "f|sequence_conflict" || "${conflict##*$'\n'}" == "false|sequence_conflict" ]] || { echo "ERROR: sequence conflict not detected: ${conflict}" >&2; exit 1; }

# A second concurrent session shares the same 100-byte term budget.
psql_admin --dbname "${test_db}" --tuples-only --no-align --command "SET ROLE pvnaive_app; SELECT accepted FROM pvnaive.direct_naive_accounting_ingest('11111111-1111-1111-1111-111111111111','alice','direct-1','22222222-2222-2222-2222-222222222222','44444444-4444-4444-4444-444444444444',1,'2026-08-29T18:00:02Z',true,0,0,false);" >/dev/null
claim1="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command "SET ROLE pvnaive_app; SELECT granted_bytes FROM pvnaive.direct_naive_accounting_claim('11111111-1111-1111-1111-111111111111','direct-1','22222222-2222-2222-2222-222222222222','33333333-3333-3333-3333-333333333333',2,'upload',80,'2026-08-29T18:00:03Z');")"
claim2="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command "SET ROLE pvnaive_app; SELECT granted_bytes FROM pvnaive.direct_naive_accounting_claim('11111111-1111-1111-1111-111111111111','direct-1','22222222-2222-2222-2222-222222222222','44444444-4444-4444-4444-444444444444',2,'upload',80,'2026-08-29T18:00:03Z');")"
[[ "${claim1##*$'\n'}" == 80 && "${claim2##*$'\n'}" == 20 ]] || { echo "ERROR: shared quota reservation failed: ${claim1}/${claim2}" >&2; exit 1; }

settle1="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command "SET ROLE pvnaive_app; SELECT accepted||'|'||upload_delta||'|'||remaining_bytes FROM pvnaive.direct_naive_accounting_ingest('11111111-1111-1111-1111-111111111111','alice','direct-1','22222222-2222-2222-2222-222222222222','33333333-3333-3333-3333-333333333333',2,'2026-08-29T18:00:04Z',true,80,0,false);")"
[[ "${settle1##*$'\n'}" == "t|80|0" || "${settle1##*$'\n'}" == "true|80|0" ]] || { echo "ERROR: settle1 failed: ${settle1}" >&2; exit 1; }
settle2="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command "SET ROLE pvnaive_app; SELECT accepted||'|'||upload_delta||'|'||quota_depleted FROM pvnaive.direct_naive_accounting_ingest('11111111-1111-1111-1111-111111111111','alice','direct-1','22222222-2222-2222-2222-222222222222','44444444-4444-4444-4444-444444444444',2,'2026-08-29T18:00:05Z',true,20,0,false);")"
[[ "${settle2##*$'\n'}" == "t|20|t" || "${settle2##*$'\n'}" == "true|20|true" ]] || { echo "ERROR: settle2 failed: ${settle2}" >&2; exit 1; }

usage="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command "SELECT upload_bytes||'|'||download_bytes||'|'||reserved_bytes FROM pvnaive.direct_naive_accounting_terms WHERE service_term_id='cccccccc-cccc-cccc-cccc-cccccccccccc';")"
[[ "${usage}" == "100|0|0" ]] || { echo "ERROR: exact usage mismatch: ${usage}" >&2; exit 1; }
state="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command "SELECT state FROM pvnaive.service_terms WHERE id='cccccccc-cccc-cccc-cccc-cccccccccccc';")"
[[ "${state}" == "quota_depleted" ]] || { echo "ERROR: quota state not depleted: ${state}" >&2; exit 1; }

# Immutable event ledger must reject mutation.
set +e
psql_admin --dbname "${test_db}" --command "UPDATE pvnaive.direct_naive_accounting_events SET upload_delta=999 WHERE source_sequence=2;" >/dev/null 2>&1
mutation_rc=$?
set -e
((mutation_rc != 0)) || { echo 'ERROR: accounting event ledger is mutable' >&2; exit 1; }

echo 'DIRECT_NAIVE_ACCOUNTING_MIGRATION_TEST=PASSED'
