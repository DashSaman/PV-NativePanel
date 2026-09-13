# PVNaive — Canonical Project Status

Last updated: 2026-09-13 23:35 Asia/Tehran

## Verified GitHub state
- Current canonical `main` tip: `61c613777ec2dcf3f1559cbb8b8df80f8e389af0`; push CI run `34779414593` completed SUCCESS.
- Exact verified runtime-bearing baseline remains `f0cdab3eda205eecb3285577ec5df32f45d0ddb7`; its merged-main CI `34779141898` completed SUCCESS across web, Go, PostgreSQL 18 database gates, runtime rehearsal, and bundle.
- PR #102 / schema22 self-service auth context is MERGED and schema22 is the canonical repository schema.
- PR #101 / Karing remains OPEN/DRAFT at `216d53670066033403fe95f61b0402bb710186a3`; exact-head CI/accounting/forwardproxy are green, but it is stale/non-mergeable against current main and still lacks real Karing import/parse/CONNECT/cleanup evidence.
- PR #64 / Task13 remains OPEN/DRAFT at stale head `3fc14825e1b164bad558decaef47f56b792e81af`; reconstruct from current verified main before any merge consideration.

## Production truth and hard gates
- No Production host is currently connected through the command channel; there is no fresh command-level deployed SHA/schema/service/Caddy/backup/rollback receipt this cycle.
- `DEPLOY-001` is canonical issue #105. Issue #103 was closed as a duplicate to prevent split accounting. The authoritative all-in-one boot/recreate renderer source is still outside this repository; do not fabricate a repository fix without recovering trusted source provenance.
- `LINEAGE-001` remains issue #104. Historical Production records indicate applied migrations through 0027 with different numbering/content; exact read-only ledger/checksums and trusted 0022..0027 artifacts are required before designing forward compatibility.
- Production was not mutated: no backup creation, migration, restart/reload, credential/DB/Caddy change, deploy, or rollback-state change.

## Work completed this cycle
- Reconciled current main and confirmed CI `34779414593` SUCCESS on exact tip `61c61377...`.
- Rechecked open PRs and exact heads; neither runtime PR satisfies promotion gates.
- Verified remote inventory: no online executable PVNaive worker; Production Primary remains absent.
- Consolidated duplicate DEPLOY-001 tracking by closing #103 as duplicate of #105.
- Posted fresh coordinator receipts to #105, #104, #101, #64 and #100 with exact current baseline, blockers and next evidence requirements.
- Started a clean Karing reconstruction branch from current main for future worker use; no partial runtime change was committed because the full UI delta could not be safely reconstructed without a complete exact-source worktree and executable validation.

## Worker / coordinator allocation
- Worker 4 / Karing lane: queued for disposable real-client import → parse → CONNECT → cleanup/revoke proof, then reconstruct the four-file delta on latest verified main if needed.
- Worker 3 / Task13 implementation: queued to reconstruct from latest verified main.
- Worker 2 / Task13 verifier: queued for independent exact-head race/permission/HTTP1+HTTP2/accounting rehearsal.
- Worker 1: queued for independent evidence/security/accounting review where installed tools are sufficient.
- Primary / Production: queued for read-only audit only when connected.
- Coordinator / GitHub lane: continue repository-safe source-provenance recovery for DEPLOY-001 and non-destructive LINEAGE-001 reconciliation.

## Promotion order
Recover and validate DEPLOY-001 source + reconcile LINEAGE-001 → Karing real-client receipt → Task13 reconstructed exact-head live proof → fresh Production read-only audit → fresh encrypted backup → independent rollback snapshot → exact deploy-SHA lock → staged promotion → postflight with rollback retained.

Never credit assignment-only, stale-head, static-only, historical, missing-tool or inferred evidence as completion.
