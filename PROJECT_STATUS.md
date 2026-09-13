# PVNaive — Canonical Project Status

Last updated: 2026-09-13 23:30 Asia/Tehran

## Verified GitHub state
- Exact verified runtime-bearing `main`: `f0cdab3eda205eecb3285577ec5df32f45d0ddb7`.
- Push CI run `34779141898` for that exact SHA completed SUCCESS, including web, Go, PostgreSQL 18 database gates, runtime rehearsal, and bundle.
- PR #102 / schema22 self-service auth context is MERGED. Before merge its exact head `9975bde902d6b53946485c36d8b766c014266580` passed full CI plus the dedicated Schema22 Auth Context, Task16 Schema21 TDD, WS1 Exact Accounting, and WS1 Pinned Forwardproxy workflows.
- Schema22 repository contract is now internally consistent: latest-schema migration/health/backup/customer/reset fixtures are at 22; rollback metadata/checksum/bookkeeping are valid; self-service actor mutation is bound to authenticated request context.
- PR #101 / Karing remains OPEN/DRAFT at `216d53670066033403fe95f61b0402bb710186a3`; repository/unit/build evidence exists, but real Karing import/parse/CONNECT/cleanup evidence is still required.
- PR #64 / Task13 remains OPEN/DRAFT on a stale pre-current-main head and must be reconstructed before any merge consideration.

## Production truth and hard gates
- No Production host is currently connected through the command channel, so there is no fresh command-level deployed SHA/schema/service/Caddy/backup/rollback receipt in this cycle.
- Repository operator records contain last-recorded live evidence from the later Super-Z batch, but that history is not being promoted to a fresh health assertion.
- `DEPLOY-001` remains OPEN/P0: the boot Caddyfile renderer can drop active customer credentials during recreate/restart unless credential-preserving rendering is proven.
- `LINEAGE-001` remains OPEN/P1: Production migration lineage is recorded ahead/different from current repository schema numbering and must be reconciled before any repository-built deployment.
- Production was not mutated. No backup, migration, restart/reload, credential/DB/Caddy change, deploy, or rollback-state change was performed.

## Work completed this cycle
- Reconciled the newer schema22 security lane and treated its failing full database CI as the immediate integration blocker.
- Repaired stale latest-schema assertions across the generic migration contract, refresh-reuse fixture, health gate, backup/restore gate, customer lifecycle fixture, and periodic reset executor fixture.
- Repaired schema22 rollback metadata, checksum manifest, and rollback bookkeeping so one-step rollback removes schema22 from `schema_migrations` transactionally.
- Re-ran exact-head CI repeatedly until the full PostgreSQL 18 suite, runtime rehearsal, bundle, dedicated auth-context security test, Task16, accounting, and pinned-forwardproxy workflows were green.
- Marked PR #102 ready only after exact-head gates were green, merged it with an expected-head SHA lock, then verified merged-main push CI green.

## Worker / coordinator allocation
- Remote inventory currently has no online executable PVNaive worker; previous Worker 1 is offline and Primary is absent.
- Worker 4 / Karing lane: queued for disposable real-client import → parse → CONNECT → cleanup/revoke proof on PR #101 when connected.
- Worker 3 / Task13 implementation: queued to reconstruct from current verified main after schema22 integration.
- Worker 2 / Task13 verifier: queued for independent exact-head race/permission/HTTP1+HTTP2/accounting rehearsal after reconstruction.
- Primary / Production: queued for read-only audit only when connected.
- Coordinator / GitHub lane: continue repository-safe work on `DEPLOY-001` credential-preserving renderer and `LINEAGE-001` lineage reconciliation without touching Production.

## Promotion order
Resolve `DEPLOY-001` + `LINEAGE-001` in repository and independently validate → Karing real-client receipt → Task13 reconstructed exact-head live proof → fresh Production read-only audit → fresh encrypted backup → independent rollback snapshot → exact deploy-SHA lock → staged promotion → postflight with rollback retained.

Never credit assignment-only, stale-head, static-only, historical, or missing-tool evidence as completion.
