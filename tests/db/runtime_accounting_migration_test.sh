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
test_db="pvnaive_migration_test_runtime_accounting_${suffix,,}"
app_password='pvnaive-accounting-ci-only'

psql_admin() {
  psql --no-psqlrc --set ON_ERROR_STOP=1 --host "${PVNAIVE_DB_HOST}" --port "${PVNAIVE_DB_PORT}" --username "${PVNAIVE_DB_USER}" "$@"
}

psql_app() {
  PGPASSWORD="${app_password}" psql --no-psqlrc --set ON_ERROR_STOP=1 \
    --host "${PVNAIVE_DB_HOST}" --port "${PVNAIVE_DB_PORT}" --username pvnaive_app --dbname "${test_db}" "$@"
}

cleanup() {
  psql_admin --dbname postgres --command "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname='${test_db}' AND pid <> pg_backend_pid()" >/dev/null 2>&1 || true
  dropdb --if-exists --host "${PVNAIVE_DB_HOST}" --port "${PVNAIVE_DB_PORT}" --username "${PVNAIVE_DB_USER}" "${test_db}" >/dev/null 2>&1 || true
  psql_admin --dbname postgres --command 'DROP ROLE IF EXISTS pvnaive_app; DROP ROLE IF EXISTS pvnaive_owner;' >/dev/null 2>&1 || true
}
trap cleanup EXIT HUP INT TERM
cleanup

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

for relation in runtime_accounting_terms runtime_accounting_sessions runtime_accounting_events runtime_accounting_reservations; do
  exists="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command "SELECT to_regclass('pvnaive.${relation}') IS NOT NULL")"
  [[ "${exists}" == "t" || "${exists}" == "true" ]] || { echo "ERROR: missing ${relation}" >&2; exit 1; }
done

for function_name in accounting_authorize accounting_reserve accounting_ingest_event accounting_read_model; do
  count="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command "SELECT COUNT(*) FROM pg_proc p JOIN pg_namespace n ON n.oid=p.pronamespace WHERE n.nspname='pvnaive' AND p.proname='${function_name}'")"
  [[ "${count}" -ge 1 ]] || { echo "ERROR: missing pvnaive.${function_name}" >&2; exit 1; }
done

# Runtime telemetry uses narrow SECURITY DEFINER functions. The application role
# must not have direct ledger DML/SELECT access without a management context.
set +e
psql_app --command 'SELECT COUNT(*) FROM pvnaive.runtime_accounting_events' >/dev/null 2>&1
ledger_select_rc=$?
set -e
[[ "${ledger_select_rc}" -ne 0 ]] || { echo 'ERROR: pvnaive_app can directly read accounting ledger' >&2; exit 1; }

psql_admin --dbname "${test_db}" >/dev/null <<'SQL'
INSERT INTO pvnaive.actors (id, tenant_id, actor_role, email, display_name, password_hash, status)
VALUES (
  'a9000000-0000-4000-8000-000000000001', NULL, 'owner',
  'accounting-owner@example.invalid', 'Accounting Owner',
  '$argon2id$v=19$m=19456,t=2,p=1$AAAAAAAAAAAAAAAAAAAAAA$AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA',
  'active'
);

ALTER TABLE pvnaive.tenants NO FORCE ROW LEVEL SECURITY;
ALTER TABLE pvnaive.users NO FORCE ROW LEVEL SECURITY;
ALTER TABLE pvnaive.naive_runtime_credentials NO FORCE ROW LEVEL SECURITY;
ALTER TABLE pvnaive.service_terms NO FORCE ROW LEVEL SECURITY;
ALTER TABLE pvnaive.user_runtime_credentials NO FORCE ROW LEVEL SECURITY;

INSERT INTO pvnaive.users (id,tenant_id,username,display_name,status,created_by_actor_id)
SELECT 'a9000000-0000-4000-8000-000000000010',id,'acct-user','Accounting User','active','a9000000-0000-4000-8000-000000000001'
FROM pvnaive.tenants WHERE tenant_type='system' AND slug='direct';

