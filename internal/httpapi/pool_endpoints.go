package httpapi

import (
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/DashSaman/PV-NaivePanel/internal/fleet"
)

// R5 Pool Manager endpoints (GATE STEER-006, backend slice).
// Pull-model registry: the primary panel stores signed desired-state
// revisions; sibling agents PULL and keep last-known-good on registry
// unreachability. All routes are Owner-only (RBAC deny-by-default) and
// every mutation goes through the fleet.Store trusted boundary (0033
// SECURITY DEFINER functions — no direct table access).
//
// The manifest pull surface for sibling agents rides the same revision
// store; the dedicated mTLS control listener is tracked as R5-PULL-001.

const (
	// manifestValidityDefault bounds how long a published desired-state
	// manifest stays trusted after a node loses registry reachability.
	manifestValidityDefault = 48 * time.Hour
	manifestValidityMin     = 1 * time.Hour
	manifestValidityMax     = 336 * time.Hour // two weeks
	maxEndpointsPerManifest = 32
)

type poolEnrollTokenPayload struct {
	NodeName   string `json:"node_name"`
	TTLMinutes int    `json:"ttl_minutes"`
}

type poolEnrollPayload struct {
	Token          string `json:"token"`
	DisplayName    string `json:"display_name"`
	Region         string `json:"region,omitempty"`
	CapacityWeight int    `json:"capacity_weight"`
}

type poolRevisionPayload struct {
	PoolID          string              `json:"pool_id"`
	Endpoints       []poolEndpointInput `json:"endpoints"`
	Weight          float64             `json:"weight"`
	ValidityMinutes int                 `json:"validity_minutes,omitempty"`
}

type poolEndpointInput struct {
	Host string `json:"host"`
	Port int    `json:"port"`
	SNI  string `json:"sni,omitempty"`
}

type poolMaintenancePayload struct {
	State string `json:"state"`
}

func (s *server) poolNodesIndex(w http.ResponseWriter, r *http.Request) {
	if s.config.FleetStore == nil {
		writeJSON(w, http.StatusServiceUnavailable, envelope{"code": "pool_unavailable", "message": "Pool registry is unavailable."})
		return
	}
	nodes, err := s.config.FleetStore.ListNodes(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, envelope{"code": "pool_list_failed", "message": "Pool nodes could not be listed."})
		return
	}
	items := make([]map[string]any, 0, len(nodes))
	for _, n := range nodes {
		item := map[string]any{
			"id":               n.ID,
			"display_name":     n.DisplayName,
			"region":           n.Region,
			"capacity_weight":  n.CapacityWeight,
			"health":           n.Health,
			"maintenance":      n.Maintenance,
			"desired_revision": n.DesiredRevision,
			"applied_revision": n.AppliedRevision,
			"drift":            fleet.Drift(fleet.Node{DesiredRevision: uint64(n.DesiredRevision), AppliedRevision: uint64(n.AppliedRevision)}),
		}
		if n.LastSeenAt != nil {
			item["last_seen_at"] = n.LastSeenAt.UTC()
		}
		items = append(items, item)
	}
	writeJSON(w, http.StatusOK, envelope{"nodes": items, "count": len(items)})
}

func (s *server) poolEnrollTokenCreate(w http.ResponseWriter, r *http.Request) {
	if s.config.FleetStore == nil {
		writeJSON(w, http.StatusServiceUnavailable, envelope{"code": "pool_unavailable", "message": "Pool registry is unavailable."})
		return
	}
	var payload poolEnrollTokenPayload
	if err := decodeStrictJSON(r, &payload); err != nil {
		writeJSON(w, http.StatusBadRequest, envelope{"code": "invalid_request", "message": "Invalid request."})
		return
	}
	ttl := fleet.DefaultEnrollmentTTL
	if payload.TTLMinutes != 0 {
		ttl = time.Duration(payload.TTLMinutes) * time.Minute
	}
	rawToken, expiresAt, err := s.config.FleetStore.IssueEnrollmentToken(r.Context(), payload.NodeName, ttl)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, envelope{"code": "pool_enroll_token_invalid", "message": "Node name or TTL is invalid."})
		return
	}
	// The raw token is returned exactly once; only its SHA-256 hash was stored.
	writeJSON(w, http.StatusCreated, envelope{
		"token":      rawToken,
		"expires_at": expiresAt.UTC(),
		"note":       "store this token now; it is never shown again",
	})
}

