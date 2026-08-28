package runtimeagent

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

type fakeOperator struct {
	healthCalls   atomic.Int32
	inspectCalls  atomic.Int32
	validateCalls atomic.Int32
	applyCalls    atomic.Int32
	rollbackCalls atomic.Int32
}

func (f *fakeOperator) Health(context.Context) (HealthResponse, error) {
	f.healthCalls.Add(1)
	return HealthResponse{Status: "ok"}, nil
}

func (f *fakeOperator) Inspect(context.Context) (InspectResponse, error) {
	f.inspectCalls.Add(1)
	return InspectResponse{
		CaddySHA256: strings.Repeat("a", 64),
		Credentials: []InspectCredential{{Username: "import.user", Password: "fixture-secret"}},
	}, nil
}

func (f *fakeOperator) Validate(_ context.Context, req ValidateRequest) (ValidateResponse, error) {
	f.validateCalls.Add(1)
	return ValidateResponse{CandidateSHA256: strings.Repeat("b", 64)}, nil
}

func (f *fakeOperator) Apply(_ context.Context, req ApplyRequest) (ApplyResponse, error) {
	f.applyCalls.Add(1)
	return ApplyResponse{
		PreviousSHA256: req.ExpectedCaddySHA256,
		AppliedSHA256:  strings.Repeat("c", 64),
		BackupID:       "backup-20260828T000000Z-1",
		MainPID:        1234,
		NRestarts:      0,
	}, nil
}

func (f *fakeOperator) Rollback(_ context.Context, req RollbackRequest) (RollbackResponse, error) {
	f.rollbackCalls.Add(1)
	return RollbackResponse{RestoredSHA256: strings.Repeat("d", 64), MainPID: 1234, NRestarts: 0}, nil
}

func TestListenUnixCreatesOnlyUnixListener(t *testing.T) {
	socketPath := filepath.Join(t.TempDir(), "runtime-agent.sock")
	listener, err := ListenUnix(socketPath)
	if err != nil {
		t.Fatalf("ListenUnix() error = %v", err)
	}
	defer listener.Close()

	if listener.Addr().Network() != "unix" {
		t.Fatalf("listener network = %q, want unix", listener.Addr().Network())
	}
	if _, ok := listener.(*net.UnixListener); !ok {
		t.Fatalf("listener type = %T, want *net.UnixListener", listener)
	}
}

func TestHandlerRejectsMalformedOversizedAndUnknownJSON(t *testing.T) {
	op := &fakeOperator{}
	socketPath, stop := startAgentServer(t, op)
	defer stop()
	client := unixHTTPClient(socketPath)

	tests := []struct {
		name       string
		body       string
		wantStatus int
	}{
		{name: "malformed", body: `{"expected_caddy_sha256":`, wantStatus: http.StatusBadRequest},
		{name: "unknown service field", body: validValidateJSON(t, `,"service":"caddy-naive.service"`), wantStatus: http.StatusBadRequest},
		{name: "unknown path field", body: validValidateJSON(t, `,"path":"/tmp/attacker.Caddyfile"`), wantStatus: http.StatusBadRequest},
		{name: "oversized", body: `{"padding":"` + strings.Repeat("x", 70*1024) + `"}`, wantStatus: http.StatusRequestEntityTooLarge},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodPost, "http://unix/v1/validate", strings.NewReader(tt.body))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", "application/json")
			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("client.Do() error = %v", err)
			}
			defer resp.Body.Close()
			io.Copy(io.Discard, resp.Body)
			if resp.StatusCode != tt.wantStatus {
				t.Fatalf("status = %d, want %d", resp.StatusCode, tt.wantStatus)
			}
		})
	}

	if got := op.validateCalls.Load(); got != 0 {
		t.Fatalf("operator validate calls = %d, want 0 for rejected requests", got)
	}
}

func TestHandlerAcceptsTypedValidateAndRejectsWrongMethod(t *testing.T) {
	op := &fakeOperator{}
	socketPath, stop := startAgentServer(t, op)
	defer stop()
	client := unixHTTPClient(socketPath)

	body := validValidateJSON(t, "")
	req, err := http.NewRequest(http.MethodPost, "http://unix/v1/validate", strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("validate request error = %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(resp.Body)
		t.Fatalf("validate status = %d, body=%s", resp.StatusCode, data)
	}
	if got := op.validateCalls.Load(); got != 1 {
		t.Fatalf("operator validate calls = %d, want 1", got)
	}

	wrong, err := http.NewRequest(http.MethodGet, "http://unix/v1/validate", nil)
	if err != nil {
		t.Fatal(err)
	}
	wrongResp, err := client.Do(wrong)
	if err != nil {
		t.Fatalf("wrong method request error = %v", err)
	}
	defer wrongResp.Body.Close()
	if wrongResp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("wrong method status = %d, want %d", wrongResp.StatusCode, http.StatusMethodNotAllowed)
	}
}

