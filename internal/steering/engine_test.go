package steering

import (
	"testing"
	"time"
)

func agg(user, node string, rtt, jitter, retrans, load float64, samples int64) Aggregate {
	return Aggregate{
		UserID: user, NodeID: node,
		RTTMedianMs: rtt, JitterMs: jitter,
		RetransRatio: retrans, LoadRatio: load,
		SuccessRate: 1, SampleCount: samples,
	}
}

func healthAll(nodes []string, h NodeHealth) map[string]NodeHealth {
	m := make(map[string]NodeHealth, len(nodes))
	for _, n := range nodes {
		m[n] = h
	}
	return m
}

func TestConfigValidate(t *testing.T) {
	if err := DefaultConfig().Validate(); err != nil {
		t.Fatalf("defaults must be valid, got %v", err)
	}
	cfg := DefaultConfig()
	cfg.Alpha = 0.9
	if err := cfg.Validate(); err == nil {
		t.Fatal("alpha out of range must be rejected")
	}
	cfg = DefaultConfig()
	cfg.Window = time.Minute
	if err := cfg.Validate(); err == nil {
		t.Fatal("window below 30m must be rejected")
	}
}

func TestScoreDeterministicOrderingAndPenalties(t *testing.T) {
	cfg := DefaultConfig()
	aggs := []Aggregate{
		agg("u1", "node-b", 40, 2, 0.01, 0.1, 100),
		agg("u1", "node-a", 20, 1, 0.01, 0.1, 100),
		agg("u1", "node-c", 20, 1, 0.60, 0.1, 100), // same rtt/jitter as node-a, heavy retrans
	}
	got := Score(cfg, aggs)
	if got[0].NodeID != "node-a" {
		t.Fatalf("best node = %q, want node-a (tie broken by id then penalties)", got[0].NodeID)
	}
	var nodeAScore float64
	for _, s := range got {
		if s.NodeID == "node-a" {
			nodeAScore = s.Score
		}
	}
	if nodeAScore <= 0 || nodeAScore > 100 {
		t.Fatalf("node-a score %v outside (0,100]", nodeAScore)
	}
	// Determinism: same input, same output.
	again := Score(cfg, aggs)
	for i := range got {
		if got[i] != again[i] {
			t.Fatalf("score not deterministic: %v vs %v", got, again)
		}
	}
}

func TestUnknownExcludedUntilMinSamples(t *testing.T) {
	cfg := DefaultConfig() // MinSamples 8
	aggs := []Aggregate{
		agg("u1", "node-fast", 5, 0, 0, 0, 4),     // great rtt but Unknown
		agg("u1", "node-slow", 90, 1, 0.1, 0, 50), // fully observed
	}
	scored := Score(cfg, aggs)
	if scored[0].NodeID != "node-slow" || scored[0].Kind != "eligible" {
		t.Fatalf("unknown node must not win: %+v", scored)
	}
	if scored[1].NodeID != "node-fast" || scored[1].Kind != "unknown" {
		t.Fatalf("under-sampled node must be reported Unknown: %+v", scored)
	}
	if got := TopK(scored, 5); len(got) != 1 || got[0].NodeID != "node-slow" {
		t.Fatalf("TopK must filter Unknown: %+v", got)
	}

	eng, err := NewEngine(cfg)
	if err != nil {
		t.Fatal(err)
	}
	d := eng.Evaluate("u1", 1, aggs, healthAll([]string{"node-fast", "node-slow"}, HealthHealthy))
	if d.Primary != "node-slow" {
		t.Fatalf("engine must pick only eligible node, got %q", d.Primary)
	}
}

