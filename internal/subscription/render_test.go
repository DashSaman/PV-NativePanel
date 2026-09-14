package subscription

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/DashSaman/PV-NaivePanel/internal/runtimecred"
	"gopkg.in/yaml.v3"
)

func sampleNodes() []Node {
	return []Node{
		{Name: "PVNaive", Host: "203.0.113.10", Port: 443, Username: "user-1", Password: "secret-1", SNI: "placeholder.example"},
		{Name: "PVNaive-2", Host: "203.0.113.11", Port: 443, Username: "user-2", Password: "secret-2", SNI: "203.0.113.11"},
	}
}

func TestDetectFamily(t *testing.T) {
	cases := map[string]Family{
		"ClashMetaForAndroid/2.10.7":        FamilyClash,
		"mihomo/1.18.1":                     FamilyClash,
		"Stash/2.6.0 (iOS)":                 FamilyClash,
		"sing-box 1.8.0 (darwin)":           FamilySingBox,
		"Karing/1.0.30 (Android)":           FamilySingBox,
		"Hiddify-Next/2.0.5":                FamilyHiddify,
		"v2rayNG/1.8.23":                    FamilyV2Ray,
		"NekoBox/1.2.0":                     FamilyV2Ray,
		"":                                  FamilyNaive,
		"Mozilla/5.0 (Windows NT 10.0;...)": FamilyNaive,
		"naive/1.0":                         FamilyNaive,
	}
	for ua, want := range cases {
		if got := DetectFamily(ua); got != want {
			t.Fatalf("DetectFamily(%q) = %q, want %q", ua, got, want)
		}
	}
}

func TestFamilyFromQuery(t *testing.T) {
	if f, err := FamilyFromQuery(nil); err != nil || f != "" {
		t.Fatalf("empty query must return no override: %q %v", f, err)
	}
	if f, err := FamilyFromQuery(map[string][]string{"family": {"Mihomo"}}); err != nil || f != FamilyClash {
		t.Fatalf("mihomo override = %q %v", f, err)
	}
	if _, err := FamilyFromQuery(map[string][]string{"family": {"trojan"}}); err == nil {
		t.Fatal("unknown override must error")
	}
}

func TestRenderClashStructure(t *testing.T) {
	body, err := RenderClash(sampleNodes())
	if err != nil {
		t.Fatal(err)
	}
	text := string(body)
	if strings.Contains(text, "interrupt-exist-connections") {
		t.Fatal("clash render must never set interrupt-exist-connections")
	}
	var doc struct {
		Proxies []struct {
			Name     string `yaml:"name"`
			Type     string `yaml:"type"`
			Host     string `yaml:"host"`
			Port     int    `yaml:"port"`
			Username string `yaml:"username"`
			SNI      string `yaml:"sni"`
		} `yaml:"proxies"`
		Groups []struct {
			Name      string   `yaml:"name"`
			Type      string   `yaml:"type"`
			Proxies   []string `yaml:"proxies"`
			URL       string   `yaml:"url"`
			Interval  int      `yaml:"interval"`
			Tolerance int      `yaml:"tolerance"`
			Strategy  string   `yaml:"strategy"`
		} `yaml:"proxy-groups"`
		Rules []string `yaml:"rules"`
	}
	if err := yaml.Unmarshal(body, &doc); err != nil {
		t.Fatal(err)
	}
	if len(doc.Proxies) != 2 || doc.Proxies[0].Type != "naive" || doc.Proxies[0].Host != "203.0.113.10" || doc.Proxies[0].Port != 443 {
		t.Fatalf("proxies render wrong: %+v", doc.Proxies)
	}
	var urlGroup, balanceGroup *struct {
		Type      string
		URL       string
		Interval  int
		Tolerance int
		Strategy  string
		Proxies   []string
	}
	for i := range doc.Groups {
		g := doc.Groups[i]
		switch g.Type {
		case "url-test":
			urlGroup = &struct {
				Type      string
				URL       string
				Interval  int
				Tolerance int
				Strategy  string
				Proxies   []string
			}{g.Type, g.URL, g.Interval, g.Tolerance, g.Strategy, g.Proxies}
		case "load-balance":
			balanceGroup = &struct {
				Type      string
				URL       string
				Interval  int
				Tolerance int
				Strategy  string
				Proxies   []string
			}{g.Type, g.URL, g.Interval, g.Tolerance, g.Strategy, g.Proxies}
		}
	}
	if urlGroup == nil || urlGroup.Tolerance != 50 || urlGroup.Interval != 300 || urlGroup.URL != "http://www.gstatic.com/generate_204" {
		t.Fatalf("url-test group must use tolerance 50 / interval 300: %+v", urlGroup)
	}
	if len(urlGroup.Proxies) != 2 {
		t.Fatalf("url-test group must reference both nodes: %+v", urlGroup.Proxies)
	}
	if balanceGroup == nil || balanceGroup.Strategy != "round-robin" {
		t.Fatalf("load-balance group must be round-robin: %+v", balanceGroup)
	}
	if len(doc.Rules) != 1 || doc.Rules[0] != "MATCH,PV-AUTO" {
		t.Fatalf("rules wrong: %+v", doc.Rules)
	}
}

