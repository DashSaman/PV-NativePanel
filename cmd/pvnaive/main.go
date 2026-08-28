package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/DashSaman/PV-NaivePanel/internal/auth"
	"github.com/DashSaman/PV-NaivePanel/internal/httpapi"
	_ "github.com/jackc/pgx/v5/stdlib"
)

const (
	defaultListen      = "127.0.0.1:8080"
	defaultAuthKeyFile = "/etc/pvnaive/auth.key"
)

func main() {
	if err := run(); err != nil {
		log.Printf("PVNaive API stopped: %v", err)
		os.Exit(1)
	}
}

func run() error {
	listen, err := validatedListenAddress(os.Getenv("PVNAIVE_LISTEN"))
	if err != nil {
		return err
	}
	keyFile := os.Getenv("PVNAIVE_AUTH_KEY_FILE")
	if keyFile == "" {
		keyFile = defaultAuthKeyFile
	}
	mfaKey, err := os.ReadFile(keyFile)
	if err != nil {
		return fmt.Errorf("read authentication key: %w", err)
	}
	if len(mfaKey) != 32 {
		return fmt.Errorf("authentication key must be exactly 32 bytes, got %d", len(mfaKey))
	}
	defer zeroBytes(mfaKey)

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

	startupCtx, startupCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer startupCancel()
	if err := db.PingContext(startupCtx); err != nil {
		return fmt.Errorf("PostgreSQL startup check: %w", err)
	}

	store, err := auth.NewStore(db)
	if err != nil {
		return err
	}
	service, err := auth.NewService(store, mfaKey)
	if err != nil {
		return err
	}

	handler := httpapi.NewServer(httpapi.ServerConfig{AuthService: service, AuthStore: store, MFAKey: mfaKey})
	server := &http.Server{
		Addr:              listen,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    64 << 10,
	}

	runCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	serveErr := make(chan error, 1)
	go func() {
		log.Printf("PVNaive API listening on %s", listen)
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
	case <-runCtx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("graceful shutdown: %w", err)
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
		"PVNAIVE_DB_HOST":            host,
		"PVNAIVE_DB_PORT":            port,
		"PVNAIVE_DB_NAME":            name,
		"PVNAIVE_DB_USER":            user,
		"PVNAIVE_DB_CONNECT_TIMEOUT": timeout,
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
	if name != "pvnaive" {
		return "", fmt.Errorf("PVNAIVE_DB_NAME must be pvnaive, got %q", name)
	}
	if user != "pvnaive_app" {
		return "", fmt.Errorf("PVNAIVE_DB_USER must be pvnaive_app, got %q", user)
	}
	timeoutSeconds, err := strconv.Atoi(timeout)
	if err != nil || timeoutSeconds < 1 || timeoutSeconds > 60 {
		return "", fmt.Errorf("invalid PVNAIVE_DB_CONNECT_TIMEOUT %q", timeout)
	}

	// pgx still reads PGPASSFILE from the process environment. The explicit
	// connection string prevents it from silently falling back to the OS user
	// or an empty database name when only PVNAIVE_DB_* variables are present.
	return fmt.Sprintf(
		"host=%s port=%s dbname=%s user=%s connect_timeout=%s sslmode=disable",
		host, port, name, user, timeout,
	), nil
}

func validatedListenAddress(value string) (string, error) {
	if value == "" {
		value = defaultListen
	}
	host, port, err := net.SplitHostPort(value)
	if err != nil {
		return "", fmt.Errorf("invalid PVNAIVE_LISTEN: %w", err)
	}
	if host != "127.0.0.1" {
		return "", fmt.Errorf("PVNAIVE_LISTEN must use IPv4 loopback, got %q", host)
	}
	if port == "" || port == "0" {
		return "", errors.New("PVNAIVE_LISTEN requires a non-zero port")
	}
	return value, nil
}

func zeroBytes(value []byte) {
	for i := range value {
		value[i] = 0
	}
}
