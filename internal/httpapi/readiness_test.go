package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DashSaman/PV-NaivePanel/internal/auth"
)

func configuredServer(probe ReadinessProbeFunc, timeout time.Duration) *server {
	return &server{config: ServerConfig{
		AuthService: &auth.Service{}, AuthStore: &auth.Store{}, MFAKey: make([]byte, 32),
		ReadinessProbe: probe, ReadyTimeout: timeout,
	}}
}

func decodeReady(t *testing.T, rec *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var body map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	return body
}

func TestReadyConfiguredProbeOK(t *testing.T) {
	s := configuredServer(func(context.Context) error { return nil }, 0)
	rec := httptest.NewRecorder()
	s.ready(rec, httptest.NewRequest(http.MethodGet, "/api/v1/health/ready", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	body := decodeReady(t, rec)
	if body["ready"] != true || body["status"] != "ready" || body["db"] != "ok" || body["schema"] != "ok" {
		t.Fatalf("body=%#v", body)
	}
}

func TestReadyConfiguredMissingProbeFailsClosed(t *testing.T) {
	s := configuredServer(nil, 0)
	rec := httptest.NewRecorder()
	s.ready(rec, httptest.NewRequest(http.MethodGet, "/api/v1/health/ready", nil))
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if body := decodeReady(t, rec); body["ready"] != false || body["status"] != "not_ready" {
		t.Fatalf("body=%#v", body)
	}
}

func TestReadyProbeFailureAndSchemaMismatchFailClosedWithoutLeak(t *testing.T) {
	for _, errValue := range []error{errors.New("postgres password=secret connection refused"), errors.New("schema have=15 want=16")} {
		s := configuredServer(func(context.Context) error { return errValue }, 0)
		rec := httptest.NewRecorder()
		s.ready(rec, httptest.NewRequest(http.MethodGet, "/api/v1/health/ready", nil))
		if rec.Code != http.StatusServiceUnavailable {
			t.Fatalf("status=%d", rec.Code)
		}
		if got := rec.Body.String(); got == "" || containsAny(got, []string{"password", "secret", "have=15", "want=16", "connection refused"}) {
			t.Fatalf("leaked body=%q", got)
		}
	}
}

func TestReadyProbeTimeoutIsBounded(t *testing.T) {
	probe := func(ctx context.Context) error { <-ctx.Done(); return ctx.Err() }
	s := configuredServer(probe, 25*time.Millisecond)
	rec := httptest.NewRecorder()
	start := time.Now()
	s.ready(rec, httptest.NewRequest(http.MethodGet, "/api/v1/health/ready", nil))
	elapsed := time.Since(start)
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status=%d", rec.Code)
	}
	if elapsed > 250*time.Millisecond {
		t.Fatalf("elapsed=%v", elapsed)
	}
}

func TestReadyZeroConfigRemainsScaffoldForIsolatedTests(t *testing.T) {
	s := &server{}
	rec := httptest.NewRecorder()
	s.ready(rec, httptest.NewRequest(http.MethodGet, "/api/v1/health/ready", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d", rec.Code)
	}
	if body := decodeReady(t, rec); body["ready"] != false || body["status"] != "scaffold" {
		t.Fatalf("body=%#v", body)
	}
}

func TestLiveUnaffectedByReadiness(t *testing.T) {
	rec := httptest.NewRecorder()
	live(rec, httptest.NewRequest(http.MethodGet, "/api/v1/health/live", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d", rec.Code)
	}
}

func containsAny(s string, xs []string) bool {
	for _, x := range xs {
		if len(x) > 0 && strings.Contains(s, x) {
			return true
		}
	}
	return false
}
