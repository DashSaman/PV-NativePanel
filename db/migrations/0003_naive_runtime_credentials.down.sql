-- pvnaive:migration-version 0003
-- Source: PVNaive Naive runtime credential management rollback
-- pvnaive:migration-name naive_runtime_credentials
-- pvnaive:transactional true
-- pvnaive:destructive true

SET LOCAL lock_timeout = '10s';
SET LOCAL statement_timeout = '120s';
SET LOCAL idle_in_transaction_session_timeout = '60s';
SET LOCAL ROLE pvnaive_owner;

DROP INDEX IF EXISTS pvnaive.runtime_revisions_scope_idempotency_uidx;

ALTER TABLE pvnaive.runtime_revisions
    DROP COLUMN IF EXISTS idempotency_key;

DROP TABLE IF EXISTS pvnaive.naive_runtime_credentials;
