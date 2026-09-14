# PVNaive Handoff

Checkpoint: 2026-09-14 02:32 Asia/Tehran

- Verified `main`: `879219990539b676050877023ceb68f2951f39ea`; CI `34787434163` SUCCESS.
- Batch-2 closed the prior renderer/build-provenance and migration-lineage blockers with recorded live evidence: repo lineage is 0001..0028, Production schema 28, boot credential reconciliation preserves 22 active credentials, Docker uses caddy-admin reload, and TLS storage is persistent.
- Active Task13 is draft PR #108, exact head `41b7bcea78b2f3e1298077e6e7ba4d7cba724bed` on current main. CI `34788082645`, Exact Accounting `34788082580`, and Pinned Forwardproxy `34788082595` are SUCCESS. Worker-1 Go/web/contracts/forwardproxy validation is also green. PRs #64/#107 were closed superseded.
- Task13 is NOT merge-ready until an independent real HTTP/1.1 + HTTP/2 pinned-Caddy kill/accounting rehearsal proves target-only termination, sibling survival, forged-tuple rejection, idempotent repeat kill, credential survival, unchanged Caddy lifecycle and exactly-once final accounting.
- Karing PR #101 remains draft and blocked only on a real disposable Karing import/parse/CONNECT/cleanup receipt; static/unit/build evidence is insufficient.
- Fresh external Production probe: nip.io live/ready/panel are 200 and TLS-valid. `namir.softarg.ir` currently fails TLS handshake, consistent with the documented Let's Encrypt duplicate limit; recorded retry-after is 2026-09-15 03:17:36 UTC. Do not restart/recreate to force issuance.
- Production Primary is not connected, so shell-level deployed identity, backup freshness/encryption and rollback snapshot still need a fresh read-only audit. No Production mutation was performed this cycle.

Execution allocation:
- Worker 4 → Karing real-client acceptance.
- Worker 3 → PR #108 review/fixes only for new exact-head findings.
- Worker 2 → real Task13 HTTP1/HTTP2 protocol + accounting proof.
- Worker 1 → exact-head verification / safe branch prep.
- Primary → read-only Production audit.
- Coordinator → CI, integration, docs and promotion safety.

Next runtime promotion requires the missing real Task13 proof, fresh Production audit, encrypted backup and rollback snapshot. Keep rollback retained through postflight.

## 2026-09-14 02:53 coordinator checkpoint
- Verified canonical main `3b49e0b9dd10cd720dbbf33be50361f3ec003dce`; push CI `34789203279` SUCCESS.
- Task13 PR #108 refreshed by fast-forwarding its exact implementation history with current docs/spec main; new head `d42f1db4112fe43e71f4cd1b7feff941d78094af`, GitHub mergeable. Independent `git diff --check`, Docker Go 1.25 gofmt/vet/test and web 19/64 + build PASS. Fresh exact-head CI/accounting/forwardproxy are running; real pinned-Caddy HTTP/1.1 + HTTP/2 acceptance remains mandatory.
- Production remains mutation-free. Latest fresh external read-only probe: nip.io live/ready/panel healthy; `namir.softarg.ir` still inside documented Let's Encrypt retry window. Primary shell-level deployed identity/backups/rollback remain unverified.
- Opened #109 as the independent next roadmap lane: R1 / STEER-001 trusted-boundary network telemetry. Worker 3=forwardproxy sampling, Worker 2=DB/ingest/replay semantics, Worker 1=independent schema/CI/test-harness review, Worker 4=E2E rehearsal, Primary=read-only Production.

## 2026-09-14 04:11 coordinator checkpoint
- Repaired canonical pre-doc main is `21687777e6dd51ccf50650d4360b3c31c05421ad`; CI `34793340297` SUCCESS across Go, PostgreSQL18 migration/health/backup-restore, web and full rehearsal. PR #111 is integrated.
- Task13 PR #108 remains DRAFT. Its prior head `f7d8dd5aa8f33b1bc09e3f19bd26ffb219e650d9` is stale versus repaired runtime-bearing main. Worker 3 must reconstruct/reconcile on exact latest main and rerun exact-head CI/accounting/forwardproxy; Worker 2 then supplies the independent real HTTP1+HTTP2 kill/session/accounting proof. No stale-head completion credit.
- Karing PR #101 remains DRAFT/non-mergeable at `216d53670066033403fe95f61b0402bb710186a3`; Worker 4 still owes a disposable real Karing import/parse/CONNECT/cleanup-revoke receipt before latest-main reconstruction.
- R1 #109 is unblocked by main repair: Worker 3 RED-first authoritative TCP_INFO sampling; Worker 2 RED-first >=0030 DB ingest/replay/idempotency with network samples separate from byte-accounting truth; Worker 1 independent schema/security/CI review; Worker 4 E2E later.
- No executable PVNaive worker or Production Primary is currently online. Production remains unmodified; Primary must first perform a fresh read-only deployed SHA/schema/services/Caddy/socket/backups/disk/rollback audit when connected. Do not restart/recreate to chase owner-domain ACME while its documented retry window remains active.
- Promotion order remains: real Task13 + Karing gates → fresh Production audit → encrypted backup → independent rollback snapshot → exact SHA lock → staged deploy → postflight → retained rollback.
