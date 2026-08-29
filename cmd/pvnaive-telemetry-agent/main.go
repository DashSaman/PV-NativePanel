package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/DashSaman/PV-NaivePanel/internal/telemetry"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	if err := run(); err != nil {
		log.Printf("PVNaive Telemetry Agent stopped: %v", err)
		os.Exit(1)
	}
}

func run() error {
	if os.Geteuid() == 0 {
		return errors.New("telemetry agent must not run as root")
	}
	dsn, err := databaseDSN(os.Getenv)
	if err != nil {
		return err
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("open PostgreSQL: %w", err)
	}
	defer db.Close()
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(5)
	db.SetConnMaxIdleTime(5 * time.Minute)
	db.SetConnMaxLifetime(30 * time.Minute)

	startupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(startupCtx); err != nil {
		return fmt.Errorf("PostgreSQL startup check: %w", err)
	}
	store, err := telemetry.NewPostgresStore(db)
	if err != nil {
		return err
	}
	listener, err := telemetry.ListenUnix(telemetry.DefaultTelemetrySocketPath)
	if err != nil {
		return fmt.Errorf("listen telemetry socket: %w", err)
	}
	defer listener.Close()
	defer os.Remove(telemetry.DefaultTelemetrySocketPath)

	server := &http.Server{
		Handler:           telemetry.NewTelemetryHandler(store),
		ReadHeaderTimeout: 3 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       30 * time.Second,
		MaxHeaderBytes:    8 << 10,
	}

	runCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	serveErr := make(chan error, 1)
	go func() {
		log.Printf("PVNaive Telemetry Agent listening on fixed Unix socket")
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
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown telemetry agent: %w", err)
		}
		return <-serveErr
	}
}

func databaseDSN(getenv func(string) string) (string, error) {
	host := getenv("PVNAIVE_DB_HOST")
	port := getenv("PVNAIVE_DB_PORT")
	name := getenv("PVNAIVE_DB_NAME")
	user := getenv("PVNAIVE_DB_USER")
	timeout := getenv("PVNAIVE_DB_CONNECT_TIMEOUT")
	for key, value := range map[string]string{
		"PVNAIVE_DB_HOST": host, "PVNAIVE_DB_PORT": port, "PVNAIVE_DB_NAME": name,
		"PVNAIVE_DB_USER": user, "PVNAIVE_DB_CONNECT_TIMEOUT": timeout,
	} {
		if value == "" {
			return "", fmt.Errorf("%s is required", key)
		}
	}
	if host != "127.0.0.1" {
		return "", fmt.Errorf("PVNAIVE_DB_HOST must be IPv4 loopback, got %q", host)
	}
	portNumber, err := strconv.Atoi(port)
	if err != nil || portNumber < 1 || portNumber > 65535 {
		return "", fmt.Errorf("invalid PVNAIVE_DB_PORT %q", port)
	}
	if name != "pvnaive" || user != "pvnaive_app" {
		return "", errors.New("telemetry agent requires the pvnaive database and pvnaive_app role")
	}
	timeoutSeconds, err := strconv.Atoi(timeout)
	if err != nil || timeoutSeconds < 1 || timeoutSeconds > 60 {
		return "", fmt.Errorf("invalid PVNAIVE_DB_CONNECT_TIMEOUT %q", timeout)
	}
	return fmt.Sprintf("host=%s port=%s dbname=%s user=%s connect_timeout=%s sslmode=disable", host, port, name, user, timeout), nil
}
