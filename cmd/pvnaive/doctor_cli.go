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

func doctorChecks() []ops.Check {
	runtimeSocket := os.Getenv("PVNAIVE_RUNTIME_AGENT_SOCKET")
	if runtimeSocket == "" {
		runtimeSocket = runtimeagent.DefaultSocketPath
	}
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
		ops.FileModeCheck("auth-key-mode", "/etc/pvnaive/auth.key", 0o600),
		ops.FileModeCheck("runtime-key-mode", "/etc/pvnaive/runtime.key", 0o600),
		ops.FileModeCheck("backup-key-mode", "/etc/pvnaive/backup.agekey", 0o600),
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
