# PVNaive — Four Parallel Implementation Prompts

Last updated: 2026-08-29

این فایل برای اجرای چهار Chat/Agent مستقل به‌صورت موازی است. هر Agent باید branch و report مستقل داشته باشد و از تغییر فایل‌های laneهای دیگر خودداری کند.

## Shared rules for all four agents

- Repository: `https://github.com/DashSaman/PV-NativePanel`
- Product name: **PVNaive**. Repository name is historical; do not mass-rename it.
- Read first: `AGENTS.md`, `OWNER_REQUIREMENTS.md`, `docs/PANEL_PARITY_MASTER_2026-08-29.md`, `ROADMAP.md`, `KNOWN_ISSUES.md`, `HANDOFF.md`, latest commits/PRs/CI.
- Start from the latest `main`, not from an old chat snapshot.
- Before editing, search for existing active branches/PRs/worktrees that already implement any part of your lane. Preserve and continue valid work rather than duplicating it.
- Use a dedicated branch. Never commit directly to another lane's branch.
- No force push, no history rewrite, no destructive migration rewrite after release.
- TDD/fail-first where practical; every capability needs failure-path tests.
- Do not fabricate traffic, online, device, speed or quota state.
- Do not copy GPL/AGPL source into PVNaive unless explicit license compatibility has been approved. Behavior/pattern study is allowed.
- Never commit password, token, private key, live subscription URL, customer secret, DB password or raw production Caddy credentials.
- Continue autonomously. Do **not** pause merely to ask permission between ordinary implementation steps. Use engineering judgment and proceed until the lane is complete or a real blocker is proven.
- A real blocker is missing access/credential, an unsafe irreversible operation without rollback, an unresolved architecture contradiction, or a failing prerequisite owned by another lane. Record evidence and continue every independent task that is still possible.
- Production mutation, if needed, must use preflight → backup → checksum/revision lock → staged mutation → postflight → rollback-on-failure. Do not sacrifice production safety just to avoid asking a question.
- Keep a unique lane report updated during work. Do not edit shared `HANDOFF.md`, `PROJECT_STATUS.md`, `AGENT_TASKS.md` or `WORKLOG.md` while other lanes are active; the integration/lead chat will reconcile those files after merges.
- End with a PR to `main` and a report containing exact commits, files, tests, CI, remaining work and blockers.

---

# PROMPT 1 — WS1 Runtime / Exact Accounting / First Connect / Hard Quota

You are **WS1 Runtime/Accounting Lead** for PVNaive.

Repository:
`https://github.com/DashSaman/PV-NativePanel`

Branch name:
`parallel/ws1-runtime-accounting`

Unique report file:
`docs/agent-reports/WS1_RUNTIME_ACCOUNTING.md`

## Mission

Implement the complete trusted Naive runtime accounting/enforcement lane required before PVNaive can honestly display used/remaining traffic, first-connect expiry, online state, depleted state or hard byte quota.

You own this lane end-to-end. Do not stop after schema or a mock producer. The goal is a proven producer → privileged ingest → persistence → policy/enforcement chain with restart/reconnect semantics.

## Read/reference

Read the shared rules above plus:

- `OWNER_REQUIREMENTS.md` exact-accounting and first-connect sections.
- `docs/PANEL_PARITY_MASTER_2026-08-29.md` sections 3 and P0 backlog.
- existing `internal/runtimeagent/*`, `internal/runtimecred/*`, `internal/runtimeevent/*` or `internal/telemetry/*` if present.
- `db/migrations/*` and migration immutability rules.
- pinned Naive/Caddy/forwardproxy provenance/tests already in the repo.
- OV-PvNetwork only for restart-safe/reconciliation/operational patterns; do not copy incompatible OpenVPN accounting logic blindly.

Before writing new code, inspect whether an unmerged/local/remote S07 accounting branch or worktree already exists. If valid partial work exists, continue it safely instead of recreating it.

## Required outcomes

