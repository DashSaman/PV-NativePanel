# PVNaive — Canonical Project Status

Last updated: 2026-09-07 16:48 Asia/Tehran

## Verified state
- Current `main` tip from GitHub: `e500042b28bfaa6313c5fbc55cccf8e3004ac234` (docs refresh). Exact-head combined status is empty; no post-merge CI green is claimed for this docs-only head.
- PR #64 Task13: OPEN / DRAFT / mergeable=false, head `3fc14825e1b164bad558decaef47f56b792e81af`; fresh focused exact-head evidence is supplemental only. Real HTTP/1.1 + HTTP/2 rehearsal remains required.
- PR #81 Task16: OPEN / DRAFT / mergeable=false, head `3c4310335ab4907d28bac995bba1be3545e14f6e`; Task16 Schema21 TDD, WS1 Exact Accounting, and WS1 Pinned Forwardproxy are SUCCESS on the latest observed run set, but repository-wide CI is FAILED in the database job. The failure is the generic latest-schema fixture path: `tests/db/periodic_usage_reset_executor_test.sh` reports `ERROR: schema version=21, want=20`. Preserve schema20-specific Task15 fixtures and correct only intended generic latest-schema expectations.
- PR #4 Karing: OPEN / DRAFT / mergeable=true, head `2501e39dc39e14063b6a501bc96b77bbfcae7384`; historical CI/curl evidence is supplemental only. Reproducible real-client import/parse/connect/cleanup smoke remains missing.

## CI evidence
- On PR #81 head `3c4310335ab4907d28bac995bba1be3545e14f6e`, observed workflows: Task16 Schema21 TDD `33678134359` SUCCESS; WS1 Exact Accounting `33678134326` SUCCESS; WS1 Pinned Forwardproxy `33678134350` SUCCESS; repository-wide CI `33678134360` FAILURE.
- CI failure is isolated to database job `101509296474`; logs show migrations/health/backup/restore and prior DB contracts passing, then `ERROR: schema version=21, want=20` before exit 1. No merge decision is permitted from this mixed state.

## Production truth
Persistent evidence includes historical read-only health passes on `pv-primary`, but no fresh command-level Production audit was executed in this run. No fresh encrypted backup preflight, independent rollback snapshot, deploy or postflight is claimed. No merge, deploy, migration, restart/reload, DB write, credential mutation, backup mutation or rollback mutation occurred.

## Worker truth
Persistent coordinator/worker reports were searched again. No fresh completion receipt tied to the current PR heads was found. Worker outputs without exact-head receipts remain uncredited. TrPaqet is still the documented active executable slot; worker-local PostgreSQL 14.x is not PostgreSQL18 evidence. Use isolated canaries, disposable credentials, redacted logs, independent verification, and never Production as a test lane.

## Actions in this run
- Re-verified GitHub `main`, open PR #64/#81/#4, exact-head status availability, current workflow runs and database failure logs.
- Reconciled the latest Task16 evidence: three dedicated gates green, repository-wide CI red on a generic schema-version expectation.
- Updated this canonical status with the exact failing test and run/job identifiers.
- No runtime/schema/Production change was integrated.

## Next gates
1. Task16: patch only generic latest-schema expectation(s) that still require 20; preserve Task15 schema20-specific tests; rerun normal CI + PostgreSQL18 + Exact Accounting + Pinned Forwardproxy on one exact SHA.
2. Task13: fresh isolated HTTP/1.1 + HTTP/2 rehearsal proving target-only kill, sibling survival, forged-tuple rejection, repeat-kill idempotency, credential survival, no restart/reload and exactly-once accounting.
3. Karing: reproducible real-client import/parse/connect/cleanup smoke with disposable credentials, exact profile hash and redacted logs.
4. Independent review: RLS, privilege separation, retention/purge safety, accounting/session lineage and secret redaction.
5. Production promotion only after all exact-head gates pass, fresh encrypted backup and independent rollback state exist, then staged deploy and postflight verification.

Never claim completion from stale reports, older heads, partial evidence or dirty worktrees.
