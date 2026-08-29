package telemetry

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"
)

type AuthorizationResult struct {
	ServiceTermID      string     `json:"service_term_id,omitempty"`
	Tracked            bool       `json:"tracked"`
	Allowed            bool       `json:"allowed"`
	Reason             string     `json:"reason"`
	QuotaBytes         *int64     `json:"quota_bytes,omitempty"`
	UsedBytes          int64      `json:"used_bytes"`
	ReservedBytes      int64      `json:"reserved_bytes"`
	RemainingBytes     *int64     `json:"remaining_bytes,omitempty"`
	ExpiresAt          *time.Time `json:"expires_at,omitempty"`
	FirstConnectedAt   *time.Time `json:"first_connected_at,omitempty"`
	AccountingComplete bool       `json:"accounting_complete"`
}

type ClaimRequest struct {
	RuntimeCredentialID string    `json:"runtime_credential_id"`
	NodeID              string    `json:"node_id"`
	BootID              string    `json:"boot_id"`
	SessionID           string    `json:"session_id"`
	Sequence            int64     `json:"sequence"`
	Direction           string    `json:"direction"`
	RequestedBytes      int64     `json:"requested_bytes"`
	ObservedAt          time.Time `json:"timestamp"`
}

type ClaimResult struct {
	ServiceTermID      string     `json:"service_term_id,omitempty"`
	Tracked            bool       `json:"tracked"`
	Allowed            bool       `json:"allowed"`
	Reason             string     `json:"reason"`
	GrantedBytes       int64      `json:"granted_bytes"`
	QuotaBytes         *int64     `json:"quota_bytes,omitempty"`
	UsedBytes          int64      `json:"used_bytes"`
	ReservedBytes      int64      `json:"reserved_bytes"`
	RemainingBytes     *int64     `json:"remaining_bytes,omitempty"`
	ExpiresAt          *time.Time `json:"expires_at,omitempty"`
	AccountingComplete bool       `json:"accounting_complete"`
}

type AccountingStore interface {
	Authorize(context.Context, string, time.Time) (AuthorizationResult, error)
	Ingest(context.Context, Event) (IngestResult, error)
	Claim(context.Context, ClaimRequest) (ClaimResult, error)
	Read(context.Context, string, time.Time, time.Duration) (ReadModel, error)
}

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore(db *sql.DB) (*PostgresStore, error) {
	if db == nil {
		return nil, errors.New("telemetry: PostgreSQL is required")
	}
	return &PostgresStore{db: db}, nil
}

func (s *PostgresStore) Authorize(ctx context.Context, runtimeCredentialID string, observedAt time.Time) (AuthorizationResult, error) {
	if s == nil || s.db == nil || !validUUID(runtimeCredentialID) || observedAt.IsZero() {
		return AuthorizationResult{}, ErrInvalidEvent
	}
	var serviceTerm sql.NullString
	var quota, remaining sql.NullInt64
	var expires, firstConnected sql.NullTime
	var out AuthorizationResult
	err := s.db.QueryRowContext(ctx, `
SELECT service_term_id::text, tracked, allowed, reason, quota_bytes, used_bytes,
       reserved_bytes, remaining_bytes, expires_at, first_connected_at,
       accounting_complete
FROM pvnaive.direct_naive_accounting_authorize($1::uuid, $2)`, runtimeCredentialID, observedAt.UTC()).Scan(
		&serviceTerm, &out.Tracked, &out.Allowed, &out.Reason, &quota, &out.UsedBytes,
		&out.ReservedBytes, &remaining, &expires, &firstConnected, &out.AccountingComplete,
	)
	if err != nil {
		return AuthorizationResult{}, fmt.Errorf("telemetry: authorize: %w", err)
	}
	out.ServiceTermID = nullableString(serviceTerm)
	out.QuotaBytes = nullableInt64(quota)
	out.RemainingBytes = nullableInt64(remaining)
	out.ExpiresAt = nullableTime(expires)
	out.FirstConnectedAt = nullableTime(firstConnected)
	return out, nil
}

func (s *PostgresStore) Ingest(ctx context.Context, event Event) (IngestResult, error) {
	if s == nil || s.db == nil {
		return IngestResult{}, errors.New("telemetry: PostgreSQL is required")
	}
	if err := ValidateEvent(event); err != nil {
		return IngestResult{}, err
	}
	var serviceTerm sql.NullString
	var remaining sql.NullInt64
	var firstConnected sql.NullTime
	var out IngestResult
	err := s.db.QueryRowContext(ctx, `
SELECT service_term_id::text, tracked, accepted, duplicate, reason,
       upload_delta, download_delta, quota_depleted, remaining_bytes,
       first_connected_at, accounting_complete
FROM pvnaive.direct_naive_accounting_ingest(
    $1::uuid, $2, $3, $4::uuid, $5::uuid, $6, $7, $8, $9, $10, $11
)`, event.RuntimeCredentialID, event.Username, event.NodeID, event.BootID,
		event.SessionID, event.Sequence, event.ObservedAt.UTC(), event.AuthenticatedConnect,
		event.UploadBytes, event.DownloadBytes, event.Final).Scan(
		&serviceTerm, &out.Tracked, &out.Accepted, &out.Duplicate, &out.Reason,
		&out.UploadDelta, &out.DownloadDelta, &out.QuotaDepleted, &remaining,
		&firstConnected, &out.AccountingComplete,
	)
	if err != nil {
		return IngestResult{}, fmt.Errorf("telemetry: ingest: %w", err)
	}
	out.ServiceTermID = nullableString(serviceTerm)
	out.RemainingBytes = nullableInt64(remaining)
	out.FirstConnectedAt = nullableTime(firstConnected)
	return out, nil
}

