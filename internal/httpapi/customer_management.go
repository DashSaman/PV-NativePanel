package httpapi

import (
	"net/http"
	"net/url"
)

func (s *server) listCustomers(w http.ResponseWriter, r *http.Request) {
	authenticated, ok := authenticatedFromRequest(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, envelope{"code": "authentication_required", "message": "Authentication is required."})
		return
	}
	if s.config.CustomerService == nil {
		writeJSON(w, http.StatusServiceUnavailable, envelope{"code": "customer_service_unavailable", "message": "Customer service is unavailable."})
		return
	}
	customers, err := s.config.CustomerService.ListCustomers(r.Context(), authenticated.Bound.Tx)
	if err != nil {
		writeJSON(w, http.StatusServiceUnavailable, envelope{"code": "customer_list_failed", "message": "Customers could not be loaded."})
		return
	}
	writeJSON(w, http.StatusOK, envelope{"customers": customers})
}

func (s *server) rotateCustomerSubscription(w http.ResponseWriter, r *http.Request) {
	authenticated, ok := authenticatedFromRequest(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, envelope{"code": "authentication_required", "message": "Authentication is required."})
		return
	}
	if s.config.CustomerService == nil {
		writeJSON(w, http.StatusServiceUnavailable, envelope{"code": "customer_service_unavailable", "message": "Customer service is unavailable."})
		return
	}
	if _, err := runtimeIdempotencyKey(r); err != nil {
		writeJSON(w, http.StatusBadRequest, envelope{"code": "invalid_idempotency_key", "message": "A valid Idempotency-Key is required."})
		return
	}
	userID := r.PathValue("id")
	if userID == "" {
		writeJSON(w, http.StatusBadRequest, envelope{"code": "invalid_customer", "message": "Customer id is required."})
		return
	}
	rawToken, err := s.config.CustomerService.RotateSubscription(r.Context(), authenticated.Bound.Tx, userID)
	if err != nil {
		_ = authenticated.Bound.Tx.Rollback()
		authenticated.TransactionFinalized = true
		writeJSON(w, http.StatusServiceUnavailable, envelope{"code": "subscription_rotate_failed", "message": "Subscription could not be reissued."})
		return
	}
	writeJSON(w, http.StatusCreated, envelope{
		"subscription_path": "/api/v1/subscriptions/" + url.PathEscape(rawToken),
		"delivery_notice":   "This replacement subscription token is shown once; previous active tokens were revoked.",
	})
}
