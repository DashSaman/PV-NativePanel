package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/DashSaman/PV-NaivePanel/internal/steering"
	"github.com/DashSaman/PV-NaivePanel/internal/steeringsched"
	"github.com/DashSaman/PV-NaivePanel/internal/steeringstore"
	"github.com/DashSaman/PV-NaivePanel/internal/telemetry"
)

// R2→R3 live loop wiring (docs/STEERING_SPEC_FA.md §2, docs/AGENT_TASKS.md R2/R3):
// the scheduler periodically reads R1 network aggregates through the trusted
// boundary, evaluates the rotation engine and hands actionable decisions to
// the durable decision sink (0032 migration). Every knob is operator
// configuration; out-of-range values fail startup loudly (fail-closed), never
// silently clamp.
const (
	defaultSteerTickSeconds     = 60
	defaultSteerRotationWindow  = 4 * time.Hour
	defaultSteerGrace           = 15 * time.Minute
	defaultSteerAlpha           = 0.2
	defaultSteerCandidateRatio  = 0.85
	defaultSteerHystMargin      = 0.25
	defaultSteerHystWindows     = 2
	defaultSteerMinSamples      = 8
	defaultSteerReadTimeoutSecs = 10
)

type steeringLoopConfig struct {
	Engine     steering.Config
	Tick       time.Duration
	ReadBudget time.Duration
}

func steeringLoopConfigFromEnv(getenv func(string) string) (steeringLoopConfig, error) {
	if getenv == nil {
		return steeringLoopConfig{}, errors.New("steering loop environment is unavailable")
	}
	cfg := steering.DefaultConfig()

	rotationWindowSecs, err := boundedSteerEnvInt(getenv("PVNAIVE_STEER_ROTATION_WINDOW_SECS"),
		int(defaultSteerRotationWindow/time.Second), 1800, 86400)
	if err != nil {
		return steeringLoopConfig{}, fmt.Errorf("PVNAIVE_STEER_ROTATION_WINDOW_SECS: %w", err)
	}
	cfg.Window = time.Duration(rotationWindowSecs) * time.Second

	graceSecs, err := boundedSteerEnvInt(getenv("PVNAIVE_STEER_GRACE_SECONDS"),
		int(defaultSteerGrace/time.Second), 300, 3600)
	if err != nil {
		return steeringLoopConfig{}, fmt.Errorf("PVNAIVE_STEER_GRACE_SECONDS: %w", err)
	}
	cfg.Grace = time.Duration(graceSecs) * time.Second

	cfg.CandidateRatio, err = boundedSteerEnvFloat(getenv("PVNAIVE_STEER_CANDIDATE_THRESHOLD"), defaultSteerCandidateRatio, 0.7, 0.95)
	if err != nil {
		return steeringLoopConfig{}, fmt.Errorf("PVNAIVE_STEER_CANDIDATE_THRESHOLD: %w", err)
	}
	cfg.HysteresisMargin, err = boundedSteerEnvFloat(getenv("PVNAIVE_STEER_HYSTERESIS_MARGIN"), defaultSteerHystMargin, 0.1, 0.5)
	if err != nil {
		return steeringLoopConfig{}, fmt.Errorf("PVNAIVE_STEER_HYSTERESIS_MARGIN: %w", err)
	}
	hystWindows, err := boundedSteerEnvInt(getenv("PVNAIVE_STEER_HYSTERESIS_WINDOWS"), defaultSteerHystWindows, 1, 4)
	if err != nil {
		return steeringLoopConfig{}, fmt.Errorf("PVNAIVE_STEER_HYSTERESIS_WINDOWS: %w", err)
	}
	cfg.HysteresisWindows = hystWindows

	minSamples, err := boundedSteerEnvInt(getenv("PVNAIVE_STEER_MIN_SAMPLES"), defaultSteerMinSamples, 1, 1000)
	if err != nil {
		return steeringLoopConfig{}, fmt.Errorf("PVNAIVE_STEER_MIN_SAMPLES: %w", err)
	}
	cfg.MinSamples = int64(minSamples)

	if _, err := boundedSteerEnvFloat(getenv("PVNAIVE_STEER_EWMA_ALPHA"), defaultSteerAlpha, 0.05, 0.5); err != nil {
		return steeringLoopConfig{}, fmt.Errorf("PVNAIVE_STEER_EWMA_ALPHA: %w", err)
	}

	tickSecs, err := boundedSteerEnvInt(getenv("PVNAIVE_STEER_TICK_SECS"), defaultSteerTickSeconds, 30, 3600)
	if err != nil {
		return steeringLoopConfig{}, fmt.Errorf("PVNAIVE_STEER_TICK_SECS: %w", err)
	}
	readBudgetSecs, err := boundedSteerEnvInt(getenv("PVNAIVE_STEER_READ_TIMEOUT_SECS"), defaultSteerReadTimeoutSecs, 2, 120)
	if err != nil {
		return steeringLoopConfig{}, fmt.Errorf("PVNAIVE_STEER_READ_TIMEOUT_SECS: %w", err)
	}

	// Engine.Validate is the single source of truth for parameter ranges; a
	// second validation here proves the env mapping never drifts out of the
	// spec envelope.
	if err := cfg.Validate(); err != nil {
		return steeringLoopConfig{}, fmt.Errorf("steering engine configuration: %w", err)
	}
	return steeringLoopConfig{
		Engine:     cfg,
		Tick:       time.Duration(tickSecs) * time.Second,
		ReadBudget: time.Duration(readBudgetSecs) * time.Second,
	}, nil
}

