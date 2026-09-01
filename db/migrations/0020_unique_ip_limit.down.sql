-- PVNaive unique IP limit rollback
-- pvnaive:migration-version 0020
-- pvnaive:migration-name unique_ip_limit
-- pvnaive:transactional true
-- pvnaive:destructive true

DROP FUNCTION pvnaive.direct_naive_accounting_ingest(
    uuid,text,text,uuid,uuid,bigint,timestamptz,boolean,bigint,bigint,boolean,text
);

ALTER FUNCTION pvnaive.direct_naive_accounting_ingest_v19(
    uuid,text,text,uuid,uuid,bigint,timestamptz,boolean,bigint,bigint,boolean
) RENAME TO direct_naive_accounting_ingest;

REVOKE ALL ON FUNCTION pvnaive.direct_naive_accounting_ingest(
    uuid,text,text,uuid,uuid,bigint,timestamptz,boolean,bigint,bigint,boolean
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION pvnaive.direct_naive_accounting_ingest(
    uuid,text,text,uuid,uuid,bigint,timestamptz,boolean,bigint,bigint,boolean
) TO pvnaive_app;

ALTER TABLE pvnaive.service_terms DROP COLUMN unique_ip_limit;
ALTER TABLE pvnaive.plans DROP COLUMN unique_ip_limit;

DELETE FROM pvnaive.schema_migrations WHERE version = 20;
