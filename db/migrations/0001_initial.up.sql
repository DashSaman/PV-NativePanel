-- pvnaive:migration-version 0001
-- Source: PVNaive PostgreSQL schema
-- pvnaive:migration-name initial_schema
-- pvnaive:transactional true
-- pvnaive:destructive false

SET LOCAL lock_timeout = '10s';
SET LOCAL statement_timeout = '120s';
SET LOCAL idle_in_transaction_session_timeout = '60s';
SET LOCAL ROLE pvnaive_owner;

CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE SCHEMA pvnaive AUTHORIZATION pvnaive_owner;
REVOKE ALL ON SCHEMA pvnaive FROM PUBLIC;

CREATE TABLE pvnaive.schema_migrations (
    version bigint PRIMARY KEY,
    filename text NOT NULL UNIQUE,
    checksum_sha256 text NOT NULL CHECK (checksum_sha256 ~ '^[0-9a-f]{64}$'),
    destructive boolean NOT NULL DEFAULT false CHECK (destructive = false),
    execution_ms bigint NOT NULL DEFAULT 0 CHECK (execution_ms >= 0),
    applied_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    applied_by text NOT NULL DEFAULT session_user
);

CREATE TABLE pvnaive.tenants (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_type text NOT NULL CHECK (tenant_type IN ('system', 'reseller')),
    slug text NOT NULL CHECK (slug ~ '^[a-z0-9][a-z0-9_-]{1,62}$'),
    display_name text NOT NULL CHECK (length(display_name) BETWEEN 1 AND 160),
    status text NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'suspended', 'closed')),
    created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    updated_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    UNIQUE (id, tenant_type)
);
CREATE UNIQUE INDEX tenants_slug_lower_uidx ON pvnaive.tenants (lower(slug));

CREATE TABLE pvnaive.actors (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id uuid REFERENCES pvnaive.tenants(id) ON DELETE RESTRICT,
    actor_role text NOT NULL CHECK (actor_role IN ('owner', 'admin', 'operator', 'auditor', 'reseller')),
    email text NOT NULL CHECK (length(email) BETWEEN 3 AND 320),
    display_name text NOT NULL CHECK (length(display_name) BETWEEN 1 AND 160),
    password_hash text,
    mfa_required boolean NOT NULL DEFAULT false,
    status text NOT NULL DEFAULT 'active' CHECK (status IN ('pending', 'active', 'locked', 'disabled')),
    last_login_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    updated_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    CHECK ((actor_role = 'reseller' AND tenant_id IS NOT NULL) OR
           (actor_role <> 'reseller' AND tenant_id IS NULL)),
    UNIQUE (id, tenant_id)
);
CREATE UNIQUE INDEX actors_email_lower_uidx ON pvnaive.actors (lower(email));

CREATE TABLE pvnaive.resellers (
    tenant_id uuid PRIMARY KEY REFERENCES pvnaive.tenants(id) ON DELETE RESTRICT,
    primary_actor_id uuid NOT NULL UNIQUE,
    status text NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'suspended', 'closed')),
    currency text NOT NULL CHECK (currency ~ '^[A-Z]{3}$'),
    credit_limit_minor bigint NOT NULL DEFAULT 0 CHECK (credit_limit_minor >= 0),
    low_credit_threshold_minor bigint NOT NULL DEFAULT 0 CHECK (low_credit_threshold_minor >= 0),
    max_users integer CHECK (max_users IS NULL OR max_users > 0),
    max_total_sale_bytes bigint CHECK (max_total_sale_bytes IS NULL OR max_total_sale_bytes > 0),
    price_multiplier numeric(9,4) NOT NULL DEFAULT 1.0000 CHECK (price_multiplier > 0),
    created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    updated_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    FOREIGN KEY (primary_actor_id, tenant_id) REFERENCES pvnaive.actors(id, tenant_id) ON DELETE RESTRICT
);

CREATE TABLE pvnaive.quota_policies (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id uuid NOT NULL REFERENCES pvnaive.tenants(id) ON DELETE RESTRICT,
    name text NOT NULL CHECK (length(name) BETWEEN 1 AND 160),
    total_bytes bigint CHECK (total_bytes IS NULL OR total_bytes > 0),
    reset_mode text NOT NULL CHECK (reset_mode IN ('never', 'calendar', 'interval')),
    reset_interval_seconds bigint CHECK (reset_interval_seconds IS NULL OR reset_interval_seconds > 0),
    timezone text NOT NULL DEFAULT 'UTC' CHECK (length(timezone) BETWEEN 1 AND 80),
    concurrency_limit integer CHECK (concurrency_limit IS NULL OR concurrency_limit > 0),
    device_limit integer CHECK (device_limit IS NULL OR device_limit > 0),
    created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    updated_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    UNIQUE (id, tenant_id),
    UNIQUE (tenant_id, name),
    CHECK ((reset_mode = 'interval' AND reset_interval_seconds IS NOT NULL) OR
           (reset_mode <> 'interval' AND reset_interval_seconds IS NULL))
);

CREATE TABLE pvnaive.plans (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id uuid REFERENCES pvnaive.tenants(id) ON DELETE RESTRICT,
    code text NOT NULL CHECK (code ~ '^[a-z0-9][a-z0-9_-]{1,62}$'),
    name text NOT NULL CHECK (length(name) BETWEEN 1 AND 160),
    status text NOT NULL DEFAULT 'active' CHECK (status IN ('draft', 'active', 'archived')),
    protocol_id text NOT NULL DEFAULT 'naive' CHECK (protocol_id ~ '^[a-z0-9][a-z0-9_-]{1,62}$'),
    quota_bytes bigint CHECK (quota_bytes IS NULL OR quota_bytes > 0),
    duration_seconds bigint NOT NULL CHECK (duration_seconds > 0),
    reset_interval_seconds bigint CHECK (reset_interval_seconds IS NULL OR reset_interval_seconds > 0),
    concurrency_limit integer CHECK (concurrency_limit IS NULL OR concurrency_limit > 0),
    device_limit integer CHECK (device_limit IS NULL OR device_limit > 0),
    base_price_minor bigint NOT NULL CHECK (base_price_minor >= 0),
    currency text NOT NULL CHECK (currency ~ '^[A-Z]{3}$'),
    created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    updated_at timestamptz NOT NULL DEFAULT clock_timestamp()
);
CREATE UNIQUE INDEX plans_owner_code_uidx ON pvnaive.plans (COALESCE(tenant_id, '00000000-0000-0000-0000-000000000000'::uuid), code);

