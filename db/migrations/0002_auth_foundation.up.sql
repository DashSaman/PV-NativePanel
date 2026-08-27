-- pvnaive:migration-version 0002
-- Source: PVNaive authentication foundation
-- pvnaive:migration-name auth_foundation
-- pvnaive:transactional true
-- pvnaive:destructive false

SET LOCAL lock_timeout = '10s';
SET LOCAL statement_timeout = '120s';
SET LOCAL idle_in_transaction_session_timeout = '60s';
SET LOCAL ROLE pvnaive_owner;

ALTER TABLE pvnaive.actors
    ADD COLUMN failed_login_attempts smallint NOT NULL DEFAULT 0
        CHECK (failed_login_attempts BETWEEN 0 AND 32767),
    ADD COLUMN locked_until timestamptz,
    ADD COLUMN password_changed_at timestamptz;

ALTER TABLE pvnaive.auth_sessions
    ADD COLUMN csrf_token_hash bytea,
    ADD COLUMN absolute_expires_at timestamptz;

UPDATE pvnaive.auth_sessions
   SET csrf_token_hash = public.gen_random_bytes(32),
       absolute_expires_at = GREATEST(expires_at, created_at + interval '12 hours')
 WHERE csrf_token_hash IS NULL
    OR absolute_expires_at IS NULL;

ALTER TABLE pvnaive.auth_sessions
    ALTER COLUMN csrf_token_hash SET NOT NULL,
    ALTER COLUMN csrf_token_hash SET DEFAULT public.gen_random_bytes(32),
    ALTER COLUMN absolute_expires_at SET NOT NULL,
    ALTER COLUMN absolute_expires_at SET DEFAULT (clock_timestamp() + interval '12 hours'),
    ADD CONSTRAINT auth_sessions_csrf_hash_len_chk
        CHECK (octet_length(csrf_token_hash) = 32),
    ADD CONSTRAINT auth_sessions_absolute_expiry_chk
        CHECK (expires_at <= absolute_expires_at);

CREATE TABLE pvnaive.actor_totp_factors (
    actor_id uuid PRIMARY KEY REFERENCES pvnaive.actors(id) ON DELETE CASCADE,
    secret_ciphertext bytea NOT NULL CHECK (octet_length(secret_ciphertext) >= 16),
    secret_nonce bytea NOT NULL CHECK (octet_length(secret_nonce) = 12),
    encryption_key_id text NOT NULL CHECK (length(encryption_key_id) BETWEEN 1 AND 160),
    last_used_step bigint,
    confirmed_at timestamptz,
    disabled_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    updated_at timestamptz NOT NULL DEFAULT clock_timestamp()
);

CREATE TABLE pvnaive.actor_mfa_recovery_codes (
    id uuid PRIMARY KEY DEFAULT public.gen_random_uuid(),
    actor_id uuid NOT NULL REFERENCES pvnaive.actors(id) ON DELETE CASCADE,
    code_hash bytea NOT NULL CHECK (octet_length(code_hash) = 32),
    used_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    UNIQUE (actor_id, code_hash)
);
CREATE INDEX actor_mfa_recovery_codes_unused_idx
    ON pvnaive.actor_mfa_recovery_codes (actor_id)
    WHERE used_at IS NULL;

REVOKE ALL ON pvnaive.actor_totp_factors FROM PUBLIC;
REVOKE ALL ON pvnaive.actor_mfa_recovery_codes FROM PUBLIC;
REVOKE ALL ON pvnaive.actor_totp_factors FROM pvnaive_app;
REVOKE ALL ON pvnaive.actor_mfa_recovery_codes FROM pvnaive_app;
REVOKE INSERT, UPDATE, DELETE ON pvnaive.actors FROM pvnaive_app;
REVOKE INSERT, UPDATE, DELETE ON pvnaive.auth_sessions FROM pvnaive_app;

