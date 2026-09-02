# PVNaive — Canonical Project Status

Last updated: 2026-09-02

This file is current repository + Production truth. Historical stage notes, worker reports and stale branches are evidence only; exact GitHub `main`, exact-head CI and fresh Production observations override them.

## Product / safety invariants

PVNaive remains standalone-first. Customer/service state, Runtime credentials, subscription delivery, exact direct-Naive accounting/session telemetry and privileged Runtime mutation stay separate. Never fabricate usage/online/IP/session history; never rotate credentials/tokens from read-only flows. Production changes require a fresh encrypted backup + rollback state, intended migrations only, postflight verification and exact deployed provenance.

## Repository truth

- Repository: `DashSaman/PV-NativePanel`.
- Current `main`: `61a46f4cb124cb5d4a180e9c564f51634cc8bba3` (PR #74 canonical reconciliation).
- Exact-main push CI run `33585915354`: **SUCCESS**.
- Current roadmap PR: **draft PR #64**, branch `lead/task13-reconstruct-62573fee`, exact published head `664b690239f723dc534164eceddfcb1dc9b3374f`.
- Task13 has been reconciled onto exact current main without force-push/history rewrite. Fresh compare: **41 commits ahead / 0 behind**, merge base = exact current main `61a46f4c...`.
- Fresh CI / WS1 Exact Accounting / WS1 Pinned Forwardproxy runs have started for exact head `664b6902...`; old-head green workflows remain historical evidence only and must not be reused as merge authorization.
- Task13 exact live-session kill: **IN PROGRESS / draft PR #64 / final live proof pending**.
- Task12 active-session management: **DONE / Production**, schema17.
- Task14 concurrent-session limit: **DONE / Production**, schema19.
- Task15 simultaneous unique-IP limit: **DONE / Production**, schema20.
- Security Task35 / BUG-001/002/003: **DONE in main**.
- Task16 bounded IP/session history: **IN PROGRESS / schema21 design gate remains failed until retention and pagination are server-bounded**.

No task becomes DONE from a local candidate, historical worker report or partial branch alone.

## Production truth

Production remains on the guarded Task15/schema20 rollout; no Task13 code has been deployed.

Fresh read-only observation on `pv-primary` during this orchestration run:

- `pvnaive-api`: **active**;
- `caddy-naive`: **active**;
- `pvnaive-runtime-agent`: **active**;
- `pvnaive-telemetry-agent`: **active**;
- `GET http://127.0.0.1:8080/api/v1/health/ready`: `{db:"ok", ready:true, schema:"ok", status:"ready"}`;
- `GET http://127.0.0.1:8080/api/v1/health/live`: `{service:"pvnaive-api", status:"ok"}`.

No fresh journal-health claim is made in this run. No Production mutation, deployment, migration, restart, reload, DB write or credential change was performed.

## Task13 — exact live-session kill

Draft PR #64 contains the validated Task13 implementation on a branch whose merge base is exact current main. Validated scope includes exact tuple registry/client primitives, live CONNECT registration after accounting-open/trusted-peer success, reload-safe Caddy-owned Unix listener, dedicated `pvnaive-session-control` socket permissions, ownership-checked DELETE API, per-session Web/UI kill without credential mutation, R1 patched-Caddy packaging/rollback, and PostgreSQL18 auth/tenant proof.

Two gates remain before merge:

1. **Fresh exact-head gates:** CI + WS1 Exact Accounting + WS1 Pinned Forwardproxy must pass on exact head `664b6902...`.
2. **Final live protocol/accounting proof:** fresh real HTTP/1.1 + HTTP/2 rehearsal must prove target-only termination, sibling survival, forged-tuple rejection, repeated-kill idempotency, credential survival, no Caddy lifecycle action caused by kill, and exactly-once final accounting.

Only after both gates pass may Task13 be merged. Production deployment additionally requires a fresh encrypted backup + rollback snapshot and postflight verification.

## Task16 — bounded IP/session history / schema21

No fresh current-main Task16 implementation is credited. Required proof remains server-enforced exact 30-day retention and hard-bounded pagination/read limits, tenant RLS/authorization, trusted peer/accounting facts only, final-accounting synchronization, maintenance-only purge, coherent schema21 migration/checksums, PostgreSQL18 behavior and rollback safety. RED tests for >30-day retention and oversized pagination must precede implementation.

## Persistent worker reports

`AGENT_HANDOFF.md` and `ops/DEPLOYMENT_PROGRESS.md` remain historical S04-era ledgers (last updated 2026-08-27) and contain no newer Task13/Task16 completion. They are evidence only, not current project truth.

## Worker / orchestration capacity

Fresh SentinelX listing shows **three connected hosts**: `TrPaqet`, `pv-worker-main`, and `pv-primary`; the Free plan permits one active host.

- Active/executable now: `pv-primary`, Production-only.
- `TrPaqet`: connected/healthy but fresh execution returns `upgrade_required`; assignment = Task13 final HTTP1/HTTP2 rehearsal.
- `pv-worker-main`: connected/healthy but fresh execution returns `upgrade_required`; assignment = Task16 RED retention/pagination lane.
- Do not use `pv-primary` as a development lane.

Task13 final live rehearsal and Task16 development cannot execute until a development Worker becomes executable. Human action is required to disconnect/switch hosts or increase the SentinelX active-host limit.

## Immediate execution order

1. Keep PR #64 draft while exact-head gates for `664b6902...` run.
2. On the first executable development Worker, run the final real HTTP1+HTTP2 exact-kill rehearsal with exactly-once final accounting and no kill-triggered Caddy lifecycle action.
3. Require all exact-head gates green; merge Task13 only after both CI and final live proof are green.
4. Deploy only after fresh encrypted Production backup + rollback snapshot and postflight access are available.
5. Start Task16 on the next independent Worker with RED retention/pagination tests; never use Production for schema21 development.
