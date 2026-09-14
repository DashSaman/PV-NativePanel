# PVNaive — Known Issues / Risks / Technical Debt

Last updated: 2026-08-31

This file contains only current gaps or intentionally retained historical closure evidence. Do not keep obsolete statements such as “exact accounting is unproven” after the integrated WS1 + Production proof.

## P0 bugs

### CLOSED — BUG-001 Refresh-token reuse-family handling

- Closed on main `11ef273fe5ccbebc0b96e8ce45aa97bbf13d1757` by the schema18/auth hardening merge.
- The revoked-token reuse path is now reachable without weakening refresh-hash validation, family revocation remains server-side, rollback restores the schema17 rotate function, and the auth regression/CI gates are green.
- Main CI run `33406624501` completed successfully on 2026-08-31. Reopen only on a concrete regression.

### CLOSED — BUG-002 Generic authenticated HTTP success could precede durable DB commit

- Closed by PR #47, merged as `fce39283c6449b0d1836757ee7caddb31fab9def`.
- `requireAuthentication` now buffers authenticated handler status/headers/body until the shared transaction commits.
- A commit failure discards the buffered response and emits a redacted HTTP 500; the success body/headers are never exposed first.
- The response buffer intentionally does not implement `http.Flusher`, preventing authenticated handlers from bypassing the commit boundary through streaming flushes.
- Runtime mutation finalization marks `TransactionFinalized` only after `CommitAndFinalize` succeeds.
- Injected commit-failure/no-leak tests, `go vet`, full Go tests and race tests passed on the PR head.
- PR-head CI, WS1 Exact Accounting and WS1 Pinned Forwardproxy all passed; merged-main CI run `33426149726` also passed.
- Reopen only on a concrete response-before-commit regression.

### CLOSED — BUG-003 DB/schema-backed readiness

- Main includes a bounded DB/schema-backed readiness probe and Production readiness evidence is green.
- Failure paths are fail-closed and do not disclose database details or secrets.
- Reopen only on a concrete readiness regression.

## P0 accounting / enforcement gaps

### ACCOUNTING-001 — Legacy/adopted accounting baseline policy needs explicit truth

- Exact direct-Naive accounting itself is integrated and live.
- Remaining problem: adopted/legacy Runtime history prior to the trusted accounting boundary must never appear as fake zero usage.
- Done gate: per-term baseline is either provably correct or explicitly `Unknown`; upload/download/remaining/last-online/session projection cannot double-count.

### ACCOUNTING-002 — Manual Reset Usage is not a ready capability

- `usage_reset_events` schema/foundation existence is not sufficient.
- Current ready customer routes do not implement a complete Reset Usage mutation.
- Done gate: confirmation, audit, accounting reset event/baseline, idempotency, no password/token rotation, single-user UI/API + tests.

### ACCOUNTING-003 — Periodic reset execution is not implemented

- Plan model supports `none/daily/weekly/monthly/yearly/custom` reset strategy.
- Missing: restart-safe scheduler, persisted cursor, timezone policy, exactly-once/idempotent execution, audit/history.

### ACCOUNTING-004 — Hard-quota core needs controlled Production acceptance proof

- Shared quota reservation/settlement exists; do not rewrite it.
- Required proof still includes simultaneous connections, exact exhaustion, reload/restart/server restart/reconnect, no negative remaining and no bypass.

### ACCOUNTING-005 — First-successful-CONNECT core needs controlled Production acceptance proof

- trusted CONNECT producer/core exists.
- Required proof: QR/sub/s/health/reload/failed-auth do not activate; successful authenticated CONNECT does; duplicate/concurrent/reconnect/restart behavior is idempotent.

## P0/P1 session and limit gaps

### SESSION-001 — Session kill remains after active-session delivery

