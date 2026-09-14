package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/DashSaman/PV-NaivePanel/internal/customer"
)

// runBackfillSubscriptions issues active subscription tokens for every active
// customer whose current service term has none. Customers created before the
// direct-subscription token feature (or whose token issuance failed) show a
// permanently disabled "QR / link" button in the panel; this command repairs
// that state idempotently: customers that already hold an active token are
// skipped, and each issued token is attributed to the customer's creator so
// the audit trail stays intact.
//
// Usage:
//
//	pvnaive backfill-subscriptions [--dry-run] [--actor <uuid>]
func runBackfillSubscriptions(args []string) error {
	dryRun := false
	actorOverride := ""
	for _, arg := range args {
		switch {
		case arg == "--dry-run":
			dryRun = true
		case strings.HasPrefix(arg, "--actor="):
			actorOverride = strings.TrimSpace(strings.TrimPrefix(arg, "--actor="))
		case arg == "--actor":
			// fallthrough handled below when used as "--actor <uuid>"
		default:
			if actorOverride == "" && strings.TrimSpace(arg) != "" && !strings.HasPrefix(arg, "-") {
				actorOverride = strings.TrimSpace(arg)
			}
		}
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
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("PostgreSQL check: %w", err)
	}

	runtimeService, runtimeKey, err := buildRuntimeService(db, os.Getenv)
	if err != nil {
		return err
	}
	if runtimeService == nil || len(runtimeKey) != 32 {
		return fmt.Errorf("backfill requires the runtime credential service (set PVNAIVE_RUNTIME_KEY_FILE)")
	}
	customerService, _, err := buildCustomerServices(db, runtimeService, runtimeKey, os.Getenv)
	if err != nil {
		return err
	}
	if customerService == nil {
		return fmt.Errorf("backfill requires the customer service")
	}
	customerStore := customer.NewPostgresStore()

	readTx, err := db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return fmt.Errorf("begin read transaction: %w", err)
	}
	views, err := customerStore.ListCustomersTx(ctx, readTx)
	if err != nil {
		_ = readTx.Rollback()
		return fmt.Errorf("list customers: %w", err)
	}
	if err := readTx.Rollback(); err != nil {
		return fmt.Errorf("close read transaction: %w", err)
	}

	pending := 0
	issued := 0
	failed := 0
	for _, view := range views {
		if string(view.Status) != "active" {
			continue
		}
		if view.SubscriptionAvailable {
			continue
		}
		pending++
		actor := actorOverride
		if actor == "" {
			actor = strings.TrimSpace(view.CreatedByActorID)
		}
		if actor == "" {
			actor = strings.TrimSpace(view.AssignedActorID)
		}
		if actor == "" {
			failed++
			fmt.Printf("SKIP %s: no actor available for attribution (use --actor <uuid>)\n", view.Username)
			continue
		}
		if dryRun {
			fmt.Printf("WOULD ISSUE %s (term %s)\n", view.Username, view.ServiceTermID)
			continue
		}
		issueCtx, issueCancel := context.WithTimeout(ctx, 15*time.Second)
		tx, err := db.BeginTx(issueCtx, nil)
		if err != nil {
			issueCancel()
			failed++
			fmt.Printf("FAIL %s: begin transaction: %v\n", view.Username, err)
			continue
		}
		idempotencyKey := "backfill.subscription." + view.ServiceTermID
		_, err = customerService.RotateSubscription(issueCtx, tx, actor, idempotencyKey, view.UserID)
		if err != nil {
			_ = tx.Rollback()
			issueCancel()
			failed++
			fmt.Printf("FAIL %s: %v\n", view.Username, err)
			continue
		}
		if err := tx.Commit(); err != nil {
			issueCancel()
			failed++
			fmt.Printf("FAIL %s: commit: %v\n", view.Username, err)
			continue
		}
		issueCancel()
		issued++
		fmt.Printf("ISSUED %s (term %s)\n", view.Username, view.ServiceTermID)
	}

	verb := "issued"
	if dryRun {
		verb = "would issue"
	}
	fmt.Printf("backfill-subscriptions: %d eligible, %d %s, %d failed, %d skipped\n", pending, issued, verb, failed, pending-issued-failed)
	if failed > 0 && !dryRun {
		os.Exit(1)
	}
	return nil
}
