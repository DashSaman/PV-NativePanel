# PVNaive — Canonical Project Status

Last updated: 2026-09-01

This file describes current repository + Production truth. Historical S04/S05/S06 snapshots and stale branches are evidence only and must not override this file, `CONTINUE_HERE.md`, exact GitHub `main`, or fresh Production evidence.

## Product / architecture invariants

PVNaive is PVNETWORK's standalone-first management plane for standard NaiveProxy. Keep commercial customer/service state, Runtime credentials/secrets, subscription/account delivery, exact direct-Naive accounting/session telemetry, and privileged Runtime mutation as separate boundaries. Read-only/edit flows must never silently rotate Runtime credentials or Subscription tokens.

## Repository truth

- Repository: `DashSaman/PV-NativePanel`
- Product: **PVNaive**
- Default branch: `main`
- Freshly audited main SHA: `fe6a9fea76fa48577fb8063bb246563f2696846b`
- Exact-main CI run `33453736537`: **SUCCESS**
- Only open PR found: old draft PR #4, unrelated to the current roadmap; do not merge it as current work.
- Task12 active-session management: **DONE / Production**, schema17.
- Task14 concurrent-session limit: **DONE / Production**, schema19.
- Security Task35: **DONE in main** — BUG-001 refresh-token reuse-family handling, BUG-002 commit-before-success semantics, and BUG-003 DB/schema-backed readiness are closed in repository truth.
- Task13 exact live-session kill: **IN PROGRESS / recovery + publication incomplete**.
- Task15 unique-IP limit: **IN PROGRESS / exact-current-main local schema20 candidate validated for Go/Web/static gates; PostgreSQL18 + publication still pending**.
- Task16 bounded IP/session history: **IN PROGRESS / isolated schema21 worker active behind schema20**.

Do not infer a task as DONE merely because a local candidate, worker report or historical test exists. Merge, exact-head verification and — when Production semantics change — guarded deployment/postflight evidence are required.

## Production truth

Latest independently recorded Production schema remains **19**. SentinelX freshly reports `pv-primary` connected/healthy at the control-plane level, but command execution returns `upgrade_required` because three hosts are connected while the plan permits one active host. Therefore no fresh Production-health assertion or mutation is made in this state. Repository/CI/review work continues independently.

When access returns: read-only audit first; then fresh encrypted backup + rollback snapshot; deploy only the intended exact commit/migrations; verify readiness, Runtime Agent, Telemetry Agent, Caddy/customer/accounting invariants and exact deployed provenance; roll back on any failed invariant.

## Exact accounting / session invariants

- Usage, online state, peer IP and session identity come only from authoritative direct-Naive facts; never fabricate history/device state.
- Exact live-session identity is Runtime credential ID + node ID + boot ID + opaque session ID.
- Killing one session must not revoke the credential or affect sibling sessions.
- Forced disconnect converges through normal final accounting exactly once.
- Retries/idempotency must not double-count or duplicate finalization.
- Concurrent-session and unique-IP admission must be PostgreSQL race-safe where policy is global to the service term.
- Trusted peer identity comes only from Caddy's actual `RemoteAddr`; never `Forwarded`, `X-Forwarded-For`, or client-supplied IP.
- Schema changes require forward/backward migration proof and coherent expected-schema/checksum manifests.
- No Caddy reload/restart merely to kill one live session.

## Task13 — exact live-session kill

The prior canonical state named persistent candidate commit `922a5e0e155746906f28fe46ca89a24f269acfa7`. Fresh inspection of the active persistent worker cannot resolve that commit object, so it is no longer a valid recoverable local source-of-truth claim.

Remote publication branch `lead/task13-kill-session-publish-20260901` exists but is incomplete and diverged from current main; relative to current main it currently contains only four Task13 files. A recovery bundle at `/tmp/task13-reference` contains eight exact previously reviewed files. Hash comparison confirms already-published objects such as `internal/sessioncontrol/client.go` (`1f7111ad538e5d0b12c39e8c76e99d090f4e2557`) are byte-identical to the recovered reference.

