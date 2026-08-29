package naiveruntime

import (
	"testing"

	"github.com/DashSaman/PV-NaivePanel/internal/runtimecred"
)

func TestRenderCredentialsRejectsAmbiguousCaddyDoubleQuoteBackslashPatterns(t *testing.T) {
	input := mustReadFixture(t)
	tests := map[string]string{
		"trailing backslash":          "ends-with-\\",
		"backslash immediately quote": `odd\"quote`,
	}

	for name, password := range tests {
		t.Run(name, func(t *testing.T) {
			credential := mustImportedCredential(t, "edge.user", password, runtimecred.CredentialActive)
			if _, err := RenderCredentials(input, []runtimecred.DesiredCredential{credential}); err == nil {
				t.Fatalf("RenderCredentials() accepted password %q that cannot be represented conservatively with Caddy double quotes", password)
			}
		})
	}
}
