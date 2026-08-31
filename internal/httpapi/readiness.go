package httpapi

import (
	"context"
	"database/sql"
	"errors"
)

const readinessSchemaVersionQuery = `SELECT COALESCE(MAX(version), 0) FROM pvnaive.schema_migrations`

type ReadinessProbeFunc func(context.Context) error

// NewDBReadinessProbe checks only the dependencies required to accept API traffic:
// PostgreSQL must answer within the caller's bounded context and the applied schema
// must exactly match the release's expected schema. Detailed database errors are
// intentionally not returned to HTTP clients by the readiness handler.
func NewDBReadinessProbe(db *sql.DB, expectedSchemaVersion int) ReadinessProbeFunc {
	return func(ctx context.Context) error {
		if db == nil || expectedSchemaVersion <= 0 {
			return errors.New("readiness dependency is not configured")
		}
		if err := db.PingContext(ctx); err != nil {
			return errors.New("database readiness probe failed")
		}
		var actual int
		if err := db.QueryRowContext(ctx, readinessSchemaVersionQuery).Scan(&actual); err != nil {
			return errors.New("schema readiness probe failed")
		}
		if actual != expectedSchemaVersion {
			return errors.New("schema readiness mismatch")
		}
		return nil
	}
}
