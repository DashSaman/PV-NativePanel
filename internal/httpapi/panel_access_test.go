package httpapi

import (
	"strings"
	"testing"
)

func TestValidateBasePath(t *testing.T) {
	ok := map[string]string{
		"/panel":      "/panel",
		"/my-secret1": "/my-secret1",
		"/console/":   "/console",
		"/abc":        "/abc",
	}
	for raw, want := range ok {
		got, err := ValidateBasePath(raw)
		if err != nil || got != want {
			t.Fatalf("ValidateBasePath(%q) = %q %v, want %q", raw, got, err, want)
		}
	}
	bad := []string{
		"", "/", "/ab", "/a1", // too short / empty
		"/API", "/Up-per", "/has space", // charset
		"/api", "/api/v1", "/assets/x", // reserved
		"/health", "/about", "/contact", // reserved cover routes
		"/robots.txt", "/sitemap.xml", // reserved files
		"/way-too-long-" + strings.Repeat("x", 60), // length
	}
	for _, p := range bad {
		if _, err := ValidateBasePath(p); err == nil {
			t.Fatalf("ValidateBasePath(%q) must be rejected", p)
		}
	}
}

func TestValidateListenPort(t *testing.T) {
	for _, port := range []int{8080, 9443, 1, 65535} {
		if err := ValidateListenPort(port); err != nil {
			t.Fatalf("port %d must pass: %v", port, err)
		}
	}
	for _, port := range []int{0, -1, 65536, 80, 443} {
		if err := ValidateListenPort(port); err == nil {
			t.Fatalf("port %d must be rejected", port)
		}
	}
}

func TestValidateAdminUsername(t *testing.T) {
	if got, err := ValidateAdminUsername("  admin.owner-1 "); err != nil || got != "admin.owner-1" {
		t.Fatalf("username = %q %v", got, err)
	}
	for _, bad := range []string{"ab", "", strings.Repeat("x", 65), "has space", "فارسی", "slash/x"} {
		if _, err := ValidateAdminUsername(bad); err == nil {
			t.Fatalf("username %q must be rejected", bad)
		}
	}
}

func TestPasswordStrengthScore(t *testing.T) {
	cases := map[string]int{
		"short1!":                      0, // <8
		"password123":                  1, // common deny-list cap
		"aaaaaaaaaaaa":                 1, // 12 chars but repeated run
		"abcdefghijkl":                 1, // sequence run
		"pvnaive-password1":            4, // 17 chars, lower+digit+symbol('-'), no runs
		"Corr3ct-Horse!":               3, // 14 chars, 4 classes
		"Battery-Staple-Correct-2026!": 4, // 28 chars, 4 classes
	}
	for pw, want := range cases {
		if got := PasswordStrengthScore(pw); got != want {
			t.Fatalf("PasswordStrengthScore(%q) = %d, want %d", pw, got, want)
		}
	}
}

func TestValidateNewPasswordPolicy(t *testing.T) {
	// Policy: >= 12 chars AND score >= 3.
	mustPass := []string{
		"Corr3ct-Horse!",
		"pvnaive-password1",
		"Battery-Staple-Correct-2026!",
	}
	for _, pw := range mustPass {
		if err := ValidateNewPassword(pw); err != nil {
			t.Fatalf("password %q must pass: %v", pw, err)
		}
	}
	mustFail := []string{
		"short1!",        // too short
		"password123456", // 14 chars but common-list family? not exact… still weak classes
		"abcdefghijkl",   // sequence
		"aaaaaaaaaaaa",   // repeats
		"saman12345678",  // weak-ish, 3 classes but low length score
	}
	for _, pw := range mustFail {
		if err := ValidateNewPassword(pw); err == nil {
			t.Fatalf("password %q must be rejected", pw)
		}
	}
}

func TestRedactAuditValue(t *testing.T) {
	if got := RedactAuditValue(); got != "[redacted]" {
		t.Fatalf("redaction marker = %q", got)
	}
}
