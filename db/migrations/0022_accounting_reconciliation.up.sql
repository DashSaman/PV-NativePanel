-- pvnaive:migration-version 0022
-- pvnaive:migration-name accounting_reconciliation
-- pvnaive:transactional true
-- pvnaive:destructive false

SET LOCAL lock_timeout = '10s';
SET LOCAL statement_timeout = '120s';
SET LOCAL idle_in_transaction_session_timeout = '60s';
SET LOCAL ROLE pvnaive_owner;

-- Claims can remain unsettled forever when a Caddy/telemetry failure happens
-- between the quota reservation (claim) and the settle event. The reservation
-- then permanently understates remaining quota and blocks resets. The release
-- path marks those claims terminal WITHOUT guessing any bytes into the ledger:
-- released bytes are never counted as usage; they only stop blocking quota.
ALTER TABLE pvnaive.direct_naive_accounting_claims
    ADD COLUMN released_bytes bigint,
    ADD COLUMN released_at timestamptz,
    ADD COLUMN released_reason text,
    ADD CONSTRAINT direct_naive_accounting_claims_release_shape CHECK (
        (released_at IS NULL) = (released_bytes IS NULL)
        AND (
            released_reason IS NULL
            OR (
                released_at IS NOT NULL
                AND released_reason IN ('session_final', 'session_stale', 'manual_reconciliation')
            )
        )
    );

CREATE INDEX direct_naive_accounting_claims_unsettled_stale_idx
    ON pvnaive.direct_naive_accounting_claims (claimed_at)
    WHERE settled_at IS NULL;

-- Replaces the 0009 read that folded idle-session staleness into
-- accounting_complete. Idle sessions keep exact counters; they must not make
-- counted usage disappear from the UI. accounting_complete now reflects only
-- real ledger anomalies (term flag) plus unresolved reservations.
ALTER FUNCTION pvnaive.direct_naive_accounting_read(uuid,timestamptz,bigint)
    RENAME TO direct_naive_accounting_read_v21;
REVOKE ALL ON FUNCTION pvnaive.direct_naive_accounting_read_v21(uuid,timestamptz,bigint)
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
           (
               COALESCE(t.accounting_complete, true)
               AND COALESCE(t.reserved_bytes, 0) = 0
           ),
           t.last_reset_at
      FROM pvnaive.direct_naive_accounting_read_v21(
               p_service_term_id,
               p_observed_at,
               p_stale_after_seconds
           ) AS r
      LEFT JOIN pvnaive.direct_naive_accounting_terms AS t
        ON t.service_term_id = r.service_term_id;
$$;

-- Release unsettled claims whose session can no longer produce a settle event:
--  * session already final, or
--  * session idle beyond the stale window (its telemetry is gone).
-- Safe against live sessions: an active session keeps last_observed_at fresh,
-- so its in-flight claims are never released.
CREATE OR REPLACE FUNCTION pvnaive.release_stale_accounting_reservations(
    p_stale_seconds bigint DEFAULT 900,
    p_observed_at timestamptz DEFAULT clock_timestamp(),
    p_limit bigint DEFAULT 500
)
RETURNS TABLE (
    released_claims bigint,
    released_bytes bigint
)
LANGUAGE plpgsql
VOLATILE
SECURITY DEFINER
SET search_path = pg_catalog, pvnaive
AS $$
DECLARE
    v_stale timestamptz;
    v_claim RECORD;
    v_total_claims bigint := 0;
    v_total_bytes bigint := 0;
    v_reason text;
