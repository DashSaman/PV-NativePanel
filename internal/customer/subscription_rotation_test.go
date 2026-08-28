package customer

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"
)

type rotationStore struct {
	fakeCustomerStore
	claim       bool
	claimErr    error
	claimActor  string
	claimKey    string
	revoked     int
	created     int
	lastCreated CreateSubscriptionTokenRecord
	target      SubscriptionTarget
}

func (s *rotationStore) SubscriptionTargetTx(context.Context, *sql.Tx, string) (SubscriptionTarget, error) {
	return s.target, nil
}
func (s *rotationStore) ClaimSubscriptionRotationTx(_ context.Context, _ *sql.Tx, target SubscriptionTarget, actorID, key string, _ []byte) (bool, error) {
	s.claimActor, s.claimKey = actorID, key
	return s.claim, s.claimErr
}
func (s *rotationStore) RevokeSubscriptionTokensTx(context.Context, *sql.Tx, SubscriptionTarget) error {
	s.revoked++
	return nil
}
func (s *rotationStore) CreateSubscriptionTokenTx(_ context.Context, _ *sql.Tx, record CreateSubscriptionTokenRecord) error {
	s.created++
	s.lastCreated = record
	return nil
}

func TestRotateSubscriptionClaimsIdempotencyBeforeRevokingOldToken(t *testing.T) {
	store := &rotationStore{claim: true, target: SubscriptionTarget{
		TenantID: "tenant-1", UserID: "user-1", ServiceTermID: "term-1", RuntimeCredentialID: "runtime-1",
	}}
	service := NewService(store, nil, time.Now)
	raw, err := service.RotateSubscription(context.Background(), &sql.Tx{}, "owner-1", "customer-rotate-0001", "user-1")
	if err != nil {
		t.Fatalf("RotateSubscription() error = %v", err)
	}
	if raw == "" || store.claimActor != "owner-1" || store.claimKey != "customer-rotate-0001" {
		t.Fatalf("raw=%q actor=%q key=%q", raw, store.claimActor, store.claimKey)
	}
	if store.revoked != 1 || store.created != 1 || len(store.lastCreated.TokenHash) != 32 {
		t.Fatalf("revoked=%d created=%d record=%#v", store.revoked, store.created, store.lastCreated)
	}
}

func TestRotateSubscriptionReplayNeverRevokesAgain(t *testing.T) {
	store := &rotationStore{claim: false, target: SubscriptionTarget{
		TenantID: "tenant-1", UserID: "user-1", ServiceTermID: "term-1", RuntimeCredentialID: "runtime-1",
	}}
	service := NewService(store, nil, time.Now)
	_, err := service.RotateSubscription(context.Background(), &sql.Tx{}, "owner-1", "customer-rotate-0001", "user-1")
	if !errors.Is(err, ErrSubscriptionRotationReplay) {
		t.Fatalf("RotateSubscription() error = %v, want replay", err)
	}
	if store.revoked != 0 || store.created != 0 {
		t.Fatalf("replay mutated tokens: revoked=%d created=%d", store.revoked, store.created)
	}
}
