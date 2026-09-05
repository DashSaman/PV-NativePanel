#!/usr/bin/env bash
# Explicit destructive maintenance for schema21 retained session/IP history.
set -Eeuo pipefail
umask 077

script_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd -P)"
# shellcheck source=scripts/db/lib.sh
source "${script_dir}/lib.sh"

pvnaive_require_command psql
pvnaive_db_defaults

[[ "${PVNAIVE_SESSION_HISTORY_PURGE_CONFIRM:-}" == "PURGE_OLDER_THAN_30_DAYS" ]] || \
  pvnaive_die "session-history purge requires PVNAIVE_SESSION_HISTORY_PURGE_CONFIRM=PURGE_OLDER_THAN_30_DAYS"

observed_at="${PVNAIVE_SESSION_HISTORY_OBSERVED_AT:-}"
if [[ -z "${observed_at}" ]]; then
  observed_at="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
fi
[[ "${observed_at}" =~ ^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}Z$ ]] || \
  pvnaive_die "PVNAIVE_SESSION_HISTORY_OBSERVED_AT must be UTC YYYY-MM-DDTHH:MM:SSZ"

can_assume_owner="$(pvnaive_psql_at --command "SELECT pg_has_role(current_user, 'pvnaive_owner', 'USAGE')")"
[[ "${can_assume_owner}" == "t" ]] || pvnaive_die "purge connection cannot assume pvnaive_owner"

pvnaive_psql --single-transaction --set observed_at="${observed_at}" <<'SQL'
SELECT pg_advisory_xact_lock(hashtext('pvnaive-session-history-purge'));
SET LOCAL lock_timeout = '10s';
SET LOCAL statement_timeout = '120s';
SET LOCAL ROLE pvnaive_owner;
SELECT pvnaive.direct_naive_accounting_enter_context();
WITH purged AS (
    DELETE FROM pvnaive.direct_naive_session_history
     WHERE final_at < :'observed_at'::timestamptz - interval '30 days'
    RETURNING 1
)
SELECT count(*) AS purged_rows FROM purged;
SELECT pvnaive.direct_naive_accounting_leave_context();
SQL

echo "PVNAIVE_SESSION_HISTORY_PURGE_RESULT=PASSED"