CREATE FUNCTION pvnaive.auth_lookup_actor(p_email text)
RETURNS TABLE (
    actor_id uuid,
    tenant_id uuid,
    actor_role text,
    password_hash text,
    mfa_required boolean,
    status text,
    locked_until timestamptz,
    totp_confirmed boolean
)
LANGUAGE sql
STABLE
SECURITY DEFINER
SET search_path = pg_catalog, pvnaive
AS $$
    SELECT a.id,
           a.tenant_id,
           a.actor_role,
           a.password_hash,
           a.mfa_required,
           a.status,
           a.locked_until,
           (f.confirmed_at IS NOT NULL AND f.disabled_at IS NULL)
      FROM pvnaive.actors AS a
      LEFT JOIN pvnaive.actor_totp_factors AS f ON f.actor_id = a.id
     WHERE lower(a.email) = lower(btrim(p_email))
     LIMIT 1;
$$;

CREATE FUNCTION pvnaive.auth_record_login_failure(p_actor_id uuid)
RETURNS timestamptz
LANGUAGE plpgsql
VOLATILE
SECURITY DEFINER
SET search_path = pg_catalog, pvnaive
AS $$
DECLARE
    new_locked_until timestamptz;
BEGIN
    UPDATE pvnaive.actors
       SET failed_login_attempts = LEAST(failed_login_attempts + 1, 32767),
           locked_until = CASE
               WHEN failed_login_attempts + 1 >= 5
               THEN GREATEST(
                   COALESCE(locked_until, '-infinity'::timestamptz),
                   clock_timestamp() + interval '15 minutes'
               )
               ELSE locked_until
           END,
           updated_at = clock_timestamp()
     WHERE id = p_actor_id
     RETURNING locked_until INTO new_locked_until;
    RETURN new_locked_until;
END;
$$;

CREATE FUNCTION pvnaive.auth_record_login_success(p_actor_id uuid)
RETURNS void
LANGUAGE plpgsql
VOLATILE
SECURITY DEFINER
SET search_path = pg_catalog, pvnaive
AS $$
BEGIN
    UPDATE pvnaive.actors
       SET failed_login_attempts = 0,
           locked_until = NULL,
           last_login_at = clock_timestamp(),
           updated_at = clock_timestamp()
     WHERE id = p_actor_id;
END;
$$;

CREATE FUNCTION pvnaive.auth_create_session(
    p_actor_id uuid,
    p_token_hash bytea,
    p_csrf_token_hash bytea,
    p_refresh_family_id uuid,
    p_user_agent_hash bytea,
    p_expires_at timestamptz,
    p_absolute_expires_at timestamptz
)
RETURNS uuid
LANGUAGE plpgsql
VOLATILE
SECURITY DEFINER
SET search_path = pg_catalog, pvnaive
AS $$
DECLARE
    actor_tenant_id uuid;
    created_session_id uuid;
    now_value timestamptz := clock_timestamp();
BEGIN
    IF p_token_hash IS NULL
       OR p_csrf_token_hash IS NULL
       OR octet_length(p_token_hash) <> 32
       OR octet_length(p_csrf_token_hash) <> 32
       OR (p_user_agent_hash IS NOT NULL AND octet_length(p_user_agent_hash) <> 32) THEN
        RAISE EXCEPTION 'invalid session hash material' USING ERRCODE = '22023';
    END IF;
    IF p_expires_at <= now_value
       OR p_absolute_expires_at <= now_value
       OR p_expires_at > p_absolute_expires_at
       OR p_expires_at > now_value + interval '61 minutes'
       OR p_absolute_expires_at > now_value + interval '12 hours 1 minute' THEN
        RAISE EXCEPTION 'invalid session expiry' USING ERRCODE = '22023';
    END IF;

    SELECT tenant_id
      INTO actor_tenant_id
      FROM pvnaive.actors
     WHERE id = p_actor_id
       AND status = 'active'
       AND (locked_until IS NULL OR locked_until <= now_value);
    IF NOT FOUND THEN
        RAISE EXCEPTION 'actor is not eligible for a session' USING ERRCODE = '28000';
    END IF;

    INSERT INTO pvnaive.auth_sessions (
        tenant_id, actor_id, token_hash, csrf_token_hash, refresh_family_id,
        user_agent_hash, expires_at, absolute_expires_at
    )
    VALUES (
        actor_tenant_id, p_actor_id, p_token_hash, p_csrf_token_hash, p_refresh_family_id,
        p_user_agent_hash, p_expires_at, p_absolute_expires_at
    )
    RETURNING id INTO created_session_id;

    RETURN created_session_id;
