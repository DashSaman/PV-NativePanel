# PVNaive — Canonical Roadmap

Last updated: 2026-08-30

This roadmap follows the Owner's Production/parity master prompt. Historical `PVN-*` IDs remain valid audit references; they are not reused for unrelated work. The current execution ledger below is the controlling order for Production-Ready completion.

Status vocabulary: `DONE`, `IN_PROGRESS`, `TODO`, `BLOCKED`, `PARTIAL`, `SUPERSEDED`.

## Current baseline

Audited start main: `a021aa4b62c35b775fb521d042b2f8e6dbde10b0`

Already integrated and therefore **not** future TODOs:

- secure Runtime Naive credential lifecycle + privileged Runtime Agent;
- customer create/adopt/edit/suspend/resume/revoke-safe-delete;
- quota/unlimited/add/set volume;
- expiry/no-expiry/creation/first-CONNECT/manual validity/extend days;
- plans, groups, tags, notes;
- renewal/new ServiceTerm, Next Plan/On Hold foundations;
- server-side search/filter/sort/pagination;
- supported bulk preview/idempotent execution;
- `/sub`, `/s`, local QR, Subscription reissue, password rotation separation;
- exact direct-Naive accounting and restart-safe telemetry core;
- trusted first-CONNECT producer core;
- session/presence projection core;
- shared hard-quota reservation/settlement core;
- tenant/RLS product foundations.

## Production-Ready master execution ledger

Do not reorder unless a real technical dependency is documented in `WORKLOG.md`.

| # | Status | Priority | Task | Done gate |
|---:|---|---|---|---|
| 1 | IN_PROGRESS | P0 | Audit latest main / PR / CI / Production | repository and live truth captured read-only, no secret leak, provenance risks recorded |
| 2 | IN_PROGRESS | P0 | Compare current 3x-ui / PasarGuard / Hiddify / OV-PvNetwork | current official refs + 120-feature matrix + licensing guard |
| 3 | IN_PROGRESS | P0 | Fix stale canonical project docs | canonical files match code + Production evidence; no implemented feature left “missing” |
| 4 | TODO | P0 | Reconcile useful PR #16 work | fresh branch; commit/file extraction only; newer main preserved; full CI |
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
| 26 | TODO | P1 | Notification engine/preferences/history | quota/expiry/customer/runtime/node/backup/security/disk/cert events + outbox/retry |
| 27 | TODO | P1 | Telegram + rule builder | secure token handling; rules such as traffic<10%, expiry<3d, runtime down>60s |
| 28 | TODO | P1 | Dashboard / monitoring / historical charts | real users/traffic/CPU/RAM/disk/network/runtime/online/history, no fake values |
| 29 | TODO | P1 | Application/runtime/security logs + request diagnostics/support bundle | bounded/redacted/request-ID linked |
| 30 | TODO | P1 | Doctor command/page | actionable dependency/runtime/DB/disk/TLS checks |
| 31 | TODO | P1 | Scheduled encrypted backup + retention | DB/config/runtime state/important files + failure alerts |
| 32 | TODO | P1 | Restore / verification / automated drill / UI | isolated proof + strong safeguards |
| 33 | TODO | P1 | Stabilize REST API + OpenAPI/Swagger | only ready handlers documented; auth policy explicit |
| 34 | TODO | P1 | API rate limit / mutation idempotency / stable webhooks | stable event contracts; no premature webhook surface |
| 35 | TODO | P0 | Fix auth/security BUG-001/002/003 | refresh reuse-family, commit-before-success, DB-backed readiness regression tests green |
| 36 | TODO | P0 | Full authorization/IDOR/CSRF/redaction/fuzz gates | complete Route × Owner/Admin/Reseller/Operator/Auditor matrix |
| 37 | TODO | P1 | Supply-chain security + license policy | dependency scan/secret scan/SAST/SBOM/signing/provenance/NOTICE |
| 38 | TODO | P1 | Multi-node model/auth/health/metrics/capacity/assignment/deployment | standalone remains safe; desired/applied state modeled |
| 39 | TODO | P1 | Drain/maintenance/canary/node upgrade/rollback/failover/smart selection/fleet dashboard | reconciliation-safe fleet operations |
| 40 | TODO | P0 | Fresh secure Ubuntu installer | version-pinned dependencies/Postgres/Caddy/API/agents/systemd/firewall/TLS/migrations/web + Doctor |
| 41 | TODO | P0 | Versioned upgrade | automatic pre-upgrade backup + migration safety |
| 42 | TODO | P0 | Rollback + conservative uninstall | restore path proven; data preserved by default |
| 43 | TODO | P0 | Client compatibility | Karing Windows/Android/iOS/macOS/Linux first; then only clients with verified Naive support |
| 44 | TODO | P0 | Load / capacity campaign | 50/100/200/400+ plus accounting/quota/session/restart/reconnect correctness |
| 45 | TODO | P1 | Bulk/search completion | reset/delete/unlimited/no-expiry/change-expiry/next-plan + advanced filters/sorts/columns/URL state |
| 46 | TODO | P1 | Final UI polish | professional customer/dashboard/account page + mobile/desktop/accessibility/theme |
| 47 | TODO | P0 | Final documentation reconciliation | every feature DONE/PARTIAL/BLOCKED/OPTIONAL/N/A with evidence; no stale contradictions |
| 48 | TODO | P0 | Final clean-server installation proof | clean supported Ubuntu VM reaches healthy fully configured panel/runtime |
| 49 | TODO | P0 | Final Production smoke | backed-up RC deploy, customer/sub/accounting/runtime smoke, rollback ready |
| 50 | TODO | P0 | Release Candidate | all logical P0/P1 release blockers closed; CI exact HEAD green; provenance/evidence recorded |

