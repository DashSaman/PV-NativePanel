# PVNaive Handoff

Checkpoint: 2026-09-14 15:44 Asia/Tehran

## Verified baseline
- Canonical repository `main` before this documentation refresh: `d438878b22579651d411e003a2c74206641b5401`; CI `34837174073` SUCCESS.
- Runtime commit deployed to Production remains `a4edea62594d5b60a978c39e1f28fad9ac45f6b6`; CI `34795216345` SUCCESS.
- Latest persistent Production receipt records `pvnaive:repo-live2` image `76c10697a03b`, schema 30 after forward-only 0029/0030, healthy readiness and previous `repo-live` retained for rollback.
- Recorded nip.io E2E remains ALL_GREEN. Do not infer a fresh shell/container/backup state without Primary reconnect.

## Promotion truth
- Task13 PR #108 remains OPEN/DRAFT, unmerged, current head `f7d8dd5aa8f33b1bc09e3f19bd26ffb219e650d9`, `mergeable=false`. Real pinned-Caddy HTTP/1.1 + HTTP/2 acceptance and latest-main reconciliation remain mandatory.
- Karing PR #101 remains OPEN/DRAFT, unmerged, current head `216d53670066033403fe95f61b0402bb710186a3`, `mergeable=false`. Real disposable Karing import → parse → CONNECT → cleanup/revoke remains mandatory.
- R1 / STEER-001 #109 remains independent. Production schema is 30; new DB work must use a migration strictly >0030 and never rewrite/reuse 0029 or 0030. Network telemetry is not quota truth.
- R2 / STEER-002 #110 / PR #112 advanced to `9c43a71fcf4990eb8a9c053221fd9701c10f2caa`. TopK semantics are now explicitly locked: renderers consume `Decision.Candidates`, which already carries CandidateRatio filtering; raw `Score(...)` is not renderer input. Regression-only commit added; fresh CI `34842627049`, Exact Accounting `34842627114`, and Pinned Forwardproxy `34842627193` are running. Keep DRAFT until all three are green plus Worker 1 boundary review and Worker 4 steering E2E disposition.
- Fresh Remote Desktop inventory shows both known PVNaive registrations offline and no Production Primary available.

## Worker allocation
- Worker 2: Task13 real HTTP1/HTTP2 protocol/accounting acceptance; R1 ingest/replay/idempotency; STEER-002 only if an actual caller bypasses `Decision.Candidates`.
- Worker 3: refresh/reconcile Task13 on latest green main; otherwise R1 trusted TCP_INFO sampling; no speculative STEER-002 edits.
- Worker 4: real Karing client acceptance and steering/R1 E2E.
- Worker 1: independent diff/schema/security/accounting/CI review plus STEER-002 config/spec, one-candidate and all-Unknown boundaries.
- Primary: read-only Production audit when connected.
- Coordinator: CI reconciliation, validated integration, canonical documentation and promotion gates.

No runtime merge or Production mutation was performed. Concrete work this checkpoint is the resolved STEER-002 TopK/Decision contract, regression commit `9c43a71f...`, fresh exact-head gates, and refreshed canonical state.