CREATE TABLE pvnaive.reseller_plan_terms (
    tenant_id uuid NOT NULL REFERENCES pvnaive.resellers(tenant_id) ON DELETE RESTRICT,
    plan_id uuid NOT NULL REFERENCES pvnaive.plans(id) ON DELETE RESTRICT,
    allowed boolean NOT NULL DEFAULT true,
    price_minor bigint NOT NULL CHECK (price_minor >= 0),
    max_active_subscriptions integer CHECK (max_active_subscriptions IS NULL OR max_active_subscriptions > 0),
    max_sale_bytes bigint CHECK (max_sale_bytes IS NULL OR max_sale_bytes > 0),
    created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    updated_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    PRIMARY KEY (tenant_id, plan_id)
);

CREATE TABLE pvnaive.users (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id uuid NOT NULL REFERENCES pvnaive.tenants(id) ON DELETE RESTRICT,
    username text NOT NULL CHECK (length(username) BETWEEN 1 AND 128),
    display_name text NOT NULL CHECK (length(display_name) BETWEEN 1 AND 160),
    status text NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'active', 'suspended', 'expired', 'depleted', 'revoked')),
    status_reason_code text,
    created_by_actor_id uuid NOT NULL REFERENCES pvnaive.actors(id) ON DELETE RESTRICT,
    created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    updated_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    UNIQUE (id, tenant_id)
);
CREATE UNIQUE INDEX users_tenant_username_lower_uidx ON pvnaive.users (tenant_id, lower(username));
CREATE INDEX users_tenant_status_idx ON pvnaive.users (tenant_id, status);

CREATE TABLE pvnaive.reseller_credit_ledger (
    sequence_no bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    id uuid NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    tenant_id uuid NOT NULL REFERENCES pvnaive.resellers(tenant_id) ON DELETE RESTRICT,
    delta_minor bigint NOT NULL CHECK (delta_minor <> 0),
    balance_after_minor bigint NOT NULL,
    currency text NOT NULL CHECK (currency ~ '^[A-Z]{3}$'),
    entry_type text NOT NULL CHECK (entry_type IN ('deposit', 'purchase', 'refund', 'manual_adjustment', 'chargeback')),
    reason_code text NOT NULL CHECK (length(reason_code) BETWEEN 1 AND 80),
    idempotency_key text NOT NULL CHECK (length(idempotency_key) BETWEEN 8 AND 160),
    created_by_actor_id uuid NOT NULL REFERENCES pvnaive.actors(id) ON DELETE RESTRICT,
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb CHECK (jsonb_typeof(metadata) = 'object'),
    created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    UNIQUE (id, tenant_id),
    UNIQUE (tenant_id, idempotency_key)
);
CREATE INDEX reseller_credit_ledger_tenant_sequence_idx ON pvnaive.reseller_credit_ledger (tenant_id, sequence_no DESC);

CREATE TABLE pvnaive.purchases (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id uuid NOT NULL REFERENCES pvnaive.resellers(tenant_id) ON DELETE RESTRICT,
    user_id uuid NOT NULL,
    plan_id uuid NOT NULL,
    credit_ledger_entry_id uuid NOT NULL,
    status text NOT NULL DEFAULT 'completed' CHECK (status IN ('pending', 'completed', 'voided', 'refunded')),
    price_minor bigint NOT NULL CHECK (price_minor >= 0),
    currency text NOT NULL CHECK (currency ~ '^[A-Z]{3}$'),
    idempotency_key text NOT NULL CHECK (length(idempotency_key) BETWEEN 8 AND 160),
    purchased_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    created_by_actor_id uuid NOT NULL REFERENCES pvnaive.actors(id) ON DELETE RESTRICT,
    UNIQUE (id, tenant_id),
    UNIQUE (id, tenant_id, plan_id),
    UNIQUE (tenant_id, idempotency_key),
    FOREIGN KEY (user_id, tenant_id) REFERENCES pvnaive.users(id, tenant_id) ON DELETE RESTRICT,
    FOREIGN KEY (tenant_id, plan_id) REFERENCES pvnaive.reseller_plan_terms(tenant_id, plan_id) ON DELETE RESTRICT,
    FOREIGN KEY (credit_ledger_entry_id, tenant_id) REFERENCES pvnaive.reseller_credit_ledger(id, tenant_id) ON DELETE RESTRICT
);

CREATE TABLE pvnaive.subscriptions (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id uuid NOT NULL REFERENCES pvnaive.resellers(tenant_id) ON DELETE RESTRICT,
    user_id uuid NOT NULL,
    plan_id uuid NOT NULL,
    purchase_id uuid NOT NULL,
    quota_policy_id uuid,
    status text NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'suspended', 'expired', 'depleted', 'revoked')),
    purchased_at timestamptz NOT NULL,
    service_started_at timestamptz,
    first_connected_at timestamptz,
    expires_at timestamptz NOT NULL,
    billing_period_started_at timestamptz NOT NULL,
    next_reset_at timestamptz,
    total_bytes bigint CHECK (total_bytes IS NULL OR total_bytes > 0),
    concurrency_limit integer CHECK (concurrency_limit IS NULL OR concurrency_limit > 0),
    device_limit integer CHECK (device_limit IS NULL OR device_limit > 0),
    last_usage_at timestamptz,
    last_synced_at timestamptz,
    revision bigint NOT NULL DEFAULT 1 CHECK (revision > 0),
    created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    updated_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    UNIQUE (id, tenant_id),
    UNIQUE (id, tenant_id, user_id),
    FOREIGN KEY (user_id, tenant_id) REFERENCES pvnaive.users(id, tenant_id) ON DELETE RESTRICT,
    FOREIGN KEY (purchase_id, tenant_id, plan_id) REFERENCES pvnaive.purchases(id, tenant_id, plan_id) ON DELETE RESTRICT,
    FOREIGN KEY (quota_policy_id, tenant_id) REFERENCES pvnaive.quota_policies(id, tenant_id) ON DELETE RESTRICT,
    CHECK (expires_at > purchased_at),
    CHECK (first_connected_at IS NULL OR first_connected_at >= purchased_at)
);
CREATE INDEX subscriptions_tenant_status_expires_idx ON pvnaive.subscriptions (tenant_id, status, expires_at);

