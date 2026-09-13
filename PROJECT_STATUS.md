# PVNaive — Canonical Project Status

Last updated: 2026-09-14 00:39 Asia/Tehran

## Verified GitHub state
- Current canonical `main` tip before this documentation update: `f4d3cc0b59d10c4507d27c86aabac9bd833d2898`; push CI run `34781998537` completed SUCCESS.
- The runtime-bearing change since the prior coordinator checkpoint is `6a0258775061af70e357821dfa51cd62f74fecc0`; current main CI is green. Its actual diff is test-only (7 database/stage test files); the commit message mentions a 0022 migration/SHA manifest repair, but those files were already identical at parent `db9e394...`. Do not credit nonexistent migration-file changes to this commit.
- PR #101 / Karing remains OPEN/DRAFT at `216d53670066033403fe95f61b0402bb710186a3`, stale/non-mergeable and still missing the required real Karing import/parse/CONNECT/cleanup receipt.
- PR #64 is now historical/stale Task13 evidence at `3fc14825e1b164bad558decaef47f56b792e81af`.
- Fresh Task13 reconstruction PR #107 is OPEN/DRAFT from exact current main; head `25b214cf019d3db2321d9651ab3654e46ef44342`. The prior worker reconstruction replayed without conflicts and `git diff --check` passed. Full exact-head CI/live protocol evidence is still required before merge.

## Production truth and hard gates
- Repository live-session notes report Production domain `namir.softarg.ir`, image `pvnaive:fix2`, schema through deployed 0027, BBR/fq, live self-service account security, and a resolved readiness mismatch via compose override. These are recorded operational facts, not a fresh command-level audit from this coordinator cycle.
- Production Primary is still not connected through the command channel; current deployed revision/schema/services/Caddy/backup/disk/rollback state therefore remains unverified this cycle.
- `DEPLOY-001` remains issue #105: trusted all-in-one renderer/build provenance is still required before repository implementation.
- `LINEAGE-001` remains issue #104: exact Production migration ledger/checksums plus trusted 0022..0027 artifacts are mandatory before forward-only reconciliation.
- No Production mutation was performed this cycle: no backup creation, migration, restart/reload, credential/DB/Caddy change, deploy, or rollback-state change.

## Work completed this cycle
- Confirmed exact current-main CI success at `f4d3cc0...` / run `34781998537`.
- Reconciled live worker inventory; Worker 1 / `Pak-Nasheeee-haaaaaaaaa` is online again.
- Fetched latest main on Worker 1 and created a clean detached verification worktree. `git diff --check db9e394..f4d3cc0` passed; Go execution is unavailable on that worker (`go: command not found`).
- Found an existing completed Task13 reconstruction worker artifact at `2c4a85b...`; replayed it cleanly onto current main and produced `25b214c...` with 34 changed files and no whitespace errors.
- Worker OAuth refused workflow-file modification, so the code reconstruction was pushed without changing `.github/workflows/ci.yml`; GitHub PR #107 now carries the fresh reconstruction and remains draft until authorized exact-head focused gates run.

## Worker / coordinator allocation
- Worker 4 / Karing: real disposable Karing import → parse → CONNECT → cleanup/revoke proof; then latest-main reconstruction if still needed.
- Worker 3 / Task13 implementation: review PR #107 reconstruction and resolve any code-level CI findings.
- Worker 2 / Task13 verifier: independent exact-head race/permission plus real HTTP/1.1 + HTTP/2 target-only kill/sibling/forged-tuple/idempotency/credential/accounting rehearsal.
- Worker 1 / `Pak-Nasheeee-haaaaaaaaa`: online; evidence/diff/security review and branch preparation where installed tooling permits. It currently lacks Go.
- Primary / Production: read-only audit only when connected.
- Coordinator / GitHub: CI reconciliation, DEPLOY-001 provenance, LINEAGE-001 evidence coordination and safe integration only.

## Promotion order
Reconcile DEPLOY-001 + LINEAGE-001 → Karing real-client receipt → PR #107 exact-head CI and live Task13 proof → fresh Production read-only audit → fresh encrypted backup → independent rollback snapshot → exact deploy-SHA lock → staged promotion → postflight with rollback retained.

Never credit assignment-only, stale-head, static-only, historical, missing-tool or inferred evidence as completion.
