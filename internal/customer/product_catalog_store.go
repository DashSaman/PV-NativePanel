package customer

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

func (s *PostgresStore) OperationTenantIDTx(ctx context.Context, tx *sql.Tx) (string, error) {
	if tx == nil {
		return "", errors.New("customer: transaction is required")
	}
	var tenantID string
	if err := tx.QueryRowContext(ctx, `
SELECT COALESCE(
    pvnaive.current_tenant_id(),
    (SELECT id FROM pvnaive.tenants WHERE tenant_type='system' AND slug='direct' AND status='active' LIMIT 1)
)::text`).Scan(&tenantID); err != nil {
		return "", fmt.Errorf("customer: resolve operation tenant: %w", err)
	}
	return tenantID, nil
}

func (s *PostgresStore) ListPlansTx(ctx context.Context, tx *sql.Tx) ([]PlanPreset, error) {
	if tx == nil {
		return nil, errors.New("customer: transaction is required")
	}
	rows, err := tx.QueryContext(ctx, `
SELECT id::text, name, quota_bytes,
       CASE WHEN no_expiry THEN 0 ELSE duration_seconds END,
       no_expiry, start_policy, reset_strategy, COALESCE(reset_custom_days,0),
       COALESCE(default_group_id::text,''), enabled, sort_order
FROM pvnaive.plans
WHERE status <> 'archived'
ORDER BY enabled DESC, sort_order ASC, name ASC`)
	if err != nil {
		return nil, fmt.Errorf("customer: list plans: %w", err)
	}
	defer rows.Close()
	out := make([]PlanPreset, 0)
	for rows.Next() {
		var plan PlanPreset
		var startPolicy, resetStrategy string
		if err := rows.Scan(
			&plan.ID, &plan.Name, &plan.QuotaBytes, &plan.ValiditySeconds,
			&plan.NoExpiry, &startPolicy, &resetStrategy, &plan.ResetCustomDays,
			&plan.DefaultGroupID, &plan.Enabled, &plan.SortOrder,
		); err != nil {
			return nil, fmt.Errorf("customer: scan plan: %w", err)
		}
		plan.StartPolicy = StartPolicy(startPolicy)
		plan.ResetStrategy = ResetStrategy(resetStrategy)
		plan.ResetEnforcement = false
		out = append(out, plan)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("customer: list plan rows: %w", err)
	}
	return out, nil
}

func (s *PostgresStore) CreatePlanTx(ctx context.Context, tx *sql.Tx, tenantID, actorID, code string, plan PlanPreset) (PlanPreset, error) {
	if tx == nil {
		return PlanPreset{}, errors.New("customer: transaction is required")
	}
	if plan.DefaultGroupID != "" {
		var allowed bool
		if err := tx.QueryRowContext(ctx, `
SELECT EXISTS(
    SELECT 1 FROM pvnaive.customer_groups
    WHERE id=$1::uuid AND tenant_id=$2::uuid
)`, plan.DefaultGroupID, tenantID).Scan(&allowed); err != nil || !allowed {
			return PlanPreset{}, ErrInvalidCustomerMetadata
		}
	}
	duration := plan.ValiditySeconds
	if plan.NoExpiry {
		duration = 86400
	}
	var out PlanPreset
	var startPolicy, resetStrategy string
	var groupID sql.NullString
	err := tx.QueryRowContext(ctx, `
INSERT INTO pvnaive.plans (
    tenant_id, code, name, status, protocol_id, quota_bytes, duration_seconds,
    base_price_minor, currency, start_policy, no_expiry, reset_strategy,
    reset_custom_days, enabled, sort_order, default_group_id
) VALUES (
    $1::uuid, $2, $3, 'active', 'naive', $4, $5,
    0, 'USD', $6, $7, $8, NULLIF($9,0), $10, $11, NULLIF($12,'')::uuid
)
RETURNING id::text, name, quota_bytes,
          CASE WHEN no_expiry THEN 0 ELSE duration_seconds END,
          no_expiry, start_policy, reset_strategy, COALESCE(reset_custom_days,0),
          default_group_id::text, enabled, sort_order`,
		tenantID, code, plan.Name, plan.QuotaBytes, duration, string(plan.StartPolicy),
		plan.NoExpiry, string(plan.ResetStrategy), plan.ResetCustomDays, plan.Enabled,
		plan.SortOrder, plan.DefaultGroupID,
	).Scan(
		&out.ID, &out.Name, &out.QuotaBytes, &out.ValiditySeconds, &out.NoExpiry,
		&startPolicy, &resetStrategy, &out.ResetCustomDays, &groupID, &out.Enabled, &out.SortOrder,
	)
	if err != nil {
		return PlanPreset{}, fmt.Errorf("customer: create plan: %w", err)
	}
	out.StartPolicy = StartPolicy(startPolicy)
	out.ResetStrategy = ResetStrategy(resetStrategy)
	out.ResetEnforcement = false
	if groupID.Valid {
		out.DefaultGroupID = groupID.String
	}
	if len(plan.TagIDs) > 0 {
		for _, tagID := range uniqueStrings(plan.TagIDs) {
			result, err := tx.ExecContext(ctx, `
INSERT INTO pvnaive.plan_tag_assignments (plan_id, tag_id, tenant_id)
SELECT $1::uuid,t.id,$3::uuid
FROM pvnaive.customer_tags t
WHERE t.id=$2::uuid AND t.tenant_id=$3::uuid`, out.ID, tagID, tenantID)
			if err != nil {
				return PlanPreset{}, fmt.Errorf("customer: assign plan tag: %w", err)
			}
			if count, _ := result.RowsAffected(); count != 1 {
				return PlanPreset{}, ErrInvalidCustomerMetadata
			}
		}
	}
	_ = actorID
	return out, nil
}

