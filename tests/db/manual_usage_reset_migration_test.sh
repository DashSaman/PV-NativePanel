#!/usr/bin/env bash
set -Eeuo pipefail
umask 077
repo_root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd -P)"
up="${repo_root}/db/migrations/0014_manual_usage_reset.up.sql"
down="${repo_root}/db/migrations/0014_manual_usage_reset.down.sql"
[[ -f "$up" ]] || { echo 'ERROR: missing schema14 manual usage reset migration' >&2; exit 1; }
[[ -f "$down" ]] || { echo 'ERROR: missing schema14 manual usage reset rollback' >&2; exit 1; }
grep -Fqx -- '-- pvnaive:migration-version 0014' "$up"
grep -Fqx -- '-- pvnaive:transactional true' "$up"
grep -Fqx -- '-- pvnaive:destructive false' "$up"
grep -Fqx -- '-- pvnaive:migration-version 0014' "$down"
grep -Fqx -- '-- pvnaive:transactional true' "$down"
grep -Fqx -- '-- pvnaive:destructive true' "$down"

: "${PVNAIVE_DB_HOST:=127.0.0.1}"
: "${PVNAIVE_DB_PORT:=5432}"
: "${PVNAIVE_DB_USER:=postgres}"
export PVNAIVE_DB_HOST PVNAIVE_DB_PORT PVNAIVE_DB_USER
suffix="${GITHUB_RUN_ID:-local}_${GITHUB_RUN_ATTEMPT:-1}_${BASHPID}"
suffix="${suffix//[^a-zA-Z0-9_]/_}"
test_db="pvnaive_migration_test_manual_reset_${suffix,,}"
app_password='pvnaive-reset-ci-only'
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
psql_admin -d postgres <<SQL >/dev/null
CREATE ROLE pvnaive_owner NOLOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOREPLICATION NOBYPASSRLS;
CREATE ROLE pvnaive_app LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOREPLICATION NOBYPASSRLS PASSWORD '${app_password}';
ALTER ROLE pvnaive_app SET row_security=on;
SQL
createdb -h "$PVNAIVE_DB_HOST" -p "$PVNAIVE_DB_PORT" -U "$PVNAIVE_DB_USER" --owner pvnaive_owner --encoding UTF8 --template template0 "$test_db"
export PVNAIVE_DB_NAME="$test_db"
mkdir -p "$fixture/migrations"
for version in $(seq 1 13); do prefix="$(printf '%04d' "$version")"; cp "${repo_root}/db/migrations/${prefix}_"*.sql "$fixture/migrations/"; done
( cd "$fixture/migrations" && sha256sum *.sql > SHA256SUMS )
PVNAIVE_MIGRATIONS_DIR="$fixture/migrations" "${repo_root}/scripts/db/migrate.sh" >/dev/null
[[ "$(psql_admin -d "$test_db" -Atc 'SELECT max(version) FROM pvnaive.schema_migrations')" == 13 ]]

# Management identity + a fresh service term. The reset primitive is exercised
# through pvnaive_app with a signed request context, not by a superuser shortcut.
psql_admin -d "$test_db" <<'SQL' >/dev/null
INSERT INTO pvnaive.actors(id,tenant_id,actor_role,email,display_name,password_hash,status)
VALUES('a7140000-0000-0000-0000-000000000001',NULL,'owner','reset-owner@example.invalid','Reset Owner',
'$argon2id$v=19$m=19456,t=2,p=1$AAAAAAAAAAAAAAAAAAAAAA$AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA','active');
INSERT INTO pvnaive.auth_sessions(id,tenant_id,actor_id,token_hash,refresh_family_id,user_agent_hash,expires_at,absolute_expires_at,csrf_token_hash)
VALUES('b7140000-0000-0000-0000-000000000001',NULL,'a7140000-0000-0000-0000-000000000001',decode(repeat('71',32),'hex'),
'c7140000-0000-0000-0000-000000000001',decode(repeat('72',32),'hex'),clock_timestamp()+interval '1 hour',clock_timestamp()+interval '12 hours',decode(repeat('73',32),'hex'));
INSERT INTO pvnaive.users(id,tenant_id,username,display_name,status,created_by_actor_id)
SELECT 'd7140000-0000-0000-0000-000000000001',id,'reset-user','Reset User','active','a7140000-0000-0000-0000-000000000001'
FROM pvnaive.tenants WHERE tenant_type='system' AND slug='direct';
INSERT INTO pvnaive.naive_runtime_credentials(id,username,secret_hash,secret_ciphertext,secret_nonce,encryption_key_id,status,origin,created_by_actor_id,updated_by_actor_id)
VALUES('e7140000-0000-0000-0000-000000000001','reset-user',decode(repeat('74',32),'hex'),decode(repeat('75',32),'hex'),decode(repeat('76',12),'hex'),'runtime-v1','active','panel','a7140000-0000-0000-0000-000000000001','a7140000-0000-0000-0000-000000000001');
INSERT INTO pvnaive.service_terms(id,tenant_id,user_id,quota_bytes,duration_seconds,start_policy,purchased_at,starts_at,expires_at,state,
 accounting_baseline_state,accounting_baseline_source,accounting_baseline_cutoff_at,accounting_baseline_upload_bytes,accounting_baseline_download_bytes)
