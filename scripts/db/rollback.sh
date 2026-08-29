#!/usr/bin/env bash
# Roll back only after the explicit PVNaive destructive gate passes.
set -Eeuo pipefail
umask 077

script_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd -P)"
repo_root="$(cd -- "${script_dir}/../.." && pwd -P)"
# shellcheck source=scripts/db/lib.sh
source "${script_dir}/lib.sh"

pvnaive_require_command psql
pvnaive_require_command sha256sum
pvnaive_require_command grep
pvnaive_require_command sed
[[ -z "${PVNAIVE_RUN_AS_OS_USER:-}" ]] || pvnaive_require_command runuser
pvnaive_db_defaults

can_assume_owner="$(pvnaive_psql_at --command "SELECT pg_has_role(current_user, 'pvnaive_owner', 'USAGE')")"
[[ "${can_assume_owner}" == "t" ]] || pvnaive_die "rollback connection cannot assume pvnaive_owner"

migrations_dir="${PVNAIVE_MIGRATIONS_DIR:-${repo_root}/db/migrations}"
(
  cd "${migrations_dir}"
  sha256sum --check --strict SHA256SUMS
) >/dev/null || pvnaive_die "migration checksum verification failed"

current_version="$(pvnaive_psql_at --command "SELECT COALESCE(MAX(version), 0) FROM pvnaive.schema_migrations")"
((current_version > 0)) || pvnaive_die "no applied migration exists"
version_text="$(printf '%04d' "${current_version}")"
up_filename="$(pvnaive_psql_at --command "SELECT filename FROM pvnaive.schema_migrations WHERE version = ${current_version}")"
down_filename="${up_filename%.up.sql}.down.sql"
down_file="${migrations_dir}/${down_filename}"
[[ -f "${down_file}" ]] || pvnaive_die "down migration is missing: ${down_filename}"
grep -qx -- "-- pvnaive:migration-version ${version_text}" "${down_file}" || pvnaive_die "down migration version mismatch"
grep -qx -- "-- pvnaive:transactional true" "${down_file}" || pvnaive_die "non-transactional rollback refused"
grep -qx -- "-- pvnaive:destructive true" "${down_file}" || pvnaive_die "rollback marker missing"

rollback_mode="${PVNAIVE_ALLOW_DESTRUCTIVE_ROLLBACK:-}"
chain_target="${PVNAIVE_ROLLBACK_CHAIN_TARGET_SCHEMA:-}"
expected_backup_schema_version=""

if [[ "${PVNAIVE_DISPOSABLE_DB:-0}" == "1" && "${PVNAIVE_DB_NAME}" =~ ^pvnaive_(migration|restore)_test_[a-z0-9_]+$ ]]; then
  :
else
  if ((current_version == 1)); then
    [[ "${rollback_mode}" == "DROP_PVNAIVE_SCHEMA" ]] || \
      pvnaive_die "schema-v1 production rollback requires PVNAIVE_ALLOW_DESTRUCTIVE_ROLLBACK=DROP_PVNAIVE_SCHEMA"
    expected_backup_schema_version=1
  elif [[ "${rollback_mode}" == "ROLLBACK_ONE_MIGRATION" ]]; then
    expected_backup_schema_version=$((current_version - 1))
  elif [[ "${rollback_mode}" == "ROLLBACK_MIGRATION_CHAIN" ]]; then
    [[ "${chain_target}" =~ ^[1-9][0-9]*$ ]] || pvnaive_die "migration-chain rollback requires a numeric PVNAIVE_ROLLBACK_CHAIN_TARGET_SCHEMA"
    ((10#${chain_target} < current_version)) || pvnaive_die "migration-chain rollback target must be older than current schema"
    expected_backup_schema_version=$((10#${chain_target}))
  else
    pvnaive_die "production rollback requires ROLLBACK_ONE_MIGRATION or ROLLBACK_MIGRATION_CHAIN"
  fi

  backup_file="${PVNAIVE_CONFIRMED_BACKUP:-}"
  identity_file="${PVNAIVE_BACKUP_IDENTITY_FILE:-/etc/pvnaive/backup.agekey}"
  [[ -f "${backup_file}" && -f "${identity_file}" ]] || pvnaive_die "verified encrypted backup and identity are required"
  backup_dir="$(dirname -- "${backup_file}")"
  [[ -f "${backup_dir}/SHA256SUMS" ]] || pvnaive_die "backup checksum manifest is missing"
  [[ -f "${backup_dir}/metadata.json" ]] || pvnaive_die "backup metadata is missing"
  grep -Eq '^[0-9a-f]{64}  metadata\.json$' "${backup_dir}/SHA256SUMS" || pvnaive_die "backup metadata is not checksummed"
  grep -Eq '^[0-9a-f]{64}  pvnaive\.dump\.age$' "${backup_dir}/SHA256SUMS" || pvnaive_die "backup archive is not checksummed"
  (cd "${backup_dir}" && sha256sum --check --strict SHA256SUMS) >/dev/null || pvnaive_die "backup checksum validation failed"
  grep -qx '  "product": "PVNaive",' "${backup_dir}/metadata.json" || pvnaive_die "backup metadata product mismatch"
  grep -qx '  "encrypted": true,' "${backup_dir}/metadata.json" || pvnaive_die "backup metadata encryption marker is missing"
  metadata_schema_version="$(sed -n 's/^  "schema_version": \([0-9][0-9]*\),$/\1/p' "${backup_dir}/metadata.json")"
  [[ "${metadata_schema_version}" == "${expected_backup_schema_version}" ]] || \
    pvnaive_die "backup schema version ${metadata_schema_version:-unknown} does not match required rollback safety schema ${expected_backup_schema_version}"
  pvnaive_require_command age
  pvnaive_require_command pg_restore
  age --decrypt --identity "${identity_file}" "${backup_file}" | pg_restore --list >/dev/null || \
    pvnaive_die "encrypted backup could not be decrypted and parsed"
fi

echo "ROLLBACK_VERSION=${version_text}"
echo "ROLLBACK_IS_DESTRUCTIVE=true"
if [[ "${rollback_mode}" == "ROLLBACK_MIGRATION_CHAIN" ]]; then
  echo "ROLLBACK_CHAIN_TARGET_SCHEMA=${chain_target}"
fi
{
  printf '%s\n' "SELECT pg_advisory_xact_lock(hashtext('pvnaive-schema-migrations'));"
  cat -- "${down_file}"
} | pvnaive_psql --single-transaction --file - >/dev/null

if ((current_version == 1)); then
  remaining_schema="$(pvnaive_psql_at --command "SELECT to_regnamespace('pvnaive') IS NOT NULL")"
  [[ "${remaining_schema}" == "f" ]] || pvnaive_die "schema-v1 rollback verification failed"
  echo "PVNAIVE_SCHEMA_VERSION=0"
else
  remaining_schema="$(pvnaive_psql_at --command "SELECT to_regnamespace('pvnaive') IS NOT NULL")"
  [[ "${remaining_schema}" == "t" ]] || pvnaive_die "migration rollback unexpectedly removed pvnaive schema"
  remaining_version="$(pvnaive_psql_at --command 'SELECT COALESCE(MAX(version), 0) FROM pvnaive.schema_migrations')"
  expected_remaining=$((current_version - 1))
  [[ "${remaining_version}" == "${expected_remaining}" ]] || \
    pvnaive_die "migration rollback version mismatch: got ${remaining_version}, expected ${expected_remaining}"
  echo "PVNAIVE_SCHEMA_VERSION=${remaining_version}"
fi

echo "PVNAIVE_ROLLBACK_RESULT=PASSED"
