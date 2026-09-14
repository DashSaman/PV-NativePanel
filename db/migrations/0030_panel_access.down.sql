-- pvnaive:migration-version 0030
-- pvnaive:migration-name panel_access
-- pvnaive:transactional true
-- pvnaive:destructive true

DROP FUNCTION IF EXISTS pvnaive.panel_settings_upsert(text, text, text, integer, integer, text, uuid);
DROP FUNCTION IF EXISTS pvnaive.panel_settings_read();

DROP TABLE IF EXISTS pvnaive.panel_access_audit;
DROP TABLE IF EXISTS pvnaive.panel_settings;

DELETE FROM pvnaive.schema_migrations WHERE version = 30;
