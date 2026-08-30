package customer

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func functionSource(t *testing.T, filename, marker string) string {
	t.Helper()
	_, current, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot resolve test path")
	}
	data, err := os.ReadFile(filepath.Join(filepath.Dir(current), filename))
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	start := strings.Index(text, marker)
	if start < 0 {
		t.Fatalf("%s missing %s", filename, marker)
	}
	rest := text[start+len(marker):]
	if next := strings.Index(rest, "\nfunc "); next >= 0 {
		rest = rest[:next]
	}
	return marker + rest
}

func requireBaselineColumns(t *testing.T, filename, marker string) {
	t.Helper()
	body := functionSource(t, filename, marker)
	for _, column := range []string{
		"accounting_baseline_state",
		"accounting_baseline_source",
		"accounting_baseline_cutoff_at",
		"accounting_baseline_upload_bytes",
		"accounting_baseline_download_bytes",
	} {
		if !strings.Contains(body, column) {
			t.Fatalf("%s %s does not preserve %s", filename, marker, column)
		}
	}
}

func TestServiceTermReadAndMutationReturnsPreserveAccountingBaseline(t *testing.T) {
	cases := []struct {
		file   string
		marker string
	}{
		{"store.go", "func (s *PostgresStore) ListCustomersTx"},
		{"adopt_update_store.go", "func (s *PostgresStore) UpdateCurrentServiceTermTx"},
		{"adjustments_store.go", "func (s *PostgresStore) AddCurrentServiceQuotaTx"},
		{"adjustments_store.go", "func (s *PostgresStore) ExtendCurrentServiceTx"},
		{"set_volume_store.go", "func (s *PostgresStore) SetCurrentServiceQuotaTx"},
		{"renewal_store.go", "func (s *PostgresStore) CurrentRenewalContextTx"},
	}
	for _, tc := range cases {
		t.Run(tc.file+tc.marker, func(t *testing.T) {
			requireBaselineColumns(t, tc.file, tc.marker)
		})
	}
}
