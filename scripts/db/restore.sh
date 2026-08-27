#!/usr/bin/env bash
# Restore only into a new PVNaive restore-drill database.
set -Eeuo pipefail
umask 077

script_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd -P)"
# shellcheck source=scripts/db/lib.sh
source "${script_dir}/lib.sh"

for command_name in age pg_restore psql createdb dropdb sha256sum grep sed; do pvnaive_require_command "${command_name}"; done
[[ -z "${PVNAIVE_RUN_AS_OS_USER:-}" ]] || pvnaive_require_command runuser
pvnaive_db_defaults

backup_file="${PVNAIVE_RESTORE_BACKUP:-}"
target_db="${PVNAIVE_RESTORE_TARGET_DB:-}"
identity_file="${PVNAIVE_BACKUP_IDENTITY_FILE:-/etc/pvnaive/backup.agekey}"
[[ -f "${backup_file}" ]] || pvnaive_die "restore backup does not exist"
[[ -f "${identity_file}" ]] || pvnaive_die "age identity does not exist"
[[ "${target_db}" =~ ^pvnaive_restore_test_[a-z0-9_]+$ ]] || pvnaive_die "restore target must be a new pvnaive_restore_test_* database"
pvnaive_require_identifier "restore target database" "${target_db}"

backup_dir="$(dirname -- "${backup_file}")"
[[ -f "${backup_dir}/SHA256SUMS" ]] || pvnaive_die "restore checksum manifest is missing"
[[ -f "${backup_dir}/metadata.json" ]] || pvnaive_die "restore metadata is missing"
grep -Eq '^[0-9a-f]{64}  metadata\.json$' "${backup_dir}/SHA256SUMS" || pvnaive_die "restore metadata is not checksummed"
grep -Eq '^[0-9a-f]{64}  pvnaive\.dump\.age$' "${backup_dir}/SHA256SUMS" || pvnaive_die "restore archive is not checksummed"
(cd "${backup_dir}" && sha256sum --check --strict SHA256SUMS) >/dev/null || pvnaive_die "restore checksum validation failed"
grep -qx '  "product": "PVNaive",' "${backup_dir}/metadata.json" || pvnaive_die "restore metadata product mismatch"
grep -qx '  "encrypted": true,' "${backup_dir}/metadata.json" || pvnaive_die "restore metadata encryption marker is missing"
metadata_schema_version="$(sed -n 's/^  "schema_version": \([0-9][0-9]*\),$/\1/p' "${backup_dir}/metadata.json")"
[[ "${metadata_schema_version}" =~ ^[1-9][0-9]*$ ]] || pvnaive_die "restore metadata schema version is invalid"

exists="$(pvnaive_psql_at --dbname postgres --command "SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = '${target_db}')")"
[[ "${exists}" == "f" ]] || pvnaive_die "restore target already exists"

temp_dir="$(mktemp -d)"
target_created=0
cleanup() {
  rm -rf -- "${temp_dir}"
  if [[ "${target_created}" == "1" ]]; then
    pvnaive_admin_tool dropdb --if-exists "${target_db}" >/dev/null 2>&1 || true
  fi
}
trap cleanup EXIT

age --decrypt --identity "${identity_file}" --output "${temp_dir}/restore.dump" "${backup_file}"
pg_restore --list "${temp_dir}/restore.dump" >/dev/null || pvnaive_die "restore archive parse failed"
pvnaive_admin_tool createdb --owner pvnaive_owner --encoding UTF8 --template template0 "${target_db}"
target_created=1
# The caller may run PostgreSQL tools as the postgres OS account while this
# script owns the mode-0700 temporary directory. Open the archive as the
# current account and pass it through stdin so pg_restore never needs path
# access to that directory.
pvnaive_db_tool pg_restore --dbname "${target_db}" --exit-on-error --single-transaction --no-owner --no-acl --role pvnaive_owner < "${temp_dir}/restore.dump" >/dev/null

restored_version="$(pvnaive_psql_at --dbname "${target_db}" --command 'SELECT COALESCE(MAX(version), 0) FROM pvnaive.schema_migrations')"
[[ "${restored_version}" == "${metadata_schema_version}" ]] || pvnaive_die "restored schema version does not match backup metadata"
echo "PVNAIVE_RESTORE_RESULT=PASSED"
echo "PVNAIVE_RESTORE_TARGET=${target_db}"
echo "PVNAIVE_RESTORE_SCHEMA_VERSION=${restored_version}"
target_created=0
rm -rf -- "${temp_dir}"
trap - EXIT
