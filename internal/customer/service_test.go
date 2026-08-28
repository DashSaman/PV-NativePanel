package customer

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DashSaman/PV-NaivePanel/internal/runtimecred"
)

type fakeCustomerStore struct {
	user         User
	term         ServiceTerm
	boundID      string
	bindingError error
	createdTerm  CreateServiceTermRecord
}

func (f *fakeCustomerStore) DirectTenantID(context.Context, *sql.Tx) (string, error) {
	return "tenant-direct", nil
}

func (f *fakeCustomerStore) CreateUserTx(_ context.Context, _ *sql.Tx, record CreateUserRecord) (User, error) {
	f.user = User{ID: "user-1", TenantID: record.TenantID, Username: record.Username, DisplayName: record.DisplayName, Status: UserActive, Revision: 1}
	return f.user, nil
}

func (f *fakeCustomerStore) CreateServiceTermTx(_ context.Context, _ *sql.Tx, record CreateServiceTermRecord) (ServiceTerm, error) {
	f.createdTerm = record
	f.term = ServiceTerm{
		ID: "term-1", TenantID: record.TenantID, UserID: record.UserID,
		QuotaBytes: record.QuotaBytes, DurationSeconds: record.DurationSeconds,
		StartPolicy: record.StartPolicy, PurchasedAt: record.PurchasedAt,
		StartsAt: record.StartsAt, ExpiresAt: record.ExpiresAt, State: record.State, Revision: 1,
	}
	return f.term, nil
}

func (f *fakeCustomerStore) BindRuntimeCredentialTx(_ context.Context, _ *sql.Tx, tenantID, userID, termID, runtimeCredentialID string) error {
	if f.bindingError != nil {
		return f.bindingError
	}
	f.boundID = runtimeCredentialID
	return nil
}

type fakeCustomerRuntimeMutation struct {
	view      runtimecred.CredentialView
	password  string
	committed bool
	aborted   bool
}

func (f *fakeCustomerRuntimeMutation) Credential() runtimecred.CredentialView { return f.view }
func (f *fakeCustomerRuntimeMutation) RuntimeRevisionID() string              { return "runtime-rev-1" }
func (f *fakeCustomerRuntimeMutation) CommitAndFinalize(context.Context, *sql.Tx) error {
	if f.aborted {
		return errors.New("mutation aborted")
	}
	f.committed = true
	return nil
}
func (f *fakeCustomerRuntimeMutation) AbortAndRollback(context.Context, *sql.Tx) error {
	f.aborted = true
	f.password = ""
	return nil
}
func (f *fakeCustomerRuntimeMutation) TakeGeneratedPassword() string {
	if !f.committed {
		return ""
	}
	password := f.password
	f.password = ""
	return password
}

func TestCreateCustomerCreatesBusinessRowsAndBindsStableRuntimeCredential(t *testing.T) {
	now := time.Date(2026, 8, 29, 12, 0, 0, 0, time.UTC)
	store := &fakeCustomerStore{}
	runtimeMutation := &fakeCustomerRuntimeMutation{
		view:     runtimecred.CredentialView{ID: "runtime-uuid-1", Username: "customer1", Status: runtimecred.CredentialActive},
		password: "generated-secret",
	}
	var runtimeInput runtimecred.CreateInput
	service := NewService(store, func(_ context.Context, _ *sql.Tx, actorID, idempotencyKey string, input runtimecred.CreateInput) (RuntimeMutation, error) {
		if actorID != "owner-1" || idempotencyKey != "idem-customer-0001:runtime" {
			t.Fatalf("runtime mutation identity = %q %q", actorID, idempotencyKey)
		}
		runtimeInput = input
		return runtimeMutation, nil
	}, func() time.Time { return now })
	quota := int64(50)

	result, err := service.CreateCustomer(context.Background(), nil, "owner-1", "idem-customer-0001", CreateCustomerInput{
		Username: "customer1", GeneratePassword: true, QuotaGB: &quota,
		Validity: ValidityInput{Mode: ValidityOnFirstSuccessfulConnection, DurationDays: 30},
	})
	if err != nil {
		t.Fatalf("CreateCustomer() error = %v", err)
	}
	if runtimeInput.Username != "customer1" || !runtimeInput.GeneratePassword {
		t.Fatalf("runtime input = %#v", runtimeInput)
	}
	if store.boundID != "runtime-uuid-1" {
		t.Fatalf("bound runtime ID = %q, want stable runtime UUID", store.boundID)
	}
	if store.createdTerm.QuotaBytes == nil || *store.createdTerm.QuotaBytes != 50*BytesPerCustomerGB {
		t.Fatalf("term quota bytes = %v", store.createdTerm.QuotaBytes)
	}
	if store.createdTerm.State != TermPending || store.createdTerm.StartsAt != nil || store.createdTerm.ExpiresAt != nil {
		t.Fatalf("first-use term = %#v", store.createdTerm)
	}
	if !runtimeMutation.committed || runtimeMutation.aborted {
		t.Fatalf("runtime mutation committed=%v aborted=%v", runtimeMutation.committed, runtimeMutation.aborted)
	}
	if result.GeneratedPassword != "generated-secret" || result.RuntimeCredential.ID != "runtime-uuid-1" {
		t.Fatalf("delivery result = %#v", result)
	}
}

