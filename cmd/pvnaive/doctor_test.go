package main

import "testing"

func TestRunDoctorRejectsUnknownArguments(t *testing.T) {
	if err := runDoctor([]string{"--definitely-unknown"}); err == nil {
		t.Fatal("unknown doctor argument must fail closed")
	}
}

func TestDoctorCheckSetIncludesCurrentTelemetryBoundary(t *testing.T) {
	names := doctorCheckNames()
	for _, want := range []string{"pvnaive-api.service", "pvnaive-runtime-agent.service", "pvnaive-telemetry-agent.service", "accounting-socket", "telemetry-health"} {
		found := false
		for _, got := range names {
			if got == want {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("doctor check %q missing from %#v", want, names)
		}
	}
}
