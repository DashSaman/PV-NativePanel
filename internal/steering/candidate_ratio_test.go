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

func TestCandidateSetIncludesExactThresholdAndExcludesBelow(t *testing.T) {
	got := candidateSet([]Scored{
		{NodeID: "best", Score: 100, Kind: "eligible"},
		{NodeID: "threshold", Score: 85, Kind: "eligible"},
		{NodeID: "below", Score: 84.999, Kind: "eligible"},
	}, 0.85)

	if len(got) != 2 {
		t.Fatalf("exact threshold must be included and below-threshold excluded: %+v", got)
	}
	if got[0].NodeID != "best" || got[1].NodeID != "threshold" {
		t.Fatalf("candidate order/boundary mismatch: %+v", got)
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

func TestTopKUsesRatioFilteredDecisionCandidates(t *testing.T) {
	cfg := DefaultConfig()
	cfg.CandidateRatio = 0.85
	eng, err := NewEngine(cfg)
	if err != nil {
		t.Fatal(err)
	}

	aggs := []Aggregate{
		agg("u-topk", "node-a", 10, 0, 0, 0, 100),
		agg("u-topk", "node-b", 20, 0, 0, 0, 100),
		agg("u-topk", "node-c", 30, 0, 0, 0, 100),
	}
	health := healthAll([]string{"node-a", "node-b", "node-c"}, HealthHealthy)

	d := eng.Evaluate("u-topk", 0, aggs, health)
	got := TopK(d.Candidates, cfg.TopKMobile)
	if len(got) != len(d.Candidates) {
		t.Fatalf("TopK renderer input must be the ratio-filtered decision candidates: got=%+v candidates=%+v", got, d.Candidates)
	}
	for i := range got {
		if got[i] != d.Candidates[i] {
			t.Fatalf("TopK must preserve ratio-filtered decision order: got=%+v candidates=%+v", got, d.Candidates)
		}
	}
}
