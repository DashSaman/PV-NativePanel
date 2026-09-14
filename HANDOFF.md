# PVNaive Handoff

Checkpoint: 2026-09-14 11:40 Asia/Tehran

## Verified baseline
- Canonical repository `main` before this documentation refresh: `638844337e814e806cbe90897e522aadc6d43863`; CI `34817253202` SUCCESS.
- Runtime commit deployed to Production remains `a4edea62594d5b60a978c39e1f28fad9ac45f6b6`; CI `34795216345` SUCCESS.
- Latest persistent Production receipt records `pvnaive:repo-live2` image `76c10697a03b`, schema 30 after forward-only 0029/0030, healthy readiness and previous `repo-live` retained for rollback.
- Recorded nip.io E2E remains ALL_GREEN. Owner-domain switch remains deferred until the documented ACME retry window clears on 2026-09-15.

## Promotion truth
- Task13 PR #108 is OPEN/DRAFT at `f7d8dd5aa8f33b1bc09e3f19bd26ffb219e650d9`; current GitHub metadata reports `mergeable=false`. Its exact-head repository gates are historical-green, but branch/main reconciliation and real pinned-Caddy HTTP/1.1 + HTTP/2 acceptance are still required before merge.
- Karing PR #101 is OPEN/DRAFT at `216d53670066033403fe95f61b0402bb710186a3`; current GitHub metadata reports `mergeable=false`. Real disposable Karing import → parse → CONNECT → cleanup/revoke remains mandatory.
- R1 / STEER-001 #109 remains independent. Production schema is 30; new DB work must use a migration strictly >0030 and never rewrite/reuse 0029 or 0030. Network telemetry is not quota truth.
- R2 / STEER-002 #110 has active draft PR #112. RED-only head `7a7923cf8faf6f03a2ea1578570a705f9df57153` adds CandidateRatio regression tests. WS1 Exact Accounting is green; CI and Pinned Forwardproxy are still running. Implementation must wait until the intended RED failure is observed.
- Fresh remote inventory shows both known PVNaive Desktop Commander registrations offline and no Production Primary available. Do not claim current container identity, migration ledger, backup freshness, disk headroom or rollback snapshot without a new receipt.

## Worker allocation
- Worker 2: Task13 real HTTP1/HTTP2 protocol/accounting acceptance after refreshed exact-head gates pass; independently review STEER-002 non-positive score and TopK/Decision consistency.
- Worker 3: refresh/reconcile Task13 on latest green main; implement STEER-002 minimally after RED is proven; otherwise continue R1 trusted TCP_INFO sampling.
- Worker 4: real Karing client acceptance, then minimal latest-main reconstruction; later R1/steering E2E.
- Worker 1: independent diff/schema/security/accounting/CI review of returned work and STEER-002 config/spec boundaries.
- Primary: read-only Production audit when connected.
- Coordinator: CI reconciliation, validated integration, canonical documentation and promotion gates.

No runtime merge or Production mutation was performed. New concrete work this checkpoint is the RED-first STEER-002 branch/PR plus refreshed canonical handoff/progress state.