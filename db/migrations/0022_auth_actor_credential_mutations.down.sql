-- pvnaive:migration-version 0022
-- pvnaive:migration-name auth_actor_credential_mutations
-- pvnaive:transactional true
-- pvnaive:destructive true

-- Removes the self-service actor credential mutation functions.
DROP FUNCTION IF EXISTS pvnaive.auth_update_actor_profile(uuid, text, text);
DROP FUNCTION IF EXISTS pvnaive.auth_update_actor_password(uuid, text);
DELETE FROM pvnaive.schema_migrations WHERE version = 22;
