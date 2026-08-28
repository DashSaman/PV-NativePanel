package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNaiveRuntimeRoutesAreOwnerOnlyAndReady(t *testing.T) {
	expected := map[string]string{
		"GET /api/v1/runtime/naive":                                   "runtime.naive.show",
		"GET /api/v1/runtime/naive/credentials":                       "runtime.naive.credentials.index",
		"POST /api/v1/runtime/naive/credentials":                      "runtime.naive.credentials.create",
		"PATCH /api/v1/runtime/naive/credentials/{id}":                "runtime.naive.credentials.update",
		"POST /api/v1/runtime/naive/credentials/{id}/rotate-password": "runtime.naive.credentials.rotate",
		"DELETE /api/v1/runtime/naive/credentials/{id}":               "runtime.naive.credentials.revoke",
	}
	seen := map[string]bool{}
	for _, route := range Routes {
		key := route.Method + " " + route.Path
		name, ok := expected[key]
		if !ok {
			continue
		}
		seen[key] = true
		if route.Name != name {
			t.Fatalf("%s name=%q, want %q", key, route.Name, name)
		}
		if route.Access != Owner {
			t.Fatalf("%s access=%q, want owner", key, route.Access)
		}
		if !route.Ready {
			t.Fatalf("%s is not marked ready", key)
		}
	}
	for key := range expected {
		if !seen[key] {
			t.Fatalf("missing route %s", key)
		}
	}
}

func TestRuntimeIdempotencyKeyContract(t *testing.T) {
	for _, tc := range []struct {
		name  string
		value string
		ok    bool
	}{
		{name: "missing", value: "", ok: false},
		{name: "short", value: "1234567", ok: false},
		{name: "valid", value: "runtime-create-0001", ok: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/v1/runtime/naive/credentials", nil)
			if tc.value != "" {
				req.Header.Set("Idempotency-Key", tc.value)
			}
			got, err := runtimeIdempotencyKey(req)
			if tc.ok && (err != nil || got != tc.value) {
				t.Fatalf("runtimeIdempotencyKey()=(%q,%v), want %q,nil", got, err, tc.value)
			}
			if !tc.ok && err == nil {
				t.Fatalf("runtimeIdempotencyKey() accepted %q", tc.value)
			}
		})
	}
}

func TestRuntimeExpectedRevisionUsesIfMatch(t *testing.T) {
	for _, tc := range []struct {
		value string
		want  int64
		ok    bool
	}{
		{value: "", ok: false},
		{value: "0", ok: false},
		{value: "abc", ok: false},
		{value: "7", want: 7, ok: true},
		{value: "\"9\"", want: 9, ok: true},
	} {
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/runtime/naive/credentials/id", nil)
		if tc.value != "" {
			req.Header.Set("If-Match", tc.value)
		}
		got, err := runtimeExpectedRevision(req)
		if tc.ok && (err != nil || got != tc.want) {
			t.Fatalf("If-Match %q => (%d,%v), want %d,nil", tc.value, got, err, tc.want)
		}
		if !tc.ok && err == nil {
			t.Fatalf("If-Match %q was accepted", tc.value)
		}
	}
}

func TestRuntimeJSONIsStrict(t *testing.T) {
	var payload struct {
		Username string `json:"username"`
	}
	unknown := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"username":"safe.user","unknown":true}`))
	if err := decodeRuntimeJSON(unknown, &payload); err == nil {
		t.Fatal("decodeRuntimeJSON accepted unknown field")
	}
	trailing := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"username":"safe.user"} {}`))
	if err := decodeRuntimeJSON(trailing, &payload); err == nil {
		t.Fatal("decodeRuntimeJSON accepted trailing JSON")
	}
	valid := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"username":"safe.user"}`))
	if err := decodeRuntimeJSON(valid, &payload); err != nil || payload.Username != "safe.user" {
		t.Fatalf("decodeRuntimeJSON valid payload error=%v username=%q", err, payload.Username)
	}
}
