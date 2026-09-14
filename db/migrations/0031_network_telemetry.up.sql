-- pvnaive:migration-version 0031
-- pvnaive:migration-name network_telemetry
-- pvnaive:transactional true
-- pvnaive:destructive false
-- R1 / STEER-001 trusted-boundary network telemetry (docs/STEERING_SPEC_FA.md §1).
--
-- Raw per-session TCP_INFO samples collected inside the pinned forwardproxy and
-- delivered through the same trusted Unix-socket boundary as exact accounting.
-- Identity rule: samples are keyed by runtime_credential_id and resolved to
-- user_id by the SAME join used by exact accounting — never by client headers.
--
-- Documented, deliberate refinements over the spec DDL draft (see
-- docs/STEERING_SPEC_FA.md §1 and the 0031 section of docs/AGENT_TASKS.md):
--   * session_id uuid added: the spec draft's (boot_id, session_seq) key cannot
--     distinguish two concurrent proxy sessions inside one boot; the accounting
--     session UUID + a per-session monotonic sample_seq is the precise identity.
--   * path column ('client' | 'upstream'): records which TCP path the sample
--     came from. 'upstream' samples (node <-> destination) are always available;
--     'client' samples (client <-> node, HTTP/1 hijack path) are the
--     user-perceived path when present.
--   * EWMA aggregation happens in the telemetry-agent holder (Go), exactly as
--     the spec's "فرمول به‌روزرسانی سرِ نگهدارنده، نه دیتابیس" rule requires.
--     The database only stores raw samples and persists holder-produced
--     aggregate rows behind a monotonic upsert.

CREATE TABLE pvnaive.session_network_samples (
    id             BIGINT GENERATED ALWAYS AS IDENTITY,
    user_id        UUID        NOT NULL,
    runtime_credential_id UUID NOT NULL,
    node_id        TEXT        NOT NULL,
    boot_id        UUID        NOT NULL,
    session_id     UUID        NOT NULL,
    sample_seq     BIGINT      NOT NULL CHECK (sample_seq >= 1),
    sampled_at     TIMESTAMPTZ NOT NULL,
    path           TEXT        NOT NULL CHECK (path IN ('client','upstream')),
    rtt_micros     INTEGER     NOT NULL CHECK (rtt_micros  >= 0),
    rtt_var_micros INTEGER     NOT NULL CHECK (rtt_var_micros >= 0),
    segs_out       BIGINT      NOT NULL CHECK (segs_out >= 0),
    segs_retrans   BIGINT      NOT NULL CHECK (segs_retrans >= 0 AND segs_retrans <= segs_out),
    bytes_in       BIGINT      NOT NULL CHECK (bytes_in >= 0),
    bytes_out      BIGINT      NOT NULL CHECK (bytes_out >= 0),
    duration_micros BIGINT     NOT NULL CHECK (duration_micros >= 0),
    PRIMARY KEY (id, sampled_at)
) PARTITION BY RANGE (sampled_at);

-- Idempotent ingest: a replayed batch (restart/reload) must never double-count
-- (STEER-001). One sample per (boot, session, seq, sampled_at). The partition
-- key must be part of any unique index on a partitioned table; the sampler
-- guarantees replay-identical payloads (buffered batches keep their original
-- timestamps), so identical replays dedup exactly. This matches the spec's
-- uniqueness rule ("uniqueness روی boot_id, session_seq, sampled_at").
CREATE UNIQUE INDEX session_network_samples_identity_idx
    ON pvnaive.session_network_samples (boot_id, session_id, sample_seq, sampled_at);
CREATE INDEX session_network_samples_user_path_idx
    ON pvnaive.session_network_samples (user_id, node_id, path, sampled_at DESC);

CREATE TABLE pvnaive.session_network_samples_2026_09 PARTITION OF pvnaive.session_network_samples
    FOR VALUES FROM ('2026-09-01') TO ('2026-10-01');
CREATE TABLE pvnaive.session_network_samples_2026_10 PARTITION OF pvnaive.session_network_samples
    FOR VALUES FROM ('2026-10-01') TO ('2026-11-01');
CREATE TABLE pvnaive.session_network_samples_2026_11 PARTITION OF pvnaive.session_network_samples
    FOR VALUES FROM ('2026-11-01') TO ('2026-12-01');
CREATE TABLE pvnaive.session_network_samples_2026_12 PARTITION OF pvnaive.session_network_samples
    FOR VALUES FROM ('2026-12-01') TO ('2027-01-01');
