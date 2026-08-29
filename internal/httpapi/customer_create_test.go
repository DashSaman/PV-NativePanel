package httpapi

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DashSaman/PV-NaivePanel/internal/auth"
	"github.com/DashSaman/PV-NaivePanel/internal/customer"
	"github.com/DashSaman/PV-NaivePanel/internal/runtimecred"
)

type apiCustomerStore struct{}

func (apiCustomerStore) DirectTenantID(context.Context, *sql.Tx) (string, error) {
	return "tenant-direct", nil
}
func (apiCustomerStore) CreateUserTx(_ context.Context, _ *sql.Tx, record customer.CreateUserRecord) (customer.User, error) {
	return customer.User{ID: "user-1", TenantID: record.TenantID, Username: record.Username, DisplayName: record.DisplayName, Status: customer.UserActive, Revision: 1}, nil
}
func (apiCustomerStore) CreateServiceTermTx(_ context.Context, _ *sql.Tx, record customer.CreateServiceTermRecord) (customer.ServiceTerm, error) {
	return customer.ServiceTerm{
		ID: "term-1", TenantID: record.TenantID, UserID: record.UserID,
		QuotaBytes: record.QuotaBytes, DurationSeconds: record.DurationSeconds,
		StartPolicy: record.StartPolicy, PurchasedAt: record.PurchasedAt,
		StartsAt: record.StartsAt, ExpiresAt: record.ExpiresAt, State: record.State, Revision: 1,
	}, nil
}
func (apiCustomerStore) BindRuntimeCredentialTx(context.Context, *sql.Tx, string, string, string, string) error {
	return nil
}
func (apiCustomerStore) CreateSubscriptionTokenTx(context.Context, *sql.Tx, customer.CreateSubscriptionTokenRecord) error {
	return nil
}

type apiRuntimeMutation struct {
	committed bool
	password  string
}

func (m *apiRuntimeMutation) Credential() runtimecred.CredentialView {
	return runtimecred.CredentialView{ID: "runtime-1", Username: "customer1", Status: runtimecred.CredentialActive}
}
func (m *apiRuntimeMutation) RuntimeRevisionID() string { return "runtime-revision-1" }
func (m *apiRuntimeMutation) CommitAndFinalize(context.Context, *sql.Tx) error {
	m.committed = true
	return nil
}
func (m *apiRuntimeMutation) AbortAndRollback(context.Context, *sql.Tx) error { return nil }
func (m *apiRuntimeMutation) TakeGeneratedPassword() string {
	if !m.committed {
		return ""
	}
	value := m.password
	m.password = ""
	return value
}

func TestCreateCustomerEndpointReturnsOneTimeDeliveryMaterial(t *testing.T) {
	now := time.Date(2026, 8, 29, 12, 0, 0, 0, time.UTC)
	mutation := &apiRuntimeMutation{password: "generated-secret"}
	service := customer.NewService(apiCustomerStore{}, func(context.Context, *sql.Tx, string, string, runtimecred.CreateInput) (customer.RuntimeMutation, error) {
		return mutation, nil
	}, func() time.Time { return now })

	body := `{"username":"customer1","generate_password":true,"quota_gb":50,"validity":{"mode":"on_first_successful_connection","duration_days":30}}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/users", strings.NewReader(body))
	req.Header.Set("Idempotency-Key", "customer-create-0001")
	req.Header.Set("Content-Type", "application/json")
	req = withAuthenticatedRequest(req, &auth.AuthenticatedTx{
		Principal: auth.Principal{ActorID: "owner-1", Role: "owner"},
	}, "session-token")
	res := httptest.NewRecorder()

	s := &server{config: ServerConfig{CustomerService: service}}
	s.createCustomer(res, req)

	if res.Code != http.StatusCreated {
		t.Fatalf("status=%d body=%s", res.Code, res.Body.String())
	}
	var payload map[string]any
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		t.Fatal(err)
	}
	if payload["generated_password"] != "generated-secret" {
		t.Fatalf("generated password = %v", payload["generated_password"])
	}
	path, _ := payload["subscription_path"].(string)
	if !strings.HasPrefix(path, "/sub/") || len(strings.TrimPrefix(path, "/sub/")) < 40 {
		t.Fatalf("subscription path = %q", path)
	}
	accountPath, _ := payload["account_page_path"].(string)
	if !strings.HasPrefix(accountPath, "/s/") || strings.TrimPrefix(accountPath, "/s/") != strings.TrimPrefix(path, "/sub/") {
		t.Fatalf("account page path = %q subscription=%q", accountPath, path)
	}
	usage, _ := payload["usage_capability"].(map[string]any)
	if usage["available"] != false || usage["reason"] != "exact_accounting_not_proven" {
		t.Fatalf("usage capability = %#v", usage)
	}
	if !mutation.committed {
		t.Fatal("customer runtime mutation was not committed")
	}
	authenticated, ok := authenticatedFromRequest(req)
	if !ok || !authenticated.TransactionFinalized {
		t.Fatal("handler did not mark authenticated transaction finalized")
	}
}
