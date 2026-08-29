-- pvnaive:migration-version 0006
-- Source: PVNaive direct subscription token projection rollback
-- pvnaive:migration-name direct_subscription_tokens
-- pvnaive:transactional true
-- pvnaive:destructive true

SET LOCAL lock_timeout = '10s';
SET LOCAL statement_timeout = '120s';
SET LOCAL idle_in_transaction_session_timeout = '60s';
SET LOCAL ROLE pvnaive_owner;

DROP TRIGGER IF EXISTS direct_subscription_runtime_credential_sync ON pvnaive.naive_runtime_credentials;
DROP TRIGGER IF EXISTS direct_subscription_service_term_sync ON pvnaive.service_terms;
DROP TRIGGER IF EXISTS direct_subscription_user_sync ON pvnaive.users;

DROP FUNCTION IF EXISTS pvnaive.sync_direct_subscription_runtime_credential();
DROP FUNCTION IF EXISTS pvnaive.sync_direct_subscription_service_term();
DROP FUNCTION IF EXISTS pvnaive.sync_direct_subscription_user();
DROP FUNCTION IF EXISTS pvnaive.resolve_direct_subscription_token(bytea);
DROP TABLE IF EXISTS pvnaive.direct_subscription_tokens;
DELETE FROM pvnaive.schema_migrations WHERE version = 6;
