package ops

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestCurrentR1ReleaseToolingContracts(t *testing.T) {
	root := filepath.Clean(filepath.Join("..", ".."))
	contracts := map[string][]string{
		filepath.Join(root, "scripts", "release", "build-r1-bundle.sh"): {
			"pvnaive-telemetry-agent",
			"db/migrations",
			"latest_schema",
			"ops/tmpfiles",
			"scripts/ops",
			"source_commit",
			"PVNAIVE_R1_BUILD_RESULT=PASSED",
		},
		filepath.Join(root, "scripts", "release", "deploy-r1.sh"): {
			"release_schema",
			"pvnaive-telemetry-agent.service",
			"accounting.sock",
			"caddy/caddy-pvnaive-accounting",
			"caddy-before",
			"PVNAIVE_R1_CADDY_ACTION=ONE_CONTROLLED_BINARY_SWAP_RESTART",
			"pvnaive-backup.timer",
			"pvnaive-restore-drill.timer",
			"/var/www/pvnaive-preview/current",
			"preview.before",
			"/opt/pvnaive/DEPLOYED_COMMIT",
			"/opt/pvnaive/DEPLOYED_WEB_RELEASE",
			"DEPLOYED_COMMIT.before",
			"DEPLOYED_WEB_RELEASE.before",
		},
		filepath.Join(root, "scripts", "release", "rollback-r1.sh"): {
			"pvnaive-telemetry-agent.before",
			"pvnaive-telemetry-agent.service.before",
			"caddy-before",
			"20-pvnaive-accounting.conf.before",
			"PVNAIVE_R1_CADDY_ACTION=RESTORED_PRIOR_BINARY_RESTART",
			"/var/www/pvnaive-preview/current",
			"preview.before",
			"/opt/pvnaive/DEPLOYED_COMMIT",
			"/opt/pvnaive/DEPLOYED_WEB_RELEASE",
			"DEPLOYED_COMMIT.before",
			"DEPLOYED_WEB_RELEASE.before",
		},
	}
	for path, wants := range contracts {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("release file %s missing: %v", path, err)
		}
		if out, err := exec.Command("bash", "-n", path).CombinedOutput(); err != nil {
			t.Fatalf("bash -n %s: %v: %s", path, err, out)
		}
		text := string(data)
		for _, want := range wants {
			if !strings.Contains(text, want) {
				t.Fatalf("%s missing %q", path, want)
			}
		}
		if strings.Contains(text, "schema_version\":8") {
			t.Fatalf("%s still hard-codes obsolete schema 8", path)
		}
	}
}

func TestLoadRehearsalIsExplicitlyBounded(t *testing.T) {
	path := filepath.Clean(filepath.Join("..", "..", "scripts", "release", "load-rehearsal.py"))
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("load rehearsal missing: %v", err)
	}
	text := string(data)
	for _, want := range []string{"control_plane_rehearsal_not_capacity_ceiling", "5000", "100", "127.0.0.1"} {
		if !strings.Contains(text, want) {
			t.Fatalf("load rehearsal missing %q", want)
		}
	}
	if out, err := exec.Command("python3", "-m", "py_compile", path).CombinedOutput(); err != nil {
		t.Fatalf("load rehearsal syntax invalid: %v: %s", path, err, out)
	}
}
