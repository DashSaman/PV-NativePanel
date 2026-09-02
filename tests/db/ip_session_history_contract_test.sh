#!/usr/bin/env bash
set -Eeuo pipefail
root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd -P)"
up="$root/db/migrations/0021_ip_session_history.up.sql"
down="$root/db/migrations/0021_ip_session_history.down.sql"
[[ -f "$up" && -f "$down" ]] || { echo 'RED: schema21 migration pair missing' >&2; exit 1; }
grep -Eq 'interval[[:space:]]+''30 days''|make_interval\(days[[:space:]]*=>[[:space:]]*30\)' "$up" || { echo 'RED: exact 30-day retention boundary missing' >&2; exit 1; }
grep -Eq 'p_limit[[:space:]]+integer' "$up" || { echo 'RED: server-side pagination limit parameter missing' >&2; exit 1; }
grep -Eq 'p_limit[[:space:]]*<[^\n]*1|p_limit[[:space:]]*>[^\n]*(100|200|500)' "$up" || { echo 'RED: hard pagination bounds missing' >&2; exit 1; }
grep -Eq 'LIMIT[[:space:]]+p_limit' "$up" || { echo 'RED: query is not server-bounded by p_limit' >&2; exit 1; }
grep -Eq 'ENABLE ROW LEVEL SECURITY|FORCE ROW LEVEL SECURITY' "$up" || { echo 'RED: tenant RLS history boundary missing' >&2; exit 1; }
echo 'TASK16_IP_SESSION_HISTORY_CONTRACT=PASSED'
