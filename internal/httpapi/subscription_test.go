package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DashSaman/PV-NaivePanel/internal/runtimecred"
	"github.com/DashSaman/PV-NaivePanel/internal/subscription"
)

type fakeSubscriptionStore struct {
	hash   [32]byte
	record subscription.Record
}

func (f *fakeSubscriptionStore) ResolveToken(_ context.Context, hash [32]byte) (subscription.Record, error) {
	if hash != f.hash {
		return subscription.Record{}, subscription.ErrUnavailable
	}
	return f.record, nil
}

func TestPublicSubscriptionReturnsNaiveURIUsingConfiguredProxyHost(t *testing.T) {
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
	store := &fakeSubscriptionStore{hash: hash, record: subscription.Record{
		RuntimeCredentialID: "runtime-1",
		Username:            "customer1",
		SecretCiphertext:    ciphertext,
		SecretNonce:         nonce,
		EncryptionKeyID:     "runtime-v1",
		UserState:           "active",
		TermState:           "pending",
	}}
	service, err := subscription.NewService(store, key, "runtime-v1")
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/subscriptions/"+rawToken, nil)
	req.Host = "attacker.example"
	res := httptest.NewRecorder()
	NewServer(ServerConfig{
		SubscriptionService:   service,
		SubscriptionProxyHost: "proxy.pvnaive.example:443",
	}).ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", res.Code, res.Body.String())
	}
	if got := res.Header().Get("Content-Type"); !strings.HasPrefix(got, "text/plain") {
		t.Fatalf("content-type=%q", got)
	}
	body := strings.TrimSpace(res.Body.String())
	if body != "naive+https://customer1:customer-secret-123@proxy.pvnaive.example:443" {
		t.Fatalf("subscription body=%q", body)
	}
	if strings.Contains(body, "attacker.example") {
		t.Fatal("request Host header influenced subscription output")
	}
	if res.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("subscription response is cacheable")
	}
}

func TestPublicSubscriptionInvalidTokenFailsWithoutSecretOracle(t *testing.T) {
	key := make([]byte, 32)
	store := &fakeSubscriptionStore{}
	service, err := subscription.NewService(store, key, "runtime-v1")
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/subscriptions/not-a-token", nil)
	res := httptest.NewRecorder()
	NewServer(ServerConfig{
		SubscriptionService:   service,
		SubscriptionProxyHost: "proxy.pvnaive.example:443",
	}).ServeHTTP(res, req)

	if res.Code != http.StatusNotFound {
		t.Fatalf("status=%d body=%s", res.Code, res.Body.String())
	}
	if strings.Contains(strings.ToLower(res.Body.String()), "decrypt") || strings.Contains(strings.ToLower(res.Body.String()), "token") {
		t.Fatalf("subscription error leaked resolution detail: %q", res.Body.String())
	}
}