func TestNoFlappingNearEqualNodes(t *testing.T) {
	// STEER-002 acceptance: two nodes within 15ms must not oscillate over a
	// simulated 24h of windows.
	cfg := DefaultConfig()
	cfg.HysteresisMargin = 0.25
	eng, err := NewEngine(cfg)
	if err != nil {
		t.Fatal(err)
	}
	nodes := []string{"node-a", "node-b"}
	health := healthAll(nodes, HealthHealthy)

	// node-a slightly better; node-b RTT differs by only 14ms (<15ms).
	base := []Aggregate{
		agg("u1", "node-a", 100, 2, 0.01, 0.1, 1000),
		agg("u1", "node-b", 110, 2, 0.01, 0.1, 1000),
	}
	switches := 0
	first := true
	for i := int64(0); i < 6; i++ { // 24h / 4h = 6 windows
		// Perturb inputs honestly within the near-equal band.
		aggs := []Aggregate{base[0], base[1]}
		if i%2 == 1 { // small noise still <15ms gap
			aggs[1].RTTMedianMs = 100 + 14
		}
		d := eng.Evaluate("u1", i, aggs, health)
		if first {
			first = false
			continue
		}
		if d.Changed {
			switches++
		}
	}
	if switches != 0 {
		t.Fatalf("near-equal nodes must not flap: %d switches over 6 windows", switches)
	}
}

func TestHysteresisRequiresTwoConsecutiveWindows(t *testing.T) {
	cfg := DefaultConfig() // margin 25%, windows 2
	eng, err := NewEngine(cfg)
	if err != nil {
		t.Fatal(err)
	}
	nodes := []string{"node-a", "node-b"}
	health := healthAll(nodes, HealthHealthy)

	far := []Aggregate{
		agg("u1", "node-a", 100, 0, 0, 0, 100),
		agg("u1", "node-b", 200, 0, 0, 0, 100),
	}
	better := []Aggregate{
		agg("u1", "node-a", 100, 0, 0, 0, 100),
		agg("u1", "node-b", 50, 0, 0, 0, 100), // >25% better than node-a
	}

	d0 := eng.Evaluate("u1", 0, far, health)
	if d0.Primary != "node-a" || d0.Reason != ReasonInitial {
		t.Fatalf("initial decision = %+v", d0)
	}
	d1 := eng.Evaluate("u1", 1, better, health)
	if d1.Changed || d1.Primary != "node-a" {
		t.Fatalf("first window with better candidate must not switch: %+v", d1)
	}
	d2 := eng.Evaluate("u1", 2, better, health)
	if !d2.Changed || d2.Primary != "node-b" || d2.Reason != ReasonHysteresis {
		t.Fatalf("second consecutive window must switch via hysteresis: %+v", d2)
	}

	// A broken streak resets: after settling on node-b, one strong window for
	// node-a must not be enough.
	d3 := eng.Evaluate("u1", 3, far, health)
	if d3.Changed {
		t.Fatalf("single strong window must not switch: %+v", d3)
	}
}

func TestKillSwitchImmediateFailover(t *testing.T) {
	cfg := DefaultConfig()
	eng, err := NewEngine(cfg)
	if err != nil {
		t.Fatal(err)
	}
	nodes := []string{"node-a", "node-b"}
	far := []Aggregate{
		agg("u1", "node-a", 100, 0, 0, 0, 100),
		agg("u1", "node-b", 200, 0, 0, 0, 100),
	}
	if d := eng.Evaluate("u1", 0, far, healthAll(nodes, HealthHealthy)); d.Primary != "node-a" {
		t.Fatalf("initial primary = %q, want node-a", d.Primary)
	}
	d := eng.Evaluate("u1", 1, far, map[string]NodeHealth{
		"node-a": HealthOffline, "node-b": HealthHealthy,
	})
	if !d.Changed || d.Primary != "node-b" || d.Reason != ReasonKillSwitch {
		t.Fatalf("kill-switch must fail over immediately: %+v", d)
	}
}

