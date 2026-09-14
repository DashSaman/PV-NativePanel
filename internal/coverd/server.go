package coverd

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"
)

// Server serves the cover site for one node over loopback (behind Caddy).
// Response hygiene (CAMO-001/002): no cookies, no CORS, no identifying
// headers, no directory listings, human error pages, zero panel references.
type Server struct {
	Renderer *Renderer
}

// NodeKey derives the per-node content key from a stable node identity. The
// key is not the raw node id to keep deployed identities out of dumps.
func NodeKey(nodeID string) string {
	h := sha256.Sum256([]byte("pvnaive/coverd/node/v1:" + nodeID))
	return hex.EncodeToString(h[:16])
}

func (s *Server) hygiene(w http.ResponseWriter) {
	h := w.Header()
	h.Set("Server", "web") // generic, non-identifying
	h.Set("X-Content-Type-Options", "nosniff")
	h.Set("Referrer-Policy", "no-referrer")
	// Deliberately absent: cookies, CORS headers, X-Powered-By, any panel hint.
}

func (s *Server) writeHTML(w http.ResponseWriter, status int, body string) {
	s.hygiene(w)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(body))
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimSuffix(r.URL.Path, "/")
	switch r.Method {
	case http.MethodGet, http.MethodHead:
		// serve below
	default:
		s.writeHTML(w, http.StatusMethodNotAllowed, s.Renderer.NotFound())
		return
	}

	switch path {
	case "", "/":
		s.writeHTML(w, http.StatusOK, s.Renderer.Home())
	case "/about":
		s.writeHTML(w, http.StatusOK, s.Renderer.About())
	case "/contact":
		s.writeHTML(w, http.StatusOK, s.Renderer.Contact())
	case "/robots.txt":
		s.hygiene(w)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(s.Renderer.Robots()))
	case "/sitemap.xml":
		s.hygiene(w)
		w.Header().Set("Content-Type", "application/xml; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(s.Renderer.Sitemap()))
	case "/feed.xml":
		s.hygiene(w)
		w.Header().Set("Content-Type", "application/rss+xml; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(s.Renderer.Feed()))
	default:
		// Human 404 for everything else, including probe-bait paths — the
		// response is indistinguishable from a normal Persian site's 404.
		s.writeHTML(w, http.StatusNotFound, s.Renderer.NotFound())
	}
}
