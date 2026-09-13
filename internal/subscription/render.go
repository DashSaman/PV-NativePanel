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
	FamilySingBox Family = "singbox" // sing-box / Karing
	FamilyHiddify Family = "hiddify" // Hiddify (sing-box based render)
	FamilyV2Ray   Family = "v2ray"   // v2rayNG / generic base64 link lists
)

// DetectFamily maps a User-Agent header to a render family. The empty or
// unrecognized agent maps to FamilyNaive so existing raw clients keep their
// current behavior unchanged.
func DetectFamily(userAgent string) Family {
	ua := strings.ToLower(userAgent)
	switch {
	case strings.Contains(ua, "clash"), strings.Contains(ua, "mihomo"), strings.Contains(ua, "stash"):
		return FamilyClash
	case strings.Contains(ua, "sing-box"), strings.Contains(ua, "singbox"), strings.Contains(ua, "karing"):
		return FamilySingBox
	case strings.Contains(ua, "hiddify"):
		return FamilyHiddify
	case strings.Contains(ua, "v2ray"), strings.Contains(ua, "nekoray"), strings.Contains(ua, "neko"):
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
	case "singbox", "sing-box", "karing":
		return FamilySingBox, nil
	case "hiddify":
		return FamilyHiddify, nil
	case "v2ray", "base64":
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

// naiveURI renders the machine URI for one node (naive+https scheme).
func (n Node) naiveURI() (string, error) {
	if err := n.validate(); err != nil {
		return "", err
	}
	uri, err := BuildNaiveURI(n.Username, n.Password, n.Host+":"+strconv.Itoa(n.Port))
	if err != nil {
		return "", err
	}
	return uri, nil
}

// clashProxy is the Mihomo/Clash naive proxy map. Field names follow the
// Clash.Meta (mihomo) naive proxy schema.
type clashProxy struct {
	Name     string `yaml:"name"`
	Type     string `yaml:"type"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	SNI      string `yaml:"sni,omitempty"`
}

type clashGroup struct {
	Name    string   `yaml:"name"`
	Type    string   `yaml:"type"`
	Use     []string `yaml:"use,omitempty"`
	Proxies []string `yaml:"proxies,omitempty"`
	URL     string   `yaml:"url"`
	Inter   int      `yaml:"interval"`
	Tol     int      `yaml:"tolerance,omitempty"`
	Strat   string   `yaml:"strategy,omitempty"`
}

// RenderClash renders a Mihomo/Clash profile for the given nodes:
//   - one naive proxy per node (inline; proxy-provider form arrives with the
//     R4/R5 pool registry — documented in docs/STEERING_SPEC_FA.md §3);
//   - a url-test group with tolerance 50ms / interval 300s (STEER-005 anti-flap);
//   - a load-balance round-robin group over the same nodes;
//   - never sets interrupt-exist-connections anywhere.
func RenderClash(nodes []Node) ([]byte, error) {
	if len(nodes) == 0 {
		return nil, errors.New("subscription: at least one node is required")
	}
	proxies := make([]clashProxy, 0, len(nodes))
	names := make([]string, 0, len(nodes))
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
		names = append(names, n.displayName())
	}

	const healthURL = "http://www.gstatic.com/generate_204"
	doc := map[string]any{
		"proxies": proxies,
		"proxy-groups": []clashGroup{
			{Name: "PV-AUTO", Type: "url-test", Proxies: names, URL: healthURL, Inter: 300, Tol: 50},
			{Name: "PV-RR", Type: "load-balance", Proxies: names, URL: healthURL, Inter: 300, Strat: "round-robin"},
		},
		"rules": []string{"MATCH,PV-AUTO"},
	}
	out, err := yaml.Marshal(doc)
	if err != nil {
		return nil, fmt.Errorf("subscription: render clash: %w", err)
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

// RenderSingBox renders a sing-box (and Karing/Hiddify) profile: naive
// outbounds for every node, a urltest group (interval 5m, tolerance 50ms) and
// a selector whose default is the auto group. interrupt_exist_connections is
// never set to true (STEER-005 hot-update rule).
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
func RenderMachinePayload(family Family, nodes []Node) ([]byte, string, error) {
	switch family {
	case FamilyClash:
		body, err := RenderClash(nodes)
		return body, "application/yaml; charset=utf-8", err
	case FamilySingBox, FamilyHiddify:
		body, err := RenderSingBox(nodes)
		return body, "application/json; charset=utf-8", err
	case FamilyV2Ray:
		body, err := RenderBase64List(nodes)
		return body, "text/plain; charset=utf-8", err
	case FamilyNaive:
		if len(nodes) == 0 {
			return nil, "", errors.New("subscription: at least one node is required")
		}
		uri, err := nodes[0].naiveURI()
		if err != nil {
			return nil, "", err
		}
		return []byte(uri + "\n"), "text/plain; charset=utf-8", nil
	default:
		return nil, "", fmt.Errorf("subscription: unsupported family %q", family)
	}
}
