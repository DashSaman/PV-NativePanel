# PVNaive — Canonical Roadmap

Last updated: 2026-08-30

This roadmap follows the Owner's Production/parity master prompt. Historical `PVN-*` IDs remain audit references; they are not reused for unrelated work.

Status vocabulary: `DONE`, `IN_PROGRESS`, `TODO`, `BLOCKED`, `PARTIAL`, `SUPERSEDED`.

## Current baseline

Current main / Production release after Task #4: `e9cce65d3fe8d82100b6bbb7e1231d07dc997edb`

Schema head: 11 / `0011_customer_product_management`.

Already integrated and not to be rewritten:

- secure Runtime Naive credential lifecycle + privileged Runtime Agent;
- customer create/adopt/edit/suspend/resume/revoke-safe-delete;
- quota/unlimited/add/set volume;
- expiry/no-expiry/creation/first-CONNECT/manual validity/extend days;
- plans, groups, tags, notes;
- renewal/new ServiceTerm, Next Plan/On Hold foundations;
- server-side search/filter/sort/pagination;
- supported bulk preview/idempotent execution;
- `/sub`, `/s`, local QR, Subscription reissue/password-rotation separation;
- exact direct-Naive accounting and restart-safe telemetry core;
- trusted first-CONNECT producer core;
- session/presence projection core;
- shared hard-quota reservation/settlement core;
- tenant/RLS product foundations;
- WS4 observability/Doctor/backup/restore/release foundations deployed and verified on Production.

## Production-Ready master execution ledger

Do not reorder unless a real technical dependency is recorded in `WORKLOG.md`.

