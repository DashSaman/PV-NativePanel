package httpapi

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/DashSaman/PV-NaivePanel/internal/auth"
)

func (s *server) logout(w http.ResponseWriter, r *http.Request) {
	request, ok := authenticatedFromRequest(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, envelope{"code": "authentication_required", "message": "Authentication is required."})
		return
	}
	if _, err := s.config.AuthStore.RevokeSessionByID(r.Context(), request.Bound.Tx, request.Bound.Principal.ActorID, request.Bound.Session.ID); err != nil {
		writeJSON(w, http.StatusInternalServerError, envelope{"code": "logout_failed", "message": "Logout failed."})
		return
	}
	clearAuthCookies(w)
	writeJSON(w, http.StatusOK, envelope{"status": "logged_out"})
}

func (s *server) me(w http.ResponseWriter, r *http.Request) {
	request, ok := authenticatedFromRequest(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, envelope{"code": "authentication_required", "message": "Authentication is required."})
		return
	}
	writeJSON(w, http.StatusOK, envelope{"principal": request.Bound.Principal})
}

func (s *server) sessions(w http.ResponseWriter, r *http.Request) {
	request, ok := authenticatedFromRequest(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, envelope{"code": "authentication_required", "message": "Authentication is required."})
		return
	}
	items, err := s.config.AuthStore.ListSessions(r.Context(), request.Bound.Tx, request.Bound.Principal.ActorID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, envelope{"code": "sessions_unavailable", "message": "Sessions are unavailable."})
		return
	}
	writeJSON(w, http.StatusOK, envelope{"sessions": items, "current_session_id": request.Bound.Session.ID})
}

func (s *server) deleteSession(w http.ResponseWriter, r *http.Request) {
	request, ok := authenticatedFromRequest(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, envelope{"code": "authentication_required", "message": "Authentication is required."})
		return
	}
	sessionID := strings.TrimSpace(r.PathValue("id"))
	if sessionID == "" {
		writeJSON(w, http.StatusBadRequest, envelope{"code": "invalid_request", "message": "Invalid request."})
		return
	}
	revoked, err := s.config.AuthStore.RevokeSessionByID(r.Context(), request.Bound.Tx, request.Bound.Principal.ActorID, sessionID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, envelope{"code": "session_revoke_failed", "message": "Session revoke failed."})
		return
	}
	if !revoked {
		writeJSON(w, http.StatusNotFound, envelope{"code": "session_not_found", "message": "Session not found."})
		return
	}
	if sessionID == request.Bound.Session.ID {
		clearAuthCookies(w)
	}
	writeJSON(w, http.StatusOK, envelope{"status": "revoked"})
}

func (s *server) refresh(w http.ResponseWriter, r *http.Request) {
	if s.config.AuthStore == nil {
		writeJSON(w, http.StatusUnauthorized, envelope{"code": "authentication_failed", "message": "Authentication failed."})
		return
	}
	cookie, err := r.Cookie("__Host-pvnaive_session")
	if err != nil || cookie.Value == "" {
		writeJSON(w, http.StatusUnauthorized, envelope{"code": "authentication_failed", "message": "Authentication failed."})
		return
	}
	oldHash := auth.HashOpaqueToken(cookie.Value)
	refreshMetadata, err := s.config.AuthStore.LoadRefreshSessionMetadata(r.Context(), oldHash[:])
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, envelope{"code": "authentication_failed", "message": "Authentication failed."})
		return
	}
	if err := validateCSRF(r, refreshMetadata.CSRFTokenHash); err != nil {
		writeJSON(w, http.StatusForbidden, envelope{"code": "csrf_failed", "message": "Request validation failed."})
		return
	}
	absolute := refreshMetadata.AbsoluteExpiresAt

	newRaw, newHash, err := auth.NewOpaqueToken()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, envelope{"code": "refresh_failed", "message": "Refresh failed."})
		return
	}
	csrfRaw, csrfHash, err := auth.NewOpaqueToken()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, envelope{"code": "refresh_failed", "message": "Refresh failed."})
		return
	}
	now := time.Now().UTC()
	expires := now.Add(time.Hour)
	if expires.After(absolute) {
		expires = absolute
	}
	uaHash := hashUserAgent(r.UserAgent())
	rotated, err := s.config.AuthStore.RotateSession(r.Context(), oldHash[:], newHash[:], csrfHash[:], uaHash, expires)
	if err != nil || rotated.ReuseDetected {
		clearAuthCookies(w)
		writeJSON(w, http.StatusUnauthorized, envelope{"code": "authentication_failed", "message": "Authentication failed."})
		return
	}
	setAuthCookies(w, newRaw, csrfRaw, expires)
	writeJSON(w, http.StatusOK, envelope{"status": "refreshed", "expires_at": expires, "absolute_expires_at": rotated.AbsoluteExpiresAt})
}