CREATE TABLE pvnaive.session_network_samples_2027_01 PARTITION OF pvnaive.session_network_samples
    FOR VALUES FROM ('2027-01-01') TO ('2027-02-01');
CREATE TABLE pvnaive.session_network_samples_2027_02 PARTITION OF pvnaive.session_network_samples
    FOR VALUES FROM ('2027-02-01') TO ('2027-03-01');

-- Holder-persisted EWMA aggregates, one row per (user, node, path).
CREATE TABLE pvnaive.user_node_network_agg (
    user_id            UUID   NOT NULL,
    node_id            TEXT   NOT NULL,
    path               TEXT   NOT NULL CHECK (path IN ('client','upstream')),
    rtt_ewma_micros    BIGINT NOT NULL CHECK (rtt_ewma_micros >= 0),
    jitter_ewma_micros BIGINT NOT NULL CHECK (jitter_ewma_micros >= 0),
    retrans_ratio_ewma REAL   NOT NULL CHECK (retrans_ratio_ewma >= 0 AND retrans_ratio_ewma <= 1),
    throughput_bps     BIGINT NOT NULL CHECK (throughput_bps >= 0),
    success_rate_ewma  REAL   NOT NULL CHECK (success_rate_ewma >= 0 AND success_rate_ewma <= 1),
    sample_count       BIGINT NOT NULL CHECK (sample_count >= 0),
    last_sampled_at    TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (user_id, node_id, path)
);

-- Telemetry relations are a trusted data-path boundary: browser/API callers get
-- no direct table privileges. The only app-role entry points are the narrowly
-- typed SECURITY DEFINER functions below.
REVOKE ALL ON pvnaive.session_network_samples FROM PUBLIC, pvnaive_app;
REVOKE ALL ON pvnaive.user_node_network_agg FROM PUBLIC, pvnaive_app;
REVOKE ALL ON TABLE pvnaive.session_network_samples_2026_09 FROM PUBLIC, pvnaive_app;
REVOKE ALL ON TABLE pvnaive.session_network_samples_2026_10 FROM PUBLIC, pvnaive_app;
REVOKE ALL ON TABLE pvnaive.session_network_samples_2026_11 FROM PUBLIC, pvnaive_app;
REVOKE ALL ON TABLE pvnaive.session_network_samples_2026_12 FROM PUBLIC, pvnaive_app;
REVOKE ALL ON TABLE pvnaive.session_network_samples_2027_01 FROM PUBLIC, pvnaive_app;
REVOKE ALL ON TABLE pvnaive.session_network_samples_2027_02 FROM PUBLIC, pvnaive_app;

ALTER TABLE pvnaive.session_network_samples ENABLE ROW LEVEL SECURITY;
ALTER TABLE pvnaive.user_node_network_agg ENABLE ROW LEVEL SECURITY;
ALTER TABLE pvnaive.session_network_samples_2026_09 ENABLE ROW LEVEL SECURITY;
ALTER TABLE pvnaive.session_network_samples_2026_10 ENABLE ROW LEVEL SECURITY;
ALTER TABLE pvnaive.session_network_samples_2026_11 ENABLE ROW LEVEL SECURITY;
ALTER TABLE pvnaive.session_network_samples_2026_12 ENABLE ROW LEVEL SECURITY;
ALTER TABLE pvnaive.session_network_samples_2027_01 ENABLE ROW LEVEL SECURITY;
ALTER TABLE pvnaive.session_network_samples_2027_02 ENABLE ROW LEVEL SECURITY;

-- Runtime partition creation so ingest never fails at a month boundary
-- (same pattern as pvnaive.cover_ensure_partition in 0029).
CREATE FUNCTION pvnaive.network_ensure_partition(p_month_start date)
RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog, pvnaive
AS $$
DECLARE
    v_name  text;
    v_start date;
    v_end   date;
BEGIN
    v_start := date_trunc('month', p_month_start)::date;
    v_end   := (v_start + interval '1 month')::date;
    v_name  := 'session_network_samples_' || to_char(v_start, 'YYYY_MM');
    IF NOT EXISTS (
        SELECT 1 FROM pg_catalog.pg_class c
        JOIN pg_catalog.pg_namespace n ON n.oid = c.relnamespace
        WHERE n.nspname = 'pvnaive' AND c.relname = v_name
    ) THEN
        EXECUTE format(
            'CREATE TABLE pvnaive.%I PARTITION OF pvnaive.session_network_samples FOR VALUES FROM (%L) TO (%L)',
            v_name, v_start, v_end);
        EXECUTE format('REVOKE ALL ON TABLE pvnaive.%I FROM PUBLIC, pvnaive_app', v_name);
        EXECUTE format('ALTER TABLE pvnaive.%I ENABLE ROW LEVEL SECURITY', v_name);
    END IF;
