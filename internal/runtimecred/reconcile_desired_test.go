package runtimecred

import "testing"

const reconcileTestKey = "0123456789abcdef0123456789abcdef"

func reconcileTestCredential(t *testing.T, id, username, password string, status CredentialStatus) Credential {
	t.Helper()
	ciphertext, nonce, hash, err := encryptMaterial([]byte(reconcileTestKey), password)
	if err != nil {
		t.Fatalf("encryptMaterial(%q): %v", password, err)
	}
	return Credential{
		ID:               id,
		Username:         username,
		EncryptionKeyID:  "runtime-v1",
		Status:           status,
		Origin:           CredentialPanel,
		Revision:         1,
		secretHash:       hash,
		secretCiphertext: ciphertext,
		secretNonce:      nonce,
	}
}

func TestReconcileDesiredCredentialsDecryptsOnlyActive(t *testing.T) {
	key := []byte(reconcileTestKey)
	credentials := []Credential{
		reconcileTestCredential(t, "11111111-1111-1111-1111-111111111111", "active.user", "active-secret-123", CredentialActive),
		reconcileTestCredential(t, "22222222-2222-2222-2222-222222222222", "disabled.user", "disabled-secret-123", CredentialDisabled),
		reconcileTestCredential(t, "33333333-3333-3333-3333-333333333333", "revoked.user", "revoked-secret-123", CredentialRevoked),
	}
	desired, err := ReconcileDesiredCredentials(key, credentials)
	if err != nil {
		t.Fatalf("ReconcileDesiredCredentials: %v", err)
	}
	if len(desired) != 1 {
		t.Fatalf("desired set size = %d, want 1", len(desired))
	}
	if desired[0].Username != "active.user" || desired[0].Password != "active-secret-123" {
		t.Fatalf("unexpected desired credential %+v", desired[0])
	}
	if desired[0].ID != "11111111-1111-1111-1111-111111111111" {
		t.Fatalf("desired credential ID = %q, want the persisted runtime UUID", desired[0].ID)
	}
}

func TestReconcileDesiredCredentialsEmptyOnZeroActive(t *testing.T) {
	key := []byte(reconcileTestKey)
	credentials := []Credential{
		reconcileTestCredential(t, "22222222-2222-2222-2222-222222222222", "disabled.user", "disabled-secret-123", CredentialDisabled),
		reconcileTestCredential(t, "33333333-3333-3333-3333-333333333333", "revoked.user", "revoked-secret-123", CredentialRevoked),
	}
	desired, err := ReconcileDesiredCredentials(key, credentials)
	if err != nil {
		t.Fatalf("ReconcileDesiredCredentials: %v", err)
	}
	if len(desired) != 0 {
		t.Fatalf("desired set size = %d, want 0 (fresh-install no-op state)", len(desired))
	}
}

func TestReconcileDesiredCredentialsRejectsCorruptEnvelope(t *testing.T) {
	key := []byte(reconcileTestKey)
	corrupted := reconcileTestCredential(t, "44444444-4444-4444-4444-444444444444", "corrupt.user", "corrupt-secret-123", CredentialActive)
	corrupted.secretCiphertext = []byte("too-short")
	_, err := ReconcileDesiredCredentials(key, []Credential{corrupted})
	if err == nil {
		t.Fatal("ReconcileDesiredCredentials accepted a corrupt secret envelope")
	}
}
