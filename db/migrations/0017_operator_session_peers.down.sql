-- pvnaive:migration-version 0017
-- pvnaive:transactional true
-- pvnaive:destructive true

SET LOCAL lock_timeout = '10s';
SET LOCAL statement_timeout = '120s';
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pvnaive.direct_naive_accounting_session_peers LIMIT 1) THEN
        RAISE EXCEPTION 'schema17 rollback refused: trusted session peer evidence exists';
    END IF;
END;
$$;

SET LOCAL ROLE pvnaive_owner;

DROP FUNCTION IF EXISTS pvnaive.list_active_customer_sessions(uuid,timestamptz,integer);
DROP FUNCTION IF EXISTS pvnaive.direct_naive_accounting_record_session_peer(uuid,text,uuid,uuid,inet,timestamptz);
DROP TABLE IF EXISTS pvnaive.direct_naive_accounting_session_peers;
DELETE FROM pvnaive.schema_migrations WHERE version=17;
