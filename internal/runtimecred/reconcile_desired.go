package runtimecred

import "fmt"

// ReconcileDesiredCredentials decrypts the ACTIVE runtime credentials into
// the export-safe AgentCredential form used by the boot-time config
// reconciliation. Disabled and revoked credentials are deliberately
// excluded: they must never appear in the rendered proxy config. A
// credential set with zero active members is a legitimate state (fresh
// install before the first customer), so it yields an empty slice instead
// of the service-level ErrLastActiveCredential.
//
// The returned passwords are plaintext by design (the proxy config must
// contain them); callers own zeroing AgentCredential.Password after use.
func ReconcileDesiredCredentials(key []byte, credentials []Credential) ([]AgentCredential, error) {
	desired := make([]AgentCredential, 0, len(credentials))
	for _, credential := range credentials {
		if credential.Status != CredentialActive {
			continue
		}
		plaintext, err := DecryptSecret(key, credential.secretNonce, credential.secretCiphertext)
		if err != nil {
			zeroAgentPasswords(desired)
			return nil, fmt.Errorf("runtimecred: decrypt active credential: %w", err)
		}
		password := string(plaintext)
		zeroBytes(plaintext)
		desired = append(desired, AgentCredential{
			ID:       credential.ID,
			Username: credential.Username,
			Password: password,
			Status:   credential.Status,
		})
	}
	return desired, nil
}
