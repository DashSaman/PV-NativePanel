package httpapi

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// Panel access validation (R7 / ACCESS-001..003).
// Pure functions first so the transactional-apply flow can validate before
// touching any runtime state; failures never leave half-applied state.

var basePathPattern = regexp.MustCompile(`^/[a-z0-9][a-z0-9\-_]{2,63}$`)

// reservedBasePaths must never host the panel: API + assets + health surface,
// and the coverd static routes (the cover site owns them on every node).
var reservedBasePaths = []string{
	"/api", "/assets", "/health", "/ready", "/about", "/contact",
	"/robots.txt", "/sitemap.xml", "/feed.xml",
}

// ValidateBasePath enforces the documented rules: lowercase charset, 3..64
// chars after the slash, and no reserved prefix. Returns the normalized
// (slash-trimmed-tail) path.
func ValidateBasePath(raw string) (string, error) {
	p := strings.TrimRight(strings.TrimSpace(raw), "/")
	if p == "" {
		p = "/"
	}
	if !basePathPattern.MatchString(p) {
		return "", fmt.Errorf("panel access: base path %q must match ^/[a-z0-9][a-z0-9-_]{2,63}$", raw)
	}
	for _, reserved := range reservedBasePaths {
		if p == reserved || strings.HasPrefix(p, reserved+"/") {
			return "", fmt.Errorf("panel access: base path %q collides with reserved prefix %s", p, reserved)
		}
	}
	return p, nil
}

// ValidateListenPort checks the port range. The recommended exposure keeps
// the panel loopback-only behind the reverse proxy; direct exposure is a
// caller-level decision requiring typed confirmation.
func ValidateListenPort(port int) error {
	if port < 1 || port > 65535 {
		return fmt.Errorf("panel access: listen port %d out of range", port)
	}
	if port == 80 || port == 443 {
		return errors.New("panel access: port 80/443 belongs to the edge proxy; pick a loopback port")
	}
	return nil
}

// ValidateAdminUsername enforces 3..64 chars of email-ish or name charset.
func ValidateAdminUsername(raw string) (string, error) {
	u := strings.TrimSpace(raw)
	if len(u) < 3 || len(u) > 64 {
		return "", fmt.Errorf("panel access: admin username must be 3..64 chars")
	}
	for _, r := range u {
		ok := r == '.' || r == '-' || r == '_' || r == '@' ||
			(r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')
		if !ok {
			return "", fmt.Errorf("panel access: admin username contains forbidden character %q", r)
		}
	}
	return u, nil
}

// commonPasswords is a small deny-list of the most guessable secrets. The
// full zxcvbn library is intentionally not vendored until a license-clean
// port is selected (tracked on the agent board, ACCESS-003 follow-up); this
// estimator is strictly more conservative on the passwords it knows.
var commonPasswords = map[string]struct{}{
	"password": {}, "password1": {}, "password12": {}, "password123": {},
	"123456": {}, "1234567": {}, "12345678": {}, "123456789": {}, "1234567890": {},
	"qwerty": {}, "qwertyui": {}, "qwerty123": {}, "admin": {}, "admin123": {},
	"administrator": {}, "letmein": {}, "welcome": {}, "welcome1": {},
	"iloveyou": {}, "monkey": {}, "dragon": {}, "football": {}, "abc12345": {},
	"111111": {}, "000000": {}, "121212": {}, "asdfgh": {}, "zxcvbnm": {},
	"saman": {}, "saman123": {}, "panel123": {}, "pvnaive123": {},
}

func hasLower(s string) bool {
	return strings.IndexFunc(s, func(r rune) bool { return r >= 'a' && r <= 'z' }) >= 0
}
func hasUpper(s string) bool {
	return strings.IndexFunc(s, func(r rune) bool { return r >= 'A' && r <= 'Z' }) >= 0
}
func hasDigit(s string) bool {
	return strings.IndexFunc(s, func(r rune) bool { return r >= '0' && r <= '9' }) >= 0
}
func hasSymbol(s string) bool {
	return strings.IndexFunc(s, func(r rune) bool {
		return strings.ContainsRune("!@#$%^&*()_+-=[]{}|;:,.<>?/~`'\"\\", r)
	}) >= 0
}

func looksLikeSequence(s string) bool {
	lower := strings.ToLower(s)
	runes := []rune(lower)
	asc, desc, rep := 0, 0, 0
	for i := 1; i < len(runes); i++ {
		switch {
		case runes[i] == runes[i-1]:
			rep++
		case runes[i] == runes[i-1]+1:
			asc++
		case runes[i] == runes[i-1]-1:
			desc++
		}
	}
	longRuns := rep >= 2 || asc >= 2 || desc >= 2
	return longRuns
}

// PasswordStrengthScore estimates a 0..4 strength for the password policy.
// Score 3+ satisfies the panel policy. The estimator is deliberately strict:
// a 12-char minimum is enforced regardless of the score.
func PasswordStrengthScore(password string) int {
	n := len(password)
	if n < 8 {
		return 0
	}
	score := 0
	switch {
	case n >= 16:
		score = 3
	case n >= 12:
		score = 2
	case n >= 8:
		score = 1
	}
	classes := 0
	for _, ok := range []bool{hasLower(password), hasUpper(password), hasDigit(password), hasSymbol(password)} {
		if ok {
			classes++
		}
	}
	if classes >= 3 {
		score++
	}
	if classes >= 4 && n >= 16 {
		score++
	}
	if looksLikeSequence(password) {
		score--
	}
	if n < 20 {
		if _, known := commonPasswords[strings.ToLower(password)]; known {
			score = 1
		}
	}
	if score < 0 {
		score = 0
	}
	if score > 4 {
		score = 4
	}
	return score
}

// ValidateNewPassword applies the full ACCESS-003 policy: minimum 12 chars
// and strength score >= 3.
func ValidateNewPassword(password string) error {
	if len(password) < 12 {
		return errors.New("panel access: password must be at least 12 characters")
	}
	if PasswordStrengthScore(password) < 3 {
		return errors.New("panel access: password too weak (use 16+ chars with mixed classes, avoid sequences and common words)")
	}
	return nil
}

// RedactAuditValue is the fixed marker recorded in panel_access_audit for
// secret fields — the actual hash/secret is never written to the audit log.
func RedactAuditValue() string { return "[redacted]" }
