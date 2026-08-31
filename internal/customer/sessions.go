package customer

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"
)

const maxActiveSessionStaleAfter = time.Hour

type CustomerSession struct {
	RuntimeCredentialID string    `json:"runtime_credential_id"`
	NodeID              string    `json:"node_id"`
	BootID              string    `json:"boot_id"`
	SessionID           string    `json:"session_id"`
	ServiceTermID       string    `json:"service_term_id"`
	ClientIP            string    `json:"client_ip"`
	ConnectedAt         time.Time `json:"connected_at"`
	LastActivityAt      time.Time `json:"last_activity_at"`
	DurationSeconds     int64     `json:"duration_seconds"`
	UploadBytes         int64     `json:"upload_bytes"`
	DownloadBytes       int64     `json:"download_bytes"`
}

type customerSessionStore interface {
	ListActiveSessionsTx(context.Context, *sql.Tx, string, time.Time, time.Duration) ([]CustomerSession, error)
}

func (s *Service) ListActiveSessions(ctx context.Context, tx *sql.Tx, userID string, observedAt time.Time, staleAfter time.Duration) ([]CustomerSession, error) {
	if s == nil || s.store == nil || tx == nil || strings.TrimSpace(userID) == "" || observedAt.IsZero() || staleAfter <= 0 || staleAfter > maxActiveSessionStaleAfter {
		return nil, errors.New("customer: active session capability is unavailable")
	}
	store, ok := s.store.(customerSessionStore)
	if !ok {
		return nil, errors.New("customer: active session store is unavailable")
	}
	return store.ListActiveSessionsTx(ctx, tx, userID, observedAt.UTC(), staleAfter)
}
