# PVNaive — Agent / Workstream Task Board

Last updated: 2026-09-01

This is the active execution board. `OWNER_REQUIREMENTS.md` and `ROADMAP.md` define product intent and ordering; repository/CI/Production evidence decides status.

## Shared rules

- always start from latest `main`, never a chat snapshot;
- preserve working Runtime/accounting/customer/subscription behavior;
- use isolated branches/PRs and never force-push/reset `main`;
- TDD/fail-first for behavior changes;
- route/schema presence is not delivery evidence;
- never fabricate usage, online state, IP/HWID, speed or Production proof;
- never expose password/token/key/secret values in Git/chat/CI/evidence;
- Production mutation requires fresh encrypted backup, rollback state, exact-head CI and postflight;
- Production is never a development/test worker;
- GPL/AGPL competitor code remains reference-only unless license compatibility is explicitly approved.

## Current verified baseline

- GitHub `main`: `a29b5ef434a72004af80cf489f47fffe0b0a03a8`.
- Active roadmap implementation: draft PR #62 (`lead/task13-reconstruct-a29b5ef`), exact published head `2e0f485d61f2dd70647b6f626b1f8a18178336d7`.
- Production: Task15 source `26aa74dddfd23535e45837f21531cf67ea2fd238`, schema **20**.
- Fresh current-run read-only Production audit: four core services active; readiness `ready=true`, `db=ok`, `schema=ok`; expected schema20; deploy source clean at exact Task15 commit; no checked panic/fatal/schema mismatch.
- Task12 active-session management is deployed at schema17.
- Task14 concurrent-session limit is deployed at schema19.
- Task15 trusted-RemoteAddr unique-IP limit is deployed at schema20.
- BUG-001 refresh reuse-family, BUG-002 commit-before-success and BUG-003 DB/schema readiness are closed; Task35 is DONE.

## Current ordered task queue

Do not call a task DONE merely because code exists. `PARTIAL` means implementation/history exists but exact current acceptance/Production evidence still needs reconciliation.

