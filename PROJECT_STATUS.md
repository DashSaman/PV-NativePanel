# PVNaive — Canonical Project Status

Last updated: 2026-09-16 22:42 (Asia/Tehran)

## Current verified GitHub truth

- Verified baseline before this documentation refresh: `a7ac95fd00aba1cde450ba92cac3d6fbc75dbe01`; exact-head push CI run `35062589782` is terminal SUCCESS.
- This checkpoint also refreshed `HANDOFF.md` (`76611a4965f93909fc253f68297c91abc0e3c74b`) and `CONTINUE_HERE.md` (`acfd4919fcc790c3f0d6ec4f13c59bb8cc79e194`) before this status commit.
- Code parent `9891fdc650e4e65c6bdc91c5b4bd3719c507eb2c` serves Karing a direct Clash-compatible Naive profile. Do not revert to rejected sing-box or intermediate base64 Karing approaches.
- Fresh independent clean-clone validation on the only online execution worker: Web 23/23 files, 117/117 tests PASS; TypeScript/Vite production build PASS. Go toolchain is unavailable on that worker, so exact-head GitHub CI remains the Go gate.
- #101 remains the only open PR and its stale/DRAFT sing-box implementation is superseded; it is not merge-eligible. Real Karing import/update -> CONNECT -> cleanup/revoke on current main remains required before closing/superseding it.
- #114 retains the basic real two-node mTLS proof; certificate overlap/rotation, old-cert retirement or explicit revocation, replay and fail-closed lifecycle proof remain open.

## Production truth ceiling

- Persistent verified Production ceiling remains schema **33** / `repo-fin2`.
- Fresh device inventory at 22:42 shows one online execution worker `Pak-Nasheeee-haaaaaaaaa` and one stale duplicate offline. No trusted `PVNaive-Production-Primary` is connected.
- No Production deploy/migration/restart/reload/Caddy/DB/credential/backup/rollback mutation is permitted until trusted Primary reconnect + read-only audit + fresh encrypted backup + independent rollback snapshot.

## Active roadmap and worker allocation

1. W1/current available execution lane: current-main client-claim/security/dependency review and exact-head verification; no forced dependency upgrades without compatibility evidence.
2. W2 when available: capability/content truth, registry/replay monotonicity and accounting/session invariants; then R8 ledger/per-node/per-user truthful projections.
3. W3 when available: #114 RED-first certificate overlap/rotation/revocation lifecycle; then STEER-006 scale/integration.
4. W4 when available: real disposable Karing import/update -> CONNECT -> cleanup/revoke on exact current main; then disposable R5/R6 lifecycle E2E.
5. Coordinator: integrate only exact-head validated work; keep migrations forward-only and Production backup/rollback gates mandatory.

## Safety invariants

- Accounting/session/quota/credential semantics remain canonical truth; telemetry/UI/control work cannot weaken them.
- Task13 kills only the selected session and preserves credential/sibling sessions.
- Fleet identity is TLS-client-cert authoritative; forwarded/client headers never override it.
- Applied migrations are immutable; future DB changes are forward-only and ledger-checked.
- Missing telemetry remains Unknown, never fabricated zero/health.
- Client compatibility claims require real-client evidence.
- Production promotion order: exact-head CI -> disposable rehearsal -> trusted read-only audit -> fresh encrypted backup + independent rollback snapshot -> staged promotion -> postflight -> retain rollback.
