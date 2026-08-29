-- pvnaive:migration-version 0011
-- Source: PVNaive WS2 customer product management
-- pvnaive:migration-name customer_product_management
-- pvnaive:transactional true
-- pvnaive:destructive false

SET LOCAL lock_timeout = '10s';
SET LOCAL statement_timeout = '120s';
SET LOCAL idle_in_transaction_session_timeout = '60s';
SET LOCAL ROLE pvnaive_owner;

-- The existing users.status CHECK is intentionally preserved. Product-level
-- "disabled" maps to the durable draft state so this migration stays strictly
-- additive and passes the repository's fail-closed migration policy.
ALTER TABLE pvnaive.plans
    ADD COLUMN start_policy text NOT NULL DEFAULT 'on_creation'
        CHECK (start_policy IN ('on_creation','on_first_successful_connection')),
    ADD COLUMN no_expiry boolean NOT NULL DEFAULT false,
    ADD COLUMN reset_strategy text NOT NULL DEFAULT 'none'
        CHECK (reset_strategy IN ('none','daily','weekly','monthly','yearly','custom')),
    ADD COLUMN reset_custom_days integer
        CHECK (reset_custom_days IS NULL OR (reset_custom_days > 0 AND reset_custom_days <= 3660)),
    ADD COLUMN enabled boolean NOT NULL DEFAULT true,
    ADD COLUMN sort_order integer NOT NULL DEFAULT 0;

ALTER TABLE pvnaive.plans
    ADD CONSTRAINT plans_custom_reset_days_check
    CHECK (
        (reset_strategy = 'custom' AND reset_custom_days IS NOT NULL)
        OR (reset_strategy <> 'custom' AND reset_custom_days IS NULL)
    ),
    ADD CONSTRAINT plans_no_expiry_start_policy_check
    CHECK (NOT no_expiry OR start_policy = 'on_creation');

CREATE TABLE pvnaive.customer_groups (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id uuid NOT NULL REFERENCES pvnaive.tenants(id) ON DELETE RESTRICT,
    name text NOT NULL CHECK (btrim(name) <> '' AND char_length(name) <= 120),
    enabled boolean NOT NULL DEFAULT true,
    sort_order integer NOT NULL DEFAULT 0,
    created_by_actor_id uuid REFERENCES pvnaive.actors(id) ON DELETE SET NULL,
    created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    updated_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    UNIQUE (tenant_id, name)
);

CREATE TABLE pvnaive.customer_tags (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id uuid NOT NULL REFERENCES pvnaive.tenants(id) ON DELETE RESTRICT,
    name text NOT NULL CHECK (btrim(name) <> '' AND char_length(name) <= 80),
    enabled boolean NOT NULL DEFAULT true,
    sort_order integer NOT NULL DEFAULT 0,
    created_by_actor_id uuid REFERENCES pvnaive.actors(id) ON DELETE SET NULL,
    created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    updated_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    UNIQUE (tenant_id, name)
);

ALTER TABLE pvnaive.plans
    ADD COLUMN default_group_id uuid REFERENCES pvnaive.customer_groups(id) ON DELETE SET NULL;

CREATE TABLE pvnaive.customer_profiles (
    tenant_id uuid NOT NULL REFERENCES pvnaive.tenants(id) ON DELETE RESTRICT,
    user_id uuid PRIMARY KEY REFERENCES pvnaive.users(id) ON DELETE CASCADE,
    note text NOT NULL DEFAULT '' CHECK (char_length(note) <= 4000),
    group_id uuid REFERENCES pvnaive.customer_groups(id) ON DELETE SET NULL,
    assigned_actor_id uuid REFERENCES pvnaive.actors(id) ON DELETE SET NULL,
    on_hold boolean NOT NULL DEFAULT false,
    next_plan_id uuid REFERENCES pvnaive.plans(id) ON DELETE SET NULL,
    next_plan_source_term_id uuid REFERENCES pvnaive.service_terms(id) ON DELETE SET NULL,
    next_plan_scheduled_at timestamptz,
    last_renewal_at timestamptz,
    revision bigint NOT NULL DEFAULT 0 CHECK (revision >= 0),
    updated_by_actor_id uuid REFERENCES pvnaive.actors(id) ON DELETE SET NULL,
    created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    updated_at timestamptz NOT NULL DEFAULT clock_timestamp()
);

CREATE FUNCTION pvnaive.validate_customer_profile_refs()
RETURNS trigger
LANGUAGE plpgsql
SET search_path = pg_catalog, pvnaive
AS $$
DECLARE
    v_user_tenant uuid;
    v_group_tenant uuid;
    v_plan_tenant uuid;