END;
$$;

CREATE FUNCTION pvnaive.auth_rotate_session(
    p_old_token_hash bytea,
    p_new_token_hash bytea,
    p_new_csrf_token_hash bytea,
    p_new_user_agent_hash bytea,
    p_new_expires_at timestamptz
)
RETURNS TABLE (
    session_id uuid,
    actor_id uuid,
    tenant_id uuid,
    refresh_family_id uuid,
    absolute_expires_at timestamptz,
    reuse_detected boolean
)
LANGUAGE plpgsql
VOLATILE
SECURITY DEFINER
SET search_path = pg_catalog, pvnaive
AS $$
DECLARE
    old_session pvnaive.auth_sessions%ROWTYPE;
    new_session_id uuid;
    now_value timestamptz := clock_timestamp();
BEGIN
    IF p_old_token_hash IS NULL
       OR p_new_token_hash IS NULL
       OR p_new_csrf_token_hash IS NULL
       OR octet_length(p_old_token_hash) <> 32
       OR octet_length(p_new_token_hash) <> 32
       OR octet_length(p_new_csrf_token_hash) <> 32
       OR (p_new_user_agent_hash IS NOT NULL AND octet_length(p_new_user_agent_hash) <> 32) THEN
        RAISE EXCEPTION 'invalid session hash material' USING ERRCODE = '22023';
    END IF;

    SELECT *
      INTO old_session
      FROM pvnaive.auth_sessions
     WHERE token_hash = p_old_token_hash
     FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION 'invalid session' USING ERRCODE = '28000';
    END IF;

    IF old_session.revoked_at IS NOT NULL THEN
        UPDATE pvnaive.auth_sessions
           SET revoked_at = COALESCE(revoked_at, now_value),
               reuse_detected_at = COALESCE(reuse_detected_at, now_value)
         WHERE refresh_family_id = old_session.refresh_family_id;
        RETURN QUERY
        SELECT old_session.id, old_session.actor_id, old_session.tenant_id,
               old_session.refresh_family_id, old_session.absolute_expires_at, true;
        RETURN;
    END IF;

    IF old_session.expires_at <= now_value OR old_session.absolute_expires_at <= now_value THEN
        UPDATE pvnaive.auth_sessions
           SET revoked_at = COALESCE(revoked_at, now_value)
         WHERE id = old_session.id;
        RAISE EXCEPTION 'expired session' USING ERRCODE = '28000';
    END IF;

    IF p_new_expires_at <= now_value
       OR p_new_expires_at > old_session.absolute_expires_at
       OR p_new_expires_at > now_value + interval '61 minutes' THEN
        RAISE EXCEPTION 'invalid rotated session expiry' USING ERRCODE = '22023';
    END IF;

    UPDATE pvnaive.auth_sessions
       SET revoked_at = now_value
     WHERE id = old_session.id;

    INSERT INTO pvnaive.auth_sessions (
        tenant_id, actor_id, token_hash, csrf_token_hash, refresh_family_id,
        user_agent_hash, expires_at, absolute_expires_at
    )
    VALUES (
        old_session.tenant_id, old_session.actor_id, p_new_token_hash, p_new_csrf_token_hash,
        old_session.refresh_family_id, p_new_user_agent_hash, p_new_expires_at,
        old_session.absolute_expires_at
    )
    RETURNING id INTO new_session_id;

    RETURN QUERY
    SELECT new_session_id, old_session.actor_id, old_session.tenant_id,
           old_session.refresh_family_id, old_session.absolute_expires_at, false;
END;
$$;

CREATE FUNCTION pvnaive.auth_revoke_session(p_token_hash bytea)
RETURNS boolean
LANGUAGE plpgsql
VOLATILE
SECURITY DEFINER
SET search_path = pg_catalog, pvnaive
AS $$
DECLARE
    changed integer;
BEGIN
    IF octet_length(p_token_hash) <> 32 THEN
        RETURN false;
    END IF;
    UPDATE pvnaive.auth_sessions
       SET revoked_at = COALESCE(revoked_at, clock_timestamp())
     WHERE token_hash = p_token_hash
       AND revoked_at IS NULL;
    GET DIAGNOSTICS changed = ROW_COUNT;
    RETURN changed = 1;
