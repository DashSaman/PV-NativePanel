# PVNaive — Canonical Roadmap

Last updated: 2026-09-01

This is the current Production-readiness roadmap. Historical `PVN-*` task IDs and older stage documents remain audit history; they do not override this ledger.

Status vocabulary: `DONE`, `IN_PROGRESS`, `TODO`, `BLOCKED`, `PARTIAL`, `SUPERSEDED`.

## Current baseline

- Current GitHub `main`: `a29b5ef434a72004af80cf489f47fffe0b0a03a8`.
- Current roadmap implementation PR: draft #62, Task13 exact-session kill, exact published head `2e0f485d61f2dd70647b6f626b1f8a18178336d7`.
- Production deployment: Task15 source `26aa74dddfd23535e45837f21531cf67ea2fd238`, schema **20**.
- Fresh current-run Production read-only audit: all four core services active; readiness `ready=true`, `db=ok`, `schema=ok`; expected schema20; deploy source clean at Task15 commit; no checked panic/fatal/schema mismatch.
- Task12 active-session projection is deployed at schema17.
- Task14 concurrent-session limit is deployed at schema19.
- Task15 simultaneous unique-IP limit is deployed at schema20.
- Security Task35 (BUG-001 refresh reuse-family, BUG-002 commit-before-success, BUG-003 DB/schema readiness) is closed in main.

Already integrated and not to be rewritten: secure Runtime credential lifecycle, narrow Runtime Agent, customer/product/subscription foundations, exact direct-Naive accounting/telemetry, trusted first-CONNECT identity, hard-quota reservation/settlement core, Task12 active-session projection, Task14 concurrent-session enforcement, Task15 trusted-RemoteAddr unique-IP enforcement, observability/Doctor/backup/restore/release foundations.

## Production-ready master execution ledger

`DONE` requires evidence. `PARTIAL` is used where implementation/history exists but the exact current acceptance evidence has not been fully re-reconciled in this pass.

