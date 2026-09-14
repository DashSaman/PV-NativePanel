# Continue Here — PVNaive

Verified checkpoint: 2026-09-14 07:41 Asia/Tehran

## Current GitHub truth
- Canonical `main` before this documentation refresh: `40024958b05d1025be48dfb95e5c750447457137`; push CI `34801758693` SUCCESS.
- Runtime deploy commit `a4edea62594d5b60a978c39e1f28fad9ac45f6b6`; CI `34795216345` SUCCESS.
- Open PRs: Task13 #108 DRAFT/non-mergeable at stale `f7d8dd5aa8f33b1bc09e3f19bd26ffb219e650d9`; Karing #101 DRAFT/non-mergeable at `216d53670066033403fe95f61b0402bb710186a3`.

## Production truth
- Latest persistent deployment receipt records `pvnaive:repo-live2` image `76c10697a03b`, built from `a4edea6`, healthy after forward-only schema 28→30 migrations (0029/0030).
- Recorded external E2E on `https://45.141.148.59.nip.io` is ALL_GREEN: strict TLS, panel 200, owner login 200, user create 201, subscription delivery headers, UA negotiation, and strict-TLS proxy CONNECT 204. Seven disposable E2E customers were revoked with history preserved.
- Previous `repo-live` image is retained for rollback.
- `namir.softarg.ir` remains deferred until the documented ACME retry window clears on 2026-09-15; do not restart/recreate merely to force issuance.
- Production Primary is currently not connected. This cycle did not independently re-check the live container identity, migration ledger, backup freshness/encryption, disk headroom or rollback snapshot by shell.

## Active lanes
- **Task13 #108**: Worker 3 must reconstruct/reconcile exact session-kill work onto exact latest green main, then rerun CI + Exact Accounting + Pinned Forwardproxy. Worker 2 must then independently execute the real pinned-Caddy HTTP/1.1 + HTTP/2 proof: target-only kill, sibling survival, forged-tuple rejection, repeat-kill idempotency, credential survival, unchanged Caddy lifecycle and exactly-once final accounting. No stale-head merge credit.
- **Karing #101**: Worker 4 must provide a real disposable Karing import → parse → CONNECT → cleanup/revoke receipt, then reconstruct only the validated minimal delta on latest main and rerun exact-head repository gates.
- **R1 / STEER-001 #109**: continue independently. Production schema is now 30, so any new migration must be >0030; never reuse/rewrite 0029/0030. Network telemetry must remain separate from exact byte-accounting/quota truth. Worker 3 = trusted TCP_INFO sampling; Worker 2 = DB ingest/replay/idempotency; Worker 1 = schema/security/accounting review; Worker 4 = E2E once an exact head exists.
- **Production #100**: Primary performs read-only audit first when connected. Before any next runtime deploy: fresh encrypted backup → independent rollback snapshot → exact deploy SHA lock → staged promotion → postflight → retain rollback.

Remote executor inventory at this checkpoint has no online PVNaive worker or Production Primary. Persisted assignments remain authoritative for the next available workers. No additional Production mutation was performed by this coordinator after reconciling the successful schema30 deployment receipt.