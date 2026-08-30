#!/usr/bin/env bash
set -Eeuo pipefail
umask 077
repo_root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd -P)"
up="${repo_root}/db/migrations/0012_accounting_baseline_truth.up.sql"
down="${repo_root}/db/migrations/0012_accounting_baseline_truth.down.sql"
[[ -f "${up}" && -f "${down}" ]] || { echo 'ERROR: missing schema-12 migration pair' >&2; exit 1; }
grep -Fqx -- '-- pvnaive:migration-version 0012' "${up}"
grep -Fqx -- '-- pvnaive:destructive false' "${up}"
grep -Fqx -- '-- pvnaive:destructive true' "${down}"
: "${PVNAIVE_DB_HOST:=127.0.0.1}"; : "${PVNAIVE_DB_PORT:=5432}"; : "${PVNAIVE_DB_USER:=postgres}"
export PVNAIVE_DB_HOST PVNAIVE_DB_PORT PVNAIVE_DB_USER
suffix="${GITHUB_RUN_ID:-local}_${GITHUB_RUN_ATTEMPT:-1}_${BASHPID}"; suffix="${suffix//[^a-zA-Z0-9_]/_}"
test_db="pvnaive_migration_test_baseline_${suffix,,}"
fixture="$(mktemp -d)"
psql_admin(){ psql --no-psqlrc --set ON_ERROR_STOP=1 -h "${PVNAIVE_DB_HOST}" -p "${PVNAIVE_DB_PORT}" -U "${PVNAIVE_DB_USER}" "$@"; }
cleanup(){ rm -rf -- "${fixture}" 2>/dev/null || true; psql_admin -d postgres -c "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname='${test_db}' AND pid<>pg_backend_pid()" >/dev/null 2>&1 || true; dropdb --if-exists -h "${PVNAIVE_DB_HOST}" -p "${PVNAIVE_DB_PORT}" -U "${PVNAIVE_DB_USER}" "${test_db}" >/dev/null 2>&1 || true; psql_admin -d postgres -c 'DROP ROLE IF EXISTS pvnaive_app; DROP ROLE IF EXISTS pvnaive_owner;' >/dev/null 2>&1 || true; }
trap cleanup EXIT HUP INT TERM; cleanup
psql_admin -d postgres <<'SQL' >/dev/null
CREATE ROLE pvnaive_owner NOLOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOREPLICATION NOBYPASSRLS;
CREATE ROLE pvnaive_app NOLOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOREPLICATION NOBYPASSRLS;
SQL
createdb -h "${PVNAIVE_DB_HOST}" -p "${PVNAIVE_DB_PORT}" -U "${PVNAIVE_DB_USER}" --owner pvnaive_owner --encoding UTF8 --template template0 "${test_db}"
export PVNAIVE_DB_NAME="${test_db}"
mkdir -p "${fixture}/migrations"
for version in $(seq 1 11); do prefix="$(printf '%04d' "${version}")"; cp "${repo_root}/db/migrations/${prefix}_"*.sql "${fixture}/migrations/"; done
( cd "${fixture}/migrations" && sha256sum *.sql > SHA256SUMS )
PVNAIVE_MIGRATIONS_DIR="${fixture}/migrations" "${repo_root}/scripts/db/migrate.sh" >/dev/null
[[ "$(psql_admin -d "${test_db}" -Atc 'SELECT MAX(version) FROM pvnaive.schema_migrations')" == 11 ]]

psql_admin -d "${test_db}" <<'SQL' >/dev/null
INSERT INTO pvnaive.actors(id,tenant_id,actor_role,email,display_name,status) VALUES('a5120000-0000-0000-0000-000000000001',NULL,'owner','baseline-owner@example.invalid','Baseline Owner','active');
INSERT INTO pvnaive.users(id,tenant_id,username,display_name,status,created_by_actor_id) SELECT 'b5120000-0000-0000-0000-000000000001',id,'legacy-baseline','Legacy Baseline','active','a5120000-0000-0000-0000-000000000001' FROM pvnaive.tenants WHERE slug='direct' AND tenant_type='system';
INSERT INTO pvnaive.naive_runtime_credentials(id,username,secret_hash,secret_ciphertext,secret_nonce,encryption_key_id,status,origin,created_by_actor_id,updated_by_actor_id,created_at) VALUES('c5120000-0000-0000-0000-000000000001','legacy-baseline',decode(repeat('11',32),'hex'),decode(repeat('12',32),'hex'),decode(repeat('13',12),'hex'),'runtime-v1','active','imported','a5120000-0000-0000-0000-000000000001','a5120000-0000-0000-0000-000000000001','2026-08-29T10:00:00Z');
INSERT INTO pvnaive.service_terms(id,tenant_id,user_id,quota_bytes,duration_seconds,start_policy,purchased_at,state,renewal_kind) SELECT 'd5120000-0000-0000-0000-000000000001',tenant_id,id,1000000,2592000,'on_creation','2026-08-29T12:00:00Z','active','initial' FROM pvnaive.users WHERE id='b5120000-0000-0000-0000-000000000001';
INSERT INTO pvnaive.user_runtime_credentials(id,tenant_id,user_id,service_term_id,runtime_credential_id,role,bound_at) SELECT 'e5120000-0000-0000-0000-000000000001',tenant_id,id,'d5120000-0000-0000-0000-000000000001','c5120000-0000-0000-0000-000000000001','primary','2026-08-29T12:34:56Z' FROM pvnaive.users WHERE id='b5120000-0000-0000-0000-000000000001';
SQL
cp "${repo_root}/db/migrations/0012_"*.sql "${fixture}/migrations/"
( cd "${fixture}/migrations" && sha256sum *.sql > SHA256SUMS )
PVNAIVE_MIGRATIONS_DIR="${fixture}/migrations" "${repo_root}/scripts/db/migrate.sh" >/dev/null
[[ "$(psql_admin -d "${test_db}" -Atc 'SELECT MAX(version) FROM pvnaive.schema_migrations')" == 12 ]]
legacy="$(psql_admin -d "${test_db}" -Atc "SELECT concat_ws('|',accounting_baseline_state,accounting_baseline_source,to_char(accounting_baseline_cutoff_at AT TIME ZONE 'UTC','YYYY-MM-DD HH24:MI:SS'),(accounting_baseline_upload_bytes IS NULL)::text,(accounting_baseline_download_bytes IS NULL)::text) FROM pvnaive.service_terms WHERE id='d5120000-0000-0000-0000-000000000001'")"
[[ "${legacy}" == 'unknown|legacy_unavailable|2026-08-29 12:34:56|true|true' || "${legacy}" == 'unknown|legacy_unavailable|2026-08-29 12:34:56|t|t' ]] || { echo "ERROR: backfill mismatch ${legacy}" >&2; exit 1; }

