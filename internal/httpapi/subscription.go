package httpapi

import (
	"io"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/DashSaman/PV-NaivePanel/internal/subscription"
)

func (s *server) publicSubscription(w http.ResponseWriter, r *http.Request) {
	if s.config.SubscriptionService == nil || strings.TrimSpace(s.config.SubscriptionProxyHost) == "" {
		http.NotFound(w, r)
		return
	}
	token := r.PathValue("token")
	if token == "" {
		http.NotFound(w, r)
		return
	}
	profile, err := s.config.SubscriptionService.ResolveProfile(r.Context(), token, s.config.SubscriptionProxyHost)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	setPublicSubscriptionHeaders(w)

	if strings.HasPrefix(r.URL.Path, "/s/") {
		s.renderAccountPage(w, r, token, profile)
		return
	}

	if !profile.Available || profile.DirectURI == "" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = io.WriteString(w, profile.DirectURI+"\n")
}

func setPublicSubscriptionHeaders(w http.ResponseWriter) {
	h := w.Header()
	h.Set("Cache-Control", "no-store")
	h.Set("Pragma", "no-cache")
	h.Set("X-Robots-Tag", "noindex, nofollow, noarchive, nosnippet")
	h.Set("X-Content-Type-Options", "nosniff")
	h.Set("X-Frame-Options", "DENY")
	h.Set("Referrer-Policy", "no-referrer")
	h.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
}

func subscriptionDeliveryPaths(token string) (string, string) {
	escaped := url.PathEscape(token)
	return "/sub/" + escaped, "/s/" + escaped
}

func canonicalSubscriptionURL(proxyHost, token string) (string, error) {
	subscriptionPath, _ := subscriptionDeliveryPaths(token)
	return canonicalPublicURL(proxyHost, subscriptionPath)
}

func canonicalAccountPageURL(proxyHost, token string) (string, error) {
	_, accountPath := subscriptionDeliveryPaths(token)
	return canonicalPublicURL(proxyHost, accountPath)
}

func canonicalPublicURL(proxyHost, path string) (string, error) {
	u, err := url.Parse("https://" + strings.TrimSpace(proxyHost))
	if err != nil || u.Host == "" || u.User != nil || u.Path != "" || u.RawQuery != "" || u.Fragment != "" {
		return "", subscription.ErrUnavailable
	}
	u.Path = path
	return u.String(), nil
}

func subscriptionQuotaLabel(quota *int64) string {
	if quota == nil {
		return "Unlimited"
	}
	gb := float64(*quota) / float64(1024*1024*1024)
	if math.Abs(gb-math.Round(gb)) < 0.01 {
		return strconv.FormatFloat(math.Round(gb), 'f', 0, 64) + " GB"
	}
	return strconv.FormatFloat(gb, 'f', 1, 64) + " GB"
}

func subscriptionByteLabel(value int64) string {
	gb := float64(value) / float64(1024*1024*1024)
	if math.Abs(gb-math.Round(gb)) < 0.01 {
		return strconv.FormatFloat(math.Round(gb), 'f', 0, 64) + " GB"
	}
	return strconv.FormatFloat(gb, 'f', 2, 64) + " GB"
}

func subscriptionExpiryLabel(expires *time.Time) string {
	if expires == nil {
		return "No expiry"
	}
	return expires.UTC().Format("2006-01-02 15:04 UTC")
}

func subscriptionRemainingLabel(expires *time.Time) string {
	if expires == nil {
		return "Unlimited"
	}
	remaining := time.Until(expires.UTC())
	if remaining <= 0 {
		return "Expired"
	}
	days := int(math.Ceil(remaining.Hours() / 24))
	return strconv.Itoa(days) + " days"
}

func subscriptionStatusKey(profile subscription.Profile) string {
	if profile.Available {
		if profile.TermState == "pending" {
			return "pending"
		}
		return "active"
	}
	if profile.UserState == "suspended" {
		return "suspended"
	}
	if profile.TermState == "expired" || (profile.ExpiresAt != nil && !profile.ExpiresAt.After(time.Now().UTC())) {
		return "expired"
	}
	if profile.TermState == "quota_depleted" {
		return "depleted"
	}
	if profile.UserState == "revoked" || profile.TermState == "revoked" {
		return "revoked"
	}
	return "inactive"
}

func subscriptionStatusClass(profile subscription.Profile) string {
	if profile.Available {
		return "ok"
	}
	if profile.UserState == "suspended" {
		return "warn"
	}
	return "bad"
}
