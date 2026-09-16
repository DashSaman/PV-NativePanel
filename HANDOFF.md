# PVNaive Handoff

Checkpoint: 2026-09-16 22:42 Asia/Tehran

## Verified baseline

- Exact code/docs baseline before this refresh: `a7ac95fd00aba1cde450ba92cac3d6fbc75dbe01`; push CI `35062589782` is terminal SUCCESS.
- Code parent `9891fdc650e4e65c6bdc91c5b4bd3719c507eb2c` serves Karing a direct Clash-compatible Naive profile. Do not revert to the rejected sing-box or intermediate base64 Karing paths.
- Fresh clean-clone validation on the only online execution worker: Web 23/23 files, 117/117 tests PASS; production Vite/TypeScript build PASS. Go toolchain is unavailable on that worker, so exact-head GitHub CI remains the Go gate.
- #101 is the only open PR and remains stale/DRAFT; its sing-box implementation is superseded and must not merge. Real Karing import/update -> CONNECT -> cleanup/revoke on current main is still required before superseding/closing it.
- #114 retains the real basic two-node mTLS proof. Remaining acceptance is certificate overlap/rotation, old-cert retirement or explicit revocation, replay and fail-closed lifecycle proof.

## Production blocker

- Fresh device inventory at 22:42: one online execution worker `Pak-Nasheeee-haaaaaaaaa`, one stale duplicate offline; no trusted `PVNaive-Production-Primary`.
- Production truth ceiling remains schema 33 / repo-fin2. Do not infer fresh image/schema/backup/disk/Caddy/rollback state.
- No Production mutation until Primary reconnect -> read-only audit -> fresh encrypted backup + independent rollback snapshot -> staged promotion/postflight.

## Worker queue

- W1 / currently available execution lane: current-main claim/security/dependency review and exact-head validation; do not force dependency upgrades without compatibility evidence.
- W2 when available: capability/content truth plus registry/replay monotonicity and accounting/session invariants; then R8 truthful projections.
- W3 when available: #114 RED-first certificate overlap/rotation/revocation mechanism and tests; then STEER-006.
- W4 when available: real disposable Karing acceptance on exact current main; then disposable R5/R6 lifecycle E2E.
- Coordinator: integrate only exact-head validated work; migrations forward-only; never substitute historical CI or inferred Production state.

## Locked invariants

- Accounting/session/quota/credential semantics remain canonical truth.
- Task13 kills only the selected session and preserves credential/sibling sessions.
- Missing telemetry is Unknown, never fabricated zero/health.
- TLS client certificate is authoritative fleet identity; headers cannot override it.
- Applied migrations are immutable; future DB changes are forward-only and ledger-checked.
- Client compatibility claims require real-client evidence.
