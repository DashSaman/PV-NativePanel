#!/usr/bin/env bash
set -Eeuo pipefail
umask 077

repo_root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd -P)"
test_suffix="${GITHUB_RUN_ID:-local}_${GITHUB_RUN_ATTEMPT:-1}_${BASHPID}"
test_suffix="${test_suffix//[^a-zA-Z0-9_]/_}"
test_db="pvnaive_migration_test_${test_suffix,,}"

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
    "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname='${test_db}' AND pid <> pg_backend_pid()" \
    >/dev/null 2>&1 || true
  dropdb --if-exists --host "${PVNAIVE_DB_HOST}" --port "${PVNAIVE_DB_PORT}" \
    --username "${PVNAIVE_DB_USER}" "${test_db}" >/dev/null 2>&1 || true
  psql_admin --dbname postgres --command \
    'DROP ROLE IF EXISTS pvnaive_app; DROP ROLE IF EXISTS pvnaive_owner;' >/dev/null 2>&1 || true
}
trap cleanup EXIT
cleanup

psql_admin --dbname postgres <<'SQL' >/dev/null
CREATE ROLE pvnaive_owner NOLOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOREPLICATION NOBYPASSRLS;
CREATE ROLE pvnaive_app NOLOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOREPLICATION NOBYPASSRLS;
SQL

createdb --host "${PVNAIVE_DB_HOST}" --port "${PVNAIVE_DB_PORT}" \
  --username "${PVNAIVE_DB_USER}" --owner pvnaive_owner --encoding UTF8 --template template0 "${test_db}"

export PVNAIVE_DB_NAME="${test_db}"
"${repo_root}/scripts/db/migrate.sh" >/dev/null
reapply_output="$("${repo_root}/scripts/db/migrate.sh")"
for version in 0001 0002 0003 0004 0005 0006 0007 0008 0009 0010 0011 0012 0013 0014 0015 0016 0017 0018 0019; do
  grep -Fqx "MIGRATION ${version}=ALREADY_APPLIED" <<< "${reapply_output}"
done
grep -Fqx 'PVNAIVE_SCHEMA_VERSION=19' <<< "${reapply_output}"
grep -Fqx 'PVNAIVE_MIGRATION_RESULT=PASSED' <<< "${reapply_output}"

psql_admin --dbname "${test_db}" <<'SQL' >/dev/null
INSERT INTO pvnaive.tenants (id, tenant_type, slug, display_name) VALUES
  ('10000000-0000-0000-0000-000000000001', 'reseller', 'tenant_a', 'Tenant A'),
  ('20000000-0000-0000-0000-000000000002', 'reseller', 'tenant_b', 'Tenant B');
INSERT INTO pvnaive.actors (id, tenant_id, actor_role, email, display_name, status) VALUES
  ('11000000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000001', 'reseller', 'a@example.invalid', 'Actor A', 'active'),
  ('22000000-0000-0000-0000-000000000002', '20000000-0000-0000-0000-000000000002', 'reseller', 'b@example.invalid', 'Actor B', 'active'),
  ('99000000-0000-0000-0000-000000000009', NULL, 'owner', 'owner@example.invalid', 'Owner', 'active');
INSERT INTO pvnaive.resellers (tenant_id, primary_actor_id, currency, credit_limit_minor) VALUES
  ('10000000-0000-0000-0000-000000000001', '11000000-0000-0000-0000-000000000001', 'EUR', 100000),
  ('20000000-0000-0000-0000-000000000002', '22000000-0000-0000-0000-000000000002', 'EUR', 100000);
INSERT INTO pvnaive.users (id, tenant_id, username, display_name, created_by_actor_id) VALUES
  ('11100000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000001', 'user_a', 'User A', '11000000-0000-0000-0000-000000000001'),
  ('22200000-0000-0000-0000-000000000002', '20000000-0000-0000-0000-000000000002', 'user_b', 'User B', '22000000-0000-0000-0000-000000000002');
