package customer

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

func (s *PostgresStore) SetCurrentServiceQuotaTx(ctx context.Context, tx *sql.Tx, userID string, quotaBytes *int64) (ServiceTerm, error) {
	if tx == nil {
		return ServiceTerm{}, errors.New("customer: transaction is required")
	}
	term, err := scanAdjustedServiceTerm(tx.QueryRowContext(ctx, `
UPDATE pvnaive.service_terms st
SET quota_bytes=$2,
    state=CASE WHEN st.state='quota_depleted' THEN 'active'::pvnaive.service_term_state ELSE st.state END,
    revision=st.revision+1,
    updated_at=clock_timestamp()
WHERE st.id=(
    SELECT current.id FROM pvnaive.service_terms current
    WHERE current.user_id=$1::uuid
    ORDER BY current.purchased_at DESC,current.created_at DESC
    LIMIT 1
)
  AND st.state NOT IN ('ended','revoked')
RETURNING st.id::text,st.tenant_id::text,st.user_id::text,st.quota_bytes,
          st.duration_seconds,st.start_policy,st.purchased_at,st.starts_at,
          st.first_connected_at,st.expires_at,st.state,st.revision`, userID, quotaBytes))
	if errors.Is(err, sql.ErrNoRows) {
		return ServiceTerm{}, ErrCustomerServiceNotFound
	}
	if err != nil {
		return ServiceTerm{}, fmt.Errorf("customer: set service quota: %w", err)
	}
	return term, nil
}
