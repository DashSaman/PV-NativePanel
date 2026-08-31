# PVNaive — Canonical Project Status

Last updated: 2026-09-01

This file describes the current repository + Production truth. Historical S04/S05/S06 snapshots, old task-number summaries and stale PR branches are evidence only and must not override this file, `CONTINUE_HERE.md`, the current roadmap, exact GitHub `main`, or fresh Production evidence.

## Product / architecture invariants

PVNaive is PVNETWORK's standalone-first management plane for standard NaiveProxy. Keep these boundaries separate:

1. commercial customer/service state (`User`, immutable `ServiceTerm`, plans/groups/tags),
2. Runtime Naive credentials/secrets,
3. `/sub/<opaque-token>` machine delivery and `/s/<opaque-token>` human account page,
4. exact direct-Naive accounting/session telemetry,
5. privileged Runtime mutation through the narrow local Runtime Agent.

Do not rotate Runtime credentials or Subscription tokens when merely editing quota/expiry or viewing Subscription/QR.

## Repository truth

- Repository: `DashSaman/PV-NativePanel`
- Product: **PVNaive**
- Default branch: `main`
- Freshly audited main SHA: `4098e2d22a2e802d277e424a968f685f9f20e6ac`
- Exact-main CI run `33445151447`: **SUCCESS**
- Old draft PR #4 is unrelated to the current roadmap and must not be merged as current work.
- Task12 active-session management: **DONE / Production**, schema17.
- Task14 concurrent-session limit: **DONE / Production**, schema19.
- Security Task35: **DONE in main** — BUG-001 refresh-token reuse-family handling, BUG-002 commit-before-success semantics, and BUG-003 DB/schema-backed readiness are closed in repository truth with exact-main CI green.
- Task13 exact live-session kill: **IN PROGRESS**, reconciled/tested candidate exists and is being published byte-for-byte to GitHub before PR gates.
- Task15 unique-IP limit: **IN PROGRESS**, schema20 redesign worker active; no accepted final report yet.
- Task16 bounded IP/session history: **IN PROGRESS**, schema21 worker active but integration remains ordered behind stable schema20; no accepted final report yet.

Do not infer a task as DONE merely because a local candidate or worker report exists. Merge, exact-head verification and — when Production semantics change — guarded deployment/postflight evidence are required.

## Production truth

The latest independently recorded Production schema remains **19**. Historical evidence records Task12/schema17 and Task14/schema19 Production rollouts. In the current coordinator audit, SentinelX reports `pv-primary` connected at the control-plane level, but command execution returns `upgrade_required` because three hosts are connected while the current plan permits one active host.

Therefore:

- no fresh Production health assertion is made for this session;
- no Production mutation/deployment/migration is permitted until `pv-primary` command execution is restored;
- repository/CI/review/documentation work continues independently;
- when access returns, begin with a read-only audit before any mutation;
- every mutation requires fresh backup + rollback state and exact postflight provenance.

The Production-access limitation is an operations/tooling blocker, not evidence that Production itself is unhealthy.

## Exact accounting / session invariants

The following are non-negotiable:

- usage, online state, peer IP and session identity must come from authoritative direct-Naive facts; never fabricate legacy history or device state;
- Runtime credential identity, node ID, boot ID and opaque session ID form the exact live-session identity used by control paths;
- killing one session must never revoke the whole credential or kill an unrelated sibling session;
- a forced disconnect must converge through the normal exact final-accounting close/settlement path exactly once;
- retries and idempotent operations must not double-count bytes or duplicate finalization;
- concurrent-session/unique-IP admission must be race-safe in PostgreSQL, not merely process-local;
- trusted peer identity comes from Caddy's actual `RemoteAddr`; never trust `Forwarded`, `X-Forwarded-For`, or client-supplied IP values for enforcement;
- schema changes require forward/backward migration proof and coherent expected-schema/checksum manifests;
- no Caddy reload/restart is acceptable merely to kill one live session.

