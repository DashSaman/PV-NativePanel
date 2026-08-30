-- pvnaive:migration-version 0014
-- Source: PVNaive Task #7 manual exact-accounting usage reset
-- pvnaive:migration-name manual_usage_reset
-- pvnaive:transactional true
-- pvnaive:destructive false

SET LOCAL lock_timeout = '10s';
SET LOCAL statement_timeout = '120s';
SET LOCAL idle_in_transaction_session_timeout = '60s';
SET LOCAL ROLE pvnaive_owner;

ALTER TABLE pvnaive.direct_naive_accounting_terms
    ADD COLUMN last_reset_at timestamptz;

CREATE TABLE pvnaive.direct_naive_accounting_reset_events (
    id uuid PRIMARY KEY DEFAULT public.gen_random_uuid(),
    tenant_id uuid NOT NULL REFERENCES pvnaive.tenants(id) ON DELETE RESTRICT,
    user_id uuid NOT NULL,
    service_term_id uuid NOT NULL,
    actor_id uuid NOT NULL REFERENCES pvnaive.actors(id) ON DELETE RESTRICT,
    customer_mutation_key_id uuid NOT NULL UNIQUE
        REFERENCES pvnaive.customer_mutation_keys(id) ON DELETE RESTRICT,
    reason text NOT NULL CHECK (reason IN ('manual', 'bulk', 'scheduled')),
    reset_at timestamptz NOT NULL,
    previous_upload_bytes bigint NOT NULL CHECK (previous_upload_bytes >= 0),
    previous_download_bytes bigint NOT NULL CHECK (previous_download_bytes >= 0),
    previous_used_bytes bigint NOT NULL CHECK (previous_used_bytes >= 0),
    created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    UNIQUE (id, tenant_id),
    FOREIGN KEY (user_id, tenant_id)
        REFERENCES pvnaive.users(id, tenant_id) ON DELETE RESTRICT,
    FOREIGN KEY (service_term_id, tenant_id, user_id)
        REFERENCES pvnaive.service_terms(id, tenant_id, user_id) ON DELETE RESTRICT,
    CHECK (previous_upload_bytes <= 9223372036854775807 - previous_download_bytes),
    CHECK (previous_used_bytes = previous_upload_bytes + previous_download_bytes),
    CHECK (created_at >= reset_at - interval '5 minutes')
);
CREATE INDEX direct_naive_accounting_reset_events_term_time_idx
    ON pvnaive.direct_naive_accounting_reset_events(service_term_id, reset_at DESC);

ALTER TABLE pvnaive.direct_naive_accounting_reset_events ENABLE ROW LEVEL SECURITY;
ALTER TABLE pvnaive.direct_naive_accounting_reset_events FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON pvnaive.direct_naive_accounting_reset_events
    USING (pvnaive.has_tenant_access(tenant_id, false))
    WITH CHECK (pvnaive.has_tenant_access(tenant_id, false));

CREATE TRIGGER direct_naive_accounting_reset_events_immutable
BEFORE UPDATE OR DELETE ON pvnaive.direct_naive_accounting_reset_events
FOR EACH ROW EXECUTE FUNCTION pvnaive.prevent_immutable_mutation();

REVOKE ALL ON pvnaive.direct_naive_accounting_reset_events FROM PUBLIC, pvnaive_app;
GRANT SELECT, INSERT ON pvnaive.direct_naive_accounting_reset_events TO pvnaive_app;

CREATE FUNCTION pvnaive.direct_naive_accounting_reset(
    p_service_term_id uuid,
    p_reset_at timestamptz,
    p_stale_after_seconds bigint DEFAULT 90
)
RETURNS TABLE (
    service_term_id uuid,
    tenant_id uuid,
    user_id uuid,
    resettable boolean,
    reason text,
    previous_upload_bytes bigint,
    previous_download_bytes bigint,
    previous_used_bytes bigint,
    reset_at timestamptz,
    service_state text
)
LANGUAGE plpgsql
VOLATILE
SECURITY DEFINER
SET search_path = pg_catalog, pvnaive
AS $$
DECLARE
    v_term pvnaive.service_terms%ROWTYPE;
    v_projection pvnaive.direct_naive_accounting_terms%ROWTYPE;
    v_previous_used bigint;
