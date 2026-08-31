#!/usr/bin/env bash
set -Eeuo pipefail
umask 077
repo_root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd -P)"
up="$repo_root/db/migrations/0017_operator_session_peers.up.sql"
down="$repo_root/db/migrations/0017_operator_session_peers.down.sql"
[[ -f "$up" ]] || { echo 'ERROR: missing schema17 operator session peer migration' >&2; exit 1; }
[[ -f "$down" ]] || { echo 'ERROR: missing schema17 operator session peer rollback' >&2; exit 1; }
grep -Fqx -- '-- pvnaive:migration-version 0017' "$up"
grep -Fqx -- '-- pvnaive:transactional true' "$up"
grep -Fqx -- '-- pvnaive:destructive false' "$up"
grep -Fqx -- '-- pvnaive:migration-version 0017' "$down"
grep -Fqx -- '-- pvnaive:destructive true' "$down"
: "${PVNAIVE_DB_HOST:=127.0.0.1}"
: "${PVNAIVE_DB_PORT:=5432}"
: "${PVNAIVE_DB_USER:=postgres}"
export PVNAIVE_DB_HOST PVNAIVE_DB_PORT PVNAIVE_DB_USER
suffix="${GITHUB_RUN_ID:-local}_${GITHUB_RUN_ATTEMPT:-1}_${BASHPID}"; suffix="${suffix//[^a-zA-Z0-9_]/_}"
test_db="pvnaive_migration_test_session_peers_${suffix,,}"
app_password="pvnaive-session-peer-ci"
tmp="$(mktemp -d)"
psql_admin(){ psql --no-psqlrc --set ON_ERROR_STOP=1 -h "$PVNAIVE_DB_HOST" -p "$PVNAIVE_DB_PORT" -U "$PVNAIVE_DB_USER" "$@"; }
psql_app(){ PGPASSWORD="$app_password" psql --no-psqlrc --set ON_ERROR_STOP=1 -h "$PVNAIVE_DB_HOST" -p "$PVNAIVE_DB_PORT" -U pvnaive_app -d "$test_db" "$@"; }
cleanup(){ rm -rf "$tmp" 2>/dev/null || true; psql_admin -d postgres -c "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname='${test_db}' AND pid<>pg_backend_pid()" >/dev/null 2>&1 || true; dropdb --if-exists -h "$PVNAIVE_DB_HOST" -p "$PVNAIVE_DB_PORT" -U "$PVNAIVE_DB_USER" "$test_db" >/dev/null 2>&1 || true; psql_admin -d postgres -c 'DROP ROLE IF EXISTS pvnaive_app; DROP ROLE IF EXISTS pvnaive_owner;' >/dev/null 2>&1 || true; }
trap cleanup EXIT HUP INT TERM
cleanup
psql_admin -d postgres <<SQL >/dev/null
CREATE ROLE pvnaive_owner NOLOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOREPLICATION NOBYPASSRLS;
CREATE ROLE pvnaive_app LOGIN PASSWORD '${app_password}' NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOREPLICATION NOBYPASSRLS;
SQL
createdb -h "$PVNAIVE_DB_HOST" -p "$PVNAIVE_DB_PORT" -U "$PVNAIVE_DB_USER" --owner pvnaive_owner --encoding UTF8 --template template0 "$test_db"
export PVNAIVE_DB_NAME="$test_db"
mkdir -p "$tmp/migrations"
for version in $(seq 1 16); do prefix="$(printf '%04d' "$version")"; cp "$repo_root/db/migrations/${prefix}_"*.sql "$tmp/migrations/"; done
( cd "$tmp/migrations" && sha256sum *.sql > SHA256SUMS )
PVNAIVE_MIGRATIONS_DIR="$tmp/migrations" "$repo_root/scripts/db/migrate.sh" >/dev/null
[[ "$(psql_admin -d "$test_db" -Atc 'select max(version) from pvnaive.schema_migrations')" == 16 ]]
# Fixture: two tenant-bound reseller contexts; one trusted session belongs only to tenant A.
psql_admin -d "$test_db" <<'SQL' >/dev/null
INSERT INTO pvnaive.tenants(id,tenant_type,slug,display_name,status)
VALUES('17170000-0000-0000-0000-000000000002','reseller','task12-other','Task12 Other','active');
INSERT INTO pvnaive.actors(id,tenant_id,actor_role,email,display_name,status)
SELECT '17170000-0000-0000-0000-000000000011',id,'reseller','task12-a@example.invalid','Task12 A','active' FROM pvnaive.tenants WHERE tenant_type='system' AND slug='direct';
INSERT INTO pvnaive.actors(id,tenant_id,actor_role,email,display_name,status)
VALUES('17170000-0000-0000-0000-000000000012','17170000-0000-0000-0000-000000000002','reseller','task12-b@example.invalid','Task12 B','active');
INSERT INTO pvnaive.auth_sessions(id,tenant_id,actor_id,token_hash,refresh_family_id,user_agent_hash,expires_at,absolute_expires_at,csrf_token_hash)
SELECT '17170000-0000-0000-0000-000000000021',tenant_id,id,decode(repeat('71',32),'hex'),'17170000-0000-0000-0000-000000000031',decode(repeat('81',32),'hex'),clock_timestamp()+interval '1 hour',clock_timestamp()+interval '12 hours',decode(repeat('91',32),'hex') FROM pvnaive.actors WHERE id='17170000-0000-0000-0000-000000000011';
INSERT INTO pvnaive.auth_sessions(id,tenant_id,actor_id,token_hash,refresh_family_id,user_agent_hash,expires_at,absolute_expires_at,csrf_token_hash)
SELECT '17170000-0000-0000-0000-000000000022',tenant_id,id,decode(repeat('72',32),'hex'),'17170000-0000-0000-0000-000000000032',decode(repeat('82',32),'hex'),clock_timestamp()+interval '1 hour',clock_timestamp()+interval '12 hours',decode(repeat('92',32),'hex') FROM pvnaive.actors WHERE id='17170000-0000-0000-0000-000000000012';
INSERT INTO pvnaive.users(id,tenant_id,username,display_name,status,created_by_actor_id)
SELECT '17170000-0000-0000-0000-000000000041',id,'task12-user','Task12 User','active','17170000-0000-0000-0000-000000000011' FROM pvnaive.tenants WHERE tenant_type='system' AND slug='direct';
INSERT INTO pvnaive.naive_runtime_credentials(id,username,secret_hash,secret_ciphertext,secret_nonce,encryption_key_id,status,origin,created_by_actor_id,updated_by_actor_id)
VALUES('17170000-0000-0000-0000-000000000051','task12-runtime',decode(repeat('41',32),'hex'),decode(repeat('51',16),'hex'),decode(repeat('61',12),'hex'),'runtime-v1','active','panel','17170000-0000-0000-0000-000000000011','17170000-0000-0000-0000-000000000011');
BEGIN;
SELECT pvnaive.direct_naive_accounting_enter_context();
INSERT INTO pvnaive.service_terms(id,tenant_id,user_id,quota_bytes,duration_seconds,start_policy,purchased_at,starts_at,expires_at,state,accounting_baseline_state,accounting_baseline_source,accounting_baseline_cutoff_at,accounting_baseline_upload_bytes,accounting_baseline_download_bytes)
SELECT '17170000-0000-0000-0000-000000000061',tenant_id,id,1000000,3600,'on_creation','2026-08-31 00:00:00+00','2026-08-31 00:00:00+00','2026-08-31 01:00:00+00','active','known','fresh_managed_term','2026-08-31 00:00:00+00',0,0 FROM pvnaive.users WHERE id='17170000-0000-0000-0000-000000000041';
SELECT pvnaive.direct_naive_accounting_leave_context();
COMMIT;
INSERT INTO pvnaive.user_runtime_credentials(id,tenant_id,user_id,service_term_id,runtime_credential_id,role,bound_at)
SELECT '17170000-0000-0000-0000-000000000071',tenant_id,id,'17170000-0000-0000-0000-000000000061','17170000-0000-0000-0000-000000000051','primary','2026-08-31 00:00:00+00' FROM pvnaive.users WHERE id='17170000-0000-0000-0000-000000000041';
INSERT INTO pvnaive.direct_naive_accounting_sessions(runtime_credential_id,node_id,boot_id,session_id,service_term_id,first_observed_at,last_observed_at,last_sequence,upload_cumulative,download_cumulative,final,accounting_complete)
VALUES('17170000-0000-0000-0000-000000000051','node-a','17170000-0000-0000-0000-000000000081','17170000-0000-0000-0000-000000000091','17170000-0000-0000-0000-000000000061','2026-08-31 00:10:00+00','2026-08-31 00:10:30+00',3,100,200,false,true);
SQL
cp "$repo_root/db/migrations/0017_"*.sql "$tmp/migrations/"
( cd "$tmp/migrations" && sha256sum *.sql > SHA256SUMS )
PVNAIVE_MIGRATIONS_DIR="$tmp/migrations" "$repo_root/scripts/db/migrate.sh" >/dev/null
[[ "$(psql_admin -d "$test_db" -Atc 'select max(version) from pvnaive.schema_migrations')" == 17 ]]
flags="$(psql_admin -d "$test_db" -Atc "select relrowsecurity::text||'|'||relforcerowsecurity::text from pg_class where oid='pvnaive.direct_naive_accounting_session_peers'::regclass")"
[[ "$flags" == 'true|true' || "$flags" == 't|t' ]]
# No request context: operator read is fail-closed.
set +e
psql_app -Atc "select count(*) from pvnaive.list_active_customer_sessions('17170000-0000-0000-0000-000000000041','2026-08-31 00:11:00+00',30)" >/dev/null 2>&1
no_context_rc=$?
set -e
[[ "$no_context_rc" -ne 0 ]]
# Peer registration requires the pre-existing trusted session.
recorded="$(psql_app -Atc "select service_term_id||'|'||recorded||'|'||duplicate from pvnaive.direct_naive_accounting_record_session_peer('17170000-0000-0000-0000-000000000051','node-a','17170000-0000-0000-0000-000000000081','17170000-0000-0000-0000-000000000091','203.0.113.7','2026-08-31 00:10:01+00')")"
[[ "$recorded" == '17170000-0000-0000-0000-000000000061|t|f' || "$recorded" == '17170000-0000-0000-0000-000000000061|true|false' ]]
replay="$(psql_app -Atc "select recorded||'|'||duplicate from pvnaive.direct_naive_accounting_record_session_peer('17170000-0000-0000-0000-000000000051','node-a','17170000-0000-0000-0000-000000000081','17170000-0000-0000-0000-000000000091','203.0.113.7','2026-08-31 00:10:02+00')")"
[[ "$replay" == 'f|t' || "$replay" == 'false|true' ]]
set +e
psql_app -Atc "select * from pvnaive.direct_naive_accounting_record_session_peer('17170000-0000-0000-0000-000000000051','node-a','17170000-0000-0000-0000-000000000081','17170000-0000-0000-0000-000000000091','203.0.113.8','2026-08-31 00:10:03+00')" >/dev/null 2>&1
conflict_rc=$?
set -e
[[ "$conflict_rc" -ne 0 ]]
# Tenant A sees the active session with exact peer/bytes/duration.
visible="$(psql_app -At <<'SQL'
BEGIN;
SELECT pvnaive.set_request_context(decode(repeat('71',32),'hex'));
SELECT client_ip||'|'||node_id||'|'||upload_bytes||'|'||download_bytes||'|'||duration_seconds FROM pvnaive.list_active_customer_sessions('17170000-0000-0000-0000-000000000041','2026-08-31 00:11:00+00',30);
ROLLBACK;
SQL
)"
visible="$(printf '%s\n' "$visible" | grep '^203\.0\.113\.7|' | tail -n1)"
[[ "$visible" == '203.0.113.7|node-a|100|200|60' ]]
# Tenant B cannot use the same user UUID to cross scope.
hidden="$(psql_app -At <<'SQL'
BEGIN;
SELECT pvnaive.set_request_context(decode(repeat('72',32),'hex'));
SELECT count(*) FROM pvnaive.list_active_customer_sessions('17170000-0000-0000-0000-000000000041','2026-08-31 00:11:00+00',30);
ROLLBACK;
SQL
)"
hidden="$(printf '%s\n' "$hidden" | grep -E '^[0-9]+$' | tail -n1)"
[[ "$hidden" == 0 ]]
# stale/final/incomplete are not active.
for mutation in "last_observed_at='2026-08-31 00:10:00+00'" "last_observed_at='2026-08-31 00:10:30+00',final=true" "final=false,accounting_complete=false"; do
  psql_admin -d "$test_db" -c "update pvnaive.direct_naive_accounting_sessions set $mutation where session_id='17170000-0000-0000-0000-000000000091'" >/dev/null
  count="$(psql_app -At <<'SQL'
