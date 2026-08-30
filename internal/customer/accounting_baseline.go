package customer

import (
	"errors"
	"math"
	"time"
)

type AccountingBaselineState string

const (
	AccountingBaselineKnown   AccountingBaselineState = "known"
	AccountingBaselineUnknown AccountingBaselineState = "unknown"
)

type AccountingBaselineSource string

const (
	AccountingBaselineFreshManagedTerm    AccountingBaselineSource = "fresh_managed_term"
	AccountingBaselineLegacyUnavailable   AccountingBaselineSource = "legacy_unavailable"
	AccountingBaselineAuthoritativeImport AccountingBaselineSource = "authoritative_import"
)

var ErrInvalidAccountingBaseline = errors.New("customer: invalid accounting baseline")

type AccountingBaseline struct {
	State         AccountingBaselineState  `json:"state"`
	Source        AccountingBaselineSource `json:"source"`
	CutoffAt      time.Time                `json:"cutoff_at"`
	UploadBytes   *int64                   `json:"upload_bytes"`
	DownloadBytes *int64                   `json:"download_bytes"`
}

func FreshManagedAccountingBaseline(cutoff time.Time) AccountingBaseline {
	zeroUpload := int64(0)
	zeroDownload := int64(0)
	return AccountingBaseline{
		State:         AccountingBaselineKnown,
		Source:        AccountingBaselineFreshManagedTerm,
		CutoffAt:      cutoff.UTC(),
		UploadBytes:   &zeroUpload,
		DownloadBytes: &zeroDownload,
	}
}

func UnknownLegacyAccountingBaseline(cutoff time.Time) AccountingBaseline {
	return AccountingBaseline{
		State:    AccountingBaselineUnknown,
		Source:   AccountingBaselineLegacyUnavailable,
		CutoffAt: cutoff.UTC(),
	}
}

func ComposeCustomerUsage(
	baseline AccountingBaseline,
	directUploadBytes, directDownloadBytes int64,
	quotaBytes *int64,
	accountingComplete bool,
) (CustomerUsage, UsageCapability, error) {
	if err := validateAccountingBaseline(baseline); err != nil {
		return CustomerUsage{}, UsageCapability{}, err
	}
	if directUploadBytes < 0 || directDownloadBytes < 0 || directUploadBytes > math.MaxInt64-directDownloadBytes {
		return CustomerUsage{}, UsageCapability{}, ErrInvalidAccountingBaseline
	}
	directUsed := directUploadBytes + directDownloadBytes
	usage := CustomerUsage{
		Available:           false,
		AccountingComplete:  accountingComplete,
		Baseline:            cloneAccountingBaseline(baseline),
		DirectUploadBytes:   directUploadBytes,
		DirectDownloadBytes: directDownloadBytes,
		DirectUsedBytes:     directUsed,
	}
	if !accountingComplete {
		return usage, UsageCapability{Available: false, Reason: "accounting_incomplete"}, nil
	}
	if baseline.State == AccountingBaselineUnknown {
		return usage, UsageCapability{Available: false, Reason: "historical_baseline_unknown"}, nil
	}
	if *baseline.UploadBytes > math.MaxInt64-directUploadBytes || *baseline.DownloadBytes > math.MaxInt64-directDownloadBytes {
		return CustomerUsage{}, UsageCapability{}, ErrInvalidAccountingBaseline
	}
	totalUpload := *baseline.UploadBytes + directUploadBytes
	totalDownload := *baseline.DownloadBytes + directDownloadBytes
	if totalUpload > math.MaxInt64-totalDownload {
		return CustomerUsage{}, UsageCapability{}, ErrInvalidAccountingBaseline
	}
	totalUsed := totalUpload + totalDownload
	usage.UploadBytes = &totalUpload
	usage.DownloadBytes = &totalDownload
	usage.UsedBytes = &totalUsed
	if quotaBytes != nil {
		if *quotaBytes <= 0 {
			return CustomerUsage{}, UsageCapability{}, ErrInvalidAccountingBaseline
		}
		remaining := *quotaBytes - totalUsed
		if remaining < 0 {
			remaining = 0
		}
		usage.RemainingBytes = &remaining
	}
	usage.Available = true
	return usage, UsageCapability{Available: true}, nil
}

