package telemetry

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

const (
	DefaultTelemetrySocketPath = "/run/pvnaive/accounting.sock"
	TelemetryIngestPath        = "/v1/accounting/event"
	TelemetryHealthPath        = "/v1/accounting/health"
	maxTelemetryRequestBytes   = 32 << 10
)

type IngestResult struct {
	Tracked        bool   `json:"tracked"`
	Accepted       bool   `json:"accepted"`
	Duplicate      bool   `json:"duplicate"`
	ServiceTermID  string `json:"service_term_id,omitempty"`
	UploadDelta    int64  `json:"upload_delta"`
	DownloadDelta  int64  `json:"download_delta"`
	QuotaDepleted  bool   `json:"quota_depleted"`
	RemainingBytes *int64 `json:"remaining_bytes,omitempty"`
}

type Ingestor interface {
	Ingest(context.Context, Event) (IngestResult, error)
}

type telemetryHandler struct {
	ingestor Ingestor
}

func NewTelemetryHandler(ingestor Ingestor) http.Handler {
	return &telemetryHandler{ingestor: ingestor}
}

func (h *telemetryHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch {
	case r.Method == http.MethodGet && r.URL.Path == TelemetryHealthPath:
		writeTelemetryJSON(w, http.StatusOK, map[string]string{"status": "ok"})
		return
	case r.Method == http.MethodPost && r.URL.Path == TelemetryIngestPath:
		h.handleIngest(w, r)
		return
	default:
		http.NotFound(w, r)
	}
}

func (h *telemetryHandler) handleIngest(w http.ResponseWriter, r *http.Request) {
	if h.ingestor == nil {
		writeTelemetryError(w, http.StatusServiceUnavailable, "telemetry unavailable")
		return
	}
	defer r.Body.Close()
	decoder := json.NewDecoder(io.LimitReader(r.Body, maxTelemetryRequestBytes+1))
	decoder.DisallowUnknownFields()
	var event Event
	if err := decoder.Decode(&event); err != nil {
		writeTelemetryError(w, http.StatusBadRequest, "invalid telemetry event")
		return
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		writeTelemetryError(w, http.StatusBadRequest, "invalid telemetry event")
		return
	}
	if err := ValidateEvent(event); err != nil {
		writeTelemetryError(w, http.StatusBadRequest, "invalid telemetry event")
		return
	}
	result, err := h.ingestor.Ingest(r.Context(), event)
	if err != nil {
		writeTelemetryError(w, http.StatusConflict, "telemetry event rejected")
		return
	}
	writeTelemetryJSON(w, http.StatusOK, result)
}

func writeTelemetryError(w http.ResponseWriter, status int, message string) {
	writeTelemetryJSON(w, status, map[string]string{"error": message})
}

func writeTelemetryJSON(w http.ResponseWriter, status int, value any) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
