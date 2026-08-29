-- pvnaive:migration-version 0007
-- Source: PVNaive exact per-runtime accounting rollback
-- pvnaive:migration-name exact_accounting
-- pvnaive:transactional true
-- pvnaive:destructive true

SET LOCAL lock_timeout = '10s';
SET LOCAL statement_timeout = '120s';
SET LOCAL idle_in_transaction_session_timeout = '60s';
SET LOCAL ROLE pvnaive_owner;

DROP TRIGGER IF EXISTS usage_counter_runtime_sync ON pvnaive.naive_runtime_credentials;
DROP TRIGGER IF EXISTS usage_counter_service_term_sync ON pvnaive.service_terms;
DROP TRIGGER IF EXISTS usage_counter_user_sync ON pvnaive.users;
DROP TRIGGER IF EXISTS usage_counter_binding_sync ON pvnaive.user_runtime_credentials;

DROP FUNCTION IF EXISTS pvnaive.accounting_apply_delta(uuid, uuid, bigint, bigint, bigint);
DROP FUNCTION IF EXISTS pvnaive.accounting_authorize(uuid);
DROP FUNCTION IF EXISTS pvnaive.sync_usage_counter_runtime();
DROP FUNCTION IF EXISTS pvnaive.sync_usage_counter_service_term();
DROP FUNCTION IF EXISTS pvnaive.sync_usage_counter_user();
DROP FUNCTION IF EXISTS pvnaive.sync_usage_counter_binding();

DROP TABLE IF EXISTS pvnaive.usage_connection_sequences;
DROP TABLE IF EXISTS pvnaive.usage_counters;

ALTER TABLE pvnaive.direct_subscription_tokens
    DROP CONSTRAINT IF EXISTS direct_subscription_token_recovery_material_ck,
    DROP COLUMN IF EXISTS token_encryption_key_id,
    DROP COLUMN IF EXISTS token_nonce,
    DROP COLUMN IF EXISTS token_ciphertext;

DELETE FROM pvnaive.schema_migrations WHERE version = 7;
