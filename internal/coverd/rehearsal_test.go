package coverd

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// R6 live rehearsal (CAMO-001/002): the cover origin must behave like a
// real small Persian site behind Caddy — distinct, stable per-node personas;
// hygienic responses (no cookies, no CORS, no identifying headers, no panel
// vocabulary); human error pages; robots allowed. The routing flip on a live
// node is a pure Caddy map from root → this loopback origin.

type rehearsalStore struct {
	items []Item
	err   error
}

func (s *rehearsalStore) Latest(nodeID string, limit int) ([]Item, error) {
	if s.err != nil {
		return nil, s.err
	}
	if limit > len(s.items) {
		limit = len(s.items)
	}
	return s.items[:limit], nil
}

func rehearsalRenderer(nodeID string, store ContentStore) *Renderer {
	return &Renderer{
		NodeID:  nodeID,
		Persona: PersonaForNode(nodeID),
		Store:   store,
		Now:     func() time.Time { return time.Date(2026, 9, 14, 12, 0, 0, 0, time.UTC) },
	}
}

func rehearsalFetch(t *testing.T, srv *Server, target string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, target, nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	return rec
}

func TestRehearsalPersonaStableDistinctNatural(t *testing.T) {
	a1 := PersonaForNode("node-alpha")
	a2 := PersonaForNode("node-alpha")
	if a1.ID != a2.ID {
		t.Fatal("persona assignment must be stable per node across restarts")
	}
	distinct := map[PersonaID]bool{}
	for i := 0; i < 48; i++ {
		distinct[PersonaForNode("node-pool-"+strings.Repeat("x", i%7)+"-"+string(rune('a'+i%26))).ID] = true
	}
	if len(distinct) < 3 {
		t.Fatalf("per-node diversity too narrow: %d packs", len(distinct))
	}
	// Every assigned persona must be a registered pack (no invented personas).
	for id := range distinct {
		if _, ok := PersonaByID(id); !ok {
			t.Fatalf("persona %q is not a registered pack", id)
		}
	}
}

func TestRehearsalHomeRendersPersianSiteWithHygiene(t *testing.T) {
	store := &rehearsalStore{items: []Item{
		{Source: "iranwire", Kind: "news", Title: "گزارش میدانی", URL: "https://example.invalid/a", PublishedAt: ptrTime(time.Date(2026, 9, 13, 8, 0, 0, 0, time.UTC))},
		{Source: "youtube", Kind: "video", Title: "پخش هفتگی", URL: "https://example.invalid/v"},
	}}
	srv := &Server{Renderer: rehearsalRenderer("node-alpha", store)}
	rec := rehearsalFetch(t, srv, "/")

	if rec.Code != http.StatusOK {
		t.Fatalf("home status = %d", rec.Code)
	}
	h := rec.Header()
	if got := h.Get("Set-Cookie"); got != "" {
		t.Fatalf("cover must not set cookies: %q", got)
	}
	if got := h.Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("cover must not send CORS headers: %q", got)
	}
	if got := h.Get("Server"); got == "" || strings.Contains(strings.ToLower(got), "pvnaive") {
		t.Fatalf("Server header must be generic: %q", got)
	}
	if !strings.Contains(h.Get("Content-Type"), "utf-8") {
		t.Fatalf("content type must be utf-8 html: %q", h.Get("Content-Type"))
	}
	body := rec.Body.String()
	for _, forbidden := range []string{"pvnaive", "PvNaive", "panel", "/api/", "session", "subscription"} {
		if strings.Contains(strings.ToLower(body), strings.ToLower(forbidden)) {
			t.Fatalf("cover body leaks identifying vocabulary %q", forbidden)
		}
	}
	if !strings.Contains(body, "<html") || !strings.Contains(body, "dir=\"rtl\"") {
		t.Fatal("cover home must be a Persian RTL page")
	}
}

func TestRehearsalStoreFailureStaysHygienic(t *testing.T) {
	srv := &Server{Renderer: rehearsalRenderer("node-beta", &rehearsalStore{err: errors.New("db down")})}
	for _, target := range []string{"/", "/about", "/contact"} {
		rec := rehearsalFetch(t, srv, target)
		if rec.Code != http.StatusOK {
			t.Fatalf("%s on store failure: status %d (must degrade to empty natural page, not 5xx)", target, rec.Code)
		}
		body := strings.ToLower(rec.Body.String())
		if strings.Contains(body, "pvnaive") || strings.Contains(body, "panel") {
			t.Fatalf("%s leaks identity on failure", target)
		}
	}
}

func TestRehearsalErrorPagesHumanAndHygienic(t *testing.T) {
	srv := &Server{Renderer: rehearsalRenderer("node-gamma", &rehearsalStore{items: nil})}
	notFound := rehearsalFetch(t, srv, "/definitely-missing")
	if notFound.Code != http.StatusNotFound {
		t.Fatalf("missing page status = %d", notFound.Code)
	}
	if strings.Contains(strings.ToLower(notFound.Body.String()), "pvnaive") {
		t.Fatal("404 must not identify the product")
	}
	badMethod := httptest.NewRequest(http.MethodPost, "/", nil)
	recPost := httptest.NewRecorder()
	srv.ServeHTTP(recPost, badMethod)
	if recPost.Code != http.StatusMethodNotAllowed {
		t.Fatalf("POST / status = %d", recPost.Code)
	}
	robots := rehearsalFetch(t, srv, "/robots.txt")
	if robots.Code != http.StatusOK || !strings.Contains(robots.Body.String(), "User-agent") {
		t.Fatal("robots.txt must render for crawlers")
	}
}

func ptrTime(t time.Time) *time.Time { return &t }
