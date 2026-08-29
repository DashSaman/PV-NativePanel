package customer

import (
	"testing"
	"time"
)

func TestPlanPresetValidationAndSnapshot(t *testing.T) {
	plan := PlanPreset{
		Name:            "100 GB / 30 days",
		QuotaBytes:      ptrInt64(100 * 1073741824),
		ValiditySeconds: 30 * 86400,
		StartPolicy:     StartOnFirstSuccessfulConnection,
		ResetStrategy:   ResetNone,
		Enabled:         true,
	}
	if err := plan.Validate(); err != nil {
		t.Fatalf("valid plan rejected: %v", err)
	}
	snapshot := plan.ServiceSnapshot(time.Date(2026, 8, 29, 12, 0, 0, 0, time.UTC))
	if snapshot.QuotaBytes == nil || *snapshot.QuotaBytes != 100*1073741824 {
		t.Fatalf("quota snapshot mismatch: %#v", snapshot.QuotaBytes)
	}
	if snapshot.DurationSeconds != 30*86400 || snapshot.StartPolicy != StartOnFirstSuccessfulConnection {
		t.Fatalf("validity snapshot mismatch: %#v", snapshot)
	}
}

func TestUnlimitedNoExpiryPlanIsExplicit(t *testing.T) {
	plan := PlanPreset{
		Name:          "Unlimited",
		QuotaBytes:    nil,
		NoExpiry:      true,
		StartPolicy:   StartOnCreation,
		ResetStrategy: ResetNone,
		Enabled:       true,
	}
	if err := plan.Validate(); err != nil {
		t.Fatalf("unlimited no-expiry plan rejected: %v", err)
	}
	if !plan.ServiceSnapshot(time.Now()).NoExpiry {
		t.Fatal("no-expiry semantic was lost")
	}
}

func TestCustomerStatusDimensionsNeverFabricateAccounting(t *testing.T) {
	now := time.Date(2026, 8, 29, 12, 0, 0, 0, time.UTC)
	status := DeriveStatusDimensions(StatusInput{
		UserState:              UserActive,
		TermState:              TermActive,
		QuotaBytes:             ptrInt64(50 * 1073741824),
		ExpiresAt:              ptrTime(now.Add(24 * time.Hour)),
		AccountingAvailable:    false,
		RuntimeHealthAvailable: false,
	})
	if status.Lifecycle != LifecycleActive || status.Commercial != CommercialActive {
		t.Fatalf("unexpected durable status: %#v", status)
	}
	if status.Presence != PresenceUnknown || status.Quota != QuotaUnavailable || status.Runtime != RuntimeUnknown {
		t.Fatalf("unproven telemetry/accounting was fabricated: %#v", status)
	}
}

func TestPendingFirstUseAndOnHoldAreDistinct(t *testing.T) {
	pending := DeriveStatusDimensions(StatusInput{
		UserState:   UserActive,
		TermState:   TermPending,
		StartPolicy: StartOnFirstSuccessfulConnection,
	})
	if pending.Commercial != CommercialPendingFirstUse {
		t.Fatalf("pending first-use lost: %#v", pending)
	}
	held := DeriveStatusDimensions(StatusInput{
		UserState: UserActive,
		TermState: TermActive,
		OnHold:    true,
	})
	if held.Commercial != CommercialOnHold {
		t.Fatalf("on-hold lost: %#v", held)
	}
}

func TestCustomerRBACMatrix(t *testing.T) {
	cases := []struct {
		role   string
		action CustomerAction
		want   bool
	}{
		{"owner", ActionDeleteCustomer, true},
		{"admin", ActionViewCustomers, true},
		{"admin", ActionRenewCustomer, true},
		{"admin", ActionDeleteCustomer, false},
		{"reseller", ActionViewCustomers, true},
		{"reseller", ActionCreateCustomer, true},
		{"reseller", ActionManagePlans, false},
		{"auditor", ActionViewAudit, true},
		{"auditor", ActionEditCustomer, false},
		{"operator", ActionViewCustomers, false},
	}
	for _, tc := range cases {
		if got := CustomerActionAllowed(tc.role, tc.action); got != tc.want {
			t.Fatalf("role=%s action=%s got=%v want=%v", tc.role, tc.action, got, tc.want)
		}
	}
}

func TestListQueryNormalization(t *testing.T) {
	q := CustomerListQuery{Page: -10, PageSize: 1000, Sort: "unknown", Direction: "sideways"}.Normalize()
	if q.Page != 1 || q.PageSize != 50 || q.Sort != SortUpdated || q.Direction != SortDescending {
		t.Fatalf("unsafe query normalization: %#v", q)
	}
}

func TestBulkPreviewSeparatesInvalidSkippedAndRuntimeConflicts(t *testing.T) {
	preview := BuildBulkPreview(BulkPreviewInput{
		Action:                      BulkSuspend,
		RequestedIDs:                []string{"a", "b", "missing"},
		Customers:                   []BulkCustomer{{ID: "a", Lifecycle: LifecycleActive}, {ID: "b", Lifecycle: LifecycleRevoked}},
		RuntimeCoordinatorAvailable: false,
	})
	if preview.Affected != 0 || len(preview.Conflicts) != 1 || len(preview.Skipped) != 1 || len(preview.Invalid) != 1 {
		t.Fatalf("unsafe runtime bulk preview: %#v", preview)
	}
}

func ptrInt64(v int64) *int64        { return &v }
func ptrTime(v time.Time) *time.Time { return &v }
