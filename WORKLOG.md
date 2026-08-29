# PVNaive — Worklog

Last updated: 2026-08-30

This log records significant verified transitions so future agents do not repeat completed work. Fine-grained historical evidence remains in git history, PR discussions and `ops/evidence/`; this file keeps the current durable milestone chain.

## Historical foundation milestones

| Date | Work | Evidence / result |
|---|---|---|
| 2026-08-27 | PVNaive naming, filesystem/service foundation, PostgreSQL 18/RLS, encrypted DB backup/restore, S03 deployment | historical S00-S03 evidence; DONE |
| 2026-08-28 | S04 auth foundation, Owner bootstrap/lifecycle, protected public panel/API | auth/web/DB/rehearsal evidence; DONE |
| 2026-08-28 | S04R Runtime credential management | AES-GCM Runtime secret, pinned Caddy parser/renderer, Runtime Agent, expected-SHA validate/backup/reload/rollback, revision saga, Owner API/UI; DONE |
| 2026-08-28 | S04R exact pinned Caddy multi-auth proof + bundle/rehearsal | CI `33190295766` checkpoint and artifact checksum evidence; DONE |
| 2026-08-29 | S05/S06 customer operations and read-only delivery | customer create/adopt/edit/lifecycle, quota/validity, read-only Subscription/QR; Production schema advanced through later release line; DONE baseline |

## 2026-08-29 — Parallel workstream integration

| Work / PR | What changed | Verification / current truth | Result |
|---|---|---|---|
| WS3 PR #15 | deterministic `/sub/<token>` machine output, `/s/<token>` human Account Page, local QR, security headers, explicit Subscription/password separation | tested WS3 head; real Karing app smoke intentionally still external follow-up | MERGED / CLIENT ACCEPTANCE PARTIAL |
| WS1 PR #17 | exact direct-Naive accounting, schema 9, first-successful-CONNECT semantics, session/presence projection, hard-quota primitives, dedicated telemetry boundary, pinned forwardproxy patch | source head documented CI #913 + pinned-forwardproxy success | MERGED |
| WS1 PR #18 | pending-reservation accounting completeness edge | exact-accounting/finalization CI follow-up | MERGED |
| WS1 PR #19 | shared runtime directory/accounting socket permission fix | production preflight exposed real permission issue; regression/fix followed | MERGED |
| WS1 PR #20 | bit-reproducible pinned accounting Caddy | two independent absolute workspaces required identical binary/SHA provenance | MERGED |
| WS1 PR #21 | preserve telemetry socket across Runtime Agent restart | production restart rehearsal exposed socket lifecycle bug; preservation fix | MERGED |
| WS1 PR #23 | replace shared RuntimeDirectory ownership with tmpfiles namespace | production CONNECT smoke exposed re-ownership of `accounting.sock`; durable tmpfiles fix proven live | MERGED |
| WS2 PR #22 | schema 11 customer product management | plans/renewal/search/bulk/groups/tags/reseller/RBAC foundations + Runtime UUID mapping synchronization | MERGED |
| PR #24 | encrypted backup validation SIGPIPE fix | real production backup validation moved to private temp file + `pg_restore --list FILE`; cleanup regression | MERGED |
| PR #25 | activate merged customer product features in protected panel | customer/product UI + exact-accounting read-model integration work | MERGED |
| PR #26 | fix customer list and polish customer/subscription UI | restored existing Production users in customer directory, compact operations UI, `/s` vs `/sub` UX cleanup, Persian account-page default; Go/Web/Production service evidence in PR | MERGED as `a021aa4b62c35b775fb521d042b2f8e6dbde10b0` |

## Current Production read-only audit — 2026-08-30

Lead audit against `testAmir5-3` / `namir.softarg.ir` recorded only non-secret state:

- `pvnaive-api.service`: active;
- `pvnaive-runtime-agent.service`: active;
- `pvnaive-telemetry-agent.service`: active;
- `caddy-naive.service`: active;
- current observed service restart counters: 0;
- API listener: loopback `127.0.0.1:8080`;
- local readiness: HTTP 200;
- public panel: HTTP 200;
- public readiness: HTTP 200;
- PostgreSQL schema version: 11;
- active users: 6;
- active Runtime credentials: 6;
- ServiceTerms: 6;
- active direct Subscription tokens: 6;
- direct accounting term projections: 6 complete / 0 incomplete at audit instant;
- append-only direct-accounting events: tens of thousands and actively growing;
- direct-accounting session history present with active/open rows;
- legacy `usage_ledger`: not the active direct-Naive accounting path;
- backup files exist under `/var/backups/pvnaive`;
- no PVNaive scheduled-backup systemd timer observed;
- root filesystem: 79% used at audit instant;
- Runtime/accounting Unix sockets remain distinct permission boundaries;
- Production deploy marker files lag newer binary/web mtimes, so release provenance is not authoritative yet.

No raw credential, password, Subscription token, encryption key or secret-bearing Caddy content was printed or committed.

