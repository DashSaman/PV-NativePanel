package telemetry

import (
	"errors"
	"fmt"
	"time"
)

// R1 / STEER-001 trusted-boundary network telemetry (docs/STEERING_SPEC_FA.md §1).
//
// Samples arrive from the pinned forwardproxy through the same trusted Unix
// socket as exact accounting, keyed by runtime_credential_id. The database
// resolves user_id via the exact-accounting join — client headers are never
// trusted. TCP_INFO counters are cumulative per TCP connection, so rate
// features (retrans ratio, throughput, jitter) are computed from deltas and
// ONLY within a continuous (same session, consecutive sequence) span. A span
// start contributes the RTT reading but no rate features — the EWMA holder
// never fabricates them.

var (
	ErrInvalidNetworkSample = errors.New("telemetry: invalid network sample")
)

const (
	NetworkPathClient   = "client"
	NetworkPathUpstream = "upstream"
)

// NetworkSample is one raw TCP_INFO reading for one proxy session.
// Segs/bytes/duration counters are cumulative since the connection start.
type NetworkSample struct {
	RuntimeCredentialID string    `json:"runtime_credential_id"`
	NodeID              string    `json:"node_id"`
	BootID              string    `json:"boot_id"`
	SessionID           string    `json:"session_id"`
	SampleSeq           int64     `json:"sample_seq"`
	SampledAt           time.Time `json:"sampled_at"`
	Path                string    `json:"path"`
	RTTMicros           int64     `json:"rtt_micros"`
	RTTVarMicros        int64     `json:"rtt_var_micros"`
	SegsOut             int64     `json:"segs_out"`
	SegsRetrans         int64     `json:"segs_retrans"`
	BytesIn             int64     `json:"bytes_in"`
	BytesOut            int64     `json:"bytes_out"`
	DurationMicros      int64     `json:"duration_micros"`
}

// NetworkIngestResult reports the DB decision for one sample.
type NetworkIngestResult struct {
	Tracked   bool
	Accepted  bool
	Duplicate bool
	Reason    string
	UserID    string
}

// NetworkAggregate is a holder-computed EWMA aggregate for (user, node, path).
type NetworkAggregate struct {
	UserID           string    `json:"user_id"`
	NodeID           string    `json:"node_id"`
	Path             string    `json:"path"`
	RTTEWMAUs        int64     `json:"rtt_ewma_micros"`
	JitterEWMAUs     int64     `json:"jitter_ewma_micros"`
	RetransRatioEWMA float64   `json:"retrans_ratio_ewma"`
	ThroughputBps    int64     `json:"throughput_bps"`
	SuccessRateEWMA  float64   `json:"success_rate_ewma"`
	SampleCount      int64     `json:"sample_count"`
	LastSampledAt    time.Time `json:"last_sampled_at"`
}

// ValidateNetworkSample enforces the trusted-boundary contract before the
// sample ever reaches the database.
func ValidateNetworkSample(sample NetworkSample) error {
	if !validUUID(sample.RuntimeCredentialID) ||
		!validUUID(sample.BootID) ||
		!validUUID(sample.SessionID) ||
		!validDiagnostic(sample.NodeID) ||
		sample.SampleSeq < 1 ||
		sample.SampledAt.IsZero() ||
		(sample.Path != NetworkPathClient && sample.Path != NetworkPathUpstream) ||
		sample.RTTMicros < 0 ||
		sample.RTTVarMicros < 0 ||
		sample.SegsOut < 0 ||
		sample.SegsRetrans < 0 ||
		sample.SegsRetrans > sample.SegsOut ||
		sample.BytesIn < 0 ||
		sample.BytesOut < 0 ||
		sample.DurationMicros < 0 {
		return ErrInvalidNetworkSample
	}
	return nil
}

// NetworkAggConfig carries the EWMA parameters. Defaults follow
// docs/STEERING_SPEC_FA.md §1.2 (alpha 0.2, min_samples 8).
type NetworkAggConfig struct {
	Alpha      float64
	MinSamples int64
}

func DefaultNetworkAggConfig() NetworkAggConfig {
	return NetworkAggConfig{Alpha: 0.2, MinSamples: 8}
}

// networkAggState is the EWMA holder state for one (user, node, path) row.
// It is a pure value type: Apply never mutates the receiver, so replay and
// simulation are deterministic and testable.
type networkAggState struct {
	userID    string
	nodeID    string
	path      string
	rttEWMA   float64
	jitterEWM float64
	retransEW float64
	thptBps   float64
	successEW float64
	count     int64
	lastAt    time.Time
	// last-span bookkeeping for delta features:
	spanSession  string
	spanBoot     string
	lastSeq      int64
	lastRTTUs    int64
	lastSegsOut  int64
	lastSegsRetr int64
	lastBytesIn  int64
	lastBytesOut int64
	lastDurUs    int64
	rateStarted  bool
}

