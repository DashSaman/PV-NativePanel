package subscription

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore(db *sql.DB) *PostgresStore {
	return &PostgresStore{db: db}
}

func (s *PostgresStore) ResolveToken(ctx context.Context, hash [32]byte) (Record, error) {
	if s == nil || s.db == nil {
		return Record{}, errors.New("subscription: database is required")
	}

	var record Record
	var expiresAt sql.NullTime
	err := s.db.QueryRowContext(ctx, `
SELECT
    runtime_credential_id::text,
    runtime_username,
    user_state,
    service_state,
    secret_ciphertext,
    secret_nonce,
    encryption_key_id,
    expires_at
FROM pvnaive.resolve_direct_subscription_token($1)`, hash[:]).Scan(
		&record.RuntimeCredentialID,
		&record.Username,
		&record.UserState,
		&record.TermState,
		&record.SecretCiphertext,
		&record.SecretNonce,
		&record.EncryptionKeyID,
		&expiresAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Record{}, ErrUnavailable
		}
		return Record{}, fmt.Errorf("subscription: resolve token: %w", err)
	}
	if expiresAt.Valid {
		expires := expiresAt.Time
		record.ExpiresAt = &expires
	}
	return record, nil
}
