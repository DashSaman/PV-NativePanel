package runtimecred

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"sync"
	"testing"
)

type fakeRuntimeRepository struct {
	credentials []Credential
	revisions   []RuntimeRevision
}

func (f *fakeRuntimeRepository) ListTx(context.Context, *sql.Tx) ([]Credential, error) {
	out := make([]Credential, len(f.credentials))
	copy(out, f.credentials)
	return out, nil
}

func (f *fakeRuntimeRepository) CreateTx(_ context.Context, _ *sql.Tx, credential Credential) (Credential, error) {
	if credential.ID == "" {
		credential.ID = "created-id"
	}
	credential.Revision = 1
	f.credentials = append(f.credentials, credential)
	return credential, nil
}

func (f *fakeRuntimeRepository) UpdateTx(_ context.Context, _ *sql.Tx, id string, expectedRevision int64, username string, status CredentialStatus, actorID string) (Credential, error) {
	for i := range f.credentials {
		if f.credentials[i].ID != id {
			continue
		}
		if f.credentials[i].Revision != expectedRevision {
			return Credential{}, ErrRevisionConflict
		}
		f.credentials[i].Username = username
		f.credentials[i].Status = status
		f.credentials[i].Revision++
		return f.credentials[i], nil
	}
	return Credential{}, sql.ErrNoRows
}

func (f *fakeRuntimeRepository) RotateTx(_ context.Context, _ *sql.Tx, id string, expectedRevision int64, hash [32]byte, ciphertext, nonce []byte, keyID, actorID string) (Credential, error) {
	for i := range f.credentials {
		if f.credentials[i].ID != id {
			continue
		}
		if f.credentials[i].Revision != expectedRevision {
			return Credential{}, ErrRevisionConflict
		}
		f.credentials[i].secretHash = hash
		f.credentials[i].secretCiphertext = append([]byte(nil), ciphertext...)
		f.credentials[i].secretNonce = append([]byte(nil), nonce...)
		f.credentials[i].EncryptionKeyID = keyID
		f.credentials[i].Revision++
		return f.credentials[i], nil
	}
	return Credential{}, sql.ErrNoRows
}

func (f *fakeRuntimeRepository) RevokeTx(_ context.Context, _ *sql.Tx, id string, expectedRevision int64, actorID string) (Credential, error) {
	for i := range f.credentials {
		if f.credentials[i].ID != id {
			continue
		}
		if f.credentials[i].Revision != expectedRevision {
			return Credential{}, ErrRevisionConflict
		}
		f.credentials[i].Status = CredentialRevoked
		f.credentials[i].Revision++
		return f.credentials[i], nil
	}
	return Credential{}, sql.ErrNoRows
}

func (f *fakeRuntimeRepository) FindRevisionByIdempotencyTx(context.Context, *sql.Tx, string) (*RuntimeRevision, error) {
	return nil, nil
}

func (f *fakeRuntimeRepository) CreateRuntimeRevisionTx(_ context.Context, _ *sql.Tx, revision RuntimeRevision) (RuntimeRevision, error) {
	if revision.ID == "" {
		revision.ID = "runtime-revision-id"
	}
	if revision.RevisionNo == 0 {
		revision.RevisionNo = int64(len(f.revisions) + 1)
	}
	f.revisions = append(f.revisions, revision)
	return revision, nil
}

func (f *fakeRuntimeRepository) MarkRevisionAppliedTx(_ context.Context, _ *sql.Tx, id string, metadata AppliedRuntimeMetadata) error {
	for i := range f.revisions {
		if f.revisions[i].ID == id {
			f.revisions[i].State = "applied"
			f.revisions[i].Applied = metadata
		}
	}
	return nil
}

func (f *fakeRuntimeRepository) MarkRevisionFailedTx(_ context.Context, _ *sql.Tx, id, code string) error {
	for i := range f.revisions {
		if f.revisions[i].ID == id {
			f.revisions[i].State = "failed"
			f.revisions[i].FailureCode = code
		}
	}
	return nil
}

