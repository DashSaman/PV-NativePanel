package httpapi

import (
	"crypto/sha256"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRBACMatrix(t *testing.T) {
	tests := []struct {
		access Access
		role   string
		want   bool
	}{
		{Authenticated, "owner", true},
		{Authenticated, "reseller", true},
		{Owner, "owner", true},
		{Owner, "admin", false},
		{Admin, "owner", true},
		{Admin, "admin", true},
		{Admin, "operator", false},
		{Operator, "owner", true},
		{Operator, "operator", true},
		{Operator, "admin", false},
		{Auditor, "owner", true},
		{Auditor, "auditor", true},
		{Auditor, "operator", false},
		{Reseller, "owner", true},
		{Reseller, "admin", true},
		{Reseller, "reseller", true},
	}
	for _, tt := range tests {
		if got := roleAllowed(tt.access, tt.role); got != tt.want {
			t.Fatalf("roleAllowed(%q,%q)=%v want %v", tt.access, tt.role, got, tt.want)
		}
	}
}

func TestSessionCookieSecurityContract(t *testing.T) {
	cookie := newSessionCookie("opaque-value")
	if cookie.Name != "__Host-pvnaive_session" || cookie.Value != "opaque-value" {
		t.Fatalf("unexpected session cookie: %#v", cookie)
	}
	if !cookie.Secure || !cookie.HttpOnly || cookie.Path != "/" || cookie.Domain != "" || cookie.SameSite != http.SameSiteStrictMode {
		t.Fatalf("insecure session cookie: %#v", cookie)
	}
}

func TestCSRFCookieAndRequestBinding(t *testing.T) {
	raw := "csrf-random-value"
	hash := sha256.Sum256([]byte(raw))
	cookie := newCSRFCookie(raw)
	if cookie.Name != "__Host-pvnaive_csrf" || !cookie.Secure || cookie.HttpOnly || cookie.Path != "/" || cookie.Domain != "" || cookie.SameSite != http.SameSiteStrictMode {
		t.Fatalf("insecure CSRF cookie: %#v", cookie)
	}

	req := httptest.NewRequest(http.MethodPost, "https://namir.softarg.ir/api/v1/auth/logout", strings.NewReader("{}"))
	req.AddCookie(cookie)
	req.Header.Set("X-CSRF-Token", raw)
	req.Header.Set("Origin", "https://namir.softarg.ir")
	if err := validateCSRF(req, hash[:]); err != nil {
		t.Fatalf("valid CSRF rejected: %v", err)
	}

	crossSite := req.Clone(req.Context())
	crossSite.Header = req.Header.Clone()
	crossSite.Header.Set("Sec-Fetch-Site", "cross-site")
	if err := validateCSRF(crossSite, hash[:]); err == nil {
		t.Fatal("cross-site unsafe request accepted")
	}

	mismatch := req.Clone(req.Context())
	mismatch.Header = req.Header.Clone()
	mismatch.Header.Set("X-CSRF-Token", "wrong")
	if err := validateCSRF(mismatch, hash[:]); err == nil {
		t.Fatal("mismatched CSRF token accepted")
	}
}

func TestUnsafeMethodClassification(t *testing.T) {
	for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete} {
		if !unsafeMethod(method) {
			t.Fatalf("%s should be unsafe", method)
		}
	}
	for _, method := range []string{http.MethodGet, http.MethodHead, http.MethodOptions} {
		if unsafeMethod(method) {
			t.Fatalf("%s should be safe", method)
		}
	}
}