Task12 active-session listing is integrated and deployed at schema17 with trusted Caddy `RemoteAddr`, exact session bytes/timestamps and tenant-scoped projection. Remaining gap is Task13: an exact one-session disconnect primitive plus confirmation/audit. Do not fake this by revoking the whole credential or restarting/reloading Caddy.

### SESSION-002 — Simultaneous unique-IP limit remains

Task14 concurrent-session limit is deployed at schema19 with PostgreSQL race/reconnect proof and Production evidence in `ops/evidence/TASK14-20260831-concurrent-session-limit-production-pass.md`.

Task15 remains open. A schema20 candidate was rejected before publication because it attempted to count `direct_naive_accounting_sessions.client_ip`, while the authoritative trusted peer IP delivered by Task12 is stored in `direct_naive_accounting_session_peers`, and the proposed new ingest IP parameter was not actually wired from the pinned forwardproxy/Telemetry boundary. The final design must enforce from trusted Caddy `RemoteAddr`, before payload forwarding, with PostgreSQL race proof and without fabricating identity from client headers.

### SESSION-003 — HWID identity is not proven

No fake HWID. Implement only if a stable Karing/Naive client identity can be proven; otherwise expose capability unavailable.

### SESSION-004 — Per-user speed limit is not proven/enforced

No fake UI option. Implement only after a real data-plane shaping boundary is proven.

## Product / RBAC gaps

### PRODUCT-001 — Reseller product surface is incomplete

- tenant/RLS/product foundations exist;
- full reseller CRUD, disable/revoke, wallet/credit, immutable ledger, allowed-plan/max-user/max-active-user policy and Owner oversight UI/API remain.

### PRODUCT-002 — Customer history / Audit Explorer incomplete

Audit events exist, but service-history/customer-history projection and actor/user/action/date/IP/result explorer are not complete.

### PRODUCT-003 — Some accounting/presence status projections remain capability-gated in customer views

Direct accounting is live, but not every list/status/page path consumes it consistently. Do not show fake `0`/offline values when projection is unavailable/incomplete.

## Operations gaps

### OPS-001 — Stale PR #16 contains useful unintegrated operations work

PR #16 is old and conflict-prone. It must never be blind-merged. Candidate units: metrics/network rates, request IDs/redaction/rate limits, OpenAPI, doctor/diagnostics, scheduled encrypted backup/restore drill, generic deploy/rollback, load rehearsal, notification/fleet foundations and System Dashboard.

Done gate: fresh branch from latest main, inspect/extract unit-by-unit, preserve newer behavior, TDD/full CI.

### OPS-002 — Scheduled backup is not active on Production

Historical note: this statement was true during the 2026-08-30 pre-Task4 audit. Later Task4 Production evidence records encrypted backup and restore-drill timers active. Retention/product-policy work remains separate and should not be confused with absence of the timer.

### OPS-003 — Production root filesystem was 79% used

Before large backup/load-test/artifact operations, re-check free space and set warning/retention policies. Do not consume remaining disk blindly.

### OPS-004 — Deployment provenance markers are stale/inconsistent

Historical pre-Task4 marker drift was repaired by guarded release tooling. Continue to verify exact deployed revision/build markers before each Production mutation; do not infer provenance only from mtimes.

### OPS-005 — System monitoring/logs/Doctor product completion remains partial

Live system monitoring, Doctor and support-bundle foundations are deployed, but historical charts and full operator log explorer/product pages remain incomplete.

## Notification/API gaps

### NOTIFY-001 — Notification product engine is incomplete

Schema/foundations do not equal delivery. Need event producers, preferences, outbox/history, retries, Telegram secret hygiene and rule builder.

### API-001 — Route registry is larger than actual ready implementation

`Routes` intentionally contains future contracts with `Ready=false`; `NewServer` can map unknown/unconfigured routes to `notImplemented`.

Rule: route declaration must never be used as parity evidence.

### API-002 — OpenAPI/webhook product completion remains partial

A ready-route OpenAPI foundation exists from Task4. Broader API stabilization and webhooks should wait for stable event contracts.

