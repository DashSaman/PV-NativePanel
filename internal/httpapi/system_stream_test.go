package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DashSaman/PV-NaivePanel/internal/auth"
)

func TestResolveStreamIntervalBounds(t *testing.T) {
	cases := []struct {
		name string
		raw  string
		want time.Duration
	}{
		{"empty uses default", "", systemStreamDefaultInterval},
		{"plain milliseconds", "500", 500 * time.Millisecond},
		{"duration string", "2s", 2 * time.Second},
		{"invalid falls back to default", "soon", systemStreamDefaultInterval},
		{"below floor is clamped", "1ms", systemStreamMinInterval},
		{"above ceiling is clamped", "1h", systemStreamMaxInterval},
		{"negative falls back to default", "-5s", systemStreamDefaultInterval},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := resolveStreamInterval(tc.raw); got != tc.want {
				t.Fatalf("resolveStreamInterval(%q) = %v, want %v", tc.raw, got, tc.want)
			}
		})
	}
}

func TestSystemStreamRequiresAuthenticatedContext(t *testing.T) {
	s := &server{config: ServerConfig{SystemStatus: func(*http.Request) (any, error) {
		return map[string]any{"sample": map[string]any{}}, nil
	}}}
	req := httptest.NewRequest(http.MethodGet, "/api/v1/system/stream", nil)
	res := httptest.NewRecorder()
	s.systemStream(res, req)
	if res.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d body=%s", res.Code, res.Body.String())
	}
	if got := res.Header().Get("Content-Type"); !strings.Contains(got, "application/json") {
		t.Fatalf("content-type=%q, want json error envelope", got)
	}
}

func TestSystemStreamFailsClosedWithoutProvider(t *testing.T) {
	s := &server{}
	req := httptest.NewRequest(http.MethodGet, "/api/v1/system/stream", nil)
	req = withAuthenticatedRequest(req, &auth.AuthenticatedTx{
		Principal: auth.Principal{ActorID: "operator-1", Role: "operator"},
	}, "session-token")
	res := httptest.NewRecorder()
	s.systemStream(res, req)
	if res.Code != http.StatusServiceUnavailable {
		t.Fatalf("status=%d body=%s", res.Code, res.Body.String())
	}
	if got := res.Header().Get("Content-Type"); strings.Contains(got, "text/event-stream") {
		t.Fatalf("content-type=%q, want plain json failure (no stream)", got)
	}
}