func (s *server) poolNodesEnroll(w http.ResponseWriter, r *http.Request) {
	if s.config.FleetStore == nil {
		writeJSON(w, http.StatusServiceUnavailable, envelope{"code": "pool_unavailable", "message": "Pool registry is unavailable."})
		return
	}
	request, ok := authenticatedFromRequest(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, envelope{"code": "authentication_failed", "message": "Authentication failed."})
		return
	}
	var payload poolEnrollPayload
	if err := decodeStrictJSON(r, &payload); err != nil {
		writeJSON(w, http.StatusBadRequest, envelope{"code": "invalid_request", "message": "Invalid request."})
		return
	}
	weight := payload.CapacityWeight
	if weight == 0 {
		weight = 1
	}
	nodeID, enrolled, reason, err := s.config.FleetStore.Enroll(
		r.Context(), payload.Token, payload.DisplayName, payload.Region, weight,
		request.Bound.Principal.ActorID,
	)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, envelope{"code": "pool_enroll_invalid", "message": "Enrollment request is invalid."})
		return
	}
	if !enrolled {
		writeJSON(w, http.StatusConflict, envelope{"code": "pool_enroll_rejected", "reason": reason, "message": "Enrollment token is invalid, expired or already used."})
		return
	}
	writeJSON(w, http.StatusCreated, envelope{"id": nodeID, "status": "enrolled"})
}

func (s *server) poolRevisionPublish(w http.ResponseWriter, r *http.Request) {
	if s.config.FleetStore == nil || strings.TrimSpace(s.config.FleetSigningKey) == "" {
		writeJSON(w, http.StatusServiceUnavailable, envelope{"code": "pool_unavailable", "message": "Pool revision publishing is unavailable."})
		return
	}
	nodeID := r.PathValue("id")
	if strings.TrimSpace(nodeID) == "" {
		writeJSON(w, http.StatusBadRequest, envelope{"code": "pool_node_id_required", "message": "Node id is required."})
		return
	}
	var payload poolRevisionPayload
	if err := decodeStrictJSON(r, &payload); err != nil {
		writeJSON(w, http.StatusBadRequest, envelope{"code": "invalid_request", "message": "Invalid request."})
		return
	}
	now := time.Now().UTC()
	manifest, err := buildManifest(nodeID, payload, s.config.FleetSigningKey, now)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, envelope{"code": "pool_revision_invalid", "message": err.Error()})
		return
	}
	revision, accepted, err := s.config.FleetStore.PublishRevision(r.Context(), nodeID, manifest, s.config.FleetSigningKey, now)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, envelope{"code": "pool_revision_invalid", "message": "Revision payload is invalid."})
		return
	}
	if !accepted {
		writeJSON(w, http.StatusConflict, envelope{"code": "pool_revision_rejected", "message": "Node is unknown or disabled; revisions are refused."})
		return
	}
	writeJSON(w, http.StatusCreated, envelope{"node_id": nodeID, "revision": revision, "status": "published"})
}

func (s *server) poolManifestShow(w http.ResponseWriter, r *http.Request) {
	if s.config.FleetStore == nil {
		writeJSON(w, http.StatusServiceUnavailable, envelope{"code": "pool_unavailable", "message": "Pool registry is unavailable."})
		return
	}
	nodeID := r.PathValue("id")
	revision, env, found, err := s.config.FleetStore.LatestRevision(r.Context(), nodeID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, envelope{"code": "pool_manifest_failed", "message": "Latest manifest could not be read."})
		return
	}
	if !found {
		writeJSON(w, http.StatusNotFound, envelope{"code": "pool_manifest_none", "message": "No published revision for this node."})
		return
	}
	if err := fleet.VerifyEnvelope(env, time.Now().UTC()); err != nil {
		// A stored envelope that no longer verifies is surfaced as a server
		// error, never served as trusted desired state.
		writeJSON(w, http.StatusInternalServerError, envelope{"code": "pool_manifest_unverifiable", "message": "Stored manifest failed verification."})
		return
	}
	body, err := json.Marshal(map[string]any{"revision": revision, "manifest": env.Manifest, "signature": env.Signature})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, envelope{"code": "pool_manifest_failed", "message": "Latest manifest could not be encoded."})
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(body)
}