## Current competitor audit — 2026-08-30

Official current snapshots re-checked beyond README:

- 3x-ui `f727d04f6522bb94a8fb52e8352fdcafb51c11e1` / v3.7.0;
- PasarGuard `aebf7256927710329d380d67ce96224f287ae5f6` / v5.3.0;
- Hiddify Manager default `dev@a99c811aa63fe908f1e06607b81f475b502ebf07`; stable v12.3.3;
- OV-PvNetwork `5b6a578bfe7733ebc67c08d9c431da6e32ac7ced` / public v1.0.0-rc1.

Result: new 120-feature matrix added at `docs/PANEL_PARITY_MASTER_2026-08-30.md`. GPL/AGPL source remains reference-only.

## Current PR / CI reconciliation — PR #27

Branch: `lead/parity-truth-2026-08-30`

Starting main: `a021aa4b62c35b775fb521d042b2f8e6dbde10b0`

Reason:

- PR #26 final bot head `317ad74a5f7402bfff7de0716b5c1a4b246a6e5a` recorded failures for normal CI + WS1 workflows;
- immediately preceding human commit `06c3c697bd4a7a988dab634b24a7dbef7614e947` passed all three;
- prior bot commit `a51b8797c13745bdcff19b20a032fceea86de84a` also passed all three;
- therefore Lead created a clean current-main PR to reproduce instead of calling it a code regression or bypassing tests.

PR #27 milestone commits so far:

| Commit | Change |
|---|---|
| `154799745918471b761b8b561c08b408ca44d26a` | Lead reconciliation execution plan |
| `d4fefa45cfdcd99263ba5eff0394af3943f63106` | refreshed 120-feature current competitor parity matrix |
| `6904efbe85489c5528f8a187b9b3829d54ad7c01` | short canonical feature matrix reconciled |
| `7fe62fdc936ecadee81983ae560cc174d5b862ee` | canonical project status reconciled |
| `293f999208ba4976c8cafde247541a90c20a7513` | handoff replaced with current schema 11/Production truth |
| `b951c8bc2b441962f95c24a6e313ef6aee9f8b56` | continuation pointer replaced |
| `8f8b08c1d7dfe4a2590216b636bd3fedc2dfa946` | known issues reconciled; obsolete accounting blocker closed; real P0 bugs retained |
| `9be0f3bb2d0929fbfa930d12c1ca56c9641b7ce3` | Agent board aligned to Owner 50-step sequence |
| `21f3b5a03ddfdef0251158ce3155687dfa96edc0` | roadmap realigned + legacy PVN crosswalk |

CI observation during this reconciliation:

- normal PR workflows are triggering on PR #27;
- WS1 Exact Accounting completed successfully on parity-matrix head `d4fefa45cfdcd99263ba5eff0394af3943f63106`;
- pinned-forwardproxy and normal CI were still running at the time of this log update;
- no final claim is allowed until all required workflows are green on the final exact documentation head.

## Re-audited current P0 defects — 2026-08-30

Lead source inspection confirmed these are still real:

1. refresh-token reuse-family bug: `refresh` → `BeginAuthenticated`, and `BeginAuthenticated` requires `revoked_at IS NULL` before rotation reuse detection;
2. generic commit-before-success bug: authenticated middleware runs handler before commit and ignores final `Tx.Commit()` error;
3. readiness remains configuration-based rather than bounded DB/schema-backed.

See `KNOWN_ISSUES.md` for exact done gates.

## Old PR classification

| PR | Classification | Evidence / action |
|---:|---|---|
| #4 | STILL USEFUL — small extract only | old branch adds `buildKaringSingBoxProfile` and explicit Copy Karing config UX not present on main; reconsider during client lane, do not merge wholesale |
| #5 | SUPERSEDED / MERGED ELSEWHERE | newer customer lifecycle exists on main |
| #6 | SUPERSEDED / MERGED ELSEWHERE | newer customer/product/subscription flow exists on main |
| #8 | SUPERSEDED / ARCHIVE | replaced by integrated WS1 accounting implementation |
| #16 | STILL USEFUL — manual extraction | 38 files include metrics, request middleware, OpenAPI, observability, Doctor, scheduled backup/restore drill, deploy/rollback, load rehearsal, notifications, fleet and system dashboard; old branch must never be blind-merged |

## Exact continuation

1. finish canonical docs update on PR #27;
2. trigger/check all workflows on exact final head;
3. review PR #27 and merge only when green;
4. create fresh PR #16 integration branch from latest main;
5. inspect/extract useful units one-by-one with TDD/full CI;
6. then continue Owner order: legacy accounting baseline → `/s` accounting → Manual Reset Usage → Bulk Reset Usage → periodic reset → hard-quota proof → first-CONNECT proof → sessions/limits → remaining roadmap.

## Logging rule

Every verified transition adds date, task, exact source/commit, change, tests/CI, Production evidence if applicable, result and exact next step. Real failures that reveal a defect remain in history; never erase them merely to make the log look green.
