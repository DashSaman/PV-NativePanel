# PVNaive — Canonical Handoff

Last updated: 2026-08-30

This handoff supersedes old S04/S05/S06 and stale PR #16/#29 notes. Resume only from current `main` and live Production truth.

## Repository

- Repo: `DashSaman/PV-NativePanel`
- Product: **PVNaive**
- Default branch: `main`
- Current Production release / main implementation baseline: `e9cce65d3fe8d82100b6bbb7e1231d07dc997edb`
- Schema head: 11 / `0011_customer_product_management`
- Parity reference: `docs/PANEL_PARITY_MASTER_2026-08-30.md`
- Tasks #1-#4: **DONE**
- Next mandated task: **#5 Legacy/adopted accounting baseline truth**

## Production — verified final Task #4 state

Production: `https://namir.softarg.ir/panel/`

Verified on 2026-08-30 without printing secret-bearing values:

- `pvnaive-api.service`, `pvnaive-runtime-agent.service`, `pvnaive-telemetry-agent.service`, `caddy-naive.service`, PostgreSQL: active;
- observed restart counters: 0;
- API loopback readiness healthy;
- Runtime Unix health healthy;
- Telemetry `/run/pvnaive/accounting.sock` health healthy;
- PostgreSQL schema 11;
- six active users;
- six active Naive Runtime credentials;
- six active ServiceTerms;
- six active direct Subscription tokens;
- backup and restore-drill timers active/enabled;
- public panel and public readiness HTTP 200;
- `pvnaive doctor`: 14 PASS / 1 disk WARN / 0 FAIL;
- real automated restore drill: schema/ownership/ACL/signing-key checks PASS;
- bounded loopback HTTP rehearsal: 100/100 success, 0 failures — explicitly not a capacity ceiling;
- release markers point to `e9cce65d3fe8d82100b6bbb7e1231d07dc997edb`;
- Caddy config SHA, PID and restart count remained unchanged across WS4 deploy/hotfix redeploy;
- root filesystem is around 79% used and remains a Doctor warning rather than a failure.

Fresh encrypted config/database backups and complete release rollback snapshots were taken before each Production mutation.

## Task #4 — completed

The useful parts of stale PR #16 were manually reconciled rather than merged wholesale. The final Production result includes:

1. real Linux CPU/RAM/disk/load/uptime/network metrics;
2. server-side network rates from monotonic counter deltas, with no invented first-sample rate;
3. structured redacted logging;
4. request IDs + bounded rate limits + loopback-only forwarded-IP trust;
5. ready-route OpenAPI;
6. real `/api/v1/system/status` for API/DB/Runtime/Telemetry dependencies;
7. `pvnaive doctor`;
8. redacted diagnostic bundle;
9. encrypted scheduled backup;
10. automated isolated restore drill;
11. backup/restore systemd timers;
12. dynamic-schema release builder + checksums/basic SBOM/source provenance;
13. same-schema guarded deploy and telemetry-aware rollback;
14. Production-aware dual web symlinks and DB-script release symlink;
15. deployment marker backup/update/restore;
16. bounded loopback load rehearsal;
17. notification retry/dedupe/redaction foundation;
18. secure Telegram transport foundation;
19. standalone-safe Fleet model/drift/delete-guard foundation;
20. live owner System Dashboard;
21. safe React ErrorBoundary.

Final verification history:

- PR #30 exact head `09b085a877e52fa02c095799359b6b9e89bb3492`: CI #1065, Exact Accounting #181, Pinned Forwardproxy #165 — SUCCESS;
- initial WS4 merge: `c717d162a7e9b2e31fb5822b6b16c27ad048cbbd`;
- Production postflight discovered two operational false negatives: Doctor key-mode expectation and restore validation SIGPIPE/141;
- PR #31 exact head `b740352012fd9646c25d4c70c83f64f2f86ce029`: CI #1070, Exact Accounting #185, Pinned Forwardproxy #169 — SUCCESS;
- final WS4 merge/release: `e9cce65d3fe8d82100b6bbb7e1231d07dc997edb`;
- final Production postflight: PASS.

## Durable integrated core — do not rewrite

### Runtime / accounting

- safe Naive Runtime credential lifecycle;
- narrow privileged Runtime Agent;
- expected-SHA Caddy validate/backup/apply/reload/postflight/rollback;
- exact successful authenticated CONNECT accounting;
- dedicated Telemetry Agent/accounting socket;
- idempotent boot/session/sequence/cumulative event model;
- ServiceTerm-isolated usage;
- trusted first-successful-CONNECT activation producer;
- session/presence projection;
- shared finite-quota reservation/settlement core.

### Customer / delivery

- customer create/adopt/edit/suspend/resume/revoke-safe-delete;
- quota/unlimited/add/set volume;
- creation/first-CONNECT/manual validity, no-expiry, extend-days;
- plans/groups/tags/notes/renewal/new ServiceTerm;
- search/filter/sort/pagination + supported idempotent bulk actions;
- `/sub` machine endpoint, `/s` human page, local QR;
- read-only Subscription view;
- explicit Subscription reissue;
- explicit password rotation.

## Next task — #5 Legacy/adopted accounting baseline truth

Goal: legacy/adopted accounts must never display an invented zero and must never double-count pre-adoption plus direct Naive telemetry.

Required semantics:

- a baseline is numeric only when the pre-adoption usage is provable from an authoritative source;
- otherwise baseline state is explicit `Unknown`/unavailable, not zero;
- direct post-adoption accounting remains exact and restart-safe;
- total usage may combine baseline + direct usage only when the baseline is known and the epochs cannot overlap;
- adoption must record enough provenance/epoch information to prevent double-counting;
- read-only actions, quota/expiry edits, Subscription reissue and QR rendering must not mutate baseline or rotate Runtime/Subscription secrets;
- existing six Production accounts must be classified from real evidence rather than guessed.

Implement with RED→GREEN tests, migration only if necessary, exact-head CI/Exact Accounting/Pinned Forwardproxy, backed-up Production rollout if schema/runtime changes, and live verification.

## Still open after Task #5

- full `/s` accounting/presence projection;
- Manual/Bulk/Periodic reset usage semantics;
- controlled hard-quota and first-CONNECT Production proofs;
- session list/kill/concurrent/unique-IP limits/history;
- HWID and speed-limit PoCs;
- reseller CRUD/wallet/ledger/restrictions;
- customer history/Audit Explorer;
- full notification preferences/history/rule UI and configured Telegram workflow;
- historical metrics/log UI beyond current live monitor;
- three known P0 auth/readiness defects in `KNOWN_ISSUES.md`;
- real fleet controller/multi-node operations;
- clean installer/upgrade/uninstall;
- Karing multi-OS acceptance;
- 50/100/200/400+ capacity campaign;
- final supply-chain/release signing and RC gates.

## Safety invariants

- no force-push/reset of main;
- no secret in Git/chat/CI/evidence;
- no fake usage/online/HWID/speed;
- no read-only QR/Subscription action may rotate secrets;
- Subscription reissue != password rotation;
- quota/expiry edits do not rotate password/token;
- no Runtime mutation without validate → backup → apply → verify → rollback;
- no Production mutation without fresh backup + rollback state;
- do not copy GPL/AGPL competitor code without explicit license review.
