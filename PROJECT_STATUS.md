# PVNaive — Canonical Project Status

Last updated: 2026-09-14 08:42 Asia/Tehran

## Verified GitHub state
- Exact canonical `main` before this documentation refresh: `1db7837357f528fd4a6ce14b69bb91871693eea8`; push CI `34805219116` SUCCESS.
- Runtime deploy commit `a4edea62594d5b60a978c39e1f28fad9ac45f6b6`; CI `34795216345` SUCCESS.
- Task13 PR #108 is OPEN/DRAFT at exact head `f7d8dd5aa8f33b1bc09e3f19bd26ffb219e650d9`; GitHub now reports mergeable=true. Exact-head CI `34789937594`, Exact Accounting `34789937603`, and Pinned Forwardproxy `34789937575` are SUCCESS. Fresh compare from its base to current main shows 24 later commits with no changed-file overlap with the 35 Task13 files; no reconstruction is currently required solely for mergeability.
- Karing PR #101 is OPEN/DRAFT at exact head `216d53670066033403fe95f61b0402bb710186a3`; GitHub now reports mergeable=true. Its four changed RuntimeNaive/runtime files do not overlap the 122 later main commits from its base, so no reconstruction is currently required solely for mergeability.

## Production truth
- Latest persistent deployment receipt records `pvnaive:repo-live2` image `76c10697a03b`, built from runtime `a4edea6` after the migration-guard and `ValidityInput` fixes.
- Production advanced forward-only from schema 28 to schema 30 by applying 0029 and 0030. Readiness was healthy after migration, and the previous `repo-live` image was retained for rollback.
- Recorded external E2E on `https://45.141.148.59.nip.io` is ALL_GREEN: strict TLS, panel 200, owner login 200, user create 201, subscription headers, UA negotiation, and strict-TLS proxy CONNECT 204. Seven disposable E2E users were revoked with history preserved.
- `namir.softarg.ir` remains intentionally deferred until the documented ACME retry window clears on 2026-09-15. Do not restart/recreate merely to chase certificate issuance.
- Production Primary is not currently connected. This coordinator therefore has not independently re-verified live container identity, migration ledger, backup freshness/encryption, disk headroom or rollback snapshot by shell in this cycle.

## Remaining gates
1. **Task13 #108:** branch/repository gates are green and GitHub is presently mergeable. Keep DRAFT until independent real pinned-Caddy HTTP/1.1 + HTTP/2 proof covers target-only kill, sibling survival, forged-tuple rejection, repeat-kill idempotency, account/credential survival, unchanged Caddy lifecycle and exactly-once final accounting. Re-check exact head/merge ref and rerun gates if either moves before merge.
2. **Karing #101:** keep DRAFT until real disposable Karing import → parse → CONNECT → cleanup/revoke evidence exists with client/platform/version, generated-profile SHA-256 and redacted cleanup proof. Re-check exact head/merge ref before merge and rerun repository gates if either moves.
3. **R1 / STEER-001 #109:** continue independently. Schema 30 is deployed, so any new DB migration must be strictly >0030; never reuse or rewrite 0029/0030. Network telemetry must remain physically/logically separate from exact byte-accounting and quota truth.
4. **Production #100:** when Primary reconnects, perform read-only deployed SHA/image, migration ledger/schema, services/listeners/Caddy, backup freshness/encryption, disk and rollback audit. Before any next runtime deploy: fresh encrypted backup → independent rollback snapshot → exact deploy SHA lock → staged promotion → postflight → retain rollback.

## Worker allocation
- Worker 2: Task13 real HTTP1/HTTP2 protocol/accounting acceptance.
- Worker 3: Task13 exact-head/merge-ref defect response only; otherwise R1 trusted TCP_INFO sampling.
- Worker 4: real Karing acceptance; later R1 E2E on an exact implementation head.
- Worker 1: independent diff/schema/security/accounting/CI review and validation of returned receipts.
- Primary: read-only Production audit when connected.
- Coordinator: CI reconciliation, validated integration, documentation and promotion safety.

Current remote inventory has no online PVNaive worker or Production Primary. Persisted GitHub assignments are the execution queue for the next available workers. Never credit stale-head, assignment-only, static-only, inferred, or missing-tool evidence as completion. No Production mutation was performed in this checkpoint.