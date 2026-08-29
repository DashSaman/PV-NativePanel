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

func TestEngineRetriesDeduplicatesAndSanitizes(t *testing.T) {
	channel := &fakeChannel{failures: 2}
	now := time.Unix(100, 0).UTC()
	engine := NewEngine(Config{
		MaxAttempts: 3,
		DedupeTTL:   10 * time.Minute,
		Now:         func() time.Time { return now },
		Sleep:       func(context.Context, time.Duration) error { return nil },
	})
	event := Event{Type: RuntimeDown, Key: "runtime:local", Title: "Runtime down", Body: "Authorization: Bearer top-secret password=hunter2"}
	if err := engine.Deliver(context.Background(), channel, event); err != nil { t.Fatal(err) }
	if channel.calls != 3 { t.Fatalf("calls=%d want 3", channel.calls) }
	if len(channel.messages) == 0 || strings.Contains(channel.messages[len(channel.messages)-1].Body, "top-secret") || strings.Contains(channel.messages[len(channel.messages)-1].Body, "hunter2") {
		t.Fatalf("notification leaked secret: %#v", channel.messages)
	}
	if err := engine.Deliver(context.Background(), channel, event); err != nil { t.Fatal(err) }
	if channel.calls != 3 { t.Fatalf("dedupe failed, calls=%d", channel.calls) }
	now = now.Add(11 * time.Minute)
	if err := engine.Deliver(context.Background(), channel, event); err != nil { t.Fatal(err) }
	if channel.calls != 4 { t.Fatalf("expired dedupe did not deliver, calls=%d", channel.calls) }
}

func TestQuotaNotificationPolicyRemainsCapabilityGated(t *testing.T) {
	if Enabled(QuotaWarning, Capabilities{ExactAccounting: false}) || Enabled(QuotaDepleted, Capabilities{ExactAccounting: false}) {
		t.Fatal("quota notifications must remain disabled without exact-accounting capability")
	}
	if !Enabled(QuotaWarning, Capabilities{ExactAccounting: true}) || !Enabled(QuotaDepleted, Capabilities{ExactAccounting: true}) {
		t.Fatal("quota notifications should be enabled when exact accounting is explicitly available")
	}
	if !Enabled(RuntimeDown, Capabilities{}) || !Enabled(Expired, Capabilities{}) {
		t.Fatal("runtime and expiry notifications should not depend on accounting")
	}
}