func (s *PostgresStore) Claim(ctx context.Context, request ClaimRequest) (ClaimResult, error) {
	if s == nil || s.db == nil || !validClaimRequest(request) {
		return ClaimResult{}, ErrInvalidQuotaClaim
	}
	var serviceTerm sql.NullString
	var quota, remaining sql.NullInt64
	var expires sql.NullTime
	var out ClaimResult
	err := s.db.QueryRowContext(ctx, `
SELECT service_term_id::text, tracked, allowed, reason, granted_bytes,
       quota_bytes, used_bytes, reserved_total_bytes, remaining_bytes,
       expires_at, accounting_complete
FROM pvnaive.direct_naive_accounting_claim(
    $1::uuid, $2, $3::uuid, $4::uuid, $5, $6, $7, $8
)`, request.RuntimeCredentialID, request.NodeID, request.BootID, request.SessionID,
		request.Sequence, request.Direction, request.RequestedBytes, request.ObservedAt.UTC()).Scan(
		&serviceTerm, &out.Tracked, &out.Allowed, &out.Reason, &out.GrantedBytes,
		&quota, &out.UsedBytes, &out.ReservedBytes, &remaining, &expires,
		&out.AccountingComplete,
	)
	if err != nil {
		return ClaimResult{}, fmt.Errorf("telemetry: claim: %w", err)
	}
	out.ServiceTermID = nullableString(serviceTerm)
	out.QuotaBytes = nullableInt64(quota)
	out.RemainingBytes = nullableInt64(remaining)
	out.ExpiresAt = nullableTime(expires)
	return out, nil
}

func (s *PostgresStore) Read(ctx context.Context, serviceTermID string, observedAt time.Time, staleAfter time.Duration) (ReadModel, error) {
	if s == nil || s.db == nil || !validUUID(serviceTermID) || observedAt.IsZero() || staleAfter <= 0 {
		return ReadModel{}, ErrInvalidProjection
	}
	seconds := int64(staleAfter / time.Second)
	if seconds <= 0 {
		seconds = 1
	}
	var quota, remaining sql.NullInt64
	var firstConnected, lastOnline sql.NullTime
	var sessionCount int64
	var state string
	var out ReadModel
	err := s.db.QueryRowContext(ctx, `
SELECT service_term_id::text, upload_bytes, download_bytes, used_bytes,
       quota_bytes, remaining_bytes, quota_state, first_connected_at,
       last_online, online, session_count, accounting_complete
FROM pvnaive.direct_naive_accounting_read($1::uuid, $2, $3)`, serviceTermID, observedAt.UTC(), seconds).Scan(
		&out.ServiceTermID, &out.UploadBytes, &out.DownloadBytes, &out.UsedBytes,
		&quota, &remaining, &state, &firstConnected, &lastOnline, &out.Online,
		&sessionCount, &out.AccountingComplete,
	)
	if err != nil {
		return ReadModel{}, fmt.Errorf("telemetry: read accounting: %w", err)
	}
	if sessionCount < 0 || sessionCount > math.MaxInt {
		return ReadModel{}, ErrInvalidProjection
	}
	out.SessionCount = int(sessionCount)
	out.QuotaBytes = nullableInt64(quota)
	out.RemainingBytes = nullableInt64(remaining)
	out.FirstConnectedAt = nullableTime(firstConnected)
	out.LastOnline = nullableTime(lastOnline)
	out.QuotaState = QuotaState(state)
	return out, nil
}

func validClaimRequest(request ClaimRequest) bool {
	return validUUID(request.RuntimeCredentialID) && validDiagnostic(request.NodeID) &&
		validUUID(request.BootID) && validUUID(request.SessionID) && request.Sequence >= 2 &&
		(request.Direction == "upload" || request.Direction == "download") &&
		request.RequestedBytes > 0 && !request.ObservedAt.IsZero()
}

func nullableString(value sql.NullString) string {
	if !value.Valid {
		return ""
	}
	return strings.TrimSpace(value.String)
}

func nullableInt64(value sql.NullInt64) *int64 {
	if !value.Valid {
		return nil
	}
	v := value.Int64
	return &v
}

func nullableTime(value sql.NullTime) *time.Time {
	if !value.Valid {
		return nil
	}
	v := value.Time.UTC()
	return &v
}
