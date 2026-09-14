# PVNaive — Canonical Project Status

Last updated: 2026-09-14 20:10 Asia/Tehran

## Checkpoint 2026-09-14 (Master Upgrade Pack live checkpoint — verified)

- Production runs `pvnaive:repo-fin2-163115` (commit `bc3b319` provenance, deployed 16:34 UTC via
  backup -> pg_dump -> image swap with TLS persistence -> force-recreate). Schema truth: **33**
  after forward-only migrations 0031 (R1 telemetry), 0032 (R2/R3 steering decisions), 0033
  (R5 pool registry). Postflight ALL GREEN (external): panel 200, owner login 200, SSE 3
  frames/3s, real-customer CONNECT 204x2 through the full naive path, 20 accounting events/5min.
- LIVE subsystems: R1 network telemetry (session_network_samples + EWMA aggregates), R2 steering
  scheduler (60s tick; users=0 expected until R1 samples pass MinSamples=8), R3 renderer
  (mihomo proxy-provider profile, spec-exact naive links), R5 pool registry backend (enrollment
  tokens, signed monotonic revisions, drain state machine), R7 panel access, R8 SSE stream.
- Migration gates now run pre-deploy on the server (scratch postgres:18 container):
  steering_decisions gate + pool_registry gate both PASS. The gates caught and rejected five
  real pre-production defects (headers, gofmt, quoted nullable args, superuser-mutability
  check, plpgsql shadowing) — none reached production.
- Push state: 9 commits sit on local `main` rebased onto origin/main `eccb681`
  (56de993-rebased chain through 6afbd9c: R1, R7, R8, R2 scheduler, BUG-STREAM-001,
  BUG-ACCT-001, R3, R2/R3 sink, R5 registry + R6 rehearsal). Owner's fine-grained PAT still
  lacks Contents:write (API blob probe 403) — the permission edit must be SAVED in GitHub
  ("Update token"), then `git push` from the sandbox completes immediately.
- Remaining: R5 UI (wizard/node list), R5-PULL-001 (sibling mTLS pull listener), R6 flip
  (coverd process + Caddy root route; core+rehearsal done, CAMO-001/002 proven by test),
  `namir.softarg.ir` flip after the 2026-09-15 ~03:21-03:26 UTC LE window.


## Verified GitHub state
- Canonical `main` before this documentation refresh is `714b2f5490b6eec317061b0834a2721e2c6a8fa6`; push CI `34848486540` completed SUCCESS.
- STEER-002 / issue #110 is completed and closed. PR #112 merged through code-bearing commit `f6b1bab91fa1583a647770a766c7cf9d58f9ee89`; exact pre-merge head `9c43a71fcf4990eb8a9c053221fd9701c10f2caa` passed CI `34842627049`, WS1 Exact Accounting `34842627114`, and WS1 Pinned Forwardproxy `34842627193`. Post-merge main CI is green. Production was not touched for this lane.
- Task13 PR #108 remains OPEN/DRAFT and unmerged on `f7d8dd5aa8f33b1bc09e3f19bd26ffb219e650d9`. Fresh GitHub metadata reports `mergeable=true`. Exact-head CI `34789937594`, WS1 Exact Accounting `34789937603`, and WS1 Pinned Forwardproxy `34789937575` are SUCCESS. Mandatory real pinned-Caddy HTTP/1.1 + HTTP/2 acceptance is still absent, so merge remains blocked.
- Karing PR #101 remains OPEN/DRAFT and unmerged on `216d53670066033403fe95f61b0402bb710186a3`. Fresh GitHub metadata reports `mergeable=true`. Exact-head CI `34732580376`, WS1 Exact Accounting `34732580468`, and WS1 Pinned Forwardproxy `34732580377` are SUCCESS. Mandatory real disposable Karing import → parse → CONNECT → cleanup/revoke acceptance is still absent.
- R1 / STEER-001 #109 is the next independent code roadmap lane. Production schema truth is 30; all new DB work must use migration strictly >0030 and never rewrite/reuse 0029 or 0030. Network telemetry remains separate from exact byte-accounting/quota truth.
- Fresh Remote Desktop inventory reports both known registrations offline and no Production Primary available.

## Production truth
- Latest persistent receipt remains `pvnaive:repo-live2` image `76c10697a03b`, runtime `a4edea62594d5b60a978c39e1f28fad9ac45f6b6`, schema 30 after forward-only 0029/0030, recorded healthy readiness, nip.io E2E ALL_GREEN, and previous `repo-live` retained for rollback.
- No fresh Production Primary shell audit exists in this checkpoint, so no new assertion is made for live container identity, migration ledger, encrypted-backup freshness, disk headroom or rollback snapshot.
- `namir.softarg.ir` remains deferred until the documented ACME retry window clears; do not restart/recreate Production merely to chase certificate issuance.

## Remaining gates
1. **Task13 #108:** Worker 2 runs the mandatory real pinned-Caddy HTTP/1.1 + HTTP/2 acceptance proving target-only kill, sibling survival, forged-tuple rejection, repeat-kill idempotency, credential/account survival, unchanged Caddy lifecycle, and exactly-once final accounting. Worker 3 only reconciles if a new merge-ref/exact-head defect appears. Worker 1 performs final independent diff/security/accounting review after protocol evidence exists.
2. **Karing #101:** Worker 4 runs real disposable Karing import → parse → CONNECT → cleanup/revoke and records client/platform/version, generated-profile SHA-256 and redacted logs. Unit/build proof is not a substitute.
3. **R1 / STEER-001 #109:** Worker 3 = trusted TCP_INFO sampling from authoritative socket/RemoteAddr state; Worker 2 = >0030 DB ingest/replay/idempotency; Worker 1 = schema/privilege/security/accounting-boundary review; Worker 4 = disposable E2E after an exact implementation head exists.
4. **Production #100:** when Primary reconnects, perform read-only live image/revision, migration/schema, services/listeners/Caddy, encrypted-backup, disk and rollback audit before any future promotion.

