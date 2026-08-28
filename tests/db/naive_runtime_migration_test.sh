#!/usr/bin/env bash
set -Eeuo pipefail
umask 077

repo_root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd -P)"
up="${repo_root}/db/migrations/0003_naive_runtime_credentials.up.sql"
down="${repo_root}/db/migrations/0003_naive_runtime_credentials.down.sql"
temp_migrations="$(mktemp -d)"

[[ -f "${up}" ]] || { echo "ERROR: missing 0003_naive_runtime_credentials.up.sql" >&2; exit 1; }
[[ -f "${down}" ]] || { echo "ERROR: missing 0003_naive_runtime_credentials.down.sql" >&2; exit 1; }

grep -Fqx -- '-- pvnaive:migration-version 0003' "${up}"
grep -Fqx -- '-- pvnaive:transactional true' "${up}"
grep -Fqx -- '-- pvnaive:destructive false' "${up}"
grep -Fqx -- '-- pvnaive:migration-version 0003' "${down}"
grep -Fqx -- '-- pvnaive:transactional true' "${down}"
grep -Fqx -- '-- pvnaive:destructive true' "${down}"

cp "${repo_root}"/db/migrations/0001_* "${temp_migrations}/"
cp "${repo_root}"/db/migrations/0002_* "${temp_migrations}/"
cp "${repo_root}"/db/migrations/0003_* "${temp_migrations}/"
(
  cd "${temp_migrations}"
  sha256sum *.sql > SHA256SUMS
)

: "${PVNAIVE_DB_HOST:=127.0.0.1}"
: "${PVNAIVE_DB_PORT:=5432}"
: "${PVNAIVE_DB_USER:=postgres}"
export PVNAIVE_DB_HOST PVNAIVE_DB_PORT PVNAIVE_DB_USER

test_suffix="${GITHUB_RUN_ID:-local}_${GITHUB_RUN_ATTEMPT:-1}_${BASHPID}"
test_suffix="${test_suffix//[^a-zA-Z0-9_]/_}"
test_db="pvnaive_migration_test_naive_runtime_${test_suffix,,}"

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
  rm -rf -- "${temp_migrations}"
}
trap cleanup EXIT
cleanup

# cleanup removes the initial temporary directory while normalizing stale DB/roles;
# recreate the v3-only migration fixture after cleanup.
temp_migrations="$(mktemp -d)"
cp "${repo_root}"/db/migrations/0001_* "${temp_migrations}/"
cp "${repo_root}"/db/migrations/0002_* "${temp_migrations}/"
cp "${repo_root}"/db/migrations/0003_* "${temp_migrations}/"
(
  cd "${temp_migrations}"
  sha256sum *.sql > SHA256SUMS
)

psql_admin --dbname postgres <<'SQL' >/dev/null
CREATE ROLE pvnaive_owner NOLOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOREPLICATION NOBYPASSRLS;
CREATE ROLE pvnaive_app LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOREPLICATION NOBYPASSRLS PASSWORD 'pvnaive-naive-runtime-ci-only';
ALTER ROLE pvnaive_app SET row_security = on;
SQL
createdb --host "${PVNAIVE_DB_HOST}" --port "${PVNAIVE_DB_PORT}" \
  --username "${PVNAIVE_DB_USER}" --owner pvnaive_owner --encoding UTF8 --template template0 "${test_db}"

export PVNAIVE_DB_NAME="${test_db}"
PVNAIVE_MIGRATIONS_DIR="${temp_migrations}" "${repo_root}/scripts/db/migrate.sh" >/dev/null

schema_version="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command \
  'SELECT COALESCE(MAX(version),0) FROM pvnaive.schema_migrations')"
[[ "${schema_version}" == "3" ]] || { echo "ERROR: schema version=${schema_version}, want=3" >&2; exit 1; }

contract="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command "
SELECT
  (to_regclass('pvnaive.naive_runtime_credentials') IS NOT NULL) || '|' ||
  EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema='pvnaive' AND table_name='naive_runtime_credentials' AND column_name='secret_ciphertext') || '|' ||
  EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema='pvnaive' AND table_name='naive_runtime_credentials' AND column_name='secret_nonce') || '|' ||
  EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema='pvnaive' AND table_name='naive_runtime_credentials' AND column_name='revision') || '|' ||
  EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema='pvnaive' AND table_name='runtime_revisions' AND column_name='idempotency_key') || '|' ||
  (SELECT relforcerowsecurity FROM pg_class WHERE oid='pvnaive.naive_runtime_credentials'::regclass);")"
