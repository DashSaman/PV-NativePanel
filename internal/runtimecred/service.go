package runtimecred

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

var (
	ErrRevisionConflict     = errors.New("runtimecred: revision conflict")
	ErrLastActiveCredential = errors.New("runtimecred: at least one active credential must remain")
	ErrConsistency          = errors.New("runtimecred: runtime/database consistency failure")
	ErrIdempotentReplay     = errors.New("runtimecred: idempotency key already used")
)

type CredentialView struct {
	ID        string           `json:"id"`
	Username  string           `json:"username"`
	Status    CredentialStatus `json:"status"`
	Origin    CredentialOrigin `json:"origin"`
	Revision  int64            `json:"revision"`
	CreatedAt time.Time        `json:"created_at,omitempty"`
	UpdatedAt time.Time        `json:"updated_at,omitempty"`
	RotatedAt *time.Time       `json:"rotated_at,omitempty"`
	RevokedAt *time.Time       `json:"revoked_at,omitempty"`
}

type AppliedRuntimeMetadata struct {
	PreviousSHA256 string `json:"previous_sha256"`
	AppliedSHA256  string `json:"applied_sha256"`
	BackupID       string `json:"backup_id"`
	MainPID        int    `json:"main_pid"`
	NRestarts      int    `json:"n_restarts"`
}

type RuntimeRevision struct {
	ID                   string
	RevisionNo           int64
	State                string
	IdempotencyKey       string
	ConfigChecksumSHA256 string
	ConfigCiphertext     []byte
	EncryptionKeyID      string
	Manifest             map[string]any
	CreatedByActorID     string
	FailureCode          string
	Applied              AppliedRuntimeMetadata
	CreatedAt            time.Time
	AppliedAt            *time.Time
}

type AgentInspection struct {
	CaddySHA256 string
}

type AgentCredential struct {
	ID       string
	Username string
	Password string
	Status   CredentialStatus
}

type AgentApplyRequest struct {
	ExpectedCaddySHA256 string
	Revision            string
	Credentials         []AgentCredential
}

type AgentApplyResult struct {
	PreviousSHA256 string
	AppliedSHA256  string
	BackupID       string
	MainPID        int
	NRestarts      int
}

type RuntimeAgent interface {
	Inspect(context.Context) (AgentInspection, error)
	Apply(context.Context, AgentApplyRequest) (AgentApplyResult, error)
	Rollback(context.Context, string) error
}

type RuntimeRepository interface {
	ListTx(context.Context, *sql.Tx) ([]Credential, error)
	CreateTx(context.Context, *sql.Tx, Credential) (Credential, error)
	UpdateTx(context.Context, *sql.Tx, string, int64, string, CredentialStatus, string) (Credential, error)
	RotateTx(context.Context, *sql.Tx, string, int64, [32]byte, []byte, []byte, string, string) (Credential, error)
	RevokeTx(context.Context, *sql.Tx, string, int64, string) (Credential, error)
	FindRevisionByIdempotencyTx(context.Context, *sql.Tx, string) (*RuntimeRevision, error)
	CreateRuntimeRevisionTx(context.Context, *sql.Tx, RuntimeRevision) (RuntimeRevision, error)
	MarkRevisionAppliedTx(context.Context, *sql.Tx, string, AppliedRuntimeMetadata) error
	MarkRevisionFailedTx(context.Context, *sql.Tx, string, string) error
}

type Service struct {
	repository RuntimeRepository
	agent      RuntimeAgent
	key        []byte
	keyID      string
}

type CreateInput struct {
	Username         string
	Password         string
	GeneratePassword bool
}

type UpdateInput struct {
	ID               string
	ExpectedRevision int64
	Username         string
	Status           CredentialStatus
}

type RotateInput struct {
	ID               string
	ExpectedRevision int64
	Password         string
	GeneratePassword bool
}

type RevokeInput struct {
	ID               string
	ExpectedRevision int64
}

type Mutation struct {
	mu                sync.Mutex
	agent             RuntimeAgent
	backupID          string
	credential        CredentialView
	runtimeRevisionID string
	generatedPassword string
	committed         bool
	secretTaken       bool
}

func NewService(repository RuntimeRepository, agent RuntimeAgent, key []byte, keyID string) (*Service, error) {
	if repository == nil || agent == nil {
		return nil, errors.New("runtimecred: repository and runtime agent are required")
	}
	if len(key) != runtimeKeySize {
		return nil, fmt.Errorf("runtimecred: runtime key must be %d bytes", runtimeKeySize)
	}
	if strings.TrimSpace(keyID) == "" || len(keyID) > 160 {
		return nil, errors.New("runtimecred: invalid runtime key id")
	}
	keyCopy := append([]byte(nil), key...)
	return &Service{repository: repository, agent: agent, key: keyCopy, keyID: keyID}, nil
}