BEGIN
    IF p_service_term_id IS NULL OR p_reset_at IS NULL OR p_stale_after_seconds <= 0 THEN
        RAISE EXCEPTION 'invalid accounting reset request' USING ERRCODE = '22023';
    END IF;
    IF NOT pvnaive.has_valid_context() OR pvnaive.current_actor_role() <> 'owner' THEN
        RAISE EXCEPTION 'accounting reset requires owner context' USING ERRCODE = '42501';
    END IF;

    SELECT st.* INTO v_term
      FROM pvnaive.service_terms AS st
     WHERE st.id = p_service_term_id
     FOR UPDATE;
    IF NOT FOUND THEN
        RETURN QUERY SELECT p_service_term_id, NULL::uuid, NULL::uuid, false, 'not_found'::text,
                            0::bigint, 0::bigint, 0::bigint, p_reset_at, NULL::text;
        RETURN;
    END IF;
    IF v_term.state IN ('ended', 'revoked') THEN
        RETURN QUERY SELECT v_term.id, v_term.tenant_id, v_term.user_id, false, 'service_inactive'::text,
                            0::bigint, 0::bigint, 0::bigint, p_reset_at, v_term.state;
        RETURN;
    END IF;

    INSERT INTO pvnaive.direct_naive_accounting_terms(service_term_id)
    VALUES(v_term.id)
    ON CONFLICT ON CONSTRAINT direct_naive_accounting_terms_pkey DO NOTHING;

    SELECT t.* INTO v_projection
      FROM pvnaive.direct_naive_accounting_terms AS t
     WHERE t.service_term_id = v_term.id
     FOR UPDATE;

    v_previous_used := v_projection.upload_bytes + v_projection.download_bytes;

    IF NOT v_projection.accounting_complete
       OR EXISTS (
           SELECT 1 FROM pvnaive.direct_naive_accounting_sessions AS s
            WHERE s.service_term_id = v_term.id AND NOT s.accounting_complete
       ) THEN
        RETURN QUERY SELECT v_term.id, v_term.tenant_id, v_term.user_id, false, 'accounting_incomplete'::text,
                            v_projection.upload_bytes, v_projection.download_bytes, v_previous_used,
                            p_reset_at, v_term.state;
        RETURN;
    END IF;

    IF v_projection.reserved_bytes <> 0
       OR EXISTS (
           SELECT 1 FROM pvnaive.direct_naive_accounting_claims AS c
            WHERE c.service_term_id = v_term.id AND c.settled_at IS NULL
       ) THEN
        RETURN QUERY SELECT v_term.id, v_term.tenant_id, v_term.user_id, false, 'reservation_pending'::text,
                            v_projection.upload_bytes, v_projection.download_bytes, v_previous_used,
                            p_reset_at, v_term.state;
        RETURN;
    END IF;

    IF EXISTS (
        SELECT 1 FROM pvnaive.direct_naive_accounting_sessions AS s
         WHERE s.service_term_id = v_term.id
           AND NOT s.final
           AND s.last_observed_at < p_reset_at - (p_stale_after_seconds * interval '1 second')
    ) THEN
        RETURN QUERY SELECT v_term.id, v_term.tenant_id, v_term.user_id, false, 'telemetry_stale'::text,
                            v_projection.upload_bytes, v_projection.download_bytes, v_previous_used,
                            p_reset_at, v_term.state;
        RETURN;
    END IF;

    IF v_projection.last_reset_at IS NOT NULL AND p_reset_at <= v_projection.last_reset_at THEN
        RETURN QUERY SELECT v_term.id, v_term.tenant_id, v_term.user_id, false, 'reset_time_conflict'::text,
                            v_projection.upload_bytes, v_projection.download_bytes, v_previous_used,
                            p_reset_at, v_term.state;
        RETURN;
    END IF;

    UPDATE pvnaive.direct_naive_accounting_terms
       SET upload_bytes = 0,
           download_bytes = 0,
           last_reset_at = p_reset_at,
           updated_at = p_reset_at
     WHERE direct_naive_accounting_terms.service_term_id = v_term.id;

    IF v_term.state = 'quota_depleted'
       AND (v_term.expires_at IS NULL OR v_term.expires_at > p_reset_at) THEN
        UPDATE pvnaive.service_terms
           SET state = 'active', revision = revision + 1, updated_at = p_reset_at
         WHERE id = v_term.id
         RETURNING * INTO v_term;
    END IF;

    RETURN QUERY SELECT v_term.id, v_term.tenant_id, v_term.user_id, true, 'reset'::text,
                        v_projection.upload_bytes, v_projection.download_bytes, v_previous_used,
                        p_reset_at, v_term.state;
