#!/usr/bin/env bash
set -Eeuo pipefail
umask 077

repo_root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd -P)"
up="${repo_root}/db/migrations/0006_direct_subscription_tokens.up.sql"
down="${repo_root}/db/migrations/0006_direct_subscription_tokens.down.sql"

[[ -f "${up}" ]] || { echo "ERROR: missing 0006_direct_subscription_tokens.up.sql" >&2; exit 1; }
[[ -f "${down}" ]] || { echo "ERROR: missing 0006_direct_subscription_tokens.down.sql" >&2; exit 1; }
grep -Fqx -- '-- pvnaive:migration-version 0006' "${up}"
grep -Fqx -- '-- pvnaive:transactional true' "${up}"
grep -Fqx -- '-- pvnaive:destructive false' "${up}"

: "${PVNAIVE_DB_HOST:=127.0.0.1}"
: "${PVNAIVE_DB_PORT:=5432}"
: "${PVNAIVE_DB_USER:=postgres}"
export PVNAIVE_DB_HOST PVNAIVE_DB_PORT PVNAIVE_DB_USER

test_suffix="${GITHUB_RUN_ID:-local}_${GITHUB_RUN_ATTEMPT:-1}_${BASHPID}"
test_suffix="${test_suffix//[^a-zA-Z0-9_]/_}"
test_db="pvnaive_direct_sub_${test_suffix,,}"

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
CREATE ROLE pvnaive_app LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOREPLICATION NOBYPASSRLS PASSWORD 'pvnaive-direct-sub-ci';
ALTER ROLE pvnaive_app SET row_security = on;
SQL

createdb --host "${PVNAIVE_DB_HOST}" --port "${PVNAIVE_DB_PORT}" \
  --username "${PVNAIVE_DB_USER}" --owner pvnaive_owner --encoding UTF8 --template template0 "${test_db}"
export PVNAIVE_DB_NAME="${test_db}"
"${repo_root}/scripts/db/migrate.sh" >/dev/null

schema_version="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command \
  'SELECT COALESCE(MAX(version),0) FROM pvnaive.schema_migrations')"
[[ "${schema_version}" == "6" ]] || { echo "ERROR: schema version=${schema_version}, want=6" >&2; exit 1; }

contract="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command "
SELECT
  (to_regclass('pvnaive.direct_subscription_tokens') IS NOT NULL) || '|' ||
  EXISTS (
    SELECT 1 FROM pg_proc p JOIN pg_namespace n ON n.oid=p.pronamespace
    WHERE n.nspname='pvnaive' AND p.proname='resolve_direct_subscription_token' AND p.prosecdef
  ) || '|' ||
  (SELECT relrowsecurity FROM pg_class WHERE oid='pvnaive.direct_subscription_tokens'::regclass) || '|' ||
  NOT (SELECT relforcerowsecurity FROM pg_class WHERE oid='pvnaive.direct_subscription_tokens'::regclass) || '|' ||
  ((SELECT COUNT(*) FROM pg_trigger WHERE tgrelid='pvnaive.users'::regclass AND tgname='direct_subscription_user_sync' AND NOT tgisinternal)=1) || '|' ||
  ((SELECT COUNT(*) FROM pg_trigger WHERE tgrelid='pvnaive.service_terms'::regclass AND tgname='direct_subscription_service_term_sync' AND NOT tgisinternal)=1) || '|' ||
  ((SELECT COUNT(*) FROM pg_trigger WHERE tgrelid='pvnaive.naive_runtime_credentials'::regclass AND tgname='direct_subscription_runtime_credential_sync' AND NOT tgisinternal)=1);")"
[[ "${contract}" == "true|true|true|true|true|true|true" || "${contract}" == "t|t|t|t|t|t|t" ]] || {
  echo "ERROR: direct subscription schema contract failed: ${contract}" >&2
  exit 1
}

psql_admin --dbname "${test_db}" >/dev/null <<'SQL'
INSERT INTO pvnaive.actors (id, tenant_id, actor_role, email, display_name, password_hash, status)
VALUES (
  'a6000000-0000-0000-0000-000000000001', NULL, 'owner',
  'direct-sub-owner@example.invalid', 'Direct Sub Owner',
  '$argon2id$v=19$m=19456,t=2,p=1$AAAAAAAAAAAAAAAAAAAAAA$AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA',
  'active'
);
INSERT INTO pvnaive.auth_sessions (
  id, tenant_id, actor_id, token_hash, refresh_family_id, user_agent_hash,
  expires_at, absolute_expires_at, csrf_token_hash
) VALUES (
  'b6000000-0000-0000-0000-000000000001', NULL,
  'a6000000-0000-0000-0000-000000000001', decode(repeat('16',32),'hex'),
  'c6000000-0000-0000-0000-000000000001', decode(repeat('26',32),'hex'),
  clock_timestamp() + interval '1 hour', clock_timestamp() + interval '12 hours',
  decode(repeat('36',32),'hex')
);
SQL

