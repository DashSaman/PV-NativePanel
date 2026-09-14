# Continue Here — PVNaive

Verified checkpoint: 2026-09-14 12:41 Asia/Tehran

## Current GitHub truth
- Canonical `main` before this documentation refresh: `3ea84fb9d8bdaba7fcd6fd33d1dd44d4ff04d1f6`; push CI `34821820421` SUCCESS.
- Runtime deploy commit remains `a4edea62594d5b60a978c39e1f28fad9ac45f6b6`; CI `34795216345` SUCCESS.
- Task13 #108 remains DRAFT/unmerged; branch/main reconciliation and real pinned-Caddy HTTP/1.1 + HTTP/2 acceptance are still required.
- Karing #101 remains DRAFT/unmerged; real disposable Karing import → parse → CONNECT → cleanup/revoke acceptance is still missing.
- STEER-002 #110 / PR #112 advanced from proven RED to minimal GREEN head `2ebe4637bf8f0d4fefb789fc299f5a404cc867fe`. `Decision.Candidates` now applies CandidateRatio without altering the full healthy eligible set used for primary/hysteresis, and the best candidate is always retained for zero/negative score domains. PR is mergeable but stays DRAFT while CI `34826791358`, Exact Accounting `34826791281`, and Pinned Forwardproxy `34826791365` run and independent edge-case review is pending.

## Production truth
- Latest persistent deployment receipt remains `pvnaive:repo-live2` image `76c10697a03b`, built from `a4edea6`, healthy at schema 30 after forward-only 0029/0030.
- Recorded nip.io E2E remains ALL_GREEN; previous `repo-live` remains retained for rollback.
- `namir.softarg.ir` remains deferred until the documented ACME retry window clears on 2026-09-15.
- Fresh remote inventory shows both known PVNaive Desktop Commander registrations offline and no Production Primary available, so no current shell/container/backup/disk/rollback state is asserted beyond the persistent receipt.

## Active lanes
- **Task13 #108**: Worker 3 refreshes/reconciles exact session-kill work on latest green main and reruns CI + Exact Accounting + Pinned Forwardproxy if head moves. Worker 2 then independently executes real pinned-Caddy HTTP/1.1 + HTTP/2 acceptance: target-only kill, sibling survival, forged-tuple rejection, repeat idempotency, account survival, unchanged Caddy lifecycle and exactly-once final accounting.
- **Karing #101**: Worker 4 performs real disposable Karing import → parse → CONNECT → cleanup/revoke, then refreshes only the validated minimal delta on latest green main and reruns exact-head repository gates if needed.
- **R1 / STEER-001 #109**: continue independently. New migrations must be >0030; never reuse/rewrite 0029/0030. Network telemetry remains separate from exact byte-accounting/quota truth. Worker 3 = trusted TCP_INFO sampling; Worker 2 = DB ingest/replay/idempotency; Worker 1 = schema/security/accounting review; Worker 4 = E2E after an exact implementation head exists.
- **R2 / STEER-002 #110 / PR #112**: wait for all exact-head workflows. Then Worker 1 reviews config/spec/diff boundaries, Worker 2 reviews non-positive/tie/TopK-vs-Decision consistency, Worker 3 fixes only exact-head failures if any, and Worker 4 owns steering E2E. Keep DRAFT until all gates are green.
- **Production #100**: Primary performs a read-only audit first when connected. Before any future runtime deploy: fresh encrypted backup → independent rollback snapshot → exact deploy SHA lock → staged promotion → postflight → retain rollback.

No runtime merge, deploy, migration, restart/reload, credential rotation, DB/Caddy mutation, backup mutation or rollback mutation was performed in this checkpoint.