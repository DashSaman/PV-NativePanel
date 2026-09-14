package httpapi

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// The system stream (R8 live charts) pushes the same metrics envelope as the
// polling endpoint over Server-Sent Events so the dashboard can render ~1s
// updates without hammering the poll route. Design invariants:
//   - the auth transaction is committed BEFORE the first byte (a stream must
//     never pin a database transaction for its lifetime);
//   - the connection carries no tenant data beyond what the Operator-gated
//     poll endpoint already exposes;
//   - tick intervals are clamped to a bounded range (default 1s);
//   - concurrent streams are capped so a leaked client cannot exhaust slots.
const (
	systemStreamDefaultInterval = time.Second
	systemStreamMinInterval     = 100 * time.Millisecond
	systemStreamMaxInterval     = 30 * time.Second
	systemStreamMaxClients      = 64
	systemStreamReconnectHintMS = 3000
)

// resolveStreamInterval parses the client-requested tick interval. Bare
// integers are interpreted as milliseconds; durations use time.ParseDuration.
// Anything invalid, negative, or out of range is clamped or defaulted.
func resolveStreamInterval(raw string) time.Duration {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return systemStreamDefaultInterval
	}
	var requested time.Duration
	if ms, err := strconv.Atoi(raw); err == nil && ms > 0 {
		requested = time.Duration(ms) * time.Millisecond
	} else if parsed, err := time.ParseDuration(raw); err == nil && parsed > 0 {
		requested = parsed
	} else {
		return systemStreamDefaultInterval
	}
	if requested < systemStreamMinInterval {
		return systemStreamMinInterval
	}
	if requested > systemStreamMaxInterval {
		return systemStreamMaxInterval
	}
	return requested
}

// acquireStreamSlot reserves one of the bounded concurrent stream slots.
// Production servers are built by NewServer, which pre-allocates the slot
// channel; the lazy init keeps directly-constructed test servers safe.
func (s *server) acquireStreamSlot() bool {
	if s.streams == nil {
		s.streams = make(chan struct{}, systemStreamMaxClients)
	}
	select {
	case s.streams <- struct{}{}:
		return true
	default:
		return false
	}
}

func (s *server) releaseStreamSlot() {
	if s.streams != nil {
		<-s.streams
	}
}

// flushSSE flushes the response through either the auth wrapper's buffer
// (streaming passthrough) or a plain Flusher. It deliberately does not rely
// on *responseBuffer satisfying http.Flusher (BUG-002 contract).
func flushSSE(w http.ResponseWriter) {
	if buf, ok := w.(*responseBuffer); ok {
		buf.flushToClient()
		return
	}
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (s *server) systemStream(w http.ResponseWriter, r *http.Request) {
	if s.config.SystemStatus == nil {
		writeJSON(w, http.StatusServiceUnavailable, envelope{"code": "system_metrics_unavailable", "message": "System metrics are unavailable."})
		return
	}
	authenticated, ok := authenticatedFromRequest(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, envelope{"code": "authentication_required", "message": "Authentication is required."})
		return
	}
	// Release the auth read transaction before streaming: the connection can
	// stay open for minutes and must never pin a database transaction.
	if authenticated.Bound.Tx != nil {
		if err := authenticated.Bound.Tx.Commit(); err != nil {
			authenticated.TransactionFinalized = true
			writeJSON(w, http.StatusInternalServerError, envelope{"code": "transaction_commit_failed", "message": "The operation could not be completed because the database transaction did not commit."})
			return
		}
	}
	authenticated.TransactionFinalized = true

	if !s.acquireStreamSlot() {
		writeJSON(w, http.StatusServiceUnavailable, envelope{"code": "stream_limit_reached", "message": "Too many live streams are open. Try again later."})
		return
	}
	defer s.releaseStreamSlot()

	// Escape the auth wrapper's response buffer: frames must reach the client
	// as they are written instead of being buffered until the handler returns.
	if buf, ok := w.(*responseBuffer); ok {
		buf.enableStreaming()
	}

	interval := resolveStreamInterval(r.URL.Query().Get("interval"))

	header := w.Header()
	header.Set("Content-Type", "text/event-stream; charset=utf-8")
	header.Set("Cache-Control", "no-cache")
	header.Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)

	writeFrame := func(event string, payload envelope) {
		data, err := json.Marshal(payload)
		if err != nil {
			return
		}
		_, _ = w.Write([]byte("event: " + event + "\ndata: " + string(data) + "\n\n"))
		flushSSE(w)
	}

	_, _ = w.Write([]byte("retry: " + strconv.Itoa(systemStreamReconnectHintMS) + "\n"))
	flushSSE(w)

	emitSnapshot := func() {
		snapshot, err := s.config.SystemStatus(r)
		if err != nil {
			writeFrame("error", envelope{"code": "system_metrics_unavailable", "message": "System metrics are unavailable."})
			return
		}
		writeFrame("status", envelope{"metrics": snapshot, "traffic_semantics": "server_counter_delta"})
	}

	// First frame is synchronous so the dashboard paints without waiting a tick.
	emitSnapshot()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			emitSnapshot()
		}
	}
}
