-- pvnaive:migration-version 0033
-- pvnaive:migration-name pool_registry
-- pvnaive:transactional true
-- pvnaive:destructive true

DROP FUNCTION IF EXISTS pvnaive.pool_nodes_list();
DROP FUNCTION IF EXISTS pvnaive.pool_node_maintenance_set(uuid,text);
DROP FUNCTION IF EXISTS pvnaive.pool_node_heartbeat(uuid,text,bigint);
DROP FUNCTION IF EXISTS pvnaive.pool_revision_latest(uuid);
DROP FUNCTION IF EXISTS pvnaive.pool_revision_publish(uuid,jsonb,text);
DROP FUNCTION IF EXISTS pvnaive.pool_node_enroll(text,text,text,integer,uuid);
DROP FUNCTION IF EXISTS pvnaive.pool_enrollment_token_record(text,text,integer);
DROP TABLE IF EXISTS pvnaive.pool_enrollment_tokens;
DROP TABLE IF EXISTS pvnaive.pool_manifest_revisions;
DROP TABLE IF EXISTS pvnaive.pool_nodes;

DELETE FROM pvnaive.schema_migrations WHERE version = 33;
