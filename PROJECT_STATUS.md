# PVNaive — Canonical Project Status

Last updated: 2026-09-14 02:32 Asia/Tehran

## Verified GitHub state
- Current canonical `main`: `879219990539b676050877023ceb68f2951f39ea`; push CI `34787434163` completed SUCCESS.
- Batch-2 main reconciled deployed migration lineage through schema 28 (`0022..0027` preserved from Production, hardened functions added as `0028`) and upstreamed the repo-built Docker/runtime config path. `DEPLOY-001` and `LINEAGE-001` are closed by recorded live evidence; do not reopen them from older handoffs.
- Active Task13 lane is draft PR #108 at exact head `41b7bcea78b2f3e1298077e6e7ba4d7cba724bed`, based on current main. Old PRs #64 and #107 are closed superseded, not merged.
- PR #108 exact-head gates: CI `34788082645` SUCCESS, WS1 Exact Accounting `34788082580` SUCCESS, WS1 Pinned Forwardproxy `34788082595` SUCCESS. Independent Worker-1 checks also passed: `git diff --check`, Go 1.25 vet/test, web 19 files / 64 tests + build, session-control permission contract, CI contract and forwardproxy race/session-control test.
- Karing PR #101 remains OPEN/DRAFT at `216d53670066033403fe95f61b0402bb710186a3`; no real Karing client acceptance receipt exists yet.

## Production truth / safety gates
- Batch-2 repository evidence records a repo-built live deploy, schema 28, boot reconcile `CREDENTIALS:22`, derived expected schema 28, live create-customer 201 and strict-TLS CONNECT 204 checks.
- Fresh external read-only probe this cycle: `45.141.148.59.nip.io` returned 200 for API live, API ready (`db:ok`, `schema:ok`) and `/panel/`, with valid TLS.
- `namir.softarg.ir` currently fails TLS handshake with server `tlsv1 alert internal error`. This matches the documented Let's Encrypt duplicate-certificate window; recorded retry-after is `2026-09-15 03:17:36 UTC`. TLS storage persistence is already fixed and the nip.io hostname is the temporary healthy path. Do not recreate/restart merely to chase certificate issuance.
- Production Primary is not connected through the command channel. Current container/image identity, shell-level migration ledger, backup freshness/encryption, disk and rollback snapshot are therefore not freshly command-verified this cycle.
- No Production mutation was performed by this coordinator cycle.

## Remaining promotion gates
1. Task13: obtain independent REAL HTTP/1.1 + HTTP/2 pinned-Caddy proof on PR #108 exact head: target-only kill, sibling survival, forged-tuple rejection, repeat-kill idempotency, credential survival, unchanged Caddy lifecycle and exactly-once final accounting.
2. Karing: disposable real-client import → parse → CONNECT → cleanup/revoke receipt, then reconstruct on latest main if required.
3. Production: fresh Primary read-only audit, then encrypted backup + independent rollback snapshot before any next runtime deploy.
4. Owner-domain TLS: allow the documented ACME window to clear; verify `namir.softarg.ir` externally before switching temporary subscription host back.

## Worker allocation
- Worker 4: real Karing acceptance and cleanup/revoke evidence.
- Worker 3: PR #108 implementation review/fix only if a new exact-head finding appears.
- Worker 2: independent real HTTP/1.1 + HTTP/2 Task13 protocol/accounting proof.
- Worker 1 / `Pak-Nasheeee-haaaaaaaaa`: online; exact-head static/Go-via-Docker/web/forwardproxy verification and safe branch preparation.
- Primary: read-only Production audit when connected.
- Coordinator: CI reconciliation, safe integration, canonical docs and promotion gates.

Never credit stale-head, assignment-only, static-only, inferred or missing-tool evidence as completion.

## 2026-09-14 02:53 coordinator checkpoint
- Verified canonical main `3b49e0b9dd10cd720dbbf33be50361f3ec003dce`; push CI `34789203279` SUCCESS.
- Task13 PR #108 refreshed by fast-forwarding its exact implementation history with current docs/spec main; new head `d42f1db4112fe43e71f4cd1b7feff941d78094af`, GitHub mergeable. Independent `git diff --check`, Docker Go 1.25 gofmt/vet/test and web 19/64 + build PASS. Fresh exact-head CI/accounting/forwardproxy are running; real pinned-Caddy HTTP/1.1 + HTTP/2 acceptance remains mandatory.
- Production remains mutation-free. Latest fresh external read-only probe: nip.io live/ready/panel healthy; `namir.softarg.ir` still inside documented Let's Encrypt retry window. Primary shell-level deployed identity/backups/rollback remain unverified.
- Opened #109 as the independent next roadmap lane: R1 / STEER-001 trusted-boundary network telemetry. Worker 3=forwardproxy sampling, Worker 2=DB/ingest/replay semantics, Worker 1=independent schema/CI/test-harness review, Worker 4=E2E rehearsal, Primary=read-only Production.

