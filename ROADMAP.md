# PVNaive — Canonical Roadmap

Last updated: 2026-08-31

This is the current Production-readiness roadmap. Historical `PVN-*` task IDs and older stage documents remain audit history; they do not override this ledger.

Status vocabulary: `DONE`, `IN_PROGRESS`, `TODO`, `BLOCKED`, `PARTIAL`, `SUPERSEDED`.

## Current baseline

- Current GitHub `main`: `fce39283c6449b0d1836757ee7caddb31fab9def`.
- Exact-main CI run `33426149726`: **SUCCESS**.
- Latest independently recorded Production deployment: `0645b2e4758d3cc6c197dd9dba9127e8de983d6c`.
- Production schema recorded by Task14 evidence: **19**.
- Task12 active-session projection is deployed at schema17.
- Task14 concurrent-session limit is deployed at schema19.
- Security Task35 (BUG-001 refresh reuse-family, BUG-002 commit-before-success, BUG-003 DB/schema readiness) is closed in main.
- The BUG-002 merge is not claimed deployed until a fresh Production audit and guarded same-schema deployment can run.

Already integrated and not to be rewritten: secure Runtime credential lifecycle, narrow Runtime Agent, customer/product/subscription foundations, exact direct-Naive accounting/telemetry, trusted first-CONNECT identity, hard-quota reservation/settlement core, Task12 active-session projection, Task14 concurrent-session enforcement, observability/Doctor/backup/restore/release foundations.

## Production-ready master execution ledger

`DONE` requires evidence. `PARTIAL` is used where implementation/history exists but the exact current acceptance evidence has not been fully re-reconciled in this pass.

| # | Status | Priority | Task | Done gate / current truth |
|---:|---|---|---|---|
| 1 | IN_PROGRESS | P0 | Audit latest main / PR / CI / Production | GitHub/CI current; fresh Production audit blocked by current SentinelX active-host limit |
| 2 | DONE | P0 | Competitor parity | current 120-feature parity matrix and licensing guard exist |
| 3 | IN_PROGRESS | P0 | Canonical project docs | current Task14/security/session-limit truth being reconciled |
| 4 | DONE | P0 | Reconcile useful old operations work | safe extraction/deployment of observability/Doctor/backup/restore/release foundations completed |
| 5 | PARTIAL | P0 | Legacy/adopted accounting baseline truth | merged implementation exists; canonical acceptance/Production evidence needs re-confirmation |
| 6 | PARTIAL | P0 | `/s` accounting/presence completion | merged implementation exists; re-confirm exact acceptance evidence |
| 7 | PARTIAL | P0 | Manual Reset Usage | merged implementation exists; re-confirm exact acceptance/evidence |
| 8 | PARTIAL | P0 | Bulk Reset Usage | merged implementation exists; re-confirm exact acceptance/evidence |
| 9 | PARTIAL | P0 | Periodic traffic reset | restart-safe implementation merged; re-confirm exact acceptance/evidence |
| 10 | TODO | P0 | Hard quota controlled Production proof | simultaneous race/exhaustion/reload/restart/reconnect/no negative/no bypass |
| 11 | TODO | P0 | First-successful-CONNECT controlled Production proof | reads/failed-auth/reload inert; successful authenticated CONNECT only activation |
| 12 | DONE | P0 | Session management | schema17 trusted peer IP + connected/last activity + exact session bytes deployed |
| 13 | IN_PROGRESS | P0 | Kill/disconnect session | exact one-session disconnect + confirmation/audit, no whole-credential revoke and no Caddy reload/restart |
| 14 | DONE | P0 | Concurrent session limit | schema19 Unlimited/N enforcement deployed with PostgreSQL race/reconnect proof |
| 15 | BLOCKED | P0 | Simultaneous unique-IP limit | rejected schema20 candidate used wrong IP source/unwired parameter; redesign around trusted Caddy RemoteAddr + race-safe DB admission |
| 16 | TODO | P1 | IP/session history | privacy-aware bounded retention using trusted session/peer facts |
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
| 35 | DONE | P0 | Fix auth/security BUG-001/002/003 | all three fixes merged; exact-main CI through BUG-002 merge green |
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

1. Rebase and independently review Task13 exact-session disconnect on `fce39283...`; prove close/final accounting survives the forced disconnect.
2. Redesign Task15 unique-IP admission using authoritative Task12 peer identity; reject any client-header or nonexistent-session-column shortcut.
3. Reconcile Task16 bounded session/IP history after the Task15 identity boundary is stable.
4. In parallel, start Task36 route×role/IDOR/fuzz negative gates because it does not require Production mutation.
5. As soon as `pv-primary` access is restored, run fresh read-only Production audit, exact backup/rollback preflight, deploy the same-schema BUG-002 main safely, and verify postflight before changing deployed provenance.

## Task14 Production evidence

See `ops/evidence/TASK14-20260831-concurrent-session-limit-production-pass.md` for the exact merge/deploy commit, release artifact SHA256, encrypted backup, schema18→19 migration, guarded R1 deploy and postflight/Caddy invariants.

## Definition of Done

A feature is never `DONE` merely because code exists. Required where applicable: real backend/schema/auth/UI, no secret leaks, idempotency/failure/race tests, unit/integration/web tests, vet/build, exact-head GitHub CI, rollback for Runtime/Production changes, live verification for Production-facing capabilities, canonical docs/evidence and no regressions.