CREATE TABLE pvnaive.subscription_tokens (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id uuid NOT NULL REFERENCES pvnaive.resellers(tenant_id) ON DELETE RESTRICT,
    subscription_id uuid NOT NULL,
    token_hash bytea NOT NULL UNIQUE CHECK (octet_length(token_hash) = 32),
    token_prefix text NOT NULL CHECK (length(token_prefix) BETWEEN 6 AND 16),
    status text NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'revoked', 'expired')),
    expires_at timestamptz,
    last_accessed_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    revoked_at timestamptz,
    UNIQUE (id, tenant_id),
    FOREIGN KEY (subscription_id, tenant_id) REFERENCES pvnaive.subscriptions(id, tenant_id) ON DELETE RESTRICT
);

CREATE TABLE pvnaive.credentials (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id uuid NOT NULL REFERENCES pvnaive.resellers(tenant_id) ON DELETE RESTRICT,
    user_id uuid NOT NULL,
    subscription_id uuid NOT NULL,
    protocol_id text NOT NULL CHECK (protocol_id ~ '^[a-z0-9][a-z0-9_-]{1,62}$'),
    external_key text NOT NULL CHECK (length(external_key) BETWEEN 1 AND 190),
    runtime_username text NOT NULL CHECK (length(runtime_username) BETWEEN 1 AND 190),
    secret_hash bytea NOT NULL CHECK (octet_length(secret_hash) >= 32),
    secret_ciphertext bytea NOT NULL CHECK (octet_length(secret_ciphertext) >= 32),
    secret_nonce bytea NOT NULL CHECK (octet_length(secret_nonce) >= 12),
    encryption_key_id text NOT NULL CHECK (length(encryption_key_id) BETWEEN 1 AND 160),
    status text NOT NULL DEFAULT 'active' CHECK (status IN ('staged', 'active', 'rotated', 'revoked')),
    issued_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    rotated_at timestamptz,
    revoked_at timestamptz,
    UNIQUE (id, tenant_id),
    UNIQUE (id, tenant_id, subscription_id),
    UNIQUE (protocol_id, external_key),
    FOREIGN KEY (user_id, tenant_id) REFERENCES pvnaive.users(id, tenant_id) ON DELETE RESTRICT,
    FOREIGN KEY (subscription_id, tenant_id, user_id) REFERENCES pvnaive.subscriptions(id, tenant_id, user_id) ON DELETE RESTRICT
);

CREATE TABLE pvnaive.auth_sessions (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id uuid REFERENCES pvnaive.tenants(id) ON DELETE RESTRICT,
    actor_id uuid NOT NULL REFERENCES pvnaive.actors(id) ON DELETE CASCADE,
    token_hash bytea NOT NULL UNIQUE CHECK (octet_length(token_hash) = 32),
    refresh_family_id uuid NOT NULL,
    user_agent_hash bytea CHECK (user_agent_hash IS NULL OR octet_length(user_agent_hash) = 32),
    expires_at timestamptz NOT NULL,
    last_seen_at timestamptz,
    revoked_at timestamptz,
    reuse_detected_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    UNIQUE (id, tenant_id)
);
CREATE INDEX auth_sessions_actor_active_idx ON pvnaive.auth_sessions (actor_id, expires_at) WHERE revoked_at IS NULL;

CREATE TABLE pvnaive.usage_reset_events (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id uuid NOT NULL REFERENCES pvnaive.resellers(tenant_id) ON DELETE RESTRICT,
    subscription_id uuid NOT NULL,
    reset_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    reason text NOT NULL CHECK (reason IN ('scheduled', 'manual', 'renewal', 'correction')),
    previous_used_bytes bigint NOT NULL CHECK (previous_used_bytes >= 0),
    new_period_started_at timestamptz NOT NULL,
    idempotency_key text NOT NULL CHECK (length(idempotency_key) BETWEEN 8 AND 160),
    created_by_actor_id uuid REFERENCES pvnaive.actors(id) ON DELETE RESTRICT,
    UNIQUE (id, tenant_id),
    UNIQUE (tenant_id, idempotency_key),
    FOREIGN KEY (subscription_id, tenant_id) REFERENCES pvnaive.subscriptions(id, tenant_id) ON DELETE RESTRICT
);

CREATE TABLE pvnaive.renewal_events (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id uuid NOT NULL REFERENCES pvnaive.resellers(tenant_id) ON DELETE RESTRICT,
    subscription_id uuid NOT NULL,
    purchase_id uuid NOT NULL,
    previous_expires_at timestamptz NOT NULL,
    new_expires_at timestamptz NOT NULL,
    previous_total_bytes bigint,
    new_total_bytes bigint,
    idempotency_key text NOT NULL CHECK (length(idempotency_key) BETWEEN 8 AND 160),
    created_by_actor_id uuid NOT NULL REFERENCES pvnaive.actors(id) ON DELETE RESTRICT,
    created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    UNIQUE (id, tenant_id),
    UNIQUE (tenant_id, idempotency_key),
    FOREIGN KEY (subscription_id, tenant_id) REFERENCES pvnaive.subscriptions(id, tenant_id) ON DELETE RESTRICT,
    FOREIGN KEY (purchase_id, tenant_id) REFERENCES pvnaive.purchases(id, tenant_id) ON DELETE RESTRICT,
    CHECK (new_expires_at > previous_expires_at)
);

