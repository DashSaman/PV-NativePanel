#!/usr/bin/env bash
set -Eeuo pipefail
umask 077
root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd -P)"
up="$root/db/migrations/0021_ip_session_history.up.sql"
down="$root/db/migrations/0021_ip_session_history.down.sql"
[[ -f "$up" && -f "$down" ]] || { echo 'RED: schema21 migration pair missing' >&2; exit 1; }
grep -Fqx -- '-- pvnaive:migration-version 0021' "$up" || { echo 'RED: schema21 version marker missing' >&2; exit 1; }
grep -Eq "interval[[:space:]]+'30 days'|make_interval\(days[[:space:]]*=>[[:space:]]*30\)" "$up" || { echo 'RED: exact 30-day retention boundary missing' >&2; exit 1; }
grep -Eq 'p_limit[[:space:]]+integer' "$up" || { echo 'RED: server-side pagination limit parameter missing' >&2; exit 1; }
grep -Eq 'p_limit[[:space:]]*<[^\n]*1|p_limit[[:space:]]*>[^\n]*500' "$up" || { echo 'RED: hard pagination bounds missing' >&2; exit 1; }
grep -Eq 'LIMIT[[:space:]]+p_limit' "$up" || { echo 'RED: query is not server-bounded by p_limit' >&2; exit 1; }
grep -q 'direct_naive_accounting_session_peers' "$up" || { echo 'RED: trusted peer lineage missing' >&2; exit 1; }
grep -Eq 's\.final[[:space:]]*=[[:space:]]*true|s\.final' "$up" || { echo 'RED: finalized accounting gate missing' >&2; exit 1; }
grep -Eq 's\.accounting_complete[[:space:]]*=[[:space:]]*true|s\.accounting_complete' "$up" || { echo 'RED: accounting-complete gate missing' >&2; exit 1; }
grep -q 'FORCE ROW LEVEL SECURITY' "$up" || { echo 'RED: forced tenant RLS history boundary missing' >&2; exit 1; }

