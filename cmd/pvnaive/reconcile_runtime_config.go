package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/DashSaman/PV-NaivePanel/internal/runtimeconfig"
)

const (
	reconcileDefaultConfigPath = "/etc/caddy/Caddyfile"
	reconcileDefaultBinary     = "/opt/pvnaive/bin/caddy"
	reconcileValidateTimeout   = 20 * time.Second
	// Local socket + superuser role: the credential snapshot must see
	// RLS-protected rows without a request context (trusted local auth,
	// same trust boundary migrations already rely on).
	reconcileDefaultDBHost = "/var/run/postgresql"
	reconcileDefaultDBUser = "postgres"
)

// runReconcileRuntimeConfig implements `pvnaive reconcile-runtime-config`:
// the boot-time merge of the rendered proxy config's forward_proxy
// credential block with the active credentials stored in the database
// (DEPLOY-001). The entrypoint invokes it after the initial template render
// and before the proxy process starts, so a corrected file is picked up by
// a plain process start without any reload.
//
// The database connection deliberately uses the local Unix socket with the
// postgres superuser role (trusted local auth, exactly like migrations):
// pvnaive.naive_runtime_credentials is RLS-protected and the least-privilege
// app role sees no rows without a request context, which does not exist at
// boot time. The superuser snapshot is read-only and never mutates rows.
func runReconcileRuntimeConfig(args []string) error {
	flags := flag.NewFlagSet("reconcile-runtime-config", flag.ContinueOnError)
	configPath := flags.String("config", reconcileDefaultConfigPath, "runtime proxy config file to reconcile")
	caddyBinary := flags.String("caddy-binary", "", "pinned proxy binary used to validate the candidate config")
	dbHost := flags.String("db-host", reconcileDefaultDBHost, "PostgreSQL host (Unix socket directory for the superuser snapshot)")
	dbPort := flags.String("db-port", "5432", "PostgreSQL port")
	dbName := flags.String("db-name", "pvnaive", "PostgreSQL database name")
	dbUser := flags.String("db-user", reconcileDefaultDBUser, "PostgreSQL role for the read-only snapshot")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return errors.New("reconcile-runtime-config takes no positional arguments")
	}
	binary := *caddyBinary
	if binary == "" {
		binary = os.Getenv("PVNAIVE_RUNTIME_CADDY_BINARY")
	}
	if binary == "" {
		binary = reconcileDefaultBinary
	}

	keyFile := os.Getenv("PVNAIVE_RUNTIME_KEY_FILE")
	if keyFile == "" {
		return errors.New("PVNAIVE_RUNTIME_KEY_FILE is required")
	}
	key, err := os.ReadFile(keyFile)
	if err != nil {
		return fmt.Errorf("read runtime key: %w", err)
	}
	if len(key) != 32 {
		zeroBytes(key)
		return fmt.Errorf("runtime key must be exactly 32 bytes, got %d", len(key))
	}
	defer zeroBytes(key)

	dsn := fmt.Sprintf("host=%s port=%s dbname=%s user=%s connect_timeout=5 sslmode=disable",
		*dbHost, *dbPort, *dbName, *dbUser)
	db, err := openReconcileDB(dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), reconcileValidateTimeout)
	defer cancel()
	validate := func(candidate string) error {
		validationCtx, validationCancel := context.WithTimeout(context.Background(), reconcileValidateTimeout)
		defer validationCancel()
		output, err := exec.CommandContext(validationCtx, binary, "validate", "--config", candidate, "--adapter", "caddyfile").CombinedOutput()
		if err != nil {
			return fmt.Errorf("candidate validation failed: %s", firstLine(output))
		}
		return nil
	}

	result, err := runtimeconfig.ReconcileRuntimeConfigFile(ctx, db, key, *configPath, validate)
	if err != nil {
		return err
	}
	fmt.Printf("RECONCILE_RESULT=CHANGED:%t CREDENTIALS:%d\n", result.Changed, result.Credentials)
	return nil
}

func openReconcileDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("open PostgreSQL: %w", err)
	}
	pingCtx, pingCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer pingCancel()
	if err := db.PingContext(pingCtx); err != nil {
		db.Close()
		return nil, fmt.Errorf("PostgreSQL startup check: %w", err)
	}
	return db, nil
}

func firstLine(output []byte) string {
	for i := range output {
		if output[i] == '\n' {
			return string(output[:i])
		}
	}
	return string(output)
}
