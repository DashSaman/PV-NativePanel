#!/usr/bin/env bash
set -Eeuo pipefail
umask 077
repo_root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd -P)"
test_suffix="${GITHUB_RUN_ID:-local}_${GITHUB_RUN_ATTEMPT:-1}_${BASHPID}"
test_suffix="${test_suffix//[^a-zA-Z0-9_]/_}"
test_db="pvnaive_auth_refresh_reuse_${test_suffix,,}"
: "${PVNAIVE_DB_HOST:=127.0.0.1}"
: "${PVNAIVE_DB_PORT:=5432}"
: "${PVNAIVE_DB_USER:=postgres}"
export PVNAIVE_DB_HOST PVNAIVE_DB_PORT PVNAIVE_DB_USER PVNAIVE_DB_NAME="$test_db"
psql_admin(){ psql --no-psqlrc --set ON_ERROR_STOP=1 --host "$PVNAIVE_DB_HOST" --port "$PVNAIVE_DB_PORT" --username "$PVNAIVE_DB_USER" "$@"; }
cleanup(){ psql_admin -d postgres -c "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname='${test_db}' AND pid <> pg_backend_pid()" >/dev/null 2>&1 || true; dropdb --if-exists -h "$PVNAIVE_DB_HOST" -p "$PVNAIVE_DB_PORT" -U "$PVNAIVE_DB_USER" "$test_db" >/dev/null 2>&1 || true; psql_admin -d postgres -c 'DROP ROLE IF EXISTS pvnaive_app; DROP ROLE IF EXISTS pvnaive_owner;' >/dev/null 2>&1 || true; }
trap cleanup EXIT
cleanup
psql_admin -d postgres <<'SQL' >/dev/null
CREATE ROLE pvnaive_owner NOLOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOREPLICATION NOBYPASSRLS;
CREATE ROLE pvnaive_app NOLOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOREPLICATION NOBYPASSRLS;
SQL
createdb -h "$PVNAIVE_DB_HOST" -p "$PVNAIVE_DB_PORT" -U "$PVNAIVE_DB_USER" --owner pvnaive_owner --encoding UTF8 --template template0 "$test_db"
"$repo_root/scripts/db/migrate.sh" >/dev/null
[[ "$(psql_admin -d "$test_db" -Atc 'select max(version) from pvnaive.schema_migrations')" == 20 ]]
psql_admin -d "$test_db" <<'SQL' >/dev/null
INSERT INTO pvnaive.tenants(id,tenant_type,slug,display_name) VALUES ('18180000-0000-0000-0000-000000000001','reseller','reuse','Reuse');
INSERT INTO pvnaive.actors(id,tenant_id,actor_role,email,display_name,status) VALUES ('18180000-0000-0000-0000-000000000002','18180000-0000-0000-0000-000000000001','reseller','reuse@example.invalid','Reuse','active');
INSERT INTO pvnaive.auth_sessions(id,tenant_id,actor_id,token_hash,csrf_token_hash,refresh_family_id,expires_at,absolute_expires_at)
VALUES ('18180000-0000-0000-0000-000000000003','18180000-0000-0000-0000-000000000001','18180000-0000-0000-0000-000000000002',digest('old-token','sha256'),digest('csrf-old','sha256'),'18180000-0000-0000-0000-000000000004',clock_timestamp()+interval '1 hour',clock_timestamp()+interval '8 hours');
SQL
first="$(psql_admin -d "$test_db" -AtF '|' <<'SQL'
SET ROLE pvnaive_app;
SELECT encode(csrf_token_hash,'hex'), absolute_expires_at > clock_timestamp() FROM pvnaive.auth_refresh_session_metadata(digest('old-token','sha256'));
SELECT reuse_detected FROM pvnaive.auth_rotate_session(digest('old-token','sha256'),digest('new-token','sha256'),digest('csrf-new','sha256'),NULL,clock_timestamp()+interval '30 minutes');
SQL
)"
grep -Eq '^[0-9a-f]{64}\|t$' <<<"$first"
[[ "$(tail -n1 <<<"$first")" == f ]]
reused="$(psql_admin -d "$test_db" -AtF '|' <<'SQL'
SET ROLE pvnaive_app;
SELECT encode(csrf_token_hash,'hex'), absolute_expires_at > clock_timestamp() FROM pvnaive.auth_refresh_session_metadata(digest('old-token','sha256'));
SELECT reuse_detected FROM pvnaive.auth_rotate_session(digest('old-token','sha256'),digest('attacker-new','sha256'),digest('attacker-csrf','sha256'),NULL,clock_timestamp()+interval '30 minutes');
SQL
)"
grep -Eq '^[0-9a-f]{64}\|t$' <<<"$reused"
[[ "$(tail -n1 <<<"$reused")" == t ]]
family_state="$(psql_admin -d "$test_db" -AtF '|' -c "SELECT count(*), count(*) FILTER (WHERE revoked_at IS NOT NULL), count(*) FILTER (WHERE reuse_detected_at IS NOT NULL) FROM pvnaive.auth_sessions WHERE refresh_family_id='18180000-0000-0000-0000-000000000004'")"
[[ "$family_state" == '2|2|2' ]]
if psql_admin -d "$test_db" -Atc "SET ROLE pvnaive_app; SELECT * FROM pvnaive.auth_refresh_session_metadata(digest('wrong-token','sha256'))" >/dev/null 2>&1; then
  echo 'ERROR: wrong opaque token resolved refresh metadata' >&2; exit 1
fi
echo 'PVNAIVE_AUTH_REFRESH_REUSE_TEST=PASSED'
