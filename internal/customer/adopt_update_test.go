package customer

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DashSaman/PV-NaivePanel/internal/runtimecred"
)

type adoptUpdateCustomerStore struct {
	subscriptionCustomerStore
	adoptable runtimecred.CredentialView
	updated   *UpdateServiceTermRecord
}

func (f *adoptUpdateCustomerStore) AdoptableRuntimeCredentialTx(context.Context, *sql.Tx, string) (runtimecred.CredentialView, error) {
	return f.adoptable, nil
}

func (f *adoptUpdateCustomerStore) UpdateCurrentServiceTermTx(_ context.Context, _ *sql.Tx, userID string, record UpdateServiceTermRecord) (ServiceTerm, error) {
	copy := record
	f.updated = &copy
	return ServiceTerm{
		ID: "term-existing", TenantID: "tenant-direct", UserID: userID,
		QuotaBytes: record.QuotaBytes, DurationSeconds: record.DurationSeconds,
		StartPolicy: record.StartPolicy, PurchasedAt: record.EffectiveAt,
		StartsAt: record.StartsAt, ExpiresAt: record.ExpiresAt, State: record.State, Revision: 2,
	}, nil
}

func TestAdoptRuntimeCredentialKeepsExistingCredentialAndAddsBusinessService(t *testing.T) {
	now := time.Date(2026, 8, 29, 13, 0, 0, 0, time.UTC)
	store := &adoptUpdateCustomerStore{
		adoptable: runtimecred.CredentialView{
			ID: "runtime-existing-1", Username: "amirreza", Status: runtimecred.CredentialActive,
			Origin: runtimecred.CredentialPanel, Revision: 7,
		},
	}
	service := NewService(store, func(context.Context, *sql.Tx, string, string, runtimecred.CreateInput) (RuntimeMutation, error) {
		t.Fatal("adoption must not create or rotate a runtime credential")
		return nil, nil
	}, func() time.Time { return now })
	quota := int64(80)

	result, err := service.AdoptRuntimeCredential(context.Background(), nil, "owner-1", AdoptRuntimeInput{
		RuntimeCredentialID: "runtime-existing-1",
		QuotaGB:             &quota,
		Validity:            ValidityInput{Mode: ValidityOnCreation, DurationDays: 45},
	})
	if err != nil {
		t.Fatalf("AdoptRuntimeCredential() error = %v", err)
	}
	if result.RuntimeCredential.ID != "runtime-existing-1" || result.RuntimeCredential.Username != "amirreza" {
		t.Fatalf("adoption changed runtime identity: %#v", result.RuntimeCredential)
	}
	if store.boundID != "runtime-existing-1" {
		t.Fatalf("bound runtime credential = %q", store.boundID)
	}
	if store.createdTerm.QuotaBytes == nil || *store.createdTerm.QuotaBytes != 80*BytesPerCustomerGB {
		t.Fatalf("adopted quota = %v", store.createdTerm.QuotaBytes)
	}
	if store.createdTerm.ExpiresAt == nil || !store.createdTerm.ExpiresAt.Equal(now.Add(45*24*time.Hour)) {
		t.Fatalf("adopted expiry = %v", store.createdTerm.ExpiresAt)
	}
	if result.SubscriptionToken == "" || store.issued == nil {
		t.Fatal("adoption must issue a one-time subscription token")
	}
}

func TestUpdateCustomerServiceChangesQuotaAndExpiryWithoutRuntimeMutation(t *testing.T) {
	now := time.Date(2026, 8, 29, 14, 0, 0, 0, time.UTC)
	store := &adoptUpdateCustomerStore{}
	service := NewService(store, func(context.Context, *sql.Tx, string, string, runtimecred.CreateInput) (RuntimeMutation, error) {
		t.Fatal("service-term edit must not mutate runtime credentials")
		return nil, nil
	}, func() time.Time { return now })
	quota := int64(120)
	expires := now.Add(20 * 24 * time.Hour)

	term, err := service.UpdateCustomerService(context.Background(), nil, "user-existing", UpdateServiceInput{
		QuotaGB:  &quota,
		Validity: ValidityInput{Mode: ValidityFixedExpiry, ExpiresAt: &expires},
	})
	if err != nil {
		t.Fatalf("UpdateCustomerService() error = %v", err)
	}
	if store.updated == nil {
		t.Fatal("service term was not updated")
	}
	if store.updated.QuotaBytes == nil || *store.updated.QuotaBytes != 120*BytesPerCustomerGB {
		t.Fatalf("updated quota bytes = %v", store.updated.QuotaBytes)
	}
	if store.updated.ExpiresAt == nil || !store.updated.ExpiresAt.Equal(expires) {
		t.Fatalf("updated expiry = %v", store.updated.ExpiresAt)
	}
	if store.updated.StartPolicy != StartFixedTimestamp || store.updated.State != TermActive {
		t.Fatalf("updated policy/state = %s/%s", store.updated.StartPolicy, store.updated.State)
	}
	if term.Revision != 2 {
		t.Fatalf("updated term revision = %d", term.Revision)
	}
}
