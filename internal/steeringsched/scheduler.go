// Package steeringsched closes the R1→R2 loop: it periodically reads R1
// network aggregates, evaluates the R2 rotation engine per active user and
// dispatches only actionable decisions (initial assignment, hysteresis
// switch, kill-switch failover) to a sink — the R3 renderer hook. no_change
// and no_candidates outcomes are never dispatched: callers keep serving
// last-known-good, nothing is fabricated.
package steeringsched

import (
	"context"
	"errors"
	"sort"
	"time"

	"github.com/DashSaman/PV-NaivePanel/internal/steering"
	"github.com/DashSaman/PV-NaivePanel/internal/telemetry"
)

// Backend is the R1 network-telemetry surface the scheduler consumes.
type Backend interface {
	ReadAggregates() []telemetry.NetworkAggregate
}

// DecisionSink consumes actionable rotation decisions (R3 renderer hook).
type DecisionSink interface {
	ApplyDecision(ctx context.Context, decision steering.Decision) error
}

// HealthSource supplies current fleet node health (R5 wiring point). Nodes
// absent from the map are treated as unknown by the engine (never offline).
type HealthSource func(ctx context.Context) map[string]steering.NodeHealth

// TickReport summarizes one rotation pass for observability.
type TickReport struct {
	Users        int // users with at least one fresh aggregate
	Dispatched   int // decisions handed to the sink
	NoCandidates int // users kept on last-known-good (no eligible node)
	SinkErrors   int // sink failures (tick continues with remaining users)
}

type Scheduler struct {
	engine     *steering.Engine
	backend    Backend
	sink       DecisionSink
	health     HealthSource
	staleAfter time.Duration
	now        func() time.Time
}

// New assembles the scheduler. staleAfter defaults to
// telemetry.SteeringStaleAfter; health may be nil (every node unknown).
func New(engine *steering.Engine, backend Backend, sink DecisionSink, health HealthSource) (*Scheduler, error) {
	if engine == nil {
		return nil, errors.New("steeringsched: engine is required")
	}
	if backend == nil {
		return nil, errors.New("steeringsched: backend is required")
	}
	if sink == nil {
		return nil, errors.New("steeringsched: sink is required")
	}
	return &Scheduler{
		engine:     engine,
		backend:    backend,
		sink:       sink,
		health:     health,
		staleAfter: telemetry.SteeringStaleAfter,
		now:        func() time.Time { return time.Now().UTC() },
	}, nil
}

// WithStaleAfter overrides the freshness window (chainable).
func (s *Scheduler) WithStaleAfter(d time.Duration) *Scheduler {
	if d > 0 {
		s.staleAfter = d
	}
	return s
}

// WithClock injects the clock (tests).
func (s *Scheduler) WithClock(now func() time.Time) *Scheduler {
	if now != nil {
		s.now = now
	}
	return s
}

// Tick runs one rotation pass. Users are evaluated in sorted order so engine
// state transitions are deterministic.
func (s *Scheduler) Tick(ctx context.Context) TickReport {
	report := TickReport{}
	now := s.now().UTC()
	cfg := s.engine.Config()

	perUser := make(map[string][]steering.Aggregate)
	for _, agg := range s.backend.ReadAggregates() {
		if !telemetry.SteeringEligible(agg, cfg.MinSamples, s.staleAfter, now) {
			continue
		}
		perUser[agg.UserID] = append(perUser[agg.UserID], agg.SteeringAggregate())
	}
	report.Users = len(perUser)

	health := map[string]steering.NodeHealth{}
	if s.health != nil {
		if supplied := s.health(ctx); supplied != nil {
			health = supplied
		}
	}

	users := make([]string, 0, len(perUser))
	for user := range perUser {
		users = append(users, user)
	}
	sort.Strings(users)

	for _, user := range users {
		decision := s.engine.Evaluate(user, steering.WindowIndex(now, user, cfg.Window), perUser[user], health)
		switch decision.Reason {
		case steering.ReasonNoCandidates:
			report.NoCandidates++
			continue
		case steering.ReasonNoChange:
			continue
		}
		report.Dispatched++
		if err := s.sink.ApplyDecision(ctx, decision); err != nil {
			report.SinkErrors++
		}
	}
	return report
}

// Run ticks on a fixed interval until the context is cancelled. onTick may
// be nil; it must not block for long.
func (s *Scheduler) Run(ctx context.Context, every time.Duration, onTick func(TickReport)) {
	if every <= 0 {
		every = cfgDefaultTick
	}
	ticker := time.NewTicker(every)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			report := s.Tick(ctx)
			if onTick != nil {
				onTick(report)
			}
		}
	}
}

const cfgDefaultTick = time.Minute