END;
$$;

-- Ingest one sample. Identity and user resolution follow the exact-accounting
-- join. Unknown credentials are reported as tracked=false (fail-closed: the
-- caller must not treat them as known telemetry), never guessed.
CREATE FUNCTION pvnaive.network_sample_ingest(
    p_runtime_credential_id uuid,
    p_node_id text,
    p_boot_id uuid,
    p_session_id uuid,
    p_sample_seq bigint,
    p_sampled_at timestamptz,
    p_path text,
    p_rtt_micros integer,
    p_rtt_var_micros integer,
    p_segs_out bigint,
    p_segs_retrans bigint,
    p_bytes_in bigint,
    p_bytes_out bigint,
    p_duration_micros bigint
)
RETURNS TABLE (
    tracked boolean,
    accepted boolean,
    duplicate boolean,
    reason text,
    user_id uuid
)
LANGUAGE plpgsql
VOLATILE
SECURITY DEFINER
SET search_path = pg_catalog, pvnaive
AS $$
DECLARE
    v_user_id uuid;
    v_user_tenant uuid;
BEGIN
    IF p_runtime_credential_id IS NULL OR p_boot_id IS NULL OR p_session_id IS NULL
       OR p_node_id IS NULL OR length(btrim(p_node_id)) NOT BETWEEN 1 AND 160
       OR p_sample_seq < 1 OR p_sampled_at IS NULL
       OR p_path NOT IN ('client','upstream')
       OR p_rtt_micros IS NULL OR p_rtt_micros < 0
       OR p_rtt_var_micros IS NULL OR p_rtt_var_micros < 0
       OR p_segs_out IS NULL OR p_segs_out < 0
       OR p_segs_retrans IS NULL OR p_segs_retrans < 0 OR p_segs_retrans > p_segs_out
       OR p_bytes_in IS NULL OR p_bytes_in < 0
       OR p_bytes_out IS NULL OR p_bytes_out < 0
       OR p_duration_micros IS NULL OR p_duration_micros < 0 THEN
        RAISE EXCEPTION 'invalid network sample' USING ERRCODE = '22023';
    END IF;
    PERFORM pvnaive.direct_naive_accounting_enter_context();

    SELECT u.id, u.tenant_id
      INTO v_user_id, v_user_tenant
      FROM pvnaive.user_runtime_credentials urc
      JOIN pvnaive.naive_runtime_credentials rc
        ON rc.id = urc.runtime_credential_id AND rc.status = 'active'
      JOIN pvnaive.users u
        ON u.id = urc.user_id AND u.tenant_id = urc.tenant_id AND u.status = 'active'
     WHERE urc.runtime_credential_id = p_runtime_credential_id
       AND urc.unbound_at IS NULL AND urc.role = 'primary'
     LIMIT 1;

    IF NOT FOUND THEN
        PERFORM pvnaive.direct_naive_accounting_leave_context();
        RETURN QUERY SELECT false, false, false, 'untracked'::text, NULL::uuid;
        RETURN;
    END IF;

    PERFORM pvnaive.network_ensure_partition(p_sampled_at::date);

    INSERT INTO pvnaive.session_network_samples (
        user_id, runtime_credential_id, node_id, boot_id, session_id, sample_seq,
        sampled_at, path, rtt_micros, rtt_var_micros,
        segs_out, segs_retrans, bytes_in, bytes_out, duration_micros
    )
    VALUES (
        v_user_id, p_runtime_credential_id, btrim(p_node_id), p_boot_id, p_session_id, p_sample_seq,
        p_sampled_at, p_path, p_rtt_micros, p_rtt_var_micros,
        p_segs_out, p_segs_retrans, p_bytes_in, p_bytes_out, p_duration_micros
    )
    ON CONFLICT (boot_id, session_id, sample_seq, sampled_at) DO NOTHING;

    IF NOT FOUND THEN
        PERFORM pvnaive.direct_naive_accounting_leave_context();
        RETURN QUERY SELECT true, false, true, 'duplicate'::text, v_user_id;
        RETURN;
    END IF;

    PERFORM pvnaive.direct_naive_accounting_leave_context();
    RETURN QUERY SELECT true, true, false, 'accepted'::text, v_user_id;
END;
$$;

