# Continue Here — PVNaive

Verified checkpoint: 2026-09-15 (Asia/Tehran)

## GitHub truth

- Validated code main before this docs checkpoint: `3eee556b908ec98c7a033de0ce63938b7dc0e714`. R8 race issue #118 / PR #119 is merged; exact PR head `e983468bb80fbda5dc9509aa1e4b89b96431b059` had all required GitHub gates green. Post-merge CI `34904000208` completed SUCCESS on `3eee556b908ec98c7a033de0ce63938b7dc0e714`.
- Task13 #108 is merged. Final exact head `1f65eccb6d71572f3ff4f15e942cac02e7bfa6c7` passed CI `34897871371`, Exact Accounting `34897871364`, Pinned Forwardproxy `34897871355`, plus the real HTTP/1.1+HTTP/2 target-session kill rehearsal.
- R8 chart issue #116 is closed completed on main; obsolete/conflicting PR #117 is closed unmerged.
- #101 Karing remains blocked on real-client acceptance.
- R5 registry DB gate and a real disposable two-node mTLS pull/heartbeat/drift E2E both passed on current main. #114 remains open for certificate rotation/overlap + explicit revocation proof; Production enablement stays gated.

## Production truth

- Persistent verified ceiling: schema 33 / repo-fin2 with recorded healthy panel/API/SSE/real-customer CONNECT/accounting postflight and retained rollback/backup evidence.
- No trusted Production Primary is connected. Do not infer fresh image/schema/backup/disk/rollback state and do not mutate Production.

## Execute next

1. R5: registry + real two-node mTLS pull/heartbeat/drift are green; next prove certificate rotation/overlap + explicit revocation and STEER-006 scale. TLS client cert remains authoritative.
2. R8: RED-first UI-002 ledger reconciliation/per-node/per-user projections and stream-RBAC isolation; preserve Unknown gaps and bounded histories.
3. R6: disposable default-OFF cover/persona/probe-sweep/feed-failure rehearsal; do not promote to Production yet.
4. #101: perform real disposable Karing import → parse → CONNECT → cleanup/revoke when a real client host is available.
5. Production Primary reconnect: first action is read-only identity/SHA/schema/service/Caddy/backup/disk/rollback audit. Only after fresh encrypted backup and independent rollback snapshot may a staged deploy be considered.

## Worker allocation

- W1: independent security/accounting/accessibility review across Task13/R5/R6/R8.
- W2: R8 ledger reconciliation and truthful projections.
- W3: R5 mTLS/STEER-006 + R8 RBAC stream tests; prioritize real disposable listener/client-cert integration over duplicate unit coverage.
- W4: disposable R5/R6 E2E + Karing real-client acceptance.

## Invariants

- Exact accounting/session/quota truth is never inferred from network telemetry.
- Task13 kills only the selected session and preserves credentials/siblings.
- Missing telemetry remains Unknown.
- Applied migrations are immutable and future migrations are forward-only.
- Production promotion order: exact-head CI → disposable rehearsal → trusted audit → fresh encrypted backup + rollback snapshot → staged promotion → postflight → retain rollback.
