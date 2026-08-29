package observability

import (
	"context"
	"strings"
	"testing"
	"time"
)

type fakeMetricSource struct {
	samples []RawSample
	index   int
}

func (f *fakeMetricSource) Read(context.Context) (RawSample, error) {
	value := f.samples[f.index]
	if f.index < len(f.samples)-1 {
		f.index++
	}
	return value, nil
}

func TestCollectorComputesNetworkRatesServerSide(t *testing.T) {
	source := &fakeMetricSource{samples: []RawSample{
		{Timestamp: time.Unix(100, 0).UTC(), CPUIdle: 80, CPUTotal: 100, RXBytes: 1000, TXBytes: 2000},
		{Timestamp: time.Unix(102, 0).UTC(), CPUIdle: 90, CPUTotal: 120, RXBytes: 5000, TXBytes: 8000},
	}}
	collector := NewCollector(source)

	first, err := collector.Collect(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if first.RateAvailable {
		t.Fatal("first sample must not invent a traffic rate")
	}

	second, err := collector.Collect(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !second.RateAvailable {
		t.Fatal("second sample should have a server-side rate")
	}
	if second.RXBytesPerSecond != 2000 || second.TXBytesPerSecond != 3000 {
		t.Fatalf("rx=%v tx=%v", second.RXBytesPerSecond, second.TXBytesPerSecond)
	}
	if second.SampledAt.Unix() != 102 || second.SampleWindowSeconds != 2 {
		t.Fatalf("timestamp/window=%v/%v", second.SampledAt, second.SampleWindowSeconds)
	}
}

func TestCollectorRejectsCounterRollbackInsteadOfNegativeRate(t *testing.T) {
	source := &fakeMetricSource{samples: []RawSample{
		{Timestamp: time.Unix(100, 0).UTC(), RXBytes: 5000, TXBytes: 8000},
		{Timestamp: time.Unix(101, 0).UTC(), RXBytes: 100, TXBytes: 200},
	}}
	collector := NewCollector(source)
	_, _ = collector.Collect(context.Background())
	second, err := collector.Collect(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if second.RateAvailable || second.RXBytesPerSecond != 0 || second.TXBytesPerSecond != 0 {
		t.Fatalf("counter rollback produced a rate: %#v", second)
	}
}

func TestStructuredLogRedactsKnownSecrets(t *testing.T) {
	line, err := MarshalLog(Event{
		Timestamp: time.Unix(100, 0).UTC(),
		Level:     "info",
		RequestID: "req-123",
		Component: "runtime",
		Message:   "mutation completed",
		Fields: map[string]any{
			"password":           "secret-password",
			"subscription_token": "secret-token",
			"authorization":      "Bearer secret-auth",
			"db_password":        "secret-db",
			"private_key":        "secret-key",
			"customer_id":        "customer-1",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	text := string(line)
	for _, secret := range []string{"secret-password", "secret-token", "secret-auth", "secret-db", "secret-key"} {
		if strings.Contains(text, secret) {
			t.Fatalf("secret leaked: %q in %s", secret, text)
		}
	}
	for _, want := range []string{"req-123", "runtime", "customer-1", "[REDACTED]"} {
		if !strings.Contains(text, want) {
			t.Fatalf("structured log missing %q: %s", want, text)
		}
	}
}
