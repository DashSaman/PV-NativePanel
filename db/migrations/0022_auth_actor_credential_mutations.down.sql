-- pvnaive:migration-version 0022 down migration
-- Removes the self-service actor credential mutation functions.
DROP FUNCTION IF EXISTS pvnaive.auth_update_actor_profile(uuid, text, text);
DROP FUNCTION IF EXISTS pvnaive.auth_update_actor_password(uuid, text);
