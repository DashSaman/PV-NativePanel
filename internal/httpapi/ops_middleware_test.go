package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRequestRateLimiterCapsLoginPerTrustedClientIP(t *testing.T) {
	limiter := newRequestRateLimiter()
	now := time.Date(2026, 8, 29, 18, 0, 0, 0, time.UTC)
	limiter.now = func() time.Time { return now }
	handler := limiter.middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	for i := 0; i < 12; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
		req.RemoteAddr = "127.0.0.1:5555"
		req.Header.Set("X-Forwarded-For", "203.0.113.9")
		res := httptest.NewRecorder()
		handler.ServeHTTP(res, req)
		if res.Code != http.StatusNoContent {
			t.Fatalf("request %d status=%d", i+1, res.Code)
		}
	}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
	req.RemoteAddr = "127.0.0.1:5555"
	req.Header.Set("X-Forwarded-For", "203.0.113.9")
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusTooManyRequests || res.Header().Get("Retry-After") != "60" {
		t.Fatalf("status=%d retry=%q", res.Code, res.Header().Get("Retry-After"))
	}
	now = now.Add(time.Minute)
	res = httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusNoContent {
		t.Fatalf("status after reset=%d", res.Code)
	}
}

func TestClientIPIgnoresForwardedForFromUntrustedPeer(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "198.51.100.5:1234"
	req.Header.Set("X-Forwarded-For", "203.0.113.1")
	if got := clientIP(req); got != "198.51.100.5" {
		t.Fatalf("clientIP=%q", got)
	}
}

func TestOperationalMiddlewareAddsRequestIDAndSecurityHeaders(t *testing.T) {
	handler := WithOperationalMiddleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	req := httptest.NewRequest(http.MethodGet, "/api/v1/health/live", nil)
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusNoContent {
		t.Fatalf("status=%d", res.Code)
	}
	if len(res.Header().Get("X-Request-ID")) < 16 {
		t.Fatalf("request id=%q", res.Header().Get("X-Request-ID"))
	}
	if res.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Fatalf("security headers missing: %#v", res.Header())
	}
}
