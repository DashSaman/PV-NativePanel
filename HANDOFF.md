# PVNaive — Canonical Handoff

Last updated: 2026-09-09 03:40 Asia/Tehran

## Current truth
- Verified GitHub `main` is `8ac4f3e8935c24cb95085a6b04fffb6ce1b2cea4`. CI run `34284651112` for this docs-only push completed `success` on 2026-09-08 22:16Z; no live rehearsal or Production health claim follows from that run.
- The prior canonical handoff referenced stale main `19f322...`; this cycle corrected the drift.
- #64 Task13 OPEN/DRAFT, API head `3fc14825e1b164bad558decaef47f56b792e81af`; focused exact-head CI/Accounting/Forwardproxy evidence is green, but fresh real HTTP/1.1 + HTTP/2 rehearsal remains mandatory.
- #81 Task16 OPEN/DRAFT, API head `3c4310335ab4907d28bac995bba1be3545e14f6e`; documented candidate `b96c65903e5fc314284ea777ceea236913a03842` is not the API head. Run `33626300697` database job `102257217166` failed; web/go passed and rehearsal/bundle were skipped. No exact-head four-gate closure is available.
- #4 Karing OPEN/DRAFT, head `2501e39dc39e14063b6a501bc96b77bbfcae7384`; PR text reports successful CI/rehearsal/bundle, but reproducible real-client import/parse/connect/cleanup smoke remains pending.

## Worker and Production
- Persistent-report search found no fresh exact-head completion receipt. Preserve disposable credentials, isolated canaries, redacted logs, backup-before-promotion, and independent rollback state.
- Historical notes identify TrPaqet as the active executable slot and other workers as inactive or upgrade-required; this is not a fresh command-level Production audit.
- No fresh command-level Production audit, encrypted backup, rollback, deploy, or postflight was available. No Production mutation occurred.

## Actions in this cycle
- Re-verified repository metadata, authoritative main ref, open PRs, main CI, Task16 CI job state, and persistent reports.
- Corrected canonical status drift to verified main `8ac4f3...` and recorded the concrete Task16 failure (`102257217166` in `33626300697`).
- No runtime/schema/Production change was integrated.

## Next assignments
1. Task16: reconcile API head vs documented candidate, then run/verify normal CI, PostgreSQL18 schema21, Exact Accounting, and Pinned Forwardproxy on one exact SHA; preserve Task15 schema20 fixtures.
2. Task13: execute isolated HTTP/1.1 + HTTP/2 rehearsal with target-only kill, sibling survival, forged-tuple rejection, idempotency, credential survival, no restart/reload, and exactly-once accounting.
3. Karing: real import/parse/connect/cleanup smoke with disposable credentials, exact profile hash, client/platform/version, and redacted logs.
4. Independent review: RLS, privilege separation, retention/purge, accounting/session lineage, and secret redaction.
5. Production lane: read-only audit first; only after all gates are green create encrypted backup + independent rollback snapshot, then staged promotion and postflight.

Human blockers: connected worker execution for live rehearsal, a real Karing client, Task16 exact-head CI closure, and connected Production audit/deploy access.