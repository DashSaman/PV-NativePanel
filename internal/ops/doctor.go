package ops

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/DashSaman/PV-NaivePanel/internal/observability"
)

type Status string

const (
	Pass Status = "PASS"
	Warn Status = "WARN"
	Fail Status = "FAIL"
)

type CheckResult struct {
	Name   string `json:"name"`
	Status Status `json:"status"`
	Detail string `json:"detail,omitempty"`
}

type Check interface {
	Run(context.Context) CheckResult
}

type Doctor struct {
	checks []Check
}

type Report struct {
	Results []CheckResult `json:"results"`
	Pass    int           `json:"pass"`
	Warn    int           `json:"warn"`
	Fail    int           `json:"fail"`
}

func NewDoctor(checks []Check) *Doctor {
	return &Doctor{checks: append([]Check(nil), checks...)}
}

func (d *Doctor) Run(ctx context.Context) Report {
	var report Report
	if d == nil {
		return report
	}
	for _, check := range d.checks {
		if check == nil {
			continue
		}
		result := check.Run(ctx)
		result.Name = observability.RedactText(strings.TrimSpace(result.Name))
		result.Detail = observability.RedactText(strings.TrimSpace(result.Detail))
		switch result.Status {
		case Pass:
			report.Pass++
		case Warn:
			report.Warn++
		default:
			result.Status = Fail
			report.Fail++
		}
		report.Results = append(report.Results, result)
	}
	return report
}

func (r Report) String() string {
	var builder strings.Builder
	for _, result := range r.Results {
		fmt.Fprintf(&builder, "%s %s", result.Status, result.Name)
		if result.Detail != "" {
			fmt.Fprintf(&builder, " — %s", result.Detail)
		}
		builder.WriteByte('\n')
	}
	fmt.Fprintf(&builder, "SUMMARY PASS=%d WARN=%d FAIL=%d\n", r.Pass, r.Warn, r.Fail)
	return observability.RedactText(builder.String())
}

type DiagnosticsBundle struct {
	Version string
	Entries map[string]string
}

func (b DiagnosticsBundle) SanitizedEntries() map[string]string {
	result := make(map[string]string, len(b.Entries))
	keys := make([]string, 0, len(b.Entries))
	for key := range b.Entries {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		cleanKey := strings.TrimSpace(key)
		if cleanKey == "" || strings.Contains(cleanKey, "..") || strings.ContainsAny(cleanKey, "/\\") {
			continue
		}
		result[cleanKey] = observability.RedactText(b.Entries[key])
	}
	return result
}
