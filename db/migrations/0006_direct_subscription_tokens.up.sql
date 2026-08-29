-- pvnaive:migration-version 0006
-- Source: PVNaive direct subscription token projection
-- pvnaive:migration-name direct_subscription_tokens
-- pvnaive:transactional true
-- pvnaive:destructive false

SET LOCAL lock_timeout = '10s';
SET LOCAL statement_timeout = '120s';
SET LOCAL idle_in_transaction_session_timeout = '60s';
SET LOCAL ROLE pvnaive_owner;

CREATE TABLE pvnaive.direct_subscription_tokens (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id uuid NOT NULL REFERENCES pvnaive.tenants(id) ON DELETE RESTRICT,
    user_id uuid NOT NULL,
    service_term_id uuid NOT NULL,
    runtime_credential_id uuid NOT NULL
        REFERENCES pvnaive.naive_runtime_credentials(id) ON DELETE RESTRICT,
    token_hash bytea NOT NULL UNIQUE CHECK (octet_length(token_hash) = 32),
    token_prefix text NOT NULL CHECK (length(token_prefix) BETWEEN 6 AND 16),
    status text NOT NULL DEFAULT 'active'
        CHECK (status IN ('active', 'revoked', 'expired')),
    user_state text NOT NULL
        CHECK (user_state IN ('draft', 'active', 'suspended', 'expired', 'depleted', 'revoked')),
    service_state text NOT NULL
        CHECK (service_state IN ('pending', 'active', 'expired', 'quota_depleted', 'ended', 'revoked')),
    runtime_username text NOT NULL CHECK (length(runtime_username) BETWEEN 1 AND 64),
    secret_ciphertext bytea NOT NULL CHECK (octet_length(secret_ciphertext) >= 16),
    secret_nonce bytea NOT NULL CHECK (octet_length(secret_nonce) = 12),
    encryption_key_id text NOT NULL CHECK (length(encryption_key_id) BETWEEN 1 AND 160),
    expires_at timestamptz,
    last_accessed_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    revoked_at timestamptz,
    UNIQUE (id, tenant_id),
    FOREIGN KEY (user_id, tenant_id)
        REFERENCES pvnaive.users(id, tenant_id) ON DELETE RESTRICT,
    FOREIGN KEY (service_term_id, tenant_id, user_id)
        REFERENCES pvnaive.service_terms(id, tenant_id, user_id) ON DELETE RESTRICT,
    CHECK (revoked_at IS NULL OR revoked_at >= created_at)
);

CREATE INDEX direct_subscription_tokens_tenant_user_idx
    ON pvnaive.direct_subscription_tokens (tenant_id, user_id, created_at DESC);
CREATE UNIQUE INDEX direct_subscription_tokens_one_active_term_uidx
    ON pvnaive.direct_subscription_tokens (service_term_id)
    WHERE status = 'active' AND revoked_at IS NULL;

ALTER TABLE pvnaive.direct_subscription_tokens ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON pvnaive.direct_subscription_tokens
    USING (pvnaive.has_tenant_access(tenant_id, false))
    WITH CHECK (pvnaive.has_tenant_access(tenant_id, false));

-- The public resolver cannot safely join FORCE-RLS management tables without
-- privileged request context. Keep a tightly scoped projection instead and
-- synchronize it from trusted source-table mutations using SECURITY DEFINER
-- trigger functions owned by pvnaive_owner. The projection table itself is not
-- FORCE RLS, so its owner can maintain/resolve it without granting BYPASSRLS.
CREATE FUNCTION pvnaive.sync_direct_subscription_user()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog, pvnaive
AS $$
BEGIN
    IF NEW.status IS DISTINCT FROM OLD.status THEN
        UPDATE pvnaive.direct_subscription_tokens
           SET user_state = NEW.status
         WHERE user_id = NEW.id
           AND tenant_id = NEW.tenant_id
           AND status = 'active'
           AND revoked_at IS NULL;
    END IF;
    RETURN NEW;
END;
$$;

CREATE TRIGGER direct_subscription_user_sync
AFTER UPDATE OF status ON pvnaive.users
FOR EACH ROW
EXECUTE FUNCTION pvnaive.sync_direct_subscription_user();

CREATE FUNCTION pvnaive.sync_direct_subscription_service_term()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog, pvnaive
AS $$
BEGIN
    IF NEW.state IS DISTINCT FROM OLD.state
       OR NEW.expires_at IS DISTINCT FROM OLD.expires_at THEN
        UPDATE pvnaive.direct_subscription_tokens
           SET service_state = NEW.state,
               expires_at = NEW.expires_at
         WHERE service_term_id = NEW.id
           AND tenant_id = NEW.tenant_id
           AND user_id = NEW.user_id
           AND status = 'active'
           AND revoked_at IS NULL;
    END IF;
    RETURN NEW;
