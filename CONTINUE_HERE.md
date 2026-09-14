# Continue Here — PVNaive

Verified checkpoint: 2026-09-14 11:40 Asia/Tehran

## Current GitHub truth
- Canonical `main` before this documentation refresh: `638844337e814e806cbe90897e522aadc6d43863`; push CI `34817253202` SUCCESS.
- Runtime deploy commit remains `a4edea62594d5b60a978c39e1f28fad9ac45f6b6`; CI `34795216345` SUCCESS.
- Task13 #108 is DRAFT at `f7d8dd5aa8f33b1bc09e3f19bd26ffb219e650d9`; current GitHub metadata reports `mergeable=false`. Exact-head repository workflows are historical-green, but branch/main reconciliation plus real pinned-Caddy HTTP/1.1 + HTTP/2 acceptance are still required.
- Karing #101 is DRAFT at `216d53670066033403fe95f61b0402bb710186a3`; current GitHub metadata reports `mergeable=false`. Real disposable Karing import → parse → CONNECT → cleanup/revoke acceptance is still missing.
- STEER-002 #110 has advanced into draft PR #112 from exact green main. RED-first head is `7a7923cf8faf6f03a2ea1578570a705f9df57153`; it adds CandidateRatio contract tests only. WS1 Exact Accounting is green on that head; main CI and Pinned Forwardproxy are still running. Do not add implementation until the expected RED failure is observed.

## Production truth
- Latest persistent deployment receipt remains `pvnaive:repo-live2` image `76c10697a03b`, built from `a4edea6`, healthy at schema 30 after forward-only 0029/0030.
- Recorded nip.io E2E remains ALL_GREEN; previous `repo-live` remains retained for rollback.
- `namir.softarg.ir` remains deferred until the documented ACME retry window clears on 2026-09-15.
- Fresh remote inventory shows both known PVNaive Desktop Commander registrations offline and no Production Primary available, so no current shell/container/backup/disk/rollback state is asserted beyond the persistent receipt.

## Active lanes
- **Task13 #108**: Worker 3 refreshes/reconciles exact session-kill work on latest green main and reruns CI + Exact Accounting + Pinned Forwardproxy if head moves. Worker 2 then independently executes real pinned-Caddy HTTP/1.1 + HTTP/2 acceptance: target-only kill, sibling survival, forged-tuple rejection, repeat idempotency, account survival, unchanged Caddy lifecycle and exactly-once final accounting.
- **Karing #101**: Worker 4 performs real disposable Karing import → parse → CONNECT → cleanup/revoke, then reconstructs only the validated minimal delta on latest green main and reruns exact-head repository gates.
- **R1 / STEER-001 #109**: continue independently. New migrations must be >0030; never reuse/rewrite 0029/0030. Network telemetry remains separate from exact byte-accounting/quota truth. Worker 3 = trusted TCP_INFO sampling; Worker 2 = DB ingest/replay/idempotency; Worker 1 = schema/security/accounting review; Worker 4 = E2E after an exact implementation head exists.
- **R2 / STEER-002 #110 / PR #112**: RED tests are now committed. After CI proves RED for the intended CandidateRatio gap, Worker 3 adds the minimal candidate filtering; Worker 2 reviews non-positive score semantics and TopK/Decision consistency; Worker 1 reviews config/spec boundaries; Worker 4 owns steering E2E. Preserve Unknown/hysteresis/kill-switch/accounting semantics.
- **Production #100**: Primary performs a read-only audit first when connected. Before any future runtime deploy: fresh encrypted backup → independent rollback snapshot → exact deploy SHA lock → staged promotion → postflight → retain rollback.

No runtime merge, deploy, migration, restart/reload, credential rotation, DB/Caddy mutation, backup mutation or rollback mutation was performed in this checkpoint.