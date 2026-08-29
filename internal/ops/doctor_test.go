package ops

import (
	"context"
	"strings"
	"testing"
)

type fakeCheck struct {
	result CheckResult
}

func (f fakeCheck) Run(context.Context) CheckResult { return f.result }

func TestDoctorSummarizesPassWarnFailWithoutSecrets(t *testing.T) {
	doctor := NewDoctor([]Check{
		fakeCheck{result: CheckResult{Name: "api", Status: Pass, Detail: "healthy"}},
		fakeCheck{result: CheckResult{Name: "backup", Status: Warn, Detail: "last backup is old"}},
		fakeCheck{result: CheckResult{Name: "postgres", Status: Fail, Detail: "password=super-secret connection failed"}},
	})
	report := doctor.Run(context.Background())
	if report.Pass != 1 || report.Warn != 1 || report.Fail != 1 {
		t.Fatalf("summary=%#v", report)
	}
	text := report.String()
	for _, want := range []string{"PASS api", "WARN backup", "FAIL postgres", "[REDACTED]"} {
		if !strings.Contains(text, want) {
			t.Fatalf("doctor output missing %q: %s", want, text)
		}
	}
	if strings.Contains(text, "super-secret") {
		t.Fatalf("doctor leaked secret: %s", text)
	}
}

func TestDiagnosticsBundleRejectsSecretBearingEntries(t *testing.T) {
	bundle := DiagnosticsBundle{
		Version: "1.0.0",
		Entries: map[string]string{
			"health.txt": "api=PASS",
			"logs.txt":   "Authorization: Bearer secret-value request_id=req-1",
		},
	}
	files := bundle.SanitizedEntries()
	if strings.Contains(files["logs.txt"], "secret-value") {
		t.Fatalf("diagnostics leaked authorization: %s", files["logs.txt"])
	}
	if !strings.Contains(files["logs.txt"], "req-1") {
		t.Fatalf("diagnostics removed useful request id: %s", files["logs.txt"])
	}
}
