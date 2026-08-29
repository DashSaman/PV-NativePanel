-- pvnaive:migration-version 0009
-- Source: PVNaive exact Direct Naive accounting ledger
-- pvnaive:migration-name direct_naive_exact_accounting
-- pvnaive:transactional true
-- pvnaive:destructive false

SET LOCAL lock_timeout = '10s';
SET LOCAL statement_timeout = '120s';
SET LOCAL idle_in_transaction_session_timeout = '60s';
SET LOCAL ROLE pvnaive_owner;

CREATE TABLE pvnaive.direct_naive_accounting_terms (
    service_term_id uuid PRIMARY KEY REFERENCES pvnaive.service_terms(id) ON DELETE RESTRICT,
    upload_bytes bigint NOT NULL DEFAULT 0 CHECK (upload_bytes >= 0),
    download_bytes bigint NOT NULL DEFAULT 0 CHECK (download_bytes >= 0),
    reserved_bytes bigint NOT NULL DEFAULT 0 CHECK (reserved_bytes >= 0),
    last_online timestamptz,
    last_telemetry_at timestamptz,
    accounting_complete boolean NOT NULL DEFAULT true,
    updated_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    CHECK (upload_bytes <= 9223372036854775807 - download_bytes),
    CHECK (upload_bytes + download_bytes <= 9223372036854775807 - reserved_bytes)
);

CREATE TABLE pvnaive.direct_naive_accounting_sessions (
    runtime_credential_id uuid NOT NULL REFERENCES pvnaive.naive_runtime_credentials(id) ON DELETE RESTRICT,
    node_id text NOT NULL CHECK (length(node_id) BETWEEN 1 AND 160),
    boot_id uuid NOT NULL,
    session_id uuid NOT NULL,
    service_term_id uuid NOT NULL REFERENCES pvnaive.service_terms(id) ON DELETE RESTRICT,
    first_observed_at timestamptz NOT NULL,
    last_observed_at timestamptz NOT NULL,
    last_sequence bigint NOT NULL CHECK (last_sequence >= 1),
    upload_cumulative bigint NOT NULL DEFAULT 0 CHECK (upload_cumulative >= 0),
    download_cumulative bigint NOT NULL DEFAULT 0 CHECK (download_cumulative >= 0),
    final boolean NOT NULL DEFAULT false,
    accounting_complete boolean NOT NULL DEFAULT true,
    updated_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    PRIMARY KEY (runtime_credential_id, node_id, boot_id, session_id),
    CHECK (last_observed_at >= first_observed_at),
    CHECK (upload_cumulative <= 9223372036854775807 - download_cumulative)
);
CREATE INDEX direct_naive_accounting_sessions_term_seen_idx
    ON pvnaive.direct_naive_accounting_sessions (service_term_id, last_observed_at DESC);
CREATE INDEX direct_naive_accounting_sessions_open_idx
    ON pvnaive.direct_naive_accounting_sessions (service_term_id, last_observed_at DESC)
    WHERE final = false;

CREATE TABLE pvnaive.direct_naive_accounting_claims (
    id uuid PRIMARY KEY DEFAULT public.gen_random_uuid(),
    runtime_credential_id uuid NOT NULL REFERENCES pvnaive.naive_runtime_credentials(id) ON DELETE RESTRICT,
    node_id text NOT NULL CHECK (length(node_id) BETWEEN 1 AND 160),
    boot_id uuid NOT NULL,
    session_id uuid NOT NULL,
    source_sequence bigint NOT NULL CHECK (source_sequence >= 2),
    service_term_id uuid NOT NULL REFERENCES pvnaive.service_terms(id) ON DELETE RESTRICT,
    direction text NOT NULL CHECK (direction IN ('upload', 'download')),
    requested_bytes bigint NOT NULL CHECK (requested_bytes > 0),
    reserved_bytes bigint NOT NULL CHECK (reserved_bytes > 0 AND reserved_bytes <= requested_bytes),
    settled_bytes bigint CHECK (settled_bytes IS NULL OR (settled_bytes >= 0 AND settled_bytes <= reserved_bytes)),
    claimed_at timestamptz NOT NULL,
    settled_at timestamptz,
    UNIQUE (runtime_credential_id, node_id, boot_id, session_id, source_sequence),
    CHECK ((settled_at IS NULL) = (settled_bytes IS NULL))
);
CREATE INDEX direct_naive_accounting_claims_term_unsettled_idx
    ON pvnaive.direct_naive_accounting_claims (service_term_id, claimed_at)
    WHERE settled_at IS NULL;

CREATE TABLE pvnaive.direct_naive_accounting_events (
    ledger_sequence bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    id uuid NOT NULL UNIQUE DEFAULT public.gen_random_uuid(),
    service_term_id uuid NOT NULL REFERENCES pvnaive.service_terms(id) ON DELETE RESTRICT,
    runtime_credential_id uuid NOT NULL REFERENCES pvnaive.naive_runtime_credentials(id) ON DELETE RESTRICT,
    username_diagnostic text NOT NULL CHECK (length(username_diagnostic) BETWEEN 1 AND 160),
    node_id text NOT NULL CHECK (length(node_id) BETWEEN 1 AND 160),
    boot_id uuid NOT NULL,
    session_id uuid NOT NULL,
    source_sequence bigint NOT NULL CHECK (source_sequence >= 1),
    observed_at timestamptz NOT NULL,
    authenticated_connect boolean NOT NULL CHECK (authenticated_connect = true),
    upload_cumulative bigint NOT NULL CHECK (upload_cumulative >= 0),
    download_cumulative bigint NOT NULL CHECK (download_cumulative >= 0),
    upload_delta bigint NOT NULL CHECK (upload_delta >= 0),
    download_delta bigint NOT NULL CHECK (download_delta >= 0),
    final boolean NOT NULL DEFAULT false,
    ingested_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    UNIQUE (runtime_credential_id, node_id, boot_id, session_id, source_sequence),
    CHECK (upload_cumulative <= 9223372036854775807 - download_cumulative),
    CHECK (upload_delta <= 9223372036854775807 - download_delta)
);
CREATE INDEX direct_naive_accounting_events_term_observed_idx
    ON pvnaive.direct_naive_accounting_events (service_term_id, observed_at, ledger_sequence);

