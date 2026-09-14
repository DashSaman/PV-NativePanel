package httpapi

import (
	"net/http"
	"time"

	"github.com/DashSaman/PV-NaivePanel/internal/customer"
	"github.com/DashSaman/PV-NaivePanel/internal/sessionkill"
)

const customerSessionStaleAfter = 90 * time.Second

func (s *server) listCustomerActiveSessions(w http.ResponseWriter, r *http.Request) {
	authenticated, ok := authenticatedFromRequest(r)
	if !ok || s.config.CustomerService == nil {
		writeJSON(w, http.StatusUnauthorized, envelope{"code": "authentication_required", "message": "Authentication is required."})
		return
	}
	userID := r.PathValue("id")
	if userID == "" {
		writeJSON(w, http.StatusBadRequest, envelope{"code": "invalid_request", "message": "User ID is required."})
		return
	}
	observedAt := time.Now().UTC()
	sessions, err := s.config.CustomerService.ListActiveSessions(
		r.Context(), authenticated.Bound.Tx, userID, observedAt, customerSessionStaleAfter,
	)
	if err != nil {
		writeJSON(w, http.StatusServiceUnavailable, envelope{"code": "sessions_unavailable", "message": "Active sessions could not be loaded."})
		return
	}
	if sessions == nil {
		sessions = []customer.CustomerSession{}
	}
	writeJSON(w, http.StatusOK, envelope{
		"sessions":    sessions,
		"observed_at": observedAt,
	})
}

func findCustomerSessionByID(sessions []customer.CustomerSession, sessionID string) (customer.CustomerSession, bool) {
	for _, session := range sessions {
		if session.SessionID == sessionID {
			return session, true
		}
	}
	return customer.CustomerSession{}, false
}

func (s *server) killCustomerActiveSession(w http.ResponseWriter, r *http.Request) {
	authenticated, ok := authenticatedFromRequest(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, envelope{"code": "authentication_required", "message": "Authentication is required."})
		return
	}
	if s.config.CustomerService == nil || s.config.SessionController == nil {
		writeJSON(w, http.StatusServiceUnavailable, envelope{"code": "session_kill_unavailable", "message": "Session kill is unavailable."})
		return
	}
	userID := r.PathValue("id")
	sessionID := r.PathValue("sessionId")
	if userID == "" || sessionID == "" {
		writeJSON(w, http.StatusBadRequest, envelope{"code": "invalid_request", "message": "Customer and session IDs are required."})
		return
	}

	observedAt := time.Now().UTC()
	sessions, err := s.config.CustomerService.ListActiveSessions(
		r.Context(), authenticated.Bound.Tx, userID, observedAt, customerSessionStaleAfter,
	)
	if err != nil {
		writeJSON(w, http.StatusServiceUnavailable, envelope{"code": "sessions_unavailable", "message": "Active sessions could not be loaded."})
		return
	}
	session, found := findCustomerSessionByID(sessions, sessionID)
	if !found {
		writeJSON(w, http.StatusNotFound, envelope{"code": "session_not_found", "message": "Active session was not found."})
		return
	}

	// Commit the authenticated ownership/RLS read before the external local
	// socket side effect. This prevents a later transaction commit failure
	// from turning a successful kill into a misleading buffered success.
	if err := authenticated.Bound.Tx.Commit(); err != nil {
		authenticated.TransactionFinalized = true
		writeJSON(w, http.StatusServiceUnavailable, envelope{"code": "session_kill_unavailable", "message": "Session ownership could not be finalized."})
		return
	}
	authenticated.TransactionFinalized = true

	result, err := s.config.SessionController.Kill(r.Context(), sessionkill.Key{
		RuntimeCredentialID: session.RuntimeCredentialID,
		NodeID:              session.NodeID,
		BootID:              session.BootID,
		SessionID:           session.SessionID,
	})
	if err != nil {
		writeJSON(w, http.StatusServiceUnavailable, envelope{"code": "session_kill_unavailable", "message": "Session kill could not be completed."})
		return
	}
	writeJSON(w, http.StatusOK, envelope{
		"status":             "completed",
		"found":              result.Found,
		"killed":             result.Killed,
		"session_id":         session.SessionID,
		"credential_mutated": false,
	})
}
