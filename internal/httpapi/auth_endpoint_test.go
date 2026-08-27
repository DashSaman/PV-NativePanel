package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DashSaman/PV-NaivePanel/internal/auth"
)

type endpointLoginStore struct {
	actor auth.ActorRecord
}

func (f *endpointLoginStore) LookupActor(context.Context, string) (auth.ActorRecord, error) {
	return f.actor, nil
}
func (f *endpointLoginStore) RecordLoginFailure(context.Context, string) (*time.Time, error) { return nil, nil }
func (f *endpointLoginStore) RecordLoginSuccess(context.Context, string) error { return nil }
func (f *endpointLoginStore) CreateSession(context.Context, string, []byte, []byte, string, []byte, time.Time, time.Time) (string, error) {
	return "00000000-0000-0000-0000-000000000099", nil
}
func (f *endpointLoginStore) GetTOTPFactorPreAuth(context.Context, string) (auth.TOTPFactorRecord, error) {
	return auth.TOTPFactorRecord{}, nil
}
func (f *endpointLoginStore) ConsumeTOTPStepPreAuth(context.Context, string, int64) (bool, error) { return true, nil }
func (f *endpointLoginStore) AppendAudit(context.Context, *string, string, string, string) error { return nil }

func TestLoginEndpointSetsHostOnlySecureCookies(t *testing.T) {
	hash, err := auth.HashPassword("test-password")
	if err != nil { t.Fatal(err) }
	service, err := auth.NewService(&endpointLoginStore{actor: auth.ActorRecord{
		ID: "00000000-0000-0000-0000-000000000001", Role: "owner", PasswordHash: hash, Status: "active",
	}}, make([]byte, 32))
	if err != nil { t.Fatal(err) }

	req := httptest.NewRequest(http.MethodPost, "https://namir.softarg.ir/api/v1/auth/login", strings.NewReader(`{"email":"owner@example.invalid","password":"test-password"}`))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	NewServer(ServerConfig{AuthService: service}).ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", res.Code, res.Body.String())
	}
	cookies := res.Result().Cookies()
	if len(cookies) != 2 {
		t.Fatalf("cookies=%d want 2", len(cookies))
	}
	seen := map[string]*http.Cookie{}
	for _, cookie := range cookies { seen[cookie.Name] = cookie }
	if seen["__Host-pvnaive_session"] == nil || seen["__Host-pvnaive_csrf"] == nil {
		t.Fatalf("missing auth cookies: %#v", seen)
	}
	if !seen["__Host-pvnaive_session"].HttpOnly || seen["__Host-pvnaive_csrf"].HttpOnly {
		t.Fatal("cookie HttpOnly contract violated")
	}
	var body map[string]any
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil { t.Fatal(err) }
	if body["status"] != "authenticated" { t.Fatalf("body=%v", body) }
}

func TestLoginEndpointUsesGenericFailure(t *testing.T) {
	hash, err := auth.HashPassword("real-password")
	if err != nil { t.Fatal(err) }
	service, err := auth.NewService(&endpointLoginStore{actor: auth.ActorRecord{
		ID: "00000000-0000-0000-0000-000000000001", Role: "owner", PasswordHash: hash, Status: "active",
	}}, make([]byte, 32))
	if err != nil { t.Fatal(err) }
	req := httptest.NewRequest(http.MethodPost, "https://namir.softarg.ir/api/v1/auth/login", strings.NewReader(`{"email":"owner@example.invalid","password":"wrong"}`))
	res := httptest.NewRecorder()
	NewServer(ServerConfig{AuthService: service}).ServeHTTP(res, req)
	if res.Code != http.StatusUnauthorized { t.Fatalf("status=%d", res.Code) }
	if strings.Contains(strings.ToLower(res.Body.String()), "password") || strings.Contains(strings.ToLower(res.Body.String()), "locked") {
		t.Fatalf("failure leaked auth reason: %s", res.Body.String())
	}
}