| # | Status | Priority | Task | Done gate |
|---:|---|---|---|---|
| 1 | DONE | P0 | Audit latest main / PR / CI / Production | repository and live truth captured read-only, no secret leak, provenance risks recorded |
| 2 | DONE | P0 | Compare current 3x-ui / PasarGuard / Hiddify / OV-PvNetwork | current official refs + 120-feature matrix + licensing guard |
| 3 | DONE | P0 | Fix stale canonical project docs | canonical files reconciled to current code/Production truth |
| 4 | DONE | P0 | Reconcile useful PR #16 work | manual extraction merged, backed-up Production deploy complete, Doctor/restore/load/Caddy invariants verified |
| 5 | TODO | P0 | Legacy/adopted accounting baseline truth | no fake zero; baseline provable or Unknown; no double-count |
| 6 | TODO | P0 | `/s` accounting/presence completion | real Used/Remaining/Upload/Download/Last Online/Online/Expiry/Quota or explicit unavailable |
| 7 | TODO | P0 | Manual Reset Usage | confirmation + audit + accounting reset/baseline + idempotency; no token/password rotation |
| 8 | TODO | P0 | Bulk Reset Usage | Preview → Execute using same idempotency identity |
| 9 | TODO | P0 | Periodic traffic reset | persisted scheduler/cursor/timezone/exactly-once/audit/history |
| 10 | TODO | P0 | Hard quota controlled Production proof | simultaneous connections/race/exhaustion/reload/restart/reconnect/no negative/no bypass |
| 11 | TODO | P0 | First-successful-CONNECT controlled Production proof | only successful authenticated CONNECT activates; all read/failed-auth/reload paths stay inert |
| 12 | TODO | P0 | Session management | active sessions/IP/connected/last-activity/reliable bytes |
| 13 | TODO | P0 | Kill/disconnect session | real disconnect primitive + confirmation + audit |
| 14 | TODO | P0 | Concurrent session limit | Unlimited/N enforced under races/reconnects |
| 15 | TODO | P0 | Simultaneous unique-IP limit | exact semantics, enforcement and failure tests |
| 16 | TODO | P1 | IP/session history | privacy-aware bounded retention |
| 17 | TODO | P1 | HWID/device identity PoC | implement only if trustworthy standard Naive/Karing identity exists; otherwise unavailable |
| 18 | TODO | P1 | Per-user speed-limit PoC | enforce in real data plane or do not expose option |
| 19 | TODO | P1 | Reseller CRUD | create/edit/disable/revoke/list/search |
| 20 | TODO | P0 | Full tenant-isolation / IDOR audit | no cross-reseller read/edit/renew/delete/subscription access |
| 21 | TODO | P1 | Reseller wallet/credit | audited correct balance operations |
| 22 | TODO | P1 | Immutable financial ledger | credit/debit/create/renew/refund/adjustment entries |
| 23 | TODO | P1 | Reseller plan/user restrictions | allowed plans/max users/max active users/credit/Owner oversight |
| 24 | TODO | P1 | Customer history | create/renew/volume/expiry/plan/group/tag/suspend/resume/revoke/rotate/reissue/reset events |
| 25 | TODO | P1 | Audit Explorer | actor/user/action/date/IP/result filters; strict redaction |
| 26 | PARTIAL | P1 | Notification engine/preferences/history | retry/dedupe/redaction + channel foundations exist; persistence/preferences/history/product wiring remain |
| 27 | PARTIAL | P1 | Telegram + rule builder | secure Telegram transport foundation tested; configuration/rules/history/product workflow remain |
| 28 | PARTIAL | P1 | Dashboard / monitoring / historical charts | live real CPU/RAM/disk/load/network/runtime dependency dashboard is in Production; historical charts/online aggregates remain |
| 29 | PARTIAL | P1 | Application/runtime/security logs + request diagnostics/support bundle | structured redacted logging/request IDs + support bundle exist; product log explorer remains |
| 30 | PARTIAL | P1 | Doctor command/page | `pvnaive doctor` is deployed and Production-verified; product page remains |
| 31 | PARTIAL | P1 | Scheduled encrypted backup + retention | encrypted scheduled backup/timer is active in Production; retention product policy remains |
| 32 | PARTIAL | P1 | Restore / verification / automated drill / UI | isolated automated restore drill is active and Production-verified; UI/full operator workflow remain |
| 33 | PARTIAL | P1 | Stabilize REST API + OpenAPI/Swagger | ready-route OpenAPI endpoint added; broader stabilization/version policy remains |
| 34 | PARTIAL | P1 | API rate limit / mutation idempotency / stable webhooks | request-ID/rate-limit middleware added; full mutation/webhook contract remains |
| 35 | TODO | P0 | Fix auth/security BUG-001/002/003 | refresh reuse-family, commit-before-success, DB-backed readiness regression tests green |
| 36 | TODO | P0 | Full authorization/IDOR/CSRF/redaction/fuzz gates | complete Route × Owner/Admin/Reseller/Operator/Auditor matrix |
| 37 | PARTIAL | P1 | Supply-chain security + license policy | release checksums/basic SBOM/source provenance foundation added; SAST/dependency scan/signing/NOTICE remain |
| 38 | PARTIAL | P1 | Multi-node model/auth/health/metrics/capacity/assignment/deployment | safe standalone fleet model/drift/delete guard foundation only; controller/network operations remain |
| 39 | TODO | P1 | Drain/maintenance/canary/node upgrade/rollback/failover/smart selection/fleet dashboard | reconciliation-safe fleet operations |
| 40 | TODO | P0 | Fresh secure Ubuntu installer | version-pinned dependencies/Postgres/Caddy/API/agents/systemd/firewall/TLS/migrations/web + Doctor |
| 41 | PARTIAL | P0 | Versioned upgrade | same-schema release deploy + pre-deploy backup foundation exists; generic migration upgrade remains |
| 42 | PARTIAL | P0 | Rollback + conservative uninstall | release rollback foundation exists; generic version rollback/uninstall/data policy remain |
| 43 | TODO | P0 | Client compatibility | Karing Windows/Android/iOS/macOS/Linux first; then only clients with verified Naive support |
| 44 | TODO | P0 | Load / capacity campaign | bounded local control-plane rehearsal is not capacity proof; 50/100/200/400+ campaign remains |
| 45 | TODO | P1 | Bulk/search completion | reset/delete/unlimited/no-expiry/change-expiry/next-plan + advanced filters/sorts/columns/URL state |
| 46 | PARTIAL | P1 | Final UI polish | customer dashboard and live system monitor/error boundary exist; final accessibility/product polish remains |
| 47 | TODO | P0 | Final documentation reconciliation | every feature DONE/PARTIAL/BLOCKED/OPTIONAL/N/A with evidence; no stale contradictions |
| 48 | TODO | P0 | Final clean-server installation proof | clean supported Ubuntu VM reaches healthy fully configured panel/runtime |
| 49 | TODO | P0 | Final Production smoke | backed-up RC deploy, customer/sub/accounting/runtime smoke, rollback ready |
| 50 | TODO | P0 | Release Candidate | all logical P0/P1 release blockers closed; CI exact HEAD green; provenance/evidence recorded |

## Task #4 Production evidence

Task #4 deliberately did not merge stale PR #16 wholesale. Useful pieces were manually reconciled against the newer WS1/WS2/WS3 code.