func TestCreateCustomerRollsRuntimeBackWhenBusinessBindingFails(t *testing.T) {
	store := &fakeCustomerStore{bindingError: errors.New("binding refused")}
	runtimeMutation := &fakeCustomerRuntimeMutation{
		view:     runtimecred.CredentialView{ID: "runtime-uuid-2", Username: "customer2", Status: runtimecred.CredentialActive},
		password: "secret-that-must-disappear",
	}
	service := NewService(store, func(context.Context, *sql.Tx, string, string, runtimecred.CreateInput) (RuntimeMutation, error) {
		return runtimeMutation, nil
	}, func() time.Time { return time.Date(2026, 8, 29, 12, 0, 0, 0, time.UTC) })

	_, err := service.CreateCustomer(context.Background(), nil, "owner-1", "idem-customer-0002", CreateCustomerInput{
		Username: "customer2", GeneratePassword: true,
		Validity: ValidityInput{Mode: ValidityOnCreation, DurationDays: 30},
	})
	if err == nil {
		t.Fatal("CreateCustomer() succeeded despite binding failure")
	}
	if !runtimeMutation.aborted || runtimeMutation.committed {
		t.Fatalf("runtime mutation committed=%v aborted=%v", runtimeMutation.committed, runtimeMutation.aborted)
	}
	if got := runtimeMutation.TakeGeneratedPassword(); got != "" {
		t.Fatalf("aborted customer leaked password %q", got)
	}
}

func TestCreateCustomerFixedExpiryKeepsExactExpiry(t *testing.T) {
	now := time.Date(2026, 8, 29, 12, 0, 0, 0, time.UTC)
	expiry := time.Date(2026, 10, 15, 23, 59, 0, 0, time.FixedZone("IRST", 3*3600+30*60))
	store := &fakeCustomerStore{}
	runtimeMutation := &fakeCustomerRuntimeMutation{view: runtimecred.CredentialView{ID: "runtime-uuid-3", Username: "customer3"}}
	service := NewService(store, func(context.Context, *sql.Tx, string, string, runtimecred.CreateInput) (RuntimeMutation, error) {
		return runtimeMutation, nil
	}, func() time.Time { return now })

	_, err := service.CreateCustomer(context.Background(), nil, "owner-1", "idem-customer-0003", CreateCustomerInput{
		Username: "customer3", Password: "custom password 123", GeneratePassword: false,
		Validity: ValidityInput{Mode: ValidityFixedExpiry, ExpiresAt: &expiry},
	})
	if err != nil {
		t.Fatalf("CreateCustomer() error = %v", err)
	}
	if store.createdTerm.ExpiresAt == nil || !store.createdTerm.ExpiresAt.Equal(expiry) {
		t.Fatalf("stored fixed expiry = %v, want exact instant %v", store.createdTerm.ExpiresAt, expiry)
	}
	if store.createdTerm.StartPolicy != StartFixedTimestamp || store.createdTerm.State != TermActive {
		t.Fatalf("stored fixed term policy/state = %s/%s", store.createdTerm.StartPolicy, store.createdTerm.State)
	}
}
