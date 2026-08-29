-- pvnaive:migration-version 0010
-- Source: PVNaive pending reservation accounting-completeness rollback
-- pvnaive:migration-name pending_reservation_completeness
-- pvnaive:transactional true
-- pvnaive:destructive true

SET LOCAL lock_timeout = '10s';
SET LOCAL statement_timeout = '120s';
SET LOCAL idle_in_transaction_session_timeout = '60s';
SET LOCAL ROLE pvnaive_owner;

DROP FUNCTION IF EXISTS pvnaive.direct_naive_accounting_authorize(uuid,timestamptz);
DROP FUNCTION IF EXISTS pvnaive.direct_naive_accounting_read(uuid,timestamptz,bigint);

ALTER FUNCTION pvnaive.direct_naive_accounting_authorize_v9(uuid,timestamptz)
    RENAME TO direct_naive_accounting_authorize;
ALTER FUNCTION pvnaive.direct_naive_accounting_read_v9(uuid,timestamptz,bigint)
    RENAME TO direct_naive_accounting_read;

REVOKE ALL ON FUNCTION pvnaive.direct_naive_accounting_authorize(uuid,timestamptz) FROM PUBLIC;
REVOKE ALL ON FUNCTION pvnaive.direct_naive_accounting_read(uuid,timestamptz,bigint) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION pvnaive.direct_naive_accounting_authorize(uuid,timestamptz) TO pvnaive_app;
GRANT EXECUTE ON FUNCTION pvnaive.direct_naive_accounting_read(uuid,timestamptz,bigint) TO pvnaive_app;

DELETE FROM pvnaive.schema_migrations WHERE version = 10;
