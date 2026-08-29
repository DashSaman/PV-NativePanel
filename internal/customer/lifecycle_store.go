package customer

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/DashSaman/PV-NaivePanel/internal/runtimecred"
)

func (s *PostgresStore) CustomerRuntimeTargetTx(ctx context.Context, tx *sql.Tx, userID string) (CustomerRuntimeTarget, error) {
	if tx == nil {
		return CustomerRuntimeTarget{}, errors.New("customer: transaction is required")
	}
	var target CustomerRuntimeTarget
	var userState, runtimeStatus string
	err := tx.QueryRowContext(ctx, `
SELECT
    u.id::text,
    u.username,
    u.status,
    rc.id::text,
    rc.username,
    rc.status,
    rc.revision
FROM pvnaive.users u
JOIN LATERAL (
    SELECT st.id
    FROM pvnaive.service_terms st
    WHERE st.user_id = u.id
      AND st.tenant_id = u.tenant_id
    ORDER BY st.purchased_at DESC, st.created_at DESC
    LIMIT 1
) current_term ON TRUE
JOIN pvnaive.user_runtime_credentials urc
  ON urc.user_id = u.id
 AND urc.service_term_id = current_term.id
 AND urc.unbound_at IS NULL
JOIN pvnaive.naive_runtime_credentials rc
  ON rc.id = urc.runtime_credential_id
WHERE u.id = $1::uuid
LIMIT 1`, userID).Scan(
		&target.UserID,
		&target.Username,
		&userState,
		&target.RuntimeCredentialID,
		&target.RuntimeUsername,
		&runtimeStatus,
		&target.RuntimeRevision,
	)
	if err != nil {
		return CustomerRuntimeTarget{}, fmt.Errorf("customer: resolve runtime target: %w", err)
	}
	target.UserState = UserAdminState(userState)
	target.RuntimeStatus = runtimecred.CredentialStatus(runtimeStatus)
	return target, nil
}

func (s *PostgresStore) SetUserAdminStateTx(ctx context.Context, tx *sql.Tx, userID string, state UserAdminState) error {
	if tx == nil {
		return errors.New("customer: transaction is required")
	}
	if state != UserActive && state != UserSuspended && state != UserRevoked {
		return errors.New("customer: unsupported lifecycle state")
	}
	result, err := tx.ExecContext(ctx, `
UPDATE pvnaive.users
SET status = $2,
    revision = revision + 1,
    updated_at = clock_timestamp()
WHERE id = $1::uuid`, userID, string(state))
	if err != nil {
		return fmt.Errorf("customer: update lifecycle state: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("customer: inspect lifecycle update: %w", err)
	}
	if rows != 1 {
		return ErrCustomerServiceNotFound
	}
	return nil
}
