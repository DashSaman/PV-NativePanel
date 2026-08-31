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
