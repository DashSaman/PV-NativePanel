-- pvnaive:migration-version 0033
-- pvnaive:migration-name pool_registry
-- pvnaive:transactional true
-- pvnaive:destructive false
-- R5 Pool Manager registry (docs/AGENT_TASKS.md R5, GATE STEER-006).
--
-- Durable pull-model registry for the node pool:
--   * pool_nodes               — the operator-managed node inventory (drain
--     workflow state machine lives here: active -> draining -> disabled).
--   * pool_manifest_revisions  — append-only, MONOTONIC signed desired-state
--     revisions per node (Ed25519 envelope from internal/fleet). Sibling
--     agents PULL the latest revision; the registry never pushes. A node
--     that cannot reach the registry keeps serving its last-known-good
--     manifest until expiry (ValidUntil is respected at trust time).
--   * pool_enrollment_tokens   — single-use, expiring enrollment tokens
--     (only SHA-256 hashes are stored; the raw token is shown exactly once).
--
-- Trusted boundary (same rule as 0030/0031/0032): pvnaive_app touches the
-- registry ONLY through SECURITY DEFINER functions; direct table access is
-- revoked and RLS is enabled with no policies. Even the migration-owner role
-- receives no DML grants on the revision ledger (append-only is enforced by
-- the publish function computing the next revision).

