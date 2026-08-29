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
	var quota, duration sql.NullInt64
	var startPolicy sql.NullString
	var startsAt, firstConnectedAt, expiresAt sql.NullTime
	err := s.db.QueryRowContext(ctx, `
SELECT
    runtime_credential_id::text,
    runtime_username,
    user_state,
    service_state,
    secret_ciphertext,
    secret_nonce,
    encryption_key_id,
    quota_bytes,
    duration_seconds,
    start_policy,
    starts_at,
    first_connected_at,
    expires_at
FROM pvnaive.resolve_direct_subscription_profile($1)`, hash[:]).Scan(
		&record.RuntimeCredentialID,
		&record.Username,
		&record.UserState,
		&record.TermState,
		&record.SecretCiphertext,
		&record.SecretNonce,
		&record.EncryptionKeyID,
		&quota,
		&duration,
		&startPolicy,
		&startsAt,
		&firstConnectedAt,
		&expiresAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Record{}, ErrUnavailable
		}
		return Record{}, fmt.Errorf("subscription: resolve token: %w", err)
	}
	if quota.Valid {
		value := quota.Int64
		record.QuotaBytes = &value
	}
	if duration.Valid {
		record.DurationSeconds = duration.Int64
	}
	if startPolicy.Valid {
		record.StartPolicy = startPolicy.String
	}
	if startsAt.Valid {
		value := startsAt.Time
		record.StartsAt = &value
	}
	if firstConnectedAt.Valid {
		value := firstConnectedAt.Time
		record.FirstConnectedAt = &value
	}
	if expiresAt.Valid {
		value := expiresAt.Time
		record.ExpiresAt = &value
	}
	return record, nil
}
