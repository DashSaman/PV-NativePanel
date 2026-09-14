-- pvnaive:migration-version 0026
-- pvnaive:migration-name runtime_log_events
-- pvnaive:transactional true
-- pvnaive:destructive true

DROP FUNCTION IF EXISTS pvnaive.runtime_log_events(integer);

DELETE FROM pvnaive.schema_migrations WHERE version = 26;
