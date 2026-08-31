package main

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

type fakePeriodicResetExecutor struct {
	mu      sync.Mutex
	results []periodicUsageResetBatch
	errs    []error
	calls   int
	called  chan struct{}
}

func (f *fakePeriodicResetExecutor) ExecuteDue(ctx context.Context, limit int) (periodicUsageResetBatch, error) {
	f.mu.Lock()
	idx := f.calls
	f.calls++
	var result periodicUsageResetBatch
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

func (f *fakePeriodicResetExecutor) count() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.calls
}

func TestPeriodicUsageResetConfigDefaults(t *testing.T) {
	cfg, err := periodicUsageResetConfigFromEnv(func(string) string { return "" })
	if err != nil {
		t.Fatalf("config error: %v", err)
	}
	if cfg.Interval != 30*time.Second {
		t.Fatalf("interval=%v", cfg.Interval)
	}
	if cfg.BatchLimit != 50 {
		t.Fatalf("batch=%d", cfg.BatchLimit)
	}
	if cfg.QueryTimeout != 10*time.Second {
		t.Fatalf("timeout=%v", cfg.QueryTimeout)
	}
}

func TestPeriodicUsageResetConfigRejectsUnsafeBounds(t *testing.T) {
	cases := []map[string]string{
		{"PVNAIVE_PERIODIC_RESET_INTERVAL_SECONDS": "1"},
		{"PVNAIVE_PERIODIC_RESET_INTERVAL_SECONDS": "3601"},
		{"PVNAIVE_PERIODIC_RESET_BATCH_LIMIT": "0"},
		{"PVNAIVE_PERIODIC_RESET_BATCH_LIMIT": "101"},
	}
	for _, values := range cases {
		_, err := periodicUsageResetConfigFromEnv(func(key string) string { return values[key] })
		if err == nil {
			t.Fatalf("expected validation error for %#v", values)
		}
	}
}

func TestPeriodicUsageResetSchedulerRunsImmediatelyAndStops(t *testing.T) {
	exec := &fakePeriodicResetExecutor{called: make(chan struct{}, 2)}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		runPeriodicUsageResetScheduler(ctx, exec, periodicUsageResetConfig{Interval: time.Hour, BatchLimit: 7, QueryTimeout: time.Second}, func(string, ...any) {})
		close(done)
	}()
	select {
	case <-exec.called:
		cancel()
	case <-time.After(time.Second):
		t.Fatal("scheduler did not execute immediately")
	}
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("scheduler did not stop after cancellation")
	}
	if exec.count() != 1 {
		t.Fatalf("calls=%d", exec.count())
	}
}

func TestPeriodicUsageResetSchedulerContinuesAfterTransientError(t *testing.T) {
	exec := &fakePeriodicResetExecutor{
		called:  make(chan struct{}, 4),
		results: []periodicUsageResetBatch{{}, {Processed: 1, Succeeded: 1}},
		errs:    []error{errors.New("temporary database failure"), nil},
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var mu sync.Mutex
	logs := 0
	done := make(chan struct{})
	go func() {
		runPeriodicUsageResetScheduler(ctx, exec, periodicUsageResetConfig{Interval: 5 * time.Millisecond, BatchLimit: 7, QueryTimeout: time.Second}, func(string, ...any) { mu.Lock(); logs++; mu.Unlock() })
		close(done)
	}()
	deadline := time.After(time.Second)
	for exec.count() < 2 {
		select {
		case <-exec.called:
		case <-deadline:
			t.Fatal("scheduler did not retry after transient failure")
		}
	}
	cancel()
	<-done
	mu.Lock()
	gotLogs := logs
	mu.Unlock()
	if gotLogs < 2 {
		t.Fatalf("logs=%d, want error and processed batch logs", gotLogs)
	}
}