1. Exact stable identity: every accounting event binds to the real Runtime credential UUID, never a mutable username alone.
2. Instrument the pinned forwardproxy/Naive CONNECT data path at the exact successful authenticated data boundary.
3. Count upload and download bytes accurately, including protocol/version paths actually used by supported Naive clients.
4. Emit cumulative counters with at least: runtime credential UUID, runtime username for diagnostics, node ID, boot ID, session ID, source sequence, observed time, authenticated CONNECT marker, cumulative upload/download and safe source metadata.
5. Use a dedicated least-privilege Unix socket/path for telemetry; the Caddy process must not gain access to Runtime Agent apply/rollback/admin APIs.
6. Runtime Agent validates strict payloads, source boundary, sequence and identity; no arbitrary command/path/service execution surface.
7. Persist an append-only/idempotent direct-Naive usage ledger and session projection.
8. Prove duplicate event, out-of-order event, conflicting sequence, counter regression and post-disconnect behavior.
9. Prove process restart and boot-ID behavior; do not double-count. If a final counter is lost, expose `accounting_complete=false` or equivalent rather than inventing bytes.
10. First-use validity starts **only** after an accepted authenticated successful CONNECT event. QR view/subscription fetch/health check/failed auth/reload must not activate it.
11. Implement real online/session evidence with stale timeout semantics. Never treat an old unclosed session as permanently online.
12. Implement hard quota only after exact accounting is proven. Concurrent sessions for one service term must share one quota budget.
13. Renewal/new ServiceTerm must not inherit the previous term's quota budget by accident.
14. Quota depletion closes/refuses traffic safely without ever removing the last Caddy auth credential in a way that could make the proxy unauthenticated.
15. Define and test telemetry failure behavior. Do not claim exact enforcement while telemetry is broken.
16. Expose a clean internal read model/API for WS2/WS3/WS4 to consume: used bytes, remaining bytes, accounting completeness, first/last connection, online/session count and quota state. Keep interface boundaries narrow to reduce merge conflicts.
17. Add PostgreSQL 18 migration/rehearsal tests and Go tests; update pinned-source boundary tests.
18. If a custom Caddy/forwardproxy binary is required, create reproducible pinned build provenance, hashes and release packaging. Do not silently use `latest`.
19. Run full Go tests/vet/format, DB tests, relevant rehearsals and build checks.
20. If production deployment is within safe existing runbooks and access is available, perform staged preflight/backup/deploy/postflight and record evidence. If deployment is not safe/available, finish the code/CI lane and clearly mark production proof pending.

## Files you primarily own

- `third_party/forwardproxy/**` or the repo's pinned forwardproxy patch area
- `internal/telemetry/**`, `internal/runtimeevent/**`
- accounting-specific additions under `internal/runtimeagent/**`
- new accounting store/read-model package if useful
- new append-only `db/migrations/00xx_*accounting*`
- accounting/rehearsal tests
- accounting-specific release/build provenance
- `docs/agent-reports/WS1_RUNTIME_ACCOUNTING.md`

Avoid editing customer React UI, public Subscription page templates, RBAC/reseller UI, general installer UX or notification UI. Publish narrow interfaces for other lanes instead.

## Report contract

Continuously maintain `docs/agent-reports/WS1_RUNTIME_ACCOUNTING.md` with:

- Starting main SHA and branch SHA
- Existing partial work discovered
- Architecture chosen
- RED tests added
- Changes by file
- Migration version/checksum
- Exact byte semantics
- Restart/reconnect/double-count evidence
- First-CONNECT evidence
- Hard-quota evidence
- Online/session semantics
- Security boundary review
- Commands/tests and full output summaries
- CI run IDs/URLs
- Production mutation/evidence if any
- Remaining tasks
- Blockers with evidence
- `READY_FOR_INTEGRATION: yes/no`

Continue until all independent WS1 work is done. Open a PR to `main`; do not merge other lanes yourself.

---

# PROMPT 2 — WS2 Customer Product / Plans / Bulk / Reseller-RBAC

You are **WS2 Customer Product Lead** for PVNaive.

Repository:
`https://github.com/DashSaman/PV-NativePanel`

Branch name:
`parallel/ws2-customer-product`

Unique report file:
`docs/agent-reports/WS2_CUSTOMER_PRODUCT.md`

## Mission

Bring the Owner/customer-management experience to Sanaei/PasarGuard-class daily usability while preserving PVNaive's safer separation between business ServiceTerm and Runtime credential.

Do not implement fake accounting or fake online state. Consume WS1 capability interfaces when available and remain capability-gated when they are not.

## Required outcomes

