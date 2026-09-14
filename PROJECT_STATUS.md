# PVNaive — Canonical Project Status

Last updated: 2026-09-14 16:47 Asia/Tehran

## Verified GitHub state
- Canonical `main` is `f6b1bab91fa1583a647770a766c7cf9d58f9ee89`, the guarded merge commit for STEER-002 PR #112. Main push CI `34848316732` is currently in progress; do not call this main tip post-merge green until that run completes successfully.
- STEER-002 PR #112 was promoted only after exact head `9c43a71fcf4990eb8a9c053221fd9701c10f2caa` completed CI `34842627049` SUCCESS, WS1 Exact Accounting `34842627114` SUCCESS, and WS1 Pinned Forwardproxy `34842627193` SUCCESS. Independent review confirmed no code-path overlap with the 15 intervening main commits, existing one-eligible/all-Unknown/no-flapping/hysteresis/kill-switch coverage, and the new CandidateRatio/TopK decision contract regression. PR #112 is merged; Production was not touched.
- Deployed runtime commit remains `a4edea62594d5b60a978c39e1f28fad9ac45f6b6`; its CI `34795216345` is SUCCESS.
- Task13 PR #108 remains OPEN/DRAFT, unmerged and currently `mergeable=false` on head `f7d8dd5aa8f33b1bc09e3f19bd26ffb219e650d9`. Real pinned-Caddy HTTP/1.1 + HTTP/2 acceptance plus reconciliation onto the latest green main remain mandatory.
- Karing PR #101 remains OPEN/DRAFT, unmerged and currently `mergeable=false` on head `216d53670066033403fe95f61b0402bb710186a3`. Mandatory real disposable Karing import → parse → CONNECT → cleanup/revoke evidence remains absent.
- R1 / STEER-001 #109 continues independently. Production schema is 30; new DB work must use a migration strictly >0030 and never rewrite/reuse 0029 or 0030. Network telemetry is not exact byte-accounting/quota truth.
- Fresh Remote Desktop inventory reports both known PVNaive registrations offline; no worker or Production Primary is currently executable through that channel.

## Production truth
- Latest persistent receipt remains `pvnaive:repo-live2` image `76c10697a03b`, runtime `a4edea62594d5b60a978c39e1f28fad9ac45f6b6`, schema 30 after forward-only 0029/0030, healthy readiness, recorded nip.io E2E ALL_GREEN, and previous `repo-live` retained for rollback.
- No fresh Production Primary shell audit exists in this checkpoint, so no new shell-level assertion is made for container identity, migration ledger, encrypted-backup freshness, disk headroom, or rollback snapshot.
- `namir.softarg.ir` remains deferred until the documented ACME retry window clears; do not restart/recreate Production merely to chase certificate issuance.

## Remaining gates
1. **Main after STEER-002 merge:** wait for push CI `34848316732` on `f6b1bab9...`; if green, close #110 as integrated. If red, remediate only the concrete failure before any deployment discussion.
2. **Task13 #108:** Worker 3 reconciles/refreshes onto the latest CI-green main and reruns exact-head CI + Exact Accounting + Pinned Forwardproxy if the head changes. Worker 2 then runs real pinned-Caddy HTTP/1.1 + HTTP/2 acceptance proving target-only kill, sibling survival, forged-tuple rejection, repeat idempotency, credential/account survival, unchanged Caddy lifecycle, and exactly-once final accounting.
3. **Karing #101:** Worker 4 runs real disposable Karing import → parse → CONNECT → cleanup/revoke with redacted evidence; only then refresh/reconstruct the minimal validated delta on the latest green main if needed.
4. **R1 / STEER-001 #109:** Worker 3 = trusted TCP_INFO sampling; Worker 2 = DB ingest/replay/idempotency; Worker 1 = schema/security/accounting-boundary review; Worker 4 = E2E after an exact implementation head exists.
5. **Production #100:** when Primary reconnects, perform read-only deployed image/revision, migration/schema, service/listener/Caddy, encrypted-backup, disk and rollback audit before any future promotion.

## Worker allocation
- Worker 2: Task13 real HTTP1/HTTP2 protocol/accounting acceptance; R1 ingest/replay/idempotency.
- Worker 3: Task13 branch/main reconciliation; otherwise R1 trusted TCP_INFO sampling.
- Worker 4: real Karing client acceptance; R1/steering E2E follow-up.
- Worker 1: independent diff/schema/security/accounting/CI review for promoted or newly produced exact heads.
- Primary: read-only Production audit when connected.
- Coordinator: reconcile main CI, close STEER-002 only after green post-merge CI, integrate validated work, maintain canonical docs and promotion safety.

Concrete progress this checkpoint: PR #112 was independently re-reviewed, promoted from draft and merged with expected-head protection into main. No Production deploy, migration, restart/reload, credential rotation, DB/Caddy mutation, backup mutation or rollback mutation was performed.