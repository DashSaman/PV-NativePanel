-- pvnaive:migration-version 0003
-- Source: PVNaive Naive runtime credential management
-- pvnaive:migration-name naive_runtime_credentials
-- pvnaive:transactional true
-- pvnaive:destructive false

SET LOCAL lock_timeout = '10s';
SET LOCAL statement_timeout = '120s';
SET LOCAL idle_in_transaction_session_timeout = '60s';
SET LOCAL ROLE pvnaive_owner;

CREATE TABLE pvnaive.naive_runtime_credentials (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    username text NOT NULL CHECK (length(username) BETWEEN 1 AND 64),
    secret_hash bytea NOT NULL CHECK (octet_length(secret_hash) = 32),
    secret_ciphertext bytea NOT NULL CHECK (octet_length(secret_ciphertext) >= 16),
    secret_nonce bytea NOT NULL CHECK (octet_length(secret_nonce) = 12),
    encryption_key_id text NOT NULL CHECK (length(encryption_key_id) BETWEEN 1 AND 160),
    status text NOT NULL CHECK (status IN ('active', 'disabled', 'revoked')),
    origin text NOT NULL CHECK (origin IN ('imported', 'panel')),
    created_by_actor_id uuid REFERENCES pvnaive.actors(id) ON DELETE RESTRICT,
    updated_by_actor_id uuid REFERENCES pvnaive.actors(id) ON DELETE RESTRICT,
    revision bigint NOT NULL DEFAULT 1 CHECK (revision > 0),
    created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    updated_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    rotated_at timestamptz,
    revoked_at timestamptz
);

CREATE UNIQUE INDEX naive_runtime_credentials_username_uidx
    ON pvnaive.naive_runtime_credentials (username);

ALTER TABLE pvnaive.naive_runtime_credentials ENABLE ROW LEVEL SECURITY;
ALTER TABLE pvnaive.naive_runtime_credentials FORCE ROW LEVEL SECURITY;
CREATE POLICY naive_runtime_credentials_owner_only
    ON pvnaive.naive_runtime_credentials
    USING (pvnaive.current_actor_role() = 'owner')
    WITH CHECK (pvnaive.current_actor_role() = 'owner');

GRANT SELECT, INSERT, UPDATE ON pvnaive.naive_runtime_credentials TO pvnaive_app;
REVOKE DELETE ON pvnaive.naive_runtime_credentials FROM pvnaive_app;

ALTER TABLE pvnaive.runtime_revisions
    ADD COLUMN idempotency_key text
        CHECK (idempotency_key IS NULL OR length(idempotency_key) BETWEEN 8 AND 160);

CREATE UNIQUE INDEX runtime_revisions_scope_idempotency_uidx
    ON pvnaive.runtime_revisions (
        COALESCE(tenant_id, '00000000-0000-0000-0000-000000000000'::uuid),
        protocol_id,
        idempotency_key
    )
    WHERE idempotency_key IS NOT NULL;

COMMENT ON TABLE pvnaive.naive_runtime_credentials
    IS 'Owner-only global Naive runtime credential state for the pre-S05 runtime bridge';
COMMENT ON COLUMN pvnaive.runtime_revisions.idempotency_key
    IS 'Durable mutation idempotency key scoped by tenant/protocol';