CREATE TRIGGER direct_naive_accounting_events_immutable
BEFORE UPDATE OR DELETE ON pvnaive.direct_naive_accounting_events
FOR EACH ROW EXECUTE FUNCTION pvnaive.prevent_immutable_mutation();

-- Accounting relations are a system data-path boundary. Browser/API callers do
-- not receive direct table privileges; the only app-role entry points are the
-- narrowly typed SECURITY DEFINER functions below.
REVOKE ALL ON pvnaive.direct_naive_accounting_terms FROM PUBLIC, pvnaive_app;
REVOKE ALL ON pvnaive.direct_naive_accounting_sessions FROM PUBLIC, pvnaive_app;
REVOKE ALL ON pvnaive.direct_naive_accounting_claims FROM PUBLIC, pvnaive_app;
REVOKE ALL ON pvnaive.direct_naive_accounting_events FROM PUBLIC, pvnaive_app;
REVOKE ALL ON SEQUENCE pvnaive.direct_naive_accounting_events_ledger_sequence_seq FROM PUBLIC, pvnaive_app;

-- FORCE-RLS source tables deliberately cannot be read by pvnaive_app without a
-- signed request context. Accounting has no browser actor. These private helpers
-- install a transaction-local, HMAC-signed synthetic Owner context only while a
-- SECURITY DEFINER accounting function is executing. They are never executable
-- by pvnaive_app or PUBLIC directly.
CREATE FUNCTION pvnaive.direct_naive_accounting_enter_context()
RETURNS void
LANGUAGE plpgsql
VOLATILE
SECURITY DEFINER
SET search_path = pg_catalog, pvnaive
AS $$
DECLARE
    actor_id_value constant uuid := '00000000-0000-0000-0000-000000000001'::uuid;
    actor_role_value constant text := 'owner';
    signing_key_value bytea;
    signature_value text;
BEGIN
    SELECT signing_key INTO STRICT signing_key_value
      FROM pvnaive.security_context_keys
     WHERE singleton;
    signature_value := encode(
        public.hmac(
            pvnaive.context_payload(actor_id_value, NULL, actor_role_value),
            signing_key_value,
            'sha256'
        ),
        'hex'
    );
    PERFORM set_config('pvnaive.actor_id', actor_id_value::text, true);
    PERFORM set_config('pvnaive.tenant_id', '', true);
    PERFORM set_config('pvnaive.actor_role', actor_role_value, true);
    PERFORM set_config('pvnaive.context_signature', signature_value, true);
END;
$$;

CREATE FUNCTION pvnaive.direct_naive_accounting_leave_context()
RETURNS void
LANGUAGE plpgsql
VOLATILE
SECURITY DEFINER
SET search_path = pg_catalog, pvnaive
AS $$
BEGIN
    PERFORM set_config('pvnaive.actor_id', '', true);
    PERFORM set_config('pvnaive.tenant_id', '', true);
    PERFORM set_config('pvnaive.actor_role', '', true);
    PERFORM set_config('pvnaive.context_signature', '', true);
END;
$$;
REVOKE ALL ON FUNCTION pvnaive.direct_naive_accounting_enter_context() FROM PUBLIC, pvnaive_app;
REVOKE ALL ON FUNCTION pvnaive.direct_naive_accounting_leave_context() FROM PUBLIC, pvnaive_app;

CREATE FUNCTION pvnaive.direct_naive_accounting_authorize(
    p_runtime_credential_id uuid,
    p_observed_at timestamptz
)
RETURNS TABLE (
    service_term_id uuid,
    tracked boolean,
    allowed boolean,
    reason text,
    quota_bytes bigint,
    used_bytes bigint,
    reserved_bytes bigint,
    remaining_bytes bigint,
    expires_at timestamptz,
    first_connected_at timestamptz,
    accounting_complete boolean
)
LANGUAGE plpgsql
VOLATILE
SECURITY DEFINER
SET search_path = pg_catalog, pvnaive
AS $$
DECLARE
    v_term pvnaive.service_terms%ROWTYPE;
    v_upload bigint := 0;
    v_download bigint := 0;
    v_reserved bigint := 0;
    v_complete boolean := true;
    v_used bigint := 0;
    v_remaining bigint;
    v_allowed boolean := true;
    v_reason text := 'allowed';
