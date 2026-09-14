# PVNaive — Canonical Project Status

Last updated: 2026-09-15 (Asia/Tehran)

## Current verified GitHub truth

- Validated code main before this documentation checkpoint: `3eee556b908ec98c7a033de0ce63938b7dc0e714`. PR #119 fixed the R8 SSE test-harness race with a synchronized recorder; exact PR head `e983468bb80fbda5dc9509aa1e4b89b96431b059` passed CI, Exact Accounting and Pinned Forwardproxy before guarded merge. Post-merge main CI `34904000208` was still running at this checkpoint and is not pre-declared green.
- R8 live monitoring slice is integrated on main. The superseded PR #117 was closed unmerged and issue #116 was closed completed after browser E2E and exact-head CI/accounting/forwardproxy validation on the main implementation.
- Task13 PR #108 is merged. Its final exact head `1f65eccb6d71572f3ff4f15e942cac02e7bfa6c7` passed CI `34897871371`, Exact Accounting `34897871364`, and Pinned Forwardproxy `34897871355` before merge.
- Task13 real pinned-Caddy acceptance also passed on that exact head: HTTP/1.1 + HTTP/2 negotiated, forged tuple rejected, target-only kill, sibling survival, repeated-kill idempotency, credential survival, unchanged Caddy PID, and exactly-once final accounting. Reproducible Caddy SHA256: `6c55347714b355be18d0d35e487e4f6c821b13626c4285f2f4d9a1cc1ef0487b`.
- Karing PR #101 remains unmerged until a real disposable Karing import → parse → CONNECT → cleanup/revoke acceptance is recorded.
- Fresh disposable PostgreSQL 18 validation on exact main `3eee556b...` passed `tests/db/pool_registry_migration_test.sh`: schema >=33, single-use enrollment, monotonic append-only revisions, heartbeat non-rewind, drain-before-disable, and SECURITY DEFINER/no-direct-table-access boundaries.

## Production truth ceiling

- Latest persistent verified Production checkpoint records schema **33** with the Master Upgrade Pack / `repo-fin2` generation live, healthy panel/API/SSE/real-customer CONNECT/accounting postflight, forward-only migrations 0031/0032/0033, and retained rollback/backup evidence.
- Fresh remote inventory still does not expose an identifiable `PVNaive-Production-Primary`. The online remote registration is an execution worker, not Production.
- Therefore current Production image/container identity, schema ledger, encrypted-backup freshness, disk headroom and rollback snapshot are not freshly asserted. No Production mutation is allowed until the trusted Primary reconnects and the read-only audit plus backup/rollback gates pass.

## Active roadmap lanes

1. **R5 enablement (#113/#114):** code is on main; run disposable multi-node enroll → publish → mTLS pull → heartbeat/drift and STEER-006 scale rehearsal. Keep Production enablement gated.
2. **R6-FLIP #115:** code is on main and default-OFF; run disposable cover/persona/probe-sweep/failure rehearsals. Production promotion requires trusted Primary audit + fresh encrypted backup + independent rollback snapshot.
3. **R8 remaining gates:** extend validated monitoring to ledger reconciliation/per-node/per-user truth, RBAC stream isolation, accessibility/RTL and documented 100-node performance without fabricating missing telemetry.
4. **Karing #101:** real-client acceptance only; static/unit evidence cannot substitute.
5. **#100 Production lane:** on trusted Primary reconnect, read-only identity/SHA/schema/services/Caddy/backup/disk/rollback audit first.

## Worker allocation

- Worker 1: independent security/accounting review of merged Task13 boundaries plus R5/R6/R8 RBAC/accessibility review; include the now race-clean R8 SSE harness and report findings only, no speculative rewrites.
- Worker 2: R8 UI-002 ledger reconciliation and per-node/per-user truthful projections; RED-first tests and bounded data structures.
- Worker 3: R5 mTLS pull/STEER-006 integration + R8 stream-RBAC regression; preserve TLS-cert authoritative identity and exact-accounting separation.
- Worker 4: disposable R5/R6 browser/multi-node rehearsals and real Karing acceptance when a suitable client host is available.
- Coordinator: integrate only exact-head validated work, keep migrations forward-only, and enforce Production backup/rollback gates. Fresh remote inventory currently exposes one execution worker only; queue independent work there while other registrations/Production remain offline.

## Safety invariants

- Accounting/session/quota semantics are canonical truth and must not be weakened by telemetry/UI/control work.
- Task13 kills the selected live session only; it must not revoke credentials or mutate sibling sessions.
- Never trust XFF/Forwarded/client headers for authoritative node/session identity; fleet pull identity is TLS-client-cert-only.
- Applied migrations are immutable; future DB changes are forward-only and ledger-checked.
- Unknown/unavailable telemetry stays Unknown; UI must not convert missing data into zero or fabricated health.
- Production sequence: exact-head CI → disposable rehearsal → trusted read-only audit → fresh encrypted backup + independent rollback snapshot → staged promotion → postflight → retain rollback.