INSERT INTO pvnaive.auth_sessions (tenant_id, actor_id, token_hash, refresh_family_id, expires_at) VALUES
  ('10000000-0000-0000-0000-000000000001', '11000000-0000-0000-0000-000000000001',
   public.digest('session-a', 'sha256'), '11110000-0000-0000-0000-000000000001', clock_timestamp() + interval '1 hour');
INSERT INTO pvnaive.audit_events (tenant_id, actor_id, action, object_type, outcome) VALUES
  ('10000000-0000-0000-0000-000000000001', '11000000-0000-0000-0000-000000000001', 'test.create', 'user', 'success');
SQL

spoofed_output="$(psql_admin --dbname "${test_db}" --tuples-only --no-align <<'SQL'
BEGIN;
SET LOCAL ROLE pvnaive_app;
SELECT set_config('pvnaive.tenant_id', '10000000-0000-0000-0000-000000000001', true);
SELECT 'UNSIGNED_VALID=' || pvnaive.has_valid_context();
SELECT 'UNSIGNED_COUNT=' || COUNT(*) FROM pvnaive.users;
ROLLBACK;
SQL
)"
[[ "$(sed -n 's/^UNSIGNED_VALID=//p' <<< "${spoofed_output}")" =~ ^(false|f)$ ]] || {
  echo 'ERROR: unsigned context validated' >&2
  exit 1
}
[[ "$(sed -n 's/^UNSIGNED_COUNT=//p' <<< "${spoofed_output}")" == "0" ]] || {
  echo 'ERROR: unsigned context bypassed RLS' >&2
  exit 1
}

signed_output="$(psql_admin --dbname "${test_db}" --tuples-only --no-align <<'SQL'
BEGIN;
SET LOCAL ROLE pvnaive_app;
SELECT pvnaive.set_request_context(public.digest('session-a', 'sha256'));
SELECT 'SIGNED_VALID=' || pvnaive.has_valid_context();
SELECT 'SIGNED_VISIBLE=' || COUNT(*) FROM pvnaive.users;
SELECT 'SIGNED_CROSS=' || COUNT(*) FROM pvnaive.users WHERE id='22200000-0000-0000-0000-000000000002';
ROLLBACK;
SQL
)"
[[ "$(sed -n 's/^SIGNED_VALID=//p' <<< "${signed_output}")" =~ ^(true|t)$ ]] || {
  echo 'ERROR: signed context invalid' >&2
  exit 1
}
[[ "$(sed -n 's/^SIGNED_VISIBLE=//p' <<< "${signed_output}")|$(sed -n 's/^SIGNED_CROSS=//p' <<< "${signed_output}")" == "1|0" ]] || {
  echo 'ERROR: tenant isolation failed' >&2
  exit 1
}

if psql_admin --dbname "${test_db}" --command 'DELETE FROM pvnaive.audit_events' >/dev/null 2>&1; then
  echo 'ERROR: append-only audit ledger accepted DELETE' >&2
  exit 1
fi

# Destructive SQL in the next contiguous migration (20) must fail closed.
temp_migrations="$(mktemp -d)"
cp -a "${repo_root}/db/migrations/." "${temp_migrations}/"
printf '%s\n' \
  '-- pvnaive:migration-version 0020' \
  '-- pvnaive:migration-name forbidden_drop' \
  '-- pvnaive:transactional true' \
  '-- pvnaive:destructive false' \
  'DROP TABLE pvnaive.users;' > "${temp_migrations}/0020_forbidden_drop.up.sql"
printf '%s\n' \
  '-- pvnaive:migration-version 0020' \
  '-- pvnaive:transactional true' \
  '-- pvnaive:destructive true' \
  'SELECT 1;' > "${temp_migrations}/0020_forbidden_drop.down.sql"
(
  cd "${temp_migrations}"
  sha256sum *.sql > SHA256SUMS
)
if PVNAIVE_MIGRATIONS_DIR="${temp_migrations}" "${repo_root}/scripts/db/migrate.sh" >/dev/null 2>&1; then
  echo 'ERROR: destructive migration scan did not fail closed' >&2
  exit 1
