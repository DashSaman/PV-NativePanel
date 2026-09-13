package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// UpdateActorPasswordHash replaces the actor's login password hash inside the
// caller's transaction. The caller is responsible for verifying the current
// password before invoking this and for revoking/rotating sessions afterwards.
func (s *Store) UpdateActorPasswordHash(ctx context.Context, tx *sql.Tx, actorID, passwordHash string) error {
	if tx == nil {
		return errors.New("auth: UpdateActorPasswordHash requires a transaction")
	}
	if actorID == "" || passwordHash == "" {
		return errors.New("auth: actor id and password hash are required")
	}
	const q = `UPDATE pvnaive.actors SET password_hash = $1, updated_at = clock_timestamp() WHERE id = $2`
	res, err := tx.ExecContext(ctx, q, passwordHash, actorID)
	if err != nil {
		return fmt.Errorf("auth: update actor password hash: %w", err)
	}
	if n, rowsErr := res.RowsAffected(); rowsErr == nil && n == 0 {
		return errors.New("auth: actor not found")
	}
	return nil
}

// UpdateActorProfile updates the actor's login email and/or display name.
// Empty values keep the current field. Lower-cased email uniqueness is
// enforced by the actors_email_lower_uidx index; callers must translate a
// unique-violation error into a user-facing conflict response.
func (s *Store) UpdateActorProfile(ctx context.Context, tx *sql.Tx, actorID, email, displayName string) (string, string, error) {
	if tx == nil {
		return "", "", errors.New("auth: UpdateActorProfile requires a transaction")
	}
	if actorID == "" || (email == "" && displayName == "") {
		return "", "", errors.New("auth: actor id and at least one profile field are required")
	}
	const q = `
UPDATE pvnaive.actors
SET email = COALESCE(NULLIF($2, ''), email),
    display_name = COALESCE(NULLIF($3, ''), display_name),
    updated_at = clock_timestamp()
WHERE id = $1
RETURNING email, display_name`
	var outEmail, outName string
	if err := tx.QueryRowContext(ctx, q, actorID, email, displayName).Scan(&outEmail, &outName); err != nil {
		return "", "", fmt.Errorf("auth: update actor profile: %w", err)
	}
	return outEmail, outName, nil
}