type fakeRuntimeAgent struct {
	mu          sync.Mutex
	applyCalls  int
	rollbackIDs []string
	applyErr    error
	lastApply   AgentApplyRequest
}

func (f *fakeRuntimeAgent) Inspect(context.Context) (AgentInspection, error) {
	return AgentInspection{CaddySHA256: strings.Repeat("a", 64)}, nil
}
func (f *fakeRuntimeAgent) Apply(_ context.Context, request AgentApplyRequest) (AgentApplyResult, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.applyCalls++
	f.lastApply = request
	if f.applyErr != nil {
		return AgentApplyResult{}, f.applyErr
	}
	return AgentApplyResult{PreviousSHA256: strings.Repeat("a", 64), AppliedSHA256: strings.Repeat("b", 64), BackupID: "backup-safe-1", MainPID: 1234}, nil
}
func (f *fakeRuntimeAgent) Rollback(_ context.Context, backupID string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.rollbackIDs = append(f.rollbackIDs, backupID)
	return nil
}

func TestServiceCreateGeneratedPasswordAppliesDesiredStateAndReturnsSecretOnce(t *testing.T) {
	key := bytesOf(0x11, 32)
	repo := &fakeRuntimeRepository{credentials: []Credential{storedCredential(t, key, "existing-id", "existing.user", "existing password 123", CredentialActive)}}
	agent := &fakeRuntimeAgent{}
	service, err := NewService(repo, agent, key, "runtime-v1")
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	tx := newDriverTx(t, nil)
	mutation, err := service.Create(context.Background(), tx, "actor-id", "idem-create-0001", CreateInput{Username: "new.user", GeneratePassword: true})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if agent.applyCalls != 1 {
		t.Fatalf("agent apply calls = %d, want 1", agent.applyCalls)
	}
	if len(agent.lastApply.Credentials) != 2 {
		t.Fatalf("agent desired credentials = %d, want 2", len(agent.lastApply.Credentials))
	}
	password := mutation.TakeGeneratedPassword()
	if len(password) != 32 {
		t.Fatalf("generated password length = %d, want 32", len(password))
	}
	if again := mutation.TakeGeneratedPassword(); again != "" {
		t.Fatalf("generated password was returned twice: %q", again)
	}
	if err := mutation.CommitAndFinalize(context.Background(), tx); err != nil {
		t.Fatalf("CommitAndFinalize() error = %v", err)
	}
	if len(agent.rollbackIDs) != 0 {
		t.Fatalf("unexpected rollback after successful commit: %#v", agent.rollbackIDs)
	}
}

func TestServiceRejectsDisablingLastActiveCredentialBeforeAgentApply(t *testing.T) {
	key := bytesOf(0x22, 32)
	repo := &fakeRuntimeRepository{credentials: []Credential{storedCredential(t, key, "only-id", "only.user", "only password 123", CredentialActive)}}
	agent := &fakeRuntimeAgent{}
	service, err := NewService(repo, agent, key, "runtime-v1")
	if err != nil {
		t.Fatal(err)
	}
	tx := newDriverTx(t, nil)

	_, err = service.Update(context.Background(), tx, "actor-id", "idem-update-0001", UpdateInput{
		ID: "only-id", ExpectedRevision: 1, Username: "only.user", Status: CredentialDisabled,
	})
	if !errors.Is(err, ErrLastActiveCredential) {
		t.Fatalf("Update() error = %v, want ErrLastActiveCredential", err)
	}
	if agent.applyCalls != 0 {
		t.Fatalf("agent apply calls = %d, want 0", agent.applyCalls)
	}
}