END;
$$;

CREATE FUNCTION pvnaive.auth_revoke_actor_sessions(p_actor_id uuid)
RETURNS integer
LANGUAGE plpgsql
VOLATILE
SECURITY DEFINER
SET search_path = pg_catalog, pvnaive
AS $$
DECLARE
    changed integer;
BEGIN
    UPDATE pvnaive.auth_sessions
       SET revoked_at = COALESCE(revoked_at, clock_timestamp())
     WHERE actor_id = p_actor_id
       AND revoked_at IS NULL;
    GET DIAGNOSTICS changed = ROW_COUNT;
    RETURN changed;
END;
$$;

CREATE FUNCTION pvnaive.auth_revoke_session_by_id(p_actor_id uuid, p_session_id uuid)
RETURNS boolean
LANGUAGE plpgsql
VOLATILE
SECURITY DEFINER
SET search_path = pg_catalog, pvnaive
AS $$
DECLARE
    changed integer;
BEGIN
    IF NOT pvnaive.has_valid_context()
       OR pvnaive.current_actor_id() IS DISTINCT FROM p_actor_id THEN
        RAISE EXCEPTION 'authentication context required' USING ERRCODE = '42501';
    END IF;
    UPDATE pvnaive.auth_sessions
       SET revoked_at = COALESCE(revoked_at, clock_timestamp())
     WHERE id = p_session_id
       AND actor_id = p_actor_id
       AND revoked_at IS NULL;
    GET DIAGNOSTICS changed = ROW_COUNT;
    RETURN changed = 1;
END;
$$;

CREATE FUNCTION pvnaive.auth_revoke_other_actor_sessions(p_actor_id uuid, p_current_session_id uuid)
RETURNS bigint
LANGUAGE plpgsql
VOLATILE
SECURITY DEFINER
SET search_path = pg_catalog, pvnaive
AS $$
DECLARE
    changed bigint;
BEGIN
    IF NOT pvnaive.has_valid_context()
       OR pvnaive.current_actor_id() IS DISTINCT FROM p_actor_id THEN
        RAISE EXCEPTION 'authentication context required' USING ERRCODE = '42501';
    END IF;
    UPDATE pvnaive.auth_sessions
       SET revoked_at = COALESCE(revoked_at, clock_timestamp())
     WHERE actor_id = p_actor_id
       AND id <> p_current_session_id
       AND revoked_at IS NULL;
    GET DIAGNOSTICS changed = ROW_COUNT;
    RETURN changed;
END;
$$;

CREATE FUNCTION pvnaive.auth_get_totp_factor(p_actor_id uuid)
RETURNS TABLE (
    secret_ciphertext bytea,
    secret_nonce bytea,
    encryption_key_id text,
    last_used_step bigint,
    confirmed_at timestamptz
)
LANGUAGE sql
STABLE
SECURITY DEFINER
SET search_path = pg_catalog, pvnaive
AS $$
    SELECT f.secret_ciphertext, f.secret_nonce, f.encryption_key_id,
           f.last_used_step, f.confirmed_at
      FROM pvnaive.actor_totp_factors AS f
     WHERE f.actor_id = p_actor_id
       AND f.disabled_at IS NULL;
$$;

