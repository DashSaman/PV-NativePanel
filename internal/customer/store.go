package customer

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type PostgresStore struct{}

func NewPostgresStore() *PostgresStore { return &PostgresStore{} }

func (s *PostgresStore) DirectTenantID(ctx context.Context, tx *sql.Tx) (string, error) {
	if tx == nil {
		return "", errors.New("customer: transaction is required")
	}
	var tenantID string
	if err := tx.QueryRowContext(ctx, `
SELECT id::text
FROM pvnaive.tenants
WHERE tenant_type = 'system'
  AND slug = 'direct'
  AND status = 'active'
LIMIT 1`).Scan(&tenantID); err != nil {
		return "", fmt.Errorf("customer: query direct tenant: %w", err)
	}
	return tenantID, nil
}

func (s *PostgresStore) CreateUserTx(ctx context.Context, tx *sql.Tx, record CreateUserRecord) (User, error) {
	if tx == nil {
		return User{}, errors.New("customer: transaction is required")
	}
	var user User
	var state string
	if err := tx.QueryRowContext(ctx, `
INSERT INTO pvnaive.users (
    tenant_id, username, display_name, status, created_by_actor_id
) VALUES ($1::uuid, $2, $3, 'active', $4::uuid)
RETURNING id::text, tenant_id::text, username, display_name, status,
          revision, created_at, updated_at`,
		record.TenantID, record.Username, record.DisplayName, record.ActorID,
	).Scan(
		&user.ID, &user.TenantID, &user.Username, &user.DisplayName, &state,
		&user.Revision, &user.CreatedAt, &user.UpdatedAt,
	); err != nil {
		return User{}, fmt.Errorf("customer: insert user: %w", err)
	}
	user.Status = UserAdminState(state)
	return user, nil
}

func (s *PostgresStore) CreateServiceTermTx(ctx context.Context, tx *sql.Tx, record CreateServiceTermRecord) (ServiceTerm, error) {
	if tx == nil {
		return ServiceTerm{}, errors.New("customer: transaction is required")
	}
	if err := validateAccountingBaseline(record.AccountingBaseline); err != nil {
		return ServiceTerm{}, err
	}
	var term ServiceTerm
	var startPolicy, state, baselineState, baselineSource string
	var baselineUpload, baselineDownload sql.NullInt64
	if err := tx.QueryRowContext(ctx, `
INSERT INTO pvnaive.service_terms (
    tenant_id, user_id, quota_bytes, duration_seconds, start_policy,
    purchased_at, starts_at, expires_at, state,
    accounting_baseline_state, accounting_baseline_source, accounting_baseline_cutoff_at,
    accounting_baseline_upload_bytes, accounting_baseline_download_bytes
) VALUES (
    $1::uuid, $2::uuid, $3, $4, $5, $6, $7, $8, $9,
    $10, $11, $12, $13, $14
)
RETURNING id::text, tenant_id::text, user_id::text, quota_bytes,
          duration_seconds, start_policy, purchased_at, starts_at,
          first_connected_at, expires_at, state,
          accounting_baseline_state, accounting_baseline_source, accounting_baseline_cutoff_at,
          accounting_baseline_upload_bytes, accounting_baseline_download_bytes,
          revision`,
		record.TenantID, record.UserID, record.QuotaBytes, record.DurationSeconds,
		string(record.StartPolicy), record.PurchasedAt, record.StartsAt, record.ExpiresAt,
		string(record.State), string(record.AccountingBaseline.State), string(record.AccountingBaseline.Source),
		record.AccountingBaseline.CutoffAt.UTC(), record.AccountingBaseline.UploadBytes, record.AccountingBaseline.DownloadBytes,
	).Scan(
		&term.ID, &term.TenantID, &term.UserID, &term.QuotaBytes,
		&term.DurationSeconds, &startPolicy, &term.PurchasedAt, &term.StartsAt,
		&term.FirstConnectedAt, &term.ExpiresAt, &state,
		&baselineState, &baselineSource, &term.AccountingBaseline.CutoffAt,
		&baselineUpload, &baselineDownload, &term.Revision,
	); err != nil {
		return ServiceTerm{}, fmt.Errorf("customer: insert service term: %w", err)
	}
	term.StartPolicy = StartPolicy(startPolicy)
	term.State = TermState(state)
	term.AccountingBaseline.State = AccountingBaselineState(baselineState)
	term.AccountingBaseline.Source = AccountingBaselineSource(baselineSource)
	term.AccountingBaseline.UploadBytes = nullableInt64Value(baselineUpload)
	term.AccountingBaseline.DownloadBytes = nullableInt64Value(baselineDownload)
	return term, nil
}

