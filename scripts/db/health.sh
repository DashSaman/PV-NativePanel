#!/usr/bin/env bash
# Verify PVNaive database reachability, schema, identity, privilege boundaries and loopback binding.
set -Eeuo pipefail
umask 077

script_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd -P)"
# shellcheck source=scripts/db/lib.sh
source "${script_dir}/lib.sh"

pvnaive_require_command psql
pvnaive_require_command pg_isready
pvnaive_db_defaults
expected_version="${PVNAIVE_EXPECTED_SCHEMA_VERSION:-2}"
expected_db_user="${PVNAIVE_EXPECTED_DB_USER:-pvnaive_app}"
[[ "${expected_version}" =~ ^[0-9]+$ ]] || pvnaive_die "invalid expected schema version"
[[ "${expected_db_user}" =~ ^[a-z_][a-z0-9_]{0,62}$ ]] || pvnaive_die "invalid expected database user"
[[ "${PVNAIVE_DB_USER}" == "${expected_db_user}" ]] || pvnaive_die "health check must connect as ${expected_db_user}, got ${PVNAIVE_DB_USER}"
[[ -z "${PVNAIVE_RUN_AS_OS_USER:-}" ]] || pvnaive_die "health check refuses PVNAIVE_RUN_AS_OS_USER=${PVNAIVE_RUN_AS_OS_USER}"
[[ "${PVNAIVE_DB_HOST}" == "127.0.0.1" || "${PVNAIVE_DB_HOST}" == "::1" ]] || \
  pvnaive_die "health check requires an explicit loopback database host"

pg_isready --host "${PVNAIVE_DB_HOST}" --port "${PVNAIVE_DB_PORT}" --dbname "${PVNAIVE_DB_NAME}" --username "${PVNAIVE_DB_USER}" --timeout "${PVNAIVE_DB_CONNECT_TIMEOUT}" >/dev/null

# inet_server_addr()/inet_client_addr() return PostgreSQL inet values. Use host()
# so the health contract compares canonical address text without /32 or /128.
identity_row="$(pvnaive_psql_at --command "SELECT current_user || '|' || session_user || '|' || current_setting('row_security') || '|' || COALESCE(host(inet_server_addr()), '') || '|' || COALESCE(inet_server_port()::text, '') || '|' || COALESCE(host(inet_client_addr()), '')")"
IFS='|' read -r database_user session_user row_security_setting server_address server_port client_address <<< "${identity_row}"
[[ "${database_user}" == "${expected_db_user}" ]] || pvnaive_die "database user mismatch: current_user=${database_user}, expected=${expected_db_user}"
[[ "${session_user}" == "${expected_db_user}" ]] || pvnaive_die "database session user mismatch: session_user=${session_user}, expected=${expected_db_user}"
[[ "${row_security_setting}" == "on" ]] || pvnaive_die "row_security is not forced on"
[[ "${server_address}" == "127.0.0.1" || "${server_address}" == "::1" ]] || pvnaive_die "database server endpoint is not loopback: ${server_address:-unknown}"
[[ "${client_address}" == "127.0.0.1" || "${client_address}" == "::1" ]] || pvnaive_die "database client endpoint is not loopback: ${client_address:-unknown}"
[[ "${server_port}" == "${PVNAIVE_DB_PORT}" ]] || pvnaive_die "database server port mismatch: ${server_port:-unknown}, expected ${PVNAIVE_DB_PORT}"

key_table_exists="$(pvnaive_psql_at --command "SELECT to_regclass('pvnaive.security_context_keys') IS NOT NULL")"
[[ "${key_table_exists}" == "t" ]] || pvnaive_die "RLS signing-key table is missing"

can_read_context_key="$(pvnaive_psql_at --command "SELECT has_table_privilege(current_user, 'pvnaive.security_context_keys', 'SELECT')")"
[[ "${can_read_context_key}" == "f" ]] || pvnaive_die "application role has SELECT privilege on the RLS signing-key table (current_user=${database_user})"

