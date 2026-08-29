package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DashSaman/PV-NaivePanel/internal/runtimecred"
	"github.com/DashSaman/PV-NaivePanel/internal/subscription"
)

func subscriptionContractFixture(t *testing.T, record subscription.Record) (http.Handler, string) {
	t.Helper()
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i + 1)
	}
	ciphertext, nonce, err := runtimecred.EncryptSecret(key, []byte("customer-secret-123"))
	if err != nil {
		t.Fatal(err)
	}
	rawToken, hash, err := subscription.GenerateToken()
	if err != nil {
		t.Fatal(err)
	}
	record.RuntimeCredentialID = "runtime-1"
	record.Username = "Amir22"
	record.SecretCiphertext = ciphertext
	record.SecretNonce = nonce
	record.EncryptionKeyID = "runtime-v1"
	if record.UserState == "" {
		record.UserState = "active"
	}
	if record.TermState == "" {
		record.TermState = "active"
	}
	store := &fakeSubscriptionStore{hash: hash, record: record}
	service, err := subscription.NewService(store, key, "runtime-v1")
	if err != nil {
		t.Fatal(err)
	}
	return NewServer(ServerConfig{
		SubscriptionService:   service,
		SubscriptionProxyHost: "namir.softarg.ir:443",
	}), rawToken
}

func TestMachineSubscriptionEndpointIgnoresBrowserAcceptHeader(t *testing.T) {
	handler, token := subscriptionContractFixture(t, subscription.Record{})
	req := httptest.NewRequest(http.MethodGet, "/sub/"+token, nil)
	req.Header.Set("Accept", "text/html,application/xhtml+xml")
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", res.Code, res.Body.String())
	}
	if got := res.Header().Get("Content-Type"); !strings.HasPrefix(got, "text/plain") {
		t.Fatalf("machine content-type=%q", got)
	}
	if strings.Contains(strings.ToLower(res.Body.String()), "<!doctype html") {
		t.Fatal("machine endpoint returned HTML because of Accept header")
	}
	if got := strings.TrimSpace(res.Body.String()); got != "naive+https://Amir22:customer-secret-123@namir.softarg.ir:443" {
		t.Fatalf("machine body=%q", got)
	}
}

func TestLegacySubscriptionEndpointRemainsMachineOnly(t *testing.T) {
	handler, token := subscriptionContractFixture(t, subscription.Record{})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/subscriptions/"+token, nil)
	req.Header.Set("Accept", "text/html")
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", res.Code, res.Body.String())
	}
	if got := res.Header().Get("Content-Type"); !strings.HasPrefix(got, "text/plain") {
		t.Fatalf("legacy content-type=%q", got)
	}
	if strings.Contains(strings.ToLower(res.Body.String()), "<!doctype html") {
		t.Fatal("legacy endpoint must not render the account page")
	}
}

func TestHumanAccountEndpointAlwaysHTMLAndUsesNewSubscriptionPath(t *testing.T) {
	quota := int64(50 * 1024 * 1024 * 1024)
	expires := time.Now().UTC().Add(72 * time.Hour)
	handler, token := subscriptionContractFixture(t, subscription.Record{
		QuotaBytes:      &quota,
		StartPolicy:     "on_creation",
		ExpiresAt:       &expires,
		DurationSeconds: 30 * 24 * 60 * 60,
	})
	req := httptest.NewRequest(http.MethodGet, "/s/"+token+"?lang=en", nil)
	req.Header.Set("Accept", "text/plain")
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", res.Code, res.Body.String())
	}
	if got := res.Header().Get("Content-Type"); !strings.HasPrefix(got, "text/html") {
		t.Fatalf("account content-type=%q", got)
	}
	body := res.Body.String()
	for _, want := range []string{
		"<html lang=\"en\" dir=\"ltr\">",
		"PVNaive",
		"PVNETWORK",
		"Amir22",
		"50 GB",
		"Usage unavailable",
		"/sub/" + token,
		"Subscription QR",
		"Direct Naive QR",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("account page missing %q", want)
		}
	}
	if strings.Contains(body, "/api/v1/subscriptions/"+token) {
		t.Fatal("account page advertised legacy subscription path")
	}
}

func TestAccountPageSecurityHeadersAreSecretSafe(t *testing.T) {
	handler, token := subscriptionContractFixture(t, subscription.Record{})
	req := httptest.NewRequest(http.MethodGet, "/s/"+token, nil)
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", res.Code, res.Body.String())
	}
	checks := map[string]string{
		"Cache-Control":          "no-store",
		"Pragma":                 "no-cache",
		"X-Content-Type-Options": "nosniff",
		"X-Frame-Options":        "DENY",
		"Referrer-Policy":        "no-referrer",
	}
	for header, want := range checks {
		if got := res.Header().Get(header); got != want {
			t.Fatalf("%s=%q want %q", header, got, want)
		}
	}
	if robots := res.Header().Get("X-Robots-Tag"); !strings.Contains(robots, "noindex") || !strings.Contains(robots, "noarchive") {
		t.Fatalf("X-Robots-Tag=%q", robots)
	}
	csp := res.Header().Get("Content-Security-Policy")
	for _, want := range []string{"default-src 'none'", "img-src 'self' data:", "frame-ancestors 'none'", "base-uri 'none'"} {
		if !strings.Contains(csp, want) {
			t.Fatalf("CSP=%q missing %q", csp, want)
		}
	}
}

func TestAccountPageSupportsPersianRTLWithoutActivatingFirstUse(t *testing.T) {
	handler, token := subscriptionContractFixture(t, subscription.Record{
		TermState:        "pending",
		StartPolicy:      "on_first_successful_connection",
		FirstConnectedAt: nil,
	})
	req := httptest.NewRequest(http.MethodGet, "/s/"+token+"?lang=fa", nil)
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", res.Code, res.Body.String())
	}
	body := res.Body.String()
	for _, want := range []string{"<html lang=\"fa\" dir=\"rtl\">", "منتظر اولین اتصال", "از اولین اتصال موفق", "در دسترس نیست"} {
		if !strings.Contains(body, want) {
			t.Fatalf("persian account page missing %q", want)
		}
	}
}

func TestSuspendedAccountPageIsReadableButMachineEndpointIsUnavailable(t *testing.T) {
	handler, token := subscriptionContractFixture(t, subscription.Record{UserState: "suspended"})

	pageReq := httptest.NewRequest(http.MethodGet, "/s/"+token+"?lang=en", nil)
	pageRes := httptest.NewRecorder()
	handler.ServeHTTP(pageRes, pageReq)
	if pageRes.Code != http.StatusOK || !strings.Contains(pageRes.Body.String(), "Suspended") {
		t.Fatalf("suspended account page status=%d body=%s", pageRes.Code, pageRes.Body.String())
	}

	machineReq := httptest.NewRequest(http.MethodGet, "/sub/"+token, nil)
	machineRes := httptest.NewRecorder()
	handler.ServeHTTP(machineRes, machineReq)
	if machineRes.Code != http.StatusNotFound {
		t.Fatalf("suspended machine status=%d body=%s", machineRes.Code, machineRes.Body.String())
	}
}
