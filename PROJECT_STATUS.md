# PVNaive — Canonical Project Status

Last updated: 2026-09-06 22:38 Asia/Tehran

## Verified state
- `main` current exact head: `70a42c7de2913face8a6d07e8c96a37dbc75e57b`.
- Exact-head workflow lookup for this documentation head returned no runs/status rows; this follow-up is documentation-only, so post-merge CI is not credited.
- PR #64 Task13: OPEN, DRAFT, `mergeable=false`, head `3fc14825e1b164bad558decaef47f56b792e81af`. Recorded CI/Exact Accounting/Pinned Forwardproxy evidence is supplemental; fresh real HTTP/1.1 + HTTP/2 rehearsal is still missing.
- PR #81 Task16: OPEN, DRAFT, currently reported `mergeable=true`, head `3c4310335ab4907d28bac995bba1be3545e14f6e`, base `main` but stale base/body evidence. Recorded dedicated gates are green, while repository-wide CI evidence remains blocked by the generic schema21/latest-schema fixture path (`periodic_usage_reset_executor_test.sh`, schema 21 vs expected 20 / RLS expectation drift). No current exact-head all-four-green receipt was observed.
- PR #4 Karing: OPEN, DRAFT, `mergeable=true`, head `2501e39dc39e14063b6a501bc96b77bbfcae7384`, base `s04-auth`. Historical CI is green; reproducible real Karing client smoke is still missing.

## Production truth
Persistent reports contain bounded historical/read-only observations that pv-primary services and health endpoints were previously healthy, but no fresh command-level Production audit, encrypted backup preflight, independent rollback snapshot, deploy or postflight was executable in this run. No merge, deploy, migration, restart/reload, DB write, credential mutation, backup mutation or rollback mutation occurred.

## Worker truth
Persistent coordinator/worker reports were rechecked. Historical worker output is not credited without exact GitHub corroboration and a fresh receipt. No fresh worker completion receipt tied to the current PR heads was found. The reports also document one-active-host limits and dirty/stale worker checkouts; worker-only output was not integrated.

## This run
- Re-verified exact `main` ref `70a42c7de2913face8a6d07e8c96a37dbc75e57b`, open PRs #64/#81/#4, exact-head status/workflow presence, and persistent reports.
- Confirmed that current `main` has no published post-merge CI evidence.
- Confirmed Task16 is not promotable: dedicated historical gates are insufficient while generic CI/latest-schema evidence is stale or failed and PR metadata/body references older heads.
- Confirmed Task13 still lacks the required fresh HTTP/1.1 + HTTP/2 protocol/accounting rehearsal.
- Confirmed Karing still lacks a reproducible real-client smoke.
- Refreshed canonical documentation; no runtime/schema work was integrated.

## Next gates
1. Task16: create one clean exact published head from current `main`, fix only generic schema21/latest-schema expectations, preserve schema20-specific Task15 fixtures, align PR body/base metadata, then run normal CI + Task16 PG18 + Exact Accounting + Pinned Forwardproxy on that same SHA.
2. Task13: obtain fresh real HTTP/1.1 + HTTP/2 rehearsal with target-only kill, sibling survival, forged-tuple rejection, repeat-kill idempotency, credential survival, no restart/reload and exactly-once final accounting.
3. Karing: obtain reproducible real client import/parse/connect/cleanup evidence.
4. Independent review: inspect RLS, privilege separation, retention/purge safety, accounting/session lineage and secret redaction.
5. Only after all required gates pass: fresh encrypted Production backup, independent rollback state, staged deploy and postflight verification.

Never claim completion from stale reports, older heads or partial evidence.
