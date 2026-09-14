-- pvnaive:migration-version 0025
-- pvnaive:migration-name notification_pipeline
-- pvnaive:transactional true
-- pvnaive:destructive true

DROP FUNCTION IF EXISTS pvnaive.notification_settle(uuid,boolean,boolean,timestamptz,text,timestamptz);
DROP FUNCTION IF EXISTS pvnaive.notification_record_delivery(uuid,text,smallint,text,text,text,timestamptz);
DROP FUNCTION IF EXISTS pvnaive.notification_claim_due(timestamptz,integer);
DROP FUNCTION IF EXISTS pvnaive.notification_enqueue(text,text,uuid,text,jsonb,timestamptz);
DROP FUNCTION IF EXISTS pvnaive.notification_seed_defaults();
DROP FUNCTION IF EXISTS pvnaive.notification_scan_terms(integer);

DELETE FROM pvnaive.schema_migrations WHERE version = 25;