## Security gaps

### SECURITY-001 — HTTP/IP-aware brute-force protection is incomplete

DB actor lockout is not the final edge policy. Need trusted-proxy client-IP boundary, request-rate control and progressive delay/lockout tests.

### SECURITY-002 — Whole-product authorization/IDOR matrix incomplete

WS2 tenant/RLS protections exist, but every ready Route × Owner/Admin/Reseller/Operator/Auditor negative test has not yet been completed.

### SECURITY-003 — Recovery-code login policy unresolved

Recovery codes exist for MFA-management behavior; login recovery remains a product decision. Implement fully or document unsupported.

### SECURITY-004 — Supply-chain security gates incomplete

Missing final SBOM, SAST, secret scanning, dependency vulnerability scanning, release signing/provenance policy.

### LEGAL-001 — Repository/source-use license policy needs final documentation

GPL/AGPL competitor source is reference-only unless explicit compatibility review approves code reuse. Add final root license/NOTICE/source-reference policy before RC.

## Installer / fleet / capacity gaps

### INSTALL-001 — No final generic clean-Ubuntu one-line installer

Existing stage-specific install/upgrade tooling is useful but not the required version-pinned fresh installer/doctor/uninstall lifecycle.

### FLEET-001 — Multi-node/failover/smart-node not integrated

Standalone correctness is intentionally first. Fleet work begins only after earlier P0/P1 gates.

### CLIENT-001 — Real Karing compatibility matrix incomplete

`/sub`, `/s`, direct Naive and QR contracts exist. Real current Karing Windows/Android/iOS/macOS/Linux evidence is still required. Old PR #4 has a small explicit sing-box/Karing export idea worth re-evaluating, not wholesale merging.

### LOAD-001 — Target 400-concurrent capacity is not proven

Need 50/100/200/400+ campaign measuring CPU/RAM/disk/PostgreSQL/connections/Caddy/telemetry/accounting lag/network/API latency plus accounting/quota/session-race/restart/reconnect correctness.

## CI / governance gaps

### DOCS-001 — Legacy documentation drift

Canonical files are being reconciled to current evidence, but older stage/design/evidence documents remain historical snapshots. They should not be rewritten merely to look current; where needed they must be labeled historical/superseded.

## Closed / no longer open

### CLOSED — Exact direct-Naive accounting feasibility

Integrated WS1 and Production evidence show exact direct accounting is live; old `TEST-003 exact accounting unproven` is obsolete.

### CLOSED — Direct Subscription / Account Page absence

`/sub/<token>` and `/s/<token>` are implemented with local QR and explicit read-only/reissue/password-rotation separation.

### CLOSED — Customer CRUD / plan / group / tag / renewal absence

Current main contains these product capabilities. Do not recreate them from old S05 branches.

### CLOSED — Exact multiple Naive basic_auth syntax proof

Pinned Naive Caddy validation/rehearsal already closed this historical risk. Reopen only on a real version/module regression.

## Open infrastructure issues (2026-09-14 live-server batch)


### CLOSED (2026-09-14) — DEPLOY-001: boot config renderer drops active customer credentials

