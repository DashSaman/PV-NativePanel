# PVNaive — Canonical Project Status

Last updated: 2026-09-13 00:38 Asia/Tehran

## Verified GitHub state
- Repository `DashSaman/PV-NativePanel` is accessible; `main` is the default branch.
- Current authoritative pre-reconciliation `main`: `1703950dcfe2e539aa4853910b014d2e991ea992`.
- Exact-main combined status is `pending` with zero published statuses, and the commit-specific workflow lookup returns no workflow runs; CI is therefore **not claimed green** for this checkpoint.
- The last validated runtime integration remains Task16/schema21, merged as `7efa359ccc5745c548cda9590bc5c516e9d5aa9e` from exact head `904e17c4a013e3adb5fb349c70f254ab59c925f8`.
- PR #64 / Task13 remains OPEN/DRAFT/non-mergeable at exact head `3fc14825e1b164bad558decaef47f56b792e81af`, based on stale history; fresh pinned-Caddy HTTP/1.1 + HTTP/2 rehearsal and exact accounting proof remain mandatory.
- PR #4 / Karing remains OPEN/DRAFT at exact head `2501e39dc39e14063b6a501bc96b77bbfcae7384`, based on `s04-auth`; independent real-client import/parse/connect/cleanup smoke remains mandatory.
- Obsolete documentation-only PRs #85, #86, #87, #88, #89, #91, #92, #93, #94 and #95 were closed without merge this cycle; their historical evidence remains available but they no longer clutter the active queue.

## Worker / coordinator truth
- Persistent coordinator/worker reports and the latest PR/issue discussions were inspected this cycle.
- No fresh exact-head completion receipt was found for Task13, Karing, issue #99, or issue #100.
- Historical, stale, dirty, mixed-head, unpushed, or absent-CI evidence remains uncredited.
- TrPaqet remains the isolated Task13 rehearsal lane in the persisted coordination record; other worker lanes require fresh exact-SHA receipts before credit.

## Production truth
- No connected command-level Production audit is available in this runtime. Issue #100 likewise contains dispatches but no fresh audit receipt.
- Therefore deployed SHA/schema, current service/process state, fresh encrypted backup, independent rollback snapshot, staged promotion, and postflight are **not re-verified in this cycle**.
- External Production health is not overclaimed and Production was not touched.
- Promotion remains gated by read-only audit → fresh encrypted backup → independent rollback snapshot → exact SHA lock → staged promotion → health/postflight → retained rollback.

## Actions completed this cycle
- Re-verified current `main`, exact-main CI visibility, open PRs, Task13/Karing discussions, and security/Production worker issues.
- Reconciled worker state: no validated runtime completion is available to merge.
- Closed ten obsolete documentation-only PRs without merge.
- Refreshed canonical handoff/progress documentation with the actual verified SHA and open gates.
- Re-dispatched the independent next tasks: Task13 exact-head reconstruction/rehearsal, Karing compatibility smoke, schema21 security/accounting review, and read-only Production audit/rollback inventory.

## Next executable gates
1. Task13: rebase/reconstruct onto current `main`, publish a new exact head without rewriting validated history, rerun exact-head CI/accounting/forwardproxy gates, then execute isolated real HTTP/1.1 + HTTP/2 rehearsal with target-only kill, sibling survival, forged-tuple rejection, repeat-kill idempotency, credential survival, no restart/reload, and exactly-once accounting.
2. Karing: rebase/reconstruct onto current `main` as needed, rerun exact-head CI, then run real import/parse/connect/cleanup smoke with exact profile hash, client/platform/version, disposable credentials, non-Production target, and redacted logs.
3. Security/accounting review: inspect merged Task16/schema21 on a clean current-main worktree for RLS fail-closed behavior, privilege separation, retention/purge safety, trusted session lineage, commit-before-HTTP-success semantics, and secret redaction.
4. Production: perform read-only command-level audit first. Only after all runtime gates are green, create a fresh encrypted backup and independent rollback snapshot, lock the deploy SHA, stage promotion, verify health/postflight, and retain rollback.

Never claim completion from stale reports, older heads, partial evidence, mixed-head proofs, dirty worktrees, or historical Production records.