## Worker allocation
- Worker 2: Task13 real HTTP1/HTTP2 protocol/accounting acceptance; R1 ingest/replay/idempotency.
- Worker 3: R1 trusted TCP_INFO sampling; Task13 reconciliation only for a concrete new defect.
- Worker 4: real Karing client acceptance; R1 E2E follow-up.
- Worker 1: independent diff/schema/security/accounting/CI review for new exact heads and final promotion candidates.
- Primary: read-only Production audit when connected.
- Coordinator: integrate only validated work, preserve promotion gates, and maintain canonical docs.

Concrete progress this checkpoint: post-merge main CI was verified green; STEER-002 #110 was closed completed; stale mergeability state for #108/#101 was corrected to current `mergeable=true`; exact-head workflow success for both PRs was re-verified; persistent worker/Production queues were refreshed. No Production deploy, migration, restart/reload, credential rotation, DB/Caddy mutation, backup mutation or rollback mutation was performed.

## Coordinator checkpoint — 2026-09-14 19:4x Asia/Tehran

- Canonical `main` before this docs refresh: `eccb681f2b21ea1ad3ff74137b47e1991be9db5e`; push CI `34861097897` completed SUCCESS.
- Task13 PR #108 was reconciled non-destructively with this exact main and pushed at refreshed head `44db803d70ad62dd9867d3d3a85d2b2e57f8b429`. Clean merge; local `git diff --check` PASS; Docker Go 1.25 `go vet ./...` + `go test ./... -count=1` PASS. Fresh exact-head CI/Exact Accounting/Pinned Forwardproxy were started and must all finish green before any merge consideration.
- Task13 mandatory real pinned-Caddy HTTP/1.1 + HTTP/2 acceptance remains required on the refreshed exact head; historical protocol evidence is not promoted to exact-head proof.
- Karing PR #101 remains DRAFT; real Karing import -> parse -> CONNECT -> cleanup/revoke is still mandatory.
- Production Primary is not present in fresh connected-device inventory. No deploy/migration/restart/reload/DB/Caddy/credential/backup/rollback mutation was performed; last recorded Production receipt remains the truth ceiling until a fresh read-only audit.
- R1 / STEER-001 remains independent: any new migration is strictly >0030; telemetry stays separate from quota/exact byte-accounting truth.

## Coordinator checkpoint — 2026-09-14 19:5x Asia/Tehran

- One execution worker is online again; the second known registration remains offline. Production Primary is still absent from fresh inventory.
- Task13 PR #108 refreshed exact head: `44db803d70ad62dd9867d3d3a85d2b2e57f8b429`; local exact-tree diff/vet/test PASS. Fresh WS1 Exact Accounting `34867408480` SUCCESS; CI `34867408371` and Pinned Forwardproxy `34867408392` were still in progress at this checkpoint. Real HTTP/1.1 + HTTP/2 pinned-Caddy acceptance remains mandatory before merge.
- Karing PR #101 refreshed exact head: `6691392639be9fae4656a861db4a6d16580f850d`; local exact-tree diff/vet/test PASS. Fresh exact-head CI/Accounting/Pinned workflows are running. Real Karing import -> parse -> CONNECT -> cleanup/revoke remains mandatory before merge.
- R1 / STEER-001 was redispatched across trusted TCP_INFO, >0030 ingest/replay/idempotency, independent security/schema/accounting review, and later disposable E2E. Network telemetry remains outside quota/exact byte-accounting truth.
- No Production mutation was performed. On Primary reconnect, perform read-only live image/schema/backup/disk/rollback audit before any deploy decision.

## Coordinator checkpoint — 2026-09-14 20:0x Asia/Tehran

- Canonical docs baseline `b79ad34abf90903a3b476edf1342f7d3ce79d360` completed CI SUCCESS (`34867971143`).
- Task13 exact head `44db803d70ad62dd9867d3d3a85d2b2e57f8b429`: CI `34867408371` SUCCESS; Exact Accounting `34867408480` SUCCESS; Pinned Forwardproxy `34867408392` SUCCESS on retry attempt 2. Attempt 1 was a transient sum.golang.org HTTP/2 INTERNAL_ERROR, not a source failure. Independent local exact-head reproducible build also PASS with binary SHA256 `629f58b192fcceac9b1ada6887ad2f53c397f872bb28d28856a90b66f86d99ec`. PR remains DRAFT because real HTTP/1.1 + HTTP/2 target-only kill/sibling-survival/idempotency/forged-tuple/credential-survival/exactly-once-accounting acceptance still has not been rerun on this exact head.
- Karing exact head `6691392639be9fae4656a861db4a6d16580f850d`: CI `34867775973` SUCCESS; Exact Accounting `34867775798` SUCCESS; Pinned Forwardproxy `34867775829` SUCCESS. PR remains DRAFT pending real Karing import -> parse -> CONNECT -> cleanup/revoke.
- Production Primary remains disconnected; no Production mutation. Last recorded Production receipt remains the truth ceiling until read-only audit after reconnect.
- One execution worker is online; R1 / STEER-001 tasks remain dispatched with migration strictly >0030 and telemetry isolated from quota/exact byte-accounting truth.
