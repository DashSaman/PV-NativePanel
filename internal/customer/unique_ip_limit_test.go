package customer

import (
	"testing"
	"time"
)

func task15Int(v int) *int { return &v }

func TestPlanPresetUniqueIPLimitValidationAndSnapshot(t *testing.T) {
	limited := PlanPreset{
		Name: "ip-limited", QuotaBytes: nil, ValiditySeconds: 3600,
		StartPolicy: StartOnCreation, ResetStrategy: ResetNone, Enabled: true,
		UniqueIPLimit: task15Int(3),
	}
	if err := limited.Validate(); err != nil {
		t.Fatalf("positive unique IP limit should validate: %v", err)
	}
	snapshot := limited.ServiceSnapshot(time.Unix(1_700_000_000, 0).UTC())
	if snapshot.UniqueIPLimit == nil || *snapshot.UniqueIPLimit != 3 {
		t.Fatalf("service snapshot lost unique IP limit: %#v", snapshot.UniqueIPLimit)
	}

	unlimited := limited
	unlimited.Name = "ip-unlimited"
	unlimited.UniqueIPLimit = nil
	if err := unlimited.Validate(); err != nil {
		t.Fatalf("nil unique IP limit means Unlimited and should validate: %v", err)
	}

	invalid := limited
	invalid.Name = "ip-invalid"
	invalid.UniqueIPLimit = task15Int(0)
	if err := invalid.Validate(); err == nil {
		t.Fatal("zero unique IP limit must be rejected; use nil for Unlimited")
	}
}

func TestCreateServiceTermRecordCarriesUniqueIPLimit(t *testing.T) {
	record := CreateServiceTermRecord{UniqueIPLimit: task15Int(5)}
	if record.UniqueIPLimit == nil || *record.UniqueIPLimit != 5 {
		t.Fatalf("term record lost unique IP limit: %#v", record.UniqueIPLimit)
	}
}
