package customer

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

var registerCustomerDriver sync.Once
var customerDriverState = &customerScriptDriver{conns: map[string]*customerScriptConn{}}

type customerScriptDriver struct {
	mu    sync.Mutex
	conns map[string]*customerScriptConn
}

func (d *customerScriptDriver) Open(name string) (driver.Conn, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	conn := d.conns[name]
	if conn == nil {
		return nil, errors.New("unknown scripted connection")
	}
	return conn, nil
}

type customerScriptQuery struct {
	contains string
	columns  []string
	values   []driver.Value
}

type customerScriptConn struct {
	mu      sync.Mutex
	queries []customerScriptQuery
	execs   []string
	args    [][]driver.NamedValue
}

func (c *customerScriptConn) Prepare(string) (driver.Stmt, error) { return nil, errors.New("prepare not supported") }
func (c *customerScriptConn) Close() error                        { return nil }
func (c *customerScriptConn) Begin() (driver.Tx, error)           { return customerScriptTx{}, nil }
func (c *customerScriptConn) BeginTx(context.Context, driver.TxOptions) (driver.Tx, error) {
	return customerScriptTx{}, nil
}
func (c *customerScriptConn) QueryContext(_ context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.queries) == 0 {
		return nil, errors.New("unexpected query")
	}
	next := c.queries[0]
	c.queries = c.queries[1:]
	if !strings.Contains(query, next.contains) {
		return nil, errors.New("query did not contain expected contract")
	}
	c.args = append(c.args, append([]driver.NamedValue(nil), args...))
	return &customerRows{columns: next.columns, values: next.values}, nil
}
func (c *customerScriptConn) ExecContext(_ context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.execs) == 0 {
		return nil, errors.New("unexpected exec")
	}
	next := c.execs[0]
	c.execs = c.execs[1:]
	if !strings.Contains(query, next) {
		return nil, errors.New("exec did not contain expected contract")
	}
	c.args = append(c.args, append([]driver.NamedValue(nil), args...))
	return driver.RowsAffected(1), nil
}

type customerScriptTx struct{}
func (customerScriptTx) Commit() error   { return nil }
func (customerScriptTx) Rollback() error { return nil }

type customerRows struct {
	columns []string
	values  []driver.Value
	done    bool
}
func (r *customerRows) Columns() []string { return r.columns }
func (r *customerRows) Close() error      { return nil }
func (r *customerRows) Next(dest []driver.Value) error {
	if r.done {
		return io.EOF
	}
	r.done = true
	copy(dest, r.values)
	return nil
}

func newCustomerStoreTx(t *testing.T, conn *customerScriptConn) *sql.Tx {
	t.Helper()
	registerCustomerDriver.Do(func() { sql.Register("pvnaive-customer-script", customerDriverState) })
	name := strings.ReplaceAll(t.Name(), "/", "_")
	customerDriverState.mu.Lock()
	customerDriverState.conns[name] = conn
	customerDriverState.mu.Unlock()
	db, err := sql.Open("pvnaive-customer-script", name)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = db.Close()
		customerDriverState.mu.Lock()
		delete(customerDriverState.conns, name)
		customerDriverState.mu.Unlock()
	})
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	return tx
}

func TestPostgresStoreCreatesDirectUserAndServiceTerm(t *testing.T) {
	now := time.Date(2026, 8, 29, 12, 0, 0, 0, time.UTC)
	expires := now.Add(30 * 24 * time.Hour)
	quota := int64(50 * BytesPerCustomerGB)
	conn := &customerScriptConn{queries: []customerScriptQuery{
		{contains: "FROM pvnaive.tenants", columns: []string{"id"}, values: []driver.Value{"tenant-direct"}},
		{contains: "INSERT INTO pvnaive.users", columns: []string{"id", "tenant_id", "username", "display_name", "status", "revision", "created_at", "updated_at"}, values: []driver.Value{"user-1", "tenant-direct", "customer1", "Customer One", "active", int64(1), now, now}},
		{contains: "INSERT INTO pvnaive.service_terms", columns: []string{"id", "tenant_id", "user_id", "quota_bytes", "duration_seconds", "start_policy", "purchased_at", "starts_at", "first_connected_at", "expires_at", "state", "revision"}, values: []driver.Value{"term-1", "tenant-direct", "user-1", quota, int64(2592000), "on_creation", now, now, nil, expires, "active", int64(1)}},
	}, execs: []string{"INSERT INTO pvnaive.user_runtime_credentials"}}
	tx := newCustomerStoreTx(t, conn)
	store := NewPostgresStore()

	tenantID, err := store.DirectTenantID(context.Background(), tx)
	if err != nil || tenantID != "tenant-direct" {
		t.Fatalf("DirectTenantID() = %q, %v", tenantID, err)
	}
	user, err := store.CreateUserTx(context.Background(), tx, CreateUserRecord{TenantID: tenantID, Username: "customer1", DisplayName: "Customer One", ActorID: "owner-1"})
	if err != nil || user.ID != "user-1" || user.Status != UserActive {
		t.Fatalf("CreateUserTx() = %#v, %v", user, err)
	}
	term, err := store.CreateServiceTermTx(context.Background(), tx, CreateServiceTermRecord{
		TenantID: tenantID, UserID: user.ID, QuotaBytes: &quota, DurationSeconds: 2592000,
		StartPolicy: StartOnCreation, PurchasedAt: now, StartsAt: &now, ExpiresAt: &expires, State: TermActive,
	})
	if err != nil || term.ID != "term-1" || term.QuotaBytes == nil || *term.QuotaBytes != quota {
		t.Fatalf("CreateServiceTermTx() = %#v, %v", term, err)
	}
	if err := store.BindRuntimeCredentialTx(context.Background(), tx, tenantID, user.ID, term.ID, "runtime-1"); err != nil {
		t.Fatalf("BindRuntimeCredentialTx() error = %v", err)
	}
	if len(conn.queries) != 0 || len(conn.execs) != 0 {
		t.Fatalf("unconsumed store script: queries=%d execs=%d", len(conn.queries), len(conn.execs))
	}
}
