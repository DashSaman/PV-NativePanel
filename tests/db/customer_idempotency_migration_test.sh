#!/usr/bin/env bash
set -Eeuo pipefail
umask 077

repo_root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd -P)"
test_suffix="${GITHUB_RUN_ID:-local}_${GITHUB_RUN_ATTEMPT:-1}_${BASHPID}"
test_suffix="${test_suffix//[^a-zA-Z0-9_]/_}"
test_db="pvnaive_migration_test_customer_idem_${test_suffix,,}"

: "${PVNAIVE_DB_HOST:=127.0.0.1}"
: "${PVNAIVE_DB_PORT:=5432}"
: "${PVNAIVE_DB_USER:=postgres}"
export PVNAIVE_DB_HOST PVNAIVE_DB_PORT PVNAIVE_DB_USER PVNAIVE_DB_NAME="${test_db}"

psql_admin() {
  psql --no-psqlrc --set ON_ERROR_STOP=1 --host "${PVNAIVE_DB_HOST}" --port "${PVNAIVE_DB_PORT}" --username "${PVNAIVE_DB_USER}" "$@"
}

cleanup() {
  psql_admin --dbname postgres --command "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname='${test_db}' AND pid <> pg_backend_pid()" >/dev/null 2>&1 || true
  dropdb --if-exists --host "${PVNAIVE_DB_HOST}" --port "${PVNAIVE_DB_PORT}" --username "${PVNAIVE_DB_USER}" "${test_db}" >/dev/null 2>&1 || true
  psql_admin --dbname postgres --command 'DROP ROLE IF EXISTS pvnaive_app; DROP ROLE IF EXISTS pvnaive_owner;' >/dev/null 2>&1 || true
}
trap cleanup EXIT HUP INT TERM
cleanup

psql_admin --dbname postgres <<'SQL' >/dev/null
CREATE ROLE pvnaive_owner NOLOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOREPLICATION NOBYPASSRLS;
CREATE ROLE pvnaive_app NOLOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOREPLICATION NOBYPASSRLS;
SQL
createdb --host "${PVNAIVE_DB_HOST}" --port "${PVNAIVE_DB_PORT}" --username "${PVNAIVE_DB_USER}" --owner pvnaive_owner --encoding UTF8 --template template0 "${test_db}"

migration_output="$("${repo_root}/scripts/db/migrate.sh")"
latest_version="$(find "${repo_root}/db/migrations" -maxdepth 1 -type f -name '[0-9][0-9][0-9][0-9]_*.up.sql' -printf '%f\n' | sort | tail -n1 | cut -d_ -f1 | sed 's/^0*//')"
[[ -n "${latest_version}" ]] || { echo 'ERROR: latest migration version could not be derived' >&2; exit 1; }
grep -Fqx "PVNAIVE_SCHEMA_VERSION=${latest_version}" <<< "${migration_output}" || {
  echo "ERROR: full migration stack did not advance schema to v${latest_version}" >&2
  exit 1
}

table_exists="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command "SELECT to_regclass('pvnaive.customer_mutation_keys') IS NOT NULL")"
[[ "${table_exists}" =~ ^(true|t)$ ]] || { echo 'ERROR: customer_mutation_keys table missing' >&2; exit 1; }

rls_contract="$(psql_admin --dbname "${test_db}" --tuples-only --no-align <<'SQL'
SELECT 'RLS=' || relrowsecurity || '|' || relforcerowsecurity
  FROM pg_class
 WHERE oid='pvnaive.customer_mutation_keys'::regclass;
SELECT 'POLICY=' || COUNT(*)
  FROM pg_policies
 WHERE schemaname='pvnaive' AND tablename='customer_mutation_keys' AND policyname='tenant_isolation';
SELECT 'DELETE=' || has_table_privilege('pvnaive_app', 'pvnaive.customer_mutation_keys', 'DELETE');
SQL
)"
grep -Eq '^RLS=(true|t)\|(true|t)$' <<< "${rls_contract}" || { echo 'ERROR: customer mutation RLS is not forced' >&2; exit 1; }
grep -Fqx 'POLICY=1' <<< "${rls_contract}" || { echo 'ERROR: customer mutation tenant policy missing' >&2; exit 1; }
grep -Eq '^DELETE=(false|f)$' <<< "${rls_contract}" || { echo 'ERROR: customer mutation ledger permits DELETE' >&2; exit 1; }

psql_admin --dbname "${test_db}" <<'SQL' >/dev/null
INSERT INTO pvnaive.actors (id, tenant_id, actor_role, email, display_name, status)
VALUES ('99000000-0000-0000-0000-000000000009', NULL, 'owner', 'idem-owner@example.invalid', 'Idempotency Owner', 'active');

INSERT INTO pvnaive.customer_mutation_keys (
  tenant_id, actor_id, idempotency_key, operation, request_hash, resource_type
)
SELECT id,
       '99000000-0000-0000-0000-000000000009',
       'customer-create-key-0001',
       'customer.create',
       public.digest('same-request', 'sha256'),
       'user'
  FROM pvnaive.tenants
 WHERE tenant_type='system' AND slug='direct';
SQL

if psql_admin --dbname "${test_db}" <<'SQL' >/dev/null 2>&1
INSERT INTO pvnaive.customer_mutation_keys (
  tenant_id, actor_id, idempotency_key, operation, request_hash, resource_type
)
SELECT id,
       '99000000-0000-0000-0000-000000000009',
       'customer-create-key-0001',
       'plan.create',
       public.digest('different-request', 'sha256'),
       'plan'
  FROM pvnaive.tenants
 WHERE tenant_type='system' AND slug='direct';
SQL
then
  echo 'ERROR: duplicate actor/idempotency key was accepted' >&2
  exit 1
fi

# Unwind any migrations newer than v5, then exercise the v5 idempotency rollback.
while :; do
  current_version="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command 'SELECT COALESCE(MAX(version),0) FROM pvnaive.schema_migrations')"
  [[ "${current_version}" -gt 4 ]] || break
  PVNAIVE_DISPOSABLE_DB=1 PVNAIVE_ALLOW_DESTRUCTIVE_ROLLBACK=ROLLBACK_ONE_MIGRATION \
    "${repo_root}/scripts/db/rollback.sh" >/dev/null
done
version_after_rollback="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command 'SELECT COALESCE(MAX(version),0) FROM pvnaive.schema_migrations')"
[[ "${version_after_rollback}" == "4" ]] || { echo 'ERROR: v5 rollback did not return to v4' >&2; exit 1; }
remaining="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command "SELECT to_regclass('pvnaive.customer_mutation_keys') IS NOT NULL")"
[[ "${remaining}" =~ ^(false|f)$ ]] || { echo 'ERROR: v5 rollback left customer_mutation_keys behind' >&2; exit 1; }

echo 'CUSTOMER_IDEMPOTENCY_MIGRATION_TEST=PASSED'
