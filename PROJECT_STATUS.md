# PVNaive — Canonical Project Status

Last updated: 2026-09-15 09:42 (Asia/Tehran)

## Current verified GitHub truth

- Current `main`: `d85fbace33efc09299a656692adc0978543d0d81`; GitHub CI `34927848620` is terminal SUCCESS.
- Fresh execution-worker validation on exact current main: gofmt clean, `go vet ./...` PASS, `go test ./...` PASS, Web 23/23 files / 117/117 tests PASS, production build PASS.
- Task13 #108 remains merged/accepted with target-only HTTP/1.1+HTTP/2 session kill and exactly-once final accounting.
- R8 monitoring/race fixes are integrated.
- #101 Karing is the only open PR. It remains DRAFT on stale head `6691392639be9fae4656a861db4a6d16580f850d`; it is currently GitHub-mergeable but must NOT merge until reconciled to current main, exact-head gates pass, and real Karing import → parse → CONNECT → cleanup/revoke evidence exists.
- #114 basic two-node mTLS pull/heartbeat/drift proof is retained; certificate overlap/rotation + explicit revocation/replay/fail-closed lifecycle proof remains open.

## Production truth ceiling

- Persistent verified Production ceiling remains schema **33** / `repo-fin2` checkpoint.
- Fresh device inventory shows one online execution worker named `Pak-Nasheeee-haaaaaaaaa` and one stale duplicate offline. No trusted `PVNaive-Production-Primary` is connected.
- Do not infer fresh Production image/SHA, schema ledger, encrypted-backup freshness, disk, Caddy/services or rollback state. No Production mutation until trusted Primary reconnect + read-only audit + fresh encrypted backup + independent rollback snapshot.

## Active roadmap and worker allocation

1. W4: #101/#120 real Karing exact-main acceptance; then disposable R5/R6 E2E.
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

## Coordinator checkpoint — 2026-09-15 09:42 Asia/Tehran

- `main` and CI are green as stated above; independent Go/Web/build gates were rerun successfully on the online execution worker.
- Persistent worker reports were inspected; older Task36 security report is historical/partial and not treated as fresh completion evidence.
- #101 and #114 worker instructions were refreshed in GitHub with exact current-main requirements and no-Production constraints.
- Production remains untouched because the trusted Primary is disconnected.
