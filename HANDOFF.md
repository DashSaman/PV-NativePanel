# PVNaive — Canonical Handoff

Last updated: 2026-09-08 22:39 Asia/Tehran

## Current truth
- Verified GitHub `main` is `478638bdd491173fb8bfd33f075eb157bc49ac3b` at inspection; this run reconciled `PROJECT_STATUS.md` to the observed head and then advanced `main` with documentation-only commit `5463f8837204826ac7917a71d61d01d324e80d68`. No green post-merge CI is claimed for the docs-only updates; combined status is empty and no PR-triggered workflow run was returned for the observed docs head.
- #64 Task13 OPEN/DRAFT/mergeable=false, API head `3fc14825e1b164bad558decaef47f56b792e81af`; exact-head CI/Exact Accounting/Pinned Forwardproxy are green, but fresh real HTTP/1.1 + HTTP/2 rehearsal remains mandatory.
- #81 Task16 OPEN/DRAFT/mergeable=false, API head `3c4310335ab4907d28bac995bba1be3545e14f6e`; documented candidate `b96c65903e5fc314284ea777ceea236913a03842` has specialized gates green, but repository-wide CI previously failed with `ERROR: RLS coverage check failed: 43/42`. Exact-head identity and four-gate closure remain unresolved.
- #4 Karing OPEN/DRAFT/mergeable=true, head `2501e39dc39e14063b6a501bc96b77bbfcae7384`; reproducible real-client smoke remains pending.
- #95 and older docs refresh PRs are documentation-only and are not Production evidence.

## Worker and Production
- Persistent-report search found no fresh exact-head completion receipt. Preserve disposable credentials, isolated canaries, redacted logs, backup-before-promotion, and independent rollback state.
- Historical notes identify TrPaqet as the active executable slot and other workers as inactive/upgrade-required; this is not a fresh command-level Production audit.
- No fresh command-level Production audit, encrypted backup, rollback, deploy, or postflight was available. No Production mutation occurred.

## Actions in this run
- Re-verified repository metadata, authoritative main ref, open PRs, current combined status, and persistent reports.
- Confirmed observed `main=478638bdd...` had no current status checks and no PR-triggered workflow runs.
- Reconciled `PROJECT_STATUS.md` to the actual observed main SHA; then updated this handoff to preserve the exact blockers and safety state.
- No runtime/schema/Production change was integrated.

## Next assignments
1. Task16: create one clean current-main-derived head, fix the repository-wide RLS coverage mismatch (`43/42`), preserve Task15 schema20 fixtures, and rerun all four exact-head gates on one SHA.
2. Task13: reconstruct onto current main, then execute isolated HTTP/1.1 + HTTP/2 rehearsal with target-only kill, sibling survival, forged-tuple rejection, idempotency, credential survival, no restart/reload, and exactly-once accounting.
3. Karing: real import/parse/connect/cleanup smoke with disposable credentials, exact profile hash, client/platform version, and redacted logs.
4. Independent review: RLS, privilege separation, retention/purge, accounting/session lineage, and secret redaction.
5. Production lane: read-only audit first; only after all gates are green create encrypted backup + independent rollback snapshot, then staged promotion and postflight.

Human blockers: current-main-derived worker capacity for live rehearsal, a real Karing client, Task16 repository-wide CI repair on one exact head, and connected Production audit/deploy access.
