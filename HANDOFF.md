# PVNaive Handoff

Checkpoint: 2026-09-14 16:47 Asia/Tehran

## Verified baseline
- Canonical repository main moved to guarded STEER-002 merge commit `f6b1bab91fa1583a647770a766c7cf9d58f9ee89`; push CI `34848316732` is currently in progress.
- PR #112 exact head `9c43a71fcf4990eb8a9c053221fd9701c10f2caa` was green on CI `34842627049`, Exact Accounting `34842627114`, and Pinned Forwardproxy `34842627193` before merge. Independent review covered CandidateRatio thresholding, non-positive scores, TopK-on-Decision semantics, one-eligible, all-Unknown, no-flapping, hysteresis and kill-switch behavior. PR #112 is merged.
- Runtime commit deployed to Production remains `a4edea62594d5b60a978c39e1f28fad9ac45f6b6`; CI `34795216345` SUCCESS.
- Latest persistent Production receipt records `pvnaive:repo-live2` image `76c10697a03b`, schema 30 after forward-only 0029/0030, healthy readiness and previous `repo-live` retained for rollback.
- Recorded nip.io E2E remains ALL_GREEN. Do not infer a fresh shell/container/backup state without Primary reconnect.

## Promotion truth
- STEER-002 #110 is code-integrated through merge commit `f6b1bab9...`; close the issue only after post-merge main CI `34848316732` completes SUCCESS.
- Task13 PR #108 remains OPEN/DRAFT, current head `f7d8dd5aa8f33b1bc09e3f19bd26ffb219e650d9`, currently `mergeable=false`. Latest-main reconciliation plus real pinned-Caddy HTTP/1.1 + HTTP/2 acceptance remain mandatory.
- Karing PR #101 remains OPEN/DRAFT, current head `216d53670066033403fe95f61b0402bb710186a3`, currently `mergeable=false`. Real disposable Karing import → parse → CONNECT → cleanup/revoke remains mandatory.
- R1 / STEER-001 #109 remains independent. Production schema is 30; new DB work must use a migration strictly >0030 and never rewrite/reuse 0029 or 0030. Network telemetry is not quota truth.
- Fresh Remote Desktop inventory shows both known PVNaive registrations offline and no Production Primary available.

## Worker allocation
- Worker 2: Task13 real HTTP1/HTTP2 protocol/accounting acceptance; R1 ingest/replay/idempotency.
- Worker 3: refresh/reconcile Task13 on latest green main; otherwise R1 trusted TCP_INFO sampling.
- Worker 4: real Karing client acceptance and R1/steering E2E follow-up.
- Worker 1: independent diff/schema/security/accounting/CI review of new exact heads.
- Primary: read-only Production audit when connected.
- Coordinator: reconcile post-merge main CI, close STEER-002 only after green, integrate validated work, maintain canonical docs and promotion gates.

No Production mutation was performed. Concrete work this checkpoint is the guarded merge of STEER-002 PR #112 after exact-head repository gates and independent semantic/edge-case review.