# PVNaive Handoff

Checkpoint: 2026-09-15 Asia/Tehran

## Verified baseline

- Current code main before this documentation checkpoint: `9187266c5f4b63df7849553d450975a982f6b816` (R10 Gold/dual-QR/subscription-guide). CI run `34907620721` has Go, Web, database/migration/backup-restore, pinned forwardproxy boundary, S04 auth rehearsal and full S04R/Task13 rehearsal steps PASS; GitHub has not yet recorded the entire workflow as terminal SUCCESS, so do not overstate it.
- R10 client-specific support claims are not yet acceptance-proven. Issue #120 is the new promotion gate; #101 Karing remains DRAFT/non-mergeable on stale head and still requires real disposable import → parse → CONNECT → cleanup/revoke after reconciliation to latest verified main.
- Task13 #108 is merged after exact-head CI/accounting/forwardproxy plus real HTTP/1.1+HTTP/2 target-only kill proof; exact accounting/session/credential invariants remain locked.
- R8 live monitoring is integrated and its SSE race regression is fixed/closed.
- R5 registry and basic real two-node mTLS pull/heartbeat/drift E2E are green. #114 stays OPEN for certificate overlap/rotation + explicit revocation/replay/fail-closed proof.

## Production blocker

- Fresh remote inventory shows both known Remote Desktop registrations offline; no trusted `PVNaive-Production-Primary` is connected.
- Production truth ceiling remains the last persistent schema-33 / repo-fin2 checkpoint. Do not claim fresh image/schema/backup/disk/Caddy/rollback state.
- On trusted Primary reconnect: read-only host identity → deployed SHA/image → schema ledger → services/listeners/Caddy → encrypted-backup freshness → disk → rollback snapshot. Only then may fresh backup/snapshot and staged promotion be considered.

## Remaining acceptance blockers

- #120/#101: verify client compatibility truth; real Karing exact-main acceptance is non-substitutable.
- #114: cert lifecycle rotation/overlap/revocation/replay/fail-closed proof.
- #115: disposable R6 cover/persona/probe-sweep/failure rehearsal; Production remains default-OFF/gated.
- Production: trusted Primary reconnect plus audit/backup/rollback gates.

## Worker queue

- Worker 1 — R10 client-claim/security truth review; then R5 PKI/revocation and R6/R8 RBAC/accessibility.
- Worker 2 — RED-first R10 UI/content capability tests preserving dual-QR; then R8 ledger reconciliation/projections.
- Worker 3 — verify R10 subscription/direct format semantics; then R5 cert lifecycle + STEER-006.
- Worker 4 — real disposable Karing acceptance; then R5/R6 browser/multi-node E2E.
- Coordinator — integrate exact-head validated work only. All remote registrations are currently offline, so assignments are queued persistently in GitHub rather than fabricated as running.

## Invariants

- Preserve exact accounting, session, quota and credential semantics.
- Task13 session control is target-only and cannot become credential revocation.
- Missing telemetry is Unknown, never zero-by-assumption.
- TLS client certificate is authoritative for fleet identity.
- Forward-only immutable migration ledger.
- User-facing compatibility claims require real evidence, not just build/unit success.
