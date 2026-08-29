package httpapi

import (
	"errors"
	"net/http"
	"time"

	"github.com/DashSaman/PV-NaivePanel/internal/customer"
)

func (s *server) createCustomer(w http.ResponseWriter, r *http.Request) {
	authenticated, ok := authenticatedFromRequest(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, envelope{"code": "authentication_required", "message": "Authentication is required."})
		return
	}
	if s.config.CustomerService == nil {
		writeJSON(w, http.StatusServiceUnavailable, envelope{"code": "customer_service_unavailable", "message": "Customer service is unavailable."})
		return
	}

	idempotencyKey, err := runtimeIdempotencyKey(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, envelope{"code": "invalid_idempotency_key", "message": "A valid Idempotency-Key is required."})
		return
	}

	var payload struct {
		Username         string `json:"username"`
		Password         string `json:"password"`
		GeneratePassword bool   `json:"generate_password"`
		QuotaGB          *int64 `json:"quota_gb"`
		Validity         struct {
			Mode         customer.ValidityMode `json:"mode"`
			DurationDays int                   `json:"duration_days"`
			ExpiresAt    *time.Time            `json:"expires_at"`
		} `json:"validity"`
	}
	if err := decodeRuntimeJSON(r, &payload); err != nil {
		writeJSON(w, http.StatusBadRequest, envelope{"code": "invalid_request", "message": "Invalid customer request."})
		return
	}

	result, err := s.config.CustomerService.CreateCustomer(
		r.Context(),
		authenticated.Bound.Tx,
		authenticated.Bound.Principal.ActorID,
		idempotencyKey,
		customer.CreateCustomerInput{
			Username:         payload.Username,
			Password:         payload.Password,
			GeneratePassword: payload.GeneratePassword,
			QuotaGB:          payload.QuotaGB,
			Validity: customer.ValidityInput{
				Mode:         payload.Validity.Mode,
				DurationDays: payload.Validity.DurationDays,
				ExpiresAt:    payload.Validity.ExpiresAt,
			},
		},
	)
	if err != nil {
		finishRuntimeTransaction(r)
		if customerInputError(err) {
			writeJSON(w, http.StatusBadRequest, envelope{"code": "invalid_customer_request", "message": "Customer request is invalid."})
			return
		}
		writeJSON(w, http.StatusServiceUnavailable, envelope{"code": "customer_create_failed", "message": "Customer could not be created."})
		return
	}

	// The runtime mutation finalizer commits the authenticated transaction. Mark
	// it finalized so the authentication middleware never attempts a second commit.
	authenticated.TransactionFinalized = true

	subscriptionPath, accountPagePath := subscriptionDeliveryPaths(result.SubscriptionToken)
	response := envelope{
		"user":               result.User,
		"service_term":       result.ServiceTerm,
		"runtime_credential": result.RuntimeCredential,
		"usage_capability":   result.UsageCapability,
		"subscription_path":  subscriptionPath,
		"account_page_path":  accountPagePath,
		"delivery_notice":    "Copy the subscription link and generated password now; one-time secrets are not shown again.",
	}
	if result.GeneratedPassword != "" {
		response["generated_password"] = result.GeneratedPassword
	}
	writeJSON(w, http.StatusCreated, response)
}

func customerInputError(err error) bool {
	return errors.Is(err, customer.ErrInvalidQuotaGB) ||
		errors.Is(err, customer.ErrQuotaOverflow) ||
		errors.Is(err, customer.ErrInvalidValidityMode) ||
		errors.Is(err, customer.ErrStaleValidityField) ||
		errors.Is(err, customer.ErrInvalidFixedExpiry) ||
		errors.Is(err, customer.ErrInvalidDuration)
}
