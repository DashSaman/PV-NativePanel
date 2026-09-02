-- pvnaive:migration-version 0021
-- pvnaive:migration-name ip_session_history
-- pvnaive:transactional true
-- pvnaive:destructive false

SET LOCAL lock_timeout = '10s';
SET LOCAL statement_timeout = '120s';
SET LOCAL idle_in_transaction_session_timeout = '60s';
SET LOCAL ROLE pvnaive_owner;

CREATE TABLE pvnaive.direct_naive_session_history (
    runtime_credential_id uuid NOT NULL REFERENCES pvnaive.naive_runtime_credentials(id) ON DELETE RESTRICT,
    node_id text NOT NULL CHECK (length(node_id) BETWEEN 1 AND 160),
    boot_id uuid NOT NULL,
    session_id uuid NOT NULL,
    service_term_id uuid NOT NULL REFERENCES pvnaive.service_terms(id) ON DELETE RESTRICT,
    tenant_id uuid NOT NULL REFERENCES pvnaive.tenants(id) ON DELETE RESTRICT,
    user_id uuid NOT NULL,
    client_ip inet NOT NULL,
    connected_at timestamptz NOT NULL,
    final_at timestamptz NOT NULL,
    upload_bytes bigint NOT NULL CHECK (upload_bytes >= 0),
    download_bytes bigint NOT NULL CHECK (download_bytes >= 0),
    recorded_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    PRIMARY KEY (runtime_credential_id, node_id, boot_id, session_id),
    FOREIGN KEY (service_term_id, tenant_id, user_id)
        REFERENCES pvnaive.service_terms(id, tenant_id, user_id)
        ON DELETE RESTRICT,
    CHECK (family(client_ip) IN (4, 6)),
    CHECK (final_at >= connected_at),
    CHECK (upload_bytes <= 9223372036854775807 - download_bytes)
);

CREATE INDEX direct_naive_session_history_customer_final_idx
    ON pvnaive.direct_naive_session_history (tenant_id, user_id, final_at DESC, session_id);

ALTER TABLE pvnaive.direct_naive_session_history ENABLE ROW LEVEL SECURITY;
ALTER TABLE pvnaive.direct_naive_session_history FORCE ROW LEVEL SECURITY;

CREATE POLICY direct_naive_session_history_tenant_select
    ON pvnaive.direct_naive_session_history
    FOR SELECT TO pvnaive_app
    USING (pvnaive.has_tenant_access(tenant_id, false));

CREATE POLICY direct_naive_session_history_owner_all
    ON pvnaive.direct_naive_session_history
    FOR ALL TO pvnaive_owner
    USING (pvnaive.has_tenant_access(tenant_id, false))
    WITH CHECK (pvnaive.has_tenant_access(tenant_id, false));

REVOKE ALL ON pvnaive.direct_naive_session_history FROM PUBLIC, pvnaive_app;

CREATE FUNCTION pvnaive.list_customer_session_history(
    p_user_id uuid,
    p_before timestamptz,
    p_limit integer
)
RETURNS TABLE(
    runtime_credential_id uuid,
    node_id text,
    boot_id uuid,
    session_id uuid,
    service_term_id uuid,
    client_ip text,
    connected_at timestamptz,
    final_at timestamptz,
    duration_seconds bigint,
    upload_bytes bigint,
    download_bytes bigint
)
LANGUAGE plpgsql
STABLE
SECURITY DEFINER
SET search_path = pg_catalog, pvnaive
AS $$
BEGIN
    IF p_user_id IS NULL OR p_before IS NULL OR p_limit < 1 OR p_limit > 500 THEN
        RAISE EXCEPTION 'invalid session history query' USING ERRCODE = '22023';
    END IF;
    IF NOT pvnaive.has_valid_context() THEN
        RAISE EXCEPTION 'valid request context required' USING ERRCODE = '28000';
    END IF;

    RETURN QUERY
    SELECT h.runtime_credential_id, h.node_id, h.boot_id, h.session_id,
           h.service_term_id, host(h.client_ip), h.connected_at, h.final_at,
           GREATEST(0, floor(extract(epoch FROM (h.final_at - h.connected_at)))::bigint),
           h.upload_bytes, h.download_bytes
      FROM pvnaive.direct_naive_session_history AS h
     WHERE h.user_id = p_user_id
       AND h.final_at <= p_before
       AND h.final_at >= p_before - interval '30 days'
       AND pvnaive.has_tenant_access(h.tenant_id, false)
     ORDER BY h.final_at DESC, h.session_id DESC
     LIMIT p_limit;