BEGIN
    IF p_stale_seconds IS NULL OR p_stale_seconds < 60 THEN
        RAISE EXCEPTION 'release_stale_accounting_reservations requires stale window >= 60 seconds' USING ERRCODE = '22023';
    END IF;
    IF p_limit IS NULL OR p_limit < 1 THEN
        RAISE EXCEPTION 'release_stale_accounting_reservations requires limit >= 1' USING ERRCODE = '22023';
    END IF;
    v_stale := p_observed_at - (p_stale_seconds * interval '1 second');

    FOR v_claim IN
        SELECT c.id, c.service_term_id, c.reserved_bytes, s.final, s.last_observed_at
          FROM pvnaive.direct_naive_accounting_claims c
          JOIN pvnaive.direct_naive_accounting_sessions s
            ON s.runtime_credential_id = c.runtime_credential_id
           AND s.node_id = c.node_id
           AND s.boot_id = c.boot_id
           AND s.session_id = c.session_id
         WHERE c.settled_at IS NULL
           AND c.released_at IS NULL
           AND (s.final OR s.last_observed_at < v_stale)
         ORDER BY c.claimed_at
         FOR UPDATE OF c, s SKIP LOCKED
         LIMIT p_limit
    LOOP
        IF v_claim.final THEN
            v_reason := 'session_final';
        ELSE
            v_reason := 'session_stale';
        END IF;
        UPDATE pvnaive.direct_naive_accounting_claims
           SET released_bytes = v_claim.reserved_bytes,
               released_at = p_observed_at,
               released_reason = v_reason,
               settled_bytes = 0,
               settled_at = p_observed_at
         WHERE id = v_claim.id;
        UPDATE pvnaive.direct_naive_accounting_terms
           SET reserved_bytes = GREATEST(reserved_bytes - v_claim.reserved_bytes, 0),
               updated_at = p_observed_at
         WHERE service_term_id = v_claim.service_term_id;
        v_total_claims := v_total_claims + 1;
        v_total_bytes := v_total_bytes + v_claim.reserved_bytes;
    END LOOP;

    RETURN QUERY SELECT v_total_claims, v_total_bytes;
END;
$$;

-- Evidence-based completeness verification for one ServiceTerm. Returns the
-- exact blocking reason when the term cannot be marked complete again.
-- Checks (all must hold):
--   1. no unresolved claims (settled or released) on the term;
--   2. term counters equal the sum of per-session cumulative counters;
--   3. every session's stored counters equal its last ledger event.
CREATE OR REPLACE FUNCTION pvnaive.verify_accounting_completeness(
    p_service_term_id uuid
)
RETURNS TABLE (
    ok boolean,
    blocking_reason text,
    detail text
)
LANGUAGE plpgsql
STABLE
SECURITY DEFINER
SET search_path = pg_catalog, pvnaive
AS $$
DECLARE
    v_projection pvnaive.direct_naive_accounting_terms%ROWTYPE;
    v_session_upload bigint;
    v_session_download bigint;
    v_unresolved bigint;
    v_unresolved_bytes bigint;
    v_tail_session uuid;
    v_tail_upload bigint;
    v_tail_download bigint;
    v_session_upload_at_tail bigint;
    v_session_download_at_tail bigint;
