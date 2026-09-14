# PVNaive Handoff

Checkpoint: 2026-09-14 21:4x Asia/Tehran

## Verified baseline
- Code baseline verified before this documentation refresh: `4906556b98387685472ddd82c169edaa4dbeae67`.
- Exact-head gates: CI `34879387500` SUCCESS; WS1 Exact Accounting `34879387510` SUCCESS; WS1 Pinned Forwardproxy `34879387503` SUCCESS.
- R1 / STEER-001 #109 is closed/completed. Trusted telemetry remains observational only and separate from exact byte-accounting/quota truth.
- Latest persistent Production receipt is schema 33 with the Master Upgrade Pack / repo-fin2 generation live and green real-customer CONNECT/accounting postflight. Treat it as the truth ceiling until a trusted Primary is freshly audited.

## Production blocker
- Fresh Remote Desktop inventory does **not** expose an identifiable `PVNaive-Production-Primary`.
- The sole online registration is named `Pak-Nasheeee-haaaaaaaaa`; a read-only probe for `/opt/pvnaive/src` failed because that path is absent.
- Do not use that registration for PVNaive Production. No Production mutation was performed.
- On trusted Primary reconnect: read-only verify host identity, deployed image/SHA, schema/migration ledger, services/listeners/Caddy, encrypted backups, disk headroom and rollback snapshots before any deploy decision.

## Open draft PRs
- Task13 #108: head `44db803d70ad62dd9867d3d3a85d2b2e57f8b429`; branch has diverged and is 20 commits behind current main. Real exact-head pinned-Caddy HTTP/1.1 + HTTP/2 acceptance remains mandatory before any reconstruction/merge.
- Karing #101: head `6691392639be9fae4656a861db4a6d16580f850d`; branch is 19 commits behind current main. Real disposable Karing import → parse → CONNECT → cleanup/revoke remains mandatory.

## Active worker queue
- Worker 2: #113 R5 pool-manager UI RED-first implementation; validate #114 registry/revision integration; validate #115 panel/API/data-plane invariants.
- Worker 3: #114 sibling mTLS pull listener RED-first implementation; #115 staged service/Caddy promotion mechanics and rollback contract.
- Worker 4: #113 disposable browser E2E after exact head; #114 disposable multi-node E2E; #115 disposable coverd/Caddy degradation/failure rehearsal.
- Worker 1: independent security/schema/accounting/accessibility/PKI/persona review across produced exact heads.
- Coordinator: integrate only exact-head validated work, keep migration/accounting/session semantics truthful, and maintain Production backup/rollback gates.

## Safety gates
- No invented health/telemetry; no XFF/Forwarded authority; no secret material in repo/logs.
- Applied migrations are immutable and future migrations are forward-only.
- Production sequence: exact-head CI → disposable rehearsal → trusted Primary read-only audit → fresh encrypted backup + independent rollback snapshot → staged promotion → postflight → retain rollback.

Concrete progress this cycle: exact main/workflows reconciled; stale R1 issue closed completed; Production identity mismatch detected before mutation and recorded on #100; #113/#114/#115 activated with explicit ownership. Historical handoffs remain in Git history.