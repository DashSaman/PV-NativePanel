-- pvnaive:migration-version 0016
-- Source: PVNaive Task #9 restart-safe periodic exact-accounting usage reset
-- pvnaive:migration-name periodic_usage_reset
-- pvnaive:transactional true
-- pvnaive:destructive false

SET LOCAL lock_timeout = '10s';
SET LOCAL statement_timeout = '120s';
SET LOCAL idle_in_transaction_session_timeout = '60s';
SET LOCAL ROLE pvnaive_owner;

CREATE FUNCTION pvnaive.next_usage_reset_due(
    p_anchor timestamptz,
    p_strategy text,
    p_custom_days integer DEFAULT NULL
)
RETURNS timestamptz
LANGUAGE sql
IMMUTABLE
SET search_path = pg_catalog, pvnaive
AS $$
    SELECT CASE p_strategy
        WHEN 'daily' THEN p_anchor + interval '1 day'
        WHEN 'weekly' THEN p_anchor + interval '7 days'
        WHEN 'monthly' THEN ((p_anchor AT TIME ZONE 'UTC') + interval '1 month') AT TIME ZONE 'UTC'
        WHEN 'yearly' THEN ((p_anchor AT TIME ZONE 'UTC') + interval '1 year') AT TIME ZONE 'UTC'
        WHEN 'custom' THEN p_anchor + make_interval(days => p_custom_days)
        ELSE NULL::timestamptz
    END;
$$;
REVOKE ALL ON FUNCTION pvnaive.next_usage_reset_due(timestamptz,text,integer) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION pvnaive.next_usage_reset_due(timestamptz,text,integer) TO pvnaive_app;

CREATE TABLE pvnaive.service_term_reset_schedules (
    service_term_id uuid PRIMARY KEY,
    tenant_id uuid NOT NULL,
    user_id uuid NOT NULL,
    strategy text NOT NULL CHECK (strategy IN ('none','daily','weekly','monthly','yearly','custom')),
    custom_days integer CHECK (custom_days IS NULL OR (custom_days > 0 AND custom_days <= 3660)),
    timezone_name text NOT NULL DEFAULT 'UTC' CHECK (timezone_name = 'UTC'),
    anchor_at timestamptz,
    next_due_at timestamptz,
    last_attempt_at timestamptz,
    last_completed_at timestamptz,
    retry_after_at timestamptz,
    last_error text CHECK (last_error IS NULL OR char_length(last_error) <= 160),
    consecutive_failures integer NOT NULL DEFAULT 0 CHECK (consecutive_failures >= 0),
    created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    updated_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    FOREIGN KEY (service_term_id, tenant_id, user_id)
        REFERENCES pvnaive.service_terms(id, tenant_id, user_id) ON DELETE RESTRICT,
    CHECK (
        (strategy = 'custom' AND custom_days IS NOT NULL)
        OR (strategy <> 'custom' AND custom_days IS NULL)
    ),
    CHECK (strategy <> 'none' OR next_due_at IS NULL),
    CHECK (anchor_at IS NOT NULL OR next_due_at IS NULL),
    CHECK (next_due_at IS NULL OR next_due_at > anchor_at),
    CHECK (last_completed_at IS NULL OR anchor_at IS NULL OR last_completed_at >= anchor_at)
);
CREATE INDEX service_term_reset_schedules_due_idx
    ON pvnaive.service_term_reset_schedules(next_due_at, retry_after_at, service_term_id)
    WHERE next_due_at IS NOT NULL;
CREATE INDEX service_term_reset_schedules_tenant_user_idx
    ON pvnaive.service_term_reset_schedules(tenant_id, user_id);

