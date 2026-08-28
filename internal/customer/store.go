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
	var term ServiceTerm
	var startPolicy string
	var state string
	if err := tx.QueryRowContext(ctx, `
INSERT INTO pvnaive.service_terms (
    tenant_id, user_id, quota_bytes, duration_seconds, start_policy,
    purchased_at, starts_at, expires_at, state
) VALUES (
    $1::uuid, $2::uuid, $3, $4, $5, $6, $7, $8, $9
)
RETURNING id::text, tenant_id::text, user_id::text, quota_bytes,
          duration_seconds, start_policy, purchased_at, starts_at,
          first_connected_at, expires_at, state, revision`,
		record.TenantID, record.UserID, record.QuotaBytes, record.DurationSeconds,
		string(record.StartPolicy), record.PurchasedAt, record.StartsAt, record.ExpiresAt,
		string(record.State),
	).Scan(
		&term.ID, &term.TenantID, &term.UserID, &term.QuotaBytes,
		&term.DurationSeconds, &startPolicy, &term.PurchasedAt, &term.StartsAt,
		&term.FirstConnectedAt, &term.ExpiresAt, &state, &term.Revision,
	); err != nil {
		return ServiceTerm{}, fmt.Errorf("customer: insert service term: %w", err)
	}
	term.StartPolicy = StartPolicy(startPolicy)
	term.State = TermState(state)
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
