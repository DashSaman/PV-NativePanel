package customer

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

const (
	bulkOperationsTable      = "pvnaive.customer_bulk_operations"
	bulkResetOperationsTable = "pvnaive.customer_bulk_reset_operations"
	bulkOperationKeysTable   = "pvnaive.customer_bulk_operation_keys"
)

func bulkOperationTable(action BulkAction) string {
	if action == BulkResetUsage {
		return bulkResetOperationsTable
	}
	return bulkOperationsTable
}

func (s *PostgresStore) BulkCustomersTx(ctx context.Context, tx *sql.Tx, ids []string) ([]BulkCustomer, error) {
	if tx == nil {
		return nil, errors.New("customer: transaction is required")
	}
	ids = uniqueStrings(ids)
	if len(ids) == 0 {
		return []BulkCustomer{}, nil
	}
	rows, err := tx.QueryContext(ctx, `
SELECT id::text,
       CASE status
         WHEN 'draft' THEN 'disabled'
         WHEN 'suspended' THEN 'suspended'
         WHEN 'revoked' THEN 'revoked'
         ELSE 'active'
       END
FROM pvnaive.users
WHERE id::text = ANY(string_to_array($1, ','))`, strings.Join(ids, ","))
	if err != nil {
		return nil, fmt.Errorf("customer: load bulk customers: %w", err)
	}
	defer rows.Close()
	out := make([]BulkCustomer, 0, len(ids))
	for rows.Next() {
		var item BulkCustomer
		var lifecycle string
		if err := rows.Scan(&item.ID, &lifecycle); err != nil {
			return nil, fmt.Errorf("customer: scan bulk customer: %w", err)
		}
		item.Lifecycle = LifecycleStatus(lifecycle)
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("customer: bulk customer rows: %w", err)
	}
	return out, nil
}

