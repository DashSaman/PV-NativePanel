-- PVNaive concurrent session limit rollback
-- pvnaive:migration-version 0019
-- pvnaive:migration-name concurrent_session_limit
-- pvnaive:transactional true
-- pvnaive:destructive true

DROP FUNCTION pvnaive.direct_naive_accounting_ingest(
    uuid,text,text,uuid,uuid,bigint,timestamptz,boolean,bigint,bigint,boolean
);

ALTER FUNCTION pvnaive.direct_naive_accounting_ingest_v17(
    uuid,text,text,uuid,uuid,bigint,timestamptz,boolean,bigint,bigint,boolean
) RENAME TO direct_naive_accounting_ingest;

REVOKE ALL ON FUNCTION pvnaive.direct_naive_accounting_ingest(
    uuid,text,text,uuid,uuid,bigint,timestamptz,boolean,bigint,bigint,boolean
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION pvnaive.direct_naive_accounting_ingest(
    uuid,text,text,uuid,uuid,bigint,timestamptz,boolean,bigint,bigint,boolean
) TO pvnaive_app;

ALTER TABLE pvnaive.service_terms DROP COLUMN concurrency_limit;

DELETE FROM pvnaive.schema_migrations WHERE version = 19;
