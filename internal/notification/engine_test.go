package notification

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

type fakeChannel struct {
	failures int
	calls    int
	messages []Message
}

func (f *fakeChannel) Name() string { return "fake" }

func (f *fakeChannel) Send(_ context.Context, message Message) error {
	f.calls++
	f.messages = append(f.messages, message)
	if f.calls <= f.failures {
		return errors.New("temporary delivery failure")
	}
	return nil
}

func TestEngineRetriesAndDeduplicates(t *testing.T) {
	channel := &fakeChannel{failures: 2}
	now := time.Unix(100, 0).UTC()
	engine := NewEngine(Config{
		MaxAttempts: 3,
		DedupeTTL:   10 * time.Minute,
		Now:         func() time.Time { return now },
		Sleep:       func(context.Context, time.Duration) error { return nil },
	})
	event := Event{Type: RuntimeDown, Key: "runtime:local", Title: "Runtime down", Body: "Runtime Agent is unavailable"}

	if err := engine.Deliver(context.Background(), channel, event); err != nil {
		t.Fatal(err)
	}
	if channel.calls != 3 {
		t.Fatalf("calls=%d want 3", channel.calls)
	}
	if err := engine.Deliver(context.Background(), channel, event); err != nil {
		t.Fatal(err)
	}
	if channel.calls != 3 {
		t.Fatalf("dedupe failed, calls=%d", channel.calls)
	}

	now = now.Add(11 * time.Minute)
	if err := engine.Deliver(context.Background(), channel, event); err != nil {
		t.Fatal(err)
	}
	if channel.calls != 4 {
		t.Fatalf("expired dedupe did not deliver, calls=%d", channel.calls)
	}
}

func TestNotificationSanitizationRemovesSecrets(t *testing.T) {
	message := Message{
		Title: "Backup failed",
		Body:  "Authorization: Bearer top-secret subscription_token=abc123 password=hunter2",
	}
	safe := Sanitize(message)
	for _, secret := range []string{"top-secret", "abc123", "hunter2"} {
		if strings.Contains(safe.Body, secret) {
			t.Fatalf("secret leaked into notification: %s", safe.Body)
		}
	}
	if !strings.Contains(safe.Body, "[REDACTED]") {
		t.Fatalf("notification did not signal redaction: %s", safe.Body)
	}
}

func TestEventPolicyKeepsQuotaEventsCapabilityGated(t *testing.T) {
	if Enabled(QuotaWarning, Capabilities{ExactAccounting: false}) {
		t.Fatal("quota warning must stay disabled until exact accounting exists")
	}
	if Enabled(QuotaDepleted, Capabilities{ExactAccounting: false}) {
		t.Fatal("quota depleted must stay disabled until exact accounting exists")
	}
	if !Enabled(RuntimeDown, Capabilities{}) || !Enabled(Expired, Capabilities{}) {
		t.Fatal("runtime and expiry notifications should not depend on accounting")
	}
}
