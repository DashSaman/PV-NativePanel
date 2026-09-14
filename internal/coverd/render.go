package coverd

import (
	"html"
	"sort"
	"strings"
	"time"
)

// ContentStore supplies the latest cached items for a node (backed by the
// cover_content SECURITY DEFINER functions). Implementations must return the
// last snapshot even when feeds are down — the site never renders empty.
type ContentStore interface {
	Latest(nodeID string, limit int) ([]Item, error)
}

// Renderer renders persona pages for one node.
type Renderer struct {
	NodeID   string
	Persona  Persona
	Store    ContentStore
	BasePath string // site base path, always "/" for the cover
	Now      func() time.Time
}

func (r *Renderer) now() time.Time {
	if r.Now != nil {
		return r.Now()
	}
	return time.Now()
}

// structuralSignature returns the ordered tag+class skeleton of a document,
// used by tests to prove personas are structurally distant (CAMO-003).
func structuralSignature(doc string) string {
	var out []string
	for _, token := range strings.FieldsFunc(doc, func(r rune) bool { return r == '<' || r == '>' }) {
		token = strings.TrimSpace(token)
		if token == "" {
			continue
		}
		name := token
		if sp := strings.IndexAny(token, " \t\n"); sp >= 0 {
			name = token[:sp]
		}
		if strings.HasPrefix(name, "/") || strings.HasPrefix(name, "!") || strings.HasPrefix(name, "?") {
			continue
		}
		out = append(out, name)
	}
	return strings.Join(out, ",")
}

func esc(s string) string { return html.EscapeString(s) }

func (r *Renderer) items(limit int) []Item {
	if r.Store == nil {
		return nil
	}
	items, err := r.Store.Latest(r.NodeID, limit)
	if err != nil || items == nil {
		return []Item{}
	}
	sort.SliceStable(items, func(i, j int) bool {
		ti, tj := time.Time{}, time.Time{}
		if items[i].PublishedAt != nil {
			ti = *items[i].PublishedAt
		}
		if items[j].PublishedAt != nil {
			tj = *items[j].PublishedAt
		}
		return ti.After(tj)
	})
	if len(items) > limit {
		items = items[:limit]
	}
	return items
}

// page wraps body with the persona shell: RTL Persian, inline persona CSS,
// Jalali date in the masthead, and zero panel references.
func (r *Renderer) page(title string, cssClass string, body string) string {
	p := r.Persona
	return "<!DOCTYPE html>\n<html lang=\"fa\" dir=\"rtl\">\n<head>\n" +
		"<meta charset=\"utf-8\">\n" +
		"<meta name=\"viewport\" content=\"width=device-width, initial-scale=1\">\n" +
		"<title>" + esc(title) + " | " + esc(p.Title) + "</title>\n" +
		"<meta name=\"description\" content=\"" + esc(p.Tagline) + "\">\n" +
		"<style>\n" + personaCSS(p, cssClass) + "\n</style>\n</head>\n" +
		"<body class=\"" + cssClass + "\">\n" +
		body +
		"<footer class=\"p-footer\"><p>" + esc(p.Title) + " — " + esc(p.Tagline) + "</p>" +
		"<p>" + esc(JalaliDateLong(r.now())) + "</p>" +
		"<nav class=\"p-footer-nav\"><a href=\"/about\">درباره ما</a><a href=\"/contact\">تماس با ما</a><a href=\"/feed.xml\">پی‌آر‌اس</a></nav></footer>\n" +
		"</body>\n</html>\n"
}

func masthead(p Persona, cssClass string) string {
	return "<header class=\"" + cssClass + "-mast\"><div class=\"" + cssClass + "-brand\">" +
		"<span class=\"" + cssClass + "-mark\" aria-hidden=\"true\"></span>" +
		"<strong>" + esc(p.Title) + "</strong><em>" + esc(p.Tagline) + "</em></div>" +
		"<nav class=\"" + cssClass + "-nav\"><a href=\"/\">خانه</a><a href=\"/about\">درباره</a><a href=\"/contact\">تماس</a></nav></header>"
}

