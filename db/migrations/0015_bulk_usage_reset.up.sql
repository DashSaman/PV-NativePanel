-- pvnaive:migration-version 0015
-- Source: PVNaive Task #8 Bulk Reset Usage
-- pvnaive:migration-name bulk_usage_reset
-- pvnaive:transactional true
-- pvnaive:destructive false

SET LOCAL lock_timeout = '10s';
SET LOCAL statement_timeout = '120s';
SET LOCAL idle_in_transaction_session_timeout = '60s';
SET LOCAL ROLE pvnaive_owner;

CREATE TABLE pvnaive.customer_bulk_operation_keys (
    tenant_id uuid NOT NULL REFERENCES pvnaive.tenants(id) ON DELETE RESTRICT,
    actor_id uuid NOT NULL REFERENCES pvnaive.actors(id) ON DELETE RESTRICT,
    idempotency_key text NOT NULL CHECK (char_length(idempotency_key) BETWEEN 8 AND 160),
    request_hash bytea NOT NULL CHECK (octet_length(request_hash) = 32),
    action text NOT NULL CHECK (char_length(action) BETWEEN 1 AND 64),
    created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    PRIMARY KEY (tenant_id, idempotency_key)
);

-- customer_bulk_operations is FORCE RLS. Temporarily relax FORCE only inside
-- this migration transaction so the table owner can backfill the shared
-- idempotency registry. ALTER TABLE takes an AccessExclusive lock and FORCE
-- is restored before commit, so no external request observes a weaker policy.
ALTER TABLE pvnaive.customer_bulk_operations NO FORCE ROW LEVEL SECURITY;
INSERT INTO pvnaive.customer_bulk_operation_keys (
    tenant_id,actor_id,idempotency_key,request_hash,action,created_at
)
SELECT tenant_id,actor_id,idempotency_key,request_hash,action,created_at
FROM pvnaive.customer_bulk_operations;
ALTER TABLE pvnaive.customer_bulk_operations FORCE ROW LEVEL SECURITY;

CREATE INDEX customer_bulk_operation_keys_actor_created_idx
    ON pvnaive.customer_bulk_operation_keys (tenant_id, actor_id, created_at DESC);

ALTER TABLE pvnaive.customer_bulk_operation_keys ENABLE ROW LEVEL SECURITY;
ALTER TABLE pvnaive.customer_bulk_operation_keys FORCE ROW LEVEL SECURITY;
CREATE POLICY customer_bulk_operation_keys_tenant_isolation ON pvnaive.customer_bulk_operation_keys
    FOR ALL TO pvnaive_app
    USING (pvnaive.has_tenant_access(tenant_id))
    WITH CHECK (pvnaive.has_tenant_access(tenant_id));

CREATE TABLE pvnaive.customer_bulk_reset_operations (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id uuid NOT NULL REFERENCES pvnaive.tenants(id) ON DELETE RESTRICT,
    actor_id uuid NOT NULL REFERENCES pvnaive.actors(id) ON DELETE RESTRICT,
    idempotency_key text NOT NULL CHECK (char_length(idempotency_key) BETWEEN 8 AND 160),
    request_hash bytea NOT NULL CHECK (octet_length(request_hash) = 32),
    action text NOT NULL CHECK (action = 'reset_usage'),
    request jsonb NOT NULL DEFAULT '{}'::jsonb CHECK (jsonb_typeof(request) = 'object'),
    status text NOT NULL DEFAULT 'previewed' CHECK (status IN ('previewed','executed','rejected')),
    preview jsonb NOT NULL DEFAULT '{}'::jsonb CHECK (jsonb_typeof(preview) = 'object'),
    result jsonb CHECK (result IS NULL OR jsonb_typeof(result) = 'object'),
    created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    executed_at timestamptz,
    UNIQUE (tenant_id, idempotency_key),
    FOREIGN KEY (tenant_id, idempotency_key)
        REFERENCES pvnaive.customer_bulk_operation_keys(tenant_id, idempotency_key) ON DELETE RESTRICT
);

CREATE INDEX customer_bulk_reset_operations_actor_created_idx
    ON pvnaive.customer_bulk_reset_operations (tenant_id, actor_id, created_at DESC);

ALTER TABLE pvnaive.customer_bulk_reset_operations ENABLE ROW LEVEL SECURITY;
ALTER TABLE pvnaive.customer_bulk_reset_operations FORCE ROW LEVEL SECURITY;
CREATE POLICY customer_bulk_reset_operations_tenant_isolation ON pvnaive.customer_bulk_reset_operations
    FOR ALL TO pvnaive_app
    USING (pvnaive.has_tenant_access(tenant_id))
    WITH CHECK (pvnaive.has_tenant_access(tenant_id));

GRANT SELECT, INSERT, UPDATE ON TABLE
    pvnaive.customer_bulk_operation_keys,
    pvnaive.customer_bulk_reset_operations
TO pvnaive_app;

COMMENT ON TABLE pvnaive.customer_bulk_operation_keys IS
    'Shared tenant-scoped idempotency namespace for all customer bulk Preview -> Execute operations; schema15 backfills legacy bulk keys.';
COMMENT ON TABLE pvnaive.customer_bulk_reset_operations IS
    'Owner-only reset_usage bulk ledger. Each affected customer uses the Task7 exact-accounting reset primitive in an independent transaction.';