func (s *PostgresStore) ListGroupsTx(ctx context.Context, tx *sql.Tx) ([]CustomerGroup, error) {
	if tx == nil {
		return nil, errors.New("customer: transaction is required")
	}
	rows, err := tx.QueryContext(ctx, `
SELECT id::text, name, enabled, sort_order
FROM pvnaive.customer_groups
ORDER BY enabled DESC, sort_order ASC, name ASC`)
	if err != nil {
		return nil, fmt.Errorf("customer: list groups: %w", err)
	}
	defer rows.Close()
	out := make([]CustomerGroup, 0)
	for rows.Next() {
		var group CustomerGroup
		if err := rows.Scan(&group.ID, &group.Name, &group.Enabled, &group.SortOrder); err != nil {
			return nil, fmt.Errorf("customer: scan group: %w", err)
		}
		out = append(out, group)
	}
	return out, rows.Err()
}

func (s *PostgresStore) CreateGroupTx(ctx context.Context, tx *sql.Tx, tenantID, actorID, name string, sortOrder int) (CustomerGroup, error) {
	if tx == nil {
		return CustomerGroup{}, errors.New("customer: transaction is required")
	}
	var out CustomerGroup
	if err := tx.QueryRowContext(ctx, `
INSERT INTO pvnaive.customer_groups (tenant_id,name,sort_order,created_by_actor_id)
VALUES ($1::uuid,$2,$3,$4::uuid)
RETURNING id::text,name,enabled,sort_order`, tenantID, name, sortOrder, actorID).
		Scan(&out.ID, &out.Name, &out.Enabled, &out.SortOrder); err != nil {
		return CustomerGroup{}, fmt.Errorf("customer: create group: %w", err)
	}
	return out, nil
}

func (s *PostgresStore) ListTagsTx(ctx context.Context, tx *sql.Tx) ([]CustomerTag, error) {
	if tx == nil {
		return nil, errors.New("customer: transaction is required")
	}
	rows, err := tx.QueryContext(ctx, `
SELECT id::text, name, enabled, sort_order
FROM pvnaive.customer_tags
ORDER BY enabled DESC, sort_order ASC, name ASC`)
	if err != nil {
		return nil, fmt.Errorf("customer: list tags: %w", err)
	}
	defer rows.Close()
	out := make([]CustomerTag, 0)
	for rows.Next() {
		var tag CustomerTag
		if err := rows.Scan(&tag.ID, &tag.Name, &tag.Enabled, &tag.SortOrder); err != nil {
			return nil, fmt.Errorf("customer: scan tag: %w", err)
		}
		out = append(out, tag)
	}
	return out, rows.Err()
}

func (s *PostgresStore) CreateTagTx(ctx context.Context, tx *sql.Tx, tenantID, actorID, name string, sortOrder int) (CustomerTag, error) {
	if tx == nil {
		return CustomerTag{}, errors.New("customer: transaction is required")
	}
	var out CustomerTag
	if err := tx.QueryRowContext(ctx, `
INSERT INTO pvnaive.customer_tags (tenant_id,name,sort_order,created_by_actor_id)
VALUES ($1::uuid,$2,$3,$4::uuid)
RETURNING id::text,name,enabled,sort_order`, tenantID, name, sortOrder, actorID).
		Scan(&out.ID, &out.Name, &out.Enabled, &out.SortOrder); err != nil {
		return CustomerTag{}, fmt.Errorf("customer: create tag: %w", err)
	}
	return out, nil
}