BEGIN
    IF p_runtime_credential_id IS NULL OR p_observed_at IS NULL THEN
        RAISE EXCEPTION 'invalid accounting authorization' USING ERRCODE = '22023';
    END IF;
    PERFORM pvnaive.direct_naive_accounting_enter_context();

    SELECT st.*
      INTO v_term
      FROM pvnaive.user_runtime_credentials urc
      JOIN pvnaive.naive_runtime_credentials rc
        ON rc.id = urc.runtime_credential_id AND rc.status = 'active'
      JOIN pvnaive.users u
        ON u.id = urc.user_id AND u.tenant_id = urc.tenant_id AND u.status = 'active'
      JOIN pvnaive.service_terms st
        ON st.id = urc.service_term_id AND st.tenant_id = urc.tenant_id AND st.user_id = urc.user_id
     WHERE urc.runtime_credential_id = p_runtime_credential_id
       AND urc.unbound_at IS NULL
       AND urc.role = 'primary'
     LIMIT 1;

    IF NOT FOUND THEN
        PERFORM pvnaive.direct_naive_accounting_leave_context();
        RETURN QUERY SELECT NULL::uuid, false, true, 'untracked'::text,
                            NULL::bigint, 0::bigint, 0::bigint, NULL::bigint,
                            NULL::timestamptz, NULL::timestamptz, true;
        RETURN;
    END IF;

    INSERT INTO pvnaive.direct_naive_accounting_terms(service_term_id)
    VALUES (v_term.id)
    ON CONFLICT ON CONSTRAINT direct_naive_accounting_terms_pkey DO NOTHING;

    SELECT t.upload_bytes, t.download_bytes, t.reserved_bytes, t.accounting_complete
      INTO v_upload, v_download, v_reserved, v_complete
      FROM pvnaive.direct_naive_accounting_terms t
     WHERE t.service_term_id = v_term.id;
    v_used := v_upload + v_download;

    IF v_term.state IN ('expired', 'ended', 'revoked') THEN
        v_allowed := false;
        v_reason := v_term.state;
    ELSIF v_term.expires_at IS NOT NULL AND v_term.expires_at <= p_observed_at THEN
        v_allowed := false;
        v_reason := 'expired';
    END IF;

    IF v_term.quota_bytes IS NOT NULL THEN
        v_remaining := GREATEST(v_term.quota_bytes - v_used - v_reserved, 0);
        IF v_remaining = 0 THEN
            v_allowed := false;
            IF v_used >= v_term.quota_bytes THEN
                v_reason := 'quota_depleted';
            ELSE
                v_reason := 'quota_reserved';
            END IF;
        ELSIF v_term.state = 'quota_depleted' AND v_term.first_connected_at IS NOT NULL THEN
            UPDATE pvnaive.service_terms
               SET state = 'active', revision = revision + 1, updated_at = p_observed_at
             WHERE id = v_term.id AND state = 'quota_depleted';
            v_term.state := 'active';
        END IF;
    END IF;

    PERFORM pvnaive.direct_naive_accounting_leave_context();
    RETURN QUERY SELECT v_term.id, true, v_allowed, v_reason, v_term.quota_bytes,
                        v_used, v_reserved, v_remaining, v_term.expires_at,
                        v_term.first_connected_at, v_complete;
END;
$$;

CREATE FUNCTION pvnaive.direct_naive_accounting_claim(
    p_runtime_credential_id uuid,
    p_node_id text,
    p_boot_id uuid,
    p_session_id uuid,
    p_source_sequence bigint,
    p_direction text,
    p_requested_bytes bigint,
    p_observed_at timestamptz
)
RETURNS TABLE (
    service_term_id uuid,
    tracked boolean,
    allowed boolean,
    reason text,
    granted_bytes bigint,
    quota_bytes bigint,
    used_bytes bigint,
    reserved_total_bytes bigint,
    remaining_bytes bigint,
    expires_at timestamptz,
    accounting_complete boolean
)
LANGUAGE plpgsql
VOLATILE
SECURITY DEFINER
SET search_path = pg_catalog, pvnaive
AS $$
DECLARE
    v_term pvnaive.service_terms%ROWTYPE;
    v_session pvnaive.direct_naive_accounting_sessions%ROWTYPE;
    v_claim pvnaive.direct_naive_accounting_claims%ROWTYPE;
    v_projection pvnaive.direct_naive_accounting_terms%ROWTYPE;
    v_used bigint;
    v_remaining bigint;
    v_grant bigint;
