package httpapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DashSaman/PV-NaivePanel/internal/runtimecred"
	"github.com/DashSaman/PV-NaivePanel/internal/subscription"
	"github.com/DashSaman/PV-NaivePanel/internal/telemetry"
)

type accountPageAccountingStore struct {
	model telemetry.ReadModel
	err   error
	term  string
}

func (s *accountPageAccountingStore) Authorize(context.Context, string, time.Time) (telemetry.AuthorizationResult, error) {
	return telemetry.AuthorizationResult{}, errors.New("unexpected authorize")
}
func (s *accountPageAccountingStore) Ingest(context.Context, telemetry.Event) (telemetry.IngestResult, error) {
	return telemetry.IngestResult{}, errors.New("unexpected ingest")
}
func (s *accountPageAccountingStore) Claim(context.Context, telemetry.ClaimRequest) (telemetry.ClaimResult, error) {
	return telemetry.ClaimResult{}, errors.New("unexpected claim")
}
func (s *accountPageAccountingStore) Read(_ context.Context, serviceTermID string, _ time.Time, staleAfter time.Duration) (telemetry.ReadModel, error) {
	s.term = serviceTermID
	if staleAfter != customerAccountingStaleAfter {
		return telemetry.ReadModel{}, errors.New("unexpected stale window")
	}
	return s.model, s.err
}

func accountPageAccountingFixture(t *testing.T, baseline subscription.AccountingBaseline, quota int64, model telemetry.ReadModel) (http.Handler, string, *accountPageAccountingStore) {
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
	expires := time.Now().UTC().Add(30 * 24 * time.Hour)
	termID := "11111111-1111-4111-8111-111111111111"
	store := &fakeSubscriptionStore{hash: hash, record: subscription.Record{
		ServiceTermID:       termID,
		RuntimeCredentialID: "22222222-2222-4222-8222-222222222222",
		Username:            "Amir22",
		SecretCiphertext:    ciphertext,
		SecretNonce:         nonce,
		EncryptionKeyID:     "runtime-v1",
		UserState:           "active",
		TermState:           "active",
		QuotaBytes:          &quota,
		StartPolicy:         "on_creation",
		ExpiresAt:           &expires,
		AccountingBaseline:  baseline,
	}}
	service, err := subscription.NewService(store, key, "runtime-v1")
	if err != nil {
		t.Fatal(err)
	}
	accounting := &accountPageAccountingStore{model: model}
	return NewServer(ServerConfig{
		SubscriptionService:   service,
		SubscriptionProxyHost: "namir.softarg.ir:443",
		AccountingStore:       accounting,
	}), rawToken, accounting
}

func TestAccountPageShowsPresenceButNotLifetimeUsageWhenLegacyBaselineUnknown(t *testing.T) {
	cutoff := time.Date(2026, 8, 30, 12, 0, 0, 0, time.UTC)
	last := time.Date(2026, 8, 30, 12, 34, 0, 0, time.UTC)
	quota := int64(10 * 1024 * 1024 * 1024)
	handler, token, accounting := accountPageAccountingFixture(t, subscription.AccountingBaseline{
		State: "unknown", Source: "legacy_unavailable", CutoffAt: cutoff,
	}, quota, telemetry.ReadModel{
		ServiceTermID:      "11111111-1111-4111-8111-111111111111",
		UploadBytes:        1 * 1024 * 1024 * 1024,
		DownloadBytes:      2 * 1024 * 1024 * 1024,
		UsedBytes:          3 * 1024 * 1024 * 1024,
		LastOnline:         &last,
		Online:             true,
		SessionCount:       1,
		AccountingComplete: true,
	})
	req := httptest.NewRequest(http.MethodGet, "/s/"+token+"?lang=en", nil)
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", res.Code, res.Body.String())
	}
	body := res.Body.String()
	for _, want := range []string{"Online", "2026-08-30 12:34 UTC", "Usage unavailable"} {
		if !strings.Contains(body, want) {
			t.Fatalf("missing %q in body", want)
		}
	}
	if accounting.term != "11111111-1111-4111-8111-111111111111" {
		t.Fatalf("accounting term=%q", accounting.term)
	}
	if strings.Contains(body, ">3 GB<") || strings.Contains(body, ">7 GB<") {
		t.Fatal("legacy unknown baseline exposed fabricated lifetime usage/remaining")
	}
}

