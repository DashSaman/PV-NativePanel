# PVNaive — Canonical Project Status

Last updated: 2026-09-06 19:41 Asia/Tehran

## Verified state
- `main` current exact head: `d9595ce4e5f45f0c74147cc73604a78227f633bd`.
- Exact-head workflow lookup for `main` returned no workflow runs; post-merge CI is not credited for this documentation-only head.
- PR #64 Task13: OPEN, DRAFT, `mergeable=false`, head `3fc14825e1b164bad558decaef47f56b792e81af`. Exact-head CI, Exact Accounting and Pinned Forwardproxy are green, but required fresh real HTTP/1.1 + HTTP/2 rehearsal is still missing.
- PR #81 Task16: OPEN, DRAFT, `mergeable=false`, head `3c4310335ab4907d28bac995bba1be3545e14f6e`. Task16 Schema21 TDD, Exact Accounting and Pinned Forwardproxy are green; repository CI run `33678134360` is failed in the database job at the generic periodic-usage schema expectation (`schema version=21, want=20`). No exact-head full-green credit.
- PR #4 Karing: OPEN, DRAFT, `mergeable=true`, head `2501e39dc39e14063b6a501bc96b77bbfcae7384`, base `s04-auth`. Historical CI is green; reproducible real Karing client smoke is still missing.

## Production truth
No fresh command-level Production health audit, backup preflight, rollback snapshot, deploy or postflight was executable in this run. No merge, deploy, migration, restart/reload, DB write, credential mutation, backup mutation or rollback mutation occurred.

## Worker truth
Persistent coordinator/worker reports are historical unless corroborated by exact GitHub state and fresh receipts. No fresh worker completion receipt tied to the current PR heads was found. SentinelX one-active-host limits mean connected workers may be inactive. Worker-only output was not integrated.

## This run
- Verified exact `main` ref, open PRs #64/#81/#4, exact-head workflows and the failed Task16 database job.
- Reconciled persistent reports without crediting stale completion claims.
- Updated canonical status; no runtime/schema work integrated.
- Re-issued bounded, non-Production assignments to the three active roadmap lanes.

## Next gates
1. Task16: fix only the generic latest-schema expectation on a new exact branch head, preserve Task15 schema20 fixtures, then run normal CI + Task16 PG18 + Exact Accounting + Pinned Forwardproxy on one SHA.
2. Task13: obtain fresh real HTTP/1.1 + HTTP/2 rehearsal with target-only kill, sibling survival, forged-tuple rejection, idempotent repeat kill, credential survival, no restart/reload and exactly-once accounting.
3. Karing: obtain reproducible real client import/parse/connect/cleanup evidence.
4. Only after all required gates pass: fresh encrypted Production backup, independent rollback state, staged deploy and postflight verification.

Never claim completion from stale reports, older heads or partial evidence.
