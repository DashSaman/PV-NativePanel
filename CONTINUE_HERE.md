# Continue Here — PVNaive

Verified checkpoint: 2026-09-12 18:40 Asia/Tehran

Current `main`: `f446e8517dafa22d2597336167259ea39c7feee0` after documentation-only reconciliation from verified pre-cycle `b6c82f7ecd4e84338c2076c0b1d0ecb0becbaeea`. CI for the pre-cycle SHA was not published in the current lookup; do not claim the new tip green until its own run appears.

Do not merge or deploy yet. Open gates:
- Task13 PR #64 exact head `3fc14825e1b164bad558decaef47f56b792e81af`: rebase/republish from current `main`, fresh real HTTP/1.1 + HTTP/2 rehearsal, target-only kill, sibling survival, forged tuple rejection, idempotency, credential survival, no Caddy restart/reload, exactly-once accounting.
- Karing PR #4 exact head `2501e39dc39e14063b6a501bc96b77bbfcae7384`: rebase/republish if needed, then real import/parse/connect/cleanup smoke with exact profile hash and redacted logs.
- Security/accounting: independent clean exact-head review of Task16/schema21 for RLS, privilege separation, retention/purge, trusted lineage, commit-before-success, and redaction.
- Production: connected audit, fresh encrypted backup, rollback snapshot, staged deployment and postflight are unavailable.

Use exact-head evidence only. Persistent worker reports are historical; TrPaqet is the identified isolated rehearsal lane. Never treat stale, dirty, mixed-head or unpushed evidence as completion.
