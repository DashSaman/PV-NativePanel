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
