package customer

import (
	"testing"
	"time"
)

func task14Int(v int) *int { return &v }

func TestPlanPresetConcurrencyLimitValidationAndSnapshot(t *testing.T) {
	limited := PlanPreset{
		Name: "limited", QuotaBytes: nil, ValiditySeconds: 3600,
		StartPolicy: StartOnCreation, ResetStrategy: ResetNone, Enabled: true,
		ConcurrencyLimit: task14Int(2),
	}
	if err := limited.Validate(); err != nil {
		t.Fatalf("positive concurrency limit should validate: %v", err)
	}
	snapshot := limited.ServiceSnapshot(time.Unix(1_700_000_000, 0).UTC())
	if snapshot.ConcurrencyLimit == nil || *snapshot.ConcurrencyLimit != 2 {
		t.Fatalf("service snapshot lost concurrency limit: %#v", snapshot.ConcurrencyLimit)
	}

	unlimited := limited
	unlimited.Name = "unlimited"
	unlimited.ConcurrencyLimit = nil
	if err := unlimited.Validate(); err != nil {
		t.Fatalf("nil concurrency limit means Unlimited and should validate: %v", err)
	}

	invalid := limited
	invalid.Name = "invalid"
	invalid.ConcurrencyLimit = task14Int(0)
	if err := invalid.Validate(); err == nil {
		t.Fatal("zero concurrency limit must be rejected; use nil for Unlimited")
	}
}

func TestCreateServiceTermRecordCarriesConcurrencyLimit(t *testing.T) {
	record := CreateServiceTermRecord{ConcurrencyLimit: task14Int(3)}
	if record.ConcurrencyLimit == nil || *record.ConcurrencyLimit != 3 {
		t.Fatalf("term record lost concurrency limit: %#v", record.ConcurrencyLimit)
	}
}
