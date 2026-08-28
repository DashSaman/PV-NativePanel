package customer

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

func (s *PostgresStore) ActivateFirstUseTx(ctx context.Context, tx *sql.Tx, runtimeCredentialID string, observedAt time.Time) (bool, error) {
	if tx == nil {
		return false, errors.New("customer: transaction is required")
	}
	var activated bool
	var syncedTokens int64
	err := tx.QueryRowContext(ctx, `
WITH activated AS (
    UPDATE pvnaive.service_terms AS st
       SET starts_at = $2,
           first_connected_at = $2,
           expires_at = $2 + (st.duration_seconds * interval '1 second'),
           state = 'active',
           revision = st.revision + 1,
           updated_at = $2
      FROM pvnaive.user_runtime_credentials AS urc
      JOIN pvnaive.naive_runtime_credentials AS rc
        ON rc.id = urc.runtime_credential_id
       AND rc.status = 'active'
      JOIN pvnaive.users AS u
        ON u.id = urc.user_id
       AND u.tenant_id = urc.tenant_id
       AND u.status = 'active'
     WHERE urc.runtime_credential_id = $1::uuid
       AND urc.unbound_at IS NULL
       AND urc.role = 'primary'
       AND st.id = urc.service_term_id
       AND st.tenant_id = urc.tenant_id
       AND st.user_id = urc.user_id
       AND st.state = 'pending'
       AND st.start_policy = 'on_first_successful_connection'
       AND st.starts_at IS NULL
       AND st.first_connected_at IS NULL
       AND st.expires_at IS NULL
       AND $2 >= st.purchased_at
    RETURNING st.id, st.expires_at
), synced AS (
    UPDATE pvnaive.direct_subscription_tokens AS dst
       SET service_state = 'active',
           expires_at = activated.expires_at
      FROM activated
     WHERE dst.service_term_id = activated.id
       AND dst.status = 'active'
       AND dst.revoked_at IS NULL
    RETURNING dst.id
)
SELECT EXISTS(SELECT 1 FROM activated), (SELECT COUNT(*) FROM synced)`, runtimeCredentialID, observedAt.UTC()).Scan(&activated, &syncedTokens)
	if err != nil {
		return false, fmt.Errorf("customer: activate first use: %w", err)
	}
	_ = syncedTokens // A service may have no currently issued subscription token.
	return activated, nil
}
