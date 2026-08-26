package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRouteRegistryHasNoDuplicates(t *testing.T) {
	seen := map[string]bool{}
	for _, route := range Routes {
		key := route.Method + " " + route.Path
		if seen[key] { t.Fatalf("duplicate route: %s", key) }
		seen[key] = true
		if route.Name == "" || route.Access == "" { t.Fatalf("route missing name or access: %s", key) }
	}
}

func TestOnlyAllowlistedRoutesArePublic(t *testing.T) {
	allowed := map[string]bool{
		"health.live": true, "health.ready": true, "auth.login": true,
		"auth.refresh": true, "subscriptions.show": true, "subscriptions.info": true,
	}
	for _, route := range Routes {
		if route.Access == Public && !allowed[route.Name] { t.Fatalf("unexpected public route: %s", route.Name) }
	}
}

func TestDomainActivityIsOwnerOnly(t *testing.T) {
	for _, route := range Routes {
		if strings.Contains(route.Name, "diagnostics.domains") && route.Access != Owner {
			t.Fatalf("domain diagnostics route is not owner-only: %s", route.Name)
		}
	}
}

func TestLiveHealthAndSecurityHeaders(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/health/live", nil)
	res := httptest.NewRecorder()
	NewServer().ServeHTTP(res, req)
	if res.Code != http.StatusOK { t.Fatalf("status=%d", res.Code) }
	if res.Header().Get("X-Content-Type-Options") != "nosniff" { t.Fatal("missing nosniff header") }
	if res.Header().Get("Cache-Control") != "no-store" { t.Fatal("missing no-store header") }
	var body map[string]any
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil { t.Fatal(err) }
	if body["service"] != "pvnative-api" { t.Fatalf("unexpected service: %v", body["service"]) }
}

func TestProtectedScaffoldRouteFailsClosed(t *testing.T) {
	for _, path := range []string{"/api/v1/users", "/api/v1/logs/application", "/api/v1/diagnostics/domain-activity"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		res := httptest.NewRecorder()
		NewServer().ServeHTTP(res, req)
		if res.Code != http.StatusUnauthorized { t.Fatalf("%s status=%d", path, res.Code) }
	}
}

func TestPublicUnimplementedRouteDoesNotPretendToWork(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader("{}"))
	res := httptest.NewRecorder()
	NewServer().ServeHTTP(res, req)
	if res.Code != http.StatusNotImplemented { t.Fatalf("status=%d", res.Code) }
}
