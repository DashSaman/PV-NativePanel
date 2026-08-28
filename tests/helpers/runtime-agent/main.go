package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"sync"
	"syscall"
	"time"

	"github.com/DashSaman/PV-NaivePanel/internal/runtimeagent"
	"github.com/DashSaman/PV-NaivePanel/internal/runtimecred"
)

type rehearsalOperator struct {
	mu          sync.Mutex
	credentials []runtimeagent.InspectCredential
	sha         string
	backups     map[string][]runtimeagent.InspectCredential
	sequence    int
}

func newRehearsalOperator() *rehearsalOperator {
	credentials := []runtimeagent.InspectCredential{{Username: "legacy.user", Password: "legacy-pass"}}
	return &rehearsalOperator{
		credentials: credentials,
		sha:         credentialSHA(credentials),
		backups:     map[string][]runtimeagent.InspectCredential{},
	}
}

func (o *rehearsalOperator) Health(context.Context) (runtimeagent.HealthResponse, error) {
	return runtimeagent.HealthResponse{Status: "ok"}, nil
}

func (o *rehearsalOperator) Inspect(context.Context) (runtimeagent.InspectResponse, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	return runtimeagent.InspectResponse{CaddySHA256: o.sha, Credentials: cloneCredentials(o.credentials)}, nil
}

func (o *rehearsalOperator) Validate(_ context.Context, request runtimeagent.ValidateRequest) (runtimeagent.ValidateResponse, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if request.ExpectedCaddySHA256 != o.sha {
		return runtimeagent.ValidateResponse{}, errors.New("stale expected SHA")
	}
	candidate, err := desiredCredentials(request.Desired)
	if err != nil {
		return runtimeagent.ValidateResponse{}, err
	}
	return runtimeagent.ValidateResponse{CandidateSHA256: credentialSHA(candidate)}, nil
}

func (o *rehearsalOperator) Apply(_ context.Context, request runtimeagent.ApplyRequest) (runtimeagent.ApplyResponse, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if request.ExpectedCaddySHA256 != o.sha {
		return runtimeagent.ApplyResponse{}, errors.New("stale expected SHA")
	}
	candidate, err := desiredCredentials(request.Desired)
	if err != nil {
		return runtimeagent.ApplyResponse{}, err
	}
	previousSHA := o.sha
	o.sequence++
	backupID := "backup-rehearsal-" + time.Now().UTC().Format("20060102T150405.000000000")
	o.backups[backupID] = cloneCredentials(o.credentials)
	o.credentials = candidate
	o.sha = credentialSHA(candidate)
	return runtimeagent.ApplyResponse{
		PreviousSHA256: previousSHA,
		AppliedSHA256:  o.sha,
		BackupID:       backupID,
		MainPID:        4242,
		NRestarts:      0,
	}, nil
}

func (o *rehearsalOperator) Rollback(_ context.Context, request runtimeagent.RollbackRequest) (runtimeagent.RollbackResponse, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	backup, ok := o.backups[request.BackupID]
	if !ok {
		return runtimeagent.RollbackResponse{}, errors.New("unknown backup")
	}
	o.credentials = cloneCredentials(backup)
	o.sha = credentialSHA(o.credentials)
	return runtimeagent.RollbackResponse{RestoredSHA256: o.sha, MainPID: 4242, NRestarts: 0}, nil
}

func desiredCredentials(input runtimeagent.DesiredStateInput) ([]runtimeagent.InspectCredential, error) {
	if input.Revision == "" || len(input.Credentials) == 0 {
		return nil, errors.New("empty desired state")
	}
	out := make([]runtimeagent.InspectCredential, 0, len(input.Credentials))
	seen := map[string]struct{}{}
	for _, credential := range input.Credentials {
		if credential.Status != runtimecred.CredentialActive {
			continue
		}
		if _, exists := seen[credential.Username]; exists {
			return nil, errors.New("duplicate username")
		}
		seen[credential.Username] = struct{}{}
		out = append(out, runtimeagent.InspectCredential{Username: credential.Username, Password: credential.Password})
	}
	if len(out) == 0 {
		return nil, errors.New("no active credential")
	}
	return out, nil
}

func credentialSHA(credentials []runtimeagent.InspectCredential) string {
	canonical := cloneCredentials(credentials)
	sort.SliceStable(canonical, func(i, j int) bool { return canonical[i].Username < canonical[j].Username })
	hash := sha256.New()
	for _, credential := range canonical {
		hash.Write([]byte(credential.Username))
		hash.Write([]byte{0})
		hash.Write([]byte(credential.Password))
		hash.Write([]byte{0xff})
	}
	return hex.EncodeToString(hash.Sum(nil))
}

func cloneCredentials(input []runtimeagent.InspectCredential) []runtimeagent.InspectCredential {
	out := make([]runtimeagent.InspectCredential, len(input))
	copy(out, input)
	return out
}

func main() {
	socketPath := os.Getenv("PVNAIVE_RUNTIME_AGENT_SOCKET")
	if socketPath == "" {
		log.Fatal("PVNAIVE_RUNTIME_AGENT_SOCKET is required")
	}
	listener, err := runtimeagent.ListenUnix(socketPath)
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()
	if err := os.Chmod(socketPath, 0660); err != nil {
		log.Fatal(err)
	}

	server := &http.Server{
		Handler:           runtimeagent.NewHandler(newRehearsalOperator()),
		ReadHeaderTimeout: 3 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       10 * time.Second,
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()
	if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}
