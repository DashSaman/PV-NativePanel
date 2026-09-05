# CONTINUE HERE — PVNaive

Last updated: 2026-09-05 03:38 Asia/Tehran

Start here after interruption. Re-read exact GitHub `main`, open PRs, exact-head CI and fresh Production health before any mutation.

## Current verified state

- `main`: `ad6f3c3e18b6b05474701993c3cd760d27338045` after the canonical status refresh in this run.
- No status rows or workflow runs are visible for the prior docs-only main head; post-merge CI is not proven for the new docs commit until checked again.
- Draft Task13 PR #64 exact head: `3fc14825e1b164bad558decaef47f56b792e81af`; historical gates are supplemental only; fresh real HTTP/1.1 + HTTP/2 rehearsal remains pending.
- Draft Task16 PR #81 actual GitHub head: `3c4310335ab4907d28bac995bba1be3545e14f6e`; PR body is stale. Current evidence is not a complete green set on one exact published head; generic latest-schema/RLS and pinned-build evidence must be re-established.
- Production remains on Task15/schema20; no Task13/schema21 code deployed.
- Fresh `pv-primary` API probes returned readiness `db=ok/schema=ok/ready=true` and liveness `status=ok`. Systemd service state was not re-credited because PVNaive units are not in the current SentinelX service allowlist.

## Task accounting

- Task12: DONE / Production, schema17.
- Task14: DONE / Production, schema19.
- Task15: DONE / Production, schema20.
- Task35 security P0: DONE in main.
- Task13: IN PROGRESS / draft #64 / live proof pending.
- Task16: IN PROGRESS / draft #81 / exact-head generic CI reconciliation pending.

## Next execution

- Development lane: final exact-head Task13 live HTTP1/HTTP2 rehearsal when a development host is executable, using PostgreSQL 18 and compatible Go/jq tooling.
- Development lane: narrow schema21-aware generic RLS/latest-schema fixture fix and full gate rerun on one exact head.
- `pv-primary`: Production-only audit, encrypted backup, rollback and deploy lane; do not use it as a development database/test lane.

Never claim completion from stale worker reports or partial evidence. No merge/deploy until all exact-head gates and safety prerequisites are green.
