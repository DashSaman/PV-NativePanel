package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"os/user"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/DashSaman/PV-NaivePanel/internal/runtimeagent"
)

const runtimeDir = "/run/pvnaive"

func main() {
	if err := run(); err != nil {
		log.Printf("PVNaive Runtime Agent stopped: %v", err)
		os.Exit(1)
	}
}

func run() error {
	if os.Geteuid() != 0 {
		return errors.New("runtime agent must run as root")
	}
	group, err := user.LookupGroup("pvnaive")
	if err != nil {
		return fmt.Errorf("lookup pvnaive group: %w", err)
	}
	gid, err := strconv.Atoi(group.Gid)
	if err != nil {
		return fmt.Errorf("invalid pvnaive group id: %w", err)
	}

	if err := os.MkdirAll(runtimeDir, 0750); err != nil {
		return fmt.Errorf("create runtime directory: %w", err)
	}
	if err := os.Chown(runtimeDir, 0, gid); err != nil {
		return fmt.Errorf("chown runtime directory: %w", err)
	}
	if err := os.Chmod(runtimeDir, 0750); err != nil {
		return fmt.Errorf("chmod runtime directory: %w", err)
	}
	if filepath.Dir(runtimeagent.DefaultSocketPath) != runtimeDir {
		return errors.New("compiled runtime socket path escaped fixed runtime directory")
	}

	listener, err := runtimeagent.ListenUnix(runtimeagent.DefaultSocketPath)
	if err != nil {
		return fmt.Errorf("listen runtime socket: %w", err)
	}
	defer listener.Close()
	if err := os.Chown(runtimeagent.DefaultSocketPath, 0, gid); err != nil {
		return fmt.Errorf("chown runtime socket: %w", err)
	}
	if err := os.Chmod(runtimeagent.DefaultSocketPath, 0660); err != nil {
		return fmt.Errorf("chmod runtime socket: %w", err)
	}

	server := &http.Server{
		Handler:           runtimeagent.NewHandler(unavailableOperator{}),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       30 * time.Second,
		MaxHeaderBytes:    16 << 10,
	}

	runCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	serveErr := make(chan error, 1)
	go func() {
		log.Printf("PVNaive Runtime Agent listening on Unix socket")
		err := server.Serve(listener)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			serveErr <- err
			return
		}
		serveErr <- nil
	}()

	select {
	case err := <-serveErr:
		return err
	case <-runCtx.Done():
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			return fmt.Errorf("shutdown runtime agent: %w", err)
		}
		return <-serveErr
	}
}

// unavailableOperator is deliberately fail-closed until PVN-023 wires the
// fixed Caddy validate/apply/rollback implementation. It keeps this protocol
// slice buildable without granting a placeholder privileged capability.
type unavailableOperator struct{}

func (unavailableOperator) Health(context.Context) (runtimeagent.HealthResponse, error) {
	return runtimeagent.HealthResponse{}, errors.New("runtime operator unavailable")
}
func (unavailableOperator) Inspect(context.Context) (runtimeagent.InspectResponse, error) {
	return runtimeagent.InspectResponse{}, errors.New("runtime operator unavailable")
}
func (unavailableOperator) Validate(context.Context, runtimeagent.ValidateRequest) (runtimeagent.ValidateResponse, error) {
	return runtimeagent.ValidateResponse{}, errors.New("runtime operator unavailable")
}
func (unavailableOperator) Apply(context.Context, runtimeagent.ApplyRequest) (runtimeagent.ApplyResponse, error) {
	return runtimeagent.ApplyResponse{}, errors.New("runtime operator unavailable")
}
func (unavailableOperator) Rollback(context.Context, runtimeagent.RollbackRequest) (runtimeagent.RollbackResponse, error) {
	return runtimeagent.RollbackResponse{}, errors.New("runtime operator unavailable")
}
