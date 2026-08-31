# CONTINUE HERE — PVNaive

Last updated: 2026-08-31

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

- GitHub `main`: `fce39283c6449b0d1836757ee7caddb31fab9def`.
- Main CI run `33426149726`: **SUCCESS**.
- Latest independently recorded Production deployment: `0645b2e4758d3cc6c197dd9dba9127e8de983d6c`, schema **19**.
- Task12 active-session management: **DONE / Production**, schema17.
- Task14 concurrent-session limit: **DONE / Production**, schema19; see `ops/evidence/TASK14-20260831-concurrent-session-limit-production-pass.md`.
- Security Task35: **DONE in main** — BUG-001 refresh reuse-family, BUG-002 commit-before-success and BUG-003 DB/schema-backed readiness are closed with exact-main CI green.
- The BUG-002 merge is not yet claimed deployed because a fresh `pv-primary` audit/deploy could not run during the current SentinelX host-access limitation.

## Work in progress

### Task13 — exact session kill

A substantial local candidate exists from an older pre-BUG-002 base. It touches the pinned forwardproxy/session-control path, HTTP API and UI. Do **not** blindly merge it. Rebase on current main and re-prove:

- exact one-session identity only;
- no whole-credential revoke;
- no Caddy restart/reload;
- forced disconnect still produces final accounting close/settlement;
- idempotent repeated kill;
- tenant/role isolation and redacted audit;
- forwardproxy reproducibility + Go/Web/rehearsal gates.

### Task15 — simultaneous unique-IP limit

The current schema20 candidate is **rejected / not publishable**. It attempted to count `direct_naive_accounting_sessions.client_ip`, while Task12 authoritative peer identity is stored in `direct_naive_accounting_session_peers`; its added ingest IP argument was also not wired from the pinned forwardproxy/Telemetry boundary.

Redesign around trusted Caddy `RemoteAddr` before payload forwarding with a PostgreSQL race-safe admission/reservation boundary. Never use `Forwarded`, `X-Forwarded-For` or a client-supplied identity.

### Task16 — bounded IP/session history

Keep pending until the Task15 identity/admission boundary is stable. History must be privacy-aware, tenant-scoped, bounded by explicit retention and derived only from trusted session/peer facts. Never invent legacy peer history.

### Parallel independent lane

Task36 authorization/IDOR/CSRF/redaction/fuzz work can progress without Production access and should be used to keep available workers productive.

## Production access blocker

SentinelX currently reports five connected hosts on a plan allowing one active host. `pv-worker-main` is accessible, while `pv-primary` returns `upgrade_required`. Therefore do not claim a fresh Production audit or deploy until `pv-primary` becomes active again. GitHub/CI/code-review/documentation work continues independently.

When Production access is restored, first perform a **read-only** audit, then if exact-main gates remain green execute the normal guarded flow: fresh encrypted backup + rollback snapshot → same-schema BUG-002 release deploy → readiness/Runtime/Telemetry/Caddy/customer/accounting smoke → verify exact deployed provenance. Roll back on any failed invariant.

## Safety invariants

- start from latest `main`;
- no force-push/reset of main;
- no secret values in Git/chat/CI/evidence;
- no fake usage/online/IP/HWID/speed;
- read-only subscription/QR/account views never rotate credentials or tokens;
- Runtime mutations preserve validate/backup/apply/verify/rollback;
- no Production mutation without fresh backup and rollback state;
- no task becomes DONE without fresh verification evidence.

## Immediate next sequence

1. finish and merge this current-truth docs reconciliation after exact-head CI;
2. rebase/review Task13 on current main;
3. redesign Task15 from the trusted peer boundary, beginning with a real failing race/security contract;
4. advance Task36 negative authorization gates in parallel;
5. restore `pv-primary` access and safely deploy/verify the BUG-002 main release;
6. continue remaining `ROADMAP.md` tasks without leaving an available worker idle.
