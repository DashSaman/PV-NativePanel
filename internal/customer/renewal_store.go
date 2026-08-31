package customer

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

func (s *PostgresStore) CurrentRenewalContextTx(ctx context.Context, tx *sql.Tx, userID string) (RenewalContext, error) {
	if tx == nil {
		return RenewalContext{}, errors.New("customer: transaction is required")
	}
	var out RenewalContext
	var planID sql.NullString
	var renewedFrom sql.NullString
	var startPolicy, state, baselineState, baselineSource string
	var baselineUpload, baselineDownload sql.NullInt64
	err := tx.QueryRowContext(ctx, `
SELECT u.tenant_id::text, u.id::text,
       st.id::text, st.plan_id::text, st.quota_bytes, st.duration_seconds, st.no_expiry, st.concurrency_limit,
       st.start_policy, st.purchased_at, st.starts_at, st.first_connected_at, st.expires_at,
       st.state, st.renewal_kind, st.renewed_from_term_id::text,
       st.accounting_baseline_state, st.accounting_baseline_source, st.accounting_baseline_cutoff_at,
       st.accounting_baseline_upload_bytes, st.accounting_baseline_download_bytes, st.revision,
       urc.runtime_credential_id::text
FROM pvnaive.users u
JOIN LATERAL (
    SELECT candidate.*
    FROM pvnaive.service_terms candidate
    WHERE candidate.user_id=u.id AND candidate.tenant_id=u.tenant_id
    ORDER BY candidate.purchased_at DESC, candidate.created_at DESC
    LIMIT 1
) st ON TRUE
JOIN pvnaive.user_runtime_credentials urc
  ON urc.user_id=u.id AND urc.service_term_id=st.id AND urc.unbound_at IS NULL
WHERE u.id=$1::uuid
LIMIT 1`, userID).Scan(
		&out.TenantID, &out.UserID,
		&out.Current.ID, &planID, &out.Current.QuotaBytes, &out.Current.DurationSeconds, &out.Current.NoExpiry, &out.Current.ConcurrencyLimit,
		&startPolicy, &out.Current.PurchasedAt, &out.Current.StartsAt, &out.Current.FirstConnectedAt,
		&out.Current.ExpiresAt, &state, &out.Current.RenewalKind, &renewedFrom,
		&baselineState, &baselineSource, &out.Current.AccountingBaseline.CutoffAt, &baselineUpload, &baselineDownload,
		&out.Current.Revision, &out.RuntimeCredentialID,
	)
	if err != nil {
		return RenewalContext{}, fmt.Errorf("customer: resolve renewal context: %w", err)
	}
	out.Current.TenantID = out.TenantID
	out.Current.UserID = out.UserID
	out.Current.StartPolicy = StartPolicy(startPolicy)
	out.Current.State = TermState(state)
	out.Current.AccountingBaseline.State = AccountingBaselineState(baselineState)
	out.Current.AccountingBaseline.Source = AccountingBaselineSource(baselineSource)
	out.Current.AccountingBaseline.UploadBytes = nullableInt64Value(baselineUpload)
	out.Current.AccountingBaseline.DownloadBytes = nullableInt64Value(baselineDownload)
	if planID.Valid {
		out.Current.PlanID = planID.String
	}
	if renewedFrom.Valid {
		out.Current.RenewedFromTermID = renewedFrom.String
	}
	return out, nil
}

func (s *PostgresStore) PlanByIDTx(ctx context.Context, tx *sql.Tx, planID string) (PlanPreset, error) {
	if tx == nil {
		return PlanPreset{}, errors.New("customer: transaction is required")
	}
	var out PlanPreset
	var startPolicy, resetStrategy string
	var groupID sql.NullString
	err := tx.QueryRowContext(ctx, `
SELECT id::text, name, quota_bytes,
       CASE WHEN no_expiry THEN 0 ELSE duration_seconds END,
       no_expiry, start_policy, reset_strategy, COALESCE(reset_custom_days,0), concurrency_limit,
       default_group_id::text, enabled, sort_order
FROM pvnaive.plans
WHERE id=$1::uuid AND status <> 'archived'`, planID).Scan(
		&out.ID, &out.Name, &out.QuotaBytes, &out.ValiditySeconds, &out.NoExpiry,
		&startPolicy, &resetStrategy, &out.ResetCustomDays, &out.ConcurrencyLimit, &groupID, &out.Enabled, &out.SortOrder,
	)
	if err != nil {
		return PlanPreset{}, fmt.Errorf("customer: resolve plan: %w", err)
	}
	out.StartPolicy = StartPolicy(startPolicy)
	out.ResetStrategy = ResetStrategy(resetStrategy)
	out.ResetEnforcement = PeriodicResetEnforcementAvailable
	if groupID.Valid {
		out.DefaultGroupID = groupID.String
	}
	return out, nil
}

