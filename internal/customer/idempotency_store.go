package customer

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
)

var ErrCustomerIdempotencyConflict = errors.New("customer: idempotency key conflicts with another request")

func (s *PostgresStore) ClaimSubscriptionRotationTx(
	ctx context.Context,
	tx *sql.Tx,
	target SubscriptionTarget,
	actorID, idempotencyKey string,
	requestHash []byte,
) (bool, error) {
	if tx == nil {
		return false, errors.New("customer: transaction is required")
	}
	if len(requestHash) != 32 {
		return false, errors.New("customer: request hash must be 32 bytes")
	}
	const operation = "customer.subscription.rotate"
	var insertedID string
	err := tx.QueryRowContext(ctx, `
INSERT INTO pvnaive.customer_mutation_keys (
    tenant_id, actor_id, idempotency_key, operation, request_hash,
    resource_type, resource_id, completed_at
) VALUES ($1::uuid,$2::uuid,$3,$4,$5,'user',$6::uuid,clock_timestamp())
ON CONFLICT (tenant_id, actor_id, idempotency_key) DO NOTHING
RETURNING id::text`, target.TenantID, actorID, idempotencyKey, operation, requestHash, target.UserID).Scan(&insertedID)
	if err == nil {
		return true, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return false, fmt.Errorf("customer: claim subscription rotation idempotency: %w", err)
	}

	var existingOperation string
	var existingHash []byte
	var existingResourceType string
	var existingResourceID sql.NullString
	if err := tx.QueryRowContext(ctx, `
SELECT operation, request_hash, resource_type, resource_id::text
FROM pvnaive.customer_mutation_keys
WHERE tenant_id=$1::uuid AND actor_id=$2::uuid AND idempotency_key=$3`,
		target.TenantID, actorID, idempotencyKey,
	).Scan(&existingOperation, &existingHash, &existingResourceType, &existingResourceID); err != nil {
		return false, fmt.Errorf("customer: inspect subscription rotation idempotency: %w", err)
	}
	if existingOperation != operation || existingResourceType != "user" || !existingResourceID.Valid || existingResourceID.String != target.UserID || !bytes.Equal(existingHash, requestHash) {
		return false, ErrCustomerIdempotencyConflict
	}
	return false, nil
}
