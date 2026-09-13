-- pvnaive:migration-version 0026
-- pvnaive:migration-name runtime_log_events
-- pvnaive:transactional true
-- pvnaive:destructive false

-- Read-only SECURITY DEFINER view over the newest exact-accounting ledger
-- events, for the runtime logs viewer. The direct tables are deliberately not
-- granted to the app role; this is the narrow read surface.
CREATE OR REPLACE FUNCTION pvnaive.runtime_log_events(
    p_limit integer DEFAULT 200
)
RETURNS TABLE (
    observed_at timestamptz,
    service_term_id text,
    username_diagnostic text,
    node_id text,
    authenticated_connect boolean,
    upload_cumulative bigint,
    download_cumulative bigint,
    final boolean
)
LANGUAGE sql
STABLE
SECURITY DEFINER
SET search_path = pg_catalog, pvnaive
AS $$
    SELECT e.observed_at,
           e.service_term_id::text,
           e.username_diagnostic,
           e.node_id,
           e.authenticated_connect,
           e.upload_cumulative,
           e.download_cumulative,
           e.final
      FROM pvnaive.direct_naive_accounting_events e
     ORDER BY e.ledger_sequence DESC
     LIMIT GREATEST(1, LEAST(GREATEST(p_limit, 1), 500));
$$;

REVOKE ALL ON FUNCTION pvnaive.runtime_log_events(integer) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION pvnaive.runtime_log_events(integer) TO pvnaive_app;
