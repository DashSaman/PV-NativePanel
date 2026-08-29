-- pvnaive:migration-version 0010
-- Source: PVNaive pending reservation accounting-completeness hardening
-- pvnaive:migration-name pending_reservation_completeness
-- pvnaive:transactional true
-- pvnaive:destructive false

SET LOCAL lock_timeout = '10s';
SET LOCAL statement_timeout = '120s';
SET LOCAL idle_in_transaction_session_timeout = '60s';
SET LOCAL ROLE pvnaive_owner;

ALTER FUNCTION pvnaive.direct_naive_accounting_authorize(uuid,timestamptz)
    RENAME TO direct_naive_accounting_authorize_v9;
REVOKE ALL ON FUNCTION pvnaive.direct_naive_accounting_authorize_v9(uuid,timestamptz)
    FROM PUBLIC, pvnaive_app;

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
LANGUAGE sql
VOLATILE
SECURITY DEFINER
SET search_path = pg_catalog, pvnaive
AS $$
    SELECT a.service_term_id,
           a.tracked,
           a.allowed,
           a.reason,
           a.quota_bytes,
           a.used_bytes,
           a.reserved_bytes,
           a.remaining_bytes,
           a.expires_at,
           a.first_connected_at,
           (a.accounting_complete AND COALESCE(a.reserved_bytes, 0) = 0)
      FROM pvnaive.direct_naive_accounting_authorize_v9(
               p_runtime_credential_id,
               p_observed_at
           ) AS a;
$$;

ALTER FUNCTION pvnaive.direct_naive_accounting_read(uuid,timestamptz,bigint)
    RENAME TO direct_naive_accounting_read_v9;
REVOKE ALL ON FUNCTION pvnaive.direct_naive_accounting_read_v9(uuid,timestamptz,bigint)
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
    accounting_complete boolean
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
               r.accounting_complete
               AND COALESCE(
                   (
                       SELECT t.reserved_bytes = 0
                         FROM pvnaive.direct_naive_accounting_terms AS t
                        WHERE t.service_term_id = r.service_term_id
                   ),
                   true
               )
           )
      FROM pvnaive.direct_naive_accounting_read_v9(
               p_service_term_id,
               p_observed_at,
               p_stale_after_seconds
           ) AS r;
$$;

REVOKE ALL ON FUNCTION pvnaive.direct_naive_accounting_authorize(uuid,timestamptz) FROM PUBLIC;
REVOKE ALL ON FUNCTION pvnaive.direct_naive_accounting_read(uuid,timestamptz,bigint) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION pvnaive.direct_naive_accounting_authorize(uuid,timestamptz) TO pvnaive_app;
GRANT EXECUTE ON FUNCTION pvnaive.direct_naive_accounting_read(uuid,timestamptz,bigint) TO pvnaive_app;
