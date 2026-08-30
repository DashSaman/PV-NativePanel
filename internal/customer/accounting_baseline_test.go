package customer

import (
	"math"
	"testing"
	"time"
)

func int64Ptr(value int64) *int64 { return &value }

func TestComposeCustomerUsageUnknownBaselineNeverFabricatesHistoricalTotal(t *testing.T) {
	cutoff := time.Date(2026, 8, 30, 0, 0, 0, 0, time.UTC)
	quota := int64(1000)
	baseline := AccountingBaseline{
		State:    AccountingBaselineUnknown,
		Source:   AccountingBaselineLegacyUnavailable,
		CutoffAt: cutoff,
	}

	usage, capability, err := ComposeCustomerUsage(baseline, 120, 230, &quota, true)
	if err != nil {
		t.Fatalf("ComposeCustomerUsage() error = %v", err)
	}
	if usage.DirectUploadBytes != 120 || usage.DirectDownloadBytes != 230 || usage.DirectUsedBytes != 350 {
		t.Fatalf("direct exact epoch lost: %#v", usage)
	}
	if usage.UploadBytes != nil || usage.DownloadBytes != nil || usage.UsedBytes != nil || usage.RemainingBytes != nil {
		t.Fatalf("unknown historical baseline fabricated a total: %#v", usage)
	}
	if usage.Baseline.State != AccountingBaselineUnknown || usage.Baseline.CutoffAt != cutoff {
		t.Fatalf("baseline truth lost: %#v", usage.Baseline)
	}
	if capability.Available || capability.Reason != "historical_baseline_unknown" {
		t.Fatalf("capability = %#v", capability)
	}
}

func TestComposeCustomerUsageKnownBaselineBuildsTotalWithoutDoubleCounting(t *testing.T) {
	cutoff := time.Date(2026, 8, 30, 0, 0, 0, 0, time.UTC)
	quota := int64(1000)
	baseline := AccountingBaseline{
		State:         AccountingBaselineKnown,
		Source:        AccountingBaselineAuthoritativeImport,
		CutoffAt:      cutoff,
		UploadBytes:   int64Ptr(200),
		DownloadBytes: int64Ptr(300),
	}

	usage, capability, err := ComposeCustomerUsage(baseline, 120, 230, &quota, true)
	if err != nil {
		t.Fatalf("ComposeCustomerUsage() error = %v", err)
	}
	if usage.UploadBytes == nil || *usage.UploadBytes != 320 || usage.DownloadBytes == nil || *usage.DownloadBytes != 530 || usage.UsedBytes == nil || *usage.UsedBytes != 850 {
		t.Fatalf("known total = %#v", usage)
	}
	if usage.RemainingBytes == nil || *usage.RemainingBytes != 150 {
		t.Fatalf("remaining = %v", usage.RemainingBytes)
	}
	if !capability.Available || capability.Reason != "" {
		t.Fatalf("capability = %#v", capability)
	}
}

func TestComposeCustomerUsageKnownZeroBaselineMakesFreshManagedTermExact(t *testing.T) {
	cutoff := time.Date(2026, 8, 30, 0, 0, 0, 0, time.UTC)
	baseline := AccountingBaseline{
		State:         AccountingBaselineKnown,
		Source:        AccountingBaselineFreshManagedTerm,
		CutoffAt:      cutoff,
		UploadBytes:   int64Ptr(0),
		DownloadBytes: int64Ptr(0),
	}

	usage, capability, err := ComposeCustomerUsage(baseline, 7, 11, nil, true)
	if err != nil {
		t.Fatalf("ComposeCustomerUsage() error = %v", err)
	}
	if usage.UsedBytes == nil || *usage.UsedBytes != 18 || usage.RemainingBytes != nil || !capability.Available {
		t.Fatalf("fresh managed term usage = %#v capability=%#v", usage, capability)
	}
}

func TestComposeCustomerUsageRejectsInvalidOrOverflowingBaseline(t *testing.T) {
	cutoff := time.Date(2026, 8, 30, 0, 0, 0, 0, time.UTC)
	cases := []AccountingBaseline{
		{State: AccountingBaselineKnown, Source: AccountingBaselineFreshManagedTerm, CutoffAt: cutoff},
		{State: AccountingBaselineUnknown, Source: AccountingBaselineLegacyUnavailable, CutoffAt: cutoff, UploadBytes: int64Ptr(1)},
		{State: AccountingBaselineKnown, Source: AccountingBaselineAuthoritativeImport, CutoffAt: cutoff, UploadBytes: int64Ptr(math.MaxInt64), DownloadBytes: int64Ptr(0)},
	}
	for i, baseline := range cases {
		directUpload := int64(0)
		if i == 2 {
			directUpload = 1
		}
		if _, _, err := ComposeCustomerUsage(baseline, directUpload, 0, nil, true); err == nil {
			t.Fatalf("case %d accepted invalid/overflowing baseline %#v", i, baseline)
		}
	}
}

func TestComposeCustomerUsageForPeriodMakesLegacyUsageExactAfterManualReset(t *testing.T) {
	cutoff := time.Date(2026, 8, 29, 12, 0, 0, 0, time.UTC)
	resetAt := time.Date(2026, 8, 30, 12, 0, 0, 0, time.UTC)
	quota := int64(1000)
	baseline := AccountingBaseline{State: AccountingBaselineUnknown, Source: AccountingBaselineLegacyUnavailable, CutoffAt: cutoff}

	usage, capability, err := ComposeCustomerUsageForPeriod(baseline, &resetAt, 120, 230, &quota, true)
	if err != nil {
		t.Fatalf("ComposeCustomerUsageForPeriod() error = %v", err)
	}
	if !capability.Available || capability.Reason != "usage_reset_epoch" {
		t.Fatalf("capability = %#v", capability)
	}
	if usage.UsedBytes == nil || *usage.UsedBytes != 350 || usage.RemainingBytes == nil || *usage.RemainingBytes != 650 {
		t.Fatalf("reset-period totals = %#v", usage)
	}
	if usage.UploadBytes == nil || *usage.UploadBytes != 120 || usage.DownloadBytes == nil || *usage.DownloadBytes != 230 {
		t.Fatalf("reset-period directional totals = %#v", usage)
	}
	if !usage.UsageResetApplied || usage.UsagePeriodStartedAt == nil || !usage.UsagePeriodStartedAt.Equal(resetAt) {
		t.Fatalf("reset epoch metadata = %#v", usage)
	}
	if usage.Baseline.State != AccountingBaselineUnknown || usage.Baseline.Source != AccountingBaselineLegacyUnavailable {
		t.Fatalf("immutable adoption baseline was rewritten: %#v", usage.Baseline)
	}
}

func TestComposeCustomerUsageForPeriodStillGatesIncompleteResetEpoch(t *testing.T) {
	cutoff := time.Date(2026, 8, 29, 12, 0, 0, 0, time.UTC)
	resetAt := time.Date(2026, 8, 30, 12, 0, 0, 0, time.UTC)
	baseline := AccountingBaseline{State: AccountingBaselineUnknown, Source: AccountingBaselineLegacyUnavailable, CutoffAt: cutoff}
	usage, capability, err := ComposeCustomerUsageForPeriod(baseline, &resetAt, 4, 6, nil, false)
	if err != nil {
		t.Fatalf("ComposeCustomerUsageForPeriod() error = %v", err)
	}
	if capability.Available || capability.Reason != "accounting_incomplete" || usage.UsedBytes != nil {
		t.Fatalf("incomplete reset epoch leaked totals: usage=%#v capability=%#v", usage, capability)
	}
}
