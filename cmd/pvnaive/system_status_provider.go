package main

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/DashSaman/PV-NaivePanel/internal/observability"
	"github.com/DashSaman/PV-NaivePanel/internal/runtimeagent"
	"github.com/DashSaman/PV-NaivePanel/internal/telemetry"
)

func buildSystemStatusProvider(db *sql.DB, getenv func(string) string) func(*http.Request) (any, error) {
	collector := observability.NewCollector(observability.NewLinuxSource())
	runtimeSocket := getenv("PVNAIVE_RUNTIME_AGENT_SOCKET")
	if runtimeSocket == "" {
		runtimeSocket = defaultRuntimeAgentSocket
	}
	runtimeClient := runtimeagent.NewClient(runtimeSocket)

	return func(r *http.Request) (any, error) {
		snapshot, err := collector.Collect(r.Context())
		if err != nil {
			return nil, err
		}
		dependencies := map[string]any{"api": map[string]any{"status": "ok"}}

		dbCtx, dbCancel := context.WithTimeout(r.Context(), 1500*time.Millisecond)
		dbErr := db.PingContext(dbCtx)
		dbCancel()
		dependencies["database"] = dependencyStatus(dbErr)

		runtimeCtx, runtimeCancel := context.WithTimeout(r.Context(), 1500*time.Millisecond)
		_, runtimeErr := runtimeClient.Health(runtimeCtx)
		runtimeCancel()
		dependencies["runtime"] = dependencyStatus(runtimeErr)

		telemetryCtx, telemetryCancel := context.WithTimeout(r.Context(), 1500*time.Millisecond)
		telemetryErr := telemetryHealth(telemetryCtx, telemetry.DefaultTelemetrySocketPath)
		telemetryCancel()
		dependencies["telemetry"] = dependencyStatus(telemetryErr)

		return map[string]any{
			"sample":       snapshot,
			"dependencies": dependencies,
		}, nil
	}
}

func dependencyStatus(err error) map[string]any {
	if err == nil {
		return map[string]any{"status": "ok"}
	}
	return map[string]any{"status": "unavailable"}
}

func telemetryHealth(ctx context.Context, socketPath string) error {
	transport := &http.Transport{
		DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
			var dialer net.Dialer
			return dialer.DialContext(ctx, "unix", socketPath)
		},
	}
	defer transport.CloseIdleConnections()
	client := &http.Client{Transport: transport}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://unix"+telemetry.TelemetryHealthPath, nil)
	if err != nil {
		return err
	}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("telemetry health returned HTTP %d", res.StatusCode)
	}
	return nil
}
