# PVNaive — Canonical Project Status

Last updated: 2026-09-15 (Asia/Tehran)

## Current verified GitHub truth

- Current code main: `9187266c5f4b63df7849553d450975a982f6b816` (R10 Gold/dual-QR/subscription-guide change). CI run `34907620721`: Go, Web, PostgreSQL/migration/backup-restore gates and the full S04/S04R/Task13 rehearsal steps are PASS; the workflow has not yet reached a terminal SUCCESS because downstream completion/bundle scheduling is still pending. Do not report the entire run green until GitHub records a terminal success.
- R10 changes the panel/subscription UX and `/s/<token>` bilingual guide. Repository gates validate rendering/build/runtime invariants, but they do not prove named third-party client compatibility. Issue #120 now tracks this truthfulness gate; no Production promotion of client-specific claims until exact real-client evidence exists.
- Task13 PR #108 is merged. Final exact head `1f65eccb6d71572f3ff4f15e942cac02e7bfa6c7` passed CI `34897871371`, Exact Accounting `34897871364`, Pinned Forwardproxy `34897871355`, plus real pinned-Caddy HTTP/1.1+HTTP/2 target-only kill acceptance with exactly-once final accounting.
- R8 issue #116 and test-race issue #118 are completed; PR #119 merged after exact-head CI/accounting/forwardproxy gates.
- Karing PR #101 remains DRAFT/unmerged and currently non-mergeable on stale head `6691392639be9fae4656a861db4a6d16580f850d`. Historical green gates must not be reused after reconciliation; final acceptance still requires real disposable Karing import → parse → CONNECT → cleanup/revoke.
- R5 disposable PostgreSQL 18 registry validation and real two-node mTLS pull/heartbeat/drift E2E are recorded PASS. #114 remains open only for certificate overlap/rotation + explicit revocation/replay/fail-closed lifecycle proof.

## Production truth ceiling

- Latest persistent verified Production checkpoint records schema **33** with the Master Upgrade Pack / `repo-fin2` generation live and previously recorded healthy panel/API/SSE/real-customer CONNECT/accounting postflight plus retained rollback/backup evidence.
- Fresh Remote Desktop inventory on 2026-09-15 shows both known registrations named `Pak-Nasheeee-haaaaaaaaa` offline and no identifiable `PVNaive-Production-Primary` connected.
- Therefore current Production image/container identity, schema ledger, encrypted-backup freshness, disk headroom, services/Caddy state and rollback snapshot are not freshly asserted. No Production mutation is allowed until the trusted Primary reconnects and read-only audit + backup/rollback gates pass.

## Active roadmap lanes

1. **R10 truthfulness (#120 + #101):** verify every named-client compatibility claim. Worker 4 owns real Karing acceptance; Worker 1 reviews client claims; Worker 2 adds RED-first UI/content truth tests; Worker 3 verifies subscription/direct formats against runtime semantics.
2. **R5 enablement (#113/#114):** code and basic two-node mTLS E2E exist; next prove cert overlap/rotation/revocation/replay lifecycle and STEER-006 scale. Keep Production enablement gated.
3. **R6-FLIP #115:** code is on main and default-OFF; run disposable cover/persona/probe-sweep/failure rehearsals before any promotion.
4. **R8 remaining gates:** ledger reconciliation/per-node/per-user truth, RBAC stream isolation, accessibility/RTL and bounded performance without fabricated telemetry.
5. **#100 Production lane:** on trusted Primary reconnect, read-only identity/SHA/schema/services/Caddy/backup/disk/rollback audit first.

## Worker allocation

- Worker 1: R10 named-client compatibility/security truth review; then R5 PKI/revocation and R6/R8 RBAC/accessibility review.
- Worker 2: RED-first R10 content/capability tests preserving dual-QR behavior; then R8 ledger reconciliation and truthful projections.
- Worker 3: verify R10 direct/subscription formats against server semantics; then R5 cert rotation/revocation + STEER-006 integration.
- Worker 4: real disposable Karing exact-main acceptance; then R5/R6 browser/multi-node rehearsals.
- Coordinator: integrate only exact-head validated work, keep migrations forward-only, and enforce Production backup/rollback gates. No remote execution worker is currently online, so these assignments remain queued in GitHub until capacity returns.

## Safety invariants

- Accounting/session/quota semantics are canonical truth and must not be weakened by telemetry/UI/control work.
- Task13 kills only the selected live session; it must not revoke credentials or mutate sibling sessions.
- Never trust XFF/Forwarded/client headers for authoritative node/session identity; fleet pull identity is TLS-client-cert-only.
- Applied migrations are immutable; future DB changes are forward-only and ledger-checked.
- Unknown/unavailable telemetry stays Unknown; UI must not convert missing data into zero or fabricated health.
- Client compatibility must be evidence-backed; CI rendering tests do not equal real-client support.
- Production sequence: exact-head CI → disposable rehearsal → trusted read-only audit → fresh encrypted backup + independent rollback snapshot → staged promotion → postflight → retain rollback.
