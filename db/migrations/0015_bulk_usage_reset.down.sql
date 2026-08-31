-- pvnaive:migration-version 0015
-- Source: PVNaive Task #8 Bulk Reset Usage rollback
-- pvnaive:migration-name bulk_usage_reset
-- pvnaive:transactional true
-- pvnaive:destructive true

SET LOCAL lock_timeout = '10s';
SET LOCAL statement_timeout = '120s';
SET LOCAL idle_in_transaction_session_timeout = '60s';

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pvnaive.customer_bulk_reset_operations) THEN
        RAISE EXCEPTION 'cannot roll back bulk_usage_reset while bulk reset history exists';
    END IF;
END;
$$;

SET LOCAL ROLE pvnaive_owner;
DROP TABLE pvnaive.customer_bulk_reset_operations;
DROP TABLE pvnaive.customer_bulk_operation_keys;
DELETE FROM pvnaive.schema_migrations WHERE version = 15;
