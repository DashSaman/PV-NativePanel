package httpapi

import (
	"crypto/rand"
	"encoding/base64"
	"html/template"
	"net/http"
	"strings"

	"github.com/DashSaman/PV-NaivePanel/internal/subscription"
)

type accountMessages struct {
	Title             string
	Subtitle          string
	ReadOnlyNotice    string
	Status            map[string]string
	TotalQuota        string
	Used              string
	RemainingTraffic  string
	Expiry            string
	RemainingDays     string
	StartPolicy       string
	Online            string
	LastOnline        string
	UsageUnavailable  string
	Unavailable       string
	Unlimited         string
	NoExpiry          string
	Expired           string
	DaysSuffix        string
	FromCreation      string
	FromFirstConnect  string
	FixedTimestamp    string
	Subscription      string
	DirectNaive       string
	SubscriptionQR    string
	DirectNaiveQR     string
	CopySubscription  string
	CopyDirect        string
	Copied            string
	InactiveNotice    string
	SecurityFootnote  string
}

type accountPageData struct {
	Lang               string
	Dir                string
	Nonce              string
	Username           string
	StatusLabel        string
	StatusClass        string
	QuotaLabel         string
	UsageLabel         string
	RemainingTraffic   string
	ExpiryLabel        string
	RemainingDaysLabel string
	StartLabel         string
	OnlineLabel        string
	LastOnlineLabel    string
	SubscriptionURL    string
	DirectURI          string
	SubscriptionQR     template.URL
	DirectQR           template.URL
	Available          bool
	M                  accountMessages
}

