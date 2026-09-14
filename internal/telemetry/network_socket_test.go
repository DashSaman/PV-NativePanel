package telemetry

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func networkTestBackend(t *testing.T) (*NetworkSampleBackend, *PostgresStore) {
	t.Helper()
	return &NetworkSampleBackend{store: &PostgresStore{}, holder: NewEWMAHolder(DefaultNetworkAggConfig())}, nil
}

func TestNetworkSampleRequestDecoding(t *testing.T) {
	body := `{"samples":[{"runtime_credential_id":"11111111-1111-1111-1111-111111111111","node_id":"direct-1","boot_id":"22222222-2222-2222-2222-222222222222","session_id":"33333333-3333-3333-3333-333333333333","sample_seq":1,"sampled_at":"2026-09-14T10:00:00Z","path":"upstream","rtt_micros":42000,"rtt_var_micros":1000,"segs_out":100,"segs_retrans":2,"bytes_in":1000,"bytes_out":2000,"duration_micros":60000000}]}`
	var request NetworkSampleRequest
	decoder := json.NewDecoder(bytes.NewReader([]byte(body)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(request.Samples) != 1 {
		t.Fatalf("samples: %d", len(request.Samples))
	}
	if err := ValidateNetworkSample(request.Samples[0]); err != nil {
		t.Fatalf("round-trip validation failed: %v", err)
	}
}

func TestNetworkSampleHandlerRejectionPaths(t *testing.T) {
	// No backend wired: the endpoint must report 503 without panicking.
	handler := NewTelemetryHandler(struct{}{})
	server := httptest.NewServer(handler)
	defer server.Close()

	request, err := http.NewRequest(http.MethodPost, server.URL+TelemetryNetworkSamplePath, bytes.NewReader([]byte(`{"samples":[]}`)))
	if err != nil {
		t.Fatal(err)
	}
	response, err := server.Client().Do(request)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 without backend, got %d", response.StatusCode)
	}
}

func TestNetworkSampleBackendIngestValidation(t *testing.T) {
	backend := &NetworkSampleBackend{store: &PostgresStore{}, holder: NewEWMAHolder(DefaultNetworkAggConfig())}
	if _, err := backend.Ingest(context.Background(), NetworkSampleRequest{Samples: nil}); err == nil {
		t.Fatal("empty batch must be rejected")
	}
	valid := netTestSample(1, time.Now().UTC(), "33333333-3333-3333-3333-333333333333", 1000, 10, 0, 0, 0, 1000)
	// PostgresStore without a live DB: Ingest must fail closed (DB required),
	// not silently pretend success.
	if _, err := backend.Ingest(context.Background(), NetworkSampleRequest{Samples: []NetworkSample{valid}}); err == nil {
		t.Fatal("ingest without DB must fail closed")
	}
	_ = networkTestBackend
}
