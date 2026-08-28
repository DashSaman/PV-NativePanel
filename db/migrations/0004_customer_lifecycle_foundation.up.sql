-- pvnaive:migration-version 0004
-- Source: PVNaive customer lifecycle foundation
-- pvnaive:migration-name customer_lifecycle_foundation
-- pvnaive:transactional true
-- pvnaive:destructive false

SET LOCAL lock_timeout = '10s';
SET LOCAL statement_timeout = '120s';
SET LOCAL idle_in_transaction_session_timeout = '60s';
SET LOCAL ROLE pvnaive_owner;

-- The direct system tenant is a schema-level bootstrap record. The owner role
-- normally remains subject to FORCE RLS, so temporarily remove FORCE while the
-- table is locked by this transactional migration, seed the row, then restore
-- FORCE before any application query can observe the transaction.
ALTER TABLE pvnaive.tenants NO FORCE ROW LEVEL SECURITY;

INSERT INTO pvnaive.tenants (tenant_type, slug, display_name, status)
SELECT 'system', 'direct', 'PVNaive Direct', 'active'
WHERE NOT EXISTS (
    SELECT 1 FROM pvnaive.tenants WHERE lower(slug) = 'direct'
);

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
          FROM pvnaive.tenants
         WHERE tenant_type = 'system'
           AND slug = 'direct'
           AND status = 'active'
    ) THEN
        RAISE EXCEPTION 'direct tenant slug is occupied by an incompatible tenant'
            USING ERRCODE = '23514';
    END IF;
END;
$$;

ALTER TABLE pvnaive.tenants FORCE ROW LEVEL SECURITY;

ALTER TABLE pvnaive.users
    ADD COLUMN revision bigint NOT NULL DEFAULT 1 CHECK (revision > 0);

ALTER TABLE pvnaive.plans
    ADD COLUMN revision bigint NOT NULL DEFAULT 1 CHECK (revision > 0);

CREATE TABLE pvnaive.service_terms (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id uuid NOT NULL REFERENCES pvnaive.tenants(id) ON DELETE RESTRICT,
    user_id uuid NOT NULL,
    plan_id uuid REFERENCES pvnaive.plans(id) ON DELETE RESTRICT,
    quota_bytes bigint CHECK (quota_bytes IS NULL OR quota_bytes > 0),
    duration_seconds bigint NOT NULL CHECK (duration_seconds > 0),
    reset_interval_seconds bigint CHECK (reset_interval_seconds IS NULL OR reset_interval_seconds > 0),
    start_policy text NOT NULL DEFAULT 'on_first_successful_connection'
        CHECK (start_policy IN ('on_creation', 'on_first_successful_connection', 'fixed_timestamp')),
    purchased_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    starts_at timestamptz,
    first_connected_at timestamptz,
    expires_at timestamptz,
    ended_at timestamptz,
    state text NOT NULL DEFAULT 'pending'
        CHECK (state IN ('pending', 'active', 'expired', 'quota_depleted', 'ended', 'revoked')),
    revision bigint NOT NULL DEFAULT 1 CHECK (revision > 0),
    created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    updated_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    UNIQUE (id, tenant_id),
    UNIQUE (id, tenant_id, user_id),
    FOREIGN KEY (user_id, tenant_id) REFERENCES pvnaive.users(id, tenant_id) ON DELETE RESTRICT,
    CHECK (first_connected_at IS NULL OR first_connected_at >= purchased_at),
    CHECK ((starts_at IS NULL AND expires_at IS NULL) OR
           (starts_at IS NOT NULL AND expires_at IS NOT NULL AND expires_at > starts_at)),
    CHECK (ended_at IS NULL OR starts_at IS NULL OR ended_at >= starts_at)
);

CREATE INDEX service_terms_tenant_user_state_idx
    ON pvnaive.service_terms (tenant_id, user_id, state);
CREATE INDEX service_terms_tenant_expires_idx
    ON pvnaive.service_terms (tenant_id, expires_at)
    WHERE expires_at IS NOT NULL AND state IN ('pending', 'active');