func (s *Service) List(ctx context.Context, tx *sql.Tx) ([]CredentialView, error) {
	credentials, err := s.repository.ListTx(ctx, tx)
	if err != nil {
		return nil, err
	}
	out := make([]CredentialView, 0, len(credentials))
	for _, credential := range credentials {
		out = append(out, credentialView(credential))
	}
	return out, nil
}

func (s *Service) Create(ctx context.Context, tx *sql.Tx, actorID, idempotencyKey string, input CreateInput) (*Mutation, error) {
	if err := validateMutationIdentity(actorID, idempotencyKey); err != nil {
		return nil, err
	}
	if err := ValidateUsername(input.Username); err != nil {
		return nil, err
	}
	password, generated, err := resolvePassword(input.Password, input.GeneratePassword)
	if err != nil {
		return nil, err
	}
	if replay, err := s.repository.FindRevisionByIdempotencyTx(ctx, tx, idempotencyKey); err != nil {
		return nil, err
	} else if replay != nil {
		return nil, ErrIdempotentReplay
	}

	credential, err := s.encryptCredential("", input.Username, password, CredentialActive, CredentialPanel, actorID)
	if err != nil {
		return nil, err
	}
	created, err := s.repository.CreateTx(ctx, tx, credential)
	if err != nil {
		return nil, err
	}
	return s.applyStaged(ctx, tx, actorID, idempotencyKey, created, generated)
}

func (s *Service) Update(ctx context.Context, tx *sql.Tx, actorID, idempotencyKey string, input UpdateInput) (*Mutation, error) {
	if err := validateMutationIdentity(actorID, idempotencyKey); err != nil {
		return nil, err
	}
	if input.ID == "" || input.ExpectedRevision <= 0 {
		return nil, errors.New("runtimecred: credential id and expected revision are required")
	}
	if err := ValidateUsername(input.Username); err != nil {
		return nil, err
	}
	if !validCredentialStatus(input.Status) {
		return nil, ErrInvalidCredentialStatus
	}
	credentials, err := s.repository.ListTx(ctx, tx)
	if err != nil {
		return nil, err
	}
	if wouldRemoveLastActive(credentials, input.ID, input.Status) {
		return nil, ErrLastActiveCredential
	}
	if replay, err := s.repository.FindRevisionByIdempotencyTx(ctx, tx, idempotencyKey); err != nil {
		return nil, err
	} else if replay != nil {
		return nil, ErrIdempotentReplay
	}
	updated, err := s.repository.UpdateTx(ctx, tx, input.ID, input.ExpectedRevision, input.Username, input.Status, actorID)
	if err != nil {
		return nil, err
	}
	return s.applyStaged(ctx, tx, actorID, idempotencyKey, updated, "")
}

func (s *Service) Rotate(ctx context.Context, tx *sql.Tx, actorID, idempotencyKey string, input RotateInput) (*Mutation, error) {
	if err := validateMutationIdentity(actorID, idempotencyKey); err != nil {
		return nil, err
	}
	if input.ID == "" || input.ExpectedRevision <= 0 {
		return nil, errors.New("runtimecred: credential id and expected revision are required")
	}
	password, generated, err := resolvePassword(input.Password, input.GeneratePassword)
	if err != nil {
		return nil, err
	}
	if replay, err := s.repository.FindRevisionByIdempotencyTx(ctx, tx, idempotencyKey); err != nil {
		return nil, err
	} else if replay != nil {
		return nil, ErrIdempotentReplay
	}
	ciphertext, nonce, hash, err := encryptMaterial(s.key, password)
	if err != nil {
		return nil, err
	}
	rotated, err := s.repository.RotateTx(ctx, tx, input.ID, input.ExpectedRevision, hash, ciphertext, nonce, s.keyID, actorID)
	if err != nil {
		return nil, err
	}
	return s.applyStaged(ctx, tx, actorID, idempotencyKey, rotated, generated)
}

func (s *Service) Revoke(ctx context.Context, tx *sql.Tx, actorID, idempotencyKey string, input RevokeInput) (*Mutation, error) {
	if err := validateMutationIdentity(actorID, idempotencyKey); err != nil {
		return nil, err
	}
	if input.ID == "" || input.ExpectedRevision <= 0 {
		return nil, errors.New("runtimecred: credential id and expected revision are required")
	}
	credentials, err := s.repository.ListTx(ctx, tx)
	if err != nil {
		return nil, err
	}
	if wouldRemoveLastActive(credentials, input.ID, CredentialRevoked) {
		return nil, ErrLastActiveCredential
	}
	if replay, err := s.repository.FindRevisionByIdempotencyTx(ctx, tx, idempotencyKey); err != nil {
		return nil, err
	} else if replay != nil {
		return nil, ErrIdempotentReplay
	}
	revoked, err := s.repository.RevokeTx(ctx, tx, input.ID, input.ExpectedRevision, actorID)
	if err != nil {
		return nil, err
	}
	return s.applyStaged(ctx, tx, actorID, idempotencyKey, revoked, "")
}

