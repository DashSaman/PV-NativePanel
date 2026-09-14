// Package fleetpull implements R5-PULL-001: the dedicated mTLS control
// listener over which sibling agents PULL their signed desired-state
// manifest from the primary registry (docs/STEERING_SPEC_FA.md §4).
//
// Security contract (fail-closed everywhere):
//   - Node identity comes ONLY from the TLS client certificate (URI SAN
//     urn:pvnaive:node:<uuid>, fallback CN) — never from headers, query
//     parameters or bodies.
//   - The listener requires and verifies client certificates against the
//     operator CA; a handshake without a cert never reaches the handler.
//   - Unknown/untracked nodes get 403 before any manifest bytes are read.
//   - A stored envelope that fails VerifyEnvelope is surfaced as a server
//     error and never served as trusted desired state.
//   - The heartbeat carried by the pull is plain fleet-state reporting; it
//     never becomes exact byte-accounting/quota truth (R1 boundary).
package fleetpull

import (
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/DashSaman/PV-NaivePanel/internal/fleet"
)

// nodeIDURISANPrefix is the URI SAN scheme carrying the node identity.
const nodeIDURISANPrefix = "urn:pvnaive:node:"

// DefaultMaxSkew bounds the preflight clock-skew warning (not a hard
// refusal — the manifest validity windows are the real trust gate).
const DefaultMaxSkew = 90 * time.Second

// IdentityFromTLS extracts the node id from the verified peer certificate:
// first the URI SAN (urn:pvnaive:node:<id>), then the CN. Empty when no
// verified cert or no usable identity field (caller must fail closed).
func IdentityFromTLS(peerCertificates []*x509.Certificate) string {
	if len(peerCertificates) == 0 {
		return ""
	}
	cert := peerCertificates[0]
	for _, uri := range cert.URIs {
		if uri != nil && strings.HasPrefix(uri.String(), nodeIDURISANPrefix) {
			return strings.TrimPrefix(uri.String(), nodeIDURISANPrefix)
		}
	}
	return strings.TrimSpace(cert.Subject.CommonName)
}

type pullRequest struct {
	health          string
	appliedRevision int64
	agentVersion    string
	agentNow        time.Time
	hasAgentNow     bool
}

// parsePullRequest decodes and bounds the optional agent preflight
// parameters. Everything is optional; anything malformed fails the pull
// (fail-closed input contract, never "best effort" interpretation).
func parsePullRequest(queryParams map[string]string, now time.Time) (pullRequest, error) {
	var req pullRequest
	req.health = strings.TrimSpace(queryParams["health"])
	if req.health != "" {
		if err := fleet.ValidateHealth(req.health); err != nil {
			return pullRequest{}, errors.New("health must be one of unknown/healthy/degraded/offline")
		}
	}
	if raw := strings.TrimSpace(queryParams["applied_revision"]); raw != "" {
		var parsed int64
		if _, err := fmt.Sscanf(raw, "%d", &parsed); err != nil || parsed < 0 || parsed > math.MaxInt64/2 {
			return pullRequest{}, errors.New("applied_revision must be a non-negative integer")
		}
		req.appliedRevision = parsed
	}
	req.agentVersion = strings.TrimSpace(queryParams["agent_version"])
	if len(req.agentVersion) > 64 {
		return pullRequest{}, errors.New("agent_version must be at most 64 characters")
	}
	if raw := strings.TrimSpace(queryParams["now"]); raw != "" {
		parsed, err := time.Parse(time.RFC3339Nano, raw)
		if err != nil {
			return pullRequest{}, errors.New("now must be an RFC3339 timestamp")
		}
		req.agentNow = parsed
		req.hasAgentNow = true
	}
	return req, nil
}

type pullResponse struct {
	Revision      int64          `json:"revision"`
	Manifest      fleet.Manifest `json:"manifest"`
	Signature     string         `json:"signature"`
	ServerTime    string         `json:"server_time"`
	ClockSkewMS   *int64         `json:"clock_skew_ms,omitempty"`
	ClockSkewWarn bool           `json:"clock_skew_warning,omitempty"`
	AgentVersion  string         `json:"agent_version,omitempty"`
}