fi
rm -rf -- "${temp_migrations}"

# An unlisted next migration (20) must fail checksum-manifest validation.
temp_migrations="$(mktemp -d)"
cp -a "${repo_root}/db/migrations/." "${temp_migrations}/"
printf '%s\n' \
  '-- pvnaive:migration-version 0020' \
  '-- pvnaive:migration-name unlisted_file' \
  '-- pvnaive:transactional true' \
  '-- pvnaive:destructive false' \
  'SELECT 1;' > "${temp_migrations}/0020_unlisted_file.up.sql"
printf '%s\n' \
  '-- pvnaive:migration-version 0020' \
  '-- pvnaive:transactional true' \
  '-- pvnaive:destructive true' \
  'SELECT 1;' > "${temp_migrations}/0020_unlisted_file.down.sql"
if PVNAIVE_MIGRATIONS_DIR="${temp_migrations}" "${repo_root}/scripts/db/migrate.sh" >/dev/null 2>&1; then
  echo 'ERROR: unlisted migration was accepted' >&2
  exit 1
fi
rm -rf -- "${temp_migrations}"

# A version gap from 19 to 21 must fail before executing SQL.
temp_migrations="$(mktemp -d)"
cp -a "${repo_root}/db/migrations/." "${temp_migrations}/"
printf '%s\n' \
  '-- pvnaive:migration-version 0021' \
  '-- pvnaive:migration-name version_gap' \
  '-- pvnaive:transactional true' \
  '-- pvnaive:destructive false' \
  'SELECT 1;' > "${temp_migrations}/0021_version_gap.up.sql"
printf '%s\n' \
  '-- pvnaive:migration-version 0021' \
  '-- pvnaive:transactional true' \
  '-- pvnaive:destructive true' \
  'SELECT 1;' > "${temp_migrations}/0021_version_gap.down.sql"
(
  cd "${temp_migrations}"
  sha256sum *.sql > SHA256SUMS
)
if PVNAIVE_MIGRATIONS_DIR="${temp_migrations}" "${repo_root}/scripts/db/migrate.sh" >/dev/null 2>&1; then
  echo 'ERROR: non-contiguous migration version was accepted' >&2
  exit 1
fi
rm -rf -- "${temp_migrations}"

# Applied migration immutability includes the newest repository migration.
expected_checksum="$(sha256sum "${repo_root}/db/migrations/0019_concurrent_session_limit.up.sql" | awk '{print $1}')"
psql_admin --dbname "${test_db}" --command \
  "UPDATE pvnaive.schema_migrations SET checksum_sha256=repeat('0',64) WHERE version=19" >/dev/null
if "${repo_root}/scripts/db/migrate.sh" >/dev/null 2>&1; then
  echo 'ERROR: changed applied migration checksum was accepted' >&2
  exit 1
fi
psql_admin --dbname "${test_db}" --command \
  "UPDATE pvnaive.schema_migrations SET checksum_sha256='${expected_checksum}' WHERE version=19" >/dev/null

for expected in 18 17 16 15 14 13 12 11 10 9 8 7 6 5 4 3 2 1; do
  PVNAIVE_DISPOSABLE_DB=1 "${repo_root}/scripts/db/rollback.sh" >/dev/null
  actual="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command 'SELECT COALESCE(MAX(version),0) FROM pvnaive.schema_migrations')"
  [[ "${actual}" == "${expected}" ]] || {
    echo "ERROR: rollback expected schema ${expected}, got ${actual}" >&2
    exit 1
  }
done

PVNAIVE_DISPOSABLE_DB=1 "${repo_root}/scripts/db/rollback.sh" >/dev/null
schema_exists="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command "SELECT to_regnamespace('pvnaive') IS NOT NULL")"
[[ "${schema_exists}" == "false" || "${schema_exists}" == "f" ]] || {
  echo 'ERROR: v1 rollback left schema behind' >&2
  exit 1
}

echo 'PVNAIVE_DB_MIGRATION_TEST=PASSED'
