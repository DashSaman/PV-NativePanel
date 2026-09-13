-- pvnaive:migration-version 0028
-- pvnaive:migration-name auth_actor_credential_mutations_hardened
-- pvnaive:transactional true
-- pvnaive:destructive true

-- Restores the pre-0028 (unguarded) SECURITY DEFINER bodies deployed by the
-- 0027 lineage, then removes this migration's bookkeeping row.
CREATE OR REPLACE FUNCTION pvnaive.auth_update_actor_password(p_actor_id uuid, p_password_hash text)
RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog, pvnaive
AS $$
BEGIN
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

CREATE OR REPLACE FUNCTION pvnaive.auth_update_actor_profile(p_actor_id uuid, p_email text, p_display_name text)
RETURNS TABLE (out_email text, out_display_name text)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog, pvnaive
AS $$
BEGIN
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
DELETE FROM pvnaive.schema_migrations WHERE version = 28;
