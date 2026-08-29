# PVNaive — Known Issues / Risks / Technical Debt

Last updated: 2026-08-28

This file records known problems that must not be silently forgotten or called complete. Every open item maps to a stable `PVN-*` roadmap task.

## Bugs

### BUG-001 [P0] Refresh-token reuse detection is unreachable through current HTTP flow

- Maps to: `PVN-030`
- Area: Auth / Session security
- Evidence:
  - SQL `pvnaive.auth_rotate_session` contains a revoked-token branch that marks family reuse.
  - HTTP `refresh` first calls `BeginAuthenticated(oldHash)`.
  - `BeginAuthenticated` only loads sessions with `revoked_at IS NULL`.
  - therefore reuse of a previously rotated/revoked token fails before the SQL reuse-detection branch is reached.
- Risk: refresh-family theft/reuse does not get the intended family-revocation evidence path.
- Done: RED regression proves bug, flow is redesigned minimally, reused token revokes family, full auth rehearsal remains green.

### BUG-002 [P0] HTTP success may be written before authenticated transaction commit is known

- Maps to: `PVN-031`
- Area: HTTP / DB consistency
- Evidence: authenticated middleware can run the handler before the final transaction commit result is known.
- Risk: a mutation outside the dedicated runtime saga could emit success while DB commit later fails.
- Note: the S04R runtime mutation path now has its own compensation/reconciliation handling, but this broader auth/HTTP integrity bug remains open.
- Done: generic authenticated responses cannot escape before durable commit; injected commit failure returns an error and no false success.

### BUG-003 [P1] Readiness endpoint does not prove ongoing DB readiness

- Maps to: `PVN-032`
- Area: Operations
- Evidence: readiness does not yet provide the final bounded ongoing DB/schema proof required by the roadmap.
- Risk: process may report ready after DB connectivity/schema becomes unhealthy.
- Done: bounded live DB dependency/schema probe plus failure-path tests and no secret leakage.

### BUG-004 [P1] Recovery codes are not accepted in login flow

- Maps to: `PVN-033`
- Area: MFA
- Evidence: recovery codes are generated/consumed for MFA-management behavior while login accepts TOTP.
- Risk: account recovery semantics are incomplete/ambiguous.
- Done: explicit product decision. If recovery-code login is supported, one-time consumption/replay/session/audit tests must pass; otherwise docs/UI must clearly state narrower purpose.

## Security gaps

### SECURITY-001 [P0] HTTP/IP-aware login rate limiting and progressive delay absent

- Maps to: `PVN-034`
- Current mitigation: DB actor lockout after repeated failures.
- Missing: bounded request-rate control, trusted reverse-proxy client-IP policy and progressive delay at HTTP edge.
- Pilot rule: customers must never receive Owner panel credentials; use strong Owner credentials/MFA where configured.

### SECURITY-002 [P0] Supply-chain/release security gates absent

- Maps to: `PVN-065`
- Missing today: SBOM, dependency vulnerability audit, SAST, secret scan, signed release/provenance policy.
- Current CI provides Go/Web/PG18/rehearsal/bundle checksums but is not the final signed release pipeline.

### SECURITY-003 [P1] S04 public-exposure security review not formally closed

- Maps to: `PVN-035`, `PVN-036`
- Public panel/API exists and S04R has strong secret/runtime boundaries, but formal S04 stage closure still requires remaining auth hardening and independent external postflight.

## Technical debt

### TECH-DEBT-001 [P0] `s04-auth` and `main` remain diverged

- Maps to: `PVN-070`
- Why it matters: `main` contains newer production evidence while active implementation lives on `s04-auth`.
- Rule: do not reset or force-push. Reconcile only at a fully green checkpoint while preserving production evidence.

### TECH-DEBT-002 [P1] Legacy documentation truth drift

- Maps to: `PVN-069`
- Canonical PM files are being synchronized, but older README/SECURITY/API/product-gap documents may still describe scaffold/target behavior.
- Done: actual/current vs target/future status is explicit everywhere.

### TECH-DEBT-003 [P1] Future route registry is much larger than actual implementation

- Maps to: `PVN-069`, `PVN-071`
- Many future business routes remain declarations only.
- Current implemented API includes health/auth plus Owner-only S04R runtime endpoints; business/user/subscription routes must not be inferred from registry entries.

