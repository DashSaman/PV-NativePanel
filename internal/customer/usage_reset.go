package customer

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"strings"
	"time"
)

type UsageResetReason string

const (
	UsageResetManual    UsageResetReason = "manual"
	UsageResetBulk      UsageResetReason = "bulk"
	UsageResetScheduled UsageResetReason = "scheduled"
)

var (
	ErrUsageResetUnavailable          = errors.New("customer: usage reset is unavailable")
	ErrUsageResetNotFound             = errors.New("customer: usage reset target not found")
	ErrUsageResetServiceInactive      = errors.New("customer: inactive service cannot be reset")
	ErrUsageResetAccountingIncomplete = errors.New("customer: accounting is incomplete")
	ErrUsageResetReservationPending   = errors.New("customer: accounting reservation is pending")
	ErrUsageResetTelemetryStale       = errors.New("customer: accounting telemetry is stale")
	ErrUsageResetTimeConflict         = errors.New("customer: usage reset time conflicts with current period")
)

const usageResetStaleAfter = 90 * time.Second

type UsageResetTarget struct {
	TenantID      string
	UserID        string
	ServiceTermID string
}

type AccountingResetResult struct {
	ServiceTermID         string
	TenantID              string
	UserID                string
	Resettable            bool
	Reason                string
	PreviousUploadBytes   int64
	PreviousDownloadBytes int64
	PreviousUsedBytes     int64
	ResetAt               time.Time
	ServiceState          TermState
}

type UsageResetEvent struct {
	ID                    string           `json:"id"`
	TenantID              string           `json:"-"`
	UserID                string           `json:"user_id"`
	ServiceTermID         string           `json:"service_term_id"`
	ActorID               string           `json:"-"`
	MutationKeyID         string           `json:"-"`
	Reason                UsageResetReason `json:"reason"`
	ResetAt               time.Time        `json:"reset_at"`
	PreviousUploadBytes   int64            `json:"previous_upload_bytes"`
	PreviousDownloadBytes int64            `json:"previous_download_bytes"`
	PreviousUsedBytes     int64            `json:"previous_used_bytes"`
}

type UsageResetResult struct {
	Event            UsageResetEvent `json:"reset_event"`
	IdempotentReplay bool            `json:"idempotent_replay"`
}

type usageResetStore interface {
	UsageResetTargetTx(context.Context, *sql.Tx, string) (UsageResetTarget, error)
	ClaimUsageResetTx(context.Context, *sql.Tx, UsageResetTarget, string, string, []byte) (string, bool, error)
	UsageResetEventByMutationKeyTx(context.Context, *sql.Tx, string) (UsageResetEvent, error)
	ResetDirectAccountingTx(context.Context, *sql.Tx, string, time.Time, time.Duration) (AccountingResetResult, error)
	AppendUsageResetEventTx(context.Context, *sql.Tx, UsageResetTarget, string, string, AccountingResetResult) (UsageResetEvent, error)
	AppendUsageResetAuditTx(context.Context, *sql.Tx, UsageResetEvent) error
}

func (s *Service) ResetCustomerUsage(ctx context.Context, tx *sql.Tx, actorID, idempotencyKey, userID string) (UsageResetResult, error) {
	if s == nil || s.store == nil || strings.TrimSpace(actorID) == "" || strings.TrimSpace(userID) == "" {
		return UsageResetResult{}, ErrUsageResetUnavailable
	}
	if len(idempotencyKey) < 8 || len(idempotencyKey) > 160 || strings.TrimSpace(idempotencyKey) != idempotencyKey {
		return UsageResetResult{}, errors.New("customer: invalid usage reset idempotency key")
	}
	store, ok := s.store.(usageResetStore)
	if !ok {
		return UsageResetResult{}, ErrUsageResetUnavailable
	}
	target, err := store.UsageResetTargetTx(ctx, tx, userID)
	if err != nil {
		return UsageResetResult{}, err
	}
	requestHash := sha256.Sum256([]byte("customer.usage.reset\n" + target.UserID + "\n" + target.ServiceTermID))
	mutationID, claimed, err := store.ClaimUsageResetTx(ctx, tx, target, actorID, idempotencyKey, requestHash[:])
	if err != nil {
		return UsageResetResult{}, err
	}
	if !claimed {
		event, err := store.UsageResetEventByMutationKeyTx(ctx, tx, mutationID)
		if err != nil {
			return UsageResetResult{}, err
		}
		return UsageResetResult{Event: event, IdempotentReplay: true}, nil
	}

	reset, err := store.ResetDirectAccountingTx(ctx, tx, target.ServiceTermID, s.now().UTC(), usageResetStaleAfter)
	if err != nil {
		return UsageResetResult{}, err
	}
	if !reset.Resettable {
		return UsageResetResult{}, usageResetReasonError(reset.Reason)
	}
	event, err := store.AppendUsageResetEventTx(ctx, tx, target, actorID, mutationID, reset)
	if err != nil {
		return UsageResetResult{}, err
	}
	if err := store.AppendUsageResetAuditTx(ctx, tx, event); err != nil {
		return UsageResetResult{}, err
	}
	return UsageResetResult{Event: event}, nil
}

func usageResetReasonError(reason string) error {
	switch reason {
	case "not_found":
		return ErrUsageResetNotFound
	case "service_inactive":
		return ErrUsageResetServiceInactive
	case "accounting_incomplete":
		return ErrUsageResetAccountingIncomplete
	case "reservation_pending":
		return ErrUsageResetReservationPending
	case "telemetry_stale":
		return ErrUsageResetTelemetryStale
	case "reset_time_conflict":
		return ErrUsageResetTimeConflict
	default:
		return ErrUsageResetUnavailable
	}
}
