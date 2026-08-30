-- pvnaive:migration-version 0012
-- Source: PVNaive Task #5 legacy/adopted accounting baseline truth
-- pvnaive:migration-name accounting_baseline_truth
-- pvnaive:transactional true
-- pvnaive:destructive false

SET LOCAL lock_timeout = '10s';
SET LOCAL statement_timeout = '120s';
SET LOCAL idle_in_transaction_session_timeout = '60s';
SET LOCAL ROLE pvnaive_owner;

ALTER TABLE pvnaive.service_terms
    ADD COLUMN accounting_baseline_state text,
    ADD COLUMN accounting_baseline_source text,
    ADD COLUMN accounting_baseline_cutoff_at timestamptz,
    ADD COLUMN accounting_baseline_upload_bytes bigint,
    ADD COLUMN accounting_baseline_download_bytes bigint;

-- service_terms uses FORCE ROW LEVEL SECURITY and its tenant policy is granted
-- to pvnaive_app rather than the owning migration role. The migration already
-- owns an ACCESS EXCLUSIVE table lock for ALTER TABLE, so disable RLS only for
-- this transactional owner-only backfill and restore the exact protection
-- before any constraint/trigger is committed.
ALTER TABLE pvnaive.service_terms DISABLE ROW LEVEL SECURITY;

-- Every pre-v12 ServiceTerm predates explicit baseline provenance. Do not infer
-- historical bytes from empty legacy ledgers or from the direct accounting
-- counters. The direct epoch starts at the first durable Runtime binding when
-- one exists; otherwise purchased_at is the only durable boundary available.
UPDATE pvnaive.service_terms AS st
SET accounting_baseline_state = 'unknown',
    accounting_baseline_source = 'legacy_unavailable',
    accounting_baseline_cutoff_at = COALESCE(
        (
            SELECT MIN(binding.bound_at)
            FROM pvnaive.user_runtime_credentials AS binding
            WHERE binding.service_term_id = st.id
        ),
        st.purchased_at
    ),
    accounting_baseline_upload_bytes = NULL,
    accounting_baseline_download_bytes = NULL;

ALTER TABLE pvnaive.service_terms ENABLE ROW LEVEL SECURITY;
ALTER TABLE pvnaive.service_terms FORCE ROW LEVEL SECURITY;

ALTER TABLE pvnaive.service_terms
    ALTER COLUMN accounting_baseline_state SET NOT NULL,
    ALTER COLUMN accounting_baseline_source SET NOT NULL,
    ALTER COLUMN accounting_baseline_cutoff_at SET NOT NULL,
    ADD CONSTRAINT service_terms_accounting_baseline_truth_check
    CHECK (
        (
            accounting_baseline_state = 'unknown'
            AND accounting_baseline_source = 'legacy_unavailable'
            AND accounting_baseline_upload_bytes IS NULL
            AND accounting_baseline_download_bytes IS NULL
        )
        OR
        (
            accounting_baseline_state = 'known'
            AND accounting_baseline_source = 'fresh_managed_term'
            AND accounting_baseline_upload_bytes = 0
            AND accounting_baseline_download_bytes = 0
        )
        OR
        (
            accounting_baseline_state = 'known'
            AND accounting_baseline_source = 'authoritative_import'
            AND accounting_baseline_upload_bytes IS NOT NULL
            AND accounting_baseline_download_bytes IS NOT NULL
            AND accounting_baseline_upload_bytes >= 0
            AND accounting_baseline_download_bytes >= 0
        )
    );

CREATE FUNCTION pvnaive.prevent_accounting_baseline_mutation()
RETURNS trigger
LANGUAGE plpgsql
SET search_path = pg_catalog, pvnaive
AS $$
BEGIN
    IF ROW(
        OLD.accounting_baseline_state,
        OLD.accounting_baseline_source,
        OLD.accounting_baseline_cutoff_at,
        OLD.accounting_baseline_upload_bytes,
        OLD.accounting_baseline_download_bytes
    ) IS DISTINCT FROM ROW(
        NEW.accounting_baseline_state,
        NEW.accounting_baseline_source,
        NEW.accounting_baseline_cutoff_at,
        NEW.accounting_baseline_upload_bytes,
        NEW.accounting_baseline_download_bytes
    ) THEN
        RAISE EXCEPTION 'accounting baseline provenance is immutable';
    END IF;
    RETURN NEW;
END;
$$;

CREATE TRIGGER service_terms_accounting_baseline_immutable
BEFORE UPDATE OF
    accounting_baseline_state,
    accounting_baseline_source,
    accounting_baseline_cutoff_at,
    accounting_baseline_upload_bytes,
    accounting_baseline_download_bytes
ON pvnaive.service_terms
FOR EACH ROW EXECUTE FUNCTION pvnaive.prevent_accounting_baseline_mutation();

REVOKE ALL ON FUNCTION pvnaive.prevent_accounting_baseline_mutation() FROM PUBLIC;
REVOKE ALL ON FUNCTION pvnaive.prevent_accounting_baseline_mutation() FROM pvnaive_app;

COMMENT ON COLUMN pvnaive.service_terms.accounting_baseline_state IS
    'Historical usage before accounting_baseline_cutoff_at is numeric only when known; legacy/adopted history without authoritative evidence is unknown, never zero.';
COMMENT ON COLUMN pvnaive.service_terms.accounting_baseline_source IS
    'Provenance for the historical baseline: fresh_managed_term, legacy_unavailable, or authoritative_import.';
COMMENT ON COLUMN pvnaive.service_terms.accounting_baseline_cutoff_at IS
    'Non-overlap boundary: direct Naive accounting belongs to this ServiceTerm epoch at/after this instant; pre-boundary history is represented only by the baseline fields.';
