package subscription

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// Family identifies a subscription client family for User-Agent content
// negotiation on the machine endpoint (/sub/<token>). The human page
// (/s/<token>) never changes behavior (STEER-003).
type Family string

const (
	FamilyNaive   Family = "naive"   // raw naive+https URI (default, backwards compatible)
	FamilyClash   Family = "clash"   // Mihomo / Clash.Meta / Stash
	FamilySingBox Family = "singbox" // sing-box JSON profile (Karing/Hiddify import base64 links instead)
	FamilyHiddify Family = "hiddify" // Hiddify (sing-box based render; kept for override back-compat)
	FamilyV2Ray   Family = "v2ray"   // v2rayNG / NekoBox / Karing / Hiddify base64 link lists
)

// DetectFamily maps a User-Agent header to a render family. The empty or
// unrecognized agent maps to FamilyNaive so existing raw clients keep their
// current behavior unchanged.
func DetectFamily(userAgent string) Family {
	ua := strings.ToLower(userAgent)
	switch {
	case strings.Contains(ua, "clash"), strings.Contains(ua, "mihomo"), strings.Contains(ua, "stash"),
		strings.Contains(ua, "karing"):
		// Karing fetches subscriptions through its Clash engine and its own
		// Clash-compat table marks the naive outbound as supported; its
		// V2ray link parser does NOT understand naive+https URIs (field
		// report: the base64 list produced "clash proxies/proxy-providers:
		// No server available" on Karing 1.2.23.2606 Android), so Karing
		// rides the Clash family with the direct naive profile.
		return FamilyClash
	case strings.Contains(ua, "sing-box"), strings.Contains(ua, "singbox"):
		return FamilySingBox
	case strings.Contains(ua, "hiddify"),
		strings.Contains(ua, "v2ray"), strings.Contains(ua, "nekoray"), strings.Contains(ua, "neko"),
		strings.Contains(ua, "dart"):
		// Hiddify (sing-box GUI) imports the universal base64 link list;
		// its URI importer parses the naive+https entries.
		return FamilyV2Ray
	default:
		return FamilyNaive
	}
}

// FamilyFromQuery applies an explicit ?family= override (used by tests and by
// operators pre-viewing a format). Unknown values return an error.
func FamilyFromQuery(query url.Values) (Family, error) {
	switch strings.ToLower(strings.TrimSpace(query.Get("family"))) {
	case "":
		return "", nil
	case "naive", "raw":
		return FamilyNaive, nil
	case "clash", "mihomo":
		return FamilyClash, nil
	case "singbox", "sing-box":
		return FamilySingBox, nil
	// karing mirrors the actual UA negotiation (Clash family, direct
	// naive profile). hiddify keeps the universal base64 link list.
	case "karing":
		return FamilyClash, nil
	case "hiddify", "v2ray", "base64":
		return FamilyV2Ray, nil
	default:
		return "", fmt.Errorf("subscription: unknown family override")
	}
}

// Node is one proxy endpoint to render. Nodes are passed in preference order
// (the steering primary first) and renderers preserve that order. Hosts are
// hardcoded IP/host + port + SNI in the rendered config — never a hostname
// that requires DNS resolution at connect time (STEER-005 no-DNS rule).
type Node struct {
	Name     string
	Host     string
	Port     int
	Username string
	Password string
	SNI      string
}

func (n Node) validate() error {
	if strings.TrimSpace(n.Host) == "" || n.Port <= 0 || n.Port > 65535 {
		return errors.New("subscription: node host and port are required")
	}
	if n.Username == "" || n.Password == "" {
		return errors.New("subscription: node credentials are required")
	}
	if n.SNI == "" {
		n.SNI = n.Host
	}
	return nil
}

func (n Node) displayName() string {
	if name := strings.TrimSpace(n.Name); name != "" {
		return name
	}
	return n.Host
}

// naiveURI renders the machine URI for one node (naive+https scheme) with
// the node display name as URI fragment (spec §3: naive://USER:PASS@IP:443#NAME
// — clients show the fragment as the node label; percent-encoding is handled
// by url.URL). The human-facing DirectURI (service.go) deliberately keeps no
// fragment.
func (n Node) naiveURI() (string, error) {
	if err := n.validate(); err != nil {
		return "", err
	}
	uri, err := BuildNaiveURI(n.Username, n.Password, n.Host+":"+strconv.Itoa(n.Port))
	if err != nil {
		return "", err
	}
	parsed, err := url.Parse(uri)
	if err != nil {
		return "", fmt.Errorf("subscription: render naive uri: %w", err)
	}
	parsed.Fragment = n.displayName()
	return parsed.String(), nil
}

