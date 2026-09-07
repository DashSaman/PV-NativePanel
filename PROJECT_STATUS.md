# PVNaive — Canonical Project Status

Last updated: 2026-09-07 17:45 Asia/Tehran

## Verified state
- Current `main` tip from GitHub: `8d70a6ec92dac5b4117cb23701be1ac24d7c67ae` (`docs: update canonical handoff with latest CI blocker`). Exact-head combined status is empty; no post-merge CI green is claimed for this docs-only head.
- PR #64 Task13: OPEN / DRAFT / mergeable=false, head `3fc14825e1b164bad558decaef47f56b792e81af`; fresh focused exact-head evidence is supplemental only. Real HTTP/1.1 + HTTP/2 rehearsal remains required.
- PR #81 Task16: OPEN / DRAFT / mergeable=false, head `3c4310335ab4907d28bac995bba1be3545e14f6e`; dedicated Task16/Exact Accounting/Pinned Forwardproxy evidence is historical or mixed-head and does not authorize merge. Repository-wide CI remains blocked by the generic fixture expectation `schema version=21, want=20` in `tests/db/periodic_usage_reset_executor_test.sh`; preserve Task15 schema20-specific fixtures and correct only intended generic latest-schema expectations.
- PR #4 Karing: OPEN / DRAFT / mergeable=true, head `2501e39dc39e14063b6a501bc96b77bbfcae7384`; historical CI/curl evidence is supplemental only. Reproducible real-client import/parse/connect/cleanup smoke remains missing.

## CI evidence
- Current exact-head status calls for `main`, Task13 head, and Task16 head returned no combined statuses. This is absence of status evidence, not a pass.
- Task16's latest documented failure is repository-wide CI database job `101509296474` with `schema version=21, want=20`, while dedicated gates were green on the prior observed run set. No merge decision is permitted from that mixed state.

## Production truth
- Persistent evidence includes historical read-only health passes on `pv-primary`, but no fresh command-level Production audit was available in this run. No fresh encrypted backup preflight, independent rollback snapshot, deploy or postflight is claimed. No merge, deploy, migration, restart/reload, DB write, credential mutation, backup mutation or rollback mutation occurred.

## Worker truth
- Persistent coordinator/worker reports were searched again. No fresh completion receipt tied to the current PR heads was found. Worker outputs without exact-head receipts remain uncredited. TrPaqet remains the documented active executable slot; worker-local PostgreSQL 14.x is not PostgreSQL18 evidence. Use isolated canaries, disposable credentials, redacted logs, independent verification, and never Production as a test lane.

## Actions in this run
- Re-verified GitHub `main`, open PR #64/#81/#4, exact-head status availability, canonical docs, and current persistent-report search results.
- Confirmed no validated worker completion was available for integration.
- Updated this canonical status with the actual `main` SHA and current evidence boundaries.
- No runtime/schema/Production change was integrated.

## Next gates
1. Task16: patch only generic latest-schema expectation(s) that still require 20; preserve Task15 schema20-specific tests; rerun normal CI + PostgreSQL18 + Exact Accounting + Pinned Forwardproxy on one exact SHA.
2. Task13: fresh isolated HTTP/1.1 + HTTP/2 rehearsal proving target-only kill, sibling survival, forged-tuple rejection, repeat-kill idempotency, credential survival, no restart/reload and exactly-once accounting.
3. Karing: reproducible real-client import/parse/connect/cleanup smoke with disposable credentials, exact profile hash and redacted logs.
4. Independent review: RLS, privilege separation, retention/purge safety, accounting/session lineage and secret redaction.
5. Production promotion only after all exact-head gates pass, fresh encrypted backup and independent rollback state exist, then staged deploy and postflight verification.

Never claim completion from stale reports, older heads, partial evidence or dirty worktrees.