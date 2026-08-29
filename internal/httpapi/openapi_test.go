package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOpenAPISpecContainsReleasedOperationalRoutes(t *testing.T) {
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
	if payload.Paths["/api/v1/health/live"]["get"] == nil {
		t.Fatal("health route missing")
	}
	if payload.Paths["/api/v1/system/status"]["get"] == nil {
		t.Fatal("system status route missing")
	}
	if _, ok := payload.Paths["/api/v1/users"]; ok {
		t.Fatal("unreleased scaffold route leaked into OpenAPI")
	}
}
