package fleetpull

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"
)

// ListenerConfig configures the dedicated sibling mTLS control listener.
// CertFile/KeyFile are the registry's own server certificate; ClientCA
// verifies sibling agent client certificates. All fields are mandatory —
// the listener refuses to start half-configured (fail-closed).
type ListenerConfig struct {
	Listen   string
	CertFile string
	KeyFile  string
	ClientCA string
	// Handler is the (already store-wired) fleetpull handler.
	Handler http.Handler
	// MinTLSVersion defaults to TLS 1.2 when zero.
	MinTLSVersion uint16
}

// NewListener builds the http.Server with require-and-verify mutual TLS.
// The server certificate is loaded eagerly so a broken pair fails startup,
// not the first handshake.
func NewListener(config ListenerConfig) (*http.Server, error) {
	if config.Handler == nil {
		return nil, errors.New("fleetpull: handler is required")
	}
	if config.Listen == "" || config.CertFile == "" || config.KeyFile == "" || config.ClientCA == "" {
		return nil, errors.New("fleetpull: listen, cert, key and client CA are all required")
	}
	if _, _, err := net.SplitHostPort(config.Listen); err != nil {
		return nil, fmt.Errorf("fleetpull: invalid listen address: %w", err)
	}
	serverCert, err := tls.LoadX509KeyPair(config.CertFile, config.KeyFile)
	if err != nil {
		return nil, fmt.Errorf("fleetpull: load server certificate: %w", err)
	}
	clientCAPEM, err := os.ReadFile(config.ClientCA)
	if err != nil {
		return nil, fmt.Errorf("fleetpull: read client CA: %w", err)
	}
	clientCAPool := x509.NewCertPool()
	if !clientCAPool.AppendCertsFromPEM(clientCAPEM) {
		return nil, errors.New("fleetpull: client CA file contains no usable certificates")
	}
	minVersion := uint16(tls.VersionTLS12)
	if config.MinTLSVersion != 0 {
		minVersion = config.MinTLSVersion
	}
	tlsConfig := &tls.Config{
		MinVersion:   minVersion,
		Certificates: []tls.Certificate{serverCert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    clientCAPool,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
		},
	}
	server := &http.Server{
		Addr:              config.Listen,
		Handler:           config.Handler,
		TLSConfig:         tlsConfig,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    16 << 10,
	}
	return server, nil
}

// Serve runs ListenAndServeTLS until ctx is cancelled, then shuts down
// gracefully. It returns the serve error (nil on graceful shutdown).
func Serve(ctx context.Context, server *http.Server, logger func(string, ...any)) error {
	if logger == nil {
		logger = log.Printf
	}
	serveErr := make(chan error, 1)
	go func() {
		logger("fleetpull: mTLS control listener on %s", server.Addr)
		// Certificates live in TLSConfig; empty names are the documented
		// way to reuse them.
		err := server.ListenAndServeTLS("", "")
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
			return fmt.Errorf("fleetpull: graceful shutdown: %w", err)
		}
		return <-serveErr
	}
}
