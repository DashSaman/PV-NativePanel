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

// Abandoned-reservation reconciler.
//
// A quota claim (reservation) that never receives its settling telemetry event
// would keep direct_naive_accounting_terms.reserved_bytes inflated forever and
// block every new CONNECT for the affected user (the authorize gate refuses
// service while reservations are pending). The safe release path already
// exists in SQL (0022/0023: pvnaive.release_stale_accounting_reservations —
// it only retires claims whose owning session is final or has been silent
// beyond the stale window, and never guesses bytes into the ledger), but it
// had no caller. This loop drives it on a periodic tick so an unclean restart
// costs users at most one reconcile window of connectivity instead of a
// permanent outage.
const (
	defaultReconcileTickSeconds  = 60
	defaultReconcileStaleSeconds = 120
	defaultReconcileBatch        = 500
	defaultReconcileQueryTimeout = 10 * time.Second
)

type accountingReconcileResult struct {
	ReleasedClaims int64
	ReleasedBytes  int64
}

type accountingReconcileExecutor interface {
	ReleaseStale(context.Context, int, int) (accountingReconcileResult, error)
}

type accountingReconcileDBExecutor struct{ db *sql.DB }

func (e accountingReconcileDBExecutor) ReleaseStale(ctx context.Context, staleSeconds, limit int) (accountingReconcileResult, error) {
	if e.db == nil {
		return accountingReconcileResult{}, errors.New("accounting reconcile database is unavailable")
	}
	var out accountingReconcileResult
	if err := e.db.QueryRowContext(ctx, `
SELECT released_claims, released_bytes
FROM pvnaive.release_stale_accounting_reservations($1, clock_timestamp(), $2)`,
		staleSeconds, limit).Scan(
		&out.ReleasedClaims, &out.ReleasedBytes,
	); err != nil {
		return accountingReconcileResult{}, fmt.Errorf("release stale accounting reservations: %w", err)
	}
	return out, nil
}

type accountingReconcileConfig struct {
	Tick         time.Duration
	StaleWindow  time.Duration
	BatchLimit   int
	QueryTimeout time.Duration
}

func accountingReconcileConfigFromEnv(getenv func(string) string) (accountingReconcileConfig, error) {
	if getenv == nil {
		return accountingReconcileConfig{}, errors.New("accounting reconcile environment is unavailable")
	}
	tickSeconds, err := boundedReconcileEnvInt(getenv("PVNAIVE_RECONCILE_TICK_SECS"),
		defaultReconcileTickSeconds, 30, 3600)
	if err != nil {
		return accountingReconcileConfig{}, fmt.Errorf("PVNAIVE_RECONCILE_TICK_SECS: %w", err)
	}
	staleSeconds, err := boundedReconcileEnvInt(getenv("PVNAIVE_RECONCILE_STALE_SECS"),
		defaultReconcileStaleSeconds, 60, 900)
	if err != nil {
		return accountingReconcileConfig{}, fmt.Errorf("PVNAIVE_RECONCILE_STALE_SECS: %w", err)
	}
	batchLimit, err := boundedReconcileEnvInt(getenv("PVNAIVE_RECONCILE_BATCH"),
		defaultReconcileBatch, 1, 5000)
	if err != nil {
		return accountingReconcileConfig{}, fmt.Errorf("PVNAIVE_RECONCILE_BATCH: %w", err)
	}
	return accountingReconcileConfig{
		Tick:         time.Duration(tickSeconds) * time.Second,
		StaleWindow:  time.Duration(staleSeconds) * time.Second,
		BatchLimit:   batchLimit,
		QueryTimeout: defaultReconcileQueryTimeout,
	}, nil
}

func boundedReconcileEnvInt(raw string, fallback, min, max int) (int, error) {
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

func runAccountingReconcileLoop(
	ctx context.Context,
	executor accountingReconcileExecutor,
	cfg accountingReconcileConfig,
	logf func(string, ...any),
) {
	if executor == nil || cfg.Tick <= 0 || cfg.StaleWindow <= 0 || cfg.BatchLimit < 1 || cfg.QueryTimeout <= 0 {
		if logf != nil {
			logf("PVNaive accounting reconcile loop refused invalid configuration")
		}
		return
	}
	if logf == nil {
		logf = func(string, ...any) {}
	}

	runOnce := func() {
		stepCtx, cancel := context.WithTimeout(ctx, cfg.QueryTimeout)
		defer cancel()
		result, err := executor.ReleaseStale(stepCtx, int(cfg.StaleWindow/time.Second), cfg.BatchLimit)
		if err != nil {
			if !errors.Is(err, context.Canceled) || ctx.Err() == nil {
				logf("PVNaive accounting reconcile error: %v", err)
			}
			return
		}
		if result.ReleasedClaims > 0 {
			logf(
				"PVNaive accounting reconcile released abandoned reservations: claims=%d bytes=%d",
				result.ReleasedClaims, result.ReleasedBytes,
			)
		}
	}

	runOnce()
	ticker := time.NewTicker(cfg.Tick)
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
