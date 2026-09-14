-- pvnaive:migration-version 0025
-- pvnaive:migration-name notification_pipeline
-- pvnaive:transactional true
-- pvnaive:destructive false

-- The notification background worker runs without request context (no
-- tenant/actor binding), so RLS-FORCED tables fail closed for it. Following
-- the established pattern (usage_summary, release_stale_accounting_reservations),
-- the worker's exact operations are wrapped in SECURITY DEFINER functions.
-- Each function exposes one narrow operation and nothing else.

-- Read-only scan for the threshold producer: live terms with quota/expiry
-- signals, composed from the exact accounting projection.
CREATE OR REPLACE FUNCTION pvnaive.notification_scan_terms(
    p_batch_limit integer DEFAULT 200
)
RETURNS TABLE (
    service_term_id uuid,
    quota_bytes bigint,
    used_bytes bigint,
    expires_at timestamptz
)
LANGUAGE sql
STABLE
SECURITY DEFINER
SET search_path = pg_catalog, pvnaive
AS $$
    SELECT st.id,
           st.quota_bytes,
           (COALESCE(t.upload_bytes, 0) + COALESCE(t.download_bytes, 0)),
           st.expires_at
      FROM pvnaive.service_terms st
      LEFT JOIN pvnaive.direct_naive_accounting_terms t
             ON t.service_term_id = st.id
     WHERE st.state IN ('active', 'pending')
       AND (st.quota_bytes IS NOT NULL OR st.expires_at IS NOT NULL)
     ORDER BY st.id
     LIMIT GREATEST(1, LEAST(GREATEST(p_batch_limit, 1), 1000));
$$;

-- Seed the baseline rule set once. Insert-if-absent so admin edits survive.
CREATE OR REPLACE FUNCTION pvnaive.notification_seed_defaults()
RETURNS integer
LANGUAGE plpgsql
VOLATILE
SECURITY DEFINER
SET search_path = pg_catalog, pvnaive
AS $$
DECLARE
    created integer := 0;
    candidate text;
BEGIN
    FOREACH candidate IN ARRAY ARRAY[
        'usage_80', 'usage_95', 'usage_exhausted',
        'expiry_7d', 'expiry_3d', 'expiry_1d', 'expired',
        'runtime_down', 'backup_failed', 'reseller_credit_low', 'tls_expiring'
    ] LOOP
        IF NOT EXISTS (SELECT 1 FROM pvnaive.notification_rules WHERE event_type = candidate) THEN
            INSERT INTO pvnaive.notification_rules
                (tenant_id, event_type, enabled, channels, max_attempts, retry_base_seconds)
            VALUES (NULL, candidate, true, '["in_app"]'::jsonb, 5, 60);
            created := created + 1;
        END IF;
    END LOOP;
    RETURN created;
END;
$$;

-- Enqueue one notification when (and only when) an enabled rule exists for
-- the event type. Dedup is enforced by the outbox unique index.
CREATE OR REPLACE FUNCTION pvnaive.notification_enqueue(
    p_event_type text,
    p_aggregate_type text,
    p_aggregate_id uuid,
    p_deduplication_key text,
    p_payload jsonb,
    p_now timestamptz
)
RETURNS uuid
LANGUAGE plpgsql
VOLATILE
SECURITY DEFINER
SET search_path = pg_catalog, pvnaive
AS $$
DECLARE
    v_rule_id uuid;
    v_max_attempts smallint;
    v_outbox_id uuid;
BEGIN
    SELECT id, max_attempts INTO v_rule_id, v_max_attempts
      FROM pvnaive.notification_rules
     WHERE event_type = p_event_type AND enabled = true
     ORDER BY (tenant_id IS NOT NULL), created_at ASC
     LIMIT 1;
    IF v_rule_id IS NULL THEN
        RETURN NULL;
    END IF;
    INSERT INTO pvnaive.notification_outbox
        (tenant_id, rule_id, event_type, aggregate_type, aggregate_id,
         deduplication_key, payload, status, max_attempts, next_attempt_at)
    VALUES (NULL, v_rule_id, p_event_type, p_aggregate_type, p_aggregate_id,
            p_deduplication_key, p_payload, 'pending', v_max_attempts, p_now)
    ON CONFLICT (COALESCE(tenant_id, '00000000-0000-0000-0000-000000000000'::uuid), deduplication_key)
    DO NOTHING
    RETURNING id INTO v_outbox_id;
    RETURN v_outbox_id;
END;
$$;

