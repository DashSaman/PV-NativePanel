package subscription

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"strings"
	"sync"
	"testing"
	"time"
)

var registerSubscriptionDriver sync.Once
var subscriptionDriverState = &subscriptionScriptDriver{conns: map[string]*subscriptionScriptConn{}}

type subscriptionScriptDriver struct {
	mu    sync.Mutex
	conns map[string]*subscriptionScriptConn
}

func (d *subscriptionScriptDriver) Open(name string) (driver.Conn, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	conn := d.conns[name]
	if conn == nil {
		return nil, errors.New("unknown scripted connection")
	}
	return conn, nil
}

type subscriptionScriptConn struct {
	queryContains string
	columns       []string
	values        []driver.Value
	args          []driver.NamedValue
}

func (c *subscriptionScriptConn) Prepare(string) (driver.Stmt, error) { return nil, errors.New("prepare not supported") }
func (c *subscriptionScriptConn) Close() error                        { return nil }
func (c *subscriptionScriptConn) Begin() (driver.Tx, error)           { return nil, errors.New("transaction not supported") }
func (c *subscriptionScriptConn) QueryContext(_ context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	if !strings.Contains(query, c.queryContains) {
		return nil, errors.New("query did not contain resolver contract")
	}
	c.args = append([]driver.NamedValue(nil), args...)
	return &subscriptionRows{columns: c.columns, values: c.values}, nil
}

type subscriptionRows struct {
	columns []string
	values  []driver.Value
	done    bool
}

func (r *subscriptionRows) Columns() []string { return r.columns }
func (r *subscriptionRows) Close() error      { return nil }
func (r *subscriptionRows) Next(dest []driver.Value) error {
	if r.done {
		return io.EOF
	}
	r.done = true
	copy(dest, r.values)
	return nil
}

func TestPostgresStoreResolveTokenUsesSecurityDefinerResolver(t *testing.T) {
	expires := time.Date(2026, 9, 30, 12, 0, 0, 0, time.UTC)
	conn := &subscriptionScriptConn{
		queryContains: "pvnaive.resolve_direct_subscription_token",
		columns: []string{
			"runtime_credential_id", "runtime_username", "user_state", "service_state",
			"secret_ciphertext", "secret_nonce", "encryption_key_id", "expires_at",
		},
		values: []driver.Value{
			"runtime-1", "customer1", "active", "pending",
			[]byte("ciphertext-0123456789"), []byte("123456789012"), "runtime-v1", expires,
		},
	}
	registerSubscriptionDriver.Do(func() { sql.Register("pvnaive-subscription-script", subscriptionDriverState) })
	name := strings.ReplaceAll(t.Name(), "/", "_")
	subscriptionDriverState.mu.Lock()
	subscriptionDriverState.conns[name] = conn
	subscriptionDriverState.mu.Unlock()
	db, err := sql.Open("pvnaive-subscription-script", name)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = db.Close()
		subscriptionDriverState.mu.Lock()
		delete(subscriptionDriverState.conns, name)
		subscriptionDriverState.mu.Unlock()
	})

	store := NewPostgresStore(db)
	var hash [32]byte
	for i := range hash {
		hash[i] = byte(i + 1)
	}
	record, err := store.ResolveToken(context.Background(), hash)
	if err != nil {
		t.Fatalf("ResolveToken() error = %v", err)
	}
	if record.RuntimeCredentialID != "runtime-1" || record.Username != "customer1" || record.UserState != "active" || record.TermState != "pending" {
		t.Fatalf("resolved record = %#v", record)
	}
	if record.ExpiresAt == nil || !record.ExpiresAt.Equal(expires) {
		t.Fatalf("resolved expiry = %v", record.ExpiresAt)
	}
	if len(conn.args) != 1 {
		t.Fatalf("resolver args = %d, want 1", len(conn.args))
	}
	gotHash, ok := conn.args[0].Value.([]byte)
	if !ok || len(gotHash) != 32 {
		t.Fatalf("resolver hash argument = %#v", conn.args[0].Value)
	}
}
