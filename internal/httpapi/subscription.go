package httpapi

import (
	"html/template"
	"io"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/DashSaman/PV-NaivePanel/internal/subscription"
)

type subscriptionPageData struct {
	Username        string
	StatusLabel     string
	StatusClass     string
	QuotaLabel      string
	UsageLabel      string
	ExpiryLabel     string
	RemainingLabel  string
	StartLabel      string
	SubscriptionURL string
	DirectURI       string
	QRDataURI       template.URL
	Available       bool
}

var subscriptionPageTemplate = template.Must(template.New("subscription-page").Parse(`<!doctype html>
<html lang="fa" dir="rtl">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<meta name="robots" content="noindex,nofollow,noarchive">
<title>PVNaive — وضعیت اشتراک</title>
<style>
:root{color-scheme:dark;--bg:#07090c;--panel:#10141a;--panel2:#151b22;--line:#252d38;--text:#f7f7f4;--muted:#9aa8b8;--gold:#e5b94b;--gold2:#c99522;--ok:#45d18b;--warn:#f0b94b;--bad:#ef6671}*{box-sizing:border-box}body{margin:0;min-height:100vh;background:radial-gradient(circle at 85% 0,#29200b 0,transparent 34%),var(--bg);font-family:Tahoma,Arial,sans-serif;color:var(--text)}.wrap{width:min(980px,calc(100% - 28px));margin:0 auto;padding:30px 0 48px}.top{display:flex;align-items:center;justify-content:space-between;gap:18px;margin-bottom:22px}.brand{display:flex;align-items:center;gap:12px}.mark{width:52px;height:52px;border-radius:15px;background:linear-gradient(145deg,var(--gold),var(--gold2));color:#0b0c0e;display:grid;place-items:center;font:900 25px/1 Arial;box-shadow:0 12px 40px #e5b94b24}.brand strong{display:block;font-size:23px;letter-spacing:.2px}.brand small{color:var(--gold);font-weight:700;letter-spacing:2px}.state{padding:9px 14px;border-radius:999px;border:1px solid var(--line);font-size:13px;font-weight:800}.state.ok{color:var(--ok);background:#45d18b12;border-color:#45d18b45}.state.warn{color:var(--warn);background:#f0b94b12;border-color:#f0b94b45}.state.bad{color:var(--bad);background:#ef667112;border-color:#ef667145}.hero{padding:28px;border:1px solid var(--line);background:linear-gradient(180deg,#141920,#0f1318);border-radius:24px;box-shadow:0 24px 70px #0007}.hero h1{font-size:30px;margin:5px 0 7px}.hero p{margin:0;color:var(--muted)}.grid{display:grid;grid-template-columns:repeat(4,1fr);gap:12px;margin:16px 0}.card{background:var(--panel);border:1px solid var(--line);border-radius:18px;padding:17px}.card span{display:block;color:var(--muted);font-size:12px;margin-bottom:8px}.card strong{font-size:19px}.usage{margin:15px 0 20px;background:var(--panel);border:1px solid var(--line);border-radius:18px;padding:16px}.usage-head{display:flex;justify-content:space-between;gap:10px;margin-bottom:10px}.usage-head span{color:var(--muted)}.track{height:10px;border-radius:999px;background:repeating-linear-gradient(135deg,#202833,#202833 8px,#252f3b 8px,#252f3b 16px);overflow:hidden}.usage-note{font-size:12px;color:var(--muted);margin:9px 0 0}.delivery{display:grid;grid-template-columns:minmax(0,1fr) 250px;gap:16px;margin-top:16px}.box{background:var(--panel);border:1px solid var(--line);border-radius:20px;padding:18px}.box h2{font-size:17px;margin:0 0 12px}.qr{background:#fff;border-radius:18px;padding:12px;display:block;width:100%;height:auto}.code{direction:ltr;text-align:left;display:block;white-space:nowrap;overflow:auto;background:#090c10;border:1px solid var(--line);border-radius:12px;padding:12px;color:#dfe7ef;font:12px/1.65 Consolas,monospace;margin-bottom:10px}.actions{display:flex;gap:9px;flex-wrap:wrap}.btn{border:0;border-radius:12px;padding:11px 15px;font-weight:800;cursor:pointer;background:linear-gradient(145deg,var(--gold),var(--gold2));color:#111}.btn.secondary{background:#1b222b;color:var(--text);border:1px solid var(--line)}.disabled-note{padding:13px;border:1px solid #ef667140;background:#ef66710f;color:#ffadb4;border-radius:12px}.foot{text-align:center;color:#657382;font-size:11px;margin-top:18px}@media(max-width:760px){.wrap{width:min(100% - 18px,980px);padding-top:18px}.top{align-items:flex-start}.grid{grid-template-columns:repeat(2,1fr)}.delivery{grid-template-columns:1fr}.qrbox{order:-1}.hero{padding:20px}.hero h1{font-size:24px}}@media(max-width:420px){.grid{grid-template-columns:1fr}.brand strong{font-size:20px}.mark{width:46px;height:46px}.state{font-size:11px}}
</style>
</head>
<body><main class="wrap">
<header class="top"><div class="brand"><div class="mark">R</div><div><strong>PVNaive</strong><small>PVNETWORK</small></div></div><span class="state {{.StatusClass}}">{{.StatusLabel}}</span></header>
<section class="hero"><p>وضعیت سرویس NaiveProxy</p><h1>{{.Username}}</h1><p>اطلاعات این صفحه از همان اشتراک شما خوانده می‌شود و باز کردن آن هیچ تغییری در اکانت ایجاد نمی‌کند.</p></section>
<section class="grid"><article class="card"><span>حجم کل</span><strong>{{.QuotaLabel}}</strong></article><article class="card"><span>مصرف</span><strong>{{.UsageLabel}}</strong></article><article class="card"><span>انقضا</span><strong>{{.ExpiryLabel}}</strong></article><article class="card"><span>باقی‌مانده</span><strong>{{.RemainingLabel}}</strong></article></section>
<section class="usage"><div class="usage-head"><strong>مصرف دقیق</strong><span>{{.UsageLabel}}</span></div><div class="track"></div><p class="usage-note">تا تکمیل و اثبات accounting در سطح Naive/Caddy، عدد مصرف‌شده یا باقی‌مانده جعلی نمایش داده نمی‌شود.</p></section>
<section class="delivery"><div class="box"><h2>Subscription</h2><code class="code">{{.SubscriptionURL}}</code><div class="actions"><button class="btn" type="button" data-copy="{{.SubscriptionURL}}">کپی لینک ساب</button>{{if .Available}}<button class="btn secondary" type="button" data-copy="{{.DirectURI}}">کپی Direct Naive</button>{{end}}</div>{{if .Available}}<h2 style="margin-top:20px">Direct Naive</h2><code class="code">{{.DirectURI}}</code>{{else}}<p class="disabled-note">این سرویس در حال حاضر قابل استفاده نیست. برای تمدید یا فعال‌سازی با پشتیبانی تماس بگیرید.</p>{{end}}<p style="color:var(--muted);font-size:12px;margin-bottom:0">شروع اعتبار: {{.StartLabel}}</p></div><div class="box qrbox"><h2>QR اشتراک</h2>{{if .QRDataURI}}<img class="qr" src="{{.QRDataURI}}" alt="QR Subscription">{{else}}<p class="disabled-note">QR برای این لینک قابل ساخت نیست؛ از دکمه کپی استفاده کنید.</p>{{end}}</div></section>
<p class="foot">PVNaive · Secure NaiveProxy delivery by PVNETWORK</p>
</main><script>document.querySelectorAll('[data-copy]').forEach(function(b){b.addEventListener('click',async function(){try{await navigator.clipboard.writeText(b.getAttribute('data-copy')||'');var old=b.textContent;b.textContent='کپی شد ✓';setTimeout(function(){b.textContent=old},1200)}catch(e){}})})</script></body></html>`))

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
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("X-Robots-Tag", "noindex, nofollow, noarchive")

	if browserWantsSubscriptionPage(r) {
		subURL, err := canonicalSubscriptionURL(s.config.SubscriptionProxyHost, token)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		qr, _ := localQRDataURI(subURL)
		data := subscriptionPageData{
			Username:        profile.Username,
			StatusLabel:     subscriptionStatusLabel(profile),
			StatusClass:     subscriptionStatusClass(profile),
			QuotaLabel:      subscriptionQuotaLabel(profile.QuotaBytes),
			UsageLabel:      "در دسترس نیست",
			ExpiryLabel:     subscriptionExpiryLabel(profile.ExpiresAt),
			RemainingLabel:  subscriptionRemainingLabel(profile.ExpiresAt),
			StartLabel:      subscriptionStartLabel(profile.StartPolicy),
			SubscriptionURL: subURL,
			DirectURI:       profile.DirectURI,
			QRDataURI:       template.URL(qr),
			Available:       profile.Available,
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_ = subscriptionPageTemplate.Execute(w, data)
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

func browserWantsSubscriptionPage(r *http.Request) bool {
	if r.URL.Query().Get("raw") == "1" {
		return false
	}
	return strings.Contains(strings.ToLower(r.Header.Get("Accept")), "text/html")
}

func canonicalSubscriptionURL(proxyHost, token string) (string, error) {
	u, err := url.Parse("https://" + strings.TrimSpace(proxyHost))
	if err != nil || u.Host == "" || u.User != nil || u.Path != "" || u.RawQuery != "" || u.Fragment != "" {
		return "", subscription.ErrUnavailable
	}
	u.Path = "/api/v1/subscriptions/" + token
	return u.String(), nil
}

func subscriptionQuotaLabel(quota *int64) string {
	if quota == nil {
		return "نامحدود"
	}
	gb := float64(*quota) / float64(1024*1024*1024)
	if math.Abs(gb-math.Round(gb)) < 0.01 {
		return strconv.FormatFloat(math.Round(gb), 'f', 0, 64) + " GB"
	}
	return strconv.FormatFloat(gb, 'f', 1, 64) + " GB"
}

func subscriptionExpiryLabel(expires *time.Time) string {
	if expires == nil {
		return "بدون انقضا"
	}
	return expires.UTC().Format("2006-01-02 15:04 UTC")
}

func subscriptionRemainingLabel(expires *time.Time) string {
	if expires == nil {
		return "نامحدود"
	}
	remaining := time.Until(expires.UTC())
	if remaining <= 0 {
		return "منقضی"
	}
	days := int(math.Ceil(remaining.Hours() / 24))
	return strconv.Itoa(days) + " روز"
}

func subscriptionStartLabel(policy string) string {
	switch policy {
	case "on_first_successful_connection":
		return "از اولین اتصال موفق"
	case "fixed_timestamp":
		return "تاریخ دستی"
	default:
		return "از زمان ثبت"
	}
}

func subscriptionStatusLabel(profile subscription.Profile) string {
	if profile.Available {
		if profile.TermState == "pending" {
			return "منتظر اولین اتصال"
		}
		return "فعال"
	}
	if profile.UserState == "suspended" {
		return "تعلیق"
	}
	if profile.TermState == "expired" || (profile.ExpiresAt != nil && !profile.ExpiresAt.After(time.Now().UTC())) {
		return "منقضی"
	}
	if profile.TermState == "quota_depleted" {
		return "حجم تمام"
	}
	return "غیرفعال"
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
