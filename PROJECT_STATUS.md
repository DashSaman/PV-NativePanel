# PVNaive — Canonical Project Status

Last updated: 2026-09-14 21:4x Asia/Tehran

## Current verified GitHub truth

- Exact `main`: `4906556b98387685472ddd82c169edaa4dbeae67` (`fix(ci): align migration ledger gate with current main`).
- Exact-head GitHub Actions are green: CI `34879387500` SUCCESS, WS1 Exact Accounting `34879387510` SUCCESS, WS1 Pinned Forwardproxy `34879387503` SUCCESS.
- R1 / STEER-001 issue #109 is now **closed/completed**. Its final contract remains: trusted server-side telemetry only, fail-closed identity, and telemetry never becomes exact byte-accounting/quota truth.
- No promotion PR is currently eligible to merge. Two old draft branches remain materially behind current main and still require real acceptance before any reconstruction/merge decision:
  - Task13 PR #108: head `44db803d70ad62dd9867d3d3a85d2b2e57f8b429`, 20 commits behind current main / 6 commits ahead of its merge base; real pinned-Caddy HTTP/1.1 + HTTP/2 target-only kill/sibling-survival/forged-tuple/idempotency/credential-survival/exactly-once-accounting acceptance remains mandatory.
  - Karing PR #101: head `6691392639be9fae4656a861db4a6d16580f850d`, 19 commits behind current main / 6 commits ahead of its merge base; real disposable Karing import → parse → CONNECT → cleanup/revoke remains mandatory.

## Production truth ceiling

- Latest persistent verified Production checkpoint records schema **33** with the Master Upgrade Pack / `repo-fin2` generation live, healthy panel/API/SSE/real-customer CONNECT/accounting postflight, forward-only migrations 0031/0032/0033, and retained rollback/backup evidence.
- That persistent checkpoint is the current truth ceiling. A fresh Remote Desktop inventory no longer exposes a device identifiable as `PVNaive-Production-Primary`.
- The sole online registration currently reports device name `Pak-Nasheeee-haaaaaaaaa`; a read-only probe for `/opt/pvnaive/src` returned `No such file or directory`. This registration must **not** be treated as Production.
- Therefore no fresh claim is made for live image/SHA, schema ledger, listeners/Caddy, encrypted-backup freshness, disk headroom or rollback snapshots beyond the persistent verified checkpoint.
- No deploy, migration, restart/reload, credential rotation, DB/Caddy mutation, backup mutation or rollback mutation was performed in this reconciliation.

## Active roadmap lanes

1. **#113 — R5 UI:** Worker 2 owns RED-first pool-manager wizard/node-list implementation against schema-33 backend truth. Worker 1 independently reviews contract/security/accessibility/truthful drain and revision semantics. Worker 4 performs disposable browser E2E after an exact head exists.
2. **#114 — R5-PULL-001:** Worker 3 owns RED-first sibling mTLS pull listener with TLS-authenticated peer identity, fail-closed revision handling and replay/rotation tests. Worker 1 reviews PKI/trust/rollback; Worker 2 validates registry/revision integration; Worker 4 runs disposable multi-node E2E.
3. **#115 — R6-FLIP:** Worker 4 owns disposable coverd/Caddy rehearsal and degradation evidence; Worker 3 owns staged service/Caddy promotion and rollback mechanics; Worker 1 reviews persona/security/no-secret boundaries; Worker 2 validates panel/API/data-plane invariants. Production promotion is blocked until trusted Primary identity, read-only audit, fresh encrypted backup and rollback snapshot are green.
4. **#100 — Production lane:** keep read-only until the expected Production Primary registration is re-established. First action after reconnect is identity/SHA/schema/service/Caddy/backup/disk/rollback audit; only then consider promotion.

## Safety invariants

- Accounting/session/quota semantics are canonical truth and must not be weakened by steering/telemetry work.
- Never trust XFF/Forwarded/client headers for authoritative node/session identity.
- Applied migrations are immutable; all future DB changes are forward-only and ledger-checked.
- Production promotion sequence remains: exact-head CI → disposable rehearsal → trusted read-only Production audit → fresh encrypted backup + independent rollback snapshot → staged promotion → postflight → retain rollback.
- No completion credit is given for assignment-only work; workers must return exact SHA plus executable evidence.

## Coordinator checkpoint

Concrete progress this cycle: reconciled exact current main and all three canonical workflows; corrected R1/#109 from stale-open to closed/completed; detected and documented Remote Desktop identity/provenance mismatch before any Production mutation; refreshed #100 blocker; and activated #113/#114/#115 with explicit worker ownership and safety gates. Historical checkpoints remain available in Git history.