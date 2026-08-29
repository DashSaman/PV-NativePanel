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

type apiAdoptUpdateStore struct {
	apiCustomerStore
	adoptable runtimecred.CredentialView
	updated   *customer.UpdateServiceTermRecord
}

func (s *apiAdoptUpdateStore) AdoptableRuntimeCredentialTx(context.Context, *sql.Tx, string) (runtimecred.CredentialView, error) {
	return s.adoptable, nil
}

func (s *apiAdoptUpdateStore) UpdateCurrentServiceTermTx(_ context.Context, _ *sql.Tx, userID string, record customer.UpdateServiceTermRecord) (customer.ServiceTerm, error) {
	copy := record
	s.updated = &copy
	return customer.ServiceTerm{
		ID: "term-existing", TenantID: "tenant-direct", UserID: userID,
		QuotaBytes: record.QuotaBytes, DurationSeconds: record.DurationSeconds,
		StartPolicy: record.StartPolicy, StartsAt: record.StartsAt, ExpiresAt: record.ExpiresAt,
		State: record.State, Revision: 2,
	}, nil
}

func authenticatedOwnerRequest(method, path, body string) *http.Request {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", "legacy-service-0001")
	return withAuthenticatedRequest(req, &auth.AuthenticatedTx{
		Principal: auth.Principal{ActorID: "owner-1", Role: "owner"},
	}, "session-token")
}

func TestAdoptRuntimeCustomerEndpointKeepsExistingRuntimeIdentity(t *testing.T) {
	now := time.Date(2026, 8, 29, 15, 0, 0, 0, time.UTC)
	store := &apiAdoptUpdateStore{adoptable: runtimecred.CredentialView{
		ID: "runtime-old-1", Username: "amirreza", Status: runtimecred.CredentialActive,
		Origin: runtimecred.CredentialImported, Revision: 4,
	}}
	service := customer.NewService(store, func(context.Context, *sql.Tx, string, string, runtimecred.CreateInput) (customer.RuntimeMutation, error) {
		t.Fatal("legacy adoption must not mutate Runtime")
		return nil, nil
	}, func() time.Time { return now })

	req := authenticatedOwnerRequest(http.MethodPost, "/api/v1/customers/adopt-runtime", `{"runtime_credential_id":"runtime-old-1","quota_gb":80,"validity":{"mode":"on_creation","duration_days":30}}`)
	res := httptest.NewRecorder()
	s := &server{config: ServerConfig{CustomerService: service}}
	s.adoptRuntimeCustomer(res, req)

	if res.Code != http.StatusCreated {
		t.Fatalf("status=%d body=%s", res.Code, res.Body.String())
	}
	var payload map[string]any
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		t.Fatal(err)
	}
	runtimeView, _ := payload["runtime_credential"].(map[string]any)
	if runtimeView["id"] != "runtime-old-1" || runtimeView["username"] != "amirreza" {
		t.Fatalf("runtime identity changed: %#v", runtimeView)
	}
	if _, leaked := payload["generated_password"]; leaked {
		t.Fatalf("adoption must not return a password: %#v", payload)
	}
	path, _ := payload["subscription_path"].(string)
	if !strings.HasPrefix(path, "/sub/") {
		t.Fatalf("subscription path=%q", path)
	}
	accountPath, _ := payload["account_page_path"].(string)
	if strings.TrimPrefix(accountPath, "/s/") != strings.TrimPrefix(path, "/sub/") {
		t.Fatalf("account page path=%q subscription=%q", accountPath, path)
	}
}

func TestUpdateCustomerServiceEndpointChangesOnlyCommercialSettings(t *testing.T) {
	now := time.Date(2026, 8, 29, 16, 0, 0, 0, time.UTC)
	store := &apiAdoptUpdateStore{}
	service := customer.NewService(store, func(context.Context, *sql.Tx, string, string, runtimecred.CreateInput) (customer.RuntimeMutation, error) {
		t.Fatal("service edit must not mutate Runtime")
		return nil, nil
	}, func() time.Time { return now })
	expires := now.Add(10 * 24 * time.Hour).Format(time.RFC3339)

	req := authenticatedOwnerRequest(http.MethodPatch, "/api/v1/customers/user-existing/service", `{"quota_gb":120,"validity":{"mode":"fixed_expiry","expires_at":"`+expires+`"}}`)
	req.SetPathValue("id", "user-existing")
	res := httptest.NewRecorder()
	s := &server{config: ServerConfig{CustomerService: service}}
	s.updateCustomerService(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", res.Code, res.Body.String())
	}
	if store.updated == nil || store.updated.QuotaBytes == nil || *store.updated.QuotaBytes != 120*customer.BytesPerCustomerGB {
		t.Fatalf("updated record=%#v", store.updated)
	}
	var payload map[string]any
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		t.Fatal(err)
	}
	if payload["runtime_mutated"] != false {
		t.Fatalf("runtime mutation flag=%v", payload["runtime_mutated"])
	}
}