func TestRenderSingBoxStructure(t *testing.T) {
	body, err := RenderSingBox(sampleNodes())
	if err != nil {
		t.Fatal(err)
	}
	var doc struct {
		Outbounds []struct {
			Type       string `json:"type"`
			Tag        string `json:"tag"`
			Server     string `json:"server"`
			ServerPort int    `json:"server_port"`
			Username   string `json:"username"`
			TLS        *struct {
				Enabled    bool   `json:"enabled"`
				ServerName string `json:"server_name"`
			} `json:"tls"`
			URL       string `json:"url"`
			Interval  string `json:"interval"`
			Tolerance int    `json:"tolerance"`
			Default   string `json:"default"`
			Interrupt *bool  `json:"interrupt_exist_connections"`
		} `json:"outbounds"`
		Route struct {
			Final string `json:"final"`
		} `json:"route"`
	}
	if err := json.Unmarshal(body, &doc); err != nil {
		t.Fatal(err)
	}
	if len(doc.Outbounds) != 4 {
		t.Fatalf("want 2 nodes + urltest + selector, got %d", len(doc.Outbounds))
	}
	first := doc.Outbounds[0]
	if first.Type != "naive" || first.Server != "203.0.113.10" || first.ServerPort != 443 || first.Username != "user-1" {
		t.Fatalf("naive outbound wrong: %+v", first)
	}
	if first.TLS == nil || !first.TLS.Enabled || first.TLS.ServerName != "placeholder.example" {
		t.Fatalf("tls block wrong: %+v", first.TLS)
	}
	var urltest *struct {
		Type      string `json:"type"`
		URL       string `json:"url"`
		Interval  string `json:"interval"`
		Tolerance int    `json:"tolerance"`
		Interrupt *bool  `json:"interrupt_exist_connections"`
	}
	var selector *struct {
		Type    string `json:"type"`
		Default string `json:"default"`
	}
	for i := range doc.Outbounds {
		switch doc.Outbounds[i].Type {
		case "urltest":
			urltest = &struct {
				Type      string `json:"type"`
				URL       string `json:"url"`
				Interval  string `json:"interval"`
				Tolerance int    `json:"tolerance"`
				Interrupt *bool  `json:"interrupt_exist_connections"`
			}{doc.Outbounds[i].Type, doc.Outbounds[i].URL, doc.Outbounds[i].Interval, doc.Outbounds[i].Tolerance, doc.Outbounds[i].Interrupt}
		case "selector":
			selector = &struct {
				Type    string `json:"type"`
				Default string `json:"default"`
			}{doc.Outbounds[i].Type, doc.Outbounds[i].Default}
		}
	}
	if urltest == nil || urltest.Tolerance != 50 || urltest.Interval != "5m" || urltest.URL != "http://www.gstatic.com/generate_204" {
		t.Fatalf("urltest group wrong: %+v", urltest)
	}
	if urltest.Interrupt == nil || *urltest.Interrupt {
		t.Fatalf("urltest must never interrupt existing connections: %+v", urltest.Interrupt)
	}
	if selector == nil || selector.Default != "PV-AUTO" {
		t.Fatalf("selector wrong: %+v", selector)
	}
	if doc.Route.Final != "PV" {
		t.Fatalf("route final wrong: %q", doc.Route.Final)
	}
}

