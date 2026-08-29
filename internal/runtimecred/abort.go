package runtimecred

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// AbortAndRollback is the inverse of CommitAndFinalize for a staged Runtime
// mutation whose surrounding business transaction can no longer complete.
// It restores the exact Runtime backup before rolling back the SQL transaction
// and permanently clears any one-time generated secret from this mutation.
func (m *Mutation) AbortAndRollback(ctx context.Context, tx *sql.Tx) error {
	if m == nil || tx == nil {
		return errors.New("runtimecred: mutation and transaction are required")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.committed {
		return errors.New("runtimecred: mutation already finalized")
	}

	rollbackErr := m.agent.Rollback(ctx, m.backupID)
	_ = tx.Rollback()
	m.generatedPassword = ""
	// Mark terminal even when Runtime rollback fails. Retrying the same Mutation
	// after an uncertain rollback would risk applying or exposing state twice.
	m.committed = true
	if rollbackErr != nil {
		return fmt.Errorf("%w: runtime rollback failed during mutation abort", ErrReconciliationRequired)
	}
	return nil
}
