package fleet

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func sampleManifest(now time.Time) Manifest {
	return Manifest{
		Schema: ManifestSchema,
		NodeID: "node-42",
		PoolID: "pool-premium",
		Endpoints: []ManifestEndpoint{
			{Host: "203.0.113.10", Port: 443, SNI: "placeholder.example"},
			{Host: "198.51.100.20", Port: 443},
		},
		Weight:     1.5,
		ValidFrom:  now.Add(-time.Hour),
		ValidUntil: now.Add(23 * time.Hour),
	}
}

func TestManifestSignVerifyRoundtrip(t *testing.T) {
	now := time.Date(2026, 9, 14, 10, 0, 0, 0, time.UTC)
	pubHex, privHex, err := GenerateSigningKey()
	if err != nil {
		t.Fatal(err)
	}
	m := sampleManifest(now)
	m.PubKeyHex = pubHex

	sig, err := SignManifest(m, privHex)
	if err != nil {
		t.Fatal(err)
	}
	env := ManifestEnvelope{Manifest: m, Signature: sig}
	if err := VerifyManifestEnvelope(env, now); err != nil {
		t.Fatalf("valid envelope must verify: %v", err)
	}
}

func TestManifestTamperDetection(t *testing.T) {
	now := time.Date(2026, 9, 14, 10, 0, 0, 0, time.UTC)
	pubHex, privHex, _ := GenerateSigningKey()
	m := sampleManifest(now)
	m.PubKeyHex = pubHex
	sig, err := SignManifest(m, privHex)
	if err != nil {
		t.Fatal(err)
	}

	// Mutate the weight after signing.
	m2 := m
	m2.Weight = 99
	if err := VerifyManifestEnvelope(ManifestEnvelope{Manifest: m2, Signature: sig}, now); err == nil {
		t.Fatal("tampered manifest must fail verification")
	}

	// Mutate one endpoint.
	m3 := m
	m3.Endpoints = []ManifestEndpoint{{Host: "203.0.113.10", Port: 8443}}
	if err := VerifyManifestEnvelope(ManifestEnvelope{Manifest: m3, Signature: sig}, now); err == nil {
		t.Fatal("tampered endpoint must fail verification")
	}

	// Wrong key entirely.
	otherPub, _, _ := GenerateSigningKey()
	m4 := m
	m4.PubKeyHex = otherPub
	if err := VerifyManifestEnvelope(ManifestEnvelope{Manifest: m4, Signature: sig}, now); err == nil {
		t.Fatal("pubkey is part of the signed body; swapping it must fail")
	}

	// Garbage signature.
	if err := VerifyManifestEnvelope(ManifestEnvelope{Manifest: m, Signature: "00ff"}, now); err == nil {
		t.Fatal("short signature must fail closed")
	}
}

func TestManifestExpiryRespected(t *testing.T) {
	now := time.Date(2026, 9, 14, 10, 0, 0, 0, time.UTC)
	pubHex, privHex, _ := GenerateSigningKey()
	m := sampleManifest(now)
	m.PubKeyHex = pubHex
	sig, _ := SignManifest(m, privHex)

	later := now.Add(24 * time.Hour) // beyond valid_until
	env := ManifestEnvelope{Manifest: m, Signature: sig}
	if err := VerifyManifestEnvelope(env, later); err == nil {
		t.Fatal("expired manifest must be rejected at trust time")
	}
	if err := VerifyManifestEnvelope(env, now.Add(-2*time.Hour)); err == nil {
		t.Fatal("manifest before valid_from must be rejected")
	}
}

func TestManifestStructureValidation(t *testing.T) {
	now := time.Date(2026, 9, 14, 10, 0, 0, 0, time.UTC)
	m := sampleManifest(now)

	if err := m.Validate(now); err != nil {
		t.Fatalf("sample manifest must be valid: %v", err)
	}
	bad := m
	bad.Schema = "other.v9"
	if err := bad.Validate(now); err == nil {
		t.Fatal("schema mismatch must be rejected")
	}
	bad = m
	bad.Endpoints = nil
	if err := bad.Validate(now); err == nil {
		t.Fatal("no endpoints must be rejected")
	}
	bad = m
	bad.Endpoints = []ManifestEndpoint{{Host: "203.0.113.10", Port: 99999}}
	if err := bad.Validate(now); err == nil {
		t.Fatal("out-of-range port must be rejected")
	}
	bad = m
	bad.Weight = 0
	if err := bad.Validate(now); err == nil {
		t.Fatal("non-positive weight must be rejected")
	}
}

func TestManifestCanonicalEncodingDeterministic(t *testing.T) {
	now := time.Date(2026, 9, 14, 10, 0, 0, 0, time.UTC)
	pubHex, privHex, _ := GenerateSigningKey()
	m := sampleManifest(now)
	m.PubKeyHex = pubHex

	sig1, err := SignManifest(m, privHex)
	if err != nil {
		t.Fatal(err)
	}
	sig2, err := SignManifest(m, privHex)
	if err != nil {
		t.Fatal(err)
	}
	if sig1 != sig2 {
		t.Fatal("signing the same manifest must be deterministic")
	}

	// JSON roundtrip keeps the signature valid (struct order fixed).
	blob, err := json.Marshal(ManifestEnvelope{Manifest: m, Signature: sig1})
	if err != nil {
		t.Fatal(err)
	}
	var env ManifestEnvelope
	if err := json.Unmarshal(blob, &env); err != nil {
		t.Fatal(err)
	}
	if err := VerifyManifestEnvelope(env, now); err != nil {
		t.Fatalf("json roundtrip must preserve verifiability: %v", err)
	}
}

func TestManifestKeyFormatsRejected(t *testing.T) {
	now := time.Now().UTC()
	m := sampleManifest(now)
	if _, err := SignManifest(m, "not-hex"); err == nil {
		t.Fatal("non-hex private key must be rejected")
	}
	if _, err := SignManifest(m, strings.Repeat("ab", 10)); err == nil {
		t.Fatal("wrong-length private key must be rejected")
	}
	m.PubKeyHex = "zz"
	if err := VerifyManifestEnvelope(ManifestEnvelope{Manifest: m, Signature: strings.Repeat("aa", 64)}, now); err == nil {
		t.Fatal("invalid pubkey must fail closed")
	}
}
