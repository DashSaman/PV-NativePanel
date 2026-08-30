package customer

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestCustomerReadQueriesProjectAccountingBaseline(t *testing.T) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot resolve test source path")
	}
	dir := filepath.Dir(filename)
	for _, name := range []string{"store.go", "product_list_store.go"} {
		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatal(err)
		}
		text := string(data)
		for _, column := range []string{
			"accounting_baseline_state",
			"accounting_baseline_source",
			"accounting_baseline_cutoff_at",
			"accounting_baseline_upload_bytes",
			"accounting_baseline_download_bytes",
		} {
			if !strings.Contains(text, column) {
				t.Fatalf("%s does not project %s", name, column)
			}
		}
	}
}
