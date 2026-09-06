# PVNaive — Canonical Project Status

Last updated: 2026-09-07 00:41 Asia/Tehran

## Verified state
- `main` exact head: `87d5931701e667f8c0c63b7499433d4be812ce05` (`docs(handoff): reconcile current CI and release blockers`). CI run `34057356515` completed SUCCESS for this exact head. This is documentation-only; no runtime/schema code changed in this run.
- PR #64 Task13: OPEN, DRAFT, `mergeable=false`, head `3fc14825e1b164bad558decaef47f56b792e81af`; prior exact-head CI/Exact Accounting/Pinned Forwardproxy gates are SUCCESS, but fresh real HTTP/1.1 + HTTP/2 rehearsal is still missing.
- PR #81 Task16: OPEN, DRAFT, head `3c4310335ab4907d28bac995bba1be3545e14f6e`; current branch body still records the last validated implementation head `b96c65903e5fc314284ea777ceea236913a03842`. Task16 PG18, Exact Accounting, and Pinned Forwardproxy were SUCCESS on that validated head, but no current exact-head all-four-green receipt exists; do not promote.
- PR #4 Karing: OPEN, DRAFT, current head `2501e39dc39e14063b6a501bc96b77bbfcae7384`; CI `33209239812` is SUCCESS, but reproducible real Karing client smoke is still missing.

## Production truth
Persistent reports provide bounded historical/read-only observations only. No fresh command-level Production audit, encrypted backup preflight, independent rollback snapshot, deploy or postflight was executable in this run. No merge, deploy, migration, restart/reload, DB write, credential mutation, backup mutation or rollback mutation occurred.

## Worker truth
Persistent coordinator/worker reports were rechecked. No fresh completion receipt tied to the current PR heads was found. Historical worker output is not credited without exact GitHub corroboration and a fresh receipt. One-active-host limitation remains; connected/inactive workers are not treated as executable capacity. Worker-only output was not integrated.

## This run
- Re-verified current `main`, open PRs #4/#64/#81, latest CI evidence, and persistent coordinator/worker reports.
- Confirmed exact-head `main` CI success on run `34057356515`.
- Confirmed Task13 protocol rehearsal, Task16 exact-head alignment, and Karing real-client smoke remain release blockers.
- Refreshed this canonical status file; documentation-only change.

## Next gates
1. Task16: publish one clean current head from `main`, preserve schema20-specific Task15 fixtures, and run normal CI + Task16 PG18 + Exact Accounting + Pinned Forwardproxy on the same SHA.
2. Task13: obtain fresh real HTTP/1.1 + HTTP/2 rehearsal proving target-only kill, sibling survival, forged-tuple rejection, repeat-kill idempotency, credential survival, no restart/reload and exactly-once final accounting.
3. Karing: obtain reproducible real client import/parse/connect/cleanup evidence.
4. Independent review: inspect RLS, privilege separation, retention/purge safety, accounting/session lineage and secret redaction.
5. Only after all required gates pass: fresh encrypted Production backup, independent rollback state, staged deploy and postflight verification.

Never claim completion from stale reports, older heads or partial evidence.
