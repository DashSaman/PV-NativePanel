package auth

import (
	"crypto/sha256"
	"encoding/base64"
	"testing"
)

func TestNewOpaqueToken(t *testing.T) {
	raw1, hash1, err := NewOpaqueToken()
	if err != nil {
		t.Fatalf("NewOpaqueToken: %v", err)
	}
	raw2, hash2, err := NewOpaqueToken()
	if err != nil {
		t.Fatalf("NewOpaqueToken second: %v", err)
	}
	if raw1 == raw2 || hash1 == hash2 {
		t.Fatal("opaque tokens must be unique")
	}
	decoded, err := base64.RawURLEncoding.DecodeString(raw1)
	if err != nil {
		t.Fatalf("token is not base64url: %v", err)
	}
	if len(decoded) != 32 {
		t.Fatalf("decoded token length=%d want=32", len(decoded))
	}
	if hash1 != sha256.Sum256([]byte(raw1)) {
		t.Fatal("stored token hash must be SHA-256 of raw cookie token")
	}
}

func TestHashOpaqueTokenDeterministic(t *testing.T) {
	got := HashOpaqueToken("abc")
	want := sha256.Sum256([]byte("abc"))
	if got != want {
		t.Fatalf("hash mismatch: got=%x want=%x", got, want)
	}
}