-- Persist holder-computed EWMA aggregates. Guard: never let an OLDER batch
-- overwrite a NEWER one (restart/replay safety for aggregates).
CREATE FUNCTION pvnaive.network_agg_upsert(
    p_user_id uuid,
    p_node_id text,
    p_path text,
    p_rtt_ewma_micros bigint,
    p_jitter_ewma_micros bigint,
    p_retrans_ratio_ewma real,
    p_throughput_bps bigint,
    p_success_rate_ewma real,
    p_sample_count bigint,
    p_last_sampled_at timestamptz
)
RETURNS boolean
LANGUAGE plpgsql
VOLATILE
SECURITY DEFINER
SET search_path = pg_catalog, pvnaive
AS $$
BEGIN
    IF p_user_id IS NULL OR p_node_id IS NULL OR length(btrim(p_node_id)) NOT BETWEEN 1 AND 160
       OR p_path NOT IN ('client','upstream')
       OR p_rtt_ewma_micros IS NULL OR p_rtt_ewma_micros < 0
       OR p_jitter_ewma_micros IS NULL OR p_jitter_ewma_micros < 0
       OR p_retrans_ratio_ewma IS NULL OR p_retrans_ratio_ewma < 0 OR p_retrans_ratio_ewma > 1
       OR p_throughput_bps IS NULL OR p_throughput_bps < 0
       OR p_success_rate_ewma IS NULL OR p_success_rate_ewma < 0 OR p_success_rate_ewma > 1
       OR p_sample_count IS NULL OR p_sample_count < 0
       OR p_last_sampled_at IS NULL THEN
        RAISE EXCEPTION 'invalid network aggregate' USING ERRCODE = '22023';
    END IF;

    INSERT INTO pvnaive.user_node_network_agg (
        user_id, node_id, path, rtt_ewma_micros, jitter_ewma_micros,
        retrans_ratio_ewma, throughput_bps, success_rate_ewma,
        sample_count, last_sampled_at
    )
    VALUES (
        p_user_id, btrim(p_node_id), p_path, p_rtt_ewma_micros, p_jitter_ewma_micros,
        p_retrans_ratio_ewma, p_throughput_bps, p_success_rate_ewma,
        p_sample_count, p_last_sampled_at
    )
    ON CONFLICT (user_id, node_id, path) DO UPDATE SET
        rtt_ewma_micros    = EXCLUDED.rtt_ewma_micros,
        jitter_ewma_micros = EXCLUDED.jitter_ewma_micros,
        retrans_ratio_ewma = EXCLUDED.retrans_ratio_ewma,
        throughput_bps     = EXCLUDED.throughput_bps,
        success_rate_ewma  = EXCLUDED.success_rate_ewma,
        sample_count       = EXCLUDED.sample_count,
        last_sampled_at    = EXCLUDED.last_sampled_at
    WHERE EXCLUDED.last_sampled_at > pvnaive.user_node_network_agg.last_sampled_at;

    RETURN TRUE;
END;
$$;

-- Steering reads only fresh aggregates. Staleness is explicit, never fabricated:
-- stale rows are omitted so the R2 engine sees an honest Unknown.
CREATE FUNCTION pvnaive.network_agg_read(p_stale_after_seconds bigint)
RETURNS SETOF pvnaive.user_node_network_agg
LANGUAGE sql
STABLE
SECURITY DEFINER
SET search_path = pg_catalog, pvnaive
AS $$
    SELECT *
      FROM pvnaive.user_node_network_agg
     WHERE last_sampled_at >= clock_timestamp() - make_interval(secs => GREATEST(p_stale_after_seconds, 1))
     ORDER BY user_id, node_id, path;
$$;

REVOKE ALL ON FUNCTION pvnaive.network_ensure_partition(date) FROM PUBLIC, pvnaive_app;
REVOKE ALL ON FUNCTION pvnaive.network_sample_ingest(uuid,text,uuid,uuid,bigint,timestamptz,text,integer,integer,bigint,bigint,bigint,bigint,bigint) FROM PUBLIC;
REVOKE ALL ON FUNCTION pvnaive.network_agg_upsert(uuid,text,text,bigint,bigint,real,bigint,real,bigint,timestamptz) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION pvnaive.network_sample_ingest(uuid,text,uuid,uuid,bigint,timestamptz,text,integer,integer,bigint,bigint,bigint,bigint,bigint) TO pvnaive_app;
GRANT EXECUTE ON FUNCTION pvnaive.network_agg_upsert(uuid,text,text,bigint,bigint,real,bigint,real,bigint,timestamptz) TO pvnaive_app;
GRANT EXECUTE ON FUNCTION pvnaive.network_agg_read(bigint) TO pvnaive_app;
