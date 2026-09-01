-- PVNaive unique IP limit enforcement
-- pvnaive:migration-version 0020
-- pvnaive:migration-name unique_ip_limit
-- pvnaive:transactional true
-- pvnaive:destructive false

ALTER TABLE pvnaive.service_terms
    ADD COLUMN unique_ip_limit integer
    CHECK (unique_ip_limit IS NULL OR unique_ip_limit > 0);

ALTER TABLE pvnaive.plans
    ADD COLUMN unique_ip_limit integer
    CHECK (unique_ip_limit IS NULL OR unique_ip_limit > 0);

-- Backfill existing plan-backed terms from plan policy.
UPDATE pvnaive.service_terms st
   SET unique_ip_limit = p.unique_ip_limit
  FROM pvnaive.plans p
 WHERE st.plan_id = p.id
   AND p.unique_ip_limit IS NOT NULL;

COMMENT ON COLUMN pvnaive.service_terms.unique_ip_limit IS
    'Immutable-per-term style snapshot of the maximum simultaneous unique client IPs from trusted Caddy RemoteAddr; NULL means Unlimited.';

-- Keep the schema19 implementation as a private primitive.
ALTER FUNCTION pvnaive.direct_naive_accounting_ingest(
    uuid,text,text,uuid,uuid,bigint,timestamptz,boolean,bigint,bigint,boolean
) RENAME TO direct_naive_accounting_ingest_v19;

REVOKE ALL ON FUNCTION pvnaive.direct_naive_accounting_ingest_v19(
    uuid,text,text,uuid,uuid,bigint,timestamptz,boolean,bigint,bigint,boolean
) FROM PUBLIC, pvnaive_app;

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
    p_final boolean DEFAULT false,
    p_client_ip text DEFAULT ''
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
    v_projection pvnaive.direct_naive_accounting_terms%ROWTYPE;
    v_active_count bigint := 0;
    v_ip_count bigint := 0;
    v_used bigint := 0;
    v_remaining bigint;
    v_trusted_ip text;
    v_result RECORD;
    v_peer_ip inet;