type sampleFeatures struct {
	rttUs      float64
	jitterUs   float64
	retrans    float64
	thptBps    float64
	hasRate    bool
	successful bool
}

// featuresFor derives delta-based rate features for a sample. Rate features
// exist only for continuous same-session spans (consecutive sample_seq in the
// same boot/session). This is the anti-fabrication rule: a fresh connection
// cannot produce a meaningful delta, so none is invented.
func (s *networkAggState) featuresFor(sample NetworkSample) sampleFeatures {
	f := sampleFeatures{
		rttUs:      float64(sample.RTTMicros),
		successful: true,
	}
	continuous := s.spanSession == sample.SessionID &&
		s.spanBoot == sample.BootID &&
		s.lastSeq == sample.SampleSeq-1 &&
		!s.lastAt.IsZero()
	if continuous {
		dSegs := sample.SegsOut - s.lastSegsOut
		dRetr := sample.SegsRetrans - s.lastSegsRetr
		dBytes := (sample.BytesIn - s.lastBytesIn) + (sample.BytesOut - s.lastBytesOut)
		dDurUs := sample.DurationMicros - s.lastDurUs
		dRTT := sample.RTTMicros - s.lastRTTUs
		if dSegs > 0 && dRetr >= 0 && dRetr <= dSegs {
			f.retrans = float64(dRetr) / float64(dSegs)
			f.hasRate = true
		}
		if dDurUs > 0 && dBytes >= 0 {
			f.thptBps = float64(dBytes*1_000_000) / float64(dDurUs)
			f.hasRate = true
		}
		if dRTT >= 0 || dRTT < 0 {
			f.jitterUs = absFloat(float64(dRTT))
			f.hasRate = true
		}
	}
	return f
}

func (s *networkAggState) applySample(sample NetworkSample, f sampleFeatures, alpha float64) networkAggState {
	next := *s
	if s.count == 0 {
		next.rttEWMA = f.rttUs
		next.successEW = bool01(f.successful)
	} else {
		next.rttEWMA = ewmaMix(alpha, f.rttUs, s.rttEWMA)
		next.successEW = ewmaMix(alpha, bool01(f.successful), s.successEW)
	}
	// Rate features (jitter/retrans/throughput) exist only for continuous
	// spans. Until the first real delta arrives the EWMA keeps "no data" —
	// mixing in fabricated zeros would bias quality downward.
	if f.hasRate {
		if !s.rateStarted {
			next.jitterEWM = f.jitterUs
			next.retransEW = f.retrans
			next.thptBps = f.thptBps
			next.rateStarted = true
		} else {
			next.jitterEWM = ewmaMix(alpha, f.jitterUs, s.jitterEWM)
			next.retransEW = ewmaMix(alpha, f.retrans, s.retransEW)
			next.thptBps = ewmaMix(alpha, f.thptBps, s.thptBps)
		}
	}
	next.count = s.count + 1
	next.lastAt = sample.SampledAt
	next.spanSession = sample.SessionID
	next.spanBoot = sample.BootID
	next.lastSeq = sample.SampleSeq
	next.lastRTTUs = sample.RTTMicros
	next.lastSegsOut = sample.SegsOut
	next.lastSegsRetr = sample.SegsRetrans
	next.lastBytesIn = sample.BytesIn
	next.lastBytesOut = sample.BytesOut
	next.lastDurUs = sample.DurationMicros
	return next
}

func (s networkAggState) aggregate() NetworkAggregate {
	return NetworkAggregate{
		UserID:           s.userID,
		NodeID:           s.nodeID,
		Path:             s.path,
		RTTEWMAUs:        int64(s.rttEWMA + 0.5),
		JitterEWMAUs:     int64(s.jitterEWM + 0.5),
		RetransRatioEWMA: s.retransEW,
		ThroughputBps:    int64(s.thptBps + 0.5),
		SuccessRateEWMA:  s.successEW,
		SampleCount:      s.count,
		LastSampledAt:    s.lastAt,
	}
}

// EWMAHolder maintains per (user, node, path) EWMA state from ordered sample
// streams and produces aggregates for steering. MinSamples semantics: a row
// below MinSamples stays Unknown — it is returned by Read for observation but
// flagged Unknown() so the R2 engine excludes it from scoring (never guessed).
type EWMAHolder struct {
	cfg   NetworkAggConfig
	state map[aggKey]*networkAggState
}

type aggKey struct {
	userID string
	nodeID string
	path   string
}

