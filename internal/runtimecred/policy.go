package runtimecred

import (
	"errors"
	"fmt"
)

const (
	maxUsernameBytes = 64
	minPasswordBytes = 14
	maxPasswordBytes = 128
)

func ValidateUsername(username string) error {
	if len(username) == 0 || len(username) > maxUsernameBytes {
		return fmt.Errorf("runtimecred: username length must be 1-%d bytes", maxUsernameBytes)
	}
	for i := 0; i < len(username); i++ {
		b := username[i]
		if isASCIIAlphaNumeric(b) {
			continue
		}
		switch b {
		case '.', '_', '@', '+', '-':
			continue
		default:
			return errors.New("runtimecred: username contains an unsafe character")
		}
	}
	return nil
}

func ValidatePassword(password string, imported bool) error {
	min := minPasswordBytes
	if imported {
		// The live credential may predate the new-generation policy. Import must
		// preserve a working printable value rather than silently rotate it.
		min = 1
	}
	if len(password) < min || len(password) > maxPasswordBytes {
		return fmt.Errorf("runtimecred: password length must be %d-%d bytes", min, maxPasswordBytes)
	}
	for i := 0; i < len(password); i++ {
		b := password[i]
		if b < 0x20 || b > 0x7e {
			return errors.New("runtimecred: password must contain visible ASCII only")
		}
	}
	return nil
}

func isASCIIAlphaNumeric(b byte) bool {
	return (b >= 'a' && b <= 'z') ||
		(b >= 'A' && b <= 'Z') ||
		(b >= '0' && b <= '9')
}