## 2026-09-14 04:11 coordinator checkpoint
- Schema29 repair PR #111 is merged. Exact pre-documentation main `21687777e6dd51ccf50650d4360b3c31c05421ad` is fully CI-green via `34793340297`: Go, PostgreSQL18 database/migration/health/backup-restore, web, and full rehearsal all succeeded.
- Task13 PR #108 old exact head `f7d8dd5aa8f33b1bc09e3f19bd26ffb219e650d9` had green old-head CI/accounting/forwardproxy gates but is stale against repaired runtime-bearing main. Worker 3 must reconstruct/reconcile onto exact latest main; Worker 2 must then supply the independent real pinned-Caddy HTTP/1.1 + HTTP/2 target/sibling/forged/idempotency/credential/Caddy/exactly-once-accounting proof. Keep DRAFT; no merge credit from stale-head evidence.
- Karing PR #101 remains DRAFT/non-mergeable at `216d53670066033403fe95f61b0402bb710186a3`; real disposable Karing import → parse → CONNECT → cleanup/revoke evidence is still mandatory before latest-main reconstruction and merge consideration.
- R1 / STEER-001 #109 is activated on the repaired baseline. Migration numbering is now >=0030; never reuse/rewrite 0029. Worker 3 owns RED-first trusted TCP_INFO sampling; Worker 2 owns RED-first DB ingest/replay/idempotency separated from exact byte-accounting truth; Worker 1 independent schema/security/CI review; Worker 4 later E2E.
- Remote executor inventory has no online PVNaive worker or Production Primary at this checkpoint. Production remains mutation-free. Latest recorded external probe keeps nip.io live/ready/panel healthy; owner-domain TLS remains inside the documented ACME retry window. Shell-level deployed image/schema ledger, backup freshness/encryption, disk and rollback snapshot require a fresh Primary read-only audit.
- Next Production promotion sequence is unchanged: complete real Task13 + Karing gates → fresh Primary audit → fresh encrypted backup → independent rollback snapshot → lock exact deploy SHA → staged promotion → health/postflight → retain rollback.

## 2026-09-14 04:42 coordinator checkpoint
- New verified runtime baseline is `60d3a91b40ad06dcde3e72e44ae2e5d3773eebf7`; push CI `34794364541` completed SUCCESS. This commit fixes the migration destructive-SQL guard so SQL-looking tokens inside function/dollar-quoted bodies and ordinary string literals do not false-trigger, while top-level destructive SQL and unclosed dollar quotes remain fail-closed.
- Recorded incident attached to that fix: attempted `repo-live2` boot hit the guard on schema29/0030 content and was rolled back to `repo-live`; the commit record states schema remained 28 and migrations were not applied. Treat this as recorded incident evidence, not a fresh shell-level Production audit.
- Task13 PR #108 remains OPEN/DRAFT/non-mergeable at stale head `f7d8dd5aa8f33b1bc09e3f19bd26ffb219e650d9`. All old-head green evidence remains historical only. Reconstruct/reconcile onto exact latest runtime main and rerun CI + Exact Accounting + Pinned Forwardproxy before the still-mandatory real HTTP/1.1 + HTTP/2 acceptance rehearsal.
- Karing PR #101 remains OPEN/DRAFT/non-mergeable at `216d53670066033403fe95f61b0402bb710186a3`; no disposable real Karing import/parse/CONNECT/cleanup-revoke receipt exists, so no merge or latest-main promotion credit is allowed.
- Remote executor inventory currently shows both known Worker-1 registrations offline; no Worker 2/3/4 or Production Primary is online through the command channel. Persisted assignments were refreshed on PR #108, PR #101, issue #109 and issue #100 for the next available workers.
- R1 / STEER-001 stays independent and unblocked for development on latest green main, with migration numbering >=0030 and network telemetry strictly separated from exact byte-accounting truth. No Production deployment is permitted from that lane without its own exact-head/E2E gates.
- Production was not mutated by this coordinator. No deploy, migration, restart/reload, credential change, backup mutation or rollback-state mutation was performed. Fresh Primary read-only deployed identity/schema/backups/disk/rollback evidence is still required before any staged promotion.

### 04:45 race correction
- A newer runtime-bearing commit, `a4edea62594d5b60a978c39e1f28fad9ac45f6b6`, landed on `main` during this coordinator cycle. It adds explicit JSON tags for `ValidityInput` plus decode regression tests after a live `repo-live2` probe observed product create rejecting `duration_days`/`expires_at` as unknown fields. Its CI run `34795216345` is currently IN_PROGRESS, so it is **not yet promoted to verified-green baseline**.
- Current repository tip after the coordinator documentation commits is `5b4374662a92dc04b2e40f59b555325a591ae1a0`; CI `34795322626` is IN_PROGRESS. Do not deploy or rebuild Task13/Karing against an assumed-green tip until these current-main runs settle.
