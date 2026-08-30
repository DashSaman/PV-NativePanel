-- pvnaive:migration-version 0013
-- Source: PVNaive Task #6 public account accounting/presence projection
-- pvnaive:migration-name subscription_account_projection
-- pvnaive:transactional true
-- pvnaive:destructive false

SET LOCAL lock_timeout = '10s';
SET LOCAL statement_timeout = '120s';
SET LOCAL idle_in_transaction_session_timeout = '60s';
SET LOCAL ROLE pvnaive_owner;

ALTER TABLE pvnaive.direct_subscription_tokens
    ADD COLUMN accounting_baseline_state text,
    ADD COLUMN accounting_baseline_source text,
    ADD COLUMN accounting_baseline_cutoff_at timestamptz,
    ADD COLUMN accounting_baseline_upload_bytes bigint,
    ADD COLUMN accounting_baseline_download_bytes bigint;

-- service_terms uses FORCE RLS. The migration owner temporarily disables it
-- only to backfill the already-scoped public token projection, then restores
-- the exact protection before commit.
ALTER TABLE pvnaive.service_terms DISABLE ROW LEVEL SECURITY;
UPDATE pvnaive.direct_subscription_tokens AS dst
SET accounting_baseline_state = st.accounting_baseline_state,
    accounting_baseline_source = st.accounting_baseline_source,
    accounting_baseline_cutoff_at = st.accounting_baseline_cutoff_at,
    accounting_baseline_upload_bytes = st.accounting_baseline_upload_bytes,
    accounting_baseline_download_bytes = st.accounting_baseline_download_bytes
FROM pvnaive.service_terms AS st
WHERE st.id = dst.service_term_id
  AND st.tenant_id = dst.tenant_id
  AND st.user_id = dst.user_id;
ALTER TABLE pvnaive.service_terms ENABLE ROW LEVEL SECURITY;
ALTER TABLE pvnaive.service_terms FORCE ROW LEVEL SECURITY;

ALTER TABLE pvnaive.direct_subscription_tokens
    ALTER COLUMN accounting_baseline_state SET NOT NULL,
    ALTER COLUMN accounting_baseline_source SET NOT NULL,
    ALTER COLUMN accounting_baseline_cutoff_at SET NOT NULL,
    ADD CONSTRAINT direct_subscription_accounting_baseline_truth_check
    CHECK (
        (
            accounting_baseline_state = 'unknown'
            AND accounting_baseline_source = 'legacy_unavailable'
            AND accounting_baseline_upload_bytes IS NULL
            AND accounting_baseline_download_bytes IS NULL
        )
        OR
        (
            accounting_baseline_state = 'known'
            AND accounting_baseline_source = 'fresh_managed_term'
            AND accounting_baseline_upload_bytes = 0
            AND accounting_baseline_download_bytes = 0
        )
        OR
        (
            accounting_baseline_state = 'known'
            AND accounting_baseline_source = 'authoritative_import'
            AND accounting_baseline_upload_bytes IS NOT NULL
            AND accounting_baseline_download_bytes IS NOT NULL
            AND accounting_baseline_upload_bytes >= 0
            AND accounting_baseline_download_bytes >= 0
        )
    );

CREATE FUNCTION pvnaive.resolve_direct_subscription_account_profile(p_token_hash bytea)
RETURNS TABLE (
    service_term_id uuid,
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
    expires_at timestamptz,
    accounting_baseline_state text,
    accounting_baseline_source text,
    accounting_baseline_cutoff_at timestamptz,
    accounting_baseline_upload_bytes bigint,
    accounting_baseline_download_bytes bigint
)
LANGUAGE sql
STABLE
SECURITY DEFINER
SET search_path = pg_catalog, pvnaive
AS $$
    SELECT
       dst.service_term_id,
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
       dst.expires_at,
       dst.accounting_baseline_state,
       dst.accounting_baseline_source,
       dst.accounting_baseline_cutoff_at,
       dst.accounting_baseline_upload_bytes,
       dst.accounting_baseline_download_bytes
      FROM pvnaive.direct_subscription_tokens AS dst
     WHERE p_token_hash IS NOT NULL
       AND octet_length(p_token_hash) = 32
       AND dst.token_hash = p_token_hash
       AND dst.status = 'active'
       AND dst.revoked_at IS NULL
     LIMIT 1;
$$;

REVOKE ALL ON FUNCTION pvnaive.resolve_direct_subscription_account_profile(bytea) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION pvnaive.resolve_direct_subscription_account_profile(bytea) TO pvnaive_app;
COMMENT ON FUNCTION pvnaive.resolve_direct_subscription_account_profile(bytea) IS
    'Resolves token-bound service identity, commercial metadata, and immutable accounting-baseline provenance for the public account page without bypassing management RLS.';