func TestNoCandidatesFailClosed(t *testing.T) {
	cfg := DefaultConfig()
	eng, err := NewEngine(cfg)
	if err != nil {
		t.Fatal(err)
	}
	aggs := []Aggregate{agg("u1", "node-a", 20, 0, 0, 0, 3)} // under-sampled
	health := healthAll([]string{"node-a"}, HealthHealthy)

	d := eng.Evaluate("u1", 0, aggs, health)
	if d.Reason != ReasonNoCandidates || d.Primary != "" {
		t.Fatalf("no eligible nodes must fail closed: %+v", d)
	}

	// After the first assignment, losing all candidates keeps the last-known
	// primary reported but flags no_candidates.
	seed := []Aggregate{agg("u1", "node-a", 20, 0, 0, 0, 50)}
	if d := eng.Evaluate("u1", 1, seed, health); d.Primary != "node-a" {
		t.Fatalf("seed eligible node must be assigned: %+v", d)
	}
	d2 := eng.Evaluate("u1", 2, aggs, health)
	if d2.Reason != ReasonNoCandidates || d2.Primary != "node-a" {
		t.Fatalf("last-known-good must be preserved on outage: %+v", d2)
	}
	// All nodes offline is also no-candidates.
	d3 := eng.Evaluate("u1", 3, seed, healthAll([]string{"node-a"}, HealthOffline))
	if d3.Reason != ReasonNoCandidates || d3.Changed {
		t.Fatalf("offline-only fleet must fail closed without churn: %+v", d3)
	}
}

func TestPhaseOffsetsDistributeAndAreStable(t *testing.T) {
	window := 4 * time.Hour
	seen := make(map[time.Duration]bool)
	for i := 0; i < 200; i++ {
		id := "user-" + time.Duration(i).String()
		off := PhaseOffset(id, window)
		if off < 0 || off >= window {
			t.Fatalf("phase offset %v outside [0,window)", off)
		}
		if PhaseOffset(id, window) != off {
			t.Fatal("phase offset must be deterministic")
		}
		seen[off] = true
	}
	if len(seen) < 100 {
		t.Fatalf("phase offsets should spread widely, got %d distinct", len(seen))
	}
	// WindowIndex aligns to the user's phase: the index is stable for a full
	// window starting at any phase-aligned boundary and advances by exactly
	// one at the next boundary.
	phase := PhaseOffset("u1", window)
	t0 := time.Unix(0, int64(phase)).Add(124 * window)
	w1 := WindowIndex(t0, "u1", window)
	if WindowIndex(t0.Add(window-time.Nanosecond), "u1", window) != w1 {
		t.Fatalf("window index must be stable within the window")
	}
	if WindowIndex(t0.Add(window), "u1", window) != w1+1 {
		t.Fatal("window index must advance by exactly one per window")
	}
}

func TestTopKClipsAndOrders(t *testing.T) {
	cfg := DefaultConfig()
	aggs := []Aggregate{
		agg("u1", "n1", 10, 0, 0, 0, 100),
		agg("u1", "n2", 20, 0, 0, 0, 100),
		agg("u1", "n3", 30, 0, 0, 0, 100),
		agg("u1", "n4", 40, 0, 0, 0, 3), // Unknown
		agg("u1", "n5", 50, 0, 0, 0, 100),
	}
	scored := Score(cfg, aggs)
	got := TopK(scored, 2)
	if len(got) != 2 || got[0].NodeID != "n1" || got[1].NodeID != "n2" {
		t.Fatalf("TopK must return best-first eligible only: %+v", got)
	}
	if got := TopK(scored, 100); len(got) != 4 {
		t.Fatalf("TopK must clip to eligible count: %d", len(got))
	}
	if got := TopK(scored, 0); got != nil {
		t.Fatalf("TopK with k=0 must be nil: %+v", got)
	}
}

func TestEWMAStep(t *testing.T) {
	if got := EWMA(100, 50, 0.2); got != 90 {
		t.Fatalf("EWMA(100,50,0.2) = %v, want 90", got)
	}
}
