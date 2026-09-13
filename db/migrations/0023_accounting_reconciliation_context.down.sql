-- pvnaive:migration-version 0023
-- pvnaive:migration-name accounting_reconciliation_context
-- pvnaive:transactional true
-- pvnaive:destructive true

SET LOCAL lock_timeout = '10s';
SET LOCAL statement_timeout = '120s';

SET LOCAL ROLE pvnaive_owner;

-- Restore the 0022 shapes (without the context calls). The 0022 functions are
-- re-created verbatim; rollback refuses only if release history exists.
DO $$
DECLARE
    v_unreleased bigint;
BEGIN
    SELECT count(*) INTO v_unreleased
      FROM pvnaive.direct_naive_accounting_claims
     WHERE released_at IS NOT NULL;
    IF v_unreleased > 0 THEN
        RAISE EXCEPTION 'schema23 rollback refused: % released claims exist; history would be lost', v_unreleased;
    END IF;
END;
$$;

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

DROP FUNCTION IF EXISTS pvnaive.release_stale_accounting_reservations(bigint,timestamptz,bigint);
DROP FUNCTION IF EXISTS pvnaive.verify_accounting_completeness(uuid);
DROP FUNCTION IF EXISTS pvnaive.repair_accounting_completeness(uuid);

DELETE FROM pvnaive.schema_migrations WHERE version = 23;
