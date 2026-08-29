package customer

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DashSaman/PV-NaivePanel/internal/runtimecred"
)

type currentSubscriptionStore struct {
	fakeCustomerStore
	record EncryptedSubscriptionToken
	reads  int
}

func (s *currentSubscriptionStore) CurrentSubscriptionTokenTx(_ context.Context, _ *sql.Tx, userID string) (EncryptedSubscriptionToken, error) {
	s.reads++
	if userID != "user-1" {
		testing.AllocsPerRun(1, func() {})
	}
	return s.record, nil
}

func TestCurrentSubscriptionDecryptsExistingTokenWithoutMutation(t *testing.T) {
	key := []byte("0123456789abcdef0123456789abcdef")
	raw := "current-subscription-token"
	ciphertext, nonce, err := runtimecred.EncryptSecret(key, []byte(raw))
	if err != nil {
		t.Fatalf("EncryptSecret() error = %v", err)
	}
	store := &currentSubscriptionStore{record: EncryptedSubscriptionToken{
		UserID: "user-1", Ciphertext: ciphertext, Nonce: nonce, EncryptionKeyID: "runtime-v1",
	}}
	service := NewServiceWithTokenRecovery(store, nil, func() time.Time { return time.Unix(0, 0) }, key, "runtime-v1")

	got, err := service.CurrentSubscription(context.Background(), nil, "user-1")
	if err != nil {
		t.Fatalf("CurrentSubscription() error = %v", err)
	}
	if got != raw {
		t.Fatalf("CurrentSubscription() = %q, want %q", got, raw)
	}
	if store.reads != 1 {
		t.Fatalf("current subscription store reads = %d, want 1", store.reads)
	}
}

func TestCurrentSubscriptionRejectsWrongEncryptionKeyID(t *testing.T) {
	key := []byte("0123456789abcdef0123456789abcdef")
	ciphertext, nonce, err := runtimecred.EncryptSecret(key, []byte("token"))
	if err != nil {
		t.Fatalf("EncryptSecret() error = %v", err)
	}
	store := &currentSubscriptionStore{record: EncryptedSubscriptionToken{
		UserID: "user-1", Ciphertext: ciphertext, Nonce: nonce, EncryptionKeyID: "runtime-v0",
	}}
	service := NewServiceWithTokenRecovery(store, nil, time.Now, key, "runtime-v1")

	if _, err := service.CurrentSubscription(context.Background(), nil, "user-1"); err == nil {
		t.Fatal("CurrentSubscription() succeeded with mismatched key id")
	}
}