INSERT INTO pvnaive.naive_runtime_credentials (
  id,username,secret_hash,secret_ciphertext,secret_nonce,encryption_key_id,status,origin,created_by_actor_id,updated_by_actor_id
) VALUES (
  'a9000000-0000-4000-8000-000000000020','acct-runtime',decode(repeat('44',32),'hex'),decode(repeat('54',16),'hex'),decode(repeat('64',12),'hex'),
  'runtime-v1','active','panel','a9000000-0000-4000-8000-000000000001','a9000000-0000-4000-8000-000000000001'
);

INSERT INTO pvnaive.service_terms (
  id,tenant_id,user_id,quota_bytes,duration_seconds,start_policy,purchased_at,state
)
SELECT 'a9000000-0000-4000-8000-000000000030',tenant_id,id,100,3600,'on_first_successful_connection','2026-08-29T11:00:00Z','pending'
FROM pvnaive.users WHERE id='a9000000-0000-4000-8000-000000000010';

INSERT INTO pvnaive.user_runtime_credentials (tenant_id,user_id,service_term_id,runtime_credential_id,role)
SELECT tenant_id,id,'a9000000-0000-4000-8000-000000000030','a9000000-0000-4000-8000-000000000020','primary'
FROM pvnaive.users WHERE id='a9000000-0000-4000-8000-000000000010';

ALTER TABLE pvnaive.user_runtime_credentials FORCE ROW LEVEL SECURITY;
ALTER TABLE pvnaive.service_terms FORCE ROW LEVEL SECURITY;
ALTER TABLE pvnaive.naive_runtime_credentials FORCE ROW LEVEL SECURITY;
ALTER TABLE pvnaive.users FORCE ROW LEVEL SECURITY;
ALTER TABLE pvnaive.tenants FORCE ROW LEVEL SECURITY;
SQL

runtime_id='a9000000-0000-4000-8000-000000000020'
term1='a9000000-0000-4000-8000-000000000030'
boot1='a9000000-0000-4000-8000-000000000040'
session1='a9000000-0000-4000-8000-000000000050'
session2='a9000000-0000-4000-8000-000000000051'

# Pending first-use service is known/tracked, but no timer is started merely by
# authorize. Successful CONNECT is represented only by accepted sequence 1.
authorize_before="$(psql_app --tuples-only --no-align --command "SELECT tracked||'|'||allowed||'|'||service_term_id||'|'||COALESCE(first_connected_at::text,'') FROM pvnaive.accounting_authorize('${runtime_id}','2026-08-29T12:00:00Z',30)")"
[[ "${authorize_before}" == "t|t|${term1}|" || "${authorize_before}" == "true|true|${term1}|" ]] || { echo "ERROR: pre-connect authorize=${authorize_before}" >&2; exit 1; }

first_event="$(psql_app --tuples-only --no-align --command "SELECT duplicate||'|'||upload_delta||'|'||download_delta||'|'||service_term_id FROM pvnaive.accounting_ingest_event('${runtime_id}','acct-runtime','node-a','${boot1}','${session1}',1,'2026-08-29T12:00:00Z',0,0,false,NULL)")"
[[ "${first_event}" == "f|0|0|${term1}" || "${first_event}" == "false|0|0|${term1}" ]] || { echo "ERROR: first event=${first_event}" >&2; exit 1; }

first_use="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command "ALTER TABLE pvnaive.service_terms NO FORCE ROW LEVEL SECURITY; SELECT starts_at::text||'|'||first_connected_at::text||'|'||expires_at::text||'|'||state FROM pvnaive.service_terms WHERE id='${term1}'; ALTER TABLE pvnaive.service_terms FORCE ROW LEVEL SECURITY;" | grep '|' | tail -n1)"
[[ "${first_use}" == "2026-08-29 12:00:00+00|2026-08-29 12:00:00+00|2026-08-29 13:00:00+00|active" ]] || { echo "ERROR: first-use activation=${first_use}" >&2; exit 1; }

# Reconnect creates another session under the same shared ServiceTerm.
psql_app --command "SELECT * FROM pvnaive.accounting_ingest_event('${runtime_id}','acct-runtime','node-a','${boot1}','${session2}',1,'2026-08-29T12:00:01Z',0,0,false,NULL)" >/dev/null

