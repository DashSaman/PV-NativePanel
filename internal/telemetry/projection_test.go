package telemetry

import (
	"errors"
	"testing"
	"time"
)

func int64ptr(v int64) *int64 { return &v }

func TestBuildReadModelUsesSharedServiceTermBudgetAcrossSessions(t *testing.T) {
	now := time.Date(2026, 8, 29, 12, 0, 0, 0, time.UTC)
	first := now.Add(-time.Hour)
	model, err := BuildReadModel(
		TermPolicy{ServiceTermID: testTermID, QuotaBytes: int64ptr(100)},
		TermUsage{ServiceTermID: testTermID, UploadBytes: 30, DownloadBytes: 40, FirstConnectedAt: &first},
		[]SessionSnapshot{
			{Key: SessionKey{RuntimeCredentialID: testRuntimeID, NodeID: testNodeID, BootID: testBootID, SessionID: testSessionID}, LastObservedAt: now.Add(-time.Second)},
			{Key: SessionKey{RuntimeCredentialID: testRuntimeID, NodeID: "node-b", BootID: testBootID, SessionID: "55555555-5555-4555-8555-555555555555"}, LastObservedAt: now.Add(-2 * time.Second)},
		},
		now,
		30*time.Second,
		true,
	)
	if err != nil {
		t.Fatalf("BuildReadModel() error = %v", err)
	}
	if model.UploadBytes != 30 || model.DownloadBytes != 40 || model.UsedBytes != 70 {
		t.Fatalf("usage = %+v, want upload=30 download=40 used=70", model)
	}
	if model.RemainingBytes == nil || *model.RemainingBytes != 30 {
		t.Fatalf("remaining = %v, want shared remaining 30", model.RemainingBytes)
	}
	if model.QuotaState != QuotaActive || !model.Online || model.SessionCount != 2 {
		t.Fatalf("presence/quota = %+v, want active online two sessions", model)
	}
	if model.FirstConnectedAt == nil || !model.FirstConnectedAt.Equal(first) {
		t.Fatalf("first connected = %v, want %v", model.FirstConnectedAt, first)
	}
}

func TestBuildReadModelStaleUnclosedSessionIsOfflineAndIncomplete(t *testing.T) {
	now := time.Date(2026, 8, 29, 12, 0, 0, 0, time.UTC)
	last := now.Add(-2 * time.Minute)
	model, err := BuildReadModel(
		TermPolicy{ServiceTermID: testTermID},
		TermUsage{ServiceTermID: testTermID},
		[]SessionSnapshot{{Key: SessionKey{RuntimeCredentialID: testRuntimeID, NodeID: testNodeID, BootID: testBootID, SessionID: testSessionID}, LastObservedAt: last}},
		now,
		30*time.Second,
		true,
	)
	if err != nil {
		t.Fatalf("BuildReadModel() error = %v", err)
	}
	if model.Online || model.SessionCount != 0 {
		t.Fatalf("stale session reported online: %+v", model)
	}
	if model.AccountingComplete {
		t.Fatal("stale unclosed session must mark accounting incomplete")
	}
	if model.LastOnline == nil || !model.LastOnline.Equal(last) {
		t.Fatalf("LastOnline = %v, want %v", model.LastOnline, last)
	}
}

func TestBuildReadModelFinalSessionIsOfflineButComplete(t *testing.T) {
	now := time.Date(2026, 8, 29, 12, 0, 0, 0, time.UTC)
	last := now.Add(-time.Second)
	model, err := BuildReadModel(
		TermPolicy{ServiceTermID: testTermID},
		TermUsage{ServiceTermID: testTermID, UploadBytes: 1, DownloadBytes: 2},
		[]SessionSnapshot{{Key: SessionKey{RuntimeCredentialID: testRuntimeID, NodeID: testNodeID, BootID: testBootID, SessionID: testSessionID}, LastObservedAt: last, Final: true}},
		now,
		30*time.Second,
		true,
	)
	if err != nil {
		t.Fatalf("BuildReadModel() error = %v", err)
	}
	if model.Online || model.SessionCount != 0 || !model.AccountingComplete {
		t.Fatalf("final session model = %+v, want offline complete", model)
	}
}