func TestAccountPageShowsExactTrafficTotalsForKnownBaseline(t *testing.T) {
	cutoff := time.Date(2026, 8, 30, 12, 0, 0, 0, time.UTC)
	last := time.Date(2026, 8, 30, 12, 35, 0, 0, time.UTC)
	zero := int64(0)
	quota := int64(10 * 1024 * 1024 * 1024)
	handler, token, _ := accountPageAccountingFixture(t, subscription.AccountingBaseline{
		State: "known", Source: "fresh_managed_term", CutoffAt: cutoff, UploadBytes: &zero, DownloadBytes: &zero,
	}, quota, telemetry.ReadModel{
		ServiceTermID:      "11111111-1111-4111-8111-111111111111",
		UploadBytes:        1 * 1024 * 1024 * 1024,
		DownloadBytes:      2 * 1024 * 1024 * 1024,
		UsedBytes:          3 * 1024 * 1024 * 1024,
		LastOnline:         &last,
		Online:             false,
		AccountingComplete: true,
	})
	req := httptest.NewRequest(http.MethodGet, "/s/"+token+"?lang=en", nil)
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", res.Code, res.Body.String())
	}
	body := res.Body.String()
	for _, want := range []string{"Upload", "Download", ">1 GB<", ">2 GB<", ">3 GB<", ">7 GB<", "Offline", "2026-08-30 12:35 UTC"} {
		if !strings.Contains(body, want) {
			t.Fatalf("missing %q in body", want)
		}
	}
}

func TestAccountPageDoesNotFabricateOfflineWhenAccountingIsIncomplete(t *testing.T) {
	cutoff := time.Date(2026, 8, 30, 12, 0, 0, 0, time.UTC)
	last := time.Date(2026, 8, 30, 12, 36, 0, 0, time.UTC)
	zero := int64(0)
	quota := int64(10 * 1024 * 1024 * 1024)
	handler, token, _ := accountPageAccountingFixture(t, subscription.AccountingBaseline{
		State: "known", Source: "fresh_managed_term", CutoffAt: cutoff, UploadBytes: &zero, DownloadBytes: &zero,
	}, quota, telemetry.ReadModel{
		ServiceTermID:      "11111111-1111-4111-8111-111111111111",
		UploadBytes:        1 * 1024 * 1024 * 1024,
		DownloadBytes:      2 * 1024 * 1024 * 1024,
		UsedBytes:          3 * 1024 * 1024 * 1024,
		LastOnline:         &last,
		Online:             false,
		AccountingComplete: false,
	})
	req := httptest.NewRequest(http.MethodGet, "/s/"+token+"?lang=en", nil)
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", res.Code, res.Body.String())
	}
	body := res.Body.String()
	for _, want := range []string{"Usage unavailable", "2026-08-30 12:36 UTC", "Unavailable"} {
		if !strings.Contains(body, want) {
			t.Fatalf("missing %q in body", want)
		}
	}
	if strings.Contains(body, "<strong>Offline</strong>") || strings.Contains(body, "<strong>Online</strong>") {
		t.Fatal("incomplete accounting fabricated presence state")
	}
}

func TestAccountPageShowsExactCurrentPeriodAfterLegacyUsageReset(t *testing.T) {
	quota := int64(10 * 1024 * 1024 * 1024)
	resetAt := time.Date(2026, 8, 30, 12, 0, 0, 0, time.UTC)
	handler, token, _ := accountPageAccountingFixture(t, subscription.AccountingBaseline{
		State: "unknown", Source: "legacy_unavailable",
		CutoffAt: time.Date(2026, 8, 29, 12, 0, 0, 0, time.UTC),
	}, quota, telemetry.ReadModel{
		ServiceTermID: "11111111-1111-4111-8111-111111111111",
		UploadBytes:   1 * 1024 * 1024 * 1024, DownloadBytes: 2 * 1024 * 1024 * 1024, UsedBytes: 3 * 1024 * 1024 * 1024, QuotaBytes: &quota,
		AccountingComplete: true, LastResetAt: &resetAt,
	})

	req := httptest.NewRequest(http.MethodGet, "/s/"+token+"?lang=en", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	body := rr.Body.String()
	for _, want := range []string{">3 GB<", ">1 GB<", ">2 GB<", ">7 GB<"} {
		if !strings.Contains(body, want) {
			t.Fatalf("reset-period account page missing %q: %s", want, body)
		}
	}
}
