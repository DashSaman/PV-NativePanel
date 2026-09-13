package httpapi

import (
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

	family, err := negotiateFamily(r.URL.Query(), r.Header.Get("User-Agent"))
	if err != nil {
		http.Error(w, "subscription: unknown family override", http.StatusBadRequest)
		return
	}
	nodes := make([]subscription.Node, 0, 1)
	if profile.Node != nil {
		nodes = append(nodes, *profile.Node)
	}
	body, contentType, renderErr := subscription.RenderMachinePayload(family, nodes)
	if renderErr != nil {
		http.Error(w, "subscription: render failed", http.StatusInternalServerError)
		return
	}
	setSubscriptionDeliveryHeaders(w, profile)
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(body)
}

// negotiateFamily applies an explicit ?family= override first, then falls
// back to User-Agent detection, then to the naive default.
func negotiateFamily(query url.Values, userAgent string) (subscription.Family, error) {
	override, err := subscription.FamilyFromQuery(query)
	if err != nil {
		return "", err
	}
	if override != "" {
		return override, nil
	}
	return subscription.DetectFamily(userAgent), nil
}

// setSubscriptionDeliveryHeaders adds the mandatory machine-endpoint headers
// (STEER-003): update interval and per-user accounting summary. Values are
// honest: fields without accounting truth stay 0.
func setSubscriptionDeliveryHeaders(w http.ResponseWriter, profile subscription.Profile) {
	var upload, download, total int64
	if profile.UsageAvailable {
		if profile.AccountingBaseline.UploadBytes != nil {
			upload = *profile.AccountingBaseline.UploadBytes
		}
		if profile.AccountingBaseline.DownloadBytes != nil {
			download = *profile.AccountingBaseline.DownloadBytes
		} else if profile.UsedBytes != nil {
			download = *profile.UsedBytes
		}
	}
	if profile.QuotaBytes != nil {
		total = *profile.QuotaBytes
	}
	var expire int64
	if profile.ExpiresAt != nil {
		expire = profile.ExpiresAt.UTC().Unix()
	}
	h := w.Header()
	h.Set("Profile-Update-Interval", "4")
	h.Set("Subscription-Userinfo", subscription.UserinfoHeader(upload, download, total, expire))
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
