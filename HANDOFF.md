# PVNaive Handoff

Checkpoint: 2026-09-14 18:42 Asia/Tehran

## Verified baseline
- Canonical main before this documentation refresh was `714b2f5490b6eec317061b0834a2721e2c6a8fa6`; push CI `34848486540` completed SUCCESS.
- STEER-002 / #110 is completed and closed. PR #112 merged via code-bearing `f6b1bab91fa1583a647770a766c7cf9d58f9ee89`; exact pre-merge head `9c43a71fcf4990eb8a9c053221fd9701c10f2caa` passed CI `34842627049`, Exact Accounting `34842627114`, and Pinned Forwardproxy `34842627193`.
- Runtime commit deployed to Production remains `a4edea62594d5b60a978c39e1f28fad9ac45f6b6`; CI `34795216345` SUCCESS.
- Latest persistent Production receipt records `pvnaive:repo-live2` image `76c10697a03b`, schema 30 after forward-only 0029/0030, healthy readiness, nip.io E2E ALL_GREEN, and previous `repo-live` retained for rollback.
- Fresh Remote Desktop inventory shows both known registrations offline and no Production Primary available; do not infer a fresh shell/container/backup state.

## Promotion truth
- Task13 #108 is OPEN/DRAFT, head `f7d8dd5aa8f33b1bc09e3f19bd26ffb219e650d9`, currently `mergeable=true`. Exact-head CI `34789937594`, Exact Accounting `34789937603`, and Pinned Forwardproxy `34789937575` are SUCCESS. Merge remains blocked only by missing real pinned-Caddy HTTP/1.1 + HTTP/2 acceptance and final independent review.
- Karing #101 is OPEN/DRAFT, head `216d53670066033403fe95f61b0402bb710186a3`, currently `mergeable=true`. Exact-head CI `34732580376`, Exact Accounting `34732580468`, and Pinned Forwardproxy `34732580377` are SUCCESS. Merge remains blocked by missing real Karing import → parse → CONNECT → cleanup/revoke evidence.
- R1 / STEER-001 #109 remains the active independent roadmap lane. New migrations must be strictly >0030; 0029/0030 are immutable. Network telemetry is not quota/exact-byte-accounting truth.

## Worker allocation
- Worker 2: Task13 real HTTP1/HTTP2 acceptance; R1 ingest/replay/idempotency.
- Worker 3: R1 trusted TCP_INFO sampling; Task13 reconcile only for a concrete new exact-head/merge-ref defect.
- Worker 4: real Karing client acceptance; R1 E2E after an exact implementation head exists.
- Worker 1: independent diff/schema/security/accounting/CI review of produced exact heads and final promotion candidates.
- Primary: read-only Production audit when connected.
- Coordinator: integrate validated work only, update canonical truth, preserve backup/rollback/deploy gates.

No Production mutation was performed. Concrete work this checkpoint: post-merge main CI was verified green, #110 was closed completed, #108/#101 mergeability and exact-head workflow status were freshly reconciled, and worker/Production queues were refreshed.

## Coordinator checkpoint — 2026-09-14 19:4x Asia/Tehran

- Canonical `main` before this docs refresh: `eccb681f2b21ea1ad3ff74137b47e1991be9db5e`; push CI `34861097897` completed SUCCESS.
- Task13 PR #108 was reconciled non-destructively with this exact main and pushed at refreshed head `44db803d70ad62dd9867d3d3a85d2b2e57f8b429`. Clean merge; local `git diff --check` PASS; Docker Go 1.25 `go vet ./...` + `go test ./... -count=1` PASS. Fresh exact-head CI/Exact Accounting/Pinned Forwardproxy were started and must all finish green before any merge consideration.
- Task13 mandatory real pinned-Caddy HTTP/1.1 + HTTP/2 acceptance remains required on the refreshed exact head; historical protocol evidence is not promoted to exact-head proof.
- Karing PR #101 remains DRAFT; real Karing import -> parse -> CONNECT -> cleanup/revoke is still mandatory.
- Production Primary is not present in fresh connected-device inventory. No deploy/migration/restart/reload/DB/Caddy/credential/backup/rollback mutation was performed; last recorded Production receipt remains the truth ceiling until a fresh read-only audit.
- R1 / STEER-001 remains independent: any new migration is strictly >0030; telemetry stays separate from quota/exact byte-accounting truth.
