# PVNaive — Canonical Handoff

Last updated: 2026-09-08 12:41 Asia/Tehran

## Current truth
- Verified GitHub `main` before this documentation reconciliation: `d14a2a78633a6edf398fa9545fb46b02ce17281e`; this run added status reconciliation commit `e3b3f9cc95e8cf77c46ce037ef4dc8687a7e6874`.
- The documentation commit has no green post-merge CI evidence claimed.
- #64 Task13 OPEN/DRAFT/mergeable=false, API head `3fc14825e1b164bad558decaef47f56b792e81af`; historical gates are not enough; fresh real HTTP/1.1 + HTTP/2 rehearsal remains mandatory.
- #81 Task16 OPEN/DRAFT/mergeable=false, API head `3c4310335ab4907d28bac995bba1be3545e14f6e`; documented candidate `b96c659...` is not the API head, so exact-head four-gate proof is not credited.
- #4 Karing OPEN/DRAFT/mergeable=true, head `2501e39dc39e14063b6a501bc96b77bbfcae7384`; reproducible real-client smoke remains pending.
- #95 docs refresh OPEN/DRAFT and documentation-only.

## Worker and Production
- Persistent-report search found no fresh exact-head completion receipt. Preserve disposable credentials, isolated canaries, redacted logs, backup-before-promotion, and independent rollback state.
- Historical notes identify TrPaqet as the active executable slot and other workers as inactive/upgrade-required; this is not a fresh command-level Production audit.
- No fresh command-level Production audit, encrypted backup, rollback, deploy, or postflight was available. No Production mutation occurred.

## Actions in this run
- Re-verified repository metadata, authoritative main ref, open PRs, exact-head CI/run availability, and persistent reports.
- Reconciled canonical status/handoff to the verified GitHub state.
- Reviewed Task13, Task16, Karing, and docs-only claims; none was accepted as green because fresh exact-head/live evidence is missing.
- No runtime/schema/Production change was integrated.

## Next assignments
1. Task13: reconstruct onto current main, then execute isolated HTTP/1.1 + HTTP/2 rehearsal with target-only kill, sibling survival, forged-tuple rejection, idempotency, credential survival, no restart/reload, and exactly-once accounting.
2. Task16: create one clean current-main-derived head, preserve Task15 schema20 fixtures, and run normal CI + PostgreSQL18 + WS1 Exact Accounting + WS1 Pinned Forwardproxy on one exact SHA.
3. Karing: real import/parse/connect/cleanup smoke with disposable credentials, exact profile hash, client/platform version, and redacted logs.
4. Independent review: RLS, privilege separation, retention/purge, accounting/session lineage, and secret redaction.
5. Production lane: read-only audit first; only after all gates are green create encrypted backup + independent rollback snapshot, then staged promotion and postflight.

Human blockers: current-main-derived worker capacity for live rehearsal, a real Karing client, PostgreSQL18 gate execution for Task16, and connected Production audit/deploy access.
