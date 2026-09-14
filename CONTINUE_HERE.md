# Continue Here — PVNaive

Verified checkpoint: 2026-09-14 15:44 Asia/Tehran

## Current GitHub truth
- Canonical `main` before this documentation refresh: `d438878b22579651d411e003a2c74206641b5401`; push CI `34837174073` SUCCESS.
- Runtime deploy commit remains `a4edea62594d5b60a978c39e1f28fad9ac45f6b6`; CI `34795216345` SUCCESS.
- Task13 #108 remains DRAFT/unmerged at `f7d8dd5aa8f33b1bc09e3f19bd26ffb219e650d9`, currently `mergeable=false`; branch/main reconciliation and real pinned-Caddy HTTP/1.1 + HTTP/2 acceptance are still required.
- Karing #101 remains DRAFT/unmerged at `216d53670066033403fe95f61b0402bb710186a3`, currently `mergeable=false`; real disposable Karing import → parse → CONNECT → cleanup/revoke acceptance is still missing.
- STEER-002 #110 / PR #112 advanced to `9c43a71fcf4990eb8a9c053221fd9701c10f2caa`. TopK/Decision semantics are resolved without changing primary/hysteresis: renderer TopK operates on ratio-filtered `Decision.Candidates`, never raw `Score(...)`. A regression locks this. Fresh exact-head CI `34842627049`, Exact Accounting `34842627114`, and Pinned Forwardproxy `34842627193` are running. Keep DRAFT until all are green plus independent boundary review and steering E2E disposition.

## Production truth
- Latest persistent deployment receipt remains `pvnaive:repo-live2` image `76c10697a03b`, built from `a4edea6`, healthy at schema 30 after forward-only 0029/0030.
- Recorded nip.io E2E remains ALL_GREEN; previous `repo-live` remains retained for rollback.
- Fresh remote inventory shows both known PVNaive Desktop Commander registrations offline and no Production Primary available, so no current shell/container/backup/disk/rollback state is asserted beyond the persistent receipt.

## Active lanes
- **Task13 #108**: Worker 3 refreshes/reconciles exact session-kill work on latest green main and reruns repository gates if the head moves. Worker 2 then executes real pinned-Caddy HTTP/1.1 + HTTP/2 acceptance: target-only kill, sibling survival, forged-tuple rejection, repeat idempotency, account survival, unchanged Caddy lifecycle and exactly-once final accounting.
- **Karing #101**: Worker 4 performs real disposable Karing import → parse → CONNECT → cleanup/revoke, then refreshes only the validated minimal delta on latest green main if needed.
- **R1 / STEER-001 #109**: continue independently. New migrations must be >0030; never reuse/rewrite 0029/0030. Network telemetry remains separate from exact byte-accounting/quota truth. Worker 3 = trusted TCP_INFO sampling; Worker 2 = DB ingest/replay/idempotency; Worker 1 = schema/security/accounting review; Worker 4 = E2E after an exact implementation head exists.
- **R2 / STEER-002 #110 / PR #112**: wait for fresh exact-head workflows. Worker 1 independently reviews config/spec and one-candidate/all-Unknown cases. Worker 4 owns steering E2E. Worker 2 only reports a concrete raw-score TopK bypass if found. Worker 3 changes code only for a concrete defect.
- **Production #100**: Primary performs a read-only audit first when connected. Before any future runtime deploy: fresh encrypted backup → independent rollback snapshot → exact deploy SHA lock → staged promotion → postflight → retain rollback.

No runtime merge, deploy, migration, restart/reload, credential rotation, DB/Caddy mutation, backup mutation or rollback mutation was performed in this checkpoint.