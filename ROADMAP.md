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

## 2026-09-14 operator batch — live server 45.141.148.59 (evidence-based)

Executed and verified live by agent session Super-Z (see AGENTS.md "Live environment" for credentials/location facts):

| Item | Status | Evidence |
|---|---|---|
| GitHub write access (fine-grained PAT, blob-probe 201) + 5 commits pushed to `main` | DONE | push `0c245b5..dbbcb98`, CI run 34775222448 |
| Domain switched to `namir.softarg.ir` (A record → 45.141.148.59), Let's Encrypt issued + auto-renew | DONE | panel HTTPS 200 verify=0; subscriptions now domain-based (clients re-import once) |
| BBR + fq + host sysctl tuning (`/etc/sysctl.d/99-pvnaive-tuning.conf`) | DONE | cubic→bbr verified; quic receive-buffer warning gone |
| Self-service admin password/login-identity change from inside panel (`POST /api/v1/me/password`, `PATCH /api/v1/me/profile`, migration 0027 SECURITY DEFINER, `SettingsSecurity.tsx`) | DONE live (deployed image pvnaive:fix2) | 7-step live E2E pass; repo commits `feat(panel) account security` + `fix(auth) SECURITY DEFINER` |
| UI/UX overhaul "Amber Command Deck" (Vazirmatn self-hosted font, SVG icon set `web/src/ui.tsx`, glass/aurora design system in `styles.css`, login/command-deck polish, light+dark refined, reduced-motion + focus rings) | DONE on main (build+61 web tests green) | deploy to live server still pending image rebuild |
| Competitor research snapshots (3x-ui v3.0→v3.7 full release notes, PasarGuard, Hiddify) saved | DONE | repo: `docs/competitor/3xui-v3-release-notes.md` (summary in FEATURE_MATRIX) |

## Remaining work — numbered, competitor-informed priority list

Source: full 3x-ui v3.0→v3.7 release notes, PasarGuard v5.3, Hiddify Manager v12 (official). Ordered by operator value; map to legacy task IDs where they exist.

1. **Deploy the current main (incl. UI overhaul + account security) to the live server** — rebuild all-in-one image from repo lineage; needs LINEAGE-001 reconciliation first (see KNOWN_ISSUES).
2. **Fix DEPLOY-001 boot Caddyfile renderer** (P0): entrypoint must merge active DB credentials on every boot, never render bootstrap-only.
3. **Traffic accounting truth** (Task 5/6/10): exact usage numbers are the #1 operator trust feature; 3x-ui shows per-client online+speed+totals live.
4. **Live per-client online status + realtime speed** (3x-ui v3.3.1 online-stats, v3.5.0): presence dot + last-online + per-client speed in customer table.
5. **DB backup/restore from panel + schedule** (3x-ui v3.2.5; Hiddify 6h auto-backup): encrypted backups, download, restore-with-validation, retention.
6. **Multi-format subscriptions**: Clash/Mihomo + sing-box + raw Xray JSON + auto-format-by-User-Agent (3x-ui v3.2.8/v3.6.0); subscription page templates (v3.3.0).
7. **Fail2ban-native IP limiting** + trusted-IP exemptions (3x-ui v3.4.0/v3.7.0); HWID/device limits per subscription (v3.7.0) — maps to Task 16/17.
8. **Notification event bus**: Telegram + SMTP/email + threshold alerts (offline/down/depleted/expiry) with per-event subscribe (3x-ui v3.4.0) — upgrade our simple TG channel; maps to Task 34.
9. **Scoped, expiring API tokens + in-panel OpenAPI docs** (3x-ui v3.7.0/v3.2.0) — maps to Task 33/34.
10. **Calendar-day renewals + per-client reset cycle + auto-renew cap** (3x-ui v3.7.0) — extends existing usage-reset machinery.
11. **Settings UX**: tag shipped-default values, 2FA re-confirm for sensitive changes (3x-ui v3.4.2/v3.6.0).
12. **Accessibility pass** (screen-reader/keyboard, axe-clean) — 3x-ui v3.4.2/v3.6.0; matches Task 46.
13. **Command-deck overview upgrades**: two-series throughput + TCP/UDP connection charts on dashboard (3x-ui v3.6.0) once telemetry API exposes them.
14. **Multi-node resilience** (Tasks 38/39; 3x-ui v3.2.8→v3.7.0 hardened sync): offline-node edits survive, per-node routing, mTLS reconcile.
15. **Client lifecycle billing**: renewal history, per-client external link controls, subscription last-fetch time (3x-ui v3.5.0/v3.7.0).
