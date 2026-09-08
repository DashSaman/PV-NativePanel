# PVNaive — Canonical Handoff

Last updated: 2026-09-08 03:43 Asia/Tehran

## Current truth
- Verified GitHub `main` ref at inspection: `37b9b4714ee3ca8e52f7511c9b75d80ca2a8d893`. This run adds documentation-only reconciliation commit `e1b68920138ae1080d6e77aebbb9a474c25205c0`.
- Exact-head combined status for the docs-only head is empty; no post-merge CI green is claimed.
- #64 Task13 OPEN/DRAFT/mergeable=false, API head `3fc14825e1b164bad558decaef47f56b792e81af`; stale base versus current main. Fresh current-main-derived HTTP/1.1 + HTTP/2 rehearsal remains mandatory.
- #81 Task16 OPEN/DRAFT/mergeable=false, API head `3c4310335ab4907d28bac995bba1be3545e14f6e`; body cites older implementation heads; current exact-head four-gate proof is absent.
- #4 Karing OPEN/DRAFT/mergeable=true, head `2501e39dc39e14063b6a501bc96b77bbfcae7384`; real-client smoke remains pending.

## Worker and Production
- Persistent-report search found no fresh exact-head completion receipt. Preserve disposable credentials, isolated canaries, redacted logs, backup-before-promotion, and independent rollback state.
- Historical notes identify TrPaqet as the active executable slot and inactive workers as unavailable/upgrade-required; this is not a fresh command-level Production audit.
- No fresh command-level Production audit, encrypted backup, rollback, deploy, or postflight was available. No Production mutation occurred.

## Actions in this run
- Re-verified repository metadata, authoritative main ref, open PRs, exact-head status, and persistent reports.
- Reconciled canonical docs to the actual GitHub state.
- Posted fresh execution assignments to #64, #81, and #4.
- No runtime/schema/Production change was integrated.

## Next assignments
1. Task13: reconstruct onto current main, then execute isolated HTTP/1.1 + HTTP/2 rehearsal with target-only kill, sibling survival, forged-tuple rejection, idempotency, credential survival, no restart/reload, and exactly-once accounting.
2. Task16: create one clean current-main-derived head, preserve Task15 schema20 fixtures, and run normal CI + PostgreSQL18 + Exact Accounting + Pinned Forwardproxy on one exact SHA.
3. Karing: real import/parse/connect/cleanup smoke with disposable credentials, exact profile hash, client/platform version, and redacted logs.
4. Independent review: RLS, privilege separation, retention/purge, accounting/session lineage, and secret redaction.
5. Production lane: read-only audit first; only after all gates are green create encrypted backup + independent rollback snapshot, then staged promotion and postflight.

Human blockers: current-main-derived worker capacity for live rehearsal, a real Karing client, PostgreSQL18 gate execution for Task16, and connected Production audit/deploy access.