package httpapi

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DashSaman/PV-NaivePanel/internal/customer"
)

// The product-panel create and renewal endpoints decode ValidityInput from
// client JSON with DisallowUnknownFields; the struct fields must therefore be
// tagged with the wire names the web panel and public API contract use.
// Regression: untagged fields made any validity.duration_days payload fail
// decodeRuntimeJSON with invalid_request (400) on POST /api/v1/users.
func TestDecodeProductCreatePayloadWithValidityDays(t *testing.T) {
	body := `{"username":"u1","generate_password":true,"quota_gb":50,` +
		`"validity":{"mode":"on_creation","duration_days":30}}`
	r := httptest.NewRequest("POST", "/api/v1/users", strings.NewReader(body))
	var input customer.ProductCreateCustomerInput
	if err := decodeRuntimeJSON(r, &input); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if input.Validity.Mode != customer.ValidityOnCreation || input.Validity.DurationDays != 30 {
		t.Fatalf("validity not decoded: %+v", input.Validity)
	}
}

func TestDecodeRenewalPayloadWithValidityDays(t *testing.T) {
	body := `{"validity":{"mode":"on_first_successful_connection","duration_days":90}}`
	r := httptest.NewRequest("POST", "/api/v1/users/x/renew", strings.NewReader(body))
	var input customer.RenewalInput
	if err := decodeRuntimeJSON(r, &input); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if input.Validity.Mode != customer.ValidityOnFirstSuccessfulConnection || input.Validity.DurationDays != 90 {
		t.Fatalf("validity not decoded: %+v", input.Validity)
	}
}

func TestDecodeFixedExpiryPayload(t *testing.T) {
	expires := "2030-01-01T00:00:00Z"
	body := `{"username":"u2","generate_password":true,` +
		`"validity":{"mode":"fixed_expiry","expires_at":"` + expires + `"}}`
	r := httptest.NewRequest("POST", "/api/v1/users", strings.NewReader(body))
	var input customer.ProductCreateCustomerInput
	if err := decodeRuntimeJSON(r, &input); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if input.Validity.ExpiresAt == nil {
		t.Fatal("expires_at not decoded")
	}
	if got := input.Validity.ExpiresAt.UTC().Format(time.RFC3339); got != expires {
		t.Fatalf("expires_at = %s, want %s", got, expires)
	}
}
