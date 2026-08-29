# PVNaive — Canonical Handoff

Last updated: 2026-08-30

This file supersedes the old S04/S05 branch handoff. Do not resume work from `s04-auth`, PR #5, PR #6 or the old S06 branch unless extracting one specifically reviewed piece.

## Repository

- Repo: `DashSaman/PV-NativePanel`
- Product: **PVNaive**
- Default branch: `main`
- Audited main at start of this handoff: `a021aa4b62c35b775fb521d042b2f8e6dbde10b0`
- Active Lead branch: `lead/parity-truth-2026-08-30`
- Active Lead PR: `#27`
- Current schema head: 11 (`0011_customer_product_management`)
- Current parity document: `docs/PANEL_PARITY_MASTER_2026-08-30.md`

Always re-fetch `main` before new implementation; do not trust this SHA if the repository has moved.

## Production — last read-only audit

Production: `https://namir.softarg.ir/panel/`

Observed 2026-08-30 without printing secrets:

- API active and loopback-only on `127.0.0.1:8080`;
- Runtime Agent active;
- Telemetry Agent active;
- Caddy/Naive active;
- panel and readiness return HTTP 200;
- PostgreSQL schema = 11;
- six active customer users, six active Runtime credentials and six ServiceTerms;
- six active direct Subscription tokens;
- six direct accounting term projections, all complete at audit time;
- direct accounting event/session data is live;
- no PVNaive scheduled-backup timer was observed;
- root filesystem was 79% used;
- deployment marker files lag newer binary/web mtimes and cannot be treated as authoritative release provenance.

Before any Production mutation, independently re-check all of this and take DB/config/Caddy/web/binary backups plus a rollback plan.

## What is genuinely integrated

### Runtime / accounting

- safe Naive Runtime credential management;
- privileged narrow Unix-socket Runtime Agent;
- expected-SHA Caddy validate/backup/apply/reload/postflight/rollback;
- exact successful-write direct-Naive accounting;
- dedicated accounting socket + Telemetry Agent;
- idempotent boot/session/sequence/cumulative accounting model;
- ServiceTerm usage isolation;
- trusted first-successful-CONNECT activation producer;
- session/presence projection;
- shared finite-quota reservation/settlement core.

Do **not** rewrite this lane from scratch.

### Customer / product

- customer create and Runtime adoption;
- edit/service update;
- suspend/resume/revoke-safe-delete;
- quota/unlimited, add/set volume;
- expiry/no-expiry, creation/first-CONNECT/manual validity, extend days;
- plans, groups, tags, notes;
- renewal/new ServiceTerm;
- Next Plan / On Hold foundations;
- search/filter/sort/pagination;
- supported bulk actions with preview + idempotent execution;
- tenant/RLS product foundations.

### Delivery

- `/sub/<token>` = machine Subscription;
- `/s/<token>` = human Account Page;
- local QR;
- read-only current Subscription view;
- explicit Subscription reissue;
- explicit password rotation;
- token/password/service mutations remain separate.

## What is NOT yet done

Do not mark these complete merely because schema/routes exist:

- Manual Reset Usage;
- Bulk Reset Usage;
- restart-safe periodic traffic reset execution;
- complete accounting/presence wiring in every customer and `/s` view;
- controlled hard-quota Production race/restart proof;
- controlled first-CONNECT Production activation proof;
- operator-facing session list/kill;
- concurrent session limit;
- simultaneous unique-IP limit/history;
- trustworthy HWID/device limit PoC;
- per-user speed-limit PoC/enforcement;
- full reseller CRUD/wallet/ledger/plan restriction UX/API;
- customer history + Audit Explorer;
- notification engine/preferences/history/Telegram/rule builder;
- real CPU/RAM/disk/network monitoring/history/log UIs;
- diagnostics/support bundle/Doctor;
- scheduled encrypted backup + product restore workflow;
- OpenAPI/Swagger and stable webhook contracts;
- final auth/security fixes and whole-route authorization/IDOR/fuzz gates;
- SBOM/SAST/secret/dependency scans/release signing/provenance;
- multi-node/fleet/failover/smart-node;
- generic clean Ubuntu installer/upgrade/rollback/uninstall;
- real Karing multi-OS compatibility matrix;
- 50/100/200/400+ capacity campaign;
- final UI/accessibility/clean-install/Production smoke/Release Candidate.

