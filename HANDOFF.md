# PVNaive Handoff

Checkpoint: 2026-09-14 09:42 Asia/Tehran

## Verified baseline
- Canonical repository `main` before this documentation refresh: `f1c9092070e3d4bfd3b9c4f62ccf2b9a7c452ffc`; CI `34809028966` SUCCESS.
- Runtime commit deployed to Production remains `a4edea62594d5b60a978c39e1f28fad9ac45f6b6`; CI `34795216345` SUCCESS.
- Latest persistent Production receipt records `pvnaive:repo-live2` image `76c10697a03b`, schema 30 after forward-only 0029/0030, healthy readiness and previous `repo-live` retained for rollback.
- Recorded nip.io E2E remains ALL_GREEN. Owner-domain switch remains deferred until the documented ACME retry window clears on 2026-09-15.

## Promotion truth
- Task13 PR #108 is OPEN/DRAFT at `f7d8dd5aa8f33b1bc09e3f19bd26ffb219e650d9`; fresh GitHub metadata now reports `mergeable=false`. Historical exact-head workflows are green, but current promotion requires branch refresh/reconciliation and then real HTTP/1.1 + HTTP/2 pinned-Caddy acceptance. Fresh base→main compare spans 26 later commits with no changed-file overlap, so the non-mergeable state is not explained by a simple path-overlap check; do not override GitHub's current result.
- Karing PR #101 is OPEN/DRAFT at `216d53670066033403fe95f61b0402bb710186a3`; fresh GitHub metadata also reports `mergeable=false`. Its four changed paths still have no overlap with 124 later main commits. Real disposable Karing import → parse → CONNECT → cleanup/revoke remains mandatory before promotion.
- R1 / STEER-001 #109 remains independent. Production schema is 30, so new DB work must use a migration strictly >0030 and must never rewrite/reuse 0029 or 0030. Network telemetry is not quota truth.
- Production Primary is not connected, so current container identity, migration ledger, backup freshness, disk headroom and rollback snapshot are not freshly shell-verified in this cycle.

## Worker allocation
- Worker 2: Task13 real HTTP1/HTTP2 protocol/accounting acceptance after refreshed exact-head gates pass.
- Worker 3: refresh/reconcile Task13 on latest green main; otherwise R1 trusted TCP_INFO sampling.
- Worker 4: real Karing client acceptance, then minimal latest-main reconstruction; later R1 E2E.
- Worker 1: independent diff/schema/security/accounting/CI review of returned work.
- Primary: read-only Production audit when connected.
- Coordinator: CI reconciliation, validated integration, canonical documentation and promotion gates.

Current remote inventory has no online PVNaive worker or Production Primary. Persisted GitHub assignments remain the execution queue. No additional Production mutation was performed in this checkpoint.