: "${PVNAIVE_DB_HOST:=127.0.0.1}"
: "${PVNAIVE_DB_PORT:=5432}"
: "${PVNAIVE_DB_USER:=postgres}"
export PVNAIVE_DB_HOST PVNAIVE_DB_PORT PVNAIVE_DB_USER
suffix="${GITHUB_RUN_ID:-local}_${GITHUB_RUN_ATTEMPT:-1}_${BASHPID}"; suffix="${suffix//[^a-zA-Z0-9_]/_}"
test_db="pvnaive_task16_${suffix,,}"
tmp="$(mktemp -d)"
psql_admin(){ psql --no-psqlrc --set ON_ERROR_STOP=1 -h "$PVNAIVE_DB_HOST" -p "$PVNAIVE_DB_PORT" -U "$PVNAIVE_DB_USER" "$@"; }
cleanup(){
  rm -rf "$tmp" 2>/dev/null || true
  psql_admin -d postgres -c "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname='${test_db}' AND pid<>pg_backend_pid()" >/dev/null 2>&1 || true
  dropdb --if-exists -h "$PVNAIVE_DB_HOST" -p "$PVNAIVE_DB_PORT" -U "$PVNAIVE_DB_USER" "$test_db" >/dev/null 2>&1 || true
  psql_admin -d postgres -c 'DROP ROLE IF EXISTS pvnaive_app; DROP ROLE IF EXISTS pvnaive_owner;' >/dev/null 2>&1 || true
}
trap cleanup EXIT HUP INT TERM
cleanup
psql_admin -d postgres <<'SQL' >/dev/null
CREATE ROLE pvnaive_owner NOLOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOREPLICATION NOBYPASSRLS;
CREATE ROLE pvnaive_app LOGIN PASSWORD 'pvnaive-task16-ci' NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOREPLICATION NOBYPASSRLS;
SQL
createdb -h "$PVNAIVE_DB_HOST" -p "$PVNAIVE_DB_PORT" -U "$PVNAIVE_DB_USER" --owner pvnaive_owner --encoding UTF8 --template template0 "$test_db"
export PVNAIVE_DB_NAME="$test_db"
mkdir -p "$tmp/migrations"
for version in $(seq 1 20); do prefix="$(printf '%04d' "$version")"; cp "$root/db/migrations/${prefix}_"*.sql "$tmp/migrations/"; done
cp "$up" "$down" "$tmp/migrations/"
( cd "$tmp/migrations" && sha256sum *.sql > SHA256SUMS )
PVNAIVE_MIGRATIONS_DIR="$tmp/migrations" "$root/scripts/db/migrate.sh" >/dev/null
[[ "$(psql_admin -d "$test_db" -Atc 'select max(version) from pvnaive.schema_migrations')" == 21 ]] || { echo 'RED: schema21 did not migrate on PostgreSQL18' >&2; exit 1; }
[[ "$(psql_admin -d "$test_db" -Atc "select relrowsecurity::text||'|'||relforcerowsecurity::text from pg_class c join pg_namespace n on n.oid=c.relnamespace where n.nspname='pvnaive' and c.relname='direct_naive_session_history'")" =~ ^(t|true)\|(t|true)$ ]] || { echo 'RED: history table is not ENABLE+FORCE RLS' >&2; exit 1; }
[[ "$(psql_admin -d "$test_db" -Atc "select has_function_privilege('pvnaive_app','pvnaive.list_customer_session_history(uuid,timestamptz,integer)','EXECUTE')")" =~ ^(t|true)$ ]] || { echo 'RED: bounded history read is not available to app role' >&2; exit 1; }
[[ "$(psql_admin -d "$test_db" -Atc "select has_function_privilege('pvnaive_app','pvnaive.sync_direct_naive_session_history(timestamptz)','EXECUTE')")" =~ ^(f|false)$ ]] || { echo 'RED: maintenance sync/purge leaked to app role' >&2; exit 1; }
fn="$(psql_admin -d "$test_db" -Atc "select pg_get_functiondef('pvnaive.sync_direct_naive_session_history(timestamptz)'::regprocedure)")"
[[ "$fn" == *"direct_naive_accounting_session_peers"* && "$fn" == *"s.final"* && "$fn" == *"s.accounting_complete"* && "$fn" == *"30 days"* ]] || { echo 'RED: sync does not derive only from trusted finalized accounting facts with exact retention' >&2; exit 1; }
set +e
bad_hi="$(PGPASSWORD=pvnaive-task16-ci psql --no-psqlrc -h "$PVNAIVE_DB_HOST" -p "$PVNAIVE_DB_PORT" -U pvnaive_app -d "$test_db" -v ON_ERROR_STOP=1 -c "select * from pvnaive.list_customer_session_history('00000000-0000-0000-0000-000000000001',clock_timestamp(),501)" 2>&1)"; rc_hi=$?
bad_lo="$(PGPASSWORD=pvnaive-task16-ci psql --no-psqlrc -h "$PVNAIVE_DB_HOST" -p "$PVNAIVE_DB_PORT" -U pvnaive_app -d "$test_db" -v ON_ERROR_STOP=1 -c "select * from pvnaive.list_customer_session_history('00000000-0000-0000-0000-000000000001',clock_timestamp(),0)" 2>&1)"; rc_lo=$?
set -e
[[ $rc_hi -ne 0 && "$bad_hi" == *"invalid session history query"* ]] || { echo 'RED: oversized pagination was not rejected server-side' >&2; exit 1; }
[[ $rc_lo -ne 0 && "$bad_lo" == *"invalid session history query"* ]] || { echo 'RED: zero pagination was not rejected server-side' >&2; exit 1; }
psql_admin -d "$test_db" -f "$down" >/dev/null
[[ "$(psql_admin -d "$test_db" -Atc 'select max(version) from pvnaive.schema_migrations')" == 20 ]] || { echo 'RED: schema21 rollback did not restore version 20' >&2; exit 1; }
[[ "$(psql_admin -d "$test_db" -Atc "select to_regclass('pvnaive.direct_naive_session_history') is null")" =~ ^(t|true)$ ]] || { echo 'RED: schema21 rollback left history table behind' >&2; exit 1; }
echo 'TASK16_IP_SESSION_HISTORY_PG18=PASSED'
