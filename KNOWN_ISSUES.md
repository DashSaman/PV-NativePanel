# PVNaive — Known Issues / Risks / Technical Debt

Last updated: 2026-08-30

This file contains only current gaps or intentionally retained historical closure evidence. Do not keep obsolete statements such as “exact accounting is unproven” after the integrated WS1 + Production proof.

## P0 bugs

### BUG-001 — Refresh-token reuse-family handling is unreachable for a revoked rotated token

- Area: Auth / session security
- Owner sequence: Security bugs
- Re-verified against current main `a021aa4b62c35b775fb521d042b2f8e6dbde10b0`.
- Evidence:
  - `refresh` hashes the existing session token then calls `AuthStore.BeginAuthenticated`.
  - `BeginAuthenticated` selects the session only with `s.revoked_at IS NULL`.
  - only after that does the refresh flow call `RotateSession` / SQL `auth_rotate_session`.
  - therefore reuse of an already-rotated/revoked token can fail before the SQL reuse-family branch is allowed to detect/revoke the family.
- Risk: intended family-reuse response is not reliably reached.
- Done gate: RED regression for reuse → minimal redesign → family revocation/audit → full auth rehearsal green.

### BUG-002 — Generic authenticated HTTP success can precede durable DB commit

- Area: HTTP / consistency
- Re-verified against current `internal/httpapi/server.go`.
- Evidence:
  - `requireAuthentication` binds a transaction, then calls `next.ServeHTTP(w,r)`.
  - generic handlers can call `writeJSON` before transaction outcome is known.
  - middleware then calls `bound.Tx.Commit()` and currently ignores the commit error.
- Risk: a mutation can return HTTP success even if final DB commit fails.
- Note: dedicated Runtime mutation saga has stronger compensation; this issue is broader generic HTTP integrity.
- Done gate: buffered/commit-aware response boundary or equivalent design; injected commit failure must never emit success.

### BUG-003 — DB/schema-backed readiness is missing

- Area: Operations / readiness
- Re-verified against current `server.ready`.
- Evidence: readiness checks configured auth store/service/MFA-key presence and returns ready, but performs no bounded current DB/schema probe.
- Risk: process may stay “ready” after a DB/schema dependency failure.
- Done gate: bounded DB/schema readiness probe + timeout/failure tests; no secret leakage.

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

### SESSION-001 — Operator-facing customer session management is incomplete

Trusted accounting sessions exist, but ready user session list/kill endpoints/UI are not integrated as a product capability.

### SESSION-002 — Concurrent-session and simultaneous unique-IP limits are not enforced

Must use trustworthy active-session semantics and include race/reconnect tests.

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

- backup files exist under the PVNaive backup root;
- during 2026-08-30 read-only audit only the DB-health timer was observed; no PVNaive scheduled-backup timer.
- Done gate: encrypted scheduled backup + retention + verification + restore drill + Production timer evidence.

### OPS-003 — Production root filesystem was 79% used

Before large backup/load-test/artifact operations, re-check free space and set warning/retention policies. Do not consume remaining disk blindly.

### OPS-004 — Deployment provenance markers are stale/inconsistent

`DEPLOYED_COMMIT` / web marker information lags newer binary/web mtimes. Runtime behavior is currently healthy, but marker data is not sufficient for signed/reproducible release provenance.

Done gate: one trustworthy deployed revision/build ID covering API/web/runtime artifacts and independently verifiable against the running installation.

### OPS-005 — System monitoring/logs/Doctor are not in current main

Customer dashboard exists, but real CPU/RAM/disk/network history, application/runtime/security log UI, request diagnostics/support bundle and Doctor remain to integrate/rebuild.

## Notification/API gaps

### NOTIFY-001 — Notification product engine is incomplete

Schema/foundations do not equal delivery. Need event producers, preferences, outbox/history, retries, Telegram secret hygiene and rule builder.

### API-001 — Route registry is larger than actual ready implementation

`Routes` intentionally contains future contracts with `Ready=false`; `NewServer` can map unknown/unconfigured routes to `notImplemented`.

Rule: route declaration must never be used as parity evidence.

### API-002 — OpenAPI/Swagger and webhooks are not current-main product capabilities

PR #16 has an OpenAPI candidate. Webhooks should wait for stable event contracts.

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

### CI-001 — PR #26 final-head workflow failures required reproduction

- PR #26 final bot head recorded CI + WS1 workflow failures.
- its immediately preceding human commit and prior bot commit passed all three.
- Lead PR #27 was created from exact current main to reproduce the baseline rather than guessing.
- exact final reconciliation head still requires all required workflows green before merge.

### DOCS-001 — Legacy documentation drift

This reconciliation updates canonical files, but older stage/design/evidence documents remain historical snapshots. They should not be rewritten merely to look current; where needed they must be labeled historical/superseded.

## Closed / no longer open

### CLOSED — Exact direct-Naive accounting feasibility

Integrated WS1 and Production evidence show exact direct accounting is live; old `TEST-003 exact accounting unproven` is obsolete.

### CLOSED — Direct Subscription / Account Page absence

`/sub/<token>` and `/s/<token>` are implemented with local QR and explicit read-only/reissue/password-rotation separation.

### CLOSED — Customer CRUD / plan / group / tag / renewal absence

Current main contains these product capabilities. Do not recreate them from old S05 branches.

### CLOSED — Exact multiple Naive basic_auth syntax proof

Pinned Naive Caddy validation/rehearsal already closed this historical risk. Reopen only on a real version/module regression.
