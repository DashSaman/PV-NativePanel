# PVNaive — Canonical Project Status

Last updated: 2026-09-13 21:40 Asia/Tehran

## Verified GitHub state
- Exact pre-refresh `main`: `507f074d234d6548fbbda21e06b3365353f996be`.
- Push CI run `34771275246` for that exact SHA completed SUCCESS.
- Latest main changes remain documentation-only; last validated merged runtime integration remains Task16/schema21 PR #81.
- PR #64 / Task13 remains OPEN/DRAFT at `3fc14825e1b164bad558decaef47f56b792e81af`; GitHub currently reports `mergeable=false` and the exact-head live protocol/accounting gate is unsatisfied.
- PR #101 / Karing remains OPEN/DRAFT at `216d53670066033403fe95f61b0402bb710186a3`; GitHub currently reports `mergeable=false`; repository/unit/build evidence is green but the real-client gate is unsatisfied.
- Issue #100 remains the read-only Production status lane.

## Fresh independent work this cycle
- Remote inventory still exposes only Worker 1 / `Pak-Nasheeee-haaaaaaaaa`; Production Primary and Workers 2/3/4 are not connected.
- Worker 1 ran a clean exact-main Task16/schema21 PostgreSQL 18 contract in a disposable container. `tests/db/ip_session_history_contract_test.sh` returned `TASK16_IP_SESSION_HISTORY_PG18=PASSED`; the container was removed afterward.
- Static review reconfirmed schema21 ENABLE+FORCE RLS, base-table revoke from app/public, fixed-search-path SECURITY DEFINER functions, bounded read permission for the app role, maintenance-only sync permission, explicit purge confirmation, owner-role requirement, advisory lock, single-transaction purge, and strict `final_at < observed_at - interval '30 days'` retention semantics.
- The rollback script emitted PostgreSQL warnings that `SET LOCAL` can only be used in transaction blocks, but the contract assertions still passed through rollback to schema20 with the history table absent. This warning is recorded for cleanup review and is not misreported as a failed rollback.
- Selected schema21/purge/test secret scan found identifiers only and no embedded credential values.

## Runtime gates
- Karing #101: exact head `216d5367...`; existing exact-head GitHub workflows and fresh Worker-1 `git diff --check`, `npm test` (19 files / 63 tests), and `npm run build` remain valid. Missing: real Karing import/parse/connect/cleanup receipt with disposable non-Production credentials, profile SHA-256, redacted connection evidence, and cleanup/revoke proof.
- Task13 #64: stale head `3fc14825...`; Worker 1 lacks Go and cannot supply the required race/session-control or pinned-Caddy rehearsal. Missing: current-main reconstruction plus independent exact-head HTTP/1.1 + HTTP/2 proof of target-only termination, sibling survival, forged-tuple rejection, repeat-kill idempotency, credential survival, no Caddy lifecycle action, and exactly-once final accounting.

## Worker / coordinator allocation
- Worker 1 / `Pak-Nasheeee-haaaaaaaaa`: online; independent repository/security/accounting review and evidence verification only. Continue issue #99 with static/source review of commit-before-HTTP-success and redaction paths while blocked runtime lanes wait for capable hosts.
- Worker 4 / `ubuntu-4gb-hel1-1`: real Karing smoke when connected.
- Worker 3 / `TrPaqet`: Task13 reconstruction from latest verified main when connected.
- Worker 2 / `RoboT`: independent Task13 exact-head race/permission/protocol/accounting rehearsal when connected.
- Primary / `testAmir5-3`: read-only Production audit only when connected.
- One writer per worktree; independent verifier on a different worker; no unrelated host changes.

## Production truth
- Production Primary is not connected and issue #100 has no fresh command-level receipt.
- Deployed SHA/schema, service/readiness/listeners, Caddy lifecycle/build identity, session-control socket, backup freshness/encryption, and rollback snapshot remain unverified this cycle.
- Production was not mutated. No backup, migration, restart/reload, credential/DB/Caddy change, deploy, or rollback-state change was performed.

## Promotion order
Karing real-client receipt and independent review → Task13 current-main reconstruction and exact-head live proof → fresh Production read-only audit → fresh encrypted backup → independent rollback snapshot → exact deploy-SHA lock → staged promotion → postflight with rollback retained.
