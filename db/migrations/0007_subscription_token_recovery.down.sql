-- pvnaive:migration-version 0007
-- Source: PVNaive Owner read-only subscription delivery rollback
-- pvnaive:migration-name subscription_token_recovery
-- pvnaive:transactional true
-- pvnaive:destructive true

SET LOCAL lock_timeout = '10s';
SET LOCAL statement_timeout = '120s';
SET LOCAL idle_in_transaction_session_timeout = '60s';
SET LOCAL ROLE pvnaive_owner;

ALTER TABLE pvnaive.direct_subscription_tokens
    DROP CONSTRAINT IF EXISTS direct_subscription_token_recovery_envelope_check,
    DROP COLUMN IF EXISTS token_encryption_key_id,
    DROP COLUMN IF EXISTS token_nonce,
    DROP COLUMN IF EXISTS token_ciphertext;