func itemCard(kind string, it Item, cssClass string, withTime bool) string {
	var meta string
	if withTime && it.PublishedAt != nil {
		meta = "<time class=\"" + cssClass + "-time\">" + esc(JalaliDateShort(*it.PublishedAt)) + "</time>"
	}
	thumb := ""
	if it.ThumbURL != "" {
		thumb = "<img class=\"" + cssClass + "-thumb\" src=\"" + esc(it.ThumbURL) + "\" alt=\"\" loading=\"lazy\">"
	}
	attribution := "<cite class=\"" + cssClass + "-src\">منبع: " + esc(it.Source) + "</cite>"
	return "<article class=\"" + cssClass + "-item\">" + thumb +
		"<h3><a href=\"" + esc(it.URL) + "\" rel=\"noopener\">" + esc(it.Title) + "</a></h3>" +
		meta +
		"<p>" + esc(it.Summary) + "</p>" + attribution +
		"<span class=\"" + cssClass + "-kind kind-" + esc(kind) + "\"></span></article>"
}

// Home renders the persona home page. Each persona uses a different section
// order, tag structure, and class vocabulary.
func (r *Renderer) Home() string {
	p := r.Persona
	date := JalaliDateLong(r.now())
	items := r.items(12)
	var first Item
	if len(items) > 0 {
		first = items[0]
	}
	rest := items
	if len(rest) > 0 {
		rest = rest[1:]
	}

	switch p.ID {
	case PersonaNewsPortal: // hero → grid → sidebar
		var b strings.Builder
		b.WriteString(masthead(p, "np"))
		b.WriteString("<div class=\"np-datelien\">" + esc(date) + "</div>")
		b.WriteString("<section class=\"np-hero\">")
		if first.Title != "" {
			b.WriteString(itemCard(first.Kind, first, "np", true))
		}
		b.WriteString("</section>")
		b.WriteString("<section class=\"np-grid\">")
		for _, it := range rest {
			b.WriteString(itemCard(it.Kind, it, "np", true))
		}
		b.WriteString("</section>")
		b.WriteString("<aside class=\"np-side\"><h2 class=\"np-side-h\">پرببینده‌ها</h2><ul class=\"np-side-list\">")
		for i := 0; i < 3 && i < len(items); i++ {
			b.WriteString("<li><a href=\"" + esc(items[i].URL) + "\">" + esc(items[i].Title) + "</a></li>")
		}
		b.WriteString("</ul></aside>")
		return r.page(p.Title, "np-body", b.String())

	case PersonaCultural: // centered masthead → feature → divided list
		var b strings.Builder
		b.WriteString(masthead(p, "cf"))
		b.WriteString("<section class=\"cf-feature\"><h2 class=\"cf-kicker\">" + esc(date) + "</h2>")
		if first.Title != "" {
			b.WriteString("<article class=\"cf-lead\"><h3><a href=\"" + esc(first.URL) + "\">" + esc(first.Title) + "</a></h3><p>" + esc(first.Summary) + "</p><cite class=\"cf-src\">منبع: " + esc(first.Source) + "</cite></article>")
		}
		b.WriteString("</section><ol class=\"cf-list\">")
		for _, it := range rest {
			b.WriteString("<li class=\"cf-line\"><a href=\"" + esc(it.URL) + "\">" + esc(it.Title) + "</a><span class=\"cf-dot\">" + esc(it.Source) + "</span></li>")
		}
		b.WriteString("</ol>")
		return r.page(p.Title, "cf-body", b.String())

	case PersonaCityServices: // service cards (evergreen) → announcements
		var b strings.Builder
		b.WriteString(masthead(p, "cs"))
		b.WriteString("<section class=\"cs-services\"><h2 class=\"cs-h\">خدمات</h2><div class=\"cs-cards\">")
		for _, svc := range []string{"استعلام پرونده", "ثبت درخواست", "پیگیری احکام", "دانلود فرم‌ها", "آرشیو بخشنامه‌ها", "سوالات متداول"} {
			b.WriteString("<div class=\"cs-card\"><span class=\"cs-ico\"></span><strong>" + svc + "</strong></div>")
		}
		b.WriteString("</div></section><section class=\"cs-news\"><h2 class=\"cs-h\">اطلاعیه‌ها</h2>")
		for _, it := range items {
			b.WriteString(itemCard(it.Kind, it, "cs", true))
		}
		b.WriteString("</section>")
		return r.page(p.Title, "cs-body", b.String())

	case PersonaNewsAgency: // ticker → wire list → chips
		var b strings.Builder
		b.WriteString(masthead(p, "na"))
		b.WriteString("<div class=\"na-tick\" aria-hidden=\"true\"><span>آخرین خبرها:</span>")
		for i := 0; i < 4 && i < len(items); i++ {
			b.WriteString("<em>" + esc(items[i].Title) + "</em>")
		}
		b.WriteString("</div>")
		b.WriteString("<section class=\"na-wire\"><h2 class=\"na-h\">" + esc(date) + "</h2>")
		for _, it := range items {
			b.WriteString(itemCard(it.Kind, it, "na", true))
		}
		b.WriteString("</section><div class=\"na-chips\">")
		for _, chip := range []string{"سیاسی", "اقتصادی", "فرهنگی", "ورزشی", "بین‌الملل", "استان‌ها"} {
			b.WriteString("<span class=\"na-chip\">" + chip + "</span>")
		}
		b.WriteString("</div>")
		return r.page(p.Title, "na-body", b.String())

	case PersonaResearch: // stats row → papers list
		var b strings.Builder
		b.WriteString(masthead(p, "ri"))
		b.WriteString("<section class=\"ri-stats\"><div class=\"ri-stat\"><b>۱۲</b><span>پژوهش جاری</span></div><div class=\"ri-stat\"><b>۸</b><span>گروه مطالعاتی</span></div><div class=\"ri-stat\"><b>۲۴</b><span>کارشناس</span></div></section>")
		b.WriteString("<section class=\"ri-papers\"><h2 class=\"ri-h\">مطالعات و گزارش‌ها</h2><p class=\"ri-date\">" + esc(date) + "</p>")
		for _, it := range items {
			b.WriteString("<article class=\"ri-paper\"><h3><a href=\"" + esc(it.URL) + "\">" + esc(it.Title) + "</a></h3><p>" + esc(it.Summary) + "</p><cite class=\"ri-src\">" + esc(it.Source) + "</cite></article>")
		}
		b.WriteString("</section>")
		return r.page(p.Title, "ri-body", b.String())

	default: // PersonaCharity: banner → campaigns → messages
		var b strings.Builder
		b.WriteString(masthead(p, "ch"))
		b.WriteString("<section class=\"ch-banner\"><h2 class=\"ch-slogan\">همت برای همدلی</h2><p class=\"ch-date\">" + esc(date) + "</p></section>")
		b.WriteString("<section class=\"ch-campaigns\"><h2 class=\"ch-h\">کمپین‌های فعال</h2><div class=\"ch-grid\">")
		for _, it := range items {
			b.WriteString(itemCard(it.Kind, it, "ch", false))
		}
		b.WriteString("</div></section><section class=\"ch-messages\"><h2 class=\"ch-h\">پیام‌ها و بیانیه‌ها</h2><ul class=\"ch-msgs\">")
		for _, it := range items {
			if it.Kind == "message" {
				b.WriteString("<li><a href=\"" + esc(it.URL) + "\">" + esc(it.Title) + "</a><cite>" + esc(it.Source) + "</cite></li>")
			}
		}
		b.WriteString("</ul></section>")
		return r.page(p.Title, "ch-body", b.String())
	}
}