CREATE TABLE pvnaive.user_runtime_credentials (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id uuid NOT NULL REFERENCES pvnaive.tenants(id) ON DELETE RESTRICT,
    user_id uuid NOT NULL,
    service_term_id uuid NOT NULL,
    runtime_credential_id uuid NOT NULL
        REFERENCES pvnaive.naive_runtime_credentials(id) ON DELETE RESTRICT,
    role text NOT NULL DEFAULT 'primary' CHECK (role = 'primary'),
    bound_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    unbound_at timestamptz,
    FOREIGN KEY (user_id, tenant_id) REFERENCES pvnaive.users(id, tenant_id) ON DELETE RESTRICT,
    FOREIGN KEY (service_term_id, tenant_id, user_id)
        REFERENCES pvnaive.service_terms(id, tenant_id, user_id) ON DELETE RESTRICT,
    CHECK (unbound_at IS NULL OR unbound_at >= bound_at)
);

CREATE UNIQUE INDEX user_runtime_credentials_one_active_primary_uidx
    ON pvnaive.user_runtime_credentials (service_term_id)
    WHERE role = 'primary' AND unbound_at IS NULL;

CREATE UNIQUE INDEX user_runtime_credentials_runtime_active_uidx
    ON pvnaive.user_runtime_credentials (runtime_credential_id)
    WHERE unbound_at IS NULL;

CREATE INDEX user_runtime_credentials_tenant_user_idx
    ON pvnaive.user_runtime_credentials (tenant_id, user_id, bound_at DESC);

CREATE FUNCTION pvnaive.validate_service_term_plan_scope()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog, pvnaive
AS $$
DECLARE
    selected_plan_tenant uuid;
BEGIN
    IF NEW.plan_id IS NULL THEN
        RETURN NEW;
    END IF;

    SELECT tenant_id
      INTO selected_plan_tenant
      FROM pvnaive.plans
     WHERE id = NEW.plan_id
     FOR KEY SHARE;

    IF NOT FOUND THEN
        RAISE EXCEPTION 'unknown service-term plan' USING ERRCODE = '23503';
    END IF;

    IF selected_plan_tenant IS NOT NULL AND selected_plan_tenant <> NEW.tenant_id THEN
        RAISE EXCEPTION 'cross-tenant service-term plan refused' USING ERRCODE = '23514';
    END IF;

    RETURN NEW;
END;
$$;

CREATE TRIGGER service_terms_validate_plan_before_write
BEFORE INSERT OR UPDATE OF tenant_id, plan_id ON pvnaive.service_terms
FOR EACH ROW EXECUTE FUNCTION pvnaive.validate_service_term_plan_scope();

ALTER TABLE pvnaive.service_terms ENABLE ROW LEVEL SECURITY;
ALTER TABLE pvnaive.service_terms FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON pvnaive.service_terms
    USING (pvnaive.has_tenant_access(tenant_id, false))
    WITH CHECK (pvnaive.has_tenant_access(tenant_id, false));

ALTER TABLE pvnaive.user_runtime_credentials ENABLE ROW LEVEL SECURITY;
ALTER TABLE pvnaive.user_runtime_credentials FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON pvnaive.user_runtime_credentials
    USING (pvnaive.has_tenant_access(tenant_id, false))
    WITH CHECK (pvnaive.has_tenant_access(tenant_id, false));

REVOKE ALL ON pvnaive.service_terms FROM PUBLIC;
REVOKE ALL ON pvnaive.user_runtime_credentials FROM PUBLIC;
GRANT SELECT, INSERT, UPDATE ON pvnaive.service_terms TO pvnaive_app;
GRANT SELECT, INSERT, UPDATE ON pvnaive.user_runtime_credentials TO pvnaive_app;
REVOKE DELETE ON pvnaive.service_terms FROM pvnaive_app;
REVOKE DELETE ON pvnaive.user_runtime_credentials FROM pvnaive_app;
REVOKE ALL ON FUNCTION pvnaive.validate_service_term_plan_scope() FROM PUBLIC;
REVOKE ALL ON FUNCTION pvnaive.validate_service_term_plan_scope() FROM pvnaive_app;

COMMENT ON TABLE pvnaive.service_terms IS
    'Immutable commercial snapshots for direct customer service periods; exact usage is intentionally external to this table.';
COMMENT ON TABLE pvnaive.user_runtime_credentials IS
    'Historical binding from a business service term to a stable Naive Runtime credential UUID.';
