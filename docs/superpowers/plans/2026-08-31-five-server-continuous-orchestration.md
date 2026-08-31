# Five-Server Continuous Orchestration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Keep all safe, dependency-ready PVNaive work moving across the Owner-provided five-server pool until Release Candidate gates are satisfied.

**Architecture:** Primary owns integration/release; four SSH workers use isolated workspaces for implementation and independent verification. A resource-aware supervisor restarts/reassigns abandoned lanes while preserving Production priority, Git isolation and exact-head validation.

**Tech Stack:** Linux/systemd, SSH, Git/GitHub, Go, Node/npm, PostgreSQL, OpenCode agent workers, PVNaive CI/release scripts.

**Spec:** `docs/superpowers/specs/2026-08-31-five-server-continuous-orchestration-design.md`

## Global Constraints

- Canonical inventory: `docs/DEVELOPMENT_WORKER_POOL.md`.
- Approximate per-host pressure ceiling: 70%; unrelated Production workload always wins.
- No two agents modify the same worktree.
- No secrets in Git/chat/CI/reports.
- Worker output must be transferred/committed before a worker is disposable.
- Production mutation still requires preflight + verified backup + rollback + postflight.
- Never manufacture parallelism that violates roadmap dependencies.

---

### Task 1: Register and verify worker inventory

**Files:**
- Create: `docs/DEVELOPMENT_WORKER_POOL.md`
- Modify: `AGENTS.md`

**Interfaces:**
- Consumes: Primary root SSH aliases and live host identity checks.
- Produces: canonical host/alias/IP/safety policy for every future agent.

- [x] **Step 1: Verify Primary public IPv4 and hostname from the live host.**
- [x] **Step 2: Verify `pv-worker-1..4` mappings from Primary root SSH config and live SSH hostnames.**
- [x] **Step 3: Record five-server inventory without private-key material.**
- [ ] **Step 4: Add the inventory to the mandatory `AGENTS.md` read order and execution rules.**
- [ ] **Step 5: Merge the documentation PR after diff review.**

### Task 2: Fill all safe lanes

**Files:**
- Update as evidence changes: `WORKLOG.md`, `AGENT_TASKS.md`

**Interfaces:**
- Consumes: current roadmap dependencies and host resource samples.
- Produces: one isolated assignment per available safe lane.

- [x] **Step 1: Inspect CPU/load/RAM/disk and running PVNaive agents across all five hosts.**
- [x] **Step 2: Preserve active Task 12/13/14 work rather than restarting it.**
- [x] **Step 3: Dispatch Worker 1 independent Task 12 verification under CPU/memory limits.**
- [x] **Step 4: Dispatch Worker 2 isolated P0 BUG-001 security work under CPU/memory limits.**
- [ ] **Step 5: Harvest reports/results and record exact evidence in `WORKLOG.md`.**

### Task 3: Install continuity supervision on Primary

**Files:**
- Create: `ops/dev/pvnaive-continuous-orchestrator.sh`
- Create: `ops/systemd/pvnaive-dev-orchestrator.service`
- Create: `ops/systemd/pvnaive-dev-orchestrator.path` or timer only if justified by the implementation
- Test: `tests/ops/continuous_orchestrator_test.sh`

**Interfaces:**
- Consumes: `docs/DEVELOPMENT_WORKER_POOL.md`, `ROADMAP.md`, `AGENT_TASKS.md`, Primary SSH aliases.
- Produces: supervised orchestrator process that can recover from an unexpected agent exit without duplicate workspace writers.

- [ ] **Step 1: Write a failing shell contract test that requires single-instance locking, resource gate, restart-safe state/report paths, and no embedded secrets.**
- [ ] **Step 2: Run the contract test and capture RED evidence.**
- [ ] **Step 3: Implement a minimal wrapper that acquires one lock, checks host pressure, launches the configured orchestrator command, and records exit state without deleting worker workspaces.**
- [ ] **Step 4: Add a systemd unit with `Restart=always`, bounded CPU/memory, low scheduling priority and explicit environment/config paths.**
- [ ] **Step 5: Run shell validation and contract tests to GREEN.**
- [ ] **Step 6: Install on Primary without touching PVNaive customer/runtime/Caddy/PostgreSQL services; verify restart after a controlled wrapper exit.**
- [ ] **Step 7: Record the live unit status and rollback/removal command in `WORKLOG.md`.**

### Task 4: Continuous dispatch/reconciliation loop

**Files:**
- Modify as task evidence changes: `AGENT_TASKS.md`, `WORKLOG.md`, `PROJECT_STATUS.md`, `ROADMAP.md`, `HANDOFF.md`, `KNOWN_ISSUES.md`

**Interfaces:**
- Consumes: worker reports, current `main`, PR/CI state and Production evidence.
- Produces: validated merged work and next ready assignments.

- [ ] **Step 1: On every worker completion, inspect report + diff + exact base before integration.**
- [ ] **Step 2: Re-run required verification on a different lane.**
- [ ] **Step 3: Integrate only onto latest main, create PR, require exact-head CI and resolve regressions.**
- [ ] **Step 4: For Production-facing work, perform backup/deploy/postflight/rollback gates.**
- [ ] **Step 5: Immediately assign the next dependency-ready work unit to free capacity.**
- [ ] **Step 6: Reconcile canonical docs after each completed master task.**

### Task 5: Stop condition

**Files:**
- Modify: canonical status/handoff/roadmap/worklog files.

**Interfaces:**
- Consumes: Release Candidate evidence.
- Produces: explicit RC completion or a documented genuine external blocker.

- [ ] **Step 1: Verify all P0 blockers and required P1 gates against actual main/CI/Production evidence.**
- [ ] **Step 2: Verify no important unmerged worker workspace/report remains.**
- [ ] **Step 3: Verify clean install, upgrade, rollback, backup/restore, Karing compatibility, capacity and final Production smoke gates.**
- [ ] **Step 4: Mark RC only after canonical docs contain no stale contradiction.**
