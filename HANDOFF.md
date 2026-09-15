# PVNaive Handoff

Checkpoint: 2026-09-15 Asia/Tehran

## Verified baseline

- Current code main before this documentation checkpoint: `4886b2730515dc6acac1e4a8d6eeb7b13067e2a5`. Follow-up R10 UI heads `a715083...` and `c9d56a9...` have terminal green CI (`34908530057`, `34908840088`). Coordinator independently reran Web 23/23 files, 115/115 tests plus production build on the latter; `4886b27...` only removes an unused hook import and passed the same local Web suite/build. Await exact-head GitHub CI before any promotion.
- R10 client-specific support claims are not yet acceptance-proven. Issue #120 is the new promotion gate; #101 Karing remains DRAFT/non-mergeable on stale head and still requires real disposable import → parse → CONNECT → cleanup/revoke after reconciliation to latest verified main.
- Task13 #108 is merged after exact-head CI/accounting/forwardproxy plus real HTTP/1.1+HTTP/2 target-only kill proof; exact accounting/session/credential invariants remain locked.
- R8 live monitoring is integrated and its SSE race regression is fixed/closed.
- R5 registry and basic real two-node mTLS pull/heartbeat/drift E2E are green. #114 stays OPEN for certificate overlap/rotation + explicit revocation/replay/fail-closed proof.

## Production blocker

- Fresh remote inventory shows one Remote Desktop execution-worker registration online and one duplicate offline; no trusted `PVNaive-Production-Primary` is connected.
- Production truth ceiling remains the last persistent schema-33 / repo-fin2 checkpoint. Do not claim fresh image/schema/backup/disk/Caddy/rollback state.
- On trusted Primary reconnect: read-only host identity → deployed SHA/image → schema ledger → services/listeners/Caddy → encrypted-backup freshness → disk → rollback snapshot. Only then may fresh backup/snapshot and staged promotion be considered.

## Remaining acceptance blockers

- #120/#101: verify client compatibility truth; real Karing exact-main acceptance is non-substitutable.
- #114: cert lifecycle rotation/overlap/revocation/replay/fail-closed proof.
- #115: disposable R6 cover/persona/probe-sweep/failure rehearsal; Production remains default-OFF/gated.
- Production: trusted Primary reconnect plus audit/backup/rollback gates.

## Worker queue

- Worker 1 — R10 client-claim/security truth review; then R5 PKI/revocation and R6/R8 RBAC/accessibility.
- Worker 2 — RED-first R10 UI/content capability tests preserving dual-QR; then R8 ledger reconciliation/projections.
- Worker 3 — verify R10 subscription/direct format semantics; then R5 cert lifecycle + STEER-006.
- Worker 4 — real disposable Karing acceptance; then R5/R6 browser/multi-node E2E.
- Coordinator — integrate exact-head validated work only. One execution-worker registration is online; assignments remain persisted in GitHub. Do not treat that worker as Production or fabricate task completion without receipts.

## Invariants

- Preserve exact accounting, session, quota and credential semantics.
- Task13 session control is target-only and cannot become credential revocation.
- Missing telemetry is Unknown, never zero-by-assumption.
- TLS client certificate is authoritative for fleet identity.
- Forward-only immutable migration ledger.
- User-facing compatibility claims require real evidence, not just build/unit success.

## Coordinator checkpoint — 2026-09-15

- Current code head: `edfae4efc7e5da7f714757fe74804901f24a9d13`. The gofmt regression is repaired; the follow-on 0034 rollback defects (destructive marker + schema ledger removal) are repaired and migration checksums refreshed.
- Disposable PostgreSQL 18 `tests/db/migration_test.sh` passes on the exact checkout. Exact-head GitHub CI `34920505396` is terminal SUCCESS; docs-tip CI `34920603677` is also terminal SUCCESS. The internal #121 CI/0034 recovery gate is cleared.
- Production remains untouched and blocked on a trusted `PVNaive-Production-Primary` reconnect plus read-only audit, fresh encrypted backup and independent rollback snapshot.
- #101 still requires real Karing import → parse → CONNECT → cleanup/revoke.

## Latest coordinator refresh — 2026-09-15 07:39 Asia/Tehran

- `main=f437f352...`; CI 34920603677 SUCCESS. Code head `edfae4ef...`; CI 34920505396 SUCCESS.
- Current-main Web independently PASS: 23 files / 117 tests + production build.
- Production Primary remains disconnected; no mutation.
- Keep workers advancing #120/#101, #114, #115 and R8 gates while Production is blocked.
