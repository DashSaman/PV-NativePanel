# PVNaive — Canonical Project Status

Last updated: 2026-09-06 23:43 Asia/Tehran

## Verified state
- `main` exact head: `547c1deccd6acde76fd2a1b5babf005a77a5af11` (`docs(handoff): record current GitHub and worker truth`). Combined status is empty and no PR-triggered workflow run is attached to this docs-only head; post-merge CI is not credited.
- PR #64 Task13: OPEN, DRAFT, `mergeable=false`, head `3fc14825e1b164bad558decaef47f56b792e81af`; exact-head CI `33623363327`, Exact Accounting `33623363299`, and Pinned Forwardproxy `33623363389` are SUCCESS. Fresh real HTTP/1.1 + HTTP/2 rehearsal is still missing.
- PR #81 Task16: OPEN, DRAFT, `mergeable=false`, head `3c4310335ab4907d28bac995bba1be3545e14f6e`; Task16 PG18 `33678134359`, Exact Accounting `33678134326`, and Pinned Forwardproxy `33678134350` are SUCCESS, but repository CI `33678134360` is FAILURE in `tests/db/periodic_usage_reset_executor_test.sh` because a generic fixture still expects schema 20 while the branch is schema 21. No exact-head all-four-green receipt exists.
- PR #4 Karing: OPEN, DRAFT, `mergeable=false`, head `2501e39dc39e14063b6a501bc96b77bbfcae7384`; CI `33209239812` is SUCCESS, but reproducible real Karing client smoke is still missing.

## Production truth
Persistent reports provide bounded historical/read-only observations only. No fresh command-level Production audit, encrypted backup preflight, independent rollback snapshot, deploy or postflight was executable in this run. No merge, deploy, migration, restart/reload, DB write, credential mutation, backup mutation or rollback mutation occurred.

## Worker truth
Persistent coordinator/worker reports were rechecked. No fresh completion receipt tied to the current PR heads was found. Historical worker output is not credited without exact GitHub corroboration and a fresh receipt. One-active-host limitations and inactive connected workers remain documented; worker-only output was not integrated.

## This run
- Re-verified `main`, PRs #64/#81/#4, exact-head workflows, combined status, and persistent reports.
- Reconciled current exact workflow outcomes: Task13 fully green on its exact head but protocol rehearsal pending; Task16 dedicated gates green but repository CI still failed on a stale generic schema21 fixture; Karing CI green but real-client smoke pending.
- Refreshed this canonical status file; documentation-only change.

## Next gates
1. Task16: create one clean published head from current `main`, narrow-fix only generic schema21/latest-schema expectations, preserve schema20-specific Task15 fixtures, then run normal CI + Task16 PG18 + Exact Accounting + Pinned Forwardproxy on the same SHA.
2. Task13: obtain fresh real HTTP/1.1 + HTTP/2 rehearsal proving target-only kill, sibling survival, forged-tuple rejection, repeat-kill idempotency, credential survival, no restart/reload and exactly-once final accounting.
3. Karing: obtain reproducible real client import/parse/connect/cleanup evidence.
4. Independent review: inspect RLS, privilege separation, retention/purge safety, accounting/session lineage and secret redaction.
5. Only after all required gates pass: fresh encrypted Production backup, independent rollback state, staged deploy and postflight verification.

Never claim completion from stale reports, older heads or partial evidence.
