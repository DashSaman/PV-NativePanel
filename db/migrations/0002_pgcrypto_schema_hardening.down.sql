-- pvnaive:migration-version 0002
-- Source: PVNaive PostgreSQL crypto schema hardening rollback
-- pvnaive:migration-name pgcrypto_schema_hardening
-- pvnaive:transactional true
-- pvnaive:destructive true

SET LOCAL lock_timeout = '10s';
SET LOCAL statement_timeout = '120s';
SET LOCAL ROLE pvnaive_owner;

ALTER EXTENSION pgcrypto SET SCHEMA public;
DROP SCHEMA pvnaive_crypto;
