# PVNaive — Agent / Workstream Task Board

Last updated: 2026-09-13 23:35 Asia/Tehran

This board is the active execution queue. Exact repository, CI and fresh Production evidence decide completion.

## Shared rules
- Start from latest verified `main`; never from chat snapshots or stale PR heads.
- Preserve truthful accounting, commit-before-success, session identity, trusted peer/IP semantics and credential lifecycle.
- TDD/fail-first for behavior changes; one writer per branch/worktree and independent verification for promotion-sensitive work.
- Never expose passwords, tokens, keys or secret values in Git, CI, comments or evidence.
- Never credit route/schema presence, static inspection, assignment-only comments, stale heads, missing-tool runs or historical Production evidence as acceptance.
- Production is not a development worker. Mutation requires exact-head green CI, fresh encrypted backup, rollback snapshot, reconciled migration lineage and postflight.

## Verified repository baseline
- Current canonical `main`: `61c613777ec2dcf3f1559cbb8b8df80f8e389af0`; push CI `34779414593`: SUCCESS.
- Runtime-bearing baseline: `f0cdab3eda205eecb3285577ec5df32f45d0ddb7`; schema22 is canonical and merged-main CI is green.

## Hard deployment blockers
| ID | Priority | Status | Required next proof |
|---|---|---|---|
| DEPLOY-001 / #105 | P0 | BLOCKED_SOURCE | recover exact trusted all-in-one renderer/build context; RED complete-credential preservation regression before implementation; #103 closed duplicate |
| LINEAGE-001 / #104 | P1 | BLOCKED_PRIMARY | exact read-only Production migration ledger/checksums and trusted 0022..0027 artifacts; then forward-only compatibility fixture/tests without rewriting history |
| PROD-AUDIT / #100 | P0 | BLOCKED_HOST | Primary offline; fresh deployed revision/schema/services/Caddy/backup/disk/rollback receipt required |

## Active roadmap lanes
| Task | Status | Next action / acceptance gate |
|---|---|---|
| Schema22 self-service auth context | DONE_REPO | merged and green; no Production claim |
| Task13 exact kill/disconnect | IN_PROGRESS/BLOCKED_HOST | PR #64 stale; Worker 3 reconstructs from current main, Worker 2 independently proves exact-head HTTP1+HTTP2 target-only kill, sibling survival, forged tuple rejection, idempotency and exactly-once final accounting |
| Task16 bounded IP/session history | DONE_REPO | retained and green under schema22 regression; no Production claim |
| Karing compatibility | IN_PROGRESS/BLOCKED_CLIENT | PR #101 exact-head repo gates green but stale/non-mergeable; Worker 4 obtains real disposable import → parse → CONNECT → cleanup/revoke receipt and reconstructs on current main if required; Worker 1 reviews receipt |
| Credential-preserving renderer | BLOCKED_SOURCE | coordinator/first capable worker must recover trusted renderer provenance before writing implementation |
| Migration lineage reconciliation | BLOCKED_PRIMARY | coordinator can prepare strategy, but exact Production ledger/artifacts are mandatory before equivalence/compatibility claims |
| Task36 auth/IDOR/CSRF/redaction/fuzz | QUEUED | independent negative-gate lane after P0/P1 blockers are controlled |
| Installer/upgrade/rollback proof | QUEUED | only after lineage/renderer blockers are resolved |
| Final load/capacity + RC | QUEUED | only after runtime/client/deployment blockers and exact Production smoke are complete |

## Worker allocation
Current remote inventory: **no online executable PVNaive worker**.

- Worker 4 / `ubuntu-4gb-hel1-1`: queued — Karing smoke and latest-main reconstruction if needed.
- Worker 3 / `TrPaqet`: queued — Task13 reconstruction from current verified main.
- Worker 2 / `RoboT`: queued — independent Task13 race/permission/protocol/accounting rehearsal.
- Worker 1 / `Pak-Nasheeee-haaaaaaaaa`: queued — independent evidence/security/accounting review where installed tools are sufficient.
- Primary / `testAmir5-3`: queued — read-only Production audit only.
- Coordinator / GitHub lane: active — exact state reconciliation, DEPLOY-001 source-provenance recovery, LINEAGE-001 coordination and safe integration only.

## Promotion order
1. Recover and validate DEPLOY-001 source/renderer behavior and reconcile LINEAGE-001.
2. Obtain Karing real-client acceptance on a current-main reconstruction.
3. Reconstruct and independently validate Task13 exact head.
4. Reconnect Primary and perform fresh read-only Production audit.
5. Create fresh encrypted backup and independent rollback snapshot.
6. Lock exact deploy SHA and perform staged promotion only if every gate remains green.
7. Run postflight and retain rollback until acceptance is complete.

## Mandatory work-unit report
```text
TASK #:
STATUS:
WHAT CHANGED:
FILES:
TESTS:
CI:
PRODUCTION:
EVIDENCE:
REMAINING:
NEXT TASK:
```
