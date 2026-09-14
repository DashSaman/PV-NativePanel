# PVNaive — Canonical Project Status

Last updated: 2026-09-14 06:37 Asia/Tehran

## Verified GitHub state
- Canonical `main` before this documentation refresh: `fabd62f4fd71c23c67c3901286153d34d3c30842`; push CI `34798441279` SUCCESS.
- Runtime deploy commit `a4edea62594d5b60a978c39e1f28fad9ac45f6b6`; CI `34795216345` SUCCESS.
- Task13 PR #108 remains OPEN/DRAFT/non-mergeable at stale `f7d8dd5aa8f33b1bc09e3f19bd26ffb219e650d9`.
- Karing PR #101 remains OPEN/DRAFT/non-mergeable at `216d53670066033403fe95f61b0402bb710186a3`.

## Production truth
- Latest persistent deployment receipt records `pvnaive:repo-live2` image `76c10697a03b`, built from runtime `a4edea6` after the migration-guard and `ValidityInput` fixes.
- Production advanced forward-only from schema 28 to schema 30 by applying 0029 and 0030. Readiness was healthy after migration, and the previous `repo-live` image was retained for rollback.
- Recorded external E2E on `https://45.141.148.59.nip.io` is ALL_GREEN: strict TLS, panel 200, owner login 200, user create 201, subscription headers, UA negotiation, and strict-TLS proxy CONNECT 204. Seven disposable E2E users were revoked with history preserved.
- `namir.softarg.ir` remains intentionally deferred until the documented ACME retry window clears on 2026-09-15. Do not restart/recreate merely to chase certificate issuance.
- Production Primary is not currently connected. This coordinator therefore has not independently re-verified live container identity, migration ledger, backup freshness/encryption, disk headroom or rollback snapshot by shell in this cycle.

## Remaining gates
1. **Task13 #108:** reconstruct/reconcile onto exact latest green main; rerun exact-head CI + Exact Accounting + Pinned Forwardproxy; then obtain independent real pinned-Caddy HTTP/1.1 + HTTP/2 proof of target-only kill, sibling survival, forged-tuple rejection, repeat-kill idempotency, account survival, unchanged Caddy lifecycle and exactly-once final accounting.
2. **Karing #101:** obtain real disposable Karing import → parse → CONNECT → cleanup/revoke evidence; then reconstruct only the validated minimal delta on latest main and rerun exact-head repository gates.
3. **R1 / STEER-001 #109:** continue independently. Schema 30 is deployed, so any new DB migration must be >0030; never reuse or rewrite 0029/0030. Network telemetry must remain physically/logically separate from exact byte-accounting and quota truth.
4. **Production #100:** when Primary reconnects, perform read-only deployed SHA/image, migration ledger/schema, services/listeners/Caddy, backup freshness/encryption, disk and rollback audit. Before any next runtime deploy: fresh encrypted backup → independent rollback snapshot → exact deploy SHA lock → staged promotion → postflight → retain rollback.

## Worker allocation
- Worker 3: Task13 latest-main reconstruction; R1 trusted TCP_INFO sampling once the Task13 branch is stable.
- Worker 2: Task13 real HTTP1/HTTP2 protocol/accounting acceptance after the refreshed exact head is green; R1 DB ingest/replay/idempotency independently.
- Worker 4: real Karing acceptance; later R1 E2E on an exact implementation head.
- Worker 1: independent diff/schema/security/accounting/CI review and safe branch preparation when online.
- Primary: read-only Production audit when connected.
- Coordinator: CI reconciliation, validated integration, documentation and promotion safety.

Current remote inventory at this checkpoint has no online PVNaive worker or Production Primary. Persisted GitHub assignments are the execution queue for the next available workers. Never credit stale-head, assignment-only, static-only, inferred, or missing-tool evidence as completion.