reserve1="$(psql_app --tuples-only --no-align --command "SELECT reservation_id||'|'||service_term_id||'|'||granted_bytes FROM pvnaive.accounting_reserve('${runtime_id}','node-a','${boot1}','${session1}','upload',60,'2026-08-29T12:00:02Z')")"
reservation1="${reserve1%%|*}"
[[ "${reserve1}" == "${reservation1}|${term1}|60" && "${reservation1}" =~ ^[0-9a-f-]{36}$ ]] || { echo "ERROR: reserve1=${reserve1}" >&2; exit 1; }

reserve2="$(psql_app --tuples-only --no-align --command "SELECT reservation_id||'|'||service_term_id||'|'||granted_bytes FROM pvnaive.accounting_reserve('${runtime_id}','node-a','${boot1}','${session2}','upload',60,'2026-08-29T12:00:03Z')")"
reservation2="${reserve2%%|*}"
[[ "${reserve2}" == "${reservation2}|${term1}|40" && "${reservation2}" =~ ^[0-9a-f-]{36}$ ]] || { echo "ERROR: reserve2=${reserve2}, want shared grant=40" >&2; exit 1; }

commit1="$(psql_app --tuples-only --no-align --command "SELECT duplicate||'|'||upload_delta||'|'||download_delta FROM pvnaive.accounting_ingest_event('${runtime_id}','acct-runtime','node-a','${boot1}','${session1}',2,'2026-08-29T12:00:04Z',60,0,false,'${reservation1}')")"
[[ "${commit1}" == "f|60|0" || "${commit1}" == "false|60|0" ]] || { echo "ERROR: commit1=${commit1}" >&2; exit 1; }

# Exact duplicate is replay-safe and cannot double-count.
duplicate1="$(psql_app --tuples-only --no-align --command "SELECT duplicate||'|'||upload_delta||'|'||download_delta FROM pvnaive.accounting_ingest_event('${runtime_id}','acct-runtime','node-a','${boot1}','${session1}',2,'2026-08-29T12:00:04Z',60,0,false,'${reservation1}')")"
[[ "${duplicate1}" == "t|0|0" || "${duplicate1}" == "true|0|0" ]] || { echo "ERROR: duplicate=${duplicate1}" >&2; exit 1; }

# Same sequence with a changed cumulative counter must fail closed.
set +e
psql_app --command "SELECT * FROM pvnaive.accounting_ingest_event('${runtime_id}','acct-runtime','node-a','${boot1}','${session1}',2,'2026-08-29T12:00:04Z',61,0,false,'${reservation1}')" >/dev/null 2>&1
sequence_conflict_rc=$?
set -e
[[ "${sequence_conflict_rc}" -ne 0 ]] || { echo 'ERROR: same-sequence conflict was accepted' >&2; exit 1; }

psql_app --command "SELECT * FROM pvnaive.accounting_ingest_event('${runtime_id}','acct-runtime','node-a','${boot1}','${session2}',2,'2026-08-29T12:00:05Z',40,0,false,'${reservation2}')" >/dev/null

read1="$(psql_app --tuples-only --no-align --command "SELECT upload_bytes||'|'||download_bytes||'|'||used_bytes||'|'||quota_bytes||'|'||remaining_bytes||'|'||quota_state FROM pvnaive.accounting_read_model('${runtime_id}','2026-08-29T12:00:06Z',30)")"
[[ "${read1}" == '100|0|100|100|0|depleted' ]] || { echo "ERROR: read1=${read1}" >&2; exit 1; }

