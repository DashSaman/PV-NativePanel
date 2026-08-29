package customer

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

func (s *PostgresStore) SetSubscriptionTokenRecoveryTx(
	ctx context.Context,
	tx *sql.Tx,
	tokenHash, ciphertext, nonce []byte,
	keyID string,
) error {
	if tx == nil {
		return errors.New("customer: transaction is required")
	}
	if len(tokenHash) != 32 || len(ciphertext) < 16 || len(nonce) != 12 || keyID == "" {
		return errors.New("customer: invalid subscription recovery envelope")
	}
	result, err := tx.ExecContext(ctx, `
UPDATE pvnaive.direct_subscription_tokens
SET token_ciphertext = $2,
    token_nonce = $3,
    token_encryption_key_id = $4
WHERE token_hash = $1
  AND status = 'active'
  AND revoked_at IS NULL`, tokenHash, ciphertext, nonce, keyID)
	if err != nil {
		return fmt.Errorf("customer: persist subscription recovery envelope: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("customer: inspect subscription recovery update: %w", err)
	}
	if rows != 1 {
		return errors.New("customer: active subscription recovery target not found")
	}
	return nil
}

func (s *PostgresStore) CurrentSubscriptionTokenTx(ctx context.Context, tx *sql.Tx, userID string) (EncryptedSubscriptionToken, error) {
	if tx == nil {
		return EncryptedSubscriptionToken{}, errors.New("customer: transaction is required")
	}
	var record EncryptedSubscriptionToken
	err := tx.QueryRowContext(ctx, `
SELECT user_id::text, token_ciphertext, token_nonce, token_encryption_key_id
FROM pvnaive.direct_subscription_tokens
WHERE user_id = $1::uuid
  AND status = 'active'
  AND revoked_at IS NULL
  AND token_ciphertext IS NOT NULL
  AND token_nonce IS NOT NULL
  AND token_encryption_key_id IS NOT NULL
  AND user_state = 'active'
  AND service_state IN ('pending','active')
  AND (expires_at IS NULL OR expires_at > clock_timestamp())
ORDER BY created_at DESC
LIMIT 1`, userID).Scan(&record.UserID, &record.Ciphertext, &record.Nonce, &record.EncryptionKeyID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return EncryptedSubscriptionToken{}, ErrSubscriptionNotRetrievable
		}
		return EncryptedSubscriptionToken{}, fmt.Errorf("customer: query current subscription: %w", err)
	}
	return record, nil
}

func (s *Service) persistSubscriptionRecovery(
	ctx context.Context,
	tx *sql.Tx,
	tokenHash, ciphertext, nonce []byte,
	keyID string,
) error {
	if len(ciphertext) == 0 && len(nonce) == 0 && keyID == "" {
		return nil
	}
	store, ok := s.store.(interface {
		SetSubscriptionTokenRecoveryTx(context.Context, *sql.Tx, []byte, []byte, []byte, string) error
	})
	if !ok {
		return errors.New("customer: subscription recovery store capability is unavailable")
	}
	return store.SetSubscriptionTokenRecoveryTx(ctx, tx, tokenHash, ciphertext, nonce, keyID)
}
