# Five-Server Continuous Orchestration Design

Date: 2026-08-31

## Goal

Use the Owner-provided five-server pool continuously and safely to complete PVNaive to Release Candidate without turning temporary development workers into product fleet architecture or endangering unrelated Production workloads.

## Architecture

Primary (`testAmir5-3`) is the single integration/release authority. Four SSH-reachable workers execute isolated implementation, test and review lanes. Every coding lane gets a separate workspace/worktree; worker output is never merged merely because an agent reports success. Integration remains gated by independent review, exact-head tests/CI and Production safety requirements.

The canonical inventory and host-specific constraints live in `docs/DEVELOPMENT_WORKER_POOL.md`.

## Scheduler

A lane is one of:

- implementation;
- independent verification;
- schema/DB/security review;
- Web/build/regression;
- integration/release.

The scheduler continuously chooses ready work from the Owner-ordered roadmap. It may parallelize only tasks whose file/dependency boundaries are compatible. Dependency order is more important than artificial CPU utilization.

Before heavy work, each host is sampled for CPU/load, RAM, disk and existing PVNaive/unrelated jobs. Approximately 70% host pressure is the normal upper target. When unrelated Production workload rises, PVNaive work yields by lowering concurrency/quota or moving to another host.

## Continuity

Long-running worker/orchestrator processes are supervised so unexpected process exit does not silently abandon a lane. A supervisor must:

1. detect exited/failed worker processes;
2. preserve reports/workspaces;
3. avoid duplicate writers on one workspace;
4. restart/reassign only when the task is still valid against current main;
5. stop scheduling only when Release Candidate gates are actually complete or a genuine external blocker is recorded.

Continuity is best-effort across software-process failures; it cannot guarantee execution through physical host loss, network/provider outage, credential expiry, disk exhaustion, or unavailable external services.

## Data and Git safety

- GitHub `main` is the repository source of truth.
- Temporary workers are never the sole durable copy of important work.
- No force-push/reset of `main`.
- No two agents write the same worktree.
- Workers do not receive unnecessary Production secrets.
- Production DB/test boundaries remain explicit; disposable PostgreSQL is preferred for migration/security tests.

## Production boundary

Development pool authorization does not relax Production mutation rules. Before Production changes, perform current preflight, fresh verified backup, exact migration/release validation, rollback preparation and postflight. Do not modify real customer quota/password/token simply to prove a feature; use isolated canaries.

## Completion

The scheduler may stop assigning work only when `ROADMAP.md` Release Candidate gates are reconciled against real code/CI/Production evidence and the canonical handoff/status/worklog contain no unresolved release blocker or unique unmerged worker output.
