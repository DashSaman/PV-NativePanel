# PVNaive — Canonical Project Status

Last updated: 2026-09-05 02:38 Asia/Tehran

This file records verified repository and Production truth. Historical worker/stage notes are evidence only; exact GitHub state, exact-head CI and fresh Production observations override them.

## Product / safety invariants

PVNaive remains standalone-first. Never fabricate usage/online/IP/session history. Never rotate credentials/tokens from read-only flows. Production changes require a fresh encrypted backup + rollback state, intended migrations only, exact artifact provenance and postflight verification.

## Repository truth

- Repository: `DashSaman/PV-NativePanel`.
- Current `main`: `eb212dfb40836a6bb91546e2bcc9f0cfc1afde7b` (docs-only checkpoint).
- No status rows or workflow runs are visible for this exact `main` head; post-merge CI is not proven.
- Task13: draft PR #64, exact head `3fc14825e1b164bad558decaef47f56b792e81af`; focused gates are green, but fresh real HTTP/1.1 + HTTP/2 rehearsal remains pending.
- Task16: draft PR #81, actual GitHub head `3c4310335ab4907d28bac995bba1be3545e14f6e`; the PR body claims an older head and is stale. Dedicated Task16/Exact Accounting/Pinned Forwardproxy gates are green; generic CI run `33678134360` failed in `database`. Failed jobs were re-run in this run; result is pending.

No task becomes DONE from a local candidate, historical report or partial branch alone.

## Production truth

Production remains on Task15/schema20; no Task13 or schema21 code has been deployed.

Fresh read-only observation on `pv-primary` at 2026-09-05 02:38 Asia/Tehran:

- `pvnaive-api.service`: active;
- `caddy-naive.service`: active;
- `pvnaive-runtime-agent.service`: active;
- `pvnaive-telemetry-agent.service`: active;
- readiness body: `db=ok`, `schema=ok`, `ready=true`, but HTTPS probe returned curl HTTP 000 due to TLS alert `internal error`;
- liveness body: `status=ok`, but HTTPS probe returned curl HTTP 000 due to the same TLS alert.

The service processes are active and API bodies report ready/ok, but end-to-end local HTTPS health is NOT currently verified. No Production mutation, deploy, restart, reload, migration, DB write or credential change was performed in this run.

## Persistent worker state

- `pv-primary` is the only executable SentinelX lane under the current one-active-host limit.
- `TrPaqet` and `pv-worker-main` are connected but inactive; no fresh development-worker completion was creditable.
- The persistent checkout at `/opt/openobserve/pvnaive-dev` is dirty (23 unstaged changes, 7 untracked files, HEAD `a5638c0900c5`), so it is not a valid integration source.
- `AGENT_HANDOFF.md` and `ops/DEPLOYMENT_PROGRESS.md` remain historical S04-era ledgers and were not treated as current truth.

## Immediate execution order

1. Keep #64 and #81 draft / DO NOT MERGE.
2. Collect the result of the PR #81 failed-job rerun.
3. Investigate and resolve the Production TLS alert independently, without touching application/database state.
4. On the first executable development Worker, run the final exact-head Task13 HTTP/1.1 + HTTP/2 rehearsal.
5. Reconcile PR #81 to one exact published head and apply only the narrow schema21-aware generic RLS/latest-schema fixture fix.
6. Only if Task13 live proof, Task16 full gates, and Production HTTPS/backup/rollback checks pass: create a fresh encrypted Production backup + rollback state, then consider promotion.
