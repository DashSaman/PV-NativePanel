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

func TestPublicSubscriptionBrowserGetsPVNaiveStatusPage(t *testing.T) {
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
	quota := int64(50 * 1024 * 1024 * 1024)
	expires := time.Date(2026, 9, 28, 14, 0, 0, 0, time.UTC)
	store := &fakeSubscriptionStore{hash: hash, record: subscription.Record{
		RuntimeCredentialID: "runtime-1",
		Username:            "Amir22",
		SecretCiphertext:    ciphertext,
		SecretNonce:         nonce,
		EncryptionKeyID:     "runtime-v1",
		UserState:           "active",
		TermState:           "active",
		QuotaBytes:          &quota,
		DurationSeconds:     30 * 24 * 60 * 60,
		StartPolicy:         "on_creation",
		ExpiresAt:           &expires,
	}}
	service, err := subscription.NewService(store, key, "runtime-v1")
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/subscriptions/"+rawToken, nil)
	req.Host = "attacker.example"
	req.Header.Set("Accept", "text/html,application/xhtml+xml")
	res := httptest.NewRecorder()
	NewServer(ServerConfig{
		SubscriptionService:   service,
		SubscriptionProxyHost: "namir.softarg.ir:443",
	}).ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", res.Code, res.Body.String())
	}
	if got := res.Header().Get("Content-Type"); !strings.HasPrefix(got, "text/html") {
		t.Fatalf("content-type=%q", got)
	}
	body := res.Body.String()
	for _, want := range []string{"PVNaive", "Amir22", "50 GB", "QR", "مصرف دقیق", "data:image/png;base64,"} {
		if !strings.Contains(body, want) {
			t.Fatalf("browser page missing %q: %s", want, body)
		}
	}
	if strings.Contains(body, "attacker.example") {
		t.Fatal("untrusted request Host leaked into browser subscription page")
	}
	if !strings.Contains(body, "namir.softarg.ir") {
		t.Fatal("browser page did not use configured canonical host")
	}
	if res.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("browser subscription response is cacheable")
	}
}
