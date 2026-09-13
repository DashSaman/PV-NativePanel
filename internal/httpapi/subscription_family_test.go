package httpapi

import (
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/DashSaman/PV-NaivePanel/internal/subscription"
)

// headerRecorder implements http.ResponseWriter minimally to capture headers.
type headerRecorder struct {
	h http.Header
}

func newHeaderRecorder() *headerRecorder { return &headerRecorder{h: http.Header{}} }

func (r *headerRecorder) Header() http.Header       { return r.h }
func (r *headerRecorder) Write([]byte) (int, error) { return 0, nil }
func (r *headerRecorder) WriteHeader(int)           {}

func TestNegotiateFamily(t *testing.T) {
	// Query override wins.
	f, err := negotiateFamily(url.Values{"family": {"mihomo"}}, "v2rayNG/1.8")
	if err != nil || f != subscription.FamilyClash {
		t.Fatalf("override = %q %v", f, err)
	}
	// Unknown override errors.
	if _, err := negotiateFamily(url.Values{"family": {"bogus"}}, ""); err == nil {
		t.Fatal("unknown override must error")
	}
	// UA detection applies without override.
	f, err = negotiateFamily(url.Values{}, "Karing/1.0.30 (Android)")
	if err != nil || f != subscription.FamilySingBox {
		t.Fatalf("karing UA = %q %v", f, err)
	}
	// Default stays naive for unknown/empty UAs (backwards compatible).
	f, err = negotiateFamily(url.Values{}, "Mozilla/5.0")
	if err != nil || f != subscription.FamilyNaive {
		t.Fatalf("default = %q %v", f, err)
	}
}

func TestSetSubscriptionDeliveryHeadersHonestValues(t *testing.T) {
	quota := int64(10 * 1024 * 1024 * 1024)
	used := int64(1 * 1024 * 1024 * 1024)
	up := int64(100)
	down := int64(200)
	expires := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)

	// Full accounting available.
	p := subscription.Profile{
		Available:      true,
		QuotaBytes:     &quota,
		UsedBytes:      &used,
		UsageAvailable: true,
		ExpiresAt:      &expires,
	}
	p.AccountingBaseline.UploadBytes = &up
	p.AccountingBaseline.DownloadBytes = &down
	rr := newHeaderRecorder()
	setSubscriptionDeliveryHeaders(rr, p)
	if got := rr.h.Get("Profile-Update-Interval"); got != "4" {
		t.Fatalf("update interval = %q", got)
	}
	want := subscription.UserinfoHeader(up, down, quota, expires.UTC().Unix())
	if got := rr.h.Get("Subscription-Userinfo"); got != want {
		t.Fatalf("userinfo = %q, want %q", got, want)
	}

	// No accounting truth: honest zeros, no fabrication.
	p2 := subscription.Profile{Available: true, ExpiresAt: &expires}
	rr2 := newHeaderRecorder()
	setSubscriptionDeliveryHeaders(rr2, p2)
	want2 := subscription.UserinfoHeader(0, 0, 0, expires.UTC().Unix())
	if got := rr2.h.Get("Subscription-Userinfo"); got != want2 {
		t.Fatalf("userinfo without accounting = %q, want %q", got, want2)
	}
}
