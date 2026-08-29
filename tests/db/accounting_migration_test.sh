#!/usr/bin/env bash
set -Eeuo pipefail
umask 077

repo_root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd -P)"
up="${repo_root}/db/migrations/0007_exact_accounting.up.sql"
down="${repo_root}/db/migrations/0007_exact_accounting.down.sql"

[[ -f "${up}" ]] || { echo 'ERROR: missing 0007_exact_accounting.up.sql' >&2; exit 1; }
[[ -f "${down}" ]] || { echo 'ERROR: missing 0007_exact_accounting.down.sql' >&2; exit 1; }
grep -Fqx -- '-- pvnaive:migration-version 0007' "${up}"
grep -Fqx -- '-- pvnaive:transactional true' "${up}"
grep -Fqx -- '-- pvnaive:destructive false' "${up}"
grep -Fqx -- '-- pvnaive:migration-version 0007' "${down}"
grep -Fqx -- '-- pvnaive:transactional true' "${down}"
grep -Fqx -- '-- pvnaive:destructive true' "${down}"

: "${PVNAIVE_DB_HOST:=127.0.0.1}"
: "${PVNAIVE_DB_PORT:=5432}"
: "${PVNAIVE_DB_USER:=postgres}"
export PVNAIVE_DB_HOST PVNAIVE_DB_PORT PVNAIVE_DB_USER

test_suffix="${GITHUB_RUN_ID:-local}_${GITHUB_RUN_ATTEMPT:-1}_${BASHPID}"
test_suffix="${test_suffix//[^a-zA-Z0-9_]/_}"
test_db="pvnaive_migration_test_accounting_${test_suffix,,}"

psql_admin() {
  psql --no-psqlrc --set ON_ERROR_STOP=1 \
    --host "${PVNAIVE_DB_HOST}" --port "${PVNAIVE_DB_PORT}" \
    --username "${PVNAIVE_DB_USER}" "$@"
}

psql_app() {
  PGPASSWORD='pvnaive-accounting-ci' psql --no-psqlrc --set ON_ERROR_STOP=1 \
    --host "${PVNAIVE_DB_HOST}" --port "${PVNAIVE_DB_PORT}" \
    --username pvnaive_app --dbname "${test_db}" "$@"
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
CREATE ROLE pvnaive_app LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOREPLICATION NOBYPASSRLS PASSWORD 'pvnaive-accounting-ci';
ALTER ROLE pvnaive_app SET row_security = on;
SQL

createdb --host "${PVNAIVE_DB_HOST}" --port "${PVNAIVE_DB_PORT}" \
  --username "${PVNAIVE_DB_USER}" --owner pvnaive_owner --encoding UTF8 --template template0 "${test_db}"
export PVNAIVE_DB_NAME="${test_db}"
"${repo_root}/scripts/db/migrate.sh" >/dev/null

schema_version="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command \
  'SELECT COALESCE(MAX(version),0) FROM pvnaive.schema_migrations')"
[[ "${schema_version}" == 7 ]] || { echo "ERROR: schema version=${schema_version}, want=7" >&2; exit 1; }

schema_contract="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command "
SELECT
  (to_regclass('pvnaive.usage_counters') IS NOT NULL) || '|' ||
  (to_regclass('pvnaive.usage_connection_sequences') IS NOT NULL) || '|' ||
  EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema='pvnaive' AND table_name='direct_subscription_tokens' AND column_name='token_ciphertext') || '|' ||
  EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema='pvnaive' AND table_name='direct_subscription_tokens' AND column_name='token_nonce') || '|' ||
  EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema='pvnaive' AND table_name='direct_subscription_tokens' AND column_name='token_encryption_key_id') || '|' ||
  EXISTS (SELECT 1 FROM pg_proc p JOIN pg_namespace n ON n.oid=p.pronamespace WHERE n.nspname='pvnaive' AND p.proname='accounting_authorize' AND p.prosecdef) || '|' ||
  EXISTS (SELECT 1 FROM pg_proc p JOIN pg_namespace n ON n.oid=p.pronamespace WHERE n.nspname='pvnaive' AND p.proname='accounting_apply_delta' AND p.prosecdef) || '|' ||
  (SELECT relrowsecurity AND NOT relforcerowsecurity FROM pg_class WHERE oid='pvnaive.usage_counters'::regclass);")"
[[ "${schema_contract}" == 'true|true|true|true|true|true|true|true' || "${schema_contract}" == 't|t|t|t|t|t|t|t' ]] || {
  echo "ERROR: schema7 accounting contract failed: ${schema_contract}" >&2
  exit 1
}

