# Continue Here — PVNaive

Verified checkpoint: 2026-09-15 19:39 Asia/Tehran

## GitHub truth

- Verified code main before docs refresh: `51e5084d78fba6866258aee69f7a0418b038b1c7`; exact-main CI `34988909834` SUCCESS.
- New validated work since the prior checkpoint: illustrated node tutorial `4b366cda...` and subscription browser-delivery fix `51e5084d...`; both exact-head CI-green. The subscription endpoint now uses inline Content-Disposition rather than forced attachment while retaining the filename parameter.
- #101 is the only open PR: DRAFT, stale head `669139263...`, fresh GitHub inspection says `mergeable=false`. Reconstruct/reconcile from exact current main, preserve newer subscription/runtime behavior, rerun exact-head CI/accounting/pinned-forwardproxy, then real Karing import → parse → CONNECT → cleanup/revoke. Never reuse historical CI for merge.
- #114: basic real two-node mTLS pull proof is retained; remaining gate is certificate overlap/rotation + old-cert retirement or explicit revocation + replay/fail-closed lifecycle proof.

## Production truth

- Persistent ceiling: schema 33 / repo-fin2.
- Fresh 19:39 inventory: one execution worker online; duplicate registration offline. No trusted `PVNaive-Production-Primary` is connected.
- Do not infer fresh Production image/schema/backup/disk/Caddy/rollback state. Do not mutate Production.

## Execute next

1. W4 — #101/#120 reconstruct from exact current main and obtain real Karing acceptance, then disposable R5/R6 E2E.
2. W3 — #114 RED-first cert overlap/rotation/revocation lifecycle, then STEER-006.
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

At 19:39, current main/open PR/CI/device state was re-inspected. Newly landed main work was reconciled against terminal exact-head CI. #101, #114 and #100 were refreshed with exact current state and next worker assignments. Canonical status/handoff/continuation files were advanced. Production remained untouched because the trusted Primary is still disconnected.