## Task13 — exact live-session kill

Persistent worker candidate commit: `922a5e0e155746906f28fe46ca89a24f269acfa7`.

Previously verified candidate gates include full Go, targeted race/session-control/rehearsal tests, Web tests/build, pinned-forwardproxy and reproducible-Caddy checks while preserving BUG-002 and schema19 behavior. Publication branch: `lead/task13-kill-session-publish-20260901`.

Because the worker has no HTTPS push credential, publication is being reconstructed with GitHub Git objects. Every published file must match the worker's `git hash-object` exactly. Any mismatched upload is discarded and must never be attached to the branch. Do not open or merge a Task13 PR until the complete 18-path candidate is present and PR CI + Exact Accounting + Pinned Forwardproxy are green.

## Task15 — unique-IP limit / schema20

The previously attempted schema20 design is rejected. The accepted design must:

- source peer IP only from trusted Caddy `RemoteAddr`;
- use `direct_naive_accounting_session_peers` / authoritative live-session state rather than an invented or stale client-IP field;
- perform fail-closed admission before payload forwarding;
- serialize competing admissions with a PostgreSQL race-safe boundary keyed to the proper customer/service identity;
- allow legitimate reconnect from an already-counted IP without consuming a second unique slot;
- release closed/stale sessions correctly;
- include PostgreSQL18 concurrency proof plus negative spoofing tests;
- preserve exact accounting semantics under rejection, retry and disconnect.

No schema20 candidate is publishable until those gates pass.

## Task16 — bounded session/IP history / schema21

Integration is ordered behind stable Task15/schema20. Required invariants:

- explicit bounded retention (30-day target from the current worker contract);
- tenant-scoped owner/reseller reads with RLS/authorization proof;
- history derived only from trusted session/peer/accounting facts;
- no invented legacy peer history;
- exact final-accounting synchronization rather than a second competing source of truth;
- purge authority restricted to maintenance operations, not ordinary app/API access;
- safe schema21 rollback and coherent migration/checksum manifests.

## Security / independent lane

Task36 authorization/IDOR/CSRF/redaction/fuzz work remains an independent lane and should advance whenever worker capacity is genuinely available. Do not starve the active schema20/schema21 workers or overload the only currently active worker host.

## Release / Production safety

Before any Production mutation:

1. verify exact GitHub main and required CI gates;
2. perform a fresh read-only Production audit;
3. create a fresh encrypted backup and rollback snapshot;
4. record exact preflight release/schema/service state;
5. deploy only the intended commit and migrations;
6. verify readiness, Runtime Agent, Telemetry Agent, Caddy/customer/accounting invariants and exact deployed provenance;
7. roll back immediately on any failed invariant;
8. record evidence and only then update roadmap/task status.

Never force-push/reset main, print secrets, fake usage/online/IP/HWID/speed, or silently rotate credentials/tokens from read-only flows.

## Immediate execution order

1. complete byte-identical Task13 publication and run branch/PR gates;
2. merge Task13 only when all required gates are green;
3. reconcile Task15 only after trusted-IP + PostgreSQL18 concurrency proof;
4. reconcile Task16 only after schema20 is stable and schema21 retention/RLS/finalization proofs pass;
5. advance Task36 when worker capacity allows;
6. restore `pv-primary` execution access and perform a read-only audit;
7. deploy eligible exact-main changes only with fresh backup/rollback/postflight proof;
8. update `HANDOFF.md`, `ROADMAP.md`, `AGENT_TASKS.md`, `KNOWN_ISSUES.md` and evidence from verified merged/deployed truth only.

## Read next

1. `CONTINUE_HERE.md`
2. `OWNER_REQUIREMENTS.md`
3. `ROADMAP.md`
4. `AGENT_TASKS.md`
5. `KNOWN_ISSUES.md`
6. `HANDOFF.md`
7. newest `ops/evidence/*`
8. latest GitHub `main`, open PRs, exact-head CI and fresh Production state before any mutation.
