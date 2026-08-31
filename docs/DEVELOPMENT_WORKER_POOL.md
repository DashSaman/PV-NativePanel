# PVNaive — Five-Server Development Worker Pool

Last verified: 2026-08-31 UTC

This file is the canonical inventory and operating policy for the temporary development/orchestration servers available to PVNaive. These machines are development workers unless explicitly documented otherwise. They are **not** automatically PVNaive fleet nodes and must not be treated as product architecture.

## Canonical inventory

| Lane | SSH name | Hostname | Public IPv4 | Primary responsibility |
|---|---|---|---|---|
| Primary | local / `testAmir5-3` | `testAmir5-3` | `91.107.182.147` | orchestration, integration, Git/PR/CI, Production-safe release and final verification |
| Worker 1 | `pv-worker-1` | `Pak-Nasheeee-haaaaaaaaa` | `91.107.138.246` | independent verification, focused backend/Web tests, lightweight review; also a safe spare lane |
| Worker 2 | `pv-worker-2` | `RoboT` | `91.107.240.235` | Go/security/regression work, isolated implementation, disposable test work |
| Worker 3 | `pv-worker-3` | `TrPaqet` | `188.132.174.199` | schema/DB/accounting/security/static review and isolated feature work |
| Worker 4 | `pv-worker-4` | `ubuntu-4gb-hel1-1` | `65.109.182.111` | Web/build/integration/regression/disposable DB and isolated feature work |

The Primary has root SSH aliases for all four workers. Do not copy its private SSH keys into Git, chat, CI, worker reports, or application artifacts.

## Standard workspace

Workers use isolated PVNaive workspaces under:

`/opt/pvnaive-worker/workspace`

Worker-local Go/Node toolchains are intentionally separate from unrelated host applications. Source `/opt/pvnaive-worker/env.sh` where present before running PVNaive tests/builds.

Important work must never exist only on a temporary worker. A validated result must be committed/pushed through the normal Git workflow or transferred back to Primary before a worker is discarded.

## Resource scheduler policy

The Owner authorizes active use of all five servers. The target ceiling is approximately **70% total CPU/RAM pressure per host**, dynamically adjusted around unrelated Production workloads.

Before dispatching a meaningful job, inspect at least:

- CPU/load and core count;
- available RAM/swap;
- root/workspace disk availability;
- top unrelated Production processes;
- existing PVNaive agent/test jobs.

Rules:

1. Use free capacity aggressively while the host remains below the safe ceiling.
2. Production workload always wins. If unrelated services push the host toward the ceiling, throttle, pause, or move PVNaive work.
3. Prefer `systemd-run`, `CPUQuota`, `MemoryMax`, `nice`, `ionice`, and bounded timeouts for heavy transient jobs.
4. Never stop/restart/reconfigure unrelated services merely to free capacity for development.
5. No two agents may modify the same worktree.
6. Independent verification should preferably run on a different worker from the implementation lane.
7. Keep at least one usable verification/recovery lane when a task affects schema, Runtime, Caddy, auth, accounting, quota, or release safety.

## Host-specific safety notes

### Worker 1 — `Pak-Nasheeee-haaaaaaaaa`

This host has unrelated WordPress/Apache workload. Prior Apache/WordPress pressure tuning and swap configuration must not be undone. Use dynamic resource limits and back off when the site workload increases.

### Worker 2 — `RoboT`

Unrelated services may include Xray, x-ui, MySQL, Apache, Hedioum, GRE watchdog and PVNetwork services. Do not stop/restart/reconfigure them for PVNaive development.

### Worker 3 — `TrPaqet`

Unrelated services may include Paqet, Xray, x-ui, OpenVPN, WaterWall, tunnel watchdogs and OV-node components. Do not modify them. Prefer isolated DB/schema/accounting/security work.

### Worker 4 — `ubuntu-4gb-hel1-1`

Unrelated services may include Paqet, Xray, x-ui, OpenVPN, WaterWall, PostgreSQL, Nginx, OV Panel and tunnel services. Do not modify those services. Use isolated workspace/tooling and disposable DBs only.

### Primary — `testAmir5-3`

Primary is also the PVNaive Production host. Keep orchestration and integration lightweight. Any PVNaive Production mutation still requires the repository's normal preflight, fresh backup, rollback and postflight gates. Development worker management does not waive Production safety.

## Continuous execution / no-idle policy

The project should operate as a queue, not as one long single-agent task.

Execution loop:

`INSPECT -> DECOMPOSE -> DISPATCH -> IMPLEMENT -> INDEPENDENT VERIFY -> REVIEW -> INTEGRATE -> CI -> BACKUP -> DEPLOY -> POSTFLIGHT -> DOCUMENT -> NEXT`

While release work remains:

- when a worker finishes, assign the next non-conflicting ready task;
- when an agent fails or exits, preserve its workspace/report, diagnose the failure, then restart or reassign the task;
- do not leave a worker idle merely because another lane is busy if an independent ready task exists;
- do not create parallel work that violates dependency order just to keep CPUs busy;
- do not merge unverified worker output;
- do not call the project complete until Release Candidate gates in `ROADMAP.md` are satisfied and canonical docs are reconciled.

The continuity mechanism may auto-restart orchestration/worker processes, but it cannot override real external blockers such as host/network/provider outages, missing credentials, exhausted disk, or a required product decision. Such blockers must be recorded explicitly rather than hidden behind a fake `DONE` state.

## 2026-08-31 live allocation checkpoint

At the time this inventory was first committed:

- Primary: independent final review of Task 12 session management;
- Worker 1: independent Task 12 Go/Web verification on a synchronized copy of the Primary working tree;
- Worker 2: isolated P0 `BUG-001` refresh-token reuse-family security work;
- Worker 3: isolated Task 14 concurrent-session-limit work;
- Worker 4: isolated Task 13 real session disconnect work.

This allocation is a timestamped checkpoint, not a permanent task ownership map. `AGENT_TASKS.md` and `WORKLOG.md` remain the task-level truth and must be updated as lanes move.