var accountPageTemplate = template.Must(template.New("pvnaive-account").Parse(`<!doctype html>
<html lang="{{.Lang}}" dir="{{.Dir}}">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width,initial-scale=1,viewport-fit=cover">
<meta name="robots" content="noindex,nofollow,noarchive,nosnippet">
<title>PVNaive — {{.M.Title}}</title>
<style nonce="{{.Nonce}}">
:root{color-scheme:light dark;--bg:#f4f6f8;--surface:#fff;--surface2:#f8fafc;--line:#dbe1e8;--text:#141a22;--muted:#667384;--brand:#b88718;--ok:#087a4c;--warn:#9a6410;--bad:#b42336;--shadow:0 18px 50px #15223814}*{box-sizing:border-box}body{margin:0;min-height:100vh;background:linear-gradient(155deg,#efe4c9 0,var(--bg) 26rem);font-family:Inter,ui-sans-serif,Tahoma,Arial,sans-serif;color:var(--text)}.wrap{width:min(1040px,calc(100% - 28px));margin:auto;padding:28px 0 46px}.top{display:flex;justify-content:space-between;align-items:center;gap:18px;margin-bottom:18px}.brand{display:flex;align-items:center;gap:12px}.mark{width:50px;height:50px;border-radius:15px;background:linear-gradient(145deg,#e4bd55,#a97812);display:grid;place-items:center;color:#111;font-size:22px;font-weight:900}.brand strong{display:block;font-size:22px}.brand small{display:block;color:var(--brand);font-weight:800;letter-spacing:1.5px}.state{padding:8px 13px;border:1px solid var(--line);border-radius:999px;font-size:13px;font-weight:800}.state.ok{color:var(--ok)}.state.warn{color:var(--warn)}.state.bad{color:var(--bad)}.hero,.panel,.card{background:var(--surface);border:1px solid var(--line);box-shadow:var(--shadow)}.hero{border-radius:24px;padding:25px}.hero h1{margin:4px 0 8px;font-size:30px;overflow-wrap:anywhere}.hero p{margin:0;color:var(--muted);line-height:1.8}.grid{display:grid;grid-template-columns:repeat(4,minmax(0,1fr));gap:12px;margin:15px 0}.card{border-radius:17px;padding:16px}.card span{display:block;font-size:12px;color:var(--muted);margin-bottom:8px}.card strong{font-size:17px;overflow-wrap:anywhere}.delivery{display:grid;grid-template-columns:minmax(0,1.5fr) minmax(260px,.8fr);gap:14px}.panel{border-radius:20px;padding:18px;min-width:0}.panel h2{font-size:17px;margin:0 0 12px}.code{direction:ltr;text-align:left;display:block;white-space:nowrap;overflow:auto;background:var(--surface2);border:1px solid var(--line);border-radius:12px;padding:12px;font:12px/1.7 ui-monospace,SFMono-Regular,Consolas,monospace}.actions{display:flex;flex-wrap:wrap;gap:9px;margin:10px 0 18px}.btn{border:0;border-radius:11px;padding:10px 13px;background:#c99b2b;color:#111;font-weight:800;cursor:pointer}.btn.secondary{background:var(--surface2);border:1px solid var(--line);color:var(--text)}.qrgrid{display:grid;grid-template-columns:1fr 1fr;gap:12px}.qrbox{background:var(--surface2);border:1px solid var(--line);border-radius:16px;padding:13px}.qrbox h3{font-size:14px;margin:0 0 10px}.qr{width:100%;height:auto;display:block;background:#fff;padding:9px;border-radius:12px}.unavailable{border:1px dashed var(--line);border-radius:12px;padding:14px;color:var(--muted);font-size:13px}.foot{text-align:center;color:var(--muted);font-size:11px;margin-top:18px}@media(prefers-color-scheme:dark){:root{--bg:#080b0f;--surface:#11161d;--surface2:#0b1016;--line:#28313d;--text:#f2f5f8;--muted:#9ba8b8;--brand:#e1b64c;--ok:#4bd59a;--warn:#f0b94b;--bad:#f17582;--shadow:0 20px 55px #0008}body{background:radial-gradient(circle at 85% 0,#30270d 0,transparent 32rem),var(--bg)}}@media(max-width:780px){.grid{grid-template-columns:repeat(2,minmax(0,1fr))}.delivery{grid-template-columns:1fr}.wrap{width:min(100% - 18px,1040px);padding-top:18px}.hero{padding:20px}.hero h1{font-size:24px}}@media(max-width:480px){.grid,.qrgrid{grid-template-columns:1fr}.top{align-items:flex-start}.state{font-size:11px}.mark{width:44px;height:44px}.brand strong{font-size:19px}}
</style>
</head>
<body><main class="wrap">
<header class="top"><div class="brand"><div class="mark">PV</div><div><strong>PVNaive</strong><small>PVNETWORK</small></div></div><span class="state {{.StatusClass}}">{{.StatusLabel}}</span></header>
<section class="hero"><p>{{.M.Subtitle}}</p><h1>{{.Username}}</h1><p>{{.M.ReadOnlyNotice}}</p></section>
<section class="grid">
<article class="card"><span>{{.M.TotalQuota}}</span><strong>{{.QuotaLabel}}</strong></article>
<article class="card"><span>{{.M.Used}}</span><strong>{{.UsageLabel}}</strong></article>
<article class="card"><span>{{.M.RemainingTraffic}}</span><strong>{{.RemainingTraffic}}</strong></article>
<article class="card"><span>{{.M.Expiry}}</span><strong>{{.ExpiryLabel}}</strong></article>
<article class="card"><span>{{.M.RemainingDays}}</span><strong>{{.RemainingDaysLabel}}</strong></article>
<article class="card"><span>{{.M.StartPolicy}}</span><strong>{{.StartLabel}}</strong></article>
<article class="card"><span>{{.M.Online}}</span><strong>{{.OnlineLabel}}</strong></article>
<article class="card"><span>{{.M.LastOnline}}</span><strong>{{.LastOnlineLabel}}</strong></article>
</section>
<section class="delivery"><div class="panel"><h2>{{.M.Subscription}}</h2><code class="code" id="subscription-value">{{.SubscriptionURL}}</code><div class="actions"><button class="btn" type="button" data-copy-target="subscription-value">{{.M.CopySubscription}}</button>{{if .Available}}<button class="btn secondary" type="button" data-copy-target="direct-value">{{.M.CopyDirect}}</button>{{end}}</div><h2>{{.M.DirectNaive}}</h2>{{if .Available}}<code class="code" id="direct-value">{{.DirectURI}}</code>{{else}}<div class="unavailable">{{.M.InactiveNotice}}</div>{{end}}</div><div class="panel"><div class="qrgrid"><div class="qrbox"><h3>{{.M.SubscriptionQR}}</h3>{{if .SubscriptionQR}}<img class="qr" src="{{.SubscriptionQR}}" alt="{{.M.SubscriptionQR}}">{{else}}<div class="unavailable">{{.M.Unavailable}}</div>{{end}}</div><div class="qrbox"><h3>{{.M.DirectNaiveQR}}</h3>{{if .DirectQR}}<img class="qr" src="{{.DirectQR}}" alt="{{.M.DirectNaiveQR}}">{{else}}<div class="unavailable">{{.M.Unavailable}}</div>{{end}}</div></div></div></section>
<p class="foot">{{.M.SecurityFootnote}} · PVNaive / PVNETWORK</p>
</main><script nonce="{{.Nonce}}">document.querySelectorAll('[data-copy-target]').forEach(function(b){b.addEventListener('click',async function(){var n=document.getElementById(b.getAttribute('data-copy-target'));if(!n)return;try{await navigator.clipboard.writeText(n.textContent||'');var o=b.textContent;b.textContent={{printf "%q" .M.Copied}};setTimeout(function(){b.textContent=o},1200)}catch(e){}})})</script></body></html>`))