psql_admin --dbname "${test_db}" >/dev/null <<'SQL'
INSERT INTO pvnaive.actors (id, tenant_id, actor_role, email, display_name, password_hash, status)
VALUES (
  'a7000000-0000-0000-0000-000000000001', NULL, 'owner',
  'accounting-owner@example.invalid', 'Accounting Owner',
  '$argon2id$v=19$m=19456,t=2,p=1$AAAAAAAAAAAAAAAAAAAAAA$AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA',
  'active'
);
INSERT INTO pvnaive.auth_sessions (
  id, tenant_id, actor_id, token_hash, refresh_family_id, user_agent_hash,
  expires_at, absolute_expires_at, csrf_token_hash
) VALUES (
  'b7000000-0000-0000-0000-000000000001', NULL,
  'a7000000-0000-0000-0000-000000000001', decode(repeat('17',32),'hex'),
  'c7000000-0000-0000-0000-000000000001', decode(repeat('27',32),'hex'),
  clock_timestamp() + interval '1 hour', clock_timestamp() + interval '12 hours',
  decode(repeat('37',32),'hex')
);
SQL

psql_app >/dev/null <<'SQL'
BEGIN;
SELECT pvnaive.set_request_context(decode(repeat('17',32),'hex'));

INSERT INTO pvnaive.users (id, tenant_id, username, display_name, status, created_by_actor_id)
SELECT 'd7000000-0000-0000-0000-000000000001', id, 'acct-user', 'Accounting User', 'active',
       'a7000000-0000-0000-0000-000000000001'
FROM pvnaive.tenants WHERE tenant_type='system' AND slug='direct';

INSERT INTO pvnaive.naive_runtime_credentials (
  id, username, secret_hash, secret_ciphertext, secret_nonce, encryption_key_id,
  status, origin, created_by_actor_id, updated_by_actor_id
) VALUES (
  'e7000000-0000-0000-0000-000000000001', 'acct-user', decode(repeat('47',32),'hex'),
  decode(repeat('57',16),'hex'), decode(repeat('67',12),'hex'), 'runtime-v1',
  'active', 'panel', 'a7000000-0000-0000-0000-000000000001', 'a7000000-0000-0000-0000-000000000001'
),(
  'e7000000-0000-0000-0000-000000000002', 'unmanaged-user', decode(repeat('48',32),'hex'),
  decode(repeat('58',16),'hex'), decode(repeat('68',12),'hex'), 'runtime-v1',
  'active', 'panel', 'a7000000-0000-0000-0000-000000000001', 'a7000000-0000-0000-0000-000000000001'
);

INSERT INTO pvnaive.service_terms (
  id, tenant_id, user_id, quota_bytes, duration_seconds, start_policy,
  purchased_at, starts_at, expires_at, state
)
SELECT 'f7000000-0000-0000-0000-000000000001', u.tenant_id, u.id,
       1000, 2592000, 'fixed_timestamp',
       clock_timestamp() - interval '2 days', clock_timestamp() - interval '2 days',
       clock_timestamp() + interval '30 days', 'active'
FROM pvnaive.users u WHERE u.id='d7000000-0000-0000-0000-000000000001';

INSERT INTO pvnaive.user_runtime_credentials (
  tenant_id, user_id, service_term_id, runtime_credential_id, role
)
SELECT u.tenant_id, u.id, 'f7000000-0000-0000-0000-000000000001',
       'e7000000-0000-0000-0000-000000000001', 'primary'
FROM pvnaive.users u WHERE u.id='d7000000-0000-0000-0000-000000000001';

-- A schema-6 style token remains valid and explicitly unrecoverable until reissued.
INSERT INTO pvnaive.direct_subscription_tokens (
  tenant_id, user_id, service_term_id, runtime_credential_id,
  token_hash, token_prefix, status, user_state, service_state,
  runtime_username, secret_ciphertext, secret_nonce, encryption_key_id, expires_at
)
SELECT u.tenant_id, u.id, 'f7000000-0000-0000-0000-000000000001',
       'e7000000-0000-0000-0000-000000000001', decode(repeat('77',32),'hex'),
       'legacy77', 'active', 'active', 'active', r.username,
       r.secret_ciphertext, r.secret_nonce, r.encryption_key_id, clock_timestamp() + interval '30 days'
FROM pvnaive.users u
JOIN pvnaive.naive_runtime_credentials r ON r.id='e7000000-0000-0000-0000-000000000001'
WHERE u.id='d7000000-0000-0000-0000-000000000001';
COMMIT;
SQL

