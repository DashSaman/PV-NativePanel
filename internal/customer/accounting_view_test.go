package customer

import (
	"testing"
	"time"
)

func TestApplyAccountingSnapshotExposesExactUsageAndPresence(t *testing.T) {
	quota := int64(1000)
	remaining := int64(500)
	lastOnline := time.Date(2026, 8, 29, 20, 15, 0, 0, time.UTC)
	view := CustomerView{
		Status:       UserActive,
		ServiceState: TermActive,
		QuotaBytes:   &quota,
		StartPolicy:  StartOnCreation,
	}

	ApplyAccountingSnapshot(&view, AccountingSnapshot{
		Present:            true,
		AccountingComplete: true,
		UploadBytes:        200,
		DownloadBytes:      300,
		UsedBytes:          500,
		RemainingBytes:     &remaining,
		Online:             true,
		OnlineSessions:     2,
		LastOnline:         &lastOnline,
	})

	if !view.UsageCapability.Available || view.UsageCapability.Reason != "" {
		t.Fatalf("exact accounting should be available: %#v", view.UsageCapability)
	}
	if view.UploadBytes != 200 || view.DownloadBytes != 300 || view.UsedBytes != 500 {
		t.Fatalf("exact byte projection lost: %#v", view)
	}
	if view.RemainingBytes == nil || *view.RemainingBytes != 500 {
		t.Fatalf("remaining bytes lost: %#v", view.RemainingBytes)
	}
	if !view.Online || view.OnlineSessions != 2 || view.LastOnline == nil || !view.LastOnline.Equal(lastOnline) {
		t.Fatalf("presence projection lost: %#v", view)
	}
	if view.StatusDimensions.Presence != PresenceOnline || view.StatusDimensions.Quota != QuotaHealthy {
		t.Fatalf("status dimensions did not consume exact accounting: %#v", view.StatusDimensions)
	}
}

func TestApplyAccountingSnapshotKeepsIncompleteUsageHonest(t *testing.T) {
	quota := int64(1000)
	remaining := int64(700)
	view := CustomerView{
		Status:       UserActive,
		ServiceState: TermActive,
		QuotaBytes:   &quota,
		StartPolicy:  StartOnCreation,
	}

	ApplyAccountingSnapshot(&view, AccountingSnapshot{
		Present:            true,
		AccountingComplete: false,
		UsedBytes:          300,
		RemainingBytes:     &remaining,
		Online:             false,
	})

	if view.UsageCapability.Available || view.UsageCapability.Reason != "accounting_incomplete" {
		t.Fatalf("incomplete accounting must remain unavailable: %#v", view.UsageCapability)
	}
	if view.UsedBytes != 300 {
		t.Fatalf("known ledger bytes should remain visible: %d", view.UsedBytes)
	}
	if view.StatusDimensions.Presence != PresenceUnknown || view.StatusDimensions.Quota != QuotaUnavailable {
		t.Fatalf("incomplete accounting must not fabricate presence/quota: %#v", view.StatusDimensions)
	}
}