Final implementation/documentation head before merge: `09b085a877e52fa02c095799359b6b9e89bb3492`.

Replacement non-draft PR #30 passed:

- CI #1065 — SUCCESS;
- WS1 Exact Accounting #181 — SUCCESS;
- WS1 Pinned Forwardproxy #165 — SUCCESS.

PR #30 merged as `c717d162a7e9b2e31fb5822b6b16c27ad048cbbd` and was backed up/deployed to Production. Postflight found two operational false negatives rather than customer-data regressions:

1. Doctor expected root-only `0600` for `auth.key`/`runtime.key`, while the required secure service-readable layout is `0640 root:pvnaive`;
2. restore pre-validation used `age | pg_restore --list` under `pipefail`, causing a SIGPIPE/141 false failure on an otherwise valid custom archive.

Both were fixed with RED→GREEN tests in PR #31. Exact hotfix head `b740352012fd9646c25d4c70c83f64f2f86ce029` passed:

- CI #1070 — SUCCESS;
- WS1 Exact Accounting #185 — SUCCESS;
- WS1 Pinned Forwardproxy #169 — SUCCESS.

PR #31 merged as final Task #4 release `e9cce65d3fe8d82100b6bbb7e1231d07dc997edb`.

Production final verification on 2026-08-30:

- schema remains 11;
- API, Runtime Agent, Telemetry Agent, Caddy and PostgreSQL active with observed restart counters 0;
- backup and restore-drill timers active/enabled;
- six active users, six active Runtime credentials, six active ServiceTerms, six active direct Subscription tokens;
- public panel and public readiness HTTP 200;
- `pvnaive doctor`: 14 PASS / 1 disk WARN / 0 FAIL;
- real systemd restore drill: schema/ownership/ACL/signing-key checks PASS;
- bounded loopback rehearsal: 100/100 requests, 0 failures (not a capacity proof);
- Caddy SHA, PID and restart count unchanged through deployment;
- release and legacy deployment markers point to `e9cce65d...`;
- fresh encrypted config/database backups and rollback snapshots exist before both Production mutations.

## Historical PVN task reconciliation

The old 72-task ledger remains in git history. Useful crosswalk:

- `PVN-001..027`: foundation/auth/Runtime credential chain — DONE.
- `PVN-028/029`: old S04R rollout milestones — SUPERSEDED by later live schema-11 Production state.
- `PVN-030..035`: map to Master #35/#36 security gates.
- `PVN-037`: customer CRUD/lifecycle — DONE.
- `PVN-038`: quota/reset model — PARTIAL because reset execution remains.
- `PVN-039`: maps to Master #19-23 reseller/wallet/ledger/restrictions.
- `PVN-040`: user-bound Naive credential lifecycle — DONE.
- `PVN-041`: current supported bulk actions — DONE baseline; reset bulk remains Master #8.
- `PVN-042`: search/filter/sort/pagination — DONE baseline; accounting projection completion remains.
- `PVN-043`: tenant/RBAC foundation — PARTIAL.
- `PVN-044`: customer/product UI — PARTIAL.
- `PVN-045..048`: exact direct accounting/restart-safe event path — DONE.
- `PVN-049`: hard quota/reset enforcement — PARTIAL.
- `PVN-050`: maps to session/device/speed #14-18.
- `PVN-051`: runtime/system monitoring — PARTIAL with live system status.
- `PVN-052`: Subscription lifecycle/renderer — DONE.
- `PVN-053`: human account page — PARTIAL because accounting/presence completion remains.
- `PVN-054`: maps to real client compatibility #43.
- `PVN-055`: local QR/delivery UX — DONE baseline.
- `PVN-056/057`: notification/Telegram — PARTIAL foundations.
- `PVN-058..062`: backup/restore/installer/upgrade/rollback — PARTIAL foundations; installer remains.
- `PVN-063`: observability/diagnostics — PARTIAL with Production Doctor/metrics/support bundle.
- `PVN-065`: supply-chain security — PARTIAL.
- `PVN-066`: capacity — TODO.
- `PVN-067`: final release gates — TODO.
- `PVN-069`: final documentation truth — TODO.
- `PVN-071`: auth/API/web fuzz quality gates — TODO.
- `PVN-072`: license/source policy — TODO.

## Definition of Done

A feature is never `DONE` merely because code exists. Apply the Owner DoD: real backend/schema/auth/UI where applicable, no secret leaks, idempotency/failure tests, unit/integration/web tests, vet/build, exact-head CI, rollback if Runtime-affecting, live verification if Production-facing, docs/evidence and no regressions.
