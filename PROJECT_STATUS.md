# PVNaive — Canonical Project Status

Last updated: 2026-09-09 03:40 Asia/Tehran

## Verified GitHub state
- Authoritative GitHub `main` currently returns `8ac4f3e8935c24cb95085a6b04fffb6ce1b2cea4` (`docs: refresh continue-here with current main and Task16 rerun`).
- This head is documentation-only. GitHub Actions CI run `34284651112` (`push` on `main`) completed `success` on 2026-09-08 22:16Z; this proves only the docs commit's CI, not live rehearsal or Production health.
- Existing canonical docs contained stale main references (`19f322...`); corrected in this cycle to the verified branch ref above.
- PR #64 Task13: OPEN / DRAFT, API head `3fc14825e1b164bad558decaef47f56b792e81af`; focused CI/Exact Accounting/Pinned Forwardproxy evidence is green, but fresh real HTTP/1.1 + HTTP/2 rehearsal remains required.
- PR #81 Task16: OPEN / DRAFT, API head `3c4310335ab4907d28bac995bba1be3545e14f6e`; documented candidate `b96c65903e5fc314284ea777ceea236913a03842` is not the API head. Historical run `33626300697` has database failure (`job 102257217166`), while web/go succeeded; skipped rehearsal/bundle. Any rerun must be verified on one exact head before crediting. Preserve Task15 schema20-specific fixtures.
- PR #4 Karing: OPEN / DRAFT, head `2501e39dc39e14063b6a501bc96b77bbfcae7384`; PR text reports CI/rehearsal/bundle success, but reproducible real-client Karing import/parse/connect/cleanup smoke is still required.
- PR #95 and older docs-only PRs are non-runtime evidence.

## Worker / coordinator truth
- Persistent-report search returned no fresh completion receipt tied to current PR heads. Worker-only, stale, dirty, mixed-head, or historical evidence is uncredited.
- Historical notes identify TrPaqet as the active executable development slot; other workers are inactive or upgrade-required. This is not a fresh command-level Production audit.

## Production truth
- No fresh command-level Production audit, deployed SHA/schema verification, encrypted backup preflight, independent rollback snapshot, staged deploy, or postflight was available in this run.
- No merge, deploy, migration, restart/reload, DB write, credential mutation, backup mutation, or rollback mutation occurred.
- Production must not be used as a test lane. Promotion requires exact-head gates, fresh encrypted backup, independent rollback state, staged deploy, and postflight verification.

## Actions in this run
- Re-verified authoritative `main`, open PRs, current CI runs/jobs, and persistent reports.
- Confirmed `main` CI run `34284651112` is green and docs-only.
- Confirmed Task16 repository-wide database failure is still the blocking evidence; no successful rerun result was available in this inspection.
- Corrected canonical documentation drift to the verified `main` SHA and recorded the exact blockers.
- No runtime/schema/Production change was integrated.

## Next executable gates
1. Task16: reconcile API head vs documented candidate; run/verify normal CI, PostgreSQL18 schema21 gate, Exact Accounting, and Pinned Forwardproxy on one exact SHA. Do not credit historical mixed-head evidence.
2. Task13: execute isolated HTTP/1.1 + HTTP/2 rehearsal on exact `3fc14825...` with target-only kill, sibling survival, forged-tuple rejection, repeat-kill idempotency, credential survival, no restart/reload, and exactly-once accounting.
3. Karing: real import/parse/connect/cleanup smoke with disposable credentials, exact profile hash, client/platform/version, and redacted logs.
4. Independent review: RLS, privilege separation, retention/purge safety, accounting/session lineage, and secret redaction.
5. Production promotion only after all exact-head gates pass and fresh backup/rollback evidence exists.

Never claim completion from stale reports, older heads, partial evidence, mixed-head proofs, or dirty worktrees.
