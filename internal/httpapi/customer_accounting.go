package httpapi

import (
	"net/http"
	"time"

	"github.com/DashSaman/PV-NaivePanel/internal/customer"
	"github.com/DashSaman/PV-NaivePanel/internal/telemetry"
)

const customerAccountingStaleAfter = 90 * time.Second

func (s *server) listProductCustomersAccounting(w http.ResponseWriter, r *http.Request) {
	authenticated, ok := authenticatedFromRequest(r)
	if !ok || s.config.CustomerService == nil {
		writeJSON(w, http.StatusUnauthorized, envelope{"code": "authentication_required", "message": "Authentication is required."})
		return
	}
	query, err := parseCustomerListQuery(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, envelope{"code": "invalid_customer_filter", "message": "Customer filters are invalid."})
		return
	}
	page, err := s.config.CustomerService.SearchCustomers(r.Context(), authenticated.Bound.Tx, query)
	if err != nil {
		writeJSON(w, http.StatusServiceUnavailable, envelope{"code": "customer_list_failed", "message": "Customers could not be loaded."})
		return
	}
	if s.config.AccountingStore != nil {
		now := time.Now().UTC()
		for i := range page.Customers {
			termID := page.Customers[i].ServiceTermID
			if termID == "" {
				continue
			}
			model, readErr := s.config.AccountingStore.Read(r.Context(), termID, now, customerAccountingStaleAfter)
			if readErr != nil {
				continue
			}
			applyCustomerAccounting(&page.Customers[i], model, now)
		}
	}
	writeJSON(w, http.StatusOK, envelope{
		"customers": page.Customers,
		"page":      page.Page,
		"page_size": page.PageSize,
		"total":     page.Total,
	})
}

func applyCustomerAccounting(view *customer.CustomerView, model telemetry.ReadModel, now time.Time) {
	if view == nil {
		return
	}
	complete := model.AccountingComplete
	view.Usage = &customer.CustomerUsage{
		Available:          complete,
		AccountingComplete: complete,
		UploadBytes:        model.UploadBytes,
		DownloadBytes:      model.DownloadBytes,
		UsedBytes:          model.UsedBytes,
		RemainingBytes:     model.RemainingBytes,
		LastOnline:         model.LastOnline,
		Online:             model.Online,
		SessionCount:       model.SessionCount,
	}
	if complete {
		view.UsageCapability = customer.UsageCapability{Available: true}
	} else {
		view.UsageCapability = customer.UsageCapability{Available: false, Reason: "accounting_incomplete"}
	}
	if view.FirstConnectedAt == nil && model.FirstConnectedAt != nil {
		view.FirstConnectedAt = model.FirstConnectedAt
	}

	used := model.UsedBytes
	var presence *customer.PresenceStatus
	if complete {
		value := customer.PresenceOffline
		if model.Online {
			value = customer.PresenceOnline
		}
		presence = &value
	}
	view.StatusDimensions = customer.DeriveStatusDimensions(customer.StatusInput{
		UserState:              view.Status,
		TermState:              view.ServiceState,
		StartPolicy:            view.StartPolicy,
		QuotaBytes:             view.QuotaBytes,
		ExpiresAt:              view.ExpiresAt,
		OnHold:                 view.OnHold,
		AccountingAvailable:    complete,
		UsedBytes:              &used,
		Presence:               presence,
		RuntimeHealthAvailable: false,
		Now:                    now,
	})
}
