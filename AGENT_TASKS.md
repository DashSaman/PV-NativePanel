# PVNaive — Agent / Workstream Task Board

Last updated: 2026-09-13 23:30 Asia/Tehran

This board is the active execution queue. `OWNER_REQUIREMENTS.md` and `ROADMAP.md` define product intent; exact repository, CI and fresh Production evidence decide completion.

## Shared rules
- Start from latest verified `main`; never from chat snapshots or stale PR heads.
- Preserve truthful accounting, commit-before-success, session identity, trusted peer/IP semantics and credential lifecycle.
- TDD/fail-first for behavior changes; one writer per branch/worktree and independent verification for promotion-sensitive work.
- Never expose passwords, tokens, keys or secret values in Git, CI, comments or evidence.
- Never credit route/schema presence, static inspection, assignment-only comments, stale heads or historical Production evidence as acceptance.
- Production is not a development worker. Mutation requires exact-head green CI, fresh encrypted backup, rollback snapshot, reconciled migration lineage and postflight.

## Verified repository baseline
- Runtime-bearing `main`: `f0cdab3eda205eecb3285577ec5df32f45d0ddb7`.
- Push CI `34779141898`: SUCCESS across web, Go, PostgreSQL 18 database suite, runtime rehearsal and bundle.
- PR #102 schema22 auth-context/security repair: MERGED after exact-head full CI + dedicated Schema22 Auth Context + Task16 + Exact Accounting + Pinned Forwardproxy gates were green.
- Repository canonical schema: 22.

## Hard deployment blockers
| ID | Priority | Status | Required next proof |
|---|---|---|---|
| DEPLOY-001 | P0 | OPEN | boot/recreate Caddy renderer must preserve the complete active credential set; add regression + recreate/restart proof before deploy |
| LINEAGE-001 | P1 | OPEN | reconcile recorded Production migration lineage with repository lineage without rewriting applied history; document safe forward compatibility path |
| PROD-AUDIT | P0 | BLOCKED | Primary is offline; fresh read-only deployed revision/schema/services/Caddy/backup/disk/rollback receipt required |

## Active roadmap lanes
| Task | Status | Next action / acceptance gate |
|---|---|---|
| Schema22 self-service auth context | DONE_REPO | merged at `f0cdab3...`; merged-main CI green; Production deployment intentionally not attempted |
| Task13 exact kill/disconnect | IN_PROGRESS/BLOCKED_HOST | PR #64 is stale; Worker 3 reconstructs from current main, Worker 2 independently proves exact-head HTTP1+HTTP2 target-only kill, sibling survival, forged tuple rejection, idempotency and exactly-once final accounting |
| Task16 bounded IP/session history | DONE_REPO | schema21 contract is retained and green under schema22 latest-schema regression; no Production claim |
| Karing compatibility | IN_PROGRESS/BLOCKED_CLIENT | PR #101 needs real disposable import → parse → CONNECT → cleanup/revoke evidence; Worker 4 owns smoke, Worker 1 reviews receipt |
| Production audit | BLOCKED_HOST | Primary read-only only when connected; no mutation until DEPLOY-001 + LINEAGE-001 are resolved and promotion gates are green |
| Credential-preserving renderer | NEXT | Coordinator may implement repository-only TDD fix independently of Production connectivity |
| Migration lineage reconciliation | NEXT | Coordinator may analyze repository/recorded Production lineage and produce non-destructive compatibility plan independently |
| Task36 auth/IDOR/CSRF/redaction/fuzz | QUEUED | independent negative-gate lane after higher P0/P1 blockers are controlled |
| Installer/upgrade/rollback proof | QUEUED | only after lineage/renderer blockers are resolved |
| Final load/capacity + RC | QUEUED | only after runtime/client/deployment blockers and exact Production smoke are complete |

## Worker allocation
Current remote inventory: **no online executable PVNaive worker**.

- Worker 4 / `ubuntu-4gb-hel1-1`: queued — real Karing smoke on PR #101 with disposable non-Production credentials.
- Worker 3 / `TrPaqet`: queued — Task13 reconstruction from current verified main.
- Worker 2 / `RoboT`: queued — independent Task13 race/permission/protocol/accounting rehearsal on reconstructed exact head.
- Worker 1 / `Pak-Nasheeee-haaaaaaaaa`: queued — independent evidence/security/accounting review where installed tools are sufficient; do not credit missing Go/Karing capabilities.
- Primary / `testAmir5-3`: queued — read-only Production audit only.
- Coordinator / GitHub lane: active — DEPLOY-001 and LINEAGE-001 repository-safe work, CI reconciliation and integration.

## Promotion order
1. Resolve DEPLOY-001 and LINEAGE-001 in repository with independent tests/evidence.
2. Obtain Karing real-client acceptance.
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