BEGIN
    SELECT tenant_id INTO v_user_tenant FROM pvnaive.users WHERE id = NEW.user_id;
    IF v_user_tenant IS NULL OR v_user_tenant <> NEW.tenant_id THEN
        RAISE EXCEPTION 'customer profile tenant must match user tenant';
    END IF;

    IF NEW.group_id IS NOT NULL THEN
        SELECT tenant_id INTO v_group_tenant FROM pvnaive.customer_groups WHERE id = NEW.group_id;
        IF v_group_tenant IS NULL OR v_group_tenant <> NEW.tenant_id THEN
            RAISE EXCEPTION 'customer group must belong to customer tenant';
        END IF;
    END IF;

    IF NEW.next_plan_id IS NOT NULL THEN
        SELECT tenant_id INTO v_plan_tenant FROM pvnaive.plans WHERE id = NEW.next_plan_id;
        IF v_plan_tenant IS NOT NULL AND v_plan_tenant <> NEW.tenant_id THEN
            RAISE EXCEPTION 'next plan must be global or belong to customer tenant';
        END IF;
    END IF;
    RETURN NEW;
END;
$$;

CREATE TRIGGER customer_profile_refs_guard
BEFORE INSERT OR UPDATE OF tenant_id, user_id, group_id, next_plan_id
ON pvnaive.customer_profiles
FOR EACH ROW EXECUTE FUNCTION pvnaive.validate_customer_profile_refs();

CREATE TABLE pvnaive.customer_tag_assignments (
    tenant_id uuid NOT NULL REFERENCES pvnaive.tenants(id) ON DELETE RESTRICT,
    user_id uuid NOT NULL REFERENCES pvnaive.users(id) ON DELETE CASCADE,
    tag_id uuid NOT NULL REFERENCES pvnaive.customer_tags(id) ON DELETE CASCADE,
    assigned_by_actor_id uuid REFERENCES pvnaive.actors(id) ON DELETE SET NULL,
    assigned_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    PRIMARY KEY (user_id, tag_id)
);

CREATE TABLE pvnaive.plan_tag_assignments (
    plan_id uuid NOT NULL REFERENCES pvnaive.plans(id) ON DELETE CASCADE,
    tag_id uuid NOT NULL REFERENCES pvnaive.customer_tags(id) ON DELETE CASCADE,
    tenant_id uuid NOT NULL REFERENCES pvnaive.tenants(id) ON DELETE RESTRICT,
    PRIMARY KEY (plan_id, tag_id)
);

CREATE FUNCTION pvnaive.validate_customer_tag_assignment_refs()
RETURNS trigger
LANGUAGE plpgsql
SET search_path = pg_catalog, pvnaive
AS $$
DECLARE
    v_user_tenant uuid;
    v_tag_tenant uuid;
BEGIN
    SELECT tenant_id INTO v_user_tenant FROM pvnaive.users WHERE id = NEW.user_id;
    SELECT tenant_id INTO v_tag_tenant FROM pvnaive.customer_tags WHERE id = NEW.tag_id;
    IF v_user_tenant IS NULL OR v_tag_tenant IS NULL
       OR v_user_tenant <> NEW.tenant_id OR v_tag_tenant <> NEW.tenant_id THEN
        RAISE EXCEPTION 'customer tag assignment must stay inside one tenant';
    END IF;
    RETURN NEW;
END;
$$;

CREATE TRIGGER customer_tag_assignment_refs_guard
BEFORE INSERT OR UPDATE OF tenant_id, user_id, tag_id
ON pvnaive.customer_tag_assignments
FOR EACH ROW EXECUTE FUNCTION pvnaive.validate_customer_tag_assignment_refs();

CREATE FUNCTION pvnaive.validate_plan_tag_assignment_refs()
RETURNS trigger
LANGUAGE plpgsql
SET search_path = pg_catalog, pvnaive
AS $$
DECLARE
    v_plan_tenant uuid;
    v_tag_tenant uuid;
BEGIN
    SELECT tenant_id INTO v_plan_tenant FROM pvnaive.plans WHERE id = NEW.plan_id;
    SELECT tenant_id INTO v_tag_tenant FROM pvnaive.customer_tags WHERE id = NEW.tag_id;
    IF v_plan_tenant IS NULL OR v_tag_tenant IS NULL
       OR v_plan_tenant <> NEW.tenant_id OR v_tag_tenant <> NEW.tenant_id THEN
        RAISE EXCEPTION 'plan tag assignment must stay inside one tenant';
    END IF;
    RETURN NEW;
END;
$$;

