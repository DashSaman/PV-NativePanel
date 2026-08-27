#!/usr/bin/env bash
# Verify PVNaive database reachability, schema and loopback binding.
set -Eeuo pipefail
umask 077

script_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd -P)"
# shellcheck source=scripts/db/lib.sh
source "${script_dir}/lib.sh"

pvnaive_require_command psql
pvnaive_require_command pg_isready
pvnaive_db_defaults
expected_version="${PVNAIVE_EXPECTED_SCHEMA_VERSION:-1}"
[[ "${expected_version}" =~ ^[0-9]+$ ]] || pvnaive_die "invalid expected schema version"

pg_isready --host "${PVNAIVE_DB_HOST}" --port "${PVNAIVE_DB_PORT}" --dbname "${PVNAIVE_DB_NAME}" --username "${PVNAIVE_DB_USER}" --timeout "${PVNAIVE_DB_CONNECT_TIMEOUT}" >/dev/null

health_row="$(pvnaive_psql_at --command "
WITH required(name) AS (
  VALUES ('actors'), ('backups'), ('credentials'), ('notification_deliveries'),
         ('notification_outbox'), ('notification_rules'), ('plans'), ('purchases'),
         ('quota_policies'), ('renewal_events'), ('reseller_credit_ledger'),
         ('reseller_plan_terms'), ('resellers'), ('runtime_health'),
         ('runtime_revisions'), ('schema_migrations'), ('sessions'), ('subscriptions'),
         ('subscription_tokens'), ('tenants'), ('usage_ledger'), ('usage_reset_events'),
         ('users'), ('audit_events'), ('auth_sessions'), ('log_metadata')
), checks AS (
  SELECT
    (SELECT COALESCE(MAX(version), 0) FROM pvnaive.schema_migrations) AS schema_version,
    (SELECT COUNT(*) FROM required r WHERE to_regclass('pvnaive.' || r.name) IS NOT NULL) AS required_tables,
    (SELECT COUNT(*) FROM pg_class c JOIN pg_namespace n ON n.oid = c.relnamespace
      WHERE n.nspname = 'pvnaive' AND c.relkind = 'r' AND c.relrowsecurity) AS rls_tables,
    (SELECT COUNT(*) FROM pvnaive.schema_migrations WHERE destructive) AS destructive_migrations,
    has_table_privilege(current_user, 'pvnaive.security_context_keys', 'SELECT') AS can_read_context_key,
    current_setting('row_security') AS row_security_setting,
    inet_server_addr() AS server_address,
    current_user AS database_user
)
SELECT schema_version || '|' || required_tables || '|' || rls_tables || '|' || destructive_migrations || '|' ||
       can_read_context_key || '|' || row_security_setting || '|' || server_address || '|' || database_user FROM checks;")"

IFS='|' read -r schema_version required_tables rls_tables destructive_migrations can_read_context_key row_security_setting server_address database_user <<< "${health_row}"
[[ "${schema_version}" == "${expected_version}" ]] || pvnaive_die "schema version ${schema_version}, expected ${expected_version}"
[[ "${required_tables}" == "26" ]] || pvnaive_die "required table check failed: ${required_tables}/26"
[[ "${rls_tables}" == "25" ]] || pvnaive_die "RLS coverage check failed: ${rls_tables}/25"
[[ "${destructive_migrations}" == "0" ]] || pvnaive_die "destructive migration record detected"
[[ "${can_read_context_key}" == "f" ]] || pvnaive_die "application role can read the RLS signing key"
[[ "${row_security_setting}" == "on" ]] || pvnaive_die "row_security is not forced on"
[[ "${server_address}" == "127.0.0.1" || "${server_address}" == "::1" ]] || pvnaive_die "database connection is not loopback"
[[ "${database_user}" == "${PVNAIVE_DB_USER}" ]] || pvnaive_die "database user mismatch"

echo "PVNAIVE_DB_HEALTH=OK"
echo "PVNAIVE_SCHEMA_VERSION=${schema_version}"
echo "PVNAIVE_DB_SERVER_ADDRESS=${server_address}"