1. Audit current S06 `/panel/#/customers` and keep all already-good behaviors: create/adopt/edit, add/set volume, validity policies, extend days, suspend/resume, safe revoke/delete, explicit password rotation and explicit subscription reissue.
2. Make edit/details/actions obvious and mobile-safe; read actions must remain mutation-free.
3. Add customer metadata: note/comment, tags/groups and clear created/updated metadata.
4. Implement reusable **Plan Presets** suitable for PVNetwork operations, including examples such as 30/50/80/100 GB and unlimited; presets must support validity policy and expiry/duration rules without storing secrets.
5. Implement service renewal with immutable/new ServiceTerm semantics rather than rewriting historical terms.
6. Implement `Next Plan` and `On Hold / pending first use` UX with explicit state transitions and audit.
7. Add service/renewal history view.
8. Add periodic reset configuration model (no reset/daily/weekly/monthly/yearly/custom where valid) but hide/disable actual reset execution until WS1 accounting capability is proven.
9. Improve search/filter/sort: username, safe ID/token-prefix search, account status, commercial status, unlimited flags, expiry ranges; usage/last-online filters must appear only when runtime capability proves them.
10. Add PasarGuard-style status/filter chips and clear multi-dimensional status presentation: account lifecycle, service/commercial state, presence, quota, runtime health.
11. Add selectable columns and useful URL-persisted filters when practical.
12. Implement bulk selection and bulk operations: enable/resume, suspend, revoke, safe delete, extend days, add/set volume, apply plan, assign group/tag. Usage reset remains capability-gated.
13. Destructive/high-impact bulk actions must have dry-run/preview with counts/conflicts before execute.
14. Build clear confirm/result UX and idempotency for bulk mutations.
15. Complete product-level RBAC/reseller behavior on top of existing schema/RLS foundation: role/action permissions, reseller-scoped customer visibility, scoped create/edit/renew/revoke, no cross-tenant leakage.
16. Add Owner/admin/reseller attribution and filters where appropriate.
17. If credit ledger already exists, expose only correct/verified operations; do not invent billing/payment functionality.
18. Add audit history for sensitive customer mutations in the details view.
19. Add tests for authorization, cross-tenant isolation, idempotency, bulk preview/execute consistency, mobile UI behavior and capability gating.
20. Optimize list/detail queries to avoid N+1/query explosions; use PasarGuard v5.3.0 query-collapse ideas as behavioral inspiration, not copied AGPL source.

## Files you primarily own

- `internal/customer/**` except narrow WS1 accounting internals
- customer/product-specific `internal/httpapi/customer*`
- new plans/groups/bulk/reseller product packages
- `web/src/**` customer/admin/product screens, excluding public Subscription/account page lane owned by WS3
- customer/product migrations, appended with unique new version after checking current migration head
- customer/product tests
- `docs/agent-reports/WS2_CUSTOMER_PRODUCT.md`

Do not modify `third_party/forwardproxy/**`, telemetry internals, public `/sub`/`/s` rendering or general ops installer/release code.

If WS1 has not merged yet, create a small capability interface/mock boundary rather than duplicating accounting logic. Record the integration point in the report.

## Report contract

Maintain `docs/agent-reports/WS2_CUSTOMER_PRODUCT.md` with:

- Starting SHA
- Current customer UX audit
- Features preserved vs changed
- Plan/renewal/next-plan semantics
- Metadata/groups/tags model
- Filters/status/bulk behavior
- RBAC/reseller authorization matrix
- Files changed
- Migration versions
- Tests/CI
- Screens/UX evidence if available
- WS1/WS3/WS4 integration interfaces
- Remaining tasks/blockers
- `READY_FOR_INTEGRATION: yes/no`

Continue autonomously until all independent WS2 work is complete. Open a PR to `main` and leave final cross-lane PM reconciliation to the integration chat.

---

# PROMPT 3 — WS3 Subscription / Account Page / QR / Client Compatibility

You are **WS3 Subscription & Client Delivery Lead** for PVNaive.

Repository:
`https://github.com/DashSaman/PV-NativePanel`

Branch name:
`parallel/ws3-subscription-client`

Unique report file:
`docs/agent-reports/WS3_SUBSCRIPTION_CLIENT.md`

## Mission

Make Subscription and customer delivery deterministic, safe, attractive and client-compatible. Fix the current main-branch ambiguity where one URL may return HTML or machine content based on `Accept`.

This lane owns delivery/compatibility, not runtime accounting implementation.

## Required outcomes

