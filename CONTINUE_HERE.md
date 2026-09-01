# CONTINUE HERE — PVNaive

Last updated: 2026-09-01

If a Chat/Agent session is interrupted, start here. Historical S04/S05/S06 notes are evidence, not current execution truth.

## First read

1. `OWNER_REQUIREMENTS.md`
2. `ROADMAP.md`
3. `AGENT_TASKS.md`
4. `KNOWN_ISSUES.md`
5. `HANDOFF.md`
6. newest `ops/evidence/*`
7. latest GitHub `main`, open PRs and exact-head CI before touching code or Production

## Current verified repository state

- GitHub `main`: `fe6a9fea76fa48577fb8063bb246563f2696846b` (Merge PR #50).
- Exact-main CI run `33453736537`: **SUCCESS**.
- The only open PR found in the fresh audit is old draft PR #4; it is unrelated to the current roadmap and must not be merged as current work.
- Latest independently recorded Production schema remains **19**.
- Task12 active-session management: **DONE / Production**, schema17.
- Task14 concurrent-session limit: **DONE / Production**, schema19.
- Security Task35: **DONE in main** — BUG-001 refresh reuse-family, BUG-002 commit-before-success and BUG-003 DB/schema-backed readiness are closed in repository truth.

## Task13 — exact session kill: recovery / publication incomplete

The prior canonical docs named local candidate commit `922a5e0e155746906f28fe46ca89a24f269acfa7`. Fresh inspection of the currently active persistent worker shows that commit object is no longer present there, so do **not** treat that SHA as a recoverable local source of truth.

The remote publication branch `lead/task13-kill-session-publish-20260901` exists but is incomplete and diverged from current main. Relative to current main it currently carries only four Task13 files: `internal/sessioncontrol/client.go`, `internal/sessioncontrol/client_test.go`, `internal/sessioncontrol/protocol.go`, and `internal/sessionkill/registry.go`.

A persistent recovery bundle was found at `/tmp/task13-reference` containing eight exact previously reviewed Task13 files. Four recovered blobs match the already-published branch objects, including `internal/sessioncontrol/client.go` blob `1f7111ad538e5d0b12c39e8c76e99d090f4e2557`. A recovery worktree from current main is active. The recovered sessionkill/sessioncontrol packages compile and test, while HTTP API compilation correctly exposes missing integration wiring (`ServerConfig.SessionKiller` and the real delete-session route). Reconstruct the remaining integration from evidence/tests; do not invent a claim that the old 18-file local candidate still exists.

Prior evidence in `ops/evidence/TASK13-20260901-prepublication-verification.md` remains evidence that a full reconciled candidate previously passed Go/race/Web/pinned-forwardproxy/reproducible-Caddy and real HTTP/1.1+HTTP/2 exact-kill rehearsal. It is **not** permission to merge the current partial branch. Task13 remains **IN PROGRESS / NOT MERGED / NOT DEPLOYED** until the complete source tree is recovered/reconstructed, reverified on current main, published, and all exact-head gates are green.

Task13 invariants remain: exact full runtime-credential/node/boot/session tuple; no whole-credential revoke; no Caddy reload/restart; exactly-once normal final accounting after forced disconnect; idempotent repeated kill; tenant/role isolation; redacted audit; HTTP/1.1 and HTTP/2 support; pinned forwardproxy/reproducible Caddy gates.

## Task15 — simultaneous unique-IP limit / schema20

A fresh candidate now exists locally on branch `lead/task15-unique-ip-schema20-20260901`, commit `2b175991f1d5628dc084f4ffddfea6b63d960bf8`, whose parent is exactly current main `fe6a9fea...`.

Fresh local gates on this exact-main candidate are green: `git diff --check`, `go vet ./...`, full `go test ./...`, focused Go race tests, Web 18 files / 61 tests, Web production build, and shell syntax. Review also fixed a bad race-test assumption: PostgreSQL lock acquisition order is not deterministic, so the test now proves the persisted winner equals the caller that actually returned accepted instead of assuming the lower session ID wins.

The accepted schema20 boundary uses only trusted Caddy `RemoteAddr`, propagates it through exact accounting, locks the ServiceTerm row for race-safe admission, counts canonical active `direct_naive_accounting_session_peers`, rejects over-limit before acceptance, and records a peer only when schema19 actually accepts a new session. Never use `Forwarded`, `X-Forwarded-For`, or client-supplied identity.

The worker has PostgreSQL 14.24 only, which cannot execute the modern repository migration baseline (`security_invoker` unsupported). Therefore PostgreSQL18 migration/concurrency proof is still mandatory in GitHub CI after publication. The candidate is **NOT merged/deployed**. See `ops/evidence/TASK15-20260901-schema20-candidate-verification.md`.

Publication blocker: worker HTTPS `git push` still has no GitHub credential. Use authenticated Git/GitHub Git-data publication, then require exact-head CI including `tests/db/unique_ip_limit_migration_test.sh` before merge.

## Task16 — bounded IP/session history / schema21

A real isolated worker has been restarted on branch `worker/task16-schema21-20260901` from Task15 candidate `2b175991...`, preserving contiguous schema20→schema21 ordering. Integration remains blocked behind stable/published schema20. Required invariants: exact 30-day bounded retention, tenant-scoped RLS, trusted session/peer/accounting facts only, final-accounting synchronization, no fabricated legacy history, maintenance-only purge authority, bounded pagination/read paths, and safe rollback. Do not mark Task16 DONE without its final worker report and independent verification.

## Parallel independent lane

Task36 authorization/IDOR/CSRF/redaction/fuzz remains queued. The active worker host has only 2 CPUs / ~2 GB RAM and Task16 is currently consuming substantial CPU, so do not start another heavy agent until capacity is genuinely available. Advance Task36 when this lane frees or another host becomes active.

## Production access blocker

SentinelX freshly reports three connected hosts while the current Free plan allows one active host. `pv-primary` is connected/healthy at the control-plane level but even a read-only command returns `upgrade_required`; the worker host is the active host. Therefore no fresh Production-health assertion, backup, deployment, migration, Caddy mutation or schema change was performed in this run.

Human action required before the next Production audit/deploy: make `pv-primary` an active SentinelX host by disconnecting another connected host or changing the plan. When access returns, first run a **read-only** audit. Only then use the guarded flow: fresh encrypted backup + rollback snapshot → intended release/migration only → readiness/Runtime/Telemetry/Caddy/customer/accounting smoke → exact deployed provenance. Roll back on any failed invariant.

## Safety invariants

- start from latest `main`;
- no force-push/reset of main;
- no secret values in Git/chat/CI/evidence;
- no fake usage/online/IP/HWID/speed;
- read-only subscription/QR/account views never rotate credentials or tokens;
- Runtime mutations preserve validate/backup/apply/verify/rollback;
- no Production mutation without fresh backup and rollback state;
- no task becomes DONE without fresh verification evidence;
- accounting/session identity remains truthful and exact under retries, kills, races and disconnects.

## Immediate next sequence

1. finish Task13 source recovery/reconstruction on current main, then rerun full Go/race/Web/rehearsal/pinned-forwardproxy/reproducible-Caddy gates before publishing a complete branch;
2. publish Task15 exact candidate through authenticated Git-data transport and let PostgreSQL18 CI prove schema20 migration/concurrency semantics;
3. merge neither Task13 nor Task15 until their exact-head required gates are green;
4. reconcile Task16 only behind stable schema20 and independently verify its schema21 retention/RLS/finalization/rollback proof;
5. start Task36 only when worker capacity is available;
6. restore `pv-primary` execution access, run read-only Production audit, then deploy only eligible exact-main changes with fresh backup/rollback/postflight evidence;
7. update `PROJECT_STATUS.md`, `HANDOFF.md`, roadmap accounting and task states only from verified merged/deployed truth.
