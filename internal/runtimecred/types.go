package runtimecred

import "time"

type CredentialStatus string

const (
	CredentialActive   CredentialStatus = "active"
	CredentialDisabled CredentialStatus = "disabled"
	CredentialRevoked  CredentialStatus = "revoked"
)

type CredentialOrigin string

const (
	CredentialImported CredentialOrigin = "imported"
	CredentialPanel    CredentialOrigin = "panel"
)

// Credential is an internal persistence-domain object. Secret material is
// deliberately unexported so accidental json.Marshal on this type cannot
// expose ciphertext, nonce or hash to browser-facing DTOs.
type Credential struct {
	ID              string
	Username        string
	EncryptionKeyID string
	Status          CredentialStatus
	Origin          CredentialOrigin
	Revision        int64
	CreatedAt       time.Time
	UpdatedAt       time.Time
	RotatedAt       *time.Time
	RevokedAt       *time.Time

	secretHash       [32]byte
	secretCiphertext []byte
	secretNonce      []byte
}

// DesiredCredential is internal desired runtime state. Password remains
// unexported; callers construct it through NewDesiredCredential and the
// Caddy renderer may retrieve it through Password only inside trusted code.
type DesiredCredential struct {
	ID       string
	Username string
	Status   CredentialStatus
	password string
}

func NewDesiredCredential(id, username, password string, status CredentialStatus) (DesiredCredential, error) {
	if err := ValidateUsername(username); err != nil {
		return DesiredCredential{}, err
	}
	if err := ValidatePassword(password, false); err != nil {
		return DesiredCredential{}, err
	}
	if status != CredentialActive && status != CredentialDisabled && status != CredentialRevoked {
		return DesiredCredential{}, ErrInvalidCredentialStatus
	}
	return DesiredCredential{ID: id, Username: username, Status: status, password: password}, nil
}

// NewImportedDesiredCredential exists only for importing the already-working
// live credential without silently rotating a legacy short password.
func NewImportedDesiredCredential(id, username, password string, status CredentialStatus) (DesiredCredential, error) {
	if err := ValidateUsername(username); err != nil {
		return DesiredCredential{}, err
	}
	if err := ValidatePassword(password, true); err != nil {
		return DesiredCredential{}, err
	}
	if status != CredentialActive && status != CredentialDisabled && status != CredentialRevoked {
		return DesiredCredential{}, ErrInvalidCredentialStatus
	}
	return DesiredCredential{ID: id, Username: username, Status: status, password: password}, nil
}

// Password is intentionally not represented as a serializable field.
func (c DesiredCredential) Password() string {
	return c.password
}

type DesiredState struct {
	Revision    string
	Credentials []DesiredCredential
}

var ErrInvalidCredentialStatus = invalidCredentialStatusError{}

type invalidCredentialStatusError struct{}

func (invalidCredentialStatusError) Error() string { return "runtimecred: invalid credential status" }
