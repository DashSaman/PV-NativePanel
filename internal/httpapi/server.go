package httpapi

import (
	"encoding/json"
	"net/http"
)

type envelope map[string]any

func NewServer() http.Handler {
	mux := http.NewServeMux()
	for _, route := range Routes {
		route := route
		handler := http.HandlerFunc(notImplemented)
		if route.Name == "health.live" {
			handler = http.HandlerFunc(live)
		}
		if route.Name == "health.ready" {
			handler = http.HandlerFunc(ready)
		}
		var wrapped http.Handler = handler
		if route.Access != Public {
			wrapped = requireAuthentication(wrapped)
		}
		mux.Handle(route.Method+" "+route.Path, wrapped)
	}
	return securityHeaders(limitBody(mux))
}

func live(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, envelope{"status": "ok", "service": "pvnaive-api"})
}

func ready(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, envelope{"status": "scaffold", "ready": false})
}

func notImplemented(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusNotImplemented, envelope{"code": "not_implemented", "message": "This endpoint is not available yet."})
}

func requireAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusUnauthorized, envelope{"code": "authentication_required", "message": "Authentication is required."})
	})
}

func limitBody(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		next.ServeHTTP(w, r)
	})
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "no-referrer")
		h.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		h.Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'; base-uri 'none'")
		h.Set("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, payload envelope) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