func boundedSteerEnvInt(raw string, fallback, min, max int) (int, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return fallback, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < min || value > max {
		return 0, fmt.Errorf("must be an integer in [%d,%d]", min, max)
	}
	return value, nil
}

func boundedSteerEnvFloat(raw string, fallback, min, max float64) (float64, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return fallback, nil
	}
	value, err := strconv.ParseFloat(raw, 64)
	if err != nil || value < min || value > max {
		return 0, fmt.Errorf("must be a number in [%g,%g]", min, max)
	}
	return value, nil
}

// aggregatesBackend adapts the telemetry store to steeringsched.Backend. A
// read failure yields no aggregates for this tick — the honest fail-closed
// behavior (users keep last-known-good; nothing is fabricated).
type aggregatesBackend struct {
	store      *telemetry.PostgresStore
	staleAfter time.Duration
	budget     time.Duration
}

func (b aggregatesBackend) ReadAggregates() []telemetry.NetworkAggregate {
	if b.store == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), b.budget)
	defer cancel()
	aggregates, err := b.store.ReadNetworkAggregates(ctx, b.staleAfter)
	if err != nil {
		return nil
	}
	return aggregates
}

func startSteeringLoop(
	ctx context.Context,
	db *sql.DB,
	aggregates *telemetry.PostgresStore,
	cfg steeringLoopConfig,
	logf func(string, ...any),
) error {
	sink, err := steeringstore.NewStore(db)
	if err != nil {
		return err
	}
	engine, err := steering.NewEngine(cfg.Engine)
	if err != nil {
		return fmt.Errorf("steering engine: %w", err)
	}
	backend := aggregatesBackend{store: aggregates, staleAfter: telemetry.SteeringStaleAfter, budget: cfg.ReadBudget}
	scheduler, err := steeringsched.New(engine, backend, sink, nil)
	if err != nil {
		return err
	}
	go scheduler.Run(ctx, cfg.Tick, func(report steeringsched.TickReport) {
		if logf != nil {
			logf(
				"PVNaive steering tick: users=%d dispatched=%d no_candidates=%d sink_errors=%d",
				report.Users, report.Dispatched, report.NoCandidates, report.SinkErrors,
			)
		}
	})
	return nil
}