CREATE FUNCTION pvnaive.auth_upsert_totp_factor(
    p_actor_id uuid,
    p_secret_ciphertext bytea,
    p_secret_nonce bytea,
    p_encryption_key_id text
)
RETURNS void
LANGUAGE plpgsql
VOLATILE
SECURITY DEFINER
SET search_path = pg_catalog, pvnaive
AS $$
BEGIN
    IF NOT pvnaive.has_valid_context()
       OR pvnaive.current_actor_id() IS DISTINCT FROM p_actor_id THEN
        RAISE EXCEPTION 'authentication context required' USING ERRCODE = '42501';
    END IF;
    IF p_secret_ciphertext IS NULL
       OR p_secret_nonce IS NULL
       OR octet_length(p_secret_ciphertext) < 16
       OR octet_length(p_secret_nonce) <> 12
       OR length(p_encryption_key_id) NOT BETWEEN 1 AND 160 THEN
        RAISE EXCEPTION 'invalid TOTP factor material' USING ERRCODE = '22023';
    END IF;

    INSERT INTO pvnaive.actor_totp_factors (
        actor_id, secret_ciphertext, secret_nonce, encryption_key_id,
        last_used_step, confirmed_at
    )
    VALUES (
        p_actor_id, p_secret_ciphertext, p_secret_nonce, p_encryption_key_id,
        NULL, NULL
    )
    ON CONFLICT (actor_id) DO UPDATE
       SET secret_ciphertext = EXCLUDED.secret_ciphertext,
           secret_nonce = EXCLUDED.secret_nonce,
           encryption_key_id = EXCLUDED.encryption_key_id,
           last_used_step = NULL,
           confirmed_at = NULL,
           disabled_at = NULL,
           updated_at = clock_timestamp();

    UPDATE pvnaive.actor_mfa_recovery_codes
       SET used_at = COALESCE(used_at, clock_timestamp())
     WHERE actor_id = p_actor_id
       AND used_at IS NULL;
    UPDATE pvnaive.actors SET mfa_required = false, updated_at = clock_timestamp()
     WHERE id = p_actor_id;
END;
$$;

CREATE FUNCTION pvnaive.auth_confirm_totp_factor(
    p_actor_id uuid,
    p_used_step bigint,
    p_recovery_code_hashes bytea[]
)
RETURNS void
LANGUAGE plpgsql
VOLATILE
SECURITY DEFINER
SET search_path = pg_catalog, pvnaive
AS $$
DECLARE
    recovery_hash bytea;
BEGIN
    IF NOT pvnaive.has_valid_context()
       OR pvnaive.current_actor_id() IS DISTINCT FROM p_actor_id THEN
        RAISE EXCEPTION 'authentication context required' USING ERRCODE = '42501';
    END IF;
    IF p_used_step < 0 OR array_length(p_recovery_code_hashes, 1) <> 10 THEN
        RAISE EXCEPTION 'invalid MFA confirmation material' USING ERRCODE = '22023';
    END IF;

    UPDATE pvnaive.actor_totp_factors
       SET confirmed_at = COALESCE(confirmed_at, clock_timestamp()),
           last_used_step = p_used_step,
           updated_at = clock_timestamp()
     WHERE actor_id = p_actor_id
       AND disabled_at IS NULL;
    IF NOT FOUND THEN
        RAISE EXCEPTION 'TOTP factor is not enrolled' USING ERRCODE = '22023';
    END IF;

    UPDATE pvnaive.actor_mfa_recovery_codes
       SET used_at = COALESCE(used_at, clock_timestamp())
     WHERE actor_id = p_actor_id
       AND used_at IS NULL;
    FOREACH recovery_hash IN ARRAY p_recovery_code_hashes LOOP
        IF octet_length(recovery_hash) <> 32 THEN
            RAISE EXCEPTION 'invalid recovery-code hash' USING ERRCODE = '22023';
        END IF;
        INSERT INTO pvnaive.actor_mfa_recovery_codes (actor_id, code_hash)
        VALUES (p_actor_id, recovery_hash);
    END LOOP;

    UPDATE pvnaive.actors
       SET mfa_required = true,
           updated_at = clock_timestamp()
     WHERE id = p_actor_id;
END;
$$;

CREATE FUNCTION pvnaive.auth_consume_totp_step(p_actor_id uuid, p_step bigint)
RETURNS boolean
LANGUAGE plpgsql
VOLATILE
SECURITY DEFINER
SET search_path = pg_catalog, pvnaive
AS $$
DECLARE
    changed integer;
BEGIN
    IF p_step < 0 THEN
        RETURN false;
    END IF;
    UPDATE pvnaive.actor_totp_factors
       SET last_used_step = p_step,
           updated_at = clock_timestamp()
     WHERE actor_id = p_actor_id
       AND confirmed_at IS NOT NULL
       AND disabled_at IS NULL
       AND (last_used_step IS NULL OR p_step > last_used_step);
    GET DIAGNOSTICS changed = ROW_COUNT;
    RETURN changed = 1;
END;
$$;

CREATE FUNCTION pvnaive.auth_consume_recovery_code(p_actor_id uuid, p_code_hash bytea)
RETURNS boolean
LANGUAGE plpgsql
VOLATILE
SECURITY DEFINER
SET search_path = pg_catalog, pvnaive
AS $$
DECLARE
    changed integer;
