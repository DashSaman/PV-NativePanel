package customer

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type productCreateStore interface {
	OperationTenantIDTx(context.Context, *sql.Tx) (string, error)
	PlanByIDTx(context.Context, *sql.Tx, string) (PlanPreset, error)
	CreateProductServiceTermTx(context.Context, *sql.Tx, CreateServiceTermRecord) (ServiceTerm, error)
	ApplyCustomerMetadataTx(context.Context, *sql.Tx, string, string, string, string, []string, bool) error
}

func (s *PostgresStore) CreateProductServiceTermTx(ctx context.Context, tx *sql.Tx, record CreateServiceTermRecord) (ServiceTerm, error) {
	if tx == nil {
		return ServiceTerm{}, errors.New("customer: transaction is required")
	}
	var term ServiceTerm
	var planID sql.NullString
	var startPolicy string
	var state string
	var renewedFrom sql.NullString
	if err := tx.QueryRowContext(ctx, `
INSERT INTO pvnaive.service_terms (
    tenant_id, user_id, plan_id, quota_bytes, duration_seconds, no_expiry,
    start_policy, purchased_at, starts_at, expires_at, state, renewal_kind, renewed_from_term_id
) VALUES (
    $1::uuid, $2::uuid, NULLIF($3,'')::uuid, $4, $5, $6,
    $7, $8, $9, $10, $11, COALESCE(NULLIF($12,''),'initial'), NULLIF($13,'')::uuid
)
RETURNING id::text, tenant_id::text, user_id::text, plan_id::text, quota_bytes,
          duration_seconds, no_expiry, start_policy, purchased_at, starts_at,
          first_connected_at, expires_at, state, renewal_kind, renewed_from_term_id::text, revision`,
		record.TenantID,
		record.UserID,
		record.PlanID,
		record.QuotaBytes,
		record.DurationSeconds,
		record.NoExpiry,
		string(record.StartPolicy),
		record.PurchasedAt,
		record.StartsAt,
		record.ExpiresAt,
		string(record.State),
		record.RenewalKind,
		record.RenewedFromTermID,
	).Scan(
		&term.ID,
		&term.TenantID,
		&term.UserID,
		&planID,
		&term.QuotaBytes,
		&term.DurationSeconds,
		&term.NoExpiry,
		&startPolicy,
		&term.PurchasedAt,
		&term.StartsAt,
		&term.FirstConnectedAt,
		&term.ExpiresAt,
		&state,
		&term.RenewalKind,
		&renewedFrom,
		&term.Revision,
	); err != nil {
		return ServiceTerm{}, fmt.Errorf("customer: insert product service term: %w", err)
	}
	if planID.Valid {
		term.PlanID = planID.String
	}
	if renewedFrom.Valid {
		term.RenewedFromTermID = renewedFrom.String
	}
	term.StartPolicy = StartPolicy(startPolicy)
	term.State = TermState(state)
	return term, nil
}

func (s *PostgresStore) ApplyCustomerMetadataTx(
	ctx context.Context,
	tx *sql.Tx,
	tenantID, userID, actorID, groupID string,
	tagIDs []string,
	onHold bool,
) error {
	if tx == nil {
		return errors.New("customer: transaction is required")
	}
	if groupID != "" {
		var allowed bool
		if err := tx.QueryRowContext(ctx, `
SELECT EXISTS(
    SELECT 1 FROM pvnaive.customer_groups
    WHERE id=$1::uuid AND tenant_id=$2::uuid
)`, groupID, tenantID).Scan(&allowed); err != nil || !allowed {
			return ErrInvalidCustomerMetadata
		}
	}
	if _, err := tx.ExecContext(ctx, `
INSERT INTO pvnaive.customer_profiles (
    tenant_id, user_id, group_id, assigned_actor_id, on_hold, updated_by_actor_id
) VALUES (
    $1::uuid, $2::uuid, NULLIF($4,'')::uuid, $3::uuid, $5, $3::uuid
)
ON CONFLICT (user_id) DO UPDATE SET
    group_id=EXCLUDED.group_id,
    assigned_actor_id=EXCLUDED.assigned_actor_id,
    on_hold=EXCLUDED.on_hold,
    updated_by_actor_id=EXCLUDED.updated_by_actor_id,
    revision=pvnaive.customer_profiles.revision+1,
    updated_at=clock_timestamp()`, tenantID, userID, actorID, groupID, onHold); err != nil {
		return fmt.Errorf("customer: apply customer metadata: %w", err)
	}
	for _, tagID := range uniqueStrings(tagIDs) {
		result, err := tx.ExecContext(ctx, `
INSERT INTO pvnaive.customer_tag_assignments (tenant_id,user_id,tag_id,assigned_by_actor_id)
SELECT $1::uuid,$2::uuid,t.id,$4::uuid
FROM pvnaive.customer_tags t
WHERE t.id=$3::uuid AND t.tenant_id=$1::uuid
ON CONFLICT (user_id,tag_id) DO NOTHING`, tenantID, userID, tagID, actorID)
		if err != nil {
			return fmt.Errorf("customer: assign customer tag: %w", err)
		}
		if count, _ := result.RowsAffected(); count == 0 {
			var exists bool
			_ = tx.QueryRowContext(ctx, `
SELECT EXISTS(
    SELECT 1 FROM pvnaive.customer_tag_assignments
    WHERE tenant_id=$1::uuid AND user_id=$2::uuid AND tag_id=$3::uuid
)`, tenantID, userID, tagID).Scan(&exists)
			if !exists {
				return ErrInvalidCustomerMetadata
			}
		}
	}
	return nil
}
