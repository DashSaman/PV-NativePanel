# PVNaive Handoff

Checkpoint: 2026-09-14 12:41 Asia/Tehran

## Verified baseline
- Canonical repository `main` before this documentation refresh: `3ea84fb9d8bdaba7fcd6fd33d1dd44d4ff04d1f6`; CI `34821820421` SUCCESS.
- Runtime commit deployed to Production remains `a4edea62594d5b60a978c39e1f28fad9ac45f6b6`; CI `34795216345` SUCCESS.
- Latest persistent Production receipt records `pvnaive:repo-live2` image `76c10697a03b`, schema 30 after forward-only 0029/0030, healthy readiness and previous `repo-live` retained for rollback.
- Recorded nip.io E2E remains ALL_GREEN. Owner-domain switch remains deferred until the documented ACME retry window clears on 2026-09-15.

## Promotion truth
- Task13 PR #108 remains OPEN/DRAFT and unmerged. Real pinned-Caddy HTTP/1.1 + HTTP/2 acceptance and refreshed branch/main reconciliation remain mandatory before merge.
- Karing PR #101 remains OPEN/DRAFT and unmerged. Real disposable Karing import → parse → CONNECT → cleanup/revoke remains mandatory.
- R1 / STEER-001 #109 remains independent. Production schema is 30; new DB work must use a migration strictly >0030 and never rewrite/reuse 0029 or 0030. Network telemetry is not quota truth.
- R2 / STEER-002 #110 draft PR #112 now has minimal implementation head `2ebe4637bf8f0d4fefb789fc299f5a404cc867fe`. CandidateRatio filters `Decision.Candidates` only; primary/hysteresis continue to use the full healthy eligible set, and the best candidate is unconditional for non-positive score safety. PR is mergeable but remains DRAFT while exact-head CI `34826791358`, Exact Accounting `34826791281`, and Pinned Forwardproxy `34826791365` run and independent edge-case review is pending.
- Fresh remote inventory shows both known PVNaive Desktop Commander registrations offline and no Production Primary available. Do not claim current container identity, migration ledger, backup freshness, disk headroom or rollback snapshot without a new receipt.

## Worker allocation
- Worker 2: Task13 real HTTP1/HTTP2 protocol/accounting acceptance; STEER-002 non-positive/tie/TopK-vs-Decision review; R1 ingest/replay/idempotency.
- Worker 3: refresh/reconcile Task13 on latest green main; only remediate STEER-002 exact-head failures if any; otherwise continue R1 trusted TCP_INFO sampling.
- Worker 4: real Karing client acceptance, then minimal latest-main reconstruction; later R1/steering E2E.
- Worker 1: independent diff/schema/security/accounting/CI review and STEER-002 config/spec boundaries.
- Primary: read-only Production audit when connected.
- Coordinator: CI reconciliation, validated integration, canonical documentation and promotion gates.

No runtime merge or Production mutation was performed. Concrete work this checkpoint is the STEER-002 minimal GREEN implementation plus refreshed persistent assignments and canonical state.