counter="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command "
SELECT upload_bytes || '|' || download_bytes || '|' || active_binding
FROM pvnaive.usage_counters
WHERE service_term_id='f7000000-0000-0000-0000-000000000001';")"
[[ "${counter}" == '0|0|true' || "${counter}" == '0|0|t' ]] || {
  echo "ERROR: managed binding was not initialized at exact zero: ${counter}" >&2
  exit 1
}

legacy_recovery="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command "
SELECT (token_ciphertext IS NULL) || '|' || (token_nonce IS NULL) || '|' || (token_encryption_key_id IS NULL)
FROM pvnaive.direct_subscription_tokens
WHERE token_hash=decode(repeat('77',32),'hex');")"
[[ "${legacy_recovery}" == 'true|true|true' || "${legacy_recovery}" == 't|t|t' ]] || {
  echo "ERROR: legacy subscription token unexpectedly gained fabricated recovery material: ${legacy_recovery}" >&2
  exit 1
}

authorized="$(psql_app --tuples-only --no-align --command "
SELECT tracked || '|' || allowed || '|' || reason || '|' || quota_bytes || '|' || used_bytes || '|' || remaining_bytes
FROM pvnaive.accounting_authorize('e7000000-0000-0000-0000-000000000001');")"
[[ "${authorized}" == 'true|true|allowed|1000|0|1000' || "${authorized}" == 't|t|allowed|1000|0|1000' ]] || {
  echo "ERROR: initial accounting authorization mismatch: ${authorized}" >&2
  exit 1
}

unmanaged="$(psql_app --tuples-only --no-align --command "
SELECT tracked || '|' || allowed || '|' || reason
FROM pvnaive.accounting_authorize('e7000000-0000-0000-0000-000000000002');")"
[[ "${unmanaged}" == 'false|true|unmanaged' || "${unmanaged}" == 'f|t|unmanaged' ]] || {
  echo "ERROR: unmanaged Runtime compatibility mismatch: ${unmanaged}" >&2
  exit 1
}
unmanaged_rows="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command "SELECT count(*) FROM pvnaive.usage_counters WHERE runtime_credential_id='e7000000-0000-0000-0000-000000000002';")"
[[ "${unmanaged_rows}" == 0 ]] || { echo 'ERROR: unmanaged Runtime gained a fabricated usage row' >&2; exit 1; }

first_delta="$(psql_app --tuples-only --no-align --command "
SELECT tracked || '|' || accepted || '|' || idempotent || '|' || continue_allowed || '|' || reason || '|' || upload_bytes || '|' || download_bytes || '|' || used_bytes || '|' || remaining_bytes
FROM pvnaive.accounting_apply_delta(
  'e7000000-0000-0000-0000-000000000001',
  '17000000-0000-0000-0000-000000000001', 1, 100, 50
);")"
[[ "${first_delta}" == 'true|true|false|true|allowed|100|50|150|850' || "${first_delta}" == 't|t|f|t|allowed|100|50|150|850' ]] || {
  echo "ERROR: first exact delta mismatch: ${first_delta}" >&2
  exit 1
}

duplicate="$(psql_app --tuples-only --no-align --command "
SELECT accepted || '|' || idempotent || '|' || upload_bytes || '|' || download_bytes || '|' || used_bytes
FROM pvnaive.accounting_apply_delta(
  'e7000000-0000-0000-0000-000000000001',
  '17000000-0000-0000-0000-000000000001', 1, 100, 50
);")"
[[ "${duplicate}" == 'true|true|100|50|150' || "${duplicate}" == 't|t|100|50|150' ]] || {
  echo "ERROR: duplicate sequence was not idempotent: ${duplicate}" >&2
  exit 1
}

if psql_app --command "SELECT * FROM pvnaive.accounting_apply_delta('e7000000-0000-0000-0000-000000000001','17000000-0000-0000-0000-000000000001',3,1,1);" >/dev/null 2>&1; then
  echo 'ERROR: sequence gap was accepted' >&2
  exit 1
fi
if psql_app --command "SELECT * FROM pvnaive.accounting_apply_delta('e7000000-0000-0000-0000-000000000001','17000000-0000-0000-0000-000000000001',2,-1,0);" >/dev/null 2>&1; then
  echo 'ERROR: negative accounting delta was accepted' >&2
  exit 1
fi

exhausted="$(psql_app --tuples-only --no-align --command "
SELECT accepted || '|' || idempotent || '|' || continue_allowed || '|' || reason || '|' || used_bytes || '|' || remaining_bytes
FROM pvnaive.accounting_apply_delta(
  'e7000000-0000-0000-0000-000000000001',
  '17000000-0000-0000-0000-000000000001', 2, 700, 200
);")"
[[ "${exhausted}" == 'true|false|false|quota_depleted|1050|0' || "${exhausted}" == 't|f|f|quota_depleted|1050|0' ]] || {
  echo "ERROR: quota depletion result mismatch: ${exhausted}" >&2
  exit 1
}