CREATE TABLE pvnaive.sessions (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id uuid NOT NULL REFERENCES pvnaive.resellers(tenant_id) ON DELETE RESTRICT,
    subscription_id uuid NOT NULL,
    credential_id uuid NOT NULL,
    runtime_session_key text NOT NULL CHECK (length(runtime_session_key) BETWEEN 1 AND 190),
    connected_at timestamptz NOT NULL,
    disconnected_at timestamptz,
    last_seen_at timestamptz NOT NULL,
    upload_bytes bigint NOT NULL DEFAULT 0 CHECK (upload_bytes >= 0),
    download_bytes bigint NOT NULL DEFAULT 0 CHECK (download_bytes >= 0),
    close_reason_code text,
    client_metadata jsonb NOT NULL DEFAULT '{}'::jsonb CHECK (jsonb_typeof(client_metadata) = 'object'),
    UNIQUE (id, tenant_id),
    UNIQUE (credential_id, runtime_session_key),
    FOREIGN KEY (subscription_id, tenant_id) REFERENCES pvnaive.subscriptions(id, tenant_id) ON DELETE RESTRICT,
    FOREIGN KEY (credential_id, tenant_id, subscription_id) REFERENCES pvnaive.credentials(id, tenant_id, subscription_id) ON DELETE RESTRICT,
    CHECK (disconnected_at IS NULL OR disconnected_at >= connected_at)
);
CREATE INDEX sessions_tenant_live_idx ON pvnaive.sessions (tenant_id, last_seen_at DESC) WHERE disconnected_at IS NULL;

CREATE TABLE pvnaive.usage_ledger (
    sequence_no bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    id uuid NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    tenant_id uuid NOT NULL REFERENCES pvnaive.resellers(tenant_id) ON DELETE RESTRICT,
    subscription_id uuid NOT NULL,
    credential_id uuid NOT NULL,
    session_id uuid,
    runtime_boot_id uuid NOT NULL,
    source_sequence bigint NOT NULL CHECK (source_sequence >= 0),
    direction text NOT NULL CHECK (direction IN ('upload', 'download')),
    bytes_delta bigint NOT NULL CHECK (bytes_delta > 0),
    observed_at timestamptz NOT NULL,
    ingested_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    source text NOT NULL CHECK (source IN ('runtime', 'reconciliation', 'correction')),
    UNIQUE (id, tenant_id),
    UNIQUE (credential_id, runtime_boot_id, source_sequence, direction),
    FOREIGN KEY (subscription_id, tenant_id) REFERENCES pvnaive.subscriptions(id, tenant_id) ON DELETE RESTRICT,
    FOREIGN KEY (credential_id, tenant_id, subscription_id) REFERENCES pvnaive.credentials(id, tenant_id, subscription_id) ON DELETE RESTRICT,
    FOREIGN KEY (session_id, tenant_id) REFERENCES pvnaive.sessions(id, tenant_id) ON DELETE RESTRICT
);
CREATE INDEX usage_ledger_subscription_observed_idx ON pvnaive.usage_ledger (tenant_id, subscription_id, observed_at);

CREATE TABLE pvnaive.notification_rules (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id uuid REFERENCES pvnaive.tenants(id) ON DELETE RESTRICT,
    event_type text NOT NULL CHECK (event_type IN (
        'usage_80', 'usage_95', 'usage_exhausted', 'expiry_7d', 'expiry_3d', 'expiry_1d',
        'expired', 'tls_expiring', 'runtime_down', 'backup_failed', 'reseller_credit_low'
    )),
    enabled boolean NOT NULL DEFAULT true,
    channels jsonb NOT NULL DEFAULT '["in_app"]'::jsonb CHECK (jsonb_typeof(channels) = 'array'),
    max_attempts smallint NOT NULL DEFAULT 5 CHECK (max_attempts BETWEEN 1 AND 20),
    retry_base_seconds integer NOT NULL DEFAULT 60 CHECK (retry_base_seconds BETWEEN 1 AND 86400),
    created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    updated_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    UNIQUE (id, tenant_id)
);
CREATE UNIQUE INDEX notification_rules_scope_event_uidx ON pvnaive.notification_rules (
    COALESCE(tenant_id, '00000000-0000-0000-0000-000000000000'::uuid), event_type
);

CREATE TABLE pvnaive.notification_outbox (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id uuid REFERENCES pvnaive.tenants(id) ON DELETE RESTRICT,
    rule_id uuid REFERENCES pvnaive.notification_rules(id) ON DELETE RESTRICT,
    event_type text NOT NULL,
    aggregate_type text NOT NULL CHECK (length(aggregate_type) BETWEEN 1 AND 80),
    aggregate_id uuid,
    deduplication_key text NOT NULL CHECK (length(deduplication_key) BETWEEN 8 AND 190),
    payload jsonb NOT NULL CHECK (jsonb_typeof(payload) = 'object'),
    status text NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'processing', 'delivered', 'failed', 'cancelled')),
    attempt_count smallint NOT NULL DEFAULT 0 CHECK (attempt_count >= 0),
    max_attempts smallint NOT NULL DEFAULT 5 CHECK (max_attempts BETWEEN 1 AND 20),
    next_attempt_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    locked_at timestamptz,
    locked_by text,
    last_error_code text,
    created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    processed_at timestamptz,
    UNIQUE (id, tenant_id),
    CHECK (attempt_count <= max_attempts)
);
CREATE UNIQUE INDEX notification_outbox_dedup_uidx ON pvnaive.notification_outbox (
    COALESCE(tenant_id, '00000000-0000-0000-0000-000000000000'::uuid), deduplication_key
);
CREATE INDEX notification_outbox_ready_idx ON pvnaive.notification_outbox (next_attempt_at, created_at)
    WHERE status IN ('pending', 'failed');

CREATE TABLE pvnaive.notification_deliveries (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id uuid REFERENCES pvnaive.tenants(id) ON DELETE RESTRICT,
    outbox_id uuid NOT NULL REFERENCES pvnaive.notification_outbox(id) ON DELETE RESTRICT,
    channel text NOT NULL CHECK (channel IN ('in_app', 'webhook', 'email', 'telegram')),
    attempt_no smallint NOT NULL CHECK (attempt_no > 0),
    idempotency_key text NOT NULL UNIQUE CHECK (length(idempotency_key) BETWEEN 8 AND 190),
    status text NOT NULL CHECK (status IN ('delivered', 'retryable_failure', 'permanent_failure')),
    provider_message_ref text,
    error_code text,
    attempted_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    delivered_at timestamptz,
    UNIQUE (outbox_id, channel, attempt_no),
    UNIQUE (id, tenant_id)
);

