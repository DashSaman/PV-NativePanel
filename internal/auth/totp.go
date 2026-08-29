package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base32"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	totpSecretBytes = 20
	totpPeriod      = int64(30)
	totpDigits      = 6
)

var totpEncoding = base32.StdEncoding.WithPadding(base32.NoPadding)

func GenerateTOTPSecret() (string, error) {
	secret := make([]byte, totpSecretBytes)
	if _, err := rand.Read(secret); err != nil {
		return "", fmt.Errorf("auth: generate TOTP secret: %w", err)
	}
	return totpEncoding.EncodeToString(secret), nil
}

func TOTPCode(secret string, at time.Time) (string, int64, error) {
	key, err := decodeTOTPSecret(secret)
	if err != nil {
		return "", 0, err
	}
	step := at.Unix() / totpPeriod
	if step < 0 {
		return "", 0, errors.New("auth: TOTP time before Unix epoch")
	}
	return hotpCode(key, uint64(step)), step, nil
}

func ValidateTOTP(secret, code string, now time.Time, lastUsedStep *int64) (int64, bool, error) {
	if len(code) != totpDigits {
		return 0, false, nil
	}
	if _, err := strconv.Atoi(code); err != nil {
		return 0, false, nil
	}
	key, err := decodeTOTPSecret(secret)
	if err != nil {
		return 0, false, err
	}
	current := now.Unix() / totpPeriod
	if current < 0 {
		return 0, false, errors.New("auth: TOTP time before Unix epoch")
	}

	for _, candidate := range []int64{current, current - 1, current + 1} {
		if candidate < 0 {
			continue
		}
		if lastUsedStep != nil && candidate <= *lastUsedStep {
			continue
		}
		expected := hotpCode(key, uint64(candidate))
		if hmac.Equal([]byte(expected), []byte(code)) {
			return candidate, true, nil
		}
	}
	return 0, false, nil
}

func decodeTOTPSecret(secret string) ([]byte, error) {
	secret = strings.ToUpper(strings.TrimSpace(secret))
	decoded, err := totpEncoding.DecodeString(secret)
	if err != nil || len(decoded) < 16 {
		return nil, errors.New("auth: invalid TOTP secret")
	}
	return decoded, nil
}

func hotpCode(key []byte, counter uint64) string {
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], counter)
	mac := hmac.New(sha1.New, key)
	_, _ = mac.Write(buf[:])
	sum := mac.Sum(nil)
	offset := sum[len(sum)-1] & 0x0f
	binaryCode := (uint32(sum[offset])&0x7f)<<24 |
		uint32(sum[offset+1])<<16 |
		uint32(sum[offset+2])<<8 |
		uint32(sum[offset+3])
	value := binaryCode % 1_000_000
	return fmt.Sprintf("%06d", value)
}

func GenerateRecoveryCodes(count int) ([]string, [][32]byte, error) {
	if count < 1 || count > 100 {
		return nil, nil, errors.New("auth: recovery-code count must be between 1 and 100")
	}
	codes := make([]string, 0, count)
	hashes := make([][32]byte, 0, count)
	seen := make(map[string]struct{}, count)
	for len(codes) < count {
		buf := make([]byte, 16)
		if _, err := rand.Read(buf); err != nil {
			return nil, nil, fmt.Errorf("auth: generate recovery code: %w", err)
		}
		code := base64.RawURLEncoding.EncodeToString(buf)
		if _, exists := seen[code]; exists {
			continue
		}
		seen[code] = struct{}{}
		codes = append(codes, code)
		hashes = append(hashes, HashRecoveryCode(code))
	}
	return codes, hashes, nil
}

func HashRecoveryCode(code string) [32]byte {
	return sha256.Sum256([]byte(code))
}
