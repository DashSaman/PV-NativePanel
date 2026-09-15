package main

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

type fakeReconcileExecutor struct {
	mu      sync.Mutex
	results []accountingReconcileResult
	errs    []error
	calls   int
	called  chan struct{}
}

func (f *fakeReconcileExecutor) ReleaseStale(ctx context.Context, staleSeconds, limit int) (accountingReconcileResult, error) {
	f.mu.Lock()
	idx := f.calls
	f.calls++
	var result accountingReconcileResult
	var err error
	if idx < len(f.results) {
		result = f.results[idx]
	}
	if idx < len(f.errs) {
		err = f.errs[idx]
	}
	ch := f.called
	f.mu.Unlock()
	if ch != nil {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
	return result, err
}

func (f *fakeReconcileExecutor) count() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.calls
}

func TestAccountingReconcileConfigDefaults(t *testing.T) {
	cfg, err := accountingReconcileConfigFromEnv(func(string) string { return "" })
	if err != nil {
		t.Fatalf("config error: %v", err)
	}
	if cfg.Tick != 60*time.Second {
		t.Fatalf("tick=%v", cfg.Tick)
	}
	if cfg.StaleWindow != 120*time.Second {
		t.Fatalf("stale=%v", cfg.StaleWindow)
	}
	if cfg.BatchLimit != 500 {
		t.Fatalf("batch=%d", cfg.BatchLimit)
	}
	if cfg.QueryTimeout != 10*time.Second {
		t.Fatalf("timeout=%v", cfg.QueryTimeout)
	}
}

func TestAccountingReconcileConfigRejectsUnsafeBounds(t *testing.T) {
	cases := []map[string]string{
		{"PVNAIVE_RECONCILE_TICK_SECS": "29"},
		{"PVNAIVE_RECONCILE_TICK_SECS": "3601"},
		{"PVNAIVE_RECONCILE_STALE_SECS": "59"},
		{"PVNAIVE_RECONCILE_STALE_SECS": "901"},
		{"PVNAIVE_RECONCILE_BATCH": "0"},
		{"PVNAIVE_RECONCILE_BATCH": "5001"},
	}
	for _, values := range cases {
		_, err := accountingReconcileConfigFromEnv(func(key string) string { return values[key] })
		if err == nil {
			t.Fatalf("expected validation error for %#v", values)
		}
	}
}

func TestAccountingReconcileLoopRunsImmediatelyAndStops(t *testing.T) {
	exec := &fakeReconcileExecutor{called: make(chan struct{}, 2)}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		runAccountingReconcileLoop(ctx, exec, accountingReconcileConfig{Tick: time.Hour, StaleWindow: 120 * time.Second, BatchLimit: 7, QueryTimeout: time.Second}, func(string, ...any) {})
		close(done)
	}()
	select {
	case <-exec.called:
		cancel()
	case <-time.After(time.Second):
		t.Fatal("reconcile loop did not execute immediately")
	}
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("reconcile loop did not stop after cancellation")
	}
	if exec.count() != 1 {
		t.Fatalf("calls=%d", exec.count())
	}
}

func TestAccountingReconcileLoopContinuesAfterTransientError(t *testing.T) {
	exec := &fakeReconcileExecutor{
		called:  make(chan struct{}, 4),
		results: []accountingReconcileResult{{}, {ReleasedClaims: 2, ReleasedBytes: 4096}},
		errs:    []error{errors.New("temporary database failure"), nil},
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var mu sync.Mutex
	logs := 0
	done := make(chan struct{})
	go func() {
		runAccountingReconcileLoop(ctx, exec, accountingReconcileConfig{Tick: 5 * time.Millisecond, StaleWindow: 120 * time.Second, BatchLimit: 7, QueryTimeout: time.Second}, func(string, ...any) { mu.Lock(); logs++; mu.Unlock() })
		close(done)
	}()
	deadline := time.After(time.Second)
	for exec.count() < 2 {
		select {
		case <-exec.called:
		case <-deadline:
			t.Fatal("reconcile loop did not retry after transient failure")
		}
	}
	cancel()
	<-done
	mu.Lock()
	gotLogs := logs
	mu.Unlock()
	if gotLogs < 2 {
		t.Fatalf("logs=%d, want error and released-batch logs", gotLogs)
	}
}
