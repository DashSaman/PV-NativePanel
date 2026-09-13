-- pvnaive:migration-version 0022
-- pvnaive:transactional true
-- pvnaive:destructive false
-- Self-service actor credential mutations (0002 removed direct DML on
-- pvnaive.actors from pvnaive_app). The panel needs a controlled way for an
-- authenticated actor to rotate their own password and login identity, so the
-- app role gets narrowly-scoped SECURITY DEFINER functions instead of table
-- grants, matching the auth_* function family from 0002.
CREATE FUNCTION pvnaive.auth_update_actor_password(p_actor_id uuid, p_password_hash text)
RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog, pvnaive
AS $$
BEGIN
    IF NOT pvnaive.has_valid_context()
       OR pvnaive.current_actor_id() IS DISTINCT FROM p_actor_id THEN
        RAISE EXCEPTION 'authentication context required' USING ERRCODE = '42501';
    END IF;
    IF p_password_hash IS NULL OR p_password_hash NOT LIKE '$argon2id$%' THEN
        RAISE EXCEPTION 'auth_update_actor_password: password hash format rejected';
    END IF;
    UPDATE pvnaive.actors
       SET password_hash = p_password_hash,
           updated_at = clock_timestamp()
     WHERE id = p_actor_id;
    IF NOT FOUND THEN
        RAISE EXCEPTION 'auth_update_actor_password: actor % not found', p_actor_id;
    END IF;
END;
$$;

CREATE FUNCTION pvnaive.auth_update_actor_profile(p_actor_id uuid, p_email text, p_display_name text)
RETURNS TABLE (out_email text, out_display_name text)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog, pvnaive
AS $$
BEGIN
    IF NOT pvnaive.has_valid_context()
       OR pvnaive.current_actor_id() IS DISTINCT FROM p_actor_id THEN
        RAISE EXCEPTION 'authentication context required' USING ERRCODE = '42501';
    END IF;
    UPDATE pvnaive.actors
       SET email = COALESCE(NULLIF(p_email, ''), email),
           display_name = COALESCE(NULLIF(p_display_name, ''), display_name),
           updated_at = clock_timestamp()
     WHERE id = p_actor_id
    RETURNING email, display_name INTO out_email, out_display_name;
    IF NOT FOUND THEN
        RAISE EXCEPTION 'auth_update_actor_profile: actor % not found', p_actor_id;
    END IF;
    RETURN NEXT;
END;
$$;

GRANT EXECUTE ON FUNCTION pvnaive.auth_update_actor_password(uuid, text) TO pvnaive_app;
GRANT EXECUTE ON FUNCTION pvnaive.auth_update_actor_profile(uuid, text, text) TO pvnaive_app;