SELECT 'f7140000-0000-0000-0000-000000000001',tenant_id,id,1000,2592000,'on_creation','2026-08-30T10:00:00Z','2026-08-30T10:00:00Z','2026-09-29T10:00:00Z','active',
'known','fresh_managed_term','2026-08-30T10:00:00Z',0,0
FROM pvnaive.users WHERE id='d7140000-0000-0000-0000-000000000001';
INSERT INTO pvnaive.user_runtime_credentials(id,tenant_id,user_id,service_term_id,runtime_credential_id,role,bound_at)
SELECT '17140000-0000-0000-0000-000000000001',tenant_id,id,'f7140000-0000-0000-0000-000000000001','e7140000-0000-0000-0000-000000000001','primary','2026-08-30T10:00:00Z'
FROM pvnaive.users WHERE id='d7140000-0000-0000-0000-000000000001';
INSERT INTO pvnaive.direct_subscription_tokens(tenant_id,user_id,service_term_id,runtime_credential_id,token_hash,token_prefix,status,user_state,service_state,runtime_username,secret_ciphertext,secret_nonce,encryption_key_id,
 quota_bytes,duration_seconds,start_policy,starts_at,first_connected_at,accounting_baseline_state,accounting_baseline_source,accounting_baseline_cutoff_at,accounting_baseline_upload_bytes,accounting_baseline_download_bytes)
SELECT u.tenant_id,u.id,st.id,rc.id,decode(repeat('77',32),'hex'),'resetTok','active',u.status,st.state,rc.username,rc.secret_ciphertext,rc.secret_nonce,rc.encryption_key_id,
 st.quota_bytes,st.duration_seconds,st.start_policy,st.starts_at,st.first_connected_at,st.accounting_baseline_state,st.accounting_baseline_source,st.accounting_baseline_cutoff_at,st.accounting_baseline_upload_bytes,st.accounting_baseline_download_bytes
FROM pvnaive.users u JOIN pvnaive.service_terms st ON st.user_id=u.id JOIN pvnaive.naive_runtime_credentials rc ON rc.id='e7140000-0000-0000-0000-000000000001'
WHERE u.id='d7140000-0000-0000-0000-000000000001';
SQL

# Build 300 exact bytes with a live session through the real accounting functions.
psql_admin -d "$test_db" <<'SQL' >/dev/null
SELECT * FROM pvnaive.direct_naive_accounting_ingest('e7140000-0000-0000-0000-000000000001','reset-user','node-reset','27140000-0000-0000-0000-000000000001','37140000-0000-0000-0000-000000000001',1,'2026-08-30T11:58:00Z',true,0,0,false);
SELECT * FROM pvnaive.direct_naive_accounting_claim('e7140000-0000-0000-0000-000000000001','node-reset','27140000-0000-0000-0000-000000000001','37140000-0000-0000-0000-000000000001',2,'upload',100,'2026-08-30T11:58:10Z');
SELECT * FROM pvnaive.direct_naive_accounting_ingest('e7140000-0000-0000-0000-000000000001','reset-user','node-reset','27140000-0000-0000-0000-000000000001','37140000-0000-0000-0000-000000000001',2,'2026-08-30T11:58:20Z',true,100,0,false);
SELECT * FROM pvnaive.direct_naive_accounting_claim('e7140000-0000-0000-0000-000000000001','node-reset','27140000-0000-0000-0000-000000000001','37140000-0000-0000-0000-000000000001',3,'download',200,'2026-08-30T11:58:30Z');
SELECT * FROM pvnaive.direct_naive_accounting_ingest('e7140000-0000-0000-0000-000000000001','reset-user','node-reset','27140000-0000-0000-0000-000000000001','37140000-0000-0000-0000-000000000001',3,'2026-08-30T11:58:40Z',true,100,200,false);
UPDATE pvnaive.service_terms SET state='quota_depleted' WHERE id='f7140000-0000-0000-0000-000000000001';
SQL
pre_identity="$(psql_admin -d "$test_db" -Atc "SELECT encode(rc.secret_hash,'hex')||'|'||encode(dst.token_hash,'hex') FROM pvnaive.naive_runtime_credentials rc JOIN pvnaive.direct_subscription_tokens dst ON dst.runtime_credential_id=rc.id WHERE rc.id='e7140000-0000-0000-0000-000000000001'")"
pre_ledger="$(psql_admin -d "$test_db" -Atc "SELECT count(*) FROM pvnaive.direct_naive_accounting_events WHERE service_term_id='f7140000-0000-0000-0000-000000000001'")"
[[ "$pre_ledger" == 3 ]]

