package runtimecred

import (
	"strings"
	"testing"
)

func TestValidateUsernameAcceptsConservativeSafeSet(t *testing.T) {
	for _, username := range []string{
		"a",
		"user.name_01",
		"owner+device@example-name",
		strings.Repeat("a", 64),
	} {
		if err := ValidateUsername(username); err != nil {
			t.Fatalf("ValidateUsername(%q) error = %v", username, err)
		}
	}
}

func TestValidateUsernameRejectsInjectionAndOutOfRange(t *testing.T) {
	for _, username := range []string{
		"",
		strings.Repeat("a", 65),
		"user:name",
		"user name",
		"user\tname",
		"user\nname",
		"user\rname",
		"user\"name",
		"user'name",
		"user\\name",
		"user{name",
		"user}name",
		"نام",
	} {
		if err := ValidateUsername(username); err == nil {
			t.Fatalf("ValidateUsername(%q) unexpectedly succeeded", username)
		}
	}
}

func TestValidateNewPasswordPolicy(t *testing.T) {
	valid := []string{
		"12345678901234",
		"safe:P@ss word!",
		strings.Repeat("x", 128),
	}
	for _, password := range valid {
		if err := ValidatePassword(password, false); err != nil {
			t.Fatalf("ValidatePassword(new %q) error = %v", password, err)
		}
	}

	invalid := []string{
		"",
		"1234567890123",
		strings.Repeat("x", 129),
		"1234567890123\nX",
		"1234567890123\rX",
		"1234567890123\x00X",
		"1234567890123\x1fX",
		"1234567890123\x7fX",
		"1234567890123é",
	}
	for _, password := range invalid {
		if err := ValidatePassword(password, false); err == nil {
			t.Fatalf("ValidatePassword(new %q) unexpectedly succeeded", password)
		}
	}
}

func TestValidateImportedPasswordPreservesLegacyPrintableValue(t *testing.T) {
	for _, password := range []string{
		"x",
		"legacy: short password!",
		"quote\"brace{}\\colon:",
		strings.Repeat("z", 128),
	} {
		if err := ValidatePassword(password, true); err != nil {
			t.Fatalf("ValidatePassword(imported %q) error = %v", password, err)
		}
	}
}

func TestValidateImportedPasswordStillRejectsUnsafeControls(t *testing.T) {
	for _, password := range []string{
		"",
		strings.Repeat("x", 129),
		"legacy\nsecret",
		"legacy\rsecret",
		"legacy\x00secret",
		"legacy\x01secret",
		"legacy\x7fsecret",
	} {
		if err := ValidatePassword(password, true); err == nil {
			t.Fatalf("ValidatePassword(imported %q) unexpectedly succeeded", password)
		}
	}
}
