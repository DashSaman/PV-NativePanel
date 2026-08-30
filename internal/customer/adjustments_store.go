package customer

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

func scanAdjustedServiceTerm(row *sql.Row) (ServiceTerm, error) {
	var term ServiceTerm
	var startPolicy, state, baselineState, baselineSource string
	var baselineUpload, baselineDownload sql.NullInt64
	if err := row.Scan(
		&term.ID, &term.TenantID, &term.UserID, &term.QuotaBytes,
		&term.DurationSeconds, &startPolicy, &term.PurchasedAt, &term.StartsAt,
		&term.FirstConnectedAt, &term.ExpiresAt, &state,
		&baselineState, &baselineSource, &term.AccountingBaseline.CutoffAt,
		&baselineUpload, &baselineDownload, &term.Revision,
	); err != nil {
		return ServiceTerm{}, err
	}
	term.StartPolicy = StartPolicy(startPolicy)
	term.State = TermState(state)
	term.AccountingBaseline.State = AccountingBaselineState(baselineState)
	term.AccountingBaseline.Source = AccountingBaselineSource(baselineSource)
	term.AccountingBaseline.UploadBytes = nullableInt64Value(baselineUpload)
	term.AccountingBaseline.DownloadBytes = nullableInt64Value(baselineDownload)
	return term, nil
}

func (s *PostgresStore) AddCurrentServiceQuotaTx(ctx context.Context, tx *sql.Tx, userID string, deltaBytes int64) (ServiceTerm, error) {
	if tx == nil {
		return ServiceTerm{}, errors.New("customer: transaction is required")
	}
	term, err := scanAdjustedServiceTerm(tx.QueryRowContext(ctx, `
UPDATE pvnaive.service_terms st
SET quota_bytes = st.quota_bytes + $2,
    revision = st.revision + 1,
    updated_at = clock_timestamp()
WHERE st.id = (
    SELECT current.id
    FROM pvnaive.service_terms current
    WHERE current.user_id = $1::uuid
    ORDER BY current.purchased_at DESC, current.created_at DESC
    LIMIT 1
)
  AND st.quota_bytes IS NOT NULL
RETURNING st.id::text, st.tenant_id::text, st.user_id::text, st.quota_bytes,
          st.duration_seconds, st.start_policy, st.purchased_at, st.starts_at,
          st.first_connected_at, st.expires_at, st.state,
          st.accounting_baseline_state, st.accounting_baseline_source, st.accounting_baseline_cutoff_at,
          st.accounting_baseline_upload_bytes, st.accounting_baseline_download_bytes,
          st.revision`, userID, deltaBytes))
	if errors.Is(err, sql.ErrNoRows) {
		var unlimited bool
		checkErr := tx.QueryRowContext(ctx, `
SELECT quota_bytes IS NULL
FROM pvnaive.service_terms
WHERE user_id = $1::uuid
ORDER BY purchased_at DESC, created_at DESC
LIMIT 1`, userID).Scan(&unlimited)
		if checkErr == nil && unlimited {
			return ServiceTerm{}, ErrUnlimitedQuotaAddition
		}
		return ServiceTerm{}, ErrCustomerServiceNotFound
	}
	if err != nil {
		return ServiceTerm{}, fmt.Errorf("customer: add service quota: %w", err)
	}
	return term, nil
}

func (s *PostgresStore) ExtendCurrentServiceTx(ctx context.Context, tx *sql.Tx, userID string, seconds int64, now time.Time) (ServiceTerm, error) {
	if tx == nil {
		return ServiceTerm{}, errors.New("customer: transaction is required")
	}
	term, err := scanAdjustedServiceTerm(tx.QueryRowContext(ctx, `
UPDATE pvnaive.service_terms st
SET duration_seconds = st.duration_seconds + $2,
    expires_at = CASE
        WHEN st.start_policy = 'on_first_successful_connection' AND st.first_connected_at IS NULL THEN NULL
        ELSE GREATEST(COALESCE(st.expires_at, $3), $3) + make_interval(secs => $2::double precision)
    END,
    state = CASE
        WHEN st.state IN ('ended', 'revoked') THEN st.state
        WHEN st.start_policy = 'on_first_successful_connection' AND st.first_connected_at IS NULL THEN 'pending'::pvnaive.service_term_state
        ELSE 'active'::pvnaive.service_term_state
    END,
    revision = st.revision + 1,
    updated_at = clock_timestamp()
WHERE st.id = (
    SELECT current.id
    FROM pvnaive.service_terms current
    WHERE current.user_id = $1::uuid
    ORDER BY current.purchased_at DESC, current.created_at DESC
    LIMIT 1
)
  AND st.state NOT IN ('ended', 'revoked')
RETURNING st.id::text, st.tenant_id::text, st.user_id::text, st.quota_bytes,
          st.duration_seconds, st.start_policy, st.purchased_at, st.starts_at,
          st.first_connected_at, st.expires_at, st.state,
          st.accounting_baseline_state, st.accounting_baseline_source, st.accounting_baseline_cutoff_at,
          st.accounting_baseline_upload_bytes, st.accounting_baseline_download_bytes,
          st.revision`, userID, seconds, now))
	if errors.Is(err, sql.ErrNoRows) {
		return ServiceTerm{}, ErrCustomerServiceNotFound
	}
	if err != nil {
		return ServiceTerm{}, fmt.Errorf("customer: extend service validity: %w", err)
	}
	return term, nil
}
