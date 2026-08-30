package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/DashSaman/PV-NaivePanel/internal/ops"
	"github.com/DashSaman/PV-NaivePanel/internal/runtimeagent"
	"github.com/DashSaman/PV-NaivePanel/internal/telemetry"
)

func doctorCheckNames() []string {
	return []string{
		"pvnaive-api.service",
		"pvnaive-runtime-agent.service",
		"pvnaive-telemetry-agent.service",
		"caddy-naive.service",
		"postgresql.service",
		"runtime-socket",
		"accounting-socket",
		"telemetry-health",
		"api-ready",
		"api-loopback",
		"disk",
		"backup",
		"auth-key-mode",
		"runtime-key-mode",
		"backup-key-mode",
	}
}

func doctorKeyModes() map[string]os.FileMode {
	return map[string]os.FileMode{
		// The API runs as pvnaive and must be able to read these root-owned
		// encryption keys through the pvnaive group. Root-only 0600 would make
		// a healthy Production API unreadable after restart.
		"auth-key-mode":    0o640,
		"runtime-key-mode": 0o640,
		// The age identity is used only by privileged backup/restore tooling.
		"backup-key-mode": 0o600,
	}
}

func doctorChecks() []ops.Check {
	runtimeSocket := os.Getenv("PVNAIVE_RUNTIME_AGENT_SOCKET")
	if runtimeSocket == "" {
		runtimeSocket = runtimeagent.DefaultSocketPath
	}
	keyModes := doctorKeyModes()
	return []ops.Check{
		ops.ServiceCheck("pvnaive-api.service"),
		ops.ServiceCheck("pvnaive-runtime-agent.service"),
		ops.ServiceCheck("pvnaive-telemetry-agent.service"),
		ops.ServiceCheck("caddy-naive.service"),
		ops.ServiceCheck("postgresql.service"),
		ops.UnixSocketCheck("runtime-socket", runtimeSocket),
		ops.UnixSocketCheck("accounting-socket", telemetry.DefaultTelemetrySocketPath),
		ops.UnixHTTPCheck("telemetry-health", telemetry.DefaultTelemetrySocketPath, telemetry.TelemetryHealthPath),
		ops.HTTPCheck("api-ready", "http://127.0.0.1:8080/api/v1/health/ready"),
		ops.PortCheck("api-loopback", "127.0.0.1:8080", true),
		ops.DiskCheck("/", 80, 90),
		ops.BackupCheck("/var/backups/pvnaive/database", 26*time.Hour),
		ops.FileModeCheck("auth-key-mode", "/etc/pvnaive/auth.key", keyModes["auth-key-mode"]),
		ops.FileModeCheck("runtime-key-mode", "/etc/pvnaive/runtime.key", keyModes["runtime-key-mode"]),
		ops.FileModeCheck("backup-key-mode", "/etc/pvnaive/backup.agekey", keyModes["backup-key-mode"]),
	}
}

func runDoctor(args []string) error {
	jsonOutput := false
	for _, arg := range args {
		switch arg {
		case "--json":
			jsonOutput = true
		case "--help", "-h":
			fmt.Fprintln(os.Stdout, "usage: pvnaive doctor [--json]")
			return nil
		default:
			return fmt.Errorf("unknown doctor argument %q", arg)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	report := ops.NewDoctor(doctorChecks()).Run(ctx)
	if jsonOutput {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(report); err != nil {
			return fmt.Errorf("encode doctor report: %w", err)
		}
	} else {
		fmt.Fprint(os.Stdout, report.String())
	}
	if report.Fail > 0 {
		return errors.New("doctor found failing checks")
	}
	return nil
}
