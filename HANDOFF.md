# PVNaive Handoff

Checkpoint: 2026-09-15 Asia/Tehran

## Verified baseline

- Validated code main before this documentation commit: `deaf0f8edb1bdbedbaf71a39518d1a3fe941b693`.
- Task13 PR #108 is merged after exact-head CI `34897871371`, Exact Accounting `34897871364`, Pinned Forwardproxy `34897871355`, and real pinned-Caddy HTTP/1.1+HTTP/2 acceptance all passed.
- Task13 accepted binary SHA256: `6c55347714b355be18d0d35e487e4f6c821b13626c4285f2f4d9a1cc1ef0487b`; target-only kill, sibling survival, forged-tuple rejection, idempotency, credential survival, unchanged Caddy lifecycle and exactly-once final accounting were verified.
- R8 issue #116 is completed on main. PR #117 was closed unmerged as stale/superseded by the stronger main implementation and fixes.
- R5 UI/pull and R6 gated cover flip code remain on main; Production enablement remains gated.

## Production blocker

- No identifiable `PVNaive-Production-Primary` is connected. The currently online remote is an execution worker and must not be treated as Production.
- On trusted Primary reconnect, perform read-only host identity, deployed image/SHA, schema/migration ledger, services/listeners/Caddy, encrypted-backup freshness, disk headroom and rollback snapshot checks before any mutation.
- Production truth ceiling remains the persistent schema-33 / repo-fin2 checkpoint; do not claim fresher live state without that audit.

## Remaining acceptance blockers

- #101 Karing: requires a real disposable Karing import → parse → CONNECT → cleanup/revoke run.
- R5/R6: require disposable multi-node / cover-site rehearsals before any Production enablement.
- Production: trusted Primary reconnect plus audit/backup/rollback gates.

## Worker queue

- Worker 1 — independent security/accounting review: Task13 post-merge boundaries, R5/R6 privilege boundaries, R8 accessibility/RBAC honesty.
- Worker 2 — R8 UI-002: ledger reconciliation + truthful per-node/per-user projections, RED-first.
- Worker 3 — R5 pull/STEER-006 integration and R8 stream-RBAC isolation regression.
- Worker 4 — R5/R6 browser/multi-node E2E and real Karing acceptance when a suitable client is available.
- Coordinator — exact-head integration only; no Production mutation without trusted audit + fresh backup + rollback.

## Invariants

- Preserve exact accounting, session, quota and credential semantics.
- Task13 session control is target-only and cannot become credential revocation.
- Missing telemetry is Unknown, never zero-by-assumption.
- TLS client certificate is authoritative for fleet identity.
- Forward-only immutable migration ledger.