[[ "${contract}" == "true|true|true|true|true|true" || "${contract}" == "t|t|t|t|t|t" ]] || {
  echo "ERROR: naive runtime schema contract failed: ${contract}" >&2
  exit 1
}

psql_admin --dbname "${test_db}" --command "
INSERT INTO pvnaive.actors (id, tenant_id, actor_role, email, display_name, password_hash, status)
VALUES
('a1000000-0000-0000-0000-000000000001', NULL, 'owner', 'runtime-owner@example.invalid', 'Runtime Owner', '\$argon2id\$v=19\$m=19456,t=2,p=1\$AAAAAAAAAAAAAAAAAAAAAA\$AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA', 'active'),
('a1000000-0000-0000-0000-000000000002', NULL, 'admin', 'runtime-admin@example.invalid', 'Runtime Admin', '\$argon2id\$v=19\$m=19456,t=2,p=1\$AAAAAAAAAAAAAAAAAAAAAA\$AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA', 'active');

INSERT INTO pvnaive.auth_sessions (
  id, tenant_id, actor_id, token_hash, refresh_family_id, user_agent_hash,
  expires_at, absolute_expires_at, csrf_token_hash
) VALUES
(
 'b1000000-0000-0000-0000-000000000001', NULL, 'a1000000-0000-0000-0000-000000000001',
 decode(repeat('11',32),'hex'), 'c1000000-0000-0000-0000-000000000001', decode(repeat('21',32),'hex'),
 clock_timestamp() + interval '1 hour', clock_timestamp() + interval '12 hours', decode(repeat('31',32),'hex')
),
(
 'b1000000-0000-0000-0000-000000000002', NULL, 'a1000000-0000-0000-0000-000000000002',
 decode(repeat('12',32),'hex'), 'c1000000-0000-0000-0000-000000000002', decode(repeat('22',32),'hex'),
 clock_timestamp() + interval '1 hour', clock_timestamp() + interval '12 hours', decode(repeat('32',32),'hex')
);" >/dev/null

owner_count="$(PGPASSWORD='pvnaive-naive-runtime-ci-only' psql --no-psqlrc --set ON_ERROR_STOP=1 \
  --host "${PVNAIVE_DB_HOST}" --port "${PVNAIVE_DB_PORT}" --username pvnaive_app --dbname "${test_db}" \
  --tuples-only --no-align <<'SQL'
BEGIN;
SELECT pvnaive.set_request_context(decode(repeat('11',32),'hex'));
INSERT INTO pvnaive.naive_runtime_credentials (
  id, username, secret_hash, secret_ciphertext, secret_nonce, encryption_key_id,
  status, origin, created_by_actor_id, updated_by_actor_id
) VALUES (
  'd1000000-0000-0000-0000-000000000001', 'imported-owner', decode(repeat('41',32),'hex'),
  decode(repeat('51',16),'hex'), decode(repeat('61',12),'hex'), 'runtime-v1',
  'active', 'imported', 'a1000000-0000-0000-0000-000000000001', 'a1000000-0000-0000-0000-000000000001'
);
SELECT COUNT(*) FROM pvnaive.naive_runtime_credentials;
COMMIT;
SQL
)"
owner_count="$(printf '%s\n' "${owner_count}" | grep -E '^[0-9]+$' | tail -n1)"
[[ "${owner_count}" == "1" ]] || { echo "ERROR: owner context could not access runtime credentials: ${owner_count}" >&2; exit 1; }

admin_count="$(PGPASSWORD='pvnaive-naive-runtime-ci-only' psql --no-psqlrc --set ON_ERROR_STOP=1 \
  --host "${PVNAIVE_DB_HOST}" --port "${PVNAIVE_DB_PORT}" --username pvnaive_app --dbname "${test_db}" \
  --tuples-only --no-align <<'SQL'