BEGIN
    IF p_code_hash IS NULL OR octet_length(p_code_hash) <> 32 THEN
        RETURN false;
    END IF;
    UPDATE pvnaive.actor_mfa_recovery_codes
       SET used_at = clock_timestamp()
     WHERE actor_id = p_actor_id
       AND code_hash = p_code_hash
       AND used_at IS NULL;
    GET DIAGNOSTICS changed = ROW_COUNT;
    RETURN changed = 1;
END;
$$;

CREATE FUNCTION pvnaive.auth_remove_mfa(p_actor_id uuid)
RETURNS void
LANGUAGE plpgsql
VOLATILE
SECURITY DEFINER
SET search_path = pg_catalog, pvnaive
AS $$
BEGIN
    IF NOT pvnaive.has_valid_context()
       OR pvnaive.current_actor_id() IS DISTINCT FROM p_actor_id THEN
        RAISE EXCEPTION 'authentication context required' USING ERRCODE = '42501';
    END IF;
    UPDATE pvnaive.actor_mfa_recovery_codes
       SET used_at = COALESCE(used_at, clock_timestamp())
     WHERE actor_id = p_actor_id
       AND used_at IS NULL;
    UPDATE pvnaive.actor_totp_factors
       SET disabled_at = COALESCE(disabled_at, clock_timestamp()),
           updated_at = clock_timestamp()
     WHERE actor_id = p_actor_id
       AND disabled_at IS NULL;
    UPDATE pvnaive.actors
       SET mfa_required = false,
           updated_at = clock_timestamp()
     WHERE id = p_actor_id;
END;
$$;

CREATE POLICY auth_audit_owner_insert ON pvnaive.audit_events
FOR INSERT
TO pvnaive_owner
WITH CHECK (true);

CREATE FUNCTION pvnaive.auth_append_audit(
    p_actor_id uuid,
    p_action text,
    p_outcome text,
    p_reason_code text DEFAULT NULL,
    p_request_id uuid DEFAULT NULL,
    p_source_ip inet DEFAULT NULL
)
RETURNS uuid
LANGUAGE plpgsql
VOLATILE
SECURITY DEFINER
SET search_path = pg_catalog, pvnaive
AS $$
DECLARE
    actor_tenant_id uuid;
    audit_id uuid;
BEGIN
    IF p_outcome IS NULL
       OR p_action IS NULL
       OR p_outcome NOT IN ('success', 'denied', 'failure')
       OR length(p_action) NOT BETWEEN 1 AND 120 THEN
        RAISE EXCEPTION 'invalid audit event' USING ERRCODE = '22023';
    END IF;
    IF p_actor_id IS NOT NULL THEN
        SELECT tenant_id INTO actor_tenant_id
          FROM pvnaive.actors
         WHERE id = p_actor_id;
        IF NOT FOUND THEN
            RAISE EXCEPTION 'unknown audit actor' USING ERRCODE = '22023';
        END IF;
    END IF;

    audit_id := public.gen_random_uuid();
    INSERT INTO pvnaive.audit_events (
        id, tenant_id, actor_id, action, object_type, request_id,
        source_ip, outcome, reason_code
    )
    VALUES (
        audit_id, actor_tenant_id, p_actor_id, p_action, 'auth', p_request_id,
        p_source_ip, p_outcome, p_reason_code
    );
    RETURN audit_id;
END;
$$;

