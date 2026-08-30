package ops

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestDiagnosticBundleScriptContract(t *testing.T) {
	path := filepath.Clean(filepath.Join("..", "..", "scripts", "ops", "diagnostic-bundle.sh"))
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("diagnostic bundle script missing: %v", err)
	}
	if out, err := exec.Command("bash", "-n", path).CombinedOutput(); err != nil {
		t.Fatalf("bash -n failed: %v: %s", err, out)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	for _, want := range []string{"pvnaive doctor --json", "pvnaive-telemetry-agent.service", "accounting.sock", "[PRESENT]", "[REDACTED]"} {
		if !strings.Contains(text, want) {
			t.Fatalf("diagnostic bundle contract missing %q", want)
		}
	}
	if strings.Contains(text, "cat /etc/pvnaive/api.env") || strings.Contains(text, "cat /etc/pvnaive/db.env") {
		t.Fatal("diagnostic bundle must not dump raw environment files")
	}
}
