# PVNaive — Canonical Project Status

Last updated: 2026-09-15 (Asia/Tehran)

## Current verified GitHub truth

- Current code main before this documentation refresh: `4886b2730515dc6acac1e4a8d6eeb7b13067e2a5`. R10 landed at `9187266...`; follow-up UI fixes `a715083...` (gold button consistency) and `c9d56a9...` (row action menu viewport positioning) both have terminal green CI (`34908530057`, `34908840088`). Coordinator independently reran Web 23/23 files, 115/115 tests and production build on `c9d56a9...`; then removed an unused `useRef` import in `4886b27...` with the same Web suite/build green. Exact-head GitHub CI for `4886b27...` must be observed before promotion.
- R10 changes the panel/subscription UX and `/s/<token>` bilingual guide. Repository gates validate rendering/build/runtime invariants, but they do not prove named third-party client compatibility. Issue #120 now tracks this truthfulness gate; no Production promotion of client-specific claims until exact real-client evidence exists.
- Task13 PR #108 is merged. Final exact head `1f65eccb6d71572f3ff4f15e942cac02e7bfa6c7` passed CI `34897871371`, Exact Accounting `34897871364`, Pinned Forwardproxy `34897871355`, plus real pinned-Caddy HTTP/1.1+HTTP/2 target-only kill acceptance with exactly-once final accounting.
- R8 issue #116 and test-race issue #118 are completed; PR #119 merged after exact-head CI/accounting/forwardproxy gates.
- Karing PR #101 remains DRAFT/unmerged and currently non-mergeable on stale head `6691392639be9fae4656a861db4a6d16580f850d`. Historical green gates must not be reused after reconciliation; final acceptance still requires real disposable Karing import → parse → CONNECT → cleanup/revoke.
- R5 disposable PostgreSQL 18 registry validation and real two-node mTLS pull/heartbeat/drift E2E are recorded PASS. #114 remains open only for certificate overlap/rotation + explicit revocation/replay/fail-closed lifecycle proof.

## Production truth ceiling

- Latest persistent verified Production checkpoint records schema **33** with the Master Upgrade Pack / `repo-fin2` generation live and previously recorded healthy panel/API/SSE/real-customer CONNECT/accounting postflight plus retained rollback/backup evidence.
- Fresh Remote Desktop inventory on 2026-09-15 shows one registration named `Pak-Nasheeee-haaaaaaaaa` online and one stale duplicate offline; the online host is an execution worker (`/root/pvnaive-orch`) and is not identified as `PVNaive-Production-Primary`.
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
- Coordinator: integrate only exact-head validated work, keep migrations forward-only, and enforce Production backup/rollback gates. One execution worker is currently online; assignments remain persisted in GitHub and may execute there, but Production work stays blocked because this host is not the trusted Primary.

## Safety invariants

- Accounting/session/quota semantics are canonical truth and must not be weakened by telemetry/UI/control work.
- Task13 kills only the selected live session; it must not revoke credentials or mutate sibling sessions.
- Never trust XFF/Forwarded/client headers for authoritative node/session identity; fleet pull identity is TLS-client-cert-only.
- Applied migrations are immutable; future DB changes are forward-only and ledger-checked.
- Unknown/unavailable telemetry stays Unknown; UI must not convert missing data into zero or fabricated health.
- Client compatibility must be evidence-backed; CI rendering tests do not equal real-client support.
- Production sequence: exact-head CI → disposable rehearsal → trusted read-only audit → fresh encrypted backup + independent rollback snapshot → staged promotion → postflight → retain rollback.

## Coordinator checkpoint — 2026-09-15

- Current code head: `edfae4efc7e5da7f714757fe74804901f24a9d13`. The gofmt regression is repaired; the follow-on 0034 rollback defects (destructive marker + schema ledger removal) are repaired and migration checksums refreshed.
- Disposable PostgreSQL 18 `tests/db/migration_test.sh` passes on the exact checkout. Exact-head GitHub CI `34920505396` is terminal SUCCESS; docs-tip CI `34920603677` is also terminal SUCCESS. The internal #121 CI/0034 recovery gate is cleared.
- Production remains untouched and blocked on a trusted `PVNaive-Production-Primary` reconnect plus read-only audit, fresh encrypted backup and independent rollback snapshot.
- #101 still requires real Karing import → parse → CONNECT → cleanup/revoke.

## Coordinator refresh — 2026-09-15 07:39 Asia/Tehran

- Current main `f437f352b855775a5ba736f26595ff932a0f510f`; CI `34920603677` SUCCESS. Exact code head `edfae4efc7e5da7f714757fe74804901f24a9d13`; CI `34920505396` SUCCESS.
- Execution worker independently reran current-main Web: 23/23 files, 117/117 tests PASS; production build PASS. Host lacks native Go toolchain; GitHub exact-head Go/DB/rehearsal CI remains the authoritative green evidence.
- Fresh device inventory still has no trusted Production Primary. No Production mutation performed.
- #120/#101 real-client truth, #114 cert lifecycle, #115 disposable R6 rehearsal, and R8 truth/RBAC remain active independent lanes.
