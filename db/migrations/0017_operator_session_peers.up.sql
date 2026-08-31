-- pvnaive:migration-version 0017
-- pvnaive:transactional true
-- pvnaive:destructive false

SET LOCAL lock_timeout = '10s';
SET LOCAL statement_timeout = '120s';
SET LOCAL idle_in_transaction_session_timeout = '60s';
SET LOCAL ROLE pvnaive_owner;

CREATE TABLE pvnaive.direct_naive_accounting_session_peers (
    runtime_credential_id uuid NOT NULL,
    node_id text NOT NULL CHECK (length(node_id) BETWEEN 1 AND 160),
    boot_id uuid NOT NULL,
    session_id uuid NOT NULL,
    service_term_id uuid NOT NULL REFERENCES pvnaive.service_terms(id) ON DELETE RESTRICT,
    tenant_id uuid NOT NULL REFERENCES pvnaive.tenants(id) ON DELETE RESTRICT,
    user_id uuid NOT NULL,
    client_ip inet NOT NULL,
    first_recorded_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    PRIMARY KEY (runtime_credential_id, node_id, boot_id, session_id),
    FOREIGN KEY (runtime_credential_id, node_id, boot_id, session_id)
        REFERENCES pvnaive.direct_naive_accounting_sessions(runtime_credential_id, node_id, boot_id, session_id)
        ON DELETE RESTRICT,
    FOREIGN KEY (service_term_id, tenant_id, user_id)
        REFERENCES pvnaive.service_terms(id, tenant_id, user_id)
        ON DELETE RESTRICT,
    CHECK (family(client_ip) IN (4, 6))
);
CREATE INDEX direct_naive_session_peers_customer_active_idx
    ON pvnaive.direct_naive_accounting_session_peers (tenant_id, user_id, first_recorded_at DESC);

ALTER TABLE pvnaive.direct_naive_accounting_session_peers ENABLE ROW LEVEL SECURITY;
ALTER TABLE pvnaive.direct_naive_accounting_session_peers FORCE ROW LEVEL SECURITY;
CREATE POLICY direct_naive_session_peers_tenant_isolation ON pvnaive.direct_naive_accounting_session_peers
    FOR SELECT TO pvnaive_app
    USING (pvnaive.has_tenant_access(tenant_id, false));
CREATE POLICY direct_naive_session_peers_internal_owner ON pvnaive.direct_naive_accounting_session_peers
    FOR ALL TO pvnaive_owner
    USING (pvnaive.has_tenant_access(tenant_id, false))
    WITH CHECK (pvnaive.has_tenant_access(tenant_id, false));
REVOKE ALL ON pvnaive.direct_naive_accounting_session_peers FROM PUBLIC, pvnaive_app;

CREATE FUNCTION pvnaive.direct_naive_accounting_record_session_peer(
    p_runtime_credential_id uuid,
    p_node_id text,
    p_boot_id uuid,
    p_session_id uuid,
    p_client_ip inet,
    p_observed_at timestamptz
)
RETURNS TABLE(service_term_id uuid, recorded boolean, duplicate boolean)
LANGUAGE plpgsql
VOLATILE
SECURITY DEFINER
SET search_path = pg_catalog, pvnaive
AS $$
DECLARE
    v_session pvnaive.direct_naive_accounting_sessions%ROWTYPE;
    v_term pvnaive.service_terms%ROWTYPE;
    v_existing pvnaive.direct_naive_accounting_session_peers%ROWTYPE;
