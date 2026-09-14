package telemetry

import (
	"math"
	"testing"
	"time"
)

func netTestSample(seq int64, at time.Time, session string, rttUs, segsOut, segsRetr, bytesIn, bytesOut, durUs int64) NetworkSample {
	return NetworkSample{
		RuntimeCredentialID: "11111111-1111-1111-1111-111111111111",
		NodeID:              "direct-1",
		BootID:              "22222222-2222-2222-2222-222222222222",
		SessionID:           session,
		SampleSeq:           seq,
		SampledAt:           at,
		Path:                NetworkPathUpstream,
		RTTMicros:           rttUs,
		RTTVarMicros:        100,
		SegsOut:             segsOut,
		SegsRetrans:         segsRetr,
		BytesIn:             bytesIn,
		BytesOut:            bytesOut,
		DurationMicros:      durUs,
	}
}

func TestValidateNetworkSample(t *testing.T) {
	base := netTestSample(1, time.Now().UTC(), "33333333-3333-3333-3333-333333333333", 1000, 100, 1, 10, 20, 1000)
	if err := ValidateNetworkSample(base); err != nil {
		t.Fatalf("valid sample rejected: %v", err)
	}
	cases := map[string]func(*NetworkSample){
		"bad credential": func(s *NetworkSample) { s.RuntimeCredentialID = "nope" },
		"bad boot":       func(s *NetworkSample) { s.BootID = "" },
		"bad session":    func(s *NetworkSample) { s.SessionID = "zz" },
		"empty node":     func(s *NetworkSample) { s.NodeID = "  " },
		"zero seq":       func(s *NetworkSample) { s.SampleSeq = 0 },
		"zero time":      func(s *NetworkSample) { s.SampledAt = time.Time{} },
		"bad path":       func(s *NetworkSample) { s.Path = "sideways" },
		"negative rtt":   func(s *NetworkSample) { s.RTTMicros = -1 },
		"negative var":   func(s *NetworkSample) { s.RTTVarMicros = -1 },
		"negative segs":  func(s *NetworkSample) { s.SegsOut = -5 },
		"retrans>out":    func(s *NetworkSample) { s.SegsRetrans = 101; s.SegsOut = 100 },
		"negative bytes": func(s *NetworkSample) { s.BytesIn = -1 },
		"negative dur":   func(s *NetworkSample) { s.DurationMicros = -1 },
	}
	for name, mutate := range cases {
		sample := base
		mutate(&sample)
		if err := ValidateNetworkSample(sample); err == nil {
			t.Fatalf("%s: expected rejection", name)
		}
	}
}

func TestEWMAHolderFirstSampleAndUnknownPolicy(t *testing.T) {
	holder := NewEWMAHolder(DefaultNetworkAggConfig())
	session := "33333333-3333-3333-3333-333333333333"
	base := time.Date(2026, 9, 14, 10, 0, 0, 0, time.UTC)
	sample := netTestSample(1, base, session, 20000, 100, 2, 1000, 2000, 1_000_000)

	agg, known := holder.Apply("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb", "direct-1", []NetworkSample{sample})
	if known {
		t.Fatal("single sample must stay Unknown (min_samples=8)")
	}
	if agg.RTTEWMAUs != 20000 || agg.SampleCount != 1 {
		t.Fatalf("first sample: rtt=%d count=%d", agg.RTTEWMAUs, agg.SampleCount)
	}
	if agg.RetransRatioEWMA != 0 || agg.ThroughputBps != 0 {
		t.Fatalf("span start must not fabricate rate features: %+v", agg)
	}
}

func TestEWMAHolderParityVector(t *testing.T) {
	// Deterministic EWMA parity: rtt series 20000, 30000 with alpha=0.2 gives
	// 0.2*30000 + 0.8*20000 = 22000 exactly.
	holder := NewEWMAHolder(DefaultNetworkAggConfig())
	session := "33333333-3333-3333-3333-333333333333"
	base := time.Date(2026, 9, 14, 10, 0, 0, 0, time.UTC)
	userID := "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"

	series := []NetworkSample{
		netTestSample(1, base, session, 20000, 100, 0, 0, 0, 1_000_000),
		netTestSample(2, base.Add(7*time.Second), session, 30000, 200, 4, 1000, 3000, 8_000_000),
	}
	agg, _ := holder.Apply(userID, "direct-1", series)
	if agg.RTTEWMAUs != 22000 {
		t.Fatalf("EWMA parity: want 22000, got %d", agg.RTTEWMAUs)
	}
	// Jitter = |30000-20000| = 10000 (first delta in continuous span).
	if agg.JitterEWMAUs != 10000 {
		t.Fatalf("jitter parity: want 10000, got %d", agg.JitterEWMAUs)
	}
	// Retrans delta: (4-0)/(200-100)=0.04.
	if math.Abs(agg.RetransRatioEWMA-0.04) > 1e-9 {
		t.Fatalf("retrans parity: want 0.04, got %f", agg.RetransRatioEWMA)
	}
	// Throughput: (1000+3000) bytes over 7s => 4000*1e6/7e6 ≈ 571 bps.
	wantThpt := int64(571)
	if math.Abs(float64(agg.ThroughputBps-wantThpt)) > 1 {
		t.Fatalf("throughput parity: want ~%d, got %d", wantThpt, agg.ThroughputBps)
	}
}

