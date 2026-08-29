package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	argonMemory      = uint32(19456)
	argonIterations  = uint32(2)
	argonParallelism = uint8(1)
	argonSaltLength  = 16
	argonKeyLength   = uint32(32)
)

var errInvalidPasswordHash = errors.New("auth: invalid password hash")

func HashPassword(password string) (string, error) {
	if password == "" {
		return "", errors.New("auth: password must not be empty")
	}

	salt := make([]byte, argonSaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("auth: generate password salt: %w", err)
	}
	key := argon2.IDKey([]byte(password), salt, argonIterations, argonMemory, argonParallelism, argonKeyLength)

	return fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		argonMemory,
		argonIterations,
		argonParallelism,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(key),
	), nil
}

func VerifyPassword(password, encoded string) (bool, error) {
	memory, iterations, parallelism, salt, expected, err := parsePasswordHash(encoded)
	if err != nil {
		return false, err
	}
	actual := argon2.IDKey([]byte(password), salt, iterations, memory, parallelism, uint32(len(expected)))
	return subtle.ConstantTimeCompare(actual, expected) == 1, nil
}

func parsePasswordHash(encoded string) (uint32, uint32, uint8, []byte, []byte, error) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[0] != "" || parts[1] != "argon2id" || parts[2] != "v=19" {
		return 0, 0, 0, nil, nil, errInvalidPasswordHash
	}

	params := strings.Split(parts[3], ",")
	if len(params) != 3 {
		return 0, 0, 0, nil, nil, errInvalidPasswordHash
	}
	memory64, err := parsePHCParam(params[0], "m")
	if err != nil {
		return 0, 0, 0, nil, nil, errInvalidPasswordHash
	}
	iterations64, err := parsePHCParam(params[1], "t")
	if err != nil {
		return 0, 0, 0, nil, nil, errInvalidPasswordHash
	}
	parallelism64, err := parsePHCParam(params[2], "p")
	if err != nil {
		return 0, 0, 0, nil, nil, errInvalidPasswordHash
	}
	if memory64 != uint64(argonMemory) || iterations64 != uint64(argonIterations) || parallelism64 != uint64(argonParallelism) {
		return 0, 0, 0, nil, nil, errInvalidPasswordHash
	}

	salt, err := base64.RawStdEncoding.Strict().DecodeString(parts[4])
	if err != nil || len(salt) != argonSaltLength {
		return 0, 0, 0, nil, nil, errInvalidPasswordHash
	}
	expected, err := base64.RawStdEncoding.Strict().DecodeString(parts[5])
	if err != nil || len(expected) != int(argonKeyLength) {
		return 0, 0, 0, nil, nil, errInvalidPasswordHash
	}

	return uint32(memory64), uint32(iterations64), uint8(parallelism64), salt, expected, nil
}

func parsePHCParam(raw, name string) (uint64, error) {
	prefix := name + "="
	if !strings.HasPrefix(raw, prefix) {
		return 0, errInvalidPasswordHash
	}
	value, err := strconv.ParseUint(strings.TrimPrefix(raw, prefix), 10, 32)
	if err != nil {
		return 0, errInvalidPasswordHash
	}
	return value, nil
}
