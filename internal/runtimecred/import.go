package runtimecred

import (
	"context"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
)

var (
	ErrRuntimeAlreadyOwned = errors.New("runtimecred: live runtime credentials are already owned")
	ErrImportEquivalence   = errors.New("runtimecred: imported desired state is not byte-equivalent to live Caddyfile")
)

type runtimeImportAgent interface {
	InspectCurrent(context.Context) (AgentInspection, []AgentCredential, error)
	Validate(context.Context, AgentApplyRequest) (string, error)
}

func (s *Service) ImportCurrent(ctx context.Context, tx *sql.Tx, actorID, idempotencyKey string) ([]CredentialView, error) {
	if s == nil || s.repository == nil || s.agent == nil {
		return nil, errors.New("runtimecred: runtime service is not initialized")
	}
	if err := validateMutationIdentity(actorID, idempotencyKey); err != nil {
		return nil, err
	}
	existing, err := s.repository.ListTx(ctx, tx)
	if err != nil {
		return nil, err
	}
	if len(existing) != 0 {
		return nil, ErrRuntimeAlreadyOwned
	}
	if replay, err := s.repository.FindRevisionByIdempotencyTx(ctx, tx, idempotencyKey); err != nil {
		return nil, err
	} else if replay != nil {
		return nil, ErrIdempotentReplay
	}

	importer, ok := s.agent.(runtimeImportAgent)
	if !ok {
		return nil, errors.New("runtimecred: runtime agent does not support guarded import")
	}
	inspection, live, err := importer.InspectCurrent(ctx)
	if err != nil {
		return nil, fmt.Errorf("runtimecred: inspect current runtime for import: %w", err)
	}
	if !validRuntimeSHA(inspection.CaddySHA256) {
		zeroAgentPasswords(live)
		return nil, errors.New("runtimecred: import inspection returned invalid Caddy SHA")
	}
	if len(live) == 0 || len(live) > 256 {
		zeroAgentPasswords(live)
		return nil, errors.New("runtimecred: import requires 1-256 live credentials")
	}

	seen := make(map[string]struct{}, len(live))
	for i := range live {
		credential := &live[i]
		if credential.Status == "" {
			credential.Status = CredentialActive
		}
		if credential.Status != CredentialActive {
			zeroAgentPasswords(live)
			return nil, errors.New("runtimecred: inspected import credential is not active")
		}
		if err := ValidateUsername(credential.Username); err != nil {
			zeroAgentPasswords(live)
			return nil, err
		}
		if _, duplicate := seen[credential.Username]; duplicate {
			zeroAgentPasswords(live)
			return nil, ErrUsernameConflict
		}
		seen[credential.Username] = struct{}{}
		if err := ValidatePassword(credential.Password, true); err != nil {
			zeroAgentPasswords(live)
			return nil, err
		}

		plaintext := []byte(credential.Password)
		ciphertext, nonce, err := EncryptSecret(s.key, plaintext)
		hash := HashSecret(plaintext)
		zeroBytes(plaintext)
		if err != nil {
			zeroAgentPasswords(live)
			return nil, err
		}
		stored := Credential{
			Username:         credential.Username,
			EncryptionKeyID:  s.keyID,
			Status:           CredentialActive,
			Origin:           CredentialImported,
			CreatedByActorID: actorID,
			UpdatedByActorID: actorID,
			Revision:         1,
			secretHash:       hash,
			secretCiphertext: ciphertext,
			secretNonce:      nonce,
		}
		if _, err := s.repository.CreateTx(ctx, tx, stored); err != nil {
			zeroAgentPasswords(live)
			return nil, err
		}
	}
	zeroAgentPasswords(live)

	stored, err := s.repository.ListTx(ctx, tx)
	if err != nil {
		return nil, err
	}
	desired, canonical, _, err := s.buildDesired(stored)
	if err != nil {
		return nil, err
	}
	zeroBytes(canonical)
	candidateSHA, err := importer.Validate(ctx, AgentApplyRequest{
		ExpectedCaddySHA256: inspection.CaddySHA256,
		Revision:            "import-current",
		Credentials:         desired,
	})
	zeroAgentPasswords(desired)
	if err != nil {
		return nil, fmt.Errorf("runtimecred: validate imported runtime state: %w", err)
	}
	if !validRuntimeSHA(candidateSHA) || !strings.EqualFold(candidateSHA, inspection.CaddySHA256) {
		return nil, ErrImportEquivalence
	}

	views := make([]CredentialView, 0, len(stored))
	for _, credential := range stored {
		views = append(views, credentialView(credential))
	}
	return views, nil
}

func validRuntimeSHA(value string) bool {
	if len(value) != 64 {
		return false
	}
	decoded, err := hex.DecodeString(value)
	return err == nil && len(decoded) == 32
}
