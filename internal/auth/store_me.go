package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// UpdateActorPasswordHash rotates the actor's login password hash through the
// SECURITY DEFINER function installed by migration 0022 (0002 revoked direct
// DML on pvnaive.actors from pvnaive_app). The caller must verify the current
// password first and revoke/rotate sessions afterwards.
func (s *Store) UpdateActorPasswordHash(ctx context.Context, tx *sql.Tx, actorID, passwordHash string) error {
	if tx == nil {
		return errors.New("auth: UpdateActorPasswordHash requires a transaction")
	}
	if actorID == "" || passwordHash == "" {
		return errors.New("auth: actor id and password hash are required")
	}
	const q = `SELECT pvnaive.auth_update_actor_password($1, $2)`
	if _, err := tx.ExecContext(ctx, q, actorID, passwordHash); err != nil {
		return fmt.Errorf("auth: update actor password hash: %w", err)
	}
	return nil
}

// UpdateActorProfile updates the actor's login email and/or display name via
// the SECURITY DEFINER function from migration 0022. Empty values keep the
// current field. Lower-cased email uniqueness is enforced by the
// actors_email_lower_uidx index; callers must translate a unique-violation
// error into a user-facing conflict response.
func (s *Store) UpdateActorProfile(ctx context.Context, tx *sql.Tx, actorID, email, displayName string) (string, string, error) {
	if tx == nil {
		return "", "", errors.New("auth: UpdateActorProfile requires a transaction")
	}
	if actorID == "" || (email == "" && displayName == "") {
		return "", "", errors.New("auth: actor id and at least one profile field are required")
	}
	const q = `SELECT out_email, out_display_name FROM pvnaive.auth_update_actor_profile($1, $2, $3)`
	var outEmail, outName string
	if err := tx.QueryRowContext(ctx, q, actorID, email, displayName).Scan(&outEmail, &outName); err != nil {
		return "", "", fmt.Errorf("auth: update actor profile: %w", err)
	}
	return outEmail, outName, nil
}