func (s *PostgresStore) CreateRenewalTermTx(ctx context.Context, tx *sql.Tx, record CreateRenewalTermRecord) (ServiceTerm, error) {
	if tx == nil {
		return ServiceTerm{}, errors.New("customer: transaction is required")
	}
	if err := validateAccountingBaseline(record.AccountingBaseline); err != nil {
		return ServiceTerm{}, err
	}
	var out ServiceTerm
	var planID, renewedFrom sql.NullString
	var startPolicy, state, baselineState, baselineSource string
	var baselineUpload, baselineDownload sql.NullInt64
	err := tx.QueryRowContext(ctx, `
INSERT INTO pvnaive.service_terms (
    tenant_id, user_id, plan_id, quota_bytes, duration_seconds, no_expiry, concurrency_limit,
    start_policy, purchased_at, starts_at, expires_at, state, renewal_kind, renewed_from_term_id,
    accounting_baseline_state, accounting_baseline_source, accounting_baseline_cutoff_at,
    accounting_baseline_upload_bytes, accounting_baseline_download_bytes
) VALUES (
    $1::uuid,$2::uuid,NULLIF($3,'')::uuid,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14::uuid,
    $15,$16,$17,$18,$19
)
RETURNING id::text, tenant_id::text, user_id::text, plan_id::text, quota_bytes,
          duration_seconds, no_expiry, concurrency_limit, start_policy, purchased_at, starts_at,
          first_connected_at, expires_at, state, renewal_kind, renewed_from_term_id::text,
          accounting_baseline_state, accounting_baseline_source, accounting_baseline_cutoff_at,
          accounting_baseline_upload_bytes, accounting_baseline_download_bytes, revision`,
		record.TenantID, record.UserID, record.PlanID, record.QuotaBytes, record.DurationSeconds,
		record.NoExpiry, record.ConcurrencyLimit, string(record.StartPolicy), record.PurchasedAt, record.StartsAt, record.ExpiresAt,
		string(record.State), record.RenewalKind, record.RenewedFromTermID,
		string(record.AccountingBaseline.State), string(record.AccountingBaseline.Source), record.AccountingBaseline.CutoffAt.UTC(),
		record.AccountingBaseline.UploadBytes, record.AccountingBaseline.DownloadBytes,
	).Scan(
		&out.ID, &out.TenantID, &out.UserID, &planID, &out.QuotaBytes,
		&out.DurationSeconds, &out.NoExpiry, &out.ConcurrencyLimit, &startPolicy, &out.PurchasedAt, &out.StartsAt,
		&out.FirstConnectedAt, &out.ExpiresAt, &state, &out.RenewalKind, &renewedFrom,
		&baselineState, &baselineSource, &out.AccountingBaseline.CutoffAt,
		&baselineUpload, &baselineDownload, &out.Revision,
	)
	if err != nil {
		return ServiceTerm{}, fmt.Errorf("customer: create renewal service term: %w", err)
	}
	if planID.Valid {
		out.PlanID = planID.String
	}
	if renewedFrom.Valid {
		out.RenewedFromTermID = renewedFrom.String
	}
	out.StartPolicy = StartPolicy(startPolicy)
	out.State = TermState(state)
	out.AccountingBaseline.State = AccountingBaselineState(baselineState)
	out.AccountingBaseline.Source = AccountingBaselineSource(baselineSource)
	out.AccountingBaseline.UploadBytes = nullableInt64Value(baselineUpload)
	out.AccountingBaseline.DownloadBytes = nullableInt64Value(baselineDownload)
	return out, nil
}

func (s *PostgresStore) RebindRuntimeCredentialTx(ctx context.Context, tx *sql.Tx, tenantID, userID, oldTermID, newTermID, runtimeID string) error {
	if tx == nil {
		return errors.New("customer: transaction is required")
	}
	result, err := tx.ExecContext(ctx, `
UPDATE pvnaive.user_runtime_credentials
SET unbound_at=clock_timestamp()
WHERE tenant_id=$1::uuid AND user_id=$2::uuid AND service_term_id=$3::uuid
  AND runtime_credential_id=$5::uuid AND unbound_at IS NULL`, tenantID, userID, oldTermID, newTermID, runtimeID)
	if err != nil {
		return fmt.Errorf("customer: unbind previous renewal runtime: %w", err)
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		return errors.New("customer: active runtime binding changed during renewal")
	}
	if _, err := tx.ExecContext(ctx, `
INSERT INTO pvnaive.user_runtime_credentials (
    tenant_id,user_id,service_term_id,runtime_credential_id,role
) VALUES ($1::uuid,$2::uuid,$3::uuid,$4::uuid,'primary')`, tenantID, userID, newTermID, runtimeID); err != nil {
		return fmt.Errorf("customer: bind renewal runtime: %w", err)
	}
	return nil
}

func (s *PostgresStore) ProjectSubscriptionToTermTx(ctx context.Context, tx *sql.Tx, userID string, term ServiceTerm) error {
	if tx == nil {
		return errors.New("customer: transaction is required")
	}
	_, err := tx.ExecContext(ctx, `
UPDATE pvnaive.direct_subscription_tokens dst
SET service_term_id=$2::uuid,
    service_state=$3,
    quota_bytes=$4,
    duration_seconds=$5,
    start_policy=$6,
    starts_at=$7,
    first_connected_at=$8,
    expires_at=$9
WHERE dst.user_id=$1::uuid
  AND dst.status='active'
  AND dst.revoked_at IS NULL`,
		userID, term.ID, string(term.State), term.QuotaBytes, term.DurationSeconds,
		string(term.StartPolicy), term.StartsAt, term.FirstConnectedAt, term.ExpiresAt,
	)
	if err != nil {
		return fmt.Errorf("customer: project renewal to subscription: %w", err)
	}
	return nil
}

