# Continue Here — PVNaive

Verified checkpoint: 2026-09-12 23:37 Asia/Tehran

Latest verified `main` before this documentation cycle: `f4f32154fb4081cf5e06efb93c7d4c6706fab8c7`. Exact-main status/workflow lookup returned no CI evidence; do not claim the current tip green until a published run is observed.

Do not merge or deploy yet. Open gates:
- Task13 PR #64 exact head `3fc14825e1b164bad558decaef47f56b792e81af`: rebase/republish from current `main`, rerun exact-head CI/accounting/forwardproxy gates, then run the real HTTP/1.1 + HTTP/2 rehearsal proving target-only kill, sibling survival, forged tuple rejection, repeat-kill idempotency, credential survival, no Caddy restart/reload, and exactly-once accounting.
- Karing PR #4 exact head `2501e39dc39e14063b6a501bc96b77bbfcae7384`: rebase/republish if needed, then real import/parse/connect/cleanup smoke with exact profile hash, client/platform/version, and redacted logs.
- Security/accounting issue #99: independent clean exact-head review of Task16/schema21 for RLS, privilege separation, retention/purge, trusted lineage, commit-before-success, and redaction.
- Production issue #100: connected read-only audit, fresh encrypted backup, rollback snapshot, staged deployment, health verification, and postflight are unavailable.

Worker dispatch for the next cycle:
- TrPaqet: isolated Task13 rehearsal lane only; do not modify Paqet/Xray/OpenVPN/WaterWall/OV services.
- Worker review lane: exact-head Task16/schema21 security/accounting review on a clean worktree.
- Worker compatibility lane: Karing real-client smoke evidence with exact profile hash and redacted logs.
- Production lane: read-only audit and rollback-readiness inventory only; no mutation without all gates.

Use exact-head evidence only. Persistent worker reports are historical; never treat stale, dirty, mixed-head, or unpushed evidence as completion.
