package telemetry

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"
)

const (
	DefaultTelemetrySocketPath = "/run/pvnaive/accounting.sock"
	TelemetryAuthorizePath     = "/v1/accounting/authorize"
	TelemetryClaimPath         = "/v1/accounting/claim"
	TelemetryIngestPath        = "/v1/accounting/event"
	TelemetryHealthPath        = "/v1/accounting/health"
	maxTelemetryRequestBytes   = 32 << 10
)

type AuthorizeRequest struct {
	RuntimeCredentialID string    `json:"runtime_credential_id"`
	ObservedAt          time.Time `json:"timestamp"`
}

type IngestResult struct {
	Tracked            bool       `json:"tracked"`
	Accepted           bool       `json:"accepted"`
	Duplicate          bool       `json:"duplicate"`
	Reason             string     `json:"reason"`
	ServiceTermID      string     `json:"service_term_id,omitempty"`
	UploadDelta        int64      `json:"upload_delta"`
	DownloadDelta      int64      `json:"download_delta"`
	QuotaDepleted      bool       `json:"quota_depleted"`
	RemainingBytes     *int64     `json:"remaining_bytes,omitempty"`
	FirstConnectedAt   *time.Time `json:"first_connected_at,omitempty"`
	AccountingComplete bool       `json:"accounting_complete"`
}

type Ingestor interface {
	Ingest(context.Context, Event) (IngestResult, error)
}

type Authorizer interface {
	Authorize(context.Context, string, time.Time) (AuthorizationResult, error)
}

type Claimer interface {
	Claim(context.Context, ClaimRequest) (ClaimResult, error)
}

type telemetryHandler struct {
	backend any
}

func NewTelemetryHandler(backend any) http.Handler {
	return &telemetryHandler{backend: backend}
}

func (h *telemetryHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch {
	case r.Method == http.MethodGet && r.URL.Path == TelemetryHealthPath:
		writeTelemetryJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	case r.Method == http.MethodPost && r.URL.Path == TelemetryAuthorizePath:
		h.handleAuthorize(w, r)
	case r.Method == http.MethodPost && r.URL.Path == TelemetryClaimPath:
		h.handleClaim(w, r)
	case r.Method == http.MethodPost && r.URL.Path == TelemetryIngestPath:
		h.handleIngest(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (h *telemetryHandler) handleAuthorize(w http.ResponseWriter, r *http.Request) {
	backend, ok := h.backend.(Authorizer)
	if !ok || backend == nil {
		writeTelemetryError(w, http.StatusServiceUnavailable, "telemetry unavailable")
		return
	}
	var request AuthorizeRequest
	if !decodeStrictJSON(w, r, &request) {
		return
	}
	if !validUUID(request.RuntimeCredentialID) || request.ObservedAt.IsZero() {
		writeTelemetryError(w, http.StatusBadRequest, "invalid authorization request")
		return
	}
	result, err := backend.Authorize(r.Context(), request.RuntimeCredentialID, request.ObservedAt)
	if err != nil {
		writeTelemetryError(w, http.StatusServiceUnavailable, "authorization unavailable")
		return
	}
	writeTelemetryJSON(w, http.StatusOK, result)
}

func (h *telemetryHandler) handleClaim(w http.ResponseWriter, r *http.Request) {
	backend, ok := h.backend.(Claimer)
	if !ok || backend == nil {
		writeTelemetryError(w, http.StatusServiceUnavailable, "telemetry unavailable")
		return
	}
	var request ClaimRequest
	if !decodeStrictJSON(w, r, &request) {
		return
	}
	if !validClaimRequest(request) {
		writeTelemetryError(w, http.StatusBadRequest, "invalid quota claim")
		return
	}
	result, err := backend.Claim(r.Context(), request)
	if err != nil {
		writeTelemetryError(w, http.StatusConflict, "quota claim rejected")
		return
	}
	writeTelemetryJSON(w, http.StatusOK, result)
}

func (h *telemetryHandler) handleIngest(w http.ResponseWriter, r *http.Request) {
	backend, ok := h.backend.(Ingestor)
	if !ok || backend == nil {
		writeTelemetryError(w, http.StatusServiceUnavailable, "telemetry unavailable")
		return
	}
	var event Event
	if !decodeStrictJSON(w, r, &event) {
		return
	}
	if err := ValidateEvent(event); err != nil {
		writeTelemetryError(w, http.StatusBadRequest, "invalid telemetry event")
		return
	}
	result, err := backend.Ingest(r.Context(), event)
	if err != nil {
		writeTelemetryError(w, http.StatusConflict, "telemetry event rejected")
		return
	}
	writeTelemetryJSON(w, http.StatusOK, result)
}

func decodeStrictJSON(w http.ResponseWriter, r *http.Request, target any) bool {
	defer r.Body.Close()
	decoder := json.NewDecoder(io.LimitReader(r.Body, maxTelemetryRequestBytes+1))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		writeTelemetryError(w, http.StatusBadRequest, "invalid telemetry request")
		return false
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		writeTelemetryError(w, http.StatusBadRequest, "invalid telemetry request")
		return false
	}
	return true
}

func writeTelemetryError(w http.ResponseWriter, status int, message string) {
	writeTelemetryJSON(w, status, map[string]string{"error": message})
}

func writeTelemetryJSON(w http.ResponseWriter, status int, value any) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
