package fleet

import (
	"strings"
	"testing"
	"time"
)

func publishableManifest() Manifest {
	return Manifest{
		Schema:    ManifestSchema,
		NodeID:    "11111111-2222-3333-4444-555555555555",
		PoolID:    "pool-eu-1",
		Endpoints: []ManifestEndpoint{{Host: "198.51.100.10", Port: 443}},
		PubKeyHex: "a1b2c3d4e5f6a7b8a1b2c3d4e5f6a7b8a1b2c3d4e5f6a7b8a1b2c3d4e5f6a7b8",
		Weight:    1,
		ValidFrom: time.Date(2026, 9, 14, 0, 0, 0, 0, time.UTC),
		// Long window: expiry is enforced at TRUST time (VerifyEnvelope),
		// publish-time validation only requires a well-formed window.
		ValidUntil: time.Now().Add(365 * 24 * time.Hour).UTC(),
	}
}

func TestPrepareRevisionSignsAndVerifies(t *testing.T) {
	pubHex, privHex, err := GenerateSigningKey()
	if err != nil {
		t.Fatalf("signing key: %v", err)
	}
	m := publishableManifest()
	m.PubKeyHex = pubHex
	now := time.Now().UTC()
	env, err := PrepareRevision(m, privHex, now)
	if err != nil {
		t.Fatalf("prepare revision: %v", err)
	}
	if err := VerifyEnvelope(env, now); err != nil {
		t.Fatalf("published envelope must verify: %v", err)
	}
	// Tampered body must fail trust verification.
	tampered := env
	tampered.Manifest.Weight = env.Manifest.Weight + 1
	if err := VerifyEnvelope(tampered, now); err == nil {
		t.Fatal("tampered manifest envelope must fail verification")
	}
}

func TestPrepareRevisionRejectsInvertedWindow(t *testing.T) {
	_, privHex, err := GenerateSigningKey()
	if err != nil {
		t.Fatalf("signing key: %v", err)
	}
	m := publishableManifest()
	m.ValidFrom, m.ValidUntil = m.ValidUntil, m.ValidFrom
	if _, err := PrepareRevision(m, privHex, time.Now().UTC()); err == nil {
		t.Fatal("inverted validity window must be refused at publish time")
	}
}

func TestPrepareRevisionRequiresValidKey(t *testing.T) {
	if _, err := PrepareRevision(publishableManifest(), "zz-not-hex", time.Now().UTC()); err == nil {
		t.Fatal("invalid private key must be refused")
	}
}

func TestValidationContract(t *testing.T) {
	cases := []struct {
		name string
		fn   func() error
		ok   bool
	}{
		{"empty name", func() error { return ValidateNodeName("   ") }, false},
		{"long name", func() error { return ValidateNodeName(strings.Repeat("n", 121)) }, false},
		{"ok name", func() error { return ValidateNodeName(" node-eu-1 ") }, true},
		{"long region", func() error { return ValidateRegion(strings.Repeat("r", 61)) }, false},
		{"empty region ok", func() error { return ValidateRegion("") }, true},
		{"zero weight", func() error { return ValidateWeight(0) }, false},
		{"ok weight", func() error { return ValidateWeight(10000) }, true},
		{"bad health", func() error { return ValidateHealth("shiny") }, false},
		{"ok health", func() error { return ValidateHealth(string(NodeDegraded)) }, true},
		{"bad maintenance", func() error { return ValidateMaintenance("yanked") }, false},
		{"ok maintenance", func() error { return ValidateMaintenance(string(string(MaintenanceDraining))) }, true},
		{"ttl too short", func() error { return ValidateEnrollmentTTL(time.Minute) }, false},
		{"ttl too long", func() error { return ValidateEnrollmentTTL(25 * time.Hour) }, false},
		{"ttl ok", func() error { return ValidateEnrollmentTTL(DefaultEnrollmentTTL) }, true},
	}
	for _, tc := range cases {
		if err := tc.fn(); (err == nil) != tc.ok {
			t.Fatalf("%s: expected ok=%v, got %v", tc.name, tc.ok, err)
		}
	}
}

func TestTokenHashRoundTrip(t *testing.T) {
	raw, hash, err := PrepareEnrollmentToken()
	if err != nil {
		t.Fatalf("prepare token: %v", err)
	}
	if len(raw) != 64 || !ValidTokenHash(hash) {
		t.Fatalf("token shape invalid: raw_len=%d hash=%q", len(raw), hash)
	}
	if HashEnrollmentToken(raw) != hash {
		t.Fatal("hash round trip mismatch")
	}
	if !ValidTokenHash(strings.Repeat("0a", 32)) {
		t.Fatal("valid hex hash rejected")
	}
	if ValidTokenHash(strings.Repeat("0A", 32)) || ValidTokenHash(strings.Repeat("0g", 32)) || ValidTokenHash("ab") {
		t.Fatal("invalid hash shapes accepted")
	}
}

func TestStoreUnavailableFailsClosed(t *testing.T) {
	var nilStore *Store
	if _, _, err := nilStore.IssueEnrollmentToken(t.Context(), "n", DefaultEnrollmentTTL); err == nil {
		t.Fatal("nil store must fail closed (issue token)")
	}
	if _, _, _, err := nilStore.Enroll(t.Context(), "raw", "n", "", 1, "actor"); err == nil {
		t.Fatal("nil store must fail closed (enroll)")
	}
	if _, _, err := nilStore.PublishRevision(t.Context(), "n", publishableManifest(), "k", time.Now()); err == nil {
		t.Fatal("nil store must fail closed (publish)")
	}
	if _, _, _, err := nilStore.LatestRevision(t.Context(), "n"); err == nil {
		t.Fatal("nil store must fail closed (latest)")
	}
	if _, err := nilStore.Heartbeat(t.Context(), "n", string(NodeHealthy), 1); err == nil {
		t.Fatal("nil store must fail closed (heartbeat)")
	}
	if _, err := nilStore.ListNodes(t.Context()); err == nil {
		t.Fatal("nil store must fail closed (list)")
	}
	if _, err := nilStore.SetMaintenance(t.Context(), "n", string(MaintenanceDraining)); err == nil {
		t.Fatal("nil store must fail closed (maintenance)")
	}
}

func TestStoreInputValidationBeforeBoundary(t *testing.T) {
	store := NewStore(nil) // db nil: validation must fire before the boundary check for bad input
	if err := ValidateNodeName(""); err == nil {
		t.Fatal("precondition")
	}
	if _, _, err := store.IssueEnrollmentToken(t.Context(), "   ", DefaultEnrollmentTTL); err == nil {
		t.Fatal("malformed name must be refused")
	}
	if _, _, err := store.IssueEnrollmentToken(t.Context(), "ok", time.Second); err == nil {
		t.Fatal("out-of-range TTL must be refused")
	}
	if _, _, _, err := store.Enroll(t.Context(), "raw", "ok", "eu", 0, "actor"); err == nil {
		t.Fatal("zero weight must be refused")
	}
	if _, _, _, err := store.Enroll(t.Context(), "raw", "ok", "eu", 1, ""); err == nil {
		t.Fatal("missing actor must be refused")
	}
	if _, err := store.Heartbeat(t.Context(), "n", string(NodeHealthy), -1); err == nil {
		t.Fatal("negative applied revision must be refused")
	}
}
