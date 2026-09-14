# PVNaive — Canonical Project Status

Last updated: 2026-09-14 15:44 Asia/Tehran

## Verified GitHub state
- Canonical `main` before this documentation refresh: `d438878b22579651d411e003a2c74206641b5401`; push CI `34837174073` SUCCESS.
- Deployed runtime commit remains `a4edea62594d5b60a978c39e1f28fad9ac45f6b6`; its CI `34795216345` is SUCCESS.
- Task13 PR #108 remains OPEN/DRAFT, unmerged and currently `mergeable=false` on head `f7d8dd5aa8f33b1bc09e3f19bd26ffb219e650d9`. Real pinned-Caddy HTTP/1.1 + HTTP/2 acceptance plus branch/main reconciliation remain mandatory.
- Karing PR #101 remains OPEN/DRAFT, unmerged and currently `mergeable=false` on head `216d53670066033403fe95f61b0402bb710186a3`. Mandatory real disposable Karing import → parse → CONNECT → cleanup/revoke evidence is still absent.
- R2 / STEER-002 #110 / draft PR #112 advanced to head `9c43a71fcf4990eb8a9c053221fd9701c10f2caa`. Coordinator resolved the TopK/Decision ambiguity without changing scoring/hysteresis: renderers must call `TopK` on ratio-filtered `Decision.Candidates`, not raw `Score(...)`. A regression now locks that contract. Fresh exact-head CI `34842627049`, WS1 Exact Accounting `34842627114`, and WS1 Pinned Forwardproxy `34842627193` are running; do not merge until all succeed plus independent boundary/E2E review completes.
- Fresh Remote Desktop inventory reports both known PVNaive registrations offline; no worker or Production Primary is currently executable through that channel.

## Production truth
- Latest persistent receipt remains `pvnaive:repo-live2` image `76c10697a03b`, runtime `a4edea62594d5b60a978c39e1f28fad9ac45f6b6`, schema 30 after forward-only 0029/0030, healthy readiness, recorded nip.io E2E ALL_GREEN, and previous `repo-live` retained for rollback.
- No fresh Production Primary shell audit exists in this checkpoint, so no new shell-level assertion is made for container identity, migration ledger, encrypted-backup freshness, disk headroom, or rollback snapshot.
- `namir.softarg.ir` remains deferred until the documented ACME retry window clears; do not restart/recreate Production merely to chase certificate issuance.

## Remaining gates
1. **Task13 #108:** Worker 3 reconciles/refreshes the branch on latest green main and reruns exact-head gates if the head moves. Worker 2 then runs real pinned-Caddy HTTP/1.1 + HTTP/2 acceptance proving target-only kill, sibling survival, forged-tuple rejection, repeat idempotency, credential/account survival, unchanged Caddy lifecycle, and exactly-once final accounting.
2. **Karing #101:** Worker 4 runs real disposable Karing import → parse → CONNECT → cleanup/revoke with redacted evidence; only then reconstruct/refresh the minimal validated delta on latest green main if needed.
3. **R1 / STEER-001 #109:** continue independently. Any new DB migration must be strictly >0030. Network telemetry remains separate from exact byte-accounting and quota truth.
4. **R2 / STEER-002 #110 / PR #112:** wait for fresh exact-head CI/accounting/forwardproxy; Worker 1 independently checks config/spec plus one-candidate/all-Unknown boundaries; Worker 4 owns steering E2E; Worker 2 reports only a concrete raw-score TopK caller if one exists; Worker 3 changes code only for a real defect.
5. **Production #100:** when Primary reconnects, perform read-only deployed image/revision, migration/schema, service/listener/Caddy, backup, disk and rollback audit before any future promotion.

## Worker allocation
- Worker 2: Task13 real HTTP1/HTTP2 acceptance; R1 ingest/replay/idempotency; STEER-002 only if a concrete TopK bypass caller is found.
- Worker 3: Task13 branch reconciliation; otherwise R1 trusted TCP_INFO sampling; no speculative STEER-002 changes.
- Worker 4: real Karing acceptance and steering/R1 E2E.
- Worker 1: independent diff/schema/security/accounting/CI review and STEER-002 config/spec boundary review.
- Primary: read-only Production audit when connected.
- Coordinator: CI reconciliation, validated integration, canonical documentation and promotion safety.

No runtime merge, deploy, migration, restart/reload, credential rotation, DB/Caddy mutation, backup mutation, or rollback mutation was performed in this checkpoint. Concrete progress is the STEER-002 TopK/Decision contract disposition plus regression commit and fresh exact-head gate run.