BEGIN
    IF p_service_term_id IS NULL THEN
        RAISE EXCEPTION 'verify_accounting_completeness requires a service term' USING ERRCODE = '22023';
    END IF;

    SELECT * INTO v_projection
      FROM pvnaive.direct_naive_accounting_terms t
     WHERE t.service_term_id = p_service_term_id;
    IF NOT FOUND THEN
        RETURN QUERY SELECT true, 'no_projection'::text, 'term has no accounting rows'::text;
        RETURN;
    END IF;

    SELECT count(*), COALESCE(sum(c.reserved_bytes), 0)
      INTO v_unresolved, v_unresolved_bytes
      FROM pvnaive.direct_naive_accounting_claims c
     WHERE c.service_term_id = p_service_term_id
       AND c.settled_at IS NULL;
    IF v_unresolved > 0 THEN
        RETURN QUERY SELECT false, 'unsettled_claims'::text,
            format('count=%s reserved_bytes=%s; run reservation release first', v_unresolved, v_unresolved_bytes)::text;
        RETURN;
    END IF;

    SELECT COALESCE(sum(s.upload_cumulative), 0), COALESCE(sum(s.download_cumulative), 0)
      INTO v_session_upload, v_session_download
      FROM pvnaive.direct_naive_accounting_sessions s
     WHERE s.service_term_id = p_service_term_id;
    IF v_projection.upload_bytes <> v_session_upload OR v_projection.download_bytes <> v_session_download THEN
        RETURN QUERY SELECT false, 'counter_mismatch'::text,
            format('term=(up=%s down=%s) sessions_sum=(up=%s down=%s)',
                   v_projection.upload_bytes, v_projection.download_bytes,
                   v_session_upload, v_session_download)::text;
        RETURN;
    END IF;

    -- Every session's stored counters must equal the cumulative counters of
    -- its newest ledger event (the ledger tail proves the session row).
    WITH tails AS (
        SELECT DISTINCT ON (e.runtime_credential_id, e.node_id, e.boot_id, e.session_id)
               e.runtime_credential_id, e.node_id, e.boot_id, e.session_id,
               e.upload_cumulative, e.download_cumulative
          FROM pvnaive.direct_naive_accounting_events e
         WHERE e.service_term_id = p_service_term_id
         ORDER BY e.runtime_credential_id, e.node_id, e.boot_id, e.session_id,
                  e.source_sequence DESC
    )
    SELECT s.session_id, s.upload_cumulative, s.download_cumulative,
           tails.upload_cumulative, tails.download_cumulative
      INTO v_tail_session, v_session_upload_at_tail, v_session_download_at_tail,
           v_tail_upload, v_tail_download
      FROM pvnaive.direct_naive_accounting_sessions s
      LEFT JOIN tails
        ON tails.runtime_credential_id = s.runtime_credential_id
       AND tails.node_id = s.node_id
       AND tails.boot_id = s.boot_id
       AND tails.session_id = s.session_id
     WHERE s.service_term_id = p_service_term_id
       AND (tails.upload_cumulative IS DISTINCT FROM s.upload_cumulative
            OR tails.download_cumulative IS DISTINCT FROM s.download_cumulative)
     LIMIT 1;
    IF FOUND THEN
        RETURN QUERY SELECT false, 'ledger_tail_mismatch'::text,
            format('session=%s session=(up=%s down=%s) ledger_tail=(up=%s down=%s)',
                   v_tail_session, v_session_upload_at_tail, v_session_download_at_tail,
                   v_tail_upload, v_tail_download)::text;
        RETURN;
    END IF;

    RETURN QUERY SELECT true, 'verified'::text, 'term counters consistent with session and ledger evidence'::text;
END;
$$;

-- Repair: re-check the evidence and, when it holds, restore
-- accounting_complete=true on the term and its sessions. Never invents usage.
CREATE OR REPLACE FUNCTION pvnaive.repair_accounting_completeness(
    p_service_term_id uuid
)
RETURNS TABLE (
    repaired boolean,
    ok boolean,
    blocking_reason text,
    detail text
)
LANGUAGE plpgsql
VOLATILE
SECURITY DEFINER
SET search_path = pg_catalog, pvnaive
AS $$
DECLARE
    v_verify RECORD;
    v_already boolean;
BEGIN
    IF p_service_term_id IS NULL THEN
        RAISE EXCEPTION 'repair_accounting_completeness requires a service term' USING ERRCODE = '22023';
    END IF;
    SELECT accounting_complete INTO v_already
      FROM pvnaive.direct_naive_accounting_terms
     WHERE service_term_id = p_service_term_id
     FOR UPDATE;
    IF NOT FOUND THEN
        RETURN QUERY SELECT false, true, 'no_projection'::text, 'term has no accounting rows'::text;
        RETURN;
    END IF;

    SELECT v.ok, v.blocking_reason, v.detail
      INTO v_verify
      FROM pvnaive.verify_accounting_completeness(p_service_term_id) v;

    IF NOT v_verify.ok THEN
        RETURN QUERY SELECT false, false, v_verify.blocking_reason, v_verify.detail;
        RETURN;
    END IF;
    IF v_already THEN
        RETURN QUERY SELECT false, true, 'already_complete'::text, 'term already marked complete'::text;
        RETURN;
    END IF;

    UPDATE pvnaive.direct_naive_accounting_terms
       SET accounting_complete = true,
           updated_at = clock_timestamp()
     WHERE service_term_id = p_service_term_id;
    UPDATE pvnaive.direct_naive_accounting_sessions
       SET accounting_complete = true,
           updated_at = clock_timestamp()
     WHERE service_term_id = p_service_term_id;

    RETURN QUERY SELECT true, true, 'repaired'::text, 'accounting completeness restored after evidence verification'::text;
