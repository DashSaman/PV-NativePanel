package steering

import "testing"

func TestEvaluateFiltersCandidatesByConfiguredRatio(t *testing.T) {
	cfg := DefaultConfig()
	cfg.CandidateRatio = 0.85
	eng, err := NewEngine(cfg)
	if err != nil {
		t.Fatal(err)
	}

	aggs := []Aggregate{
		agg("u-ratio", "node-a", 10, 0, 0, 0, 100),
		agg("u-ratio", "node-b", 20, 0, 0, 0, 100),
		agg("u-ratio", "node-c", 30, 0, 0, 0, 100),
	}
	health := healthAll([]string{"node-a", "node-b", "node-c"}, HealthHealthy)

	d := eng.Evaluate("u-ratio", 0, aggs, health)
	if len(d.Candidates) != 1 {
		t.Fatalf("CandidateRatio=0.85 must exclude scores below 85%% of best; candidates=%+v", d.Candidates)
	}
	if d.Candidates[0].NodeID != "node-a" {
		t.Fatalf("best node must remain in candidate set: %+v", d.Candidates)
	}
}

func TestEvaluateCandidateRatioKeepsBestWhenScoresAreNonPositive(t *testing.T) {
	cfg := DefaultConfig()
	cfg.CandidateRatio = 0.85
	cfg.WeightJitter = 2
	cfg.WeightRetrans = 2
	cfg.WeightLoad = 2
	eng, err := NewEngine(cfg)
	if err != nil {
		t.Fatal(err)
	}

	aggs := []Aggregate{
		agg("u-negative", "node-a", 10, 10, 1, 1, 100),
		agg("u-negative", "node-b", 20, 10, 1, 1, 100),
	}
	health := healthAll([]string{"node-a", "node-b"}, HealthHealthy)

	d := eng.Evaluate("u-negative", 0, aggs, health)
	if len(d.Candidates) == 0 {
		t.Fatalf("candidate ratio filtering must never drop the best node when the score domain is non-positive")
	}
	if d.Candidates[0].NodeID != "node-a" {
		t.Fatalf("best node must remain first in candidate set: %+v", d.Candidates)
	}
}
