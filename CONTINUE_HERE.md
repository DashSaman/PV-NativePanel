# Continue Here — PVNaive

Verified checkpoint: 2026-09-15 (Asia/Tehran)

## GitHub truth

- Current code main before docs refresh: `4886b2730515dc6acac1e4a8d6eeb7b13067e2a5`. R10 follow-up UI commits `a715083...` and `c9d56a9...` are terminal CI green (`34908530057`, `34908840088`); coordinator also reran Web 23/23 files, 115/115 tests and production build. `4886b27...` is a one-line unused-import cleanup with the same local Web gates green; wait for its exact-head GitHub CI before promotion.
- R10 is not Production-approved merely because CI is passing: named-client compatibility is a separate truth gate in #120. #101 Karing remains DRAFT/non-mergeable and must be reconciled to latest verified main before a real import → parse → CONNECT → cleanup/revoke acceptance.
- Task13 #108 is merged and accepted with real HTTP/1.1+HTTP/2 target-only kill plus exactly-once final accounting.
- R8 monitoring/race fixes are integrated.
- R5 registry and real basic two-node mTLS pull/heartbeat/drift E2E are green; #114 remains for cert rotation/overlap + explicit revocation/replay/fail-closed lifecycle proof.

## Production truth

- Persistent verified ceiling: schema 33 / repo-fin2 checkpoint.
- One execution-worker Remote Desktop registration is online and one duplicate is offline; no trusted Production Primary is connected.
- Do not infer fresh image/schema/backup/disk/rollback/Caddy state and do not mutate Production.

## Execute next

1. #120/#101: validate every advertised client claim. Real Karing exact-main acceptance first; remove or qualify any unverified client claims rather than presenting them as supported.
2. #114: implement/prove certificate overlap/rotation + explicit revocation/replay/fail-closed behavior; TLS client cert remains authoritative.
3. R8: ledger reconciliation/per-node/per-user projections and stream-RBAC isolation with Unknown gaps preserved.
4. #115: disposable default-OFF cover/persona/probe-sweep/failure rehearsal.
5. Production Primary reconnect: read-only identity/SHA/schema/services/Caddy/backup/disk/rollback audit first; only then fresh encrypted backup + independent rollback snapshot and staged promotion.

## Worker allocation

- W1: R10 support-claim/security review → R5 PKI/revocation → R6/R8 RBAC/accessibility.
- W2: RED-first R10 truth tests → R8 ledger/projections.
- W3: R10 export/direct format semantics → R5 cert lifecycle/STEER-006.
- W4: real Karing acceptance → disposable R5/R6 E2E.

One registered execution worker is online. Keep the queue in GitHub and execute only with receipts; do not treat that worker as Production.

## Invariants

- Exact accounting/session/quota truth is never inferred from network telemetry.
- Task13 kills only the selected session and preserves credentials/siblings.
- Missing telemetry remains Unknown.
- Applied migrations are immutable and future migrations are forward-only.
- Client compatibility claims are evidence-backed.
- Production promotion order: exact-head CI → disposable rehearsal → trusted audit → fresh encrypted backup + rollback snapshot → staged promotion → postflight → retain rollback.

## Coordinator checkpoint — 2026-09-15

- Current code head: `edfae4efc7e5da7f714757fe74804901f24a9d13`. The gofmt regression is repaired; the follow-on 0034 rollback defects (destructive marker + schema ledger removal) are repaired and migration checksums refreshed.
- Disposable PostgreSQL 18 `tests/db/migration_test.sh` passes on the exact checkout. Exact-head GitHub CI `34920505396` is terminal SUCCESS; docs-tip CI `34920603677` is also terminal SUCCESS. The internal #121 CI/0034 recovery gate is cleared.
- Production remains untouched and blocked on a trusted `PVNaive-Production-Primary` reconnect plus read-only audit, fresh encrypted backup and independent rollback snapshot.
- #101 still requires real Karing import → parse → CONNECT → cleanup/revoke.

## Latest verified continuation point — 2026-09-15 07:39 Asia/Tehran

Current main is `f437f352b855775a5ba736f26595ff932a0f510f` with CI `34920603677` SUCCESS. Exact code head `edfae4efc7e5da7f714757fe74804901f24a9d13` has CI `34920505396` SUCCESS. Worker-side current-main Web suite is 23/23 files, 117/117 tests PASS and build PASS. Production Primary is still disconnected; do not deploy until read-only audit + fresh encrypted backup + independent rollback snapshot. Advance #120/#101, #114, #115 and R8 independently.