cp "${repo_root}/db/migrations/0014_"*.sql "$fixture/migrations/"
( cd "$fixture/migrations" && sha256sum *.sql > SHA256SUMS )
PVNAIVE_MIGRATIONS_DIR="$fixture/migrations" "${repo_root}/scripts/db/migrate.sh" >/dev/null
[[ "$(psql_admin -d "$test_db" -Atc 'SELECT max(version) FROM pvnaive.schema_migrations')" == 14 ]]

reset="$(PGPASSWORD="$app_password" psql --no-psqlrc --set ON_ERROR_STOP=1 -h "$PVNAIVE_DB_HOST" -p "$PVNAIVE_DB_PORT" -U pvnaive_app -d "$test_db" -At -F'|' <<'SQL'
BEGIN;
SELECT pvnaive.set_request_context(decode(repeat('71',32),'hex'));
SELECT resettable,reason,previous_upload_bytes,previous_download_bytes,previous_used_bytes,service_state
FROM pvnaive.direct_naive_accounting_reset('f7140000-0000-0000-0000-000000000001','2026-08-30T12:00:00Z',90);
COMMIT;
SQL
)"
reset="$(printf '%s\n' "$reset" | grep -E '^(t|true)\|' | tail -n1)"
[[ "$reset" == 't|reset|100|200|300|active' || "$reset" == 'true|reset|100|200|300|active' ]] || { echo "ERROR: reset result=$reset" >&2; exit 1; }

projection="$(psql_admin -d "$test_db" -At -F'|' -c "SELECT upload_bytes,download_bytes,reserved_bytes,to_char(last_reset_at AT TIME ZONE 'UTC','YYYY-MM-DD HH24:MI:SS') FROM pvnaive.direct_naive_accounting_terms WHERE service_term_id='f7140000-0000-0000-0000-000000000001'")"
[[ "$projection" == '0|0|0|2026-08-30 12:00:00' ]] || { echo "ERROR: reset projection=$projection" >&2; exit 1; }
[[ "$(psql_admin -d "$test_db" -Atc "SELECT state FROM pvnaive.service_terms WHERE id='f7140000-0000-0000-0000-000000000001'")" == active ]]
[[ "$(psql_admin -d "$test_db" -Atc "SELECT count(*) FROM pvnaive.direct_naive_accounting_events WHERE service_term_id='f7140000-0000-0000-0000-000000000001'")" == "$pre_ledger" ]]
[[ "$(psql_admin -d "$test_db" -Atc "SELECT encode(rc.secret_hash,'hex')||'|'||encode(dst.token_hash,'hex') FROM pvnaive.naive_runtime_credentials rc JOIN pvnaive.direct_subscription_tokens dst ON dst.runtime_credential_id=rc.id WHERE rc.id='e7140000-0000-0000-0000-000000000001'")" == "$pre_identity" ]]
read_reset="$(psql_admin -d "$test_db" -At -F'|' -c "SELECT upload_bytes,download_bytes,used_bytes,to_char(last_reset_at AT TIME ZONE 'UTC','YYYY-MM-DD HH24:MI:SS') FROM pvnaive.direct_naive_accounting_read('f7140000-0000-0000-0000-000000000001','2026-08-30T12:00:05Z',90)")"
[[ "$read_reset" == '0|0|0|2026-08-30 12:00:00' ]] || { echo "ERROR: read reset epoch=$read_reset" >&2; exit 1; }

