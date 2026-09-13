-- pvnaive:migration-version 0024
-- pvnaive:migration-name notification_preferences
-- pvnaive:transactional true
-- pvnaive:destructive false

-- Per-actor notification preferences.
-- Stores muted event types and delivery preferences per panel actor so the
-- in-app notification list can honour "what this user wants to see".
-- Preferences are a JSON object, e.g.:
--   {"muted_event_types": ["backup_success"], "ui_compact": true}
CREATE TABLE pvnaive.notification_preferences (
    actor_id uuid PRIMARY KEY REFERENCES pvnaive.actors(id) ON DELETE CASCADE,
    tenant_id uuid REFERENCES pvnaive.tenants(id) ON DELETE RESTRICT,
    preferences jsonb NOT NULL DEFAULT '{}'::jsonb CHECK (jsonb_typeof(preferences) = 'object'),
    created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    updated_at timestamptz NOT NULL DEFAULT clock_timestamp()
);

GRANT SELECT, INSERT, UPDATE, DELETE ON pvnaive.notification_preferences TO pvnaive_app;

ALTER TABLE pvnaive.notification_preferences ENABLE ROW LEVEL SECURITY;
ALTER TABLE pvnaive.notification_preferences FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON pvnaive.notification_preferences
    USING (pvnaive.has_tenant_access(tenant_id, false))
    WITH CHECK (pvnaive.has_tenant_access(tenant_id, false));
