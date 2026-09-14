#!/usr/bin/env bash
set -Eeuo pipefail
umask 077

# R2/R3 decision-sink migration gate (0032): proves the append-only decision
# audit, the replay dedup identity, the monotonic current-state guard, the
# honest untracked-user behavior, and the trusted boundary (pvnaive_app can
# apply/read only through the SECURITY DEFINER functions — no table access,
# no UPDATE/DELETE anywhere).

repo_root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd -P)"
suffix="${GITHUB_RUN_ID:-local}_${GITHUB_RUN_ATTEMPT:-1}_${BASHPID}"
suffix="${suffix//[^a-zA-Z0-9_]/_}"
test_db="pvnaive_r2_steering_${suffix,,}"
: "${PVNAIVE_DB_HOST:=127.0.0.1}"
: "${PVNAIVE_DB_PORT:=5432}"
: "${PVNAIVE_DB_USER:=postgres}"
export PVNAIVE_DB_HOST PVNAIVE_DB_PORT PVNAIVE_DB_USER

psql_admin() {
  psql --no-psqlrc --set ON_ERROR_STOP=1 --host "${PVNAIVE_DB_HOST}" \
    --port "${PVNAIVE_DB_PORT}" --username "${PVNAIVE_DB_USER}" "$@"
}
cleanup() {
  psql_admin --dbname postgres --command \
    "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname='${test_db}' AND pid <> pg_backend_pid()" \
    >/dev/null 2>&1 || true
  dropdb --if-exists --host "${PVNAIVE_DB_HOST}" --port "${PVNAIVE_DB_PORT}" \
    --username "${PVNAIVE_DB_USER}" "${test_db}" >/dev/null 2>&1 || true
  psql_admin --dbname postgres --command 'DROP ROLE IF EXISTS pvnaive_app; DROP ROLE IF EXISTS pvnaive_owner;' >/dev/null 2>&1 || true
}
trap cleanup EXIT
cleanup

psql_admin --dbname postgres <<'SQL' >/dev/null
CREATE ROLE pvnaive_owner NOLOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOREPLICATION NOBYPASSRLS;
CREATE ROLE pvnaive_app NOLOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOREPLICATION NOBYPASSRLS;
SQL
createdb --host "${PVNAIVE_DB_HOST}" --port "${PVNAIVE_DB_PORT}" --username "${PVNAIVE_DB_USER}" \
  --owner pvnaive_owner --encoding UTF8 --template template0 "${test_db}"
PVNAIVE_DB_NAME="${test_db}" "${repo_root}/scripts/db/migrate.sh" >/dev/null

schema="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command 'SELECT COALESCE(MAX(version),0) FROM pvnaive.schema_migrations')"
[[ "${schema}" =~ ^[0-9]+$ && "${schema}" -ge 32 ]] || { echo "ERROR: expected schema >=32, got ${schema}" >&2; exit 1; }

# 1) Schema contract.
contract="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command "
SELECT concat_ws('|',
 (to_regclass('pvnaive.steering_decisions') IS NOT NULL)::text,
 (to_regclass('pvnaive.user_steering_state') IS NOT NULL)::text,
 (to_regprocedure('pvnaive.steering_decision_apply(uuid,text,text,bigint,text,timestamptz,jsonb)') IS NOT NULL)::text,
 (to_regprocedure('pvnaive.steering_state_read(uuid)') IS NOT NULL)::text
);")"
[[ "${contract}" == 'true|true|true|true' ]] || {
  echo "ERROR: steering decision schema contract failed: ${contract}" >&2; exit 1;
}

# Fixture: one known user.
psql_admin --dbname "${test_db}" <<'SQL' >/dev/null
INSERT INTO pvnaive.actors (id, tenant_id, actor_role, email, display_name, status)
VALUES ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', NULL, 'owner', 'owner@r2.invalid', 'R2 Owner', 'active');

INSERT INTO pvnaive.users (id, tenant_id, username, display_name, status, created_by_actor_id)
SELECT 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', id, 'r2bob', 'R2 Bob', 'active', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa'
FROM pvnaive.tenants WHERE slug='direct' AND tenant_type='system';
SQL

apply() {
  # Nullable TEXT/JSONB arguments must reach SQL as quoted literals or bare NULL.
  local prev_sql="NULL" details_sql="NULL"
  [[ "${3}" != "NULL" ]] && prev_sql="'${3}'"
  [[ "${7}" != "NULL" ]] && details_sql="'${7}'::jsonb"
  psql_admin --dbname "${test_db}" --tuples-only --no-align --command "SET ROLE pvnaive_app; SELECT concat_ws('|',tracked::text,accepted::text,duplicate::text) FROM pvnaive.steering_decision_apply('${1}','${2}',${prev_sql},${4},'${5}','${6}',${details_sql});"
}

# 2) First decision (initial assignment) is accepted.
first="$(apply 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb' 'direct-1' NULL 7 initial '2026-09-14T10:00:00Z' NULL)"
[[ "${first##*$'\n'}" == 'true|true|false' ]] || { echo "ERROR: first decision: ${first}" >&2; exit 1; }