PGPASSWORD='pvnaive-direct-sub-ci' psql --no-psqlrc --set ON_ERROR_STOP=1 \
  --host "${PVNAIVE_DB_HOST}" --port "${PVNAIVE_DB_PORT}" --username pvnaive_app --dbname "${test_db}" \
  >/dev/null <<'SQL'
BEGIN;
SELECT pvnaive.set_request_context(decode(repeat('16',32),'hex'));

INSERT INTO pvnaive.users (id, tenant_id, username, display_name, status, created_by_actor_id)
SELECT 'd6000000-0000-0000-0000-000000000001', id, 'sub-user', 'Sub User', 'active',
       'a6000000-0000-0000-0000-000000000001'
FROM pvnaive.tenants WHERE tenant_type='system' AND slug='direct';

INSERT INTO pvnaive.naive_runtime_credentials (
  id, username, secret_hash, secret_ciphertext, secret_nonce, encryption_key_id,
  status, origin, created_by_actor_id, updated_by_actor_id
) VALUES (
  'e6000000-0000-0000-0000-000000000001', 'sub-user', decode(repeat('46',32),'hex'),
  decode(repeat('56',16),'hex'), decode(repeat('66',12),'hex'), 'runtime-v1',
  'active', 'panel', 'a6000000-0000-0000-0000-000000000001', 'a6000000-0000-0000-0000-000000000001'
);

INSERT INTO pvnaive.service_terms (
  id, tenant_id, user_id, quota_bytes, duration_seconds, start_policy,
  purchased_at, state
)
SELECT 'f6000000-0000-0000-0000-000000000001', u.tenant_id, u.id,
       53687091200, 2592000, 'on_first_successful_connection', clock_timestamp(), 'pending'
FROM pvnaive.users u WHERE u.id='d6000000-0000-0000-0000-000000000001';

INSERT INTO pvnaive.user_runtime_credentials (
  tenant_id, user_id, service_term_id, runtime_credential_id, role
)
SELECT u.tenant_id, u.id, 'f6000000-0000-0000-0000-000000000001',
       'e6000000-0000-0000-0000-000000000001', 'primary'
FROM pvnaive.users u WHERE u.id='d6000000-0000-0000-0000-000000000001';

INSERT INTO pvnaive.direct_subscription_tokens (
  tenant_id, user_id, service_term_id, runtime_credential_id,
  token_hash, token_prefix, status, user_state, service_state,
  runtime_username, secret_ciphertext, secret_nonce, encryption_key_id, expires_at
)
SELECT u.tenant_id, u.id, 'f6000000-0000-0000-0000-000000000001',
       'e6000000-0000-0000-0000-000000000001', decode(repeat('76',32),'hex'),
       'testtoken', 'active', 'active', 'pending', r.username,
       r.secret_ciphertext, r.secret_nonce, r.encryption_key_id, NULL
FROM pvnaive.users u
JOIN pvnaive.naive_runtime_credentials r ON r.id='e6000000-0000-0000-0000-000000000001'
WHERE u.id='d6000000-0000-0000-0000-000000000001';
COMMIT;
SQL

resolve_projection() {
  PGPASSWORD='pvnaive-direct-sub-ci' psql --no-psqlrc --set ON_ERROR_STOP=1 \
    --host "${PVNAIVE_DB_HOST}" --port "${PVNAIVE_DB_PORT}" --username pvnaive_app --dbname "${test_db}" \
    --tuples-only --no-align --command "
SELECT runtime_credential_id::text || '|' || runtime_username || '|' || user_state || '|' || service_state || '|' || encryption_key_id
FROM pvnaive.resolve_direct_subscription_token(decode(repeat('76',32),'hex'));"
}

resolved="$(resolve_projection)"
[[ "${resolved}" == "e6000000-0000-0000-0000-000000000001|sub-user|active|pending|runtime-v1" ]] || {
  echo "ERROR: resolver returned unexpected projection: ${resolved}" >&2
  exit 1
}

