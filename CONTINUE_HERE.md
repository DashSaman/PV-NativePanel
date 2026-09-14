# Continue Here — PVNaive

Verified checkpoint: 2026-09-14 09:42 Asia/Tehran

## Current GitHub truth
- Canonical `main` before this documentation refresh: `f1c9092070e3d4bfd3b9c4f62ccf2b9a7c452ffc`; push CI `34809028966` SUCCESS.
- Runtime deploy commit remains `a4edea62594d5b60a978c39e1f28fad9ac45f6b6`; CI `34795216345` SUCCESS.
- Task13 #108 is DRAFT at `f7d8dd5aa8f33b1bc09e3f19bd26ffb219e650d9`; fresh GitHub metadata reports `mergeable=false`. Prior exact-head workflows remain historical-green, but the branch must be refreshed/reconciled and revalidated before merge.
- Karing #101 is DRAFT at `216d53670066033403fe95f61b0402bb710186a3`; fresh GitHub metadata also reports `mergeable=false`. Real-client acceptance is still missing.

## Production truth
- Latest persistent deployment receipt remains `pvnaive:repo-live2` image `76c10697a03b`, built from `a4edea6`, healthy at schema 30 after forward-only 0029/0030.
- Recorded nip.io E2E remains ALL_GREEN; previous `repo-live` remains retained for rollback.
- `namir.softarg.ir` remains deferred until the documented ACME retry window clears on 2026-09-15.
- Production Primary is not connected, so this cycle did not independently re-check live container identity, migration ledger, backup freshness, disk headroom or rollback state by shell.

## Active lanes
- **Task13 #108**: Worker 3 refreshes/reconciles the exact session-kill work on latest green main and reruns CI + Exact Accounting + Pinned Forwardproxy if the head moves. Worker 2 then independently executes real pinned-Caddy HTTP/1.1 + HTTP/2 acceptance: target-only kill, sibling survival, forged-tuple rejection, repeat idempotency, account survival, unchanged Caddy lifecycle and exactly-once final accounting.
- **Karing #101**: Worker 4 performs real disposable Karing import → parse → CONNECT → cleanup/revoke, then reconstructs only the validated minimal delta on latest green main and reruns exact-head repository gates.
- **R1 / STEER-001 #109**: continue independently. New migrations must be >0030; never reuse/rewrite 0029/0030. Network telemetry remains separate from exact byte-accounting/quota truth. Worker 3 = trusted TCP_INFO sampling; Worker 2 = DB ingest/replay/idempotency; Worker 1 = schema/security/accounting review; Worker 4 = E2E after an exact implementation head exists.
- **Production #100**: Primary performs a read-only audit first when connected. Before any future runtime deploy: fresh encrypted backup → independent rollback snapshot → exact deploy SHA lock → staged promotion → postflight → retain rollback.

Remote executor inventory at this checkpoint has no online PVNaive worker or Production Primary; both known worker registrations are offline. Persisted assignments remain authoritative for the next available workers. No additional Production mutation was performed.