// About renders the persona about page.
func (r *Renderer) About() string {
	p := r.Persona
	body := masthead(p, "pg") + "<main class=\"pg-main\"><h2>درباره ما</h2>" +
		"<p>" + esc(p.Title) + " با هدف «" + esc(p.Tagline) + "» فعالیت خود را آغاز کرده است. ما با تکیه بر کارشناسان باتجربه و بهره‌گیری از روش‌های نوین، تلاش می‌کنیم خدماتی شایستهٔ اعتماد شهروندان ارائه کنیم.</p>" +
		"<p>راهبردهای ما شفافیت، پاسخ‌گویی و احترام به حقوق مخاطبان است. برای ارتباط با تیم ما از صفحهٔ تماس استفاده کنید.</p></main>"
	return r.page("درباره ما", "pg-body", body)
}

// Contact renders the persona contact page.
func (r *Renderer) Contact() string {
	p := r.Persona
	body := masthead(p, "pg") + "<main class=\"pg-main\"><h2>تماس با ما</h2>" +
		"<address class=\"pg-addr\"><p>نشانی: تهران، خیابان اصلی، پلاک ۱۲، طبقه ۳</p>" +
		"<p>تلفن پاسخگویی: ۰۲۱-۱۲۳۴۵۶۷۸ (داخلی ۱۰)</p>" +
		"<p>ساعات پاسخگویی: شنبه تا چهارشنبه، ۸ تا ۱۴</p></address></main>"
	return r.page("تماس با ما", "pg-body", body)
}