CREATE TABLE pvnaive.scheduled_usage_reset_attempts (
    id uuid PRIMARY KEY DEFAULT public.gen_random_uuid(),
    service_term_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    user_id uuid NOT NULL,
    scheduled_due_at timestamptz NOT NULL,
    attempted_at timestamptz NOT NULL,
    outcome text NOT NULL CHECK (outcome IN ('success','deferred','skipped')),
    reason text NOT NULL CHECK (btrim(reason) <> '' AND char_length(reason) <= 160),
    reset_event_id uuid REFERENCES pvnaive.direct_naive_accounting_reset_events(id) ON DELETE RESTRICT,
    created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    FOREIGN KEY (service_term_id, tenant_id, user_id)
        REFERENCES pvnaive.service_terms(id, tenant_id, user_id) ON DELETE RESTRICT,
    CHECK ((outcome = 'success' AND reset_event_id IS NOT NULL) OR (outcome <> 'success' AND reset_event_id IS NULL)),
    CHECK (created_at >= attempted_at - interval '5 minutes')
);
CREATE INDEX scheduled_usage_reset_attempts_term_time_idx
    ON pvnaive.scheduled_usage_reset_attempts(service_term_id, attempted_at DESC);
CREATE UNIQUE INDEX scheduled_usage_reset_attempts_one_success_uidx
    ON pvnaive.scheduled_usage_reset_attempts(service_term_id, scheduled_due_at)
    WHERE outcome = 'success';

CREATE TRIGGER scheduled_usage_reset_attempts_immutable
BEFORE UPDATE OR DELETE ON pvnaive.scheduled_usage_reset_attempts
FOR EACH ROW EXECUTE FUNCTION pvnaive.prevent_immutable_mutation();

-- Stable, non-login attribution identity for scheduled history. It is not used
-- to establish request context and has no password/session credentials.
INSERT INTO pvnaive.actors(id, tenant_id, actor_role, email, display_name, password_hash, mfa_required, status)
VALUES(
    '00000000-0000-0000-0000-000000000016', NULL, 'owner',
    'scheduler@pvnaive.invalid', 'PVNaive Scheduler', NULL, false, 'disabled'
)
ON CONFLICT (id) DO NOTHING;

CREATE FUNCTION pvnaive.init_service_term_reset_schedule()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog, pvnaive
AS $$
DECLARE
    v_strategy text := 'none';
    v_custom_days integer;
    v_anchor timestamptz;
BEGIN
    IF TG_OP = 'INSERT' THEN
        IF NEW.renewal_kind = 'renew_current' AND NEW.renewed_from_term_id IS NOT NULL THEN
            SELECT s.strategy, s.custom_days
              INTO v_strategy, v_custom_days
              FROM pvnaive.service_term_reset_schedules AS s
             WHERE s.service_term_id = NEW.renewed_from_term_id;
        END IF;

        IF NOT FOUND OR v_strategy IS NULL THEN
            v_strategy := 'none';
            v_custom_days := NULL;
        END IF;

        IF NOT (NEW.renewal_kind = 'renew_current' AND NEW.renewed_from_term_id IS NOT NULL AND FOUND)
           AND NEW.plan_id IS NOT NULL THEN
            SELECT p.reset_strategy, p.reset_custom_days
              INTO v_strategy, v_custom_days
              FROM pvnaive.plans AS p
             WHERE p.id = NEW.plan_id;
            IF NOT FOUND THEN
                v_strategy := 'none';
                v_custom_days := NULL;
            END IF;
        END IF;

        v_anchor := NEW.starts_at;
        INSERT INTO pvnaive.service_term_reset_schedules(
            service_term_id, tenant_id, user_id, strategy, custom_days,
            timezone_name, anchor_at, next_due_at
        ) VALUES (
            NEW.id, NEW.tenant_id, NEW.user_id, v_strategy, v_custom_days,
            'UTC', v_anchor,
            CASE WHEN v_anchor IS NULL OR v_strategy = 'none' THEN NULL
                 ELSE pvnaive.next_usage_reset_due(v_anchor, v_strategy, v_custom_days) END
        )
        ON CONFLICT (service_term_id) DO NOTHING;
        RETURN NEW;
    END IF;

    IF TG_OP = 'UPDATE' AND OLD.starts_at IS NULL AND NEW.starts_at IS NOT NULL THEN
        UPDATE pvnaive.service_term_reset_schedules
           SET anchor_at = NEW.starts_at,
               next_due_at = CASE WHEN strategy = 'none' THEN NULL
                                  ELSE pvnaive.next_usage_reset_due(NEW.starts_at, strategy, custom_days) END,
               updated_at = clock_timestamp()
         WHERE service_term_id = NEW.id
           AND anchor_at IS NULL;
    END IF;
    RETURN NEW;
