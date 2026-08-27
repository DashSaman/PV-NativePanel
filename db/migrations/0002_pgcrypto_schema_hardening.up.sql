-- pvnaive:migration-version 0002
-- Source: PVNaive PostgreSQL crypto schema hardening
-- pvnaive:migration-name pgcrypto_schema_hardening
-- pvnaive:transactional true
-- pvnaive:destructive false

SET LOCAL lock_timeout = '10s';
SET LOCAL statement_timeout = '120s';
SET LOCAL idle_in_transaction_session_timeout = '60s';
SET LOCAL ROLE pvnaive_owner;

CREATE SCHEMA pvnaive_crypto AUTHORIZATION pvnaive_owner;
REVOKE ALL ON SCHEMA pvnaive_crypto FROM PUBLIC;

ALTER EXTENSION pgcrypto SET SCHEMA pvnaive_crypto;

-- pgcrypto is deliberately not exposed to the application role. The NOLOGIN
-- owner can execute only the HMAC primitive required by SECURITY DEFINER
-- request-context functions.
REVOKE ALL ON ALL FUNCTIONS IN SCHEMA pvnaive_crypto FROM PUBLIC;
GRANT USAGE ON SCHEMA pvnaive_crypto TO pvnaive_owner;
GRANT EXECUTE ON FUNCTION pvnaive_crypto.hmac(bytea, bytea, text) TO pvnaive_owner;

CREATE OR REPLACE FUNCTION pvnaive.has_valid_context()
RETURNS boolean
LANGUAGE plpgsql
STABLE
SECURITY DEFINER
SET search_path = pg_catalog, pvnaive, pvnaive_crypto
AS $$
DECLARE
    actor_id_value uuid;
    tenant_id_value uuid;
    actor_role_value text;
    provided_signature text;
    expected_signature text;
    signing_key_value bytea;
BEGIN
    actor_id_value := NULLIF(current_setting('pvnaive.actor_id', true), '')::uuid;
    tenant_id_value := NULLIF(current_setting('pvnaive.tenant_id', true), '')::uuid;
    actor_role_value := NULLIF(current_setting('pvnaive.actor_role', true), '');
    provided_signature := NULLIF(current_setting('pvnaive.context_signature', true), '');
    IF actor_id_value IS NULL OR actor_role_value IS NULL OR provided_signature IS NULL THEN
        RETURN false;
    END IF;

    SELECT signing_key
      INTO signing_key_value
      FROM pvnaive.security_context_keys
     WHERE singleton;

    expected_signature := encode(
        hmac(
            pvnaive.context_payload(actor_id_value, tenant_id_value, actor_role_value),
            signing_key_value,
            'sha256'
        ),
        'hex'
    );
    RETURN expected_signature = provided_signature;
EXCEPTION WHEN OTHERS THEN
    RETURN false;
END;
$$;

CREATE OR REPLACE FUNCTION pvnaive.set_request_context(p_session_token_hash bytea)
RETURNS TABLE (actor_id uuid, tenant_id uuid, actor_role text)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog, pvnaive, pvnaive_crypto
AS $$
DECLARE
    selected_actor pvnaive.actors%ROWTYPE;
    signing_key_value bytea;
    signature_value text;
BEGIN
    IF p_session_token_hash IS NULL OR octet_length(p_session_token_hash) <> 32 THEN
        RAISE EXCEPTION 'invalid session context' USING ERRCODE = '28000';
    END IF;

    SELECT a.* INTO selected_actor
      FROM pvnaive.auth_sessions AS s
      JOIN pvnaive.actors AS a ON a.id = s.actor_id
     WHERE s.token_hash = p_session_token_hash
       AND s.tenant_id IS NOT DISTINCT FROM a.tenant_id
       AND s.revoked_at IS NULL
       AND s.expires_at > clock_timestamp()
       AND a.status = 'active';

    IF NOT FOUND THEN
        RAISE EXCEPTION 'invalid session context' USING ERRCODE = '28000';
    END IF;

    SELECT signing_key
      INTO signing_key_value
      FROM pvnaive.security_context_keys
     WHERE singleton;

    signature_value := encode(
        hmac(
            pvnaive.context_payload(selected_actor.id, selected_actor.tenant_id, selected_actor.actor_role),
            signing_key_value,
            'sha256'
        ),
        'hex'
    );

    PERFORM set_config('pvnaive.actor_id', selected_actor.id::text, true);
    PERFORM set_config('pvnaive.tenant_id', COALESCE(selected_actor.tenant_id::text, ''), true);
    PERFORM set_config('pvnaive.actor_role', selected_actor.actor_role, true);
    PERFORM set_config('pvnaive.context_signature', signature_value, true);

    RETURN QUERY SELECT selected_actor.id, selected_actor.tenant_id, selected_actor.actor_role;
END;
$$;

COMMENT ON SCHEMA pvnaive_crypto IS 'PVNaive private pgcrypto extension schema; no direct application access';
