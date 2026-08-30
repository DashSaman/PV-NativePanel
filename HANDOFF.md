# PVNaive — Canonical Handoff

Last updated: 2026-08-30

This handoff supersedes old S04/S05/S06 branch notes. Do not resume from stale PR branches wholesale.

## Repository

- Repo: `DashSaman/PV-NativePanel`
- Product: **PVNaive**
- Default branch: `main`
- Current main before Task #4 merge: `1ced722f9ee46fc4fb1a005ae8653f52339786b0`
- Active branch: `lead/ws4-safe-extract-2026-08-30`
- Active PR: `#29`
- Schema head: 11 / `0011_customer_product_management`
- Parity reference: `docs/PANEL_PARITY_MASTER_2026-08-30.md`

Tasks #1-#3 are merged. Task #4 is at final-documentation/exact-head verification and must not be called Production-complete until the merged commit is deployed and smoke-tested.

## Production — latest pre-deploy audit

Production: `https://namir.softarg.ir/panel/`

Observed without printing secrets:

- `pvnaive-api.service`, `pvnaive-runtime-agent.service`, `pvnaive-telemetry-agent.service`, `caddy-naive.service`: active;
- restart counters observed at 0;
- API loopback `127.0.0.1:8080` healthy;
- Runtime Unix health healthy;
- Telemetry `/run/pvnaive/accounting.sock` health healthy;
- PostgreSQL schema 11;
- six active users, six active Naive Runtime credentials, six ServiceTerms, six active direct Subscription tokens;
- direct accounting events/sessions are live;
- root filesystem about 79% used;
- Caddy panel docroot is `/var/www/pvnaive-preview/current`;
- `/opt/pvnaive/web/current` and `/opt/pvnaive/db/current` are also release symlinks;
- backup age recipient/key exist with restricted permissions;
- pre-WS4 scheduled PVNaive backup/restore timers were not active;
- legacy deploy markers lagged running artifacts.

Before mutation, re-check live state and take fresh DB/config/Caddy/web/binary/unit/tmpfiles/marker rollback state.

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

## Task #4 — manually extracted from stale PR #16

Never merge PR #16 wholesale. PR #29 manually reconciles the still-useful ideas against current main.

Completed on the Task #4 branch:

1. real Linux CPU/RAM/disk/load/uptime/network metrics;
2. server-side network rates from monotonic counter deltas, with no invented first-sample rate;
3. structured redacted logging;
4. request ID + bounded rate-limit middleware + loopback-only forwarded-IP trust;
5. ready-route OpenAPI;
6. real `/api/v1/system/status` with API/DB/Runtime/Telemetry dependency state;
7. `pvnaive doctor`;
8. redacted diagnostic bundle;
9. encrypted scheduled backup;
10. isolated restore drill;
11. backup/restore systemd timers;
12. dynamic-schema release builder + checksums/SBOM/source provenance;
13. same-schema guarded deploy and telemetry-aware rollback;
14. Production-aware dual web release symlinks (`/opt/pvnaive/web/current` + `/var/www/pvnaive-preview/current`);
15. deploy-marker backup/update/restore;
16. bounded loopback load rehearsal, explicitly not capacity proof;
17. notification retry/dedupe/redaction foundation;
18. Telegram transport foundation with token-leak prevention;
19. fleet model/drift/delete-guard foundation only;
20. live owner System Dashboard with API/DB/Runtime/Telemetry badges and real metrics;
21. safe React ErrorBoundary without raw exception rendering.

Notification/fleet foundations are not equivalent to a finished notification product or multi-node controller.

## Verification

Unit #7 implementation head `6a61512449884432477b1c107de6ac80e9e0d69c` passed:

- CI #1060, including Web tests/build, Go vet/tests, PostgreSQL checks, full S04R rehearsal and production bundle;
- Exact Accounting #176;
- Pinned Forwardproxy #160.

Documentation commits follow that head. **Require all three workflows again on the final exact PR head before merge.**

## Still open

- legacy/adopted accounting baseline truth;
- full `/s` accounting/presence;
- Manual/Bulk/Periodic reset usage semantics;
- controlled hard-quota and first-CONNECT Production proofs;
- session list/kill/concurrent/unique-IP limits/history;
- HWID and speed-limit PoCs;
- reseller wallet/ledger/restrictions;
- customer history/Audit Explorer;
- full notification preferences/history/rule UI and configured Telegram workflow;
- historical metrics/log UI beyond current live monitor;
- three known P0 auth/readiness defects in `KNOWN_ISSUES.md`;
- real fleet controller/multi-node operations;
- clean installer/upgrade/uninstall;
- Karing multi-OS acceptance;
- 50/100/200/400+ capacity campaign;
- final supply-chain/release signing and RC gates.

## Exact continuation

1. finish Task #4 canonical docs;
2. verify CI + Exact Accounting + Pinned Forwardproxy on the final exact PR #29 head;
3. final diff/secret/rollback review;
4. merge PR #29 only with expected head SHA;
5. on Production: re-audit → fresh backups → build artifact from exact merged commit → guarded same-schema deploy → verify API/Runtime/Telemetry/panel/timers/markers/Caddy unchanged → run Doctor and bounded smoke;
6. if any postflight fails, use the generated release backup with `rollback-r1.sh` and re-verify;
7. only after Production success start Task #5: legacy/adopted accounting baseline truth.

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
