#!/usr/bin/env bash
set -Eeuo pipefail
umask 077

repo_root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd -P)"
test_suffix="${GITHUB_RUN_ID:-local}_${GITHUB_RUN_ATTEMPT:-1}_${BASHPID}"
test_suffix="${test_suffix//[^a-zA-Z0-9_]/_}"
test_db="pvnaive_s05_chain_test_${test_suffix,,}"
temp_root="$(mktemp -d)"

: "${PVNAIVE_DB_HOST:=127.0.0.1}"
: "${PVNAIVE_DB_PORT:=5432}"
: "${PVNAIVE_DB_USER:=postgres}"
export PVNAIVE_DB_HOST PVNAIVE_DB_PORT PVNAIVE_DB_USER

psql_admin() {
  psql --no-psqlrc --set ON_ERROR_STOP=1 \
    --host "${PVNAIVE_DB_HOST}" --port "${PVNAIVE_DB_PORT}" \
    --username "${PVNAIVE_DB_USER}" "$@"
}

cleanup() {
  psql_admin --dbname postgres --command \
    "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname='${test_db}' AND pid <> pg_backend_pid()" \
    >/dev/null 2>&1 || true
  dropdb --if-exists --host "${PVNAIVE_DB_HOST}" --port "${PVNAIVE_DB_PORT}" \
    --username "${PVNAIVE_DB_USER}" "${test_db}" >/dev/null 2>&1 || true
  psql_admin --dbname postgres --command \
    'DROP ROLE IF EXISTS pvnaive_app; DROP ROLE IF EXISTS pvnaive_owner;' >/dev/null 2>&1 || true
  rm -rf -- "${temp_root}"
}
trap cleanup EXIT
cleanup

psql_admin --dbname postgres <<'SQL' >/dev/null
CREATE ROLE pvnaive_owner NOLOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOREPLICATION NOBYPASSRLS;
CREATE ROLE pvnaive_app NOLOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOREPLICATION NOBYPASSRLS;
SQL
createdb --host "${PVNAIVE_DB_HOST}" --port "${PVNAIVE_DB_PORT}" \
  --username "${PVNAIVE_DB_USER}" --owner pvnaive_owner --encoding UTF8 --template template0 "${test_db}"

v2_migrations="${temp_root}/migrations-v2"
mkdir -p "${v2_migrations}"
cp "${repo_root}"/db/migrations/0001_* "${v2_migrations}/"
cp "${repo_root}"/db/migrations/0002_* "${v2_migrations}/"
(
  cd "${v2_migrations}"
  sha256sum *.sql > SHA256SUMS
)

PVNAIVE_DB_NAME="${test_db}" PVNAIVE_MIGRATIONS_DIR="${v2_migrations}" \
  "${repo_root}/scripts/db/migrate.sh" >/dev/null
schema="$(psql_admin --dbname "${test_db}" --tuples-only --no-align \
  --command 'SELECT COALESCE(MAX(version),0) FROM pvnaive.schema_migrations')"
[[ "${schema}" == 2 ]] || { echo "ERROR: chain fixture expected schema 2, got ${schema}" >&2; exit 1; }

psql_admin --dbname "${test_db}" --command \
  "INSERT INTO pvnaive.tenants (id, tenant_type, slug, display_name) VALUES ('55000000-0000-0000-0000-000000000005', 'system', 'chain_base', 'Chain Base')" >/dev/null

age-keygen -o "${temp_root}/backup.agekey" >/dev/null 2>&1
age-keygen -y "${temp_root}/backup.agekey" >"${temp_root}/backup.recipient"
backup_output="$(
  PVNAIVE_DB_NAME="${test_db}" \
  PVNAIVE_BACKUP_ROOT="${temp_root}/backups" \
  PVNAIVE_BACKUP_RECIPIENT_FILE="${temp_root}/backup.recipient" \
  PVNAIVE_BACKUP_IDENTITY_FILE="${temp_root}/backup.agekey" \
    "${repo_root}/scripts/db/backup.sh"
)"
backup_file="$(awk -F= '/^PVNAIVE_BACKUP_PATH=/ {value=$2} END {print value}' <<<"${backup_output}")"
[[ -f "${backup_file}" ]] || { echo 'ERROR: schema2 chain backup missing' >&2; exit 1; }
grep -Fq '"schema_version": 2,' "$(dirname -- "${backup_file}")/metadata.json"

PVNAIVE_DB_NAME="${test_db}" PVNAIVE_MIGRATIONS_DIR="${repo_root}/db/migrations" \
  "${repo_root}/scripts/db/migrate.sh" >/dev/null
schema="$(psql_admin --dbname "${test_db}" --tuples-only --no-align \
  --command 'SELECT COALESCE(MAX(version),0) FROM pvnaive.schema_migrations')"
[[ "${schema}" == 7 ]] || { echo "ERROR: chain fixture expected schema 7, got ${schema}" >&2; exit 1; }

# The old single-step gate must remain strict: a schema2 backup is not fresh
# enough to authorize a normal 7 -> 6 rollback.
if PVNAIVE_DB_NAME="${test_db}" \
  PVNAIVE_MIGRATIONS_DIR="${repo_root}/db/migrations" \
  PVNAIVE_ALLOW_DESTRUCTIVE_ROLLBACK=ROLLBACK_ONE_MIGRATION \
  PVNAIVE_CONFIRMED_BACKUP="${backup_file}" \
  PVNAIVE_BACKUP_IDENTITY_FILE="${temp_root}/backup.agekey" \
    "${repo_root}/scripts/db/rollback.sh" >/dev/null 2>&1; then
  echo 'ERROR: single-step rollback accepted an older chain-base backup' >&2
  exit 1
fi

for expected in 6 5 4 3 2; do
  rollback_output="$(
    PVNAIVE_DB_NAME="${test_db}" \
    PVNAIVE_MIGRATIONS_DIR="${repo_root}/db/migrations" \
    PVNAIVE_ALLOW_DESTRUCTIVE_ROLLBACK=ROLLBACK_MIGRATION_CHAIN \
    PVNAIVE_ROLLBACK_CHAIN_TARGET_SCHEMA=2 \
    PVNAIVE_CONFIRMED_BACKUP="${backup_file}" \
    PVNAIVE_BACKUP_IDENTITY_FILE="${temp_root}/backup.agekey" \
      "${repo_root}/scripts/db/rollback.sh"
  )"
  grep -Fqx "PVNAIVE_SCHEMA_VERSION=${expected}" <<<"${rollback_output}"
  grep -Fqx 'ROLLBACK_CHAIN_TARGET_SCHEMA=2' <<<"${rollback_output}"
  grep -Fqx 'PVNAIVE_ROLLBACK_RESULT=PASSED' <<<"${rollback_output}"
  schema="$(psql_admin --dbname "${test_db}" --tuples-only --no-align \
    --command 'SELECT COALESCE(MAX(version),0) FROM pvnaive.schema_migrations')"
  [[ "${schema}" == "${expected}" ]] || {
    echo "ERROR: migration-chain rollback expected ${expected}, got ${schema}" >&2
    exit 1
  }
done

seed="$(psql_admin --dbname "${test_db}" --tuples-only --no-align \
  --command "SELECT slug FROM pvnaive.tenants WHERE id='55000000-0000-0000-0000-000000000005'")"
[[ "${seed}" == chain_base ]] || { echo 'ERROR: chain rollback lost base-schema data' >&2; exit 1; }

echo 'PVNAIVE_DB_ROLLBACK_CHAIN_TEST=PASSED'