BEGIN
    IF p_runtime_credential_id IS NULL OR p_boot_id IS NULL OR p_session_id IS NULL
       OR p_node_id IS NULL OR length(btrim(p_node_id)) NOT BETWEEN 1 AND 160
       OR p_source_sequence < 2 OR p_direction NOT IN ('upload','download')
       OR p_requested_bytes <= 0 OR p_observed_at IS NULL THEN
        RAISE EXCEPTION 'invalid accounting claim' USING ERRCODE = '22023';
    END IF;
    PERFORM pvnaive.direct_naive_accounting_enter_context();

    SELECT st.*
      INTO v_term
      FROM pvnaive.user_runtime_credentials urc
      JOIN pvnaive.naive_runtime_credentials rc
        ON rc.id = urc.runtime_credential_id AND rc.status = 'active'
      JOIN pvnaive.users u
        ON u.id = urc.user_id AND u.tenant_id = urc.tenant_id AND u.status = 'active'
      JOIN pvnaive.service_terms st
        ON st.id = urc.service_term_id AND st.tenant_id = urc.tenant_id AND st.user_id = urc.user_id
     WHERE urc.runtime_credential_id = p_runtime_credential_id
       AND urc.unbound_at IS NULL AND urc.role = 'primary'
     FOR UPDATE OF st
     LIMIT 1;

    IF NOT FOUND THEN
        PERFORM pvnaive.direct_naive_accounting_leave_context();
        RETURN QUERY SELECT NULL::uuid, false, true, 'untracked'::text,
                            p_requested_bytes, NULL::bigint, 0::bigint, 0::bigint,
                            NULL::bigint, NULL::timestamptz, true;
        RETURN;
    END IF;

    INSERT INTO pvnaive.direct_naive_accounting_terms(service_term_id)
    VALUES (v_term.id)
    ON CONFLICT ON CONSTRAINT direct_naive_accounting_terms_pkey DO NOTHING;
    SELECT * INTO v_projection
      FROM pvnaive.direct_naive_accounting_terms
     WHERE direct_naive_accounting_terms.service_term_id = v_term.id
     FOR UPDATE;

    SELECT * INTO v_session
      FROM pvnaive.direct_naive_accounting_sessions s
     WHERE s.runtime_credential_id = p_runtime_credential_id
       AND s.node_id = btrim(p_node_id)
       AND s.boot_id = p_boot_id
       AND s.session_id = p_session_id
     FOR UPDATE;
    IF NOT FOUND OR v_session.service_term_id <> v_term.id OR v_session.final THEN
        PERFORM pvnaive.direct_naive_accounting_leave_context();
        RETURN QUERY SELECT v_term.id, true, false, 'session_not_open'::text,
                            0::bigint, v_term.quota_bytes,
                            (v_projection.upload_bytes + v_projection.download_bytes),
                            v_projection.reserved_bytes,
                            CASE WHEN v_term.quota_bytes IS NULL THEN NULL::bigint ELSE GREATEST(v_term.quota_bytes - v_projection.upload_bytes - v_projection.download_bytes - v_projection.reserved_bytes, 0) END,
                            v_term.expires_at, false;
        RETURN;
    END IF;

    SELECT * INTO v_claim
      FROM pvnaive.direct_naive_accounting_claims c
     WHERE c.runtime_credential_id = p_runtime_credential_id
       AND c.node_id = btrim(p_node_id)
       AND c.boot_id = p_boot_id
       AND c.session_id = p_session_id
       AND c.source_sequence = p_source_sequence;
    IF FOUND THEN
        IF v_claim.direction = p_direction AND v_claim.requested_bytes = p_requested_bytes THEN
            v_used := v_projection.upload_bytes + v_projection.download_bytes;
            v_remaining := CASE WHEN v_term.quota_bytes IS NULL THEN NULL ELSE GREATEST(v_term.quota_bytes - v_used - v_projection.reserved_bytes, 0) END;
            PERFORM pvnaive.direct_naive_accounting_leave_context();
            RETURN QUERY SELECT v_term.id, true, true, 'duplicate_claim'::text,
                                v_claim.reserved_bytes, v_term.quota_bytes, v_used,
                                v_projection.reserved_bytes, v_remaining, v_term.expires_at,
                                v_projection.accounting_complete;
            RETURN;
        END IF;
        UPDATE pvnaive.direct_naive_accounting_terms
           SET accounting_complete=false, updated_at=p_observed_at
         WHERE direct_naive_accounting_terms.service_term_id=v_term.id;
        PERFORM pvnaive.direct_naive_accounting_leave_context();
        RETURN QUERY SELECT v_term.id, true, false, 'claim_conflict'::text,
                            0::bigint, v_term.quota_bytes,
                            (v_projection.upload_bytes + v_projection.download_bytes),
                            v_projection.reserved_bytes, 0::bigint, v_term.expires_at, false;
        RETURN;
    END IF;

    IF p_source_sequence <> v_session.last_sequence + 1 THEN
        UPDATE pvnaive.direct_naive_accounting_terms
           SET accounting_complete=false, updated_at=p_observed_at
         WHERE direct_naive_accounting_terms.service_term_id=v_term.id;
        UPDATE pvnaive.direct_naive_accounting_sessions
           SET accounting_complete=false, updated_at=p_observed_at
         WHERE runtime_credential_id=p_runtime_credential_id AND node_id=btrim(p_node_id)
           AND boot_id=p_boot_id AND session_id=p_session_id;
        PERFORM pvnaive.direct_naive_accounting_leave_context();
        RETURN QUERY SELECT v_term.id, true, false, 'sequence_not_next'::text,
                            0::bigint, v_term.quota_bytes,
                            (v_projection.upload_bytes + v_projection.download_bytes),
                            v_projection.reserved_bytes, 0::bigint, v_term.expires_at, false;
        RETURN;
    END IF;

    v_used := v_projection.upload_bytes + v_projection.download_bytes;
    IF v_term.state IN ('expired','ended','revoked') OR (v_term.expires_at IS NOT NULL AND v_term.expires_at <= p_observed_at) THEN
        PERFORM pvnaive.direct_naive_accounting_leave_context();
        RETURN QUERY SELECT v_term.id, true, false,
                            CASE WHEN v_term.expires_at IS NOT NULL AND v_term.expires_at <= p_observed_at THEN 'expired' ELSE v_term.state END,
                            0::bigint, v_term.quota_bytes, v_used, v_projection.reserved_bytes,
                            CASE WHEN v_term.quota_bytes IS NULL THEN NULL::bigint ELSE GREATEST(v_term.quota_bytes-v_used-v_projection.reserved_bytes,0) END,
                            v_term.expires_at, v_projection.accounting_complete;
        RETURN;
    END IF;

    IF v_term.quota_bytes IS NULL THEN
        v_grant := p_requested_bytes;
        v_remaining := NULL;
    ELSE
        v_remaining := GREATEST(v_term.quota_bytes - v_used - v_projection.reserved_bytes, 0);
        v_grant := LEAST(p_requested_bytes, v_remaining);
        IF v_grant = 0 THEN
            PERFORM pvnaive.direct_naive_accounting_leave_context();
            RETURN QUERY SELECT v_term.id, true, false,
                                CASE WHEN v_used >= v_term.quota_bytes THEN 'quota_depleted' ELSE 'quota_reserved' END,
                                0::bigint, v_term.quota_bytes, v_used, v_projection.reserved_bytes,
                                0::bigint, v_term.expires_at, v_projection.accounting_complete;
            RETURN;
        END IF;
        v_remaining := v_remaining - v_grant;
        IF v_term.state = 'quota_depleted' AND v_term.first_connected_at IS NOT NULL THEN
            UPDATE pvnaive.service_terms
               SET state='active', revision=revision+1, updated_at=p_observed_at
             WHERE id=v_term.id AND state='quota_depleted';
        END IF;
    END IF;

    INSERT INTO pvnaive.direct_naive_accounting_claims (
        runtime_credential_id,node_id,boot_id,session_id,source_sequence,service_term_id,
        direction,requested_bytes,reserved_bytes,claimed_at
    ) VALUES (
        p_runtime_credential_id,btrim(p_node_id),p_boot_id,p_session_id,p_source_sequence,v_term.id,
        p_direction,p_requested_bytes,v_grant,p_observed_at
    );
    UPDATE pvnaive.direct_naive_accounting_terms
       SET reserved_bytes=reserved_bytes+v_grant,
           last_telemetry_at=p_observed_at,
           updated_at=p_observed_at
     WHERE direct_naive_accounting_terms.service_term_id=v_term.id
     RETURNING direct_naive_accounting_terms.reserved_bytes INTO v_projection.reserved_bytes;

    PERFORM pvnaive.direct_naive_accounting_leave_context();
    RETURN QUERY SELECT v_term.id, true, true, 'granted'::text,
                        v_grant, v_term.quota_bytes, v_used, v_projection.reserved_bytes,
                        v_remaining, v_term.expires_at, v_projection.accounting_complete;