func TestEWMAHolderSpanBoundaryNoCrossSessionDeltas(t *testing.T) {
	holder := NewEWMAHolder(DefaultNetworkAggConfig())
	userID := "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"
	base := time.Date(2026, 9, 14, 10, 0, 0, 0, time.UTC)
	sessionA := "33333333-3333-3333-3333-333333333333"
	sessionB := "44444444-4444-4444-4444-444444444444"

	// Session A ends with cumulative counters at 1000/500/20; session B starts
	// its own counters near zero. A naive cross-session delta would go negative
	// or fabricate retrans spikes; the holder must treat B as a new span.
	series := []NetworkSample{
		netTestSample(1, base, sessionA, 20000, 1000, 20, 500, 500, 10_000_000),
		netTestSample(2, base.Add(7*time.Second), sessionB, 30000, 3, 0, 0, 0, 100_000),
	}
	agg, _ := holder.Apply(userID, "direct-1", series)
	if agg.JitterEWMAUs != 0 || agg.RetransRatioEWMA != 0 || agg.ThroughputBps != 0 {
		t.Fatalf("cross-session deltas leaked: %+v", agg)
	}
	if agg.SampleCount != 2 {
		t.Fatalf("sample count: %d", agg.SampleCount)
	}
}

func TestEWMAHolderReplayDeterministic(t *testing.T) {
	run := func() NetworkAggregate {
		holder := NewEWMAHolder(DefaultNetworkAggConfig())
		session := "33333333-3333-3333-3333-333333333333"
		base := time.Date(2026, 9, 14, 10, 0, 0, 0, time.UTC)
		series := make([]NetworkSample, 0, 10)
		segs := int64(0)
		for i := int64(1); i <= 10; i++ {
			segs += 50
			retr := segs / 25
			series = append(series, netTestSample(i, base.Add(time.Duration(i)*7*time.Second), session, 20000+i*100, segs, retr, i*1000, i*2000, i*7_000_000))
		}
		agg, known := holder.Apply("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb", "direct-1", series)
		if !known {
			t.Fatal("10 samples must reach MinSamples=8")
		}
		return agg
	}
	a, b := run(), run()
	if a != b {
		t.Fatalf("replay not deterministic:\n%+v\n%+v", a, b)
	}
}

func TestEWMAHolderSeedRestartsKeepMemory(t *testing.T) {
	first := NewEWMAHolder(DefaultNetworkAggConfig())
	session := "33333333-3333-3333-3333-333333333333"
	base := time.Date(2026, 9, 14, 10, 0, 0, 0, time.UTC)
	userID := "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"
	agg, known := first.Apply(userID, "direct-1", []NetworkSample{
		netTestSample(1, base, session, 20000, 100, 0, 0, 0, 1_000_000),
	})
	if known {
		t.Fatal("must be Unknown before MinSamples")
	}

	// "Restart": fresh holder seeded from persisted aggregate.
	second := NewEWMAHolder(DefaultNetworkAggConfig())
	second.Seed(agg)
	agg2, _ := second.Apply(userID, "direct-1", []NetworkSample{
		netTestSample(2, base.Add(7*time.Second), session, 30000, 200, 4, 1000, 3000, 8_000_000),
	})
	if agg2.SampleCount != 2 || agg2.RTTEWMAUs != 22000 {
		t.Fatalf("seed failed: count=%d rtt=%d", agg2.SampleCount, agg2.RTTEWMAUs)
	}
	// Seed refuses to move state backwards.
	stale := agg2
	stale.LastSampledAt = base
	stale.SampleCount = 1
	second.Seed(stale)
	var snapshot NetworkAggregate
	for _, candidate := range second.Snapshot() {
		if candidate.UserID == userID && candidate.NodeID == "direct-1" {
			snapshot = candidate
		}
	}
	if snapshot.SampleCount != 2 {
		t.Fatalf("stale seed overwrote state: %+v", snapshot)
	}
}

func TestSteeringEligibleFreshness(t *testing.T) {
	now := time.Date(2026, 9, 14, 12, 0, 0, 0, time.UTC)
	agg := NetworkAggregate{SampleCount: 8, LastSampledAt: now.Add(-time.Hour)}
	if !SteeringEligible(agg, 8, SteeringStaleAfter, now) {
		t.Fatal("fresh aggregate must be eligible")
	}
	agg.LastSampledAt = now.Add(-5 * time.Hour)
	if SteeringEligible(agg, 8, SteeringStaleAfter, now) {
		t.Fatal("stale aggregate must be Unknown")
	}
	agg.SampleCount = 7
	agg.LastSampledAt = now.Add(-time.Hour)
	if SteeringEligible(agg, 8, SteeringStaleAfter, now) {
		t.Fatal("below MinSamples must be Unknown")
	}
}
