# Continue Here — PVNaive

Verified checkpoint: 2026-09-15 09:42 Asia/Tehran

## GitHub truth

- Verified pre-docs main: `d85fbace33efc09299a656692adc0978543d0d81`; CI `34927848620` SUCCESS.
- Fresh exact-main worker validation: gofmt clean, `go vet ./...` PASS, `go test ./...` PASS, Web 23/23 files / 117/117 tests PASS, production build PASS.
- #101 is the only open PR: DRAFT, stale head `669139263...`, currently mergeable by GitHub mechanics but acceptance-blocked. Reconcile minimally to current main, rerun exact-head gates, then real Karing import → parse → CONNECT → cleanup/revoke. Never reuse historical CI for merge.
- #114: basic real two-node mTLS pull proof is retained; remaining gate is certificate overlap/rotation + explicit revocation/replay/fail-closed lifecycle proof.

## Production truth

- Persistent ceiling: schema 33 / repo-fin2.
- One execution worker is online; duplicate registration offline. No trusted `PVNaive-Production-Primary` is connected.
- Do not infer fresh Production image/schema/backup/disk/Caddy/rollback state. Do not mutate Production.

## Execute next

1. W4 — #101/#120 real Karing exact-main acceptance, then disposable R5/R6 E2E.
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

Current main/PR/CI/device state and persistent worker reports were inspected. The old Task36 security report is partial historical evidence, not a new completion. #101 and #114 instructions were refreshed in GitHub. Production remained untouched.
