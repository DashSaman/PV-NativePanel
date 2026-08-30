-- pvnaive:migration-version 0012
-- Source: PVNaive Task #5 accounting baseline truth rollback
-- pvnaive:migration-name accounting_baseline_truth
-- pvnaive:transactional true
-- pvnaive:destructive true

SET LOCAL lock_timeout = '10s';
SET LOCAL statement_timeout = '120s';
SET LOCAL idle_in_transaction_session_timeout = '60s';
SET LOCAL ROLE pvnaive_owner;

DROP TRIGGER IF EXISTS service_terms_accounting_baseline_immutable ON pvnaive.service_terms;
DROP FUNCTION IF EXISTS pvnaive.prevent_accounting_baseline_mutation();

ALTER TABLE pvnaive.service_terms
    DROP CONSTRAINT IF EXISTS service_terms_accounting_baseline_truth_check,
    DROP COLUMN IF EXISTS accounting_baseline_download_bytes,
    DROP COLUMN IF EXISTS accounting_baseline_upload_bytes,
    DROP COLUMN IF EXISTS accounting_baseline_cutoff_at,
    DROP COLUMN IF EXISTS accounting_baseline_source,
    DROP COLUMN IF EXISTS accounting_baseline_state;

DELETE FROM pvnaive.schema_migrations WHERE version = 12;
