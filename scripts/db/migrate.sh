#!/usr/bin/env bash
# Apply only checksum-verified, non-destructive PVNaive migrations.
set -Eeuo pipefail
umask 077

script_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd -P)"
repo_root="$(cd -- "${script_dir}/../.." && pwd -P)"
# shellcheck source=scripts/db/lib.sh
source "${script_dir}/lib.sh"

pvnaive_require_command psql
pvnaive_require_command sha256sum
pvnaive_require_command grep
pvnaive_require_command sort
pvnaive_require_command sed
pvnaive_require_command tr
[[ -z "${PVNAIVE_RUN_AS_OS_USER:-}" ]] || pvnaive_require_command runuser
pvnaive_db_defaults

migrations_dir="${PVNAIVE_MIGRATIONS_DIR:-${repo_root}/db/migrations}"
manifest="${migrations_dir}/SHA256SUMS"
[[ -d "${migrations_dir}" ]] || pvnaive_die "migration directory not found: ${migrations_dir}"
[[ -f "${manifest}" ]] || pvnaive_die "migration checksum manifest is missing"

(
  cd "${migrations_dir}"
  sha256sum --check --strict SHA256SUMS
) >/dev/null || pvnaive_die "migration checksum verification failed"

mapfile -t migration_files < <(find "${migrations_dir}" -maxdepth 1 -type f -name '[0-9][0-9][0-9][0-9]_*.up.sql' -print | LC_ALL=C sort)
((${#migration_files[@]} > 0)) || pvnaive_die "no versioned up migrations found"

# Validate the complete migration set before applying any file. This prevents a
# malformed later migration from leaving an earlier migration committed.
expected_version=1
for migration_file in "${migration_files[@]}"; do
  filename="$(basename -- "${migration_file}")"
  [[ "${filename}" =~ ^([0-9]{4})_([a-z0-9_]+)\.up\.sql$ ]] || pvnaive_die "invalid migration filename: ${filename}"
  version_text="${BASH_REMATCH[1]}"
  version=$((10#${version_text}))
  ((version == expected_version)) || pvnaive_die "migration sequence must be contiguous; expected $(printf '%04d' "${expected_version}"), found ${version_text}"
  expected_version=$((expected_version + 1))

  grep -qx -- "-- pvnaive:migration-version ${version_text}" "${migration_file}" || pvnaive_die "missing version header: ${filename}"
  grep -qx -- "-- pvnaive:transactional true" "${migration_file}" || pvnaive_die "non-transactional migration refused: ${filename}"
  grep -qx -- "-- pvnaive:destructive false" "${migration_file}" || pvnaive_die "destructive migration refused: ${filename}"

  normalized_sql="$(sed -E 's/--.*$//' "${migration_file}" | tr '\n' ' ')"
  if grep -Eiq '(^|[[:space:];])(DROP[[:space:]]+(TABLE|SCHEMA|DATABASE|ROLE|TYPE|INDEX|VIEW|MATERIALIZED[[:space:]]+VIEW|FUNCTION|PROCEDURE|TRIGGER|EXTENSION)|TRUNCATE([[:space:]]+TABLE)?|DELETE[[:space:]]+FROM|ALTER[[:space:]]+TABLE[^;]*[[:space:]]DROP[[:space:]]|COPY[^;]*[[:space:]]PROGRAM[[:space:]])' <<< "${normalized_sql}" || \
     grep -Eiq '(^|[[:space:];])\\(i|ir|include|include_relative)[[:space:]]' <<< "${normalized_sql}"; then
    pvnaive_die "destructive SQL pattern refused: ${filename}"
  fi

  checksum="$(sha256sum "${migration_file}" | awk '{print $1}')"
  [[ "${checksum}" =~ ^[0-9a-f]{64}$ ]] || pvnaive_die "invalid checksum: ${filename}"
  mapfile -t manifest_checksums < <(awk -v file="${filename}" '$2 == file {print $1}' "${manifest}")
  ((${#manifest_checksums[@]} == 1)) || pvnaive_die "migration must appear exactly once in SHA256SUMS: ${filename}"
  [[ "${manifest_checksums[0]}" == "${checksum}" ]] || pvnaive_die "manifest checksum mismatch: ${filename}"

  down_filename="${filename%.up.sql}.down.sql"
  [[ -f "${migrations_dir}/${down_filename}" ]] || pvnaive_die "down migration is missing: ${down_filename}"
  down_entries="$(awk -v file="${down_filename}" '$2 == file {count++} END {print count+0}' "${manifest}")"
  [[ "${down_entries}" == "1" ]] || pvnaive_die "down migration must appear exactly once in SHA256SUMS: ${down_filename}"
done

can_assume_owner="$(pvnaive_psql_at --command "SELECT pg_has_role(current_user, 'pvnaive_owner', 'USAGE')")"
[[ "${can_assume_owner}" == "t" ]] || pvnaive_die "migration connection cannot assume pvnaive_owner"

migrations_table="$(pvnaive_psql_at --command "SELECT to_regclass('pvnaive.schema_migrations') IS NOT NULL")"
if [[ "${migrations_table}" == "t" ]]; then
  current_version="$(pvnaive_psql_at --command 'SELECT COALESCE(MAX(version), 0) FROM pvnaive.schema_migrations')"
else
  current_version=0
fi

for migration_file in "${migration_files[@]}"; do
  filename="$(basename -- "${migration_file}")"
  [[ "${filename}" =~ ^([0-9]{4})_([a-z0-9_]+)\.up\.sql$ ]] || pvnaive_die "invalid migration filename: ${filename}"
  version_text="${BASH_REMATCH[1]}"
  version=$((10#${version_text}))
  checksum="$(sha256sum "${migration_file}" | awk '{print $1}')"

  if ((version <= current_version)); then
    stored="$(pvnaive_psql_at --command "SELECT filename || '|' || checksum_sha256 FROM pvnaive.schema_migrations WHERE version = ${version}")"
    [[ "${stored}" == "${filename}|${checksum}" ]] || pvnaive_die "immutable migration mismatch at version ${version_text}"
    echo "MIGRATION ${version_text}=ALREADY_APPLIED"
    continue
  fi
  ((version == current_version + 1)) || pvnaive_die "database migration gap before ${version_text}"

  echo "MIGRATION ${version_text}=APPLYING"
  {
    printf '%s\n' "SELECT pg_advisory_xact_lock(hashtext('pvnaive-schema-migrations'));"
    cat -- "${migration_file}"
    printf "\nINSERT INTO pvnaive.schema_migrations (version, filename, checksum_sha256, destructive, execution_ms) VALUES (%d, '%s', '%s', false, GREATEST(0, EXTRACT(MILLISECONDS FROM clock_timestamp() - transaction_timestamp())::bigint));\n" \
      "${version}" "${filename}" "${checksum}"
  } | pvnaive_psql --single-transaction --file - >/dev/null

  stored="$(pvnaive_psql_at --command "SELECT filename || '|' || checksum_sha256 FROM pvnaive.schema_migrations WHERE version = ${version}")"
  [[ "${stored}" == "${filename}|${checksum}" ]] || pvnaive_die "post-migration verification failed at ${version_text}"
  echo "MIGRATION ${version_text}=APPLIED"
  current_version="${version}"
done

database_version="$(pvnaive_psql_at --command 'SELECT COALESCE(MAX(version), 0) FROM pvnaive.schema_migrations')"
[[ "${database_version}" == "${current_version}" ]] || pvnaive_die "final schema version verification failed"
echo "PVNAIVE_SCHEMA_VERSION=${database_version}"
echo "PVNAIVE_MIGRATION_RESULT=PASSED"
