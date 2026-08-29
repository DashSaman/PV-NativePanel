package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSystemStatusReturnsRealProviderDataAndSemantics(t *testing.T) {
	s := &server{config: ServerConfig{SystemStatus: func(*http.Request) (any, error) {
		return map[string]any{"sample": map[string]any{"cpu_percent": 12.5}, "dependencies": map[string]any{"database": map[string]any{"status": "ok"}}}, nil
	}}}
	req := httptest.NewRequest(http.MethodGet, "/api/v1/system/status", nil)
	res := httptest.NewRecorder()
	s.systemStatus(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", res.Code, res.Body.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(res.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload["traffic_semantics"] != "server_counter_delta" {
		t.Fatalf("traffic semantics=%v", payload["traffic_semantics"])
	}
	if payload["metrics"] == nil {
		t.Fatal("metrics payload missing")
	}
}

func TestSystemStatusFailsClosedWithoutProvider(t *testing.T) {
	s := &server{}
	req := httptest.NewRequest(http.MethodGet, "/api/v1/system/status", nil)
	res := httptest.NewRecorder()
	s.systemStatus(res, req)
	if res.Code != http.StatusServiceUnavailable {
		t.Fatalf("status=%d body=%s", res.Code, res.Body.String())
	}
}
