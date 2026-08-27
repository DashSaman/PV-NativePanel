package httpapi

import (
	"crypto/sha256"
	"crypto/subtle"
	"errors"
	"net/http"
	"net/url"
	"strings"
)

func roleAllowed(access Access, role string) bool {
	switch access {
	case Public:
		return true
	case Authenticated:
		return role == "owner" || role == "admin" || role == "operator" || role == "auditor" || role == "reseller"
	case Owner:
		return role == "owner"
	case Admin:
		return role == "owner" || role == "admin"
	case Operator:
		return role == "owner" || role == "operator"
	case Auditor:
		return role == "owner" || role == "auditor"
	case Reseller:
		return role == "owner" || role == "reseller"
	default:
		return false
	}
}

func newSessionCookie(value string) *http.Cookie {
	return &http.Cookie{
		Name:     "__Host-pvnaive_session",
		Value:    value,
		Path:     "/",
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}
}

func newCSRFCookie(value string) *http.Cookie {
	return &http.Cookie{
		Name:     "__Host-pvnaive_csrf",
		Value:    value,
		Path:     "/",
		Secure:   true,
		HttpOnly: false,
		SameSite: http.SameSiteStrictMode,
	}
}

func unsafeMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return false
	default:
		return true
	}
}

func validateCSRF(r *http.Request, expectedHash []byte) error {
	if !unsafeMethod(r.Method) {
		return nil
	}
	if len(expectedHash) != sha256.Size {
		return errors.New("csrf: invalid session binding")
	}
	if strings.EqualFold(strings.TrimSpace(r.Header.Get("Sec-Fetch-Site")), "cross-site") {
		return errors.New("csrf: cross-site request")
	}
	if origin := strings.TrimSpace(r.Header.Get("Origin")); origin != "" {
		parsed, err := url.Parse(origin)
		if err != nil || parsed.Scheme != "https" || !strings.EqualFold(parsed.Host, r.Host) {
			return errors.New("csrf: origin mismatch")
		}
	}
	cookie, err := r.Cookie("__Host-pvnaive_csrf")
	if err != nil || cookie.Value == "" {
		return errors.New("csrf: cookie missing")
	}
	header := r.Header.Get("X-CSRF-Token")
	if header == "" || subtle.ConstantTimeCompare([]byte(cookie.Value), []byte(header)) != 1 {
		return errors.New("csrf: double-submit mismatch")
	}
	sum := sha256.Sum256([]byte(header))
	if subtle.ConstantTimeCompare(sum[:], expectedHash) != 1 {
		return errors.New("csrf: session binding mismatch")
	}
	return nil
}
