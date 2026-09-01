# PVNaive — Canonical Handoff

Last updated: 2026-09-01

Resume from this file plus `CONTINUE_HERE.md`, `PROJECT_STATUS.md`, newest `ops/evidence/*`, and a fresh GitHub/Production audit. Older S04/S05/S06/task checkpoints are historical.

## Repository / release truth

- Repository: `DashSaman/PV-NativePanel`
- Product: **PVNaive**
- Current GitHub `main`: `fe6a9fea76fa48577fb8063bb246563f2696846b`
- Exact-main CI `33453736537`: **SUCCESS**
- Only open unrelated PR at audit start: old draft #4.
- Latest independently recorded Production schema: **19**.
- Task12: **DONE / Production**, schema17.
- Task14: **DONE / Production**, schema19.
- Task35 BUG-001/002/003: **DONE in main**.

## Production access

Fresh SentinelX state: three hosts connected, one active allowed. `pv-primary` is connected/healthy in the control plane but even read-only command execution returns `upgrade_required`; the persistent worker host is the active host. Therefore no fresh Production-health claim, backup, deploy, migration, Caddy mutation or schema change was made in this coordinator run.

Human action required before Production work: make `pv-primary` active by disconnecting another host or changing the SentinelX plan. When restored: read-only audit first → fresh encrypted backup + rollback snapshot → intended exact release/migration only → readiness/Runtime/Telemetry/Caddy/customer/accounting postflight → exact deployed provenance; rollback on any failed invariant.

## Task13 — exact one-session disconnect: IN PROGRESS / RECOVERY

The previously documented persistent candidate commit `922a5e0e155746906f28fe46ca89a24f269acfa7` cannot be resolved on the currently active worker and must not be treated as a still-present local candidate.

Remote branch `lead/task13-kill-session-publish-20260901` is partial and diverged from current main; it currently publishes only four Task13 files relative to main. A recovery bundle `/tmp/task13-reference` contains eight exact previously reviewed files. Hash checks prove published objects such as sessioncontrol `client.go` match the recovery bundle byte-for-byte.

A current-main recovery worktree has been created. Recovered `internal/sessionkill` and `internal/sessioncontrol` compile/test. HTTP API compilation intentionally exposes remaining missing wiring: `ServerConfig.SessionKiller` and the real `users.sessions.delete` route are absent on current main. Continue reconstructing from preserved evidence/repository patterns; do not merge the partial branch or claim the old 18-file source tree still exists.

Historical evidence `ops/evidence/TASK13-20260901-prepublication-verification.md` proves a full candidate previously passed Go/race/Web/rehearsal/pinned-forwardproxy/reproducible-Caddy plus real HTTP/1.1+HTTP/2 exact-kill rehearsal. The reconstructed current-main source must repeat those gates before publication/merge.

Task13 invariants: exact full runtime-credential/node/boot/session tuple; no credential-wide revoke; no Caddy reload/restart; forced disconnect produces normal exact final accounting once; repeated kill idempotent; tenant/role isolation; redacted audit; sibling sessions survive.

## Task15 — unique-IP limit: IN PROGRESS / LOCAL CANDIDATE VERIFIED

Current local branch: `lead/task15-unique-ip-schema20-20260901`.
Commit: `2b175991f1d5628dc084f4ffddfea6b63d960bf8`.
Parent: exact main `fe6a9fea76fa48577fb8063bb246563f2696846b`.

Fresh PASS: `git diff --check`, `go vet ./...`, full Go tests, focused Go race tests, Web 18 files / 61 tests, Web production build, schema20 test-script syntax.

The redesign uses only trusted Caddy `RemoteAddr`, propagates it through exact accounting, serializes admission by locking the ServiceTerm row, counts canonical active `direct_naive_accounting_session_peers`, and records peer state only after the schema19 accounting call actually accepts the new session. `Forwarded`/`X-Forwarded-For` are not enforcement inputs.

Review fixed the concurrency test so it does not pretend PostgreSQL lock acquisition has a deterministic session-ID winner. It now proves exactly one accepted + one limited result and that the persisted session equals the caller that actually won.

Blockers before integration:

- worker has PostgreSQL 14.24 only; this cannot prove the modern migration baseline (`security_invoker` unsupported). PostgreSQL18 CI must run `tests/db/unique_ip_limit_migration_test.sh` after publication;
- worker HTTPS `git push` has no GitHub credential. Publish via authenticated Git/GitHub Git-data transport, then require exact-head gates.

See `ops/evidence/TASK15-20260901-schema20-candidate-verification.md`. Task15 is **not merged/deployed**.

## Task16 — bounded IP/session history: IN PROGRESS

A real isolated worker is active on `worker/task16-schema21-20260901`, based on Task15 candidate `2b175991...` to preserve contiguous schema20→21 ordering. Keep integration behind stable schema20. Required: exact 30-day retention, tenant RLS, canonical session/peer/accounting facts only, final-accounting synchronization, no fabricated legacy history, maintenance-only purge, bounded pagination, coherent checksum/rollback proof. Await durable worker report and independent verification before any DONE claim.

## Task36 — security negative matrix: QUEUED

Task36 remains independent, but the only active worker host has 2 CPUs / ~2 GB RAM and the Task16 agent is consuming substantial CPU. Do not start another heavy agent until capacity frees or another host becomes active. This is deliberate resource safety, not a roadmap dependency.

## Current execution rules

- latest GitHub `main` is the source of truth;
- no force push/reset of main;
- no secrets in Git/chat/CI/evidence;
- no fake usage/online/IP/HWID/speed;
- read-only account/subscription/QR actions never rotate credentials/tokens;
- no Runtime mutation without validate → backup → apply → verify → rollback;
- no Production mutation without fresh backup and rollback state;
- no feature becomes DONE without fresh verification evidence;
- accounting/session semantics remain exact under retry, race, kill and disconnect.

## Exact next sequence

1. finish Task13 source recovery/reconstruction on current main and repeat all exact-kill gates;
2. publish Task15 exact candidate and require PostgreSQL18 + normal CI gates;
3. merge neither Task13 nor Task15 unless exact-head gates are green;
4. reconcile Task16 only after stable schema20 and independent schema21 retention/RLS/finalization/rollback proof;
5. start Task36 when worker capacity allows;
6. restore `pv-primary`, run read-only audit, then guarded backup/deploy/postflight for eligible exact-main changes;
7. update roadmap/task accounting only from verified merged/deployed truth.
