# Continue Here — PVNaive

Verified checkpoint: 2026-09-13 01:41 Asia/Tehran

Pre-cycle `main`: `655526a5a572f8d62bc33319d7f1e2c9705bd57b`. Its combined status was pending with zero published statuses and commit-specific workflow lookup returned no runs. This cycle refreshed canonical docs, closed completed Task16 issue #79, and removed duplicate follow-up issues #97/#98. Re-read `main` before binding any runtime work.

Do not merge or deploy yet. Open gates:
- Task13 PR #64 exact head `3fc14825e1b164bad558decaef47f56b792e81af`: reconstruct/rebase from current `main` without force-rewriting validated history, rerun exact-head CI/accounting/pinned-forwardproxy/focused gates, then run the real isolated pinned-Caddy HTTP/1.1 + HTTP/2 rehearsal proving target-only kill, sibling survival, forged-tuple rejection, repeat-kill idempotency, credential survival, no Caddy restart/reload, and exactly-once accounting.
- Karing PR #4 exact head `2501e39dc39e14063b6a501bc96b77bbfcae7384`: reconcile the export change from current `main`, rerun exact-head CI, then run real import/parse/connect/cleanup smoke using disposable credentials and a non-Production target; record exact profile hash, client/platform/version and redacted logs.
- Security/accounting issue #99: independent clean current-main review of merged Task16/schema21 for RLS fail-closed behavior, app/maintenance privilege separation, bounded retention/purge safety, trusted lineage, commit-before-HTTP-success semantics and redaction.
- Production issue #100: connected read-only audit must precede any backup/deploy activity. Record exact host/time, deployed SHA/schema, services/readiness/listeners, Caddy binary/pinned SHA/MainPID/NRestarts, session-control socket permissions, backup inventory/freshness/encryption, rollback snapshot availability, disk/capacity and postflight prerequisites.

Current worker lanes:
- TrPaqet: Task13 current-main reconstruction plus isolated rehearsal only; preserve co-hosted Paqet/Xray/OpenVPN/WaterWall/OV services.
- Security review lane: issue #99 exact-main clean-worktree review; any fix goes to a separate PR.
- Compatibility lane: PR #4 current-main-derived Karing export plus independent real-client smoke.
- Production lane: issue #100 read-only audit/rollback-readiness inventory only; no mutation until all runtime gates are green.

Repository truth:
- Task16/schema21 PR #81 is merged and issue #79 is closed completed.
- Duplicate worker issues #97/#98 are closed as superseded by #99/#100.
- Historical `AGENT_HANDOFF.md` and `ops/DEPLOYMENT_PROGRESS.md` are S04-era records from 2026-08-27 and must not override fresh GitHub state or these canonical files.

Use exact-head evidence only. Never credit stale, dirty, mixed-head, unpushed, historical, or absent-CI evidence as completion. Production promotion remains read-only audit → fresh encrypted backup → independent rollback snapshot → exact SHA lock → staged promotion → health/postflight → retained rollback.