END;
$$;
REVOKE ALL ON FUNCTION pvnaive.init_service_term_reset_schedule() FROM PUBLIC, pvnaive_app;

-- Existing ServiceTerms predate periodic enforcement. Snapshot the current Plan
-- policy once at adoption; later Plan edits never rewrite these schedule rows.
-- FORCE RLS is relaxed only under the migration transaction's AccessExclusive
-- locks and restored before commit, so no external transaction sees weaker RLS.
ALTER TABLE pvnaive.service_terms NO FORCE ROW LEVEL SECURITY;
ALTER TABLE pvnaive.plans NO FORCE ROW LEVEL SECURITY;
INSERT INTO pvnaive.service_term_reset_schedules(
    service_term_id, tenant_id, user_id, strategy, custom_days,
    timezone_name, anchor_at, next_due_at
)
SELECT st.id, st.tenant_id, st.user_id,
       COALESCE(p.reset_strategy, 'none'),
       CASE WHEN p.reset_strategy = 'custom' THEN p.reset_custom_days ELSE NULL END,
       'UTC', st.starts_at,
       CASE WHEN st.starts_at IS NULL OR COALESCE(p.reset_strategy, 'none') = 'none' THEN NULL
            ELSE pvnaive.next_usage_reset_due(
                st.starts_at,
                COALESCE(p.reset_strategy, 'none'),
                CASE WHEN p.reset_strategy = 'custom' THEN p.reset_custom_days ELSE NULL END
            ) END
  FROM pvnaive.service_terms AS st
  LEFT JOIN pvnaive.plans AS p ON p.id = st.plan_id
ON CONFLICT (service_term_id) DO NOTHING;
ALTER TABLE pvnaive.plans FORCE ROW LEVEL SECURITY;
ALTER TABLE pvnaive.service_terms FORCE ROW LEVEL SECURITY;

CREATE TRIGGER service_terms_init_reset_schedule
AFTER INSERT ON pvnaive.service_terms
FOR EACH ROW EXECUTE FUNCTION pvnaive.init_service_term_reset_schedule();
CREATE TRIGGER service_terms_start_reset_schedule
AFTER UPDATE OF starts_at ON pvnaive.service_terms
FOR EACH ROW
WHEN (OLD.starts_at IS NULL AND NEW.starts_at IS NOT NULL)
EXECUTE FUNCTION pvnaive.init_service_term_reset_schedule();

ALTER TABLE pvnaive.service_term_reset_schedules ENABLE ROW LEVEL SECURITY;
ALTER TABLE pvnaive.service_term_reset_schedules FORCE ROW LEVEL SECURITY;
CREATE POLICY service_term_reset_schedules_tenant_isolation ON pvnaive.service_term_reset_schedules
    FOR ALL TO pvnaive_app
    USING (pvnaive.has_tenant_access(tenant_id, false))
    WITH CHECK (pvnaive.has_tenant_access(tenant_id, false));
CREATE POLICY service_term_reset_schedules_internal_owner ON pvnaive.service_term_reset_schedules
    FOR ALL TO pvnaive_owner
    USING (pvnaive.has_tenant_access(tenant_id, false))
    WITH CHECK (pvnaive.has_tenant_access(tenant_id, false));

ALTER TABLE pvnaive.scheduled_usage_reset_attempts ENABLE ROW LEVEL SECURITY;
ALTER TABLE pvnaive.scheduled_usage_reset_attempts FORCE ROW LEVEL SECURITY;
CREATE POLICY scheduled_usage_reset_attempts_tenant_isolation ON pvnaive.scheduled_usage_reset_attempts
    FOR ALL TO pvnaive_app
    USING (pvnaive.has_tenant_access(tenant_id, false))
    WITH CHECK (pvnaive.has_tenant_access(tenant_id, false));
CREATE POLICY scheduled_usage_reset_attempts_internal_owner ON pvnaive.scheduled_usage_reset_attempts
    FOR ALL TO pvnaive_owner
    USING (pvnaive.has_tenant_access(tenant_id, false))
    WITH CHECK (pvnaive.has_tenant_access(tenant_id, false));

