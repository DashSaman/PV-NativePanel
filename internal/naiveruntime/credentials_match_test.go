package naiveruntime

import (
	"strings"
	"testing"

	"github.com/DashSaman/PV-NaivePanel/internal/runtimecred"
)

func credentialsMatchSeed() []byte {
	return []byte(`{
	order forward_proxy before file_server
	default_sni panel.example.invalid
}

:443, panel.example.invalid {
	encode gzip

	forward_proxy {
		basic_auth live.user "live-password-123"
		hide_ip
		hide_via
		probe_resistance
		pvnaive_accounting_socket /run/pvnaive/accounting.sock
		pvnaive_node_id pvnaive-node-1
		pvnaive_runtime_credential live.user 11111111-1111-1111-1111-111111111111
	}

	root * /opt/pvnaive/camouflage
	file_server
}
`)
}

func TestCredentialsMatchTrueForEquivalentSet(t *testing.T) {
	desired := []runtimecred.DesiredCredential{mustDesiredCredential(t, "11111111-1111-1111-1111-111111111111", "live.user", "live-password-123")}
	match, err := CredentialsMatch(credentialsMatchSeed(), desired)
	if err != nil {
		t.Fatalf("CredentialsMatch: %v", err)
	}
	if !match {
		t.Fatal("CredentialsMatch = false for an identical live set, want true")
	}
}

func TestCredentialsMatchFalseForDifferentPassword(t *testing.T) {
	desired := []runtimecred.DesiredCredential{mustDesiredCredential(t, "11111111-1111-1111-1111-111111111111", "live.user", "rotated-password-456")}
	match, err := CredentialsMatch(credentialsMatchSeed(), desired)
	if err != nil {
		t.Fatalf("CredentialsMatch: %v", err)
	}
	if match {
		t.Fatal("CredentialsMatch = true for a rotated password, want false")
	}
}

func TestCredentialsMatchFalseForDifferentUserCount(t *testing.T) {
	desired := []runtimecred.DesiredCredential{
		mustDesiredCredential(t, "11111111-1111-1111-1111-111111111111", "live.user", "live-password-123"),
		mustDesiredCredential(t, "22222222-2222-2222-2222-222222222222", "second.user", "second-password-123"),
	}
	match, err := CredentialsMatch(credentialsMatchSeed(), desired)
	if err != nil {
		t.Fatalf("CredentialsMatch: %v", err)
	}
	if match {
		t.Fatal("CredentialsMatch = true when the database has an extra active credential, want false")
	}
}

func TestCredentialsMatchErrorOnMissingCredentialBlock(t *testing.T) {
	configWithoutCredentials := strings.Replace(
		strings.Replace(string(credentialsMatchSeed()), "basic_auth live.user \"live-password-123\"\n", "", 1),
		"pvnaive_runtime_credential live.user 11111111-1111-1111-1111-111111111111\n", "", 1)
	_, err := CredentialsMatch([]byte(configWithoutCredentials), nil)
	if err == nil || !strings.Contains(err.Error(), "no basic_auth credentials") {
		t.Fatalf("CredentialsMatch on empty block error = %v, want no-basic_auth error", err)
	}
}

func mustDesiredCredential(t *testing.T, id, username, password string) runtimecred.DesiredCredential {
	t.Helper()
	credential, err := runtimecred.NewImportedDesiredCredential(id, username, password, runtimecred.CredentialActive)
	if err != nil {
		t.Fatalf("NewImportedDesiredCredential(%q): %v", username, err)
	}
	return credential
}
