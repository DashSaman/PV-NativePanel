# PVNaive — Canonical Handoff

Last updated: 2026-09-07 02:39 Asia/Tehran

## Current truth
- `main` was verified at `87dbb530550772c24f305e5d01be45db17ca0d86` and advanced by docs-only commits in this run to `ad7aaa1ab621c944207531fd65e61b03ae5b8075`. No post-merge CI is credited for these docs-only heads.
- #64 Task13 OPEN/DRAFT, head `3fc14825e1b164bad558decaef47f56b792e81af`; focused/CI evidence is historical/supplemental; fresh real HTTP/1.1 + HTTP/2 rehearsal is still required.
- #81 Task16 OPEN/DRAFT; latest known implementation heads and historical receipts are not a current exact-head four-gate proof. Generic schema21/latest-schema fixture alignment remains required; Task15 schema20-specific fixtures must remain unchanged.
- #4 Karing OPEN/DRAFT, head `2501e39dc39e14063b6a501bc96b77bbfcae7384`; historical CI `33209239812` is SUCCESS, but reproducible real-client smoke is pending.
- Stale documentation PRs #85/#86/#87/#88/#89/#91/#92/#93/#94/#95 were not merged; they target older bases and are not current canonical truth.

## Production
Persistent reports contain bounded historical/read-only observations only. No fresh command-level Production audit, encrypted backup preflight, independent rollback snapshot, deploy or postflight was executable in this run. No merge, deploy, migration, restart/reload, DB write, credential change, backup mutation or rollback mutation occurred.

## Worker/release rules
Worker reports remain historical unless matched to an exact GitHub head and fresh receipt. Dirty/stale worker trees are not completion evidence. Do not integrate worker-only output. Use disposable credentials and isolated canaries. Promotion requires all exact-head gates green, fresh encrypted backup, independent rollback state, provenance and postflight verification.

## This run
- Re-verified repository, current `main`, open PRs, exact-head CI availability, and persistent coordinator/worker reports.
- Rejected stale docs PRs and historical worker claims as promotion evidence.
- Updated canonical status, continuation, and handoff records only; no runtime/schema work integrated and no Production mutation performed.
- Reassigned independent next actions on Task16, Task13, and Karing.

## Next assignments
1. Task16/schema lane: create/refresh a clean branch from latest exact `main`, narrow-fix only generic schema21/latest-schema expectations, preserve schema20-specific Task15 fixtures, align PR metadata, then run normal CI + Task16 PostgreSQL18 + Exact Accounting + Pinned Forwardproxy on one SHA.
2. Task13/protocol lane: clean exact-head checkout, run fresh HTTP/1.1 + HTTP/2 rehearsal outside Production with target-only kill, sibling survival, forged-tuple rejection, repeat-kill idempotency, credential survival, no restart/reload, and exactly-once accounting.
3. Karing compatibility lane: run reproducible real client import/parse/connect/cleanup smoke with disposable credentials and redacted logs.
4. Independent review lane: inspect RLS, privilege separation, retention/purge safety, accounting/session lineage and secret redaction.
5. Production-only lane: read-only health first; backup/rollback then staged promotion only after every release gate is green.

Human action/blocker: provide an executable development worker slot with compatible Go/PostgreSQL18 tooling (or enable multi-host capacity). Until then, no Task13 live rehearsal or Task16 repair can be freshly validated.