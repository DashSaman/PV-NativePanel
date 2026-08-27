#!/usr/bin/env bash
# Roll back the complete PVNaive schema only after the explicit destructive gate passes.
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
pvnaive_require_command sort
[[ -z "${PVNAIVE_RUN_AS_OS_USER:-}" ]] || pvnaive_require_command runuser
pvnaive_db_defaults

can_assume_owner="$(pvnaive_psql_at --command "SELECT pg_has_role(current_user, 'pvnaive_owner', 'USAGE')")"
[[ "${can_assume_owner}" == "t" ]] || pvnaive_die "rollback connection cannot assume pvnaive_owner"

migrations_dir="${PVNAIVE_MIGRATIONS_DIR:-${repo_root}/db/migrations}"
manifest="${migrations_dir}/SHA256SUMS"
[[ -d "${migrations_dir}" ]] || pvnaive_die "migration directory not found: ${migrations_dir}"
[[ -f "${manifest}" ]] || pvnaive_die "migration checksum manifest is missing"
(
  cd "${migrations_dir}"
  sha256sum --check --strict SHA256SUMS
) >/dev/null || pvnaive_die "migration checksum verification failed"

migrations_table="$(pvnaive_psql_at --command "SELECT to_regclass('pvnaive.schema_migrations') IS NOT NULL")"
[[ "${migrations_table}" == "t" ]] || pvnaive_die "schema_migrations table is missing"
current_version="$(pvnaive_psql_at --command "SELECT COALESCE(MAX(version), 0) FROM pvnaive.schema_migrations")"
((current_version > 0)) || pvnaive_die "no applied migration exists"