func TestResponseDTOsRedactSecretsExceptExplicitInspect(t *testing.T) {
	safeResponses := []any{
		HealthResponse{Status: "ok"},
		ValidateResponse{CandidateSHA256: strings.Repeat("b", 64)},
		ApplyResponse{PreviousSHA256: strings.Repeat("a", 64), AppliedSHA256: strings.Repeat("b", 64), BackupID: "backup-1", MainPID: 1},
		RollbackResponse{RestoredSHA256: strings.Repeat("a", 64), MainPID: 1},
	}
	for _, value := range safeResponses {
		encoded, err := json.Marshal(value)
		if err != nil {
			t.Fatalf("json.Marshal(%T) error = %v", value, err)
		}
		lower := bytes.ToLower(encoded)
		for _, forbidden := range [][]byte{[]byte("password"), []byte("ciphertext"), []byte("nonce"), []byte("secret")} {
			if bytes.Contains(lower, forbidden) {
				t.Fatalf("%T JSON unexpectedly contains %q: %s", value, forbidden, encoded)
			}
		}
	}

	inspect := InspectResponse{Credentials: []InspectCredential{{Username: "import.user", Password: "fixture-secret"}}}
	encoded, err := json.Marshal(inspect)
	if err != nil {
		t.Fatalf("json.Marshal(InspectResponse) error = %v", err)
	}
	if !bytes.Contains(encoded, []byte(`"password":"fixture-secret"`)) {
		t.Fatalf("explicit internal InspectResponse did not carry import password: %s", encoded)
	}
}

func TestHandlerMapsOperatorFailureWithoutLeakingErrorBody(t *testing.T) {
	op := failingOperator{err: errors.New("fixture-secret must never leave agent")}
	socketPath, stop := startAgentServer(t, op)
	defer stop()
	client := unixHTTPClient(socketPath)

	req, _ := http.NewRequest(http.MethodGet, "http://unix/v1/health", nil)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("client.Do() error = %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
	if bytes.Contains(body, []byte("fixture-secret")) {
		t.Fatalf("operator error leaked through HTTP response: %s", body)
	}
}

type failingOperator struct{ err error }

func (f failingOperator) Health(context.Context) (HealthResponse, error) {
	return HealthResponse{}, f.err
}
func (f failingOperator) Inspect(context.Context) (InspectResponse, error) {
	return InspectResponse{}, f.err
}
func (f failingOperator) Validate(context.Context, ValidateRequest) (ValidateResponse, error) {
	return ValidateResponse{}, f.err
}
func (f failingOperator) Apply(context.Context, ApplyRequest) (ApplyResponse, error) {
	return ApplyResponse{}, f.err
}
func (f failingOperator) Rollback(context.Context, RollbackRequest) (RollbackResponse, error) {
	return RollbackResponse{}, f.err
}

func startAgentServer(t *testing.T, op Operator) (string, func()) {
	t.Helper()
	socketPath := filepath.Join(t.TempDir(), "runtime-agent.sock")
	listener, err := ListenUnix(socketPath)
	if err != nil {
		t.Fatalf("ListenUnix() error = %v", err)
	}
	server := &http.Server{Handler: NewHandler(op), ReadHeaderTimeout: time.Second}
	done := make(chan struct{})
	go func() {
		_ = server.Serve(listener)
		close(done)
	}()
	return socketPath, func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = server.Shutdown(ctx)
		<-done
	}
}

func unixHTTPClient(socketPath string) *http.Client {
	transport := &http.Transport{
		DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
			var dialer net.Dialer
			return dialer.DialContext(ctx, "unix", socketPath)
		},
	}
	return &http.Client{Transport: transport, Timeout: 2 * time.Second}
}

func validValidateJSON(t *testing.T, extra string) string {
	t.Helper()
	base := `{"expected_caddy_sha256":"` + strings.Repeat("a", 64) + `","desired":{"revision":"rev-1","credentials":[{"id":"cred-1","username":"user.one","password":"safe password 123","status":"active"}]}`
	return base + extra + `}`
}
