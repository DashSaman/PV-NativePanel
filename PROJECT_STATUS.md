# PVNaive — Canonical Project Status

Last updated: 2026-09-14 09:42 Asia/Tehran

## Verified GitHub state
- Canonical `main` before this documentation refresh: `f1c9092070e3d4bfd3b9c4f62ccf2b9a7c452ffc`; push CI `34809028966` SUCCESS.
- Deployed runtime commit remains `a4edea62594d5b60a978c39e1f28fad9ac45f6b6`; its CI `34795216345` is SUCCESS.
- Task13 PR #108 is OPEN/DRAFT at `f7d8dd5aa8f33b1bc09e3f19bd26ffb219e650d9`. Fresh GitHub metadata now reports `mergeable=false`. Its prior exact-head CI `34789937594`, Exact Accounting `34789937603`, and Pinned Forwardproxy `34789937575` remain historical SUCCESS. Fresh compare from base to current main spans 26 later commits and shows no changed-file overlap with the 35 Task13 files. Treat GitHub's current mergeability result as authoritative until the branch is refreshed and revalidated.
- Karing PR #101 is OPEN/DRAFT at `216d53670066033403fe95f61b0402bb710186a3`. Fresh GitHub metadata also reports `mergeable=false`. Its four changed RuntimeNaive/runtime paths still have no overlap with the 124 later main commits from its base.

## Production truth
- Latest persistent receipt remains `pvnaive:repo-live2` image `76c10697a03b`, schema 30 after forward-only 0029/0030, healthy readiness, recorded nip.io E2E ALL_GREEN, and previous `repo-live` retained for rollback.
- `namir.softarg.ir` remains intentionally deferred until the documented ACME retry window clears on 2026-09-15.
- Production Primary is not connected, so no fresh shell-level container, migration-ledger, backup, disk, or rollback audit is asserted in this cycle.

## Remaining gates
1. **Task13 #108:** Worker 3 refreshes/reconciles the branch on latest green main and reruns exact-head CI + Exact Accounting + Pinned Forwardproxy if the head moves. Worker 2 then performs real pinned-Caddy HTTP/1.1 + HTTP/2 acceptance for target-only kill, sibling survival, forged-tuple rejection, repeat idempotency, account survival, unchanged Caddy lifecycle and exactly-once final accounting.
2. **Karing #101:** Worker 4 provides real disposable Karing import → parse → CONNECT → cleanup/revoke evidence, then reconstructs only the validated minimal delta on latest green main and reruns exact-head repository gates.
3. **R1 / STEER-001 #109:** continue independently. Any new DB migration must be strictly >0030. Network telemetry remains separate from exact byte-accounting and quota truth.
4. **Production #100:** when Primary reconnects, perform read-only deployed image/revision, migration/schema, service/listener/Caddy, backup, disk and rollback audit before any future promotion.

## Worker allocation
- Worker 2: Task13 real HTTP1/HTTP2 acceptance after refreshed green exact head.
- Worker 3: Task13 refresh/reconcile; otherwise R1 trusted TCP_INFO sampling.
- Worker 4: real Karing acceptance and latest-main reconstruction; later R1 E2E.
- Worker 1: independent diff/schema/security/accounting/CI review.
- Primary: read-only Production audit when connected.
- Coordinator: CI reconciliation, validated integration, canonical documentation and promotion safety.

Current remote inventory has no online PVNaive worker or Production Primary; both known worker registrations are offline. Persisted GitHub assignments remain the execution queue. No Production mutation was performed in this checkpoint.