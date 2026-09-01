# CONTINUE HERE — PVNaive

Last updated: 2026-09-01

If a Chat/Agent session is interrupted, start here. Historical S04/S05/S06 notes are evidence, not the current execution source of truth.

## First read

1. `OWNER_REQUIREMENTS.md`
2. `ROADMAP.md`
3. `AGENT_TASKS.md`
4. `KNOWN_ISSUES.md`
5. `HANDOFF.md`
6. newest `ops/evidence/*`
7. latest GitHub `main`, open PRs and exact-head CI before touching code or Production

## Current verified state

- GitHub `main`: `4098e2d22a2e802d277e424a968f685f9f20e6ac`.
- Exact-main CI run `33445151447`: **SUCCESS**.
- The only unrelated open PR found in the fresh audit is old draft PR #4; do not merge it as part of the current roadmap.
- Latest independently recorded Production deployment remains schema **19**. A fresh read-only Production audit could not run in the current session because SentinelX reports three connected hosts on a plan allowing one active host; `pv-primary` is connected but returns `upgrade_required` for command execution.
- Task12 active-session management: **DONE / Production**, schema17.
- Task14 concurrent-session limit: **DONE / Production**, schema19; see `ops/evidence/TASK14-20260831-concurrent-session-limit-production-pass.md`.
- Security Task35: **DONE in main** — BUG-001 refresh reuse-family, BUG-002 commit-before-success and BUG-003 DB/schema-backed readiness are closed with exact-main CI green.
- The BUG-002 merge is still not claimed freshly deployed because `pv-primary` has not been available for a new guarded audit/deploy proof.

## Work in progress

### Task13 — exact session kill

The latest-main reconciliation candidate is commit `922a5e0e155746906f28fe46ca89a24f269acfa7` in the persistent worker workspace. It was previously verified with the full Go suite, targeted race/session-control/rehearsal tests, Web tests/build, pinned-forwardproxy and reproducible-Caddy gates while preserving BUG-002 and schema19 semantics.

Publication is being reconstructed through GitHub's Git object API because the worker has read/fetch access but no HTTPS push credential. Publication branch: `lead/task13-kill-session-publish-20260901`. Only hash-identical blobs are acceptable. If an uploaded blob SHA differs from the worker's `git hash-object`, discard it and do not attach it to the branch. Do not open/merge the Task13 PR until all 18 candidate paths are present and the branch CI/accounting/pinned-forwardproxy gates are green.

Task13 invariants remain:

- exact one-session identity only;
- no whole-credential revoke;
- no Caddy restart/reload;
- forced disconnect still produces the normal final accounting close/settlement exactly once;
- idempotent repeated kill;
- tenant/role isolation and redacted audit;
- HTTP/1.1 and HTTP/2 behavior remain proven;
- forwardproxy reproducibility + Go/Web/rehearsal gates must remain green.

### Task15 — simultaneous unique-IP limit

A worker is actively rebuilding schema20 from the trusted peer boundary. The old schema20 candidate remains rejected because it counted the wrong source and did not wire trusted Caddy `RemoteAddr` into enforcement.

Required boundary: trusted Caddy `RemoteAddr` before payload forwarding + `direct_naive_accounting_session_peers`/authoritative live-session state + PostgreSQL race-safe admission/reservation. Never use `Forwarded`, `X-Forwarded-For`, or client-supplied identity. Require a PostgreSQL18 concurrency proof before integration. No final worker report has been accepted yet.

### Task16 — bounded IP/session history

A second worker is active on schema21. Keep integration ordered behind stable Task15/schema20. History must be privacy-aware, tenant-scoped, explicitly bounded to the agreed retention window, synchronized with exact final accounting, and derived only from trusted session/peer facts. Purge authority must remain maintenance-only rather than app-accessible. No final worker report has been accepted yet.

### Parallel independent lane

Task36 authorization/IDOR/CSRF/redaction/fuzz work remains an independent next lane. Do not overload the only currently active worker host while Task15 and Task16 are consuming it; start Task36 when capacity is genuinely available or another host becomes active.

## Production access blocker

SentinelX currently reports three connected hosts on a plan allowing one active host. The worker host is active; `pv-primary` is connected/healthy at the control-plane level but command execution returns `upgrade_required`. Therefore do not claim a fresh Production audit or deploy until `pv-primary` becomes active again. GitHub/CI/code-review/documentation work continues independently.

When Production access is restored, first perform a **read-only** audit. Only after exact-main gates are green use the guarded flow: fresh encrypted backup + rollback snapshot → apply only the intended release/migration → readiness/Runtime/Telemetry/Caddy/customer/accounting smoke → verify exact deployed provenance. Roll back on any failed invariant. Never infer Production state from GitHub alone.

## Safety invariants

- start from latest `main`;
- no force-push/reset of main;
- no secret values in Git/chat/CI/evidence;
- no fake usage/online/IP/HWID/speed;
- read-only subscription/QR/account views never rotate credentials or tokens;
- Runtime mutations preserve validate/backup/apply/verify/rollback;
- no Production mutation without fresh backup and rollback state;
- no task becomes DONE without fresh verification evidence;
- accounting/session identity must remain truthful and exact under retries, kills, races and disconnects.

## Immediate next sequence

1. finish hash-identical publication of all 18 Task13 candidate paths on `lead/task13-kill-session-publish-20260901`;
2. run PR CI + exact-accounting + pinned-forwardproxy gates and merge only if all are green;
3. reconcile Task15 worker output only after trusted-IP and PostgreSQL18 race proofs pass;
4. reconcile Task16 only after schema20 is stable and its schema21/retention/RLS/finalization gates pass;
5. start Task36 on available capacity without starving the current schema workers;
6. restore `pv-primary` execution access, perform read-only audit, then deploy guarded releases only with backup/rollback and fresh postflight evidence;
7. update `PROJECT_STATUS.md`, `HANDOFF.md`, roadmap accounting and task states from merged/deployed evidence only.