## Current task #1-3 evidence

The Lead reconciliation branch `lead/parity-truth-2026-08-30` / PR #27 records:

- current main + PR/CI audit;
- read-only Production schema/service/accounting/disk/backup/provenance audit;
- current competitor snapshots;
- `docs/PANEL_PARITY_MASTER_2026-08-30.md`;
- canonical documentation reconciliation.

Task #1-3 cannot become `DONE` until exact final PR #27 head passes required CI and review.

## Historical PVN task reconciliation

The old 72-task ledger is preserved in git history and still provides useful stable IDs. Its status descriptions were stale after later workstreams merged. Current crosswalk:

### Historical work now definitely completed/integrated

- `PVN-001..027`: foundation/auth/Runtime credential development chain — DONE.
- `PVN-028/029`: old S04R preflight/rollout milestones are superseded by later live Production state (schema 11, active Runtime credentials/services); do not rerun the old stage blindly.
- `PVN-037`: user/customer CRUD/lifecycle — DONE in newer main.
- `PVN-038`: plan/quota/reset **model** lifecycle — PARTIAL overall because reset execution remains.
- `PVN-040`: user-bound Naive credential lifecycle — DONE.
- `PVN-041`: bulk operations with dry-run — DONE for supported current action set; Reset Usage bulk remains new backlog.
- `PVN-042`: search/filter/sort/pagination + computed status — PARTIAL overall because accounting/presence wiring is incomplete in some projections.
- `PVN-043`: tenant/RBAC foundations — PARTIAL; full route-wide matrix remains.
- `PVN-044`: customer/product UI — PARTIAL; current daily UI exists, final product polish/status wiring remains.
- `PVN-045..048`: exact direct accounting / restart-safe event path — DONE by integrated WS1.
- `PVN-049`: hard quota/reset enforcement — PARTIAL; quota core integrated, reset execution + Production quota proof remain.
- `PVN-052`: Subscription lifecycle/renderer — DONE.
- `PVN-053`: human account page — PARTIAL only because full accounting/presence projection remains.
- `PVN-055`: local QR/delivery UX — DONE baseline.

### Historical items still open and mapped directly

- `PVN-030` → Master #35 BUG-001 refresh reuse-family.
- `PVN-031` → Master #35 BUG-002 commit-before-success.
- `PVN-032` → Master #35 BUG-003 DB readiness.
- `PVN-033` → Master #35/#36 recovery-code login policy.
- `PVN-034/035` → Master #36 auth abuse/security exposure.
- `PVN-039` → Master #19-23 reseller/wallet/ledger/restrictions.
- `PVN-050` → Master #14-18 session/device/speed capability.
- `PVN-051` → Master #28 runtime status integration.
- `PVN-054` → Master #43 real client compatibility.
- `PVN-056/057` → Master #26-27 notifications/Telegram.
- `PVN-058..062` → Master #31-32, #40-42 installer/upgrade/rollback/backup/restore.
- `PVN-063` → Master #28-30 observability/logs/diagnostics.
- `PVN-065` → Master #37 supply-chain security.
- `PVN-066` → Master #44 capacity.
- `PVN-067` → Master #48-50 final release gates.
- `PVN-069` → Master #47 documentation truth.
- `PVN-071` → Master #36 API/web authorization/fuzz quality gates.
- `PVN-072` → Master #37 license/source policy.

`PVN-070` old branch-divergence wording is superseded by the current PR classification/integration process; never force-reset old branches.

## Definition of Done

A feature is never `DONE` merely because code exists. Apply the Owner's 20-point DoD: real backend/schema/auth/UI where applicable, no secret leaks, idempotency/failure tests, unit/integration/web tests, vet/build, exact-head CI, rollback if Runtime-affecting, live verification if Production-facing, docs/evidence and no regressions.
