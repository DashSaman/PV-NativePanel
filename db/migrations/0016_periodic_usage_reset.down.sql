-- pvnaive:migration-version 0016
-- Source: PVNaive Task #9 restart-safe periodic exact-accounting usage reset rollback
-- pvnaive:migration-name periodic_usage_reset
-- pvnaive:transactional true
-- pvnaive:destructive true

SET LOCAL lock_timeout = '10s';
SET LOCAL statement_timeout = '120s';
SET LOCAL idle_in_transaction_session_timeout = '60s';

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pvnaive.scheduled_usage_reset_attempts)
       OR EXISTS (SELECT 1 FROM pvnaive.direct_naive_accounting_reset_events WHERE reason = 'scheduled') THEN
        RAISE EXCEPTION 'cannot roll back periodic_usage_reset while scheduled reset history exists';
    END IF;
END;
$$;

SET LOCAL ROLE pvnaive_owner;
DROP FUNCTION IF EXISTS pvnaive.execute_due_scheduled_usage_resets(integer);
DROP TRIGGER IF EXISTS service_terms_start_reset_schedule ON pvnaive.service_terms;
DROP TRIGGER IF EXISTS service_terms_init_reset_schedule ON pvnaive.service_terms;
DROP FUNCTION IF EXISTS pvnaive.init_service_term_reset_schedule();
DROP TABLE IF EXISTS pvnaive.scheduled_usage_reset_attempts;
DROP TABLE IF EXISTS pvnaive.service_term_reset_schedules;
DROP FUNCTION IF EXISTS pvnaive.next_usage_reset_due(timestamptz,text,integer);
DELETE FROM pvnaive.actors
 WHERE id = '00000000-0000-0000-0000-000000000016'
   AND email = 'scheduler@pvnaive.invalid';
DELETE FROM pvnaive.schema_migrations WHERE version = 16;
