-- pvnaive:migration-version 0007
-- Source: PVNaive Owner read-only subscription delivery
-- pvnaive:migration-name subscription_token_recovery
-- pvnaive:transactional true
-- pvnaive:destructive false

SET LOCAL lock_timeout = '10s';
SET LOCAL statement_timeout = '120s';
SET LOCAL idle_in_transaction_session_timeout = '60s';
SET LOCAL ROLE pvnaive_owner;

ALTER TABLE pvnaive.direct_subscription_tokens
    ADD COLUMN token_ciphertext bytea,
    ADD COLUMN token_nonce bytea,
    ADD COLUMN token_encryption_key_id text;

ALTER TABLE pvnaive.direct_subscription_tokens
    ADD CONSTRAINT direct_subscription_token_recovery_envelope_check
    CHECK (
        (token_ciphertext IS NULL AND token_nonce IS NULL AND token_encryption_key_id IS NULL)
        OR
        (
            token_ciphertext IS NOT NULL
            AND octet_length(token_ciphertext) >= 16
            AND token_nonce IS NOT NULL
            AND octet_length(token_nonce) = 12
            AND token_encryption_key_id IS NOT NULL
            AND length(token_encryption_key_id) BETWEEN 1 AND 160
        )
    );

COMMENT ON COLUMN pvnaive.direct_subscription_tokens.token_ciphertext IS
    'AES-GCM encrypted copy of the active opaque token for authenticated Owner read-only QR/subscription recovery; public lookup still uses token_hash.';
COMMENT ON COLUMN pvnaive.direct_subscription_tokens.token_nonce IS
    'AES-GCM nonce for token_ciphertext.';
COMMENT ON COLUMN pvnaive.direct_subscription_tokens.token_encryption_key_id IS
    'Application key identifier used for token_ciphertext; raw subscription tokens are never stored in plaintext.';