END;
$$;

-- Fleet-level summary for the admin dashboard (exact counters only).
CREATE OR REPLACE FUNCTION pvnaive.usage_summary()
RETURNS TABLE (
    terms_total bigint,
    terms_active bigint,
    terms_depleted bigint,
    terms_suspended bigint,
    terms_incomplete bigint,
    users_online bigint,
    sessions_active bigint,
    upload_bytes bigint,
    download_bytes bigint,
    used_bytes bigint,
    reserved_bytes bigint,
    quota_bytes_total bigint
)
LANGUAGE sql
STABLE
SECURITY DEFINER
SET search_path = pg_catalog, pvnaive
AS $$
    SELECT
        count(*)::bigint,
        count(*) FILTER (WHERE st.state = 'active')::bigint,
        count(*) FILTER (WHERE st.state = 'quota_depleted')::bigint,
        count(*) FILTER (WHERE st.state = 'suspended')::bigint,
        count(*) FILTER (WHERE t.accounting_complete = false)::bigint,
        count(*) FILTER (WHERE t.last_online IS NOT NULL AND t.last_online >= clock_timestamp() - interval '90 seconds')::bigint,
        (
            SELECT count(*) FROM pvnaive.direct_naive_accounting_sessions s
             WHERE s.final = false
               AND s.last_observed_at >= clock_timestamp() - interval '90 seconds'
        )::bigint,
        COALESCE(sum(t.upload_bytes), 0)::bigint,
        COALESCE(sum(t.download_bytes), 0)::bigint,
        COALESCE(sum(t.upload_bytes + t.download_bytes), 0)::bigint,
        COALESCE(sum(t.reserved_bytes), 0)::bigint,
        COALESCE(sum(st.quota_bytes), 0)::bigint
      FROM pvnaive.direct_naive_accounting_terms t
      JOIN pvnaive.service_terms st ON st.id = t.service_term_id;
$$;

-- Exact daily usage from the immutable ledger. p_service_term_id NULL means
-- all terms (fleet chart). p_days bounds the window (1..366).
CREATE OR REPLACE FUNCTION pvnaive.usage_history(
    p_service_term_id uuid DEFAULT NULL,
    p_days integer DEFAULT 14
)
RETURNS TABLE (
    day date,
    upload_bytes bigint,
    download_bytes bigint
)
LANGUAGE sql
STABLE
SECURITY DEFINER
SET search_path = pg_catalog, pvnaive
AS $$
    SELECT e.observed_at::date AS day,
           COALESCE(sum(e.upload_delta), 0)::bigint AS upload_bytes,
           COALESCE(sum(e.download_delta), 0)::bigint AS download_bytes
      FROM pvnaive.direct_naive_accounting_events e
     WHERE e.observed_at >= (current_date - (GREATEST(LEAST(p_days, 366), 1))::int)::date
       AND (p_service_term_id IS NULL OR e.service_term_id = p_service_term_id)
     GROUP BY e.observed_at::date
     ORDER BY day;
$$;

