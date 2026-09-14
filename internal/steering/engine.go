// Package steering implements the R2 scoring and steering engine for the
// Master Upgrade Pack: deterministic per-user best-node selection from
// server-side network aggregates (R1), time-windowed rotation with per-user
// phase offsets, hysteresis, kill-switch failover and an honest Unknown policy.
//
// Design rules (docs/STEERING_SPEC_FA.md):
//   - every parameter comes from Config, nothing hardcoded in logic;
//   - aggregates with fewer than MinSamples samples are Unknown and are
//     excluded from scoring — never fabricated, never default-guessed;
//   - primary switches only at window boundaries, only with hysteresis
//     (>= margin better for >= HysteresisWindows consecutive windows), except
//     on node health failure (kill-switch: immediate failover);
//   - all decisions are pure functions of (state, inputs) so simulation tests
//     are deterministic (gate STEER-002).
package steering

import (
	"errors"
	"fmt"
	"hash/fnv"
	"sort"
	"time"
)

// Aggregate is the per (user, node) network aggregate produced by R1
// (EWMA-smoothed). LoadRatio and SuccessRate are in [0,1]; RetransRatio in
// [0,1]. SampleCount is the number of contributing samples.
type Aggregate struct {
	UserID       string
	NodeID       string
	RTTMedianMs  float64
	JitterMs     float64
	RetransRatio float64
	LoadRatio    float64
	SuccessRate  float64
	SampleCount  int64
}

// NodeHealth mirrors fleet node health as seen by the steering scheduler.
type NodeHealth string

const (
	HealthHealthy  NodeHealth = "healthy"
	HealthDegraded NodeHealth = "degraded"
	HealthOffline  NodeHealth = "offline"
)

// Reason explains why a decision produced its outcome (audit log payload).
type Reason string

const (
	ReasonInitial      Reason = "initial"       // first assignment for the user
	ReasonNoChange     Reason = "no_change"     // primary retained
	ReasonHysteresis   Reason = "hysteresis"    // margin met for N consecutive windows
	ReasonKillSwitch   Reason = "kill_switch"   // primary unhealthy -> immediate failover
	ReasonNoCandidates Reason = "no_candidates" // no eligible node (all Unknown/unhealthy)
)

// Scored is a node with its computed score (higher is better).
type Scored struct {
	NodeID string
	Score  float64
	Kind   string // "eligible" | "unknown"
}

// Config carries every engine parameter. Defaults live in DefaultConfig and
// are expected to be loaded from operator configuration, not hardcoded here.
type Config struct {
	Window            time.Duration // rotation window (default 4h)
	Grace             time.Duration // overlap validity of the previous node (default 15m)
	Alpha             float64       // EWMA smoothing factor used by R1 ingest (default 0.2)
	CandidateRatio    float64       // candidates: score >= best*CandidateRatio (default 0.85)
	HysteresisMargin  float64       // switch if candidate >= primary*(1+margin) (default 0.25)
	HysteresisWindows int           // consecutive windows required (default 2)
	MinSamples        int64         // aggregates below this are Unknown (default 8)
	WeightJitter      float64       // jitter penalty weight (default 0.2)
	WeightRetrans     float64       // retransmission penalty weight (default 0.3)
	WeightLoad        float64       // node load penalty weight (default 0.2)
	TopKMobile        int           // per-user subset size for mobile formats (default 10)
}

// DefaultConfig returns the documented defaults from the spec.
func DefaultConfig() Config {
	return Config{
		Window:            4 * time.Hour,
		Grace:             15 * time.Minute,
		Alpha:             0.2,
		CandidateRatio:    0.85,
		HysteresisMargin:  0.25,
		HysteresisWindows: 2,
		MinSamples:        8,
		WeightJitter:      0.2,
		WeightRetrans:     0.3,
		WeightLoad:        0.2,
		TopKMobile:        10,
	}
}

// Validate rejects configurations that would make decisions meaningless.
func (c Config) Validate() error {
	switch {
	case c.Window < 30*time.Minute || c.Window > 24*time.Hour:
		return errors.New("steering: rotation window must be between 30m and 24h")
	case c.Grace < 5*time.Minute || c.Grace > time.Hour:
		return errors.New("steering: grace must be between 5m and 1h")
	case c.Alpha < 0.05 || c.Alpha > 0.5:
		return errors.New("steering: ewma alpha must be in [0.05, 0.5]")
	case c.CandidateRatio < 0.7 || c.CandidateRatio > 0.95:
		return errors.New("steering: candidate ratio must be in [0.7, 0.95]")
	case c.HysteresisMargin < 0.1 || c.HysteresisMargin > 0.5:
		return errors.New("steering: hysteresis margin must be in [0.1, 0.5]")
	case c.HysteresisWindows < 1 || c.HysteresisWindows > 4:
		return errors.New("steering: hysteresis windows must be in [1, 4]")
	case c.MinSamples < 1:
		return errors.New("steering: minimum samples must be positive")
	case c.WeightJitter < 0 || c.WeightRetrans < 0 || c.WeightLoad < 0:
		return errors.New("steering: penalty weights must be non-negative")
	case c.TopKMobile < 3 || c.TopKMobile > 50:
		return errors.New("steering: mobile top-k must be in [3, 50]")
	}
	return nil
}

