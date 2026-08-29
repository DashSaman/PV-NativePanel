#!/usr/bin/env bash
set -Eeuo pipefail
umask 077

expected_host="testAmir5-3"
marker="/opt/pvnaive/S03_DATABASE.json"
db_env="/etc/pvnaive/db.env"
helper="/opt/pvnaive/bin/pvnaive-password"
tty="/dev/tty"

fail() {
  printf 'ERROR: %s\n' "$1" >&2
  exit 1
}

sql_quote() {
  local value="$1"
  value="${value//\'/\'\'}"
  printf "'%s'" "${value}"
}

[[ ${EUID} -eq 0 ]] || fail "run as root"
[[ "$(hostname)" == "${expected_host}" ]] || fail "unexpected host"
[[ -f "${marker}" ]] || fail "S03_DATABASE.json is missing"
[[ -r "${db_env}" ]] || fail "database environment is missing"
[[ -c "${tty}" ]] || fail "/dev/tty is required"
[[ -x "${helper}" ]] || fail "pvnaive-password helper is missing"
command -v runuser >/dev/null 2>&1 || fail "runuser is missing"
command -v psql >/dev/null 2>&1 || fail "psql is missing"

grep -Fqx '  "stage": "S03-DATABASE",' "${marker}" || fail "invalid S03 marker"
grep -Fqx '  "host": "testAmir5-3",' "${marker}" || fail "S03 marker host mismatch"

set -a
# shellcheck disable=SC1090
source "${db_env}"
set +a

[[ "${PVNAIVE_DB_HOST:-}" == "127.0.0.1" ]] || fail "database host is not IPv4 loopback"
[[ "${PVNAIVE_DB_NAME:-}" == "pvnaive" ]] || fail "unexpected database name"
[[ "${PVNAIVE_DB_PORT:-}" =~ ^[0-9]+$ ]] || fail "invalid database port"

postgres_psql() {
  runuser -u postgres -- psql --no-psqlrc --set ON_ERROR_STOP=1 \
    --host /var/run/postgresql --port "${PVNAIVE_DB_PORT}" --username postgres "$@"
}

schema_version="$(postgres_psql --dbname "${PVNAIVE_DB_NAME}" --tuples-only --no-align \
  --command 'SELECT COALESCE(MAX(version), 0) FROM pvnaive.schema_migrations')"
[[ "${schema_version}" == "2" ]] || fail "schema_migrations must be at version 2"

owner_count="$(postgres_psql --dbname "${PVNAIVE_DB_NAME}" --tuples-only --no-align \
  --command "SELECT COUNT(*) FROM pvnaive.actors WHERE actor_role = 'owner'")"
[[ "${owner_count}" == "0" ]] || fail "owner already exists"

printf 'Owner email: ' >"${tty}"
IFS= read -r owner_email <"${tty}"
printf 'Owner display name: ' >"${tty}"
IFS= read -r owner_display <"${tty}"

owner_email="${owner_email,,}"
[[ "${owner_email}" =~ ^[^[:space:]@]+@[^[:space:]@]+\.[^[:space:]@]+$ ]] || fail "invalid owner email"
((${#owner_email} >= 3 && ${#owner_email} <= 320)) || fail "owner email length is invalid"
((${#owner_display} >= 1 && ${#owner_display} <= 160)) || fail "owner display name length is invalid"

printf 'Owner password (minimum 14 characters): ' >"${tty}"
IFS= read -rs secret_one <"${tty}"
printf '\nConfirm owner password: ' >"${tty}"
IFS= read -rs secret_two <"${tty}"
printf '\n' >"${tty}"

[[ "${secret_one}" == "${secret_two}" ]] || {
  secret_one=
  secret_two=
  fail "password confirmation does not match"
}
((${#secret_one} >= 14)) || {
  secret_one=
  secret_two=
  fail "password is shorter than policy"
}

phc_hash="$(printf '%s\n' "${secret_one}" | "${helper}")" || {
  secret_one=
  secret_two=
  fail "pvnaive-password failed"
}
secret_one=
secret_two=
[[ "${phc_hash}" == '$argon2id$'* ]] || fail "password helper returned an invalid hash"

email_sql="$(sql_quote "${owner_email}")"
display_sql="$(sql_quote "${owner_display}")"
hash_sql="$(sql_quote "${phc_hash}")"
phc_hash=

sql_file="$(mktemp /run/pvnaive-owner-bootstrap.XXXXXX.sql)"
cleanup() {
  rm -f -- "${sql_file:-}"
}
trap cleanup EXIT HUP INT TERM
chmod 0600 "${sql_file}"
cat >"${sql_file}" <<SQL
BEGIN;
SELECT pg_advisory_xact_lock(hashtext('pvnaive-owner-bootstrap'));
SET LOCAL ROLE pvnaive_owner;
DO \$block\$
BEGIN
  IF EXISTS (SELECT 1 FROM pvnaive.actors WHERE actor_role = 'owner') THEN
    RAISE EXCEPTION 'owner already exists' USING ERRCODE = '23505';
  END IF;
END
\$block\$;
INSERT INTO pvnaive.actors (
  tenant_id, actor_role, email, display_name, password_hash, mfa_required, status, password_changed_at
) VALUES (
  NULL, 'owner', ${email_sql}, ${display_sql}, ${hash_sql}, false, 'active', clock_timestamp()
);
COMMIT;
SQL
unset email_sql display_sql hash_sql

# The file is created and populated by root, but psql below deliberately runs
# as the postgres OS user. Hand ownership to postgres only after the complete
# SQL payload has been written, while retaining mode 0600 so no other account
# can read the password hash or bootstrap data.
chown postgres:postgres "${sql_file}"
chmod 0600 "${sql_file}"

postgres_psql --dbname "${PVNAIVE_DB_NAME}" --file "${sql_file}" >/dev/null
rm -f -- "${sql_file}"
trap - EXIT HUP INT TERM

owner_count="$(postgres_psql --dbname "${PVNAIVE_DB_NAME}" --tuples-only --no-align \
  --command "SELECT COUNT(*) FROM pvnaive.actors WHERE actor_role = 'owner' AND status = 'active' AND password_hash IS NOT NULL")"
[[ "${owner_count}" == "1" ]] || fail "owner verification failed"

echo "PVNAIVE_OWNER_BOOTSTRAP=PASSED"
echo "OWNER_EMAIL=${owner_email}"