END;
$$;

CREATE TRIGGER direct_subscription_service_term_sync
AFTER UPDATE OF state, expires_at ON pvnaive.service_terms
FOR EACH ROW
EXECUTE FUNCTION pvnaive.sync_direct_subscription_service_term();

CREATE FUNCTION pvnaive.sync_direct_subscription_runtime_credential()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog, pvnaive
AS $$
BEGIN
    IF NEW.status <> 'active' THEN
        UPDATE pvnaive.direct_subscription_tokens
           SET status = 'revoked',
               revoked_at = COALESCE(revoked_at, clock_timestamp())
         WHERE runtime_credential_id = NEW.id
           AND status = 'active'
           AND revoked_at IS NULL;
        RETURN NEW;
    END IF;

    IF NEW.username IS DISTINCT FROM OLD.username
       OR NEW.secret_ciphertext IS DISTINCT FROM OLD.secret_ciphertext
       OR NEW.secret_nonce IS DISTINCT FROM OLD.secret_nonce
       OR NEW.encryption_key_id IS DISTINCT FROM OLD.encryption_key_id THEN
        UPDATE pvnaive.direct_subscription_tokens
           SET runtime_username = NEW.username,
               secret_ciphertext = NEW.secret_ciphertext,
               secret_nonce = NEW.secret_nonce,
               encryption_key_id = NEW.encryption_key_id
         WHERE runtime_credential_id = NEW.id
           AND status = 'active'
           AND revoked_at IS NULL;
    END IF;
    RETURN NEW;
END;
$$;

CREATE TRIGGER direct_subscription_runtime_credential_sync
AFTER UPDATE OF status, username, secret_ciphertext, secret_nonce, encryption_key_id
ON pvnaive.naive_runtime_credentials
FOR EACH ROW
EXECUTE FUNCTION pvnaive.sync_direct_subscription_runtime_credential();

CREATE FUNCTION pvnaive.resolve_direct_subscription_token(p_token_hash bytea)
RETURNS TABLE (
    tenant_id uuid,
    user_id uuid,
    service_term_id uuid,
    runtime_credential_id uuid,
    runtime_username text,
    user_state text,
    service_state text,
    secret_ciphertext bytea,
    secret_nonce bytea,
    encryption_key_id text,
    expires_at timestamptz
)
LANGUAGE sql
STABLE
SECURITY DEFINER
SET search_path = pg_catalog, pvnaive
AS $$
    SELECT
       dst.tenant_id,
       dst.user_id,
       dst.service_term_id,
       dst.runtime_credential_id,
       dst.runtime_username,
       dst.user_state,
       dst.service_state,
       dst.secret_ciphertext,
       dst.secret_nonce,
       dst.encryption_key_id,
       dst.expires_at
      FROM pvnaive.direct_subscription_tokens AS dst
     WHERE p_token_hash IS NOT NULL
       AND octet_length(p_token_hash) = 32
       AND dst.token_hash = p_token_hash
       AND dst.status = 'active'
       AND dst.revoked_at IS NULL
       AND dst.user_state = 'active'
       AND dst.service_state IN ('pending', 'active')
       AND (dst.expires_at IS NULL OR dst.expires_at > clock_timestamp())
     LIMIT 1;
$$;

REVOKE ALL ON pvnaive.direct_subscription_tokens FROM PUBLIC;
GRANT SELECT, INSERT, UPDATE ON pvnaive.direct_subscription_tokens TO pvnaive_app;
REVOKE DELETE ON pvnaive.direct_subscription_tokens FROM pvnaive_app;

REVOKE ALL ON FUNCTION pvnaive.resolve_direct_subscription_token(bytea) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION pvnaive.resolve_direct_subscription_token(bytea) TO pvnaive_app;

REVOKE ALL ON FUNCTION pvnaive.sync_direct_subscription_user() FROM PUBLIC;
REVOKE ALL ON FUNCTION pvnaive.sync_direct_subscription_service_term() FROM PUBLIC;
REVOKE ALL ON FUNCTION pvnaive.sync_direct_subscription_runtime_credential() FROM PUBLIC;

COMMENT ON TABLE pvnaive.direct_subscription_tokens IS
    'Revocable opaque-token projection for direct customer subscription delivery; raw tokens are never stored.';
COMMENT ON FUNCTION pvnaive.resolve_direct_subscription_token(bytea) IS
    'Resolves an active opaque subscription token from a trigger-synchronized projection without requiring management RLS context.';
