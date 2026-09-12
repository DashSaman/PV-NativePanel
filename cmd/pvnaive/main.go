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
	"strings"
	"syscall"
	"time"

	"github.com/DashSaman/PV-NaivePanel/internal/auth"
	"github.com/DashSaman/PV-NaivePanel/internal/customer"
	"github.com/DashSaman/PV-NaivePanel/internal/httpapi"
	"github.com/DashSaman/PV-NaivePanel/internal/runtimeagent"
	"github.com/DashSaman/PV-NaivePanel/internal/runtimecred"
	"github.com/DashSaman/PV-NaivePanel/internal/sessioncontrol"
	"github.com/DashSaman/PV-NaivePanel/internal/subscription"
	"github.com/DashSaman/PV-NaivePanel/internal/telemetry"
	_ "github.com/jackc/pgx/v5/stdlib"
)

const (
	defaultListen             = "127.0.0.1:8080"
	defaultAuthKeyFile        = "/etc/pvnaive/auth.key"
	defaultRuntimeKeyID       = "runtime-v1"
	defaultRuntimeAgentSocket = runtimeagent.DefaultSocketPath
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "doctor" {
		if err := runDoctor(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "PVNaive doctor: %v\n", err)
			os.Exit(1)
		}
		return
	}
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
	accountingStore, err := telemetry.NewPostgresStore(db)
	if err != nil {
		return err
	}

	runtimeService, runtimeKey, err := buildRuntimeService(db, os.Getenv)
	if err != nil {
		return err
	}
	if runtimeKey != nil {
		defer zeroBytes(runtimeKey)
	}

	customerService, subscriptionService, err := buildCustomerServices(db, runtimeService, runtimeKey, os.Getenv)
	if err != nil {
		return err
	}
	subscriptionHost, err := subscriptionProxyHost(os.Getenv, customerService != nil && subscriptionService != nil)
	if err != nil {
		return err
	}
	systemStatus := buildSystemStatusProvider(db, os.Getenv)
	expectedSchema, err := expectedSchemaVersion(os.Getenv)
	if err != nil {
		return err
	}
	readinessProbe := httpapi.NewDBReadinessProbe(db, expectedSchema)
	sessionControlSocket := strings.TrimSpace(os.Getenv("PVNAIVE_SESSION_CONTROL_SOCKET"))
	if sessionControlSocket == "" {
		sessionControlSocket = sessioncontrol.DefaultSocketPath
	}
	sessionController := sessioncontrol.NewClient(sessionControlSocket)

	var periodicResetConfig *periodicUsageResetConfig
	if customerService != nil {
		cfg, err := periodicUsageResetConfigFromEnv(os.Getenv)
		if err != nil {
			return fmt.Errorf("periodic usage reset scheduler configuration: %w", err)
		}
		periodicResetConfig = &cfg
	}

	handler := httpapi.NewServer(httpapi.ServerConfig{
		AuthService:           service,
		AuthStore:             store,
		MFAKey:                mfaKey,
		RuntimeService:        runtimeService,
		CustomerService:       customerService,
		AccountingStore:       accountingStore,
		SessionController:     sessionController,
		SubscriptionService:   subscriptionService,
		SubscriptionProxyHost: subscriptionHost,
		SystemStatus:          systemStatus,
		ReadinessProbe:        readinessProbe,
	})
	handler = httpapi.WithOperationalMiddleware(handler)
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
	if periodicResetConfig != nil {
		go runPeriodicUsageResetScheduler(
			runCtx,
			periodicUsageResetDBExecutor{db: db},
			*periodicResetConfig,
			log.Printf,
		)
	}
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