func (s *server) mfaEnroll(w http.ResponseWriter, r *http.Request) {
	request, ok := authenticatedFromRequest(r)
	if !ok || len(s.config.MFAKey) != 32 {
		writeJSON(w, http.StatusServiceUnavailable, envelope{"code": "mfa_unavailable", "message": "MFA is unavailable."})
		return
	}
	secret, err := auth.GenerateTOTPSecret()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, envelope{"code": "mfa_enroll_failed", "message": "MFA enrollment failed."})
		return
	}
	ciphertext, nonce, err := auth.EncryptSecret(s.config.MFAKey, []byte(secret))
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, envelope{"code": "mfa_enroll_failed", "message": "MFA enrollment failed."})
		return
	}
	if err := s.config.AuthStore.UpsertTOTPFactor(r.Context(), request.Bound.Tx, request.Bound.Principal.ActorID, ciphertext, nonce, "auth-key-v1"); err != nil {
		writeJSON(w, http.StatusInternalServerError, envelope{"code": "mfa_enroll_failed", "message": "MFA enrollment failed."})
		return
	}
	issuer := url.QueryEscape("PVNaive")
	label := url.QueryEscape("PVNaive:" + request.Bound.Principal.Email)
	uri := fmt.Sprintf("otpauth://totp/%s?secret=%s&issuer=%s&algorithm=SHA1&digits=6&period=30", label, secret, issuer)
	writeJSON(w, http.StatusOK, envelope{"status": "pending_confirmation", "secret": secret, "otpauth_uri": uri})
}

func (s *server) mfaConfirm(w http.ResponseWriter, r *http.Request) {
	request, ok := authenticatedFromRequest(r)
	if !ok || len(s.config.MFAKey) != 32 {
		writeJSON(w, http.StatusServiceUnavailable, envelope{"code": "mfa_unavailable", "message": "MFA is unavailable."})
		return
	}
	var payload struct {
		Code string `json:"code"`
	}
	if err := decodeStrictJSON(r, &payload); err != nil || payload.Code == "" {
		writeJSON(w, http.StatusBadRequest, envelope{"code": "invalid_request", "message": "Invalid request."})
		return
	}
	factor, err := s.config.AuthStore.GetTOTPFactor(r.Context(), request.Bound.Tx, request.Bound.Principal.ActorID)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, envelope{"code": "mfa_not_enrolled", "message": "MFA is not enrolled."})
		return
	}
	secret, err := auth.DecryptSecret(s.config.MFAKey, factor.Nonce, factor.Ciphertext)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, envelope{"code": "mfa_confirm_failed", "message": "MFA confirmation failed."})
		return
	}
	step, valid, err := auth.ValidateTOTP(string(secret), payload.Code, time.Now().UTC(), factor.LastUsedStep)
	for i := range secret {
		secret[i] = 0
	}
	if err != nil || !valid {
		writeJSON(w, http.StatusUnauthorized, envelope{"code": "mfa_invalid", "message": "MFA code is invalid."})
		return
	}
	codes, hashes, err := auth.GenerateRecoveryCodes(10)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, envelope{"code": "mfa_confirm_failed", "message": "MFA confirmation failed."})
		return
	}
	if err := s.config.AuthStore.ConfirmTOTPFactor(r.Context(), request.Bound.Tx, request.Bound.Principal.ActorID, step, hashes); err != nil {
		writeJSON(w, http.StatusInternalServerError, envelope{"code": "mfa_confirm_failed", "message": "MFA confirmation failed."})
		return
	}
	writeJSON(w, http.StatusOK, envelope{"status": "confirmed", "recovery_codes": codes})
}

