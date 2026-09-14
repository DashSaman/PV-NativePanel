package coverd

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type fakeStore struct {
	items []Item
	err   error
}

func (f *fakeStore) Latest(nodeID string, limit int) ([]Item, error) {
	if f.err != nil {
		return nil, f.err
	}
	if len(f.items) > limit {
		return f.items[:limit], nil
	}
	return f.items, nil
}

func sampleItems() []Item {
	past := time.Now().Add(-2 * time.Hour).UTC()
	older := past.Add(-time.Hour)
	return []Item{
		{Source: "khamenei.ir", Kind: "message", Title: "بیانیه‌ای دربارهٔ همت ملی", URL: "https://example.org/a", Summary: "خلاصهٔ متن پیام در قالب چند سطر.", PublishedAt: &past},
		{Source: "irna", Kind: "news", Title: "گشایش نمایشگاه دستاوردها", URL: "https://example.org/b", Summary: "رویدادی با حضور مسئولان استانی.", PublishedAt: &older},
	}
}

func fixedTime() time.Time { return time.Date(2026, 9, 14, 9, 30, 0, 0, time.UTC) }

func rendererFor(id PersonaID, store ContentStore) *Renderer {
	p, ok := PersonaByID(id)
	if !ok {
		panic("coverd: missing persona " + string(id))
	}
	return &Renderer{NodeID: "node-test", Persona: p, Store: store, Now: fixedTime}
}

func TestPersonaForNodeStableAndDistinct(t *testing.T) {
	a := PersonaForNode("node-alpha")
	if a.ID != PersonaForNode("node-alpha").ID {
		t.Fatal("persona assignment must be stable per node")
	}
	seen := map[PersonaID]bool{}
	for i := 0; i < 64; i++ {
		seen[PersonaForNode(string(rune('a'+i%26))+"-node-"+string(rune('a'+i%26))).ID] = true
	}
	if len(seen) < 3 {
		t.Fatalf("expected wide persona spread across nodes, got %d packs", len(seen))
	}
}

func TestPersonaStructuralDiversity(t *testing.T) {
	store := &fakeStore{items: sampleItems()}
	sigs := map[string]PersonaID{}
	for _, p := range Personas {
		r := rendererFor(p.ID, store)
		sig := structuralSignature(r.Home())
		if other, clash := sigs[sig]; clash {
			t.Fatalf("persona %s renders structurally identical to %s (CAMO-003)", p.ID, other)
		}
		sigs[sig] = p.ID
	}
}

func TestCoverHygieneNoPanelArtifacts(t *testing.T) {
	store := &fakeStore{items: sampleItems()}
	forbidden := []string{"pvnaive", "panel", "naive", "connect", "proxy", "admin", "/api", "csrf"}
	for _, p := range Personas {
		r := rendererFor(p.ID, store)
		for _, page := range []string{r.Home(), r.About(), r.Contact(), r.NotFound(), r.Sitemap(), r.Feed()} {
			lower := strings.ToLower(page)
			for _, word := range forbidden {
				if strings.Contains(lower, word) {
					t.Fatalf("persona %s leaked artifact %q (CAMO-001)", p.ID, word)
				}
			}
		}
	}
}

func TestCoverServerHeadersAndRoutes(t *testing.T) {
	store := &fakeStore{items: sampleItems()}
	srv := &Server{Renderer: rendererFor(PersonaNewsPortal, store)}

	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != 200 {
		t.Fatalf("home status = %d", rec.Code)
	}
	if got := rec.Header().Get("Set-Cookie"); got != "" {
		t.Fatalf("cover must set no cookies, got %q", got)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("cover must set no CORS headers, got %q", got)
	}
	if got := rec.Header().Get("X-Powered-By"); got != "" {
		t.Fatalf("cover must not reveal implementation, got %q", got)
	}
	if strings.Contains(rec.Header().Get("Server"), "caddy") || strings.Contains(rec.Header().Get("Server"), "go") {
		t.Fatalf("identifying Server header: %q", rec.Header().Get("Server"))
	}

	for path, want := range map[string]int{
		"/about": 200, "/contact": 200, "/robots.txt": 200,
		"/sitemap.xml": 200, "/feed.xml": 200,
		"/admin": 404, "/wp-login.php": 404, "/.env": 404,
		"/some/random/deep/link": 404,
	} {
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		if rec.Code != want {
			t.Fatalf("%s status = %d, want %d", path, rec.Code, want)
		}
		if want == 404 && !strings.Contains(rec.Body.String(), "صفحهٔ مورد نظر یافت نشد") {
			t.Fatalf("%s must render the human 404 page", path)
		}
	}

	rec = httptest.NewRecorder()
	srv.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/", nil))
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("POST status = %d, want 405", rec.Code)
	}
}

func TestStaleStoreStillRenders(t *testing.T) {
	// Store error (sources down): the site must still render shell pages,
	// never an empty or broken page (CAMO-004).
	srv := &Server{Renderer: rendererFor(PersonaNewsAgency, &fakeStore{err: errors.New("db down")})}
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != 200 || !strings.Contains(rec.Body.String(), "خبرگزاری منطقه‌ای پارس‌ران") {
		t.Fatal("site must keep serving last-known shell when store errors")
	}
}

func TestJalaliConversionKnownDates(t *testing.T) {
	cases := []struct {
		greg time.Time
		want string
	}{
		{time.Date(2024, 3, 20, 12, 0, 0, 0, time.UTC), "۱ فروردین ۱۴۰۳"}, // Nowruz 1403
		{time.Date(2025, 3, 20, 12, 0, 0, 0, time.UTC), "۳۰ اسفند ۱۴۰۳"},  // 1403 was a leap year
		{time.Date(2025, 3, 21, 12, 0, 0, 0, time.UTC), "۱ فروردین ۱۴۰۴"}, // Nowruz 1404
	}
	for _, tc := range cases {
		jy, jm, jd := Jalali(tc.greg)
		got := PersianDigits(itoa(jd)) + " " + jalaliMonths[jm-1] + " " + PersianDigits(itoa(jy))
		if got != tc.want {
			t.Fatalf("Jalali(%s) = %q, want %q", tc.greg, got, tc.want)
		}
	}
}

func itoa(v int) string {
	if v == 0 {
		return "0"
	}
	digits := ""
	for v > 0 {
		digits = string(rune('0'+v%10)) + digits
		v /= 10
	}
	return digits
}

func TestPersianDigits(t *testing.T) {
	if got := PersianDigits("1405/06/24"); got != "۱۴۰۵/۰۶/۲۴" {
		t.Fatalf("PersianDigits = %q", got)
	}
}