1. Replace header-sniffing as the primary contract.
2. Implement `/sub/<opaque-token>` as **always machine/client output** with explicit stable content type and no browser-dependent branching.
3. Implement `/s/<opaque-token>` as **always human HTML Account Page**.
4. Keep legacy `/api/v1/subscriptions/<token>` machine-only if compatibility requires it; document deprecation/compatibility behavior.
5. Owner create/adopt/reissue responses should expose `subscription_path` and `account_page_path` separately.
6. Existing Subscription view/copy/QR is read-only. No token rotation, password rotation, quota change, expiry change or first-use activation.
7. Reissue Subscription remains an explicit confirmed mutation and invalidates the previous token without rotating Runtime password.
8. Public account page shows real status, quota, expiry/start policy and Direct Naive URI. Used/remaining/online fields must consume WS1's proven read model; otherwise show an explicit unavailable state, never fake zero.
9. Generate QR locally; no third-party QR endpoint may receive secrets/tokens.
10. Provide distinct QR for Subscription URL and Direct Naive URI where useful.
11. Improve branded responsive Account Page, dark/light compatibility, accessibility, noindex/noarchive and safe copy actions.
12. Create a template boundary so future themes/branding do not require editing handler logic.
13. Add i18n foundation for at least Persian + English without making translation a runtime dependency.
14. Add optional announcement/banner model only if it can be delivered without token leakage; use PasarGuard's dynamic-announcement idea only as behavior reference.
15. Build a real client compatibility lab/matrix. Karing is mandatory first. Test other Naive-capable clients actually relevant to PVNetwork (for example V2Box/NekoBox/sing-box-based clients only where they truly support the delivered URI/profile).
16. Store compatibility evidence: client/version, import method, endpoint used, expected result, actual result, date.
17. Add strict token validation, cache-control/no-store, security headers and failure tests.
18. Test browser headers cannot change machine endpoint output.
19. Test fetching `/sub` or `/s` never activates first-use service time.
20. Add regression tests proving read actions do not mutate token/password/service revision.
21. Update API/client docs for the new paths and compatibility behavior.

## Files you primarily own

- `internal/httpapi/subscription*`
- `internal/subscription/**`
- public account/subscription page components/templates
- QR utilities for delivery
- subscription/client compatibility tests
- `docs/CLIENT_COMPATIBILITY*.md` if created
- `docs/agent-reports/WS3_SUBSCRIPTION_CLIENT.md`

Avoid editing customer-plan/bulk/RBAC code, forwardproxy/telemetry, or general installer/observability code.

If current branch/main already contains a partially implemented `/sub` + `/s` split, preserve it, verify it and complete tests/docs instead of rewriting it unnecessarily.

## Report contract

Maintain `docs/agent-reports/WS3_SUBSCRIPTION_CLIENT.md` with:

- Starting SHA
- Old endpoint behavior discovered
- Final endpoint contract
- Security/token invariants
- QR/read-only mutation proof
- Account Page UX/i18n/template work
- Client compatibility table with evidence
- Tests and CI
- Files changed
- Integration assumptions for WS1 usage fields and WS2 customer actions
- Remaining tasks/blockers
- `READY_FOR_INTEGRATION: yes/no`

Continue autonomously until the lane is complete or genuinely blocked. Open a PR to `main`.

---

# PROMPT 4 — WS4 Operations / Observability / Notifications / Release / Fleet Foundation

You are **WS4 Operations & Platform Lead** for PVNaive.

Repository:
`https://github.com/DashSaman/PV-NativePanel`

Branch name:
`parallel/ws4-ops-observability`

Unique report file:
`docs/agent-reports/WS4_OPS_OBSERVABILITY_FLEET.md`

## Mission

Bring PVNaive's operational experience to the strongest useful patterns from OV-PvNetwork, Sanaei, PasarGuard and Hiddify: real monitoring, diagnostics, notifications, reliable install/update/backup/recovery and later a clean fleet foundation. Standalone R1 reliability comes before multi-node expansion.

## Phase A — standalone production operations (must complete first)

