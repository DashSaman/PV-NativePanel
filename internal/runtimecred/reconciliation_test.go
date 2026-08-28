package runtimecred

import (
	"context"
	"errors"
	"testing"
)

type rollbackFailingAgent struct {
	fakeRuntimeAgent
}

func (a *rollbackFailingAgent) Rollback(context.Context, string) error {
	return errors.New("rollback injected failure")
}

func TestCommitFailureAndRollbackFailureRequiresReconciliation(t *testing.T) {
	key := bytesOf(0x77, 32)
	repo := &fakeRuntimeRepository{credentials: []Credential{storedCredential(t, key, "existing-id", "existing.user", "existing password 123", CredentialActive)}}
	agent := &rollbackFailingAgent{}
	service, err := NewService(repo, agent, key, "runtime-v1")
	if err != nil {
		t.Fatal(err)
	}
	tx := newDriverTx(t, errors.New("commit injected failure"))

	mutation, err := service.Create(context.Background(), tx, "actor-id", "idem-reconcile-0001", CreateInput{Username: "new.user", GeneratePassword: true})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if err := mutation.CommitAndFinalize(context.Background(), tx); !errors.Is(err, ErrReconciliationRequired) {
		t.Fatalf("CommitAndFinalize() error = %v, want ErrReconciliationRequired", err)
	}
	if password := mutation.TakeGeneratedPassword(); password != "" {
		t.Fatalf("generated password leaked after reconciliation failure: %q", password)
	}
}
