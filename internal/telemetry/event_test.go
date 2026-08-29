package telemetry

import (
	"errors"
	"testing"
	"time"
)

const (
	testRuntimeID = "11111111-1111-4111-8111-111111111111"
	testTermID    = "22222222-2222-4222-8222-222222222222"
	testNodeID    = "node-a"
	testBootID    = "33333333-3333-4333-8333-333333333333"
	testSessionID = "44444444-4444-4444-8444-444444444444"
)

func eventAt(seq int64, upload int64, download int64, at time.Time) Event {
	return Event{
		RuntimeCredentialID: testRuntimeID,
		Username:            "alice",
		NodeID:              testNodeID,
		BootID:              testBootID,
		SessionID:           testSessionID,
		Sequence:            seq,
		ObservedAt:          at,
		AuthenticatedConnect: true,
		UploadBytes:         upload,
		DownloadBytes:       download,
	}
}

func TestSessionStateDerivesOnlyMonotonicCumulativeDelta(t *testing.T) {
	base := time.Date(2026, 8, 29, 12, 0, 0, 0, time.UTC)
	var state SessionState

	first, err := state.Apply(eventAt(1, 10, 20, base))
	if err != nil {
		t.Fatalf("first Apply() error = %v", err)
	}
	if first.UploadDelta != 10 || first.DownloadDelta != 20 || first.Duplicate {
		t.Fatalf("first result = %+v, want delta 10/20 non-duplicate", first)
	}

	second, err := state.Apply(eventAt(2, 25, 55, base.Add(time.Second)))
	if err != nil {
		t.Fatalf("second Apply() error = %v", err)
	}
	if second.UploadDelta != 15 || second.DownloadDelta != 35 || second.Duplicate {
		t.Fatalf("second result = %+v, want delta 15/35 non-duplicate", second)
	}
}

func TestSessionStateDuplicateIsIdempotent(t *testing.T) {
	base := time.Date(2026, 8, 29, 12, 0, 0, 0, time.UTC)
	var state SessionState
	e := eventAt(1, 10, 20, base)
	if _, err := state.Apply(e); err != nil {
		t.Fatalf("first Apply() error = %v", err)
	}
	result, err := state.Apply(e)
	if err != nil {
		t.Fatalf("duplicate Apply() error = %v", err)
	}
	if !result.Duplicate || result.UploadDelta != 0 || result.DownloadDelta != 0 {
		t.Fatalf("duplicate result = %+v, want idempotent zero delta", result)
	}
}

func TestSessionStateSameSequenceConflictFailsClosed(t *testing.T) {
	base := time.Date(2026, 8, 29, 12, 0, 0, 0, time.UTC)
	var state SessionState
	if _, err := state.Apply(eventAt(1, 10, 20, base)); err != nil {
		t.Fatalf("first Apply() error = %v", err)
	}
	_, err := state.Apply(eventAt(1, 11, 20, base))
	if !errors.Is(err, ErrSequenceConflict) {
		t.Fatalf("conflicting Apply() error = %v, want ErrSequenceConflict", err)
	}
}

func TestSessionStateRejectsGapOutOfOrderAndCounterRegression(t *testing.T) {
	base := time.Date(2026, 8, 29, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		name string
		event Event
		want error
	}{
		{name: "gap", event: eventAt(3, 30, 30, base.Add(2 * time.Second)), want: ErrSequenceGap},
		{name: "out_of_order", event: eventAt(0, 0, 0, base), want: ErrInvalidEvent},
		{name: "counter_regression", event: eventAt(2, 9, 21, base.Add(time.Second)), want: ErrCounterRegression},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var state SessionState
			if _, err := state.Apply(eventAt(1, 10, 20, base)); err != nil {
				t.Fatalf("first Apply() error = %v", err)
			}
			_, err := state.Apply(tt.event)
			if !errors.Is(err, tt.want) {
				t.Fatalf("Apply() error = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestEventRequiresTrustedAuthenticatedConnectIdentity(t *testing.T) {
	base := time.Date(2026, 8, 29, 12, 0, 0, 0, time.UTC)
	valid := eventAt(1, 0, 0, base)
	tests := []struct {
		name   string
		mutate func(*Event)
	}{
		{name: "runtime_uuid", mutate: func(e *Event) { e.RuntimeCredentialID = "alice" }},
		{name: "node", mutate: func(e *Event) { e.NodeID = "" }},
		{name: "boot_uuid", mutate: func(e *Event) { e.BootID = "" }},
		{name: "session_uuid", mutate: func(e *Event) { e.SessionID = "" }},
		{name: "sequence", mutate: func(e *Event) { e.Sequence = 0 }},
		{name: "observed_at", mutate: func(e *Event) { e.ObservedAt = time.Time{} }},
		{name: "auth_marker", mutate: func(e *Event) { e.AuthenticatedConnect = false }},
		{name: "negative_upload", mutate: func(e *Event) { e.UploadBytes = -1 }},
		{name: "negative_download", mutate: func(e *Event) { e.DownloadBytes = -1 }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := valid
			tt.mutate(&e)
			if err := ValidateEvent(e); !errors.Is(err, ErrInvalidEvent) {
				t.Fatalf("ValidateEvent() error = %v, want ErrInvalidEvent", err)
			}
		})
	}
}

func TestSessionKeySeparatesReconnectAndBootRestart(t *testing.T) {
	base := time.Date(2026, 8, 29, 12, 0, 0, 0, time.UTC)
	first := eventAt(1, 0, 0, base)
	reconnect := first
	reconnect.SessionID = "55555555-5555-4555-8555-555555555555"
	restarted := first
	restarted.BootID = "66666666-6666-4666-8666-666666666666"
	if first.Key() == reconnect.Key() {
		t.Fatal("reconnect reused the same session key")
	}
	if first.Key() == restarted.Key() {
		t.Fatal("boot restart reused the same session key")
	}
}

func TestFinalEventClosesSessionWithoutInventingBytes(t *testing.T) {
	base := time.Date(2026, 8, 29, 12, 0, 0, 0, time.UTC)
	var state SessionState
	if _, err := state.Apply(eventAt(1, 10, 20, base)); err != nil {
		t.Fatalf("first Apply() error = %v", err)
	}
	final := eventAt(2, 15, 25, base.Add(time.Second))
	final.Final = true
	result, err := state.Apply(final)
	if err != nil {
		t.Fatalf("final Apply() error = %v", err)
	}
	if !result.Final || result.UploadDelta != 5 || result.DownloadDelta != 5 {
		t.Fatalf("final result = %+v, want exact 5/5 delta and final", result)
	}
	if !state.Complete() {
		t.Fatal("final accepted counter must mark the session complete")
	}
}