END;
$$;

CREATE FUNCTION pvnaive.direct_naive_accounting_ingest(
    p_runtime_credential_id uuid,
    p_username_diagnostic text,
    p_node_id text,
    p_boot_id uuid,
    p_session_id uuid,
    p_source_sequence bigint,
    p_observed_at timestamptz,
    p_authenticated_connect boolean,
    p_upload_cumulative bigint,
    p_download_cumulative bigint,
    p_final boolean DEFAULT false
)
RETURNS TABLE (
    service_term_id uuid,
    tracked boolean,
    accepted boolean,
    duplicate boolean,
    reason text,
    upload_delta bigint,
    download_delta bigint,
    quota_depleted boolean,
    remaining_bytes bigint,
    first_connected_at timestamptz,
    accounting_complete boolean
)
LANGUAGE plpgsql
VOLATILE
SECURITY DEFINER
SET search_path = pg_catalog, pvnaive
AS $$
DECLARE
    v_term pvnaive.service_terms%ROWTYPE;
    v_session pvnaive.direct_naive_accounting_sessions%ROWTYPE;
    v_event pvnaive.direct_naive_accounting_events%ROWTYPE;
    v_claim pvnaive.direct_naive_accounting_claims%ROWTYPE;
    v_projection pvnaive.direct_naive_accounting_terms%ROWTYPE;
    v_upload_delta bigint := 0;
    v_download_delta bigint := 0;
    v_used bigint := 0;
    v_remaining bigint;
    v_quota_depleted boolean := false;
    v_reason text := 'accepted';