CREATE TABLE pvnaive.pool_nodes (
    id                  UUID        PRIMARY KEY,
    display_name        TEXT        NOT NULL CHECK (length(btrim(display_name)) BETWEEN 1 AND 120),
    region              TEXT        CHECK (region IS NULL OR length(btrim(region)) BETWEEN 1 AND 60),
    capacity_weight     INTEGER     NOT NULL DEFAULT 1 CHECK (capacity_weight BETWEEN 1 AND 10000),
    health              TEXT        NOT NULL DEFAULT 'unknown'
        CHECK (health IN ('unknown','healthy','degraded','offline')),
    maintenance         TEXT        NOT NULL DEFAULT 'active'
        CHECK (maintenance IN ('active','draining','disabled')),
    applied_revision    BIGINT      NOT NULL DEFAULT 0 CHECK (applied_revision >= 0),
    created_by_actor_id UUID        NOT NULL REFERENCES pvnaive.actors(id),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_seen_at        TIMESTAMPTZ,
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE pvnaive.pool_manifest_revisions (
    node_id     UUID        NOT NULL REFERENCES pvnaive.pool_nodes(id) ON DELETE CASCADE,
    revision    BIGINT      NOT NULL CHECK (revision >= 1),
    manifest    JSONB       NOT NULL,
    signature   TEXT        NOT NULL CHECK (length(btrim(signature)) BETWEEN 64 AND 256),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (node_id, revision)
);

CREATE TABLE pvnaive.pool_enrollment_tokens (
    token_hash      TEXT        PRIMARY KEY CHECK (token_hash ~ '^[0-9a-f]{64}$'),
    node_name       TEXT        NOT NULL CHECK (length(btrim(node_name)) BETWEEN 1 AND 120),
    expires_at      TIMESTAMPTZ NOT NULL,
    used_by_node_id UUID        REFERENCES pvnaive.pool_nodes(id),
    used_at         TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (used_by_node_id IS NULL OR used_at IS NOT NULL)
);

REVOKE ALL ON pvnaive.pool_nodes FROM PUBLIC, pvnaive_app;
REVOKE ALL ON pvnaive.pool_manifest_revisions FROM PUBLIC, pvnaive_app;
REVOKE ALL ON pvnaive.pool_enrollment_tokens FROM PUBLIC, pvnaive_app;

ALTER TABLE pvnaive.pool_nodes ENABLE ROW LEVEL SECURITY;
ALTER TABLE pvnaive.pool_manifest_revisions ENABLE ROW LEVEL SECURITY;
ALTER TABLE pvnaive.pool_enrollment_tokens ENABLE ROW LEVEL SECURITY;

-- Record a new single-use enrollment token (only its SHA-256 hash reaches the
-- database; the raw token is returned to the operator exactly once by the app).
CREATE FUNCTION pvnaive.pool_enrollment_token_record(
    p_token_hash text,
    p_node_name text,
    p_ttl_seconds integer
)
RETURNS TABLE (expires_at timestamptz)
LANGUAGE plpgsql
VOLATILE
SECURITY DEFINER
SET search_path = pg_catalog, pvnaive
AS $$
BEGIN
    IF p_token_hash IS NULL OR p_token_hash !~ '^[0-9a-f]{64}$'
       OR p_node_name IS NULL OR length(btrim(p_node_name)) NOT BETWEEN 1 AND 120
       OR p_ttl_seconds IS NULL OR p_ttl_seconds NOT BETWEEN 300 AND 86400 THEN
        RAISE EXCEPTION 'invalid enrollment token record' USING ERRCODE = '22023';
    END IF;
    INSERT INTO pvnaive.pool_enrollment_tokens (token_hash, node_name, expires_at)
    VALUES (p_token_hash, btrim(p_node_name), now() + make_interval(secs => p_ttl_seconds::double precision))
    ON CONFLICT (token_hash) DO NOTHING;
    IF NOT FOUND THEN
        RAISE EXCEPTION 'enrollment token hash collision' USING ERRCODE = '23505';
    END IF;
    RETURN QUERY SELECT t.expires_at FROM pvnaive.pool_enrollment_tokens t
      WHERE t.token_hash = p_token_hash;
END;
$$;

-- Consume a token: creates the node inventory row (single-use, unexpired).
-- Invalid/expired/used tokens are reported honestly as enrolled=false.
CREATE FUNCTION pvnaive.pool_node_enroll(
    p_token_hash text,
    p_display_name text,
    p_region text,
    p_capacity_weight integer,
    p_actor_id uuid
)
RETURNS TABLE (node_id uuid, enrolled boolean, reason text)
LANGUAGE plpgsql
VOLATILE
SECURITY DEFINER
SET search_path = pg_catalog, pvnaive
AS $$
DECLARE
    v_node_id uuid;
    v_consumed integer := 0;
BEGIN
    IF p_token_hash IS NULL OR p_token_hash !~ '^[0-9a-f]{64}$'
       OR p_display_name IS NULL OR length(btrim(p_display_name)) NOT BETWEEN 1 AND 120
       OR (p_region IS NOT NULL AND length(btrim(p_region)) NOT BETWEEN 1 AND 60)
       OR p_capacity_weight IS NULL OR p_capacity_weight NOT BETWEEN 1 AND 10000 THEN
        RAISE EXCEPTION 'invalid enrollment request' USING ERRCODE = '22023';
    END IF;

    UPDATE pvnaive.pool_enrollment_tokens
       SET used_at = now()
     WHERE token_hash = p_token_hash
       AND used_by_node_id IS NULL
       AND expires_at > now();
    GET DIAGNOSTICS v_consumed = ROW_COUNT;
    IF v_consumed = 0 THEN
        RETURN QUERY SELECT NULL::uuid, false, 'token_invalid'::text;
        RETURN;
    END IF;

    IF p_actor_id IS NULL OR NOT EXISTS
       (SELECT 1 FROM pvnaive.actors a WHERE a.id = p_actor_id) THEN
        RAISE EXCEPTION 'enrollment requires a known actor' USING ERRCODE = '23503';
    END IF;

    INSERT INTO pvnaive.pool_nodes (id, display_name, region, capacity_weight, created_by_actor_id)
    VALUES (gen_random_uuid(), btrim(p_display_name), btrim(p_region), p_capacity_weight, p_actor_id)
    RETURNING id INTO v_node_id;

    UPDATE pvnaive.pool_enrollment_tokens
       SET used_by_node_id = v_node_id
     WHERE token_hash = p_token_hash;

    RETURN QUERY SELECT v_node_id, true, 'enrolled'::text;
END;
$$;

-- Publish the next signed desired-state revision. Monotonic by construction:
-- the next revision is computed under an advisory lock per node; the ledger
-- is append-only and never updated. Disabled nodes refuse new revisions
-- (drain workflow: disabled means "leaving the pool").
CREATE FUNCTION pvnaive.pool_revision_publish(
    p_node_id uuid,
    p_manifest jsonb,
    p_signature text
)
RETURNS TABLE (revision bigint, accepted boolean)
LANGUAGE plpgsql
VOLATILE
SECURITY DEFINER
SET search_path = pg_catalog, pvnaive
AS $$
DECLARE
    v_next bigint;
    v_state text;
BEGIN
    IF p_node_id IS NULL OR p_manifest IS NULL
       OR p_signature IS NULL OR length(btrim(p_signature)) NOT BETWEEN 64 AND 256 THEN
        RAISE EXCEPTION 'invalid revision publish' USING ERRCODE = '22023';
    END IF;
    PERFORM pg_advisory_xact_lock(hashtext('pool-revision-' || p_node_id::text));

    SELECT maintenance INTO v_state FROM pvnaive.pool_nodes WHERE id = p_node_id;
    IF v_state IS NULL THEN
        RETURN QUERY SELECT 0::bigint, false;
        RETURN;
    END IF;
    IF v_state = 'disabled' THEN
        RETURN QUERY SELECT 0::bigint, false;
        RETURN;
    END IF;

    SELECT COALESCE(MAX(r.revision), 0) + 1 INTO v_next
      FROM pvnaive.pool_manifest_revisions r
     WHERE r.node_id = p_node_id;

    INSERT INTO pvnaive.pool_manifest_revisions (node_id, revision, manifest, signature)
    VALUES (p_node_id, v_next, p_manifest, btrim(p_signature));

    RETURN QUERY SELECT v_next, true;
END;
$$;

-- Pull-side read: the latest signed envelope for one node. Empty result =
-- nothing published yet (the agent keeps its last-known-good manifest).
CREATE FUNCTION pvnaive.pool_revision_latest(p_node_id uuid)
RETURNS TABLE (revision bigint, manifest jsonb, signature text)
LANGUAGE sql
STABLE
SECURITY DEFINER
SET search_path = pg_catalog, pvnaive
AS $$
    SELECT r.revision, r.manifest, r.signature
      FROM pvnaive.pool_manifest_revisions r
     WHERE r.node_id = p_node_id
     ORDER BY r.revision DESC
     LIMIT 1;
$$;

-- Agent heartbeat: last-seen + health + applied revision (monotonic guard:
-- an older applied revision can never rewind the recorded progress).
CREATE FUNCTION pvnaive.pool_node_heartbeat(
    p_node_id uuid,
    p_health text,
    p_applied_revision bigint
)
RETURNS TABLE (tracked boolean)
LANGUAGE plpgsql
VOLATILE
SECURITY DEFINER
SET search_path = pg_catalog, pvnaive
AS $$
BEGIN
    IF p_node_id IS NULL
       OR p_health IS NULL OR p_health NOT IN ('unknown','healthy','degraded','offline')
       OR p_applied_revision IS NULL OR p_applied_revision < 0 THEN
        RAISE EXCEPTION 'invalid heartbeat' USING ERRCODE = '22023';
    END IF;
    UPDATE pvnaive.pool_nodes n
       SET last_seen_at = now(),
           health = p_health,
           applied_revision = GREATEST(n.applied_revision, p_applied_revision),
           updated_at = now()
     WHERE n.id = p_node_id;
    IF NOT FOUND THEN
        RETURN QUERY SELECT false;
        RETURN;
    END IF;
    RETURN QUERY SELECT true;
END;
$$;

-- Drain workflow state machine (never yank a live node):
--   active -> draining -> disabled. Re-activation is allowed from draining
--   or disabled (operator fixed the node and wants it back). A direct
--   active -> disabled jump is refused.
CREATE FUNCTION pvnaive.pool_node_maintenance_set(
    p_node_id uuid,
    p_state text
)
RETURNS TABLE (tracked boolean)
LANGUAGE plpgsql
VOLATILE
SECURITY DEFINER
SET search_path = pg_catalog, pvnaive
AS $$
DECLARE
    v_current text;
BEGIN
    IF p_node_id IS NULL OR p_state NOT IN ('active','draining','disabled') THEN
        RAISE EXCEPTION 'invalid maintenance state' USING ERRCODE = '22023';
    END IF;
    SELECT maintenance INTO v_current FROM pvnaive.pool_nodes WHERE id = p_node_id;
    IF v_current IS NULL THEN
        RETURN QUERY SELECT false;
        RETURN;
    END IF;
    IF v_current = 'active' AND p_state = 'disabled' THEN
        RAISE EXCEPTION 'drain required before disabling a live node' USING ERRCODE = '23514';
    END IF;
    UPDATE pvnaive.pool_nodes
       SET maintenance = p_state, updated_at = now()
     WHERE id = p_node_id;
    RETURN QUERY SELECT true;
END;
$$;

-- Owner node list (drift visibility: desired = max published revision).
CREATE FUNCTION pvnaive.pool_nodes_list()
RETURNS TABLE (
    node_id           uuid,
    display_name      text,
    region            text,
    capacity_weight   integer,
    health            text,
    maintenance       text,
    desired_revision  bigint,
    applied_revision  bigint,
    last_seen_at      timestamptz
)
LANGUAGE sql
STABLE
SECURITY DEFINER
SET search_path = pg_catalog, pvnaive
AS $$
    SELECT n.id, n.display_name, n.region, n.capacity_weight,
           n.health, n.maintenance,
           COALESCE((SELECT MAX(r.revision) FROM pvnaive.pool_manifest_revisions r
                      WHERE r.node_id = n.id), 0::bigint) AS desired_revision,
           n.applied_revision,
           n.last_seen_at
      FROM pvnaive.pool_nodes n
     ORDER BY n.created_at, n.id;
$$;

REVOKE ALL ON FUNCTION pvnaive.pool_enrollment_token_record(text,text,integer) FROM PUBLIC;
REVOKE ALL ON FUNCTION pvnaive.pool_node_enroll(text,text,text,integer,uuid) FROM PUBLIC;
REVOKE ALL ON FUNCTION pvnaive.pool_revision_publish(uuid,jsonb,text) FROM PUBLIC;
REVOKE ALL ON FUNCTION pvnaive.pool_revision_latest(uuid) FROM PUBLIC;
REVOKE ALL ON FUNCTION pvnaive.pool_node_heartbeat(uuid,text,bigint) FROM PUBLIC;
REVOKE ALL ON FUNCTION pvnaive.pool_node_maintenance_set(uuid,text) FROM PUBLIC;
REVOKE ALL ON FUNCTION pvnaive.pool_nodes_list() FROM PUBLIC;
GRANT EXECUTE ON FUNCTION pvnaive.pool_enrollment_token_record(text,text,integer) TO pvnaive_app;
GRANT EXECUTE ON FUNCTION pvnaive.pool_node_enroll(text,text,text,integer,uuid) TO pvnaive_app;
GRANT EXECUTE ON FUNCTION pvnaive.pool_revision_publish(uuid,jsonb,text) TO pvnaive_app;
GRANT EXECUTE ON FUNCTION pvnaive.pool_revision_latest(uuid) TO pvnaive_app;
GRANT EXECUTE ON FUNCTION pvnaive.pool_node_heartbeat(uuid,text,bigint) TO pvnaive_app;
GRANT EXECUTE ON FUNCTION pvnaive.pool_node_maintenance_set(uuid,text) TO pvnaive_app;
GRANT EXECUTE ON FUNCTION pvnaive.pool_nodes_list() TO pvnaive_app;