CREATE TRIGGER plan_tag_assignment_refs_guard
BEFORE INSERT OR UPDATE OF tenant_id, plan_id, tag_id
ON pvnaive.plan_tag_assignments
FOR EACH ROW EXECUTE FUNCTION pvnaive.validate_plan_tag_assignment_refs();

CREATE FUNCTION pvnaive.validate_plan_default_group_ref()
RETURNS trigger
LANGUAGE plpgsql
SET search_path = pg_catalog, pvnaive
AS $$
DECLARE
    v_group_tenant uuid;
BEGIN
    IF NEW.default_group_id IS NOT NULL THEN
        SELECT tenant_id INTO v_group_tenant FROM pvnaive.customer_groups WHERE id = NEW.default_group_id;
        IF NEW.tenant_id IS NULL OR v_group_tenant IS NULL OR v_group_tenant <> NEW.tenant_id THEN
            RAISE EXCEPTION 'plan default group must belong to plan tenant';
        END IF;
    END IF;
    RETURN NEW;
END;
$$;

CREATE TRIGGER plan_default_group_ref_guard
BEFORE INSERT OR UPDATE OF tenant_id, default_group_id
ON pvnaive.plans
FOR EACH ROW EXECUTE FUNCTION pvnaive.validate_plan_default_group_ref();

ALTER TABLE pvnaive.service_terms
    ADD COLUMN no_expiry boolean NOT NULL DEFAULT false,
    ADD COLUMN renewal_kind text NOT NULL DEFAULT 'initial'
        CHECK (renewal_kind IN ('initial','renew_current','renew_plan','custom','next_plan')),
    ADD COLUMN renewed_from_term_id uuid REFERENCES pvnaive.service_terms(id) ON DELETE SET NULL;

COMMENT ON COLUMN pvnaive.service_terms.no_expiry IS
    'When true, expires_at is intentionally NULL and duration_seconds is retained only for backward-compatible storage; product logic must ignore it.';
COMMENT ON COLUMN pvnaive.plans.no_expiry IS
    'When true, duration_seconds is a backward-compatible storage placeholder and is ignored by product renewal semantics.';

CREATE TABLE pvnaive.customer_bulk_operations (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id uuid NOT NULL REFERENCES pvnaive.tenants(id) ON DELETE RESTRICT,
    actor_id uuid NOT NULL REFERENCES pvnaive.actors(id) ON DELETE RESTRICT,
    idempotency_key text NOT NULL CHECK (char_length(idempotency_key) BETWEEN 8 AND 160),
    request_hash bytea NOT NULL CHECK (octet_length(request_hash) = 32),
    action text NOT NULL CHECK (action IN (
        'enable','suspend','revoke','safe_delete','extend_days','add_volume','set_volume',
        'apply_plan','assign_group','add_tag','remove_tag','reissue_subscription'
    )),
    request jsonb NOT NULL DEFAULT '{}'::jsonb CHECK (jsonb_typeof(request) = 'object'),
    status text NOT NULL DEFAULT 'previewed' CHECK (status IN ('previewed','executed','rejected')),
    preview jsonb NOT NULL DEFAULT '{}'::jsonb CHECK (jsonb_typeof(preview) = 'object'),
    result jsonb CHECK (result IS NULL OR jsonb_typeof(result) = 'object'),
    created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    executed_at timestamptz,
    UNIQUE (tenant_id, idempotency_key)
);

CREATE INDEX customer_groups_tenant_sort_idx
    ON pvnaive.customer_groups (tenant_id, enabled, sort_order, name);
CREATE INDEX customer_tags_tenant_sort_idx
    ON pvnaive.customer_tags (tenant_id, enabled, sort_order, name);
CREATE INDEX customer_profiles_tenant_group_idx
    ON pvnaive.customer_profiles (tenant_id, group_id, updated_at DESC);
CREATE INDEX customer_profiles_next_plan_idx
    ON pvnaive.customer_profiles (tenant_id, next_plan_id)
    WHERE next_plan_id IS NOT NULL;
CREATE INDEX customer_tag_assignments_tenant_tag_idx
    ON pvnaive.customer_tag_assignments (tenant_id, tag_id, user_id);
CREATE INDEX plan_tag_assignments_tenant_tag_idx
    ON pvnaive.plan_tag_assignments (tenant_id, tag_id, plan_id);
CREATE INDEX customer_bulk_operations_actor_created_idx
    ON pvnaive.customer_bulk_operations (tenant_id, actor_id, created_at DESC);
CREATE INDEX users_tenant_updated_idx
    ON pvnaive.users (tenant_id, updated_at DESC, id);
