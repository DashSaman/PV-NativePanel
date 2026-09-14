package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestValidatePanelAccessPayload(t *testing.T) {
	port := 8080
	grace := 10
	base := panelAccessPayload{
		AdminUsername:   "admin.owner-1",
		CurrentPassword: "current-secret-value",
		BasePath:        "/panel",
		ListenPort:      &port,
		GraceMinutes:    &grace,
	}
	if _, _, _, _, err := validatePanelAccessPayload(base, false); err != nil {
		t.Fatalf("valid payload rejected: %v", err)
	}

	// Step-up: current password is mandatory for EVERY change.
	noStepUp := base
	noStepUp.CurrentPassword = ""
	if _, _, _, _, err := validatePanelAccessPayload(noStepUp, false); err == nil {
		t.Fatal("missing current password must be rejected")
	}

	// Username charset and length.
	badUser := base
	badUser.AdminUsername = "Admin With Spaces"
	if _, _, _, _, err := validatePanelAccessPayload(badUser, false); err == nil {
		t.Fatal("bad username must be rejected")
	}

	// Base path reserved list.
	badPath := base
	badPath.BasePath = "/api"
	if _, _, _, _, err := validatePanelAccessPayload(badPath, false); err == nil {
		t.Fatal("reserved base path must be rejected")
	}

	// Edge port belongs to the proxy, not the panel loopback listener.
	badPort := 443
	badPortPayload := base
	badPortPayload.ListenPort = &badPort
	if _, _, _, _, err := validatePanelAccessPayload(badPortPayload, false); err == nil {
		t.Fatal("port 443 must be rejected")
	}

	// Grace bounds.
	badGrace := 61
	badGracePayload := base
	badGracePayload.GraceMinutes = &badGrace
	if _, _, _, _, err := validatePanelAccessPayload(badGracePayload, false); err == nil {
		t.Fatal("grace minutes above 60 must be rejected")
	}

	// Exposure mode whitelist.
	badExposure := base
	badExposure.ExposureMode = "hybrid"
	if _, _, _, _, err := validatePanelAccessPayload(badExposure, false); err == nil {
		t.Fatal("unknown exposure mode must be rejected")
	}

	// New password policy (12+ chars and strength >= 3).
	weak := base
	weak.NewPassword = "password1234"
	if _, _, _, _, err := validatePanelAccessPayload(weak, false); err == nil {
		t.Fatal("weak new password must be rejected")
	}
	strong := base
	strong.NewPassword = "orbital-Quartz-77-panel"
	if _, _, _, _, err := validatePanelAccessPayload(strong, false); err != nil {
		t.Fatalf("strong new password rejected: %v", err)
	}
}

func TestPanelAccessRoutesRegistered(t *testing.T) {
	found := map[string]bool{}
	for _, route := range Routes {
		if route.Path == "/api/v1/panel-access" {
			found[route.Method+"-"+string(route.Access)] = true
			if route.Name != "panel.access.show" && route.Name != "panel.access.update" {
				t.Fatalf("unexpected route name %q", route.Name)
			}
		}
	}
	if !found["GET-owner"] || !found["PUT-owner"] {
		t.Fatalf("panel-access routes missing or wrong access: %v", found)
	}
}

func TestPanelAccessShowWithoutStore(t *testing.T) {
	server := NewServer(ServerConfig{})
	request := httptest.NewRequest(http.MethodGet, "/api/v1/panel-access", nil)
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, request)
	// Unauthenticated caller must never reach the handler body: 401 or the
	// 503 envelope are both acceptable; a 200 leaking defaults is not.
	if recorder.Code == http.StatusOK {
		t.Fatalf("panel-access must not leak settings to unauthenticated callers: %d", recorder.Code)
	}
}

func TestSanitizePanelAccessError(t *testing.T) {
	msg := sanitizePanelAccessError(errInvalidBasePathFixture)
	if strings.Contains(msg, "panel access: ") {
		t.Fatalf("prefix not stripped: %q", msg)
	}
}

type errInvalidBasePathFixtureType struct{}

func (errInvalidBasePathFixtureType) Error() string { return "panel access: base path rejected" }

var errInvalidBasePathFixture = errInvalidBasePathFixtureType{}