# 3) Replay (engine restart re-emits the same window/reason) is a duplicate
#    and adds NO second audit row.
replay="$(apply 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb' 'direct-1' NULL 7 initial '2026-09-14T10:05:00Z' NULL)"
[[ "${replay##*$'\n'}" == 'true|false|true' ]] || { echo "ERROR: replay dedup: ${replay}" >&2; exit 1; }
count="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command "SELECT count(*) FROM pvnaive.steering_decisions;")"
[[ "${count}" == 1 ]] || { echo "ERROR: replay double-counted: count=${count}" >&2; exit 1; }

# 4) A distinct reason in the same window (kill_switch) is its own row.
kill="$(apply 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb' 'direct-2' 'direct-1' 7 kill_switch '2026-09-14T10:06:00Z' '{"candidates":[{"node":"direct-2","score":1.2}]}')"
[[ "${kill##*$'\n'}" == 'true|true|false' ]] || { echo "ERROR: kill switch decision: ${kill}" >&2; exit 1; }
count="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command "SELECT count(*) FROM pvnaive.steering_decisions;")"
[[ "${count}" == 2 ]] || { echo "ERROR: kill row missing: count=${count}" >&2; exit 1; }

# 5) Current state advanced to the kill_switch node, and an OLDER window can
#    never overwrite a newer one.
state="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command "SET ROLE pvnaive_app; SELECT concat_ws('|',node_id,reason,window_index::text) FROM pvnaive.steering_state_read('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb');")"
[[ "${state##*$'\n'}" == 'direct-2|kill_switch|7' ]] || { echo "ERROR: current state: ${state}" >&2; exit 1; }
older="$(apply 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb' 'direct-1' NULL 6 hysteresis '2026-09-14T11:00:00Z' NULL)"
[[ "${older##*$'\n'}" == 'true|true|false' ]] || { echo "ERROR: older window apply: ${older}" >&2; exit 1; }
state_after_older="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command "SET ROLE pvnaive_app; SELECT node_id FROM pvnaive.steering_state_read('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb');")"
[[ "${state_after_older##*$'\n'}" == 'direct-2' ]] || {
  echo "ERROR: older window overwrote newer state: ${state_after_older}" >&2; exit 1;
}
# ...yet the audit stays append-only complete.
audit_total="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command "SELECT count(*) FROM pvnaive.steering_decisions;")"
[[ "${audit_total}" == 3 ]] || { echo "ERROR: audit row count: ${audit_total}" >&2; exit 1; }

# 6) Unknown users are honestly untracked — never fabricated.
untracked="$(apply 'eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee' 'direct-1' NULL 7 initial '2026-09-14T10:00:00Z' NULL)"
[[ "${untracked##*$'\n'}" == 'false|false|false' ]] || { echo "ERROR: untracked user: ${untracked}" >&2; exit 1; }

# 7) Malformed decisions are rejected (non-actionable reason).
set +e
bad_rc="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command "SET ROLE pvnaive_app; SELECT pvnaive.steering_decision_apply('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb','direct-1',NULL,7,'no_change','2026-09-14T10:00:00Z',NULL);" >/dev/null 2>&1; echo $?)"
set -e
[[ "${bad_rc}" != 0 ]] || { echo 'ERROR: non-actionable reason accepted' >&2; exit 1; }

# 8) Unknown user reads honestly empty.
none="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command "SET ROLE pvnaive_app; SELECT count(*) FROM pvnaive.steering_state_read('eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee');")"
[[ "${none##*$'\n'}" == 0 ]] || { echo "ERROR: unknown user state read: ${none}" >&2; exit 1; }

# 9) Trusted boundary: pvnaive_app has NO direct table privileges (RLS + grants).
set +e
denied_rc="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command "SET ROLE pvnaive_app; SELECT count(*) FROM pvnaive.steering_decisions;" >/dev/null 2>&1; echo $?)"
set -e
[[ "${denied_rc}" != 0 ]] || { echo 'ERROR: pvnaive_app read steering_decisions directly' >&2; exit 1; }

# 10) Append-only enforcement at the role level: even the owner role cannot
#     UPDATE/DELETE audit rows through grants (no DML grants were issued).
set +e
update_rc="$(psql_admin --dbname "${test_db}" --command "SET ROLE pvnaive_owner; UPDATE pvnaive.steering_decisions SET node_id='tampered';" >/dev/null 2>&1; echo $?)"
delete_rc="$(psql_admin --dbname "${test_db}" --command "SET ROLE pvnaive_owner; DELETE FROM pvnaive.steering_decisions;" >/dev/null 2>&1; echo $?)"
set -e
[[ "${update_rc}" != 0 && "${delete_rc}" != 0 ]] || {
  echo 'ERROR: steering_decisions rows are mutable — append-only contract broken' >&2; exit 1;
}

echo "steering decisions migration gate: OK"
