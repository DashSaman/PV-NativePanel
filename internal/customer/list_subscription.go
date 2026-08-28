package customer

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"strings"

	"github.com/DashSaman/PV-NaivePanel/internal/subscription"
)

var ErrSubscriptionRotationReplay = errors.New("customer: subscription rotation idempotent replay")

type customerListStore interface {
	ListCustomersTx(context.Context, *sql.Tx) ([]CustomerView, error)
}

type subscriptionRotationStore interface {
	SubscriptionTargetTx(context.Context, *sql.Tx, string) (SubscriptionTarget, error)
	ClaimSubscriptionRotationTx(context.Context, *sql.Tx, SubscriptionTarget, string, string, []byte) (bool, error)
	RevokeSubscriptionTokensTx(context.Context, *sql.Tx, SubscriptionTarget) error
	CreateSubscriptionTokenTx(context.Context, *sql.Tx, CreateSubscriptionTokenRecord) error
}

func (s *Service) ListCustomers(ctx context.Context, tx *sql.Tx) ([]CustomerView, error) {
	if s == nil || s.store == nil {
		return nil, errors.New("customer: service dependencies are required")
	}
	store, ok := s.store.(customerListStore)
	if !ok {
		return nil, errors.New("customer: list capability is unavailable")
	}
	return store.ListCustomersTx(ctx, tx)
}

func (s *Service) RotateSubscription(ctx context.Context, tx *sql.Tx, actorID, idempotencyKey, userID string) (string, error) {
	if s == nil || s.store == nil || tx == nil || strings.TrimSpace(userID) == "" || strings.TrimSpace(actorID) == "" {
		return "", errors.New("customer: subscription rotation requires service, transaction, actor and user id")
	}
	if len(idempotencyKey) < 8 || len(idempotencyKey) > 160 || strings.TrimSpace(idempotencyKey) != idempotencyKey {
		return "", errors.New("customer: invalid subscription rotation idempotency key")
	}
	store, ok := s.store.(subscriptionRotationStore)
	if !ok {
		return "", errors.New("customer: subscription rotation capability is unavailable")
	}
	target, err := store.SubscriptionTargetTx(ctx, tx, userID)
	if err != nil {
		return "", err
	}
	requestHash := sha256.Sum256([]byte("customer.subscription.rotate\n" + target.UserID))
	claimed, err := store.ClaimSubscriptionRotationTx(ctx, tx, target, actorID, idempotencyKey, requestHash[:])
	if err != nil {
		return "", err
	}
	if !claimed {
		return "", ErrSubscriptionRotationReplay
	}

	raw, hash, err := subscription.GenerateToken()
	if err != nil {
		return "", err
	}
	prefix := raw
	if len(prefix) > 10 {
		prefix = prefix[:10]
	}
	if err := store.RevokeSubscriptionTokensTx(ctx, tx, target); err != nil {
		return "", err
	}
	if err := store.CreateSubscriptionTokenTx(ctx, tx, CreateSubscriptionTokenRecord{
		TenantID:            target.TenantID,
		UserID:              target.UserID,
		ServiceTermID:       target.ServiceTermID,
		RuntimeCredentialID: target.RuntimeCredentialID,
		TokenHash:           append([]byte(nil), hash[:]...),
		TokenPrefix:         prefix,
		ExpiresAt:           target.ExpiresAt,
	}); err != nil {
		return "", err
	}
	return raw, nil
}
