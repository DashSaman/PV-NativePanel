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
[[ -r "${recipient_file}" ]] || pvnaive_die "age recipient file is not readable"
recipient="$(tr -d '[:space:]' < "${recipient_file}")"
[[ "${recipient}" == age1* ]] || pvnaive_die "invalid age recipient"

stamp="$(date -u +%Y%m%dT%H%M%SZ)"
install -d -m 0700 "${backup_root}"
temp_dir="$(mktemp -d "${backup_root}/.tmp.${stamp}.XXXXXX")"
final_dir="${backup_root}/${stamp}"
cleanup() { rm -rf -- "${temp_dir}"; }
trap cleanup EXIT
[[ ! -e "${final_dir}" ]] || pvnaive_die "backup destination already exists"

pvnaive_db_tool pg_dump --format custom --compress 6 --no-owner --no-acl --dbname "${PVNAIVE_DB_NAME}" > "${temp_dir}/pvnaive.dump"
pg_restore --list "${temp_dir}/pvnaive.dump" >/dev/null || pvnaive_die "pg_dump archive validation failed"
age --recipient "${recipient}" --output "${temp_dir}/pvnaive.dump.age" "${temp_dir}/pvnaive.dump"
rm -f -- "${temp_dir}/pvnaive.dump"

schema_version="$(pvnaive_psql_at --command 'SELECT COALESCE(MAX(version), 0) FROM pvnaive.schema_migrations')"
((schema_version > 0)) || pvnaive_die "database has no applied PVNaive migration"
server_version="$(pvnaive_psql_at --command 'SHOW server_version')"
dump_size="$(stat -c '%s' "${temp_dir}/pvnaive.dump.age")"
printf '{\n  "product": "PVNaive",\n  "created_at_utc": "%s",\n  "database": "%s",\n  "schema_version": %s,\n  "postgres_version": "%s",\n  "encrypted": true,\n  "size_bytes": %s\n}\n' \
  "${stamp}" "${PVNAIVE_DB_NAME}" "${schema_version}" "${server_version}" "${dump_size}" > "${temp_dir}/metadata.json"
(
  cd "${temp_dir}"
  sha256sum metadata.json pvnaive.dump.age > SHA256SUMS
)
chmod 0600 "${temp_dir}"/*
mv -- "${temp_dir}" "${final_dir}"
trap - EXIT
echo "PVNAIVE_BACKUP_RESULT=PASSED"
echo "PVNAIVE_BACKUP_PATH=${final_dir}/pvnaive.dump.age"
