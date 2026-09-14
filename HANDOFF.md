# PVNaive Handoff

Checkpoint: 2026-09-14 14:40 Asia/Tehran

## Verified baseline
- Canonical repository `main` before this documentation refresh: `60db02ddd324d4463d4bfed528fab6c0bb0467fe`; CI `34832005854` SUCCESS.
- Runtime commit deployed to Production remains `a4edea62594d5b60a978c39e1f28fad9ac45f6b6`; CI `34795216345` SUCCESS.
- Latest persistent Production receipt records `pvnaive:repo-live2` image `76c10697a03b`, schema 30 after forward-only 0029/0030, healthy readiness and previous `repo-live` retained for rollback.
- Recorded nip.io E2E remains ALL_GREEN. Owner-domain switch remains deferred until the documented ACME retry window clears on 2026-09-15.

## Promotion truth
- Task13 PR #108 remains OPEN/DRAFT and unmerged. Real pinned-Caddy HTTP/1.1 + HTTP/2 acceptance and refreshed branch/main reconciliation remain mandatory before merge.
- Karing PR #101 remains OPEN/DRAFT and unmerged. Real disposable Karing import → parse → CONNECT → cleanup/revoke remains mandatory.
- R1 / STEER-001 #109 remains independent. Production schema is 30; new DB work must use a migration strictly >0030 and never rewrite/reuse 0029 or 0030. Network telemetry is not quota truth.
- R2 / STEER-002 #110 draft PR #112 exact head `74f62de1f14459ef510c9af24125fccde7eb9da6` is fully green on CI `34831867967`, Exact Accounting `34831867953`, and Pinned Forwardproxy `34831867963`. Coordinator review confirms ordering/best-retention semantics, but #110's explicit TopK/decision consistency requirement still needs independent disposition because this PR only filters `Decision.Candidates` while `TopK` remains ratio-agnostic. Keep DRAFT until Worker 2/1 review and Worker 4 E2E disposition are recorded.
- Fresh remote inventory shows both known PVNaive Desktop Commander registrations offline and no Production Primary available. Do not claim current container identity, migration ledger, backup freshness, disk headroom or rollback snapshot without a new receipt.

## Worker allocation
- Worker 2: Task13 real HTTP1/HTTP2 protocol/accounting acceptance; STEER-002 TopK-vs-Decision/non-positive/tie review; R1 ingest/replay/idempotency.
- Worker 3: refresh/reconcile Task13 on latest green main; only remediate STEER-002 if a concrete review defect is found; otherwise continue R1 trusted TCP_INFO sampling.
- Worker 4: real Karing client acceptance, then minimal latest-main reconstruction; steering/R1 E2E.
- Worker 1: independent diff/schema/security/accounting/CI review and STEER-002 config/spec plus one-candidate/all-Unknown boundaries.
- Primary: read-only Production audit when connected.
- Coordinator: CI reconciliation, validated integration, canonical documentation and promotion gates.

No runtime merge or Production mutation was performed. Concrete work this checkpoint is the completed green exact-head STEER-002 gate set, coordinator semantic review, explicit TopK/Decision blocker documentation, refreshed worker assignments and canonical state.