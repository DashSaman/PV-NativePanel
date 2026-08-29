package auth

import (
	"bytes"
	"crypto/rand"
	"testing"
)

func TestEncryptSecretRoundTrip(t *testing.T) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		t.Fatal(err)
	}
	plaintext := []byte("JBSWY3DPEHPK3PXP")
	ciphertext, nonce, err := EncryptSecret(key, plaintext)
	if err != nil {
		t.Fatalf("EncryptSecret: %v", err)
	}
	if len(nonce) != 12 {
		t.Fatalf("nonce bytes=%d want=12", len(nonce))
	}
	if bytes.Contains(ciphertext, plaintext) {
		t.Fatal("ciphertext contains plaintext")
	}
	got, err := DecryptSecret(key, nonce, ciphertext)
	if err != nil {
		t.Fatalf("DecryptSecret: %v", err)
	}
	if !bytes.Equal(got, plaintext) {
		t.Fatalf("plaintext mismatch: got=%q want=%q", got, plaintext)
	}
}

func TestDecryptSecretRejectsTampering(t *testing.T) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		t.Fatal(err)
	}
	ciphertext, nonce, err := EncryptSecret(key, []byte("secret"))
	if err != nil {
		t.Fatal(err)
	}
	ciphertext[len(ciphertext)-1] ^= 0xff
	if _, err := DecryptSecret(key, nonce, ciphertext); err == nil {
		t.Fatal("tampered ciphertext decrypted")
	}
}

func TestEnvelopeRejectsInvalidKeyOrNonce(t *testing.T) {
	if _, _, err := EncryptSecret(make([]byte, 31), []byte("secret")); err == nil {
		t.Fatal("31-byte key accepted")
	}
	key := make([]byte, 32)
	if _, err := DecryptSecret(key, make([]byte, 11), []byte("ciphertext")); err == nil {
		t.Fatal("11-byte nonce accepted")
	}
}
