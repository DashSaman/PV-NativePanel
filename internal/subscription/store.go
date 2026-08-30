package subscription

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
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
	var quota, duration, baselineUpload, baselineDownload sql.NullInt64
	var startPolicy sql.NullString
	var startsAt, firstConnectedAt, expiresAt sql.NullTime
	var baselineState, baselineSource string
	var baselineCutoff time.Time
	err := s.db.QueryRowContext(ctx, `
SELECT
    service_term_id::text,
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
    expires_at,
    accounting_baseline_state,
    accounting_baseline_source,
    accounting_baseline_cutoff_at,
    accounting_baseline_upload_bytes,
    accounting_baseline_download_bytes
FROM pvnaive.resolve_direct_subscription_account_profile($1)`, hash[:]).Scan(
		&record.ServiceTermID,
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
		&baselineState,
		&baselineSource,
		&baselineCutoff,
		&baselineUpload,
		&baselineDownload,
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
	record.AccountingBaseline = AccountingBaseline{
		State: baselineState, Source: baselineSource, CutoffAt: baselineCutoff.UTC(),
		UploadBytes: nullableInt64Value(baselineUpload), DownloadBytes: nullableInt64Value(baselineDownload),
	}
	return record, nil
}

func nullableInt64Value(value sql.NullInt64) *int64 {
	if !value.Valid {
		return nil
	}
	v := value.Int64
	return &v
}