-- Terms currently flagged incomplete, for the reconciliation report.
CREATE OR REPLACE FUNCTION pvnaive.list_incomplete_accounting_terms()
RETURNS TABLE (
    service_term_id uuid
)
LANGUAGE sql
STABLE
SECURITY DEFINER
SET search_path = pg_catalog, pvnaive
AS $$
    SELECT t.service_term_id
      FROM pvnaive.direct_naive_accounting_terms t
     WHERE t.accounting_complete = false
     ORDER BY t.updated_at;
$$;

-- Current ServiceTerm for one user (newest purchased term), for the
-- per-user usage endpoint.
CREATE OR REPLACE FUNCTION pvnaive.customer_current_usage_term(
    p_user_id uuid
)
RETURNS TABLE (
    service_term_id uuid,
    quota_bytes bigint
)
LANGUAGE sql
STABLE
SECURITY DEFINER
SET search_path = pg_catalog, pvnaive
AS $$
    SELECT st.id, st.quota_bytes
      FROM pvnaive.service_terms st
     WHERE st.user_id = p_user_id
     ORDER BY st.purchased_at DESC, st.created_at DESC
     LIMIT 1;
$$;

REVOKE ALL ON FUNCTION pvnaive.direct_naive_accounting_read(uuid,timestamptz,bigint) FROM PUBLIC;
REVOKE ALL ON FUNCTION pvnaive.release_stale_accounting_reservations(bigint,timestamptz,bigint) FROM PUBLIC;
REVOKE ALL ON FUNCTION pvnaive.verify_accounting_completeness(uuid) FROM PUBLIC;
REVOKE ALL ON FUNCTION pvnaive.repair_accounting_completeness(uuid) FROM PUBLIC;
REVOKE ALL ON FUNCTION pvnaive.usage_summary() FROM PUBLIC;
REVOKE ALL ON FUNCTION pvnaive.usage_history(uuid,integer) FROM PUBLIC;
REVOKE ALL ON FUNCTION pvnaive.list_incomplete_accounting_terms() FROM PUBLIC;
REVOKE ALL ON FUNCTION pvnaive.customer_current_usage_term(uuid) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION pvnaive.direct_naive_accounting_read(uuid,timestamptz,bigint) TO pvnaive_app;
GRANT EXECUTE ON FUNCTION pvnaive.release_stale_accounting_reservations(bigint,timestamptz,bigint) TO pvnaive_app;
GRANT EXECUTE ON FUNCTION pvnaive.verify_accounting_completeness(uuid) TO pvnaive_app;
GRANT EXECUTE ON FUNCTION pvnaive.repair_accounting_completeness(uuid) TO pvnaive_app;
GRANT EXECUTE ON FUNCTION pvnaive.usage_summary() TO pvnaive_app;
GRANT EXECUTE ON FUNCTION pvnaive.usage_history(uuid,integer) TO pvnaive_app;
GRANT EXECUTE ON FUNCTION pvnaive.list_incomplete_accounting_terms() TO pvnaive_app;
GRANT EXECUTE ON FUNCTION pvnaive.customer_current_usage_term(uuid) TO pvnaive_app;

COMMENT ON COLUMN pvnaive.direct_naive_accounting_claims.released_bytes IS
    'Reservation bytes returned to quota without being counted; released bytes are never guessed into the ledger.';
COMMENT ON FUNCTION pvnaive.release_stale_accounting_reservations(bigint,timestamptz,bigint) IS
    'Reaper for reservations whose session can no longer settle; releases quota without inventing usage.';
COMMENT ON FUNCTION pvnaive.verify_accounting_completeness(uuid) IS
    'Evidence check: unresolved claims, term-vs-session counter equality, ledger-tail equality.';
COMMENT ON FUNCTION pvnaive.repair_accounting_completeness(uuid) IS
    'Restores accounting_complete only after verify_accounting_completeness proves consistency.';
COMMENT ON FUNCTION pvnaive.usage_summary() IS
    'Fleet-level exact accounting totals for the admin dashboard.';
COMMENT ON FUNCTION pvnaive.usage_history(uuid,integer) IS
    'Exact daily upload/download from the immutable accounting ledger (UTC day buckets).';
