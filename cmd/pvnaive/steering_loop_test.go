package main

import (
	"strings"
	"testing"
	"time"

	"github.com/DashSaman/PV-NaivePanel/internal/telemetry"
)

func TestSteeringLoopConfigDefaults(t *testing.T) {
	cfg, err := steeringLoopConfigFromEnv(func(string) string { return "" })
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Engine.Window != 4*time.Hour {
		t.Fatalf("default window = %s", cfg.Engine.Window)
	}
	if cfg.Engine.Grace != 15*time.Minute {
		t.Fatalf("default grace = %s", cfg.Engine.Grace)
	}
	if cfg.Engine.CandidateRatio != 0.85 || cfg.Engine.HysteresisMargin != 0.25 || cfg.Engine.HysteresisWindows != 2 {
		t.Fatalf("defaults wrong: %+v", cfg.Engine)
	}
	if cfg.Engine.MinSamples != 8 {
		t.Fatalf("default min samples = %d", cfg.Engine.MinSamples)
	}
	if cfg.Tick != 60*time.Second || cfg.ReadBudget != 10*time.Second {
		t.Fatalf("loop timing wrong: tick=%s read=%s", cfg.Tick, cfg.ReadBudget)
	}
	if err := cfg.Engine.Validate(); err != nil {
		t.Fatalf("defaults must satisfy engine validation: %v", err)
	}
}

func TestSteeringLoopConfigEnvOverridesAndClamps(t *testing.T) {
	env := map[string]string{
		"PVNAIVE_STEER_ROTATION_WINDOW_SECS": "7200",
		"PVNAIVE_STEER_GRACE_SECONDS":        "600",
		"PVNAIVE_STEER_CANDIDATE_THRESHOLD":  "0.9",
		"PVNAIVE_STEER_HYSTERESIS_MARGIN":    "0.3",
		"PVNAIVE_STEER_HYSTERESIS_WINDOWS":   "3",
		"PVNAIVE_STEER_MIN_SAMPLES":          "16",
		"PVNAIVE_STEER_TICK_SECS":            "120",
	}
	cfg, err := steeringLoopConfigFromEnv(func(key string) string { return env[key] })
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Engine.Window != 2*time.Hour || cfg.Engine.Grace != 10*time.Minute {
		t.Fatalf("window/grace override wrong: %s %s", cfg.Engine.Window, cfg.Engine.Grace)
	}
	if cfg.Engine.CandidateRatio != 0.9 || cfg.Engine.HysteresisMargin != 0.3 || cfg.Engine.HysteresisWindows != 3 {
		t.Fatalf("engine overrides wrong: %+v", cfg.Engine)
	}
	if cfg.Engine.MinSamples != 16 || cfg.Tick != 2*time.Minute {
		t.Fatalf("loop overrides wrong: min=%d tick=%s", cfg.Engine.MinSamples, cfg.Tick)
	}
}

func TestSteeringLoopConfigRejectsOutOfRangeValuesLoudly(t *testing.T) {
	cases := map[string]string{
		"PVNAIVE_STEER_ROTATION_WINDOW_SECS": "60",   // below 30m
		"PVNAIVE_STEER_GRACE_SECONDS":        "1",    // below 5m
		"PVNAIVE_STEER_CANDIDATE_THRESHOLD":  "0.5",  // below 0.7
		"PVNAIVE_STEER_HYSTERESIS_MARGIN":    "0.05", // below 0.1
		"PVNAIVE_STEER_HYSTERESIS_WINDOWS":   "9",    // above 4
		"PVNAIVE_STEER_MIN_SAMPLES":          "0",    // below 1
		"PVNAIVE_STEER_TICK_SECS":            "5",    // below 30
		"PVNAIVE_STEER_EWMA_ALPHA":           "2",    // above 0.5
	}
	for key, value := range cases {
		env := map[string]string{key: value}
		if _, err := steeringLoopConfigFromEnv(func(k string) string { return env[k] }); err == nil {
			t.Fatalf("%s=%s must fail startup loudly", key, value)
		} else if !strings.Contains(err.Error(), key) {
			t.Fatalf("error must name the offending variable: %v", err)
		}
	}
}

func TestAggregatesBackendFailClosedOnStoreError(t *testing.T) {
	backend := aggregatesBackend{store: nil, staleAfter: telemetry.SteeringStaleAfter, budget: time.Second}
	if got := backend.ReadAggregates(); got != nil {
		t.Fatalf("nil store must yield no aggregates, got %+v", got)
	}
}
