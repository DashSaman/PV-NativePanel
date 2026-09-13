-- pvnaive:migration-version 0022
-- pvnaive:migration-name accounting_reconciliation
-- pvnaive:transactional true
-- pvnaive:destructive true

SET LOCAL lock_timeout = '10s';
SET LOCAL statement_timeout = '120s';

DO $$
DECLARE
    v_unreleased bigint;
BEGIN
    SELECT count(*) INTO v_unreleased
      FROM pvnaive.direct_naive_accounting_claims
     WHERE released_at IS NOT NULL;
    IF v_unreleased > 0 THEN
        RAISE EXCEPTION 'schema22 rollback refused: % released claims exist; history would be lost', v_unreleased;
    END IF;
END;
$$;

SET LOCAL ROLE pvnaive_owner;

DROP FUNCTION IF EXISTS pvnaive.list_incomplete_accounting_terms();
DROP FUNCTION IF EXISTS pvnaive.customer_current_usage_term(uuid);
DROP FUNCTION IF EXISTS pvnaive.usage_history(uuid,integer);
DROP FUNCTION IF EXISTS pvnaive.usage_summary();
DROP FUNCTION IF EXISTS pvnaive.repair_accounting_completeness(uuid);
DROP FUNCTION IF EXISTS pvnaive.verify_accounting_completeness(uuid);
DROP FUNCTION IF EXISTS pvnaive.release_stale_accounting_reservations(bigint,timestamptz,bigint);

ALTER FUNCTION pvnaive.direct_naive_accounting_read(uuid,timestamptz,bigint)
    RENAME TO direct_naive_accounting_read_v22_dropped;
DROP FUNCTION pvnaive.direct_naive_accounting_read_v22_dropped(uuid,timestamptz,bigint);
ALTER FUNCTION pvnaive.direct_naive_accounting_read_v21(uuid,timestamptz,bigint)
    RENAME TO direct_naive_accounting_read;
GRANT EXECUTE ON FUNCTION pvnaive.direct_naive_accounting_read(uuid,timestamptz,bigint) TO pvnaive_app;

DROP INDEX IF EXISTS pvnaive.direct_naive_accounting_claims_unsettled_stale_idx;
ALTER TABLE pvnaive.direct_naive_accounting_claims
    DROP CONSTRAINT IF EXISTS direct_naive_accounting_claims_release_shape,
    DROP COLUMN IF EXISTS released_bytes,
    DROP COLUMN IF EXISTS released_at,
    DROP COLUMN IF EXISTS released_reason;

DELETE FROM pvnaive.schema_migrations WHERE version = 22;
