package httpapi

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/DashSaman/PV-NaivePanel/internal/observability"
)

type responseStatusRecorder struct {
	http.ResponseWriter
	status int
}

func (w *responseStatusRecorder) WriteHeader(status int) {
	if w.status == 0 {
		w.status = status
	}
	w.ResponseWriter.WriteHeader(status)
}

func (w *responseStatusRecorder) Write(body []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	return w.ResponseWriter.Write(body)
}

func requestInstrumentation(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := newRequestID()
		w.Header().Set("X-Request-ID", requestID)
		start := time.Now()
		recorder := &responseStatusRecorder{ResponseWriter: w}
		next.ServeHTTP(recorder, r)
		status := recorder.status
		if status == 0 {
			status = http.StatusOK
		}
		encoded, err := observability.MarshalLog(observability.Event{
			Timestamp: start,
			Level:     "info",
			RequestID: requestID,
			Component: "http",
			Message:   "request completed",
			Fields: map[string]any{
				"method":      r.Method,
				"path":        r.URL.Path,
				"status":      status,
				"duration_ms": time.Since(start).Milliseconds(),
				"client_ip":   clientIP(r),
			},
		})
		if err == nil {
			log.Print(string(encoded))
		}
	})
}

func newRequestID() string {
	var raw [12]byte
	if _, err := rand.Read(raw[:]); err == nil {
		return hex.EncodeToString(raw[:])
	}
	return time.Now().UTC().Format("20060102T150405.000000000")
}

type rateWindow struct {
	start time.Time
	count int
}

type requestRateLimiter struct {
	mu      sync.Mutex
	windows map[string]rateWindow
	now     func() time.Time
}

func newRequestRateLimiter() *requestRateLimiter {
	return &requestRateLimiter{windows: make(map[string]rateWindow), now: time.Now}
}

func (l *requestRateLimiter) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		class, limit := rateClass(r)
		if limit > 0 && !l.allow(class+":"+clientIP(r), limit, time.Minute) {
			w.Header().Set("Retry-After", "60")
			writeJSON(w, http.StatusTooManyRequests, envelope{"code": "rate_limited", "message": "Too many requests. Try again shortly."})
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (l *requestRateLimiter) allow(key string, limit int, duration time.Duration) bool {
	if l == nil || limit <= 0 {
		return true
	}
	now := l.now().UTC()
	l.mu.Lock()
	defer l.mu.Unlock()
	window := l.windows[key]
	if window.start.IsZero() || now.Sub(window.start) >= duration || now.Before(window.start) {
		window = rateWindow{start: now}
	}
	if window.count >= limit {
		l.windows[key] = window
		return false
	}
	window.count++
	l.windows[key] = window
	if len(l.windows) > 4096 {
		for storedKey, stored := range l.windows {
			if now.Sub(stored.start) >= 10*duration || now.Before(stored.start) {
				delete(l.windows, storedKey)
			}
		}
	}
	return true
}

func rateClass(r *http.Request) (string, int) {
	path := r.URL.Path
	switch {
	case r.Method == http.MethodPost && path == "/api/v1/auth/login":
		return "login", 12
	case strings.HasPrefix(path, "/sub/"), strings.HasPrefix(path, "/s/"), strings.HasPrefix(path, "/api/v1/subscriptions/"):
		return "subscription", 180
	default:
		return "api", 900
	}
}

func clientIP(r *http.Request) string {
	peerHost, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		peerHost = r.RemoteAddr
	}
	peer := net.ParseIP(strings.TrimSpace(peerHost))
	if peer != nil && peer.IsLoopback() {
		if forwarded := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); forwarded != "" {
			first := strings.TrimSpace(strings.Split(forwarded, ",")[0])
			if parsed := net.ParseIP(first); parsed != nil {
				return parsed.String()
			}
		}
	}
	if peer != nil {
		return peer.String()
	}
	return "unknown"
}
