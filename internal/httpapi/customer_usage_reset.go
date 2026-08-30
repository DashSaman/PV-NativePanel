package httpapi

import (
	"errors"
	"net/http"

	"github.com/DashSaman/PV-NaivePanel/internal/customer"
)

func (s *server) resetCustomerUsage(w http.ResponseWriter, r *http.Request) {
	authenticated, key, userID, ok := customerOperationIdentity(w, r, s)
	if !ok {
		return
	}
	var payload struct {
		Confirm bool `json:"confirm"`
	}
	if err := decodeStrictJSON(r, &payload); err != nil || !payload.Confirm {
		writeJSON(w, http.StatusBadRequest, envelope{
			"code":    "reset_confirmation_required",
			"message": "Explicit confirmation is required before resetting usage.",
		})
		return
	}

	result, err := s.config.CustomerService.ResetCustomerUsage(
		r.Context(), authenticated.Bound.Tx, authenticated.Bound.Principal.ActorID, key, userID,
	)
	if err != nil {
		finishRuntimeTransaction(r)
		writeUsageResetError(w, err)
		return
	}
	if err := authenticated.Bound.Tx.Commit(); err != nil {
		authenticated.TransactionFinalized = true
		writeJSON(w, http.StatusServiceUnavailable, envelope{
			"code":    "usage_reset_commit_failed",
			"message": "Usage reset was not confirmed because the database transaction did not commit.",
		})
		return
	}
	authenticated.TransactionFinalized = true
	writeJSON(w, http.StatusOK, envelope{
		"reset_event":           result.Event,
		"idempotent_replay":     result.IdempotentReplay,
		"runtime_mutated":       false,
		"password_rotated":      false,
		"subscription_reissued": false,
		"message":               "Current usage period was reset without rotating Runtime credentials, Password, or Subscription.",
	})
}

func writeUsageResetError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, customer.ErrUsageResetNotFound):
		writeJSON(w, http.StatusNotFound, envelope{"code": "customer_not_found", "message": "Managed customer service was not found."})
	case errors.Is(err, customer.ErrCustomerIdempotencyConflict):
		writeJSON(w, http.StatusConflict, envelope{"code": "idempotency_conflict", "message": "This Idempotency-Key belongs to a different customer mutation."})
	case errors.Is(err, customer.ErrUsageResetServiceInactive):
		writeJSON(w, http.StatusConflict, envelope{"code": "usage_reset_service_inactive", "message": "Ended or revoked service usage cannot be reset."})
	case errors.Is(err, customer.ErrUsageResetAccountingIncomplete):
		writeJSON(w, http.StatusConflict, envelope{"code": "usage_reset_accounting_incomplete", "message": "Exact accounting is incomplete; usage was not reset."})
	case errors.Is(err, customer.ErrUsageResetReservationPending):
		writeJSON(w, http.StatusConflict, envelope{"code": "usage_reset_reservation_pending", "message": "A quota reservation is still pending; usage was not reset."})
	case errors.Is(err, customer.ErrUsageResetTelemetryStale):
		writeJSON(w, http.StatusConflict, envelope{"code": "usage_reset_telemetry_stale", "message": "An open session has stale telemetry; usage was not reset."})
	case errors.Is(err, customer.ErrUsageResetTimeConflict):
		writeJSON(w, http.StatusConflict, envelope{"code": "usage_reset_time_conflict", "message": "The reset time conflicts with the current accounting period."})
	default:
		writeJSON(w, http.StatusServiceUnavailable, envelope{"code": "usage_reset_failed", "message": "Usage could not be reset safely."})
	}
}
