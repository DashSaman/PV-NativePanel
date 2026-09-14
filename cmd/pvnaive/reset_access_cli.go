package main

import (
	"bufio"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/DashSaman/PV-NaivePanel/internal/auth"
	"golang.org/x/term"
)

// adminResetAccess implements the on-host recovery CLI (ACCESS-002):
//
//	pvnaive admin reset-access
//
// It runs ONLY on the host (root/pvnaive, TTY required), never accepts the
// password from argv or env, prompts twice, applies through the same SECURITY
// DEFINER path as the panel, and prints the resulting effective state. It
// never writes secrets to stdout/stderr/audit.
func adminResetAccess(args []string) error {
	if len(args) != 0 {
		return errors.New("usage: pvnaive admin reset-access")
	}
	if os.Geteuid() == 0 {
		fmt.Fprintln(os.Stderr, "WARNING: running as root; the API normally runs as the pvnaive system user.")
	}
	if !termIsInteractive() {
		return errors.New("reset-access requires an interactive terminal; passwords are never read from argv/env/pipes")
	}

	dsn, err := databaseDSN(os.Getenv)
	if err != nil {
		return fmt.Errorf("database environment: %w", err)
	}

	username, err := promptLine("Admin username (3..64 chars): ", false)
	if err != nil {
		return err
	}
	password, err := promptNewPasswordTwice()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	store, closeStore, err := openAuthStore(ctx, dsn)
	if err != nil {
		return err
	}
	defer closeStore()

	hash, err := auth.HashPassword(password)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	if err := store.UpsertPanelSettings(ctx, auth.PanelAccessInput{
		AdminUsername: strings.TrimSpace(username),
		PasswordHash:  hash,
		BasePath:      "/panel",
		ListenPort:    8080,
		GraceMinutes:  10,
		ExposureMode:  "reverse_proxy",
	}, nil); err != nil {
		return fmt.Errorf("apply reset: %w", err)
	}
	_ = store.AppendAudit(ctx, nil, "panel.access.recovery", "success", "host CLI reset-access")

	settings, err := store.ReadPanelSettings(ctx)
	if err != nil || settings == nil {
		return errors.New("post-apply verification failed")
	}
	fmt.Println("RESET_ACCESS=APPLIED")
	fmt.Printf("ADMIN_USERNAME=%s\n", settings.AdminUsername)
	fmt.Printf("BASE_PATH=%s\n", settings.BasePath)
	fmt.Printf("LISTEN_PORT=%d\n", settings.ListenPort)
	fmt.Println("NOTE: restart the pvnaive-api service so the new credentials are loaded.")
	return nil
}

func openAuthStore(ctx context.Context, dsn string) (*auth.Store, func(), error) {
	db, err := openPgxDB(dsn)
	if err != nil {
		return nil, nil, err
	}
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := db.PingContext(pingCtx); err != nil {
		_ = db.Close()
		return nil, nil, fmt.Errorf("PostgreSQL ping: %w", err)
	}
	store, err := auth.NewStore(db)
	if err != nil {
		_ = db.Close()
		return nil, nil, err
	}
	return store, func() { _ = db.Close() }, nil
}

func openPgxDB(dsn string) (*sql.DB, error) {
	return sql.Open("pgx", dsn)
}

func promptNewPasswordTwice() (string, error) {
	first, err := promptPassword("New admin password (min 14 chars): ")
	if err != nil {
		return "", err
	}
	second, err := promptPassword("Repeat new admin password: ")
	if err != nil {
		return "", err
	}
	if first != second {
		return "", errors.New("passwords do not match")
	}
	if len(first) < 14 {
		return "", errors.New("password must be at least 14 characters")
	}
	return first, nil
}

func promptLine(label string, _ bool) (string, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(label)
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("read input: %w", err)
	}
	line = strings.TrimRight(line, "\r\n")
	if strings.TrimSpace(line) == "" {
		return "", errors.New("empty input")
	}
	return line, nil
}

func promptPassword(label string) (string, error) {
	fmt.Print(label)
	raw, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", fmt.Errorf("read password: %w", err)
	}
	fmt.Println()
	value := strings.TrimRight(string(raw), "\r\n")
	if value == "" {
		return "", errors.New("empty password")
	}
	return value, nil
}

func termIsInteractive() bool {
	info, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}