func ComposeCustomerUsageForPeriod(
	baseline AccountingBaseline,
	lastResetAt *time.Time,
	directUploadBytes, directDownloadBytes int64,
	quotaBytes *int64,
	accountingComplete bool,
) (CustomerUsage, UsageCapability, error) {
	if lastResetAt == nil {
		return ComposeCustomerUsage(baseline, directUploadBytes, directDownloadBytes, quotaBytes, accountingComplete)
	}
	if err := validateAccountingBaseline(baseline); err != nil {
		return CustomerUsage{}, UsageCapability{}, err
	}
	resetAt := lastResetAt.UTC()
	if resetAt.IsZero() || resetAt.Before(baseline.CutoffAt.UTC()) || directUploadBytes < 0 || directDownloadBytes < 0 || directUploadBytes > math.MaxInt64-directDownloadBytes {
		return CustomerUsage{}, UsageCapability{}, ErrInvalidAccountingBaseline
	}
	directUsed := directUploadBytes + directDownloadBytes
	usage := CustomerUsage{
		Available:            false,
		AccountingComplete:   accountingComplete,
		Baseline:             cloneAccountingBaseline(baseline),
		DirectUploadBytes:    directUploadBytes,
		DirectDownloadBytes:  directDownloadBytes,
		DirectUsedBytes:      directUsed,
		UsagePeriodStartedAt: &resetAt,
		UsageResetApplied:    true,
	}
	if !accountingComplete {
		return usage, UsageCapability{Available: false, Reason: "accounting_incomplete"}, nil
	}
	upload := directUploadBytes
	download := directDownloadBytes
	used := directUsed
	usage.UploadBytes = &upload
	usage.DownloadBytes = &download
	usage.UsedBytes = &used
	if quotaBytes != nil {
		if *quotaBytes <= 0 {
			return CustomerUsage{}, UsageCapability{}, ErrInvalidAccountingBaseline
		}
		remaining := *quotaBytes - used
		if remaining < 0 {
			remaining = 0
		}
		usage.RemainingBytes = &remaining
	}
	usage.Available = true
	return usage, UsageCapability{Available: true, Reason: "usage_reset_epoch"}, nil
}

func validateAccountingBaseline(baseline AccountingBaseline) error {
	if baseline.CutoffAt.IsZero() {
		return ErrInvalidAccountingBaseline
	}
	switch baseline.State {
	case AccountingBaselineUnknown:
		if baseline.Source != AccountingBaselineLegacyUnavailable || baseline.UploadBytes != nil || baseline.DownloadBytes != nil {
			return ErrInvalidAccountingBaseline
		}
	case AccountingBaselineKnown:
		if baseline.Source != AccountingBaselineFreshManagedTerm && baseline.Source != AccountingBaselineAuthoritativeImport {
			return ErrInvalidAccountingBaseline
		}
		if baseline.UploadBytes == nil || baseline.DownloadBytes == nil || *baseline.UploadBytes < 0 || *baseline.DownloadBytes < 0 {
			return ErrInvalidAccountingBaseline
		}
		if baseline.Source == AccountingBaselineFreshManagedTerm && (*baseline.UploadBytes != 0 || *baseline.DownloadBytes != 0) {
			return ErrInvalidAccountingBaseline
		}
		if *baseline.UploadBytes > math.MaxInt64-*baseline.DownloadBytes {
			return ErrInvalidAccountingBaseline
		}
	default:
		return ErrInvalidAccountingBaseline
	}
	return nil
}

func cloneAccountingBaseline(value AccountingBaseline) AccountingBaseline {
	value.CutoffAt = value.CutoffAt.UTC()
	if value.UploadBytes != nil {
		copy := *value.UploadBytes
		value.UploadBytes = &copy
	}
	if value.DownloadBytes != nil {
		copy := *value.DownloadBytes
		value.DownloadBytes = &copy
	}
	return value
}