func (s *PostgresStore) BindRuntimeCredentialTx(ctx context.Context, tx *sql.Tx, tenantID, userID, termID, runtimeCredentialID string) error {
	if tx == nil {
		return errors.New("customer: transaction is required")
	}
	if _, err := tx.ExecContext(ctx, `
INSERT INTO pvnaive.user_runtime_credentials (
    tenant_id, user_id, service_term_id, runtime_credential_id, role
) VALUES ($1::uuid, $2::uuid, $3::uuid, $4::uuid, 'primary')`,
		tenantID, userID, termID, runtimeCredentialID,
	); err != nil {
		return fmt.Errorf("customer: bind runtime credential: %w", err)
	}
	return nil
}

func (s *PostgresStore) ListCustomersTx(ctx context.Context, tx *sql.Tx) ([]CustomerView, error) {
	if tx == nil {
		return nil, errors.New("customer: transaction is required")
	}
	rows, err := tx.QueryContext(ctx, `
SELECT
    u.id::text,
    u.username,
    u.status,
    st.id::text,
    st.state,
    st.quota_bytes,
    st.duration_seconds,
    st.start_policy,
    st.starts_at,
    st.first_connected_at,
    st.expires_at,
    st.accounting_baseline_state,
    st.accounting_baseline_source,
    st.accounting_baseline_cutoff_at,
    st.accounting_baseline_upload_bytes,
    st.accounting_baseline_download_bytes,
    urc.runtime_credential_id::text,
    EXISTS (
        SELECT 1
        FROM pvnaive.direct_subscription_tokens dst
        WHERE dst.user_id = u.id
          AND dst.service_term_id = st.id
          AND dst.runtime_credential_id = urc.runtime_credential_id
          AND dst.status = 'active'
    ) AS subscription_available
FROM pvnaive.users u
JOIN LATERAL (
    SELECT candidate.*
    FROM pvnaive.service_terms candidate
    WHERE candidate.user_id = u.id
      AND candidate.tenant_id = u.tenant_id
    ORDER BY candidate.purchased_at DESC, candidate.created_at DESC
    LIMIT 1
) st ON TRUE
JOIN pvnaive.user_runtime_credentials urc
  ON urc.user_id = u.id
 AND urc.service_term_id = st.id
 AND urc.unbound_at IS NULL
WHERE u.tenant_id = st.tenant_id
ORDER BY u.created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("customer: list customers: %w", err)
	}
	defer rows.Close()

	out := make([]CustomerView, 0)
	for rows.Next() {
		var view CustomerView
		var userState, termState, startPolicy, baselineState, baselineSource string
		var baselineUpload, baselineDownload sql.NullInt64
		if err := rows.Scan(
			&view.UserID,
			&view.Username,
			&userState,
			&view.ServiceTermID,
			&termState,
			&view.QuotaBytes,
			&view.DurationSeconds,
			&startPolicy,
			&view.StartsAt,
			&view.FirstConnectedAt,
			&view.ExpiresAt,
			&baselineState,
			&baselineSource,
			&view.AccountingBaseline.CutoffAt,
			&baselineUpload,
			&baselineDownload,
			&view.RuntimeCredentialID,
			&view.SubscriptionAvailable,
		); err != nil {
			return nil, fmt.Errorf("customer: scan customer: %w", err)
		}
		view.Status = UserAdminState(userState)
		view.ServiceState = TermState(termState)
		view.StartPolicy = StartPolicy(startPolicy)
		view.AccountingBaseline.State = AccountingBaselineState(baselineState)
		view.AccountingBaseline.Source = AccountingBaselineSource(baselineSource)
		view.AccountingBaseline.UploadBytes = nullableInt64Value(baselineUpload)
		view.AccountingBaseline.DownloadBytes = nullableInt64Value(baselineDownload)
		view.UsageCapability = DefaultUsageCapability()
		out = append(out, view)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("customer: list customers rows: %w", err)
	}
	return out, nil
}

func (s *PostgresStore) SubscriptionTargetTx(ctx context.Context, tx *sql.Tx, userID string) (SubscriptionTarget, error) {
	if tx == nil {
		return SubscriptionTarget{}, errors.New("customer: transaction is required")
	}
	var target SubscriptionTarget
	var expiresAt sql.NullTime
	err := tx.QueryRowContext(ctx, `
SELECT u.tenant_id::text, u.id::text, st.id::text, urc.runtime_credential_id::text, st.expires_at
FROM pvnaive.users u
JOIN LATERAL (
    SELECT candidate.*
    FROM pvnaive.service_terms candidate
    WHERE candidate.user_id = u.id
      AND candidate.tenant_id = u.tenant_id
    ORDER BY candidate.purchased_at DESC, candidate.created_at DESC
    LIMIT 1
) st ON TRUE
JOIN pvnaive.user_runtime_credentials urc
  ON urc.user_id = u.id
 AND urc.service_term_id = st.id
 AND urc.unbound_at IS NULL
JOIN pvnaive.naive_runtime_credentials rc
  ON rc.id = urc.runtime_credential_id
 AND rc.status = 'active'
WHERE u.id = $1::uuid
  AND u.status = 'active'
  AND st.state IN ('pending','active')`, userID).Scan(
		&target.TenantID, &target.UserID, &target.ServiceTermID,
		&target.RuntimeCredentialID, &expiresAt,
	)
	if err != nil {
		return SubscriptionTarget{}, fmt.Errorf("customer: resolve subscription target: %w", err)
	}
	if expiresAt.Valid {
		expires := expiresAt.Time
		target.ExpiresAt = &expires
	}
	return target, nil
}

func (s *PostgresStore) RevokeSubscriptionTokensTx(ctx context.Context, tx *sql.Tx, target SubscriptionTarget) error {
	if tx == nil {
		return errors.New("customer: transaction is required")
	}
	_, err := tx.ExecContext(ctx, `
UPDATE pvnaive.direct_subscription_tokens
SET status='revoked', revoked_at=clock_timestamp()
WHERE user_id=$1::uuid
  AND service_term_id=$2::uuid
  AND status='active'`, target.UserID, target.ServiceTermID)
	if err != nil {
		return fmt.Errorf("customer: revoke subscription tokens: %w", err)
	}
	return nil
}

func (s *PostgresStore) CreateSubscriptionTokenTx(ctx context.Context, tx *sql.Tx, record CreateSubscriptionTokenRecord) error {
	if tx == nil {
		return errors.New("customer: transaction is required")
	}
	if len(record.TokenHash) != 32 {
		return errors.New("customer: subscription token hash must be 32 bytes")
	}
	if len(record.TokenPrefix) < 6 || len(record.TokenPrefix) > 16 {
		return errors.New("customer: subscription token prefix must be 6-16 bytes")
	}
	result, err := tx.ExecContext(ctx, `
INSERT INTO pvnaive.direct_subscription_tokens (
    tenant_id, user_id, service_term_id, runtime_credential_id,
    token_hash, token_prefix, status, user_state, service_state,
    runtime_username, secret_ciphertext, secret_nonce, encryption_key_id, expires_at,
    quota_bytes, duration_seconds, start_policy, starts_at, first_connected_at,
    accounting_baseline_state, accounting_baseline_source, accounting_baseline_cutoff_at,
    accounting_baseline_upload_bytes, accounting_baseline_download_bytes
)
SELECT
    u.tenant_id, u.id, st.id, rc.id,
    $5, $6, 'active', u.status, st.state,
    rc.username, rc.secret_ciphertext, rc.secret_nonce, rc.encryption_key_id, $7,
    st.quota_bytes, st.duration_seconds, st.start_policy, st.starts_at, st.first_connected_at,
    st.accounting_baseline_state, st.accounting_baseline_source, st.accounting_baseline_cutoff_at,
    st.accounting_baseline_upload_bytes, st.accounting_baseline_download_bytes
FROM pvnaive.users AS u
JOIN pvnaive.service_terms AS st
  ON st.id = $3::uuid
 AND st.tenant_id = u.tenant_id
 AND st.user_id = u.id
JOIN pvnaive.user_runtime_credentials AS urc
  ON urc.service_term_id = st.id
 AND urc.runtime_credential_id = $4::uuid
 AND urc.unbound_at IS NULL
JOIN pvnaive.naive_runtime_credentials AS rc
  ON rc.id = urc.runtime_credential_id
WHERE u.id = $2::uuid
  AND u.tenant_id = $1::uuid
  AND rc.status = 'active'`,
		record.TenantID, record.UserID, record.ServiceTermID, record.RuntimeCredentialID,
		record.TokenHash, record.TokenPrefix, record.ExpiresAt,
	)
	if err != nil {
		return fmt.Errorf("customer: insert subscription token: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("customer: inspect subscription token insert: %w", err)
	}
	if rows != 1 {
		return errors.New("customer: subscription token projection scope not found")
	}
	return nil
}

func nullableInt64Value(value sql.NullInt64) *int64 {
	if !value.Valid {
		return nil
	}
	copy := value.Int64
	return &copy
}