psql_admin -d "${test_db}" <<'SQL' >/dev/null
INSERT INTO pvnaive.users(id,tenant_id,username,display_name,status,created_by_actor_id) SELECT 'b5120000-0000-0000-0000-000000000002',id,'fresh-baseline','Fresh Baseline','active','a5120000-0000-0000-0000-000000000001' FROM pvnaive.tenants WHERE slug='direct' AND tenant_type='system';
INSERT INTO pvnaive.service_terms(id,tenant_id,user_id,quota_bytes,duration_seconds,start_policy,purchased_at,state,renewal_kind,accounting_baseline_state,accounting_baseline_source,accounting_baseline_cutoff_at,accounting_baseline_upload_bytes,accounting_baseline_download_bytes) SELECT 'd5120000-0000-0000-0000-000000000002',tenant_id,id,2000000,2592000,'on_creation','2026-08-30T01:00:00Z','active','initial','known','fresh_managed_term','2026-08-30T01:00:00Z',0,0 FROM pvnaive.users WHERE id='b5120000-0000-0000-0000-000000000002';
INSERT INTO pvnaive.users(id,tenant_id,username,display_name,status,created_by_actor_id) SELECT 'b5120000-0000-0000-0000-000000000003',id,'authoritative-baseline','Authoritative','active','a5120000-0000-0000-0000-000000000001' FROM pvnaive.tenants WHERE slug='direct' AND tenant_type='system';
INSERT INTO pvnaive.service_terms(id,tenant_id,user_id,quota_bytes,duration_seconds,start_policy,purchased_at,state,renewal_kind,accounting_baseline_state,accounting_baseline_source,accounting_baseline_cutoff_at,accounting_baseline_upload_bytes,accounting_baseline_download_bytes) SELECT 'd5120000-0000-0000-0000-000000000003',tenant_id,id,5000000,2592000,'on_creation','2026-08-30T01:30:00Z','active','initial','known','authoritative_import','2026-08-30T01:30:00Z',100,200 FROM pvnaive.users WHERE id='b5120000-0000-0000-0000-000000000003';
SQL
set +e
psql_admin -d "${test_db}" -c "UPDATE pvnaive.service_terms SET accounting_baseline_upload_bytes=999 WHERE id='d5120000-0000-0000-0000-000000000003'" >/dev/null 2>&1; immutable_rc=$?
psql_admin -d "${test_db}" -c "INSERT INTO pvnaive.service_terms(tenant_id,user_id,quota_bytes,duration_seconds,start_policy,purchased_at,state,accounting_baseline_state,accounting_baseline_source,accounting_baseline_cutoff_at,accounting_baseline_upload_bytes,accounting_baseline_download_bytes) SELECT tenant_id,id,1,60,'on_creation',clock_timestamp(),'active','unknown','legacy_unavailable',clock_timestamp(),1,NULL FROM pvnaive.users WHERE id='b5120000-0000-0000-0000-000000000002'" >/dev/null 2>&1; invalid_rc=$?
set -e
(( immutable_rc != 0 && invalid_rc != 0 )) || { echo 'ERROR: baseline mutability/truth constraint failed' >&2; exit 1; }
psql_admin -d "${test_db}" -c "UPDATE pvnaive.service_terms SET quota_bytes=6000000 WHERE id='d5120000-0000-0000-0000-000000000003'" >/dev/null
[[ "$(psql_admin -d "${test_db}" -Atc "SELECT concat_ws('|',quota_bytes::text,accounting_baseline_state,accounting_baseline_source,accounting_baseline_upload_bytes::text,accounting_baseline_download_bytes::text) FROM pvnaive.service_terms WHERE id='d5120000-0000-0000-0000-000000000003'")" == '6000000|known|authoritative_import|100|200' ]]
PVNAIVE_DISPOSABLE_DB=1 PVNAIVE_ALLOW_DESTRUCTIVE_ROLLBACK=ROLLBACK_ONE_MIGRATION PVNAIVE_MIGRATIONS_DIR="${fixture}/migrations" "${repo_root}/scripts/db/rollback.sh" >/dev/null
rolled="$(psql_admin -d "${test_db}" -Atc "SELECT COALESCE(MAX(version),0) || '|' || ((SELECT count(*) FROM information_schema.columns WHERE table_schema='pvnaive' AND table_name='service_terms' AND column_name LIKE 'accounting_baseline_%')=0)::text FROM pvnaive.schema_migrations")"
[[ "${rolled}" == '11|true' || "${rolled}" == '11|t' ]] || { echo "ERROR: rollback mismatch ${rolled}" >&2; exit 1; }
echo 'PVNAIVE_ACCOUNTING_BASELINE_MIGRATION_TEST=PASSED'