func buildRuntimeService(db *sql.DB, getenv func(string) string) (*runtimecred.Service, []byte, error) {
	keyFile := getenv("PVNAIVE_RUNTIME_KEY_FILE")
	if keyFile == "" {
		// Runtime management remains optional for isolated auth rehearsals.
		return nil, nil, nil
	}
	key, err := os.ReadFile(keyFile)
	if err != nil {
		return nil, nil, fmt.Errorf("read runtime key: %w", err)
	}
	if len(key) != 32 {
		zeroBytes(key)
		return nil, nil, fmt.Errorf("runtime key must be exactly 32 bytes, got %d", len(key))
	}
	store, err := runtimecred.NewStore(db)
	if err != nil {
		zeroBytes(key)
		return nil, nil, err
	}
	keyID := getenv("PVNAIVE_RUNTIME_KEY_ID")
	if keyID == "" {
		keyID = defaultRuntimeKeyID
	}
	socketPath := getenv("PVNAIVE_RUNTIME_AGENT_SOCKET")
	if socketPath == "" {
		socketPath = defaultRuntimeAgentSocket
	}
	client := runtimeagent.NewClient(socketPath)
	adapter := runtimeagent.NewRuntimeCredAdapter(client)
	service, err := runtimecred.NewService(store, adapter, key, keyID)
	if err != nil {
		zeroBytes(key)
		return nil, nil, err
	}
	return service, key, nil
}

func buildCustomerServices(
	db *sql.DB,
	runtimeService *runtimecred.Service,
	runtimeKey []byte,
	getenv func(string) string,
) (*customer.Service, *subscription.Service, error) {
	if runtimeService == nil {
		return nil, nil, nil
	}
	if db == nil || len(runtimeKey) != 32 {
		return nil, nil, errors.New("customer services require PostgreSQL and the runtime encryption key")
	}
	keyID := getenv("PVNAIVE_RUNTIME_KEY_ID")
	if keyID == "" {
		keyID = defaultRuntimeKeyID
	}
	customerStore := customer.NewPostgresStore()
	createRuntime := func(ctx context.Context, tx *sql.Tx, actorID, idempotencyKey string, input runtimecred.CreateInput) (customer.RuntimeMutation, error) {
		return runtimeService.Create(ctx, tx, actorID, idempotencyKey, input)
	}
	customerService := customer.NewServiceWithTokenRecovery(customerStore, createRuntime, time.Now, runtimeKey, keyID)
	if err := customerService.ConfigureRuntimeOperations(
		func(ctx context.Context, tx *sql.Tx, actorID, idempotencyKey string, input runtimecred.UpdateInput) (customer.RuntimeMutation, error) {
			return runtimeService.Update(ctx, tx, actorID, idempotencyKey, input)
		},
		func(ctx context.Context, tx *sql.Tx, actorID, idempotencyKey string, input runtimecred.RotateInput) (customer.RuntimeMutation, error) {
			return runtimeService.Rotate(ctx, tx, actorID, idempotencyKey, input)
		},
		func(ctx context.Context, tx *sql.Tx, actorID, idempotencyKey string, input runtimecred.RevokeInput) (customer.RuntimeMutation, error) {
			return runtimeService.Revoke(ctx, tx, actorID, idempotencyKey, input)
		},
	); err != nil {
		return nil, nil, err
	}
	subscriptionService, err := subscription.NewService(subscription.NewPostgresStore(db), runtimeKey, keyID)
	if err != nil {
		return nil, nil, err
	}
	return customerService, subscriptionService, nil
}

func subscriptionProxyHost(getenv func(string) string, enabled bool) (string, error) {
	if !enabled {
		return "", nil
	}
	host := strings.TrimSpace(getenv("PVNAIVE_NAIVE_PUBLIC_HOST"))
	if host == "" {
		return "", errors.New("PVNAIVE_NAIVE_PUBLIC_HOST is required when Naive customer subscriptions are enabled")
	}
	if _, err := subscription.BuildNaiveURI("pvnaive-probe", "probe-secret", host); err != nil {
		return "", fmt.Errorf("invalid PVNAIVE_NAIVE_PUBLIC_HOST: %w", err)
	}
	return host, nil
}

func expectedSchemaVersion(getenv func(string) string) (int, error) {
	raw := strings.TrimSpace(getenv("PVNAIVE_EXPECTED_SCHEMA_VERSION"))
	version, err := strconv.Atoi(raw)
	if err != nil || version <= 0 {
		return 0, fmt.Errorf("PVNAIVE_EXPECTED_SCHEMA_VERSION must be a positive integer")
	}
	return version, nil
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
