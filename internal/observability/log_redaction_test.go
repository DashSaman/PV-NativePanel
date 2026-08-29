package observability

import (
	"strings"
	"testing"
)

func TestRedactTextRemovesCredentialsAndSubscriptionTokens(t *testing.T) {
	token := strings.Repeat("A", 43)
	input := "GET /sub/" + token + " direct=naive+https://alice:SuperSecret@example.com authorization: Bearer abcdef"
	got := RedactText(input)
	for _, secret := range []string{token, "SuperSecret", "abcdef"} {
		if strings.Contains(got, secret) {
			t.Fatalf("secret %q leaked in %q", secret, got)
		}
	}
	if !strings.Contains(got, "[REDACTED]") {
		t.Fatalf("redaction marker missing: %q", got)
	}
}
