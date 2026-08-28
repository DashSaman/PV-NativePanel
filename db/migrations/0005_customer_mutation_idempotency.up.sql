-- pvnaive:migration-version 0005
-- Source: PVNaive direct customer mutation idempotency
-- pvnaive:migration-name customer_mutation_idempotency
-- pvnaive:transactional true
-- pvnaive:destructive false

SET LOCAL lock_timeout = '10s';
SET LOCAL statement_timeout = '120s';
SET LOCAL idle_in_transaction_session_timeout = '60s';
SET LOCAL ROLE pvnaive_owner;

CREATE TABLE pvnaive.customer_mutation_keys (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id uuid NOT NULL REFERENCES pvnaive.tenants(id) ON DELETE RESTRICT,
    actor_id uuid NOT NULL REFERENCES pvnaive.actors(id) ON DELETE RESTRICT,
    idempotency_key text NOT NULL CHECK (length(idempotency_key) BETWEEN 8 AND 160),
    operation text NOT NULL CHECK (length(operation) BETWEEN 1 AND 120),
    request_hash bytea NOT NULL CHECK (octet_length(request_hash) = 32),
    resource_type text NOT NULL CHECK (length(resource_type) BETWEEN 1 AND 80),
    resource_id uuid,
    created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    completed_at timestamptz,
    UNIQUE (tenant_id, actor_id, idempotency_key),
    CHECK (completed_at IS NULL OR completed_at >= created_at)
);

CREATE INDEX customer_mutation_keys_tenant_created_idx
    ON pvnaive.customer_mutation_keys (tenant_id, created_at DESC);

ALTER TABLE pvnaive.customer_mutation_keys ENABLE ROW LEVEL SECURITY;
ALTER TABLE pvnaive.customer_mutation_keys FORCE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation ON pvnaive.customer_mutation_keys
    USING (pvnaive.has_tenant_access(tenant_id, false))
    WITH CHECK (pvnaive.has_tenant_access(tenant_id, false));

REVOKE ALL ON pvnaive.customer_mutation_keys FROM PUBLIC;
GRANT SELECT, INSERT, UPDATE ON pvnaive.customer_mutation_keys TO pvnaive_app;
REVOKE DELETE ON pvnaive.customer_mutation_keys FROM pvnaive_app;

COMMENT ON TABLE pvnaive.customer_mutation_keys IS
    'Tenant-scoped idempotency ledger for direct customer and plan mutations. Request hashes prevent one key from being reused for a different payload.';
