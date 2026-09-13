# Continue Here — PVNaive

Verified checkpoint: 2026-09-13 03:41 Asia/Tehran

Pre-docs `main`: `7ab4d8a8bbb5af7ecef1135743bda40ef7bfa472`. Exact combined status was pending with zero statuses and no commit-specific workflow runs.

Do not deploy yet. Current executable lanes:
- **Karing PR #101**: clean current-main reconstruction. RED commit `95768324b5f5dd43ddaf153c0b115678fa24a49e` added only the regression test and correctly failed the web `npm test` gate. GREEN commit `6168c8445ce5b9358c9cbae12be98951d3153845` adds the minimal profile builder; web tests and build pass. Next: preserve current-main `RuntimeNaive.tsx` behavior while wiring the Karing copy action, then require all exact-head workflows green and run an independent real Karing import/parse/connect/cleanup smoke with disposable credentials, exact profile hash, client/platform/version and redacted logs. Keep DRAFT. Legacy PR #4 stays open until #101 fully supersedes it.
- **Task13 PR #64**: OPEN/DRAFT/non-mergeable at `3fc14825e1b164bad558decaef47f56b792e81af`. Reconstruct validated delta from current `main`, rerun exact-head CI/accounting/pinned-forwardproxy/focused gates, then isolated real HTTP/1.1 + HTTP/2 rehearsal proving target-only kill, sibling survival, forged-tuple rejection, repeat-kill idempotency, credential survival, no Caddy restart/reload and exactly-once accounting.
- **Security/accounting issue #99**: independent clean current-main review of merged schema21 for RLS fail-closed behavior, privilege separation, bounded retention/purge safety, trusted lineage, commit-before-HTTP-success and redaction. Any fix goes to a separate PR.
- **Production issue #100**: read-only audit only. Record exact host/time, deployed SHA/schema, service/readiness/listeners, Caddy binary/pinned SHA/MainPID/NRestarts, session-control socket mode/ownership, disk/capacity, backup inventory/freshness/encryption, independent rollback snapshot availability and postflight prerequisites.

Worker truth: no new completion receipt arrived after the prior dispatches. TrPaqet remains the persisted isolated Task13 rehearsal lane. Historical `AGENT_HANDOFF.md` and `ops/DEPLOYMENT_PROGRESS.md` are S04-era records from 2026-08-27 and must not override fresh exact-SHA evidence.

Use exact-head evidence only. Never credit stale, dirty, mixed-head, unpushed, historical or absent-CI evidence. Production promotion remains read-only audit → fresh encrypted backup → independent rollback snapshot → exact SHA lock → staged promotion → health/postflight → retained rollback.