func TestServiceAgentFailureDoesNotProduceCommittableMutation(t *testing.T) {
	key := bytesOf(0x33, 32)
	repo := &fakeRuntimeRepository{credentials: []Credential{storedCredential(t, key, "existing-id", "existing.user", "existing password 123", CredentialActive)}}
	agent := &fakeRuntimeAgent{applyErr: errors.New("apply failed")}
	service, _ := NewService(repo, agent, key, "runtime-v1")
	tx := newDriverTx(t, nil)

	mutation, err := service.Create(context.Background(), tx, "actor-id", "idem-create-0002", CreateInput{Username: "new.user", Password: "new password 123"})
	if err == nil || mutation != nil {
		t.Fatalf("Create() mutation=%#v err=%v, want nil mutation and error", mutation, err)
	}
}

func TestCommitFailureCompensatesAppliedRuntime(t *testing.T) {
	key := bytesOf(0x44, 32)
	repo := &fakeRuntimeRepository{credentials: []Credential{storedCredential(t, key, "existing-id", "existing.user", "existing password 123", CredentialActive)}}
	agent := &fakeRuntimeAgent{}
	service, _ := NewService(repo, agent, key, "runtime-v1")
	tx := newDriverTx(t, errors.New("commit injected failure"))

	mutation, err := service.Create(context.Background(), tx, "actor-id", "idem-create-0003", CreateInput{Username: "new.user", Password: "new password 123"})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if err := mutation.CommitAndFinalize(context.Background(), tx); !errors.Is(err, ErrConsistency) {
		t.Fatalf("CommitAndFinalize() error = %v, want ErrConsistency", err)
	}
	if len(agent.rollbackIDs) != 1 || agent.rollbackIDs[0] != "backup-safe-1" {
		t.Fatalf("rollback ids = %#v, want backup-safe-1", agent.rollbackIDs)
	}
}

func TestCredentialViewJSONNeverContainsSecretMaterial(t *testing.T) {
	view := CredentialView{ID: "id", Username: "user", Status: CredentialActive, Origin: CredentialPanel, Revision: 1}
	encoded, err := json.Marshal(view)
	if err != nil {
		t.Fatal(err)
	}
	lower := strings.ToLower(string(encoded))
	for _, forbidden := range []string{"password", "ciphertext", "nonce", "secret_hash", "secret"} {
		if strings.Contains(lower, forbidden) {
			t.Fatalf("CredentialView JSON contains forbidden field %q: %s", forbidden, encoded)
		}
	}
}

func storedCredential(t *testing.T, key []byte, id, username, password string, status CredentialStatus) Credential {
	t.Helper()
	ciphertext, nonce, err := EncryptSecret(key, []byte(password))
	if err != nil {
		t.Fatal(err)
	}
	return Credential{
		ID: id, Username: username, EncryptionKeyID: "runtime-v1", Status: status, Origin: CredentialImported, Revision: 1,
		secretHash: HashSecret([]byte(password)), secretCiphertext: ciphertext, secretNonce: nonce,
	}
}

func bytesOf(value byte, count int) []byte {
	out := make([]byte, count)
	for i := range out {
		out[i] = value
	}
	return out
}

type txConnector struct{ commitErr error }
type txConn struct{ commitErr error }
type txDriverTx struct{ commitErr error }

func (c txConnector) Connect(context.Context) (driver.Conn, error) { return txConn{commitErr: c.commitErr}, nil }
func (c txConnector) Driver() driver.Driver                           { return txDriver{} }
type txDriver struct{}
func (txDriver) Open(string) (driver.Conn, error)                     { return txConn{}, nil }
func (c txConn) Prepare(string) (driver.Stmt, error)                  { return nil, errors.New("not implemented") }
func (c txConn) Close() error                                        { return nil }
func (c txConn) Begin() (driver.Tx, error)                            { return txDriverTx{commitErr: c.commitErr}, nil }
func (t txDriverTx) Commit() error                                   { return t.commitErr }
func (t txDriverTx) Rollback() error                                 { return nil }

func newDriverTx(t *testing.T, commitErr error) *sql.Tx {
	t.Helper()
	db := sql.OpenDB(txConnector{commitErr: commitErr})
	t.Cleanup(func() { _ = db.Close() })
	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	return tx
}

var _ io.Closer = txConn{}
