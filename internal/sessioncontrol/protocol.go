// Package sessioncontrol provides the API-side client for the local-only
// session-control Unix socket served by the forwardproxy overlay. The
// overlay owns the live Naive CONNECT connections and the in-process
// pvnaiveSessionRegistry; this client translates an operator-driven kill
// request into a single, exact-tuple connection close via that registry.
package sessioncontrol

const (
	DefaultSocketPath = "/run/pvnaive/session-control.sock"
	maxRequestBytes  = 4096
	maxResponseBytes = 4096
)

type KillRequest struct {
	RuntimeCredentialID string `json:"runtime_credential_id"`
	NodeID string `json:"node_id"`
	BootID string `json:"boot_id"`
	SessionID string `json:"session_id"`
}

type KillResult struct { Found bool `json:"found"`; Killed bool `json:"killed"` }
