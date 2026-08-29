#!/usr/bin/env bash
set -Eeuo pipefail
umask 077

repo_root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd -P)"
test_suffix="${GITHUB_RUN_ID:-local}_${GITHUB_RUN_ATTEMPT:-1}_${BASHPID}"
test_suffix="${test_suffix//[^a-zA-Z0-9_]/_}"
test_db="pvnaive_backup_collision_${test_suffix,,}"
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
    "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname='${test_db}' AND pid <> pg_backend_pid()" >/dev/null 2>&1 || true
  dropdb --if-exists --host "${PVNAIVE_DB_HOST}" --port "${PVNAIVE_DB_PORT}" \
    --username "${PVNAIVE_DB_USER}" "${test_db}" >/dev/null 2>&1 || true
  psql_admin --dbname postgres --command \
    'DROP ROLE IF EXISTS pvnaive_app; DROP ROLE IF EXISTS pvnaive_owner;' >/dev/null 2>&1 || true
  rm -rf -- "${temp_root}"
}
trap cleanup EXIT
cleanup
mkdir -p "${temp_root}"

psql_admin --dbname postgres <<'SQL' >/dev/null
CREATE ROLE pvnaive_owner NOLOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOREPLICATION NOBYPASSRLS;
CREATE ROLE pvnaive_app NOLOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOREPLICATION NOBYPASSRLS;
SQL
createdb --host "${PVNAIVE_DB_HOST}" --port "${PVNAIVE_DB_PORT}" \
  --username "${PVNAIVE_DB_USER}" --owner pvnaive_owner --encoding UTF8 --template template0 "${test_db}"

export PVNAIVE_DB_NAME="${test_db}"
"${repo_root}/scripts/db/migrate.sh" >/dev/null

age-keygen -o "${temp_root}/backup.agekey" >/dev/null 2>&1
age-keygen -y "${temp_root}/backup.agekey" > "${temp_root}/backup.recipient"
mkdir -p "${temp_root}/fakebin"
cat >"${temp_root}/fakebin/date" <<'DATE'
#!/usr/bin/env bash
if [[ "$*" == '-u +%Y%m%dT%H%M%SZ' ]]; then
  printf '%s\n' '20260827T234607Z'
else
  exec /usr/bin/date "$@"
fi
DATE
chmod 0755 "${temp_root}/fakebin/date"

run_backup() {
  PATH="${temp_root}/fakebin:${PATH}" \
  PVNAIVE_BACKUP_ROOT="${temp_root}/backups" \
  PVNAIVE_BACKUP_RECIPIENT_FILE="${temp_root}/backup.recipient" \
  PVNAIVE_BACKUP_IDENTITY_FILE="${temp_root}/backup.agekey" \
    "${repo_root}/scripts/db/backup.sh"
}

first_output="$(run_backup)"
second_output="$(run_backup)"
first_path="$(awk -F= '/^PVNAIVE_BACKUP_PATH=/ {print $2}' <<<"${first_output}")"
second_path="$(awk -F= '/^PVNAIVE_BACKUP_PATH=/ {print $2}' <<<"${second_output}")"

[[ -f "${first_path}" ]] || { echo 'ERROR: first backup missing' >&2; exit 1; }
[[ -f "${second_path}" ]] || { echo 'ERROR: second backup missing' >&2; exit 1; }
[[ "${first_path}" != "${second_path}" ]] || { echo 'ERROR: same-second backups collided' >&2; exit 1; }

echo 'PVNAIVE_BACKUP_COLLISION_TEST=PASSED'
