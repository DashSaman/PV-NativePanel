# PVNaive — Canonical Project Status

Last updated: 2026-09-07 04:42 Asia/Tehran

## Verified state
- Current `main` tip from GitHub recent commit history: `27e8c291d2fde20f7cd90f939414d6e09f396306` (`docs: refresh canonical handoff checkpoint`). No post-merge CI result for this docs-only head is available in the connected GitHub view; do not claim post-merge green.
- PR #64 Task13: OPEN / DRAFT, head `3fc14825e1b164bad558decaef47f56b792e81af`; current combined status is empty. Historical focused gates are supplemental only. Fresh real HTTP/1.1 + HTTP/2 rehearsal remains required.
- PR #81 Task16: OPEN / DRAFT, head `3c4310335ab4907d28bac995bba1be3545e14f6e`; current combined status is empty. The last verified branch evidence reports generic latest-schema fixture mismatch and a prior pinned Caddy transient; no current single-SHA four-gate proof is available. Preserve schema20-specific Task15 fixtures.
- PR #4 Karing: OPEN / DRAFT, head `2501e39dc39e14063b6a501bc96b77bbfcae7384`; historical CI is green, but reproducible real-client import/parse/connect/cleanup smoke is still missing.
- Stale documentation PRs (#85/#86/#87/#88/#89/#91/#92/#93/#94/#95) remain non-canonical because they target older bases/claims.

## Production truth
No connected Production command/deployment lane or fresh command-level audit was available in this run. Persistent evidence is bounded and historical/read-only. No fresh health pass, encrypted backup preflight, independent rollback snapshot, deploy or postflight is claimed. No merge, deploy, migration, restart/reload, DB write, credential mutation, backup mutation or rollback mutation occurred.

## Worker truth
Persistent coordinator/worker reports were searched again. No fresh completion receipt tied to the current PR heads was found. Historical, dirty or stale worker output is not credited and was not integrated. The one-active-host / limited development capacity constraint remains the main execution blocker.

## Actions in this run
- Re-verified repository metadata, latest main history, open PRs, PR heads/bodies, exact-head status availability and persistent reports.
- Reconciled stale documentation and historical worker claims as non-promotion evidence.
- Updated canonical status documentation only; no runtime/schema change was integrated.
- Reassigned independent next actions on Task16, Task13 and Karing.

## Next gates
1. Task16: clean current-head branch; narrow-fix only generic schema21/latest-schema expectations; preserve Task15 schema20 fixtures; run normal CI + Task16 PostgreSQL18 + Exact Accounting + Pinned Forwardproxy on one exact SHA.
2. Task13: fresh isolated HTTP/1.1 + HTTP/2 rehearsal proving target-only kill, sibling survival, forged-tuple rejection, repeat-kill idempotency, credential survival, no restart/reload and exactly-once accounting.
3. Karing: reproducible real-client import/parse/connect/cleanup smoke with disposable credentials and redacted logs.
4. Independent review: RLS, privilege separation, retention/purge safety, accounting/session lineage and secret redaction.
5. Production promotion only after all exact-head gates pass, fresh encrypted backup and independent rollback state exist, then staged deploy and postflight verification.

Never claim completion from stale reports, older heads, partial evidence or dirty worktrees.