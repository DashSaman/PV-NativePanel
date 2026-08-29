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

type recordingIngestor struct {
	events []Event
	result IngestResult
	err    error
}

func (r *recordingIngestor) Ingest(_ context.Context, event Event) (IngestResult, error) {
	r.events = append(r.events, event)
	return r.result, r.err
}

func validSocketEvent() Event {
	return Event{
		RuntimeCredentialID: "11111111-1111-1111-1111-111111111111",
		Username:            "alice",
		NodeID:              "direct-1",
		BootID:              "22222222-2222-2222-2222-222222222222",
		SessionID:           "33333333-3333-3333-3333-333333333333",
		Sequence:            1,
		ObservedAt:          time.Date(2026, 8, 29, 18, 0, 0, 0, time.UTC),
		AuthenticatedConnect: true,
		UploadBytes:         100,
		DownloadBytes:       200,
	}
}

func TestTelemetryHandlerAcceptsOnlyFixedIngestRoute(t *testing.T) {
	ingestor := &recordingIngestor{result: IngestResult{Tracked: true, Accepted: true}}
	handler := NewTelemetryHandler(ingestor)
	payload, err := json.Marshal(validSocketEvent())
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodPost, TelemetryIngestPath, bytes.NewReader(payload))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if len(ingestor.events) != 1 {
		t.Fatalf("expected one ingested event, got %d", len(ingestor.events))
	}
}

func TestTelemetryHandlerCannotReachManagementOperations(t *testing.T) {
	ingestor := &recordingIngestor{}
	handler := NewTelemetryHandler(ingestor)
	for _, path := range []string{
		"/v1/runtime/apply",
		"/v1/runtime/rollback",
		"/v1/runtime/validate",
		"/v1/runtime/inspect",
		"/v1/accounting/arbitrary",
	} {
		req := httptest.NewRequest(http.MethodPost, path, bytes.NewBufferString(`{"service":"caddy","path":"/etc/passwd","command":"reload"}`))
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Fatalf("path %s must be unreachable, got %d", path, rec.Code)
		}
	}
	if len(ingestor.events) != 0 {
		t.Fatalf("management-shaped requests must not reach ingestor")
	}
}

func TestTelemetryHandlerRejectsUnknownFieldsAndNonConnect(t *testing.T) {
	ingestor := &recordingIngestor{}
	handler := NewTelemetryHandler(ingestor)

	body := `{"runtime_credential_id":"11111111-1111-1111-1111-111111111111","username":"alice","node_id":"direct-1","boot_id":"22222222-2222-2222-2222-222222222222","session_id":"33333333-3333-3333-3333-333333333333","sequence":1,"timestamp":"2026-08-29T18:00:00Z","authenticated_connect":true,"upload_bytes":1,"download_bytes":2,"command":"reload"}`
	req := httptest.NewRequest(http.MethodPost, TelemetryIngestPath, bytes.NewBufferString(body))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("unknown field must be rejected, got %d", rec.Code)
	}

	event := validSocketEvent()
	event.AuthenticatedConnect = false
	payload, _ := json.Marshal(event)
	req = httptest.NewRequest(http.MethodPost, TelemetryIngestPath, bytes.NewReader(payload))
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("non-authenticated CONNECT telemetry must be rejected, got %d", rec.Code)
	}
}

func TestTelemetryHandlerHealthIsReadOnly(t *testing.T) {
	handler := NewTelemetryHandler(&recordingIngestor{})
	req := httptest.NewRequest(http.MethodGet, TelemetryHealthPath, nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected health 200, got %d", rec.Code)
	}
	var got map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got["status"] != "ok" {
		t.Fatalf("unexpected health response: %#v", got)
	}
}
