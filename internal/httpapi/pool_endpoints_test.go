package httpapi

import (
	"strings"
	"testing"
	"time"

	"github.com/DashSaman/PV-NaivePanel/internal/fleet"
)

// The R5 pool surface is Owner-only, deny-by-default (GATE STEER-006).
func TestPoolRoutesAreRegisteredAndOwnerOnly(t *testing.T) {
	want := map[string]string{
		"pool.nodes.index":        "GET",
		"pool.enrolltoken.create": "POST",
		"pool.nodes.enroll":       "POST",
		"pool.revision.publish":   "POST",
		"pool.manifest.show":      "GET",
		"pool.maintenance.set":    "POST",
	}
	seen := 0
	for _, route := range Routes {
		method, ok := want[route.Name]
		if !ok {
			continue
		}
		seen++
		if route.Method != method {
			t.Fatalf("%s: method %s, want %s", route.Name, route.Method, method)
		}
		if route.Access != Owner {
			t.Fatalf("%s: access %q, want owner-only", route.Name, route.Access)
		}
		if !strings.HasPrefix(route.Path, "/api/v1/pool/") {
			t.Fatalf("%s: unexpected path %q", route.Name, route.Path)
		}
	}
	if seen != len(want) {
		t.Fatalf("pool routes registered = %d, want %d", seen, len(want))
	}
}

func testSigningKey(t *testing.T) (string, string) {
	t.Helper()
	pubHex, privHex, err := fleet.GenerateSigningKey()
	if err != nil {
		t.Fatalf("signing key: %v", err)
	}
	return pubHex, privHex
}

func validPoolRevisionPayload() poolRevisionPayload {
	return poolRevisionPayload{
		PoolID: "pool-eu-1",
		Endpoints: []poolEndpointInput{
			{Host: "198.51.100.10", Port: 443, SNI: "node1.example.invalid"},
		},
		Weight: 1,
	}
}

func TestBuildManifestDerivesOperatorKey(t *testing.T) {
	pubHex, privHex := testSigningKey(t)
	now := time.Now().UTC()
	m, err := buildManifest("node-1", validPoolRevisionPayload(), privHex, now)
	if err != nil {
		t.Fatalf("buildManifest: %v", err)
	}
	if m.PubKeyHex != pubHex {
		t.Fatal("manifest pubkey must be derived from the operator signing key")
	}
	if m.NodeID != "node-1" || m.PoolID != "pool-eu-1" || len(m.Endpoints) != 1 {
		t.Fatalf("manifest fields wrong: %+v", m)
	}
	if !m.ValidUntil.After(now.Add(manifestValidityDefault - time.Minute)) {
		t.Fatalf("default validity not applied: %s", m.ValidUntil)
	}
	env := fleet.ManifestEnvelope{Manifest: m}
	env.Signature, err = fleet.SignManifest(m, privHex)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	if err := fleet.VerifyEnvelope(env, now); err != nil {
		t.Fatalf("published manifest must verify: %v", err)
	}
}

func TestBuildManifestValidation(t *testing.T) {
	_, privHex := testSigningKey(t)
	now := time.Now().UTC()

	cases := []struct {
		name   string
		key    string
		mutate func(*poolRevisionPayload)
	}{
		{"zero weight", privHex, func(p *poolRevisionPayload) { p.Weight = 0 }},
		{"negative weight", privHex, func(p *poolRevisionPayload) { p.Weight = -1 }},
		{"huge weight", privHex, func(p *poolRevisionPayload) { p.Weight = 5000 }},
		{"empty pool id", privHex, func(p *poolRevisionPayload) { p.PoolID = " " }},
		{"long pool id", privHex, func(p *poolRevisionPayload) { p.PoolID = strings.Repeat("p", 61) }},
		{"no endpoints", privHex, func(p *poolRevisionPayload) { p.Endpoints = nil }},
		{"too many endpoints", privHex, func(p *poolRevisionPayload) {
			p.Endpoints = make([]poolEndpointInput, maxEndpointsPerManifest+1)
		}},
		{"bad port", privHex, func(p *poolRevisionPayload) { p.Endpoints[0].Port = 0 }},
		{"port too high", privHex, func(p *poolRevisionPayload) { p.Endpoints[0].Port = 65536 }},
		{"empty host", privHex, func(p *poolRevisionPayload) { p.Endpoints[0].Host = "" }},
		{"long sni", privHex, func(p *poolRevisionPayload) { p.Endpoints[0].SNI = strings.Repeat("s", 256) }},
		{"validity too short", privHex, func(p *poolRevisionPayload) { p.ValidityMinutes = 10 }},
		{"validity too long", privHex, func(p *poolRevisionPayload) { p.ValidityMinutes = 100000 }},
		{"signing key garbage", "not-hex", func(*poolRevisionPayload) {}},
		{"signing key short", strings.Repeat("ab", 31), func(*poolRevisionPayload) {}},
	}
	for _, tc := range cases {
		payload := validPoolRevisionPayload()
		tc.mutate(&payload)
		if _, err := buildManifest("node-1", payload, tc.key, now); err == nil {
			t.Fatalf("%s: must be rejected", tc.name)
		}
	}

	// Happy paths: default validity and explicit validity.
	if _, err := buildManifest("node-1", validPoolRevisionPayload(), privHex, now); err != nil {
		t.Fatalf("valid payload rejected: %v", err)
	}
	payload := validPoolRevisionPayload()
	payload.ValidityMinutes = 120
	m, err := buildManifest("node-1", payload, privHex, now)
	if err != nil {
		t.Fatalf("explicit validity rejected: %v", err)
	}
	if got := m.ValidUntil.Sub(m.ValidFrom); got != 2*time.Hour {
		t.Fatalf("explicit validity = %s, want 2h", got)
	}
}