func (s *server) renderAccountPage(w http.ResponseWriter, r *http.Request, token string, profile subscription.Profile) {
	nonceBytes := make([]byte, 18)
	if _, err := rand.Read(nonceBytes); err != nil {
		http.Error(w, "account page unavailable", http.StatusServiceUnavailable)
		return
	}
	nonce := base64.RawURLEncoding.EncodeToString(nonceBytes)
	lang := accountLanguage(r)
	messages := messagesForLanguage(lang)
	dir := "rtl"
	if lang == "en" {
		dir = "ltr"
	}
	subscriptionURL, err := canonicalSubscriptionURL(s.config.SubscriptionProxyHost, token)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	subscriptionQR, _ := localQRDataURI(subscriptionURL)
	directQR := ""
	if profile.Available && profile.DirectURI != "" {
		directQR, _ = localQRDataURI(profile.DirectURI)
	}
	usageLabel := messages.UsageUnavailable
	remainingTraffic := messages.UsageUnavailable
	if profile.UsageAvailable && profile.UsedBytes != nil {
		usageLabel = subscriptionByteLabel(*profile.UsedBytes)
		if profile.RemainingBytes != nil {
			remainingTraffic = subscriptionByteLabel(*profile.RemainingBytes)
		} else if profile.QuotaBytes == nil {
			remainingTraffic = messages.Unlimited
		}
	}
	data := accountPageData{
		Lang:               lang,
		Dir:                dir,
		Nonce:              nonce,
		Username:           profile.Username,
		StatusLabel:        messages.Status[subscriptionStatusKey(profile)],
		StatusClass:        subscriptionStatusClass(profile),
		QuotaLabel:         accountQuotaLabel(profile.QuotaBytes, messages),
		UsageLabel:         usageLabel,
		RemainingTraffic:   remainingTraffic,
		ExpiryLabel:        accountExpiryLabel(profile, messages),
		RemainingDaysLabel: accountRemainingDays(profile, messages),
		StartLabel:         accountStartLabel(profile.StartPolicy, messages),
		OnlineLabel:        messages.Unavailable,
		LastOnlineLabel:    messages.Unavailable,
		SubscriptionURL:    subscriptionURL,
		DirectURI:          profile.DirectURI,
		SubscriptionQR:     template.URL(subscriptionQR),
		DirectQR:           template.URL(directQR),
		Available:          profile.Available && profile.DirectURI != "",
		M:                  messages,
	}
	w.Header().Set("Content-Security-Policy", "default-src 'none'; img-src 'self' data:; style-src 'nonce-"+nonce+"'; script-src 'nonce-"+nonce+"'; frame-ancestors 'none'; base-uri 'none'; form-action 'none'; connect-src 'none'")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = accountPageTemplate.Execute(w, data)
}

func accountLanguage(r *http.Request) string {
	if lang := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("lang"))); lang == "en" || lang == "fa" {
		return lang
	}
	if strings.HasPrefix(strings.ToLower(strings.TrimSpace(r.Header.Get("Accept-Language"))), "en") {
		return "en"
	}
	return "fa"
}

