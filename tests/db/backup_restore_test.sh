#!/usr/bin/env bash
set -Eeuo pipefail
umask 077

repo_root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd -P)"
test_suffix="${GITHUB_RUN_ID:-local}_${GITHUB_RUN_ATTEMPT:-1}"
test_suffix="${test_suffix//[^a-zA-Z0-9_]/_}"
source_db="pvnaive_backup_test_${test_suffix,,}"
restore_db="pvnaive_restore_test_${test_suffix,,}"
temp_root="$(mktemp -d)"

: "${PVNAIVE_DB_HOST:=127.0.0.1}"
: "${PVNAIVE_DB_PORT:=5432}"
: "${PVNAIVE_DB_USER:=postgres}"
export PVNAIVE_DB_HOST PVNAIVE_DB_PORT PVNAIVE_DB_USER

psql_admin() {
  psql --no-psqlrc --set ON_ERROR_STOP=1 --host "${PVNAIVE_DB_HOST}" --port "${PVNAIVE_DB_PORT}" --username "${PVNAIVE_DB_USER}" "$@"
}

cleanup_database_objects() {
  for database_name in "${restore_db}" "${source_db}"; do
    psql_admin --dbname postgres --command "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = '${database_name}' AND pid <> pg_backend_pid()" >/dev/null 2>&1 || true
    dropdb --if-exists --host "${PVNAIVE_DB_HOST}" --port "${PVNAIVE_DB_PORT}" --username "${PVNAIVE_DB_USER}" "${database_name}" >/dev/null 2>&1 || true
  done
  psql_admin --dbname postgres --command 'DROP ROLE IF EXISTS pvnaive_app; DROP ROLE IF EXISTS pvnaive_owner;' >/dev/null 2>&1 || true
}

cleanup() {
  cleanup_database_objects
  rm -rf -- "${temp_root}"
}
trap cleanup EXIT
cleanup_database_objects

psql_admin --dbname postgres <<'SQL' >/dev/null
CREATE ROLE pvnaive_owner NOLOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOREPLICATION NOBYPASSRLS;
CREATE ROLE pvnaive_app NOLOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOREPLICATION NOBYPASSRLS;
SQL
createdb --host "${PVNAIVE_DB_HOST}" --port "${PVNAIVE_DB_PORT}" --username "${PVNAIVE_DB_USER}" --owner pvnaive_owner --encoding UTF8 --template template0 "${source_db}"

export PVNAIVE_DB_NAME="${source_db}"
"${repo_root}/scripts/db/migrate.sh" >/dev/null
psql_admin --dbname "${source_db}" --command \
  "INSERT INTO pvnaive.tenants (id, tenant_type, slug, display_name) VALUES ('44000000-0000-0000-0000-000000000004', 'reseller', 'backup_test', 'Backup Test')" >/dev/null

age-keygen -o "${temp_root}/backup.agekey" >/dev/null 2>&1
age-keygen -y "${temp_root}/backup.agekey" > "${temp_root}/backup.recipient"
backup_output="$(
  PVNAIVE_BACKUP_ROOT="${temp_root}/backups" \
  PVNAIVE_BACKUP_RECIPIENT_FILE="${temp_root}/backup.recipient" \
  PVNAIVE_BACKUP_IDENTITY_FILE="${temp_root}/backup.agekey" \
    "${repo_root}/scripts/db/backup.sh"
)"
grep -Fqx 'PVNAIVE_BACKUP_RESULT=PASSED' <<< "${backup_output}"
backup_file="$(awk -F= '/^PVNAIVE_BACKUP_PATH=/ {value=$2} END {print value}' <<< "${backup_output}")"
[[ -f "${backup_file}" ]] || { echo "ERROR: backup archive was not created" >&2; exit 1; }
backup_dir="$(dirname -- "${backup_file}")"
(cd "${backup_dir}" && sha256sum --check --strict SHA256SUMS) >/dev/null
[[ ! -e "${backup_dir}/pvnaive.dump" ]] || { echo "ERROR: plaintext backup archive exists" >&2; exit 1; }

restore_output="$(
  PVNAIVE_RESTORE_BACKUP="${backup_file}" \
  PVNAIVE_RESTORE_TARGET_DB="${restore_db}" \
  PVNAIVE_BACKUP_IDENTITY_FILE="${temp_root}/backup.agekey" \
    "${repo_root}/scripts/db/restore.sh"
)"
grep -Fqx 'PVNAIVE_RESTORE_RESULT=PASSED' <<< "${restore_output}"
restored_row="$(psql_admin --dbname "${restore_db}" --tuples-only --no-align --command "SELECT slug FROM pvnaive.tenants WHERE id = '44000000-0000-0000-0000-000000000004'")"
[[ "${restored_row}" == "backup_test" ]] || { echo "ERROR: restored data verification failed" >&2; exit 1; }
plaintext_dump="$(find "${temp_root}" -type f -name '*.dump' -print -quit)"
[[ -z "${plaintext_dump}" ]] || {
  echo "ERROR: plaintext dump file was materialized: ${plaintext_dump}" >&2
  exit 1
}

echo "PVNAIVE_DB_BACKUP_RESTORE_TEST=PASSED"
