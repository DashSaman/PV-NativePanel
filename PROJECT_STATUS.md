# PVNaive — Canonical Project Status

Last updated: 2026-09-14 12:41 Asia/Tehran

## Verified GitHub state
- Canonical `main` before this documentation refresh: `3ea84fb9d8bdaba7fcd6fd33d1dd44d4ff04d1f6`; push CI `34821820421` SUCCESS.
- Deployed runtime commit remains `a4edea62594d5b60a978c39e1f28fad9ac45f6b6`; its CI `34795216345` is SUCCESS.
- Task13 PR #108 remains OPEN/DRAFT and unmerged. Exact-head repository evidence is historical; branch/main reconciliation plus mandatory real pinned-Caddy HTTP/1.1 + HTTP/2 acceptance remain required.
- Karing PR #101 remains OPEN/DRAFT and unmerged. Mandatory real disposable Karing import → parse → CONNECT → cleanup/revoke evidence is still absent.
- R2 / STEER-002 #110 advanced from proven RED to minimal GREEN implementation on draft PR #112 exact head `2ebe4637bf8f0d4fefb789fc299f5a404cc867fe`. CandidateRatio now filters only `Decision.Candidates`; the full healthy eligible set still drives primary selection/hysteresis. The best node is retained unconditionally so zero/negative score domains cannot produce an empty candidate set. GitHub currently reports the PR mergeable, but exact-head CI `34826791358`, WS1 Exact Accounting `34826791281`, and WS1 Pinned Forwardproxy `34826791365` are still running, so no merge credit is claimed yet.
- Fresh Remote Desktop inventory reports both known PVNaive registrations offline; no worker or Production Primary is currently available through that channel.

## Production truth
- Latest persistent receipt remains `pvnaive:repo-live2` image `76c10697a03b`, schema 30 after forward-only 0029/0030, healthy readiness, recorded nip.io E2E ALL_GREEN, and previous `repo-live` retained for rollback.
- `namir.softarg.ir` remains intentionally deferred until the documented ACME retry window clears on 2026-09-15.
- No fresh Production Primary shell audit exists in this checkpoint, so no new shell-level assertion is made for container identity, migration ledger, encrypted-backup freshness, disk headroom, or rollback snapshot.

## Remaining gates
1. **Task13 #108:** Worker 3 reconciles/refreshes the branch on latest green main and reruns exact-head CI + Exact Accounting + Pinned Forwardproxy if the head moves. Worker 2 then performs real pinned-Caddy HTTP/1.1 + HTTP/2 acceptance proving target-only kill, sibling survival, forged-tuple rejection, repeat idempotency, credential/account survival, unchanged Caddy lifecycle, and exactly-once final accounting.
2. **Karing #101:** Worker 4 runs real disposable Karing import → parse → CONNECT → cleanup/revoke with redacted evidence; only then refresh/reconstruct the minimal validated delta on latest green main and rerun exact-head repository gates if needed.
3. **R1 / STEER-001 #109:** continue independently. Any new DB migration must be strictly >0030. Network telemetry remains separate from exact byte-accounting and quota truth.
4. **R2 / STEER-002 #110 / PR #112:** wait for all three exact-head workflows to complete; then Worker 1 independently reviews config/spec/diff boundaries, Worker 2 reviews non-positive/tie/TopK-vs-Decision behavior, Worker 3 fixes only exact-head failures if any, and Worker 4 performs steering E2E. Keep DRAFT until these gates are green.
5. **Production #100:** when Primary is available, perform read-only deployed image/revision, migration/schema, service/listener/Caddy, backup, disk and rollback audit before any future promotion.

## Worker allocation
- Worker 2: Task13 real HTTP1/HTTP2 acceptance; STEER-002 non-positive/tie/TopK-vs-Decision review; R1 ingest/replay/idempotency.
- Worker 3: Task13 refresh/reconcile; STEER-002 exact-head failure remediation only; otherwise R1 trusted TCP_INFO sampling.
- Worker 4: real Karing acceptance and latest-main reconstruction; later steering/R1 E2E.
- Worker 1: independent diff/schema/security/accounting/CI review and STEER-002 config/spec boundary review.
- Primary: read-only Production audit when connected.
- Coordinator: CI reconciliation, validated integration, canonical documentation and promotion safety.

No runtime merge, deploy, migration, restart/reload, credential rotation, DB/Caddy mutation, backup mutation, or rollback mutation was performed in this checkpoint.