package httpapi

import (
	"net/http"
	"strings"
	"time"

	"github.com/DashSaman/PV-NaivePanel/internal/customer"
	"github.com/DashSaman/PV-NaivePanel/internal/subscription"
)

// publicSubscriptionInfo serves the read-only JSON profile behind
// GET /api/v1/subscriptions/{token}/info. The token itself is the bearer
// credential: anyone holding it can already fetch the raw Naive URI via
// /sub/{token} and the rendered account page via /s/{token}. This endpoint
// exposes the same information in a machine-readable form so subscription
// clients and support tooling can present live quota/expiry state without
// scraping HTML. No secret material beyond the DirectURI (which the token
// holder can already obtain) is returned.
func (s *server) publicSubscriptionInfo(w http.ResponseWriter, r *http.Request) {
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

	subscriptionURL, err := canonicalSubscriptionURL(s.config.SubscriptionProxyHost, token)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	accountPageURL, err := canonicalAccountPageURL(s.config.SubscriptionProxyHost, token)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	payload := envelope{
		"username":         profile.Username,
		"status":           subscriptionStatusKey(profile),
		"available":        profile.Available,
		"quota_bytes":      profile.QuotaBytes,
		"quota_label":      subscriptionQuotaLabel(profile.QuotaBytes),
		"expires_at":       subscriptionTimeValue(profile.ExpiresAt),
		"expiry_label":     subscriptionExpiryLabel(profile.ExpiresAt),
		"remaining_label":  subscriptionRemainingLabel(profile.ExpiresAt),
		"start_policy":     profile.StartPolicy,
		"usage_available":  profile.UsageAvailable,
		"subscription_url": subscriptionURL,
		"account_page_url": accountPageURL,
	}
	if profile.Available && profile.DirectURI != "" {
		payload["direct_uri"] = profile.DirectURI
	}

	if used, remaining, ok := s.subscriptionUsageSnapshot(r, profile); ok {
		payload["used_bytes"] = used
		payload["remaining_bytes"] = remaining
		payload["used_label"] = subscriptionByteLabel(*used)
		if remaining != nil {
			payload["remaining_traffic_label"] = subscriptionByteLabel(*remaining)
		}
	}

	writeJSON(w, http.StatusOK, payload)
}

// subscriptionUsageSnapshot composes the current accounting period usage for
// the profile. It returns ok=false whenever accounting is unavailable so the
// endpoint stays truthful instead of inventing numbers.
func (s *server) subscriptionUsageSnapshot(r *http.Request, profile subscription.Profile) (*int64, *int64, bool) {
	if s.config.AccountingStore == nil || profile.ServiceTermID == "" {
		return nil, nil, false
	}
	model, readErr := s.config.AccountingStore.Read(r.Context(), profile.ServiceTermID, time.Now().UTC(), customerAccountingStaleAfter)
	if readErr != nil {
		return nil, nil, false
	}
	baseline := customer.AccountingBaseline{
		State:         customer.AccountingBaselineState(profile.AccountingBaseline.State),
		Source:        customer.AccountingBaselineSource(profile.AccountingBaseline.Source),
		CutoffAt:      profile.AccountingBaseline.CutoffAt,
		UploadBytes:   profile.AccountingBaseline.UploadBytes,
		DownloadBytes: profile.AccountingBaseline.DownloadBytes,
	}
	usage, capability, composeErr := customer.ComposeCustomerUsageForPeriod(
		baseline, model.LastResetAt, model.UploadBytes, model.DownloadBytes, profile.QuotaBytes, model.AccountingComplete,
	)
	if composeErr != nil || !capability.Available {
		return nil, nil, false
	}
	if usage.UsedBytes == nil {
		return nil, nil, false
	}
	return usage.UsedBytes, usage.RemainingBytes, true
}

func subscriptionTimeValue(value *time.Time) any {
	if value == nil {
		return nil
	}
	return value.UTC().Format(time.RFC3339Nano)
}