func (s *Service) applyStaged(ctx context.Context, tx *sql.Tx, actorID, idempotencyKey string, changed Credential, generatedPassword string) (*Mutation, error) {
	credentials, err := s.repository.ListTx(ctx, tx)
	if err != nil {
		return nil, err
	}
	desired, canonical, manifest, err := s.buildDesired(credentials)
	if err != nil {
		return nil, err
	}
	inspection, err := s.agent.Inspect(ctx)
	if err != nil {
		return nil, fmt.Errorf("runtimecred: inspect runtime: %w", err)
	}
	if len(inspection.CaddySHA256) != 64 {
		return nil, errors.New("runtimecred: runtime inspection returned invalid Caddy SHA")
	}

	configCiphertext, configNonce, err := EncryptSecret(s.key, canonical)
	if err != nil {
		return nil, err
	}
	packedCiphertext := append(append([]byte(nil), configNonce...), configCiphertext...)
	checksum := sha256.Sum256(canonical)
	revision, err := s.repository.CreateRuntimeRevisionTx(ctx, tx, RuntimeRevision{
		State:                "staged",
		IdempotencyKey:       idempotencyKey,
		ConfigChecksumSHA256: hex.EncodeToString(checksum[:]),
		ConfigCiphertext:     packedCiphertext,
		EncryptionKeyID:      s.keyID,
		Manifest:             manifest,
		CreatedByActorID:     actorID,
	})
	zeroBytes(canonical)
	if err != nil {
		return nil, err
	}

	request := AgentApplyRequest{
		ExpectedCaddySHA256: inspection.CaddySHA256,
		Revision:            fmt.Sprintf("%d", revision.RevisionNo),
		Credentials:         desired,
	}
	applied, err := s.agent.Apply(ctx, request)
	zeroAgentPasswords(request.Credentials)
	if err != nil {
		_ = s.repository.MarkRevisionFailedTx(ctx, tx, revision.ID, "runtime_apply_failed")
		_ = tx.Rollback()
		return nil, fmt.Errorf("runtimecred: runtime apply failed: %w", err)
	}
	metadata := AppliedRuntimeMetadata{
		PreviousSHA256: applied.PreviousSHA256,
		AppliedSHA256:  applied.AppliedSHA256,
		BackupID:       applied.BackupID,
		MainPID:        applied.MainPID,
		NRestarts:      applied.NRestarts,
	}
	if err := s.repository.MarkRevisionAppliedTx(ctx, tx, revision.ID, metadata); err != nil {
		rollbackErr := s.agent.Rollback(ctx, applied.BackupID)
		_ = tx.Rollback()
		if rollbackErr != nil {
			return nil, fmt.Errorf("%w: mark revision applied failed and runtime rollback failed", ErrConsistency)
		}
		return nil, fmt.Errorf("%w: mark revision applied failed", ErrConsistency)
	}
	return &Mutation{
		agent:             s.agent,
		backupID:          applied.BackupID,
		credential:        credentialView(changed),
		runtimeRevisionID: revision.ID,
		generatedPassword: generatedPassword,
	}, nil
}

func (s *Service) buildDesired(credentials []Credential) ([]AgentCredential, []byte, map[string]any, error) {
	active := 0
	desired := make([]AgentCredential, 0, len(credentials))
	safeManifest := make([]map[string]any, 0, len(credentials))
	for _, credential := range credentials {
		safeManifest = append(safeManifest, map[string]any{
			"id": credential.ID, "username": credential.Username, "status": credential.Status, "revision": credential.Revision,
		})
		if credential.Status != CredentialActive {
			continue
		}
		active++
		plaintext, err := DecryptSecret(s.key, credential.secretNonce, credential.secretCiphertext)
		if err != nil {
			zeroAgentPasswords(desired)
			return nil, nil, nil, fmt.Errorf("runtimecred: decrypt active credential: %w", err)
		}
		password := string(plaintext)
		zeroBytes(plaintext)
		desired = append(desired, AgentCredential{
			ID: credential.ID, Username: credential.Username, Password: password, Status: credential.Status,
		})
	}
	if active == 0 {
		return nil, nil, nil, ErrLastActiveCredential
	}
	canonical, err := json.Marshal(struct {
		Protocol    string            `json:"protocol"`
		Credentials []AgentCredential `json:"credentials"`
	}{Protocol: "naive", Credentials: desired})
	if err != nil {
		zeroAgentPasswords(desired)
		return nil, nil, nil, err
	}
	manifest := map[string]any{"protocol": "naive", "credentials": safeManifest}
	return desired, canonical, manifest, nil
}

