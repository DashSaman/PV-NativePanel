#!/usr/bin/env bash
set -Eeuo pipefail
umask 077

repo_root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd -P)"
test_suffix="${GITHUB_RUN_ID:-local}_${GITHUB_RUN_ATTEMPT:-1}"
test_suffix="${test_suffix//[^a-zA-Z0-9_]/_}"
test_db="pvnaive_migration_test_${test_suffix,,}"

: "${PVNAIVE_DB_HOST:=127.0.0.1}"
: "${PVNAIVE_DB_PORT:=5432}"
: "${PVNAIVE_DB_USER:=postgres}"
export PVNAIVE_DB_HOST PVNAIVE_DB_PORT PVNAIVE_DB_USER

psql_admin() {
  psql --no-psqlrc --set ON_ERROR_STOP=1 --host "${PVNAIVE_DB_HOST}" --port "${PVNAIVE_DB_PORT}" --username "${PVNAIVE_DB_USER}" "$@"
}

cleanup() {
  psql_admin --dbname postgres --command "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = '${test_db}' AND pid <> pg_backend_pid()" >/dev/null 2>&1 || true
  dropdb --if-exists --host "${PVNAIVE_DB_HOST}" --port "${PVNAIVE_DB_PORT}" --username "${PVNAIVE_DB_USER}" "${test_db}" >/dev/null 2>&1 || true
  psql_admin --dbname postgres --command 'DROP ROLE IF EXISTS pvnaive_app; DROP ROLE IF EXISTS pvnaive_owner;' >/dev/null 2>&1 || true
}
trap cleanup EXIT
cleanup

psql_admin --dbname postgres <<'SQL' >/dev/null
CREATE ROLE pvnaive_owner NOLOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOREPLICATION NOBYPASSRLS;
CREATE ROLE pvnaive_app NOLOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOREPLICATION NOBYPASSRLS;
SQL
createdb --host "${PVNAIVE_DB_HOST}" --port "${PVNAIVE_DB_PORT}" --username "${PVNAIVE_DB_USER}" --owner pvnaive_owner --encoding UTF8 --template template0 "${test_db}"

export PVNAIVE_DB_NAME="${test_db}"
"${repo_root}/scripts/db/migrate.sh"
"${repo_root}/scripts/db/migrate.sh" | grep -q 'MIGRATION 0001=ALREADY_APPLIED'

psql_admin --dbname "${test_db}" <<'SQL' >/dev/null
INSERT INTO pvnaive.tenants (id, tenant_type, slug, display_name) VALUES
  ('10000000-0000-0000-0000-000000000001', 'reseller', 'tenant_a', 'Tenant A'),
  ('20000000-0000-0000-0000-000000000002', 'reseller', 'tenant_b', 'Tenant B');
INSERT INTO pvnaive.actors (id, tenant_id, actor_role, email, display_name, status) VALUES
  ('11000000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000001', 'reseller', 'a@example.invalid', 'Actor A', 'active'),
  ('22000000-0000-0000-0000-000000000002', '20000000-0000-0000-0000-000000000002', 'reseller', 'b@example.invalid', 'Actor B', 'active');
INSERT INTO pvnaive.actors (id, tenant_id, actor_role, email, display_name, status) VALUES
  ('99000000-0000-0000-0000-000000000009', NULL, 'owner', 'owner@example.invalid', 'Owner', 'active');
INSERT INTO pvnaive.resellers (tenant_id, primary_actor_id, currency, credit_limit_minor) VALUES
  ('10000000-0000-0000-0000-000000000001', '11000000-0000-0000-0000-000000000001', 'EUR', 100000),
  ('20000000-0000-0000-0000-000000000002', '22000000-0000-0000-0000-000000000002', 'EUR', 100000);
INSERT INTO pvnaive.users (id, tenant_id, username, display_name, created_by_actor_id) VALUES
  ('11100000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000001', 'user_a', 'User A', '11000000-0000-0000-0000-000000000001'),
  ('22200000-0000-0000-0000-000000000002', '20000000-0000-0000-0000-000000000002', 'user_b', 'User B', '22000000-0000-0000-0000-000000000002');
INSERT INTO pvnaive.auth_sessions (tenant_id, actor_id, token_hash, refresh_family_id, expires_at) VALUES
  ('10000000-0000-0000-0000-000000000001', '11000000-0000-0000-0000-000000000001', digest('session-a', 'sha256'), '11110000-0000-0000-0000-000000000001', clock_timestamp() + interval '1 hour');
INSERT INTO pvnaive.audit_events (tenant_id, actor_id, action, object_type, outcome) VALUES
  ('10000000-0000-0000-0000-000000000001', '11000000-0000-0000-0000-000000000001', 'test.create', 'user', 'success');
INSERT INTO pvnaive.plans (id, tenant_id, code, name, duration_seconds, base_price_minor, currency) VALUES
  ('33000000-0000-0000-0000-000000000003', NULL, 'test_plan', 'Test Plan', 2592000, 1000, 'EUR');
