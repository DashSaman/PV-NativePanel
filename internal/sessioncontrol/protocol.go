// Package sessioncontrol provides the API-side client for the local-only
// session-control Unix socket served by the forwardproxy overlay. The
// overlay owns the live Naive CONNECT connections and the in-process
// pvnaiveSessionRegistry; this client translates an operator-driven kill
// request into a single, exact-tuple connection close via that registry.
//
// The control boundary is local-only: the socket is a Unix-domain listener
// at a fixed path under /run/pvnaive, reachable only from processes on
// the same host. Authentication of the operator is performed by the
// API-side HTTP handler (CSRF + RBAC + user-ownership) before this
// client is ever called. The client itself carries no credentials.
package sessioncontrol

const (
	// DefaultSocketPath is the fixed Unix-domain socket path the overlay
	// listens on for session-control commands.
	DefaultSocketPath = "/run/pvnaive/session-control.sock"

	maxRequestBytes  = 4096
	maxResponseBytes = 4096
)

// KillRequest is the wire format for an exact-session-kill command. Every
// field must be populated with the exact, trusted tuple that the caller
// obtained from the accounting read model — partial or forged tuples are
// rejected by the overlay and never match a live session they do not own.
type KillRequest struct {
	RuntimeCredentialID string `json:"runtime_credential_id"`
	NodeID              string `json:"node_id"`
	BootID              string `json:"boot_id"`
	SessionID           string `json:"session_id"`
}

// KillResult is the wire-format response from the overlay kill endpoint.
type KillResult struct {
	// Found reports whether a matching live session was registered in the
	// overlay's in-process registry at the time the kill was processed.
	Found bool `json:"found"`
	// Killed reports whether this request actually closed the live
	// connection. false for an already-cancelled (idempotent repeat) or
	// already-unregistered session.
	Killed bool `json:"killed"`
}
