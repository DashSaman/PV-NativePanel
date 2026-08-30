-- pvnaive:migration-version 0014
-- Source: PVNaive Task #7 manual exact-accounting usage reset rollback
-- pvnaive:migration-name manual_usage_reset
-- pvnaive:transactional true
-- pvnaive:destructive true

SET LOCAL lock_timeout = '10s';
SET LOCAL statement_timeout = '120s';
SET LOCAL idle_in_transaction_session_timeout = '60s';
SET LOCAL ROLE pvnaive_owner;

DROP FUNCTION IF EXISTS pvnaive.direct_naive_accounting_reset(uuid,timestamptz,bigint);
DROP FUNCTION IF EXISTS pvnaive.direct_naive_accounting_read(uuid,timestamptz,bigint);
ALTER FUNCTION pvnaive.direct_naive_accounting_read_v13(uuid,timestamptz,bigint)
    RENAME TO direct_naive_accounting_read;
GRANT EXECUTE ON FUNCTION pvnaive.direct_naive_accounting_read(uuid,timestamptz,bigint) TO pvnaive_app;

DROP TABLE IF EXISTS pvnaive.direct_naive_accounting_reset_events;
ALTER TABLE pvnaive.direct_naive_accounting_terms DROP COLUMN IF EXISTS last_reset_at;

DELETE FROM pvnaive.schema_migrations WHERE version = 14;
