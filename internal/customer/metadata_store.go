package customer

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

func (s *PostgresStore) UpdateCustomerMetadataTx(
	ctx context.Context,
	tx *sql.Tx,
	actorID, userID string,
	input CustomerMetadataInput,
) (CustomerMetadata, error) {
	if tx == nil {
		return CustomerMetadata{}, errors.New("customer: transaction is required")
	}

	var tenantID string
	if err := tx.QueryRowContext(ctx, `SELECT tenant_id::text FROM pvnaive.users WHERE id=$1::uuid`, userID).Scan(&tenantID); err != nil {
		return CustomerMetadata{}, fmt.Errorf("customer: resolve metadata tenant: %w", err)
	}

	noteSet := input.Note != nil
	note := ""
	if input.Note != nil {
		note = *input.Note
	}
	groupSet := input.GroupID != nil
	groupID := ""
	if input.GroupID != nil {
		groupID = *input.GroupID
		if groupID != "" {
			var allowed bool
			if err := tx.QueryRowContext(ctx, `
SELECT EXISTS(
    SELECT 1 FROM pvnaive.customer_groups
    WHERE id=$1::uuid AND tenant_id=$2::uuid
)`, groupID, tenantID).Scan(&allowed); err != nil || !allowed {
				return CustomerMetadata{}, ErrInvalidCustomerMetadata
			}
		}
	}
	holdSet := input.OnHold != nil
	onHold := false
	if input.OnHold != nil {
		onHold = *input.OnHold
	}

	if _, err := tx.ExecContext(ctx, `
INSERT INTO pvnaive.customer_profiles (
    tenant_id, user_id, note, group_id, assigned_actor_id, on_hold, updated_by_actor_id
) VALUES (
    $1::uuid, $2::uuid,
    CASE WHEN $4 THEN $3 ELSE '' END,
    CASE WHEN $6 THEN NULLIF($5,'')::uuid ELSE NULL END,
    $7::uuid,
    CASE WHEN $9 THEN $8 ELSE false END,
    $7::uuid
)
ON CONFLICT (user_id) DO UPDATE SET
    note=CASE WHEN $4 THEN EXCLUDED.note ELSE pvnaive.customer_profiles.note END,
    group_id=CASE WHEN $6 THEN EXCLUDED.group_id ELSE pvnaive.customer_profiles.group_id END,
    assigned_actor_id=$7::uuid,
    on_hold=CASE WHEN $9 THEN EXCLUDED.on_hold ELSE pvnaive.customer_profiles.on_hold END,
    updated_by_actor_id=$7::uuid,
    revision=pvnaive.customer_profiles.revision+1,
    updated_at=clock_timestamp()`, tenantID, userID, note, noteSet, groupID, groupSet, actorID, onHold, holdSet); err != nil {
		return CustomerMetadata{}, fmt.Errorf("customer: update metadata profile: %w", err)
	}

	for _, tagID := range input.AddTagIDs {
		result, err := tx.ExecContext(ctx, `
INSERT INTO pvnaive.customer_tag_assignments (tenant_id,user_id,tag_id,assigned_by_actor_id)
SELECT u.tenant_id,u.id,t.id,$3::uuid
FROM pvnaive.users u
JOIN pvnaive.customer_tags t ON t.id=$2::uuid AND t.tenant_id=u.tenant_id
WHERE u.id=$1::uuid
ON CONFLICT (user_id,tag_id) DO NOTHING`, userID, tagID, actorID)
		if err != nil {
			return CustomerMetadata{}, fmt.Errorf("customer: add metadata tag: %w", err)
		}
		if count, _ := result.RowsAffected(); count == 0 {
			var exists bool
			_ = tx.QueryRowContext(ctx, `
SELECT EXISTS(
  SELECT 1 FROM pvnaive.customer_tag_assignments
  WHERE user_id=$1::uuid AND tag_id=$2::uuid AND tenant_id=$3::uuid
)`, userID, tagID, tenantID).Scan(&exists)
			if !exists {
				return CustomerMetadata{}, ErrInvalidCustomerMetadata
			}
		}
	}
	for _, tagID := range input.RemoveTagIDs {
		if _, err := tx.ExecContext(ctx, `
DELETE FROM pvnaive.customer_tag_assignments
WHERE user_id=$1::uuid AND tag_id=$2::uuid AND tenant_id=$3::uuid`, userID, tagID, tenantID); err != nil {
			return CustomerMetadata{}, fmt.Errorf("customer: remove metadata tag: %w", err)
		}
	}

	var out CustomerMetadata
	var group CustomerGroup
	var groupIDOut sql.NullString
	var groupName sql.NullString
	var groupEnabled sql.NullBool
	var groupSort sql.NullInt64
	if err := tx.QueryRowContext(ctx, `
SELECT cp.note, cp.on_hold, COALESCE(cp.assigned_actor_id::text,''),
       cg.id::text, cg.name, cg.enabled, cg.sort_order
FROM pvnaive.customer_profiles cp
LEFT JOIN pvnaive.customer_groups cg ON cg.id=cp.group_id AND cg.tenant_id=cp.tenant_id
WHERE cp.user_id=$1::uuid`, userID).Scan(
		&out.Note, &out.OnHold, &out.AssignedActorID,
		&groupIDOut, &groupName, &groupEnabled, &groupSort,
	); err != nil {
		return CustomerMetadata{}, fmt.Errorf("customer: read metadata profile: %w", err)
	}
	if groupIDOut.Valid {
		group.ID = groupIDOut.String
		group.Name = groupName.String
		group.Enabled = groupEnabled.Bool
		group.SortOrder = int(groupSort.Int64)
		out.Group = &group
	}
	out.Tags = make([]CustomerTag, 0)
	rows, err := tx.QueryContext(ctx, `
SELECT t.id::text,t.name,t.enabled,t.sort_order
FROM pvnaive.customer_tag_assignments a
JOIN pvnaive.customer_tags t ON t.id=a.tag_id AND t.tenant_id=a.tenant_id
WHERE a.user_id=$1::uuid
ORDER BY t.sort_order,t.name`, userID)
	if err != nil {
		return CustomerMetadata{}, fmt.Errorf("customer: read metadata tags: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var tag CustomerTag
		if err := rows.Scan(&tag.ID, &tag.Name, &tag.Enabled, &tag.SortOrder); err != nil {
			return CustomerMetadata{}, fmt.Errorf("customer: scan metadata tag: %w", err)
		}
		out.Tags = append(out.Tags, tag)
	}
	if err := rows.Err(); err != nil {
		return CustomerMetadata{}, fmt.Errorf("customer: metadata tags: %w", err)
	}
	return out, nil
}