BEGIN
    IF p_runtime_credential_id IS NULL OR p_boot_id IS NULL OR p_session_id IS NULL
       OR p_username_diagnostic IS NULL OR length(btrim(p_username_diagnostic)) NOT BETWEEN 1 AND 160
       OR p_node_id IS NULL OR length(btrim(p_node_id)) NOT BETWEEN 1 AND 160
       OR p_source_sequence < 1 OR p_observed_at IS NULL OR p_authenticated_connect IS DISTINCT FROM true
       OR p_upload_cumulative < 0 OR p_download_cumulative < 0
       OR p_upload_cumulative > 9223372036854775807 - p_download_cumulative THEN
        RAISE EXCEPTION 'invalid accounting event' USING ERRCODE = '22023';
    END IF;
    PERFORM pvnaive.direct_naive_accounting_enter_context();

    SELECT st.*
      INTO v_term
      FROM pvnaive.user_runtime_credentials urc
      JOIN pvnaive.naive_runtime_credentials rc
        ON rc.id=urc.runtime_credential_id AND rc.status='active'
      JOIN pvnaive.users u
        ON u.id=urc.user_id AND u.tenant_id=urc.tenant_id AND u.status='active'
      JOIN pvnaive.service_terms st
        ON st.id=urc.service_term_id AND st.tenant_id=urc.tenant_id AND st.user_id=urc.user_id
     WHERE urc.runtime_credential_id=p_runtime_credential_id
       AND urc.unbound_at IS NULL AND urc.role='primary'
     FOR UPDATE OF st
     LIMIT 1;

    IF NOT FOUND THEN
        PERFORM pvnaive.direct_naive_accounting_leave_context();
        RETURN QUERY SELECT NULL::uuid,false,false,false,'untracked'::text,0::bigint,0::bigint,false,NULL::bigint,NULL::timestamptz,true;
        RETURN;
    END IF;

    INSERT INTO pvnaive.direct_naive_accounting_terms(service_term_id)
    VALUES(v_term.id) ON CONFLICT ON CONSTRAINT direct_naive_accounting_terms_pkey DO NOTHING;
    SELECT * INTO v_projection FROM pvnaive.direct_naive_accounting_terms t
     WHERE t.service_term_id=v_term.id FOR UPDATE;

    SELECT * INTO v_event
      FROM pvnaive.direct_naive_accounting_events e
     WHERE e.runtime_credential_id=p_runtime_credential_id
       AND e.node_id=btrim(p_node_id) AND e.boot_id=p_boot_id AND e.session_id=p_session_id
       AND e.source_sequence=p_source_sequence;
    IF FOUND THEN
        IF v_event.service_term_id=v_term.id
           AND v_event.username_diagnostic=btrim(p_username_diagnostic)
           AND v_event.observed_at=p_observed_at
           AND v_event.authenticated_connect=p_authenticated_connect
           AND v_event.upload_cumulative=p_upload_cumulative
           AND v_event.download_cumulative=p_download_cumulative
           AND v_event.final=p_final THEN
            v_used := v_projection.upload_bytes+v_projection.download_bytes;
            v_remaining := CASE WHEN v_term.quota_bytes IS NULL THEN NULL ELSE GREATEST(v_term.quota_bytes-v_used-v_projection.reserved_bytes,0) END;
            PERFORM pvnaive.direct_naive_accounting_leave_context();
            RETURN QUERY SELECT v_term.id,true,true,true,'duplicate'::text,0::bigint,0::bigint,
                                (v_term.quota_bytes IS NOT NULL AND v_used>=v_term.quota_bytes),v_remaining,
                                v_term.first_connected_at,v_projection.accounting_complete;
            RETURN;
        END IF;
        UPDATE pvnaive.direct_naive_accounting_terms SET accounting_complete=false,updated_at=p_observed_at
         WHERE direct_naive_accounting_terms.service_term_id=v_term.id;
        PERFORM pvnaive.direct_naive_accounting_leave_context();
        RETURN QUERY SELECT v_term.id,true,false,false,'sequence_conflict'::text,0::bigint,0::bigint,false,0::bigint,v_term.first_connected_at,false;
        RETURN;
    END IF;

    SELECT * INTO v_session
      FROM pvnaive.direct_naive_accounting_sessions s
     WHERE s.runtime_credential_id=p_runtime_credential_id
       AND s.node_id=btrim(p_node_id) AND s.boot_id=p_boot_id AND s.session_id=p_session_id
     FOR UPDATE;

    IF NOT FOUND THEN
        IF p_source_sequence<>1 OR p_upload_cumulative<>0 OR p_download_cumulative<>0 THEN
            UPDATE pvnaive.direct_naive_accounting_terms SET accounting_complete=false,updated_at=p_observed_at
             WHERE direct_naive_accounting_terms.service_term_id=v_term.id;
            PERFORM pvnaive.direct_naive_accounting_leave_context();
            RETURN QUERY SELECT v_term.id,true,false,false,'missing_open_event'::text,0::bigint,0::bigint,false,
                                CASE WHEN v_term.quota_bytes IS NULL THEN NULL::bigint ELSE GREATEST(v_term.quota_bytes-v_projection.upload_bytes-v_projection.download_bytes-v_projection.reserved_bytes,0) END,
                                v_term.first_connected_at,false;
            RETURN;
        END IF;

        IF v_term.state='pending' AND v_term.start_policy='on_first_successful_connection'
           AND v_term.starts_at IS NULL AND v_term.first_connected_at IS NULL AND v_term.expires_at IS NULL
           AND p_observed_at>=v_term.purchased_at THEN
            UPDATE pvnaive.service_terms
               SET starts_at=p_observed_at,
                   first_connected_at=p_observed_at,
                   expires_at=p_observed_at+(duration_seconds*interval '1 second'),
                   state='active', revision=revision+1, updated_at=p_observed_at
             WHERE id=v_term.id
             RETURNING * INTO v_term;
        END IF;

        INSERT INTO pvnaive.direct_naive_accounting_sessions (
            runtime_credential_id,node_id,boot_id,session_id,service_term_id,
            first_observed_at,last_observed_at,last_sequence,upload_cumulative,download_cumulative,final
        ) VALUES (
            p_runtime_credential_id,btrim(p_node_id),p_boot_id,p_session_id,v_term.id,
            p_observed_at,p_observed_at,1,0,0,p_final
        );
        INSERT INTO pvnaive.direct_naive_accounting_events (
            service_term_id,runtime_credential_id,username_diagnostic,node_id,boot_id,session_id,
            source_sequence,observed_at,authenticated_connect,upload_cumulative,download_cumulative,
            upload_delta,download_delta,final
        ) VALUES (
            v_term.id,p_runtime_credential_id,btrim(p_username_diagnostic),btrim(p_node_id),p_boot_id,p_session_id,
            1,p_observed_at,true,0,0,0,0,p_final
        );
        UPDATE pvnaive.direct_naive_accounting_terms
           SET last_online=p_observed_at,last_telemetry_at=p_observed_at,updated_at=p_observed_at
         WHERE direct_naive_accounting_terms.service_term_id=v_term.id;
        v_remaining := CASE WHEN v_term.quota_bytes IS NULL THEN NULL ELSE GREATEST(v_term.quota_bytes-v_projection.upload_bytes-v_projection.download_bytes-v_projection.reserved_bytes,0) END;
        PERFORM pvnaive.direct_naive_accounting_leave_context();
        RETURN QUERY SELECT v_term.id,true,true,false,'accepted'::text,0::bigint,0::bigint,false,v_remaining,v_term.first_connected_at,v_projection.accounting_complete;
        RETURN;
    END IF;

    IF v_session.service_term_id<>v_term.id THEN
        UPDATE pvnaive.direct_naive_accounting_terms SET accounting_complete=false,updated_at=p_observed_at WHERE direct_naive_accounting_terms.service_term_id=v_term.id;
        PERFORM pvnaive.direct_naive_accounting_leave_context();
        RETURN QUERY SELECT v_term.id,true,false,false,'service_term_mismatch'::text,0::bigint,0::bigint,false,0::bigint,v_term.first_connected_at,false;
        RETURN;
    END IF;
    IF v_session.final THEN
        PERFORM pvnaive.direct_naive_accounting_leave_context();
        RETURN QUERY SELECT v_term.id,true,false,false,'session_closed'::text,0::bigint,0::bigint,false,0::bigint,v_term.first_connected_at,v_projection.accounting_complete;
        RETURN;
    END IF;
    IF p_source_sequence<>v_session.last_sequence+1 THEN
        UPDATE pvnaive.direct_naive_accounting_terms SET accounting_complete=false,updated_at=p_observed_at WHERE direct_naive_accounting_terms.service_term_id=v_term.id;
        UPDATE pvnaive.direct_naive_accounting_sessions SET accounting_complete=false,updated_at=p_observed_at
         WHERE runtime_credential_id=p_runtime_credential_id AND node_id=btrim(p_node_id) AND boot_id=p_boot_id AND session_id=p_session_id;
        PERFORM pvnaive.direct_naive_accounting_leave_context();
        RETURN QUERY SELECT v_term.id,true,false,false,
                            CASE WHEN p_source_sequence<v_session.last_sequence+1 THEN 'out_of_order' ELSE 'sequence_gap' END,
                            0::bigint,0::bigint,false,0::bigint,v_term.first_connected_at,false;
        RETURN;
    END IF;
    IF p_observed_at<v_session.last_observed_at THEN
        UPDATE pvnaive.direct_naive_accounting_terms SET accounting_complete=false,updated_at=clock_timestamp() WHERE direct_naive_accounting_terms.service_term_id=v_term.id;
        UPDATE pvnaive.direct_naive_accounting_sessions SET accounting_complete=false,updated_at=clock_timestamp()
         WHERE runtime_credential_id=p_runtime_credential_id AND node_id=btrim(p_node_id) AND boot_id=p_boot_id AND session_id=p_session_id;
        PERFORM pvnaive.direct_naive_accounting_leave_context();
        RETURN QUERY SELECT v_term.id,true,false,false,'out_of_order'::text,0::bigint,0::bigint,false,0::bigint,v_term.first_connected_at,false;
        RETURN;
    END IF;
    IF p_upload_cumulative<v_session.upload_cumulative OR p_download_cumulative<v_session.download_cumulative THEN
        UPDATE pvnaive.direct_naive_accounting_terms SET accounting_complete=false,updated_at=p_observed_at WHERE direct_naive_accounting_terms.service_term_id=v_term.id;
        UPDATE pvnaive.direct_naive_accounting_sessions SET accounting_complete=false,updated_at=p_observed_at
         WHERE runtime_credential_id=p_runtime_credential_id AND node_id=btrim(p_node_id) AND boot_id=p_boot_id AND session_id=p_session_id;
        PERFORM pvnaive.direct_naive_accounting_leave_context();
        RETURN QUERY SELECT v_term.id,true,false,false,'counter_regression'::text,0::bigint,0::bigint,false,0::bigint,v_term.first_connected_at,false;
        RETURN;
    END IF;

    v_upload_delta:=p_upload_cumulative-v_session.upload_cumulative;
    v_download_delta:=p_download_cumulative-v_session.download_cumulative;
    IF v_upload_delta>9223372036854775807-v_download_delta THEN
        RAISE EXCEPTION 'accounting delta overflow' USING ERRCODE='22003';
    END IF;

    SELECT * INTO v_claim
      FROM pvnaive.direct_naive_accounting_claims c
     WHERE c.runtime_credential_id=p_runtime_credential_id
       AND c.node_id=btrim(p_node_id) AND c.boot_id=p_boot_id AND c.session_id=p_session_id
       AND c.source_sequence=p_source_sequence
     FOR UPDATE;

    IF v_upload_delta+v_download_delta>0 THEN
        IF NOT FOUND OR v_claim.settled_at IS NOT NULL
           OR (v_claim.direction='upload' AND (v_download_delta<>0 OR v_upload_delta>v_claim.reserved_bytes))
           OR (v_claim.direction='download' AND (v_upload_delta<>0 OR v_download_delta>v_claim.reserved_bytes)) THEN
            UPDATE pvnaive.direct_naive_accounting_terms SET accounting_complete=false,updated_at=p_observed_at WHERE direct_naive_accounting_terms.service_term_id=v_term.id;
            UPDATE pvnaive.direct_naive_accounting_sessions SET accounting_complete=false,updated_at=p_observed_at
             WHERE runtime_credential_id=p_runtime_credential_id AND node_id=btrim(p_node_id) AND boot_id=p_boot_id AND session_id=p_session_id;
            PERFORM pvnaive.direct_naive_accounting_leave_context();
            RETURN QUERY SELECT v_term.id,true,false,false,'unreserved_bytes'::text,0::bigint,0::bigint,false,0::bigint,v_term.first_connected_at,false;
            RETURN;
        END IF;
    END IF;

    IF FOUND AND v_claim.settled_at IS NULL THEN
        UPDATE pvnaive.direct_naive_accounting_claims
           SET settled_bytes=CASE WHEN v_claim.direction='upload' THEN v_upload_delta ELSE v_download_delta END,
               settled_at=p_observed_at
         WHERE id=v_claim.id;
        UPDATE pvnaive.direct_naive_accounting_terms
           SET reserved_bytes=reserved_bytes-v_claim.reserved_bytes
         WHERE direct_naive_accounting_terms.service_term_id=v_term.id;
    END IF;

    INSERT INTO pvnaive.direct_naive_accounting_events (
        service_term_id,runtime_credential_id,username_diagnostic,node_id,boot_id,session_id,
        source_sequence,observed_at,authenticated_connect,upload_cumulative,download_cumulative,
        upload_delta,download_delta,final
    ) VALUES (
        v_term.id,p_runtime_credential_id,btrim(p_username_diagnostic),btrim(p_node_id),p_boot_id,p_session_id,
        p_source_sequence,p_observed_at,true,p_upload_cumulative,p_download_cumulative,
        v_upload_delta,v_download_delta,p_final
    );

    UPDATE pvnaive.direct_naive_accounting_sessions
       SET last_observed_at=p_observed_at,last_sequence=p_source_sequence,
           upload_cumulative=p_upload_cumulative,download_cumulative=p_download_cumulative,
           final=p_final,updated_at=p_observed_at
     WHERE runtime_credential_id=p_runtime_credential_id AND node_id=btrim(p_node_id)
       AND boot_id=p_boot_id AND session_id=p_session_id;

    UPDATE pvnaive.direct_naive_accounting_terms
       SET upload_bytes=upload_bytes+v_upload_delta,
           download_bytes=download_bytes+v_download_delta,
           last_online=p_observed_at,last_telemetry_at=p_observed_at,updated_at=p_observed_at
     WHERE direct_naive_accounting_terms.service_term_id=v_term.id
     RETURNING * INTO v_projection;

    v_used:=v_projection.upload_bytes+v_projection.download_bytes;
    IF v_term.quota_bytes IS NOT NULL AND v_used>=v_term.quota_bytes THEN
        v_quota_depleted:=true;
        UPDATE pvnaive.service_terms
           SET state='quota_depleted',revision=revision+1,updated_at=p_observed_at
         WHERE id=v_term.id AND state='active';
        v_term.state:='quota_depleted';
    END IF;
    v_remaining:=CASE WHEN v_term.quota_bytes IS NULL THEN NULL ELSE GREATEST(v_term.quota_bytes-v_used-v_projection.reserved_bytes,0) END;

    PERFORM pvnaive.direct_naive_accounting_leave_context();
    RETURN QUERY SELECT v_term.id,true,true,false,v_reason,v_upload_delta,v_download_delta,
                        v_quota_depleted,v_remaining,v_term.first_connected_at,v_projection.accounting_complete;
