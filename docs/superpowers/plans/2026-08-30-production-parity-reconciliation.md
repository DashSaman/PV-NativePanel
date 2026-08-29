# PVNaive Production / Parity Reconciliation Plan

Date: 2026-08-30
Branch: `lead/parity-truth-2026-08-30`
Starting main: `a021aa4b62c35b775fb521d042b2f8e6dbde10b0`

## Goal

Re-establish repository/production truth before new feature work, refresh competitor parity against current official snapshots, reconcile stale project status documents, reproduce the exact-main CI gate on an isolated branch, and prepare a safe commit-by-commit integration lane for useful PR #16 work.

This plan implements the Owner master prompt ordering. It does not mutate Production.

## Safety invariants

- No direct commits to `main`.
- No Production mutation in this reconciliation branch.
- No secret/password/token/key output or commit.
- Route declarations, schema columns and placeholders are not counted as implemented capabilities.
- Existing working Runtime/accounting/customer/subscription behavior is preserved rather than rewritten.
- GPL/AGPL competitors are behavior/architecture references only unless license compatibility is explicitly approved.
- A feature is only marked DONE when code/tests/authorization/UI where applicable and evidence support it.

## Task 1 — Repository + Production truth audit

1. Confirm current `main`, recent commits, open/merged PRs and workflow evidence.
2. Read canonical status/design/workstream documents requested by the Owner.
3. Read-only Production inspection of API, PostgreSQL, Caddy, Runtime Agent, Telemetry Agent, web panel, service state, accounting, disk and backup state.
4. Record provenance drift or operational risks without attempting a deployment.
5. Reproduce CI using this clean branch. A fresh PR must run the normal workflows; do not infer green exact-main status from older runs.

Current audit evidence to carry forward:

- main = `a021aa4b62c35b775fb521d042b2f8e6dbde10b0`.
- Production DB schema = 11.
- API, Runtime Agent, Telemetry Agent and Caddy are active; panel/readiness return 200.
- Production has six active users/runtime credentials/service terms and six complete direct-accounting term projections.
- Exact-accounting event/session data is present and active; no fake usage is required.
- No PVNaive scheduled-backup timer was observed; existing backup files are present.
- root filesystem was observed at 79% used and must be considered before backup/load-test work.
- Production deploy markers are stale/inconsistent with newer binaries/web deployment evidence; provenance needs hardening.
- PR #26 final bot commit had failed workflow records, while its immediately preceding human commit and prior bot commit passed CI + both WS1 workflows. Root cause must be reproduced rather than guessed.

## Task 2 — Current competitor parity matrix

Refresh official reference snapshots and rewrite the parity matrix using source/API/model evidence:

- PVNaive: `a021aa4b62c35b775fb521d042b2f8e6dbde10b0`.
- 3x-ui: `f727d04f6522bb94a8fb52e8352fdcafb51c11e1` / v3.7.0.
- PasarGuard: `aebf7256927710329d380d67ce96224f287ae5f6` / v5.3.0.
- Hiddify Manager: default `dev` at `a99c811aa63fe908f1e06607b81f475b502ebf07`, stable v12.3.3.
- OV-PvNetwork: `5b6a578bfe7733ebc67c08d9c431da6e32ac7ced` / public v1.0.0-rc1 snapshot.

Required matrix columns:

`Feature | PVNaive | 3x-ui | PasarGuard | Hiddify | OV-PvNetwork | Priority | Action`

Cover the Owner's 120-feature list plus Hiddify-specific domain/CDN/Cloudflare/WARP/proxy-mode items. Mark protocol-specific or irrelevant capabilities OPTIONAL / PROTOCOL-SPECIFIC / N/A rather than adding them for feature count.

## Task 3 — Documentation truth reconciliation

After source/prod audit, update canonical truth files:

- `FEATURE_MATRIX.md`
- `PROJECT_STATUS.md`
- `HANDOFF.md`
- `AGENT_TASKS.md`
- `ROADMAP.md`
- `KNOWN_ISSUES.md`
- `WORKLOG.md`
- `CONTINUE_HERE.md`
- parity master/gap matrix as needed

Key rules:

- Do not leave implemented customer CRUD, exact accounting, plans/groups/tags, `/sub`, `/s`, QR or other proven features marked missing.
- Do not promote manual usage reset, periodic reset execution, session kill/limits, scheduled backup, Karing real-client acceptance, reseller wallet/ledger UI or other unproven features to DONE.
- Keep schema/route foundations distinct from complete product features.

## Task 4 — PR #16 reconciliation design

Classify old PRs #4/#5/#6/#8 and document whether they are SUPERSEDED, MERGED ELSEWHERE, STILL USEFUL or ARCHIVE.

For PR #16:

1. List changed files and commits.
2. Compare each useful unit to current main.
3. Never merge the stale branch wholesale.
4. Extract only capabilities still absent from current main on a fresh integration branch.
5. Preserve newer main behavior and schema 11.
6. TDD each extracted unit and require full CI before merge.

## Verification gate for this branch

Before this branch can merge:

- parity/status documents match current code and Production evidence;
- no secrets are present;
- fresh PR workflows for the exact branch head have completed successfully, or any workflow infrastructure failure has a documented root cause and fix;
- no runtime/Production mutation has occurred;
- next exact implementation task is stated unambiguously.

## After merge

Proceed in the Owner-mandated sequence:

1. safe PR #16 extraction/integration;
2. legacy accounting baseline truth;
3. `/s` accounting projection truth;
4. manual reset usage (TDD);
5. bulk reset usage;
6. periodic reset scheduler;
7. hard quota production proof;
8. first-successful-CONNECT production proof;
9. session management and limits;
10. reseller/RBAC/ledger and remaining ordered backlog.
