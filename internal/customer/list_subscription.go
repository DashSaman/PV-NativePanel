package customer

import (
	"context"
	"database/sql"
	"errors"

	"github.com/DashSaman/PV-NaivePanel/internal/subscription"
)

type customerListStore interface {
	ListCustomersTx(context.Context, *sql.Tx) ([]CustomerView, error)
}

type subscriptionRotationStore interface {
	SubscriptionTargetTx(context.Context, *sql.Tx, string) (SubscriptionTarget, error)
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

func (s *Service) RotateSubscription(ctx context.Context, tx *sql.Tx, userID string) (string, error) {
	if s == nil || s.store == nil || tx == nil || userID == "" {
		return "", errors.New("customer: subscription rotation requires service, transaction and user id")
	}
	store, ok := s.store.(subscriptionRotationStore)
	if !ok {
		return "", errors.New("customer: subscription rotation capability is unavailable")
	}
	target, err := store.SubscriptionTargetTx(ctx, tx, userID)
	if err != nil {
		return "", err
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
