package customer

import (
	"context"
	"database/sql/driver"
	"testing"
	"time"
)

func TestListActiveSessionsUsesProvidedTransaction(t *testing.T) {
	observed := time.Date(2026, 8, 31, 7, 45, 0, 0, time.UTC)
	connected := observed.Add(-45 * time.Second)
	last := observed.Add(-3 * time.Second)
	conn := &customerScriptConn{queries: []customerScriptQuery{{
		contains: "FROM pvnaive.list_active_customer_sessions",
		columns: []string{
			"runtime_credential_id", "node_id", "boot_id", "session_id", "service_term_id",
			"client_ip", "connected_at", "last_activity_at", "duration_seconds", "upload_bytes", "download_bytes",
		},
		values: []driver.Value{
			"17170000-0000-0000-0000-000000000081", "node-a",
			"17170000-0000-0000-0000-000000000082", "17170000-0000-0000-0000-000000000083",
			"17170000-0000-0000-0000-000000000084", "203.0.113.9", connected, last,
			int64(45), int64(1234), int64(5678),
		},
	}}}
	tx := newCustomerStoreTx(t, conn)
	service := NewService(NewPostgresStore(), nil, func() time.Time { return observed })

	sessions, err := service.ListActiveSessions(
		context.Background(), tx, "17170000-0000-0000-0000-000000000041", observed, 90*time.Second,
	)
	if err != nil {
		t.Fatalf("ListActiveSessions() error = %v", err)
	}
	if len(sessions) != 1 {
		t.Fatalf("ListActiveSessions() len = %d, want 1", len(sessions))
	}
	got := sessions[0]
	if got.ClientIP != "203.0.113.9" || got.NodeID != "node-a" || got.UploadBytes != 1234 || got.DownloadBytes != 5678 {
		t.Fatalf("ListActiveSessions() = %#v", got)
	}
	if len(conn.queries) != 0 {
		t.Fatalf("unconsumed store query count = %d", len(conn.queries))
	}
}
