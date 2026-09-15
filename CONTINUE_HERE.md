# Continue Here — PVNaive

Verified checkpoint: 2026-09-15 12:37 Asia/Tehran

## GitHub truth

- Current verified main: `8209701f1b4561b190b423b7e9d38bd3baf15ced`; CI `34945421993` SUCCESS.
- Latest code validation: gofmt clean, `go vet ./...` PASS, `go test ./...` PASS, Web 23/23 files / 117/117 tests PASS, production build PASS. Newer commits are documentation-only.
- #101 is the only open PR: DRAFT, stale head `669139263...`, fresh GitHub inspection says `mergeable=false`. Reconstruct/reconcile from exact current main, rerun exact-head CI/accounting/pinned-forwardproxy, then real Karing import → parse → CONNECT → cleanup/revoke. Never reuse historical CI for merge.
- #114: basic real two-node mTLS pull proof is retained; remaining gate is certificate overlap/rotation + explicit revocation/replay/fail-closed lifecycle proof.

## Production truth

- Persistent ceiling: schema 33 / repo-fin2.
- Fresh inventory: one execution worker online; duplicate registration offline. No trusted `PVNaive-Production-Primary` is connected.
- Do not infer fresh Production image/schema/backup/disk/Caddy/rollback state. Do not mutate Production.

## Execute next

1. W4 — #101/#120 reconstruct from exact current main and obtain real Karing acceptance, then disposable R5/R6 E2E.
2. W3 — #114 cert overlap/rotation/revocation, then STEER-006.
3. W1 — independent PKI/client-truth/RBAC review, especially fail-closed revocation and TLS identity authority.
4. W2 — replay/registry monotonicity tests, then R8 ledger/per-node/per-user truthful projections.
5. On trusted Production Primary reconnect only: read-only identity/SHA/image/schema/services/Caddy/backup/disk/rollback audit → fresh encrypted backup + independent rollback snapshot → staged promotion/postflight.

## Invariants

- Exact accounting/session/quota truth is never inferred from telemetry.
- Task13 is selected-session-only and preserves credentials/siblings.
- Missing telemetry stays Unknown.
- TLS client certificate is authoritative fleet identity.
- Applied migrations immutable; future changes forward-only.
- Client compatibility requires real-client evidence.

## Latest actions

At 11:38, current main/PR/CI/device state was re-inspected. CI `34941672566` is now terminal SUCCESS. #101 is mechanically non-mergeable on its stale head and was re-dispatched to W4 from exact current main. #114 lifecycle work was re-dispatched across W1-W4. #100 records the still-disconnected Production Primary. Production remained untouched.