# Prove the effective boundary with a real SELECT permission check. LIMIT 0 keeps
# the signing key out of output while PostgreSQL still performs access control.
if pvnaive_psql_at --command 'SELECT signing_key FROM pvnaive.security_context_keys LIMIT 0' >/dev/null 2>&1; then
  pvnaive_die "application role can directly SELECT the RLS signing key"
fi

if ((expected_version >= 2)); then
  mfa_tables="$(pvnaive_psql_at --command "SELECT (to_regclass('pvnaive.actor_totp_factors') IS NOT NULL)::text || '|' || (to_regclass('pvnaive.actor_mfa_recovery_codes') IS NOT NULL)::text")"
  [[ "${mfa_tables}" == "true|true" || "${mfa_tables}" == "t|t" ]] || pvnaive_die "S04 MFA tables are missing"
  mfa_direct="$(pvnaive_psql_at --command "SELECT has_table_privilege(current_user, 'pvnaive.actor_totp_factors', 'SELECT')::text || '|' || has_table_privilege(current_user, 'pvnaive.actor_mfa_recovery_codes', 'SELECT')::text")"
  [[ "${mfa_direct}" == "false|false" || "${mfa_direct}" == "f|f" ]] || pvnaive_die "application role has direct SELECT on MFA secret tables"
fi

health_row="$(pvnaive_psql_at --command "
WITH required(name) AS (
  VALUES ('actors'), ('backups'), ('credentials'), ('notification_deliveries'),
         ('notification_outbox'), ('notification_rules'), ('plans'), ('purchases'),
         ('quota_policies'), ('renewal_events'), ('reseller_credit_ledger'),
         ('reseller_plan_terms'), ('resellers'), ('runtime_health'),
         ('runtime_revisions'), ('schema_migrations'), ('sessions'), ('subscriptions'),
         ('subscription_tokens'), ('tenants'), ('usage_ledger'), ('usage_reset_events'),
         ('users'), ('audit_events'), ('auth_sessions'), ('log_metadata'),
         ('actor_totp_factors'), ('actor_mfa_recovery_codes')
), checks AS (
  SELECT
    (SELECT COALESCE(MAX(version), 0) FROM pvnaive.schema_migrations) AS schema_version,
    (SELECT COUNT(*) FROM required r WHERE to_regclass('pvnaive.' || r.name) IS NOT NULL) AS required_tables,
    (SELECT COUNT(*) FROM pg_class c JOIN pg_namespace n ON n.oid = c.relnamespace
      WHERE n.nspname = 'pvnaive' AND c.relkind = 'r' AND c.relrowsecurity) AS rls_tables,
    (SELECT COUNT(*) FROM pvnaive.schema_migrations WHERE destructive) AS destructive_migrations
)
SELECT schema_version || '|' || required_tables || '|' || rls_tables || '|' || destructive_migrations FROM checks;")"

IFS='|' read -r schema_version required_tables rls_tables destructive_migrations <<< "${health_row}"
[[ "${schema_version}" == "${expected_version}" ]] || pvnaive_die "schema version ${schema_version}, expected ${expected_version}"
if ((expected_version >= 2)); then
  [[ "${required_tables}" == "28" ]] || pvnaive_die "required table check failed: ${required_tables}/28"
else
  [[ "${required_tables}" == "26" ]] || pvnaive_die "required table check failed: ${required_tables}/26"
fi
[[ "${rls_tables}" == "25" ]] || pvnaive_die "RLS coverage check failed: ${rls_tables}/25"
[[ "${destructive_migrations}" == "0" ]] || pvnaive_die "destructive migration record detected"

echo "PVNAIVE_DB_HEALTH=OK"
echo "PVNAIVE_SCHEMA_VERSION=${schema_version}"
echo "PVNAIVE_DB_USER=${database_user}"
echo "PVNAIVE_DB_SERVER_ADDRESS=${server_address}"
echo "PVNAIVE_DB_SERVER_PORT=${server_port}"
echo "PVNAIVE_DB_CLIENT_ADDRESS=${client_address}"
echo "PVNAIVE_SECRET_DIRECT_SELECT=DENIED"
if ((expected_version >= 2)); then
  echo "PVNAIVE_MFA_DIRECT_SELECT=DENIED"
fi
