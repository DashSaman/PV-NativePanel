-- pvnaive:migration-version 0013
-- Source: PVNaive Task #6 public account projection rollback
-- pvnaive:migration-name subscription_account_projection
-- pvnaive:transactional true
-- pvnaive:destructive true

SET LOCAL lock_timeout = '10s';
SET LOCAL statement_timeout = '120s';
SET LOCAL idle_in_transaction_session_timeout = '60s';
SET LOCAL ROLE pvnaive_owner;

DROP FUNCTION IF EXISTS pvnaive.resolve_direct_subscription_account_profile(bytea);

ALTER TABLE pvnaive.direct_subscription_tokens
    DROP CONSTRAINT IF EXISTS direct_subscription_accounting_baseline_truth_check,
    DROP COLUMN IF EXISTS accounting_baseline_download_bytes,
    DROP COLUMN IF EXISTS accounting_baseline_upload_bytes,
    DROP COLUMN IF EXISTS accounting_baseline_cutoff_at,
    DROP COLUMN IF EXISTS accounting_baseline_source,
    DROP COLUMN IF EXISTS accounting_baseline_state;

DELETE FROM pvnaive.schema_migrations WHERE version = 13;
