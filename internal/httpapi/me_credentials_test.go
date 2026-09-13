package httpapi

import (
	"net/http"
	"strings"
	"testing"
)

func TestValidateMePasswordChange(t *testing.T) {
	cases := []struct {
		name    string
		current string
		next    string
		wantErr bool
	}{
		{"accepts policy-compliant pair", "current-password-1", "a-fresh-new-password", false},
		{"requires current password", "", "a-fresh-new-password", true},
		{"rejects short new password", "current-password-1", "short", true},
		{"rejects empty new password", "current-password-1", "", true},
		{"rejects oversized new password", "current-password-1", strings.Repeat("x", 1025), true},
		{"accepts maximum length", "current-password-1", strings.Repeat("x", 1024), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateMePasswordChange(tc.current, tc.next)
			if tc.wantErr && err == nil {
				t.Fatalf("expected error for %q/%q", tc.current, tc.next)
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidateMeProfileUpdate(t *testing.T) {
	cases := []struct {
		name        string
		email       string
		displayName string
		wantErr     bool
	}{
		{"accepts email change", "Owner@Example.com", "", false},
		{"accepts display name change", "", "Saman K.", false},
		{"accepts both", "Owner@Example.com", "Saman K.", false},
		{"rejects empty payload", "", "", true},
		{"rejects email without at", "owner.example.com", "", true},
		{"rejects email with two ats", "a@b@c.example", "", true},
		{"rejects email without dot in domain", "owner@localhost", "", true},
		{"rejects email with whitespace", "owner @example.com", "", true},
		{"rejects too short email", "a@", "", true},
		{"rejects too long email", strings.Repeat("a", 310) + "@example.com", "", true},
		{"rejects too long display name", "", strings.Repeat("ن", 161), true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateMeProfileUpdate(tc.email, tc.displayName)
			if tc.wantErr && err == nil {
				t.Fatalf("expected error for email=%q display_name=%q", tc.email, tc.displayName)
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestMeCredentialRoutesRegisteredReadyAndAuthenticated(t *testing.T) {
	want := map[string]struct {
		method string
		access Access
		path   string
	}{
		"me.password.update": {method: http.MethodPost, access: Authenticated, path: "/api/v1/me/password"},
		"me.profile.update":  {method: http.MethodPatch, access: Authenticated, path: "/api/v1/me/profile"},
	}
	found := map[string]bool{}
	for _, route := range Routes {
		spec, ok := want[route.Name]
		if !ok {
			continue
		}
		found[route.Name] = true
		if route.Method != spec.method || route.Access != spec.access || route.Path != spec.path {
			t.Fatalf("%s registered as %s %s (access=%s), want %s %s (access=%s)",
				route.Name, route.Method, route.Path, route.Access, spec.method, spec.path, spec.access)
		}
		if !route.Ready {
			t.Fatalf("%s must be Ready so the panel can use it", route.Name)
		}
	}
	for name := range want {
		if !found[name] {
			t.Fatalf("route %s missing from Routes", name)
		}
	}
}