// EWMA is the exponential moving average step used by R1 ingest and kept here
// so engine simulations and the telemetry sampler share one definition.
func EWMA(prev, sample, alpha float64) float64 {
	return alpha*sample + (1-alpha)*prev
}

// PhaseOffset derives the deterministic per-user rotation phase:
// phase = fnv64a(userID) mod window. Users therefore do not rotate
// simultaneously. The zero window disables the calculation.
func PhaseOffset(userID string, window time.Duration) time.Duration {
	if window <= 0 {
		return 0
	}
	h := fnv.New64a()
	_, _ = h.Write([]byte(userID))
	return time.Duration(h.Sum64() % uint64(window))
}

// WindowIndex maps (now, user) to a stable window counter, aligned to the
// user's phase offset so boundaries differ per user.
func WindowIndex(now time.Time, userID string, window time.Duration) int64 {
	if window <= 0 {
		return 0
	}
	shifted := now.Add(-PhaseOffset(userID, window))
	return int64(shifted.UnixNano() / int64(window))
}

// Score computes deterministic scores for the given aggregates.
//
//   - aggregates with SampleCount < MinSamples are returned with Kind
//     "unknown" and are never eligible;
//   - the RTT component ranks eligible nodes linearly from ~100 (best) down;
//   - jitter penalty is relative to the worst eligible jitter;
//   - retransmission and load penalties are proportional to their ratios.
//
// Ties on RTT are broken by NodeID so results are fully deterministic.
func Score(cfg Config, aggs []Aggregate) []Scored {
	type entry struct {
		agg   Aggregate
		score float64
	}

	var elig []entry
	var unknown []Scored
	for _, a := range aggs {
		if a.SampleCount < cfg.MinSamples {
			unknown = append(unknown, Scored{NodeID: a.NodeID, Kind: "unknown"})
			continue
		}
		elig = append(elig, entry{agg: a})
	}

	sort.Slice(elig, func(i, j int) bool {
		if elig[i].agg.RTTMedianMs != elig[j].agg.RTTMedianMs {
			return elig[i].agg.RTTMedianMs < elig[j].agg.RTTMedianMs
		}
		return elig[i].agg.NodeID < elig[j].agg.NodeID
	})

	n := float64(len(elig))
	var maxJitter float64
	for i := range elig {
		if elig[i].agg.JitterMs > maxJitter {
			maxJitter = elig[i].agg.JitterMs
		}
	}

	out := make([]Scored, 0, len(aggs))
	for i := range elig {
		rttComponent := 100.0
		if n > 1 {
			rttComponent = (n - float64(i)) * (100.0 / n)
		}
		jitterPenalty := 0.0
		if maxJitter > 0 {
			jitterPenalty = cfg.WeightJitter * 100 * (elig[i].agg.JitterMs / maxJitter)
		}
		retransPenalty := cfg.WeightRetrans * 100 * clamp01(elig[i].agg.RetransRatio)
		loadPenalty := cfg.WeightLoad * 100 * clamp01(elig[i].agg.LoadRatio)
		elig[i].score = rttComponent - jitterPenalty - retransPenalty - loadPenalty
		out = append(out, Scored{NodeID: elig[i].agg.NodeID, Score: elig[i].score, Kind: "eligible"})
	}
	for _, u := range unknown {
		out = append(out, u)
	}

	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Kind != out[j].Kind {
			return out[i].Kind == "eligible"
		}
		if out[i].Kind == "eligible" && out[i].Score != out[j].Score {
			return out[i].Score > out[j].Score
		}
		return out[i].NodeID < out[j].NodeID
	})
	return out
}

// TopK returns the first k eligible entries (Unknown never selected), clipped
// to the available count. Used for per-user node subset rendering (R5).
func TopK(scored []Scored, k int) []Scored {
	if k <= 0 {
		return nil
	}
	var out []Scored
	for _, s := range scored {
		if len(out) >= k {
			break
		}
		if s.Kind == "eligible" {
			out = append(out, s)
		}
	}
	return out
}

// Decision is the outcome of evaluating one rotation window for a user.
type Decision struct {
	UserID     string
	WindowIdx  int64
	Primary    string // chosen node; "" when none eligible and no previous
	Previous   string // primary before this decision ("" when none)
	Reason     Reason
	Candidates []Scored // eligible scores, ordered best-first
	Changed    bool
}

// userState tracks the hysteresis bookkeeping between windows.
type userState struct {
	primary          string
	candidateStreak  int
	pendingCandidate string
}