REVOKE ALL ON pvnaive.service_term_reset_schedules FROM PUBLIC, pvnaive_app;
REVOKE ALL ON pvnaive.scheduled_usage_reset_attempts FROM PUBLIC, pvnaive_app;
GRANT SELECT ON pvnaive.service_term_reset_schedules TO pvnaive_app;
GRANT SELECT ON pvnaive.scheduled_usage_reset_attempts TO pvnaive_app;

CREATE FUNCTION pvnaive.execute_due_scheduled_usage_resets(p_limit integer DEFAULT 50)
RETURNS TABLE(processed integer, succeeded integer, deferred integer, skipped integer)
LANGUAGE plpgsql
VOLATILE
SECURITY DEFINER
SET search_path = pg_catalog, pvnaive
AS $$
DECLARE
    v_now timestamptz := clock_timestamp();
    v_due record;
    v_last_reset timestamptz;
    v_resettable boolean;
    v_reset_reason text;
    v_previous_upload bigint;
    v_previous_download bigint;
    v_previous_used bigint;
    v_service_state text;
    v_mutation_key text;
    v_request_hash bytea;
    v_mutation_id uuid;
    v_existing_operation text;
    v_existing_hash bytea;
    v_existing_resource uuid;
    v_event_id uuid;
    v_event_reset_at timestamptz;
    v_processed integer := 0;
    v_succeeded integer := 0;
    v_deferred integer := 0;
    v_skipped integer := 0;
    v_system_actor constant uuid := '00000000-0000-0000-0000-000000000016'::uuid;
