-- pvnaive:migration-version 0001
-- Source: PVNaive PostgreSQL schema rollback
-- pvnaive:migration-name initial_schema
-- pvnaive:transactional true
-- pvnaive:destructive true
-- This file is never executed by migrate.sh. rollback.sh requires an explicit
-- destructive confirmation and a verified encrypted backup outside disposable CI databases.

SET LOCAL lock_timeout = '10s';
SET LOCAL statement_timeout = '120s';
SET LOCAL ROLE pvnaive_owner;
DROP SCHEMA pvnaive CASCADE;
DROP EXTENSION IF EXISTS pgcrypto;
