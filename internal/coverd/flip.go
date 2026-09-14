package coverd

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)

// FlipConfig is the validated configuration for exposing the cover site of
// THIS node (R6-FLIP-001). The flip is deliberately explicit: nothing in
// the default environment enables it.
type FlipConfig struct {
	// Enabled turns the cover listener on. Default OFF — root keeps 404.
	Enabled bool
	// NodeID is the stable identity of THIS node (persona + content key
	// derivation). Required when enabled.
	NodeID string
	// Listen is the loopback bind address Caddy proxies to. Must be an
	// explicit loopback address — the cover service is never exposed
	// directly.
	Listen string
	// PersonaOverride pins a persona id; empty means derive (stored
	// persona first, then the stable deterministic assignment).
	PersonaOverride string
}

// Validate enforces the flip invariants.
func (c FlipConfig) Validate() error {
	if !c.Enabled {
		return nil
	}
	nodeID := strings.TrimSpace(c.NodeID)
	if len(nodeID) < 1 || len(nodeID) > 120 {
		return errors.New("coverd: PVNAIVE_COVERD_NODE_ID must be 1..120 characters when the cover is enabled")
	}
	if c.Listen == "" {
		return errors.New("coverd: PVNAIVE_COVERD_LISTEN is required when the cover is enabled")
	}
	host, port, err := net.SplitHostPort(c.Listen)
	if err != nil {
		return fmt.Errorf("coverd: invalid PVNAIVE_COVERD_LISTEN: %w", err)
	}
	if host != "127.0.0.1" && host != "::1" {
		return fmt.Errorf("coverd: PVNAIVE_COVERD_LISTEN must bind loopback (got %q) — exposure happens only through the reverse proxy", host)
	}
	if port == "" || port == "0" {
		return errors.New("coverd: PVNAIVE_COVERD_LISTEN requires a non-zero port")
	}
	if c.PersonaOverride != "" {
		if _, ok := PersonaByID(PersonaID(c.PersonaOverride)); !ok {
			return fmt.Errorf("coverd: unknown persona override %q", c.PersonaOverride)
		}
	}
	return nil
}

// PersonaForConfig resolves the effective persona for the flip:
// override > stored > deterministic node assignment.
func PersonaForConfig(config FlipConfig, stored func() (string, bool)) (Persona, error) {
	if err := config.Validate(); err != nil {
		return Persona{}, err
	}
	if config.PersonaOverride != "" {
		persona, ok := PersonaByID(PersonaID(config.PersonaOverride))
		if !ok {
			return Persona{}, fmt.Errorf("coverd: unknown persona override %q", config.PersonaOverride)
		}
		return persona, nil
	}
	if stored != nil {
		if raw, ok := stored(); ok {
			if persona, found := PersonaByID(PersonaID(strings.TrimSpace(raw))); found {
				return persona, nil
			}
		}
	}
	return PersonaForNode(strings.TrimSpace(config.NodeID)), nil
}

// NewNodeServer assembles the cover HTTP handler for one node: renderer
// with the resolved persona, content from the provided store (the DB
// snapshot store in production; fakes in rehearsals).
func NewNodeServer(config FlipConfig, store ContentStore, stored func() (string, bool), now func() time.Time) (*Server, error) {
	persona, err := PersonaForConfig(config, stored)
	if err != nil {
		return nil, err
	}
	if store == nil {
		return nil, errors.New("coverd: content store is required")
	}
	renderer := &Renderer{
		NodeID:  strings.TrimSpace(config.NodeID),
		Persona: persona,
		Store:   store,
	}
	if now != nil {
		renderer.Now = now
	}
	return &Server{Renderer: renderer}, nil
}

// Serve runs the cover listener on an already-validated loopback address
// until ctx is cancelled, then shuts down gracefully. Plain HTTP: the
// bind is loopback-only and TLS terminates at the reverse proxy.
func Serve(ctx context.Context, server *http.Server, logger func(string, ...any)) error {
	if logger == nil {
		logger = log.Printf
	}
	serveErr := make(chan error, 1)
	go func() {
		logger("coverd: cover site listening on %s", server.Addr)
		err := server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			serveErr <- err
			return
		}
		serveErr <- nil
	}()
	select {
	case err := <-serveErr:
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("coverd: graceful shutdown: %w", err)
		}
		return <-serveErr
	}
}
