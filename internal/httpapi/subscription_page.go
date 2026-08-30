package httpapi

import (
	"crypto/rand"
	"encoding/base64"
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/DashSaman/PV-NaivePanel/internal/customer"
	"github.com/DashSaman/PV-NaivePanel/internal/subscription"
)

type accountMessages struct {
	Title            string
	Subtitle         string
	ReadOnlyNotice   string
	Status           map[string]string
	TotalQuota       string
	Used             string
	Upload           string
	Download         string
	RemainingTraffic string
	Expiry           string
	RemainingDays    string
	StartPolicy      string
	Online           string
	OnlineValue      string
	OfflineValue     string
	LastOnline       string
	UsageUnavailable string
	Unavailable      string
	Unlimited        string
	NoExpiry         string
	Expired          string
	DaysSuffix       string
	FromCreation     string
	FromFirstConnect string
	FixedTimestamp   string
	Subscription     string
	DirectNaive      string
	SubscriptionQR   string
	DirectNaiveQR    string
	CopySubscription string
	CopyDirect       string
	Copied           string
	InactiveNotice   string
	SecurityFootnote string
	Overview         string
	Connection       string
	ConnectionHint   string
	HelpTitle        string
	StepOne          string
	StepTwo          string
	StepThree        string
	PrivateBadge     string
	SubscriptionHint string
	DirectHint       string
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
	UploadLabel        string
	DownloadLabel      string
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
:root{color-scheme:dark;--bg:#080a0e;--surface:#11151b;--surface2:#0d1117;--surface3:#171c24;--line:#222a35;--line2:#303947;--text:#f5f7fa;--muted:#8f99a8;--gold:#d6a84b;--gold2:#f0cf7c;--green:#47c98a;--orange:#e5a14d;--red:#db6a72;--blue:#62a3ef;--shadow:0 20px 60px #0005}*{box-sizing:border-box}html{background:var(--bg)}body{margin:0;min-height:100vh;background:radial-gradient(circle at 78% -10%,#30240b 0,transparent 28rem),var(--bg);font-family:Tahoma,"Segoe UI",Arial,sans-serif;color:var(--text)}a{color:inherit;text-decoration:none}.account-shell{width:min(1120px,calc(100% - 30px));margin:auto;padding:24px 0 44px}.account-nav{display:flex;align-items:center;justify-content:space-between;gap:18px;margin-bottom:18px}.brand{display:flex;align-items:center;gap:11px}.mark{width:44px;height:44px;border:1px solid #6f5621;border-radius:13px;background:linear-gradient(145deg,#2b2416,#11151b);display:grid;place-items:center;color:var(--gold2);font-size:16px;font-weight:900;box-shadow:inset 0 1px 0 #fff1}.brand strong,.brand small{display:block}.brand strong{font-size:18px}.brand small{color:var(--gold);font-size:8px;font-weight:800;letter-spacing:1.6px;margin-top:2px}.nav-tools{display:flex;align-items:center;gap:8px}.lang-switch{display:flex;gap:3px;padding:3px;border:1px solid var(--line);border-radius:9px;background:var(--surface2)}.lang-switch a{min-width:32px;padding:6px 7px;border-radius:6px;text-align:center;color:var(--muted);font-size:9px;font-weight:800}.lang-switch a.active{background:var(--surface3);color:var(--gold2)}.state{display:inline-flex;align-items:center;gap:6px;padding:7px 10px;border:1px solid var(--line);border-radius:999px;font-size:9px;font-weight:800;background:var(--surface2)}.state:before{content:"";width:6px;height:6px;border-radius:50%;background:currentColor;box-shadow:0 0 9px currentColor}.state.ok{color:var(--green)}.state.warn{color:var(--orange)}.state.bad{color:var(--red)}.hero-card,.service-card,.connect-card,.qr-panel,.help-card,.service-strip{background:linear-gradient(150deg,var(--surface),var(--surface2));border:1px solid var(--line);box-shadow:var(--shadow)}.hero-card{position:relative;overflow:hidden;border-radius:18px;padding:24px 25px;margin-bottom:12px}.hero-card:after{content:"";position:absolute;width:220px;height:220px;left:-100px;top:-120px;border-radius:50%;background:radial-gradient(circle,#d6a84b18,transparent 70%)}.hero-main{display:flex;align-items:flex-end;justify-content:space-between;gap:18px;position:relative;z-index:1}.hero-copy small{display:block;color:var(--gold);font-size:9px;font-weight:800;letter-spacing:.8px;margin-bottom:7px}.hero-copy h1{font-size:29px;margin:0 0 8px;overflow-wrap:anywhere}.hero-copy p{max-width:760px;margin:0;color:var(--muted);font-size:10px;line-height:1.8}.private-badge{display:inline-flex;align-items:center;gap:6px;padding:7px 9px;border:1px solid #5d4921;border-radius:9px;background:#2a2112;color:var(--gold2);font-size:9px;white-space:nowrap}.section-label{display:flex;align-items:center;justify-content:space-between;gap:12px;margin:16px 2px 9px}.section-label h2{font-size:14px;margin:0}.section-label span{font-size:9px;color:var(--muted)}.service-grid{display:grid;grid-template-columns:repeat(4,minmax(0,1fr));gap:9px}.service-card{border-radius:13px;padding:14px;min-height:88px}.service-card span{display:block;color:var(--muted);font-size:9px;margin-bottom:8px}.service-card strong{display:block;font-size:17px;overflow-wrap:anywhere}.service-card.primary strong{color:var(--gold2)}.service-strip{display:grid;grid-template-columns:repeat(4,minmax(0,1fr));gap:0;border-radius:13px;margin-top:9px;overflow:hidden}.service-strip>div{padding:11px 13px;border-left:1px solid var(--line)}.service-strip>div:last-child{border-left:0}.service-strip span,.service-strip strong{display:block}.service-strip span{color:var(--muted);font-size:8px;margin-bottom:4px}.service-strip strong{font-size:10px;overflow-wrap:anywhere}.connection-grid{display:grid;grid-template-columns:minmax(0,1.55fr) minmax(300px,.75fr);gap:10px}.connect-card,.qr-panel{border-radius:15px;padding:17px;min-width:0}.connect-head{display:flex;justify-content:space-between;gap:12px;align-items:flex-start;margin-bottom:13px}.connect-head h2{font-size:15px;margin:0 0 4px}.connect-head p{margin:0;color:var(--muted);font-size:9px;line-height:1.7}.connection-item{padding:12px;border:1px solid var(--line);border-radius:11px;background:var(--surface2);margin-top:8px}.connection-item:first-of-type{margin-top:0}.connection-item-head{display:flex;align-items:center;justify-content:space-between;gap:10px;margin-bottom:7px}.connection-item-head strong{font-size:10px}.connection-item-head small{color:var(--muted);font-size:8px}.code{direction:ltr;text-align:left;display:block;width:100%;white-space:nowrap;overflow:auto;border-radius:8px;padding:9px 10px;background:#080b10;border:1px solid var(--line);font:10px/1.7 ui-monospace,SFMono-Regular,Consolas,monospace;color:#dce3eb}.actions{display:flex;gap:7px;margin-top:8px;flex-wrap:wrap}.btn{min-height:34px;border:0;border-radius:8px;padding:7px 11px;background:linear-gradient(180deg,var(--gold2),var(--gold));color:#0a0d10;font-size:9px;font-weight:900;cursor:pointer}.btn.secondary{background:var(--surface3);border:1px solid var(--line2);color:var(--text)}.btn:active{transform:translateY(1px)}.unavailable{border:1px dashed var(--line2);border-radius:9px;padding:10px;color:var(--muted);font-size:9px;line-height:1.7}.qr-panel h2{font-size:14px;margin:0 0 12px}.qrgrid{display:grid;grid-template-columns:1fr 1fr;gap:9px}.qrbox{display:grid;justify-items:center;gap:7px;background:var(--surface2);border:1px solid var(--line);border-radius:11px;padding:10px;text-align:center}.qrbox h3{font-size:9px;margin:0;color:var(--muted)}.qr{width:100%;max-width:142px;height:auto;display:block;background:#fff;padding:6px;border-radius:8px}.help-card{border-radius:15px;padding:17px;margin-top:10px}.help-card h2{font-size:14px;margin:0 0 12px}.steps{display:grid;grid-template-columns:repeat(3,minmax(0,1fr));gap:8px}.step{display:grid;grid-template-columns:28px 1fr;gap:9px;align-items:start;padding:11px;border:1px solid var(--line);border-radius:10px;background:var(--surface2)}.step b{width:28px;height:28px;display:grid;place-items:center;border-radius:8px;background:#2a2112;color:var(--gold2);border:1px solid #5d4921;font-size:10px}.step span{font-size:9px;line-height:1.8;color:var(--muted)}.foot{text-align:center;color:var(--muted);font-size:8px;margin-top:15px}.foot strong{color:var(--gold2)}
@media(prefers-color-scheme:light){:root{color-scheme:light;--bg:#f3f5f7;--surface:#fff;--surface2:#f8f9fb;--surface3:#eef1f4;--line:#dfe4ea;--line2:#cfd6de;--text:#161a20;--muted:#697584;--gold:#a7781c;--gold2:#815b12;--green:#23875e;--orange:#ad6f25;--red:#b24f58;--blue:#376fae;--shadow:0 16px 45px #18202b12}body{background:radial-gradient(circle at 78% -10%,#e8d7ae 0,transparent 28rem),var(--bg)}.mark{background:linear-gradient(145deg,#fff8e8,#fff);border-color:#dec98d}.private-badge{background:#fff7e5;border-color:#dec98d}.code{background:#f4f6f8;color:#202730}.step b{background:#fff7e5;border-color:#dec98d}}
@media(max-width:820px){.account-shell{width:min(100% - 18px,1120px);padding-top:16px}.service-grid{grid-template-columns:repeat(2,minmax(0,1fr))}.service-strip{grid-template-columns:repeat(2,minmax(0,1fr))}.service-strip>div{border-bottom:1px solid var(--line)}.connection-grid{grid-template-columns:1fr}.steps{grid-template-columns:1fr}.hero-main{align-items:flex-start;flex-direction:column}.hero-card{padding:20px}}
@media(max-width:520px){.account-nav{align-items:flex-start}.nav-tools{align-items:flex-end;flex-direction:column}.hero-copy h1{font-size:24px}.service-grid{grid-template-columns:1fr 1fr}.service-card{min-height:78px;padding:12px}.service-card strong{font-size:14px}.service-strip{grid-template-columns:1fr 1fr}.qrgrid{grid-template-columns:1fr 1fr}.connect-card,.qr-panel,.help-card{padding:13px}.connection-item-head{align-items:flex-start;flex-direction:column}.actions{display:grid;grid-template-columns:1fr}.actions .btn{width:100%}.brand strong{font-size:16px}}
@media(max-width:370px){.service-grid,.service-strip,.qrgrid{grid-template-columns:1fr}}
@media(prefers-reduced-motion:reduce){*{scroll-behavior:auto!important;transition:none!important}}
</style>
</head>
<body><main class="account-shell">
<nav class="account-nav"><div class="brand"><div class="mark">PV</div><div><strong>PVNaive</strong><small>PVNETWORK</small></div></div><div class="nav-tools"><div class="lang-switch"><a href="?lang=fa" class="{{if eq .Lang "fa"}}active{{end}}">FA</a><a href="?lang=en" class="{{if eq .Lang "en"}}active{{end}}">EN</a></div><span class="state {{.StatusClass}}">{{.StatusLabel}}</span></div></nav>
<section class="hero-card"><div class="hero-main"><div class="hero-copy"><small>{{.M.Subtitle}}</small><h1>{{.Username}}</h1><p>{{.M.ReadOnlyNotice}}</p></div><span class="private-badge">◈ {{.M.PrivateBadge}}</span></div></section>
<div class="section-label"><h2>{{.M.Overview}}</h2><span>{{.M.Title}}</span></div>
<section class="service-grid">
<article class="service-card primary"><span>{{.M.TotalQuota}}</span><strong>{{.QuotaLabel}}</strong></article>
<article class="service-card"><span>{{.M.Used}}</span><strong>{{.UsageLabel}}</strong></article>
<article class="service-card"><span>{{.M.Upload}}</span><strong>{{.UploadLabel}}</strong></article>
<article class="service-card"><span>{{.M.Download}}</span><strong>{{.DownloadLabel}}</strong></article>
<article class="service-card"><span>{{.M.RemainingTraffic}}</span><strong>{{.RemainingTraffic}}</strong></article>
<article class="service-card"><span>{{.M.Expiry}}</span><strong>{{.ExpiryLabel}}</strong></article>
</section>
<section class="service-strip">
<div><span>{{.M.RemainingDays}}</span><strong>{{.RemainingDaysLabel}}</strong></div>
<div><span>{{.M.StartPolicy}}</span><strong>{{.StartLabel}}</strong></div>
<div><span>{{.M.Online}}</span><strong>{{.OnlineLabel}}</strong></div>
<div><span>{{.M.LastOnline}}</span><strong>{{.LastOnlineLabel}}</strong></div>
</section>
<div class="section-label"><h2>{{.M.Connection}}</h2><span>{{.M.ConnectionHint}}</span></div>
<section class="connection-grid">
<div class="connect-card">
<div class="connect-head"><div><h2>{{.M.Subscription}}</h2><p>{{.M.SubscriptionHint}}</p></div></div>
<div class="connection-item"><div class="connection-item-head"><strong>{{.M.Subscription}}</strong><small>{{.M.SubscriptionQR}}</small></div><code class="code" id="subscription-value">{{.SubscriptionURL}}</code><div class="actions"><button class="btn" type="button" data-copy-target="subscription-value">{{.M.CopySubscription}}</button></div></div>
<div class="connection-item"><div class="connection-item-head"><strong>{{.M.DirectNaive}}</strong><small>{{.M.DirectHint}}</small></div>{{if .Available}}<code class="code" id="direct-value">{{.DirectURI}}</code><div class="actions"><button class="btn secondary" type="button" data-copy-target="direct-value">{{.M.CopyDirect}}</button></div>{{else}}<div class="unavailable">{{.M.InactiveNotice}}</div>{{end}}</div>
</div>
<div class="qr-panel"><h2>QR</h2><div class="qrgrid"><div class="qrbox"><h3>{{.M.SubscriptionQR}}</h3>{{if .SubscriptionQR}}<img class="qr" src="{{.SubscriptionQR}}" alt="{{.M.SubscriptionQR}}">{{else}}<div class="unavailable">{{.M.Unavailable}}</div>{{end}}</div><div class="qrbox"><h3>{{.M.DirectNaiveQR}}</h3>{{if .DirectQR}}<img class="qr" src="{{.DirectQR}}" alt="{{.M.DirectNaiveQR}}">{{else}}<div class="unavailable">{{.M.Unavailable}}</div>{{end}}</div></div></div>
</section>
<section class="help-card"><h2>{{.M.HelpTitle}}</h2><div class="steps"><div class="step"><b>1</b><span>{{.M.StepOne}}</span></div><div class="step"><b>2</b><span>{{.M.StepTwo}}</span></div><div class="step"><b>3</b><span>{{.M.StepThree}}</span></div></div></section>
<p class="foot"><strong>PVNaive</strong> · {{.M.SecurityFootnote}}</p>
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
	uploadLabel := messages.UsageUnavailable
	downloadLabel := messages.UsageUnavailable
	remainingTraffic := messages.UsageUnavailable
	onlineLabel := messages.Unavailable
	lastOnlineLabel := messages.Unavailable
	if s.config.AccountingStore != nil && profile.ServiceTermID != "" {
		now := time.Now().UTC()
		model, readErr := s.config.AccountingStore.Read(r.Context(), profile.ServiceTermID, now, customerAccountingStaleAfter)
		if readErr == nil {
			if model.LastOnline != nil {
				lastOnlineLabel = model.LastOnline.UTC().Format("2006-01-02 15:04 UTC")
			}
			if model.AccountingComplete {
				onlineLabel = messages.OfflineValue
				if model.Online {
					onlineLabel = messages.OnlineValue
				}
			}
			baseline := customer.AccountingBaseline{
				State:         customer.AccountingBaselineState(profile.AccountingBaseline.State),
				Source:        customer.AccountingBaselineSource(profile.AccountingBaseline.Source),
				CutoffAt:      profile.AccountingBaseline.CutoffAt,
				UploadBytes:   profile.AccountingBaseline.UploadBytes,
				DownloadBytes: profile.AccountingBaseline.DownloadBytes,
			}
			usage, capability, composeErr := customer.ComposeCustomerUsage(baseline, model.UploadBytes, model.DownloadBytes, profile.QuotaBytes, model.AccountingComplete)
			if composeErr == nil && capability.Available && usage.UsedBytes != nil && usage.UploadBytes != nil && usage.DownloadBytes != nil {
				usageLabel = subscriptionByteLabel(*usage.UsedBytes)
				uploadLabel = subscriptionByteLabel(*usage.UploadBytes)
				downloadLabel = subscriptionByteLabel(*usage.DownloadBytes)
				if usage.RemainingBytes != nil {
					remainingTraffic = subscriptionByteLabel(*usage.RemainingBytes)
				} else if profile.QuotaBytes == nil {
					remainingTraffic = messages.Unlimited
				}
			}
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
		UploadLabel:        uploadLabel,
		DownloadLabel:      downloadLabel,
		RemainingTraffic:   remainingTraffic,
		ExpiryLabel:        accountExpiryLabel(profile, messages),
		RemainingDaysLabel: accountRemainingDays(profile, messages),
		StartLabel:         accountStartLabel(profile.StartPolicy, messages),
		OnlineLabel:        onlineLabel,
		LastOnlineLabel:    lastOnlineLabel,
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
	return "fa"
}

func messagesForLanguage(lang string) accountMessages {
	if lang == "en" {
		return accountMessages{
			Title: "Account status", Subtitle: "NaiveProxy account", ReadOnlyNotice: "This page is read-only. Opening or copying it never changes your token, password, quota, expiry, first-use state, or Runtime credential.",
			Status:     map[string]string{"active": "Active", "pending": "Pending first connection", "suspended": "Suspended", "expired": "Expired", "depleted": "Quota depleted", "revoked": "Revoked", "inactive": "Inactive"},
			TotalQuota: "Total quota", Used: "Used", Upload: "Upload", Download: "Download", RemainingTraffic: "Remaining", Expiry: "Expiry", RemainingDays: "Remaining days", StartPolicy: "Start policy", Online: "Online", OnlineValue: "Online", OfflineValue: "Offline", LastOnline: "Last online",
			UsageUnavailable: "Usage unavailable", Unavailable: "Unavailable", Unlimited: "Unlimited", NoExpiry: "No expiry", Expired: "Expired", DaysSuffix: "days",
			FromCreation: "From creation", FromFirstConnect: "From first successful connection", FixedTimestamp: "Fixed expiry",
			Subscription: "Subscription", DirectNaive: "Direct Naive", SubscriptionQR: "Subscription QR", DirectNaiveQR: "Direct Naive QR", CopySubscription: "Copy Subscription", CopyDirect: "Copy Direct Naive", Copied: "Copied ✓",
			InactiveNotice: "Direct access is unavailable while this service is inactive.", SecurityFootnote: "Private account page — do not share publicly",
			Overview: "Service overview", Connection: "Quick connection", ConnectionHint: "Choose the method that fits your client", HelpTitle: "Connect with Karing", PrivateBadge: "Private account",
			SubscriptionHint: "Recommended for Karing: add this as a Subscription URL so future profile updates can be refreshed from the same link.", DirectHint: "Direct Naive is useful for one-off/manual imports.",
			StepOne: "Copy the Subscription URL or scan its QR code.", StepTwo: "In Karing, open Add Profile and choose Subscription / Link import.", StepThree: "Save the profile, refresh the subscription, then connect.",
		}
	}
	return accountMessages{
		Title: "وضعیت حساب", Subtitle: "سرویس NaiveProxy", ReadOnlyNotice: "این صفحه فقط خواندنی است؛ باز کردن یا کپی کردن آن توکن، رمز، حجم، انقضا، اولین اتصال یا اطلاعات Runtime را تغییر نمی‌دهد.",
		Status:     map[string]string{"active": "فعال", "pending": "منتظر اولین اتصال", "suspended": "تعلیق", "expired": "منقضی", "depleted": "حجم تمام", "revoked": "لغوشده", "inactive": "غیرفعال"},
		TotalQuota: "حجم کل", Used: "مصرف", Upload: "آپلود", Download: "دانلود", RemainingTraffic: "حجم باقی‌مانده", Expiry: "انقضا", RemainingDays: "روز باقی‌مانده", StartPolicy: "شروع اعتبار", Online: "آنلاین", OnlineValue: "آنلاین", OfflineValue: "آفلاین", LastOnline: "آخرین آنلاین",
		UsageUnavailable: "در دسترس نیست", Unavailable: "در دسترس نیست", Unlimited: "نامحدود", NoExpiry: "بدون انقضا", Expired: "منقضی", DaysSuffix: "روز",
		FromCreation: "از زمان ثبت", FromFirstConnect: "از اولین اتصال موفق", FixedTimestamp: "تاریخ دستی",
		Subscription: "Subscription", DirectNaive: "Direct Naive", SubscriptionQR: "QR اشتراک", DirectNaiveQR: "QR مستقیم Naive", CopySubscription: "کپی لینک ساب", CopyDirect: "کپی Direct Naive", Copied: "کپی شد ✓",
		InactiveNotice: "تا زمانی که سرویس غیرفعال است اتصال مستقیم در دسترس نیست.", SecurityFootnote: "صفحه خصوصی حساب — عمومی منتشر نکنید",
		Overview: "نمای کلی سرویس", Connection: "اتصال سریع", ConnectionHint: "روش مناسب کلاینت خودت را انتخاب کن", HelpTitle: "اتصال با Karing", PrivateBadge: "حساب خصوصی",
		SubscriptionHint: "روش پیشنهادی برای Karing: این لینک را به‌عنوان Subscription اضافه کن تا بروزرسانی‌های بعدی از همین لینک دریافت شوند.", DirectHint: "Direct Naive برای ورود دستی یا اتصال مستقیم قابل استفاده است.",
		StepOne: "لینک Subscription را کپی کن یا QR آن را اسکن کن.", StepTwo: "در Karing وارد Add Profile شو و Subscription / Link را انتخاب کن.", StepThree: "پروفایل را ذخیره، Subscription را Refresh و سپس اتصال را فعال کن.",
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
