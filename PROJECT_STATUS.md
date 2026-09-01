# PVNaive — Canonical Project Status

Last updated: 2026-09-02

This file is current repository + Production truth. Historical stage notes, worker reports and stale branches are evidence only; exact GitHub `main`, exact-head CI and fresh Production observations override them.

## Product / safety invariants

PVNaive remains standalone-first. Customer/service state, Runtime credentials, subscription delivery, exact direct-Naive accounting/session telemetry and privileged Runtime mutation stay separate. Never fabricate usage/online/IP/session history; never rotate credentials/tokens from read-only flows. Production changes require a fresh encrypted backup + rollback state, intended migrations only, postflight verification and exact deployed provenance.

## Repository truth

- Repository: `DashSaman/PV-NativePanel`.
- Current `main`: `898536d8e44585fc696895bc74831b52416fba93` (PR #71 documentation reconciliation).
- Exact-main push CI run `33560195019`: **SUCCESS**.
- Current roadmap PR: **draft PR #64**, branch `lead/task13-reconstruct-62573fee`, exact published head `5bc42d8dedd682eaf560a99777b21b9e82062c79`.
- Exact-head #64 workflows remain green: CI run `33557107036` **SUCCESS**, WS1 Exact Accounting run `33557107038` **SUCCESS**, WS1 Pinned Forwardproxy run `33557107045` **SUCCESS**.
- PR #64 is now **diverged from current main**: compared with `898536d8...`, it is 38 commits ahead and 25 commits behind, with merge base `62573fee8b88e4f951224da10e6a26d5b5838a54`; GitHub currently reports `mergeable=false`. Do not merge or force-update it until the Task13 tree is reconciled onto current main and revalidated.
- Task13 exact live-session kill: **IN PROGRESS / draft PR #64**. Validated scope includes exact tuple registry/client primitives, live CONNECT registration after accounting-open/trusted-peer success, reload-safe Caddy-owned Unix listener, dedicated `pvnaive-session-control` socket permissions, ownership-checked DELETE API, per-session Web/UI kill, R1 packaging/install/rollback support, and PostgreSQL18 DB/auth proof.
- Task12 active-session management: **DONE / Production**, schema17.
- Task14 concurrent-session limit: **DONE / Production**, schema19.
- Task15 simultaneous unique-IP limit: **DONE / Production**, schema20.
- Security Task35 / BUG-001/002/003: **DONE in main**.
- Task16 bounded IP/session history: **IN PROGRESS / schema21 design gate remains failed until retention and pagination are server-bounded**.

No task becomes DONE from a local candidate, historical worker report or partial branch alone.

## Production truth

Production remains on the guarded Task15/schema20 rollout; no Task13 code has been deployed.

Fresh read-only observation on `pv-primary` at this checkpoint:

- `pvnaive-api`: **active**;
- `caddy-naive`: **active**;
- `pvnaive-runtime-agent`: **active**;
- `pvnaive-telemetry-agent`: **active**;
- `GET http://127.0.0.1:8080/api/v1/health/ready`: `{db:"ok", ready:true, schema:"ok", status:"ready"}`;
- `GET http://127.0.0.1:8080/api/v1/health/live`: `{service:"pvnaive-api", status:"ok"}`;
- no `panic`, `fatal`, or `schema mismatch` matches were found in the inspected recent service journal window.

The API process environment could not be re-read because `/proc/<pid>/environ` is permission-denied for the SentinelX account, so no fresh process-env schema claim is made. No Production mutation, deployment, migration, restart, reload, DB write or credential change was performed.

## Task13 — exact live-session kill

Draft PR #64 contains the data-plane registry/control path, dedicated socket permission boundary, ownership-checked API, Web/UI kill action, and R1 packaging/install/rollback support for the patched reproducible Caddy binary and drop-in. Its current exact head remains green on CI, Exact Accounting and Pinned Forwardproxy.

Two gates now block merge:

1. **Branch reconciliation:** #64 is 25 commits behind current main and currently `mergeable=false`; rebuild/rebase/cherry-pick the validated Task13 delta onto current main without wholesale historical branch replacement, then rerun the exact-head gates.
2. **Final live protocol/accounting proof:** fresh real HTTP/1.1 + HTTP/2 rehearsal must prove target-only termination, sibling survival, forged-tuple rejection, repeated-kill idempotency, credential survival, no Caddy lifecycle action caused by a kill request, and exactly-once final accounting.

Only after both gates pass may the final R1 artifact be backed up, deployed and postflight-verified on Production.

## Task16 — bounded IP/session history / schema21

No fresh current-main Task16 implementation is credited. Required proof remains server-enforced exact 30-day retention and hard-bounded pagination/read limits, tenant RLS/authorization, trusted peer/accounting facts only, final-accounting synchronization, maintenance-only purge, coherent schema21 migration/checksums, PostgreSQL18 behavior and rollback safety. RED tests for >30-day retention and oversized pagination must precede implementation.

## Worker / orchestration capacity

Fresh SentinelX listing shows **three connected hosts**: `TrPaqet`, `pv-worker-main`, and `pv-primary`; the Free plan permits one active host.

- Active/executable now: `pv-primary`, Production-only.
- `TrPaqet`: connected/healthy but fresh execution returns `upgrade_required` under the one-active-host limit.
- `pv-worker-main`: connected/healthy but fresh execution returns `upgrade_required` under the same limit.
- Do not use `pv-primary` as a development lane.

True Task13/Task16 development cannot resume until a development Worker becomes executable. Human action is required to disconnect/switch hosts or increase the SentinelX active-host limit.

## Immediate execution order

1. Keep PR #64 draft and do not merge while it is diverged/mergeable=false.
2. On the first executable development Worker, reconstruct the validated Task13 delta from current `main` (no force-push of unreviewed history), run local focused/full tests, and publish a clean current-main head.
3. Add/run the final real HTTP1+HTTP2 exact-kill rehearsal with exactly-once final accounting and no kill-triggered Caddy lifecycle action; require all exact-head GitHub gates green.
4. Merge/deploy Task13 only after a fresh encrypted Production backup + rollback snapshot and postflight access are available.
5. Start Task16 on the next independent Worker with RED retention/pagination tests; never use Production for schema21 development.