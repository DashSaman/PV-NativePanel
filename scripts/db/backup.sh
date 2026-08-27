#!/usr/bin/env bash
# Create an encrypted and checksum-verified PVNaive database backup.
set -Eeuo pipefail
umask 077

script_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd -P)"
# shellcheck source=scripts/db/lib.sh
source "${script_dir}/lib.sh"

for command_name in pg_dump pg_restore age sha256sum; do pvnaive_require_command "${command_name}"; done
[[ -z "${PVNAIVE_RUN_AS_OS_USER:-}" ]] || pvnaive_require_command runuser
pvnaive_db_defaults

backup_root="${PVNAIVE_BACKUP_ROOT:-/var/backups/pvnaive/database}"
recipient_file="${PVNAIVE_BACKUP_RECIPIENT_FILE:-/etc/pvnaive/backup.recipient}"
identity_file="${PVNAIVE_BACKUP_IDENTITY_FILE:-/etc/pvnaive/backup.agekey}"
[[ -r "${recipient_file}" ]] || pvnaive_die "age recipient file is not readable"
[[ -r "${identity_file}" ]] || pvnaive_die "age identity file is not readable"
recipient="$(tr -d '[:space:]' < "${recipient_file}")"
[[ "${recipient}" == age1* ]] || pvnaive_die "invalid age recipient"
pvnaive_validate_storage_root "${backup_root}"

stamp="$(date -u +%Y%m%dT%H%M%SZ)"
install -d -m 0700 "${backup_root}"
temp_dir=""
final_dir="${backup_root}/${stamp}"
cleanup() {
  local code="$1"
  trap - EXIT HUP INT TERM
  if [[ -n "${temp_dir}" && -d "${temp_dir}" ]]; then
    rm -rf -- "${temp_dir}" || true
  fi
  exit "${code}"
}
trap 'cleanup "$?"' EXIT
trap 'exit 129' HUP
trap 'exit 130' INT
trap 'exit 143' TERM
temp_dir="$(mktemp -d "${backup_root}/.tmp.${stamp}.XXXXXX")"
[[ ! -e "${final_dir}" ]] || pvnaive_die "backup destination already exists"

# Preserve ownership and ACL information. Restore drills run as PostgreSQL
# superuser so the archive can reconstruct pvnaive_owner ownership and the
# pvnaive_app GRANT/REVOKE boundary exactly.
pvnaive_db_tool pg_dump --format custom --compress=6 --dbname "${PVNAIVE_DB_NAME}" |
  age --recipient "${recipient}" --output "${temp_dir}/pvnaive.dump.age" ||
  pvnaive_die "encrypted pg_dump stream failed"
age --decrypt --identity "${identity_file}" "${temp_dir}/pvnaive.dump.age" |
  pg_restore --list >/dev/null || pvnaive_die "encrypted pg_dump archive validation failed"

schema_version="$(pvnaive_psql_at --command 'SELECT COALESCE(MAX(version), 0) FROM pvnaive.schema_migrations')"
((schema_version > 0)) || pvnaive_die "database has no applied PVNaive migration"
server_version="$(pvnaive_psql_at --command 'SHOW server_version')"
dump_size="$(stat -c '%s' "${temp_dir}/pvnaive.dump.age")"
printf '{\n  "product": "PVNaive",\n  "created_at_utc": "%s",\n  "database": "%s",\n  "schema_version": %s,\n  "postgres_version": "%s",\n  "encrypted": true,\n  "ownership_and_acls": true,\n  "size_bytes": %s\n}\n' \
  "${stamp}" "${PVNAIVE_DB_NAME}" "${schema_version}" "${server_version}" "${dump_size}" > "${temp_dir}/metadata.json"
(
  cd "${temp_dir}"
  sha256sum metadata.json pvnaive.dump.age > SHA256SUMS
)
chmod 0600 "${temp_dir}"/*
mv -- "${temp_dir}" "${final_dir}"
temp_dir=""
trap - EXIT HUP INT TERM
echo "PVNAIVE_BACKUP_RESULT=PASSED"
echo "PVNAIVE_BACKUP_PATH=${final_dir}/pvnaive.dump.age"
