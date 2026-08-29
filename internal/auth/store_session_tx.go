package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

func (s *Store) GetActorPasswordHash(ctx context.Context, tx *sql.Tx, actorID string) (string, error) {
	if tx == nil || actorID == "" {
		return "", errors.New("auth: bound transaction and actor ID are required")
	}
	var hash sql.NullString
	if err := tx.QueryRowContext(ctx, `SELECT password_hash FROM pvnaive.actors WHERE id=$1::uuid`, actorID).Scan(&hash); err != nil {
		return "", fmt.Errorf("auth: get actor password hash: %w", err)
	}
	if !hash.Valid || hash.String == "" {
		return "", errors.New("auth: actor has no password hash")
	}
	return hash.String, nil
}

func (s *Store) RevokeOtherActorSessions(ctx context.Context, tx *sql.Tx, actorID, currentSessionID string) (int64, error) {
	if tx == nil || actorID == "" || currentSessionID == "" {
		return 0, errors.New("auth: bound transaction, actor ID and current session ID are required")
	}
	var count int64
	if err := tx.QueryRowContext(ctx, `SELECT pvnaive.auth_revoke_other_actor_sessions($1::uuid,$2::uuid)`, actorID, currentSessionID).Scan(&count); err != nil {
		return 0, fmt.Errorf("auth: revoke other actor sessions: %w", err)
	}
	return count, nil
}

func (s *Store) RotateSessionTx(ctx context.Context, tx *sql.Tx, oldHash, newHash, newCSRFHash, newUserAgentHash []byte, newExpiresAt time.Time) (RotatedSession, error) {
	if tx == nil {
		return RotatedSession{}, errors.New("auth: bound transaction is required")
	}
	for label, value := range map[string][]byte{"old session token": oldHash, "new session token": newHash, "new CSRF token": newCSRFHash} {
		if err := requireSHA256(label, value); err != nil {
			return RotatedSession{}, err
		}
	}
	if len(newUserAgentHash) != 0 && len(newUserAgentHash) != 32 {
		return RotatedSession{}, errors.New("auth: user-agent hash must be 32 bytes")
	}
	var out RotatedSession
	var tenant sql.NullString
	err := tx.QueryRowContext(ctx, `SELECT session_id::text, actor_id::text, tenant_id::text, refresh_family_id::text, absolute_expires_at, reuse_detected FROM pvnaive.auth_rotate_session($1,$2,$3,$4,$5)`, oldHash, newHash, newCSRFHash, nullableBytes(newUserAgentHash), newExpiresAt).
		Scan(&out.SessionID, &out.ActorID, &tenant, &out.RefreshFamilyID, &out.AbsoluteExpiresAt, &out.ReuseDetected)
	if err != nil {
		return RotatedSession{}, fmt.Errorf("auth: rotate session in transaction: %w", err)
	}
	if tenant.Valid {
		v := tenant.String
		out.TenantID = &v
	}
	return out, nil
}