func TestBuildReadModelTelemetryFailureRemovesExactnessClaim(t *testing.T) {
	now := time.Date(2026, 8, 29, 12, 0, 0, 0, time.UTC)
	model, err := BuildReadModel(
		TermPolicy{ServiceTermID: testTermID, QuotaBytes: int64ptr(100)},
		TermUsage{ServiceTermID: testTermID, UploadBytes: 10, DownloadBytes: 20},
		nil,
		now,
		30*time.Second,
		false,
	)
	if err != nil {
		t.Fatalf("BuildReadModel() error = %v", err)
	}
	if model.AccountingComplete {
		t.Fatal("telemetry failure must make accounting_complete=false")
	}
}

func TestBuildReadModelQuotaDepletionAndUnlimited(t *testing.T) {
	now := time.Date(2026, 8, 29, 12, 0, 0, 0, time.UTC)
	depleted, err := BuildReadModel(
		TermPolicy{ServiceTermID: testTermID, QuotaBytes: int64ptr(50)},
		TermUsage{ServiceTermID: testTermID, UploadBytes: 20, DownloadBytes: 35},
		nil,
		now,
		30*time.Second,
		true,
	)
	if err != nil {
		t.Fatalf("depleted BuildReadModel() error = %v", err)
	}
	if depleted.QuotaState != QuotaDepleted || depleted.RemainingBytes == nil || *depleted.RemainingBytes != 0 {
		t.Fatalf("depleted = %+v, want depleted with zero remaining", depleted)
	}

	unlimited, err := BuildReadModel(
		TermPolicy{ServiceTermID: testTermID},
		TermUsage{ServiceTermID: testTermID, UploadBytes: 20, DownloadBytes: 35},
		nil,
		now,
		30*time.Second,
		true,
	)
	if err != nil {
		t.Fatalf("unlimited BuildReadModel() error = %v", err)
	}
	if unlimited.QuotaState != QuotaUnlimited || unlimited.RemainingBytes != nil {
		t.Fatalf("unlimited = %+v, want unlimited and nil remaining", unlimited)
	}
}

func TestBuildReadModelDoesNotCarryUsageAcrossServiceTerms(t *testing.T) {
	now := time.Date(2026, 8, 29, 12, 0, 0, 0, time.UTC)
	_, err := BuildReadModel(
		TermPolicy{ServiceTermID: "77777777-7777-4777-8777-777777777777", QuotaBytes: int64ptr(100)},
		TermUsage{ServiceTermID: testTermID, UploadBytes: 90, DownloadBytes: 0},
		nil,
		now,
		30*time.Second,
		true,
	)
	if !errors.Is(err, ErrServiceTermMismatch) {
		t.Fatalf("BuildReadModel() error = %v, want ErrServiceTermMismatch", err)
	}
}

func TestBuildReadModelRejectsInvalidStaleWindowAndOverflow(t *testing.T) {
	now := time.Date(2026, 8, 29, 12, 0, 0, 0, time.UTC)
	_, err := BuildReadModel(TermPolicy{ServiceTermID: testTermID}, TermUsage{ServiceTermID: testTermID}, nil, now, 0, true)
	if !errors.Is(err, ErrInvalidProjection) {
		t.Fatalf("zero stale timeout error = %v, want ErrInvalidProjection", err)
	}
	_, err = BuildReadModel(
		TermPolicy{ServiceTermID: testTermID},
		TermUsage{ServiceTermID: testTermID, UploadBytes: int64(^uint64(0) >> 1), DownloadBytes: 1},
		nil,
		now,
		30*time.Second,
		true,
	)
	if !errors.Is(err, ErrInvalidProjection) {
		t.Fatalf("overflow error = %v, want ErrInvalidProjection", err)
	}
}