CREATE TABLE pvnaive.runtime_revisions (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id uuid REFERENCES pvnaive.tenants(id) ON DELETE RESTRICT,
    protocol_id text NOT NULL CHECK (protocol_id ~ '^[a-z0-9][a-z0-9_-]{1,62}$'),
    revision_no bigint NOT NULL CHECK (revision_no > 0),
    state text NOT NULL CHECK (state IN ('staged', 'validated', 'applied', 'failed', 'rolled_back')),
    config_checksum_sha256 text NOT NULL CHECK (config_checksum_sha256 ~ '^[0-9a-f]{64}$'),
    config_ciphertext bytea NOT NULL,
    encryption_key_id text NOT NULL CHECK (length(encryption_key_id) BETWEEN 1 AND 160),
    manifest jsonb NOT NULL DEFAULT '{}'::jsonb CHECK (jsonb_typeof(manifest) = 'object'),
    previous_revision_id uuid REFERENCES pvnaive.runtime_revisions(id) ON DELETE RESTRICT,
    created_by_actor_id uuid NOT NULL REFERENCES pvnaive.actors(id) ON DELETE RESTRICT,
    created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    validated_at timestamptz,
    applied_at timestamptz,
    failure_code text,
    UNIQUE (id, tenant_id)
);
CREATE UNIQUE INDEX runtime_revisions_scope_no_uidx ON pvnaive.runtime_revisions (
    COALESCE(tenant_id, '00000000-0000-0000-0000-000000000000'::uuid), protocol_id, revision_no
);

CREATE TABLE pvnaive.runtime_health (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id uuid REFERENCES pvnaive.tenants(id) ON DELETE RESTRICT,
    protocol_id text NOT NULL,
    runtime_boot_id uuid,
    runtime_revision_id uuid REFERENCES pvnaive.runtime_revisions(id) ON DELETE SET NULL,
    health_status text NOT NULL CHECK (health_status IN ('healthy', 'degraded', 'down', 'unknown')),
    latency_ms integer CHECK (latency_ms IS NULL OR latency_ms >= 0),
    error_code text,
    details jsonb NOT NULL DEFAULT '{}'::jsonb CHECK (jsonb_typeof(details) = 'object'),
    observed_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    UNIQUE (id, tenant_id)
);
CREATE INDEX runtime_health_latest_idx ON pvnaive.runtime_health (protocol_id, observed_at DESC);

CREATE TABLE pvnaive.audit_events (
    sequence_no bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    id uuid NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    tenant_id uuid REFERENCES pvnaive.tenants(id) ON DELETE RESTRICT,
    actor_id uuid REFERENCES pvnaive.actors(id) ON DELETE SET NULL,
    action text NOT NULL CHECK (length(action) BETWEEN 1 AND 120),
    object_type text NOT NULL CHECK (length(object_type) BETWEEN 1 AND 80),
    object_id uuid,
    request_id uuid,
    source_ip inet,
    outcome text NOT NULL CHECK (outcome IN ('success', 'denied', 'failure')),
    reason_code text,
    before_state jsonb CHECK (before_state IS NULL OR jsonb_typeof(before_state) = 'object'),
    after_state jsonb CHECK (after_state IS NULL OR jsonb_typeof(after_state) = 'object'),
    occurred_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    UNIQUE (id, tenant_id)
);
CREATE INDEX audit_events_scope_time_idx ON pvnaive.audit_events (tenant_id, occurred_at DESC);

CREATE TABLE pvnaive.log_metadata (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id uuid REFERENCES pvnaive.tenants(id) ON DELETE RESTRICT,
    log_class text NOT NULL CHECK (log_class IN ('application', 'runtime', 'security')),
    severity text NOT NULL CHECK (severity IN ('debug', 'info', 'warning', 'error', 'critical')),
    event_code text NOT NULL CHECK (length(event_code) BETWEEN 1 AND 120),
    request_id uuid,
    actor_id uuid REFERENCES pvnaive.actors(id) ON DELETE SET NULL,
    user_id uuid REFERENCES pvnaive.users(id) ON DELETE SET NULL,
    source_component text NOT NULL CHECK (length(source_component) BETWEEN 1 AND 80),
    fields jsonb NOT NULL DEFAULT '{}'::jsonb CHECK (
        jsonb_typeof(fields) = 'object' AND
        NOT (fields ?| ARRAY['password', 'secret', 'token', 'authorization', 'cookie', 'path', 'query'])
    ),
    occurred_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    expires_at timestamptz,
    UNIQUE (id, tenant_id)
);
CREATE INDEX log_metadata_scope_time_idx ON pvnaive.log_metadata (tenant_id, log_class, occurred_at DESC);

CREATE TABLE pvnaive.backups (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id uuid REFERENCES pvnaive.tenants(id) ON DELETE RESTRICT,
    backup_type text NOT NULL CHECK (backup_type IN ('database', 'configuration', 'full')),
    status text NOT NULL CHECK (status IN ('running', 'completed', 'failed', 'verified', 'restore_tested')),
    storage_ref text NOT NULL CHECK (length(storage_ref) BETWEEN 1 AND 500),
    encrypted boolean NOT NULL CHECK (encrypted = true),
    encryption_key_id text NOT NULL CHECK (length(encryption_key_id) BETWEEN 1 AND 160),
    checksum_sha256 text CHECK (checksum_sha256 IS NULL OR checksum_sha256 ~ '^[0-9a-f]{64}$'),
    size_bytes bigint CHECK (size_bytes IS NULL OR size_bytes >= 0),
    database_server_version text,
    schema_version bigint,
    started_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    completed_at timestamptz,
    verified_at timestamptz,
    restore_tested_at timestamptz,
    failure_code text,
    UNIQUE (id, tenant_id)
);

CREATE TABLE pvnaive.security_context_keys (
    singleton boolean PRIMARY KEY DEFAULT true CHECK (singleton),
    signing_key bytea NOT NULL CHECK (octet_length(signing_key) = 32),
    created_at timestamptz NOT NULL DEFAULT clock_timestamp()
);
INSERT INTO pvnaive.security_context_keys (singleton, signing_key) VALUES (true, gen_random_bytes(32));

