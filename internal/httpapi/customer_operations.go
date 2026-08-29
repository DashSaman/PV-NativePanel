package httpapi

import (
	"errors"
	"net/http"

	"github.com/DashSaman/PV-NaivePanel/internal/customer"
	"github.com/DashSaman/PV-NaivePanel/internal/runtimecred"
)

func customerOperationIdentity(w http.ResponseWriter, r *http.Request, s *server) (*authenticatedRequest, string, string, bool) {
	authenticated, ok := authenticatedFromRequest(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, envelope{"code": "authentication_required", "message": "Authentication is required."})
		return nil, "", "", false
	}
	if s.config.CustomerService == nil {
		writeJSON(w, http.StatusServiceUnavailable, envelope{"code": "customer_service_unavailable", "message": "Customer service is unavailable."})
		return nil, "", "", false
	}
	idempotencyKey, err := runtimeIdempotencyKey(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, envelope{"code": "invalid_idempotency_key", "message": "A valid Idempotency-Key is required."})
		return nil, "", "", false
	}
	userID := r.PathValue("id")
	if userID == "" {
		writeJSON(w, http.StatusBadRequest, envelope{"code": "invalid_customer", "message": "Customer id is required."})
		return nil, "", "", false
	}
	return authenticated, idempotencyKey, userID, true
}

func writeCustomerRuntimeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, runtimecred.ErrRevisionConflict):
		writeJSON(w, http.StatusConflict, envelope{"code": "revision_conflict", "message": "Customer Runtime state changed. Refresh and retry."})
	case errors.Is(err, runtimecred.ErrLastActiveCredential):
		writeJSON(w, http.StatusConflict, envelope{"code": "last_active_credential", "message": "At least one active Naive credential must remain."})
	case errors.Is(err, runtimecred.ErrIdempotentReplay):
		writeJSON(w, http.StatusConflict, envelope{"code": "idempotency_replay", "message": "This customer operation was already submitted. Refresh before retrying."})
	case errors.Is(err, runtimecred.ErrReconciliationRequired):
		writeJSON(w, http.StatusServiceUnavailable, envelope{"code": "runtime_reconciliation_required", "message": "Runtime reconciliation is required before more customer mutations."})
	case errors.Is(err, customer.ErrCustomerLifecycleUnavailable), errors.Is(err, customer.ErrCustomerServiceNotFound):
		writeJSON(w, http.StatusNotFound, envelope{"code": "customer_not_found", "message": "Managed customer was not found."})
	default:
		writeJSON(w, http.StatusServiceUnavailable, envelope{"code": "customer_operation_failed", "message": "Customer operation could not be completed safely."})
	}
}

func (s *server) suspendCustomer(w http.ResponseWriter, r *http.Request) {
	authenticated, key, userID, ok := customerOperationIdentity(w, r, s)
	if !ok {
		return
	}
	credential, err := s.config.CustomerService.SuspendCustomer(r.Context(), authenticated.Bound.Tx, authenticated.Bound.Principal.ActorID, key, userID)
	if err != nil {
		finishRuntimeTransaction(r)
		writeCustomerRuntimeError(w, err)
		return
	}
	authenticated.TransactionFinalized = true
	writeJSON(w, http.StatusOK, envelope{"status": "suspended", "runtime_credential": credential})
}

func (s *server) resumeCustomer(w http.ResponseWriter, r *http.Request) {
	authenticated, key, userID, ok := customerOperationIdentity(w, r, s)
	if !ok {
		return
	}
	credential, err := s.config.CustomerService.ResumeCustomer(r.Context(), authenticated.Bound.Tx, authenticated.Bound.Principal.ActorID, key, userID)
	if err != nil {
		finishRuntimeTransaction(r)
		writeCustomerRuntimeError(w, err)
		return
	}
	authenticated.TransactionFinalized = true
	writeJSON(w, http.StatusOK, envelope{"status": "active", "runtime_credential": credential})
}

func (s *server) deleteCustomer(w http.ResponseWriter, r *http.Request) {
	authenticated, key, userID, ok := customerOperationIdentity(w, r, s)
	if !ok {
		return
	}
	credential, err := s.config.CustomerService.RevokeCustomer(r.Context(), authenticated.Bound.Tx, authenticated.Bound.Principal.ActorID, key, userID)
	if err != nil {
		finishRuntimeTransaction(r)
		writeCustomerRuntimeError(w, err)
		return
	}
	authenticated.TransactionFinalized = true
	writeJSON(w, http.StatusOK, envelope{
		"status":             "revoked",
		"runtime_credential": credential,
		"message":            "Customer access was safely revoked. Database history was preserved.",
	})
}

func (s *server) rotateCustomerPassword(w http.ResponseWriter, r *http.Request) {
	authenticated, key, userID, ok := customerOperationIdentity(w, r, s)
	if !ok {
		return
	}
	var payload struct {
		Password         string `json:"password"`
		GeneratePassword bool   `json:"generate_password"`
	}
	if err := decodeRuntimeJSON(r, &payload); err != nil {
		writeJSON(w, http.StatusBadRequest, envelope{"code": "invalid_request", "message": "Invalid password rotation request."})
		return
	}
	credential, generatedPassword, err := s.config.CustomerService.RotateCustomerPassword(
		r.Context(), authenticated.Bound.Tx, authenticated.Bound.Principal.ActorID, key, userID, payload.Password, payload.GeneratePassword,
	)
	if err != nil {
		finishRuntimeTransaction(r)
		writeCustomerRuntimeError(w, err)
		return
	}
	authenticated.TransactionFinalized = true
	response := envelope{
		"runtime_credential": credential,
		"delivery_notice":    "Password rotation is independent from Subscription. The active Subscription token was not reissued.",
	}
	if generatedPassword != "" {
		response["generated_password"] = generatedPassword
	}
	writeJSON(w, http.StatusOK, response)
}
