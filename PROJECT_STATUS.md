# PVNaive — Canonical Project Status

Last updated: 2026-09-14 11:40 Asia/Tehran

## Verified GitHub state
- Canonical `main` before this documentation refresh: `638844337e814e806cbe90897e522aadc6d43863`; push CI `34817253202` SUCCESS.
- Deployed runtime commit remains `a4edea62594d5b60a978c39e1f28fad9ac45f6b6`; its CI `34795216345` is SUCCESS.
- Task13 PR #108 is OPEN/DRAFT at exact head `f7d8dd5aa8f33b1bc09e3f19bd26ffb219e650d9`; GitHub currently reports `mergeable=false`. Exact-head repository workflows are historical-green but branch/main reconciliation plus mandatory real pinned-Caddy HTTP/1.1 + HTTP/2 acceptance are still absent.
- Karing PR #101 is OPEN/DRAFT at exact head `216d53670066033403fe95f61b0402bb710186a3`; GitHub currently reports `mergeable=false`. Mandatory real disposable Karing import → parse → CONNECT → cleanup/revoke evidence is still absent.
- R2 / STEER-002 #110 advanced to draft PR #112. RED-only head `7a7923cf8faf6f03a2ea1578570a705f9df57153` adds two CandidateRatio contract regressions: below-threshold candidates must be excluded, and the best candidate must survive non-positive score domains. WS1 Exact Accounting is already SUCCESS; CI and Pinned Forwardproxy are still running, so no implementation or merge credit is claimed yet.
- Fresh Remote Desktop inventory reports both known PVNaive registrations offline; no worker or Production Primary is currently available through that channel.

## Production truth
- Latest persistent receipt remains `pvnaive:repo-live2` image `76c10697a03b`, schema 30 after forward-only 0029/0030, healthy readiness, recorded nip.io E2E ALL_GREEN, and previous `repo-live` retained for rollback.
- `namir.softarg.ir` remains intentionally deferred until the documented ACME retry window clears on 2026-09-15.
- No fresh Production Primary shell audit exists in this checkpoint, so no new shell-level assertion is made for container identity, migration ledger, encrypted-backup freshness, disk headroom, or rollback snapshot.

## Remaining gates
1. **Task13 #108:** Worker 3 refreshes/reconciles the branch on latest green main and reruns exact-head CI + Exact Accounting + Pinned Forwardproxy if the head moves. Worker 2 then performs real pinned-Caddy HTTP/1.1 + HTTP/2 acceptance proving target-only kill, sibling survival, forged-tuple rejection, repeat idempotency, credential/account survival, unchanged Caddy lifecycle, and exactly-once final accounting.
2. **Karing #101:** Worker 4 runs real disposable Karing import → parse → CONNECT → cleanup/revoke with redacted evidence; only then reconstruct/refresh the minimal validated delta on latest green main and rerun exact-head repository gates.
3. **R1 / STEER-001 #109:** continue independently. Any new DB migration must be strictly >0030. Network telemetry remains separate from exact byte-accounting and quota truth.
4. **R2 / STEER-002 #110 / PR #112:** observe the intended RED failure first. Then Worker 3 adds the smallest candidate-set filtering fix; Worker 2 independently checks zero/negative score behavior and TopK/Decision consistency; Worker 1 reviews config/spec boundaries; Worker 4 performs steering E2E. Preserve Unknown/hysteresis/kill-switch/accounting semantics.
5. **Production #100:** when Primary is available, perform read-only deployed image/revision, migration/schema, service/listener/Caddy, backup, disk and rollback audit before any future promotion.

## Worker allocation
- Worker 2: Task13 real HTTP1/HTTP2 acceptance after refreshed green exact head; STEER-002 non-positive score and TopK/Decision review.
- Worker 3: Task13 refresh/reconcile; STEER-002 minimal implementation after RED is proven; otherwise R1 trusted TCP_INFO sampling.
- Worker 4: real Karing acceptance and latest-main reconstruction; later steering/R1 E2E.
- Worker 1: independent diff/schema/security/accounting/CI review and STEER-002 config/spec boundary review.
- Primary: read-only Production audit when connected.
- Coordinator: CI reconciliation, validated integration, canonical documentation and promotion safety.

No runtime merge, deploy, migration, restart/reload, credential rotation, DB/Caddy mutation, backup mutation, or rollback mutation was performed in this checkpoint.