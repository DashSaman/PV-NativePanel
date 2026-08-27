package database

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"
)

type fakeTx struct {
	query string
	args  []any
	err   error
}

func (f *fakeTx) ExecContext(_ context.Context, query string, args ...any) (sql.Result, error) {
	f.query = query
	f.args = args
	return nil, f.err
}

func TestBindRequestContextUsesOnlyTokenHash(t *testing.T) {
	tx := &fakeTx{}
	hash := make([]byte, 32)
	if err := bindRequestContext(context.Background(), tx, hash); err != nil {
		t.Fatalf("bindRequestContext() error = %v", err)
	}
	if tx.query != bindRequestContextSQL {
		t.Fatalf("query = %q", tx.query)
	}
	if len(tx.args) != 1 || len(tx.args[0].([]byte)) != 32 {
		t.Fatalf("unexpected args: %#v", tx.args)
	}
	if strings.Contains(strings.ToLower(tx.query), "tenant_id") {
		t.Fatal("request-supplied tenant ID must not be part of context binding")
	}
}

func TestBindRequestContextFailsClosed(t *testing.T) {
	if err := BindRequestContext(context.Background(), nil, make([]byte, 32)); err == nil {
		t.Fatal("nil transaction accepted")
	}
	if err := bindRequestContext(context.Background(), &fakeTx{}, make([]byte, 31)); err == nil {
		t.Fatal("short token hash accepted")
	}
	want := errors.New("boom")
	err := bindRequestContext(context.Background(), &fakeTx{err: want}, make([]byte, 32))
	if !errors.Is(err, want) {
		t.Fatalf("wrapped error = %v", err)
	}
}

func TestClearRequestContext(t *testing.T) {
	tx := &fakeTx{}
	if err := clearRequestContext(context.Background(), tx); err != nil {
		t.Fatalf("clearRequestContext() error = %v", err)
	}
	if tx.query != clearRequestContextSQL || len(tx.args) != 0 {
		t.Fatalf("unexpected clear call: %q %#v", tx.query, tx.args)
	}
}
