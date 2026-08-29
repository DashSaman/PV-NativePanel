package httpapi

import (
	"testing"
	"time"

	"github.com/DashSaman/PV-NaivePanel/internal/customer"
	"github.com/DashSaman/PV-NaivePanel/internal/telemetry"
)

func TestApplyCustomerAccountingPublishesExactUsageAndPresence(t *testing.T) {
	quota := int64(1000)
	remaining := int64(150)
	lastOnline := time.Date(2026, 8, 29, 20, 0, 0, 0, time.UTC)
	view := customer.CustomerView{
		UserID: "user-1",
		QuotaBytes: &quota,
		Status: customer.UserActive,
		ServiceState: customer.TermActive,
		StartPolicy: customer.StartOnCreation,
	}
	model := telemetry.ReadModel{
		ServiceTermID: "11111111-1111-4111-8111-111111111111",
		UploadBytes: 300,
		DownloadBytes: 550,
		UsedBytes: 850,
		QuotaBytes: &quota,
		RemainingBytes: &remaining,
		QuotaState: telemetry.QuotaActive,
		LastOnline: &lastOnline,
		Online: true,
		SessionCount: 2,
		AccountingComplete: true,
	}

	applyCustomerAccounting(&view, model, time.Date(2026, 8, 29, 20, 0, 30, 0, time.UTC))

	if !view.UsageCapability.Available || view.Usage == nil {
		t.Fatal("exact accounting was not published as available")
	}
	if view.Usage.UsedBytes != 850 || view.Usage.RemainingBytes == nil || *view.Usage.RemainingBytes != 150 {
		t.Fatalf("unexpected usage snapshot: %#v", view.Usage)
	}
	if !view.Usage.Online || view.Usage.SessionCount != 2 || view.StatusDimensions.Presence != customer.PresenceOnline {
		t.Fatalf("presence not derived from WS1 read model: %#v", view.StatusDimensions)
	}
	if view.StatusDimensions.Quota != customer.QuotaWarning {
		t.Fatalf("quota status=%s want=%s", view.StatusDimensions.Quota, customer.QuotaWarning)
	}
}

func TestApplyCustomerAccountingKeepsIncompleteProjectionCapabilityGated(t *testing.T) {
	quota := int64(1000)
	view := customer.CustomerView{QuotaBytes: &quota, Status: customer.UserActive, ServiceState: customer.TermActive, StartPolicy: customer.StartOnCreation}
	model := telemetry.ReadModel{ServiceTermID: "11111111-1111-4111-8111-111111111111", UsedBytes: 100, QuotaBytes: &quota, AccountingComplete: false}

	applyCustomerAccounting(&view, model, time.Now().UTC())

	if view.UsageCapability.Available {
		t.Fatal("incomplete accounting must not be advertised as exact")
	}
	if view.Usage == nil || view.Usage.AccountingComplete {
		t.Fatalf("incomplete projection metadata was lost: %#v", view.Usage)
	}
	if view.StatusDimensions.Presence != customer.PresenceUnknown {
		t.Fatalf("incomplete accounting must keep presence unknown, got %s", view.StatusDimensions.Presence)
	}
}
