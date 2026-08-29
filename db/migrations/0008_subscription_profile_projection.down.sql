-- pvnaive:migration-version 0008
-- pvnaive:migration-name subscription_profile_projection
-- pvnaive:transactional true
-- pvnaive:destructive true

SET LOCAL lock_timeout = '10s';
SET LOCAL statement_timeout = '120s';
SET LOCAL idle_in_transaction_session_timeout = '60s';
SET LOCAL ROLE pvnaive_owner;

DROP FUNCTION IF EXISTS pvnaive.resolve_direct_subscription_profile(bytea);
DROP TRIGGER IF EXISTS direct_subscription_service_profile_sync ON pvnaive.service_terms;
DROP FUNCTION IF EXISTS pvnaive.sync_direct_subscription_service_profile();

ALTER TABLE pvnaive.direct_subscription_tokens
    DROP COLUMN first_connected_at,
    DROP COLUMN starts_at,
    DROP COLUMN start_policy,
    DROP COLUMN duration_seconds,
    DROP COLUMN quota_bytes;

DELETE FROM pvnaive.schema_migrations WHERE version = 8;
