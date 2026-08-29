-- pvnaive:migration-version 0007
-- Source: PVNaive exact per-runtime accounting foundation
-- pvnaive:migration-name exact_accounting
-- pvnaive:transactional true
-- pvnaive:destructive false

SET LOCAL lock_timeout = '10s';
SET LOCAL statement_timeout = '120s';
SET LOCAL idle_in_transaction_session_timeout = '60s';
SET LOCAL ROLE pvnaive_owner;

ALTER TABLE pvnaive.direct_subscription_tokens
    ADD COLUMN token_ciphertext bytea,
    ADD COLUMN token_nonce bytea,
    ADD COLUMN token_encryption_key_id text,
    ADD CONSTRAINT direct_subscription_token_recovery_material_ck CHECK (
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

CREATE TABLE pvnaive.usage_counters (
    service_term_id uuid PRIMARY KEY,
    tenant_id uuid NOT NULL,
    user_id uuid NOT NULL,
    runtime_credential_id uuid NOT NULL
        REFERENCES pvnaive.naive_runtime_credentials(id) ON DELETE RESTRICT,
    user_state text NOT NULL
        CHECK (user_state IN ('draft', 'active', 'suspended', 'expired', 'depleted', 'revoked')),
    service_state text NOT NULL
        CHECK (service_state IN ('pending', 'active', 'expired', 'quota_depleted', 'ended', 'revoked')),
    runtime_state text NOT NULL
        CHECK (runtime_state IN ('active', 'disabled', 'revoked')),
    quota_bytes bigint CHECK (quota_bytes IS NULL OR quota_bytes > 0),
    expires_at timestamptz,
    upload_bytes bigint NOT NULL DEFAULT 0 CHECK (upload_bytes >= 0),
    download_bytes bigint NOT NULL DEFAULT 0 CHECK (download_bytes >= 0),
    active_binding boolean NOT NULL DEFAULT true,
    updated_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    FOREIGN KEY (service_term_id, tenant_id, user_id)
        REFERENCES pvnaive.service_terms(id, tenant_id, user_id) ON DELETE RESTRICT,
    CHECK ((upload_bytes::numeric + download_bytes::numeric) <= 9223372036854775807::numeric)
);

CREATE UNIQUE INDEX usage_counters_runtime_active_uidx
    ON pvnaive.usage_counters (runtime_credential_id)
    WHERE active_binding;
CREATE INDEX usage_counters_tenant_user_idx
    ON pvnaive.usage_counters (tenant_id, user_id, active_binding);

CREATE TABLE pvnaive.usage_connection_sequences (
    runtime_credential_id uuid NOT NULL
        REFERENCES pvnaive.naive_runtime_credentials(id) ON DELETE RESTRICT,
    connection_id uuid NOT NULL,
    last_sequence bigint NOT NULL CHECK (last_sequence > 0),
    last_upload_delta bigint NOT NULL CHECK (last_upload_delta >= 0),
    last_download_delta bigint NOT NULL CHECK (last_download_delta >= 0),
    updated_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    PRIMARY KEY (runtime_credential_id, connection_id)
);

ALTER TABLE pvnaive.usage_counters ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON pvnaive.usage_counters
    USING (pvnaive.has_tenant_access(tenant_id, false))
    WITH CHECK (pvnaive.has_tenant_access(tenant_id, false));

REVOKE ALL ON pvnaive.usage_counters FROM PUBLIC;
REVOKE ALL ON pvnaive.usage_connection_sequences FROM PUBLIC;
GRANT SELECT ON pvnaive.usage_counters TO pvnaive_app;

CREATE FUNCTION pvnaive.sync_usage_counter_binding()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog, pvnaive
AS $$
DECLARE
    v_user_state text;
    v_service_state text;
    v_runtime_state text;
    v_quota_bytes bigint;
    v_expires_at timestamptz;
BEGIN
    IF TG_OP = 'UPDATE'
       AND OLD.unbound_at IS NULL
       AND NEW.unbound_at IS NOT NULL THEN
        UPDATE pvnaive.usage_counters AS uc
           SET active_binding = false,
               updated_at = clock_timestamp()
         WHERE uc.service_term_id = NEW.service_term_id
           AND uc.runtime_credential_id = NEW.runtime_credential_id
           AND uc.active_binding;
        RETURN NEW;
    END IF;

    IF NEW.unbound_at IS NOT NULL THEN
        RETURN NEW;
    END IF;

    SELECT u.status, st.state, rc.status, st.quota_bytes, st.expires_at
      INTO STRICT v_user_state, v_service_state, v_runtime_state, v_quota_bytes, v_expires_at
      FROM pvnaive.users AS u
      JOIN pvnaive.service_terms AS st
        ON st.id = NEW.service_term_id
       AND st.tenant_id = NEW.tenant_id
       AND st.user_id = NEW.user_id
      JOIN pvnaive.naive_runtime_credentials AS rc
        ON rc.id = NEW.runtime_credential_id
     WHERE u.id = NEW.user_id
       AND u.tenant_id = NEW.tenant_id;

    INSERT INTO pvnaive.usage_counters (
        service_term_id, tenant_id, user_id, runtime_credential_id,
        user_state, service_state, runtime_state, quota_bytes, expires_at,
        upload_bytes, download_bytes, active_binding, updated_at
    ) VALUES (
        NEW.service_term_id, NEW.tenant_id, NEW.user_id, NEW.runtime_credential_id,
        v_user_state, v_service_state, v_runtime_state, v_quota_bytes, v_expires_at,
        0, 0, true, clock_timestamp()
    )
    ON CONFLICT (service_term_id) DO UPDATE
       SET tenant_id = EXCLUDED.tenant_id,
           user_id = EXCLUDED.user_id,
           runtime_credential_id = EXCLUDED.runtime_credential_id,
           user_state = EXCLUDED.user_state,
           service_state = EXCLUDED.service_state,
           runtime_state = EXCLUDED.runtime_state,
           quota_bytes = EXCLUDED.quota_bytes,
           expires_at = EXCLUDED.expires_at,
           active_binding = true,
           updated_at = clock_timestamp();

    RETURN NEW;
END;
$$;

CREATE FUNCTION pvnaive.sync_usage_counter_user()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog, pvnaive
AS $$
BEGIN
    UPDATE pvnaive.usage_counters AS uc
       SET user_state = NEW.status,
           updated_at = clock_timestamp()
     WHERE uc.user_id = NEW.id
       AND uc.tenant_id = NEW.tenant_id
       AND uc.active_binding;
    RETURN NEW;
END;
$$;

CREATE FUNCTION pvnaive.sync_usage_counter_service_term()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog, pvnaive
AS $$
BEGIN
    UPDATE pvnaive.usage_counters AS uc
       SET service_state = NEW.state,
           quota_bytes = NEW.quota_bytes,
           expires_at = NEW.expires_at,
           updated_at = clock_timestamp()
     WHERE uc.service_term_id = NEW.id
       AND uc.tenant_id = NEW.tenant_id
       AND uc.user_id = NEW.user_id
       AND uc.active_binding;
    RETURN NEW;
END;
$$;

CREATE FUNCTION pvnaive.sync_usage_counter_runtime()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog, pvnaive
AS $$
BEGIN
    UPDATE pvnaive.usage_counters AS uc
       SET runtime_state = NEW.status,
           updated_at = clock_timestamp()
     WHERE uc.runtime_credential_id = NEW.id
       AND uc.active_binding;
    RETURN NEW;
END;
$$;

CREATE TRIGGER usage_counter_binding_sync
AFTER INSERT OR UPDATE OF unbound_at ON pvnaive.user_runtime_credentials
FOR EACH ROW
EXECUTE FUNCTION pvnaive.sync_usage_counter_binding();

CREATE TRIGGER usage_counter_user_sync
AFTER UPDATE OF status ON pvnaive.users
FOR EACH ROW
EXECUTE FUNCTION pvnaive.sync_usage_counter_user();

CREATE TRIGGER usage_counter_service_term_sync
AFTER UPDATE OF state, quota_bytes, expires_at ON pvnaive.service_terms
FOR EACH ROW
EXECUTE FUNCTION pvnaive.sync_usage_counter_service_term();

CREATE TRIGGER usage_counter_runtime_sync
AFTER UPDATE OF status ON pvnaive.naive_runtime_credentials
FOR EACH ROW
EXECUTE FUNCTION pvnaive.sync_usage_counter_runtime();

-- FORCE RLS source tables are temporarily unforced while the migration owns
-- their exclusive ALTER locks, solely to seed exact zero counters for existing
-- active managed bindings. FORCE is restored before transaction commit.
ALTER TABLE pvnaive.users NO FORCE ROW LEVEL SECURITY;
ALTER TABLE pvnaive.service_terms NO FORCE ROW LEVEL SECURITY;
ALTER TABLE pvnaive.user_runtime_credentials NO FORCE ROW LEVEL SECURITY;
ALTER TABLE pvnaive.naive_runtime_credentials NO FORCE ROW LEVEL SECURITY;

INSERT INTO pvnaive.usage_counters (
    service_term_id, tenant_id, user_id, runtime_credential_id,
    user_state, service_state, runtime_state, quota_bytes, expires_at,
    upload_bytes, download_bytes, active_binding, updated_at
)
SELECT
    urc.service_term_id,
    urc.tenant_id,
    urc.user_id,
    urc.runtime_credential_id,
    u.status,
    st.state,
    rc.status,
    st.quota_bytes,
    st.expires_at,
    0,
    0,
    true,
    clock_timestamp()
FROM pvnaive.user_runtime_credentials AS urc
JOIN pvnaive.users AS u
  ON u.id = urc.user_id
 AND u.tenant_id = urc.tenant_id
JOIN pvnaive.service_terms AS st
  ON st.id = urc.service_term_id
 AND st.tenant_id = urc.tenant_id
 AND st.user_id = urc.user_id
JOIN pvnaive.naive_runtime_credentials AS rc
  ON rc.id = urc.runtime_credential_id
WHERE urc.unbound_at IS NULL
ON CONFLICT (service_term_id) DO NOTHING;

ALTER TABLE pvnaive.users FORCE ROW LEVEL SECURITY;
ALTER TABLE pvnaive.service_terms FORCE ROW LEVEL SECURITY;
ALTER TABLE pvnaive.user_runtime_credentials FORCE ROW LEVEL SECURITY;
ALTER TABLE pvnaive.naive_runtime_credentials FORCE ROW LEVEL SECURITY;

CREATE FUNCTION pvnaive.accounting_authorize(p_runtime_credential_id uuid)
RETURNS TABLE (
    tracked boolean,
    allowed boolean,
    reason text,
    quota_bytes bigint,
    upload_bytes bigint,
    download_bytes bigint,
    used_bytes bigint,
    remaining_bytes bigint,
    expires_at timestamptz,
    accounting_updated_at timestamptz
)
LANGUAGE plpgsql
STABLE
SECURITY DEFINER
SET search_path = pg_catalog, pvnaive
AS $$
DECLARE
    v_counter pvnaive.usage_counters%ROWTYPE;
    v_used numeric;
    v_allowed boolean := true;
    v_reason text := 'allowed';
    v_remaining bigint;
BEGIN
    IF p_runtime_credential_id IS NULL THEN
        RAISE EXCEPTION 'runtime credential id is required' USING ERRCODE = '22023';
    END IF;

    SELECT uc.*
      INTO v_counter
      FROM pvnaive.usage_counters AS uc
     WHERE uc.runtime_credential_id = p_runtime_credential_id
       AND uc.active_binding
     LIMIT 1;

    IF NOT FOUND THEN
        RETURN QUERY
        SELECT false, true, 'unmanaged'::text,
               NULL::bigint, 0::bigint, 0::bigint, 0::bigint, NULL::bigint,
               NULL::timestamptz, NULL::timestamptz;
        RETURN;
    END IF;

    v_used := v_counter.upload_bytes::numeric + v_counter.download_bytes::numeric;

    IF v_counter.runtime_state <> 'active' THEN
        v_allowed := false;
        v_reason := 'runtime_' || v_counter.runtime_state;
    ELSIF v_counter.user_state <> 'active' THEN
        v_allowed := false;
        v_reason := 'user_' || v_counter.user_state;
    ELSIF v_counter.service_state = 'expired' THEN
        v_allowed := false;
        v_reason := 'expired';
    ELSIF v_counter.service_state = 'quota_depleted' THEN
        v_allowed := false;
        v_reason := 'quota_depleted';
    ELSIF v_counter.service_state NOT IN ('pending', 'active') THEN
        v_allowed := false;
        v_reason := 'service_' || v_counter.service_state;
    ELSIF v_counter.expires_at IS NOT NULL
       AND v_counter.expires_at <= clock_timestamp() THEN
        v_allowed := false;
        v_reason := 'expired';
    ELSIF v_counter.quota_bytes IS NOT NULL
       AND v_used >= v_counter.quota_bytes::numeric THEN
        v_allowed := false;
        v_reason := 'quota_depleted';
    END IF;

    IF v_counter.quota_bytes IS NULL THEN
        v_remaining := NULL;
    ELSE
        v_remaining := GREATEST(v_counter.quota_bytes::numeric - v_used, 0)::bigint;
    END IF;

    RETURN QUERY
    SELECT true,
           v_allowed,
           v_reason,
           v_counter.quota_bytes,
           v_counter.upload_bytes,
           v_counter.download_bytes,
           v_used::bigint,
           v_remaining,
           v_counter.expires_at,
           v_counter.updated_at;
END;
$$;

CREATE FUNCTION pvnaive.accounting_apply_delta(
    p_runtime_credential_id uuid,
    p_connection_id uuid,
    p_sequence bigint,
    p_upload_delta bigint,
    p_download_delta bigint
)
RETURNS TABLE (
    tracked boolean,
    accepted boolean,
    idempotent boolean,
    continue_allowed boolean,
    reason text,
    upload_bytes bigint,
    download_bytes bigint,
    used_bytes bigint,
    remaining_bytes bigint
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog, pvnaive
AS $$
DECLARE
    v_counter pvnaive.usage_counters%ROWTYPE;
    v_sequence pvnaive.usage_connection_sequences%ROWTYPE;
    v_has_sequence boolean := false;
    v_new_upload numeric;
    v_new_download numeric;
    v_new_used numeric;
    v_allowed boolean;
    v_reason text;
    v_used bigint;
    v_remaining bigint;
BEGIN
    IF p_runtime_credential_id IS NULL
       OR p_connection_id IS NULL
       OR p_sequence IS NULL
       OR p_sequence < 1
       OR p_upload_delta IS NULL
       OR p_download_delta IS NULL
       OR p_upload_delta < 0
       OR p_download_delta < 0 THEN
        RAISE EXCEPTION 'invalid accounting delta input' USING ERRCODE = '22023';
    END IF;

    SELECT uc.*
      INTO v_counter
      FROM pvnaive.usage_counters AS uc
     WHERE uc.runtime_credential_id = p_runtime_credential_id
       AND uc.active_binding
     FOR UPDATE;

    IF NOT FOUND THEN
        RETURN QUERY
        SELECT false, false, false, true, 'unmanaged'::text,
               0::bigint, 0::bigint, 0::bigint, NULL::bigint;
        RETURN;
    END IF;

    SELECT ucs.*
      INTO v_sequence
      FROM pvnaive.usage_connection_sequences AS ucs
     WHERE ucs.runtime_credential_id = p_runtime_credential_id
       AND ucs.connection_id = p_connection_id
     FOR UPDATE;
    v_has_sequence := FOUND;

    IF v_has_sequence AND p_sequence = v_sequence.last_sequence THEN
        IF p_upload_delta <> v_sequence.last_upload_delta
           OR p_download_delta <> v_sequence.last_download_delta THEN
            RAISE EXCEPTION 'accounting sequence replay payload mismatch'
                USING ERRCODE = '22023';
        END IF;

        SELECT a.allowed, a.reason, a.used_bytes, a.remaining_bytes
          INTO v_allowed, v_reason, v_used, v_remaining
          FROM pvnaive.accounting_authorize(p_runtime_credential_id) AS a;

        RETURN QUERY
        SELECT true, true, true, v_allowed, v_reason,
               v_counter.upload_bytes, v_counter.download_bytes,
               v_used, v_remaining;
        RETURN;
    END IF;

    IF (v_has_sequence AND p_sequence <> v_sequence.last_sequence + 1)
       OR (NOT v_has_sequence AND p_sequence <> 1) THEN
        RAISE EXCEPTION 'accounting sequence gap' USING ERRCODE = '22023';
    END IF;

    v_new_upload := v_counter.upload_bytes::numeric + p_upload_delta::numeric;
    v_new_download := v_counter.download_bytes::numeric + p_download_delta::numeric;
    v_new_used := v_new_upload + v_new_download;

    IF v_new_upload > 9223372036854775807::numeric
       OR v_new_download > 9223372036854775807::numeric
       OR v_new_used > 9223372036854775807::numeric THEN
        RAISE EXCEPTION 'accounting counter overflow' USING ERRCODE = '22003';
    END IF;

    UPDATE pvnaive.usage_counters AS uc
       SET upload_bytes = v_new_upload::bigint,
           download_bytes = v_new_download::bigint,
           updated_at = clock_timestamp()
     WHERE uc.service_term_id = v_counter.service_term_id;

    INSERT INTO pvnaive.usage_connection_sequences (
        runtime_credential_id, connection_id, last_sequence,
        last_upload_delta, last_download_delta, updated_at
    ) VALUES (
        p_runtime_credential_id, p_connection_id, p_sequence,
        p_upload_delta, p_download_delta, clock_timestamp()
    )
    ON CONFLICT (runtime_credential_id, connection_id) DO UPDATE
       SET last_sequence = EXCLUDED.last_sequence,
           last_upload_delta = EXCLUDED.last_upload_delta,
           last_download_delta = EXCLUDED.last_download_delta,
           updated_at = clock_timestamp();

    SELECT a.allowed, a.reason, a.used_bytes, a.remaining_bytes
      INTO v_allowed, v_reason, v_used, v_remaining
      FROM pvnaive.accounting_authorize(p_runtime_credential_id) AS a;

    RETURN QUERY
    SELECT true, true, false, v_allowed, v_reason,
           v_new_upload::bigint, v_new_download::bigint,
           v_used, v_remaining;
END;
$$;

REVOKE ALL ON FUNCTION pvnaive.sync_usage_counter_binding() FROM PUBLIC;
REVOKE ALL ON FUNCTION pvnaive.sync_usage_counter_user() FROM PUBLIC;
REVOKE ALL ON FUNCTION pvnaive.sync_usage_counter_service_term() FROM PUBLIC;
REVOKE ALL ON FUNCTION pvnaive.sync_usage_counter_runtime() FROM PUBLIC;
REVOKE ALL ON FUNCTION pvnaive.accounting_authorize(uuid) FROM PUBLIC;
REVOKE ALL ON FUNCTION pvnaive.accounting_apply_delta(uuid, uuid, bigint, bigint, bigint) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION pvnaive.accounting_authorize(uuid) TO pvnaive_app;
GRANT EXECUTE ON FUNCTION pvnaive.accounting_apply_delta(uuid, uuid, bigint, bigint, bigint) TO pvnaive_app;

COMMENT ON TABLE pvnaive.usage_counters IS
    'Exact zero-origin upload/download counters for managed Runtime credential bindings; no historical usage is fabricated.';
COMMENT ON TABLE pvnaive.usage_connection_sequences IS
    'Per-connection monotonic sequence ledger used to reject gaps and make identical delta retries idempotent.';
COMMENT ON FUNCTION pvnaive.accounting_authorize(uuid) IS
    'Returns trusted current accounting policy for a stable Runtime credential UUID; unmanaged credentials remain compatible.';
COMMENT ON FUNCTION pvnaive.accounting_apply_delta(uuid, uuid, bigint, bigint, bigint) IS
    'Atomically applies exact successful-write deltas with replay/gap/overflow protection and returns post-delta policy.';