// Engine evaluates rotation windows per user. It is safe for use as a plain
// struct owned by one scheduler goroutine; persistence of state across
// restarts is the caller's responsibility (R4/R5 wiring).
type Engine struct {
	cfg   Config
	users map[string]*userState
}

// NewEngine returns an engine with the given validated configuration.
func NewEngine(cfg Config) (*Engine, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("steering: %w", err)
	}
	return &Engine{cfg: cfg, users: make(map[string]*userState)}, nil
}

// Config exposes the active configuration (read-only use).
func (e *Engine) Config() Config { return e.cfg }

// Primary returns the current primary node for a user ("" when unassigned).
func (e *Engine) Primary(user string) string {
	if st, ok := e.users[user]; ok {
		return st.primary
	}
	return ""
}

// candidateSet applies the rendering/subset CandidateRatio contract without
// changing the full eligible set used by primary selection and hysteresis.
// The best node is always retained; this matters when all scores are zero or
// negative because best*ratio would otherwise be greater than the best score.
func candidateSet(eligible []Scored, ratio float64) []Scored {
	if len(eligible) == 0 {
		return nil
	}
	best := eligible[0].Score
	threshold := best * ratio
	out := make([]Scored, 0, len(eligible))
	out = append(out, eligible[0])
	for _, s := range eligible[1:] {
		if s.Score >= threshold {
			out = append(out, s)
		}
	}
	return out
}

// Evaluate processes one rotation window for a user. aggs carries the latest
// per-node aggregates for the user; health carries current node health.
//
// Semantics:
//   - Unknown aggregates never win;
//   - no eligible node -> ReasonNoCandidates (fail-closed: callers keep the
//     last-known-good rendering; nothing is fabricated);
//   - first assignment and kill-switch (primary offline) bypass hysteresis;
//   - otherwise a switch needs a candidate >= primary*(1+margin) for
//     HysteresisWindows consecutive windows.
func (e *Engine) Evaluate(user string, windowIdx int64, aggs []Aggregate, health map[string]NodeHealth) Decision {
	st := e.users[user]
	if st == nil {
		st = &userState{}
		e.users[user] = st
	}

	scored := Score(e.cfg, aggs)
	var eligible []Scored
	for _, s := range scored {
		if s.Kind == "eligible" && health[s.NodeID] != HealthOffline {
			eligible = append(eligible, s)
		}
	}

	d := Decision{UserID: user, WindowIdx: windowIdx, Previous: st.primary, Candidates: candidateSet(eligible, e.cfg.CandidateRatio)}
	if len(eligible) == 0 {
		// Fail-closed: no fabricated choice. Keep the previous primary recorded
		// so callers can serve last-known-good, but report no candidates.
		d.Reason = ReasonNoCandidates
		d.Primary = st.primary
		return d
	}

	best := eligible[0]
	primScore := 0.0
	hasPrimary := st.primary != ""
	for _, s := range eligible {
		if s.NodeID == st.primary {
			primScore = s.Score
		}
	}

	switch {
	case !hasPrimary:
		st.primary = best.NodeID
		st.candidateStreak = 0
		st.pendingCandidate = ""
		d.Primary = best.NodeID
		d.Reason = ReasonInitial
		d.Changed = true
	case health[st.primary] == HealthOffline:
		// Kill-switch: immediate failover, hysteresis bypassed.
		target := best.NodeID
		if target == st.primary {
			// Primary reported offline but is still the best eligible entry
			// (health map inconsistency); do not churn.
			d.Primary = st.primary
			d.Reason = ReasonNoChange
			return d
		}
		st.primary = target
		st.candidateStreak = 0
		st.pendingCandidate = ""
		d.Primary = target
		d.Reason = ReasonKillSwitch
		d.Changed = true
	case best.NodeID == st.primary:
		st.candidateStreak = 0
		st.pendingCandidate = ""
		d.Primary = st.primary
		d.Reason = ReasonNoChange
	case primScore > 0 && best.Score >= primScore*(1+e.cfg.HysteresisMargin):
		if st.pendingCandidate == best.NodeID {
			st.candidateStreak++
		} else {
			st.pendingCandidate = best.NodeID
			st.candidateStreak = 1
		}
		if st.candidateStreak >= e.cfg.HysteresisWindows {
			st.primary = best.NodeID
			st.candidateStreak = 0
			st.pendingCandidate = ""
			d.Primary = best.NodeID
			d.Reason = ReasonHysteresis
			d.Changed = true
		} else {
			d.Primary = st.primary
			d.Reason = ReasonNoChange
		}
	default:
		st.candidateStreak = 0
		st.pendingCandidate = ""
		d.Primary = st.primary
		d.Reason = ReasonNoChange
	}
	return d
}

func clamp01(v float64) float64 {
	switch {
	case v < 0:
		return 0
	case v > 1:
		return 1
	default:
		return v
	}
}
