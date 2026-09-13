# PVNaive Handoff

Checkpoint: 2026-09-13 23:35 Asia/Tehran

- Current canonical `main`: `61c613777ec2dcf3f1559cbb8b8df80f8e389af0`; push CI `34779414593` SUCCESS.
- Verified runtime-bearing baseline: `f0cdab3eda205eecb3285577ec5df32f45d0ddb7`; schema22 is canonical and merged-main CI is green.
- Karing PR #101 remains OPEN/DRAFT at `216d53670066033403fe95f61b0402bb710186a3`; exact-head repository gates are green but branch drift plus missing disposable real-client import/parse/CONNECT/cleanup evidence block merge.
- Task13 PR #64 remains OPEN/DRAFT at stale `3fc14825e1b164bad558decaef47f56b792e81af`; reconstruct from current main, then independently prove HTTP/1.1 + HTTP/2 target kill, sibling survival, forged-tuple rejection, repeated-kill idempotency, credential survival, unchanged Caddy lifecycle and exactly-once final accounting.
- Remote inventory has no online executable PVNaive worker. Production Primary is absent, so issue #100 has no fresh command-level audit receipt.
- DEPLOY-001 is tracked canonically in #105; #103 is closed duplicate. Recover the exact trusted all-in-one renderer/build context before implementation; do not guess or recreate source from Production behavior.
- LINEAGE-001 remains #104; obtain exact read-only Production migration ledger/checksums and trusted 0022..0027 artifacts before a forward-only compatibility fixture can be accepted.
- Production remained untouched: no backup creation, deploy, migration, restart/reload, DB/credential/Caddy write or rollback-state change.

Execution allocation when hosts reconnect:
- Worker 4: Karing disposable real-client smoke and cleanup/revoke evidence; reconstruct on current main if still stale.
- Worker 3: Task13 current-main reconstruction.
- Worker 2: independent Task13 exact-head race/permission/protocol/accounting verification.
- Worker 1: independent evidence/security/accounting review only where installed tooling is sufficient.
- Primary: read-only Production audit only until all promotion gates are green.
- Coordinator: DEPLOY-001 source provenance and LINEAGE-001 non-destructive reconciliation.

Promotion order: DEPLOY-001 + LINEAGE-001 → Karing real-client proof → Task13 exact-head proof → fresh Production audit → encrypted backup + independent rollback snapshot → exact deploy-SHA lock → staged deploy → postflight with rollback retained.

Never credit stale, mixed-head, assignment-only, static-only, missing-tool, historical or inferred evidence as completion.
