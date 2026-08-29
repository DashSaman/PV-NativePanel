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

func TestPublicAccountPageUsesExplicitHumanEndpoint(t *testing.T) {
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
	expires := time.Now().UTC().Add(30 * 24 * time.Hour)
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

	req := httptest.NewRequest(http.MethodGet, "/s/"+rawToken+"?lang=fa", nil)
	req.Host = "attacker.example"
	req.Header.Set("Accept", "text/plain")
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
	for _, want := range []string{"PVNaive", "PVNETWORK", "Amir22", "50 GB", "در دسترس نیست", "data:image/png;base64,", "/sub/" + rawToken, `class="account-shell"`, `class="service-grid"`, `class="connect-card"`, `class="account-nav"`} {
		if !strings.Contains(body, want) {
			t.Fatalf("account page missing %q: %s", want, body)
		}
	}
	if strings.Contains(body, "attacker.example") {
		t.Fatal("untrusted request Host leaked into account page")
	}
	if !strings.Contains(body, "namir.softarg.ir") {
		t.Fatal("account page did not use configured canonical host")
	}
	if res.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("account page response is cacheable")
	}
}

func TestAccountLanguageDefaultsToPersianUnlessExplicitlyEnglish(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/s/example", nil)
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	if got := accountLanguage(req); got != "fa" {
		t.Fatalf("default language=%q want fa", got)
	}
	req = httptest.NewRequest(http.MethodGet, "/s/example?lang=en", nil)
	if got := accountLanguage(req); got != "en" {
		t.Fatalf("explicit English language=%q want en", got)
	}
}