# A live session continues from its cumulative counters; only post-reset delta is charged.
psql_admin -d "$test_db" <<'SQL' >/dev/null
SELECT * FROM pvnaive.direct_naive_accounting_claim('e7140000-0000-0000-0000-000000000001','node-reset','27140000-0000-0000-0000-000000000001','37140000-0000-0000-0000-000000000001',4,'upload',10,'2026-08-30T12:00:10Z');
SELECT * FROM pvnaive.direct_naive_accounting_ingest('e7140000-0000-0000-0000-000000000001','reset-user','node-reset','27140000-0000-0000-0000-000000000001','37140000-0000-0000-0000-000000000001',4,'2026-08-30T12:00:20Z',true,110,200,false);
SQL
[[ "$(psql_admin -d "$test_db" -At -F'|' -c "SELECT upload_bytes,download_bytes FROM pvnaive.direct_naive_accounting_terms WHERE service_term_id='f7140000-0000-0000-0000-000000000001'")" == '10|0' ]]
[[ "$(psql_admin -d "$test_db" -Atc "SELECT count(*) FROM pvnaive.direct_naive_accounting_events WHERE service_term_id='f7140000-0000-0000-0000-000000000001'")" == 4 ]]

# Unsafe boundaries refuse to reset instead of guessing.
psql_admin -d "$test_db" -c "UPDATE pvnaive.direct_naive_accounting_terms SET reserved_bytes=5 WHERE service_term_id='f7140000-0000-0000-0000-000000000001'" >/dev/null
reserved="$(PGPASSWORD="$app_password" psql --no-psqlrc --set ON_ERROR_STOP=1 -h "$PVNAIVE_DB_HOST" -p "$PVNAIVE_DB_PORT" -U pvnaive_app -d "$test_db" -At -F'|' <<'SQL'
BEGIN; SELECT pvnaive.set_request_context(decode(repeat('71',32),'hex'));
SELECT resettable,reason FROM pvnaive.direct_naive_accounting_reset('f7140000-0000-0000-0000-000000000001','2026-08-30T12:01:00Z',90); ROLLBACK;
SQL
)"
[[ "$(printf '%s\n' "$reserved" | grep -E '^(f|false)\|' | tail -n1)" =~ ^(f|false)\|reservation_pending$ ]]
psql_admin -d "$test_db" -c "UPDATE pvnaive.direct_naive_accounting_terms SET reserved_bytes=0,accounting_complete=false WHERE service_term_id='f7140000-0000-0000-0000-000000000001'" >/dev/null
incomplete="$(PGPASSWORD="$app_password" psql --no-psqlrc --set ON_ERROR_STOP=1 -h "$PVNAIVE_DB_HOST" -p "$PVNAIVE_DB_PORT" -U pvnaive_app -d "$test_db" -At -F'|' <<'SQL'
BEGIN; SELECT pvnaive.set_request_context(decode(repeat('71',32),'hex'));
SELECT resettable,reason FROM pvnaive.direct_naive_accounting_reset('f7140000-0000-0000-0000-000000000001','2026-08-30T12:01:00Z',90); ROLLBACK;
SQL
)"
[[ "$(printf '%s\n' "$incomplete" | grep -E '^(f|false)\|' | tail -n1)" =~ ^(f|false)\|accounting_incomplete$ ]]
psql_admin -d "$test_db" -c "UPDATE pvnaive.direct_naive_accounting_terms SET accounting_complete=true WHERE service_term_id='f7140000-0000-0000-0000-000000000001'" >/dev/null
stale="$(PGPASSWORD="$app_password" psql --no-psqlrc --set ON_ERROR_STOP=1 -h "$PVNAIVE_DB_HOST" -p "$PVNAIVE_DB_PORT" -U pvnaive_app -d "$test_db" -At -F'|' <<'SQL'
BEGIN; SELECT pvnaive.set_request_context(decode(repeat('71',32),'hex'));
SELECT resettable,reason FROM pvnaive.direct_naive_accounting_reset('f7140000-0000-0000-0000-000000000001','2026-08-30T12:03:00Z',90); ROLLBACK;
SQL
)"
[[ "$(printf '%s\n' "$stale" | grep -E '^(f|false)\|' | tail -n1)" =~ ^(f|false)\|telemetry_stale$ ]]

