package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

const bindRequestContextSQL = `SELECT pvnaive.set_request_context($1)`
const clearRequestContextSQL = `SELECT pvnaive.clear_request_context()`

// execer keeps the SQL call testable. The exported functions below accept
// *sql.Tx explicitly, so callers cannot accidentally pass *sql.DB and lose the
// transaction-local PostgreSQL settings on a connection-pool hop.
type execer interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}

// BindRequestContext asks PostgreSQL to resolve an active management session
// and install a transaction-local, signed RLS context. The caller passes only
// a SHA-256 session-token hash; tenant IDs from HTTP input are never trusted.
func BindRequestContext(ctx context.Context, tx *sql.Tx, sessionTokenHash []byte) error {
	if tx == nil {
		return errors.New("database: nil transaction")
	}
	return bindRequestContext(ctx, tx, sessionTokenHash)
}

func bindRequestContext(ctx context.Context, tx execer, sessionTokenHash []byte) error {
	if len(sessionTokenHash) != 32 {
		return errors.New("database: session token hash must be 32 bytes")
	}
	if _, err := tx.ExecContext(ctx, bindRequestContextSQL, sessionTokenHash); err != nil {
		return fmt.Errorf("database: bind signed request context: %w", err)
	}
	return nil
}

// ClearRequestContext is defense-in-depth for long transactions. Transaction
// completion clears LOCAL settings automatically, but explicit clearing makes
// early reuse inside a transaction fail closed.
func ClearRequestContext(ctx context.Context, tx *sql.Tx) error {
	if tx == nil {
		return errors.New("database: nil transaction")
	}
	return clearRequestContext(ctx, tx)
}

func clearRequestContext(ctx context.Context, tx execer) error {
	if _, err := tx.ExecContext(ctx, clearRequestContextSQL); err != nil {
		return fmt.Errorf("database: clear request context: %w", err)
	}
	return nil
}
