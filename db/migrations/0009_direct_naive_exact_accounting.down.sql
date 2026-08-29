-- pvnaive:migration-version 0009
-- Source: PVNaive exact Direct Naive accounting ledger rollback
-- pvnaive:migration-name direct_naive_exact_accounting
-- pvnaive:transactional true
-- pvnaive:destructive true

SET LOCAL lock_timeout = '10s';
SET LOCAL statement_timeout = '120s';
SET LOCAL idle_in_transaction_session_timeout = '60s';
SET LOCAL ROLE pvnaive_owner;

DROP FUNCTION IF EXISTS pvnaive.direct_naive_accounting_read(uuid,timestamptz,bigint);
DROP FUNCTION IF EXISTS pvnaive.direct_naive_accounting_ingest(uuid,text,text,uuid,uuid,bigint,timestamptz,boolean,bigint,bigint,boolean);
DROP FUNCTION IF EXISTS pvnaive.direct_naive_accounting_claim(uuid,text,uuid,uuid,bigint,text,bigint,timestamptz);
DROP FUNCTION IF EXISTS pvnaive.direct_naive_accounting_authorize(uuid,timestamptz);
DROP FUNCTION IF EXISTS pvnaive.direct_naive_accounting_leave_context();
DROP FUNCTION IF EXISTS pvnaive.direct_naive_accounting_enter_context();
DROP TABLE IF EXISTS pvnaive.direct_naive_accounting_events;
DROP TABLE IF EXISTS pvnaive.direct_naive_accounting_claims;
DROP TABLE IF EXISTS pvnaive.direct_naive_accounting_sessions;
DROP TABLE IF EXISTS pvnaive.direct_naive_accounting_terms;

DELETE FROM pvnaive.schema_migrations WHERE version = 9;
