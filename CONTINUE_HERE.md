# Continue Here — PVNaive

Verified checkpoint: 2026-09-14 21:4x Asia/Tehran

## GitHub truth
- Verified code baseline before this docs refresh: `4906556b98387685472ddd82c169edaa4dbeae67`.
- Exact-head CI `34879387500`, WS1 Exact Accounting `34879387510`, and WS1 Pinned Forwardproxy `34879387503` are all SUCCESS.
- R1 / STEER-001 #109 is closed/completed. Telemetry remains separate from exact byte-accounting/quota truth.
- Task13 #108 and Karing #101 remain old DRAFT branches, materially behind main. Do not merge either without reconstructing/reconciling on current main and satisfying their real-client/protocol acceptance gates.

## Production truth
- Persistent verified Production checkpoint: schema 33, Master Upgrade Pack / repo-fin2 generation live, healthy panel/API/SSE/real-customer CONNECT/accounting postflight, and retained backup/rollback evidence.
- Current Remote Desktop inventory does not contain a trustworthy `PVNaive-Production-Primary`. The sole online registration identifies as `Pak-Nasheeee-haaaaaaaaa`, and `/opt/pvnaive/src` is absent there.
- Treat Production as inaccessible until the expected Primary identity reconnects. No Production mutation was made.

## Execute next
1. #113: Worker 2 starts RED-first R5 pool-manager UI; Worker 1 reviews security/accessibility/truthful state; Worker 4 runs disposable browser E2E after exact head.
2. #114: Worker 3 starts RED-first authenticated sibling mTLS pull listener; Worker 1 reviews PKI/trust/rollback; Worker 2 validates registry/revision integration; Worker 4 runs disposable multi-node E2E.
3. #115: Worker 4 prepares disposable coverd/Caddy rehearsal; Worker 3 owns staged promotion/rollback mechanics; Worker 1 reviews persona/security/no-secret boundaries; Worker 2 checks panel/API/data-plane invariants. Do not touch Production until trusted Primary audit + backup + rollback gates are green.
4. #100: first trusted-Primary action is read-only host/SHA/image/schema/services/Caddy/backup/disk/rollback audit.
5. #108/#101: keep DRAFT. Reconcile only when real pinned-Caddy HTTP1/2 and real Karing acceptance hosts are available; historical CI is not sufficient for current-main merge.

## Invariants
- Preserve exact accounting, session, quota and credential semantics.
- No XFF/Forwarded/client-header authority for trusted identity.
- No fabricated health/telemetry and no secret-bearing evidence.
- Forward-only migrations; never rewrite applied history.
- Promotion order: exact-head CI → disposable rehearsal → trusted Production read-only audit → fresh encrypted backup + rollback snapshot → staged promotion → postflight → retain rollback.

Concrete progress in the latest cycle: current main and all canonical gates verified green, #109 closed completed, Remote Desktop identity mismatch caught before any Production action, #100 updated, and #113/#114/#115 redispatched.