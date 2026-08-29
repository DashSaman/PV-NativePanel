-- pvnaive:migration-version 0008
-- Source: PVNaive public subscription profile projection
-- pvnaive:migration-name subscription_profile_projection
-- pvnaive:transactional true
-- pvnaive:destructive false

SET LOCAL lock_timeout = '10s';
SET LOCAL statement_timeout = '120s';
SET LOCAL idle_in_transaction_session_timeout = '60s';
SET LOCAL ROLE pvnaive_owner;

ALTER TABLE pvnaive.direct_subscription_tokens
    ADD COLUMN quota_bytes bigint CHECK (quota_bytes IS NULL OR quota_bytes > 0),
    ADD COLUMN duration_seconds bigint CHECK (duration_seconds IS NULL OR duration_seconds > 0),
    ADD COLUMN start_policy text CHECK (
        start_policy IS NULL OR start_policy IN ('on_creation', 'on_first_successful_connection', 'fixed_timestamp')
    ),
    ADD COLUMN starts_at timestamptz,
    ADD COLUMN first_connected_at timestamptz;

-- Backfill from the authoritative service term while the table is locked. The
-- FORCE-RLS relaxation is transactional and restored before commit, matching
-- the bootstrap pattern already used by the lifecycle migration.
LOCK TABLE pvnaive.service_terms IN SHARE ROW EXCLUSIVE MODE;
ALTER TABLE pvnaive.service_terms NO FORCE ROW LEVEL SECURITY;
UPDATE pvnaive.direct_subscription_tokens AS dst
   SET quota_bytes = st.quota_bytes,
       duration_seconds = st.duration_seconds,
       start_policy = st.start_policy,
       starts_at = st.starts_at,
       first_connected_at = st.first_connected_at,
       expires_at = st.expires_at,
       service_state = st.state
  FROM pvnaive.service_terms AS st
 WHERE st.id = dst.service_term_id
   AND st.tenant_id = dst.tenant_id
   AND st.user_id = dst.user_id;
ALTER TABLE pvnaive.service_terms FORCE ROW LEVEL SECURITY;

-- Keep the existing state/expiry trigger intact and add a complementary
-- non-destructive trigger for commercial profile metadata.
CREATE FUNCTION pvnaive.sync_direct_subscription_service_profile()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog, pvnaive
AS $$
BEGIN
    IF NEW.quota_bytes IS DISTINCT FROM OLD.quota_bytes
       OR NEW.duration_seconds IS DISTINCT FROM OLD.duration_seconds
       OR NEW.start_policy IS DISTINCT FROM OLD.start_policy
       OR NEW.starts_at IS DISTINCT FROM OLD.starts_at
       OR NEW.first_connected_at IS DISTINCT FROM OLD.first_connected_at THEN
        UPDATE pvnaive.direct_subscription_tokens
           SET quota_bytes = NEW.quota_bytes,
               duration_seconds = NEW.duration_seconds,
               start_policy = NEW.start_policy,
               starts_at = NEW.starts_at,
               first_connected_at = NEW.first_connected_at
         WHERE service_term_id = NEW.id
           AND tenant_id = NEW.tenant_id
           AND user_id = NEW.user_id
           AND status = 'active'
           AND revoked_at IS NULL;
    END IF;
    RETURN NEW;
END;
$$;

CREATE TRIGGER direct_subscription_service_profile_sync
AFTER UPDATE OF quota_bytes, duration_seconds, start_policy, starts_at, first_connected_at
ON pvnaive.service_terms
FOR EACH ROW
EXECUTE FUNCTION pvnaive.sync_direct_subscription_service_profile();

CREATE FUNCTION pvnaive.resolve_direct_subscription_profile(p_token_hash bytea)
RETURNS TABLE (
    runtime_credential_id uuid,
    runtime_username text,
    user_state text,
    service_state text,
    secret_ciphertext bytea,
    secret_nonce bytea,
    encryption_key_id text,
    quota_bytes bigint,
    duration_seconds bigint,
    start_policy text,
    starts_at timestamptz,
    first_connected_at timestamptz,
    expires_at timestamptz
)
LANGUAGE sql
STABLE
SECURITY DEFINER
SET search_path = pg_catalog, pvnaive
AS $$
    SELECT
       dst.runtime_credential_id,
       dst.runtime_username,
       dst.user_state,
       dst.service_state,
       dst.secret_ciphertext,
       dst.secret_nonce,
       dst.encryption_key_id,
       dst.quota_bytes,
       dst.duration_seconds,
       dst.start_policy,
       dst.starts_at,
       dst.first_connected_at,
       dst.expires_at
      FROM pvnaive.direct_subscription_tokens AS dst
     WHERE p_token_hash IS NOT NULL
       AND octet_length(p_token_hash) = 32
       AND dst.token_hash = p_token_hash
       AND dst.status = 'active'
       AND dst.revoked_at IS NULL
     LIMIT 1;
$$;

REVOKE ALL ON FUNCTION pvnaive.resolve_direct_subscription_profile(bytea) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION pvnaive.resolve_direct_subscription_profile(bytea) TO pvnaive_app;

REVOKE ALL ON FUNCTION pvnaive.sync_direct_subscription_service_profile() FROM PUBLIC;
REVOKE ALL ON FUNCTION pvnaive.sync_direct_subscription_service_profile() FROM pvnaive_app;

COMMENT ON FUNCTION pvnaive.resolve_direct_subscription_profile(bytea) IS
    'Resolves token-bound customer status/quota/validity metadata for the public subscription status page without bypassing management RLS.';
