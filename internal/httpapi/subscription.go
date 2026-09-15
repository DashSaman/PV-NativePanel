package httpapi

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/DashSaman/PV-NaivePanel/internal/fleet"
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

	// Browsers opening the machine endpoint must land on the human account
	// page instead of downloading a file (user-visible bug: /sub/<token>
	// opened a download dialog). An explicit ?family= override always
	// serves the raw payload so operators can preview formats from a
	// browser, and non-browser agents (Karing, NekoBox, curl, …) keep the
	// exact raw behavior they depend on.
	userAgent := r.Header.Get("User-Agent")
	override, overrideErr := subscription.FamilyFromQuery(r.URL.Query())
	if overrideErr != nil {
		http.Error(w, "subscription: unknown family override", http.StatusBadRequest)
		return
	}
	if override == "" && looksLikeBrowser(userAgent) {
		http.Redirect(w, r, "/s/"+url.PathEscape(token), http.StatusFound)
		return
	}

	if !profile.Available || profile.DirectURI == "" {
		http.NotFound(w, r)
		return
	}

	family, err := negotiateFamily(r.URL.Query(), userAgent)
	if err != nil {
		http.Error(w, "subscription: unknown family override", http.StatusBadRequest)
		return
	}
	nodes := make([]subscription.Node, 0, 4)
	if profile.Node != nil {
		nodes = append(nodes, *profile.Node)
		// Multi-server auto-switch: healthy, in-sync pool nodes join the
		// rendered profile right after the primary. sing-box/Karing gets a
		// urltest group that probes every node and sticks to the fastest;
		// Mihomo gets PV-AUTO (url-test) via the provider payload.
		nodes = append(nodes, s.fleetSubscriptionNodes(r.Context(), *profile.Node)...)
	}
	opts, err := s.renderOptions(r, token)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	body, contentType, renderErr := subscription.RenderMachinePayload(family, nodes, opts)
	if renderErr != nil {
		http.Error(w, "subscription: render failed", http.StatusInternalServerError)
		return
	}
	setSubscriptionDeliveryHeaders(w, profile)
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(body)
}

// fleetSubscriptionNodes maps serving pool nodes (healthy, in-sync, actively
// serving) into subscription render nodes. Endpoints come from each node's
// latest signed manifest revision; credentials mirror the primary node
// because the panel reconciles the same runtime credential set to every
// pool member (R5). Failure-tolerant: any store error just renders fewer
// nodes — the primary is always present.
func (s *server) fleetSubscriptionNodes(ctx context.Context, primary subscription.Node) []subscription.Node {
	if s.config.FleetStore == nil {
		return nil
	}
	inventory, err := s.config.FleetStore.ListNodes(ctx)
	if err != nil {
		return nil
	}
	latest := func(nodeID string) (fleet.ManifestEnvelope, bool) {
		_, env, found, err := s.config.FleetStore.LatestRevision(ctx, nodeID)
		if err != nil || !found {
			return fleet.ManifestEnvelope{}, false
		}
		return env, true
	}
	return fleetNodesToSubscriptionNodes(inventory, primary, latest)
}

// fleetNodesToSubscriptionNodes is the pure mapping used by
// fleetSubscriptionNodes: serving nodes in preference order, endpoints
// deduplicated against the primary, capped to keep profiles small.
func fleetNodesToSubscriptionNodes(
	inventory []fleet.PoolNode,
	primary subscription.Node,
	latestManifest func(nodeID string) (fleet.ManifestEnvelope, bool),
) []subscription.Node {
	const maxExtraNodes = 8
	extra := make([]subscription.Node, 0, len(inventory))
	for i := range inventory {
		node := inventory[i]
		if node.Health != "healthy" || node.Maintenance != "active" {
			continue
		}
		if node.AppliedRevision <= 0 || node.AppliedRevision != node.DesiredRevision {
			continue // not in sync: do not advertise a stale node
		}
		env, ok := latestManifest(node.ID)
		if !ok {
			continue
		}
		for _, endpoint := range env.Manifest.Endpoints {
			if len(extra) >= maxExtraNodes {
				return extra
			}
			if endpoint.Host == primary.Host && endpoint.Port == primary.Port {
				continue // same as the primary entry
			}
			sni := strings.TrimSpace(endpoint.SNI)
			if sni == "" {
				sni = endpoint.Host
			}
			extra = append(extra, subscription.Node{
				Name:     node.DisplayName,
				Host:     endpoint.Host,
				Port:     endpoint.Port,
				Username: primary.Username,
				Password: primary.Password,
				SNI:      sni,
			})
		}
	}
	return extra
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

// looksLikeBrowser reports whether the User-Agent is an interactive web
// browser. Every mainstream browser engine (Chromium, Gecko, WebKit, Trident,
// Blink) prefixes its UA with "Mozilla/5.0", while machine clients (Karing,
// sing-box, NekoBox, NekoRay, v2rayNG, Clash, curl, wget, …) do not. A blank
// UA is treated as a machine client: raw delivery stays the default.
func looksLikeBrowser(userAgent string) bool {
	ua := strings.ToLower(strings.TrimSpace(userAgent))
	return strings.Contains(ua, "mozilla/")
}

// renderOptions builds the machine render context. The Mihomo provider URL
// is the canonical absolute /sub/<token>?family=mihomo address embedded in
// the full profile; a fetch carrying that explicit override is Mihomo's own
// proxy-provider engine pulling the node list, so it renders the bare
// provider payload instead (STEER-003 hot updates).
func (s *server) renderOptions(r *http.Request, token string) (subscription.RenderOptions, error) {
	base, err := canonicalSubscriptionURL(s.config.SubscriptionProxyHost, token)
	if err != nil {
		return subscription.RenderOptions{}, err
	}
	return subscription.RenderOptions{
		ProviderURL:     base + "?family=mihomo",
		ProviderPayload: subscription.ProviderPayloadOverride(r.URL.Query()),
	}, nil
}

// setSubscriptionDeliveryHeaders adds the mandatory machine-endpoint headers
// (STEER-003): update interval, per-user accounting summary and a filename
// carrying the account remark (clients use it to name the imported profile).
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
	if remark := subscriptionProfileRemark(profile); remark != "" {
		// "inline" keeps the endpoint renderable in a browser (opening the
		// subscription URL shows plain text instead of triggering a download)
		// while machine clients still pick up the filename parameter to name
		// the imported profile.
		h.Set("Content-Disposition", fmt.Sprintf("inline; filename=%q", remark))
	}
}

// subscriptionProfileRemark builds the client-visible profile name
// ("PVNaive-<username>"); unsafe characters collapse to a safe fallback.
func subscriptionProfileRemark(profile subscription.Profile) string {
	name := strings.TrimSpace(profile.Username)
	if name == "" {
		return "PVNaive"
	}
	return "PVNaive-" + name
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
