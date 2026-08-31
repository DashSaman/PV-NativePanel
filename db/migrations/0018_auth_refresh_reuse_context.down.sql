-- pvnaive:migration-version 0018
-- pvnaive:transactional true
-- pvnaive:destructive true

SET LOCAL lock_timeout = '10s';
SET LOCAL statement_timeout = '120s';
SET LOCAL ROLE pvnaive_owner;

DROP FUNCTION IF EXISTS pvnaive.auth_refresh_session_metadata(bytea);
DROP FUNCTION IF EXISTS pvnaive.auth_rotate_session(bytea,bytea,bytea,bytea,timestamptz);
ALTER FUNCTION pvnaive.auth_rotate_session_v17(bytea,bytea,bytea,bytea,timestamptz)
    RENAME TO auth_rotate_session;
REVOKE ALL ON FUNCTION pvnaive.auth_rotate_session(bytea,bytea,bytea,bytea,timestamptz) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION pvnaive.auth_rotate_session(bytea,bytea,bytea,bytea,timestamptz) TO pvnaive_app;
DELETE FROM pvnaive.schema_migrations WHERE version = 18;
