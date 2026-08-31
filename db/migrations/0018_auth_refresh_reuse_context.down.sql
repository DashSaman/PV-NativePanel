-- pvnaive:migration-version 0018
-- pvnaive:transactional true
-- pvnaive:destructive false

SET LOCAL lock_timeout = '10s';
SET LOCAL statement_timeout = '120s';
SET LOCAL ROLE pvnaive_owner;

DROP FUNCTION IF EXISTS pvnaive.auth_refresh_session_metadata(bytea);
DELETE FROM pvnaive.schema_migrations WHERE version = 18;
