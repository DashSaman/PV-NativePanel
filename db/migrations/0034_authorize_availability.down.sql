-- pvnaive:migration-version 0034
-- pvnaive:migration-name authorize_availability
-- pvnaive:transactional true
-- pvnaive:destructive false

SET LOCAL lock_timeout = '10s';
SET LOCAL statement_timeout = '120s';
SET LOCAL idle_in_transaction_session_timeout = '60s';
SET LOCAL ROLE pvnaive_owner;

DROP FUNCTION IF EXISTS pvnaive.direct_naive_accounting_authorize(uuid,timestamptz);

ALTER FUNCTION pvnaive.direct_naive_accounting_authorize_v10(uuid,timestamptz)
    RENAME TO direct_naive_accounting_authorize;

REVOKE ALL ON FUNCTION pvnaive.direct_naive_accounting_authorize(uuid,timestamptz) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION pvnaive.direct_naive_accounting_authorize(uuid,timestamptz) TO pvnaive_app;