CREATE INDEX service_terms_user_renewal_idx
    ON pvnaive.service_terms (tenant_id, user_id, purchased_at DESC, created_at DESC);

ALTER TABLE pvnaive.customer_groups ENABLE ROW LEVEL SECURITY;
ALTER TABLE pvnaive.customer_groups FORCE ROW LEVEL SECURITY;
CREATE POLICY customer_groups_tenant_isolation ON pvnaive.customer_groups
    FOR ALL TO pvnaive_app
    USING (pvnaive.has_tenant_access(tenant_id))
    WITH CHECK (pvnaive.has_tenant_access(tenant_id));

ALTER TABLE pvnaive.customer_tags ENABLE ROW LEVEL SECURITY;
ALTER TABLE pvnaive.customer_tags FORCE ROW LEVEL SECURITY;
CREATE POLICY customer_tags_tenant_isolation ON pvnaive.customer_tags
    FOR ALL TO pvnaive_app
    USING (pvnaive.has_tenant_access(tenant_id))
    WITH CHECK (pvnaive.has_tenant_access(tenant_id));

ALTER TABLE pvnaive.customer_profiles ENABLE ROW LEVEL SECURITY;
ALTER TABLE pvnaive.customer_profiles FORCE ROW LEVEL SECURITY;
CREATE POLICY customer_profiles_tenant_isolation ON pvnaive.customer_profiles
    FOR ALL TO pvnaive_app
    USING (pvnaive.has_tenant_access(tenant_id))
    WITH CHECK (pvnaive.has_tenant_access(tenant_id));

ALTER TABLE pvnaive.customer_tag_assignments ENABLE ROW LEVEL SECURITY;
ALTER TABLE pvnaive.customer_tag_assignments FORCE ROW LEVEL SECURITY;
CREATE POLICY customer_tag_assignments_tenant_isolation ON pvnaive.customer_tag_assignments
    FOR ALL TO pvnaive_app
    USING (pvnaive.has_tenant_access(tenant_id))
    WITH CHECK (pvnaive.has_tenant_access(tenant_id));

ALTER TABLE pvnaive.plan_tag_assignments ENABLE ROW LEVEL SECURITY;
ALTER TABLE pvnaive.plan_tag_assignments FORCE ROW LEVEL SECURITY;
CREATE POLICY plan_tag_assignments_tenant_isolation ON pvnaive.plan_tag_assignments
    FOR ALL TO pvnaive_app
    USING (pvnaive.has_tenant_access(tenant_id))
    WITH CHECK (pvnaive.has_tenant_access(tenant_id));

ALTER TABLE pvnaive.customer_bulk_operations ENABLE ROW LEVEL SECURITY;
ALTER TABLE pvnaive.customer_bulk_operations FORCE ROW LEVEL SECURITY;
CREATE POLICY customer_bulk_operations_tenant_isolation ON pvnaive.customer_bulk_operations
    FOR ALL TO pvnaive_app
    USING (pvnaive.has_tenant_access(tenant_id))
    WITH CHECK (pvnaive.has_tenant_access(tenant_id));

GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE
    pvnaive.customer_groups,
    pvnaive.customer_tags,
    pvnaive.customer_profiles,
    pvnaive.customer_tag_assignments,
    pvnaive.plan_tag_assignments
TO pvnaive_app;
GRANT SELECT, INSERT, UPDATE ON TABLE pvnaive.customer_bulk_operations TO pvnaive_app;

REVOKE ALL ON FUNCTION pvnaive.validate_customer_profile_refs() FROM PUBLIC;
REVOKE ALL ON FUNCTION pvnaive.validate_customer_profile_refs() FROM pvnaive_app;
REVOKE ALL ON FUNCTION pvnaive.validate_customer_tag_assignment_refs() FROM PUBLIC;
REVOKE ALL ON FUNCTION pvnaive.validate_customer_tag_assignment_refs() FROM pvnaive_app;
REVOKE ALL ON FUNCTION pvnaive.validate_plan_tag_assignment_refs() FROM PUBLIC;
REVOKE ALL ON FUNCTION pvnaive.validate_plan_tag_assignment_refs() FROM pvnaive_app;
REVOKE ALL ON FUNCTION pvnaive.validate_plan_default_group_ref() FROM PUBLIC;
REVOKE ALL ON FUNCTION pvnaive.validate_plan_default_group_ref() FROM pvnaive_app;

COMMENT ON TABLE pvnaive.customer_profiles IS
    'Mutable customer-management metadata. Commercial history remains in service_terms.';
COMMENT ON TABLE pvnaive.customer_bulk_operations IS
    'Idempotent dry-run/execute records. Runtime-sensitive actions require a safe per-customer coordinator.';
