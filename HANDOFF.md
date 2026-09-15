# PVNaive Handoff

Checkpoint: 2026-09-15 19:39 Asia/Tehran

## Verified baseline

- Verified code main before this docs refresh: `51e5084d78fba6866258aee69f7a0418b038b1c7`; exact-main CI `34988909834` SUCCESS.
- New validated main work: node tutorial `4b366cda...` (CI SUCCESS) and subscription browser-delivery fix `51e5084d...` (`Content-Disposition: inline`; focused subscription/httpapi tests recorded green; exact-main CI SUCCESS).
- Task13 #108 remains accepted with target-only HTTP/1.1+HTTP/2 kill and exactly-once accounting; R8 fixes remain integrated.
- #101 Karing remains DRAFT on stale `669139263...`; fresh GitHub inspection reports `mergeable=false`. Merge is forbidden until reconstruction/reconciliation from current main + fresh exact-head gates + real Karing import/parse/CONNECT/cleanup-revoke acceptance.
- #114 retains basic real two-node mTLS proof; cert overlap/rotation, old-cert retirement or explicit revocation, replay and fail-closed lifecycle proof remain the acceptance gap.

## Production blocker

- Fresh device inventory at 19:39: one online execution worker `Pak-Nasheeee-haaaaaaaaa`, one duplicate offline; no trusted `PVNaive-Production-Primary`.
- Production truth ceiling remains schema 33 / repo-fin2. Do not claim fresh image/schema/backup/disk/Caddy/rollback state.
- On Primary reconnect: read-only identity/SHA/image/schema/services/listeners/Caddy/backup/disk/rollback audit first; then fresh encrypted backup + independent rollback snapshot; only then staged promotion.

## Worker queue

- W1: independent client-truth + PKI/RBAC review; focus fail-closed revocation, TLS identity authority and advertised-client truth.
- W2: registry/replay monotonicity assertions; then R8 ledger/per-node/per-user truth.
- W3: #114 RED-first certificate overlap/rotation/revocation lifecycle, then STEER-006.
- W4: #101 reconstruct from exact current main, preserving newer subscription/runtime behavior + real Karing acceptance; then disposable R5/R6 E2E.
- Coordinator: merge only exact-head validated work; never substitute historical CI or inferred Production state.

## Invariants

- Exact accounting/session/quota/credential semantics are locked.
- Task13 is selected-session-only, not credential revocation.
- Missing telemetry remains Unknown.
- TLS client cert is authoritative fleet identity; headers cannot override it.
- Applied migrations immutable; future DB work forward-only.
- Client compatibility requires real-client evidence.

## Latest actions

- Re-inspected current main, open PRs, CI, Production lane and fresh device inventory.
- Reconciled and accepted the two newly landed main commits only after terminal exact-head CI evidence.
- Refreshed #101, #114 and #100 with current exact baseline and worker instructions.
- Updated canonical status/handoff; Production untouched: no deploy/migration/restart/Caddy/DB/credential/backup/rollback mutation.

### Fresh coordinator reconciliation — 2026-09-15 22:41 Asia/Tehran
Current exact-main is `9891fdc650e4e65c6bdc91c5b4bd3719c507eb2c`, CI `35008842524` SUCCESS. Karing now receives the direct Clash-compatible Naive profile; do not reconstruct/merge stale #101 sing-box code and do not revert to the intermediate base64 Karing path. Real Karing CONNECT + cleanup/revoke remains the acceptance gate. Production Primary remains unavailable; no Production mutation.
