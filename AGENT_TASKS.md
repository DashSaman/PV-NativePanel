# PVNaive — Agent / Workstream Task Board

Last updated: 2026-08-28

The current Chat environment does **not** expose an independent sub-agent execution tool. Therefore the roles below are durable workstreams/ownership labels, not a claim that Claude/Gemini/Qwen/DeepSeek are currently running. If a future environment provides real isolated agents, dispatch only independent tasks and keep these ownership boundaries.

## Mandatory report format

Before work:

```text
AGENT: <role>
TASK-ID: <PVN-ID>
GOAL: <one outcome>
FILES: <allowed scope>
DEPENDENCIES: <IDs/evidence>
```

After a work unit:

```text
STATUS: RED | GREEN | BLOCKED | REVIEW
CHANGES: <what changed>
FILES MODIFIED: <paths>
TESTS: <commands/CI run>
RESULT: <evidence, not just Done>
NEXT STEP: <exact next action>
BLOCKERS: <none or evidence>
```

No workstream may mark a task `DONE`; final verification/ledger update belongs to Agent-REVIEW / Lead Engineer.

## Active assignments

| Workstream | Task | Status | Scope | Handoff condition |
|---|---|---|---|---|
| Agent-PM / Lead | PVN-068 | IN_PROGRESS | canonical audit/docs/task ledger | all PM files + AGENTS committed, cross-links verified |
| Agent-BACKEND/SEC | PVN-020 | IN_PROGRESS | `internal/runtimecred/*` only for current slice | targeted tests green + full Go/CI green |
| Agent-QA | PVN-020 | WAITING_ON_IMPL | validate existing RED then GREEN; secret leakage/policy cases | CI evidence |
| Agent-ARCH | PVN-021 | READY_AFTER_020 | Caddy parser/renderer contract | failing tests first; byte preservation/injection proof |
| Agent-DEVOPS | PVN-022/023 | WAITING | Unix socket agent and privileged operator | fixed capability boundary + reload-only rehearsal |
| Agent-DB | PVN-024 | WAITING | runtime store/revision saga | compensation and idempotency tests |
| Agent-BACKEND | PVN-025 | WAITING | typed Owner-only API | RBAC/CSRF/idempotency/revision tests |
| Agent-FRONTEND | PVN-026 | WAITING | `/runtime/naive` UI | no secret leak; one-time generated secret UX |
| Agent-QA/DEVOPS | PVN-027 | WAITING | disposable full S04R rehearsal | all runtime agent/Caddy order/failure gates |
| Agent-REVIEW | PVN-028/029 | WAITING | live read-only preflight then guarded import | one server step at a time; evidence before mutation |
| Agent-SEC | PVN-030/031/034 | READY_AFTER_S04R_CODE | critical auth fixes | RED→GREEN regressions + full auth rehearsal |
| Agent-DOCS | PVN-069 | READY | reconcile stale product/security/API docs | actual vs planned clearly labeled |
| Agent-ARCH/REVIEW | PVN-070 | WAIT_GREEN_CHECKPOINT | branch divergence integration | no force/reset; clean reviewed diff + CI |
| Agent-QA/SEC | PVN-071 | FUTURE | authorization/E2E/fuzz gates | CI enforced |
| Agent-PM/OWNER | PVN-072 | NEEDS_OWNER_BUSINESS_DECISION | license policy | explicit license choice + NOTICE strategy |

## File conflict rules

- `internal/runtimecred/*`: PVN-020, then PVN-024; do not parallel-edit those tasks.
- `internal/runtimeagent/*`: PVN-022 then PVN-023; can design tests separately but integrate sequentially.
- `internal/httpapi/*`: PVN-025 and auth-hardening tasks must be coordinated; do not parallel-write `server.go`/auth handlers.
- `web/src/App.tsx`/routing: only one frontend task at a time until routing is decomposed.
- `db/migrations/*`: migrations are append-only after release; checksum changes must be reviewed.
- canonical PM files may be updated after any verified task, but sequential writes to the same file are required.

## Workstream responsibilities

### Agent-ARCH

Architecture, capability boundaries, ADR/spec consistency, avoiding feature bloat. Current next architecture task: `PVN-021` after PVN-020 green.

### Agent-BACKEND

Go domain/service/API code. Must use strict DTOs, fail closed, no secret logging, and transaction integrity.

### Agent-FRONTEND

React/TypeScript UI. Only display verified capabilities; placeholders must be explicit; accessibility/responsive checks required.

### Agent-DB

PostgreSQL migrations/RLS/ledger/saga persistence. PostgreSQL 18 disposable tests and one-step rollback evidence required before live use.

### Agent-SEC

Auth/authorization/secret boundaries, threat-model regression tests, supply-chain gates. Security findings are not waived for schedule pressure.

### Agent-QA

RED→GREEN evidence, regression/failure injection, authorization matrix, parser/fuzz/property cases, full CI verification.

### Agent-DEVOPS

systemd, installer, Caddy lifecycle, backup/restore, update/rollback. Production actions are one step at a time and never infer success from code alone.

### Agent-RESEARCH

Competitor/reference research and licensing notes. Prefer source/commit-pinned evidence; distinguish public implementation from claims/production-derived behavior.

### Agent-DOCS

Actual-vs-target documentation and handoff; never turn a plan or route registry into a feature claim.

### Agent-REVIEW / Lead Engineer

Reviews diff, tests, regression, security, integration and docs before changing `ROADMAP.md` status to DONE. Uses `verification-before-completion`; evidence before assertion.

## Current execution queue

1. Finish `PVN-068` canonical PM layer.
2. Continue `PVN-020` from observed RED; implement minimal secret/policy/types code.
3. Fresh full CI. If green, update PM docs and mark PVN-020 DONE.
4. Immediately start `PVN-021` by writing failing parser/renderer tests.
5. Continue dependency chain without requesting routine technical approval.

Production is untouched until `PVN-027` is fully green and `PVN-028` read-only preflight is explicitly ready.