PGPASSWORD='pvnaive-direct-sub-ci' psql --no-psqlrc --set ON_ERROR_STOP=1 \
  --host "${PVNAIVE_DB_HOST}" --port "${PVNAIVE_DB_PORT}" --username pvnaive_app --dbname "${test_db}" >/dev/null <<'SQL'
BEGIN;
SELECT pvnaive.set_request_context(decode(repeat('16',32),'hex'));
UPDATE pvnaive.users SET status='suspended' WHERE id='d6000000-0000-0000-0000-000000000001';
COMMIT;
SQL
suspended_count="$(PGPASSWORD='pvnaive-direct-sub-ci' psql --no-psqlrc --set ON_ERROR_STOP=1 --host "${PVNAIVE_DB_HOST}" --port "${PVNAIVE_DB_PORT}" --username pvnaive_app --dbname "${test_db}" --tuples-only --no-align --command "SELECT count(*) FROM pvnaive.resolve_direct_subscription_token(decode(repeat('76',32),'hex'));")"
[[ "${suspended_count}" == "0" ]] || { echo "ERROR: suspended user subscription still resolved" >&2; exit 1; }

PGPASSWORD='pvnaive-direct-sub-ci' psql --no-psqlrc --set ON_ERROR_STOP=1 \
  --host "${PVNAIVE_DB_HOST}" --port "${PVNAIVE_DB_PORT}" --username pvnaive_app --dbname "${test_db}" >/dev/null <<'SQL'
BEGIN;
SELECT pvnaive.set_request_context(decode(repeat('16',32),'hex'));
UPDATE pvnaive.users SET status='active' WHERE id='d6000000-0000-0000-0000-000000000001';
UPDATE pvnaive.naive_runtime_credentials
   SET username='sub-user-rotated',
       secret_ciphertext=decode(repeat('57',16),'hex'),
       secret_nonce=decode(repeat('67',12),'hex')
 WHERE id='e6000000-0000-0000-0000-000000000001';
COMMIT;
SQL
rotated="$(resolve_projection)"
[[ "${rotated}" == "e6000000-0000-0000-0000-000000000001|sub-user-rotated|active|pending|runtime-v1" ]] || {
  echo "ERROR: runtime rotation did not synchronize subscription projection: ${rotated}" >&2
  exit 1
}

PGPASSWORD='pvnaive-direct-sub-ci' psql --no-psqlrc --set ON_ERROR_STOP=1 \
  --host "${PVNAIVE_DB_HOST}" --port "${PVNAIVE_DB_PORT}" --username pvnaive_app --dbname "${test_db}" >/dev/null <<'SQL'
BEGIN;
SELECT pvnaive.set_request_context(decode(repeat('16',32),'hex'));
UPDATE pvnaive.naive_runtime_credentials SET status='disabled' WHERE id='e6000000-0000-0000-0000-000000000001';
COMMIT;
SQL
revoked_count="$(PGPASSWORD='pvnaive-direct-sub-ci' psql --no-psqlrc --set ON_ERROR_STOP=1 --host "${PVNAIVE_DB_HOST}" --port "${PVNAIVE_DB_PORT}" --username pvnaive_app --dbname "${test_db}" --tuples-only --no-align --command "SELECT count(*) FROM pvnaive.resolve_direct_subscription_token(decode(repeat('76',32),'hex'));")"
[[ "${revoked_count}" == "0" ]] || { echo "ERROR: disabled runtime credential subscription still resolved" >&2; exit 1; }
token_state="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command "SELECT status || '|' || (revoked_at IS NOT NULL) FROM pvnaive.direct_subscription_tokens WHERE token_hash=decode(repeat('76',32),'hex')")"
[[ "${token_state}" == "revoked|true" || "${token_state}" == "revoked|t" ]] || { echo "ERROR: runtime disable did not revoke subscription token: ${token_state}" >&2; exit 1; }

missing="$(PGPASSWORD='pvnaive-direct-sub-ci' psql --no-psqlrc --set ON_ERROR_STOP=1 \
  --host "${PVNAIVE_DB_HOST}" --port "${PVNAIVE_DB_PORT}" --username pvnaive_app --dbname "${test_db}" \
  --tuples-only --no-align --command "SELECT count(*) FROM pvnaive.resolve_direct_subscription_token(decode(repeat('77',32),'hex'));")"
[[ "${missing}" == "0" ]] || { echo "ERROR: unknown token resolved" >&2; exit 1; }

echo "DIRECT_SUBSCRIPTION_MIGRATION_TEST=PASSED"