BEGIN;
SELECT pvnaive.set_request_context(decode(repeat('71',32),'hex'));
SELECT count(*) FROM pvnaive.list_active_customer_sessions('17170000-0000-0000-0000-000000000041','2026-08-31 00:11:00+00',30);
ROLLBACK;
SQL
)"
  count="$(printf '%s\n' "$count" | grep -E '^[0-9]+$' | tail -n1)"; [[ "$count" == 0 ]]
done
# With peer evidence present rollback must refuse. After fixture peer removal disposable rollback is clean.
set +e
PVNAIVE_DISPOSABLE_DB=1 PVNAIVE_MIGRATIONS_DIR="$tmp/migrations" "$repo_root/scripts/db/rollback.sh" >/dev/null 2>&1
blocked_rc=$?
set -e
[[ "$blocked_rc" -ne 0 ]]
psql_admin -d "$test_db" -c "delete from pvnaive.direct_naive_accounting_session_peers" >/dev/null
PVNAIVE_DISPOSABLE_DB=1 PVNAIVE_MIGRATIONS_DIR="$tmp/migrations" "$repo_root/scripts/db/rollback.sh" >/dev/null
[[ "$(psql_admin -d "$test_db" -Atc 'select max(version) from pvnaive.schema_migrations')" == 16 ]]
[[ "$(psql_admin -d "$test_db" -Atc "select to_regclass('pvnaive.direct_naive_accounting_session_peers') is null")" =~ ^(t|true)$ ]]
echo 'PVNAIVE_OPERATOR_SESSION_PEERS_MIGRATION_TEST=PASSED'
