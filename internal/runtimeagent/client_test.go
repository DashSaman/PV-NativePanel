package runtimeagent

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/DashSaman/PV-NaivePanel/internal/runtimecred"
)

func TestClientUsesUnixSocketAndCanonicalPaths(t *testing.T) {
	socketPath := filepath.Join(t.TempDir(), "client.sock")
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		t.Fatalf("net.Listen(unix) error = %v", err)
	}
	defer listener.Close()

	seen := make(chan string, 1)
	server := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen <- r.Method + " " + r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(HealthResponse{Status: "ok"})
	})}
	defer server.Close()
	go server.Serve(listener)

	client := NewClient(socketPath)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	response, err := client.Health(ctx)
	if err != nil {
		t.Fatalf("Client.Health() error = %v", err)
	}
	if response.Status != "ok" {
		t.Fatalf("health status = %q, want ok", response.Status)
	}
	select {
	case got := <-seen:
		if got != "GET /v1/health" {
			t.Fatalf("request = %q, want GET /v1/health", got)
		}
	case <-ctx.Done():
		t.Fatal("Unix server did not receive client request")
	}
}

func TestClientRejectsNonSuccessAndMalformedResponse(t *testing.T) {
	tests := []struct {
		name   string
		status int
		body   string
	}{
		{name: "non-success", status: http.StatusBadRequest, body: `{"error":{"code":"bad_request"}}`},
		{name: "malformed success", status: http.StatusOK, body: `{"status":`},
		{name: "unknown success field", status: http.StatusOK, body: `{"status":"ok","password":"leak"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			socketPath := filepath.Join(t.TempDir(), "client.sock")
			listener, err := net.Listen("unix", socketPath)
			if err != nil {
				t.Fatal(err)
			}
			defer listener.Close()
			server := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.status)
				_, _ = w.Write([]byte(tt.body))
			})}
			defer server.Close()
			go server.Serve(listener)

			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			_, err = NewClient(socketPath).Health(ctx)
			if err == nil {
				t.Fatalf("Client.Health() unexpectedly accepted status=%d body=%s", tt.status, tt.body)
			}
		})
	}
}

func TestProtocolRequestTypesHaveNoArbitraryExecutionFields(t *testing.T) {
	encoded, err := json.Marshal(ApplyRequest{
		ExpectedCaddySHA256: strings.Repeat("a", 64),
		Desired: DesiredStateInput{
			Revision: "rev-1",
			Credentials: []CredentialInput{{
				ID:       "cred-1",
				Username: "user.one",
				Password: "safe password 123",
				Status:   runtimecred.CredentialActive,
			}},
		},
	})
	if err != nil {
		t.Fatalf("json.Marshal(ApplyRequest) error = %v", err)
	}
	lower := strings.ToLower(string(encoded))
	for _, forbidden := range []string{"command", "argv", "service", "path", "binary", "systemctl"} {
		if strings.Contains(lower, `"`+forbidden+`"`) {
			t.Fatalf("ApplyRequest JSON exposed arbitrary execution field %q: %s", forbidden, encoded)
		}
	}
}