A current-main recovery worktree proves the recovered sessionkill/sessioncontrol packages compile/test. HTTP API compilation exposes the missing integration boundary rather than hiding it: `ServerConfig.SessionKiller` and the real delete-session route are absent on current main. Reconstruct remaining wiring/tests from preserved evidence and repository patterns, then rerun full verification. Do not merge the partial branch or claim the old 18-file local candidate still exists.

`ops/evidence/TASK13-20260901-prepublication-verification.md` remains historical proof that a complete reconciled candidate previously passed full Go, focused race/rehearsal, Web, pinned-forwardproxy, reproducible-Caddy and real HTTP/1.1/HTTP/2 exact-kill rehearsal. It does not substitute for publishing/retesting the recovered current-main source.

## Task15 — unique-IP limit / schema20

Local candidate branch: `lead/task15-unique-ip-schema20-20260901`.
Local candidate commit: `2b175991f1d5628dc084f4ffddfea6b63d960bf8`.
Parent: exact current main `fe6a9fea76fa48577fb8063bb246563f2696846b`.

Fresh local verification PASS:

- `git diff --check`
- `go vet ./...` with repository Go 1.25 toolchain
- `go test ./...`
- `go test -race ./internal/customer ./internal/httpapi ./internal/telemetry`
- Web: 18 test files / 61 tests
- Web production build
- shell syntax for the schema20 PostgreSQL gate

The candidate sources the peer only from trusted Caddy `RemoteAddr`, passes it through the exact accounting event, serializes admission by locking the ServiceTerm row, counts canonical active `direct_naive_accounting_session_peers`, and records peer state only when schema19 actually accepts a new session. Review fixed the PostgreSQL race test so it does not assume a deterministic lock winner; it proves the persisted winner is whichever concurrent caller actually returned accepted.

The active worker only has PostgreSQL 14.24 and no PostgreSQL18/container runtime. The modern migration baseline uses `security_invoker`, unsupported by PG14; therefore a local DB failure is environmental and **not** accepted as schema20 proof. The candidate's CI workflow includes `tests/db/unique_ip_limit_migration_test.sh`; PostgreSQL18 migration + concurrency proof is mandatory after publication.

Worker HTTPS `git push` still lacks credentials. Candidate is not repository truth, merged or deployed. See `ops/evidence/TASK15-20260901-schema20-candidate-verification.md`.

## Task16 — bounded session/IP history / schema21

An isolated worker is active on `worker/task16-schema21-20260901`, based on Task15 candidate `2b175991...` to keep schema20→schema21 contiguous. Integration remains ordered behind stable schema20.

Required invariants: exact 30-day retention; tenant-scoped RLS/authorization; trusted session/peer/accounting facts only; no invented legacy peer history; final-accounting synchronization; maintenance-only purge (not ordinary app/API authority); bounded pagination/read paths; safe rollback and coherent migration/checksum manifests. No DONE claim until a durable worker report and independent validation exist.

## Security / independent lane

Task36 authorization/IDOR/CSRF/redaction/fuzz remains independent, but the active host has only 2 CPUs / ~2 GB RAM and the Task16 agent currently consumes substantial CPU. Start Task36 when this worker frees or another host is active rather than violating the resource ceiling.

## Human-action blocker

To resume Production audit/deploy, make `pv-primary` an active SentinelX host: disconnect another connected host or change the plan. This is an operations/tooling blocker, not evidence that Production itself is unhealthy.

## Immediate execution order

1. recover/reconstruct the complete Task13 current-main source and re-run full exact-kill verification before publication/PR;
2. publish exact Task15 commit through authenticated Git/GitHub Git-data transport and require PostgreSQL18 CI + all normal gates;
3. merge neither Task13 nor Task15 unless exact-head gates are green;
4. reconcile Task16 only behind stable schema20 and independently prove schema21 retention/RLS/finalization/rollback;
5. advance Task36 when worker capacity allows;
6. restore `pv-primary`, perform read-only audit, then deploy eligible exact-main changes only with fresh backup/rollback/postflight proof;
7. update handoff/roadmap/task accounting only from merged/deployed evidence.
