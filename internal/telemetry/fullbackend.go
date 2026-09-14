package telemetry

import (
	"database/sql"
	"time"

	"github.com/DashSaman/PV-NaivePanel/internal/steering"
)

// FullBackend couples exact accounting (PostgresStore) with R1 network
// telemetry (NetworkSampleBackend) so a single Unix-socket handler serves both
// boundaries. The embedding never widens table privileges: every method still
// goes through the same SECURITY DEFINER functions.
type FullBackend struct {
	*PostgresStore
	*NetworkSampleBackend
}

func NewFullBackend(db *sql.DB, cfg NetworkAggConfig) (*FullBackend, error) {
	store, err := NewPostgresStore(db)
	if err != nil {
		return nil, err
	}
	if cfg.Alpha <= 0 || cfg.Alpha > 1 || cfg.MinSamples < 1 {
		cfg = DefaultNetworkAggConfig()
	}
	network, err := NewNetworkSampleBackend(store, NewEWMAHolder(cfg))
	if err != nil {
		return nil, err
	}
	return &FullBackend{PostgresStore: store, NetworkSampleBackend: network}, nil
}

// SteeringStaleAfter is the default freshness window for steering consumption:
// aggregates older than this are honestly Unknown for scoring.
const SteeringStaleAfter = 4 * time.Hour

// SteeringAggregate maps an R1 aggregate onto the R2 engine input. RTT is the
// EWMA (the engine treats it as the latency input for ranking); LoadRatio is
// NOT a network-telemetry quantity — steering supplies it from node health and
// it stays zero here, never fabricated.
func (a NetworkAggregate) SteeringAggregate() steering.Aggregate {
	return steering.Aggregate{
		UserID:       a.UserID,
		NodeID:       a.NodeID,
		RTTMedianMs:  float64(a.RTTEWMAUs) / 1000.0,
		JitterMs:     float64(a.JitterEWMAUs) / 1000.0,
		RetransRatio: a.RetransRatioEWMA,
		LoadRatio:    0,
		SuccessRate:  a.SuccessRateEWMA,
		SampleCount:  a.SampleCount,
	}
}

// SteeringEligible reports whether an aggregate may be scored by the R2
// engine: enough samples and still fresh. Unknown is honest, never guessed.
func SteeringEligible(agg NetworkAggregate, minSamples int64, staleAfter time.Duration, now time.Time) bool {
	if minSamples < 1 || agg.SampleCount < minSamples || agg.LastSampledAt.IsZero() {
		return false
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	return !agg.LastSampledAt.Before(now.Add(-staleAfter))
}