# Reset history relation must be append-only and tied to the global mutation ledger.
psql_admin -d "$test_db" <<'SQL' >/dev/null
INSERT INTO pvnaive.customer_mutation_keys(id,tenant_id,actor_id,idempotency_key,operation,request_hash,resource_type,resource_id)
SELECT '47140000-0000-0000-0000-000000000001',u.tenant_id,'a7140000-0000-0000-0000-000000000001','reset-usage-test-0001','customer.usage.reset',decode(repeat('78',32),'hex'),'user',u.id
FROM pvnaive.users u WHERE u.id='d7140000-0000-0000-0000-000000000001';
INSERT INTO pvnaive.direct_naive_accounting_reset_events(tenant_id,user_id,service_term_id,actor_id,customer_mutation_key_id,reason,reset_at,previous_upload_bytes,previous_download_bytes,previous_used_bytes)
SELECT u.tenant_id,u.id,'f7140000-0000-0000-0000-000000000001','a7140000-0000-0000-0000-000000000001','47140000-0000-0000-0000-000000000001','manual','2026-08-30T12:00:00Z',100,200,300
FROM pvnaive.users u WHERE u.id='d7140000-0000-0000-0000-000000000001';
UPDATE pvnaive.customer_mutation_keys SET completed_at=clock_timestamp() WHERE id='47140000-0000-0000-0000-000000000001';
INSERT INTO pvnaive.audit_events(tenant_id,actor_id,action,object_type,object_id,outcome,before_state,after_state)
SELECT u.tenant_id,'a7140000-0000-0000-0000-000000000001','customer.usage.reset','service_term','f7140000-0000-0000-0000-000000000001','success',
       '{"upload_bytes":100,"download_bytes":200,"used_bytes":300}'::jsonb,
       '{"upload_bytes":0,"download_bytes":0,"used_bytes":0}'::jsonb
FROM pvnaive.users u WHERE u.id='d7140000-0000-0000-0000-000000000001';
SQL
[[ "$(psql_admin -d "$test_db" -Atc "SELECT count(*) FROM pvnaive.direct_naive_accounting_reset_events WHERE customer_mutation_key_id='47140000-0000-0000-0000-000000000001'")" == 1 ]]
[[ "$(psql_admin -d "$test_db" -Atc "SELECT count(*) FROM pvnaive.audit_events WHERE action='customer.usage.reset' AND object_id='f7140000-0000-0000-0000-000000000001'")" == 1 ]]
[[ "$(psql_admin -d "$test_db" -Atc "SELECT completed_at IS NOT NULL FROM pvnaive.customer_mutation_keys WHERE id='47140000-0000-0000-0000-000000000001'")" =~ ^(t|true)$ ]]
set +e
psql_admin -d "$test_db" -c "DELETE FROM pvnaive.direct_naive_accounting_reset_events" >/dev/null 2>&1; delete_rc=$?
psql_admin -d "$test_db" -c "UPDATE pvnaive.direct_naive_accounting_reset_events SET previous_used_bytes=0" >/dev/null 2>&1; update_rc=$?
set -e
(( delete_rc != 0 && update_rc != 0 )) || { echo 'ERROR: reset history is mutable' >&2; exit 1; }

PVNAIVE_DISPOSABLE_DB=1 PVNAIVE_ALLOW_DESTRUCTIVE_ROLLBACK=ROLLBACK_ONE_MIGRATION PVNAIVE_MIGRATIONS_DIR="$fixture/migrations" "${repo_root}/scripts/db/rollback.sh" >/dev/null
[[ "$(psql_admin -d "$test_db" -Atc 'SELECT max(version) FROM pvnaive.schema_migrations')" == 13 ]]
[[ "$(psql_admin -d "$test_db" -Atc "SELECT to_regclass('pvnaive.direct_naive_accounting_reset_events') IS NULL")" =~ ^(t|true)$ ]]
[[ "$(psql_admin -d "$test_db" -Atc "SELECT count(*) FROM information_schema.columns WHERE table_schema='pvnaive' AND table_name='direct_naive_accounting_terms' AND column_name='last_reset_at'")" == 0 ]]
echo 'PVNAIVE_MANUAL_USAGE_RESET_MIGRATION_TEST=PASSED'
