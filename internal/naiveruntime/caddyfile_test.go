package naiveruntime

import (
	"bytes"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/DashSaman/PV-NaivePanel/internal/runtimecred"
)

func TestInspectCaddyfileFindsSingleForwardProxyAndCredentials(t *testing.T) {
	input := mustReadFixture(t)

	inspection, err := InspectCaddyfile(input)
	if err != nil {
		t.Fatalf("InspectCaddyfile() error = %v", err)
	}
	want := []string{"legacy-user", "second.user"}
	if !reflect.DeepEqual(inspection.Usernames, want) {
		t.Fatalf("inspection usernames = %#v, want %#v", inspection.Usernames, want)
	}
}

func TestInspectCaddyfileRejectsZeroMultipleAndCredentialAmbiguity(t *testing.T) {
	tests := map[string]string{
		"zero forward proxy": `example.com {
    file_server
}
`,
		"multiple forward proxy": `a.example {
    forward_proxy {
        basic_auth one password-one-123
    }
}
b.example {
    forward_proxy {
        basic_auth two password-two-123
    }
}
`,
		"no credentials": `example.com {
    forward_proxy {
        hide_ip
    }
}
`,
		"non-contiguous credentials": `example.com {
    forward_proxy {
        basic_auth one password-one-123
        hide_ip
        basic_auth two password-two-123
    }
}
`,
	}

	for name, input := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := InspectCaddyfile([]byte(input)); err == nil {
				t.Fatal("InspectCaddyfile() unexpectedly succeeded")
			}
		})
	}
}

func TestInspectCaddyfileIgnoresCommentsAndQuotedBraceText(t *testing.T) {
	input := []byte(`# forward_proxy { basic_auth fake fake }
example.com {
    header X-Debug "forward_proxy { not-a-block }"
    forward_proxy {
        basic_auth real-user "real password 123"
        probe_resistance "value with } and { and # characters"
        # } forward_proxy { this is still a comment
        hide_ip
    }
    file_server
}
`)

	inspection, err := InspectCaddyfile(input)
	if err != nil {
		t.Fatalf("InspectCaddyfile() error = %v", err)
	}
	if !reflect.DeepEqual(inspection.Usernames, []string{"real-user"}) {
		t.Fatalf("inspection usernames = %#v", inspection.Usernames)
	}
}

func TestInspectCaddyfilePreservesLiteralBackslashesLikeCaddyLexer(t *testing.T) {
	input := []byte(`example.com {
    forward_proxy {
        basic_auth real-user "two\\slashes"
        hide_ip
    }
}
`)

	inspection, err := InspectCaddyfile(input)
	if err != nil {
		t.Fatalf("InspectCaddyfile() error = %v", err)
	}
	if len(inspection.credentials) != 1 {
		t.Fatalf("credential count = %d, want 1", len(inspection.credentials))
	}
	if got, want := inspection.credentials[0].password, `two\\slashes`; got != want {
		t.Fatalf("parsed password = %q, want %q", got, want)
	}
}

