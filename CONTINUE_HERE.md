# CONTINUE HERE — PVNaive

Last updated: 2026-09-03

Re-read exact GitHub `main`, open PRs, exact-head CI and fresh Production health before any mutation.

## Verified state

- `main`: `0b921abe9b2bd1d827023f494fda11a407fe34d3`.
- Task13 PR #64: exact head `3fc14825e1b164bad558decaef47f56b792e81af`; GitHub gates green, real HTTP/1.1 + HTTP/2 rehearsal pending.
- Task16 PR #81: exact head `3c4310335ab4907d28bac995bba1be3545e14f6e`; TDD/Exact Accounting/Pinned Forwardproxy green, normal CI red at `RLS coverage check failed: 43/42`.
- Production read-only HTTP probes returned 200 for ready/live with `db=ok`, `schema=ok`, `ready=true`, `status=ready` and `service=pvnaive-api`, `status=ok`.
- Docker inventory was unavailable due to socket permission denial; do not infer container state beyond the HTTP evidence.
- No Production mutation occurred.

## Next execution

1. Keep #64/#81 draft and do not merge on partial evidence.
2. Fix only generic latest-schema CI expectations; keep Task15 schema20-specific tests unchanged.
3. Run Task13 live rehearsal on development infrastructure.
4. On green gates only: fresh encrypted Production backup + rollback state, exact artifact verification, controlled deploy and postflight.

## Worker assignments

- `TrPaqet`: Task13 live rehearsal.
- `pv-worker-main`: Task16 CI fixture correction and exact-head rerun when executable.
- `pv-primary`: Production-only audit/backup/rollback/deploy lane.

Historical worker reports do not override exact GitHub or fresh Production evidence.