# Renewal/new ServiceTerm must not inherit usage from term1.
psql_admin --dbname "${test_db}" >/dev/null <<SQL
ALTER TABLE pvnaive.service_terms NO FORCE ROW LEVEL SECURITY;
ALTER TABLE pvnaive.user_runtime_credentials NO FORCE ROW LEVEL SECURITY;
UPDATE pvnaive.user_runtime_credentials SET unbound_at='2026-08-29T12:10:00Z' WHERE service_term_id='${term1}';
INSERT INTO pvnaive.service_terms (id,tenant_id,user_id,quota_bytes,duration_seconds,start_policy,purchased_at,state)
SELECT 'a9000000-0000-4000-8000-000000000031',tenant_id,id,200,3600,'on_first_successful_connection','2026-08-29T12:10:00Z','pending'
FROM pvnaive.users WHERE id='a9000000-0000-4000-8000-000000000010';
INSERT INTO pvnaive.user_runtime_credentials (tenant_id,user_id,service_term_id,runtime_credential_id,role,bound_at)
SELECT tenant_id,id,'a9000000-0000-4000-8000-000000000031','${runtime_id}','primary','2026-08-29T12:10:00Z'
FROM pvnaive.users WHERE id='a9000000-0000-4000-8000-000000000010';
ALTER TABLE pvnaive.user_runtime_credentials FORCE ROW LEVEL SECURITY;
ALTER TABLE pvnaive.service_terms FORCE ROW LEVEL SECURITY;
SQL

term2_read="$(psql_app --tuples-only --no-align --command "SELECT service_term_id||'|'||used_bytes||'|'||quota_bytes||'|'||remaining_bytes FROM pvnaive.accounting_read_model('${runtime_id}','2026-08-29T12:10:01Z',30)")"
[[ "${term2_read}" == 'a9000000-0000-4000-8000-000000000031|0|200|200' ]] || { echo "ERROR: renewal carry-over detected: ${term2_read}" >&2; exit 1; }

# New boot while an old-boot session has no trusted final counter is not allowed
# to silently claim complete accounting. No bytes are estimated.
boot2='a9000000-0000-4000-8000-000000000041'
session3='a9000000-0000-4000-8000-000000000052'
psql_app --command "SELECT * FROM pvnaive.accounting_ingest_event('${runtime_id}','acct-runtime','node-a','${boot2}','${session3}',1,'2026-08-29T12:10:02Z',0,0,false,NULL)" >/dev/null
incomplete="$(psql_app --tuples-only --no-align --command "SELECT accounting_complete FROM pvnaive.accounting_read_model('${runtime_id}','2026-08-29T12:10:03Z',30)")"
[[ "${incomplete}" == 'f' || "${incomplete}" == 'false' ]] || { echo "ERROR: restart without old final counter claimed complete accounting" >&2; exit 1; }

# Sequence gap and counter regression are both rejected on the new session.
set +e
psql_app --command "SELECT * FROM pvnaive.accounting_ingest_event('${runtime_id}','acct-runtime','node-a','${boot2}','${session3}',3,'2026-08-29T12:10:04Z',1,0,false,NULL)" >/dev/null 2>&1
gap_rc=$?
psql_app --command "SELECT * FROM pvnaive.accounting_ingest_event('${runtime_id}','acct-runtime','node-a','${boot2}','${session3}',2,'2026-08-29T12:10:04Z',-1,0,false,NULL)" >/dev/null 2>&1
regression_rc=$?
set -e
[[ "${gap_rc}" -ne 0 ]] || { echo 'ERROR: sequence gap was accepted' >&2; exit 1; }
[[ "${regression_rc}" -ne 0 ]] || { echo 'ERROR: negative/counter regression event was accepted' >&2; exit 1; }

# Stale unclosed session is offline and remains incomplete; the system never
# invents a final counter to make it look healthy.
stale="$(psql_app --tuples-only --no-align --command "SELECT online||'|'||session_count||'|'||accounting_complete FROM pvnaive.accounting_read_model('${runtime_id}','2026-08-29T12:20:00Z',30)")"
[[ "${stale}" == 'f|0|f' || "${stale}" == 'false|0|false' ]] || { echo "ERROR: stale read=${stale}" >&2; exit 1; }

# Append-only means even a privileged connection cannot rewrite history.
set +e
psql_admin --dbname "${test_db}" --command "UPDATE pvnaive.runtime_accounting_events SET upload_bytes=999 WHERE runtime_credential_id='${runtime_id}'" >/dev/null 2>&1
append_only_rc=$?
set -e
[[ "${append_only_rc}" -ne 0 ]] || { echo 'ERROR: accounting event history was mutable' >&2; exit 1; }

echo 'PVNAIVE_RUNTIME_ACCOUNTING_MIGRATION_TEST=PASSED'
