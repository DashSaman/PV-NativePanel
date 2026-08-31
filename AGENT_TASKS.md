# PVNaive — Agent / Workstream Task Board

Last updated: 2026-08-31

This is the active execution board. `OWNER_REQUIREMENTS.md` and `ROADMAP.md` define product intent and ordering; repository/CI/Production evidence decides status.

## Shared rules

- always start from latest `main`, never a chat snapshot;
- preserve working Runtime/accounting/customer/subscription behavior;
- use isolated branches/PRs and never force-push/reset `main`;
- TDD/fail-first for behavior changes;
- route/schema presence is not delivery evidence;
- never fabricate usage, online state, IP/HWID, speed or Production proof;
- never expose password/token/key/secret values in Git/chat/CI/evidence;
- Production mutation requires fresh backup, rollback state, exact-head CI and postflight;
- GPL/AGPL competitor code remains reference-only unless license compatibility is explicitly approved.

## Current verified baseline

- GitHub `main`: `fce39283c6449b0d1836757ee7caddb31fab9def` — PR #47 BUG-002 merge.
- Exact-main CI run `33426149726`: **SUCCESS**.
- Latest independently recorded Production release: `0645b2e4758d3cc6c197dd9dba9127e8de983d6c`, schema **19**, from Task14 rollout evidence.
- Task12 active-session management is deployed at schema17.
- Task14 concurrent-session limit is deployed at schema19 with PostgreSQL race/reconnect proof.
- BUG-001 refresh reuse-family, BUG-002 commit-before-success and BUG-003 DB/schema readiness are closed; Task35 is DONE.
- A fresh live Production audit/deploy after the BUG-002 merge is currently blocked by SentinelX host-plan access to `pv-primary`; do not claim `fce39283...` is deployed until that gate is available and verified.

## Current ordered task queue

Do not call a task DONE merely because code exists. `PARTIAL` means implementation/history exists but the exact current acceptance/Production evidence still needs reconciliation.

| # | Workstream | Status | Current gate / next action |
|---:|---|---|---|
| 1 | Latest repo/PR/CI/Production audit | IN_PROGRESS | GitHub/CI reconciled; fresh `pv-primary` read-only audit blocked by current SentinelX active-host limit |
| 2 | Competitor parity | DONE | current 120-feature matrix exists; re-check only when product surface changes materially |
| 3 | Canonical docs truth | IN_PROGRESS | this reconciliation branch removes stale Task14/BUG-002/session-limit claims |
| 4 | WS4 safe operations extraction | DONE | observability/Doctor/backup/restore/release foundations deployed previously |
| 5 | Legacy/adopted accounting baseline | PARTIAL | implementation merged historically; re-confirm exact Production evidence before promoting canonical DONE |
| 6 | `/s` accounting/presence projection | PARTIAL | implementation merged historically; re-confirm acceptance evidence |
| 7 | Manual Reset Usage | PARTIAL | implementation merged historically; re-confirm acceptance/evidence |
| 8 | Bulk Reset Usage | PARTIAL | implementation merged historically; re-confirm acceptance/evidence |
| 9 | Periodic reset execution | PARTIAL | merged restart-safe implementation exists; re-confirm acceptance/evidence |
| 10 | Hard quota controlled Production proof | TODO | prove race/exhaustion/reload/restart/reconnect/no bypass on controlled Production canary |
| 11 | First-successful-CONNECT Production proof | TODO | prove reads/failed-auth/reload inert and successful CONNECT only activation path |
| 12 | Customer active sessions | DONE | trusted Caddy peer IP, active timestamps and exact per-session bytes deployed at schema17 |
| 13 | Exact kill/disconnect session | IN_PROGRESS | rebase/review existing candidate on latest main; must kill one live CONNECT without credential revoke or Caddy reload/restart |
| 14 | Concurrent session limit | DONE | schema19 Unlimited/N enforcement deployed; race/reconnect proof and Production evidence recorded |
| 15 | Simultaneous unique-IP limit | BLOCKED | current schema20 candidate rejected: it counted the wrong session IP source and did not wire trusted RemoteAddr into the enforcement boundary; redesign required |
| 16 | Bounded IP/session history | TODO | privacy-aware retention using trusted peer/session facts; reconcile migration numbering after Task15 design |
| 17 | HWID/device identity PoC | TODO | implement only if a trustworthy standard Naive/Karing identity exists |
| 18 | Per-user speed-limit PoC | TODO | expose only if a real data-plane enforcement boundary is proven |
| 19 | Reseller CRUD | TODO | create/edit/disable/revoke/list/search |
| 20 | Full tenant isolation / IDOR audit | TODO | negative Route × role matrix and cross-reseller mutations/reads |
| 21 | Reseller wallet/credit | TODO | correct audited balance mutation semantics |
| 22 | Immutable financial ledger | TODO | credit/debit/create/renew/refund/adjustment history |
| 23 | Reseller plan/restriction quotas | TODO | allowed plans/max users/max active/credit |
| 24 | Customer history | TODO | project all sensitive customer/service actions |
| 25 | Audit Explorer | TODO | actor/user/action/date/IP/result filters with redaction |
| 26 | Notification engine/preferences/history | PARTIAL | transport/foundation exists; product persistence/preferences/event wiring/history remain |
| 27 | Telegram + rules | PARTIAL | secure transport foundation exists; configuration/rules/history/product UX remain |
| 28 | Dashboard/monitoring/history | PARTIAL | live system monitoring exists; historical charts/online aggregates remain |
| 29 | Logs/request diagnostics/support bundle | PARTIAL | request IDs/redaction/support bundle exist; product log explorer remains |
| 30 | Doctor | PARTIAL | CLI deployed previously; product page/workflow remains |
| 31 | Scheduled encrypted backup + retention | PARTIAL | timer/backup foundation deployed; retention product policy remains |
| 32 | Restore/verification/drill/UI | PARTIAL | automated drill foundation deployed; operator UI/full workflow remains |
| 33 | API/OpenAPI | PARTIAL | ready-route OpenAPI exists; broader stabilization/version policy remains |
| 34 | API rate limit/idempotency/webhooks | PARTIAL | request/rate-limit foundation exists; stable mutation/webhook contracts remain |
| 35 | Security BUG-001/002/003 | DONE | all three fixes merged; exact-main CI through BUG-002 merge is green |
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

While only `pv-worker-main` is accessible through the current SentinelX host plan, keep independent lanes isolated in worktrees and use GitHub CI for full toolchain proof:

- Lead: integration, CI, canonical truth and safe release gating;
- Task13 lane: exact one-session disconnect, rebase existing candidate after BUG-002;
- Task15 lane: redesign unique-IP enforcement around trusted Caddy `RemoteAddr` and PostgreSQL race-safe admission;
- Task16 lane: bounded privacy-aware history after Task15 identity/enforcement boundary stabilizes;
- independent security/release lane: Task36 authorization/IDOR/fuzz preparation that does not depend on Production access.

When additional SentinelX hosts become active again, redistribute these lanes immediately and keep total CPU/RAM pressure around the Owner-requested 70% ceiling, backing off when other services raise host load.

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