BEGIN;
SELECT pvnaive.set_request_context(decode(repeat('12',32),'hex'));
SELECT COUNT(*) FROM pvnaive.naive_runtime_credentials;
ROLLBACK;
SQL
)"
admin_count="$(printf '%s\n' "${admin_count}" | grep -E '^[0-9]+$' | tail -n1)"
[[ "${admin_count}" == "0" ]] || { echo "ERROR: admin context can read owner-only runtime credentials: ${admin_count}" >&2; exit 1; }

set +e
PGPASSWORD='pvnaive-naive-runtime-ci-only' psql --no-psqlrc --set ON_ERROR_STOP=1 \
  --host "${PVNAIVE_DB_HOST}" --port "${PVNAIVE_DB_PORT}" --username pvnaive_app --dbname "${test_db}" \
  >/dev/null 2>&1 <<'SQL'
BEGIN;
SELECT pvnaive.set_request_context(decode(repeat('12',32),'hex'));
INSERT INTO pvnaive.naive_runtime_credentials (
  username, secret_hash, secret_ciphertext, secret_nonce, encryption_key_id,
  status, origin, created_by_actor_id, updated_by_actor_id
) VALUES (
  'admin-must-fail', decode(repeat('42',32),'hex'), decode(repeat('52',16),'hex'),
  decode(repeat('62',12),'hex'), 'runtime-v1', 'active', 'panel',
  'a1000000-0000-0000-0000-000000000002', 'a1000000-0000-0000-0000-000000000002'
);
COMMIT;
SQL
admin_insert_rc=$?
set -e
[[ "${admin_insert_rc}" -ne 0 ]] || { echo "ERROR: admin context inserted owner-only runtime credential" >&2; exit 1; }

psql_admin --dbname "${test_db}" --command "
INSERT INTO pvnaive.runtime_revisions (
  id, tenant_id, protocol_id, revision_no, state, config_checksum_sha256,
  config_ciphertext, encryption_key_id, manifest, created_by_actor_id, idempotency_key
) VALUES (
  'e1000000-0000-0000-0000-000000000001', NULL, 'naive', 1, 'staged', repeat('a',64),
  decode('00','hex'), 'runtime-v1', '{}'::jsonb, 'a1000000-0000-0000-0000-000000000001', 'idem-runtime-0001'
);" >/dev/null

set +e
psql_admin --dbname "${test_db}" --command "
INSERT INTO pvnaive.runtime_revisions (
  id, tenant_id, protocol_id, revision_no, state, config_checksum_sha256,
  config_ciphertext, encryption_key_id, manifest, created_by_actor_id, idempotency_key
) VALUES (
  'e1000000-0000-0000-0000-000000000002', NULL, 'naive', 2, 'staged', repeat('b',64),
  decode('01','hex'), 'runtime-v1', '{}'::jsonb, 'a1000000-0000-0000-0000-000000000001', 'idem-runtime-0001'
);" >/dev/null 2>&1
idempotency_rc=$?
set -e
[[ "${idempotency_rc}" -ne 0 ]] || { echo "ERROR: duplicate runtime idempotency key was accepted" >&2; exit 1; }

PVNAIVE_DISPOSABLE_DB=1 PVNAIVE_ALLOW_DESTRUCTIVE_ROLLBACK=ROLLBACK_ONE_MIGRATION \
  PVNAIVE_MIGRATIONS_DIR="${temp_migrations}" \
  "${repo_root}/scripts/db/rollback.sh" >/dev/null

remaining="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command \
  "SELECT (to_regclass('pvnaive.actor_totp_factors') IS NOT NULL) || '|' || (SELECT COALESCE(MAX(version),0) FROM pvnaive.schema_migrations)")"
[[ "${remaining}" == "true|2" || "${remaining}" == "t|2" ]] || {
  echo "ERROR: v3 rollback did not preserve v2: ${remaining}" >&2
  exit 1
}

table_after_rollback="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command \
  "SELECT to_regclass('pvnaive.naive_runtime_credentials') IS NULL")"
[[ "${table_after_rollback}" == "true" || "${table_after_rollback}" == "t" ]] || {
  echo "ERROR: naive runtime table survived v3 rollback" >&2
  exit 1
}

echo "PVNAIVE_NAIVE_RUNTIME_MIGRATION_TEST=PASSED"
