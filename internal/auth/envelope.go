package auth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
)

const mfaEncryptionKeyBytes = 32

func EncryptSecret(key, plaintext []byte) ([]byte, []byte, error) {
	if len(key) != mfaEncryptionKeyBytes {
		return nil, nil, errors.New("auth: MFA encryption key must be 32 bytes")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, fmt.Errorf("auth: create AES cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, fmt.Errorf("auth: create GCM: %w", err)
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, nil, fmt.Errorf("auth: generate GCM nonce: %w", err)
	}
	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)
	return ciphertext, nonce, nil
}

func DecryptSecret(key, nonce, ciphertext []byte) ([]byte, error) {
	if len(key) != mfaEncryptionKeyBytes {
		return nil, errors.New("auth: MFA encryption key must be 32 bytes")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("auth: create AES cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("auth: create GCM: %w", err)
	}
	if len(nonce) != gcm.NonceSize() {
		return nil, errors.New("auth: invalid GCM nonce size")
	}
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, errors.New("auth: decrypt MFA secret")
	}
	return plaintext, nil
}
