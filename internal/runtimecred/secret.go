package runtimecred

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
)

const (
	runtimeKeySize       = 32
	runtimeNonceSize     = 12
	generatedSecretBytes = 24
)

// EncryptSecret encrypts runtime credential material with AES-256-GCM.
// The caller is responsible for keeping the key outside the database and for
// clearing plaintext buffers when practical.
func EncryptSecret(key, plaintext []byte) (ciphertext, nonce []byte, err error) {
	if len(key) != runtimeKeySize {
		return nil, nil, fmt.Errorf("runtimecred: encryption key must be %d bytes", runtimeKeySize)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, fmt.Errorf("runtimecred: initialize AES: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, fmt.Errorf("runtimecred: initialize GCM: %w", err)
	}
	if gcm.NonceSize() != runtimeNonceSize {
		return nil, nil, errors.New("runtimecred: unexpected GCM nonce size")
	}

	nonce = make([]byte, runtimeNonceSize)
	if _, err := rand.Read(nonce); err != nil {
		return nil, nil, fmt.Errorf("runtimecred: generate nonce: %w", err)
	}
	ciphertext = gcm.Seal(nil, nonce, plaintext, nil)
	return ciphertext, nonce, nil
}

// DecryptSecret authenticates and decrypts AES-256-GCM runtime credential
// material. Authentication failures deliberately return no plaintext.
func DecryptSecret(key, nonce, ciphertext []byte) ([]byte, error) {
	if len(key) != runtimeKeySize {
		return nil, fmt.Errorf("runtimecred: encryption key must be %d bytes", runtimeKeySize)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("runtimecred: initialize AES: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("runtimecred: initialize GCM: %w", err)
	}
	if len(nonce) != gcm.NonceSize() {
		return nil, fmt.Errorf("runtimecred: nonce must be %d bytes", gcm.NonceSize())
	}
	if len(ciphertext) < gcm.Overhead() {
		return nil, errors.New("runtimecred: ciphertext is too short")
	}

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, errors.New("runtimecred: secret authentication failed")
	}
	return plaintext, nil
}

func HashSecret(secret []byte) [32]byte {
	return sha256.Sum256(secret)
}

// GeneratePassword returns 24 bytes of CSPRNG entropy encoded as unpadded
// base64url. That produces a 32-character renderer-safe password.
func GeneratePassword() (string, error) {
	raw := make([]byte, generatedSecretBytes)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("runtimecred: generate password: %w", err)
	}
	password := base64.RawURLEncoding.EncodeToString(raw)
	for i := range raw {
		raw[i] = 0
	}
	return password, nil
}
