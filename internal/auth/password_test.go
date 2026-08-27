package auth

import (
	"strings"
	"testing"
)

func TestHashPasswordAndVerify(t *testing.T) {
	hash, err := HashPassword("correct horse battery staple")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	if !strings.HasPrefix(hash, "$argon2id$v=19$m=19456,t=2,p=1$") {
		t.Fatalf("unexpected PHC parameters: %q", hash)
	}

	ok, err := VerifyPassword("correct horse battery staple", hash)
	if err != nil {
		t.Fatalf("VerifyPassword: %v", err)
	}
	if !ok {
		t.Fatal("expected password to verify")
	}

	ok, err = VerifyPassword("wrong password", hash)
	if err != nil {
		t.Fatalf("VerifyPassword wrong password: %v", err)
	}
	if ok {
		t.Fatal("wrong password verified")
	}
}

func TestVerifyPasswordRejectsMalformedOrUnsafePHC(t *testing.T) {
	cases := []string{
		"",
		"not-a-phc-string",
		"$argon2i$v=19$m=19456,t=2,p=1$c2FsdA$YWJjZA",
		"$argon2id$v=19$m=1024,t=1,p=1$c2FsdHNhbHQ$YWJjZA",
		"$argon2id$v=16$m=19456,t=2,p=1$c2FsdHNhbHQ$YWJjZA",
	}
	for _, encoded := range cases {
		if ok, err := VerifyPassword("password", encoded); err == nil || ok {
			t.Fatalf("expected malformed/unsafe PHC to be rejected: %q (ok=%v err=%v)", encoded, ok, err)
		}
	}
}

func TestHashPasswordRejectsEmptyPassword(t *testing.T) {
	if _, err := HashPassword(""); err == nil {
		t.Fatal("expected empty password to be rejected")
	}
}
