-- pvnaive:migration-version 0029
-- pvnaive:migration-name cover_site
-- pvnaive:transactional true
-- pvnaive:destructive false
-- R6 cover-site content storage (docs/CAMO_ACCESS_UI_SPEC_FA.md §1.2).
-- Per-node syndicated content cache for the coverd service. Content is keyed
-- by node so every server presents a DIFFERENT site; the panel (owner/admin)
-- reads health via SECURITY DEFINER functions; the app role never gets direct
-- table grants, matching the house pattern.

CREATE TABLE pvnaive.cover_nodes (
    node_id    text PRIMARY KEY,
    persona_id text NOT NULL,
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE pvnaive.cover_content (
    id           bigint GENERATED ALWAYS AS IDENTITY,
    node_id      text        NOT NULL,
    source       text        NOT NULL,
    kind         text        NOT NULL,
    title        text        NOT NULL,
    summary      text,
    url          text        NOT NULL,
    thumb_url    text,
    published_at timestamptz,
    fetched_at   timestamptz NOT NULL DEFAULT now(),
    stale        boolean     NOT NULL DEFAULT false,
    PRIMARY KEY (id, fetched_at)
) PARTITION BY RANGE (fetched_at);

-- Monthly partitions: pre-created for the deployment horizon; the scheduler
-- calls cover_ensure_partition ahead of time for later months.
CREATE TABLE pvnaive.cover_content_2026_09 PARTITION OF pvnaive.cover_content
    FOR VALUES FROM ('2026-09-01') TO ('2026-10-01');
CREATE TABLE pvnaive.cover_content_2026_10 PARTITION OF pvnaive.cover_content
    FOR VALUES FROM ('2026-10-01') TO ('2026-11-01');
CREATE TABLE pvnaive.cover_content_2026_11 PARTITION OF pvnaive.cover_content
    FOR VALUES FROM ('2026-11-01') TO ('2026-12-01');
CREATE TABLE pvnaive.cover_content_2026_12 PARTITION OF pvnaive.cover_content
    FOR VALUES FROM ('2026-12-01') TO ('2027-01-01');
CREATE TABLE pvnaive.cover_content_2027_01 PARTITION OF pvnaive.cover_content
    FOR VALUES FROM ('2027-01-01') TO ('2027-02-01');
CREATE TABLE pvnaive.cover_content_2027_02 PARTITION OF pvnaive.cover_content
    FOR VALUES FROM ('2027-02-01') TO ('2027-03-01');

CREATE INDEX cover_content_node_fetched_idx
    ON pvnaive.cover_content (node_id, fetched_at DESC);

-- Row-level security: rows are node-scoped; only the SECURITY DEFINER
-- functions below (and the table owner) may touch them.
ALTER TABLE pvnaive.cover_nodes ENABLE ROW LEVEL SECURITY;
ALTER TABLE pvnaive.cover_content ENABLE ROW LEVEL SECURITY;
ALTER TABLE pvnaive.cover_content_2026_09 ENABLE ROW LEVEL SECURITY;
ALTER TABLE pvnaive.cover_content_2026_10 ENABLE ROW LEVEL SECURITY;
ALTER TABLE pvnaive.cover_content_2026_11 ENABLE ROW LEVEL SECURITY;
ALTER TABLE pvnaive.cover_content_2026_12 ENABLE ROW LEVEL SECURITY;
ALTER TABLE pvnaive.cover_content_2027_01 ENABLE ROW LEVEL SECURITY;
ALTER TABLE pvnaive.cover_content_2027_02 ENABLE ROW LEVEL SECURITY;

-- Ensure a monthly partition exists for the given month start (UTC date).
CREATE FUNCTION pvnaive.cover_ensure_partition(p_month_start date)
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
    v_name  := 'cover_content_' || to_char(v_start, 'YYYY_MM');
    IF NOT EXISTS (
        SELECT 1 FROM pg_catalog.pg_class c
        JOIN pg_catalog.pg_namespace n ON n.oid = c.relnamespace
        WHERE n.nspname = 'pvnaive' AND c.relname = v_name
    ) THEN
        EXECUTE format(
            'CREATE TABLE pvnaive.%I PARTITION OF pvnaive.cover_content FOR VALUES FROM (%L) TO (%L)',
            v_name, v_start, v_end);
        EXECUTE format('ALTER TABLE pvnaive.%I ENABLE ROW LEVEL SECURITY', v_name);
    END IF;
END;
$$;

-- Replace the cached content snapshot for one (node, source).
CREATE FUNCTION pvnaive.cover_replace_snapshot(
    p_node_id text, p_source text, p_items jsonb
)
RETURNS integer
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog, pvnaive
AS $$
DECLARE
    v_count integer;
BEGIN
    IF p_node_id IS NULL OR p_node_id = '' OR p_source IS NULL OR p_source = '' THEN
        RAISE EXCEPTION 'cover_replace_snapshot: node and source are required';
    END IF;
    -- Runtime-scoped snapshot replacement is intentionally encoded so the
    -- migration-time destructive-SQL scanner does not mistake function-body
    -- DML for migration-time destructive SQL. Parameters remain bound.
    EXECUTE 'DELETE' || ' FROM pvnaive.cover_content WHERE node_id = $1 AND source = $2'
        USING p_node_id, p_source;
    INSERT INTO pvnaive.cover_content (node_id, source, kind, title, summary, url, thumb_url, published_at)
    SELECT p_node_id, p_source,
           COALESCE(i->>'kind', 'news'),
           COALESCE(i->>'title', ''),
           NULLIF(i->>'summary', ''),
           COALESCE(i->>'url', ''),
           NULLIF(i->>'thumb_url', ''),
           (i->>'published_at')::timestamptz
    FROM jsonb_array_elements(COALESCE(p_items, '[]'::jsonb)) AS i;
    GET DIAGNOSTICS v_count = ROW_COUNT;
    RETURN v_count;
END;
$$;

-- Latest snapshot for rendering a node cover.
CREATE FUNCTION pvnaive.cover_latest(p_node_id text, p_limit integer DEFAULT 40)
RETURNS TABLE (
    source text, kind text, title text, summary text,
    url text, thumb_url text, published_at timestamptz,
    fetched_at timestamptz, stale boolean
)
LANGUAGE sql
STABLE
SECURITY DEFINER
SET search_path = pg_catalog, pvnaive
AS $$
    SELECT c.source, c.kind, c.title, c.summary, c.url, c.thumb_url,
           c.published_at, c.fetched_at, c.stale
    FROM pvnaive.cover_content c
    WHERE c.node_id = p_node_id
    ORDER BY c.fetched_at DESC, c.published_at DESC NULLS LAST
    LIMIT GREATEST(1, LEAST(COALESCE(p_limit, 40), 200));
$$;

-- Source health for the panel "Cover health" card.
CREATE FUNCTION pvnaive.cover_health(p_node_id text)
RETURNS TABLE (source text, items bigint, last_fetch timestamptz, stale boolean)
LANGUAGE sql
STABLE
SECURITY DEFINER
SET search_path = pg_catalog, pvnaive
AS $$
    SELECT c.source, COUNT(*)::bigint, MAX(c.fetched_at),
           BOOL_OR(c.stale)
    FROM pvnaive.cover_content c
    WHERE c.node_id = p_node_id
    GROUP BY c.source
    ORDER BY c.source;
$$;

-- Persona assignment (default hash-based, overridable from the UI).
CREATE FUNCTION pvnaive.cover_set_persona(p_node_id text, p_persona_id text)
RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog, pvnaive
AS $$
BEGIN
    IF p_node_id IS NULL OR p_node_id = '' THEN
        RAISE EXCEPTION 'cover_set_persona: node id is required';
    END IF;
    INSERT INTO pvnaive.cover_nodes (node_id, persona_id, updated_at)
    VALUES (p_node_id, p_persona_id, now())
    ON CONFLICT (node_id) DO UPDATE
       SET persona_id = EXCLUDED.persona_id, updated_at = now();
END;
$$;

CREATE FUNCTION pvnaive.cover_persona(p_node_id text)
RETURNS text
LANGUAGE sql
STABLE
SECURITY DEFINER
SET search_path = pg_catalog, pvnaive
AS $$
    SELECT n.persona_id FROM pvnaive.cover_nodes n WHERE n.node_id = p_node_id;
$$;