func (s *server) poolMaintenanceSet(w http.ResponseWriter, r *http.Request) {
	if s.config.FleetStore == nil {
		writeJSON(w, http.StatusServiceUnavailable, envelope{"code": "pool_unavailable", "message": "Pool registry is unavailable."})
		return
	}
	nodeID := r.PathValue("id")
	var payload poolMaintenancePayload
	if err := decodeStrictJSON(r, &payload); err != nil {
		writeJSON(w, http.StatusBadRequest, envelope{"code": "invalid_request", "message": "Invalid request."})
		return
	}
	tracked, err := s.config.FleetStore.SetMaintenance(r.Context(), nodeID, payload.State)
	if errors.Is(err, fleet.ErrDrainRequired) {
		writeJSON(w, http.StatusConflict, envelope{"code": "pool_drain_required", "message": "A live node must be drained before it is disabled."})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusBadRequest, envelope{"code": "pool_maintenance_invalid", "message": "Maintenance state is invalid."})
		return
	}
	if !tracked {
		writeJSON(w, http.StatusNotFound, envelope{"code": "pool_node_unknown", "message": "Node does not exist."})
		return
	}
	writeJSON(w, http.StatusOK, envelope{"id": nodeID, "maintenance": payload.State})
}

// buildManifest derives the operator public key from the signing key, clamps
// the validity window and validates every endpoint input before the store
// (fail-closed input contract, mirroring the panel-access validator style).
func buildManifest(nodeID string, payload poolRevisionPayload, signingKey string, now time.Time) (fleet.Manifest, error) {
	if strings.TrimSpace(payload.PoolID) == "" || len(payload.PoolID) > 60 {
		return fleet.Manifest{}, errors.New("pool_id is required (1..60 characters)")
	}
	if len(payload.Endpoints) == 0 || len(payload.Endpoints) > maxEndpointsPerManifest {
		return fleet.Manifest{}, errors.New("at least one endpoint is required (max 32)")
	}
	if payload.Weight <= 0 || payload.Weight > 1000 {
		return fleet.Manifest{}, errors.New("weight must be within 0.1..1000")
	}
	validity := manifestValidityDefault
	if payload.ValidityMinutes != 0 {
		validity = time.Duration(payload.ValidityMinutes) * time.Minute
	}
	if validity < manifestValidityMin || validity > manifestValidityMax {
		return fleet.Manifest{}, errors.New("validity_minutes must be 60..20160")
	}
	decoded, err := hex.DecodeString(strings.TrimSpace(signingKey))
	if err != nil || len(decoded) != ed25519.PrivateKeySize {
		return fleet.Manifest{}, errors.New("operator signing key is malformed")
	}
	priv := ed25519.PrivateKey(decoded)
	pubKey, ok := priv.Public().(ed25519.PublicKey)
	if !ok {
		return fleet.Manifest{}, errors.New("operator signing key is malformed")
	}
	endpoints := make([]fleet.ManifestEndpoint, 0, len(payload.Endpoints))
	for _, e := range payload.Endpoints {
		if strings.TrimSpace(e.Host) == "" || len(e.Host) > 255 {
			return fleet.Manifest{}, errors.New("endpoint host is required (max 255 characters)")
		}
		if e.Port <= 0 || e.Port > 65535 {
			return fleet.Manifest{}, errors.New("endpoint port must be 1..65535")
		}
		if len(e.SNI) > 255 {
			return fleet.Manifest{}, errors.New("endpoint SNI too long")
		}
		endpoints = append(endpoints, fleet.ManifestEndpoint{Host: strings.TrimSpace(e.Host), Port: e.Port, SNI: strings.TrimSpace(e.SNI)})
	}
	return fleet.Manifest{
		Schema:     fleet.ManifestSchema,
		NodeID:     nodeID,
		PoolID:     strings.TrimSpace(payload.PoolID),
		Endpoints:  endpoints,
		PubKeyHex:  hex.EncodeToString(pubKey),
		Weight:     payload.Weight,
		ValidFrom:  now,
		ValidUntil: now.Add(validity),
	}, nil
}
