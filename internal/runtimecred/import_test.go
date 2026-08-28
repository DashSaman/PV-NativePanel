package runtimecred

import (
	"context"
	"strings"
	"testing"
)

type guardedImportAgent struct {
	inspection    AgentInspection
	candidateSHA  string
	validateCalls int
	applyCalls    int
}

func (a *guardedImportAgent) Inspect(context.Context) (AgentInspection, error) {
	return a.inspection, nil
}

func (a *guardedImportAgent) Validate(_ context.Context, request AgentApplyRequest) (string, error) {
	a.validateCalls++
	if request.ExpectedCaddySHA256 != a.inspection.CaddySHA256 {
		return "", ErrConsistency
	}
	return a.candidateSHA, nil
}

func (a *guardedImportAgent) Apply(context.Context, AgentApplyRequest) (AgentApplyResult, error) {
	a.applyCalls++
	return AgentApplyResult{}, nil
}

func (a *guardedImportAgent) Rollback(context.Context, string) error { return nil }

func TestImportCurrentEncryptsLiveCredentialWithoutRuntimeMutation(t *testing.T) {
	key := bytesOf(0x55, 32)
	sha := strings.Repeat("a", 64)
	repo := &fakeRuntimeRepository{}
	agent := &guardedImportAgent{
		inspection: AgentInspection{
			CaddySHA256: sha,
			Credentials: []AgentCredential{{Username: "live.user", Password: "legacy-pass", Status: CredentialActive}},
		},
		candidateSHA: sha,
	}
	service, err := NewService(repo, agent, key, "runtime-v1")
	if err != nil {
		t.Fatal(err)
	}
	tx := newDriverTx(t, nil)
	credentials, err := service.ImportCurrent(context.Background(), tx, "actor-id", "idem-import-0001")
	if err != nil {
		t.Fatal(err)
	}
	if len(credentials) != 1 || credentials[0].Username != "live.user" || credentials[0].Origin != CredentialImported {
		t.Fatalf("unexpected imported credentials: %+v", credentials)
	}
	if agent.validateCalls != 1 || agent.applyCalls != 0 {
		t.Fatalf("validate=%d apply=%d, want validate=1 apply=0", agent.validateCalls, agent.applyCalls)
	}
	if len(repo.credentials) != 1 {
		t.Fatalf("stored credentials=%d, want 1", len(repo.credentials))
	}
	plaintext, err := DecryptSecret(key, repo.credentials[0].secretNonce, repo.credentials[0].secretCiphertext)
	if err != nil {
		t.Fatal(err)
	}
	if string(plaintext) != "legacy-pass" {
		t.Fatalf("import changed live password")
	}
	zeroBytes(plaintext)
}

func TestImportCurrentFailsClosedWhenCandidateDiffersFromLiveCaddy(t *testing.T) {
	key := bytesOf(0x66, 32)
	repo := &fakeRuntimeRepository{}
	agent := &guardedImportAgent{
		inspection: AgentInspection{
			CaddySHA256: strings.Repeat("a", 64),
			Credentials: []AgentCredential{{Username: "live.user", Password: "legacy-pass", Status: CredentialActive}},
		},
		candidateSHA: strings.Repeat("b", 64),
	}
	service, _ := NewService(repo, agent, key, "runtime-v1")
	tx := newDriverTx(t, nil)
	if _, err := service.ImportCurrent(context.Background(), tx, "actor-id", "idem-import-0002"); err == nil {
		t.Fatal("ImportCurrent accepted non-equivalent candidate")
	}
	if agent.applyCalls != 0 {
		t.Fatalf("import mutated runtime: apply calls=%d", agent.applyCalls)
	}
}
