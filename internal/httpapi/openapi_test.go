package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOpenAPISpecContainsCurrentReadyRoutesOnly(t *testing.T) {
	handler := WithOperationalMiddleware(http.NotFoundHandler())
	req := httptest.NewRequest(http.MethodGet, "/api/v1/openapi.json", nil)
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", res.Code, res.Body.String())
	}
	var payload struct {
		OpenAPI string                               `json:"openapi"`
		Paths   map[string]map[string]map[string]any `json:"paths"`
	}
	if err := json.Unmarshal(res.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.OpenAPI != "3.1.0" {
		t.Fatalf("openapi=%q", payload.OpenAPI)
	}
	for _, path := range []string{"/api/v1/health/live", "/api/v1/system/status", "/api/v1/users", "/api/v1/plans"} {
		if payload.Paths[path] == nil {
			t.Fatalf("ready route %s missing", path)
		}
	}
	if _, ok := payload.Paths["/api/v1/resellers"]; ok {
		t.Fatal("unready reseller route leaked into OpenAPI")
	}
}