BEGIN
    -- Preserve every pre-schema20 validation/conflict/finalization path by
    -- delegating events that cannot represent a brand-new successful open.
    IF p_runtime_credential_id IS NULL OR p_boot_id IS NULL OR p_session_id IS NULL
       OR p_source_sequence IS DISTINCT FROM 1
       OR p_observed_at IS NULL
       OR p_authenticated_connect IS DISTINCT FROM true
       OR p_upload_cumulative IS DISTINCT FROM 0
       OR p_download_cumulative IS DISTINCT FROM 0
       OR p_final IS DISTINCT FROM false THEN
        RETURN QUERY
        SELECT * FROM pvnaive.direct_naive_accounting_ingest_v19(
            p_runtime_credential_id,p_username_diagnostic,p_node_id,p_boot_id,p_session_id,
            p_source_sequence,p_observed_at,p_authenticated_connect,p_upload_cumulative,
            p_download_cumulative,p_final
        );
        RETURN;
    END IF;

    PERFORM pvnaive.direct_naive_accounting_enter_context();

    -- Same exact session identity must keep the schema19 duplicate/conflict
    -- semantics even when the term is already at its IP limit.
    IF EXISTS (
        SELECT 1
          FROM pvnaive.direct_naive_accounting_sessions s
         WHERE s.runtime_credential_id = p_runtime_credential_id
           AND s.node_id = btrim(p_node_id)
           AND s.boot_id = p_boot_id
           AND s.session_id = p_session_id
    ) THEN
        PERFORM pvnaive.direct_naive_accounting_leave_context();
        RETURN QUERY
        SELECT * FROM pvnaive.direct_naive_accounting_ingest_v19(
            p_runtime_credential_id,p_username_diagnostic,p_node_id,p_boot_id,p_session_id,
            p_source_sequence,p_observed_at,p_authenticated_connect,p_upload_cumulative,
            p_download_cumulative,p_final
        );
        RETURN;
    END IF;

    -- This row lock is the atomic concurrency boundary. Competing opens for the
    -- same ServiceTerm serialize here before either can create a session row.
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
       AND urc.unbound_at IS NULL
       AND urc.role='primary'
     FOR UPDATE OF st
     LIMIT 1;

    -- No term found or no IP limit: delegate to schema19 unchanged.
    IF NOT FOUND OR v_term.unique_ip_limit IS NULL THEN
        PERFORM pvnaive.direct_naive_accounting_leave_context();
        RETURN QUERY
        SELECT * FROM pvnaive.direct_naive_accounting_ingest_v19(
            p_runtime_credential_id,p_username_diagnostic,p_node_id,p_boot_id,p_session_id,
            p_source_sequence,p_observed_at,p_authenticated_connect,p_upload_cumulative,
            p_download_cumulative,p_final
        );
        RETURN;
    END IF;

    -- Only trusted Caddy RemoteAddr peer metadata counts.
    -- Never Forwarded or X-Forwarded-For. Empty means no peer was recorded.
    v_trusted_ip := btrim(COALESCE(NULLIF(p_client_ip, ''), ''));

    -- Fail closed: a non-empty p_client_ip is mandatory for unique-IP
    -- enforcement. An empty value means the Caddy overlay did not supply
    -- a parseable RemoteAddr, so we must not guess or default to 'unknown'.
    IF v_trusted_ip = '' THEN
        PERFORM pvnaive.direct_naive_accounting_leave_context();
        RETURN QUERY SELECT
            v_term.id,true,false,false,'missing_trusted_peer'::text,
            0::bigint,0::bigint,false,NULL::bigint,
            v_term.first_connected_at,true;
        RETURN;
    END IF;

    -- Validate/canonicalize client_ip early before any ::inet cast.  A
    -- malformed IP must produce a clean rejection, not a raw DB error.
    IF v_trusted_ip !~ '^([0-9]{1,3}\.){3}[0-9]{1,3}$'
       AND v_trusted_ip !~ '^([0-9a-fA-F]{0,4}:){2,7}[0-9a-fA-F]{0,4}$' THEN
        PERFORM pvnaive.direct_naive_accounting_leave_context();
        RETURN QUERY SELECT
            v_term.id,true,false,false,'invalid_client_ip'::text,
            0::bigint,0::bigint,false,NULL::bigint,
            v_term.first_connected_at,true;
        RETURN;
    END IF;

    -- Fail closed if any qualifying active session lacks trusted peer
    -- evidence. We never invent an IP for an un-peer'd session.
    IF EXISTS (
        SELECT 1
          FROM pvnaive.direct_naive_accounting_sessions s
         WHERE s.service_term_id = v_term.id
           AND s.final = false
           AND s.accounting_complete = true
           AND s.last_observed_at >= p_observed_at - interval '90 seconds'
           AND NOT EXISTS (
               SELECT 1
                 FROM pvnaive.direct_naive_accounting_session_peers p
                WHERE p.runtime_credential_id = s.runtime_credential_id
                  AND p.node_id = s.node_id
                  AND p.boot_id = s.boot_id
                  AND p.session_id = s.session_id
           )
    ) THEN
        PERFORM pvnaive.direct_naive_accounting_leave_context();
        RETURN QUERY SELECT
            v_term.id,true,false,false,'active_session_missing_peer'::text,
            0::bigint,0::bigint,false,NULL::bigint,
            v_term.first_connected_at,true;
        RETURN;
    END IF;

    -- Count distinct active IPs for this term including the new one.
    -- JOIN to canonical session_peers; same IP multi-session counts once.
    SELECT count(*)
      INTO v_ip_count
      FROM (
          SELECT DISTINCT host(p.client_ip) AS client_ip
            FROM pvnaive.direct_naive_accounting_sessions s
            JOIN pvnaive.direct_naive_accounting_session_peers p
              ON p.runtime_credential_id = s.runtime_credential_id
             AND p.node_id = s.node_id
             AND p.boot_id = s.boot_id
             AND p.session_id = s.session_id
           WHERE s.service_term_id = v_term.id
             AND s.final = false
             AND s.accounting_complete = true
             AND s.last_observed_at >= p_observed_at - interval '90 seconds'
          UNION
          SELECT v_trusted_ip
      ) distinct_ips;

    IF v_ip_count > v_term.unique_ip_limit THEN
        INSERT INTO pvnaive.direct_naive_accounting_terms(service_term_id)
        VALUES(v_term.id)
        ON CONFLICT ON CONSTRAINT direct_naive_accounting_terms_pkey DO NOTHING;

        SELECT * INTO v_projection
          FROM pvnaive.direct_naive_accounting_terms t
         WHERE t.service_term_id=v_term.id;
        v_used := v_projection.upload_bytes + v_projection.download_bytes;
        v_remaining := CASE
            WHEN v_term.quota_bytes IS NULL THEN NULL
            ELSE GREATEST(v_term.quota_bytes-v_used-v_projection.reserved_bytes,0)
        END;
        PERFORM pvnaive.direct_naive_accounting_leave_context();
        RETURN QUERY SELECT
            v_term.id,true,false,false,'unique_ip_limit'::text,
            0::bigint,0::bigint,
            (v_term.quota_bytes IS NOT NULL AND v_used>=v_term.quota_bytes),
            v_remaining,v_term.first_connected_at,v_projection.accounting_complete;
        RETURN;
    END IF;

    -- The ServiceTerm lock remains held for this statement/transaction while
    -- schema19 performs the actual append-only open and first-CONNECT logic.
    -- Capture the result instead of RETURN QUERY so we can inspect it before
    -- deciding whether to record the trusted peer.  RETURN QUERY does not
    -- terminate PL/pgSQL: if schema19 rejects (e.g. concurrent_session_limit)
    -- we must NOT insert a peer row for a session that was never created.
    SELECT * INTO v_result
    FROM pvnaive.direct_naive_accounting_ingest_v19(
        p_runtime_credential_id,p_username_diagnostic,p_node_id,p_boot_id,p_session_id,
        p_source_sequence,p_observed_at,p_authenticated_connect,p_upload_cumulative,
        p_download_cumulative,p_final
    );

    -- v_trusted_ip was already validated for inet cast above. Safe to convert.
    v_peer_ip := v_trusted_ip::inet;

    -- Record the trusted peer ONLY when schema19 actually accepted a new
    -- first-open session.  The four conditions together guarantee:
    --   tracked=true  → a service term was resolved
    --   accepted=true → the session was actually created (not rejected)
    --   duplicate=false → this is not a replay of an existing session
    --   v_peer_ip IS NOT NULL → the IP validated to a canonical inet
    -- If schema19 rejects, we return its exact rejection with zero peer
    -- mutation and zero exception.
    IF v_result.tracked AND v_result.accepted
       AND NOT v_result.duplicate AND v_peer_ip IS NOT NULL THEN
        INSERT INTO pvnaive.direct_naive_accounting_session_peers(
            runtime_credential_id,node_id,boot_id,session_id,service_term_id,
            tenant_id,user_id,client_ip,first_recorded_at,updated_at
        )
        SELECT p_runtime_credential_id,btrim(p_node_id),p_boot_id,p_session_id,
               v_term.id,v_term.tenant_id,v_term.user_id,
               v_peer_ip,p_observed_at,p_observed_at
        ON CONFLICT (runtime_credential_id,node_id,boot_id,session_id) DO UPDATE
          SET updated_at = GREATEST(
              pvnaive.direct_naive_accounting_session_peers.updated_at,
              EXCLUDED.updated_at
          );
    END IF;

    -- Return exactly the schema19 result (accepted or rejected), unmodified.
    RETURN QUERY
    SELECT v_result.service_term_id, v_result.tracked, v_result.accepted,
           v_result.duplicate, v_result.reason, v_result.upload_delta,
           v_result.download_delta, v_result.quota_depleted,
           v_result.remaining_bytes, v_result.first_connected_at,
           v_result.accounting_complete;
END;
$$;

REVOKE ALL ON FUNCTION pvnaive.direct_naive_accounting_ingest(
    uuid,text,text,uuid,uuid,bigint,timestamptz,boolean,bigint,bigint,boolean,text
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION pvnaive.direct_naive_accounting_ingest(
    uuid,text,text,uuid,uuid,bigint,timestamptz,boolean,bigint,bigint,boolean,text
) TO pvnaive_app;

COMMENT ON FUNCTION pvnaive.direct_naive_accounting_ingest(
    uuid,text,text,uuid,uuid,bigint,timestamptz,boolean,bigint,bigint,boolean,text
) IS 'Schema20 race-safe public ingest boundary enforcing ServiceTerm unique-IP limits from trusted Caddy RemoteAddr before first-CONNECT mutation.';