func TestRenderCredentialsPreservesEveryNonCredentialByte(t *testing.T) {
	input := mustReadFixture(t)
	first := mustImportedCredential(t, "new.owner", `quote"brace{}\\colon: safe password`, runtimecred.CredentialActive)
	second := mustImportedCredential(t, "mobile+01@example", "another safe password 123", runtimecred.CredentialActive)
	desired := []runtimecred.DesiredCredential{first, second}

	output, err := RenderCredentials(input, desired)
	if err != nil {
		t.Fatalf("RenderCredentials() error = %v", err)
	}
	if bytes.Equal(output, input) {
		t.Fatal("RenderCredentials() did not change the credential span")
	}
	if got, want := stripBasicAuthLines(output), stripBasicAuthLines(input); !bytes.Equal(got, want) {
		t.Fatalf("non-credential Caddyfile bytes changed\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}

	inspection, err := InspectCaddyfile(output)
	if err != nil {
		t.Fatalf("InspectCaddyfile(rendered) error = %v", err)
	}
	wantUsers := []string{"new.owner", "mobile+01@example"}
	if !reflect.DeepEqual(inspection.Usernames, wantUsers) {
		t.Fatalf("rendered usernames = %#v, want %#v", inspection.Usernames, wantUsers)
	}
	if bytes.Contains(output, []byte("legacy-user")) || bytes.Contains(output, []byte("second.user")) {
		t.Fatal("rendered output retained an old credential")
	}

	again, err := RenderCredentials(input, desired)
	if err != nil {
		t.Fatalf("second RenderCredentials() error = %v", err)
	}
	if !bytes.Equal(output, again) {
		t.Fatal("RenderCredentials() is not deterministic for identical input/state")
	}
}

func TestRenderCredentialsUsesCaddyCompatibleQuotedEscapes(t *testing.T) {
	input := mustReadFixture(t)
	credential := mustImportedCredential(t, "escape.user", `slash\path and "quote"`, runtimecred.CredentialActive)

	output, err := RenderCredentials(input, []runtimecred.DesiredCredential{credential})
	if err != nil {
		t.Fatalf("RenderCredentials() error = %v", err)
	}
	want := []byte(`        basic_auth escape.user "slash\path and \"quote\""` + "\n")
	if !bytes.Contains(output, want) {
		t.Fatalf("rendered basic_auth bytes are not Caddy-compatible\n--- got ---\n%s\n--- want line ---\n%s", output, want)
	}
}

func TestRenderCredentialsRendersOnlyActiveCredentials(t *testing.T) {
	input := mustReadFixture(t)
	active := mustImportedCredential(t, "active.user", "active password 123", runtimecred.CredentialActive)
	disabled := mustImportedCredential(t, "disabled.user", "disabled password 123", runtimecred.CredentialDisabled)
	revoked := mustImportedCredential(t, "revoked.user", "revoked password 123", runtimecred.CredentialRevoked)

	output, err := RenderCredentials(input, []runtimecred.DesiredCredential{disabled, active, revoked})
	if err != nil {
		t.Fatalf("RenderCredentials() error = %v", err)
	}
	inspection, err := InspectCaddyfile(output)
	if err != nil {
		t.Fatalf("InspectCaddyfile(rendered) error = %v", err)
	}
	if !reflect.DeepEqual(inspection.Usernames, []string{"active.user"}) {
		t.Fatalf("rendered usernames = %#v, want only active.user", inspection.Usernames)
	}
}

func TestRenderCredentialsRejectsDuplicateUnsafeAndNoActiveCredential(t *testing.T) {
	input := mustReadFixture(t)
	one := mustImportedCredential(t, "duplicate.user", "first safe password", runtimecred.CredentialActive)
	two := mustImportedCredential(t, "duplicate.user", "second safe password", runtimecred.CredentialActive)
	if _, err := RenderCredentials(input, []runtimecred.DesiredCredential{one, two}); err == nil {
		t.Fatal("RenderCredentials() accepted duplicate usernames")
	}

	unsafe := runtimecred.DesiredCredential{Username: "attacker\nhide_ip", Status: runtimecred.CredentialActive}
	if _, err := RenderCredentials(input, []runtimecred.DesiredCredential{unsafe}); err == nil {
		t.Fatal("RenderCredentials() accepted an unsafe username")
	}

	if _, err := RenderCredentials(input, nil); err == nil {
		t.Fatal("RenderCredentials() accepted an empty credential set")
	}
	disabled := mustImportedCredential(t, "disabled.user", "disabled password 123", runtimecred.CredentialDisabled)
	if _, err := RenderCredentials(input, []runtimecred.DesiredCredential{disabled}); err == nil {
		t.Fatal("RenderCredentials() accepted a state with no active credential")
	}
}

func mustReadFixture(t *testing.T) []byte {
	t.Helper()
	input, err := os.ReadFile("testdata/live_like.Caddyfile")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	return input
}

func mustImportedCredential(t *testing.T, username, password string, status runtimecred.CredentialStatus) runtimecred.DesiredCredential {
	t.Helper()
	credential, err := runtimecred.NewImportedDesiredCredential("test-id-"+username, username, password, status)
	if err != nil {
		t.Fatalf("NewImportedDesiredCredential(%q) error = %v", username, err)
	}
	return credential
}

func stripBasicAuthLines(input []byte) []byte {
	lines := strings.SplitAfter(string(input), "\n")
	var out strings.Builder
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(strings.TrimSuffix(line, "\n")), "basic_auth ") {
			continue
		}
		out.WriteString(line)
	}
	return []byte(out.String())
}
