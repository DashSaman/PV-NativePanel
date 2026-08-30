package ops

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestScheduledBackupAndRestoreContracts(t *testing.T) {
	root := filepath.Clean(filepath.Join("..", ".."))
	files := map[string][]string{
		filepath.Join(root, "scripts", "ops", "backup-run.sh"): {
			"/var/backups/pvnaive/scheduled",
			"/var/backups/pvnaive/database",
			"pvnaive-telemetry-agent.service",
			"pvnaive-runtime.conf",
			"config.tar.age",
			"PVNAIVE_SCHEDULED_BACKUP_RESULT=PASSED",
		},
		filepath.Join(root, "scripts", "ops", "restore-drill.sh"): {
			"PVNAIVE_RESTORE_TARGET_DB",
			"dropdb --if-exists --force",
			"PVNAIVE_RESTORE_DRILL_RESULT=PASSED",
		},
		filepath.Join(root, "scripts", "db", "restore.sh"): {
			"pvnaive_validate_encrypted_archive",
			"PVNAIVE_RESTORE_RESULT=PASSED",
		},
		filepath.Join(root, "ops", "systemd", "pvnaive-backup.timer"): {
			"OnCalendar=",
			"Persistent=true",
		},
		filepath.Join(root, "ops", "systemd", "pvnaive-restore-drill.timer"): {
			"OnCalendar=",
			"Persistent=true",
		},
	}
	for path, wants := range files {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("required ops file %s missing: %v", path, err)
		}
		text := string(data)
		for _, want := range wants {
			if !strings.Contains(text, want) {
				t.Fatalf("%s missing %q", path, want)
			}
		}
		if strings.HasSuffix(path, ".sh") {
			if out, err := exec.Command("bash", "-n", path).CombinedOutput(); err != nil {
				t.Fatalf("bash -n %s: %v: %s", path, err, out)
			}
		}
	}
}
