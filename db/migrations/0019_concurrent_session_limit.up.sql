-- PVNaive concurrent session limit enforcement
-- pvnaive:migration-version 0019
-- pvnaive:migration-name concurrent_session_limit
-- pvnaive:transactional true
-- pvnaive:destructive false

ALTER TABLE pvnaive.service_terms
    ADD COLUMN concurrency_limit integer
    CHECK (concurrency_limit IS NULL OR concurrency_limit > 0);

-- Existing plan-backed terms inherit the plan policy once. NULL remains
-- explicit Unlimited for custom/adopted terms and plans without a limit.
UPDATE pvnaive.service_terms st
   SET concurrency_limit = p.concurrency_limit
  FROM pvnaive.plans p
 WHERE st.plan_id = p.id
   AND p.concurrency_limit IS NOT NULL;

COMMENT ON COLUMN pvnaive.service_terms.concurrency_limit IS
    'Immutable-per-term style snapshot of the maximum simultaneous non-final direct Naive sessions; NULL means Unlimited.';

-- Keep the schema17 implementation as a private primitive. The public wrapper
-- serializes new opens on the ServiceTerm row before delegating, which keeps
-- exact accounting/idempotency behavior unchanged while closing the race.
ALTER FUNCTION pvnaive.direct_naive_accounting_ingest(
    uuid,text,text,uuid,uuid,bigint,timestamptz,boolean,bigint,bigint,boolean
) RENAME TO direct_naive_accounting_ingest_v17;

REVOKE ALL ON FUNCTION pvnaive.direct_naive_accounting_ingest_v17(
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
    v_projection pvnaive.direct_naive_accounting_terms%ROWTYPE;
    v_active_count bigint := 0;
    v_used bigint := 0;
    v_remaining bigint;
BEGIN
    -- Preserve every pre-schema19 validation/conflict/finalization path by
    -- delegating events that cannot represent a brand-new successful open.
    IF p_runtime_credential_id IS NULL OR p_boot_id IS NULL OR p_session_id IS NULL
       OR p_source_sequence IS DISTINCT FROM 1
       OR p_observed_at IS NULL
       OR p_authenticated_connect IS DISTINCT FROM true
       OR p_upload_cumulative IS DISTINCT FROM 0
       OR p_download_cumulative IS DISTINCT FROM 0
       OR p_final IS DISTINCT FROM false THEN
        RETURN QUERY
        SELECT * FROM pvnaive.direct_naive_accounting_ingest_v17(
            p_runtime_credential_id,p_username_diagnostic,p_node_id,p_boot_id,p_session_id,
            p_source_sequence,p_observed_at,p_authenticated_connect,p_upload_cumulative,
            p_download_cumulative,p_final
        );
        RETURN;
    END IF;

    PERFORM pvnaive.direct_naive_accounting_enter_context();

    -- Same exact session identity must keep the schema17 duplicate/conflict
    -- semantics even when the term is already at its limit.
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
        SELECT * FROM pvnaive.direct_naive_accounting_ingest_v17(
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

    IF NOT FOUND OR v_term.concurrency_limit IS NULL THEN
        PERFORM pvnaive.direct_naive_accounting_leave_context();
        RETURN QUERY
        SELECT * FROM pvnaive.direct_naive_accounting_ingest_v17(
            p_runtime_credential_id,p_username_diagnostic,p_node_id,p_boot_id,p_session_id,
            p_source_sequence,p_observed_at,p_authenticated_connect,p_upload_cumulative,
            p_download_cumulative,p_final
        );
        RETURN;
    END IF;

    SELECT count(*)
      INTO v_active_count
      FROM pvnaive.direct_naive_accounting_sessions s
     WHERE s.service_term_id = v_term.id
       AND s.final = false
       AND s.accounting_complete = true
       AND s.last_observed_at >= p_observed_at - interval '90 seconds';

    IF v_active_count >= v_term.concurrency_limit THEN
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
            v_term.id,true,false,false,'concurrent_session_limit'::text,
            0::bigint,0::bigint,
            (v_term.quota_bytes IS NOT NULL AND v_used>=v_term.quota_bytes),
            v_remaining,v_term.first_connected_at,v_projection.accounting_complete;
        RETURN;
    END IF;

    -- The ServiceTerm lock remains held for this statement/transaction while
    -- schema17 performs the actual append-only open and first-CONNECT logic.
    RETURN QUERY
    SELECT * FROM pvnaive.direct_naive_accounting_ingest_v17(
        p_runtime_credential_id,p_username_diagnostic,p_node_id,p_boot_id,p_session_id,
        p_source_sequence,p_observed_at,p_authenticated_connect,p_upload_cumulative,
        p_download_cumulative,p_final
    );
END;
$$;

REVOKE ALL ON FUNCTION pvnaive.direct_naive_accounting_ingest(
    uuid,text,text,uuid,uuid,bigint,timestamptz,boolean,bigint,bigint,boolean
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION pvnaive.direct_naive_accounting_ingest(
    uuid,text,text,uuid,uuid,bigint,timestamptz,boolean,bigint,bigint,boolean
) TO pvnaive_app;

COMMENT ON FUNCTION pvnaive.direct_naive_accounting_ingest(
    uuid,text,text,uuid,uuid,bigint,timestamptz,boolean,bigint,bigint,boolean
) IS 'Schema19 race-safe public ingest boundary enforcing ServiceTerm concurrent-session limits before first-CONNECT mutation.';