denied="$(psql_app --tuples-only --no-align --command "
SELECT tracked || '|' || allowed || '|' || reason || '|' || used_bytes || '|' || remaining_bytes
FROM pvnaive.accounting_authorize('e7000000-0000-0000-0000-000000000001');")"
[[ "${denied}" == 'true|false|quota_depleted|1050|0' || "${denied}" == 't|f|quota_depleted|1050|0' ]] || {
  echo "ERROR: exhausted account was not denied: ${denied}" >&2
  exit 1
}

psql_app >/dev/null <<'SQL'
BEGIN;
SELECT pvnaive.set_request_context(decode(repeat('17',32),'hex'));
UPDATE pvnaive.service_terms
SET quota_bytes=2000
WHERE id='f7000000-0000-0000-0000-000000000001';
COMMIT;
SQL
resumed="$(psql_app --tuples-only --no-align --command "
SELECT allowed || '|' || reason || '|' || used_bytes || '|' || remaining_bytes
FROM pvnaive.accounting_authorize('e7000000-0000-0000-0000-000000000001');")"
[[ "${resumed}" == 'true|allowed|1050|950' || "${resumed}" == 't|allowed|1050|950' ]] || {
  echo "ERROR: quota extension did not resume same Runtime credential: ${resumed}" >&2
  exit 1
}

psql_app >/dev/null <<'SQL'
BEGIN;
SELECT pvnaive.set_request_context(decode(repeat('17',32),'hex'));
UPDATE pvnaive.service_terms
SET expires_at=clock_timestamp() - interval '1 day'
WHERE id='f7000000-0000-0000-0000-000000000001';
COMMIT;
SQL
expired="$(psql_app --tuples-only --no-align --command "
SELECT allowed || '|' || reason FROM pvnaive.accounting_authorize('e7000000-0000-0000-0000-000000000001');")"
[[ "${expired}" == 'false|expired' || "${expired}" == 'f|expired' ]] || {
  echo "ERROR: expired policy was not denied: ${expired}" >&2
  exit 1
}

# Restore a non-expired unlimited policy, then prove BIGINT overflow is rejected.
psql_app >/dev/null <<'SQL'
BEGIN;
SELECT pvnaive.set_request_context(decode(repeat('17',32),'hex'));
UPDATE pvnaive.service_terms
SET quota_bytes=NULL, expires_at=clock_timestamp() + interval '30 days'
WHERE id='f7000000-0000-0000-0000-000000000001';
COMMIT;
SQL
if psql_app --command "SELECT * FROM pvnaive.accounting_apply_delta('e7000000-0000-0000-0000-000000000001','17000000-0000-0000-0000-000000000001',3,9223372036854775807,0);" >/dev/null 2>&1; then
  echo 'ERROR: BIGINT usage overflow was accepted' >&2
  exit 1
fi
post_overflow="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command "SELECT upload_bytes || '|' || download_bytes FROM pvnaive.usage_counters WHERE service_term_id='f7000000-0000-0000-0000-000000000001';")"
[[ "${post_overflow}" == '800|250' ]] || { echo "ERROR: overflow attempt mutated counters: ${post_overflow}" >&2; exit 1; }

# Schema-7 rollback on a disposable DB must preserve schema-6 customer/token data.
PVNAIVE_DISPOSABLE_DB=1 PVNAIVE_DB_NAME="${test_db}" "${repo_root}/scripts/db/rollback.sh" >/dev/null
rolled_version="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command 'SELECT COALESCE(MAX(version),0) FROM pvnaive.schema_migrations')"
[[ "${rolled_version}" == 6 ]] || { echo "ERROR: rollback schema=${rolled_version}, want=6" >&2; exit 1; }
post_rollback="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command "
SELECT
  (to_regclass('pvnaive.usage_counters') IS NULL) || '|' ||
  (to_regclass('pvnaive.direct_subscription_tokens') IS NOT NULL) || '|' ||
  EXISTS (SELECT 1 FROM pvnaive.direct_subscription_tokens WHERE token_hash=decode(repeat('77',32),'hex')) || '|' ||
  NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema='pvnaive' AND table_name='direct_subscription_tokens' AND column_name='token_ciphertext');")"
[[ "${post_rollback}" == 'true|true|true|true' || "${post_rollback}" == 't|t|t|t' ]] || {
  echo "ERROR: schema7 rollback damaged schema6 state: ${post_rollback}" >&2
  exit 1
}

echo 'PVNAIVE_ACCOUNTING_MIGRATION_TEST=PASSED'
