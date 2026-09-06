# PVNaive — Canonical Project Status

Last updated: 2026-09-06 07:39 Asia/Tehran

This file records verified repository truth and bounded Production truth. Historical worker/stage notes are evidence only; exact GitHub state, exact-head CI and fresh Production observations override them.

## Safety invariants

PVNaive remains standalone-first. Never fabricate usage/online/IP/session history. Never rotate credentials/tokens from read-only flows. Production changes require a fresh encrypted backup + rollback state, intended migrations only, exact artifact provenance and postflight verification.

## Repository truth

- Repository: `DashSaman/PV-NativePanel`.
- Current `main`: `cf8e90298ad3c0dca20fb1010ce12fe763f84df6`.
- Exact `main` combined status: no status rows returned. Post-merge CI is not credited.
- Task13: draft PR #64, head `3fc14825e1b164bad558decaef47f56b792e81af`; focused/exact-head historical gates are successful, but fresh real HTTP/1.1 + HTTP/2 rehearsal remains incomplete. PR metadata/body still contains stale historical main references.
- Task16: draft PR #81, head `3c4310335ab4907d28bac995bba1be3545e14f6e`; PR metadata/body records historical schema21 evidence and prior generic database CI failure. No fresh exact-head all-green proof was observed in this run.
- PR #4 (Karing export): draft, head `2501e39dc39e14063b6a501bc96b77bbfcae7384`, base `s04-auth`; CI evidence is historical and one real Karing client smoke is still required.
- Documentation-only PRs remain open/stale and are not promotion authority.

## Production truth

- No fresh command-level Production audit was available in this run; no Production health pass is claimed.
- No Task13, Task16 or PR #4 code is authorized as deployed from current evidence.
- No deploy, migration, restart, reload, DB write, credential change, backup mutation or rollback mutation was performed.

## Persistent worker state

- Persistent reports were searched; no fresh worker completion receipt tied to the current GitHub heads was available for reconciliation.
- Historical reports remain non-authoritative without exact GitHub corroboration and fresh receipts.
- Latest bounded lane plan: TrPaqet → Task13 rehearsal; PostgreSQL18-capable worker → Task16 fixture/CI reconciliation; independent worker → regression/security review; `pv-primary` → Production-only when an executable slot is available.

## This run — 2026-09-06 07:39 Asia/Tehran

- Verified current main ref, open PRs #4/#64/#81, exact-head status presence and current PR metadata.
- Confirmed no combined status rows for current `main`.
- Reconciled that no historical worker completion can be credited.
- Confirmed no merge/deploy safety gate is green enough for mutation.
- Refreshed canonical status without integrating unvalidated runtime/schema work.

## Immediate execution order

1. Keep #64, #81 and #4 draft / DO NOT MERGE until their specific evidence gates are complete.
2. Task13: reconstruct validated work onto current main and run fresh HTTP/1.1 + HTTP/2 proof outside Production.
3. Task16: fix/reconcile generic schema21 fixtures on a clean branch, preserve schema20-specific Task15 fixtures, align PR metadata and rerun all required gates on one exact SHA.
4. PR #4: obtain one real Karing client smoke with version/platform/import/connect evidence.
5. Only after required evidence is green: obtain fresh encrypted Production backup + rollback state, then consider promotion.

Never claim completion from stale worker reports or partial evidence. No merge/deploy until all exact-head gates and safety prerequisites are green.
