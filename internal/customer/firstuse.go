package customer

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"
)

type firstUseStore interface {
	ActivateFirstUseTx(context.Context, *sql.Tx, string, time.Time) (bool, error)
}

func (s *Service) ActivateFirstUse(ctx context.Context, tx *sql.Tx, runtimeCredentialID string, observedAt time.Time) (bool, error) {
	if s == nil || s.store == nil {
		return false, errors.New("customer: service dependencies are required")
	}
	store, ok := s.store.(firstUseStore)
	if !ok {
		return false, errors.New("customer: first-use activation capability is unavailable")
	}
	runtimeCredentialID = strings.TrimSpace(runtimeCredentialID)
	if runtimeCredentialID == "" {
		return false, errors.New("customer: runtime credential id is required")
	}
	if observedAt.IsZero() {
		return false, errors.New("customer: first-use observed time is required")
	}
	return store.ActivateFirstUseTx(ctx, tx, runtimeCredentialID, observedAt.UTC())
}