| # | Workstream | Status | Current gate / next action |
|---:|---|---|---|
| 1 | Latest repo/PR/CI/Production audit | IN_PROGRESS | fresh current-run schema20 Production read-only audit complete; repeat before every mutation |
| 2 | Competitor parity | DONE | current parity matrix exists; re-check only on material product-surface change |
| 3 | Canonical docs truth | IN_PROGRESS | PR #63 reconciles main/schema20/Task13/Task16/worker truth |
| 4 | WS4 safe operations extraction | DONE | observability/Doctor/backup/restore/release foundations deployed previously |
| 5 | Legacy/adopted accounting baseline | PARTIAL | implementation merged historically; re-confirm exact Production evidence before canonical DONE |
| 6 | `/s` accounting/presence projection | PARTIAL | implementation merged historically; re-confirm acceptance evidence |
| 7 | Manual Reset Usage | PARTIAL | implementation merged historically; re-confirm acceptance/evidence |
| 8 | Bulk Reset Usage | PARTIAL | implementation merged historically; re-confirm acceptance/evidence |
| 9 | Periodic reset execution | PARTIAL | merged restart-safe implementation exists; re-confirm acceptance/evidence |
| 10 | Hard quota controlled Production proof | TODO | prove race/exhaustion/reload/restart/reconnect/no bypass on controlled canary |
| 11 | First-successful-CONNECT Production proof | TODO | prove reads/failed-auth/reload inert and successful CONNECT only activation path |
| 12 | Customer active sessions | DONE | trusted Caddy peer IP, active timestamps and exact per-session bytes deployed at schema17 |
| 13 | Exact kill/disconnect session | IN_PROGRESS | PR #62: exact-tuple primitives + local handler published; next forwardproxy/listener → API auth → UI → real HTTP1/2/final-accounting proof |
| 14 | Concurrent session limit | DONE | schema19 Unlimited/N enforcement deployed; race/reconnect proof recorded |
| 15 | Simultaneous unique-IP limit | DONE | schema20 trusted Caddy `RemoteAddr` + race-safe admission deployed and verified |
| 16 | Bounded IP/session history | IN_PROGRESS | schema21 blocked until RED tests prove exact 30-day retention and bounded pagination cannot be caller-bypassed |
| 17 | HWID/device identity PoC | TODO | implement only if trustworthy standard Naive/Karing identity exists |
| 18 | Per-user speed-limit PoC | TODO | expose only if real data-plane enforcement exists |
| 19 | Reseller CRUD | TODO | create/edit/disable/revoke/list/search |
| 20 | Full tenant isolation / IDOR audit | TODO | negative Route × role matrix and cross-reseller reads/mutations |
| 21 | Reseller wallet/credit | TODO | correct audited balance mutation semantics |
| 22 | Immutable financial ledger | TODO | credit/debit/create/renew/refund/adjustment history |
| 23 | Reseller plan/restriction quotas | TODO | allowed plans/max users/max active/credit |
| 24 | Customer history | TODO | project all sensitive customer/service actions |
| 25 | Audit Explorer | TODO | actor/user/action/date/IP/result filters with redaction |
| 26 | Notification engine/preferences/history | PARTIAL | transport/foundation exists; product persistence/preferences/event wiring/history remain |
| 27 | Telegram + rules | PARTIAL | secure transport foundation exists; configuration/rules/history/product UX remain |
| 28 | Dashboard/monitoring/history | PARTIAL | live monitoring exists; historical charts/online aggregates remain |
| 29 | Logs/request diagnostics/support bundle | PARTIAL | request IDs/redaction/support bundle exist; product log explorer remains |
| 30 | Doctor | PARTIAL | CLI deployed previously; product page/workflow remains |
| 31 | Scheduled encrypted backup + retention | PARTIAL | timer/backup foundation deployed; retention product policy remains |
| 32 | Restore/verification/drill/UI | PARTIAL | automated drill foundation deployed; operator UI/full workflow remains |
| 33 | API/OpenAPI | PARTIAL | ready-route OpenAPI exists; broader stabilization/version policy remains |
| 34 | API rate limit/idempotency/webhooks | PARTIAL | request/rate-limit foundation exists; stable mutation/webhook contracts remain |
| 35 | Security BUG-001/002/003 | DONE | all three fixes merged and current main remains green |
| 36 | Authorization/IDOR/CSRF/redaction/fuzz | TODO | complete full route × role quality gate |
| 37 | Supply-chain security | PARTIAL | checksums/basic SBOM/provenance foundation exists; SAST/dependency/secret scan/signing/NOTICE remain |
| 38 | Multi-node/fleet | PARTIAL | standalone-safe model/drift foundations exist; controller/network operations remain |
| 39 | Drain/maintenance/canary/failover/smart node | TODO | reconciliation-safe fleet operations |
| 40 | Fresh installer | TODO | secure version-pinned clean Ubuntu install + Doctor |
| 41 | Upgrade | PARTIAL | same-schema guarded release deploy exists; generic migration upgrade remains |
| 42 | Rollback/uninstall | PARTIAL | release rollback exists; generic version rollback/uninstall/data-retention policy remain |
| 43 | Client compatibility | TODO | Karing Windows/Android/iOS/macOS/Linux acceptance first |
| 44 | Load/capacity | TODO | 50/100/200/400+ correctness and resource campaign |
| 45 | Bulk/search completion | TODO | remaining bulk actions + advanced filters/sorts/columns/URL state |
| 46 | Final UI polish | PARTIAL | current product UI exists; accessibility/responsive/theme/final polish remain |
| 47 | Final docs reconciliation | IN_PROGRESS | remove remaining stale historical-as-current statements |
| 48 | Clean-server installation proof | TODO | fresh supported Ubuntu VM reaches healthy full stack |
| 49 | Final Production smoke | TODO | exact RC backed-up deploy + customer/sub/accounting/runtime smoke + rollback readiness |
| 50 | Release Candidate | TODO | no logical P0/P1 blocker, exact-head CI green, provenance/evidence complete |

## Parallel work allocation

Current SentinelX plan permits one active host. In this run `TrPaqet` was the development lane, then disconnected; the active slot moved to `pv-primary`. `pv-worker-main` and `ubuntu-4gb-hel1-1` currently return `upgrade_required`, so do not move development onto Production.

When development worker capacity is available, allocate in this order:

- **Lead / integration:** GitHub CI, canonical truth, reviews and safe release gating.
- **Worker A — Task13:** exact one-session disconnect on PR #62; preserve trusted `RemoteAddr`, schema20 unique-IP admission and BUG-002/final-accounting semantics.
- **Worker B — Task16:** RED tests for retention >30 days and oversized pagination, then minimal schema21 server enforcement/RLS/purge/PG18/rollback.
- **Worker C — Task36:** route×role/IDOR/CSRF/redaction/fuzz negative gates; independent of Production mutation.

If only one development worker is executable, Task13 owns it until its exact-head integration/rehearsal gate is complete; Task16 and Task36 remain prepared independent lanes, not falsely reported as running.

## Mandatory work-unit report

```text
TASK #:
STATUS:
WHAT CHANGED:
FILES:
TESTS:
CI:
PRODUCTION:
EVIDENCE:
REMAINING:
NEXT TASK:
```

Lead reconciles canonical documents only from evidence, then immediately advances the next independent task.