-- Claim due outbox rows for dispatch. SKIP LOCKED keeps parallel replicas
-- from double-sending.
CREATE OR REPLACE FUNCTION pvnaive.notification_claim_due(
    p_now timestamptz,
    p_batch integer
)
RETURNS TABLE (
    outbox_id uuid,
    event_type text,
    channels text,
    attempt_count smallint,
    max_attempts smallint,
    retry_base_seconds integer,
    payload text
)
LANGUAGE plpgsql
VOLATILE
SECURITY DEFINER
SET search_path = pg_catalog, pvnaive
AS $$
BEGIN
    RETURN QUERY
    WITH due AS (
        SELECT o.id
          FROM pvnaive.notification_outbox o
         WHERE o.status IN ('pending', 'failed')
           AND o.next_attempt_at <= p_now
           AND o.attempt_count < o.max_attempts
         ORDER BY o.next_attempt_at ASC
         LIMIT GREATEST(1, LEAST(GREATEST(p_batch, 1), 200))
         FOR UPDATE SKIP LOCKED
    )
    UPDATE pvnaive.notification_outbox o
       SET status = 'processing', locked_at = p_now, locked_by = 'api-dispatcher'
      FROM due,
           pvnaive.notification_rules r
     WHERE o.id = due.id
       AND r.id = o.rule_id
    RETURNING o.id,
              o.event_type,
              COALESCE(r.channels::text, '["in_app"]'),
              o.attempt_count,
              o.max_attempts,
              r.retry_base_seconds,
              o.payload::text;
END;
$$;

-- Record one channel delivery attempt (idempotent per attempt).
CREATE OR REPLACE FUNCTION pvnaive.notification_record_delivery(
    p_outbox_id uuid,
    p_channel text,
    p_attempt_no smallint,
    p_idempotency_key text,
    p_status text,
    p_error_code text,
    p_now timestamptz
)
RETURNS void
LANGUAGE plpgsql
VOLATILE
SECURITY DEFINER
SET search_path = pg_catalog, pvnaive
AS $$
BEGIN
    INSERT INTO pvnaive.notification_deliveries
        (outbox_id, channel, attempt_no, idempotency_key, status, error_code,
         attempted_at, delivered_at)
    VALUES
        (p_outbox_id, p_channel, p_attempt_no, p_idempotency_key, p_status,
         NULLIF(p_error_code, ''), p_now,
         CASE WHEN p_status = 'delivered' THEN p_now END)
    ON CONFLICT (outbox_id, channel, attempt_no) DO NOTHING;
END;
$$;

-- Settle a claimed row: delivered, terminal failure, or scheduled retry.
CREATE OR REPLACE FUNCTION pvnaive.notification_settle(
    p_outbox_id uuid,
    p_delivered boolean,
    p_terminal boolean,
    p_next_attempt timestamptz,
    p_error_code text,
    p_now timestamptz
)
RETURNS void
LANGUAGE plpgsql
VOLATILE
SECURITY DEFINER
SET search_path = pg_catalog, pvnaive
AS $$
BEGIN
    IF p_delivered THEN
        UPDATE pvnaive.notification_outbox
           SET status = 'delivered', processed_at = p_now, locked_by = NULL,
               locked_at = NULL, last_error_code = NULL,
               attempt_count = attempt_count + 1
         WHERE id = p_outbox_id;
    ELSIF p_terminal THEN
        UPDATE pvnaive.notification_outbox
           SET status = 'failed', processed_at = p_now, locked_by = NULL,
               locked_at = NULL, last_error_code = NULLIF(p_error_code, ''),
               attempt_count = attempt_count + 1
         WHERE id = p_outbox_id;
    ELSE
        UPDATE pvnaive.notification_outbox
           SET status = 'failed', locked_by = NULL, locked_at = NULL,
               last_error_code = NULLIF(p_error_code, ''),
               attempt_count = attempt_count + 1,
               next_attempt_at = COALESCE(p_next_attempt, p_now + interval '1 minute')
         WHERE id = p_outbox_id;
    END IF;
END;
$$;

REVOKE ALL ON FUNCTION pvnaive.notification_scan_terms(integer) FROM PUBLIC;
REVOKE ALL ON FUNCTION pvnaive.notification_seed_defaults() FROM PUBLIC;
REVOKE ALL ON FUNCTION pvnaive.notification_enqueue(text,text,uuid,text,jsonb,timestamptz) FROM PUBLIC;
REVOKE ALL ON FUNCTION pvnaive.notification_claim_due(timestamptz,integer) FROM PUBLIC;
REVOKE ALL ON FUNCTION pvnaive.notification_record_delivery(uuid,text,smallint,text,text,text,timestamptz) FROM PUBLIC;
REVOKE ALL ON FUNCTION pvnaive.notification_settle(uuid,boolean,boolean,timestamptz,text,timestamptz) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION pvnaive.notification_scan_terms(integer) TO pvnaive_app;
GRANT EXECUTE ON FUNCTION pvnaive.notification_seed_defaults() TO pvnaive_app;
GRANT EXECUTE ON FUNCTION pvnaive.notification_enqueue(text,text,uuid,text,jsonb,timestamptz) TO pvnaive_app;
GRANT EXECUTE ON FUNCTION pvnaive.notification_claim_due(timestamptz,integer) TO pvnaive_app;
GRANT EXECUTE ON FUNCTION pvnaive.notification_record_delivery(uuid,text,smallint,text,text,text,timestamptz) TO pvnaive_app;
GRANT EXECUTE ON FUNCTION pvnaive.notification_settle(uuid,boolean,boolean,timestamptz,text,timestamptz) TO pvnaive_app;
