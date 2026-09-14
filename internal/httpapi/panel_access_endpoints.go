package httpapi

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/DashSaman/PV-NaivePanel/internal/auth"
)

// R7 panel access endpoints (ACCESS-001..004).
//   GET /api/v1/panel-access — Owner: effective settings (secrets never returned).
//   PUT /api/v1/panel-access — Owner: step-up protected (current password),
//     validates everything before touching state; every change is audited by
//     panel_settings_upsert with redacted secret markers.

type panelAccessPayload struct {
	AdminUsername string `json:"admin_username,omitempty"`
	// CurrentPassword is mandatory: step-up proof for every change.
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password,omitempty"`
	BasePath        string `json:"base_path,omitempty"`
	ListenPort      *int   `json:"listen_port,omitempty"`
	GraceMinutes    *int   `json:"grace_minutes,omitempty"`
	ExposureMode    string `json:"exposure_mode,omitempty"`
}

func (s *server) panelAccessShow(w http.ResponseWriter, r *http.Request) {
	if s.config.AuthStore == nil {
		writeJSON(w, http.StatusServiceUnavailable, envelope{"code": "panel_access_unavailable", "message": "Panel access management is unavailable."})
		return
	}
	settings, err := s.config.AuthStore.ReadPanelSettings(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, envelope{"code": "panel_access_read_failed", "message": "Panel access could not be read."})
		return
	}
	if settings == nil {
		username, found, err := s.config.AuthStore.EffectiveAdminUsername(r.Context())
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, envelope{"code": "panel_access_read_failed", "message": "Panel access could not be read."})
			return
		}
		writeJSON(w, http.StatusOK, envelope{
			"status":         "defaults",
			"admin_username": username,
			"configured":     found,
			"base_path":      "/panel",
			"listen_port":    8080,
			"grace_minutes":  10,
			"exposure_mode":  "reverse_proxy",
		})
		return
	}
	writeJSON(w, http.StatusOK, envelope{
		"status":         "configured",
		"configured":     true,
		"admin_username": settings.AdminUsername,
		"base_path":      settings.BasePath,
		"listen_port":    settings.ListenPort,
		"grace_minutes":  settings.GraceMinutes,
		"exposure_mode":  settings.ExposureMode,
		"session_ttl":    int((settings.SessionTTL + time.Second - 1) / time.Second),
		"updated_at":     settings.UpdatedAt,
	})
}

func validatePanelAccessPayload(payload panelAccessPayload, hasSettings bool) (adminUsername, basePath string, listenPort, graceMinutes int, err error) {
	// Step-up: the current password is mandatory for every change.
	if payload.CurrentPassword == "" {
		return "", "", 0, 0, errors.New("current password is required")
	}
	adminUsername, err = ValidateAdminUsername(payload.AdminUsername)
	if err != nil {
		return "", "", 0, 0, err
	}
	basePath, err = ValidateBasePath(payload.BasePath)
	if err != nil {
		return "", "", 0, 0, err
	}
	listenPort = 8080
	if payload.ListenPort != nil {
		listenPort = *payload.ListenPort
		if err := ValidateListenPort(listenPort); err != nil {
			return "", "", 0, 0, err
		}
	} else if !hasSettings {
		listenPort = 8080
	}
	graceMinutes = 10
	if payload.GraceMinutes != nil {
		graceMinutes = *payload.GraceMinutes
		if graceMinutes < 1 || graceMinutes > 60 {
			return "", "", 0, 0, errors.New("panel access: grace minutes must be 1..60")
		}
	}
	if payload.ExposureMode != "" && payload.ExposureMode != "reverse_proxy" && payload.ExposureMode != "direct" {
		return "", "", 0, 0, errors.New("panel access: exposure mode must be reverse_proxy or direct")
	}
	if payload.NewPassword != "" {
		if err := ValidateNewPassword(payload.NewPassword); err != nil {
			return "", "", 0, 0, err
		}
	}
	return adminUsername, basePath, listenPort, graceMinutes, nil
}

func (s *server) panelAccessUpdate(w http.ResponseWriter, r *http.Request) {
	if s.config.AuthStore == nil {
		writeJSON(w, http.StatusServiceUnavailable, envelope{"code": "panel_access_unavailable", "message": "Panel access management is unavailable."})
		return
	}
	request, ok := authenticatedFromRequest(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, envelope{"code": "authentication_failed", "message": "Authentication failed."})
		return
	}
	var payload panelAccessPayload
	if err := decodeStrictJSON(r, &payload); err != nil {
		writeJSON(w, http.StatusBadRequest, envelope{"code": "invalid_request", "message": "Invalid request."})
		return
	}

	settings, err := s.config.AuthStore.ReadPanelSettings(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, envelope{"code": "panel_access_read_failed", "message": "Panel access could not be read."})
		return
	}

	adminUsername, basePath, listenPort, graceMinutes, err := validatePanelAccessPayload(payload, settings != nil)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, envelope{"code": "invalid_request", "message": sanitizePanelAccessError(err)})
		return
	}

	// Step-up proof: verify the current password against the actor record.
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

	input := auth.PanelAccessInput{
		AdminUsername: adminUsername,
		BasePath:      basePath,
		ListenPort:    listenPort,
		GraceMinutes:  graceMinutes,
		ExposureMode:  payload.ExposureMode,
	}
	if settings != nil && settings.ExposureMode != "" && input.ExposureMode == "" {
		input.ExposureMode = settings.ExposureMode
	}
	if payload.NewPassword != "" {
		hash, err := auth.HashPassword(payload.NewPassword)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, envelope{"code": "panel_access_update_failed", "message": "Panel access update failed."})
			return
		}
		input.PasswordHash = hash
	}
	if settings != nil && input.PasswordHash == "" {
		input.PasswordHash = settings.PasswordHash
	}
	actorID := request.Bound.Principal.ActorID
	if err := s.config.AuthStore.UpsertPanelSettings(r.Context(), input, &actorID); err != nil {
		writeJSON(w, http.StatusInternalServerError, envelope{"code": "panel_access_update_failed", "message": "Panel access update failed."})
		return
	}
	_ = s.config.AuthStore.AppendAudit(r.Context(), &actorID, "panel.access.update", "success", "")
	writeJSON(w, http.StatusOK, envelope{
		"status":         "updated",
		"admin_username": input.AdminUsername,
		"base_path":      input.BasePath,
		"listen_port":    input.ListenPort,
		"grace_minutes":  input.GraceMinutes,
		"exposure_mode":  input.ExposureMode,
		"note":           "Runtime exposure changes apply on the next reconciled reload; the previous base path stays accepted during the grace window.",
	})
}

func sanitizePanelAccessError(err error) string {
	msg := err.Error()
	msg = strings.ReplaceAll(msg, "panel access: ", "")
	return msg
}
