package fleetpull

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/DashSaman/PV-NaivePanel/internal/fleet"
)

func selfSignedCert(t *testing.T, commonName string, uris ...string) *x509.Certificate {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: commonName},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
	}
	for _, raw := range uris {
		parsed, err := url.Parse(raw)
		if err != nil {
			t.Fatalf("parse uri: %v", err)
		}
		template.URIs = append(template.URIs, parsed)
	}
	der, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("create certificate: %v", err)
	}
	parsed, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatalf("parse certificate: %v", err)
	}
	return parsed
}

func TestIdentityFromTLS(t *testing.T) {
	uriCert := selfSignedCert(t, "ignored-cn", "urn:pvnaive:node:0d3c6a1e-1111-4222-8333-444455556666")
	if got := IdentityFromTLS([]*x509.Certificate{uriCert}); got != "0d3c6a1e-1111-4222-8333-444455556666" {
		t.Fatalf("URI SAN identity = %q", got)
	}
	cnCert := selfSignedCert(t, "9c1b2a3e-9999-4222-8333-444455556666")
	if got := IdentityFromTLS([]*x509.Certificate{cnCert}); got != "9c1b2a3e-9999-4222-8333-444455556666" {
		t.Fatalf("CN fallback identity = %q", got)
	}
	if got := IdentityFromTLS(nil); got != "" {
		t.Fatalf("no cert identity = %q, want empty", got)
	}
	blank := selfSignedCert(t, "   ")
	if got := IdentityFromTLS([]*x509.Certificate{blank}); got != "" {
		t.Fatalf("blank CN identity = %q, want empty", got)
	}
}

func TestParsePullRequest(t *testing.T) {
	now := time.Now()
	if _, err := parsePullRequest(map[string]string{}, now); err != nil {
		t.Fatalf("empty request: %v", err)
	}
	req, err := parsePullRequest(map[string]string{
		"health":           "healthy",
		"applied_revision": "7",
		"agent_version":    "pvnaive-agent/1.0",
		"now":              now.UTC().Format(time.RFC3339Nano),
	}, now)
	if err != nil {
		t.Fatalf("valid request: %v", err)
	}
	if req.health != "healthy" || req.appliedRevision != 7 || !req.hasAgentNow {
		t.Fatalf("parsed request = %+v", req)
	}
	bad := []map[string]string{
		{"health": "blazing"},
		{"applied_revision": "-1"},
		{"applied_revision": "not-a-number"},
		{"now": "yesterday"},
		{"agent_version": string(make([]byte, 65))},
	}
	for index, params := range bad {
		if _, err := parsePullRequest(params, now); err == nil {
			t.Fatalf("case %d accepted malformed params %v", index, params)
		}
	}
}

func TestHandlerFailClosed(t *testing.T) {
	handler := &Handler{Store: nil, Now: func() time.Time { return time.Now() }}

	// Nil store with a valid cert identity -> 503 (config error surfaces
	// only to authenticated callers), no panic.
	request := httptest.NewRequest(http.MethodGet, "/fleet/v1/manifest", nil)
	request.TLS = &tls.ConnectionState{PeerCertificates: []*x509.Certificate{selfSignedCert(t, "0d3c6a1e-1111-4222-8333-444455556666")}}
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("nil store status = %d, want 503", recorder.Code)
	}

	// Plain HTTP (no TLS) -> 401 identity required.
	handler = &Handler{Now: func() time.Time { return time.Now() }}
	recorder = httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/fleet/v1/manifest", nil))
	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("no TLS status = %d, want 401", recorder.Code)
	}

	// Wrong method -> 405.
	recorder = httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/fleet/v1/manifest", nil))
	if recorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf("POST status = %d, want 405", recorder.Code)
	}

	// Unknown route -> 404.
	recorder = httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/fleet/v1/other", nil))
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("unknown route status = %d, want 404", recorder.Code)
	}
}

func TestHandlerTLSIdentityGate(t *testing.T) {
	handler := &Handler{Now: func() time.Time { return time.Now() }}
	// TLS state present but no peer certificate -> 401.
	request := httptest.NewRequest(http.MethodGet, "/fleet/v1/manifest", nil)
	request.TLS = &tls.ConnectionState{}
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("empty peers status = %d, want 401", recorder.Code)
	}
}

func TestHandlerServeManifest(t *testing.T) {
	// The handler is exercised with a real store-backed flow in the db
	// gate (tests/db); here the envelope path is covered with the real
	// verifier against a signed manifest published through PrepareRevision.
	pubHex, privHex, err := fleet.GenerateSigningKey()
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	nodeID := "7f6c5d4e-3333-4222-8333-444455556666"
	manifest := fleet.Manifest{
		Schema:     fleet.ManifestSchema,
		NodeID:     nodeID,
		PoolID:     "eu-primary",
		Endpoints:  []fleet.ManifestEndpoint{{Host: "edge.example.net", Port: 443, SNI: "edge.example.net"}},
		PubKeyHex:  pubHex,
		Weight:     2,
		ValidFrom:  time.Now().Add(-time.Minute),
		ValidUntil: time.Now().Add(time.Hour),
	}
	envelope, err := fleet.PrepareRevision(manifest, privHex, time.Now())
	if err != nil {
		t.Fatalf("prepare revision: %v", err)
	}
	if err := fleet.VerifyEnvelope(envelope, time.Now()); err != nil {
		t.Fatalf("verify envelope: %v", err)
	}
	encoded, err := json.Marshal(envelope)
	if err != nil {
		t.Fatalf("marshal envelope: %v", err)
	}
	if len(encoded) == 0 || envelope.Signature == "" {
		t.Fatalf("envelope missing signature")
	}
}
