# Continue Here — PVNaive

Verified checkpoint: 2026-09-14 18:42 Asia/Tehran

## Current GitHub truth
- Canonical main before this documentation refresh was `714b2f5490b6eec317061b0834a2721e2c6a8fa6`; push CI `34848486540` SUCCESS.
- STEER-002 / #110 is completed and closed. PR #112 merged via code-bearing `f6b1bab91fa1583a647770a766c7cf9d58f9ee89`; exact pre-merge head `9c43a71fcf4990eb8a9c053221fd9701c10f2caa` passed CI `34842627049`, Exact Accounting `34842627114`, and Pinned Forwardproxy `34842627193`.
- Task13 #108 remains DRAFT/unmerged at `f7d8dd5aa8f33b1bc09e3f19bd26ffb219e650d9`; fresh metadata says `mergeable=true`. Exact-head CI `34789937594`, Exact Accounting `34789937603`, and Pinned Forwardproxy `34789937575` are SUCCESS. Real pinned-Caddy HTTP/1.1 + HTTP/2 acceptance is still mandatory.
- Karing #101 remains DRAFT/unmerged at `216d53670066033403fe95f61b0402bb710186a3`; fresh metadata says `mergeable=true`. Exact-head CI `34732580376`, Exact Accounting `34732580468`, and Pinned Forwardproxy `34732580377` are SUCCESS. Real disposable Karing import → parse → CONNECT → cleanup/revoke is still mandatory.

## Production truth
- Latest persistent deployment receipt remains `pvnaive:repo-live2` image `76c10697a03b`, runtime `a4edea62594d5b60a978c39e1f28fad9ac45f6b6`, schema 30 after forward-only 0029/0030, recorded healthy readiness + nip.io E2E ALL_GREEN, previous `repo-live` retained for rollback.
- Fresh Remote Desktop inventory shows both known registrations offline and no Production Primary available. Do not assert current live image, migration ledger, encrypted-backup freshness, disk headroom or rollback snapshot beyond the persistent receipt.

## Active lanes
- **Task13 #108:** Worker 2 performs real pinned-Caddy HTTP/1.1 + HTTP/2 acceptance: target-only kill, sibling survival, forged-tuple rejection, repeat idempotency, credential/account survival, unchanged Caddy lifecycle, exactly-once final accounting. Worker 3 reconciles only for a concrete merge-ref/exact-head defect. Worker 1 performs final independent review after protocol evidence exists.
- **Karing #101:** Worker 4 performs real disposable Karing import → parse → CONNECT → cleanup/revoke with client/platform/version, generated-profile SHA-256 and redacted logs. Static/unit/build proof is not a substitute.
- **R1 / STEER-001 #109:** continue independently. New migrations must be strictly >0030; never reuse/rewrite 0029/0030. Worker 3 = trusted TCP_INFO sampling from authoritative socket/RemoteAddr state; Worker 2 = DB ingest/replay/idempotency; Worker 1 = schema/privilege/security/accounting review; Worker 4 = E2E after an exact implementation head exists. Telemetry remains separate from exact byte-accounting/quota truth.
- **Production #100:** Primary performs read-only audit first when connected. Before any future runtime deploy: fresh encrypted backup → independent rollback snapshot → exact deploy SHA lock → staged promotion → postflight → retain rollback.

No Production deploy, migration, restart/reload, credential rotation, DB/Caddy mutation, backup mutation or rollback mutation was performed in this checkpoint.

## Coordinator checkpoint — 2026-09-14 19:4x Asia/Tehran

- Canonical `main` before this docs refresh: `eccb681f2b21ea1ad3ff74137b47e1991be9db5e`; push CI `34861097897` completed SUCCESS.
- Task13 PR #108 was reconciled non-destructively with this exact main and pushed at refreshed head `44db803d70ad62dd9867d3d3a85d2b2e57f8b429`. Clean merge; local `git diff --check` PASS; Docker Go 1.25 `go vet ./...` + `go test ./... -count=1` PASS. Fresh exact-head CI/Exact Accounting/Pinned Forwardproxy were started and must all finish green before any merge consideration.
- Task13 mandatory real pinned-Caddy HTTP/1.1 + HTTP/2 acceptance remains required on the refreshed exact head; historical protocol evidence is not promoted to exact-head proof.
- Karing PR #101 remains DRAFT; real Karing import -> parse -> CONNECT -> cleanup/revoke is still mandatory.
- Production Primary is not present in fresh connected-device inventory. No deploy/migration/restart/reload/DB/Caddy/credential/backup/rollback mutation was performed; last recorded Production receipt remains the truth ceiling until a fresh read-only audit.
- R1 / STEER-001 remains independent: any new migration is strictly >0030; telemetry stays separate from quota/exact byte-accounting truth.