END;
$$;
REVOKE ALL ON FUNCTION pvnaive.direct_naive_accounting_reset(uuid,timestamptz,bigint) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION pvnaive.direct_naive_accounting_reset(uuid,timestamptz,bigint) TO pvnaive_app;

-- Extend the exact read projection with its explicit reset epoch. Keep the v13
-- reader intact so rollback can restore the exact previous interface.
ALTER FUNCTION pvnaive.direct_naive_accounting_read(uuid,timestamptz,bigint)
    RENAME TO direct_naive_accounting_read_v13;
REVOKE ALL ON FUNCTION pvnaive.direct_naive_accounting_read_v13(uuid,timestamptz,bigint)
    FROM PUBLIC, pvnaive_app;

CREATE FUNCTION pvnaive.direct_naive_accounting_read(
    p_service_term_id uuid,
    p_observed_at timestamptz,
    p_stale_after_seconds bigint DEFAULT 90
)
RETURNS TABLE (
    service_term_id uuid,
    upload_bytes bigint,
    download_bytes bigint,
    used_bytes bigint,
    quota_bytes bigint,
    remaining_bytes bigint,
    quota_state text,
    first_connected_at timestamptz,
    last_online timestamptz,
    online boolean,
    session_count bigint,
    accounting_complete boolean,
    last_reset_at timestamptz
)
LANGUAGE sql
STABLE
SECURITY DEFINER
SET search_path = pg_catalog, pvnaive
AS $$
    SELECT r.service_term_id,
           r.upload_bytes,
           r.download_bytes,
           r.used_bytes,
           r.quota_bytes,
           r.remaining_bytes,
           r.quota_state,
           r.first_connected_at,
           r.last_online,
           r.online,
           r.session_count,
           r.accounting_complete,
           t.last_reset_at
      FROM pvnaive.direct_naive_accounting_read_v13(
               p_service_term_id,
               p_observed_at,
               p_stale_after_seconds
           ) AS r
      LEFT JOIN pvnaive.direct_naive_accounting_terms AS t
        ON t.service_term_id = r.service_term_id;
$$;
REVOKE ALL ON FUNCTION pvnaive.direct_naive_accounting_read(uuid,timestamptz,bigint) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION pvnaive.direct_naive_accounting_read(uuid,timestamptz,bigint) TO pvnaive_app;

COMMENT ON TABLE pvnaive.direct_naive_accounting_reset_events IS
    'Append-only management history for exact-accounting reset epochs. Reset never deletes immutable CONNECT ledger events or rotates credentials/tokens.';
COMMENT ON COLUMN pvnaive.direct_naive_accounting_terms.last_reset_at IS
    'Start of the current exact usage period after an explicit reset; NULL means the ServiceTerm adoption baseline still governs lifetime composition.';
