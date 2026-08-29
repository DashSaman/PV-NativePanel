-- pvnaive:migration-version 0002
-- Source: PVNaive authentication foundation rollback
-- pvnaive:migration-name auth_foundation
-- pvnaive:transactional true
-- pvnaive:destructive true

SET LOCAL lock_timeout = '10s';
SET LOCAL statement_timeout = '120s';
SET LOCAL ROLE pvnaive_owner;

DROP FUNCTION pvnaive.auth_append_audit(uuid, text, text, text, uuid, inet);
DROP POLICY auth_audit_owner_insert ON pvnaive.audit_events;
DROP FUNCTION pvnaive.auth_remove_mfa(uuid);
DROP FUNCTION pvnaive.auth_consume_recovery_code(uuid, bytea);
DROP FUNCTION pvnaive.auth_consume_totp_step(uuid, bigint);
DROP FUNCTION pvnaive.auth_confirm_totp_factor(uuid, bigint, bytea[]);
DROP FUNCTION pvnaive.auth_upsert_totp_factor(uuid, bytea, bytea, text);
DROP FUNCTION pvnaive.auth_get_totp_factor(uuid);
DROP FUNCTION pvnaive.auth_revoke_other_actor_sessions(uuid, uuid);
DROP FUNCTION pvnaive.auth_revoke_session_by_id(uuid, uuid);
DROP FUNCTION pvnaive.auth_revoke_actor_sessions(uuid);
DROP FUNCTION pvnaive.auth_revoke_session(bytea);
DROP FUNCTION pvnaive.auth_rotate_session(bytea, bytea, bytea, bytea, timestamptz);
DROP FUNCTION pvnaive.auth_create_session(uuid, bytea, bytea, uuid, bytea, timestamptz, timestamptz);
DROP FUNCTION pvnaive.auth_record_login_success(uuid);
DROP FUNCTION pvnaive.auth_record_login_failure(uuid);
DROP FUNCTION pvnaive.auth_lookup_actor(text);

DROP TABLE pvnaive.actor_mfa_recovery_codes;
DROP TABLE pvnaive.actor_totp_factors;

ALTER TABLE pvnaive.auth_sessions
    DROP CONSTRAINT auth_sessions_absolute_expiry_chk,
    DROP CONSTRAINT auth_sessions_csrf_hash_len_chk,
    DROP COLUMN absolute_expires_at,
    DROP COLUMN csrf_token_hash;

ALTER TABLE pvnaive.actors
    DROP COLUMN password_changed_at,
    DROP COLUMN locked_until,
    DROP COLUMN failed_login_attempts;

GRANT INSERT, UPDATE, DELETE ON pvnaive.actors TO pvnaive_app;
GRANT INSERT, UPDATE, DELETE ON pvnaive.auth_sessions TO pvnaive_app;

DELETE FROM pvnaive.schema_migrations WHERE version = 2;
