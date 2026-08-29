package httpapi

import (
	"errors"
	"net/http"
	"time"

	"github.com/DashSaman/PV-NaivePanel/internal/customer"
)

type customerServiceSettingsPayload struct {
	QuotaGB  *int64 `json:"quota_gb"`
	Validity struct {
		Mode         customer.ValidityMode `json:"mode"`
		DurationDays int                   `json:"duration_days"`
		ExpiresAt    *time.Time            `json:"expires_at"`
	} `json:"validity"`
}

func validityInputFromPayload(payload customerServiceSettingsPayload) customer.ValidityInput {
	return customer.ValidityInput{
		Mode:         payload.Validity.Mode,
		DurationDays: payload.Validity.DurationDays,
		ExpiresAt:    payload.Validity.ExpiresAt,
	}
}

func (s *server) adoptRuntimeCustomer(w http.ResponseWriter, r *http.Request) {
	authenticated, ok := authenticatedFromRequest(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, envelope{"code": "authentication_required", "message": "Authentication is required."})
		return
	}
	if s.config.CustomerService == nil {
		writeJSON(w, http.StatusServiceUnavailable, envelope{"code": "customer_service_unavailable", "message": "Customer service is unavailable."})
		return
	}

	var payload struct {
		RuntimeCredentialID string `json:"runtime_credential_id"`
		customerServiceSettingsPayload
	}
	if err := decodeRuntimeJSON(r, &payload); err != nil {
		writeJSON(w, http.StatusBadRequest, envelope{"code": "invalid_request", "message": "Invalid legacy account request."})
		return
	}
	result, err := s.config.CustomerService.AdoptRuntimeCredential(
		r.Context(),
		authenticated.Bound.Tx,
		authenticated.Bound.Principal.ActorID,
		customer.AdoptRuntimeInput{
			RuntimeCredentialID: payload.RuntimeCredentialID,
			QuotaGB:             payload.QuotaGB,
			Validity:            validityInputFromPayload(payload.customerServiceSettingsPayload),
		},
	)
	if err != nil {
		switch {
		case customerInputError(err):
			writeJSON(w, http.StatusBadRequest, envelope{"code": "invalid_customer_request", "message": "Volume or validity is invalid."})
		case errors.Is(err, customer.ErrRuntimeCredentialNotAdoptable):
			writeJSON(w, http.StatusConflict, envelope{"code": "runtime_already_managed", "message": "This Runtime account is already managed or unavailable."})
		default:
			writeJSON(w, http.StatusServiceUnavailable, envelope{"code": "customer_adopt_failed", "message": "Existing Runtime account could not be added to customer management."})
		}
		return
	}
	subscriptionPath, accountPagePath := subscriptionDeliveryPaths(result.SubscriptionToken)
	writeJSON(w, http.StatusCreated, envelope{
		"user":               result.User,
		"service_term":       result.ServiceTerm,
		"runtime_credential": result.RuntimeCredential,
		"usage_capability":   result.UsageCapability,
		"subscription_path":  subscriptionPath,
		"account_page_path":  accountPagePath,
		"delivery_notice":    "Existing username, password and Runtime credential were preserved. Only service metadata and a new subscription link were added.",
	})
}

func (s *server) updateCustomerService(w http.ResponseWriter, r *http.Request) {
	authenticated, ok := authenticatedFromRequest(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, envelope{"code": "authentication_required", "message": "Authentication is required."})
		return
	}
	if s.config.CustomerService == nil {
		writeJSON(w, http.StatusServiceUnavailable, envelope{"code": "customer_service_unavailable", "message": "Customer service is unavailable."})
		return
	}
	userID := r.PathValue("id")
	if userID == "" {
		writeJSON(w, http.StatusBadRequest, envelope{"code": "invalid_customer", "message": "Customer id is required."})
		return
	}
	var payload customerServiceSettingsPayload
	if err := decodeRuntimeJSON(r, &payload); err != nil {
		writeJSON(w, http.StatusBadRequest, envelope{"code": "invalid_request", "message": "Invalid service edit request."})
		return
	}
	term, err := s.config.CustomerService.UpdateCustomerService(
		r.Context(), authenticated.Bound.Tx, userID,
		customer.UpdateServiceInput{QuotaGB: payload.QuotaGB, Validity: validityInputFromPayload(payload)},
	)
	if err != nil {
		switch {
		case customerInputError(err):
			writeJSON(w, http.StatusBadRequest, envelope{"code": "invalid_customer_request", "message": "Volume or validity is invalid."})
		case errors.Is(err, customer.ErrCustomerServiceNotFound):
			writeJSON(w, http.StatusNotFound, envelope{"code": "customer_not_found", "message": "Managed customer service was not found."})
		default:
			writeJSON(w, http.StatusServiceUnavailable, envelope{"code": "customer_service_update_failed", "message": "Customer service settings could not be updated."})
		}
		return
	}
	writeJSON(w, http.StatusOK, envelope{
		"service_term":    term,
		"runtime_mutated": false,
		"message":         "Volume and validity updated without changing the Runtime credential or password.",
	})
}
