package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// GetTOTPFactorPreAuth is intentionally narrow: login has no signed request
// context yet, so PostgreSQL exposes only the encrypted factor through the
// SECURITY DEFINER helper created by migration 0002.
func (s *Store) GetTOTPFactorPreAuth(ctx context.Context, actorID string) (TOTPFactorRecord, error) {
	if s == nil || s.db == nil || actorID == "" {
		return TOTPFactorRecord{}, errors.New("auth: initialized store and actor ID are required")
	}
	var out TOTPFactorRecord
	var step sql.NullInt64
	var confirmed sql.NullTime
	if err := s.db.QueryRowContext(ctx, `SELECT secret_ciphertext,secret_nonce,encryption_key_id,last_used_step,confirmed_at FROM pvnaive.auth_get_totp_factor($1::uuid)`, actorID).
		Scan(&out.Ciphertext, &out.Nonce, &out.KeyID, &step, &confirmed); err != nil {
		return TOTPFactorRecord{}, fmt.Errorf("auth: get pre-auth TOTP factor: %w", err)
	}
	if step.Valid {
		v := step.Int64
		out.LastUsedStep = &v
	}
	if confirmed.Valid {
		v := confirmed.Time
		out.ConfirmedAt = &v
	}
	return out, nil
}

func (s *Store) ConsumeTOTPStepPreAuth(ctx context.Context, actorID string, step int64) (bool, error) {
	if s == nil || s.db == nil || actorID == "" || step < 0 {
		return false, errors.New("auth: invalid pre-auth TOTP consume request")
	}
	var ok bool
	if err := s.db.QueryRowContext(ctx, `SELECT pvnaive.auth_consume_totp_step($1::uuid,$2)`, actorID, step).Scan(&ok); err != nil {
		return false, fmt.Errorf("auth: consume pre-auth TOTP step: %w", err)
	}
	return ok, nil
}

// PasswordChangedAt is queried only through authenticated RLS context where
// needed later; this helper exists to keep login's pre-auth database surface
// minimal.
var _ = time.Time{}
