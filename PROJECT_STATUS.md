# PVNaive — Canonical Project Status

Last updated: 2026-09-07 15:46 Asia/Tehran

## Verified state
- Current `main` tip from GitHub: `f23d9bf00c4b94f397a87111d9d56ca0708b41cf` (`docs: refresh canonical handoff for automation run 11`). Exact-head combined status is empty; no post-merge CI green is claimed for this docs-only head.
- PR #64 Task13: OPEN / DRAFT / mergeable=false, head `3fc14825e1b164bad558decaef47f56b792e81af`; focused exact-head tests are supplemental only. Fresh real HTTP/1.1 + HTTP/2 rehearsal remains required. Known tooling/worker-capacity blocker remains unresolved.
- PR #81 Task16: OPEN / DRAFT / mergeable=false, head `3c4310335ab4907d28bac995bba1be3545e14f6e`; dedicated Task16/Exact Accounting/Pinned Forwardproxy evidence exists on prior exact heads, but a fresh single-SHA four-gate proof is not verified. Preserve schema20-specific Task15 fixtures. Persistent reports record prior generic schema21 fixture/RLS mismatches and stale worker candidates; none are credited as current completion.
- PR #4 Karing: OPEN / DRAFT / mergeable=true, head `2501e39dc39e14063b6a501bc96b77bbfcae7384`; historical CI/curl evidence is supplemental only. Reproducible real-client import/parse/connect/cleanup smoke remains missing.

## Production truth
Persistent evidence includes historical read-only `/api/v1/health/ready` and `/api/v1/health/live` passes on `pv-primary`, but no fresh command-level Production audit was executed in this run. No fresh encrypted backup preflight, independent rollback snapshot, deploy or postflight is claimed. No merge, deploy, migration, restart/reload, DB write, credential mutation, backup mutation or rollback mutation occurred.

## Worker truth
Persistent coordinator/worker reports were searched again. No fresh completion receipt tied to the current PR heads was found. A historical `pv-worker-main` schema/auth-refresh candidate was explicitly stale/dirty and environment-blocked (`psql` unavailable); it was not credited or integrated. Use isolated canaries, disposable credentials, redacted logs, independent verification, and never Production as a test lane.

## Actions in this run
- Re-verified current GitHub `main`, open PR #64/#81/#4, exact-head status availability, and persistent coordinator/worker reports.
- Added fresh execution dispatch comments to all three active lanes.
- Corrected canonical project status to the verified current `main` SHA and updated the release blockers/evidence rules.
- No runtime/schema/Production change was integrated.

## Next gates
1. Task16: rebuild/update from exact current `main`; narrow-fix only generic schema21/latest-schema expectations; preserve Task15 schema20-specific fixtures; prove normal CI + PostgreSQL18 + Exact Accounting + Pinned Forwardproxy on one SHA.
2. Task13: fresh isolated HTTP/1.1 + HTTP/2 rehearsal proving target-only kill, sibling survival, forged-tuple rejection, repeat-kill idempotency, credential survival, no restart/reload and exactly-once accounting.
3. Karing: reproducible real-client import/parse/connect/cleanup smoke with disposable credentials, exact profile hash and redacted logs.
4. Independent review: RLS, privilege separation, retention/purge safety, accounting/session lineage and secret redaction.
5. Production promotion only after all exact-head gates pass, fresh encrypted backup and independent rollback state exist, then staged deploy and postflight verification.

Never claim completion from stale reports, older heads, partial evidence or dirty worktrees.