CREATE FUNCTION pvnaive.prevent_immutable_mutation()
RETURNS trigger
LANGUAGE plpgsql
SET search_path = pg_catalog, pvnaive
AS $$
BEGIN
    RAISE EXCEPTION 'PVNaive append-only relation % does not permit %', TG_TABLE_NAME, TG_OP
        USING ERRCODE = '55000';
END;
$$;

CREATE FUNCTION pvnaive.validate_credit_ledger_insert()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog, pvnaive
AS $$
DECLARE
    expected_currency text;
    credit_limit bigint;
    previous_balance bigint;
BEGIN
    SELECT r.currency, r.credit_limit_minor
      INTO expected_currency, credit_limit
      FROM pvnaive.resellers AS r
     WHERE r.tenant_id = NEW.tenant_id
     FOR UPDATE;

    IF NOT FOUND THEN
        RAISE EXCEPTION 'unknown reseller tenant' USING ERRCODE = '23503';
    END IF;
    IF NEW.currency <> expected_currency THEN
        RAISE EXCEPTION 'credit ledger currency mismatch' USING ERRCODE = '23514';
    END IF;

    SELECT COALESCE((
        SELECT l.balance_after_minor
          FROM pvnaive.reseller_credit_ledger AS l
         WHERE l.tenant_id = NEW.tenant_id
         ORDER BY l.sequence_no DESC
         LIMIT 1
    ), 0) INTO previous_balance;

    IF NEW.balance_after_minor <> previous_balance + NEW.delta_minor THEN
        RAISE EXCEPTION 'credit ledger balance mismatch' USING ERRCODE = '23514';
    END IF;
    IF NEW.balance_after_minor < -credit_limit THEN
        RAISE EXCEPTION 'reseller credit limit exceeded' USING ERRCODE = '23514';
    END IF;
    RETURN NEW;
END;
$$;

CREATE TRIGGER reseller_credit_validate_before_insert
BEFORE INSERT ON pvnaive.reseller_credit_ledger
FOR EACH ROW EXECUTE FUNCTION pvnaive.validate_credit_ledger_insert();

CREATE FUNCTION pvnaive.validate_purchase_plan_scope()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog, pvnaive
AS $$
DECLARE
    term_allowed boolean;
    plan_tenant_id uuid;
BEGIN
    SELECT terms.allowed, plans.tenant_id
      INTO term_allowed, plan_tenant_id
      FROM pvnaive.reseller_plan_terms AS terms
      JOIN pvnaive.plans AS plans ON plans.id = terms.plan_id
     WHERE terms.tenant_id = NEW.tenant_id
       AND terms.plan_id = NEW.plan_id
       FOR KEY SHARE OF terms, plans;

    IF NOT FOUND OR NOT term_allowed THEN
        RAISE EXCEPTION 'plan is not allowed for reseller tenant' USING ERRCODE = '23514';
    END IF;
    IF plan_tenant_id IS NOT NULL AND plan_tenant_id <> NEW.tenant_id THEN
        RAISE EXCEPTION 'cross-tenant plan reference refused' USING ERRCODE = '23514';
    END IF;
    RETURN NEW;
END;
$$;

CREATE TRIGGER purchases_validate_plan_before_write
BEFORE INSERT OR UPDATE OF tenant_id, plan_id ON pvnaive.purchases
FOR EACH ROW EXECUTE FUNCTION pvnaive.validate_purchase_plan_scope();

CREATE FUNCTION pvnaive.validate_auth_session_scope()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog, pvnaive
AS $$
DECLARE
    actor_tenant_id uuid;
BEGIN
    SELECT tenant_id INTO actor_tenant_id
      FROM pvnaive.actors
     WHERE id = NEW.actor_id
       FOR KEY SHARE;
    IF NOT FOUND OR actor_tenant_id IS DISTINCT FROM NEW.tenant_id THEN
        RAISE EXCEPTION 'session tenant does not match actor tenant' USING ERRCODE = '23514';
    END IF;
    RETURN NEW;
END;
$$;

CREATE TRIGGER auth_sessions_validate_scope_before_write
BEFORE INSERT OR UPDATE OF tenant_id, actor_id ON pvnaive.auth_sessions
FOR EACH ROW EXECUTE FUNCTION pvnaive.validate_auth_session_scope();

DO $$
DECLARE
    relation_name text;
BEGIN
    FOREACH relation_name IN ARRAY ARRAY[
        'reseller_credit_ledger', 'usage_ledger', 'usage_reset_events',
        'renewal_events', 'audit_events'
    ] LOOP
        EXECUTE format(
            'CREATE TRIGGER %I_immutable BEFORE UPDATE OR DELETE ON pvnaive.%I FOR EACH ROW EXECUTE FUNCTION pvnaive.prevent_immutable_mutation()',
            relation_name, relation_name
        );
    END LOOP;
END;
$$;

CREATE FUNCTION pvnaive.context_payload(p_actor_id uuid, p_tenant_id uuid, p_actor_role text)
RETURNS bytea
LANGUAGE sql
IMMUTABLE
PARALLEL SAFE
AS $$
    SELECT convert_to(
        p_actor_id::text || '|' || COALESCE(p_tenant_id::text, '') || '|' || p_actor_role,
        'UTF8'
    );
$$;

CREATE FUNCTION pvnaive.has_valid_context()
RETURNS boolean
LANGUAGE plpgsql
STABLE
SECURITY DEFINER
SET search_path = pg_catalog, pvnaive
AS $$
DECLARE
    actor_id_value uuid;
    tenant_id_value uuid;
    actor_role_value text;
    provided_signature text;
    expected_signature text;
    signing_key_value bytea;
BEGIN
    actor_id_value := NULLIF(current_setting('pvnaive.actor_id', true), '')::uuid;
    tenant_id_value := NULLIF(current_setting('pvnaive.tenant_id', true), '')::uuid;
    actor_role_value := NULLIF(current_setting('pvnaive.actor_role', true), '');
    provided_signature := NULLIF(current_setting('pvnaive.context_signature', true), '');
    IF actor_id_value IS NULL OR actor_role_value IS NULL OR provided_signature IS NULL THEN
        RETURN false;
    END IF;

    SELECT signing_key INTO signing_key_value FROM pvnaive.security_context_keys WHERE singleton;
    expected_signature := encode(
        hmac(pvnaive.context_payload(actor_id_value, tenant_id_value, actor_role_value), signing_key_value, 'sha256'),
        'hex'
    );
    RETURN expected_signature = provided_signature;