END;
$$;

CREATE FUNCTION pvnaive.sync_direct_naive_session_history(
    p_observed_at timestamptz
)
RETURNS TABLE(inserted_count bigint, purged_count bigint)
LANGUAGE plpgsql
VOLATILE
SECURITY DEFINER
SET search_path = pg_catalog, pvnaive
AS $$
DECLARE
    v_inserted bigint := 0;
    v_purged bigint := 0;
BEGIN
    IF p_observed_at IS NULL THEN
        RAISE EXCEPTION 'invalid history maintenance timestamp' USING ERRCODE = '22023';
    END IF;

    PERFORM pvnaive.direct_naive_accounting_enter_context();

    INSERT INTO pvnaive.direct_naive_session_history(
        runtime_credential_id, node_id, boot_id, session_id, service_term_id,
        tenant_id, user_id, client_ip, connected_at, final_at,
        upload_bytes, download_bytes
    )
    SELECT s.runtime_credential_id, s.node_id, s.boot_id, s.session_id,
           s.service_term_id, p.tenant_id, p.user_id, p.client_ip,
           s.first_observed_at, s.last_observed_at,
           s.upload_cumulative, s.download_cumulative
      FROM pvnaive.direct_naive_accounting_sessions AS s
      JOIN pvnaive.direct_naive_accounting_session_peers AS p
        ON p.runtime_credential_id = s.runtime_credential_id
       AND p.node_id = s.node_id
       AND p.boot_id = s.boot_id
       AND p.session_id = s.session_id
       AND p.service_term_id = s.service_term_id
      JOIN pvnaive.service_terms AS st
        ON st.id = s.service_term_id
       AND st.tenant_id = p.tenant_id
       AND st.user_id = p.user_id
     WHERE s.final = true
       AND s.accounting_complete = true
       AND s.last_observed_at <= p_observed_at
       AND s.last_observed_at >= p_observed_at - interval '30 days'
    ON CONFLICT (runtime_credential_id, node_id, boot_id, session_id) DO NOTHING;

    GET DIAGNOSTICS v_inserted = ROW_COUNT;

    DELETE FROM pvnaive.direct_naive_session_history AS h
     WHERE h.final_at < p_observed_at - interval '30 days';
    GET DIAGNOSTICS v_purged = ROW_COUNT;

    PERFORM pvnaive.direct_naive_accounting_leave_context();
    RETURN QUERY SELECT v_inserted, v_purged;
END;
$$;

REVOKE ALL ON FUNCTION pvnaive.list_customer_session_history(uuid,timestamptz,integer) FROM PUBLIC;
REVOKE ALL ON FUNCTION pvnaive.sync_direct_naive_session_history(timestamptz) FROM PUBLIC, pvnaive_app;
GRANT EXECUTE ON FUNCTION pvnaive.list_customer_session_history(uuid,timestamptz,integer) TO pvnaive_app;
GRANT EXECUTE ON FUNCTION pvnaive.sync_direct_naive_session_history(timestamptz) TO pvnaive_owner;

COMMENT ON TABLE pvnaive.direct_naive_session_history IS
    'Bounded 30-day customer session/IP history materialized only from finalized exact accounting sessions joined to trusted Caddy RemoteAddr peer evidence.';
COMMENT ON FUNCTION pvnaive.list_customer_session_history(uuid,timestamptz,integer) IS
    'Tenant-scoped bounded session history read. p_limit is mandatory and hard-capped at 500.';
COMMENT ON FUNCTION pvnaive.sync_direct_naive_session_history(timestamptz) IS
    'Maintenance-only materialization and purge boundary; keeps exactly the trailing 30 days and accepts only final accounting-complete sessions with trusted exact peer lineage.';