func messagesForLanguage(lang string) accountMessages {
	if lang == "en" {
		return accountMessages{
			Title: "Account status", Subtitle: "NaiveProxy account", ReadOnlyNotice: "This page is read-only. Opening or copying it never changes your token, password, quota, expiry, first-use state, or Runtime credential.",
			Status: map[string]string{"active": "Active", "pending": "Pending first connection", "suspended": "Suspended", "expired": "Expired", "depleted": "Quota depleted", "revoked": "Revoked", "inactive": "Inactive"},
			TotalQuota: "Total quota", Used: "Used", RemainingTraffic: "Remaining", Expiry: "Expiry", RemainingDays: "Remaining days", StartPolicy: "Start policy", Online: "Online", LastOnline: "Last online",
			UsageUnavailable: "Usage unavailable", Unavailable: "Unavailable", Unlimited: "Unlimited", NoExpiry: "No expiry", Expired: "Expired", DaysSuffix: "days",
			FromCreation: "From creation", FromFirstConnect: "From first successful connection", FixedTimestamp: "Fixed expiry",
			Subscription: "Subscription", DirectNaive: "Direct Naive", SubscriptionQR: "Subscription QR", DirectNaiveQR: "Direct Naive QR", CopySubscription: "Copy Subscription", CopyDirect: "Copy Direct Naive", Copied: "Copied ✓",
			InactiveNotice: "Direct access is unavailable while this service is inactive.", SecurityFootnote: "Private account page — do not share publicly",
		}
	}
	return accountMessages{
		Title: "وضعیت حساب", Subtitle: "سرویس NaiveProxy", ReadOnlyNotice: "این صفحه فقط خواندنی است؛ باز کردن یا کپی کردن آن توکن، رمز، حجم، انقضا، اولین اتصال یا اطلاعات Runtime را تغییر نمی‌دهد.",
		Status: map[string]string{"active": "فعال", "pending": "منتظر اولین اتصال", "suspended": "تعلیق", "expired": "منقضی", "depleted": "حجم تمام", "revoked": "لغوشده", "inactive": "غیرفعال"},
		TotalQuota: "حجم کل", Used: "مصرف", RemainingTraffic: "حجم باقی‌مانده", Expiry: "انقضا", RemainingDays: "روز باقی‌مانده", StartPolicy: "شروع اعتبار", Online: "آنلاین", LastOnline: "آخرین آنلاین",
		UsageUnavailable: "در دسترس نیست", Unavailable: "در دسترس نیست", Unlimited: "نامحدود", NoExpiry: "بدون انقضا", Expired: "منقضی", DaysSuffix: "روز",
		FromCreation: "از زمان ثبت", FromFirstConnect: "از اولین اتصال موفق", FixedTimestamp: "تاریخ دستی",
		Subscription: "Subscription", DirectNaive: "Direct Naive", SubscriptionQR: "QR اشتراک", DirectNaiveQR: "QR مستقیم Naive", CopySubscription: "کپی لینک ساب", CopyDirect: "کپی Direct Naive", Copied: "کپی شد ✓",
		InactiveNotice: "تا زمانی که سرویس غیرفعال است اتصال مستقیم در دسترس نیست.", SecurityFootnote: "صفحه خصوصی حساب — عمومی منتشر نکنید",
	}
}

func accountQuotaLabel(quota *int64, messages accountMessages) string {
	if quota == nil {
		return messages.Unlimited
	}
	return subscriptionQuotaLabel(quota)
}

func accountExpiryLabel(profile subscription.Profile, messages accountMessages) string {
	if profile.ExpiresAt == nil {
		return messages.NoExpiry
	}
	return profile.ExpiresAt.UTC().Format("2006-01-02 15:04 UTC")
}

func accountRemainingDays(profile subscription.Profile, messages accountMessages) string {
	if profile.ExpiresAt == nil {
		return messages.Unlimited
	}
	label := subscriptionRemainingLabel(profile.ExpiresAt)
	if langDays := strings.TrimSuffix(label, " days"); langDays != label && messages.DaysSuffix != "days" {
		return langDays + " " + messages.DaysSuffix
	}
	if label == "Expired" {
		return messages.Expired
	}
	return label
}

func accountStartLabel(policy string, messages accountMessages) string {
	switch policy {
	case "on_first_successful_connection":
		return messages.FromFirstConnect
	case "fixed_timestamp":
		return messages.FixedTimestamp
	default:
		return messages.FromCreation
	}
}
