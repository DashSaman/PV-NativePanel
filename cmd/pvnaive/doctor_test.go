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

func TestDoctorKeyModesMatchProductionServiceAccess(t *testing.T) {
	modes := doctorKeyModes()
	if got := modes["auth-key-mode"]; got != 0o640 {
		t.Fatalf("auth key mode = %#o, want 0640 so root:pvnaive is readable by the API service", got)
	}
	if got := modes["runtime-key-mode"]; got != 0o640 {
		t.Fatalf("runtime key mode = %#o, want 0640 so root:pvnaive is readable by the API service", got)
	}
	if got := modes["backup-key-mode"]; got != 0o600 {
		t.Fatalf("backup identity mode = %#o, want root-only 0600", got)
	}
}
