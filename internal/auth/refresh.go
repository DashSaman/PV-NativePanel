package auth

import (
	"context"
	"errors"
	"fmt"
	"time"
)

type RefreshSessionMetadata struct {
	CSRFTokenHash     []byte
	AbsoluteExpiresAt time.Time
}

func (s *Store) LoadRefreshSessionMetadata(ctx context.Context, tokenHash []byte) (RefreshSessionMetadata, error) {
	if err := requireSHA256("session token", tokenHash); err != nil {
		return RefreshSessionMetadata{}, err
	}
	if s == nil || s.db == nil {
		return RefreshSessionMetadata{}, errors.New("auth: store is not initialized")
	}
	var out RefreshSessionMetadata
	if err := s.db.QueryRowContext(ctx, `SELECT csrf_token_hash, absolute_expires_at FROM pvnaive.auth_refresh_session_metadata($1)`, tokenHash).Scan(&out.CSRFTokenHash, &out.AbsoluteExpiresAt); err != nil {
		return RefreshSessionMetadata{}, fmt.Errorf("auth: load refresh session metadata: %w", err)
	}
	if len(out.CSRFTokenHash) != 32 || out.AbsoluteExpiresAt.IsZero() {
		return RefreshSessionMetadata{}, errors.New("auth: invalid refresh session metadata")
	}
	return out, nil
}
