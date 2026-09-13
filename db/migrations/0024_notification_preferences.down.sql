-- pvnaive:migration-version 0024
-- pvnaive:migration-name notification_preferences
-- pvnaive:transactional true
-- pvnaive:destructive true

DROP TABLE IF EXISTS pvnaive.notification_preferences;

DELETE FROM pvnaive.schema_migrations WHERE version = 24;
