package subscription

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/DashSaman/PV-NaivePanel/internal/runtimecred"
)

type fakeStore struct {
	wantHash [32]byte
	record   Record
	err      error
}

func (f *fakeStore) ResolveToken(_ context.Context, tokenHash [32]byte) (Record, error) {
	if f.err != nil {
		return Record{}, f.err
	}
	if tokenHash != f.wantHash {
		return Record{}, errors.New("unexpected token hash")
	}
	return f.record, nil
}

func TestGenerateTokenProducesCanonicalOpaqueTokenAndHash(t *testing.T) {
	raw, hash, err := GenerateToken()
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}
	decoded, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		t.Fatalf("token is not canonical base64url: %v", err)
	}
	if len(decoded) != 32 {
		t.Fatalf("decoded token bytes = %d, want 32", len(decoded))
	}
	if got := sha256.Sum256(decoded); got != hash {
		t.Fatalf("token hash mismatch: got %x want %x", got, hash)
	}
	if strings.ContainsAny(raw, "+/=") {
		t.Fatalf("token is not URL-safe: %q", raw)
	}
}

func TestServiceResolveRendersNaiveURIFromEncryptedRuntimeSecret(t *testing.T) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i + 1)
	}
	ciphertext, nonce, err := runtimecred.EncryptSecret(key, []byte("safe password !@#"))
	if err != nil {
		t.Fatal(err)
	}
	rawBytes := make([]byte, 32)
	for i := range rawBytes {
		rawBytes[i] = byte(0x80 + i)
	}
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
		t.Fatalf("NewService() error = %v", err)
	}
	text, err := service.Resolve(context.Background(), raw, "proxy.example.com:443")
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	want := "naive+https://customer.name:safe%20password%20%21%40%23@proxy.example.com:443"
	if text != want {
		t.Fatalf("Resolve() = %q, want %q", text, want)
	}
}

func TestServiceResolveRejectsRevokedOrExpiredService(t *testing.T) {
	key := make([]byte, 32)
	ciphertext, nonce, err := runtimecred.EncryptSecret(key, []byte("password 123456"))
	if err != nil {
		t.Fatal(err)
	}
	rawBytes := make([]byte, 32)
	raw := base64.RawURLEncoding.EncodeToString(rawBytes)
	hash := sha256.Sum256(rawBytes)
	past := time.Now().UTC().Add(-time.Minute)
	for _, tc := range []Record{
		{Username: "a", SecretCiphertext: ciphertext, SecretNonce: nonce, EncryptionKeyID: "runtime-v1", UserState: "revoked", TermState: "active"},
		{Username: "a", SecretCiphertext: ciphertext, SecretNonce: nonce, EncryptionKeyID: "runtime-v1", UserState: "active", TermState: "revoked"},
		{Username: "a", SecretCiphertext: ciphertext, SecretNonce: nonce, EncryptionKeyID: "runtime-v1", UserState: "active", TermState: "active", ExpiresAt: &past},
	} {
		service, err := NewService(&fakeStore{wantHash: hash, record: tc}, key, "runtime-v1")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := service.Resolve(context.Background(), raw, "proxy.example.com"); !errors.Is(err, ErrUnavailable) {
			t.Fatalf("Resolve(%#v) error = %v, want ErrUnavailable", tc, err)
		}
	}
}

func TestServiceResolveAllowsPendingFirstUseWithoutActivatingIt(t *testing.T) {
	key := make([]byte, 32)
	ciphertext, nonce, err := runtimecred.EncryptSecret(key, []byte("password 123456"))
	if err != nil {
		t.Fatal(err)
	}
	rawBytes := make([]byte, 32)
	raw := base64.RawURLEncoding.EncodeToString(rawBytes)
	hash := sha256.Sum256(rawBytes)
	store := &fakeStore{wantHash: hash, record: Record{
		Username: "pending-user", SecretCiphertext: ciphertext, SecretNonce: nonce,
		EncryptionKeyID: "runtime-v1", UserState: "active", TermState: "pending",
	}}
	service, err := NewService(store, key, "runtime-v1")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.Resolve(context.Background(), raw, "proxy.example.com"); err != nil {
		t.Fatalf("pending first-use subscription must render without starting term: %v", err)
	}
}
