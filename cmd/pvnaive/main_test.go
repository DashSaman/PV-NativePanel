package main

import (
	"strings"
	"testing"
)

func TestDatabaseDSNUsesPVNaiveEnvironment(t *testing.T) {
	env := map[string]string{
		"PVNAIVE_DB_HOST":            "127.0.0.1",
		"PVNAIVE_DB_PORT":            "5432",
		"PVNAIVE_DB_NAME":            "pvnaive",
		"PVNAIVE_DB_USER":            "pvnaive_app",
		"PVNAIVE_DB_CONNECT_TIMEOUT": "5",
	}

	dsn, err := databaseDSN(func(key string) string { return env[key] })
	if err != nil {
		t.Fatalf("databaseDSN returned error: %v", err)
	}
	for _, want := range []string{
		"host=127.0.0.1",
		"port=5432",
		"dbname=pvnaive",
		"user=pvnaive_app",
		"connect_timeout=5",
	} {
		if !strings.Contains(dsn, want) {
			t.Fatalf("DSN %q does not contain %q", dsn, want)
		}
	}
}

func TestDatabaseDSNRejectsMissingRequiredValues(t *testing.T) {
	env := map[string]string{
		"PVNAIVE_DB_HOST":            "127.0.0.1",
		"PVNAIVE_DB_PORT":            "5432",
		"PVNAIVE_DB_NAME":            "pvnaive",
		"PVNAIVE_DB_CONNECT_TIMEOUT": "5",
	}

	if _, err := databaseDSN(func(key string) string { return env[key] }); err == nil {
		t.Fatal("databaseDSN accepted missing PVNAIVE_DB_USER")
	}
}

func TestSubscriptionProxyHostRequiresValidExplicitHostWhenEnabled(t *testing.T) {
	env := map[string]string{"PVNAIVE_NAIVE_PUBLIC_HOST": "proxy.example.test"}
	got, err := subscriptionProxyHost(func(key string) string { return env[key] }, true)
	if err != nil {
		t.Fatalf("subscriptionProxyHost returned error: %v", err)
	}
	if got != "proxy.example.test" {
		t.Fatalf("subscriptionProxyHost=%q", got)
	}

	if _, err := subscriptionProxyHost(func(string) string { return "" }, true); err == nil {
		t.Fatal("enabled subscription accepted missing public host")
	}
	if _, err := subscriptionProxyHost(func(string) string { return "https://bad.example/path" }, true); err == nil {
		t.Fatal("subscription accepted URL instead of host[:port]")
	}
}

func TestSubscriptionProxyHostMayBeUnsetWhenCustomerRuntimeIsDisabled(t *testing.T) {
	got, err := subscriptionProxyHost(func(string) string { return "" }, false)
	if err != nil || got != "" {
		t.Fatalf("disabled subscription host=%q err=%v", got, err)
	}
}

func TestExpectedSchemaVersion(t *testing.T) {
	for _, tc := range []struct {
		value string
		want  int
		ok    bool
	}{
		{"16", 16, true}, {"1", 1, true}, {"", 0, false}, {"0", 0, false}, {"-1", 0, false}, {"abc", 0, false}, {"16x", 0, false},
	} {
		got, err := expectedSchemaVersion(func(string) string { return tc.value })
		if tc.ok {
			if err != nil || got != tc.want {
				t.Fatalf("value=%q got=%d err=%v", tc.value, got, err)
			}
		} else if err == nil {
			t.Fatalf("value=%q unexpectedly accepted as %d", tc.value, got)
		}
	}
}

func TestBuildFleetPullListener(t *testing.T) {
	// Disabled when nothing is set (the default deployment posture).
	server, err := buildFleetPullListener(func(string) string { return "" }, nil)
	if err != nil || server != nil {
		t.Fatalf("unset env: server=%v err=%v, want nil/nil", server, err)
	}
	// Partial configuration must fail loudly, never half-enable.
	getenv := func(key string) string {
		if key == "PVNAIVE_FLEET_PULL_LISTEN" {
			return "203.0.113.10:9443"
		}
		return ""
	}
	if _, err := buildFleetPullListener(getenv, nil); err == nil {
		t.Fatalf("partial configuration unexpectedly accepted")
	}
	// Loopback is allowed for single-host rehearsals; malformed hosts are not.
	getenvFull := func(key string) string {
		switch key {
		case "PVNAIVE_FLEET_PULL_LISTEN":
			return "127.0.0.1:9443"
		case "PVNAIVE_FLEET_PULL_CERT_FILE":
			return "/tmp/cert.pem"
		case "PVNAIVE_FLEET_PULL_KEY_FILE":
			return "/tmp/key.pem"
		case "PVNAIVE_FLEET_PULL_CLIENT_CA_FILE":
			return "/tmp/ca.pem"
		}
		return ""
	}
	if _, err := buildFleetPullListener(getenvFull, nil); err == nil {
		t.Fatalf("missing certificate files should fail during load")
	}
	// Hostname listen addresses are refused (explicit IPv4 only).
	getenvHost := func(key string) string {
		if key == "PVNAIVE_FLEET_PULL_LISTEN" {
			return "registry.example.net:9443"
		}
		if key == "PVNAIVE_FLEET_PULL_CERT_FILE" {
			return "/tmp/cert.pem"
		}
		if key == "PVNAIVE_FLEET_PULL_KEY_FILE" {
			return "/tmp/key.pem"
		}
		return "/tmp/ca.pem"
	}
	if _, err := buildFleetPullListener(getenvHost, nil); err == nil {
		t.Fatalf("hostname listen address unexpectedly accepted")
	}
}
