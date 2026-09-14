package fleet

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

// ManifestSchema tags the signed node manifest payload version (R4).
const ManifestSchema = "pvnaive.node.manifest.v1"

// ManifestEndpoint is one connectable endpoint of a pool node. Hosts are
// expected to be IPs or hostnames already hardcoded into rendered client
// configs — never a name requiring per-connect DNS resolution.
type ManifestEndpoint struct {
	Host string `json:"host"`
	Port int    `json:"port"`
	SNI  string `json:"sni,omitempty"`
}

// Manifest is the signed desired-state document for a pool node
// (docs/STEERING_SPEC_FA.md §4). It is distributed out-of-band and verified
// before trust; no central controller is required for verification.
type Manifest struct {
	Schema     string             `json:"schema"`
	NodeID     string             `json:"node_id"`
	PoolID     string             `json:"pool_id"`
	Endpoints  []ManifestEndpoint `json:"endpoints"`
	PubKeyHex  string             `json:"pubkey"`
	Weight     float64            `json:"weight"`
	ValidFrom  time.Time          `json:"valid_from"`
	ValidUntil time.Time          `json:"valid_until"`
}

// SignatureHex wraps the hex-encoded Ed25519 signature over the canonical
// manifest JSON (without the signature itself).
type ManifestEnvelope struct {
	Manifest  Manifest `json:"manifest"`
	Signature string   `json:"signature"` // hex ed25519 signature
}

// GenerateSigningKey produces a fresh Ed25519 keypair (hex-encoded). Private
// keys are operator secrets: they are distributed out-of-band and never
// stored in the repository or rendered configs.
func GenerateSigningKey() (pubHex string, privHex string, err error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return "", "", fmt.Errorf("fleet: generate signing key: %w", err)
	}
	return hex.EncodeToString(pub), hex.EncodeToString(priv), nil
}

// Validate checks the manifest's structural integrity. Expired manifests are
// valid documents but MustAccept rejects them at trust time.
func (m Manifest) Validate(now time.Time) error {
	if m.Schema != ManifestSchema {
		return fmt.Errorf("fleet: manifest schema %q, want %q", m.Schema, ManifestSchema)
	}
	if strings.TrimSpace(m.NodeID) == "" {
		return errors.New("fleet: manifest node id is required")
	}
	if strings.TrimSpace(m.PoolID) == "" {
		return errors.New("fleet: manifest pool id is required")
	}
	if len(m.Endpoints) == 0 {
		return errors.New("fleet: manifest requires at least one endpoint")
	}
	for _, e := range m.Endpoints {
		if strings.TrimSpace(e.Host) == "" || e.Port <= 0 || e.Port > 65535 {
			return errors.New("fleet: manifest endpoint host/port invalid")
		}
	}
	if m.Weight <= 0 {
		return errors.New("fleet: manifest weight must be positive")
	}
	if !m.ValidFrom.Before(m.ValidUntil) {
		return errors.New("fleet: manifest validity window is empty")
	}
	if now.After(m.ValidUntil) {
		return fmt.Errorf("fleet: manifest expired at %s (expiry must be respected)", m.ValidUntil.UTC().Format(time.RFC3339))
	}
	if now.Before(m.ValidFrom) {
		return fmt.Errorf("fleet: manifest not valid before %s", m.ValidFrom.UTC().Format(time.RFC3339))
	}
	return nil
}

// canonicalFingerprint builds the deterministic JSON body that gets signed.
// Map-free struct encoding keeps the byte layout stable across releases.
func canonicalFingerprint(m Manifest) ([]byte, error) {
	// The envelope's manifest minus nothing: the Manifest struct encodes
	// deterministically because all fields are fixed-order typed structs.
	body := struct {
		Schema     string             `json:"schema"`
		NodeID     string             `json:"node_id"`
		PoolID     string             `json:"pool_id"`
		Endpoints  []ManifestEndpoint `json:"endpoints"`
		PubKeyHex  string             `json:"pubkey"`
		Weight     float64            `json:"weight"`
		ValidFrom  time.Time          `json:"valid_from"`
		ValidUntil time.Time          `json:"valid_until"`
	}{
		Schema:     m.Schema,
		NodeID:     m.NodeID,
		PoolID:     m.PoolID,
		Endpoints:  m.Endpoints,
		PubKeyHex:  m.PubKeyHex,
		Weight:     m.Weight,
		ValidFrom:  m.ValidFrom.UTC(),
		ValidUntil: m.ValidUntil.UTC(),
	}
	return json.Marshal(body)
}

// SignManifest produces the hex Ed25519 signature over the canonical
// manifest JSON. privHex is the operator's out-of-band private key.
func SignManifest(m Manifest, privHex string) (string, error) {
	priv, err := hex.DecodeString(strings.TrimSpace(privHex))
	if err != nil || len(priv) != ed25519.PrivateKeySize {
		return "", errors.New("fleet: invalid private key (want 64-byte hex)")
	}
	body, err := canonicalFingerprint(m)
	if err != nil {
		return "", err
	}
	sig := ed25519.Sign(ed25519.PrivateKey(priv), body)
	return hex.EncodeToString(sig), nil
}

// VerifyManifestEnvelope verifies a signed manifest envelope against the
// embedded public key, then validates structure and the validity window.
// Verification fails closed: any error means the manifest is not trusted.
func VerifyManifestEnvelope(env ManifestEnvelope, now time.Time) error {
	pub, err := hex.DecodeString(strings.TrimSpace(env.Manifest.PubKeyHex))
	if err != nil || len(pub) != ed25519.PublicKeySize {
		return errors.New("fleet: manifest pubkey invalid (want 32-byte hex)")
	}
	sig, err := hex.DecodeString(strings.TrimSpace(env.Signature))
	if err != nil || len(sig) != ed25519.SignatureSize {
		return errors.New("fleet: manifest signature invalid")
	}
	body, err := canonicalFingerprint(env.Manifest)
	if err != nil {
		return err
	}
	if !ed25519.Verify(ed25519.PublicKey(pub), body, sig) {
		return errors.New("fleet: manifest signature verification failed")
	}
	return env.Manifest.Validate(now)
}
