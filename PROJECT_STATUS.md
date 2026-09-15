# PVNaive — Canonical Project Status

Last updated: 2026-09-15 12:37 (Asia/Tehran)

## Current verified GitHub truth

- Current verified `main`: `8209701f1b4561b190b423b7e9d38bd3baf15ced`; GitHub CI `34945421993` is terminal SUCCESS.
- Previous independent execution-worker validation remains: gofmt clean, `go vet ./...` PASS, `go test ./...` PASS, Web 23/23 files / 117/117 tests PASS, production build PASS. No newer code commit exists after that validation; latest commits are documentation-only.
- Task13 #108 remains merged/accepted with target-only HTTP/1.1+HTTP/2 session kill and exactly-once final accounting.
- R8 monitoring/race fixes are integrated.
- #101 Karing is the only open PR. It remains DRAFT on stale head `6691392639be9fae4656a861db4a6d16580f850d`; fresh GitHub inspection reports `mergeable=false`. Do NOT merge until reconstructed/reconciled from current main, exact-head gates pass, and real Karing import → parse → CONNECT → cleanup/revoke evidence exists.
- #114 basic two-node mTLS pull/heartbeat/drift proof is retained; certificate overlap/rotation + explicit revocation/replay/fail-closed lifecycle proof remains open.

## Production truth ceiling

- Persistent verified Production ceiling remains schema **33** / `repo-fin2` checkpoint.
- Fresh device inventory at 11:38 shows one online execution worker named `Pak-Nasheeee-haaaaaaaaa` and one stale duplicate offline. No trusted `PVNaive-Production-Primary` is connected.
- Do not infer fresh Production image/SHA, schema ledger, encrypted-backup freshness, disk, Caddy/services or rollback state. No Production mutation until trusted Primary reconnect + read-only audit + fresh encrypted backup + independent rollback snapshot.

## Active roadmap and worker allocation

1. W4: #101/#120 reconstruct from exact current main and obtain real Karing acceptance; then disposable R5/R6 E2E.
2. W3: #114 certificate overlap/rotation/revocation; then STEER-006 scale/integration.
3. W1: independent PKI/client-truth/RBAC review; verify fail-closed identity and revocation semantics.
4. W2: replay/registry monotonicity assertions; then R8 ledger/per-node/per-user truthful projections.
5. Coordinator: integrate only exact-head validated work; keep migrations forward-only and Production backup/rollback gates mandatory.

## Safety invariants

- Accounting/session/quota semantics remain canonical truth; telemetry/UI/control work cannot weaken them.
- Task13 kills only the selected session and preserves credential/sibling sessions.
- Fleet identity is TLS-client-cert authoritative; forwarded/client headers never override it.
- Applied migrations are immutable; future DB changes are forward-only and ledger-checked.
- Missing telemetry remains Unknown, never fabricated zero/health.
- Client compatibility claims require real-client evidence.
- Production promotion order: exact-head CI → disposable rehearsal → trusted read-only audit → fresh encrypted backup + independent rollback snapshot → staged promotion → postflight → retain rollback.

## Coordinator checkpoint — 2026-09-15 12:37 Asia/Tehran

- Re-inspected current main, CI, the only open PR, Production lane and fresh connected-device inventory.
- CI `34945421993` for exact main `8209701f...` is terminal SUCCESS; independent exact-main Go/Web/build rerun is also GREEN.
- #101 is now mechanically non-mergeable on its stale head and remains acceptance-blocked; W4 instructions were refreshed against exact current main.
- #114 worker allocation was refreshed for certificate lifecycle/replay/fail-closed evidence.
- #100 records the fresh disconnected-Primary state and preserves schema 33 / repo-fin2 as the Production truth ceiling.
- Production remained untouched: no deploy, migration, restart/reload, Caddy/DB/credential/backup/rollback mutation.