BEGIN
    IF p_runtime_credential_id IS NULL OR p_boot_id IS NULL OR p_session_id IS NULL
       OR p_node_id IS NULL OR length(btrim(p_node_id)) NOT BETWEEN 1 AND 160
       OR p_node_id <> btrim(p_node_id) OR p_client_ip IS NULL OR p_observed_at IS NULL THEN
        RAISE EXCEPTION 'invalid session peer' USING ERRCODE = '22023';
    END IF;

    PERFORM pvnaive.direct_naive_accounting_enter_context();

    SELECT s.* INTO v_session
      FROM pvnaive.direct_naive_accounting_sessions AS s
     WHERE s.runtime_credential_id = p_runtime_credential_id
       AND s.node_id = p_node_id
       AND s.boot_id = p_boot_id
       AND s.session_id = p_session_id
     FOR UPDATE;
    IF NOT FOUND THEN
        PERFORM pvnaive.direct_naive_accounting_leave_context();
        RAISE EXCEPTION 'accounting session not found' USING ERRCODE = '23503';
    END IF;
    IF p_observed_at < v_session.first_observed_at THEN
        PERFORM pvnaive.direct_naive_accounting_leave_context();
        RAISE EXCEPTION 'session peer predates accounting session' USING ERRCODE = '22023';
    END IF;

    SELECT st.* INTO STRICT v_term
      FROM pvnaive.service_terms AS st
     WHERE st.id = v_session.service_term_id;

    SELECT p.* INTO v_existing
      FROM pvnaive.direct_naive_accounting_session_peers AS p
     WHERE p.runtime_credential_id = p_runtime_credential_id
       AND p.node_id = p_node_id
       AND p.boot_id = p_boot_id
       AND p.session_id = p_session_id
     FOR UPDATE;

    IF FOUND THEN
        IF v_existing.client_ip <> p_client_ip
           OR v_existing.service_term_id <> v_session.service_term_id
           OR v_existing.tenant_id <> v_term.tenant_id
           OR v_existing.user_id <> v_term.user_id THEN
            PERFORM pvnaive.direct_naive_accounting_leave_context();
            RAISE EXCEPTION 'session peer identity conflict' USING ERRCODE = '23505';
        END IF;
        UPDATE pvnaive.direct_naive_accounting_session_peers
           SET updated_at = GREATEST(updated_at, p_observed_at)
         WHERE runtime_credential_id = p_runtime_credential_id
           AND node_id = p_node_id AND boot_id = p_boot_id AND session_id = p_session_id;
        PERFORM pvnaive.direct_naive_accounting_leave_context();
        RETURN QUERY SELECT v_session.service_term_id, false, true;
        RETURN;
    END IF;

    INSERT INTO pvnaive.direct_naive_accounting_session_peers(
        runtime_credential_id,node_id,boot_id,session_id,service_term_id,tenant_id,user_id,
        client_ip,first_recorded_at,updated_at
    ) VALUES (
        p_runtime_credential_id,p_node_id,p_boot_id,p_session_id,v_session.service_term_id,
        v_term.tenant_id,v_term.user_id,p_client_ip,p_observed_at,p_observed_at
    );
    PERFORM pvnaive.direct_naive_accounting_leave_context();
    RETURN QUERY SELECT v_session.service_term_id, true, false;
END;
$$;

CREATE FUNCTION pvnaive.list_active_customer_sessions(
    p_user_id uuid,
    p_observed_at timestamptz,
    p_stale_after_seconds integer DEFAULT 90
)
RETURNS TABLE(
    runtime_credential_id uuid,
    node_id text,
    boot_id uuid,
    session_id uuid,
    service_term_id uuid,
    client_ip text,
    connected_at timestamptz,
    last_activity_at timestamptz,
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
    IF p_user_id IS NULL OR p_observed_at IS NULL OR p_stale_after_seconds < 1 OR p_stale_after_seconds > 3600 THEN
        RAISE EXCEPTION 'invalid active session query' USING ERRCODE = '22023';
    END IF;
    IF NOT pvnaive.has_valid_context() THEN
        RAISE EXCEPTION 'valid request context required' USING ERRCODE = '28000';
    END IF;

    RETURN QUERY
    SELECT s.runtime_credential_id,s.node_id,s.boot_id,s.session_id,s.service_term_id,
           host(p.client_ip),s.first_observed_at,s.last_observed_at,
           GREATEST(0, floor(extract(epoch FROM (p_observed_at - s.first_observed_at)))::bigint),
           s.upload_cumulative,s.download_cumulative
      FROM pvnaive.direct_naive_accounting_sessions AS s
      JOIN pvnaive.direct_naive_accounting_session_peers AS p
        ON p.runtime_credential_id=s.runtime_credential_id AND p.node_id=s.node_id
       AND p.boot_id=s.boot_id AND p.session_id=s.session_id
      JOIN pvnaive.service_terms AS st ON st.id=s.service_term_id
     WHERE st.user_id=p_user_id
       AND p.user_id=p_user_id
       AND p.tenant_id=st.tenant_id
       AND pvnaive.has_tenant_access(st.tenant_id,false)
       AND s.final=false
       AND s.accounting_complete=true
       AND s.last_observed_at >= p_observed_at - make_interval(secs => p_stale_after_seconds)
       AND s.first_observed_at <= p_observed_at
     ORDER BY s.last_observed_at DESC,s.session_id;
END;
$$;

REVOKE ALL ON FUNCTION pvnaive.direct_naive_accounting_record_session_peer(uuid,text,uuid,uuid,inet,timestamptz) FROM PUBLIC;
REVOKE ALL ON FUNCTION pvnaive.list_active_customer_sessions(uuid,timestamptz,integer) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION pvnaive.direct_naive_accounting_record_session_peer(uuid,text,uuid,uuid,inet,timestamptz) TO pvnaive_app;
GRANT EXECUTE ON FUNCTION pvnaive.list_active_customer_sessions(uuid,timestamptz,integer) TO pvnaive_app;

COMMENT ON TABLE pvnaive.direct_naive_accounting_session_peers IS 'Trusted CONNECT peer identity captured from Caddy RemoteAddr; not client headers';
COMMENT ON FUNCTION pvnaive.list_active_customer_sessions(uuid,timestamptz,integer) IS 'Tenant-scoped active session projection; excludes stale/final/incomplete/unattributed sessions';
