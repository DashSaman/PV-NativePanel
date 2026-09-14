package coverd

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type emptyStore struct{}

func (emptyStore) Latest(string, int) ([]Item, error) { return []Item{}, nil }

func TestFlipConfigValidate(t *testing.T) {
	// Default OFF is always valid (the flip is explicit).
	if err := (FlipConfig{}).Validate(); err != nil {
		t.Fatalf("default config must be valid: %v", err)
	}
	valid := FlipConfig{Enabled: true, NodeID: "node-alpha", Listen: "127.0.0.1:9444"}
	if err := valid.Validate(); err != nil {
		t.Fatalf("valid config rejected: %v", err)
	}
	bad := []FlipConfig{
		{Enabled: true, Listen: "127.0.0.1:9444"},                     // missing node id
		{Enabled: true, NodeID: "node-alpha"},                         // missing listen
		{Enabled: true, NodeID: "node-alpha", Listen: "0.0.0.0:9444"}, // non-loopback bind
		{Enabled: true, NodeID: "node-alpha", Listen: "127.0.0.1:0"},  // zero port
		{Enabled: true, NodeID: "node-alpha", Listen: "127.0.0.1:9444", PersonaOverride: "no-such-persona"},
		{Enabled: true, NodeID: strings.Repeat("x", 121), Listen: "127.0.0.1:9444"}, // oversized id
	}
	for index, config := range bad {
		if err := config.Validate(); err == nil {
			t.Fatalf("case %d unexpectedly accepted: %+v", index, config)
		}
	}
}

func TestPersonaForConfig(t *testing.T) {
	nodeID := "flip-node-alpha"
	derived, err := PersonaForConfig(FlipConfig{Enabled: true, NodeID: nodeID, Listen: "127.0.0.1:9444"}, nil)
	if err != nil {
		t.Fatalf("derived persona: %v", err)
	}
	if derived.ID != PersonaForNode(nodeID).ID {
		t.Fatalf("derived persona mismatch: %s vs %s", derived.ID, PersonaForNode(nodeID).ID)
	}
	stored, err := PersonaForConfig(
		FlipConfig{Enabled: true, NodeID: nodeID, Listen: "127.0.0.1:9444"},
		func() (string, bool) { return string(PersonaResearch), true },
	)
	if err != nil {
		t.Fatalf("stored persona: %v", err)
	}
	if stored.ID != PersonaResearch {
		t.Fatalf("stored persona must win over derivation, got %s", stored.ID)
	}
	overridden, err := PersonaForConfig(
		FlipConfig{Enabled: true, NodeID: nodeID, Listen: "127.0.0.1:9444", PersonaOverride: string(PersonaCharity)},
		func() (string, bool) { return string(PersonaResearch), true },
	)
	if err != nil {
		t.Fatalf("override persona: %v", err)
	}
	if overridden.ID != PersonaCharity {
		t.Fatalf("override must win over stored, got %s", overridden.ID)
	}
}

func TestNewNodeServerRendersCover(t *testing.T) {
	config := FlipConfig{Enabled: true, NodeID: "flip-node-beta", Listen: "127.0.0.1:9444"}
	server, err := NewNodeServer(config, emptyStore{}, nil, func() time.Time { return time.Unix(1726358400, 0) })
	if err != nil {
		t.Fatalf("build node server: %v", err)
	}
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("home status = %d, want 200", recorder.Code)
	}
	body := recorder.Body.String()
	if strings.Contains(body, "pvnaive") || strings.Contains(body, "panel") {
		t.Fatalf("cover markup leaks panel vocabulary")
	}
	if recorder.Header().Get("Server") != "web" {
		t.Fatalf("Server header = %q, want generic", recorder.Header().Get("Server"))
	}
	if recorder.Header().Get("Set-Cookie") != "" {
		t.Fatalf("cover must never set cookies")
	}
	// Probe-bait path stays a natural human 404.
	recorder = httptest.NewRecorder()
	server.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/wp-admin/setup.php", nil))
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("bait path status = %d, want 404", recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), "یافت نشد") {
		t.Fatalf("404 page lost its human error text")
	}
}

func TestNewNodeServerRejectsBadInput(t *testing.T) {
	if _, err := NewNodeServer(FlipConfig{Enabled: true, NodeID: "x", Listen: "127.0.0.1:9444", PersonaOverride: "bogus"}, emptyStore{}, nil, nil); err == nil {
		t.Fatalf("bogus persona override unexpectedly accepted")
	}
	valid := FlipConfig{Enabled: true, NodeID: "node", Listen: "127.0.0.1:9444"}
	if _, err := NewNodeServer(valid, nil, nil, nil); err == nil {
		t.Fatalf("nil store unexpectedly accepted")
	}
}
