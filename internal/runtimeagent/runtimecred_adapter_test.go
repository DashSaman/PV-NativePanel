package runtimeagent

import (
	"context"
	"testing"

	"github.com/DashSaman/PV-NaivePanel/internal/runtimecred"
)

type fakeRuntimeRPC struct {
	inspect InspectResponse
	apply   ApplyResponse
	last    ApplyRequest
	backup  string
}

func (f *fakeRuntimeRPC) Inspect(context.Context) (InspectResponse, error) { return f.inspect, nil }
func (f *fakeRuntimeRPC) Apply(_ context.Context, request ApplyRequest) (ApplyResponse, error) {
	f.last = request
	return f.apply, nil
}
func (f *fakeRuntimeRPC) Rollback(_ context.Context, request RollbackRequest) (RollbackResponse, error) {
	f.backup = request.BackupID
	return RollbackResponse{}, nil
}

func TestRuntimeCredAdapterMapsTypedProtocol(t *testing.T) {
	rpc := &fakeRuntimeRPC{
		inspect: InspectResponse{CaddySHA256: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},
		apply:   ApplyResponse{PreviousSHA256: "old", AppliedSHA256: "new", BackupID: "backup-1", MainPID: 42, NRestarts: 0},
	}
	adapter := NewRuntimeCredAdapter(rpc)
	inspection, err := adapter.Inspect(context.Background())
	if err != nil || inspection.CaddySHA256 != rpc.inspect.CaddySHA256 {
		t.Fatalf("inspect=(%+v,%v)", inspection, err)
	}
	result, err := adapter.Apply(context.Background(), runtimecred.AgentApplyRequest{
		ExpectedCaddySHA256: rpc.inspect.CaddySHA256,
		Revision:            "7",
		Credentials:         []runtimecred.AgentCredential{{ID: "c1", Username: "alice", Password: "safe-password-123", Status: runtimecred.CredentialActive}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if rpc.last.Desired.Revision != "7" || len(rpc.last.Desired.Credentials) != 1 || rpc.last.Desired.Credentials[0].Username != "alice" {
		t.Fatalf("unexpected apply mapping: %+v", rpc.last)
	}
	if result.BackupID != "backup-1" || result.MainPID != 42 {
		t.Fatalf("unexpected result: %+v", result)
	}
	if err := adapter.Rollback(context.Background(), "backup-1"); err != nil {
		t.Fatal(err)
	}
	if rpc.backup != "backup-1" {
		t.Fatalf("rollback backup=%q", rpc.backup)
	}
}
