# PVNaive — Agent / Workstream Task Board

Last updated: 2026-09-14 00:39 Asia/Tehran

This board is the active execution queue. Exact repository, CI and fresh Production evidence decide completion.

## Shared rules
- Start from latest verified `main`; preserve truthful accounting/session identity/credential lifecycle.
- TDD for behavior changes; one writer per worktree; independent verification for promotion-sensitive work.
- Never expose secrets or credit stale/static/assignment-only/missing-tool evidence as acceptance.
- Production mutation requires fresh audit, exact-head green gates, encrypted backup, rollback snapshot and reconciled lineage.

## Verified repository baseline
- Pre-doc-update main: `f4d3cc0b59d10c4507d27c86aabac9bd833d2898`; CI `34781998537`: SUCCESS.
- Fresh Task13 reconstruction: PR #107 head `25b214cf019d3db2321d9651ab3654e46ef44342`, replayed conflict-free from the completed worker artifact; `git diff --check` passed.
- Karing PR #101 remains stale/DRAFT at `216d53670066033403fe95f61b0402bb710186a3` and still lacks real-client proof.

## Hard deployment blockers
| ID | Priority | Status | Required next proof |
|---|---|---|---|
| DEPLOY-001 / #105 | P0 | BLOCKED_SOURCE | recover exact trusted renderer/build provenance; preservation regression before implementation |
| LINEAGE-001 / #104 | P1 | BLOCKED_PRIMARY | exact read-only Production ledger/checksums + trusted 0022..0027 artifacts; forward-only compatibility tests |
| PROD-AUDIT / #100 | P0 | BLOCKED_PRIMARY | fresh deployed revision/schema/services/Caddy/backup/disk/rollback receipt |

## Active roadmap lanes
| Task | Status | Next action / acceptance gate |
|---|---|---|
| Task13 exact kill/disconnect | IN_PROGRESS | PR #107 current-main reconstruction; exact-head CI/focused gates then independent HTTP1+HTTP2 live accounting/session proof |
| Karing compatibility | BLOCKED_CLIENT | real disposable import → parse → CONNECT → cleanup/revoke; reconstruct on latest main after proof if needed |
| Credential-preserving renderer | BLOCKED_SOURCE | trusted source provenance before implementation |
| Migration lineage reconciliation | BLOCKED_PRIMARY | exact Production ledger/artifacts before compatibility claims |
| Task36 security negatives | QUEUED | advance independently when a suitably tooled worker is free |
| Installer/upgrade/rollback | QUEUED | after renderer/lineage blockers are controlled |

## Worker allocation
Current remote inventory: Worker 1 / `Pak-Nasheeee-haaaaaaaaa` ONLINE; it lacks Go.

- Worker 4 / `ubuntu-4gb-hel1-1`: Karing real-client acceptance + cleanup/revoke evidence.
- Worker 3 / `TrPaqet`: review/fix PR #107 code reconstruction against current main.
- Worker 2 / `RoboT`: independently run exact-head Task13 race/permission/protocol/accounting rehearsal.
- Worker 1 / `Pak-Nasheeee-haaaaaaaaa`: branch preparation, static/evidence/security review; no Go acceptance claim.
- Primary / `testAmir5-3`: read-only Production audit only when connected.
- Coordinator / GitHub: CI, provenance/lineage coordination and safe integration only.

## Promotion order
1. Control DEPLOY-001 and LINEAGE-001.
2. Obtain Karing real-client proof.
3. Validate PR #107 exact head with focused + real protocol/accounting rehearsal.
4. Fresh Production read-only audit.
5. Fresh encrypted backup + independent rollback snapshot.
6. Exact deploy SHA, staged promotion, postflight, rollback retained.
