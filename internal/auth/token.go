package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

const opaqueTokenBytes = 32

func NewOpaqueToken() (string, [32]byte, error) {
	buf := make([]byte, opaqueTokenBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", [32]byte{}, fmt.Errorf("auth: generate opaque token: %w", err)
	}
	raw := base64.RawURLEncoding.EncodeToString(buf)
	return raw, HashOpaqueToken(raw), nil
}

func HashOpaqueToken(raw string) [32]byte {
	return sha256.Sum256([]byte(raw))
}
