package steeringsched

import (
	"context"
	"testing"
	"time"

	"github.com/DashSaman/PV-NaivePanel/internal/steering"
	"github.com/DashSaman/PV-NaivePanel/internal/telemetry"
)

type fakeBackend struct {
	aggs []telemetry.NetworkAggregate
}

func (f *fakeBackend) ReadAggregates() []telemetry.NetworkAggregate { return f.aggs }

type recordingSink struct {
	calls []steering.Decision
	err   error
}

func (r *recordingSink) ApplyDecision(_ context.Context, d steering.Decision) error {
	r.calls = append(r.calls, d)
	return r.err
}

func agg(user, node string, rttUs int64, samples int64, lastAt time.Time) telemetry.NetworkAggregate {
	return telemetry.NetworkAggregate{
		UserID: user, NodeID: node,
		RTTEWMAUs: rttUs, SampleCount: samples, LastSampledAt: lastAt,
	}
}

func newScheduler(t *testing.T, backend Backend, sink DecisionSink, health map[string]steering.NodeHealth, now time.Time) *Scheduler {
	t.Helper()
	engine, err := steering.NewEngine(steering.DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	s, err := New(engine, backend, sink, func(context.Context) map[string]steering.NodeHealth { return health })
	if err != nil {
		t.Fatal(err)
	}
	return s.WithClock(func() time.Time { return now })
}

func TestTickDispatchesInitialDecision(t *testing.T) {
	now := time.Date(2026, 9, 14, 12, 0, 0, 0, time.UTC)
	sink := &recordingSink{}
	s := newScheduler(t, &fakeBackend{aggs: []telemetry.NetworkAggregate{
		agg("user-1", "node-a", 40_000, 20, now.Add(-time.Minute)),
	}}, sink, map[string]steering.NodeHealth{"node-a": steering.HealthHealthy}, now)

	report := s.Tick(context.Background())
	if report.Dispatched != 1 || report.SinkErrors != 0 {
		t.Fatalf("report=%+v", report)
	}
	if len(sink.calls) != 1 {
		t.Fatalf("sink calls=%d", len(sink.calls))
	}
	d := sink.calls[0]
	if d.UserID != "user-1" || d.Primary != "node-a" {
		t.Fatalf("decision=%+v", d)
	}
	if d.Reason != steering.ReasonInitial || !d.Changed {
		t.Fatalf("reason=%v changed=%v", d.Reason, d.Changed)
	}
	wantIdx := steering.WindowIndex(now, "user-1", steering.DefaultConfig().Window)
	if d.WindowIdx != wantIdx {
		t.Fatalf("windowIdx=%d want=%d", d.WindowIdx, wantIdx)
	}
}

func TestTickSkipsStaleAndThinAggregates(t *testing.T) {
	now := time.Date(2026, 9, 14, 12, 0, 0, 0, time.UTC)
	sink := &recordingSink{}
	s := newScheduler(t, &fakeBackend{aggs: []telemetry.NetworkAggregate{
		agg("user-stale", "node-a", 40_000, 20, now.Add(-5*time.Hour)), // older than 4h
		agg("user-thin", "node-a", 40_000, 3, now.Add(-time.Minute)),   // below MinSamples
	}}, sink, map[string]steering.NodeHealth{"node-a": steering.HealthHealthy}, now)

	report := s.Tick(context.Background())
	if report.Users != 0 || report.Dispatched != 0 || len(sink.calls) != 0 {
		t.Fatalf("report=%+v sink=%d", report, len(sink.calls))
	}
}

func TestTickDoesNotDispatchNoChange(t *testing.T) {
	now := time.Date(2026, 9, 14, 12, 0, 0, 0, time.UTC)
	sink := &recordingSink{}
	s := newScheduler(t, &fakeBackend{aggs: []telemetry.NetworkAggregate{
		agg("user-1", "node-a", 40_000, 20, now.Add(-time.Minute)),
	}}, sink, map[string]steering.NodeHealth{"node-a": steering.HealthHealthy}, now)

	if report := s.Tick(context.Background()); report.Dispatched != 1 {
		t.Fatalf("first tick report=%+v", report)
	}
	report := s.Tick(context.Background())
	if report.Dispatched != 0 || len(sink.calls) != 1 {
		t.Fatalf("second tick report=%+v sink=%d", report, len(sink.calls))
	}
}

func TestTickDispatchesKillSwitchFailover(t *testing.T) {
	now := time.Date(2026, 9, 14, 12, 0, 0, 0, time.UTC)
	sink := &recordingSink{}
	health := map[string]steering.NodeHealth{"node-a": steering.HealthHealthy, "node-b": steering.HealthHealthy}
	s := newScheduler(t, &fakeBackend{aggs: []telemetry.NetworkAggregate{
		agg("user-1", "node-a", 40_000, 20, now.Add(-time.Minute)),
		agg("user-1", "node-b", 90_000, 20, now.Add(-time.Minute)),
	}}, sink, health, now)

	if report := s.Tick(context.Background()); report.Dispatched != 1 {
		t.Fatalf("first tick report=%+v", report)
	}
	if sink.calls[0].Primary != "node-a" {
		t.Fatalf("initial primary=%v", sink.calls[0].Primary)
	}

	health["node-a"] = steering.HealthOffline
	report := s.Tick(context.Background())
	if report.Dispatched != 1 || len(sink.calls) != 2 {
		t.Fatalf("second tick report=%+v sink=%d", report, len(sink.calls))
	}
	d := sink.calls[1]
	if d.Primary != "node-b" || d.Reason != steering.ReasonKillSwitch {
		t.Fatalf("failover decision=%+v", d)
	}
}

func TestTickContinuesWhenSinkFails(t *testing.T) {
	now := time.Date(2026, 9, 14, 12, 0, 0, 0, time.UTC)
	sink := &recordingSink{err: context.DeadlineExceeded}
	s := newScheduler(t, &fakeBackend{aggs: []telemetry.NetworkAggregate{
		agg("user-1", "node-a", 40_000, 20, now.Add(-time.Minute)),
		agg("user-2", "node-a", 60_000, 20, now.Add(-time.Minute)),
	}}, sink, map[string]steering.NodeHealth{"node-a": steering.HealthHealthy}, now)

	report := s.Tick(context.Background())
	if report.SinkErrors != 2 || report.Dispatched != 2 || len(sink.calls) != 2 {
		t.Fatalf("report=%+v sink=%d", report, len(sink.calls))
	}
}

func TestRunStopsOnContextCancel(t *testing.T) {
	now := time.Date(2026, 9, 14, 12, 0, 0, 0, time.UTC)
	sink := &recordingSink{}
	s := newScheduler(t, &fakeBackend{aggs: []telemetry.NetworkAggregate{
		agg("user-1", "node-a", 40_000, 20, now.Add(-time.Minute)),
	}}, sink, map[string]steering.NodeHealth{"node-a": steering.HealthHealthy}, now)

	ctx, cancel := context.WithCancel(context.Background())
	ticks := 0
	done := make(chan struct{})
	go func() {
		s.Run(ctx, 10*time.Millisecond, func(TickReport) { ticks++ })
		close(done)
	}()
	time.Sleep(80 * time.Millisecond)
	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not return after cancel")
	}
	if ticks < 1 {
		t.Fatalf("ticks=%d", ticks)
	}
}

func TestNewRejectsNilCollaborators(t *testing.T) {
	engine, err := steering.NewEngine(steering.DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := New(nil, &fakeBackend{}, &recordingSink{}, nil); err == nil {
		t.Fatal("nil engine accepted")
	}
	if _, err := New(engine, nil, &recordingSink{}, nil); err == nil {
		t.Fatal("nil backend accepted")
	}
	if _, err := New(engine, &fakeBackend{}, nil, nil); err == nil {
		t.Fatal("nil sink accepted")
	}
}
