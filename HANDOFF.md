# PVNaive Handoff

Checkpoint: 2026-09-14 08:42 Asia/Tehran

## Verified baseline
- Exact canonical repository `main` before this documentation refresh: `1db7837357f528fd4a6ce14b69bb91871693eea8`; CI `34805219116` SUCCESS.
- Runtime commit deployed to Production: `a4edea62594d5b60a978c39e1f28fad9ac45f6b6`; CI `34795216345` SUCCESS.
- Latest persistent Production receipt records `pvnaive:repo-live2` image `76c10697a03b`, schema 30 after forward-only 0029/0030, healthy readiness, preserved owner/account state, and previous `repo-live` retained for rollback.
- Full recorded nip.io E2E is ALL_GREEN: strict TLS, panel 200, login 200, user creation 201, subscription headers, UA negotiation, and strict-TLS proxy CONNECT 204. Seven disposable E2E users were revoked after the test with history retained.
- Owner-domain switch remains intentionally deferred until the documented ACME retry window clears on 2026-09-15. Do not restart/recreate merely to force certificate issuance.

## Promotion truth
- Task13 PR #108 is OPEN/DRAFT at `f7d8dd5aa8f33b1bc09e3f19bd26ffb219e650d9` and GitHub currently reports mergeable=true. Exact-head CI `34789937594`, Exact Accounting `34789937603`, and Pinned Forwardproxy `34789937575` are SUCCESS. The 24 main commits since its base have no changed-file overlap with the 35 Task13 files, so no current reconstruction is needed solely for mergeability. Promotion is still blocked on independent real HTTP/1.1 + HTTP/2 pinned-Caddy proof: target-only kill, sibling survival, forged-tuple rejection, repeat idempotency, account survival, unchanged Caddy lifecycle and exactly-once final accounting.
- Karing PR #101 is OPEN/DRAFT at `216d53670066033403fe95f61b0402bb710186a3` and GitHub currently reports mergeable=true. Its four RuntimeNaive/runtime changed files do not overlap the 122 later main commits from its base. Promotion still requires real disposable Karing import → parse → CONNECT → cleanup/revoke evidence. Static/unit/build evidence is not acceptance.
- R1 / STEER-001 #109 remains independent. Production schema is 30, so new DB work must use a migration strictly >0030 and must never rewrite/reuse 0029 or 0030. Network telemetry is not quota truth and must remain separate from exact byte-accounting rows.
- Production Primary is not connected through the current command channel. Therefore this coordinator has not freshly shell-verified current container identity, migration ledger, encrypted-backup freshness, disk headroom or rollback snapshot in this cycle.

## Worker allocation
- Worker 2: Task13 real HTTP1/HTTP2 protocol/accounting acceptance.
- Worker 3: respond only to a new Task13 exact-head/merge-ref defect; otherwise R1 trusted TCP_INFO sampling.
- Worker 4: real Karing client acceptance; later R1 E2E when an exact implementation head exists.
- Worker 1: independent diff/schema/security/accounting/CI review and validation of returned receipts.
- Primary: read-only Production audit when connected; before any next deploy, confirm backup and rollback prerequisites.
- Coordinator: CI reconciliation, validated integration, canonical documentation and promotion gates.

Current remote inventory has no online PVNaive worker or Production Primary. Persisted GitHub assignments remain the execution queue for the next available worker. No additional Production mutation was performed in this checkpoint.