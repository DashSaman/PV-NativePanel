# PVNaive — Agent / Workstream Task Board

Last updated: 2026-09-14 02:32 Asia/Tehran

## Shared rules
- Start from latest verified `main`; preserve truthful session/accounting identity and credential lifecycle semantics.
- TDD for behavior changes; independent proof for promotion-sensitive work.
- Production mutation requires fresh exact-head gates, fresh read-only audit, encrypted backup and rollback snapshot.

## Verified baseline
- `main` `879219990539b676050877023ceb68f2951f39ea`; CI `34787434163` SUCCESS.
- Renderer/build provenance and migration-lineage reconciliation are CLOSED by batch-2 evidence. Repo schema is 28.
- Task13 active PR #108 head `41b7bcea78b2f3e1298077e6e7ba4d7cba724bed`: CI/accounting/pinned-forwardproxy green; independent Go/web/contracts/forwardproxy checks green.
- Karing PR #101 remains DRAFT and lacks real-client proof.
- Production nip.io fallback is externally live/ready/TLS-valid; owner domain TLS is in documented ACME duplicate-limit window.

## Active lanes
| Lane | Status | Acceptance / next action |
|---|---|---|
| Task13 exact kill/disconnect | IN_PROGRESS | Worker 2 real HTTP1+HTTP2 pinned-Caddy proof; keep #108 DRAFT until target/sibling/forged/idempotency/credential/Caddy/exactly-once accounting gates pass |
| Karing compatibility | BLOCKED_CLIENT | Worker 4 disposable real import → parse → CONNECT → cleanup/revoke receipt; then latest-main reconstruction if needed |
| Production audit | BLOCKED_PRIMARY | Primary read-only deployed identity/schema/services/Caddy/backups/disk/rollback receipt |
| Owner-domain TLS | WAIT_ACME | Do not mutate; external recheck after recorded retry-after 2026-09-15 03:17:36 UTC |
| Task36 security negatives | QUEUED | Worker 1 may advance non-conflicting evidence/tests while runtime lanes wait |
| Installer/upgrade/rollback | QUEUED | Advance tests/docs only unless Production promotion prerequisites are fully green |

## Worker allocation
- Worker 4 / `ubuntu-4gb-hel1-1`: Karing real-client acceptance.
- Worker 3 / `TrPaqet`: PR #108 review/fix if new exact-head finding appears.
- Worker 2 / `RoboT`: independent real Task13 HTTP1/HTTP2 protocol/accounting rehearsal.
- Worker 1 / `Pak-Nasheeee-haaaaaaaaa`: ONLINE; exact-head verification, Docker-Go tests, static/security review and safe branch prep.
- Primary / `testAmir5-3`: read-only Production audit when connected.
- Coordinator: reconcile CI/receipts, integrate only validated work, keep docs canonical.

## Promotion order
Task13 real proof + Karing real-client proof → fresh Production audit → encrypted backup + rollback snapshot → exact deploy SHA → staged deploy → postflight → retain rollback. Owner-domain ACME recovery is externally verified separately and must not be forced by unsafe restarts.
