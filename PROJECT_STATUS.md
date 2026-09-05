# PVNaive — Canonical Project Status

Last updated: 2026-09-05 03:38 Asia/Tehran

This file records verified repository and Production truth. Historical worker/stage notes are evidence only; exact GitHub state, exact-head CI and fresh Production observations override them.

## Product / safety invariants

PVNaive remains standalone-first. Never fabricate usage/online/IP/session history. Never rotate credentials/tokens from read-only flows. Production changes require a fresh encrypted backup + rollback state, intended migrations only, exact artifact provenance and postflight verification.

## Repository truth

- Repository: `DashSaman/PV-NativePanel`.
- Current `main`: `241d67ca880fe633e38409b7ad9ab7e7e4d96ab6`.
- No status rows or workflow runs are visible for the current docs-only `main` head; post-merge CI is not proven.
- Task13: draft PR #64, exact GitHub head `3fc14825e1b164bad558decaef47f56b792e81af`; historical focused gates are green, but fresh real HTTP/1.1 + HTTP/2 rehearsal remains pending. A prior rehearsal attempt was blocked by PostgreSQL 14 / Go 1.18.1 / credential constraints; no PASS was credited.
- Task16: draft PR #81, actual GitHub head `3c4310335ab4907d28bac995bba1be3545e14f6e`; the PR body still claims an older head and is stale. Historical dedicated Task16/Exact Accounting/Pinned Forwardproxy evidence exists, but current generic CI and exact-head evidence are not all green on one published head; no merge authorization exists.

No task becomes DONE from a local candidate, historical report or partial branch alone.

## Production truth

Production remains on Task15/schema20; no Task13 or schema21 code has been deployed.

Fresh read-only observation on `pv-primary` at 2026-09-05 03:38 Asia/Tehran:

- `/api/v1/health/ready` returned `{"db":"ok","ready":true,"schema":"ok","status":"ready"}`;
- `/api/v1/health/live` returned `{"service":"pvnaive-api","status":"ok"}`;
- no deploy, migration, restart, reload, DB write, credential change, backup mutation or rollback mutation was performed.

The active services were not re-credited from systemd because the units are not present in the current SentinelX service allowlist; health claims above are limited to the successful API probes.

## Persistent worker state

- Three SentinelX hosts are connected: `TrPaqet`, `pv-primary`, and `pv-worker-main`.
- The Free-plan one-active-host limit made `pv-primary` the only executable lane in this run; the two development hosts returned `upgrade_required`.
- The persistent checkout at `/opt/openobserve/pvnaive-dev` on `pv-primary` is dirty (23 unstaged changes, 7 untracked files, HEAD `a5638c0900c5`), so it is not a valid integration source.
- No fresh worker completion was creditable.

## Immediate execution order

1. Keep #64 and #81 draft / DO NOT MERGE.
2. Re-run or inspect the latest exact-head CI evidence for #81 after the generic schema21/latest-schema fixture correction; do not reuse stale green receipts.
3. On the first executable development Worker, run the final exact-head Task13 HTTP/1.1 + HTTP/2 rehearsal with PostgreSQL 18 and compatible Go/jq tooling.
4. Reconcile PR #81 to one exact published head and apply only the narrow schema21-aware generic fixture fix; preserve Task15 schema20-specific tests.
5. Only if Task13 live proof and Task16 full gates pass: create a fresh encrypted Production backup + rollback state, then consider promotion.
