-- pvnaive:migration-version 0031
-- pvnaive:migration-name network_telemetry
-- pvnaive:transactional true
-- pvnaive:destructive true

DROP FUNCTION IF EXISTS pvnaive.network_agg_read(bigint);
DROP FUNCTION IF EXISTS pvnaive.network_agg_upsert(uuid,text,text,bigint,bigint,real,bigint,real,bigint,timestamptz);
DROP FUNCTION IF EXISTS pvnaive.network_sample_ingest(uuid,text,uuid,uuid,bigint,timestamptz,text,integer,integer,bigint,bigint,bigint,bigint,bigint);
DROP FUNCTION IF EXISTS pvnaive.network_ensure_partition(date);

DROP TABLE IF EXISTS pvnaive.user_node_network_agg;
DROP TABLE IF EXISTS pvnaive.session_network_samples;

DELETE FROM pvnaive.schema_migrations WHERE version = 31;
