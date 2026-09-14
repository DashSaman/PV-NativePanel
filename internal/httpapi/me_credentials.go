package httpapi

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/DashSaman/PV-NaivePanel/internal/auth"
)

const (
	mePasswordMinBytes  = 14
	mePasswordMaxBytes  = 1024
	meEmailMinBytes     = 3
	meEmailMaxBytes     = 320
	meDisplayNameMaxLen = 160
)

type mePasswordPayload struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

type meProfilePayload struct {
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
}

// validateMePasswordChange enforces the same policy as the
// pvnaive-password helper: 14..1024 bytes, current password must be supplied.
func validateMePasswordChange(current, next string) error {
	if current == "" {
		return errors.New("current password is required")
	}
	if n := len(next); n < mePasswordMinBytes {
		return errors.New("new password must be at least 14 characters")
	} else if n > mePasswordMaxBytes {
		return errors.New("new password must be at most 1024 characters")
	}
	return nil
}

// validateMeProfileUpdate accepts a change to the login email and/or the
// display name; both empty means "nothing to update".
func validateMeProfileUpdate(email, displayName string) error {
	if email == "" && displayName == "" {
		return errors.New("nothing to update")
	}
	if email != "" {
		if n := len(email); n < meEmailMinBytes || n > meEmailMaxBytes {
			return errors.New("email length is outside policy")
		}
		if strings.ContainsAny(email, " \t\r\n") {
			return errors.New("email must not contain whitespace")
		}
		if strings.Count(email, "@") != 1 {
			return errors.New("email must contain exactly one @")
		}
		local, domain, _ := strings.Cut(email, "@")
		if local == "" || domain == "" || !strings.Contains(domain, ".") {
			return errors.New("email domain is invalid")
		}
	}
	if len(displayName) > meDisplayNameMaxLen {
		return errors.New("display name is too long")
	}
	return nil
}

// mePasswordUpdate lets the authenticated actor rotate their own login
// password. The current password is verified, other sessions are revoked and
// the current session is rotated so the caller stays signed in.
func (s *server) mePasswordUpdate(w http.ResponseWriter, r *http.Request) {
	request, ok := authenticatedFromRequest(r)
	if !ok || s.config.AuthStore == nil {
		writeJSON(w, http.StatusServiceUnavailable, envelope{"code": "me_password_unavailable", "message": "Password change is unavailable."})
		return
	}
	var payload mePasswordPayload
	if err := decodeStrictJSON(r, &payload); err != nil || validateMePasswordChange(payload.CurrentPassword, payload.NewPassword) != nil {
		writeJSON(w, http.StatusBadRequest, envelope{"code": "invalid_request", "message": "Invalid request. Provide the current password and a new password of 14-1024 characters."})
		return
	}
	currentHash, err := s.config.AuthStore.GetActorPasswordHash(r.Context(), request.Bound.Tx, request.Bound.Principal.ActorID)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, envelope{"code": "authentication_failed", "message": "Authentication failed."})
		return
	}
	currentOK, err := auth.VerifyPassword(payload.CurrentPassword, currentHash)
	if err != nil || !currentOK {
		writeJSON(w, http.StatusUnauthorized, envelope{"code": "authentication_failed", "message": "Authentication failed."})
		return
	}
	newHash, err := auth.HashPassword(payload.NewPassword)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, envelope{"code": "me_password_update_failed", "message": "Password update failed."})
		return
	}
	if err := s.config.AuthStore.UpdateActorPasswordHash(r.Context(), request.Bound.Tx, request.Bound.Principal.ActorID, newHash); err != nil {
		writeJSON(w, http.StatusInternalServerError, envelope{"code": "me_password_update_failed", "message": "Password update failed."})
		return
	}
	if _, err := s.config.AuthStore.RevokeOtherActorSessions(r.Context(), request.Bound.Tx, request.Bound.Principal.ActorID, request.Bound.Session.ID); err != nil {
		writeJSON(w, http.StatusInternalServerError, envelope{"code": "me_password_update_failed", "message": "Password update failed."})
		return
	}
	newRaw, newTokenHash, err := auth.NewOpaqueToken()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, envelope{"code": "me_password_update_failed", "message": "Password update failed."})
		return
	}
	csrfRaw, csrfHash, err := auth.NewOpaqueToken()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, envelope{"code": "me_password_update_failed", "message": "Password update failed."})
		return
	}
	oldHash := auth.HashOpaqueToken(request.RawSessionToken)
	expires := time.Now().UTC().Add(time.Hour)
	if expires.After(request.Bound.Session.AbsoluteExpiresAt) {
		expires = request.Bound.Session.AbsoluteExpiresAt
	}
	rotated, err := s.config.AuthStore.RotateSessionTx(r.Context(), request.Bound.Tx, oldHash[:], newTokenHash[:], csrfHash[:], hashUserAgent(r.UserAgent()), expires)
	if err != nil || rotated.ReuseDetected {
		writeJSON(w, http.StatusInternalServerError, envelope{"code": "me_password_update_failed", "message": "Password update failed."})
		return
	}
	setAuthCookies(w, newRaw, csrfRaw, expires)
	_ = s.config.AuthStore.AppendAudit(r.Context(), &request.Bound.Principal.ActorID, "me.password.rotate", "success", "")
	writeJSON(w, http.StatusOK, envelope{"status": "updated", "sessions_revoked": true})
}

// meProfileUpdate lets the authenticated actor change their login email
// and/or display name. A unique-violation on the lower(email) index maps to
// HTTP 409.
func (s *server) meProfileUpdate(w http.ResponseWriter, r *http.Request) {
	request, ok := authenticatedFromRequest(r)
	if !ok || s.config.AuthStore == nil {
		writeJSON(w, http.StatusServiceUnavailable, envelope{"code": "me_profile_unavailable", "message": "Profile update is unavailable."})
		return
	}
	var payload meProfilePayload
	if err := decodeStrictJSON(r, &payload); err != nil || validateMeProfileUpdate(payload.Email, payload.DisplayName) != nil {
		writeJSON(w, http.StatusBadRequest, envelope{"code": "invalid_request", "message": "Invalid request. Provide a valid email and/or a display name of at most 160 characters."})
		return
	}
	email, displayName, err := s.config.AuthStore.UpdateActorProfile(r.Context(), request.Bound.Tx, request.Bound.Principal.ActorID, strings.ToLower(strings.TrimSpace(payload.Email)), strings.TrimSpace(payload.DisplayName))
	if err != nil {
		var pgErr interface{ SQLState() string }
		if errors.As(err, &pgErr) && pgErr.SQLState() == "23505" {
			writeJSON(w, http.StatusConflict, envelope{"code": "email_conflict", "message": "This email is already in use."})
			return
		}
		writeJSON(w, http.StatusInternalServerError, envelope{"code": "me_profile_update_failed", "message": "Profile update failed."})
		return
	}
	_ = s.config.AuthStore.AppendAudit(r.Context(), &request.Bound.Principal.ActorID, "me.profile.update", "success", "")
	writeJSON(w, http.StatusOK, envelope{"status": "updated", "email": email, "display_name": displayName})
}
