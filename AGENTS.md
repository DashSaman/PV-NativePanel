# AGENTS.md — PVNaive mandatory agent instructions

## Mission

PVNaive is a **standalone-first** management plane for standard NaiveProxy on one external server. Controller/fleet/multi-node and Iran-specific topology are not R1 dependencies. Unsupported accounting/session/device/speed/quota capabilities must never be presented as implemented.

## Single source of truth / read order

Before any change, read in this order:

1. `PROJECT_STATUS.md` — canonical development snapshot and numerical progress.
2. `HANDOFF.md` — exact current task and continuation.
3. `ROADMAP.md` — permanent `PVN-*` task IDs, priority, dependency, Done gates.
4. `docs/PILOT_INSTALL_FA.md` — current existing-server S04R Pilot install/runbook.
5. `KNOWN_ISSUES.md` — open bugs/security/debt/test/deploy risks.
6. `AGENT_TASKS.md` — workstream ownership and file-conflict rules.
7. `WORKLOG.md` — significant completed/failed work; do not repeat it.
8. `FEATURE_MATRIX.md` — actual-vs-target gap analysis.
9. relevant design/spec/plan under `docs/superpowers/`.
10. before **any production action**, independently read `main:CONTINUE_HERE.md`, `main:ops/S04_LIVE_STATE.md`, and newest `main:ops/evidence/*`.

Older handoff/roadmap/feature documents may lag the canonical files above. Never use a stale file alone to infer live state.

## Permanent task IDs

- Every development unit belongs to a stable `PVN-*` ID in `ROADMAP.md`.
- Bugs/security/test/deploy debt may additionally use `BUG-*`, `SECURITY-*`, `TEST-*`, `DEPLOY-*`, `TECH-DEBT-*`, but must map to a `PVN-*` task.
- Do not renumber completed IDs.
- Do not duplicate work before checking ROADMAP, WORKLOG, PR/branch changes and ownership.

## Before starting a task

Record:

```text
AGENT
TASK-ID
GOAL
FILES
DEPENDENCIES
```

Then:

1. inspect current code, not only docs;
2. verify no active task owns the same files;
3. for production code/bugfixes, follow TDD: failing test → observe correct RED → minimal implementation → GREEN;
4. preserve unrelated changes;
5. record architecture/safety changes in relevant spec/handoff.

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

- Do not make R1 depend on Controller/fleet/Iran topology.
- Do not alter Naive wire protocol without ADR, benchmark and client-compatibility evidence.
- Do not build default random chaff/fake browsing traffic.
- Do not use estimated access-log traffic as exact billing.
- Do not commit passwords, tokens, subscription secrets, runtime secrets, private keys, raw secret-bearing Caddyfiles or production dumps.
- Do not place Web UI/API on the data-plane availability path.
- Do not close SSH or casually change firewall.
- Do not run destructive migration/uninstall without backup + validation + rollback plan.
- Do not use unpinned `latest` for production dependencies/artifacts.
- Do not fabricate unsupported accounting/session/device/speed/quota capabilities.
- Do not let the unprivileged API gain arbitrary root shell/path/service/URL access.
- Caddy live credential changes are `expected SHA → exact backup → validate → install → reload-only → verify → exact rollback`; never restart as routine behavior.
- Do not reset/force-push branch history to resolve `main`/`s04-auth` divergence.

## Production interaction rule

On target `testAmir5-3`, use **one server command/step at a time**. Wait for full output before interpreting or issuing the next live mutation. Prefer read-only preflight first. Never ask the Owner to paste secrets.

No server mutation is allowed just because code or CI passes. Each stage requires its own live preflight, backup/rollback and postflight.

For the current Pilot use `docs/PILOT_INSTALL_FA.md`.

## Current critical execution chain

Development/rehearsal is complete through `PVN-027`.

Current edge:

`PVN-028 read-only live preflight → PVN-029 guarded S04R upgrade/import/Karing smoke`

Only after the Pilot evidence:

`PVN-030 refresh reuse fix + PVN-031 commit integrity + PVN-032 DB readiness + PVN-033 recovery decision + PVN-034 rate limit → PVN-035 security review → PVN-036 formal S04 closure`

Then continue user/business/accounting/subscription stages.

## Pilot capability boundary

After PVN-029 passes live, the Owner may create and hand a customer a Naive credential/link. The Pilot is **not** a customer self-service portal and does not yet implement quota, exact usage, expiry, reseller, subscription lifecycle, device/session/speed controls or billing.

Never give customers Owner panel credentials.

## Definition of Done

A task is not DONE unless all applicable items are true:

- implementation is complete;
- diff/code review completed;
- targeted and relevant regression tests pass;
- failure paths are exercised;
- security/secret/authorization implications reviewed;
- integration remains healthy;
- migration/rollback documented and tested when applicable;
- UI build/responsive behavior is checked when applicable;
- docs reflect actual behavior, not target behavior;
- WORKLOG/ROADMAP/PROJECT_STATUS/HANDOFF updated;
- fresh verification evidence exists before completion claims.

## Context recovery

If context may be lost, update repository state first. A new Chat/Agent must be able to continue from canonical files without the old conversation.
