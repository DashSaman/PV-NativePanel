package runtimecred

import (
	"context"
	"errors"
	"testing"
)

func TestRuntimeMutationAbortRollsBackAppliedRuntimeAndClearsSecret(t *testing.T) {
	agent := &fakeRuntimeAgent{}
	tx := newDriverTx(t, nil)
	mutation := &Mutation{
		agent:             agent,
		backupID:          "backup-before-business-bind",
		generatedPassword: "must-not-survive-abort",
	}

	if err := mutation.AbortAndRollback(context.Background(), tx); err != nil {
		t.Fatalf("AbortAndRollback() error = %v", err)
	}
	if len(agent.rollbackIDs) != 1 || agent.rollbackIDs[0] != "backup-before-business-bind" {
		t.Fatalf("rollback IDs = %#v, want one exact backup", agent.rollbackIDs)
	}
	if got := mutation.TakeGeneratedPassword(); got != "" {
		t.Fatalf("aborted mutation leaked generated password %q", got)
	}
	if err := mutation.CommitAndFinalize(context.Background(), tx); err == nil {
		t.Fatal("aborted mutation was still committable")
	}
}

func TestRuntimeMutationAbortReturnsReconciliationRequiredWhenRollbackFails(t *testing.T) {
	agent := &failingRollbackAgent{fakeRuntimeAgent: fakeRuntimeAgent{}}
	tx := newDriverTx(t, nil)
	mutation := &Mutation{agent: agent, backupID: "backup-fail", generatedPassword: "secret"}

	err := mutation.AbortAndRollback(context.Background(), tx)
	if !errors.Is(err, ErrReconciliationRequired) {
		t.Fatalf("AbortAndRollback() error = %v, want ErrReconciliationRequired", err)
	}
	if got := mutation.TakeGeneratedPassword(); got != "" {
		t.Fatalf("failed rollback path leaked generated password %q", got)
	}
}

type failingRollbackAgent struct {
	fakeRuntimeAgent
}

func (f *failingRollbackAgent) Rollback(context.Context, string) error {
	return errors.New("rollback failed")
}