END;
$$;

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
    accounting_complete boolean
)
LANGUAGE plpgsql
STABLE
SECURITY DEFINER
SET search_path = pg_catalog, pvnaive
AS $$
DECLARE
    v_term pvnaive.service_terms%ROWTYPE;
    v_projection pvnaive.direct_naive_accounting_terms%ROWTYPE;
    v_sessions bigint;
    v_has_stale boolean;
    v_used bigint;
BEGIN
    IF p_service_term_id IS NULL OR p_observed_at IS NULL OR p_stale_after_seconds<=0 THEN
        RAISE EXCEPTION 'invalid accounting read request' USING ERRCODE='22023';
    END IF;
    PERFORM pvnaive.direct_naive_accounting_enter_context();
    SELECT * INTO v_term FROM pvnaive.service_terms WHERE id=p_service_term_id;
    IF NOT FOUND THEN
        PERFORM pvnaive.direct_naive_accounting_leave_context();
        RETURN;
    END IF;
    SELECT * INTO v_projection FROM pvnaive.direct_naive_accounting_terms t WHERE t.service_term_id=p_service_term_id;
    IF NOT FOUND THEN
        v_projection.service_term_id:=p_service_term_id;
        v_projection.upload_bytes:=0;
        v_projection.download_bytes:=0;
        v_projection.reserved_bytes:=0;
        v_projection.accounting_complete:=true;
    END IF;
    SELECT count(*) FILTER (WHERE NOT s.final AND s.last_observed_at >= p_observed_at-(p_stale_after_seconds*interval '1 second')),
           COALESCE(bool_or(NOT s.final AND s.last_observed_at < p_observed_at-(p_stale_after_seconds*interval '1 second')),false)
      INTO v_sessions,v_has_stale
      FROM pvnaive.direct_naive_accounting_sessions s
     WHERE s.service_term_id=p_service_term_id;
    v_used:=v_projection.upload_bytes+v_projection.download_bytes;
    PERFORM pvnaive.direct_naive_accounting_leave_context();
    RETURN QUERY SELECT p_service_term_id,v_projection.upload_bytes,v_projection.download_bytes,v_used,
                        v_term.quota_bytes,
                        CASE WHEN v_term.quota_bytes IS NULL THEN NULL::bigint ELSE GREATEST(v_term.quota_bytes-v_used-v_projection.reserved_bytes,0) END,
                        CASE WHEN v_term.quota_bytes IS NULL THEN 'unlimited'::text WHEN v_used>=v_term.quota_bytes THEN 'depleted'::text ELSE 'active'::text END,
                        v_term.first_connected_at,v_projection.last_online,(v_sessions>0),v_sessions,
                        (v_projection.accounting_complete AND NOT v_has_stale);
