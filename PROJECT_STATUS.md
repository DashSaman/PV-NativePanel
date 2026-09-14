# PVNaive — Canonical Project Status

Last updated: 2026-09-14 10:45 Asia/Tehran

## Verified GitHub state
- Canonical `main` before this documentation refresh: `415c8ebfdd5cef38c49e13fbb99e014088143c33`; push CI `34812744583` SUCCESS.
- Deployed runtime commit remains `a4edea62594d5b60a978c39e1f28fad9ac45f6b6`; its CI `34795216345` is SUCCESS.
- Task13 PR #108 is OPEN/DRAFT at exact head `f7d8dd5aa8f33b1bc09e3f19bd26ffb219e650d9`; GitHub currently reports `mergeable=false`. Exact-head CI `34789937594`, Exact Accounting `34789937603`, and Pinned Forwardproxy `34789937575` are SUCCESS but remain historical until branch/main reconciliation is complete. Mandatory real pinned-Caddy HTTP/1.1 + HTTP/2 acceptance is still absent.
- Karing PR #101 is OPEN/DRAFT at exact head `216d53670066033403fe95f61b0402bb710186a3`; GitHub currently reports `mergeable=false`. Repository gates on that exact head were previously green; mandatory real Karing import → parse → CONNECT → cleanup/revoke evidence is still absent.
- No new worker completion was posted after the 09:42 coordinator dispatches. Remote device enumeration is unavailable in this non-interactive run, so this checkpoint does not claim fresh executor-online status.

## Production truth
- Latest persistent receipt remains `pvnaive:repo-live2` image `76c10697a03b`, schema 30 after forward-only 0029/0030, healthy readiness, recorded nip.io E2E ALL_GREEN, and previous `repo-live` retained for rollback.
- `namir.softarg.ir` remains intentionally deferred until the documented ACME retry window clears on 2026-09-15.
- Production Primary is not freshly connected/audited in this run, so no new shell-level assertion is made for container identity, migration ledger, encrypted-backup freshness, disk headroom, or rollback snapshot.

## Remaining gates
1. **Task13 #108:** Worker 3 refreshes/reconciles the branch on latest green main and reruns exact-head CI + Exact Accounting + Pinned Forwardproxy if the head moves. Worker 2 then performs the real pinned-Caddy HTTP/1.1 + HTTP/2 acceptance proving target-only kill, sibling survival, forged-tuple rejection, repeat idempotency, credential/account survival, unchanged Caddy lifecycle, and exactly-once final accounting.
2. **Karing #101:** Worker 4 runs real disposable Karing import → parse → CONNECT → cleanup/revoke with redacted evidence; only then reconstruct/refresh the minimal validated delta on latest green main and rerun exact-head repository gates.
3. **R1 / STEER-001 #109:** continue independently. Any new DB migration must be strictly >0030. Network telemetry remains separate from exact byte-accounting and quota truth.
4. **Production #100:** when Primary is available, perform read-only deployed image/revision, migration/schema, service/listener/Caddy, backup, disk and rollback audit before any future promotion.

## Worker allocation
- Worker 2: Task13 real HTTP1/HTTP2 acceptance after refreshed green exact head.
- Worker 3: Task13 refresh/reconcile; otherwise R1 trusted TCP_INFO sampling.
- Worker 4: real Karing acceptance and latest-main reconstruction; later R1 E2E.
- Worker 1: independent diff/schema/security/accounting/CI review.
- Primary: read-only Production audit when connected.
- Coordinator: CI reconciliation, validated integration, canonical documentation and promotion safety.

No runtime merge, deploy, migration, restart/reload, credential rotation, DB/Caddy mutation, backup mutation, or rollback mutation was performed in this checkpoint.