- Closed by `internal/runtimeconfig.ReconcileRuntimeConfigFile` + the `pvnaive reconcile-runtime-config` subcommand: the all-in-one entrypoint now reconciles the rendered config's credential block with active DB credentials on EVERY boot (superuser snapshot over the local socket, production renderer, pinned-binary validation, atomic swap; failure is non-fatal with a loud warning). Zero-active is a no-op (fresh-install seed).
- The `docker/` build context (Dockerfile, entrypoint, template, compose, installer) is upstreamed into the repo and the deployed image `pvnaive:repo-live` was built from the repo tree. Live evidence: boot log `RECONCILE_RESULT=CHANGED:false CREDENTIALS:22` after recreate.
- TLS storage persistence added at the same time (`XDG_DATA_HOME=/var/lib/pvnaive/tls-data` + `./data/tls` compose volume) — recreates no longer re-obtain certificates (the cause of the 2026-09-14 Let's Encrypt rate-limit outage; see AGENTS.md batch-2 notes for the dual-name interim state and the namir auto-retry window).

### CLOSED (2026-09-14) — LINEAGE-001: deployed migration lineage ahead of repo lineage

- Repo `db/migrations` is now `0001..0028`: deployed server files 0022..0027 upstreamed byte-identical (UP-file checksums are immutable in `schema_migrations`), server down files normalized to repo convention (version header + transactional/destructive markers + self-delete; 0024..0027 lacked the self-delete and broke `rollback.sh`), manifest regenerated.
- The hardened self-service functions are new **0028_auth_actor_credential_mutations_hardened** (authentication-context guard over the deployed 0027 bodies; CREATE OR REPLACE upgrades live in place). `tests/db/auth_actor_credential_mutations_migration_test.sh` now asserts schema 28. Full 35-gate CI database suite green.

### CLOSED (2026-09-14) — readiness 503 / baked PVNAIVE_EXPECTED_SCHEMA_VERSION

- The image no longer bakes the expected schema: `docker/entrypoint.sh` derives it from the bundled migration count (operator override still wins). Live boot: `derived PVNAIVE_EXPECTED_SCHEMA_VERSION=28`, readiness `{"ready":true,"schema":"ok"}` without any override env.

### RESOLVED (2026-09-14) — pvbootstrap is obsolete in the runtime-mapped proxy architecture

- The custom `forward_proxy` Caddy module now REQUIRES a runtime UUID mapping (DB-backed identity) for every `basic_auth` user; adding the bootstrap account without a mapping is rejected at config load: `missing or invalid runtime UUID mapping for configured user "pvbootstrap"`.
- The live Caddyfile intentionally contains only customer credentials (22 entries); `pvbootstrap` cannot proxy and must not be re-added to the rendered Caddyfile.
- Docs that still advertise the `pvbootstrap`/`Bt7xKp9mVq2wRt8z` proxy account are stale — treat the installer bootstrap account as API/owner bootstrap only.
- Note for agents: never attempt to reload a Caddyfile containing unmapped users — validate first (`caddy validate`), and remember the running config survives a failed reload.

### R1-NET-001 — Client-path TCP_INFO coverage limited to HTTP/1 hijack sessions (OPEN, by design of net/http)
`net/http` does not expose the client-side TCP conn for H2/H3 requests, so the R1 sampler
records `client`-path samples only on HTTP/1 hijack sessions; H2/H3 sessions contribute
`upstream`-path samples (node↔destination). Steering therefore prefers fresh client-path
aggregates and falls back to upstream-path aggregates per (user, node). Revisit if Caddy
exposes per-request conn info. Evidence: third_party/forwardproxy overlay + WORKLOG
2026-09-14 03:5x entry.

### R1-NET-002 — R1 image not yet deployed to production (OPEN, procedure-bound)
Migration 0031 + telemetry agent + sampler Caddy are merged to main but the production image
on 45.141.148.59 is still `pvnaive:repo-live2` (a4edea6, schema 30). Deploy requires the #100
procedure (fresh encrypted backup → independent rollback snapshot → exact SHA lock → staged
promotion → postflight). Do NOT deploy without that receipt. This is also the STEER-001 live
evidence blocker.

### R7-NET-001 — Panel-access runtime apply (Caddyfile base path/port) not yet wired (OPEN)
Migration 0030 + settings API + recovery CLI are merged, but the runtime reconciliation
that regenerates the Caddyfile (base path, listen port, graceful dual-accept window) is not
implemented — panel_settings changes currently affect the API contract only and require an
operator-coordinated reload. Default exposure remains reverse_proxy on /panel. Track with
the R8/R5 integration lanes; do NOT claim ACCESS-001 live until the flip is rehearsed.