END;
$$;

REVOKE ALL ON FUNCTION pvnaive.direct_naive_accounting_authorize(uuid,timestamptz) FROM PUBLIC;
REVOKE ALL ON FUNCTION pvnaive.direct_naive_accounting_claim(uuid,text,uuid,uuid,bigint,text,bigint,timestamptz) FROM PUBLIC;
REVOKE ALL ON FUNCTION pvnaive.direct_naive_accounting_ingest(uuid,text,text,uuid,uuid,bigint,timestamptz,boolean,bigint,bigint,boolean) FROM PUBLIC;
REVOKE ALL ON FUNCTION pvnaive.direct_naive_accounting_read(uuid,timestamptz,bigint) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION pvnaive.direct_naive_accounting_authorize(uuid,timestamptz) TO pvnaive_app;
GRANT EXECUTE ON FUNCTION pvnaive.direct_naive_accounting_claim(uuid,text,uuid,uuid,bigint,text,bigint,timestamptz) TO pvnaive_app;
GRANT EXECUTE ON FUNCTION pvnaive.direct_naive_accounting_ingest(uuid,text,text,uuid,uuid,bigint,timestamptz,boolean,bigint,bigint,boolean) TO pvnaive_app;
GRANT EXECUTE ON FUNCTION pvnaive.direct_naive_accounting_read(uuid,timestamptz,bigint) TO pvnaive_app;

COMMENT ON TABLE pvnaive.direct_naive_accounting_events IS
    'Append-only exact successful-payload ledger for authenticated Direct Naive CONNECT sessions.';
COMMENT ON TABLE pvnaive.direct_naive_accounting_claims IS
    'Persistent pre-write reservations preventing concurrent sessions from double-spending one ServiceTerm quota.';
COMMENT ON COLUMN pvnaive.direct_naive_accounting_terms.accounting_complete IS
    'False means exact final usage cannot be claimed; unknown bytes are never guessed into the ledger.';
