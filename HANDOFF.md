# PVNaive — Canonical Handoff

Last updated: 2026-09-06 16:38 Asia/Tehran

Resume from this file plus `CONTINUE_HERE.md`, `PROJECT_STATUS.md`, exact GitHub `main`, open PRs, newest evidence and fresh Production health. Older worker checkpoints are historical evidence.

## Repository / release truth

- Current `main`: `d3319d29d6a868e482a8f37bac97a84f99b279b7` at inspection start; documentation reconciliation commits from this run follow it.
- Post-merge CI for exact main head `d3319d29...` is SUCCESS: run `34032536681`.
- #64 Task13 draft head `3fc14825e1b164bad558decaef47f56b792e81af`; fresh real HTTP/1.1 + HTTP/2 rehearsal is still pending.
- #81 Task16 draft current head `3c4310335ab4907d28bac995bba1be3545e14f6e`; observed workflow evidence targets older head `b96c65903e5fc314284ea777ceea236913a03842`; current exact-head all-green is unproven.
- #4 Karing draft head `2501e39dc39e14063b6a501bc96b77bbfcae7384`; historical CI 402 is green, but real Karing smoke is pending.
- #91–#95 are stale documentation PRs and not promotion authority.

## Production state

No fresh command-level Production audit was executable in this run. No Production health pass is claimed. No restart, reload, migration, DB write, credential rotation, backup mutation, rollback mutation or deployment occurred.

## Gates and blockers

- #64: require fresh exact-head HTTP/1.1 + HTTP/2 rehearsal proving target-only kill, sibling survival, forged-tuple rejection, idempotency, credential survival, no restart/reload and exactly-once accounting.
- #81: require all four gates on one exact current SHA; never reuse older-head green runs.
- #4: require reproducible real Karing client smoke.
- Promotion requires fresh encrypted backup, rollback state, exact artifact provenance and postflight verification.
- Persistent reports are historical unless corroborated by exact GitHub state and fresh receipts; worker capacity is bounded by the one-active-host limit.

## This run — 2026-09-06 16:38 Asia/Tehran

- Verified current main, open PRs, exact-main CI, current PR heads and persistent coordinator/worker reports.
- Confirmed exact-main CI success on run `34032536681` for `d3319d29...`.
- Found no fresh worker completion receipt tied to current PR heads.
- Updated canonical status/continuation/handoff docs directly on main; no runtime/schema work was integrated.
- No merge or deploy occurred.

## Next assignments

- Task16 lane: reconcile branch/current head, run normal CI + Task16 PG18 + Exact Accounting + Pinned Forwardproxy on one exact SHA, preserve Task15 schema20 fixtures.
- Task13 lane: run fresh HTTP/1.1 + HTTP/2 rehearsal on exact head outside Production with full session/accounting receipt.
- Karing lane: execute real client smoke and attach reproducible import/connect evidence.
- Independent review lane: security/RLS/accounting/rollback review for #64/#81.
- Production-only lane: read-only health first; only after all gates pass, create encrypted backup + rollback state and consider deploy.

Keep truthful accounting/session semantics under retry, race, kill and disconnect.