// Handler serves GET /fleet/v1/manifest over the mTLS listener. Store may
// be nil in tests; a nil store fails closed (503).
type Handler struct {
	Store   *fleet.Store
	Now     func() time.Time
	MaxSkew time.Duration
}

func (h *Handler) now() time.Time {
	if h.Now != nil {
		return h.Now()
	}
	return time.Now()
}

func (h *Handler) maxSkew() time.Duration {
	if h.MaxSkew > 0 {
		return h.MaxSkew
	}
	return DefaultMaxSkew
}

type errorEnvelope struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	body, err := json.Marshal(errorEnvelope{Code: code, Message: message})
	if err != nil {
		http.Error(w, code, status)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_, _ = w.Write(body)
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		writeError(w, http.StatusMethodNotAllowed, "fleet_method_not_allowed", "Only GET is supported.")
		return
	}
	if r.URL.Path != "/fleet/v1/manifest" {
		writeError(w, http.StatusNotFound, "fleet_route_unknown", "Unknown fleet control route.")
		return
	}
	if r.TLS == nil || len(r.TLS.PeerCertificates) == 0 {
		// The TLS config already demands a client cert; reaching here means
		// a non-TLS connection hit the handler. Fail closed, serve nothing.
		writeError(w, http.StatusUnauthorized, "fleet_tls_identity_required", "A verified client certificate is required.")
		return
	}
	nodeID := IdentityFromTLS(r.TLS.PeerCertificates)
	if nodeID == "" {
		writeError(w, http.StatusForbidden, "fleet_identity_missing", "Client certificate carries no node identity.")
		return
	}
	req, err := parsePullRequest(map[string]string{
		"health":           r.URL.Query().Get("health"),
		"applied_revision": r.URL.Query().Get("applied_revision"),
		"agent_version":    r.URL.Query().Get("agent_version"),
		"now":              r.URL.Query().Get("now"),
	}, h.now())
	if err != nil {
		writeError(w, http.StatusBadRequest, "fleet_preflight_invalid", err.Error())
		return
	}
	if h.Store == nil {
		writeError(w, http.StatusServiceUnavailable, "fleet_registry_unavailable", "Fleet registry is unavailable.")
		return
	}
	now := h.now()
	if req.health != "" {
		tracked, heartbeatErr := h.Store.Heartbeat(r.Context(), nodeID, req.health, req.appliedRevision)
		if heartbeatErr != nil {
			writeError(w, http.StatusInternalServerError, "fleet_heartbeat_failed", "Heartbeat could not be recorded.")
			return
		}
		if !tracked {
			writeError(w, http.StatusForbidden, "fleet_node_unknown", "Node is not enrolled in this registry.")
			return
		}
	}
	revision, env, found, latestErr := h.Store.LatestRevision(r.Context(), nodeID)
	if latestErr != nil {
		writeError(w, http.StatusInternalServerError, "fleet_manifest_failed", "Manifest could not be read.")
		return
	}
	if !found {
		writeError(w, http.StatusNotFound, "fleet_manifest_none", "No published revision for this node.")
		return
	}
	if verifyErr := fleet.VerifyEnvelope(env, now); verifyErr != nil {
		writeError(w, http.StatusInternalServerError, "fleet_manifest_unverifiable", "Stored manifest failed verification.")
		return
	}
	response := pullResponse{
		Revision:     revision,
		Manifest:     env.Manifest,
		Signature:    env.Signature,
		ServerTime:   now.UTC().Format(time.RFC3339Nano),
		AgentVersion: req.agentVersion,
	}
	if req.hasAgentNow {
		skew := now.Sub(req.agentNow)
		millis := skew.Milliseconds()
		response.ClockSkewMS = &millis
		response.ClockSkewWarn = absDuration(skew) > h.maxSkew()
	}
	body, encodeErr := json.Marshal(response)
	if encodeErr != nil {
		writeError(w, http.StatusInternalServerError, "fleet_manifest_failed", "Manifest could not be encoded.")
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(body)
}

func absDuration(value time.Duration) time.Duration {
	if value < 0 {
		return -value
	}
	return value
}
