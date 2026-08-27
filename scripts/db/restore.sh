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
grep -qx '  "ownership_and_acls": true,' "${backup_dir}/metadata.json" || pvnaive_die "restore metadata ownership/ACL marker is missing"
metadata_schema_version="$(sed -n 's/^  "schema_version": \([0-9][0-9]*\),$/\1/p' "${backup_dir}/metadata.json")"
[[ "${metadata_schema_version}" =~ ^[1-9][0-9]*$ ]] || pvnaive_die "restore metadata schema version is invalid"

# Ownership and ACL replay requires a PostgreSQL superuser. S03 authenticates as
# postgres over the local Unix socket and runs the client under the postgres OS
# account, but the check is role-based rather than name-based.
is_superuser="$(pvnaive_psql_at --dbname postgres --command "SELECT rolsuper FROM pg_roles WHERE rolname = current_user")"
[[ "${is_superuser}" == "t" ]] || pvnaive_die "restore connection must be a PostgreSQL superuser to preserve ownership and ACLs"

exists="$(pvnaive_psql_at --dbname postgres --command "SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = '${target_db}')")"
[[ "${exists}" == "f" ]] || pvnaive_die "restore target already exists"

target_created=0
cleanup() {
  local code="$1"
  trap - EXIT HUP INT TERM
  if [[ "${target_created}" == "1" ]]; then
    if ! pvnaive_admin_tool dropdb --if-exists --force "${target_db}" >/dev/null 2>&1; then
      echo "ERROR: cleanup could not drop restore target: ${target_db}" >&2
    fi
  fi
  exit "${code}"
}
trap 'cleanup "$?"' EXIT
trap 'exit 129' HUP
trap 'exit 130' INT
trap 'exit 143' TERM

age --decrypt --identity "${identity_file}" "${backup_file}" |
  pg_restore --list >/dev/null || pvnaive_die "restore archive parse failed"
target_created=1
pvnaive_admin_tool createdb --owner pvnaive_owner --encoding UTF8 --template template0 "${target_db}"
# Decrypt directly to pg_restore. No plaintext database archive is written to
# disk. The superuser connection intentionally preserves archive ownership and
# ACL entries and bypasses FORCE RLS while loading backed-up table data.
age --decrypt --identity "${identity_file}" "${backup_file}" |
  pvnaive_db_tool pg_restore --dbname "${target_db}" --exit-on-error --single-transaction >/dev/null ||
  pvnaive_die "restore execution failed"

restored_version="$(pvnaive_psql_at --dbname "${target_db}" --command 'SELECT COALESCE(MAX(version), 0) FROM pvnaive.schema_migrations')"
[[ "${restored_version}" == "${metadata_schema_version}" ]] || pvnaive_die "restored schema version does not match backup metadata"

restored_owner="$(pvnaive_psql_at --dbname "${target_db}" --command "SELECT pg_get_userbyid(c.relowner) FROM pg_class AS c JOIN pg_namespace AS n ON n.oid = c.relnamespace WHERE n.nspname = 'pvnaive' AND c.relname = 'users'")"
[[ "${restored_owner}" == "pvnaive_owner" ]] || pvnaive_die "restored object ownership mismatch"

app_schema_usage="$(pvnaive_psql_at --dbname "${target_db}" --command "SELECT has_schema_privilege('pvnaive_app', 'pvnaive', 'USAGE')")"
app_users_select="$(pvnaive_psql_at --dbname "${target_db}" --command "SELECT has_table_privilege('pvnaive_app', 'pvnaive.users', 'SELECT')")"
app_key_select="$(pvnaive_psql_at --dbname "${target_db}" --command "SELECT has_table_privilege('pvnaive_app', 'pvnaive.security_context_keys', 'SELECT')")"
app_context_execute="$(pvnaive_psql_at --dbname "${target_db}" --command "SELECT has_function_privilege('pvnaive_app', 'pvnaive.set_request_context(bytea)', 'EXECUTE')")"
[[ "${app_schema_usage}" == "t" ]] || pvnaive_die "restored pvnaive_app schema privilege is missing"
[[ "${app_users_select}" == "t" ]] || pvnaive_die "restored pvnaive_app table privilege is missing"
[[ "${app_key_select}" == "f" ]] || pvnaive_die "restored pvnaive_app can read the RLS signing key"
[[ "${app_context_execute}" == "t" ]] || pvnaive_die "restored pvnaive_app context function privilege is missing"

restored_pgcrypto="$(pvnaive_psql_at --dbname "${target_db}" --command "SELECT EXISTS (SELECT 1 FROM pg_extension WHERE extname = 'pgcrypto')")"
[[ "${restored_pgcrypto}" == "t" ]] || pvnaive_die "pgcrypto extension was not restored"

source_key_hash="$(pvnaive_psql_at --dbname "${PVNAIVE_DB_NAME}" --command "SELECT encode(digest(signing_key, 'sha256'), 'hex') FROM pvnaive.security_context_keys WHERE singleton")"
restored_key_hash="$(pvnaive_psql_at --dbname "${target_db}" --command "SELECT encode(digest(signing_key, 'sha256'), 'hex') FROM pvnaive.security_context_keys WHERE singleton")"
[[ -n "${source_key_hash}" && "${source_key_hash}" == "${restored_key_hash}" ]] || pvnaive_die "restored RLS signing key does not match source backup"
unset source_key_hash restored_key_hash

target_created=0
trap - EXIT HUP INT TERM
echo "PVNAIVE_RESTORE_RESULT=PASSED"
echo "PVNAIVE_RESTORE_TARGET=${target_db}"
echo "PVNAIVE_RESTORE_SCHEMA_VERSION=${restored_version}"
echo "PVNAIVE_RESTORE_OWNERSHIP=PASSED"
echo "PVNAIVE_RESTORE_ACLS=PASSED"
echo "PVNAIVE_RESTORE_SIGNING_KEY=PASSED"
