package runtimecred

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"testing"
)

func TestEncryptDecryptSecretRoundTrip(t *testing.T) {
	key := bytes.Repeat([]byte{0x42}, 32)
	plaintext := []byte("runtime-secret-value")

	ciphertext, nonce, err := EncryptSecret(key, plaintext)
	if err != nil {
		t.Fatalf("EncryptSecret() error = %v", err)
	}
	if len(nonce) != 12 {
		t.Fatalf("nonce length = %d, want 12", len(nonce))
	}
	if bytes.Equal(ciphertext, plaintext) {
		t.Fatal("ciphertext must not equal plaintext")
	}
	got, err := DecryptSecret(key, nonce, ciphertext)
	if err != nil {
		t.Fatalf("DecryptSecret() error = %v", err)
	}
	if !bytes.Equal(got, plaintext) {
		t.Fatalf("roundtrip plaintext = %q, want %q", got, plaintext)
	}
}

func TestEncryptSecretUsesUniqueNonce(t *testing.T) {
	key := bytes.Repeat([]byte{0x11}, 32)
	plaintext := []byte("same-secret")

	_, nonce1, err := EncryptSecret(key, plaintext)
	if err != nil {
		t.Fatalf("first EncryptSecret() error = %v", err)
	}
	_, nonce2, err := EncryptSecret(key, plaintext)
	if err != nil {
		t.Fatalf("second EncryptSecret() error = %v", err)
	}
	if bytes.Equal(nonce1, nonce2) {
		t.Fatal("two encryptions reused a GCM nonce")
	}
}

func TestDecryptSecretRejectsWrongKeyAndTamper(t *testing.T) {
	key := bytes.Repeat([]byte{0x21}, 32)
	wrongKey := bytes.Repeat([]byte{0x22}, 32)
	ciphertext, nonce, err := EncryptSecret(key, []byte("secret"))
	if err != nil {
		t.Fatalf("EncryptSecret() error = %v", err)
	}

	if _, err := DecryptSecret(wrongKey, nonce, ciphertext); err == nil {
		t.Fatal("DecryptSecret() accepted the wrong key")
	}

	tampered := append([]byte(nil), ciphertext...)
	tampered[len(tampered)-1] ^= 0x01
	if _, err := DecryptSecret(key, nonce, tampered); err == nil {
		t.Fatal("DecryptSecret() accepted tampered ciphertext")
	}
}

func TestSecretEnvelopeRequires32ByteKey(t *testing.T) {
	for _, keyLen := range []int{0, 16, 31, 33, 64} {
		key := make([]byte, keyLen)
		if _, _, err := EncryptSecret(key, []byte("secret")); err == nil {
			t.Fatalf("EncryptSecret() accepted key length %d", keyLen)
		}
		if _, err := DecryptSecret(key, make([]byte, 12), make([]byte, 16)); err == nil {
			t.Fatalf("DecryptSecret() accepted key length %d", keyLen)
		}
	}
}

func TestHashSecretIsSHA256(t *testing.T) {
	secret := []byte("hash-me")
	want := sha256.Sum256(secret)
	if got := HashSecret(secret); got != want {
		t.Fatalf("HashSecret() = %x, want %x", got, want)
	}
}

func TestGeneratePasswordUses24RandomBytesBase64URL(t *testing.T) {
	first, err := GeneratePassword()
	if err != nil {
		t.Fatalf("GeneratePassword() error = %v", err)
	}
	second, err := GeneratePassword()
	if err != nil {
		t.Fatalf("GeneratePassword() second error = %v", err)
	}
	if first == second {
		t.Fatal("two generated passwords are identical")
	}
	if bytes.ContainsRune([]byte(first), '=') {
		t.Fatalf("generated password contains padding: %q", first)
	}
	decoded, err := base64.RawURLEncoding.DecodeString(first)
	if err != nil {
		t.Fatalf("generated password is not raw base64url: %v", err)
	}
	if len(decoded) != 24 {
		t.Fatalf("generated password decodes to %d bytes, want 24", len(decoded))
	}
	if err := ValidatePassword(first, false); err != nil {
		t.Fatalf("generated password fails new-password policy: %v", err)
	}
}
