-- pvnaive:migration-version 0006
-- Source: PVNaive direct subscription token projection rollback
-- pvnaive:migration-name direct_subscription_tokens
-- pvnaive:transactional true
-- pvnaive:destructive true

SET LOCAL lock_timeout = '10s';
SET LOCAL statement_timeout = '120s';
SET LOCAL idle_in_transaction_session_timeout = '60s';
SET LOCAL ROLE pvnaive_owner;

DROP FUNCTION IF EXISTS pvnaive.resolve_direct_subscription_token(bytea);
DROP TABLE IF EXISTS pvnaive.direct_subscription_tokens;
DELETE FROM pvnaive.schema_migrations WHERE version = 6;
