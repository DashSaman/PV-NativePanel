# PVNaive — Canonical Project Status

Last updated: 2026-09-06 16:38 Asia/Tehran

This file records verified repository truth and bounded Production truth. Historical worker/stage notes are evidence only; exact GitHub state, exact-head CI and fresh Production observations override them.

## Safety invariants

PVNaive remains standalone-first. Never fabricate usage/online/IP/session history. Never rotate credentials/tokens from read-only flows. Production changes require a fresh encrypted backup + rollback state, intended migrations only, exact artifact provenance and postflight verification.

## Repository truth

- Repository: `DashSaman/PV-NativePanel`.
- Current `main` at start of this run: `d3319d29d6a868e482a8f37bac97a84f99b279b7` (verified directly from GitHub).
- Post-merge CI for that exact main head is SUCCESS: CI run `34032536681` completed successfully on 2026-09-06 12:14–12:18Z.
- Task13: draft PR #64, head `3fc14825e1b164bad558decaef47f56b792e81af`; focused checks are supplemental only; fresh real HTTP/1.1 + HTTP/2 rehearsal remains incomplete.
- Task16: draft PR #81, current GitHub head `3c4310335ab4907d28bac995bba1be3545e14f6e`; the observed workflow evidence is attached to older head `b96c65903e5fc314284ea777ceea236913a03842`, so current exact-head all-green is not proven.
- PR #4 (Karing export): draft, head `2501e39dc39e14063b6a501bc96b77bbfcae7384`, base `s04-auth`; historical CI run 402 is SUCCESS, but no reproducible real Karing client smoke evidence is attached.
- Documentation-only PRs #91–#95 remain open/stale and are not promotion authority; canonical docs are updated directly on `main`.

## Production truth

- No fresh command-level Production audit was executable in this run. Historical read-only evidence is not re-credited as a fresh pass.
- No Task13, Task16 or PR #4 code is authorized as deployed from current evidence.
- No deploy, migration, restart, reload, DB write, credential change, backup mutation or rollback mutation was performed.

## Persistent worker state

- Persistent reports were searched; they are historical unless corroborated by exact GitHub state and fresh receipts.
- No fresh worker completion receipt tied to the current heads was available for reconciliation.
- Current bounded assignments: TrPaqet → Task13 rehearsal; PostgreSQL18-capable worker → Task16 exact-head reconciliation and CI fix verification; independent worker → regression/security review; `pv-primary` → Production-only when executable access is available.
- Persistent evidence records a one-active-host SentinelX limit; connected workers can be inactive and therefore cannot be treated as executable.

## This run — 2026-09-06 16:38 Asia/Tehran

- Re-verified the repository default branch, current `main`, open PRs #4/#64/#81 and documentation PRs #91–#95.
- Verified exact-main post-merge CI SUCCESS on run `34032536681` for `d3319d29...`.
- Checked exact-head workflow evidence: #81 current head still differs from older green/failing evidence; #64 still lacks fresh protocol rehearsal; #4 still lacks real Karing smoke.
- Confirmed no new validated worker completion or fresh Production receipt.
- Updated canonical documentation directly on `main`; no runtime/schema work integrated.
- No merge/deploy/migration/restart/reload/DB write/credential/backup/rollback mutation performed.

## Immediate execution order

1. Task16: run the full four-gate suite on one exact current SHA; if CI fails, fix only generic latest-schema/RLS expectations and preserve schema20-specific Task15 fixtures.
2. Task13: reconstruct validated work onto current main and run fresh HTTP/1.1 + HTTP/2 proof outside Production.
3. PR #4: obtain one reproducible real Karing client smoke with version/platform/import/connect evidence.
4. Independent review: inspect Task13/Task16 diffs for security, accounting, RLS and rollback regressions.
5. Only after required evidence is green: obtain fresh encrypted Production backup + rollback state, then consider promotion.

Never claim completion from stale worker reports or partial evidence. No merge/deploy until all exact-head gates and safety prerequisites are green.