func (s *server) mfaRemove(w http.ResponseWriter, r *http.Request) {
	request, ok := authenticatedFromRequest(r)
	if !ok || len(s.config.MFAKey) != 32 {
		writeJSON(w, http.StatusServiceUnavailable, envelope{"code": "mfa_unavailable", "message": "MFA is unavailable."})
		return
	}
	var payload struct {
		Password     string `json:"password"`
		TOTPCode     string `json:"totp_code"`
		RecoveryCode string `json:"recovery_code"`
	}
	if err := decodeStrictJSON(r, &payload); err != nil || payload.Password == "" || (payload.TOTPCode == "" && payload.RecoveryCode == "") {
		writeJSON(w, http.StatusBadRequest, envelope{"code": "invalid_request", "message": "Invalid request."})
		return
	}
	passwordHash, err := s.config.AuthStore.GetActorPasswordHash(r.Context(), request.Bound.Tx, request.Bound.Principal.ActorID)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, envelope{"code": "authentication_failed", "message": "Authentication failed."})
		return
	}
	passwordOK, err := auth.VerifyPassword(payload.Password, passwordHash)
	if err != nil || !passwordOK {
		writeJSON(w, http.StatusUnauthorized, envelope{"code": "authentication_failed", "message": "Authentication failed."})
		return
	}
	verified := false
	if payload.RecoveryCode != "" {
		hash := auth.HashRecoveryCode(payload.RecoveryCode)
		verified, err = s.config.AuthStore.ConsumeRecoveryCode(r.Context(), request.Bound.Tx, request.Bound.Principal.ActorID, hash)
	} else {
		factor, factorErr := s.config.AuthStore.GetTOTPFactor(r.Context(), request.Bound.Tx, request.Bound.Principal.ActorID)
		if factorErr == nil {
			secret, decryptErr := auth.DecryptSecret(s.config.MFAKey, factor.Nonce, factor.Ciphertext)
			if decryptErr == nil {
				step, valid, validateErr := auth.ValidateTOTP(string(secret), payload.TOTPCode, time.Now().UTC(), factor.LastUsedStep)
				for i := range secret {
					secret[i] = 0
				}
				if validateErr == nil && valid {
					verified, err = s.config.AuthStore.ConsumeTOTPStep(r.Context(), request.Bound.Tx, request.Bound.Principal.ActorID, step)
				}
			}
		}
	}
	if err != nil || !verified {
		writeJSON(w, http.StatusUnauthorized, envelope{"code": "authentication_failed", "message": "Authentication failed."})
		return
	}
	if err := s.config.AuthStore.RemoveMFA(r.Context(), request.Bound.Tx, request.Bound.Principal.ActorID); err != nil {
		writeJSON(w, http.StatusInternalServerError, envelope{"code": "mfa_remove_failed", "message": "MFA removal failed."})
		return
	}
	if _, err := s.config.AuthStore.RevokeOtherActorSessions(r.Context(), request.Bound.Tx, request.Bound.Principal.ActorID, request.Bound.Session.ID); err != nil {
		writeJSON(w, http.StatusInternalServerError, envelope{"code": "mfa_remove_failed", "message": "MFA removal failed."})
		return
	}
	newRaw, newHash, err := auth.NewOpaqueToken()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, envelope{"code": "mfa_remove_failed", "message": "MFA removal failed."})
		return
	}
	csrfRaw, csrfHash, err := auth.NewOpaqueToken()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, envelope{"code": "mfa_remove_failed", "message": "MFA removal failed."})
		return
	}
	oldHash := auth.HashOpaqueToken(request.RawSessionToken)
	expires := time.Now().UTC().Add(time.Hour)
	if expires.After(request.Bound.Session.AbsoluteExpiresAt) {
		expires = request.Bound.Session.AbsoluteExpiresAt
	}
	rotated, err := s.config.AuthStore.RotateSessionTx(r.Context(), request.Bound.Tx, oldHash[:], newHash[:], csrfHash[:], hashUserAgent(r.UserAgent()), expires)
	if err != nil || rotated.ReuseDetected {
		writeJSON(w, http.StatusInternalServerError, envelope{"code": "mfa_remove_failed", "message": "MFA removal failed."})
		return
	}
	setAuthCookies(w, newRaw, csrfRaw, expires)
	writeJSON(w, http.StatusOK, envelope{"status": "removed", "sessions_revoked": true})
}

func decodeStrictJSON(r *http.Request, dst any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(dst)
}

func setAuthCookies(w http.ResponseWriter, sessionRaw, csrfRaw string, expires time.Time) {
	session := newSessionCookie(sessionRaw)
	session.Expires = expires
	csrf := newCSRFCookie(csrfRaw)
	csrf.Expires = expires
	http.SetCookie(w, session)
	http.SetCookie(w, csrf)
}

func clearAuthCookies(w http.ResponseWriter) {
	for _, cookie := range []*http.Cookie{newSessionCookie(""), newCSRFCookie("")} {
		cookie.MaxAge = -1
		cookie.Expires = time.Unix(1, 0).UTC()
		http.SetCookie(w, cookie)
	}
}

func hashUserAgent(userAgent string) []byte {
	if userAgent == "" {
		return nil
	}
	sum := sha256.Sum256([]byte(userAgent))
	return sum[:]
}

var _ = errors.New
