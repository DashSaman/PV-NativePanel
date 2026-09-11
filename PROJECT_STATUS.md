# PVNaive — Canonical Project Status

Last updated: 2026-09-11 18:38 Asia/Tehran

## Verified GitHub state
- Current `main`: `239f595a9a34a283564df3a2e531dd055dbf1280`.
- Current `main` combined status is `pending` with zero published status entries; no fresh post-update CI green result is claimed for the docs-only checkpoint.
- Task16 PR #81 is merged as `7efa359ccc5745c548cda9590bc5c516e9d5aa9e` after four exact-head green gates on `904e17c4a013e3adb5fb349c70f254ab59c925f8`.
- PR #64 Task13 remains OPEN / DRAFT at exact head `3fc14825e1b164bad558decaef47f56b792e81af`; focused checks are supplemental; fresh real HTTP/1.1 + HTTP/2 rehearsal remains mandatory.
- PR #4 Karing remains OPEN / DRAFT at exact head `2501e39dc39e14063b6a501bc96b77bbfcae7384`; independent real-client import/parse/connect/cleanup proof remains missing.
- PR #95 and older documentation PRs remain stale-base reconciliation attempts; keep uncredited unless rebased and freshly validated.

## Worker / coordinator truth
- Task16 completion remains credited only for the merged, four-gate-validated head above.
- Persistent reports contain no fresh exact-head completion receipt for Task13 or Karing; only historical/focused evidence is present.
- Historical worker-only, stale, dirty, mixed-head, and unpushed output remains uncredited.
- Persistent material identifies TrPaqet as the active executable development slot; other lanes are inactive or upgrade-required under the one-active-host constraint.
- Fresh exact-head dispatches were posted to PR #64 and PR #4 in this cycle.

## Production truth
- No fresh command-level Production audit, deployed SHA/schema verification, encrypted backup, independent rollback snapshot, staged deploy, or postflight was available through connected tools in this cycle.
- No Production mutation occurred.
- Promotion still requires read-only audit, fresh encrypted backup, exact SHA lock, staged deploy, health/postflight, and rollback readiness.

## Next executable gates
1. Task13: isolated real HTTP/1.1 + HTTP/2 rehearsal proving target-only kill, sibling survival, forged-tuple rejection, repeat-kill idempotency, credential survival, no restart/reload, and exactly-once accounting.
2. Karing: real import/parse/connect/cleanup smoke with disposable credentials, exact profile hash, client/platform/version, and redacted logs.
3. Independent review: RLS, privilege separation, retention/purge safety, accounting/session lineage, and secret redaction after Task16 merge.
4. Read-only Production audit, then backup/rollback/staged promotion only after all runtime evidence is complete.

Never claim completion from stale reports, older heads, partial evidence, mixed-head proofs, or dirty worktrees.
