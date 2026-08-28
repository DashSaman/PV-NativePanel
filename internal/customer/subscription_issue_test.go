package customer

import (
	"context"
	"database/sql"
	"encoding/base64"
	"testing"
	"time"

	"github.com/DashSaman/PV-NaivePanel/internal/runtimecred"
)

type subscriptionCustomerStore struct {
	fakeCustomerStore
	issued *CreateSubscriptionTokenRecord
}

func (f *subscriptionCustomerStore) CreateSubscriptionTokenTx(_ context.Context, _ *sql.Tx, record CreateSubscriptionTokenRecord) error {
	copy := record
	f.issued = &copy
	return nil
}

func TestCreateCustomerIssuesOpaqueSubscriptionTokenAndPersistsOnlyHash(t *testing.T) {
	now := time.Date(2026, 8, 29, 12, 0, 0, 0, time.UTC)
	store := &subscriptionCustomerStore{}
	mutation := &fakeCustomerRuntimeMutation{
		view:     runtimecred.CredentialView{ID: "runtime-sub-1", Username: "sub.customer", Status: runtimecred.CredentialActive},
		password: "generated-secret",
	}
	service := NewService(store, func(context.Context, *sql.Tx, string, string, runtimecred.CreateInput) (RuntimeMutation, error) {
		return mutation, nil
	}, func() time.Time { return now })
	quota := int64(20)

	result, err := service.CreateCustomer(context.Background(), nil, "owner-1", "idem-subscription-0001", CreateCustomerInput{
		Username: "sub.customer", GeneratePassword: true, QuotaGB: &quota,
		Validity: ValidityInput{Mode: ValidityOnFirstSuccessfulConnection, DurationDays: 30},
	})
	if err != nil {
		t.Fatalf("CreateCustomer() error = %v", err)
	}
	if result.SubscriptionToken == "" {
		t.Fatal("CreateCustomer() did not return one-time subscription token")
	}
	raw, err := base64.RawURLEncoding.DecodeString(result.SubscriptionToken)
	if err != nil || len(raw) != 32 {
		t.Fatalf("subscription token is not canonical 256-bit base64url: len=%d err=%v", len(raw), err)
	}
	if store.issued == nil {
		t.Fatal("subscription token hash was not persisted")
	}
	if store.issued.RawToken != "" {
		t.Fatal("raw subscription token was persisted")
	}
	if store.issued.TenantID != "tenant-direct" || store.issued.UserID != "user-1" || store.issued.ServiceTermID != "term-1" || store.issued.RuntimeCredentialID != "runtime-sub-1" {
		t.Fatalf("subscription projection scope = %#v", store.issued)
	}
	if len(store.issued.TokenHash) != 32 || len(store.issued.TokenPrefix) < 6 {
		t.Fatalf("subscription hash/prefix invalid: %#v", store.issued)
	}
	if store.issued.ExpiresAt != nil {
		t.Fatalf("first-use token should not receive a fixed expiry before activation: %v", store.issued.ExpiresAt)
	}
}
