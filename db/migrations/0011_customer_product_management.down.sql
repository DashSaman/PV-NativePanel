-- pvnaive:migration-version 0011
-- Source: PVNaive WS2 customer product management rollback
-- pvnaive:migration-name customer_product_management
-- pvnaive:transactional true
-- pvnaive:destructive true

SET LOCAL lock_timeout = '10s';
SET LOCAL statement_timeout = '120s';
SET LOCAL idle_in_transaction_session_timeout = '60s';
SET LOCAL ROLE pvnaive_owner;

DROP TABLE IF EXISTS pvnaive.customer_bulk_operations;

DROP TRIGGER IF EXISTS plan_default_group_ref_guard ON pvnaive.plans;
DROP FUNCTION IF EXISTS pvnaive.validate_plan_default_group_ref();
DROP TRIGGER IF EXISTS plan_tag_assignment_refs_guard ON pvnaive.plan_tag_assignments;
DROP FUNCTION IF EXISTS pvnaive.validate_plan_tag_assignment_refs();
DROP TRIGGER IF EXISTS customer_tag_assignment_refs_guard ON pvnaive.customer_tag_assignments;
DROP FUNCTION IF EXISTS pvnaive.validate_customer_tag_assignment_refs();
DROP TABLE IF EXISTS pvnaive.plan_tag_assignments;
DROP TABLE IF EXISTS pvnaive.customer_tag_assignments;
DROP TRIGGER IF EXISTS customer_profile_refs_guard ON pvnaive.customer_profiles;
DROP FUNCTION IF EXISTS pvnaive.validate_customer_profile_refs();
DROP TABLE IF EXISTS pvnaive.customer_profiles;

ALTER TABLE pvnaive.plans DROP COLUMN IF EXISTS default_group_id;
DROP TABLE IF EXISTS pvnaive.customer_tags;
DROP TABLE IF EXISTS pvnaive.customer_groups;

ALTER TABLE pvnaive.service_terms
    DROP COLUMN IF EXISTS renewed_from_term_id,
    DROP COLUMN IF EXISTS renewal_kind,
    DROP COLUMN IF EXISTS no_expiry;

ALTER TABLE pvnaive.plans
    DROP CONSTRAINT IF EXISTS plans_no_expiry_start_policy_check,
    DROP CONSTRAINT IF EXISTS plans_custom_reset_days_check,
    DROP COLUMN IF EXISTS sort_order,
    DROP COLUMN IF EXISTS enabled,
    DROP COLUMN IF EXISTS reset_custom_days,
    DROP COLUMN IF EXISTS reset_strategy,
    DROP COLUMN IF EXISTS no_expiry,
    DROP COLUMN IF EXISTS start_policy;

DELETE FROM pvnaive.schema_migrations WHERE version = 11;
