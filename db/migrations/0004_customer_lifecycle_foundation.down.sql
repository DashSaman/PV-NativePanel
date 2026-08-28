-- pvnaive:migration-version 0004
-- Source: PVNaive customer lifecycle foundation rollback
-- pvnaive:migration-name customer_lifecycle_foundation
-- pvnaive:transactional true
-- pvnaive:destructive true

SET LOCAL lock_timeout = '10s';
SET LOCAL statement_timeout = '120s';
SET LOCAL idle_in_transaction_session_timeout = '60s';
SET LOCAL ROLE pvnaive_owner;

DROP TABLE IF EXISTS pvnaive.user_runtime_credentials;
DROP TRIGGER IF EXISTS service_terms_validate_plan_before_write ON pvnaive.service_terms;
DROP FUNCTION IF EXISTS pvnaive.validate_service_term_plan_scope();
DROP TABLE IF EXISTS pvnaive.service_terms;

ALTER TABLE pvnaive.users
    DROP COLUMN IF EXISTS revision;

ALTER TABLE pvnaive.plans
    DROP COLUMN IF EXISTS revision;

DELETE FROM pvnaive.schema_migrations WHERE version = 4;
