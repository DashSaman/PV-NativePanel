-- pvnaive:migration-version 0021
-- pvnaive:migration-name ip_session_history
-- pvnaive:transactional true
-- pvnaive:destructive true

SET LOCAL lock_timeout = '10s';
SET LOCAL statement_timeout = '120s';

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pvnaive.direct_naive_session_history LIMIT 1) THEN
        RAISE EXCEPTION 'schema21 rollback refused: retained session history exists';
    END IF;
END;
$$;

SET LOCAL ROLE pvnaive_owner;

DROP FUNCTION IF EXISTS pvnaive.list_customer_session_history(uuid,timestamptz,integer);
DROP FUNCTION IF EXISTS pvnaive.sync_direct_naive_session_history(timestamptz);
DROP TABLE IF EXISTS pvnaive.direct_naive_session_history;
DELETE FROM pvnaive.schema_migrations WHERE version = 21;
