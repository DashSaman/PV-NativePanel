package auth

import (
	"encoding/base32"
	"encoding/base64"
	"testing"
	"time"
)

func TestTOTPCodeMatchesRFC6238SHA1SixDigitProjection(t *testing.T) {
	// RFC 6238 SHA1 test secret "12345678901234567890".
	secret := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString([]byte("12345678901234567890"))
	code, step, err := TOTPCode(secret, time.Unix(59, 0).UTC())
	if err != nil {
		t.Fatalf("TOTPCode: %v", err)
	}
	if step != 1 {
		t.Fatalf("step=%d want=1", step)
	}
	if code != "287082" {
		t.Fatalf("code=%q want=287082", code)
	}
}

func TestValidateTOTPAcceptsWindowAndRejectsReplay(t *testing.T) {
	secret := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString([]byte("12345678901234567890"))
	now := time.Unix(90, 0).UTC() // step 3
	previousCode, previousStep, err := TOTPCode(secret, time.Unix(60, 0).UTC())
	if err != nil {
		t.Fatal(err)
	}
	step, ok, err := ValidateTOTP(secret, previousCode, now, nil)
	if err != nil || !ok || step != previousStep {
		t.Fatalf("window validation step=%d ok=%v err=%v want step=%d", step, ok, err, previousStep)
	}

	lastUsed := previousStep
	if step, ok, err := ValidateTOTP(secret, previousCode, now, &lastUsed); err != nil || ok || step != 0 {
		t.Fatalf("replay accepted step=%d ok=%v err=%v", step, ok, err)
	}
}

func TestGenerateTOTPSecret(t *testing.T) {
	secret, err := GenerateTOTPSecret()
	if err != nil {
		t.Fatalf("GenerateTOTPSecret: %v", err)
	}
	decoded, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(secret)
	if err != nil {
		t.Fatalf("secret is not base32: %v", err)
	}
	if len(decoded) != 20 {
		t.Fatalf("secret entropy bytes=%d want=20", len(decoded))
	}
}

func TestRecoveryCodesAreHighEntropyUniqueAndHashOnly(t *testing.T) {
	codes, hashes, err := GenerateRecoveryCodes(10)
	if err != nil {
		t.Fatalf("GenerateRecoveryCodes: %v", err)
	}
	if len(codes) != 10 || len(hashes) != 10 {
		t.Fatalf("counts codes=%d hashes=%d", len(codes), len(hashes))
	}
	seen := map[string]bool{}
	for i, code := range codes {
		if seen[code] {
			t.Fatalf("duplicate recovery code %q", code)
		}
		seen[code] = true
		decoded, err := base64.RawURLEncoding.DecodeString(code)
		if err != nil {
			t.Fatalf("code %d is not base64url: %v", i, err)
		}
		if len(decoded) != 16 {
			t.Fatalf("code %d entropy bytes=%d want=16", i, len(decoded))
		}
		if hashes[i] != HashRecoveryCode(code) {
			t.Fatalf("code %d hash mismatch", i)
		}
	}
}

func TestGenerateRecoveryCodesRejectsInvalidCount(t *testing.T) {
	for _, count := range []int{0, -1, 101} {
		if _, _, err := GenerateRecoveryCodes(count); err == nil {
			t.Fatalf("count=%d should fail", count)
		}
	}
}