func TestSystemStreamRejectsWhenSlotBudgetExhausted(t *testing.T) {
	s := &server{config: ServerConfig{SystemStatus: func(*http.Request) (any, error) {
		return map[string]any{"sample": map[string]any{}}, nil
	}}}
	s.streams = make(chan struct{}, 1)
	s.streams <- struct{}{}
	req := httptest.NewRequest(http.MethodGet, "/api/v1/system/stream", nil)
	req = withAuthenticatedRequest(req, &auth.AuthenticatedTx{
		Principal: auth.Principal{ActorID: "operator-1", Role: "operator"},
	}, "session-token")
	res := httptest.NewRecorder()
	s.systemStream(res, req)
	if res.Code != http.StatusServiceUnavailable {
		t.Fatalf("status=%d body=%s", res.Code, res.Body.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(res.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload["code"] != "stream_limit_reached" {
		t.Fatalf("code=%v", payload["code"])
	}
}

func TestSystemStreamEmitsStatusFrames(t *testing.T) {
	snapshot := map[string]any{
		"sample":       map[string]any{"cpu_percent": 7.5},
		"dependencies": map[string]any{"database": map[string]any{"status": "ok"}},
	}
	s := &server{config: ServerConfig{SystemStatus: func(*http.Request) (any, error) {
		return snapshot, nil
	}}}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/system/stream?interval=100", nil)
	req = req.WithContext(ctx)
	req = withAuthenticatedRequest(req, &auth.AuthenticatedTx{
		Principal: auth.Principal{ActorID: "operator-1", Role: "operator"},
	}, "session-token")
	res := httptest.NewRecorder()

	done := make(chan struct{})
	go func() {
		s.systemStream(res, req)
		close(done)
	}()

	deadline := time.Now().Add(2 * time.Second)
	var body string
	for {
		body = res.Body.String()
		if strings.Count(body, "event: status") >= 2 {
			break
		}
		if time.Now().After(deadline) {
			cancel()
			t.Fatalf("expected at least 2 status frames within deadline, got: %q", body)
		}
		time.Sleep(10 * time.Millisecond)
	}
	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("handler did not return after client cancel")
	}

	if res.Code != http.StatusOK {
		t.Fatalf("status=%d", res.Code)
	}
	if got := res.Header().Get("Content-Type"); !strings.HasPrefix(got, "text/event-stream") {
		t.Fatalf("content-type=%q", got)
	}
	frames := parseSSEFramesForTest(body)
	statusFrames := 0
	for _, frame := range frames {
		if frame.event != "status" {
			continue
		}
		statusFrames++
		var payload map[string]any
		if err := json.Unmarshal([]byte(frame.data), &payload); err != nil {
			t.Fatalf("frame %d data is not JSON: %v (%q)", statusFrames, err, frame.data)
		}
		if payload["traffic_semantics"] != "server_counter_delta" {
			t.Fatalf("traffic semantics=%v", payload["traffic_semantics"])
		}
		if payload["metrics"] == nil {
			t.Fatal("metrics payload missing")
		}
	}
	if statusFrames < 2 {
		t.Fatalf("expected >=2 status frames, got %d", statusFrames)
	}
	if !strings.Contains(body, "retry: 3000") {
		t.Fatal("stream must advertise a reconnect hint")
	}
}

func TestSystemStreamEmitsErrorFramesButStaysOpen(t *testing.T) {
	calls := 0
	s := &server{config: ServerConfig{SystemStatus: func(*http.Request) (any, error) {
		calls++
		return nil, context.DeadlineExceeded
	}}}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/system/stream?interval=100", nil)
	req = req.WithContext(ctx)
	req = withAuthenticatedRequest(req, &auth.AuthenticatedTx{
		Principal: auth.Principal{ActorID: "operator-1", Role: "operator"},
	}, "session-token")
	res := httptest.NewRecorder()

	done := make(chan struct{})
	go func() {
		s.systemStream(res, req)
		close(done)
	}()

	deadline := time.Now().Add(2 * time.Second)
	var body string
	for {
		body = res.Body.String()
		if strings.Count(body, "event: error") >= 2 {
			break
		}
		if time.Now().After(deadline) {
			cancel()
			t.Fatalf("expected at least 2 error frames within deadline, got: %q", body)
		}
		time.Sleep(10 * time.Millisecond)
	}
	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("handler did not return after client cancel")
	}

	if !strings.Contains(body, `"code":"system_metrics_unavailable"`) {
		t.Fatalf("error frame missing fail-closed code: %q", body)
	}
}

func TestSystemStreamFinalizesAuthTransactionBeforeStreaming(t *testing.T) {
	s := &server{config: ServerConfig{SystemStatus: func(*http.Request) (any, error) {
		return map[string]any{"sample": map[string]any{}}, nil
	}}}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/system/stream?interval=100", nil)
	req = req.WithContext(ctx)
	bound := &auth.AuthenticatedTx{Principal: auth.Principal{ActorID: "operator-1", Role: "operator"}}
	req = withAuthenticatedRequest(req, bound, "session-token")
	recorded, _ := authenticatedFromRequest(req)
	_ = recorded
	res := httptest.NewRecorder()

	done := make(chan struct{})
	go func() {
		s.systemStream(res, req)
		close(done)
	}()
	deadline := time.Now().Add(2 * time.Second)
	for strings.Count(res.Body.String(), "event: status") < 1 && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("handler did not return after client cancel")
	}
	recorded, ok := authenticatedFromRequest(req)
	if !ok || !recorded.TransactionFinalized {
		t.Fatal("stream must finalize the auth transaction before streaming")
	}
}

func TestResponseBufferStreamingPassesThroughAndCommitIsNoop(t *testing.T) {
	rec := httptest.NewRecorder()
	buf := &responseBuffer{w: rec}
	buf.enableStreaming()
	buf.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	buf.WriteHeader(http.StatusOK)
	buf.Write([]byte("event: status\ndata: {}\n\n"))
	buf.commitToClient()

	if got := rec.Header().Get("Content-Type"); got != "text/event-stream; charset=utf-8" {
		t.Fatalf("content-type=%q", got)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d", rec.Code)
	}
	if got := rec.Body.String(); got != "event: status\ndata: {}\n\n" {
		t.Fatalf("body=%q", got)
	}
}

type sseFrame struct {
	event string
	data  string
}

func parseSSEFramesForTest(body string) []sseFrame {
	var frames []sseFrame
	var current sseFrame
	for _, line := range strings.Split(body, "\n") {
		line = strings.TrimRight(line, "\r")
		switch {
		case strings.HasPrefix(line, "event: "):
			current.event = strings.TrimPrefix(line, "event: ")
		case strings.HasPrefix(line, "data: "):
			if current.data != "" {
				current.data += "\n"
			}
			current.data += strings.TrimPrefix(line, "data: ")
		case line == "":
			if current.event != "" || current.data != "" {
				frames = append(frames, current)
				current = sseFrame{}
			}
		}
	}
	return frames
}
