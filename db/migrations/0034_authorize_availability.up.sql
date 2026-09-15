-- pvnaive:migration-version 0034
-- pvnaive:migration-name authorize_availability
-- pvnaive:transactional true
-- pvnaive:destructive false

SET LOCAL lock_timeout = '10s';
SET LOCAL statement_timeout = '120s';
SET LOCAL idle_in_transaction_session_timeout = '60s';
SET LOCAL ROLE pvnaive_owner;

-- Availability fix for the CONNECT authorization gate. The previous wrapper
-- (0010) folded the sticky term-level anomaly flag into accounting_complete,
-- so one orphaned reservation or one recorded anomaly permanently blocked
-- every new CONNECT for the affected user: blocked clients cannot emit the
-- telemetry that would ever clear the flag. The gate now blocks only on real,
-- currently-pending quota reservations (reserved_bytes > 0), which the
-- 0022 release path (release_stale_accounting_reservations, driven by the
-- pvnaive process on a periodic loop) retires within the reconcile window.
-- The sticky anomaly flag remains a data-quality signal for operators in the
-- read/usage path and is unchanged there.

ALTER FUNCTION pvnaive.direct_naive_accounting_authorize(uuid,timestamptz)
    RENAME TO direct_naive_accounting_authorize_v10;
REVOKE ALL ON FUNCTION pvnaive.direct_naive_accounting_authorize_v10(uuid,timestamptz)
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
           (COALESCE(a.reserved_bytes, 0) = 0)
      FROM pvnaive.direct_naive_accounting_authorize_v9(
               p_runtime_credential_id,
               p_observed_at
           ) AS a;
$$;

REVOKE ALL ON FUNCTION pvnaive.direct_naive_accounting_authorize(uuid,timestamptz) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION pvnaive.direct_naive_accounting_authorize(uuid,timestamptz) TO pvnaive_app;
