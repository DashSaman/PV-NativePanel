package runtimeevent

import (
	"context"
	"database/sql"
	"testing"
	"time"
)

type activationProbe struct {
	calls        int
	credentialID string
	observedAt   time.Time
	activated    bool
	err          error
}

func (p *activationProbe) ActivateFirstUse(_ context.Context, _ *sql.Tx, credentialID string, observedAt time.Time) (bool, error) {
	p.calls++
	p.credentialID = credentialID
	p.observedAt = observedAt
	return p.activated, p.err
}

func TestHandleFirstUseOnlyAcceptsAuthenticatedConnect(t *testing.T) {
	observed := time.Date(2026, 8, 29, 1, 2, 3, 0, time.UTC)
	for _, event := range []FirstUseEvent{
		{RuntimeCredentialID: "runtime-1", Method: "GET", Authenticated: true, ObservedAt: observed},
		{RuntimeCredentialID: "runtime-1", Method: "CONNECT", Authenticated: false, ObservedAt: observed},
		{RuntimeCredentialID: "runtime-1", Method: "SUBSCRIPTION_FETCH", Authenticated: true, ObservedAt: observed},
	} {
		probe := &activationProbe{activated: true}
		activated, err := HandleFirstUse(context.Background(), nil, probe, event)
		if err != nil {
			t.Fatalf("HandleFirstUse(%+v) error = %v", event, err)
		}
		if activated || probe.calls != 0 {
			t.Fatalf("untrusted event %+v activated=%v calls=%d", event, activated, probe.calls)
		}
	}
}

func TestHandleFirstUseActivatesAuthenticatedConnectOnceThroughCustomerService(t *testing.T) {
	observed := time.Date(2026, 8, 29, 4, 5, 6, 0, time.UTC)
	probe := &activationProbe{activated: true}
	event := FirstUseEvent{RuntimeCredentialID: "runtime-1", Method: "CONNECT", Authenticated: true, ObservedAt: observed}
	activated, err := HandleFirstUse(context.Background(), nil, probe, event)
	if err != nil {
		t.Fatalf("HandleFirstUse() error = %v", err)
	}
	if !activated || probe.calls != 1 || probe.credentialID != "runtime-1" || !probe.observedAt.Equal(observed) {
		t.Fatalf("activation result=%v probe=%+v", activated, probe)
	}
}

func TestHandleFirstUseRejectsMalformedTrustedEvent(t *testing.T) {
	probe := &activationProbe{activated: true}
	if _, err := HandleFirstUse(context.Background(), nil, probe, FirstUseEvent{Method: "CONNECT", Authenticated: true, ObservedAt: time.Now()}); err == nil {
		t.Fatal("accepted empty runtime credential id")
	}
	if _, err := HandleFirstUse(context.Background(), nil, probe, FirstUseEvent{RuntimeCredentialID: "runtime-1", Method: "CONNECT", Authenticated: true}); err == nil {
		t.Fatal("accepted zero observed time")
	}
	if probe.calls != 0 {
		t.Fatalf("malformed event reached activator: calls=%d", probe.calls)
	}
}
