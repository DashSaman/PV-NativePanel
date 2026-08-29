package customer

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/DashSaman/PV-NaivePanel/internal/runtimecred"
)

func (s *PostgresStore) AdoptableRuntimeCredentialTx(ctx context.Context, tx *sql.Tx, runtimeCredentialID string) (runtimecred.CredentialView, error) {
	if tx == nil {
		return runtimecred.CredentialView{}, errors.New("customer: transaction is required")
	}
	var view runtimecred.CredentialView
	var status, origin string
	var rotatedAt, revokedAt sql.NullTime
	err := tx.QueryRowContext(ctx, `
SELECT
    rc.id::text,
    rc.username,
    rc.status,
    rc.origin,
    rc.revision,
    rc.created_at,
    rc.updated_at,
    rc.rotated_at,
    rc.revoked_at
FROM pvnaive.naive_runtime_credentials AS rc
WHERE rc.id = $1::uuid
  AND rc.status = 'active'
  AND NOT EXISTS (
      SELECT 1
      FROM pvnaive.user_runtime_credentials AS urc
      WHERE urc.runtime_credential_id = rc.id
        AND urc.unbound_at IS NULL
  )
FOR UPDATE`, runtimeCredentialID).Scan(
		&view.ID,
		&view.Username,
		&status,
		&origin,
		&view.Revision,
		&view.CreatedAt,
		&view.UpdatedAt,
		&rotatedAt,
		&revokedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return runtimecred.CredentialView{}, ErrRuntimeCredentialNotAdoptable
		}
		return runtimecred.CredentialView{}, fmt.Errorf("customer: resolve adoptable runtime credential: %w", err)
	}
	view.Status = runtimecred.CredentialStatus(status)
	view.Origin = runtimecred.CredentialOrigin(origin)
	if rotatedAt.Valid {
		value := rotatedAt.Time
		view.RotatedAt = &value
	}
	if revokedAt.Valid {
		value := revokedAt.Time
		view.RevokedAt = &value
	}
	return view, nil
}

func (s *PostgresStore) UpdateCurrentServiceTermTx(ctx context.Context, tx *sql.Tx, userID string, record UpdateServiceTermRecord) (ServiceTerm, error) {
	if tx == nil {
		return ServiceTerm{}, errors.New("customer: transaction is required")
	}
	var term ServiceTerm
	var startPolicy, state string
	err := tx.QueryRowContext(ctx, `
UPDATE pvnaive.service_terms AS st
SET quota_bytes = $2,
    duration_seconds = $3,
    start_policy = $4,
    starts_at = $5,
    first_connected_at = CASE WHEN $4 = 'on_first_successful_connection' THEN NULL ELSE first_connected_at END,
    expires_at = $6,
    state = $7,
    updated_at = $8,
    revision = revision + 1
WHERE st.id = (
    SELECT urc.service_term_id
    FROM pvnaive.user_runtime_credentials AS urc
    WHERE urc.user_id = $1::uuid
      AND urc.unbound_at IS NULL
    ORDER BY urc.bound_at DESC
    LIMIT 1
)
RETURNING st.id::text, st.tenant_id::text, st.user_id::text, st.quota_bytes,
          st.duration_seconds, st.start_policy, st.purchased_at, st.starts_at,
          st.first_connected_at, st.expires_at, st.state, st.revision`,
		userID,
		record.QuotaBytes,
		record.DurationSeconds,
		string(record.StartPolicy),
		record.StartsAt,
		record.ExpiresAt,
		string(record.State),
		record.EffectiveAt,
	).Scan(
		&term.ID,
		&term.TenantID,
		&term.UserID,
		&term.QuotaBytes,
		&term.DurationSeconds,
		&startPolicy,
		&term.PurchasedAt,
		&term.StartsAt,
		&term.FirstConnectedAt,
		&term.ExpiresAt,
		&state,
		&term.Revision,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ServiceTerm{}, ErrCustomerServiceNotFound
		}
		return ServiceTerm{}, fmt.Errorf("customer: update service term: %w", err)
	}
	term.StartPolicy = StartPolicy(startPolicy)
	term.State = TermState(state)
	return term, nil
}
