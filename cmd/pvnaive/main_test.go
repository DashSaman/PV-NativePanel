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