// RenderOptions carries the render context for machine payloads.
type RenderOptions struct {
	// ProviderURL is the canonical absolute /sub/<token>?family=mihomo URL
	// embedded in the Mihomo proxy-provider block (STEER-003 hot-update
	// source). Required for the FamilyClash full profile.
	ProviderURL string
	// ProviderPayload reports the ?family=mihomo explicit override: the fetch
	// is Mihomo's proxy-provider engine pulling the node list, so the response
	// must be the bare proxies document instead of the full profile.
	ProviderPayload bool
}

// ProviderPayloadOverride reports whether the ?family=mihomo override was
// passed explicitly. The spec §3 self-referential provider URL
// (…/sub/<token>?family=mihomo) is fetched by Mihomo's proxy-provider
// engine; those fetches receive the provider payload (a bare proxies list)
// while profile fetches receive the full profile embedding that URL.
func ProviderPayloadOverride(query url.Values) bool {
	return strings.ToLower(strings.TrimSpace(query.Get("family"))) == "mihomo"
}

// clashProxy is the Mihomo/Clash naive proxy map. Field names follow the
// Clash.Meta (mihomo) naive proxy schema.
type clashProxy struct {
	Name     string `yaml:"name"`
	Type     string `yaml:"type"`
	Host     string `yaml:"server"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	SNI      string `yaml:"sni,omitempty"`
}

type clashGroup struct {
	Name    string   `yaml:"name"`
	Type    string   `yaml:"type"`
	Proxies []string `yaml:"proxies,omitempty"`
	Use     []string `yaml:"use,omitempty"`
	URL     string   `yaml:"url"`
	Inter   int      `yaml:"interval"`
	Tol     int      `yaml:"tolerance,omitempty"`
	Strat   string   `yaml:"strategy,omitempty"`
}

const clashHealthURL = "http://www.gstatic.com/generate_204"

// clashProxies validates the node set and renders the mihomo naive proxy
// list (preference order preserved: steering primary first).
func clashProxies(nodes []Node) ([]clashProxy, error) {
	proxies := make([]clashProxy, 0, len(nodes))
	for _, n := range nodes {
		if err := n.validate(); err != nil {
			return nil, err
		}
		sni := strings.TrimSpace(n.SNI)
		if sni == "" {
			sni = n.Host
		}
		proxies = append(proxies, clashProxy{
			Name: n.displayName(), Type: "naive",
			Host: n.Host, Port: n.Port,
			Username: n.Username, Password: n.Password, SNI: sni,
		})
	}
	return proxies, nil
}

// RenderClashProviderPayload renders the canonical Mihomo proxy-provider
// payload: a bare `proxies:` document (STEER-003). It is served to the
// fetches triggered by the proxy-providers.pvnaive.url inside the full
// profile, so provider updates stay hot — the client refreshes its node
// list without reloading the profile.
func RenderClashProviderPayload(nodes []Node) ([]byte, error) {
	if len(nodes) == 0 {
		return nil, errors.New("subscription: at least one node is required")
	}
	proxies, err := clashProxies(nodes)
	if err != nil {
		return nil, err
	}
	out, err := yaml.Marshal(map[string]any{"proxies": proxies})
	if err != nil {
		return nil, fmt.Errorf("subscription: render clash provider payload: %w", err)
	}
	return out, nil
}

// RenderClashProfile renders the full Mihomo/Clash profile per spec §3:
//   - the node set arrives exclusively through the pvnaive http
//     proxy-provider (interval 4h + gstatic health-check) so updates are hot
//     and the profile itself never needs a reload;
//   - a url-test group with tolerance 50ms / interval 300s (STEER-005
//     anti-flap) and a load-balance round-robin group, both `use` the
//     provider;
//   - interrupt-exist-connections is never set anywhere.
//
// providerURL is the absolute canonical /sub/<token>?family=mihomo URL the
// client re-fetches on the provider interval.
func RenderClashProfile(providerURL string, nodes []Node) ([]byte, error) {
	if len(nodes) == 0 {
		return nil, errors.New("subscription: at least one node is required")
	}
	if _, err := clashProxies(nodes); err != nil {
		return nil, err
	}
	providerURL = strings.TrimSpace(providerURL)
	if providerURL == "" {
		return nil, errors.New("subscription: clash provider url is required")
	}
	// interval 14400s == the spec's 4h, emitted as integer seconds because
	// every parser in the Clash family (Mihomo/Stash/legacy) accepts it,
	// while duration strings predate some of them.
	doc := map[string]any{
		"proxy-providers": map[string]any{
			"pvnaive": map[string]any{
				"type":     "http",
				"url":      providerURL,
				"interval": 14400,
				"health-check": map[string]any{
					"enable":   true,
					"url":      clashHealthURL,
					"interval": 300,
				},
			},
		},
		"proxy-groups": []clashGroup{
			{Name: "PV-AUTO", Type: "url-test", Use: []string{"pvnaive"}, URL: clashHealthURL, Inter: 300, Tol: 50},
			{Name: "PV-RR", Type: "load-balance", Use: []string{"pvnaive"}, URL: clashHealthURL, Inter: 300, Strat: "round-robin"},
		},
		"rules": []string{"MATCH,PV-AUTO"},
	}
	out, err := yaml.Marshal(doc)
	if err != nil {
		return nil, fmt.Errorf("subscription: render clash profile: %w", err)
	}
	return out, nil
}

// RenderClashDirectProfile renders a Clash profile whose node list is
// embedded directly (no proxy-providers): naive proxies in preference
// order, a PV-AUTO url-test group (interval 5m, tolerance 50ms,
// STEER-005 anti-flap) and a MATCH rule. This is the payload Karing
// (and other Clash-family clients that support the naive proxy type,
// e.g. Stash) import: it avoids Karing's unsupported load-balance
// group and needs no provider engine. Real mihomo cores cannot speak
// naive at all - no payload fixes that; the /s page steers those
// users to Karing or NekoBox instead.
func RenderClashDirectProfile(nodes []Node) ([]byte, error) {
	if len(nodes) == 0 {
		return nil, errors.New("subscription: at least one node is required")
	}
	proxies, err := clashProxies(nodes)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(nodes))
	for _, n := range nodes {
		names = append(names, n.displayName())
	}
	doc := map[string]any{
		"proxies": proxies,
		"proxy-groups": []clashGroup{
			{Name: "PV-AUTO", Type: "url-test", Proxies: names, URL: clashHealthURL, Inter: 300, Tol: 50},
		},
		"rules": []string{"MATCH,PV-AUTO"},
	}
	out, err := yaml.Marshal(doc)
	if err != nil {
		return nil, fmt.Errorf("subscription: render clash direct profile: %w", err)
	}
	return out, nil
}

// singBoxOutbound is the subset of the sing-box outbound schema we emit.
type singBoxOutbound struct {
	Type       string           `json:"type"`
	Tag        string           `json:"tag"`
	Server     string           `json:"server,omitempty"`
	ServerPort int              `json:"server_port,omitempty"`
	Username   string           `json:"username,omitempty"`
	Password   string           `json:"password,omitempty"`
	TLS        *singBoxTLS      `json:"tls,omitempty"`
	Outbounds  []string         `json:"outbounds,omitempty"`
	URL        string           `json:"url,omitempty"`
	Interval   string           `json:"interval,omitempty"`
	Tolerance  int              `json:"tolerance,omitempty"`
	Default    string           `json:"default,omitempty"`
	Interrupt  *bool            `json:"interrupt_exist_connections,omitempty"`
	Rules      []map[string]any `json:"rules,omitempty"`
	Final      string           `json:"final,omitempty"`
	AutoDetect *bool            `json:"auto_detect_interface,omitempty"`
}

type singBoxTLS struct {
	Enabled    bool   `json:"enabled"`
	ServerName string `json:"server_name"`
}

// RenderSingBox renders a sing-box JSON profile (genuine sing-box clients
// and the operator ?family=singbox preview): naive outbounds for every
// node, a urltest group (interval 5m, tolerance 50ms) and a selector whose
// default is the auto group. interrupt_exist_connections is never set to
// true (STEER-005 hot-update rule). Note: Karing/Hiddify are deliberately
// NOT served this payload - their importers reject the naive outbound type
// - they get RenderBase64List instead.
func RenderSingBox(nodes []Node) ([]byte, error) {
	if len(nodes) == 0 {
		return nil, errors.New("subscription: at least one node is required")
	}
	outbounds := []singBoxOutbound{}
	names := make([]string, 0, len(nodes))
	for _, n := range nodes {
		if err := n.validate(); err != nil {
			return nil, err
		}
		sni := strings.TrimSpace(n.SNI)
		if sni == "" {
			sni = n.Host
		}
		outbounds = append(outbounds, singBoxOutbound{
			Type: "naive", Tag: n.displayName(),
			Server: n.Host, ServerPort: n.Port,
			Username: n.Username, Password: n.Password,
			TLS: &singBoxTLS{Enabled: true, ServerName: sni},
		})
		names = append(names, n.displayName())
	}

	const healthURL = "http://www.gstatic.com/generate_204"
	neverInterrupt := false
	outbounds = append(outbounds,
		singBoxOutbound{
			Type: "urltest", Tag: "PV-AUTO", Outbounds: names,
			URL: healthURL, Interval: "5m", Tolerance: 50,
			Interrupt: &neverInterrupt,
		},
		singBoxOutbound{
			Type: "selector", Tag: "PV", Outbounds: append([]string{"PV-AUTO"}, names...),
			Default: "PV-AUTO", Interrupt: &neverInterrupt,
		},
	)
	doc := map[string]any{
		"outbounds": outbounds,
		"route": map[string]any{
			"rules":                 []map[string]any{},
			"final":                 "PV",
			"auto_detect_interface": true,
		},
	}
	out, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("subscription: render sing-box: %w", err)
	}
	return append(out, '\n'), nil
}

// RenderBase64List renders the v2rayNG / generic base64 subscription: the
// newline-joined node URIs (preference order, primary first), base64 encoded.
func RenderBase64List(nodes []Node) ([]byte, error) {
	if len(nodes) == 0 {
		return nil, errors.New("subscription: at least one node is required")
	}
	var lines []string
	for _, n := range nodes {
		uri, err := n.naiveURI()
		if err != nil {
			return nil, err
		}
		lines = append(lines, uri)
	}
	payload := strings.Join(lines, "\n")
	return []byte(base64.StdEncoding.EncodeToString([]byte(payload))), nil
}

// UserinfoHeader renders the mandatory subscription-userinfo value
// (bytes aggregated across nodes once the fleet lands; 0 = unlimited).
func UserinfoHeader(upload, download, total, expireUnix int64) string {
	return "upload=" + strconv.FormatInt(upload, 10) +
		"; download=" + strconv.FormatInt(download, 10) +
		"; total=" + strconv.FormatInt(total, 10) +
		"; expire=" + strconv.FormatInt(expireUnix, 10)
}

// RenderMachinePayload renders the negotiated family body.
func RenderMachinePayload(family Family, nodes []Node, opts RenderOptions) ([]byte, string, error) {
	switch family {
	case FamilyClash:
		if opts.ProviderPayload {
			body, err := RenderClashProviderPayload(nodes)
			return body, "application/yaml; charset=utf-8", err
		}
		// Direct embedded profile (Karing/Stash naive support; no
		// provider engine and no load-balance group Karing lacks).
		body, err := RenderClashDirectProfile(nodes)
		return body, "application/yaml; charset=utf-8", err
	case FamilySingBox, FamilyHiddify:
		body, err := RenderSingBox(nodes)
		return body, "application/json; charset=utf-8", err
	case FamilyV2Ray:
		body, err := RenderBase64List(nodes)
		return body, "text/plain; charset=utf-8", err
	case FamilyNaive:
		// spec §3: raw naive = current primary first, then alternates,
		// one naive:// URI per line (single node ⇒ single line).
		if len(nodes) == 0 {
			return nil, "", errors.New("subscription: at least one node is required")
		}
		lines := make([]string, 0, len(nodes))
		for _, n := range nodes {
			uri, err := n.naiveURI()
			if err != nil {
				return nil, "", err
			}
			lines = append(lines, uri)
		}
		return []byte(strings.Join(lines, "\n") + "\n"), "text/plain; charset=utf-8", nil
	default:
		return nil, "", fmt.Errorf("subscription: unsupported family %q", family)
	}
}
