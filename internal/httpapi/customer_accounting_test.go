package httpapi

import (
	"testing"
	"time"

	"github.com/DashSaman/PV-NaivePanel/internal/customer"
	"github.com/DashSaman/PV-NaivePanel/internal/telemetry"
)

func knownZeroBaseline(at time.Time) customer.AccountingBaseline {
	upload, download := int64(0), int64(0)
	return customer.AccountingBaseline{
		State: customer.AccountingBaselineKnown, Source: customer.AccountingBaselineFreshManagedTerm,
		CutoffAt: at, UploadBytes: &upload, DownloadBytes: &download,
	}
}

func unknownLegacyBaseline(at time.Time) customer.AccountingBaseline {
	return customer.AccountingBaseline{
		State: customer.AccountingBaselineUnknown, Source: customer.AccountingBaselineLegacyUnavailable, CutoffAt: at,
	}
}

func TestApplyCustomerAccountingPublishesExactKnownTotalAndPresence(t *testing.T) {
	quota := int64(1000)
	remaining := int64(150)
	lastOnline := time.Date(2026, 8, 29, 20, 0, 0, 0, time.UTC)
	view := customer.CustomerView{
		UserID: "user-1", QuotaBytes: &quota, Status: customer.UserActive,
		ServiceState: customer.TermActive, StartPolicy: customer.StartOnCreation,
		AccountingBaseline: knownZeroBaseline(time.Date(2026, 8, 29, 19, 0, 0, 0, time.UTC)),
	}
	model := telemetry.ReadModel{
		ServiceTermID: "11111111-1111-4111-8111-111111111111",
		UploadBytes:   300, DownloadBytes: 550, UsedBytes: 850, QuotaBytes: &quota,
		RemainingBytes: &remaining, QuotaState: telemetry.QuotaActive, LastOnline: &lastOnline,
		Online: true, SessionCount: 2, AccountingComplete: true,
	}

	applyCustomerAccounting(&view, model, time.Date(2026, 8, 29, 20, 0, 30, 0, time.UTC))

	if !view.UsageCapability.Available || view.Usage == nil {
		t.Fatal("known complete accounting was not published as available")
	}
	if view.Usage.DirectUsedBytes != 850 || view.Usage.UsedBytes == nil || *view.Usage.UsedBytes != 850 {
		t.Fatalf("unexpected usage snapshot: %#v", view.Usage)
	}
	if view.Usage.RemainingBytes == nil || *view.Usage.RemainingBytes != 150 {
		t.Fatalf("unexpected remaining: %#v", view.Usage.RemainingBytes)
	}
	if !view.Usage.Online || view.Usage.SessionCount != 2 || view.StatusDimensions.Presence != customer.PresenceOnline {
		t.Fatalf("presence not derived from direct read model: %#v", view.StatusDimensions)
	}
	if view.StatusDimensions.Quota != customer.QuotaWarning {
		t.Fatalf("quota status=%s want=%s", view.StatusDimensions.Quota, customer.QuotaWarning)
	}
}

func TestApplyCustomerAccountingKeepsLegacyHistoricalTotalUnknown(t *testing.T) {
	quota := int64(1000)
	lastOnline := time.Date(2026, 8, 29, 20, 0, 0, 0, time.UTC)
	view := customer.CustomerView{
		UserID: "legacy-1", QuotaBytes: &quota, Status: customer.UserActive,
		ServiceState: customer.TermActive, StartPolicy: customer.StartOnCreation,
		AccountingBaseline: unknownLegacyBaseline(time.Date(2026, 8, 29, 19, 0, 0, 0, time.UTC)),
	}
	model := telemetry.ReadModel{
		ServiceTermID: "11111111-1111-4111-8111-111111111111",
		UploadBytes:   300, DownloadBytes: 550, UsedBytes: 850, QuotaBytes: &quota,
		LastOnline: &lastOnline, Online: true, SessionCount: 1, AccountingComplete: true,
	}

	applyCustomerAccounting(&view, model, time.Date(2026, 8, 29, 20, 0, 30, 0, time.UTC))

	if view.UsageCapability.Available || view.UsageCapability.Reason != "historical_baseline_unknown" {
		t.Fatalf("legacy unknown capability = %#v", view.UsageCapability)
	}
	if view.Usage == nil || view.Usage.DirectUsedBytes != 850 {
		t.Fatalf("direct exact epoch was lost: %#v", view.Usage)
	}
	if view.Usage.UploadBytes != nil || view.Usage.DownloadBytes != nil || view.Usage.UsedBytes != nil || view.Usage.RemainingBytes != nil {
		t.Fatalf("legacy historical total was fabricated: %#v", view.Usage)
	}
	if view.StatusDimensions.Presence != customer.PresenceOnline {
		t.Fatalf("trusted presence should remain available, got %s", view.StatusDimensions.Presence)
	}
	if view.StatusDimensions.Quota != customer.QuotaUnavailable {
		t.Fatalf("quota must be unavailable with unknown historical total, got %s", view.StatusDimensions.Quota)
	}
}

func TestApplyCustomerAccountingKeepsIncompleteProjectionCapabilityGated(t *testing.T) {
	quota := int64(1000)
	view := customer.CustomerView{
		QuotaBytes: &quota, Status: customer.UserActive, ServiceState: customer.TermActive,
		StartPolicy: customer.StartOnCreation, AccountingBaseline: knownZeroBaseline(time.Now().UTC()),
	}
	model := telemetry.ReadModel{
		ServiceTermID: "11111111-1111-4111-8111-111111111111", UploadBytes: 40, DownloadBytes: 60,
		UsedBytes: 100, QuotaBytes: &quota, AccountingComplete: false,
	}

	applyCustomerAccounting(&view, model, time.Now().UTC())

	if view.UsageCapability.Available || view.UsageCapability.Reason != "accounting_incomplete" {
		t.Fatalf("incomplete accounting capability = %#v", view.UsageCapability)
	}
	if view.Usage == nil || view.Usage.AccountingComplete || view.Usage.DirectUsedBytes != 100 {
		t.Fatalf("incomplete projection metadata was lost: %#v", view.Usage)
	}
	if view.Usage.UsedBytes != nil || view.Usage.RemainingBytes != nil {
		t.Fatalf("incomplete total must stay unavailable: %#v", view.Usage)
	}
	if view.StatusDimensions.Presence != customer.PresenceUnknown {
		t.Fatalf("incomplete accounting must keep presence unknown, got %s", view.StatusDimensions.Presence)
	}
}