### TECH-DEBT-004 [P1] Full protocol/accounting adapter remains incomplete

- Maps to: `PVN-045`…`PVN-051`
- S04R now safely manages credentials but does not prove exact accounting, sessions, device limits, speed limits or quota enforcement.
- Rule: capability flags and UI must remain honest until PoCs prove them.

## Test gaps

### TEST-001 [P1] Full authorization/IDOR/fuzz quality gate incomplete

- Maps to: `PVN-071`
- Existing S04R API and parser have targeted authorization/failure tests, but the final whole-product route/role/tenant/fuzz matrix is still future work.

### TEST-003 [P0] Exact Naive per-credential accounting remains unproven

- Maps to: `PVN-045`…`PVN-048`
- Until measured, billing/quota/session UI must not claim exact values.

## Deployment / release gaps

### DEPLOY-001 [P0] Formal S04 ledger not closed

- Maps to: `PVN-036`
- Implemented milestones/public preview/S04R code are real, but official `S04=PASSED` waits for the independent final gate.

### DEPLOY-002 [P0] S04R live preflight and production rollout are not yet evidenced

- Maps to: `PVN-028`, `PVN-029`
- Development/rehearsal is complete through PVN-027.
- Production schema remains v2 until the guarded Pilot upgrade is actually run.
- Read-only `S04R-preflight.sh` and guarded `S04R-upgrade.sh` are packaged and CI-gated.
- Live rule: follow `docs/PILOT_INSTALL_FA.md`; one step at a time; never skip the exact Caddy SHA lock or pre-upgrade encrypted DB backup.

### DEPLOY-003 [P0] No final general-purpose fresh installer/release path yet

- Maps to: `PVN-058`…`PVN-067`
- A stage-specific S04R Pilot upgrade path for the existing installation now exists.
- It is **not** the final fresh-server one-command installer, generic upgrade/uninstall lifecycle or signed Release Candidate.

### DEPLOY-004 [P1] Pilot is raw Runtime credential management, not customer lifecycle

- Maps to: `PVN-037`…`PVN-055`
- The Owner can create a credential and hand a copy-ready Naive link to a customer after PVN-029 live smoke.
- Missing intentionally: customer portal/login, quota, commercial expiry, reseller, subscription page/token, exact usage, device/concurrency/speed controls.

## Legal / source-use gap

### LEGAL-001 [P1] PVNaive has no explicit repository license policy recorded

- Maps to: `PVN-072`
- Repo audit did not find a root LICENSE file.
- GPL/AGPL competitor projects are architecture/product references only unless an explicit compatible licensing decision is made.

## Closed / verified S04R risks

### CLOSED TEST-002 — Exact multiple Naive `basic_auth` syntax proof

- Former maps: `PVN-021`, `PVN-027`.
- The pinned `v2.11.2-naive` Caddy asset is downloaded in CI with a fixed SHA-256.
- CI runs `caddy validate/adapt` against multiple `basic_auth` directives in one `forward_proxy` block.
- Full S04R rehearsal also exercises multiple credentials through create/disable/enable/rotate/revoke lifecycle.
- Do not reopen unless the production preflight reports a different Caddy version/module/binary or a regression reproduces failure.

### CLOSED S04R-SECRET-001 — generated secret leakage/replay

- generated password is returned only after successful mutation commit;
- idempotency replay does not return it again;
- list/GET responses do not contain plaintext/ciphertext/nonce/hash fields;
- DB rehearsal asserts encrypted envelope shape and excludes plaintext.

### CLOSED S04R-CONSISTENCY-001 — DB finalization + Runtime rollback double failure ambiguity

- Runtime saga now distinguishes a reconciliation-required state and maps it to dedicated API behavior rather than generic runtime-unavailable success/failure ambiguity.

## Closed historical failures

- missing `file` package during first S04 live attempt — fixed;
- same-second encrypted backup destination collision — regression added/fixed;
- S04 API DB environment/startup mismatch — fixed;
- S04 migration remained v2 after safe rollback refusal — later recovery/backup/postflight completed;
- full S04R CI initially omitted the rehearsal helper binary — RED observed and build wiring fixed;
- old S04 bundle omitted Runtime Agent — bundle contract RED observed and S04R bundle rebuilt correctly;
- web Runtime test typing failed production TypeScript build — fixed and build-gated.

Do not reopen a closed failure unless a regression reproduces it.
