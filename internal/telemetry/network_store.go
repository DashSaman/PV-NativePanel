package telemetry

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

// Trusted-boundary persistence for R1 network telemetry. Table privileges stay
// with the SECURITY DEFINER functions; pvnaive_app never touches the tables.

type NetworkIngestor interface {
	IngestNetworkSamples(context.Context, []NetworkSample) ([]NetworkIngestResult, error)
}

type NetworkAggReader interface {
	ReadNetworkAggregates(context.Context, time.Duration) ([]NetworkAggregate, error)
}

const maxNetworkSampleBatch = 512

// IngestNetworkSamples ingests a batch in a single transaction. DB errors are
// fail-closed (the whole batch is rejected and the caller retries the exact
// same batch — replay-safe by the identity index). Individual duplicates are
// reported per sample and are not errors.
func (s *PostgresStore) IngestNetworkSamples(ctx context.Context, samples []NetworkSample) ([]NetworkIngestResult, error) {
	if s == nil || s.db == nil {
		return nil, errors.New("telemetry: PostgreSQL is required")
	}
	if len(samples) == 0 {
		return nil, ErrInvalidNetworkSample
	}
	if len(samples) > maxNetworkSampleBatch {
		return nil, fmt.Errorf("telemetry: network batch too large: %d > %d", len(samples), maxNetworkSampleBatch)
	}
	for _, sample := range samples {
		if err := ValidateNetworkSample(sample); err != nil {
			return nil, err
		}
	}
	// Deterministic ingest order (stable replay): time, then session, then seq.
	ordered := make([]NetworkSample, len(samples))
	copy(ordered, samples)
	sort.SliceStable(ordered, func(i, j int) bool {
		if !ordered[i].SampledAt.Equal(ordered[j].SampledAt) {
			return ordered[i].SampledAt.Before(ordered[j].SampledAt)
		}
		if ordered[i].SessionID != ordered[j].SessionID {
			return ordered[i].SessionID < ordered[j].SessionID
		}
		return ordered[i].SampleSeq < ordered[j].SampleSeq
	})

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("telemetry: network batch begin: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	results := make([]NetworkIngestResult, len(ordered))
	for i, sample := range ordered {
		var out NetworkIngestResult
		err := tx.QueryRowContext(ctx, `
SELECT tracked, accepted, duplicate, reason, COALESCE(user_id::text, '')
FROM pvnaive.network_sample_ingest(
    $1::uuid, $2, $3::uuid, $4::uuid, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
)`, sample.RuntimeCredentialID, sample.NodeID, sample.BootID, sample.SessionID,
			sample.SampleSeq, sample.SampledAt.UTC(), sample.Path, sample.RTTMicros,
			sample.RTTVarMicros, sample.SegsOut, sample.SegsRetrans, sample.BytesIn,
			sample.BytesOut, sample.DurationMicros).Scan(
			&out.Tracked, &out.Accepted, &out.Duplicate, &out.Reason, &out.UserID)
		if err != nil {
			return nil, fmt.Errorf("telemetry: network sample ingest: %w", err)
		}
		results[i] = out
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("telemetry: network batch commit: %w", err)
	}
	return results, nil
}

// UpsertNetworkAggregates persists holder EWMA state. One row per aggregate;
// the SQL upsert enforces the monotonic last_sampled_at guard.
func (s *PostgresStore) UpsertNetworkAggregates(ctx context.Context, aggs []NetworkAggregate) error {
	if s == nil || s.db == nil {
		return errors.New("telemetry: PostgreSQL is required")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("telemetry: network agg begin: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	for _, agg := range aggs {
		if agg.UserID == "" || agg.NodeID == "" || agg.LastSampledAt.IsZero() {
			return ErrInvalidNetworkSample
		}
		if _, err := tx.ExecContext(ctx, `
SELECT pvnaive.network_agg_upsert(
    $1::uuid, $2, $3, $4, $5, $6::real, $7, $8::real, $9, $10
)`, agg.UserID, agg.NodeID, agg.Path, agg.RTTEWMAUs, agg.JitterEWMAUs,
			agg.RetransRatioEWMA, agg.ThroughputBps, agg.SuccessRateEWMA,
			agg.SampleCount, agg.LastSampledAt.UTC()); err != nil {
			return fmt.Errorf("telemetry: network agg upsert: %w", err)
		}
	}
	return tx.Commit()
}

// ReadNetworkAggregates returns fresh aggregates only. Stale rows are omitted
// (honest Unknown), never rewritten.
func (s *PostgresStore) ReadNetworkAggregates(ctx context.Context, staleAfter time.Duration) ([]NetworkAggregate, error) {
	if s == nil || s.db == nil || staleAfter <= 0 {
		return nil, ErrInvalidProjection
	}
	seconds := int64(staleAfter / time.Second)
	if seconds < 1 {
		seconds = 1
	}
	rows, err := s.db.QueryContext(ctx, `
SELECT user_id::text, node_id, path, rtt_ewma_micros, jitter_ewma_micros,
       retrans_ratio_ewma::float8, throughput_bps, success_rate_ewma::float8,
       sample_count, last_sampled_at
FROM pvnaive.network_agg_read($1)`, seconds)
	if err != nil {
		return nil, fmt.Errorf("telemetry: network agg read: %w", err)
	}
	defer rows.Close()
	var out []NetworkAggregate
	for rows.Next() {
		var agg NetworkAggregate
		if err := rows.Scan(&agg.UserID, &agg.NodeID, &agg.Path, &agg.RTTEWMAUs,
			&agg.JitterEWMAUs, &agg.RetransRatioEWMA, &agg.ThroughputBps,
			&agg.SuccessRateEWMA, &agg.SampleCount, &agg.LastSampledAt); err != nil {
			return nil, fmt.Errorf("telemetry: network agg scan: %w", err)
		}
		agg.LastSampledAt = agg.LastSampledAt.UTC()
		out = append(out, agg)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("telemetry: network agg rows: %w", err)
	}
	return out, nil
}

// NetworkSampleRequest is the socket payload: one batch from the proxy.
type NetworkSampleRequest struct {
	Samples []NetworkSample `json:"samples"`
}

type NetworkSampleResult struct {
	Accepted   int    `json:"accepted"`
	Duplicates int    `json:"duplicates"`
	Rejected   bool   `json:"rejected"`
	Reason     string `json:"reason,omitempty"`
}

// NetworkSampleBackend pairs the store with the EWMA holder. It implements the
// socket endpoint semantics: validate strictly, ingest transactionally,
// aggregate deterministically, persist aggregates, and report honest counters.
// Ingest failures are fail-closed for telemetry (rejected=true) — but they never
// affect the proxy data path (the proxy treats non-2xx as drop-and-continue).
type NetworkSampleBackend struct {
	store  *PostgresStore
	holder *EWMAHolder
	mu     sync.Mutex
}

func NewNetworkSampleBackend(store *PostgresStore, holder *EWMAHolder) (*NetworkSampleBackend, error) {
	if store == nil || holder == nil {
		return nil, errors.New("telemetry: network backend requires store and holder")
	}
	return &NetworkSampleBackend{store: store, holder: holder}, nil
}

// Ingest is the socket-facing entry point.
func (b *NetworkSampleBackend) IngestNetworkSample(ctx context.Context, request NetworkSampleRequest) (NetworkSampleResult, error) {
	if b == nil || b.store == nil {
		return NetworkSampleResult{Rejected: true, Reason: "unavailable"}, errors.New("telemetry: network backend unavailable")
	}
	if len(request.Samples) == 0 {
		return NetworkSampleResult{Rejected: true, Reason: "empty_batch"}, ErrInvalidNetworkSample
	}
	if len(request.Samples) > maxNetworkSampleBatch {
		return NetworkSampleResult{Rejected: true, Reason: "batch_too_large"}, ErrInvalidNetworkSample
	}
	for _, sample := range request.Samples {
		if err := ValidateNetworkSample(sample); err != nil {
			return NetworkSampleResult{Rejected: true, Reason: "invalid_sample"}, err
		}
	}

	results, err := b.store.IngestNetworkSamples(ctx, request.Samples)
	if err != nil {
		return NetworkSampleResult{Rejected: true, Reason: "ingest_failed"}, err
	}

	accepted, duplicates := 0, 0
	for _, result := range results {
		switch {
		case result.Accepted:
			accepted++
		case result.Duplicate:
			duplicates++
		}
	}
	if accepted > 0 {
		if err := b.applyToHolder(request.Samples, results); err != nil {
			// Aggregation failure must not fail the already-committed ingest;
			// the next restart reseeds from persisted aggregates.
			return NetworkSampleResult{Accepted: accepted, Duplicates: duplicates}, nil
		}
		if err := b.persistAggregates(ctx); err != nil {
			return NetworkSampleResult{Accepted: accepted, Duplicates: duplicates}, nil
		}
	}
	return NetworkSampleResult{Accepted: accepted, Duplicates: duplicates}, nil
}

func (b *NetworkSampleBackend) applyToHolder(samples []NetworkSample, results []NetworkIngestResult) error {
	// Group accepted (non-duplicate) samples by (user, node, path) preserving
	// the store's deterministic ingest order.
	series := map[aggKey][]NetworkSample{}
	for i, result := range results {
		if !result.Accepted || result.UserID == "" {
			continue
		}
		sample := samples[i]
		key := aggKey{userID: result.UserID, nodeID: sample.NodeID, path: sample.Path}
		series[key] = append(series[key], sample)
	}
	if len(series) == 0 {
		return nil
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	for key, group := range series {
		sort.SliceStable(group, func(i, j int) bool {
			if !group[i].SampledAt.Equal(group[j].SampledAt) {
				return group[i].SampledAt.Before(group[j].SampledAt)
			}
			return group[i].SampleSeq < group[j].SampleSeq
		})
		b.holder.applySeries(key, group)
	}
	return nil
}

func (b *NetworkSampleBackend) persistAggregates(ctx context.Context) error {
	b.mu.Lock()
	pending := b.holder.persistables()
	b.mu.Unlock()
	if len(pending) == 0 {
		return nil
	}
	return b.store.UpsertNetworkAggregates(ctx, pending)
}

// ReadAggregates exposes holder state to steering consumers.
func (b *NetworkSampleBackend) ReadAggregates() []NetworkAggregate {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.holder.Snapshot()
}

// RestoreAggregates seeds the holder from persisted rows (restart safety).
func (b *NetworkSampleBackend) RestoreAggregates(ctx context.Context, staleAfter time.Duration) error {
	aggs, err := b.store.ReadNetworkAggregates(ctx, staleAfter)
	if err != nil {
		return err
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	for _, agg := range aggs {
		b.holder.Seed(agg)
	}
	return nil
}

// socket handler -----------------------------------------------------------------

const TelemetryNetworkSamplePath = "/v1/accounting/network-sample"

type networkSampleIngestor interface {
	IngestNetworkSample(context.Context, NetworkSampleRequest) (NetworkSampleResult, error)
}

func (h *telemetryHandler) handleNetworkSample(w http.ResponseWriter, r *http.Request) {
	backend, ok := h.backend.(networkSampleIngestor)
	if !ok || backend == nil {
		writeTelemetryError(w, http.StatusServiceUnavailable, "network telemetry unavailable")
		return
	}
	defer r.Body.Close()
	decoder := json.NewDecoder(io.LimitReader(r.Body, maxTelemetryRequestBytes+1))
	decoder.DisallowUnknownFields()
	var request NetworkSampleRequest
	if err := decoder.Decode(&request); err != nil {
		writeTelemetryError(w, http.StatusBadRequest, "invalid network sample request")
		return
	}
	result, err := backend.IngestNetworkSample(r.Context(), request)
	if err != nil {
		status := http.StatusConflict
		if strings.Contains(result.Reason, "invalid") || result.Reason == "empty_batch" || result.Reason == "batch_too_large" {
			status = http.StatusBadRequest
		}
		writeTelemetryError(w, status, "network sample rejected")
		return
	}
	writeTelemetryJSON(w, http.StatusOK, result)
}
