# PVNaive — Canonical Project Status

Last updated: 2026-09-15 19:39 (Asia/Tehran)

## Current verified GitHub truth

- Current verified `main`: `51e5084d78fba6866258aee69f7a0418b038b1c7`; exact-main GitHub CI `34988909834` is terminal SUCCESS.
- New validated work since the 12:37 checkpoint: `4b366cda518bc4a8ed636a9beda727a5b4d6560c` adds the illustrated node setup tutorial and is CI-green; `51e5084d78fba6866258aee69f7a0418b038b1c7` changes subscription `Content-Disposition` from `attachment` to `inline` so browser-opened subscription text renders instead of forced download, while retaining the filename parameter for clients. Its exact-main CI is green.
- The subscription fix commit records focused `go test ./internal/httpapi ./internal/subscription` green; repository CI is the exact-head gate.
- Task13 #108 remains merged/accepted with target-only HTTP/1.1+HTTP/2 session kill and exactly-once final accounting. R8 monitoring/race fixes remain integrated.
- #101 Karing is the only open PR. It remains DRAFT on stale head `6691392639be9fae4656a861db4a6d16580f850d`; fresh GitHub inspection reports `mergeable=false`. Do NOT merge until reconstructed/reconciled from current main, exact-head gates pass, and real Karing import → parse → CONNECT → cleanup/revoke evidence exists.
- #114 basic two-node real mTLS pull/heartbeat/drift proof is retained; certificate overlap/rotation + old-cert retirement or explicit revocation + replay/fail-closed lifecycle proof remains open.

## Production truth ceiling

- Persistent verified Production ceiling remains schema **33** / `repo-fin2` checkpoint.
- Fresh device inventory at 19:39 shows one online execution worker named `Pak-Nasheeee-haaaaaaaaa` and one stale duplicate offline. No trusted `PVNaive-Production-Primary` is connected.
- Do not infer fresh Production image/SHA, schema ledger, encrypted-backup freshness, disk, Caddy/services or rollback state. No Production mutation until trusted Primary reconnect + read-only audit + fresh encrypted backup + independent rollback snapshot.

## Active roadmap and worker allocation

1. W4: #101/#120 reconstruct from exact current main and obtain real Karing acceptance; preserve newer subscription/runtime behavior; then disposable R5/R6 E2E.
2. W3: #114 RED-first certificate overlap/rotation/revocation lifecycle; then STEER-006 scale/integration.
3. W1: independent PKI/client-truth/RBAC review; verify fail-closed identity, revocation semantics and compatibility claims.
4. W2: registry/replay monotonicity assertions; then R8 ledger/per-node/per-user truthful projections.
5. Coordinator: integrate only exact-head validated work; keep migrations forward-only and Production backup/rollback gates mandatory.

## Safety invariants

- Accounting/session/quota semantics remain canonical truth; telemetry/UI/control work cannot weaken them.
- Task13 kills only the selected session and preserves credential/sibling sessions.
- Fleet identity is TLS-client-cert authoritative; forwarded/client headers never override it.
- Applied migrations are immutable; future DB changes are forward-only and ledger-checked.
- Missing telemetry remains Unknown, never fabricated zero/health.
- Client compatibility claims require real-client evidence.
- Production promotion order: exact-head CI → disposable rehearsal → trusted read-only audit → fresh encrypted backup + independent rollback snapshot → staged promotion → postflight → retain rollback.

## Coordinator checkpoint — 2026-09-15 19:39 Asia/Tehran

- Re-inspected current main, CI, the only open PR, Production lane and fresh connected-device inventory.
- Reconciled two new main commits and accepted them only with terminal exact-head CI evidence; no PR merge was performed.
- Refreshed #101, #114 and #100 with the exact current baseline and next worker assignments.
- Production remained untouched: no deploy, migration, restart/reload, Caddy/DB/credential/backup/rollback mutation.