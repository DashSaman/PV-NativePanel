# S04 runtime gate fixes — 2026-08-27

Production deployment remains blocked pending a fully green CI + rehearsal + bundle run.

## Failures caught before server rollout

1. The first rehearsal harness deleted its own temporary directory before starting the API. This was a test-harness bug and did not change production code.
2. The corrected rehearsal exposed real PostgreSQL boundaries:
   - `auth_append_audit` could not insert into `audit_events` because the table uses forced RLS.
   - authenticated logout attempted a direct `UPDATE` on `auth_sessions` after migration 0002 had revoked raw DML from `pvnaive_app`.

## Fixes

- Added a narrowly-scoped INSERT policy for `pvnaive_owner` so the SECURITY DEFINER authentication audit helper can append immutable auth audit rows without granting direct app-table INSERT.
- Added scoped SECURITY DEFINER helpers for revoking one actor-owned session and all other sessions of the currently authenticated actor.
- Store methods now call these helpers instead of issuing raw session UPDATE statements.
- Rehearsal cleanup no longer disables errexit for the main test flow.
- Rehearsal now asserts a successful login audit row is actually persisted.
- Migration SHA256 manifest was regenerated and verified.
- One-time patch runner completed `go mod verify`, `go vet ./...`, `go test ./...`, shell syntax checks, and removed itself.

## Gate

Do not deploy S04 to `testAmir5-3` until the ordinary CI run from this evidence commit reports Go, Web, PostgreSQL 18, end-to-end rehearsal, and production bundle all successful.
