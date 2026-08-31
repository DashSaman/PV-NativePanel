# PVNaive — Canonical Handoff

Last updated: 2026-08-31

Resume from this file plus latest GitHub/Production evidence. Older S04/S05/S06/task checkpoints are historical and must not override current `main`.

## Repository / release truth

- Repository: `DashSaman/PV-NativePanel`
- Product: **PVNaive**
- Default branch: `main`
- Current GitHub main: `fce39283c6449b0d1836757ee7caddb31fab9def`
- Main CI run `33426149726`: **SUCCESS**
- Latest independently recorded Production deployment: `0645b2e4758d3cc6c197dd9dba9127e8de983d6c`
- Latest independently recorded Production schema: **19**
- Task14 Production evidence: `ops/evidence/TASK14-20260831-concurrent-session-limit-production-pass.md`

Do not claim the BUG-002 main commit is deployed until a fresh Production audit + guarded same-schema deploy/postflight runs.

## Durable integrated core — do not rewrite

- safe Naive Runtime credential lifecycle with narrow privileged Runtime Agent;
- Caddy validate/backup/apply/verify/rollback discipline;
- exact direct-Naive accounting and Telemetry Agent/Unix-socket boundary;
- restart-safe boot/session/sequence/cumulative accounting semantics;
- ServiceTerm-isolated usage and finite-quota reservation/settlement core;
- trusted first-successful-CONNECT identity path;
- customer/product/plan/group/tag/renewal foundations;
- `/sub` machine delivery, `/s` human account page and local QR;
- explicit Subscription reissue separated from password rotation;
- observability, request IDs/redaction, system dashboard, Doctor/support-bundle foundations;
- encrypted backup + restore-drill + guarded release/rollback foundations;
- Task12 trusted active-session projection at schema17;
- Task14 concurrent-session limit at schema19.

## Verified recent closures

### Task12 — active customer sessions: DONE / Production

Trusted Caddy `RemoteAddr` peer identity, active timestamps and exact per-session bytes are deployed at schema17. Legacy sessions without trusted peer evidence are not fabricated.

### Task14 — concurrent session limit: DONE / Production

Schema19 `Unlimited/N` admission is PostgreSQL-authoritative and race/reconnect tested. Exact rollout evidence records fresh encrypted backup, schema18→19 migration, guarded R1 deploy, healthy postflight and unchanged Caddy binary/config/PID/restart count.

### Task35 — BUG-001/002/003: DONE in main

- BUG-001 refresh-token reuse-family handling: closed with schema18/auth regression proof.
- BUG-002 generic response-before-commit: PR #47 merged as `fce39283...`; authenticated responses are buffered until transaction commit and commit failure cannot leak success. Exact-main CI run `33426149726` passed.
- BUG-003 DB/schema-backed readiness: closed with bounded fail-closed readiness behavior.

## Current active lanes

### Task13 — exact one-session disconnect: IN_PROGRESS

A substantial candidate exists in the worker workspace from an older base. It must be rebased on current main because it touches HTTP server code changed by BUG-002.

Required proof before promotion:

- disconnect exactly one session identity;
- no whole-credential revoke;
- no Caddy reload/restart;
- forced disconnect still emits/settles final exact accounting;
- idempotent repeated kill;
- correct tenant/role boundary and redacted audit;
- pinned forwardproxy reproducibility, Go/Web/rehearsal gates;
- controlled Production canary before DONE.

### Task15 — simultaneous unique-IP limit: BLOCKED ON DESIGN

A schema20 candidate was rejected before publication. It attempted to count `direct_naive_accounting_sessions.client_ip`, but Task12 authoritative trusted IP lives in `direct_naive_accounting_session_peers`; its proposed extra ingest IP argument also was not wired from the pinned forwardproxy/Telemetry path.

Redesign rules:

- identity source is Caddy `RemoteAddr`, never client headers;
- admission must occur before payload forwarding and must not fabricate/retrofit peer identity;
- PostgreSQL must serialize competing opens so simultaneous different IPs cannot oversubscribe the limit;
- same-IP multi-session semantics and reconnect/stale-session rules must be explicit;
- first-CONNECT activation/accounting semantics must not be corrupted by a rejected IP admission.

### Task16 — bounded IP/session history: TODO

Implement only from trusted session/peer facts with explicit privacy-aware retention, tenant isolation and purge semantics. Never invent legacy IP history. Reconcile migration numbering after Task15 stabilizes.

### Task36 — authorization/IDOR/CSRF/redaction/fuzz: independent next lane

This can progress without Production mutation. Build the negative ready-route × Owner/Admin/Reseller/Operator/Auditor matrix and cross-tenant mutation/read tests while Task13/15 are being resolved.

## Infrastructure blocker

SentinelX currently sees five connected hosts but the active plan permits one active host. `pv-worker-main` is accessible; `pv-primary` currently returns `upgrade_required`. Therefore this handoff does **not** assert a fresh live Production check after the Task14 evidence and does not assert deployment of `fce39283...`.

As soon as `pv-primary` is active again:

1. perform read-only service/readiness/schema/deployed-marker/disk audit;
2. confirm exact main CI remains green;
3. take fresh encrypted DB/config/release rollback snapshots;
4. build/release exact current main;
5. run guarded same-schema deployment for BUG-002 code;
6. verify API/DB/Runtime/Telemetry/Caddy/customer/subscription/accounting invariants and provenance;
7. roll back immediately if any invariant fails.

## Current execution rule

Do not leave independent work waiting on the Production-access blocker. Continue Task13 rebase/review, Task15 redesign, Task16 design and Task36 negative security gates in isolated branches/worktrees. When multiple servers are accessible, distribute these lanes and keep each host near the Owner-requested maximum ~70% CPU/RAM, reducing our own concurrency if resident services push total load above that ceiling.

## Safety invariants

- latest `main` is source of truth;
- no force push/reset of main;
- no secret/password/token/key in Git/chat/CI/evidence;
- no fake usage/online/IP/HWID/speed;
- read-only account/subscription/QR actions never rotate credentials/tokens;
- no Runtime mutation without validate → backup → apply → verify → rollback;
- no Production mutation without fresh backup and rollback state;
- no feature is DONE without fresh verification evidence.

## Exact next sequence

1. merge the current canonical-doc reconciliation after its exact-head CI is green;
2. rebase and review Task13 on `fce39283...`;
3. write a new RED security/race contract for Task15 using the trusted peer boundary, then implement minimal GREEN design;
4. advance Task36 negative authorization gates in parallel;
5. restore `pv-primary` access and deploy/verify current main safely;
6. continue `ROADMAP.md` in priority/dependency order through clean install, capacity, final smoke and RC.
