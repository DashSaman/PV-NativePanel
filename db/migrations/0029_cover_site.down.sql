-- pvnaive:migration-version 0029
-- pvnaive:migration-name cover_site
-- pvnaive:transactional true
-- pvnaive:destructive true

DROP FUNCTION IF EXISTS pvnaive.cover_persona(text);
DROP FUNCTION IF EXISTS pvnaive.cover_set_persona(text, text);
DROP FUNCTION IF EXISTS pvnaive.cover_health(text);
DROP FUNCTION IF EXISTS pvnaive.cover_latest(text, integer);
DROP FUNCTION IF EXISTS pvnaive.cover_replace_snapshot(text, text, jsonb);
DROP FUNCTION IF EXISTS pvnaive.cover_ensure_partition(date);

DROP TABLE IF EXISTS pvnaive.cover_content CASCADE;
DROP TABLE IF EXISTS pvnaive.cover_nodes;

DELETE FROM pvnaive.schema_migrations WHERE version = 29;
