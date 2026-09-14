-- pvnaive:migration-version 0032
-- pvnaive:migration-name steering_decisions
-- pvnaive:transactional true
-- pvnaive:destructive false
-- R2/R3 decision-sink persistence (docs/STEERING_SPEC_FA.md §2, docs/AGENT_TASKS.md R2/R3).
--
-- The R2 engine hands actionable decisions (initial / hysteresis / kill_switch)
-- to steeringsched's DecisionSink. This migration is the durable side of that
-- hook:
--   * steering_decisions  — append-only decision audit. Rows are NEVER updated
--     or deleted by the application; the spec's "audit فقط-الحاقی" rule is
--     enforced by granting no UPDATE/DELETE anywhere and by the app role only
--     seeing the SECURITY DEFINER apply function.
--   * user_steering_state — the current (last applied) decision per user, the
--     future R5 renderer ordering input. Monotonic guard: an older window can
--     never overwrite a newer one; same-window replays (restart of the
--     in-memory engine) dedup through the audit identity index and the >=
--     guard.
-- details carries only node ids and scores (no credentials, no headers, no
-- PII beyond the user linkage accounting already holds) — redaction rule.

CREATE TABLE pvnaive.steering_decisions (
    id            BIGINT GENERATED ALWAYS AS IDENTITY,
    user_id       UUID        NOT NULL,
    node_id       TEXT        NOT NULL CHECK (length(btrim(node_id)) BETWEEN 1 AND 160),
    previous_node TEXT        CHECK (previous_node IS NULL OR length(btrim(previous_node)) BETWEEN 1 AND 160),
    window_index  BIGINT      NOT NULL CHECK (window_index >= 0),
    reason        TEXT        NOT NULL CHECK (reason IN ('initial','hysteresis','kill_switch')),
    decided_at    TIMESTAMPTZ NOT NULL,
    details       JSONB,
    recorded_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (id)
);

-- Replay identity: the in-memory engine loses state on restart and would
-- re-emit the same decision for the same window. One audit row per
-- (user, window, reason) keeps the audit honest without double rows; a
-- different reason in the same window (e.g. kill_switch after initial) is a
-- distinct, legitimate row.
CREATE UNIQUE INDEX steering_decisions_identity_idx
    ON pvnaive.steering_decisions (user_id, window_index, reason);

CREATE INDEX steering_decisions_user_time_idx
    ON pvnaive.steering_decisions (user_id, decided_at DESC);

CREATE TABLE pvnaive.user_steering_state (
    user_id       UUID        PRIMARY KEY REFERENCES pvnaive.users(id) ON DELETE CASCADE,
    node_id       TEXT        NOT NULL CHECK (length(btrim(node_id)) BETWEEN 1 AND 160),
    previous_node TEXT        CHECK (previous_node IS NULL OR length(btrim(previous_node)) BETWEEN 1 AND 160),
    window_index  BIGINT      NOT NULL CHECK (window_index >= 0),
    reason        TEXT        NOT NULL CHECK (reason IN ('initial','hysteresis','kill_switch')),
    decided_at    TIMESTAMPTZ NOT NULL,
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

REVOKE ALL ON pvnaive.steering_decisions FROM PUBLIC, pvnaive_app;
REVOKE ALL ON pvnaive.user_steering_state FROM PUBLIC, pvnaive_app;

ALTER TABLE pvnaive.steering_decisions ENABLE ROW LEVEL SECURITY;
ALTER TABLE pvnaive.user_steering_state ENABLE ROW LEVEL SECURITY;

-- Apply one actionable decision: append the audit row (idempotent on replay)
-- and advance the per-user current state behind a monotonic guard. Unknown
-- users are reported honestly as tracked=false, never fabricated.
CREATE FUNCTION pvnaive.steering_decision_apply(
    p_user_id uuid,
    p_node_id text,
    p_previous_node text,
    p_window_index bigint,
    p_reason text,
    p_decided_at timestamptz,
    p_details jsonb
)
RETURNS TABLE (tracked boolean, accepted boolean, duplicate boolean)
LANGUAGE plpgsql
VOLATILE
SECURITY DEFINER
SET search_path = pg_catalog, pvnaive
AS $$
DECLARE
    v_exists boolean;
BEGIN
    IF p_user_id IS NULL
       OR p_node_id IS NULL OR length(btrim(p_node_id)) NOT BETWEEN 1 AND 160
       OR (p_previous_node IS NOT NULL AND length(btrim(p_previous_node)) NOT BETWEEN 1 AND 160)
       OR p_window_index IS NULL OR p_window_index < 0
       OR p_reason NOT IN ('initial','hysteresis','kill_switch')
       OR p_decided_at IS NULL THEN
        RAISE EXCEPTION 'invalid steering decision' USING ERRCODE = '22023';
    END IF;

    SELECT EXISTS (SELECT 1 FROM pvnaive.users u WHERE u.id = p_user_id)
      INTO v_exists;
    IF NOT v_exists THEN
        RETURN QUERY SELECT false, false, false;
        RETURN;
    END IF;

    BEGIN
        INSERT INTO pvnaive.steering_decisions (
            user_id, node_id, previous_node, window_index, reason, decided_at, details
        ) VALUES (
            p_user_id, btrim(p_node_id), p_previous_node, p_window_index, p_reason, p_decided_at, p_details
        )
        ON CONFLICT (user_id, window_index, reason) DO NOTHING;
        IF NOT FOUND THEN
            RETURN QUERY SELECT true, false, true;
            RETURN;
        END IF;
    END;

    INSERT INTO pvnaive.user_steering_state (
        user_id, node_id, previous_node, window_index, reason, decided_at
    ) VALUES (
        p_user_id, btrim(p_node_id), p_previous_node, p_window_index, p_reason, p_decided_at
    )
    ON CONFLICT (user_id) DO UPDATE SET
        node_id       = EXCLUDED.node_id,
        previous_node = EXCLUDED.previous_node,
        window_index  = EXCLUDED.window_index,
        reason        = EXCLUDED.reason,
        decided_at    = EXCLUDED.decided_at,
        updated_at    = now()
    WHERE EXCLUDED.window_index > pvnaive.user_steering_state.window_index
       OR (EXCLUDED.window_index = pvnaive.user_steering_state.window_index
           AND EXCLUDED.decided_at >= pvnaive.user_steering_state.decided_at);

    RETURN QUERY SELECT true, true, false;
END;
$$;

-- Read the current steering state for one user. Empty result = the user has
-- no applied decision yet (honest Unknown; callers must not invent a node).
CREATE FUNCTION pvnaive.steering_state_read(p_user_id uuid)
RETURNS TABLE (
    node_id       text,
    previous_node text,
    window_index  bigint,
    reason        text,
    decided_at    timestamptz
)
LANGUAGE sql
STABLE
SECURITY DEFINER
SET search_path = pg_catalog, pvnaive
AS $$
    SELECT s.node_id, s.previous_node, s.window_index, s.reason, s.decided_at
      FROM pvnaive.user_steering_state s
     WHERE s.user_id = p_user_id;
$$;

REVOKE ALL ON FUNCTION pvnaive.steering_decision_apply(uuid,text,text,bigint,text,timestamptz,jsonb) FROM PUBLIC;
REVOKE ALL ON FUNCTION pvnaive.steering_state_read(uuid) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION pvnaive.steering_decision_apply(uuid,text,text,bigint,text,timestamptz,jsonb) TO pvnaive_app;
GRANT EXECUTE ON FUNCTION pvnaive.steering_state_read(uuid) TO pvnaive_app;