func (s *PostgresStore) RecordRenewalProfileTx(ctx context.Context, tx *sql.Tx, tenantID, userID string, renewedAt time.Time, consumeNext bool) error {
	if tx == nil {
		return errors.New("customer: transaction is required")
	}
	_, err := tx.ExecContext(ctx, `
INSERT INTO pvnaive.customer_profiles (
    tenant_id,user_id,last_renewal_at,on_hold,next_plan_id,next_plan_source_term_id,next_plan_scheduled_at,updated_at
) VALUES (
    $1::uuid,$2::uuid,$3,false,NULL,NULL,NULL,clock_timestamp()
)
ON CONFLICT (user_id) DO UPDATE SET
    last_renewal_at=EXCLUDED.last_renewal_at,
    on_hold=false,
    next_plan_id=CASE WHEN $4 THEN NULL ELSE pvnaive.customer_profiles.next_plan_id END,
    next_plan_source_term_id=CASE WHEN $4 THEN NULL ELSE pvnaive.customer_profiles.next_plan_source_term_id END,
    next_plan_scheduled_at=CASE WHEN $4 THEN NULL ELSE pvnaive.customer_profiles.next_plan_scheduled_at END,
    revision=pvnaive.customer_profiles.revision+1,
    updated_at=clock_timestamp()`, tenantID, userID, renewedAt, consumeNext)
	if err != nil {
		return fmt.Errorf("customer: record renewal profile: %w", err)
	}
	return nil
}

func (s *PostgresStore) ScheduledNextPlanTx(ctx context.Context, tx *sql.Tx, userID string) (*ScheduledNextPlan, error) {
	if tx == nil {
		return nil, errors.New("customer: transaction is required")
	}
	var out ScheduledNextPlan
	var scheduled sql.NullTime
	err := tx.QueryRowContext(ctx, `
SELECT cp.next_plan_id::text, p.name, cp.next_plan_source_term_id::text, cp.next_plan_scheduled_at
FROM pvnaive.customer_profiles cp
JOIN pvnaive.plans p ON p.id=cp.next_plan_id
WHERE cp.user_id=$1::uuid AND cp.next_plan_id IS NOT NULL`, userID).
		Scan(&out.PlanID, &out.PlanName, &out.SourceTermID, &scheduled)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("customer: load next plan: %w", err)
	}
	if scheduled.Valid {
		value := scheduled.Time
		out.ScheduledAt = &value
	}
	return &out, nil
}

func (s *PostgresStore) ScheduleNextPlanTx(ctx context.Context, tx *sql.Tx, tenantID, userID, sourceTermID, planID string, scheduledAt time.Time) (ScheduledNextPlan, error) {
	if tx == nil {
		return ScheduledNextPlan{}, errors.New("customer: transaction is required")
	}
	if _, err := tx.ExecContext(ctx, `
INSERT INTO pvnaive.customer_profiles (
    tenant_id,user_id,next_plan_id,next_plan_source_term_id,next_plan_scheduled_at,updated_at
) VALUES ($1::uuid,$2::uuid,$3::uuid,$4::uuid,$5,clock_timestamp()
)
ON CONFLICT (user_id) DO UPDATE SET
    next_plan_id=EXCLUDED.next_plan_id,
    next_plan_source_term_id=EXCLUDED.next_plan_source_term_id,
    next_plan_scheduled_at=EXCLUDED.next_plan_scheduled_at,
    revision=pvnaive.customer_profiles.revision+1,
    updated_at=clock_timestamp()`, tenantID, userID, planID, sourceTermID, scheduledAt); err != nil {
		return ScheduledNextPlan{}, fmt.Errorf("customer: schedule next plan: %w", err)
	}
	plan, err := s.PlanByIDTx(ctx, tx, planID)
	if err != nil {
		return ScheduledNextPlan{}, err
	}
	value := scheduledAt
	return ScheduledNextPlan{PlanID: plan.ID, PlanName: plan.Name, SourceTermID: sourceTermID, ScheduledAt: &value}, nil
}

func (s *PostgresStore) ClearNextPlanTx(ctx context.Context, tx *sql.Tx, actorID, userID string) error {
	if tx == nil {
		return errors.New("customer: transaction is required")
	}
	result, err := tx.ExecContext(ctx, `
UPDATE pvnaive.customer_profiles
SET next_plan_id=NULL,next_plan_source_term_id=NULL,next_plan_scheduled_at=NULL,
    updated_by_actor_id=$2::uuid,revision=revision+1,updated_at=clock_timestamp()
WHERE user_id=$1::uuid`, userID, actorID)
	if err != nil {
		return fmt.Errorf("customer: clear next plan : %w", err)
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return ErrNextPlanMissing
	}
	return nil
}
