package customer

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

func (s *PostgresStore) ListActiveSessionsTx(ctx context.Context, tx *sql.Tx, userID string, observedAt time.Time, staleAfter time.Duration) ([]CustomerSession, error) {
	if tx == nil || observedAt.IsZero() || staleAfter <= 0 || staleAfter > maxActiveSessionStaleAfter {
		return nil, errors.New("customer: transaction and bounded active-session window are required")
	}
	seconds := int64(staleAfter / time.Second)
	if seconds < 1 {
		seconds = 1
	}
	rows, err := tx.QueryContext(ctx, `
SELECT runtime_credential_id::text, node_id, boot_id::text, session_id::text,
       service_term_id::text, client_ip, connected_at, last_activity_at,
       duration_seconds, upload_bytes, download_bytes
FROM pvnaive.list_active_customer_sessions($1::uuid, $2, $3)`, userID, observedAt.UTC(), seconds)
	if err != nil {
		return nil, fmt.Errorf("customer: list active sessions: %w", err)
	}
	defer rows.Close()

	out := make([]CustomerSession, 0)
	for rows.Next() {
		var session CustomerSession
		if err := rows.Scan(
			&session.RuntimeCredentialID, &session.NodeID, &session.BootID, &session.SessionID,
			&session.ServiceTermID, &session.ClientIP, &session.ConnectedAt, &session.LastActivityAt,
			&session.DurationSeconds, &session.UploadBytes, &session.DownloadBytes,
		); err != nil {
			return nil, fmt.Errorf("customer: scan active session: %w", err)
		}
		out = append(out, session)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("customer: active session rows: %w", err)
	}
	return out, nil
}