func (s *PostgresStore) ClaimBulkPreviewTx(
	ctx context.Context,
	tx *sql.Tx,
	tenantID, actorID, idempotencyKey string,
	requestHash []byte,
	request BulkRequest,
	preview BulkPreview,
) (BulkOperation, []byte, error) {
	if tx == nil || len(requestHash) != 32 {
		return BulkOperation{}, nil, ErrInvalidBulkRequest
	}
	requestJSON, err := json.Marshal(request)
	if err != nil {
		return BulkOperation{}, nil, err
	}
	previewJSON, err := json.Marshal(preview)
	if err != nil {
		return BulkOperation{}, nil, err
	}
	// Schema15 establishes one idempotency namespace across the legacy bulk
	// ledger and the reset-specific ledger. The backfill-on-claim also covers
	// any legacy operation created during a guarded schema/code rollout window.
	if _, err := tx.ExecContext(ctx, `
INSERT INTO pvnaive.customer_bulk_operation_keys (
    tenant_id,actor_id,idempotency_key,request_hash,action,created_at
)
SELECT tenant_id,actor_id,idempotency_key,request_hash,action,created_at
FROM pvnaive.customer_bulk_operations
WHERE tenant_id=$1::uuid AND idempotency_key=$2
ON CONFLICT (tenant_id,idempotency_key) DO NOTHING`, tenantID, idempotencyKey); err != nil {
		return BulkOperation{}, nil, fmt.Errorf("customer: backfill bulk idempotency key: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `
INSERT INTO pvnaive.customer_bulk_operation_keys (
    tenant_id,actor_id,idempotency_key,request_hash,action
) VALUES ($1::uuid,$2::uuid,$3,$4,$5)
ON CONFLICT (tenant_id,idempotency_key) DO NOTHING`,
		tenantID, actorID, idempotencyKey, requestHash, string(request.Action)); err != nil {
		return BulkOperation{}, nil, fmt.Errorf("customer: claim bulk idempotency key: %w", err)
	}
	var registryActor, registryAction string
	var registryHash []byte
	if err := tx.QueryRowContext(ctx, `
SELECT actor_id::text,request_hash,action
FROM pvnaive.customer_bulk_operation_keys
WHERE tenant_id=$1::uuid AND idempotency_key=$2`, tenantID, idempotencyKey).Scan(&registryActor, &registryHash, &registryAction); err != nil {
		return BulkOperation{}, nil, fmt.Errorf("customer: inspect bulk idempotency key: %w", err)
	}
	if registryActor != actorID || registryAction != string(request.Action) || !bytes.Equal(registryHash, requestHash) {
		return BulkOperation{}, nil, ErrBulkConflict
	}

	table := bulkOperationTable(request.Action)
	query := fmt.Sprintf(`
INSERT INTO %s (
    tenant_id,actor_id,idempotency_key,request_hash,action,request,status,preview
) VALUES ($1::uuid,$2::uuid,$3,$4,$5,$6::jsonb,'previewed',$7::jsonb)
ON CONFLICT (tenant_id,idempotency_key) DO NOTHING`, table)
	if _, err := tx.ExecContext(ctx, query,
		tenantID, actorID, idempotencyKey, requestHash, string(request.Action), string(requestJSON), string(previewJSON)); err != nil {
		return BulkOperation{}, nil, fmt.Errorf("customer: claim bulk preview: %w", err)
	}
	return s.bulkOperationRow(ctx, tx, actorID, idempotencyKey, request.Action)
}

func (s *PostgresStore) BulkOperationTx(ctx context.Context, tx *sql.Tx, actorID, idempotencyKey string, action BulkAction) (BulkOperation, []byte, error) {
	if tx == nil {
		return BulkOperation{}, nil, errors.New("customer: transaction is required")
	}
	return s.bulkOperationRow(ctx, tx, actorID, idempotencyKey, action)
}

func (s *PostgresStore) bulkOperationRow(ctx context.Context, tx *sql.Tx, actorID, idempotencyKey string, expectedAction BulkAction) (BulkOperation, []byte, error) {
	var operation BulkOperation
	var action string
	var requestHash []byte
	var requestJSON, previewJSON []byte
	var resultJSON []byte
	var resultRaw sql.RawBytes
	table := bulkOperationTable(expectedAction)
	query := fmt.Sprintf(`
SELECT id::text,action,status,request_hash,request,preview,result
FROM %s
WHERE actor_id=$1::uuid AND idempotency_key=$2
LIMIT 1`, table)
	err := tx.QueryRowContext(ctx, query, actorID, idempotencyKey).Scan(
		&operation.ID, &action, &operation.Status, &requestHash, &requestJSON, &previewJSON, &resultRaw,
	)
	if err != nil {
		return BulkOperation{}, nil, fmt.Errorf("customer: load bulk operation: %w", err)
	}
	operation.Action = BulkAction(action)
	if operation.Action != expectedAction {
		return BulkOperation{}, nil, ErrBulkConflict
	}
	operation.IdempotencyKey = idempotencyKey
	if err := json.Unmarshal(requestJSON, &operation.Request); err != nil {
		return BulkOperation{}, nil, fmt.Errorf("customer: decode bulk request: %w", err)
	}
	if err := json.Unmarshal(previewJSON, &operation.Preview); err != nil {
		return BulkOperation{}, nil, fmt.Errorf("customer: decode bulk preview: %w", err)
	}
	if resultRaw != nil {
		resultJSON = append([]byte(nil), resultRaw...)
		var result BulkExecutionResult
		if len(resultJSON) > 0 && string(resultJSON) != "null" {
			if err := json.Unmarshal(resultJSON, &result); err != nil {
				return BulkOperation{}, nil, fmt.Errorf("customer: decode bulk result: %w", err)
			}
			operation.Result = &result
		}
	}
	return operation, append([]byte(nil), requestHash...), nil
}

func (s *PostgresStore) MarkBulkExecutedTx(
	ctx context.Context,
	tx *sql.Tx,
	actorID, idempotencyKey string,
	action BulkAction,
	result BulkExecutionResult,
) (BulkOperation, error) {
	if tx == nil {
		return BulkOperation{}, errors.New("customer: transaction is required")
	}
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return BulkOperation{}, err
	}
	table := bulkOperationTable(action)
	query := fmt.Sprintf(`
UPDATE %s
SET status='executed',result=$3::jsonb,executed_at=COALESCE(executed_at,clock_timestamp())
WHERE actor_id=$1::uuid AND idempotency_key=$2 AND status IN ('previewed','executed')`, table)
	command, err := tx.ExecContext(ctx, query, actorID, idempotencyKey, string(resultJSON))
	if err != nil {
		return BulkOperation{}, fmt.Errorf("customer: mark bulk executed: %w", err)
	}
	if count, _ := command.RowsAffected(); count == 0 {
		return BulkOperation{}, ErrBulkNotPreviewed
	}
	operation, _, err := s.bulkOperationRow(ctx, tx, actorID, idempotencyKey, action)
	return operation, err
}