// NotFound renders the human 404 page (never a framework default).
func (r *Renderer) NotFound() string {
	p := r.Persona
	body := masthead(p, "pg") + "<main class=\"pg-main\"><h2>صفحهٔ مورد نظر یافت نشد</h2>" +
		"<p>نشانی واردشده در سایت موجود نیست؛ از نوار بالا برای رسیدن به صفحهٔ اصلی استفاده کنید.</p>" +
		"<p><a class=\"pg-home-link\" href=\"/\">بازگشت به خانه</a></p></main>"
	return r.page("یافت نشد", "pg-body", body)
}

// Robots renders robots.txt allowing public pages only.
func (r *Renderer) Robots() string {
	return "User-agent: *\nAllow: /\nDisallow: /feed.xml\nSitemap: /sitemap.xml\n"
}

// Sitemap renders sitemap.xml for the static pages.
func (r *Renderer) Sitemap() string {
	base := "/"
	var b strings.Builder
	b.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<urlset xmlns=\"http://www.sitemaps.org/schemas/sitemap/0.9\">\n")
	for _, path := range []string{"/", "/about", "/contact"} {
		b.WriteString("<url><loc>" + esc(base+strings.TrimPrefix(path, "/")) + "</loc></url>\n")
	}
	b.WriteString("</urlset>\n")
	return b.String()
}

// Feed renders our own RSS output (attribution preserved, links out only).
func (r *Renderer) Feed() string {
	now := r.now().Format(time.RFC1123Z)
	var b strings.Builder
	b.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<rss version=\"2.0\"><channel>\n")
	b.WriteString("<title>" + esc(r.Persona.Title) + "</title>\n<link>/</link>\n<description>" + esc(r.Persona.Tagline) + "</description>\n<lastBuildDate>" + now + "</lastBuildDate>\n")
	for _, it := range r.items(30) {
		b.WriteString("<item><title>" + esc(it.Title) + "</title><link>" + esc(it.URL) + "</link><source>" + esc(it.Source) + "</source></item>\n")
	}
	b.WriteString("</channel></rss>\n")
	return b.String()
}