BEGIN
    IF p_limit IS NULL OR p_limit < 1 OR p_limit > 100 THEN
        RAISE EXCEPTION 'invalid scheduled reset batch limit' USING ERRCODE = '22023';
    END IF;

    -- Install the existing HMAC-signed synthetic Owner context only inside this
    -- narrow SECURITY DEFINER function. pvnaive_app never receives a generic
    -- context-escalation primitive.
    PERFORM pvnaive.direct_naive_accounting_enter_context();

    FOR v_due IN
        SELECT s.*, st.state AS term_state, st.expires_at AS term_expires
          FROM pvnaive.service_term_reset_schedules AS s
          JOIN pvnaive.service_terms AS st
            ON st.id = s.service_term_id
           AND st.tenant_id = s.tenant_id
           AND st.user_id = s.user_id
         WHERE s.next_due_at IS NOT NULL
           AND s.next_due_at <= v_now
           AND (s.retry_after_at IS NULL OR s.retry_after_at <= v_now)
         ORDER BY s.next_due_at, s.service_term_id
         FOR UPDATE OF s SKIP LOCKED
         LIMIT p_limit
    LOOP
        v_processed := v_processed + 1;

        IF v_due.term_state IN ('expired','ended','revoked')
           OR (v_due.term_expires IS NOT NULL AND v_due.term_expires <= v_now) THEN
            INSERT INTO pvnaive.scheduled_usage_reset_attempts(
                service_term_id,tenant_id,user_id,scheduled_due_at,attempted_at,outcome,reason
            ) VALUES (
                v_due.service_term_id,v_due.tenant_id,v_due.user_id,
                v_due.next_due_at,v_now,'skipped','service_inactive'
            );
            UPDATE pvnaive.service_term_reset_schedules
               SET next_due_at=NULL,retry_after_at=NULL,last_attempt_at=v_now,
                   last_error='service_inactive',updated_at=v_now
             WHERE service_term_id=v_due.service_term_id;
            v_skipped := v_skipped + 1;
            CONTINUE;
        END IF;

        SELECT t.last_reset_at
          INTO v_last_reset
          FROM pvnaive.direct_naive_accounting_terms AS t
         WHERE t.service_term_id = v_due.service_term_id;

        -- A manual/bulk reset that happened after this scheduled boundary
        -- already established a newer exact accounting epoch. Do not double
        -- reset; adopt that proven epoch as the cadence anchor.
        IF v_last_reset IS NOT NULL AND v_last_reset >= v_due.next_due_at THEN
            INSERT INTO pvnaive.scheduled_usage_reset_attempts(
                service_term_id,tenant_id,user_id,scheduled_due_at,attempted_at,outcome,reason
            ) VALUES (
                v_due.service_term_id,v_due.tenant_id,v_due.user_id,
                v_due.next_due_at,v_now,'skipped','explicit_reset_satisfied_period'
            );
            UPDATE pvnaive.service_term_reset_schedules
               SET anchor_at=v_last_reset,
                   next_due_at=pvnaive.next_usage_reset_due(v_last_reset,strategy,custom_days),
                   last_attempt_at=v_now,last_completed_at=v_last_reset,
                   retry_after_at=NULL,last_error=NULL,consecutive_failures=0,updated_at=v_now
             WHERE service_term_id=v_due.service_term_id;
            v_skipped := v_skipped + 1;
            CONTINUE;
        END IF;

        v_mutation_key := 'scheduled-reset:' || v_due.service_term_id::text || ':' ||
                          to_char(v_due.next_due_at AT TIME ZONE 'UTC','YYYYMMDDHH24MISSUS');
        v_request_hash := public.digest(convert_to(
            'customer.usage.reset.scheduled' || E'\n' || v_due.service_term_id::text || E'\n' ||
            to_char(v_due.next_due_at AT TIME ZONE 'UTC','YYYY-MM-DD"T"HH24:MI:SS.US'),
            'UTF8'
        ), 'sha256');

        INSERT INTO pvnaive.customer_mutation_keys(
            tenant_id,actor_id,idempotency_key,operation,request_hash,resource_type,resource_id,created_at
        ) VALUES (
            v_due.tenant_id,v_system_actor,v_mutation_key,'customer.usage.reset',
            v_request_hash,'user',v_due.user_id,v_now
        )
        ON CONFLICT (tenant_id,actor_id,idempotency_key) DO NOTHING;

        SELECT id,operation,request_hash,resource_id
          INTO v_mutation_id,v_existing_operation,v_existing_hash,v_existing_resource
          FROM pvnaive.customer_mutation_keys
         WHERE tenant_id=v_due.tenant_id
           AND actor_id=v_system_actor
           AND idempotency_key=v_mutation_key;
        IF NOT FOUND OR v_existing_operation <> 'customer.usage.reset'
           OR v_existing_hash IS DISTINCT FROM v_request_hash
           OR v_existing_resource IS DISTINCT FROM v_due.user_id THEN
            RAISE EXCEPTION 'scheduled reset idempotency conflict' USING ERRCODE = '23505';
        END IF;

        SELECT e.id,e.reset_at
          INTO v_event_id,v_event_reset_at
          FROM pvnaive.direct_naive_accounting_reset_events AS e
         WHERE e.customer_mutation_key_id=v_mutation_id;
        IF FOUND THEN
            UPDATE pvnaive.service_term_reset_schedules
               SET anchor_at=v_event_reset_at,
                   next_due_at=pvnaive.next_usage_reset_due(v_event_reset_at,strategy,custom_days),
                   last_attempt_at=v_now,last_completed_at=v_event_reset_at,
                   retry_after_at=NULL,last_error=NULL,consecutive_failures=0,updated_at=v_now
             WHERE service_term_id=v_due.service_term_id;
            v_succeeded := v_succeeded + 1;
            CONTINUE;
        END IF;

        SELECT r.resettable,r.reason,r.previous_upload_bytes,r.previous_download_bytes,
               r.previous_used_bytes,r.service_state
          INTO v_resettable,v_reset_reason,v_previous_upload,v_previous_download,
               v_previous_used,v_service_state
          FROM pvnaive.direct_naive_accounting_reset(v_due.service_term_id,v_now,90) AS r;

        IF v_resettable THEN
            INSERT INTO pvnaive.direct_naive_accounting_reset_events(
                tenant_id,user_id,service_term_id,actor_id,customer_mutation_key_id,reason,reset_at,
                previous_upload_bytes,previous_download_bytes,previous_used_bytes
            ) VALUES (
                v_due.tenant_id,v_due.user_id,v_due.service_term_id,
                v_system_actor,v_mutation_id,'scheduled',v_now,
                v_previous_upload,v_previous_download,v_previous_used
            ) RETURNING id INTO v_event_id;

            UPDATE pvnaive.customer_mutation_keys
               SET completed_at=v_now
             WHERE id=v_mutation_id AND completed_at IS NULL;

            INSERT INTO pvnaive.audit_events(
                tenant_id,actor_id,action,object_type,object_id,outcome,before_state,after_state
            ) VALUES (
                v_due.tenant_id,v_system_actor,'customer.usage.reset','service_term',
                v_due.service_term_id,'success',
                jsonb_build_object('upload_bytes',v_previous_upload,'download_bytes',v_previous_download,'used_bytes',v_previous_used),
                jsonb_build_object('upload_bytes',0,'download_bytes',0,'used_bytes',0,'period_started_at',v_now,'reason','scheduled')
            );

            INSERT INTO pvnaive.scheduled_usage_reset_attempts(
                service_term_id,tenant_id,user_id,scheduled_due_at,attempted_at,outcome,reason,reset_event_id
            ) VALUES (
                v_due.service_term_id,v_due.tenant_id,v_due.user_id,
                v_due.next_due_at,v_now,'success','reset',v_event_id
            );

            UPDATE pvnaive.service_term_reset_schedules
               SET anchor_at=v_now,
                   next_due_at=pvnaive.next_usage_reset_due(v_now,strategy,custom_days),
                   last_attempt_at=v_now,last_completed_at=v_now,
                   retry_after_at=NULL,last_error=NULL,consecutive_failures=0,updated_at=v_now
             WHERE service_term_id=v_due.service_term_id;
            v_succeeded := v_succeeded + 1;
        ELSIF v_reset_reason IN ('service_inactive','not_found') THEN
            INSERT INTO pvnaive.scheduled_usage_reset_attempts(
                service_term_id,tenant_id,user_id,scheduled_due_at,attempted_at,outcome,reason
            ) VALUES (
                v_due.service_term_id,v_due.tenant_id,v_due.user_id,
                v_due.next_due_at,v_now,'skipped',v_reset_reason
            );
            UPDATE pvnaive.service_term_reset_schedules
               SET next_due_at=NULL,retry_after_at=NULL,last_attempt_at=v_now,
                   last_error=v_reset_reason,updated_at=v_now
             WHERE service_term_id=v_due.service_term_id;
            v_skipped := v_skipped + 1;
        ELSE
            INSERT INTO pvnaive.scheduled_usage_reset_attempts(
                service_term_id,tenant_id,user_id,scheduled_due_at,attempted_at,outcome,reason
            ) VALUES (
                v_due.service_term_id,v_due.tenant_id,v_due.user_id,
                v_due.next_due_at,v_now,'deferred',COALESCE(v_reset_reason,'reset_unavailable')
            );
            UPDATE pvnaive.service_term_reset_schedules
               SET last_attempt_at=v_now,
                   retry_after_at=v_now + interval '60 seconds',
                   last_error=COALESCE(v_reset_reason,'reset_unavailable'),
                   consecutive_failures=consecutive_failures+1,updated_at=v_now
             WHERE service_term_id=v_due.service_term_id;
            v_deferred := v_deferred + 1;
        END IF;
    END LOOP;

    PERFORM pvnaive.direct_naive_accounting_leave_context();
    RETURN QUERY SELECT v_processed,v_succeeded,v_deferred,v_skipped;
EXCEPTION WHEN OTHERS THEN
    PERFORM pvnaive.direct_naive_accounting_leave_context();
    RAISE;
END;
$$;
REVOKE ALL ON FUNCTION pvnaive.execute_due_scheduled_usage_resets(integer) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION pvnaive.execute_due_scheduled_usage_resets(integer) TO pvnaive_app;

COMMENT ON TABLE pvnaive.service_term_reset_schedules IS
    'Frozen UTC periodic reset policy/cursor per ServiceTerm. Scheduler lateness advances cadence from actual successful reset time so no post-due bytes are discarded.';
COMMENT ON TABLE pvnaive.scheduled_usage_reset_attempts IS
    'Append-only periodic reset execution history; success links to the exact reset event and transient unsafe states are recorded as deferred.';