INSERT INTO pvnaive.reseller_plan_terms (tenant_id, plan_id, allowed, price_minor) VALUES
  ('10000000-0000-0000-0000-000000000001', '33000000-0000-0000-0000-000000000003', false, 900);
SQL

spoofed_count="$(psql_admin --dbname "${test_db}" --tuples-only --no-align <<'SQL'
BEGIN;
SET LOCAL ROLE pvnaive_app;
SELECT set_config('pvnaive.tenant_id', '10000000-0000-0000-0000-000000000001', true);
SELECT COUNT(*) FROM pvnaive.users;
ROLLBACK;
SQL
)"
[[ "$(echo "${spoofed_count}" | sed -n '2p')" == "0" ]] || { echo "ERROR: unsigned tenant context bypassed RLS" >&2; exit 1; }

forged_owner_count="$(psql_admin --dbname "${test_db}" --tuples-only --no-align <<'SQL'
BEGIN;
SET LOCAL ROLE pvnaive_app;
SELECT set_config('pvnaive.actor_id', '99000000-0000-0000-0000-000000000009', true);
SELECT set_config('pvnaive.actor_role', 'owner', true);
SELECT set_config('pvnaive.context_signature', repeat('0', 64), true);
SELECT COUNT(*) FROM pvnaive.users;
ROLLBACK;
SQL
)"
[[ "$(echo "${forged_owner_count}" | tail -n 1)" == "0" ]] || { echo "ERROR: forged owner context bypassed RLS" >&2; exit 1; }

scoped_counts="$(psql_admin --dbname "${test_db}" --tuples-only --no-align <<'SQL'
BEGIN;
SET LOCAL ROLE pvnaive_app;
SELECT pvnaive.set_request_context(digest('session-a', 'sha256'));
SELECT COUNT(*) FROM pvnaive.users;
SELECT COUNT(*) FROM pvnaive.users WHERE id = '22200000-0000-0000-0000-000000000002';
ROLLBACK;
SQL
)"
mapfile -t numeric_rows < <(echo "${scoped_counts}" | grep -E '^[0-9]+$')
[[ "${numeric_rows[*]}" == "1 0" ]] || { echo "ERROR: signed tenant isolation failed: ${numeric_rows[*]}" >&2; exit 1; }

if psql_admin --dbname "${test_db}" <<'SQL' >/dev/null 2>&1
BEGIN;
SET LOCAL ROLE pvnaive_app;
SELECT pvnaive.set_request_context(digest('session-a', 'sha256'));
INSERT INTO pvnaive.users (tenant_id, username, display_name, created_by_actor_id)
VALUES ('20000000-0000-0000-0000-000000000002', 'cross_tenant', 'Cross Tenant', '22000000-0000-0000-0000-000000000002');
COMMIT;
SQL
then
  echo "ERROR: cross-tenant INSERT bypassed RLS" >&2
  exit 1
fi

if psql_admin --dbname "${test_db}" <<'SQL' >/dev/null 2>&1
BEGIN;
SET LOCAL ROLE pvnaive_app;
SELECT pvnaive.set_request_context(digest('session-a', 'sha256'));
UPDATE pvnaive.auth_sessions
   SET actor_id = '99000000-0000-0000-0000-000000000009'
 WHERE token_hash = digest('session-a', 'sha256');
COMMIT;
SQL
then
  echo "ERROR: session actor/tenant privilege escalation was accepted" >&2
  exit 1
fi

psql_admin --dbname "${test_db}" <<'SQL' >/dev/null
BEGIN;
SET LOCAL ROLE pvnaive_app;
SELECT pvnaive.set_request_context(digest('session-a', 'sha256'));
INSERT INTO pvnaive.reseller_credit_ledger
  (tenant_id, delta_minor, balance_after_minor, currency, entry_type, reason_code, idempotency_key, created_by_actor_id)
VALUES
  ('10000000-0000-0000-0000-000000000001', 1000, 1000, 'EUR', 'deposit', 'test', 'credit-test-0001', '11000000-0000-0000-0000-000000000001');
COMMIT;
SQL

if psql_admin --dbname "${test_db}" <<'SQL' >/dev/null 2>&1
BEGIN;
SET LOCAL ROLE pvnaive_app;
SELECT pvnaive.set_request_context(digest('session-a', 'sha256'));
INSERT INTO pvnaive.purchases
  (tenant_id, user_id, plan_id, credit_ledger_entry_id, price_minor, currency, idempotency_key, created_by_actor_id)
SELECT '10000000-0000-0000-0000-000000000001',
       '11100000-0000-0000-0000-000000000001',
       '33000000-0000-0000-0000-000000000003',
       id, 900, 'EUR', 'purchase-test-0001',
       '11000000-0000-0000-0000-000000000001'
  FROM pvnaive.reseller_credit_ledger
 WHERE tenant_id = '10000000-0000-0000-0000-000000000001';