EXCEPTION WHEN OTHERS THEN
    RETURN false;
END;
$$;

CREATE FUNCTION pvnaive.current_actor_id()
RETURNS uuid LANGUAGE sql STABLE AS $$
    SELECT CASE WHEN pvnaive.has_valid_context()
                THEN NULLIF(current_setting('pvnaive.actor_id', true), '')::uuid END;
$$;
CREATE FUNCTION pvnaive.current_tenant_id()
RETURNS uuid LANGUAGE sql STABLE AS $$
    SELECT CASE WHEN pvnaive.has_valid_context()
                THEN NULLIF(current_setting('pvnaive.tenant_id', true), '')::uuid END;
$$;
CREATE FUNCTION pvnaive.current_actor_role()
RETURNS text LANGUAGE sql STABLE AS $$
    SELECT CASE WHEN pvnaive.has_valid_context()
                THEN NULLIF(current_setting('pvnaive.actor_role', true), '') END;
$$;

CREATE FUNCTION pvnaive.has_tenant_access(p_tenant_id uuid, p_allow_global boolean DEFAULT false)
RETURNS boolean
LANGUAGE sql
STABLE
AS $$
    SELECT pvnaive.has_valid_context() AND (
        pvnaive.current_actor_role() IN ('owner', 'admin', 'operator', 'auditor') OR
        (p_tenant_id IS NULL AND p_allow_global) OR
        p_tenant_id = pvnaive.current_tenant_id()
    );
$$;

CREATE FUNCTION pvnaive.set_request_context(p_session_token_hash bytea)
RETURNS TABLE (actor_id uuid, tenant_id uuid, actor_role text)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog, pvnaive
AS $$
DECLARE
    selected_actor pvnaive.actors%ROWTYPE;
    signing_key_value bytea;
    signature_value text;
BEGIN
    IF p_session_token_hash IS NULL OR octet_length(p_session_token_hash) <> 32 THEN
        RAISE EXCEPTION 'invalid session context' USING ERRCODE = '28000';
    END IF;

    SELECT a.* INTO selected_actor
     FROM pvnaive.auth_sessions AS s
      JOIN pvnaive.actors AS a ON a.id = s.actor_id
     WHERE s.token_hash = p_session_token_hash
       AND s.tenant_id IS NOT DISTINCT FROM a.tenant_id
       AND s.revoked_at IS NULL
       AND s.expires_at > clock_timestamp()
       AND a.status = 'active';
    IF NOT FOUND THEN
        RAISE EXCEPTION 'invalid session context' USING ERRCODE = '28000';
    END IF;

    SELECT signing_key INTO signing_key_value FROM pvnaive.security_context_keys WHERE singleton;
    signature_value := encode(
        hmac(
            pvnaive.context_payload(selected_actor.id, selected_actor.tenant_id, selected_actor.actor_role),
            signing_key_value,
            'sha256'
        ),
        'hex'
    );

    PERFORM set_config('pvnaive.actor_id', selected_actor.id::text, true);
    PERFORM set_config('pvnaive.tenant_id', COALESCE(selected_actor.tenant_id::text, ''), true);
    PERFORM set_config('pvnaive.actor_role', selected_actor.actor_role, true);
    PERFORM set_config('pvnaive.context_signature', signature_value, true);

    RETURN QUERY SELECT selected_actor.id, selected_actor.tenant_id, selected_actor.actor_role;
END;
$$;

CREATE FUNCTION pvnaive.clear_request_context()
RETURNS void
LANGUAGE plpgsql
AS $$
BEGIN
    PERFORM set_config('pvnaive.actor_id', '', true);
    PERFORM set_config('pvnaive.tenant_id', '', true);
    PERFORM set_config('pvnaive.actor_role', '', true);
    PERFORM set_config('pvnaive.context_signature', '', true);
END;
$$;

ALTER TABLE pvnaive.tenants ENABLE ROW LEVEL SECURITY;
ALTER TABLE pvnaive.tenants FORCE ROW LEVEL SECURITY;
CREATE POLICY tenants_isolation ON pvnaive.tenants
    USING (pvnaive.has_tenant_access(id, false))
    WITH CHECK (pvnaive.has_tenant_access(id, false));

DO $$
DECLARE
    relation_name text;
BEGIN
    FOREACH relation_name IN ARRAY ARRAY[
        'resellers', 'quota_policies', 'reseller_plan_terms', 'users',
        'reseller_credit_ledger', 'purchases', 'subscriptions', 'subscription_tokens',
        'credentials', 'usage_reset_events', 'renewal_events', 'sessions', 'usage_ledger'
    ] LOOP
        EXECUTE format('ALTER TABLE pvnaive.%I ENABLE ROW LEVEL SECURITY', relation_name);
        EXECUTE format('ALTER TABLE pvnaive.%I FORCE ROW LEVEL SECURITY', relation_name);
        EXECUTE format(
            'CREATE POLICY tenant_isolation ON pvnaive.%I USING (pvnaive.has_tenant_access(tenant_id, false)) WITH CHECK (pvnaive.has_tenant_access(tenant_id, false))',
            relation_name
        );
    END LOOP;

    FOREACH relation_name IN ARRAY ARRAY[
        'plans'
    ] LOOP
        EXECUTE format('ALTER TABLE pvnaive.%I ENABLE ROW LEVEL SECURITY', relation_name);
        EXECUTE format('ALTER TABLE pvnaive.%I FORCE ROW LEVEL SECURITY', relation_name);
        EXECUTE format(
            'CREATE POLICY tenant_isolation ON pvnaive.%I USING (pvnaive.has_tenant_access(tenant_id, true)) WITH CHECK (pvnaive.has_tenant_access(tenant_id, false))',
            relation_name
        );
    END LOOP;

    FOREACH relation_name IN ARRAY ARRAY[
        'notification_rules', 'notification_outbox',
        'notification_deliveries', 'runtime_revisions', 'runtime_health',
        'audit_events', 'log_metadata', 'backups'
    ] LOOP
        EXECUTE format('ALTER TABLE pvnaive.%I ENABLE ROW LEVEL SECURITY', relation_name);
        EXECUTE format('ALTER TABLE pvnaive.%I FORCE ROW LEVEL SECURITY', relation_name);
        EXECUTE format(
            'CREATE POLICY tenant_isolation ON pvnaive.%I USING (pvnaive.has_tenant_access(tenant_id, false)) WITH CHECK (pvnaive.has_tenant_access(tenant_id, false))',
            relation_name
        );
    END LOOP;