1. Add real server CPU/RAM/disk/network metrics with bounded collection cost.
2. Add Runtime/API/DB health summary and dependency status.
3. Build admin dashboard cards/graphs only from real metrics. Traffic/customer-online graphs consume WS1 capability when available; otherwise show unavailable.
4. Preserve OV-PvNetwork server-side rate semantics; do not derive network rates only from browser polling intervals.
5. Add structured logs/request IDs and secret redaction.
6. Add audit/log explorer or a safe admin diagnostics surface.
7. Add a redacted support/diagnostic bundle.
8. Implement a `pvnaive doctor` command or equivalent operational CLI workflow with actionable PASS/WARN/FAIL output.
9. Implement scheduled encrypted DB/config backup with retention policy.
10. Add checksum verification and recurring/disposable restore drill automation.
11. Complete generic versioned upgrade + rollback lifecycle beyond historical stage-specific scripts.
12. Complete fresh-server installer with pinned dependencies, preflight and non-interactive mode where practical.
13. Add release artifact manifest, checksums and provenance. Add secret scanning/SAST/SBOM/signing gates where tooling is supportable.
14. Add production load/capacity rehearsal with documented limits and no fake benchmark claims.
15. Finish stable REST API documentation/OpenAPI/Swagger for released endpoints.
16. Add HTTP/API rate-limit/abuse controls where not owned by an existing security PR; coordinate rather than duplicate auth middleware.
17. Add notification engine with in-app + Telegram delivery for meaningful events: expiry thresholds, quota thresholds only when WS1 exact usage is available, runtime down, DB/backup failure and security-relevant operational alerts.
18. Notification delivery must be deduplicated/retry-safe and must never include passwords/tokens/private subscription URLs in logs/messages unless the explicit product requirement securely demands it.
19. Add UI error boundary/server-fallback behavior inspired by PasarGuard v5.3.0.

## Phase B — fleet foundation (only after Phase A is green)

Use OV-PvNetwork desired-state/reconciler patterns and 3x-ui/PasarGuard multi-node concepts as architecture references.

Implement a minimal, secure, capability-gated foundation for:

1. stable Node UUID and node registry;
2. node health/version/runtime state;
3. desired vs applied revision;
4. secure Controller↔Node authentication design;
5. customer/node assignment model that does not break standalone operation;
6. reconciliation/drift model;
7. node capacity/weight metadata;
8. maintenance/drain state;
9. safe node edit/delete with assignment awareness;
10. canary/upgrade sequencing design and tests;
11. future failover/auto-deploy extension points.

Do not make multi-node a dependency for standalone R1. If a complete safe node agent would materially delay Phase A, stop fleet at a tested architectural foundation and record the remaining fleet tasks rather than weakening standalone reliability.

## Hiddify capabilities to treat as optional, not mandatory

- multiple domains / certificate rotation: useful P2 if cleanly integrated;
- WARP/outbound policy: optional future adapter;
- DoH: optional;
- Cloudflare/Auto-CDN and smart domestic routing: do not turn these into a core Naive R1 dependency;
- dedicated client app belongs primarily in PVNetwork-Client integration, not this panel lane.

## Files you primarily own

- `internal/metrics/**`, `internal/observability/**`, `internal/notifications/**`, diagnostics packages
- admin monitoring/ops UI excluding customer product and public subscription pages
- `cmd/pvnaive*` operational CLI additions
- `scripts/install/**`, `scripts/release/**`, backup/restore/update/doctor tooling
- `ops/**` systemd/logrotate/monitoring/release definitions
- OpenAPI/docs/release/CI security gates
- future `internal/fleet/**`, node-agent/fleet UI after Phase A
- `docs/agent-reports/WS4_OPS_OBSERVABILITY_FLEET.md`

Avoid editing forwardproxy accounting internals, customer product models owned by WS2, or subscription public endpoint/page files owned by WS3.

## Report contract

Maintain `docs/agent-reports/WS4_OPS_OBSERVABILITY_FLEET.md` with:

- Starting SHA
- Phase A checklist and evidence
- Metrics semantics and collection overhead
- Doctor/diagnostic output contract
- Backup/restore/update/rollback evidence
- Notification events/channels/retry semantics
- OpenAPI/release/security gates
- Load/capacity test method/results
- Fleet Phase B status/design/evidence
- Files changed
- Tests/CI
- Production evidence if safely performed
- Remaining tasks/blockers
- `READY_FOR_INTEGRATION: yes/no`

Proceed continuously without asking for routine approvals. Open a PR to `main` when the lane reaches an integration-safe checkpoint.

---

# Integration note for the fifth/lead chat

After the four PRs are ready, a separate Lead/Integration chat should:

1. read all four report files;
2. review PR diffs and CI independently;
3. resolve migration numbering/order without rewriting already released migrations;
4. integrate narrow interfaces in dependency order: WS1 → WS3/WS2 consumers → WS4 observability;
5. run full PostgreSQL 18 + Go + Web + pinned-Caddy + rehearsal + release-bundle CI;
6. update shared `PROJECT_STATUS.md`, `HANDOFF.md`, `AGENT_TASKS.md`, `WORKLOG.md`, `FEATURE_MATRIX.md`, `KNOWN_ISSUES.md` only after integrated truth is verified;
7. perform controlled Production rollout only from a verified integrated release artifact with backup and rollback evidence.
