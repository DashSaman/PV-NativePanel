# PVNaive Handoff

Checkpoint: 2026-09-15 09:42 Asia/Tehran

## Verified baseline

- Verified pre-docs main: `d85fbace33efc09299a656692adc0978543d0d81`; CI `34927848620` SUCCESS.
- Independent exact-main worker gates: gofmt clean; `go vet ./...` PASS; `go test ./...` PASS; Web 23/23 files / 117/117 tests PASS; production build PASS.
- Task13 #108 remains accepted with target-only HTTP/1.1+HTTP/2 kill and exactly-once accounting.
- R8 monitoring/race fixes are integrated.
- #101 Karing remains DRAFT on stale `669139263...`; GitHub currently reports it mergeable, but merge is forbidden until current-main reconciliation + fresh exact-head gates + real Karing import/parse/CONNECT/cleanup-revoke acceptance.
- #114 retains basic real two-node mTLS proof; cert overlap/rotation/revocation/replay/fail-closed remains the acceptance gap.

## Production blocker

- Fresh device inventory: one online execution worker `Pak-Nasheeee-haaaaaaaaa`, one duplicate offline; no trusted `PVNaive-Production-Primary`.
- Production truth ceiling remains schema 33 / repo-fin2. Do not claim fresh image/schema/backup/disk/Caddy/rollback state.
- On Primary reconnect: read-only identity/SHA/image/schema/services/listeners/Caddy/backup/disk/rollback audit first; then fresh encrypted backup + independent rollback snapshot; only then staged promotion.

## Worker queue

- W1: independent client-truth + PKI/RBAC review; focus fail-closed revocation and TLS identity authority.
- W2: registry/replay monotonicity assertions; then R8 ledger/per-node/per-user truth.
- W3: #114 certificate overlap/rotation/revocation, then STEER-006.
- W4: #101 real Karing exact-main acceptance, then disposable R5/R6 E2E.
- Coordinator: merge only exact-head validated work; never substitute historical CI or inferred Production state.

## Invariants

- Exact accounting/session/quota/credential semantics are locked.
- Task13 is selected-session-only, not credential revocation.
- Missing telemetry remains Unknown.
- TLS client cert is authoritative fleet identity; headers cannot override it.
- Applied migrations immutable; future DB work forward-only.
- Client compatibility requires real-client evidence.

## Latest actions

- Inspected current main, open PRs, CI, Production device inventory, and persistent worker reports.
- Historical Task36 report remains partial/stale and was not promoted to completed truth.
- Refreshed #101 and #114 worker instructions in GitHub against current main.
- Production untouched; no deploy/migration/restart/Caddy/DB/credential/backup/rollback mutation.