REVOKE ALL ON FUNCTION pvnaive.auth_lookup_actor(text) FROM PUBLIC;
REVOKE ALL ON FUNCTION pvnaive.auth_record_login_failure(uuid) FROM PUBLIC;
REVOKE ALL ON FUNCTION pvnaive.auth_record_login_success(uuid) FROM PUBLIC;
REVOKE ALL ON FUNCTION pvnaive.auth_create_session(uuid, bytea, bytea, uuid, bytea, timestamptz, timestamptz) FROM PUBLIC;
REVOKE ALL ON FUNCTION pvnaive.auth_rotate_session(bytea, bytea, bytea, bytea, timestamptz) FROM PUBLIC;
REVOKE ALL ON FUNCTION pvnaive.auth_revoke_session(bytea) FROM PUBLIC;
REVOKE ALL ON FUNCTION pvnaive.auth_revoke_actor_sessions(uuid) FROM PUBLIC;
REVOKE ALL ON FUNCTION pvnaive.auth_revoke_session_by_id(uuid, uuid) FROM PUBLIC;
REVOKE ALL ON FUNCTION pvnaive.auth_revoke_other_actor_sessions(uuid, uuid) FROM PUBLIC;
REVOKE ALL ON FUNCTION pvnaive.auth_get_totp_factor(uuid) FROM PUBLIC;
REVOKE ALL ON FUNCTION pvnaive.auth_upsert_totp_factor(uuid, bytea, bytea, text) FROM PUBLIC;
REVOKE ALL ON FUNCTION pvnaive.auth_confirm_totp_factor(uuid, bigint, bytea[]) FROM PUBLIC;
REVOKE ALL ON FUNCTION pvnaive.auth_consume_totp_step(uuid, bigint) FROM PUBLIC;
REVOKE ALL ON FUNCTION pvnaive.auth_consume_recovery_code(uuid, bytea) FROM PUBLIC;
REVOKE ALL ON FUNCTION pvnaive.auth_remove_mfa(uuid) FROM PUBLIC;
REVOKE ALL ON FUNCTION pvnaive.auth_append_audit(uuid, text, text, text, uuid, inet) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION pvnaive.auth_lookup_actor(text) TO pvnaive_app;
GRANT EXECUTE ON FUNCTION pvnaive.auth_record_login_failure(uuid) TO pvnaive_app;
GRANT EXECUTE ON FUNCTION pvnaive.auth_record_login_success(uuid) TO pvnaive_app;
GRANT EXECUTE ON FUNCTION pvnaive.auth_create_session(uuid, bytea, bytea, uuid, bytea, timestamptz, timestamptz) TO pvnaive_app;
GRANT EXECUTE ON FUNCTION pvnaive.auth_rotate_session(bytea, bytea, bytea, bytea, timestamptz) TO pvnaive_app;
GRANT EXECUTE ON FUNCTION pvnaive.auth_revoke_session(bytea) TO pvnaive_app;
GRANT EXECUTE ON FUNCTION pvnaive.auth_revoke_actor_sessions(uuid) TO pvnaive_app;
GRANT EXECUTE ON FUNCTION pvnaive.auth_revoke_session_by_id(uuid, uuid) TO pvnaive_app;
GRANT EXECUTE ON FUNCTION pvnaive.auth_revoke_other_actor_sessions(uuid, uuid) TO pvnaive_app;
GRANT EXECUTE ON FUNCTION pvnaive.auth_get_totp_factor(uuid) TO pvnaive_app;
GRANT EXECUTE ON FUNCTION pvnaive.auth_upsert_totp_factor(uuid, bytea, bytea, text) TO pvnaive_app;
GRANT EXECUTE ON FUNCTION pvnaive.auth_confirm_totp_factor(uuid, bigint, bytea[]) TO pvnaive_app;
GRANT EXECUTE ON FUNCTION pvnaive.auth_consume_totp_step(uuid, bigint) TO pvnaive_app;
GRANT EXECUTE ON FUNCTION pvnaive.auth_consume_recovery_code(uuid, bytea) TO pvnaive_app;
GRANT EXECUTE ON FUNCTION pvnaive.auth_remove_mfa(uuid) TO pvnaive_app;
GRANT EXECUTE ON FUNCTION pvnaive.auth_append_audit(uuid, text, text, text, uuid, inet) TO pvnaive_app;

COMMENT ON TABLE pvnaive.actor_totp_factors IS 'Encrypted TOTP factor material; never directly readable by pvnaive_app';
COMMENT ON TABLE pvnaive.actor_mfa_recovery_codes IS 'One-time high-entropy MFA recovery-code hashes';
COMMENT ON FUNCTION pvnaive.auth_lookup_actor(text) IS 'Narrow pre-authentication actor lookup for login; bypasses actor RLS by design';
COMMENT ON FUNCTION pvnaive.auth_rotate_session(bytea, bytea, bytea, bytea, timestamptz) IS 'Atomic opaque session rotation with refresh-family reuse detection';