func (m *Mutation) Credential() CredentialView {
	if m == nil {
		return CredentialView{}
	}
	return m.credential
}

func (m *Mutation) RuntimeRevisionID() string {
	if m == nil {
		return ""
	}
	return m.runtimeRevisionID
}

func (m *Mutation) CommitAndFinalize(ctx context.Context, tx *sql.Tx) error {
	if m == nil || tx == nil {
		return errors.New("runtimecred: mutation and transaction are required")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.committed {
		return errors.New("runtimecred: mutation already committed")
	}
	if err := tx.Commit(); err != nil {
		rollbackErr := m.agent.Rollback(ctx, m.backupID)
		_ = tx.Rollback()
		m.generatedPassword = ""
		if rollbackErr != nil {
			return fmt.Errorf("%w: database commit failed and runtime rollback failed", ErrConsistency)
		}
		return fmt.Errorf("%w: database commit failed; runtime rolled back", ErrConsistency)
	}
	m.committed = true
	return nil
}

func (m *Mutation) TakeGeneratedPassword() string {
	if m == nil {
		return ""
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if !m.committed || m.secretTaken || m.generatedPassword == "" {
		return ""
	}
	m.secretTaken = true
	password := m.generatedPassword
	m.generatedPassword = ""
	return password
}

func (s *Service) encryptCredential(id, username, password string, status CredentialStatus, origin CredentialOrigin, actorID string) (Credential, error) {
	ciphertext, nonce, hash, err := encryptMaterial(s.key, password)
	if err != nil {
		return Credential{}, err
	}
	now := time.Now().UTC()
	return Credential{
		ID: id, Username: username, EncryptionKeyID: s.keyID, Status: status, Origin: origin,
		Revision: 1, CreatedAt: now, UpdatedAt: now,
		secretHash: hash, secretCiphertext: ciphertext, secretNonce: nonce,
	}, nil
}

func encryptMaterial(key []byte, password string) ([]byte, []byte, [32]byte, error) {
	if err := ValidatePassword(password, false); err != nil {
		return nil, nil, [32]byte{}, err
	}
	plaintext := []byte(password)
	ciphertext, nonce, err := EncryptSecret(key, plaintext)
	hash := HashSecret(plaintext)
	zeroBytes(plaintext)
	if err != nil {
		return nil, nil, [32]byte{}, err
	}
	return ciphertext, nonce, hash, nil
}

func resolvePassword(password string, generate bool) (string, string, error) {
	if generate {
		if password != "" {
			return "", "", errors.New("runtimecred: choose supplied or generated password, not both")
		}
		generated, err := GeneratePassword()
		if err != nil {
			return "", "", err
		}
		return generated, generated, nil
	}
	if err := ValidatePassword(password, false); err != nil {
		return "", "", err
	}
	return password, "", nil
}

func validateMutationIdentity(actorID, idempotencyKey string) error {
	if strings.TrimSpace(actorID) == "" {
		return errors.New("runtimecred: actor id is required")
	}
	if len(idempotencyKey) < 8 || len(idempotencyKey) > 160 {
		return errors.New("runtimecred: idempotency key length must be 8-160 bytes")
	}
	return nil
}

func wouldRemoveLastActive(credentials []Credential, targetID string, nextStatus CredentialStatus) bool {
	if nextStatus == CredentialActive {
		return false
	}
	active := 0
	targetActive := false
	for _, credential := range credentials {
		if credential.Status == CredentialActive {
			active++
			if credential.ID == targetID {
				targetActive = true
			}
		}
	}
	return targetActive && active <= 1
}

func validCredentialStatus(status CredentialStatus) bool {
	return status == CredentialActive || status == CredentialDisabled || status == CredentialRevoked
}

func credentialView(credential Credential) CredentialView {
	return CredentialView{
		ID: credential.ID, Username: credential.Username, Status: credential.Status, Origin: credential.Origin,
		Revision: credential.Revision, CreatedAt: credential.CreatedAt, UpdatedAt: credential.UpdatedAt,
		RotatedAt: credential.RotatedAt, RevokedAt: credential.RevokedAt,
	}
}

func zeroBytes(value []byte) {
	for i := range value {
		value[i] = 0
	}
}

func zeroAgentPasswords(credentials []AgentCredential) {
	for i := range credentials {
		credentials[i].Password = ""
	}
}
