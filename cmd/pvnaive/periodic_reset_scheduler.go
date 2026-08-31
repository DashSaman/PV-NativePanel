package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	defaultPeriodicUsageResetInterval = 30 * time.Second
	defaultPeriodicUsageResetBatch    = 50
	defaultPeriodicUsageResetTimeout  = 10 * time.Second
)

type periodicUsageResetBatch struct {
	Processed int
	Succeeded int
	Deferred  int
	Skipped   int
}

type periodicUsageResetExecutor interface {
	ExecuteDue(context.Context, int) (periodicUsageResetBatch, error)
}

type periodicUsageResetDBExecutor struct{ db *sql.DB }

func (e periodicUsageResetDBExecutor) ExecuteDue(ctx context.Context, limit int) (periodicUsageResetBatch, error) {
	if e.db == nil {
		return periodicUsageResetBatch{}, errors.New("periodic usage reset database is unavailable")
	}
	var out periodicUsageResetBatch
	if err := e.db.QueryRowContext(ctx, `
SELECT processed,succeeded,deferred,skipped
FROM pvnaive.execute_due_scheduled_usage_resets($1)`, limit).Scan(
		&out.Processed, &out.Succeeded, &out.Deferred, &out.Skipped,
	); err != nil {
		return periodicUsageResetBatch{}, fmt.Errorf("execute due periodic usage resets: %w", err)
	}
	return out, nil
}

type periodicUsageResetConfig struct {
	Interval     time.Duration
	BatchLimit   int
	QueryTimeout time.Duration
}

func periodicUsageResetConfigFromEnv(getenv func(string) string) (periodicUsageResetConfig, error) {
	if getenv == nil {
		return periodicUsageResetConfig{}, errors.New("periodic usage reset environment is unavailable")
	}
	intervalSeconds, err := boundedPeriodicResetEnvInt(getenv("PVNAIVE_PERIODIC_RESET_INTERVAL_SECONDS"), 30, 5, 3600)
	if err != nil {
		return periodicUsageResetConfig{}, fmt.Errorf("PVNAIVE_PERIODIC_RESET_INTERVAL_SECONDS: %w", err)
	}
	batchLimit, err := boundedPeriodicResetEnvInt(getenv("PVNAIVE_PERIODIC_RESET_BATCH_LIMIT"), defaultPeriodicUsageResetBatch, 1, 100)
	if err != nil {
		return periodicUsageResetConfig{}, fmt.Errorf("PVNAIVE_PERIODIC_RESET_BATCH_LIMIT: %w", err)
	}
	return periodicUsageResetConfig{
		Interval:     time.Duration(intervalSeconds) * time.Second,
		BatchLimit:   batchLimit,
		QueryTimeout: defaultPeriodicUsageResetTimeout,
	}, nil
}

func boundedPeriodicResetEnvInt(raw string, fallback, min, max int) (int, error) {
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

func runPeriodicUsageResetScheduler(
	ctx context.Context,
	executor periodicUsageResetExecutor,
	cfg periodicUsageResetConfig,
	logf func(string, ...any),
) {
	if executor == nil || cfg.Interval <= 0 || cfg.BatchLimit < 1 || cfg.QueryTimeout <= 0 {
		if logf != nil {
			logf("PVNaive periodic usage reset scheduler refused invalid configuration")
		}
		return
	}
	if logf == nil {
		logf = func(string, ...any) {}
	}

	runOnce := func() {
		stepCtx, cancel := context.WithTimeout(ctx, cfg.QueryTimeout)
		defer cancel()
		result, err := executor.ExecuteDue(stepCtx, cfg.BatchLimit)
		if err != nil {
			if !errors.Is(err, context.Canceled) || ctx.Err() == nil {
				logf("PVNaive periodic usage reset scheduler error: %v", err)
			}
			return
		}
		if result.Processed > 0 {
			logf(
				"PVNaive periodic usage reset batch: processed=%d succeeded=%d deferred=%d skipped=%d",
				result.Processed, result.Succeeded, result.Deferred, result.Skipped,
			)
		}
	}

	runOnce()
	ticker := time.NewTicker(cfg.Interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			runOnce()
		}
	}
}
