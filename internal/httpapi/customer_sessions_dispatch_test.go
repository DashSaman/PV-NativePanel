package httpapi

import (
	"testing"

	"github.com/DashSaman/PV-NaivePanel/internal/customer"
)

func TestActiveSessionRouteUsesCustomerServiceNotTelemetryPool(t *testing.T) {
	s := &server{config: ServerConfig{CustomerService: customer.NewService(nil, nil, nil)}}
	if handler := s.customerExtraHandler("users.sessions.index"); handler == nil {
		t.Fatal("active-session route was not wired from customer service")
	}
}
