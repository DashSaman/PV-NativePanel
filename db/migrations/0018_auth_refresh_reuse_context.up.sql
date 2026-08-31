-- pvnaive:migration-version 0018
-- pvnaive:migration-name auth_refresh_reuse_context
-- pvnaive:transactional true
-- pvnaive:destructive false

SET LOCAL lock_timeout = '10s';
SET LOCAL statement_timeout = '120s';
SET LOCAL ROLE pvnaive_owner;

CREATE FUNCTION pvnaive.auth_refresh_session_metadata(p_token_hash bytea)
RETURNS TABLE (
    csrf_token_hash bytea,
    absolute_expires_at timestamptz
)
LANGUAGE plpgsql
STABLE
SECURITY DEFINER
SET search_path = pg_catalog, pvnaive
AS $$
BEGIN
    IF p_token_hash IS NULL OR octet_length(p_token_hash) <> 32 THEN
        RAISE EXCEPTION 'invalid refresh session' USING ERRCODE = '28000';
    END IF;

    RETURN QUERY
    SELECT s.csrf_token_hash, s.absolute_expires_at
      FROM pvnaive.auth_sessions AS s
      JOIN pvnaive.actors AS a ON a.id = s.actor_id
     WHERE s.token_hash = p_token_hash
       AND s.tenant_id IS NOT DISTINCT FROM a.tenant_id
       AND a.status = 'active'
     LIMIT 1;

    IF NOT FOUND THEN
        RAISE EXCEPTION 'invalid refresh session' USING ERRCODE = '28000';
    END IF;
END;
$$;

REVOKE ALL ON FUNCTION pvnaive.auth_refresh_session_metadata(bytea) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION pvnaive.auth_refresh_session_metadata(bytea) TO pvnaive_app;

COMMENT ON FUNCTION pvnaive.auth_refresh_session_metadata(bytea) IS
    'Narrow refresh-only metadata lookup that deliberately includes revoked rotated tokens so auth_rotate_session can detect and revoke refresh-family reuse; never establishes RLS request context';

-- The original schema-2 rotate function contains a latent PL/pgSQL name collision
-- in the revoked-token reuse branch: the RETURNS TABLE output parameter
-- refresh_family_id conflicts with auth_sessions.refresh_family_id. That branch
-- was unreachable before this migration. Preserve the proven normal rotation
-- implementation under a private compatibility name and put a narrow wrapper in
-- front of it that handles only the revoked-token reuse case with qualified SQL.
ALTER FUNCTION pvnaive.auth_rotate_session(bytea,bytea,bytea,bytea,timestamptz)
    RENAME TO auth_rotate_session_v17;

REVOKE ALL ON FUNCTION pvnaive.auth_rotate_session_v17(bytea,bytea,bytea,bytea,timestamptz) FROM PUBLIC;

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
    v_old pvnaive.auth_sessions%ROWTYPE;
    v_now timestamptz := clock_timestamp();
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

    SELECT s.*
      INTO v_old
      FROM pvnaive.auth_sessions AS s
     WHERE s.token_hash = p_old_token_hash
     FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION 'invalid session' USING ERRCODE = '28000';
    END IF;

    IF v_old.revoked_at IS NOT NULL THEN
        UPDATE pvnaive.auth_sessions AS s
           SET revoked_at = COALESCE(s.revoked_at, v_now),
               reuse_detected_at = COALESCE(s.reuse_detected_at, v_now)
         WHERE s.refresh_family_id = v_old.refresh_family_id;
        RETURN QUERY
        SELECT v_old.id, v_old.actor_id, v_old.tenant_id,
               v_old.refresh_family_id, v_old.absolute_expires_at, true;
        RETURN;
    END IF;

    RETURN QUERY
    SELECT r.session_id, r.actor_id, r.tenant_id, r.refresh_family_id,
           r.absolute_expires_at, r.reuse_detected
      FROM pvnaive.auth_rotate_session_v17(
          p_old_token_hash,
          p_new_token_hash,
          p_new_csrf_token_hash,
          p_new_user_agent_hash,
          p_new_expires_at
      ) AS r;
END;
$$;

REVOKE ALL ON FUNCTION pvnaive.auth_rotate_session(bytea,bytea,bytea,bytea,timestamptz) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION pvnaive.auth_rotate_session(bytea,bytea,bytea,bytea,timestamptz) TO pvnaive_app;
