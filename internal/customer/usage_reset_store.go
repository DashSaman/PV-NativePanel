package customer

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

func (s *PostgresStore) UsageResetTargetTx(ctx context.Context, tx *sql.Tx, userID string) (UsageResetTarget, error) {
	if tx == nil {
		return UsageResetTarget{}, errors.New("customer: transaction is required")
	}
	var target UsageResetTarget
	err := tx.QueryRowContext(ctx, `
SELECT u.tenant_id::text, u.id::text, st.id::text
FROM pvnaive.users AS u
JOIN LATERAL (
    SELECT candidate.id
    FROM pvnaive.service_terms AS candidate
    WHERE candidate.user_id = u.id AND candidate.tenant_id = u.tenant_id
    ORDER BY candidate.purchased_at DESC, candidate.created_at DESC
    LIMIT 1
) AS st ON TRUE
WHERE u.id = $1::uuid`, userID).Scan(&target.TenantID, &target.UserID, &target.ServiceTermID)
	if errors.Is(err, sql.ErrNoRows) {
		return UsageResetTarget{}, ErrUsageResetNotFound
	}
	if err != nil {
		return UsageResetTarget{}, fmt.Errorf("customer: resolve usage reset target: %w", err)
	}
	return target, nil
}

func (s *PostgresStore) ClaimUsageResetTx(ctx context.Context, tx *sql.Tx, target UsageResetTarget, actorID, idempotencyKey string, requestHash []byte) (string, bool, error) {
	if tx == nil || len(requestHash) != 32 {
		return "", false, errors.New("customer: invalid usage reset idempotency request")
	}
	const operation = "customer.usage.reset"
	var id string
	err := tx.QueryRowContext(ctx, `
INSERT INTO pvnaive.customer_mutation_keys (
    tenant_id, actor_id, idempotency_key, operation, request_hash, resource_type, resource_id
) VALUES ($1::uuid,$2::uuid,$3,$4,$5,'user',$6::uuid)
ON CONFLICT (tenant_id, actor_id, idempotency_key) DO NOTHING
RETURNING id::text`, target.TenantID, actorID, idempotencyKey, operation, requestHash, target.UserID).Scan(&id)
	if err == nil {
		return id, true, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return "", false, fmt.Errorf("customer: claim usage reset idempotency: %w", err)
	}
	var existingOperation, resourceType string
	var existingHash []byte
	var resourceID sql.NullString
	if err := tx.QueryRowContext(ctx, `
SELECT id::text, operation, request_hash, resource_type, resource_id::text
FROM pvnaive.customer_mutation_keys
WHERE tenant_id=$1::uuid AND actor_id=$2::uuid AND idempotency_key=$3`,
		target.TenantID, actorID, idempotencyKey).Scan(&id, &existingOperation, &existingHash, &resourceType, &resourceID); err != nil {
		return "", false, fmt.Errorf("customer: inspect usage reset idempotency: %w", err)
	}
	if existingOperation != operation || resourceType != "user" || !resourceID.Valid || resourceID.String != target.UserID || !bytes.Equal(existingHash, requestHash) {
		return "", false, ErrCustomerIdempotencyConflict
	}
	return id, false, nil
}

func (s *PostgresStore) UsageResetEventByMutationKeyTx(ctx context.Context, tx *sql.Tx, mutationID string) (UsageResetEvent, error) {
	if tx == nil {
		return UsageResetEvent{}, errors.New("customer: transaction is required")
	}
	var event UsageResetEvent
	var reason string
	err := tx.QueryRowContext(ctx, `
SELECT id::text, tenant_id::text, user_id::text, service_term_id::text,
       actor_id::text, customer_mutation_key_id::text, reason, reset_at,
       previous_upload_bytes, previous_download_bytes, previous_used_bytes
FROM pvnaive.direct_naive_accounting_reset_events
WHERE customer_mutation_key_id=$1::uuid`, mutationID).Scan(
		&event.ID, &event.TenantID, &event.UserID, &event.ServiceTermID,
		&event.ActorID, &event.MutationKeyID, &reason, &event.ResetAt,
		&event.PreviousUploadBytes, &event.PreviousDownloadBytes, &event.PreviousUsedBytes,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return UsageResetEvent{}, ErrUsageResetUnavailable
	}
	if err != nil {
		return UsageResetEvent{}, fmt.Errorf("customer: load usage reset replay: %w", err)
	}
	event.Reason = UsageResetReason(reason)
	return event, nil
}

