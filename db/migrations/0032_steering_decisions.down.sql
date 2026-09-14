-- pvnaive:migration-version 0032
-- pvnaive:migration-name steering_decisions
-- pvnaive:transactional true
-- pvnaive:destructive true

DROP FUNCTION IF EXISTS pvnaive.steering_state_read(uuid);
DROP FUNCTION IF EXISTS pvnaive.steering_decision_apply(uuid,text,text,bigint,text,timestamptz,jsonb);
DROP TABLE IF EXISTS pvnaive.user_steering_state;
DROP TABLE IF EXISTS pvnaive.steering_decisions;
