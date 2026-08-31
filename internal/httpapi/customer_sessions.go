package httpapi

import (
	"net/http"
	"time"

	"github.com/DashSaman/PV-NaivePanel/internal/customer"
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

func (s *server) notImplementedSessionKill(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusNotImplemented, envelope{
		"code":    "session_kill_not_implemented",
		"message": "Session kill is not available yet. Planned for Task 13.",
	})
}
