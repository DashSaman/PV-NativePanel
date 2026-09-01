package sessioncontrol

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DashSaman/PV-NaivePanel/internal/sessionkill"
)

func testKey() sessionkill.Key {
	return sessionkill.Key{RuntimeCredentialID: "11111111-1111-4111-8111-111111111111", NodeID: "node-a", BootID: "22222222-2222-4222-8222-222222222222", SessionID: "33333333-3333-4333-8333-333333333333"}
}

func testClient(server *httptest.Server) *Client {
	transport := &http.Transport{DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
		var dialer net.Dialer
		return dialer.DialContext(ctx, "tcp", server.Listener.Addr().String())
	}}
	return &Client{httpClient: &http.Client{Transport: transport}}
}

func TestClientKillSendsExactTupleAndReturnsResult(t *testing.T) {
	var seen KillRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/sessions/kill" { t.Fatalf("unexpected %s %s", r.Method, r.URL.Path) }
		if err := json.NewDecoder(r.Body).Decode(&seen); err != nil { t.Fatalf("decode: %v", err) }
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(KillResult{Found: true, Killed: true})
	}))
	defer server.Close()
	key := testKey()
	result, err := testClient(server).Kill(context.Background(), key)
	if err != nil { t.Fatalf("Kill: %v", err) }
	if !result.Found || !result.Killed { t.Fatalf("expected found=true killed=true, got %+v", result) }
	if seen.RuntimeCredentialID != key.RuntimeCredentialID || seen.NodeID != key.NodeID || seen.BootID != key.BootID || seen.SessionID != key.SessionID { t.Fatalf("wrong tuple sent: %+v", seen) }
}

func TestClientKillIdempotentReturnsFoundNotKilled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { _ = json.NewEncoder(w).Encode(KillResult{Found: true, Killed: false}) }))
	defer server.Close()
	result, err := testClient(server).Kill(context.Background(), testKey())
	if err != nil { t.Fatalf("Kill: %v", err) }
	if !result.Found || result.Killed { t.Fatalf("expected found=true killed=false, got %+v", result) }
}

func TestClientKillNotFoundReturnsBothFalse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { _ = json.NewEncoder(w).Encode(KillResult{}) }))
	defer server.Close()
	result, err := testClient(server).Kill(context.Background(), testKey())
	if err != nil { t.Fatalf("Kill: %v", err) }
	if result.Found || result.Killed { t.Fatalf("expected both false, got %+v", result) }
}

func TestClientKillServerDownReturnsError(t *testing.T) {
	client := &Client{httpClient: &http.Client{Transport: &http.Transport{DialContext: func(_ context.Context, _, _ string) (net.Conn, error) { return nil, &net.OpError{Op: "dial", Net: "unix", Err: errors.New("connection refused")} }}, Timeout: 2 * time.Second}}
	if _, err := client.Kill(context.Background(), testKey()); err == nil { t.Fatal("expected error for unreachable socket") }
}

func TestClientKillRejectsTrailingJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte(`{"found":true,"killed":true}{"extra":"data"}`)) }))
	defer server.Close()
	if _, err := testClient(server).Kill(context.Background(), testKey()); err == nil { t.Fatal("expected error for trailing JSON") }
}
