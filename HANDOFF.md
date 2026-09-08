# PVNaive — Canonical Handoff

Last updated: 2026-09-08 14:41 Asia/Tehran

## Current truth
- Verified GitHub `main` before this handoff update: `7324392bc4bafd7737c96f304634970e62051777` (documentation reconciliation from authoritative main `3da457660...`). No green post-merge CI is claimed for the docs-only update.
- #64 Task13 OPEN/DRAFT/mergeable=false, API head `3fc14825e1b164bad558decaef47f56b792e81af`; CI, WS1 Exact Accounting and WS1 Pinned Forwardproxy are SUCCESS on that exact head, but fresh real HTTP/1.1 + HTTP/2 rehearsal remains mandatory.
- #81 Task16 OPEN/DRAFT/mergeable=true, API head `3c4310335ab4907d28bac995bba1be3545e14f6e`; candidate `b96c65903e5fc314284ea777ceea236913a03842` has Task16 Schema21 TDD, Exact Accounting and Pinned Forwardproxy SUCCESS, but repository-wide CI FAILURE. Exact-head identity is unresolved.
- #4 Karing OPEN/DRAFT/mergeable=true, head `2501e39dc39e14063b6a501bc96b77bbfcae7384`; reproducible real-client smoke remains pending.
- #95 and older docs refresh PRs are documentation-only and are not Production evidence.

## Worker and Production
- Persistent-report search found no fresh exact-head completion receipt. Preserve disposable credentials, isolated canaries, redacted logs, backup-before-promotion, and independent rollback state.
- Historical notes identify TrPaqet as the active executable slot and other workers as inactive/upgrade-required; this is not a fresh command-level Production audit.
- No fresh command-level Production audit, encrypted backup, rollback, deploy, or postflight was available. No Production mutation occurred.

## Actions in this run
- Re-verified repository metadata, authoritative main ref, open PRs, exact-head workflow runs, and persistent reports.
- Confirmed Task13 exact-head CI/Accounting/Forwardproxy SUCCESS on `3fc14825...`.
- Confirmed Task16 split state: API head `3c431033...`; specialized green evidence on `b96c659...`; repository-wide CI FAILURE on that candidate.
- Added reconciliation/assignment comment to PR #64.
- Updated canonical status and handoff state; no runtime/schema/Production change was integrated.

## Next assignments
1. Task13: reconstruct onto current main, then execute isolated HTTP/1.1 + HTTP/2 rehearsal with target-only kill, sibling survival, forged-tuple rejection, idempotency, credential survival, no restart/reload, and exactly-once accounting.
2. Task16: create one clean current-main-derived head, fix repository-wide CI, preserve Task15 schema20 fixtures, and rerun all four exact-head gates on one SHA.
3. Karing: real import/parse/connect/cleanup smoke with disposable credentials, exact profile hash, client/platform version, and redacted logs.
4. Independent review: RLS, privilege separation, retention/purge, accounting/session lineage, and secret redaction.
5. Production lane: read-only audit first; only after all gates are green create encrypted backup + independent rollback snapshot, then staged promotion and postflight.

Human blockers: current-main-derived worker capacity for live rehearsal, a real Karing client, Task16 repository-wide CI repair on one exact head, and connected Production audit/deploy access.