mapfile -t applied_rows < <(pvnaive_psql_at --command "SELECT version::text || '|' || filename FROM pvnaive.schema_migrations ORDER BY version DESC")
((${#applied_rows[@]} == current_version)) || pvnaive_die "applied migration count does not match current version"

down_files=()
expected_version="${current_version}"
for row in "${applied_rows[@]}"; do
  IFS='|' read -r version filename <<< "${row}"
  [[ "${version}" =~ ^[1-9][0-9]*$ ]] || pvnaive_die "invalid applied migration version: ${version}"
  ((version == expected_version)) || pvnaive_die "applied migration sequence is not contiguous at version ${version}"
  [[ "${filename}" =~ ^[0-9]{4}_[a-z0-9_]+\.up\.sql$ ]] || pvnaive_die "invalid applied migration filename: ${filename}"
  version_text="$(printf '%04d' "${version}")"
  [[ "${filename:0:4}" == "${version_text}" ]] || pvnaive_die "applied migration filename/version mismatch: ${filename}"

  down_filename="${filename%.up.sql}.down.sql"
  down_file="${migrations_dir}/${down_filename}"
  [[ -f "${down_file}" ]] || pvnaive_die "down migration is missing: ${down_filename}"
  grep -Fqx -- "-- pvnaive:migration-version ${version_text}" "${down_file}" || pvnaive_die "down migration version mismatch: ${down_filename}"
  grep -Fqx -- "-- pvnaive:transactional true" "${down_file}" || pvnaive_die "non-transactional rollback refused: ${down_filename}"
  grep -Fqx -- "-- pvnaive:destructive true" "${down_file}" || pvnaive_die "rollback marker missing: ${down_filename}"
  down_entries="$(awk -v file="${down_filename}" '$2 == file {count++} END {print count+0}' "${manifest}")"
  [[ "${down_entries}" == "1" ]] || pvnaive_die "down migration must appear exactly once in SHA256SUMS: ${down_filename}"
  down_files+=("${down_file}")
  expected_version=$((expected_version - 1))
done
((expected_version == 0)) || pvnaive_die "rollback plan does not reach schema version zero"

if [[ "${PVNAIVE_DISPOSABLE_DB:-0}" == "1" && "${PVNAIVE_DB_NAME}" =~ ^pvnaive_(migration|restore)_test_[a-z0-9_]+$ ]]; then
  :
else
  [[ "${PVNAIVE_ALLOW_DESTRUCTIVE_ROLLBACK:-}" == "DROP_PVNAIVE_SCHEMA" ]] || \
    pvnaive_die "production rollback requires PVNAIVE_ALLOW_DESTRUCTIVE_ROLLBACK=DROP_PVNAIVE_SCHEMA"
  backup_file="${PVNAIVE_CONFIRMED_BACKUP:-}"
  identity_file="${PVNAIVE_BACKUP_IDENTITY_FILE:-/etc/pvnaive/backup.agekey}"
  [[ -f "${backup_file}" && -f "${identity_file}" ]] || pvnaive_die "verified encrypted backup and identity are required"
  backup_dir="$(dirname -- "${backup_file}")"
  [[ -f "${backup_dir}/SHA256SUMS" ]] || pvnaive_die "backup checksum manifest is missing"
  [[ -f "${backup_dir}/metadata.json" ]] || pvnaive_die "backup metadata is missing"
  grep -Eq '^[0-9a-f]{64}  metadata\.json$' "${backup_dir}/SHA256SUMS" || pvnaive_die "backup metadata is not checksummed"
  grep -Eq '^[0-9a-f]{64}  pvnaive\.dump\.age$' "${backup_dir}/SHA256SUMS" || pvnaive_die "backup archive is not checksummed"
  (cd "${backup_dir}" && sha256sum --check --strict SHA256SUMS) >/dev/null || pvnaive_die "backup checksum validation failed"
  grep -Fqx '  "product": "PVNaive",' "${backup_dir}/metadata.json" || pvnaive_die "backup metadata product mismatch"
  grep -Fqx '  "encrypted": true,' "${backup_dir}/metadata.json" || pvnaive_die "backup metadata encryption marker is missing"
  metadata_schema_version="$(sed -n 's/^  "schema_version": \([0-9][0-9]*\),$/\1/p' "${backup_dir}/metadata.json")"
  [[ "${metadata_schema_version}" == "${current_version}" ]] || pvnaive_die "backup schema version does not match rollback version"
  pvnaive_require_command age
  pvnaive_require_command pg_restore
  age --decrypt --identity "${identity_file}" "${backup_file}" | pg_restore --list >/dev/null || \
    pvnaive_die "encrypted backup could not be decrypted and parsed"
fi

echo "ROLLBACK_FROM_VERSION=$(printf '%04d' "${current_version}")"
echo "ROLLBACK_TO_VERSION=0000"
echo "ROLLBACK_STEPS=${#down_files[@]}"
echo "ROLLBACK_IS_DESTRUCTIVE=true"
{
  printf '%s\n' "SELECT pg_advisory_xact_lock(hashtext('pvnaive-schema-migrations'));"
  for down_file in "${down_files[@]}"; do
    cat -- "${down_file}"
    printf '\n'
  done
} | pvnaive_psql --single-transaction --file - >/dev/null

remaining_schema="$(pvnaive_psql_at --command "SELECT to_regnamespace('pvnaive') IS NOT NULL")"
remaining_crypto_schema="$(pvnaive_psql_at --command "SELECT to_regnamespace('pvnaive_crypto') IS NOT NULL")"
remaining_pgcrypto="$(pvnaive_psql_at --command "SELECT EXISTS (SELECT 1 FROM pg_extension WHERE extname='pgcrypto')")"
[[ "${remaining_schema}|${remaining_crypto_schema}|${remaining_pgcrypto}" == "f|f|f" ]] || \
  pvnaive_die "rollback verification failed: pvnaive=${remaining_schema} crypto_schema=${remaining_crypto_schema} pgcrypto=${remaining_pgcrypto}"

echo "PVNAIVE_ROLLBACK_RESULT=PASSED"