| # | Status | Priority | Task | Done gate / current truth |
|---:|---|---|---|---|
| 1 | IN_PROGRESS | P0 | Audit latest main / PR / CI / Production | fresh GitHub/CI and schema20 Production read-only audit current; repeat before any mutation |
| 2 | DONE | P0 | Competitor parity | current 120-feature parity matrix and licensing guard exist |
| 3 | IN_PROGRESS | P0 | Canonical project docs | current schema20/Task13/Task16/worker truth being reconciled in PR #63 |
| 4 | DONE | P0 | Reconcile useful old operations work | safe extraction/deployment of observability/Doctor/backup/restore/release foundations completed |
| 5 | PARTIAL | P0 | Legacy/adopted accounting baseline truth | merged implementation exists; canonical acceptance/Production evidence needs re-confirmation |
| 6 | PARTIAL | P0 | `/s` accounting/presence completion | merged implementation exists; re-confirm exact acceptance evidence |
| 7 | PARTIAL | P0 | Manual Reset Usage | merged implementation exists; re-confirm exact acceptance/evidence |
| 8 | PARTIAL | P0 | Bulk Reset Usage | merged implementation exists; re-confirm exact acceptance/evidence |
| 9 | PARTIAL | P0 | Periodic traffic reset | restart-safe implementation merged; re-confirm exact acceptance/evidence |
| 10 | TODO | P0 | Hard quota controlled Production proof | simultaneous race/exhaustion/reload/restart/reconnect/no negative/no bypass |
| 11 | TODO | P0 | First-successful-CONNECT controlled Production proof | reads/failed-auth/reload inert; successful authenticated CONNECT only activation |
| 12 | DONE | P0 | Session management | schema17 trusted peer IP + connected/last activity + exact session bytes deployed |
| 13 | IN_PROGRESS | P0 | Kill/disconnect session | draft PR #62; exact-tuple primitives + local control handler published; full data-plane/API/UI/final-accounting/live-protocol proof remains |
| 14 | DONE | P0 | Concurrent session limit | schema19 Unlimited/N enforcement deployed with PostgreSQL race/reconnect proof |
| 15 | DONE | P0 | Simultaneous unique-IP limit | schema20 trusted Caddy RemoteAddr + race-safe DB admission deployed and verified |
| 16 | IN_PROGRESS | P1 | IP/session history | schema21 design gate: exact 30-day retention and pagination must be server-bounded; RED tests required before implementation acceptance |
| 17 | TODO | P1 | HWID/device identity PoC | implement only if trustworthy standard Naive/Karing identity exists |
| 18 | TODO | P1 | Per-user speed-limit PoC | enforce in real data plane or do not expose |
| 19 | TODO | P1 | Reseller CRUD | create/edit/disable/revoke/list/search |
| 20 | TODO | P0 | Full tenant-isolation / IDOR audit | no cross-reseller read/edit/renew/delete/subscription access |
| 21 | TODO | P1 | Reseller wallet/credit | audited correct balance operations |
| 22 | TODO | P1 | Immutable financial ledger | credit/debit/create/renew/refund/adjustment entries |
| 23 | TODO | P1 | Reseller plan/user restrictions | allowed plans/max users/max active/credit/Owner oversight |
| 24 | TODO | P1 | Customer history | project create/renew/volume/expiry/plan/group/tag/suspend/resume/revoke/rotate/reissue/reset events |
| 25 | TODO | P1 | Audit Explorer | actor/user/action/date/IP/result filters; strict redaction |
| 26 | PARTIAL | P1 | Notification engine/preferences/history | retry/dedupe/redaction/transport foundation exists; persistence/preferences/event wiring/history remain |
| 27 | PARTIAL | P1 | Telegram + rule builder | secure transport foundation exists; configuration/rules/history/product workflow remain |
| 28 | PARTIAL | P1 | Dashboard / monitoring / historical charts | live monitoring exists; historical charts/online aggregates remain |
| 29 | PARTIAL | P1 | Logs / request diagnostics / support bundle | redacted logging/request IDs/support bundle exist; product log explorer remains |
| 30 | PARTIAL | P1 | Doctor command/page | CLI foundation deployed; complete product page/workflow remains |
| 31 | PARTIAL | P1 | Scheduled encrypted backup + retention | scheduled backup foundation deployed; retention product policy remains |
| 32 | PARTIAL | P1 | Restore / verification / drill / UI | automated restore drill foundation exists; full operator UI/workflow remains |
| 33 | PARTIAL | P1 | REST API + OpenAPI | ready-route OpenAPI exists; broader stabilization/version policy remains |
| 34 | PARTIAL | P1 | API rate limit / idempotency / webhooks | request/rate-limit foundation exists; stable mutation/webhook contracts remain |
| 35 | DONE | P0 | Fix auth/security BUG-001/002/003 | all three fixes merged and current main remains green |
| 36 | TODO | P0 | Authorization / IDOR / CSRF / redaction / fuzz | complete Route × role negative/quality matrix |
| 37 | PARTIAL | P1 | Supply-chain security + license policy | checksums/basic SBOM/provenance exist; SAST/dependency/secret scan/signing/NOTICE remain |
| 38 | PARTIAL | P1 | Multi-node model/auth/health/metrics/assignment | standalone-safe model/drift foundation only; controller/network operations remain |
| 39 | TODO | P1 | Drain/maintenance/canary/upgrade/failover/smart selection | reconciliation-safe fleet operations |
| 40 | TODO | P0 | Fresh secure Ubuntu installer | version-pinned PostgreSQL/Caddy/API/agents/systemd/firewall/TLS/migrations/web + Doctor |
| 41 | PARTIAL | P0 | Versioned upgrade | guarded same-schema release deployment exists; generic migration upgrade remains |
| 42 | PARTIAL | P0 | Rollback + conservative uninstall | release rollback exists; generic version rollback/uninstall/data policy remain |
| 43 | TODO | P0 | Client compatibility | Karing Windows/Android/iOS/macOS/Linux acceptance first |
| 44 | TODO | P0 | Load/capacity campaign | 50/100/200/400+ resource + accounting/quota/session correctness proof |
| 45 | TODO | P1 | Bulk/search completion | remaining bulk actions + advanced filters/sorts/columns/URL state |
| 46 | PARTIAL | P1 | Final UI polish | current product UI exists; accessibility/responsive/theme/final polish remain |
| 47 | IN_PROGRESS | P0 | Final documentation reconciliation | eliminate remaining historical-as-current contradictions with evidence |
| 48 | TODO | P0 | Final clean-server installation proof | fresh supported Ubuntu VM reaches fully healthy stack |
| 49 | TODO | P0 | Final Production smoke | backed-up exact RC deploy + customer/sub/accounting/runtime smoke + rollback ready |
| 50 | TODO | P0 | Release Candidate | no logical P0/P1 blocker, exact-head CI green, provenance/evidence recorded |

## Current immediate execution order

1. Finish Task13 on draft PR #62 without overwriting current Task14/15 or BUG-002 semantics: exact one-session close must drive the existing normal `accountingSession.close()` path and preserve sibling sessions.
2. Prove Task13 end-to-end on exact published tree: Go/race/Web/pinned-forwardproxy/reproducible Caddy + real HTTP/1.1 and HTTP/2 kill + exactly-once final accounting + unchanged Caddy PID/no reload.
3. Merge/deploy Task13 only after exact-head gates are green and fresh Production encrypted backup + rollback snapshot are ready.
4. As soon as a development worker is executable in parallel, start Task16 with RED tests proving requests cannot exceed 30-day retention or bounded page size; then implement minimal schema21 server-side enforcement, RLS, purge and PG18/rollback proof.
5. Continue independent Task36 route×role/IDOR/fuzz preparation when worker capacity permits because it does not require Production mutation.

## Production evidence

- Task14: `ops/evidence/TASK14-20260831-concurrent-session-limit-production-pass.md`.
- Task15: `ops/evidence/TASK15-20260901-schema20-production-pass.md`.

## Definition of Done

A feature is never `DONE` merely because code exists. Required where applicable: real backend/schema/auth/UI, no secret leaks, idempotency/failure/race tests, unit/integration/web tests, vet/build, exact-head GitHub CI, rollback for Runtime/Production changes, live verification for Production-facing capabilities, canonical docs/evidence and no regressions.