// personaCSS emits the persona-specific stylesheet. Palettes, class names,
// and layout rules differ per persona (structural fingerprint diversity).
func personaCSS(p Persona, bodyClass string) string {
	common := "*{box-sizing:border-box;margin:0;padding:0}" +
		"body{font-family:Vazirmatn,Tahoma,'Segoe UI',sans-serif;background:" + p.Background + ";color:" + p.Ink + ";line-height:1.8}" +
		"a{color:inherit;text-decoration:none}img{max-width:100%;display:block}" +
		".p-footer{padding:24px;text-align:center;font-size:12px;color:" + p.Ink + ";opacity:.75;border-top:1px solid rgba(0,0,0,.08);margin-top:40px}" +
		".p-footer-nav{display:flex;gap:16px;justify-content:center;margin-top:8px}" +
		".pg-main{max-width:760px;margin:24px auto;padding:0 16px}.pg-main h2{color:" + p.Accent + ";margin-bottom:12px}" +
		".pg-addr p{margin:6px 0}.pg-home-link{color:" + p.Accent + ";text-decoration:underline}"
	var persona string
	switch p.ID {
	case PersonaNewsPortal:
		persona = ".np-body{background:" + p.Background + "}" +
			".np-mast{display:flex;justify-content:space-between;align-items:center;padding:14px 24px;background:" + p.Accent + ";color:#fff}" +
			".np-brand strong{font-size:20px;display:block}.np-brand em{font-size:12px;font-style:normal;opacity:.85}.np-nav a{margin-inline-start:16px;font-size:14px}" +
			".np-datelien{max-width:1080px;margin:10px auto 0;padding:0 16px;font-size:12px;color:#5a6478}" +
			".np-hero{max-width:1080px;margin:12px auto;padding:0 16px}.np-item{background:#fff;border-radius:12px;padding:14px;box-shadow:0 1px 4px rgba(20,30,60,.08);margin-bottom:14px}" +
			".np-hero .np-item h3 a{font-size:22px}.np-grid{max-width:1080px;margin:0 auto;padding:0 16px;display:grid;grid-template-columns:repeat(3,1fr);gap:14px}" +
			".np-side{max-width:1080px;margin:18px auto;padding:0 16px}.np-side-h{font-size:15px;margin-bottom:8px;color:" + p.Accent + "}.np-side-list{list-style:none}.np-side-list li{padding:6px 0;border-bottom:1px dashed #dde3ee}" +
			".np-kind{display:none}.np-src{font-size:11px;color:#8a93a5}" +
			"@media(max-width:820px){.np-grid{grid-template-columns:1fr}}"
	case PersonaCultural:
		persona = ".cf-body{background:" + p.Background + "}" +
			".cf-mast{text-align:center;padding:34px 16px 22px;border-bottom:3px double " + p.Accent + "}" +
			".cf-brand strong{font-size:26px;display:block;letter-spacing:.5px}.cf-brand em{font-size:13px;color:" + p.Accent + ";font-style:normal}" +
			".cf-nav{margin-top:10px}.cf-nav a{margin:0 10px;font-size:13px}" +
			".cf-feature{max-width:720px;margin:26px auto 0;padding:0 18px;text-align:center}" +
			".cf-kicker{font-size:13px;color:" + p.Accent + ";margin-bottom:14px}.cf-lead h3{font-size:21px;margin-bottom:8px}.cf-lead p{font-size:15px}" +
			".cf-list{max-width:720px;margin:22px auto;padding:0 18px;list-style:none}.cf-line{padding:12px 0;border-bottom:1px solid #eadfcb;display:flex;justify-content:space-between;gap:12px}" +
			".cf-dot{font-size:11px;color:#a08a5f;white-space:nowrap}.cf-src{display:block;font-size:11px;color:#a08a5f;margin-top:6px}"
	case PersonaCityServices:
		persona = ".cs-body{background:" + p.Background + "}" +
			".cs-mast{display:flex;justify-content:space-between;padding:14px 22px;background:" + p.Accent + ";color:#fff}" +
			".cs-nav a{margin-inline-start:14px;font-size:14px}" +
			".cs-h{font-size:16px;margin:20px 0 10px;color:" + p.Accent + "}" +
			".cs-services,.cs-news{max-width:960px;margin:8px auto;padding:0 16px}" +
			".cs-cards{display:grid;grid-template-columns:repeat(3,1fr);gap:12px}" +
			".cs-card{background:#fff;border:1px solid #d9ece8;border-radius:10px;padding:18px 12px;text-align:center;font-size:14px}" +
			".cs-ico{display:block;width:26px;height:26px;margin:0 auto 8px;border-radius:50%;background:" + p.Accent + ";opacity:.35}" +
			".cs-item{background:#fff;border-radius:10px;padding:12px;margin-bottom:10px;border-inline-start:4px solid " + p.Accent + "}" +
			".cs-src{font-size:11px;color:#6d8781;display:block;margin-top:4px}@media(max-width:760px){.cs-cards{grid-template-columns:1fr 1fr}}"
	case PersonaNewsAgency:
		persona = ".na-body{background:" + p.Background + "}" +
			".na-mast{background:" + p.Accent + ";color:#fff;padding:12px 22px;display:flex;justify-content:space-between}" +
			".na-tick{background:" + p.Ink + ";color:#fff;padding:8px 16px;font-size:13px;display:flex;gap:14px;overflow:hidden;white-space:nowrap}.na-tick span{color:#f0c9c9}.na-tick em{font-style:normal;opacity:.85}" +
			".na-h{font-size:14px;color:" + p.Accent + ";margin:16px 0 8px}.na-wire{max-width:860px;margin:0 auto;padding:0 16px}" +
			".na-item{background:#fff;margin-bottom:10px;padding:12px;border-radius:8px;border:1px solid #e8d8d8}" +
			".na-time{font-size:11px;color:#a06a6a;margin-inline-start:8px}.na-src{font-size:11px;color:#a06a6a;display:block;margin-top:4px}" +
			".na-chips{max-width:860px;margin:16px auto;padding:0 16px;display:flex;flex-wrap:wrap;gap:8px}.na-chip{font-size:12px;background:#fff;border:1px solid " + p.Accent + ";color:" + p.Accent + ";border-radius:999px;padding:4px 12px}"
	case PersonaResearch:
		persona = ".ri-body{background:" + p.Background + "}" +
			".ri-mast{padding:26px 22px;background:" + p.Accent + ";color:#fff}.ri-nav a{margin-inline-start:14px;font-size:14px}" +
			".ri-stats{max-width:920px;margin:18px auto;padding:0 16px;display:flex;gap:12px}" +
			".ri-stat{flex:1;background:#fff;border-radius:10px;padding:16px;text-align:center}.ri-stat b{display:block;font-size:24px;color:" + p.Accent + "}.ri-stat span{font-size:12px;color:#5d7278}" +
			".ri-papers{max-width:920px;margin:14px auto;padding:0 16px}.ri-h{font-size:17px;margin-bottom:8px}.ri-date{font-size:12px;color:#7c8f95;margin-bottom:14px}" +
			".ri-paper{background:#fff;border-radius:10px;padding:16px;margin-bottom:12px;border-inline-start:5px solid " + p.Accent + "}.ri-src{font-size:11px;color:#7c8f95}"
	default: // charity
		persona = ".ch-body{background:" + p.Background + "}" +
			".ch-mast{display:flex;justify-content:space-between;padding:14px 22px;background:" + p.Accent + ";color:#fff}" +
			".ch-banner{max-width:920px;margin:18px auto;padding:30px 18px;background:linear-gradient(135deg," + p.Accent + ",#4c8a52);border-radius:14px;color:#fff;text-align:center}" +
			".ch-slogan{font-size:24px}.ch-date{font-size:12px;opacity:.9;margin-top:6px}" +
			".ch-h{font-size:16px;margin:18px 0 10px;color:" + p.Accent + "}" +
			".ch-campaigns,.ch-messages{max-width:920px;margin:0 auto;padding:0 16px}" +
			".ch-grid{display:grid;grid-template-columns:repeat(3,1fr);gap:12px}" +
			".ch-item{background:#fff;border-radius:12px;padding:14px;box-shadow:0 1px 3px rgba(30,60,30,.12)}" +
			".ch-src{font-size:11px;color:#6a8a6e;display:block;margin-top:4px}" +
			".ch-msgs{list-style:none}.ch-msgs li{padding:8px 0;border-bottom:1px dashed #d7e6d8;display:flex;justify-content:space-between}" +
			"@media(max-width:760px){.ch-grid{grid-template-columns:1fr}}"
	}
	_ = bodyClass
	return common + persona
}