COMMIT;
SQL
then
  echo "ERROR: purchase using a disallowed reseller plan was accepted" >&2
  exit 1
fi

if psql_admin --dbname "${test_db}" --command 'DELETE FROM pvnaive.reseller_credit_ledger' >/dev/null 2>&1; then
  echo "ERROR: append-only reseller credit ledger accepted DELETE" >&2
  exit 1
fi

if psql_admin --dbname "${test_db}" --command 'DELETE FROM pvnaive.audit_events' >/dev/null 2>&1; then
  echo "ERROR: append-only audit ledger accepted DELETE" >&2
  exit 1
fi

temp_migrations="$(mktemp -d)"
cp -a "${repo_root}/db/migrations/." "${temp_migrations}/"
printf '%s\n' \
  '-- pvnaive:migration-version 0002' \
  '-- pvnaive:migration-name forbidden_drop' \
  '-- pvnaive:transactional true' \
  '-- pvnaive:destructive false' \
  'DROP TABLE pvnaive.users;' > "${temp_migrations}/0002_forbidden_drop.up.sql"
(
  cd "${temp_migrations}"
  sha256sum 0001_initial.down.sql 0001_initial.up.sql 0002_forbidden_drop.up.sql > SHA256SUMS
)
if PVNAIVE_MIGRATIONS_DIR="${temp_migrations}" "${repo_root}/scripts/db/migrate.sh" >/dev/null 2>&1; then
  echo "ERROR: destructive migration scan did not fail closed" >&2
  exit 1
fi
rm -rf -- "${temp_migrations}"

temp_migrations="$(mktemp -d)"
cp -a "${repo_root}/db/migrations/." "${temp_migrations}/"
printf '%s\n' \
  '-- pvnaive:migration-version 0002' \
  '-- pvnaive:migration-name unlisted_file' \
  '-- pvnaive:transactional true' \
  '-- pvnaive:destructive false' \
  'SELECT 1;' > "${temp_migrations}/0002_unlisted_file.up.sql"
printf '%s\n' \
  '-- pvnaive:migration-version 0002' \
  '-- pvnaive:transactional true' \
  '-- pvnaive:destructive true' \
  'SELECT 1;' > "${temp_migrations}/0002_unlisted_file.down.sql"
if PVNAIVE_MIGRATIONS_DIR="${temp_migrations}" "${repo_root}/scripts/db/migrate.sh" >/dev/null 2>&1; then
  echo "ERROR: migration omitted from checksum manifest was accepted" >&2
  exit 1
fi
rm -rf -- "${temp_migrations}"

temp_migrations="$(mktemp -d)"
cp -a "${repo_root}/db/migrations/." "${temp_migrations}/"
printf '%s\n' \
  '-- pvnaive:migration-version 0003' \
  '-- pvnaive:migration-name version_gap' \
  '-- pvnaive:transactional true' \
  '-- pvnaive:destructive false' \
  'SELECT 1;' > "${temp_migrations}/0003_version_gap.up.sql"
printf '%s\n' \
  '-- pvnaive:migration-version 0003' \
  '-- pvnaive:transactional true' \
  '-- pvnaive:destructive true' \
  'SELECT 1;' > "${temp_migrations}/0003_version_gap.down.sql"
(
  cd "${temp_migrations}"
  sha256sum 0001_initial.down.sql 0001_initial.up.sql 0003_version_gap.down.sql 0003_version_gap.up.sql > SHA256SUMS
)
if PVNAIVE_MIGRATIONS_DIR="${temp_migrations}" "${repo_root}/scripts/db/migrate.sh" >/dev/null 2>&1; then
  echo "ERROR: non-contiguous migration version was accepted" >&2
  exit 1
fi
rm -rf -- "${temp_migrations}"

expected_checksum="$(sha256sum "${repo_root}/db/migrations/0001_initial.up.sql" | awk '{print $1}')"
psql_admin --dbname "${test_db}" --command "UPDATE pvnaive.schema_migrations SET checksum_sha256 = repeat('0', 64) WHERE version = 1" >/dev/null
if "${repo_root}/scripts/db/migrate.sh" >/dev/null 2>&1; then
  echo "ERROR: changed applied migration checksum was accepted" >&2
  exit 1
fi
psql_admin --dbname "${test_db}" --command "UPDATE pvnaive.schema_migrations SET checksum_sha256 = '${expected_checksum}' WHERE version = 1" >/dev/null

PVNAIVE_DISPOSABLE_DB=1 "${repo_root}/scripts/db/rollback.sh"
schema_exists="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command "SELECT to_regnamespace('pvnaive') IS NOT NULL")"
[[ "${schema_exists}" == "f" ]] || { echo "ERROR: disposable rollback left schema behind" >&2; exit 1; }

echo "PVNAIVE_DB_MIGRATION_TEST=PASSED"