## Confirmed current P0 defects

Re-audited on current main:

1. Refresh-token reuse-family handling remains unreachable for an already-revoked rotated token because `BeginAuthenticated` filters `revoked_at IS NULL` before rotation reuse detection.
2. Generic authenticated middleware can write an HTTP success before `Tx.Commit()` is known; the commit result is ignored.
3. `/health/ready` is not yet a bounded DB/schema readiness check.

Do not silently remove these from `KNOWN_ISSUES.md`.

## Old PR classification

- **PR #4 — STILL USEFUL (small extract only):** old Karing export branch is obsolete as a whole, but `buildKaringSingBoxProfile` + explicit Copy Karing config UX is not present on main and can be re-evaluated during client compatibility work. Do not merge the branch wholesale.
- **PR #5 — SUPERSEDED / MERGED ELSEWHERE:** customer lifecycle foundation is represented by newer main schema/code.
- **PR #6 — SUPERSEDED / MERGED ELSEWHERE:** Sanaei-style customer flow has newer implementations on main.
- **PR #8 — SUPERSEDED / ARCHIVE:** old exact-accounting branch was replaced by the integrated WS1 implementation; do not merge it.
- **PR #16 — STILL USEFUL, REQUIRES MANUAL EXTRACTION:** contains ops/observability/backup/OpenAPI/load/fleet foundations on an old base. Never blind merge. Review commit/file units against current main.

## PR #16 useful candidate areas

Changed files show isolated candidates for:

- real system/network metrics;
- request IDs/redaction/rate-limit middleware;
- OpenAPI;
- application observability/logging;
- doctor/diagnostics;
- encrypted scheduled backup + restore drill systemd units/scripts;
- generic deploy/rollback scripts;
- bounded load rehearsal;
- notification engine/Telegram foundation;
- fleet model foundation;
- System Dashboard/error boundary.

Each must be extracted onto a fresh branch after this truth reconciliation, with newer main behavior preserved and fresh tests/CI.

## CI state for this handoff

The final PR #26 bot head recorded failed workflow runs, but immediately preceding commits passed CI and both WS1 workflows. Lead PR #27 was created from exact current main to reproduce rather than guess.

At the time of this handoff update, PR #27 had already produced a successful WS1 Exact Accounting run on the refreshed parity head, while other workflows were still running. Before merge, trigger/check all required workflows on the **final exact head**.

## Exact next task

Do not jump to new features yet.

1. Finish Task #1/#2/#3 truth reconciliation on PR #27.
2. Require exact-head CI green.
3. Merge PR #27 only after review.
4. Then Task #4: create a fresh PR #16 integration branch from latest main and extract useful units commit-by-commit.
5. After PR #16 integration is resolved, continue Owner order with legacy/adopted accounting baseline truth, `/s` accounting, Manual Reset Usage, Bulk Reset Usage and periodic resets.

## Safety invariants

- no force-push/reset of main;
- no secret in Git/chat/CI/evidence;
- no customer deletion/password rotation without the explicit requested action;
- QR/Subscription view is read-only;
- Subscription reissue != password rotation;
- quota/expiry edits do not rotate token/password;
- no fake usage/online/HWID/speed;
- no Runtime mutation without validate → backup → apply → verify → rollback;
- no Production mutation without DB/config/Caddy/web/binary backup and rollback plan;
- do not copy GPL/AGPL competitor code without explicit license review.