func TestRenderBase64ListPrimaryFirst(t *testing.T) {
	body, err := RenderBase64List(sampleNodes())
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := base64.StdEncoding.DecodeString(string(body))
	if err != nil {
		t.Fatalf("payload must be standard base64: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(decoded)), "\n")
	if len(lines) != 2 {
		t.Fatalf("want 2 links, got %d", len(lines))
	}
	if !strings.HasPrefix(lines[0], "naive+https://user-1:secret-1@203.0.113.10:443") {
		t.Fatalf("first link must be the primary node: %q", lines[0])
	}
	if !strings.HasPrefix(lines[1], "naive+https://user-2:secret-2@203.0.113.11:443") {
		t.Fatalf("second link wrong: %q", lines[1])
	}
}

func TestRenderMachinePayloadFamilies(t *testing.T) {
	nodes := sampleNodes()
	cases := map[Family]string{
		FamilyClash:   "application/yaml; charset=utf-8",
		FamilySingBox: "application/json; charset=utf-8",
		FamilyHiddify: "application/json; charset=utf-8",
		FamilyV2Ray:   "text/plain; charset=utf-8",
		FamilyNaive:   "text/plain; charset=utf-8",
	}
	for family, ctype := range cases {
		body, got, err := RenderMachinePayload(family, nodes)
		if err != nil {
			t.Fatalf("%s: %v", family, err)
		}
		if got != ctype {
			t.Fatalf("%s content type = %q, want %q", family, got, ctype)
		}
		if len(body) == 0 {
			t.Fatalf("%s: empty body", family)
		}
	}
	if _, _, err := RenderMachinePayload(FamilyClash, nil); err == nil {
		t.Fatal("empty node list must error")
	}
}

func TestUserinfoHeaderFormat(t *testing.T) {
	got := UserinfoHeader(100, 200, 300, 1700000000)
	want := "upload=100; download=200; total=300; expire=1700000000"
	if got != want {
		t.Fatalf("userinfo = %q, want %q", got, want)
	}
}

func TestNodeValidationAndURIErrors(t *testing.T) {
	bad := []Node{{Name: "x", Host: "", Port: 443, Username: "u", Password: "p"}}
	if _, err := RenderClash(bad); err == nil {
		t.Fatal("empty host must error")
	}
	worse := []Node{{Name: "x", Host: "h", Port: 443, Username: "", Password: "p"}}
	if _, err := RenderBase64List(worse); err == nil {
		t.Fatal("missing username must error")
	}
	oddPort := []Node{{Name: "x", Host: "h", Port: 99999, Username: "u", Password: "p"}}
	if _, err := RenderSingBox(oddPort); err == nil {
		t.Fatal("out-of-range port must error")
	}
}

func TestResolveProfileFillsRenderNode(t *testing.T) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i + 1)
	}
	ciphertext, nonce, err := runtimecred.EncryptSecret(key, []byte("render-pass"))
	if err != nil {
		t.Fatal(err)
	}
	rawBytes := make([]byte, 32)
	raw := base64.RawURLEncoding.EncodeToString(rawBytes)
	hash := sha256.Sum256(rawBytes)
	expires := time.Now().UTC().Add(24 * time.Hour)
	store := &fakeStore{wantHash: hash, record: Record{
		RuntimeCredentialID: "runtime-1",
		Username:            "customer.name",
		SecretCiphertext:    ciphertext,
		SecretNonce:         nonce,
		EncryptionKeyID:     "runtime-v1",
		UserState:           "active",
		TermState:           "active",
		ExpiresAt:           &expires,
	}}
	service, err := NewService(store, key, "runtime-v1")
	if err != nil {
		t.Fatal(err)
	}
	profile, err := service.ResolveProfile(context.Background(), raw, "203.0.113.10:443")
	if err != nil {
		t.Fatal(err)
	}
	if profile.Node == nil {
		t.Fatal("available profile must carry the render node")
	}
	if profile.Node.Host != "203.0.113.10" || profile.Node.Port != 443 {
		t.Fatalf("node host/port = %s:%d", profile.Node.Host, profile.Node.Port)
	}
	if profile.Node.Username != "customer.name" || profile.Node.Password != "render-pass" {
		t.Fatal("node credentials wrong")
	}
	if profile.Node.SNI != "203.0.113.10" {
		t.Fatalf("node SNI = %q, want host fallback", profile.Node.SNI)
	}

	// Host without an explicit port defaults to 443.
	profile2, err := service.ResolveProfile(context.Background(), raw, "proxy.example.com")
	if err != nil {
		t.Fatal(err)
	}
	if profile2.Node == nil || profile2.Node.Port != 443 || profile2.Node.Host != "proxy.example.com" {
		t.Fatalf("default port render node wrong: %+v", profile2.Node)
	}
}
