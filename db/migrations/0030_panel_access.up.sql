-- pvnaive:migration-version 0030
-- pvnaive:migration-name panel_access
-- pvnaive:transactional true
-- pvnaive:destructive false
-- R7 panel access management state (docs/CAMO_ACCESS_UI_SPEC_FA.md §2.1).
-- Single-row settings for admin identity, base path, port and exposure mode.
-- Starts empty: the Go store falls back to the existing actors table until
-- the first in-panel change is applied, so deployment behavior is unchanged.

CREATE TABLE pvnaive.panel_settings (
    id             integer PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    admin_username text    NOT NULL CHECK (length(btrim(admin_username)) BETWEEN 3 AND 64),
    password_hash  text    NOT NULL CHECK (password_hash LIKE '$argon2id$%'),
    base_path      text    NOT NULL DEFAULT '/panel'
                   CHECK (base_path ~ '^/[a-z0-9][a-z0-9\-_]{2,63}$'),
    listen_port    integer NOT NULL DEFAULT 8080 CHECK (listen_port BETWEEN 1 AND 65535),
    session_ttl    interval NOT NULL DEFAULT interval '12 hours',
    grace_minutes  integer  NOT NULL DEFAULT 10 CHECK (grace_minutes BETWEEN 1 AND 60),
    exposure_mode  text     NOT NULL DEFAULT 'reverse_proxy'
                   CHECK (exposure_mode IN ('reverse_proxy','direct')),
    updated_at     timestamptz NOT NULL DEFAULT now(),
    updated_by     uuid
);

ALTER TABLE pvnaive.panel_settings ENABLE ROW LEVEL SECURITY;

-- Append-only audit for every access change. Secrets are NEVER recorded:
-- password changes log a fixed redaction marker as new_value.
CREATE TABLE pvnaive.panel_access_audit (
    id        bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    actor_id  uuid,
    field     text        NOT NULL,
    old_value text,
    new_value text,
    at        timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX panel_access_audit_at_idx ON pvnaive.panel_access_audit (at DESC);

ALTER TABLE pvnaive.panel_access_audit ENABLE ROW LEVEL SECURITY;

-- Read the settings row (STABLE, SECURITY DEFINER). Returns no rows when the
-- operator has never changed access state; the caller falls back to actors.
CREATE FUNCTION pvnaive.panel_settings_read()
RETURNS SETOF pvnaive.panel_settings
LANGUAGE sql
STABLE
SECURITY DEFINER
SET search_path = pg_catalog, pvnaive
AS $$
    SELECT * FROM pvnaive.panel_settings WHERE id = 1;
$$;

-- Upsert settings and record the per-field audit trail. Old values come from
-- the pre-update row (if any). Password hash changes are audited with a
-- redaction marker, never the hash itself.
CREATE FUNCTION pvnaive.panel_settings_upsert(
    p_admin_username text,
    p_password_hash  text,
    p_base_path      text,
    p_listen_port    integer,
    p_grace_minutes  integer,
    p_exposure_mode  text,
    p_actor_id       uuid
)
RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog, pvnaive
AS $$
DECLARE
    v_old record;
    v_new_hash text := NULLIF(p_password_hash, '');
BEGIN
    IF p_admin_username IS NULL OR length(btrim(p_admin_username)) NOT BETWEEN 3 AND 64 THEN
        RAISE EXCEPTION 'panel_settings_upsert: admin username rejected';
    END IF;
    IF v_new_hash IS NOT NULL AND v_new_hash NOT LIKE '$argon2id$%' THEN
        RAISE EXCEPTION 'panel_settings_upsert: password hash format rejected';
    END IF;
    IF p_base_path IS NULL OR p_base_path !~ '^/[a-z0-9][a-z0-9\-_]{2,63}$' THEN
        RAISE EXCEPTION 'panel_settings_upsert: base path rejected';
    END IF;
    IF p_listen_port IS NULL OR p_listen_port NOT BETWEEN 1 AND 65535 THEN
        RAISE EXCEPTION 'panel_settings_upsert: listen port rejected';
    END IF;
    IF p_grace_minutes IS NULL OR p_grace_minutes NOT BETWEEN 1 AND 60 THEN
        RAISE EXCEPTION 'panel_settings_upsert: grace minutes rejected';
    END IF;
    IF p_exposure_mode NOT IN ('reverse_proxy','direct') THEN
        RAISE EXCEPTION 'panel_settings_upsert: exposure mode rejected';
    END IF;

    SELECT * INTO v_old FROM pvnaive.panel_settings WHERE id = 1;

    INSERT INTO pvnaive.panel_settings (
        id, admin_username, password_hash, base_path, listen_port,
        grace_minutes, exposure_mode, updated_at, updated_by
    ) VALUES (
        1, btrim(p_admin_username),
        COALESCE(v_new_hash, v_old.password_hash, ''),
        p_base_path, p_listen_port, p_grace_minutes, p_exposure_mode, now(), p_actor_id
    )
    ON CONFLICT (id) DO UPDATE SET
        admin_username = EXCLUDED.admin_username,
        password_hash  = COALESCE(EXCLUDED.password_hash, pvnaive.panel_settings.password_hash),
        base_path      = EXCLUDED.base_path,
        listen_port    = EXCLUDED.listen_port,
        grace_minutes  = EXCLUDED.grace_minutes,
        exposure_mode  = EXCLUDED.exposure_mode,
        updated_at     = now(),
        updated_by     = EXCLUDED.updated_by;

    IF v_old IS NULL OR v_old.admin_username IS DISTINCT FROM btrim(p_admin_username) THEN
        INSERT INTO pvnaive.panel_access_audit (actor_id, field, old_value, new_value)
        VALUES (p_actor_id, 'admin_username', v_old.admin_username, btrim(p_admin_username));
    END IF;
    IF v_new_hash IS NOT NULL THEN
        INSERT INTO pvnaive.panel_access_audit (actor_id, field, old_value, new_value)
        VALUES (p_actor_id, 'password_hash', '[redacted]', '[redacted]');
    END IF;
    IF v_old IS NULL OR v_old.base_path IS DISTINCT FROM p_base_path THEN
        INSERT INTO pvnaive.panel_access_audit (actor_id, field, old_value, new_value)
        VALUES (p_actor_id, 'base_path', v_old.base_path, p_base_path);
    END IF;
    IF v_old IS NULL OR v_old.listen_port IS DISTINCT FROM p_listen_port THEN
        INSERT INTO pvnaive.panel_access_audit (actor_id, field, old_value, new_value)
        VALUES (p_actor_id, 'listen_port', v_old.listen_port::text, p_listen_port::text);
    END IF;
    IF v_old IS NULL OR v_old.exposure_mode IS DISTINCT FROM p_exposure_mode THEN
        INSERT INTO pvnaive.panel_access_audit (actor_id, field, old_value, new_value)
        VALUES (p_actor_id, 'exposure_mode', v_old.exposure_mode, p_exposure_mode);
    END IF;
END;
$$;
