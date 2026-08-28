package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/DashSaman/PV-NaivePanel/internal/auth"
	"github.com/DashSaman/PV-NaivePanel/internal/runtimecred"
	"github.com/DashSaman/PV-NaivePanel/internal/subscription"
)

type envelope map[string]any

type ServerConfig struct {
	AuthService           *auth.Service
	AuthStore             *auth.Store
	MFAKey                []byte
	RuntimeService        *runtimecred.Service
	SubscriptionService   *subscription.Service
	SubscriptionProxyHost string
}

type server struct {
	config ServerConfig
}

func NewServer(configs ...ServerConfig) http.Handler {
	var cfg ServerConfig
	if len(configs) > 0 {
		cfg = configs[0]
	}
	s := &server{config: cfg}
	mux := http.NewServeMux()
	for _, route := range Routes {
		route := route
		var handler http.Handler = http.HandlerFunc(notImplemented)
		switch route.Name {
		case "health.live":
			handler = http.HandlerFunc(live)
		case "health.ready":
			handler = http.HandlerFunc(s.ready)
		case "auth.login":
			if cfg.AuthService != nil {
				handler = http.HandlerFunc(s.login)
			}
		case "auth.refresh":
			if cfg.AuthStore != nil {
				handler = http.HandlerFunc(s.refresh)
			}
		case "auth.logout":
			if cfg.AuthStore != nil {
				handler = http.HandlerFunc(s.logout)
			}
		case "subscriptions.show":
			if cfg.SubscriptionService != nil && cfg.SubscriptionProxyHost != "" {
				handler = http.HandlerFunc(s.publicSubscription)
			}
		case "me.show":
			if cfg.AuthStore != nil {
				handler = http.HandlerFunc(s.me)
			}
		case "me.sessions.index":
			if cfg.AuthStore != nil {
				handler = http.HandlerFunc(s.sessions)
			}
		case "me.sessions.delete":
			if cfg.AuthStore != nil {
				handler = http.HandlerFunc(s.deleteSession)
			}
		case "me.mfa.enroll":
			if cfg.AuthStore != nil && len(cfg.MFAKey) == 32 {
				handler = http.HandlerFunc(s.mfaEnroll)
			}
		case "me.mfa.confirm":
			if cfg.AuthStore != nil && len(cfg.MFAKey) == 32 {
				handler = http.HandlerFunc(s.mfaConfirm)
			}
		case "me.mfa.delete":
			if cfg.AuthStore != nil && len(cfg.MFAKey) == 32 {
				handler = http.HandlerFunc(s.mfaRemove)
			}
		case "runtime.naive.show":
			if cfg.RuntimeService != nil {
				handler = http.HandlerFunc(s.runtimeNaiveStatus)
			}
		case "runtime.naive.credentials.index":
			if cfg.RuntimeService != nil {
				handler = http.HandlerFunc(s.runtimeNaiveCredentials)
			}
		case "runtime.naive.import":
			if cfg.RuntimeService != nil {
				handler = http.HandlerFunc(s.runtimeNaiveImport)
			}
		case "runtime.naive.credentials.create":
			if cfg.RuntimeService != nil {
				handler = http.HandlerFunc(s.runtimeNaiveCreateCredential)
			}
		case "runtime.naive.credentials.update":
			if cfg.RuntimeService != nil {
				handler = http.HandlerFunc(s.runtimeNaiveUpdateCredential)
			}
		case "runtime.naive.credentials.rotate":
			if cfg.RuntimeService != nil {
				handler = http.HandlerFunc(s.runtimeNaiveRotateCredential)
			}
		case "runtime.naive.credentials.revoke":
			if cfg.RuntimeService != nil {
				handler = http.HandlerFunc(s.runtimeNaiveRevokeCredential)
			}
		}
		if route.Access != Public {
			handler = s.requireAuthentication(route, handler)
		}
		mux.Handle(route.Method+" "+route.Path, handler)
	}
	return securityHeaders(limitBody(mux))
}

func live(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, envelope{"status": "ok", "service": "pvnaive-api"})
}

func (s *server) ready(w http.ResponseWriter, _ *http.Request) {
	ready := s.config.AuthService != nil && s.config.AuthStore != nil && len(s.config.MFAKey) == 32
	status := "scaffold"
	if ready {
		status = "ready"
	}
	writeJSON(w, http.StatusOK, envelope{"status": status, "ready": ready})
}

func notImplemented(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusNotImplemented, envelope{"code": "not_implemented", "message": "This endpoint is not available yet."})
}

func (s *server) login(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		TOTPCode string `json:"totp_code"`
	}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, envelope{"code": "invalid_request", "message": "Invalid request."})
		return
	}
	result, err := s.config.AuthService.Login(r.Context(), auth.LoginInput{
		Email: payload.Email, Password: payload.Password, TOTPCode: payload.TOTPCode, UserAgent: r.UserAgent(),
	})
	if err != nil {
		if errors.Is(err, auth.ErrMFARequired) {
			writeJSON(w, http.StatusUnauthorized, envelope{"code": "mfa_required", "message": "Additional authentication is required."})
			return
		}
		if errors.Is(err, auth.ErrInvalidCredentials) {
			writeJSON(w, http.StatusUnauthorized, envelope{"code": "authentication_failed", "message": "Authentication failed."})
			return
		}
		writeJSON(w, http.StatusInternalServerError, envelope{"code": "authentication_unavailable", "message": "Authentication is unavailable."})
		return
	}
	setAuthCookies(w, result.SessionToken, result.CSRFToken, result.ExpiresAt)
	writeJSON(w, http.StatusOK, envelope{
		"status": "authenticated", "actor_id": result.ActorID, "role": result.Role,
		"expires_at": result.ExpiresAt, "absolute_expires_at": result.AbsoluteExpiresAt,
	})
}

func (s *server) requireAuthentication(route Route, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.config.AuthStore == nil {
			writeJSON(w, http.StatusUnauthorized, envelope{"code": "authentication_required", "message": "Authentication is required."})
			return
		}
		cookie, err := r.Cookie("__Host-pvnaive_session")
		if err != nil || cookie.Value == "" {
			writeJSON(w, http.StatusUnauthorized, envelope{"code": "authentication_required", "message": "Authentication is required."})
			return
		}
		hash := auth.HashOpaqueToken(cookie.Value)
		bound, err := s.config.AuthStore.BeginAuthenticated(r.Context(), hash[:])
		if err != nil {
			writeJSON(w, http.StatusUnauthorized, envelope{"code": "authentication_required", "message": "Authentication is required."})
			return
		}
		defer bound.Tx.Rollback()
		if !roleAllowed(route.Access, bound.Principal.Role) {
			writeJSON(w, http.StatusForbidden, envelope{"code": "forbidden", "message": "Access denied."})
			return
		}
		if err := validateCSRF(r, bound.Session.CSRFTokenHash); err != nil {
			writeJSON(w, http.StatusForbidden, envelope{"code": "csrf_failed", "message": "Request validation failed."})
			return
		}
		r = withAuthenticatedRequest(r, bound, cookie.Value)
		next.ServeHTTP(w, r)
		if authenticated, ok := authenticatedFromRequest(r); ok && authenticated.TransactionFinalized {
			return
		}
		_ = bound.Tx.Commit()
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