func NewEWMAHolder(cfg NetworkAggConfig) *EWMAHolder {
	if cfg.Alpha <= 0 || cfg.Alpha > 1 {
		cfg = DefaultNetworkAggConfig()
	}
	if cfg.MinSamples < 1 {
		cfg.MinSamples = DefaultNetworkAggConfig().MinSamples
	}
	return &EWMAHolder{cfg: cfg, state: map[aggKey]*networkAggState{}}
}

// Apply ingests an ordered sample stream for one (user, node, path) series and
// returns the updated aggregate. Samples must be pre-sorted by SampledAt then
// SampleSeq; the holder never reorders (deterministic replay).
func (h *EWMAHolder) Apply(userID, nodeID string, samples []NetworkSample) (NetworkAggregate, bool) {
	key := aggKey{userID: userID, nodeID: nodeID, path: samplePath(samples)}
	st, ok := h.state[key]
	if !ok {
		st = &networkAggState{userID: userID, nodeID: nodeID, path: key.path}
		h.state[key] = st
	}
	for _, sample := range samples {
		features := st.featuresFor(sample)
		next := st.applySample(sample, features, h.cfg.Alpha)
		*st = next
	}
	agg := st.aggregate()
	return agg, agg.SampleCount >= h.cfg.MinSamples
}

// applySeries ingests a pre-grouped, ordered sample series for one exact
// (user, node, path) key. Called by the socket backend after DB commit.
func (h *EWMAHolder) applySeries(key aggKey, samples []NetworkSample) {
	st, ok := h.state[key]
	if !ok {
		st = &networkAggState{userID: key.userID, nodeID: key.nodeID, path: key.path}
		h.state[key] = st
	}
	for _, sample := range samples {
		features := st.featuresFor(sample)
		*st = st.applySample(sample, features, h.cfg.Alpha)
	}
}

// Seed restores holder state from a previously persisted aggregate so restart
// does not reset EWMA memory (STEER-001 restart safety for aggregates).
func (h *EWMAHolder) Seed(agg NetworkAggregate) {
	key := aggKey{userID: agg.UserID, nodeID: agg.NodeID, path: agg.Path}
	if agg.LastSampledAt.IsZero() || agg.SampleCount <= 0 {
		return
	}
	if existing, ok := h.state[key]; ok && !existing.lastAt.Before(agg.LastSampledAt) {
		return
	}
	h.state[key] = &networkAggState{
		userID:      agg.UserID,
		nodeID:      agg.NodeID,
		path:        agg.Path,
		rttEWMA:     float64(agg.RTTEWMAUs),
		jitterEWM:   float64(agg.JitterEWMAUs),
		retransEW:   agg.RetransRatioEWMA,
		thptBps:     float64(agg.ThroughputBps),
		successEW:   agg.SuccessRateEWMA,
		count:       agg.SampleCount,
		lastAt:      agg.LastSampledAt,
		rateStarted: agg.JitterEWMAUs > 0 || agg.RetransRatioEWMA > 0 || agg.ThroughputBps > 0,
	}
}

// Snapshot returns every current aggregate. Unknown marks rows below
// MinSamples — honest observation without scoring eligibility.
func (h *EWMAHolder) Snapshot() []NetworkAggregate {
	out := make([]NetworkAggregate, 0, len(h.state))
	for _, st := range h.state {
		out = append(out, st.aggregate())
	}
	sortAggregates(out)
	return out
}

func (h *EWMAHolder) persistables() []NetworkAggregate {
	out := make([]NetworkAggregate, 0, len(h.state))
	for _, st := range h.state {
		if st.count > 0 && !st.lastAt.IsZero() {
			out = append(out, st.aggregate())
		}
	}
	sortAggregates(out)
	return out
}

func sortAggregates(aggs []NetworkAggregate) {
	// Deterministic order for tests and stable upsert batches.
	for i := 1; i < len(aggs); i++ {
		for j := i; j > 0 && aggLess(aggs[j], aggs[j-1]); j-- {
			aggs[j], aggs[j-1] = aggs[j-1], aggs[j]
		}
	}
}

func aggLess(a, b NetworkAggregate) bool {
	if a.UserID != b.UserID {
		return a.UserID < b.UserID
	}
	if a.NodeID != b.NodeID {
		return a.NodeID < b.NodeID
	}
	return a.Path < b.Path
}

func samplePath(samples []NetworkSample) string {
	if len(samples) == 0 {
		return ""
	}
	return samples[0].Path
}

func ewmaMix(alpha, sample, old float64) float64 {
	return alpha*sample + (1-alpha)*old
}

func bool01(v bool) float64 {
	if v {
		return 1
	}
	return 0
}

func absFloat(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}

// String keeps logs redacted: node/path only, never credential material.
func (s NetworkSample) String() string {
	return fmt.Sprintf("network_sample{node=%s path=%s seq=%d}", s.NodeID, s.Path, s.SampleSeq)
}
