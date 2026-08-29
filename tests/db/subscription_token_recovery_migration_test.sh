#!/usr/bin/env bash
set -Eeuo pipefail
umask 077

repo_root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd -P)"
test_suffix="${GITHUB_RUN_ID:-local}_${GITHUB_RUN_ATTEMPT:-1}_${BASHPID}"
test_suffix="${test_suffix//[^a-zA-Z0-9_]/_}"
test_db="pvnaive_s06_token_recovery_${test_suffix,,}"
tmp="$(mktemp -d)"
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
  rm -rf -- "${tmp}"
}
trap cleanup EXIT
cleanup

psql_admin --dbname postgres <<'SQL' >/dev/null
CREATE ROLE pvnaive_owner NOLOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOREPLICATION NOBYPASSRLS;
CREATE ROLE pvnaive_app NOLOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOREPLICATION NOBYPASSRLS;
SQL
createdb --host "${PVNAIVE_DB_HOST}" --port "${PVNAIVE_DB_PORT}" --username "${PVNAIVE_DB_USER}" --owner pvnaive_owner --encoding UTF8 --template template0 "${test_db}"

v6="${tmp}/migrations-v6"
mkdir -p "${v6}"
for version in 0001 0002 0003 0004 0005 0006; do cp "${repo_root}"/db/migrations/${version}_* "${v6}/"; done
( cd "${v6}"; sha256sum *.sql > SHA256SUMS )
PVNAIVE_DB_NAME="${test_db}" PVNAIVE_MIGRATIONS_DIR="${v6}" "${repo_root}/scripts/db/migrate.sh" >/dev/null
schema="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command 'SELECT COALESCE(MAX(version),0) FROM pvnaive.schema_migrations')"
[[ "${schema}" == 6 ]] || { echo "ERROR: S06 fixture expected schema 6, got ${schema}" >&2; exit 1; }

age-keygen -o "${tmp}/backup.agekey" >/dev/null 2>&1
age-keygen -y "${tmp}/backup.agekey" >"${tmp}/backup.recipient"
backup_output="$(PVNAIVE_DB_NAME="${test_db}" PVNAIVE_BACKUP_ROOT="${tmp}/backups" PVNAIVE_BACKUP_RECIPIENT_FILE="${tmp}/backup.recipient" PVNAIVE_BACKUP_IDENTITY_FILE="${tmp}/backup.agekey" "${repo_root}/scripts/db/backup.sh")"
backup_file="$(awk -F= '$1=="PVNAIVE_BACKUP_PATH" {print $2}' <<<"${backup_output}")"
[[ -f "${backup_file}" ]] || { echo 'ERROR: schema 6 backup missing' >&2; exit 1; }
grep -Fq '"schema_version": 6,' "$(dirname -- "${backup_file}")/metadata.json"

PVNAIVE_DB_NAME="${test_db}" PVNAIVE_MIGRATIONS_DIR="${repo_root}/db/migrations" "${repo_root}/scripts/db/migrate.sh" >/dev/null
schema="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command 'SELECT COALESCE(MAX(version),0) FROM pvnaive.schema_migrations')"
[[ "${schema}" == 7 ]] || { echo "ERROR: S06 migration expected schema 7, got ${schema}" >&2; exit 1; }

contract="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command "
SELECT
  EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema='pvnaive' AND table_name='direct_subscription_tokens' AND column_name='token_ciphertext' AND data_type='bytea') || '|' ||
  EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema='pvnaive' AND table_name='direct_subscription_tokens' AND column_name='token_nonce' AND data_type='bytea') || '|' ||
  EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema='pvnaive' AND table_name='direct_subscription_tokens' AND column_name='token_encryption_key_id' AND data_type='text') || '|' ||
  EXISTS (SELECT 1 FROM pg_constraint WHERE conrelid='pvnaive.direct_subscription_tokens'::regclass AND conname='direct_subscription_token_recovery_envelope_check');")"
[[ "${contract}" == "true|true|true|true" || "${contract}" == "t|t|t|t" ]] || { echo "ERROR: schema 7 recovery contract failed: ${contract}" >&2; exit 1; }

# Legacy schema-6 tokens remain valid rows after migration: all recovery fields
# may stay NULL until the Owner explicitly reissues the link once.
legacy_allowed="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command "SELECT pg_get_constraintdef(oid) LIKE '%token_ciphertext IS NULL%' FROM pg_constraint WHERE conrelid='pvnaive.direct_subscription_tokens'::regclass AND conname='direct_subscription_token_recovery_envelope_check';")"
[[ "${legacy_allowed}" == t || "${legacy_allowed}" == true ]] || { echo 'ERROR: schema 7 does not preserve legacy NULL recovery envelope' >&2; exit 1; }

rollback_output="$(PVNAIVE_DB_NAME="${test_db}" PVNAIVE_MIGRATIONS_DIR="${repo_root}/db/migrations" PVNAIVE_ALLOW_DESTRUCTIVE_ROLLBACK=ROLLBACK_MIGRATION_CHAIN PVNAIVE_ROLLBACK_CHAIN_TARGET_SCHEMA=6 PVNAIVE_CONFIRMED_BACKUP="${backup_file}" PVNAIVE_BACKUP_IDENTITY_FILE="${tmp}/backup.agekey" "${repo_root}/scripts/db/rollback.sh")"
grep -Fqx 'PVNAIVE_SCHEMA_VERSION=6' <<<"${rollback_output}"
grep -Fqx 'PVNAIVE_ROLLBACK_RESULT=PASSED' <<<"${rollback_output}"
columns_after="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command "SELECT count(*) FROM information_schema.columns WHERE table_schema='pvnaive' AND table_name='direct_subscription_tokens' AND column_name IN ('token_ciphertext','token_nonce','token_encryption_key_id');")"
[[ "${columns_after}" == 0 ]] || { echo "ERROR: schema 7 recovery columns remain after rollback: ${columns_after}" >&2; exit 1; }

echo 'SUBSCRIPTION_TOKEN_RECOVERY_MIGRATION_TEST=PASSED'