END;
$$;

-- These two tables deliberately omit FORCE ROW LEVEL SECURITY: the NOLOGIN
-- owner executes set_request_context() as SECURITY DEFINER before a context
-- exists. The application role is still always subject to both policies.
ALTER TABLE pvnaive.actors ENABLE ROW LEVEL SECURITY;
CREATE POLICY actors_isolation ON pvnaive.actors
    USING (pvnaive.has_tenant_access(tenant_id, false))
    WITH CHECK (pvnaive.has_tenant_access(tenant_id, false));

ALTER TABLE pvnaive.auth_sessions ENABLE ROW LEVEL SECURITY;
CREATE POLICY auth_sessions_isolation ON pvnaive.auth_sessions
    USING (pvnaive.has_tenant_access(tenant_id, false))
    WITH CHECK (pvnaive.has_tenant_access(tenant_id, false));

CREATE VIEW pvnaive.subscription_usage
WITH (security_invoker = true)
AS
SELECT s.id AS subscription_id,
       s.tenant_id,
       COALESCE(SUM(l.bytes_delta) FILTER (WHERE l.direction = 'upload'), 0)::bigint AS upload_bytes,
       COALESCE(SUM(l.bytes_delta) FILTER (WHERE l.direction = 'download'), 0)::bigint AS download_bytes,
       COALESCE(SUM(l.bytes_delta), 0)::bigint AS used_bytes,
       MAX(l.ingested_at) AS last_updated_at
  FROM pvnaive.subscriptions AS s
  LEFT JOIN pvnaive.usage_ledger AS l
    ON l.subscription_id = s.id
   AND l.tenant_id = s.tenant_id
   AND l.observed_at >= s.billing_period_started_at
 GROUP BY s.id, s.tenant_id;

CREATE VIEW pvnaive.subscription_summary
WITH (security_invoker = true)
AS
SELECT s.id AS subscription_id,
       s.tenant_id,
       s.user_id,
       s.status,
       s.purchased_at,
       s.first_connected_at,
       s.expires_at,
       s.next_reset_at,
       s.total_bytes,
       u.used_bytes,
       CASE WHEN s.total_bytes IS NULL THEN NULL ELSE GREATEST(s.total_bytes - u.used_bytes, 0) END AS remaining_bytes,
       CASE WHEN s.total_bytes IS NULL THEN NULL
            ELSE LEAST(100.00, ROUND((u.used_bytes::numeric * 100) / s.total_bytes, 2)) END AS usage_percent,
       s.concurrency_limit,
       s.device_limit,
       COALESCE(u.last_updated_at, s.updated_at) AS last_updated_at
  FROM pvnaive.subscriptions AS s
  JOIN pvnaive.subscription_usage AS u
    ON u.subscription_id = s.id AND u.tenant_id = s.tenant_id;

REVOKE ALL ON ALL TABLES IN SCHEMA pvnaive FROM PUBLIC;
REVOKE ALL ON ALL SEQUENCES IN SCHEMA pvnaive FROM PUBLIC;
REVOKE ALL ON ALL FUNCTIONS IN SCHEMA pvnaive FROM PUBLIC;

GRANT USAGE ON SCHEMA pvnaive TO pvnaive_app;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA pvnaive TO pvnaive_app;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA pvnaive TO pvnaive_app;
GRANT EXECUTE ON FUNCTION pvnaive.set_request_context(bytea) TO pvnaive_app;
GRANT EXECUTE ON FUNCTION pvnaive.clear_request_context() TO pvnaive_app;
GRANT EXECUTE ON FUNCTION pvnaive.has_valid_context() TO pvnaive_app;
GRANT EXECUTE ON FUNCTION pvnaive.has_tenant_access(uuid, boolean) TO pvnaive_app;
GRANT EXECUTE ON FUNCTION pvnaive.current_actor_id() TO pvnaive_app;
GRANT EXECUTE ON FUNCTION pvnaive.current_tenant_id() TO pvnaive_app;
GRANT EXECUTE ON FUNCTION pvnaive.current_actor_role() TO pvnaive_app;

REVOKE ALL ON pvnaive.security_context_keys FROM pvnaive_app;
REVOKE INSERT, UPDATE, DELETE ON pvnaive.schema_migrations FROM pvnaive_app;
REVOKE UPDATE, DELETE ON pvnaive.reseller_credit_ledger FROM pvnaive_app;
REVOKE UPDATE, DELETE ON pvnaive.usage_ledger FROM pvnaive_app;
REVOKE UPDATE, DELETE ON pvnaive.usage_reset_events FROM pvnaive_app;
REVOKE UPDATE, DELETE ON pvnaive.renewal_events FROM pvnaive_app;
REVOKE UPDATE, DELETE ON pvnaive.audit_events FROM pvnaive_app;

ALTER DEFAULT PRIVILEGES FOR ROLE pvnaive_owner IN SCHEMA pvnaive REVOKE ALL ON TABLES FROM PUBLIC;
ALTER DEFAULT PRIVILEGES FOR ROLE pvnaive_owner IN SCHEMA pvnaive REVOKE ALL ON SEQUENCES FROM PUBLIC;
ALTER DEFAULT PRIVILEGES FOR ROLE pvnaive_owner IN SCHEMA pvnaive REVOKE ALL ON FUNCTIONS FROM PUBLIC;

COMMENT ON SCHEMA pvnaive IS 'PVNaive standalone application schema';
COMMENT ON TABLE pvnaive.reseller_credit_ledger IS 'Append-only, serialized reseller credit ledger';
COMMENT ON TABLE pvnaive.usage_ledger IS 'Append-only idempotent per-credential usage delta ledger';
COMMENT ON FUNCTION pvnaive.set_request_context(bytea) IS 'Binds a transaction-local, signed RLS context from an active session token hash';
