# AGENTS.md — PVNaive mandatory agent instructions

## Mission

PVNaive is currently a **standalone-first** management plane for standard NaiveProxy on one external server. Controller/fleet/multi-node and Iran-specific topology are not R1 dependencies. Architecture may support future adapters, but unsupported capabilities must never be presented as implemented.

## Single source of truth / read order

Before any change, read in this order:

1. `PROJECT_STATUS.md` — canonical development snapshot and numerical progress.
2. `HANDOFF.md` — exact current task and continuation.
3. `ROADMAP.md` — permanent `PVN-*` task IDs, priority, dependency, Done gates.
4. `KNOWN_ISSUES.md` — open bugs/security/debt/test/deploy risks.
5. `AGENT_TASKS.md` — workstream ownership and file-conflict rules.
6. `WORKLOG.md` — significant completed/failed work; do not repeat it.
7. `FEATURE_MATRIX.md` — actual-vs-target competitor gap analysis.
8. relevant design/spec/plan under `docs/superpowers/`.
9. `docs/ARCHITECTURE_FA.md`, `docs/DECISIONS_FA.md`, `docs/EXTENSIBILITY_FA.md`, `docs/COVER_TRAFFIC_AND_CONTENT_FA.md` as architectural background.
10. before **any production action**, independently read `main:CONTINUE_HERE.md`, `main:ops/S04_LIVE_STATE.md`, and newest `main:ops/evidence/*`.

`AGENT_HANDOFF.md`, `ops/DEPLOYMENT_PROGRESS.md`, `docs/ROADMAP_FA.md`, `docs/FEATURE_MATRIX_FA.md`, `docs/PRODUCT_GAPS_FA.md` remain useful historical documents but may lag the canonical files above. Never use a stale file alone to infer live state.

## Permanent task IDs

- Every development unit belongs to a stable `PVN-*` ID in `ROADMAP.md`.
- Bugs/security/test/deploy debt may additionally use `BUG-*`, `SECURITY-*`, `TEST-*`, `DEPLOY-*`, `TECH-DEBT-*`, but must map to a `PVN-*` task.
- Do not renumber completed IDs.
- Do not create duplicate work before checking ROADMAP, WORKLOG, open PR/branch changes and current task ownership.

## Before starting a task

Report internally/in notes:

```text
AGENT
TASK-ID
GOAL
FILES
DEPENDENCIES
```

Then:

1. inspect current code, not only filenames/docs;
2. verify no other active task owns the same files;
3. for production code/bugfixes, follow TDD: failing test → observe correct RED → minimal implementation → GREEN;
4. preserve unrelated user/agent changes;
5. record architecture changes in relevant spec/decision/handoff.

## Mandatory work report

After a work unit record:

```text
STATUS
CHANGES
FILES MODIFIED
TESTS
RESULT
NEXT STEP
BLOCKERS
```

No agent may say only “Done”. Final DONE transition belongs to Lead/Agent-REVIEW after evidence review.

## Red lines

- Do not make R1 depend on Controller, fleet or Iran topology.
- Do not alter Naive wire protocol without ADR, benchmark and client-compatibility evidence.
- Do not build default random chaff/fake browsing traffic.
- Do not hardcode public-site topic in installer/binary; use content-pack abstraction.
- Do not use estimated access-log traffic as exact billing.
- Do not commit passwords, tokens, subscription secrets, runtime secrets, private keys, raw secret-bearing Caddyfiles or production dumps.
- Do not place Web UI/API on the data-plane availability path.
- Do not close SSH or casually change firewall.
- Do not run destructive migration/uninstall without backup + validation + rollback plan.
- Do not use unpinned `latest` for production dependencies/artifacts.
- Do not fabricate unsupported accounting/session/device/speed/quota capabilities.
- Do not let unprivileged API gain arbitrary root shell/path/service/URL access.
- Caddy live changes are `validate → exact backup → install → reload-only → verify → exact rollback`; never restart unless a separately approved emergency procedure proves it necessary.
- Do not reset/force-push branch history to resolve `main`/`s04-auth` divergence.

## Production interaction rule

On target `testAmir5-3`, give **one server command/step at a time**. Wait for the Owner to paste full output before interpreting or issuing the next live command. Prefer read-only preflight first. Do not ask the Owner to paste secrets.

No server mutation is allowed just because code compiles or CI passes. Each stage requires its own production preflight, backup/rollback and independent postflight.

## Current critical execution chain

Read `HANDOFF.md` for exact current head. At this snapshot:

`PVN-020 → PVN-021 → PVN-022 → PVN-023 → PVN-024 → PVN-025 → PVN-026 → PVN-027 → PVN-028 → PVN-029`

Then critical auth hardening `PVN-030/031/034` and formal S04 closure `PVN-036` before user/business stages.

## Definition of Done

A task is not DONE unless all applicable items are true:

- implementation is complete;
- diff/code review completed;
- targeted tests and full relevant regression tests pass;
- failure paths are exercised;
- security/secret/authorization implications reviewed;
- integration remains healthy;
- migration/rollback documented and tested when applicable;
- UI is responsive/accessibility-checked when applicable;
- docs reflect actual behavior, not target behavior;
- `WORKLOG.md` updated;
- `ROADMAP.md` status updated;
- `PROJECT_STATUS.md` counts updated;
- `HANDOFF.md` exact continuation updated;
- fresh verification evidence exists before claiming completion.

## Context recovery

If context may be lost, update repository state **before anything else**: WORKLOG, PROJECT_STATUS, AGENT_TASKS if ownership changed, ROADMAP, HANDOFF and current test/CI evidence. A new Chat/Agent must be able to continue without the old conversation.
