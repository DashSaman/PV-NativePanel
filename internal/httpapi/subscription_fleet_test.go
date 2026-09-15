package httpapi

import (
	"testing"

	"github.com/DashSaman/PV-NaivePanel/internal/fleet"
	"github.com/DashSaman/PV-NaivePanel/internal/subscription"
)

func poolMappingFixture() ([]fleet.PoolNode, subscription.Node, map[string]fleet.ManifestEnvelope) {
	primary := subscription.Node{
		Name: "PVNaive-demo", Host: "naive.example.ir", Port: 443,
		Username: "demo", Password: "secret", SNI: "naive.example.ir",
	}
	inventory := []fleet.PoolNode{
		{ID: "n-healthy", DisplayName: "Frankfurt", Health: "healthy", Maintenance: "active", AppliedRevision: 3, DesiredRevision: 3},
		{ID: "n-draining", DisplayName: "Draining", Health: "healthy", Maintenance: "draining", AppliedRevision: 3, DesiredRevision: 3},
		{ID: "n-offline", DisplayName: "Offline", Health: "offline", Maintenance: "active", AppliedRevision: 3, DesiredRevision: 3},
		{ID: "n-pending", DisplayName: "Pending", Health: "healthy", Maintenance: "active", AppliedRevision: 2, DesiredRevision: 3},
		{ID: "n-nomanyl", DisplayName: "NoManifest", Health: "healthy", Maintenance: "active", AppliedRevision: 3, DesiredRevision: 3},
		{ID: "n-primary", DisplayName: "SameAsPrimary", Health: "healthy", Maintenance: "active", AppliedRevision: 3, DesiredRevision: 3},
	}
	manifests := map[string]fleet.ManifestEnvelope{
		"n-healthy": {Manifest: fleet.Manifest{Endpoints: []fleet.ManifestEndpoint{
			{Host: "198.51.100.10", Port: 443, SNI: "frankfurt.example.ir"},
			{Host: "198.51.100.10", Port: 8443},
		}}},
		"n-draining": {Manifest: fleet.Manifest{Endpoints: []fleet.ManifestEndpoint{{Host: "198.51.100.20", Port: 443}}}},
		"n-offline":  {Manifest: fleet.Manifest{Endpoints: []fleet.ManifestEndpoint{{Host: "198.51.100.30", Port: 443}}}},
		"n-pending":  {Manifest: fleet.Manifest{Endpoints: []fleet.ManifestEndpoint{{Host: "198.51.100.40", Port: 443}}}},
		"n-nomanyl":  {},
		"n-primary":  {Manifest: fleet.Manifest{Endpoints: []fleet.ManifestEndpoint{{Host: "naive.example.ir", Port: 443}}}},
	}
	return inventory, primary, manifests
}

func TestFleetNodesToSubscriptionNodesFiltersServingNodesOnly(t *testing.T) {
	inventory, primary, manifests := poolMappingFixture()
	latest := func(nodeID string) (fleet.ManifestEnvelope, bool) {
		env, ok := manifests[nodeID]
		return env, ok
	}
	got := fleetNodesToSubscriptionNodes(inventory, primary, latest)
	if len(got) != 2 {
		t.Fatalf("extra nodes = %d (%v), want 2 (healthy in-sync only, 2 endpoints)", len(got), got)
	}
	first := got[0]
	if first.Name != "Frankfurt" || first.Host != "198.51.100.10" || first.Port != 443 {
		t.Fatalf("first node = %+v", first)
	}
	if first.SNI != "frankfurt.example.ir" {
		t.Fatalf("first SNI = %q", first.SNI)
	}
	if got[1].Port != 8443 || got[1].SNI != "198.51.100.10" {
		t.Fatalf("second node = %+v (SNI must default to host)", got[1])
	}
	// Credentials always mirror the primary so the same account works fleet-wide.
	for _, node := range got {
		if node.Username != primary.Username || node.Password != primary.Password {
			t.Fatalf("node %q credentials do not mirror primary", node.Name)
		}
	}
}

func TestFleetNodesToSubscriptionNodesSkipsPrimaryDuplicate(t *testing.T) {
	inventory, primary, manifests := poolMappingFixture()
	latest := func(nodeID string) (fleet.ManifestEnvelope, bool) {
		env, ok := manifests[nodeID]
		return env, ok
	}
	got := fleetNodesToSubscriptionNodes(inventory, primary, latest)
	for _, node := range got {
		if node.Host == primary.Host && node.Port == primary.Port {
			t.Fatalf("primary endpoint duplicated as extra node: %+v", node)
		}
	}
}

func TestFleetNodesToSubscriptionNodesCapsAndToleratesEmpty(t *testing.T) {
	_, primary, _ := poolMappingFixture()
	if got := fleetNodesToSubscriptionNodes(nil, primary, func(string) (fleet.ManifestEnvelope, bool) { return fleet.ManifestEnvelope{}, false }); len(got) != 0 {
		t.Fatalf("empty inventory must map to zero extra nodes, got %d", len(got))
	}
}
