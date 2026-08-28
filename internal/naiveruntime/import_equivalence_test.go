package naiveruntime

import (
	"bytes"
	"testing"

	"github.com/DashSaman/PV-NaivePanel/internal/runtimecred"
)

func TestRenderCredentialsPreservesEquivalentLiveUnquotedCredentialBytes(t *testing.T) {
	input := []byte(`example.com {
    forward_proxy {
        basic_auth live.user safe-password-123
        hide_ip
        hide_via
    }
}
`)
	credential := mustImportedCredential(t, "live.user", "safe-password-123", runtimecred.CredentialActive)

	output, err := RenderCredentials(input, []runtimecred.DesiredCredential{credential})
	if err != nil {
		t.Fatalf("RenderCredentials() error = %v", err)
	}
	if !bytes.Equal(output, input) {
		t.Fatalf("equivalent imported live credential changed Caddyfile bytes\n--- got ---\n%s\n--- want ---\n%s", output, input)
	}
}
