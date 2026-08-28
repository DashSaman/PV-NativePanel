# PVNaive — Known Issues / Risks / Technical Debt

Last updated: 2026-08-28

This file records known problems that must not be silently forgotten or called complete. Every item maps to a stable `PVN-*` roadmap task.

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
- Evidence: authenticated middleware runs handler and then ignores `_ = bound.Tx.Commit()`.
- Risk: a mutation can emit success to the browser while the DB commit later fails.
- Done: response success cannot escape before durable commit; injected commit failure returns an error and no false success.

### BUG-003 [P1] Readiness endpoint does not prove ongoing DB readiness

- Maps to: `PVN-032`
- Area: Operations
- Evidence: `/health/ready` currently checks non-nil injected AuthService/AuthStore and MFA key length, not an active DB/schema check.
- Risk: process may report ready after DB connectivity/schema becomes unhealthy.
- Done: bounded live DB dependency/schema probe plus failure-path tests and no secret leakage.

### BUG-004 [P1] Recovery codes are not accepted in login flow

- Maps to: `PVN-033`
- Area: MFA
- Evidence: recovery codes are generated and can be consumed for MFA removal, while `LoginInput` and login HTTP payload only contain `TOTPCode`.
- Risk: account recovery semantics are incomplete/ambiguous.
- Done: explicit product decision. If recovery-code login is supported, one-time consumption/replay/session/audit tests must pass; otherwise docs/UI must clearly state its narrower purpose.

## Security gaps

### SECURITY-001 [P0] HTTP/IP-aware login rate limiting and progressive delay absent

- Maps to: `PVN-034`
- Current mitigation: DB actor lockout after repeated failures.
- Missing: bounded request-rate control/trusted reverse-proxy client-IP policy/progressive delay at HTTP edge.
- Done: tested anti-abuse layer without making the data plane depend on panel availability.

### SECURITY-002 [P0] Supply-chain/release security gates absent

- Maps to: `PVN-065`
- Missing today: SBOM, dependency vulnerability audit, SAST, secret scan, signed release/provenance policy.
- Current CI does Go/Web/PG18/rehearsal/bundle but these security release gates are not present.

### SECURITY-003 [P1] S04 public-exposure security review not formally closed

- Maps to: `PVN-035`, `PVN-036`
- Public panel/API is live and prior Caddy invariants passed, but formal S04 stage closure still requires independent external security/postflight evidence after remaining auth blockers are handled.

## Technical debt

### TECH-DEBT-001 [P0] `s04-auth` and `main` are substantially diverged

- Maps to: `PVN-070`
- Audit state: `s04-auth` 123 commits ahead and 37 commits behind `main`, merge base `d0398cd1...`.
- Why it matters: `main` contains newer production evidence, while feature implementation lives on `s04-auth`.
- Rule: do not reset or force-push. Reconcile only at a fully green checkpoint with diff/review and preservation of production evidence.

### TECH-DEBT-002 [P1] Documentation truth drift

- Maps to: `PVN-069`
- `README.fa.md` still says Auth/PostgreSQL are not implemented.
- `SECURITY.md` still labels the project as scaffold.
- `docs/PRODUCT_GAPS_FA.md` still says Auth/DB/verified CI are absent.
- `docs/API_FA.md` mixes target routes with actual implementation.
- old `docs/FEATURE_MATRIX_FA.md` says multi-node is MVP while current architecture is standalone-first.
- Done: actual/current vs target/future status is explicit everywhere.

### TECH-DEBT-003 [P1] Future route registry is much larger than actual implementation

- Maps to: `PVN-069`, `PVN-071`
- `internal/httpapi/routes.go` intentionally lists many future routes with `Ready=false`; server currently implements health/auth/me/session/MFA only and returns 501 elsewhere.
- Risk: docs/UI/readers may mistake registration for implementation.
- Rule: feature matrices count implementation + tests, not route declarations.

### TECH-DEBT-004 [P1] Protocol abstraction exists but real Naive adapter is not implemented

- Maps to: `PVN-051`
- `internal/protocol` defines capability and adapter interfaces only.
- Rule: capability flags must remain honest until PoCs prove accounting/session/speed/concurrency/device behavior.

## Test gaps

### TEST-001 [P1] End-to-end authorization/IDOR/fuzz coverage incomplete

- Maps to: `PVN-071`
- Need: route readiness coverage, role/tenant matrix, strict JSON/size limits, parser/property/fuzz failure paths, browser E2E for critical flows.

### TEST-002 [P0] Multiple Naive `basic_auth` credentials not yet proven against exact custom Caddy

- Maps to: `PVN-021`, `PVN-027`, `PVN-028`
- Must be proven with the exact installed `forward_proxy` Caddy build before production ownership/mutation.
- Do not infer syntax support from generic documentation.

### TEST-003 [P0] Exact Naive per-credential accounting remains unproven

- Maps to: `PVN-045`…`PVN-048`
- Until measured, billing/quota/session UI must not claim exact values.

## Deployment / release gaps

### DEPLOY-001 [P0] Formal S04 ledger not closed

- Maps to: `PVN-036`
- Implemented milestones/public preview are real, but official `S04=PASSED` must wait for the independent final gate defined in the roadmap.

### DEPLOY-002 [P0] S04R migration 0003 is development-only

- Maps to: `PVN-027`…`PVN-029`
- Production schema remains v2.
- Do not apply migration 0003 merely because DB CI is green; full S04R implementation/rehearsal/read-only live preflight must pass first.

### DEPLOY-003 [P0] No final fresh-install/upgrade production release path yet

- Maps to: `PVN-058`…`PVN-067`
- Existing stage scripts are valuable foundations, not the final general-purpose installer/release lifecycle.

## Legal / source-use gap

### LEGAL-001 [P1] PVNaive has no explicit repository license policy recorded

- Maps to: `PVN-072`
- Repo audit did not find a root LICENSE file.
- Competitor references include GPL/AGPL projects; their code must not be blindly copied.
- Until owner chooses policy, use concepts/architecture and original clean-room implementation; preserve any required attribution for permissive sources actually reused.

## Closed historical failures

These are not current blockers but remain in evidence/history:

- missing `file` package during first S04 live attempt — fixed;
- same-second encrypted backup destination collision — regression added/fixed;
- S04 API DB environment/startup mismatch — subsequently fixed and localhost/public deployment passed;
- S04 migration remained v2 after safe rollback refusal — later recovery/backup/postflight completed.

Do not reopen a closed historical failure unless a regression reproduces it.
