-- pvnaive:migration-version 0005
-- Source: PVNaive direct customer mutation idempotency rollback
-- pvnaive:migration-name customer_mutation_idempotency
-- pvnaive:transactional true
-- pvnaive:destructive true

SET LOCAL lock_timeout = '10s';
SET LOCAL statement_timeout = '120s';
SET LOCAL idle_in_transaction_session_timeout = '60s';
SET LOCAL ROLE pvnaive_owner;

DROP TABLE IF EXISTS pvnaive.customer_mutation_keys;
DELETE FROM pvnaive.schema_migrations WHERE version = 5;