func (s *PostgresStore) ResetDirectAccountingTx(ctx context.Context, tx *sql.Tx, serviceTermID string, resetAt time.Time, staleAfter time.Duration) (AccountingResetResult, error) {
	if tx == nil || resetAt.IsZero() || staleAfter <= 0 {
		return AccountingResetResult{}, errors.New("customer: invalid direct accounting reset request")
	}
	seconds := int64(staleAfter / time.Second)
	if seconds <= 0 {
		seconds = 1
	}
	var out AccountingResetResult
	var tenantID, userID sql.NullString
	var state sql.NullString
	if err := tx.QueryRowContext(ctx, `
SELECT service_term_id::text, tenant_id::text, user_id::text, resettable, reason,
       previous_upload_bytes, previous_download_bytes, previous_used_bytes,
       reset_at, service_state
FROM pvnaive.direct_naive_accounting_reset($1::uuid,$2,$3)`, serviceTermID, resetAt.UTC(), seconds).Scan(
		&out.ServiceTermID, &tenantID, &userID, &out.Resettable, &out.Reason,
		&out.PreviousUploadBytes, &out.PreviousDownloadBytes, &out.PreviousUsedBytes,
		&out.ResetAt, &state,
	); err != nil {
		return AccountingResetResult{}, fmt.Errorf("customer: reset direct accounting: %w", err)
	}
	if tenantID.Valid {
		out.TenantID = tenantID.String
	}
	if userID.Valid {
		out.UserID = userID.String
	}
	if state.Valid {
		out.ServiceState = TermState(state.String)
	}
	return out, nil
}

func (s *PostgresStore) AppendUsageResetEventTx(ctx context.Context, tx *sql.Tx, target UsageResetTarget, actorID, mutationID string, reason UsageResetReason, reset AccountingResetResult) (UsageResetEvent, error) {
	if tx == nil || !reset.Resettable || (reason != UsageResetManual && reason != UsageResetBulk && reason != UsageResetScheduled) {
		return UsageResetEvent{}, errors.New("customer: valid usage reset result is required")
	}
	var event UsageResetEvent
	var reasonText string
	if err := tx.QueryRowContext(ctx, `
INSERT INTO pvnaive.direct_naive_accounting_reset_events (
    tenant_id,user_id,service_term_id,actor_id,customer_mutation_key_id,reason,reset_at,
    previous_upload_bytes,previous_download_bytes,previous_used_bytes
) VALUES ($1::uuid,$2::uuid,$3::uuid,$4::uuid,$5::uuid,$6,$7,$8,$9,$10)
RETURNING id::text,tenant_id::text,user_id::text,service_term_id::text,actor_id::text,
          customer_mutation_key_id::text,reason,reset_at,previous_upload_bytes,previous_download_bytes,previous_used_bytes`,
		target.TenantID, target.UserID, target.ServiceTermID, actorID, mutationID, string(reason), reset.ResetAt.UTC(),
		reset.PreviousUploadBytes, reset.PreviousDownloadBytes, reset.PreviousUsedBytes,
	).Scan(&event.ID, &event.TenantID, &event.UserID, &event.ServiceTermID, &event.ActorID,
		&event.MutationKeyID, &reasonText, &event.ResetAt, &event.PreviousUploadBytes,
		&event.PreviousDownloadBytes, &event.PreviousUsedBytes); err != nil {
		return UsageResetEvent{}, fmt.Errorf("customer: append usage reset event: %w", err)
	}
	event.Reason = UsageResetReason(reasonText)
	if _, err := tx.ExecContext(ctx, `
UPDATE pvnaive.customer_mutation_keys
SET completed_at=clock_timestamp()
WHERE id=$1::uuid AND completed_at IS NULL`, mutationID); err != nil {
		return UsageResetEvent{}, fmt.Errorf("customer: complete usage reset idempotency: %w", err)
	}
	return event, nil
}

func (s *PostgresStore) AppendUsageResetAuditTx(ctx context.Context, tx *sql.Tx, event UsageResetEvent) error {
	if tx == nil {
		return errors.New("customer: transaction is required")
	}
	before, err := json.Marshal(map[string]any{
		"upload_bytes":   event.PreviousUploadBytes,
		"download_bytes": event.PreviousDownloadBytes,
		"used_bytes":     event.PreviousUsedBytes,
	})
	if err != nil {
		return fmt.Errorf("customer: encode usage reset audit before state: %w", err)
	}
	after, err := json.Marshal(map[string]any{"upload_bytes": 0, "download_bytes": 0, "used_bytes": 0, "period_started_at": event.ResetAt.UTC()})
	if err != nil {
		return fmt.Errorf("customer: encode usage reset audit after state: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `
INSERT INTO pvnaive.audit_events (
    tenant_id,actor_id,action,object_type,object_id,outcome,before_state,after_state
) VALUES ($1::uuid,$2::uuid,'customer.usage.reset','service_term',$3::uuid,'success',$4::jsonb,$5::jsonb)`,
		event.TenantID, event.ActorID, event.ServiceTermID, before, after); err != nil {
		return fmt.Errorf("customer: append usage reset audit: %w", err)
	}
	return nil
}
