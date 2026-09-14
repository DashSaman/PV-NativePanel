# PVNaive — Worklog

Last updated: 2026-08-30

This log records significant verified transitions so future agents do not repeat completed work. Fine-grained RED/GREEN evidence remains in git history and GitHub Actions.

## Durable historical milestones

| Date | Work | Result |
|---|---|---|
| 2026-08-27 | naming/filesystem/PostgreSQL 18/RLS/encrypted DB backup-restore/S03 foundation | DONE |
| 2026-08-28 | S04 auth + protected panel/API | DONE |
| 2026-08-28 | S04R Runtime credential lifecycle + safe Runtime Agent/Caddy mutation | DONE |
| 2026-08-29 | customer lifecycle/quota/validity/subscription/QR baseline | DONE baseline |
| 2026-08-29 | WS1 exact direct-Naive accounting/first-CONNECT/session/quota core | MERGED |
| 2026-08-29 | WS2 schema 11 customer product management | MERGED |
| 2026-08-29 | WS3 `/sub` machine + `/s` human delivery separation | MERGED |
| 2026-08-30 | Tasks #1-#3 current truth/parity/docs reconciliation | MERGED via replacement PR #28, main `1ced722f9ee46fc4fb1a005ae8653f52339786b0` |

## Task #4 — stale PR #16 safe extraction

Branch: `lead/ws4-safe-extract-2026-08-30`

PR: #29

Base main: `1ced722f9ee46fc4fb1a005ae8653f52339786b0`

Rule: no blind merge/cherry-pick from stale PR #16. Newer WS1 accounting, WS2 product and WS3 delivery behavior must win.

### Unit 1 — observability

Implemented:

- Linux CPU/RAM/disk/load/uptime/network metrics;
- server-side RX/TX rate only from monotonic counter + time deltas;
- no fabricated first sample and no negative rate after counter rollback;
- structured event logging and sensitive key/text redaction.

TDD: initial undefined-symbol RED followed by GREEN.

### Unit 2 — HTTP ops/OpenAPI/system status

Implemented:

- request IDs;
- redacted structured request completion logs;
- bounded per-IP rate limiting;
- forwarded IP trusted only from loopback reverse proxy;
- OpenAPI generated from current ready routes;
- `/api/v1/system/status` provider with real API/DB/Runtime/Telemetry dependency state.

Preserved current customer routes instead of stale PR #16 assumptions.

### Unit 3 — Doctor/diagnostics

Implemented:

- `pvnaive doctor` text/JSON checks;
- current API/Runtime/Telemetry/Caddy/PostgreSQL services/sockets;
- disk and backup freshness checks;
- strict redaction;
- root diagnostic bundle with bounded journals, listeners/disk/memory and environment variable names only.

### Unit 4 — scheduled backup/restore drill

Implemented:

- encrypted config snapshot;
- PostgreSQL backup via existing DB scripts;
- no DB password copied into backup tree;
- daily backup systemd timer;
- isolated weekly restore drill into strictly generated disposable DB;
- no automatic root-level retention/delete logic added during this extraction.

RED contract was observed before scripts/units existed; implementation then passed CI.

### Unit 5 — release/deploy/rollback/load tooling

Implemented:

- dynamic latest schema from migration filenames; obsolete schema 8 hard-code removed;
- API/password/runtime/telemetry static binaries;
- checksums, basic SBOM and source commit provenance;
- same-schema guarded deployment only — no hidden auto-migration;
- mandatory encrypted pre-deploy backup;
- telemetry-aware rollback;
- no Caddy restart/change: hash/PID/NRestarts checked before/after;
- bounded loopback control-plane rehearsal (`5000` requests / `100` concurrency hard max), explicitly not a capacity ceiling.

Exact head `e7eca858…` passed CI/Exact Accounting/Pinned Forwardproxy before later hardening.

### Unit 5.1 — real Production web layout hardening

Read-only Production audit found Caddy's live panel root is `/var/www/pvnaive-preview/current`, while `/opt/pvnaive/web/current` also exists. A test was added and failed exactly because deploy did not know the preview path.

Implementation now:

- backs up both web symlink targets;
- creates root:caddy preview release with restricted modes;
- switches both web symlinks;
- rollback restores both;
- still does not restart Caddy.

Exact head `ec6bedf…`: CI #1048 + Exact Accounting #164 + Pinned Forwardproxy #148 SUCCESS.

### Unit 5.2 — deployment provenance markers

Read-only Production audit showed stale:

- `/opt/pvnaive/DEPLOYED_COMMIT`
- `/opt/pvnaive/DEPLOYED_WEB_RELEASE`

A RED contract proved deploy lacked them. Implementation now backs up, updates and rollback-restores both markers alongside `/opt/pvnaive/release/CURRENT`/`RELEASE.json`.

Exact head `6f0ec5ec…`: CI #1051 + Exact Accounting #167 + Pinned Forwardproxy #151 SUCCESS.

### Unit 6 — Notification/Fleet foundations

Implemented additively:

- notification message model;
- retry/dedupe/redaction engine foundation;
- Telegram HTTP transport with non-2xx failure and token-leak prevention tests;
- in-app channel interface;
- standalone-safe Fleet model, drift classification and delete guard.

No real Telegram configuration/controller/fake multi-node health was wired into Production.

Exact head `d436185…`: CI #1045 + Exact Accounting #161 + Pinned Forwardproxy #145 SUCCESS.

### Unit 7 — System Dashboard/ErrorBoundary

RED #1: CI #1052 failed only because `./systemStatus` did not exist; Go/DB and both WS1 gates stayed healthy.

Implementation:

- real `/api/v1/system/status` adapter;
- owner-only live System Dashboard integrated into current customer dashboard;
- CPU/RAM/disk/load/uptime/RX/TX from backend samples;
- API/DB/Runtime/Telemetry dependency badges;
- network rate shown only when `rate_available=true`;
- no invented Accounting/Online card;
- safe ErrorBoundary that does not render raw exception details.

GREEN attempt exposed only an over-specific ASCII-digit test; the test was corrected to be locale-safe without weakening uptime validity semantics.

Implementation head `6a61512449884432477b1c107de6ac80e9e0d69c` passed:

- CI #1060 — Web tests/build, Go vet/tests, PostgreSQL gates, full S04R, production bundle;
- WS1 Exact Accounting #176;
- WS1 Pinned Forwardproxy #160.

## Current Production pre-deploy truth

Last read-only pre-Task-4 deployment audit:

- API/Runtime/Telemetry/Caddy active and healthy;
- restart counters observed at 0;
- schema 11;
- six active users/Runtime credentials/ServiceTerms/direct Subscription tokens;
- direct accounting live;
- root filesystem about 79% used;
- Caddy panel root `/var/www/pvnaive-preview/current`;
- `/opt/pvnaive/web/current` and `/opt/pvnaive/db/current` release symlinks present;
- backup age recipient/private key present with restricted permissions;
- scheduled PVNaive timers not active before Task #4 deployment;
- deployment marker drift present before new release tooling is deployed.

No Production mutation was performed during extraction/code verification.

## Final Task #4 gate now

1. canonical docs updated on PR #29;
2. run CI + Exact Accounting + Pinned Forwardproxy on final exact documentation head;
3. final diff/secret/rollback review;
4. merge with expected head SHA;
5. Production re-audit + fresh backups;
6. build from exact merged commit;
7. guarded same-schema deploy;
8. verify panel/API/Runtime/Telemetry/Caddy invariants/timers/provenance/Doctor;
9. only then mark Task #4 Production-complete and begin Task #5 legacy/adopted accounting baseline truth.

## Logging rule

Every transition records date, exact commit/source, what changed, tests/CI, Production evidence when applicable, result and exact next step. Real failures that revealed defects remain part of the evidence chain.

## 2026-08-31 continuous orchestration checkpoint — Task12

- GitHub `main`: `3cd98a1bc1358fc3b58dd8642646da122cac84c6`.
- Production remains on schema 16 / deployed commit `4f7853fced230d65644a94ed8b50cf3c6c74ca98`; readiness is green and PostgreSQL/Caddy/API/Runtime/Telemetry services are active.
- Task12 exact local head `bea9ebeba17f411e5926165483a546a988c4b20d`: independent release review `VERDICT=PASS`, no BLOCKER/HIGH; fresh `go test ./...`, 18/18 web files / 60/60 tests and production web build pass. Prior full PG18 schema17 and pinned reproducible Caddy proofs remain the release DB/data-plane gates.
- GitHub staging branch `lead/task12-session-management-published-2026-08-31` is intentionally NOT mergeable yet: its tree is 14 files behind the exact reviewed local tree. Do not merge or deploy the partial branch.
- Active independent work: BUG-001 refresh reuse DB proof; Task13 exact kill-session; Task14/15 deterministic race RED proof; Task16 bounded privacy-aware history proof.
- Task15 unit/model candidate is not promotable until real PostgreSQL concurrency proof passes. Task16 is not promotable until migration numbering is reconciled after the P0 BUG-001 schema change and PG18/retention proof passes.
- Next release action: publish the exact Task12 tree, run exact-head GitHub CI including schema17 and pinned forwardproxy/Caddy, then guarded Production backup → schema16→17 → API/web/Caddy rollout → postflight.

## 2026-09-14 02:53 coordinator checkpoint
- Verified canonical main `3b49e0b9dd10cd720dbbf33be50361f3ec003dce`; push CI `34789203279` SUCCESS.
- Task13 PR #108 refreshed by fast-forwarding its exact implementation history with current docs/spec main; new head `d42f1db4112fe43e71f4cd1b7feff941d78094af`, GitHub mergeable. Independent `git diff --check`, Docker Go 1.25 gofmt/vet/test and web 19/64 + build PASS. Fresh exact-head CI/accounting/forwardproxy are running; real pinned-Caddy HTTP/1.1 + HTTP/2 acceptance remains mandatory.
- Production remains mutation-free. Latest fresh external read-only probe: nip.io live/ready/panel healthy; `namir.softarg.ir` still inside documented Let's Encrypt retry window. Primary shell-level deployed identity/backups/rollback remain unverified.
- Opened #109 as the independent next roadmap lane: R1 / STEER-001 trusted-boundary network telemetry. Worker 3=forwardproxy sampling, Worker 2=DB/ingest/replay semantics, Worker 1=independent schema/CI/test-harness review, Worker 4=E2E rehearsal, Primary=read-only Production.


## 2026-09-14 — Master Upgrade Pack R1→R8 execution (Super-Z operator session)

Verified on dev-server golang:1.25 CI-mirror before every push; evidence in git history.

| Workstream | Slice merged | Main |
|---|---|---|
| Program docs | `docs/AGENT_TASKS.md` canonical board (owner R1→R8 merged), `docs/STEERING_SPEC_FA.md`, `docs/CAMO_ACCESS_UI_SPEC_FA.md` | 9409719, a14201a, 3b49e0b |
| R2 steering | `internal/steering`: deterministic scoring, EWMA, per-user phase offsets, hysteresis 25%/2w, kill-switch, Unknown policy; 10 simulation tests (STEER-002 core) | daee7e2 |
| R3 renderer | `internal/subscription/render.go`: UA negotiation (Clash/sing-box/Karing/Hiddify/v2ray/naive), url-test tolerance 50 + load-balance round-robin, never interrupt-exist-connections, Profile-Update-Interval + Subscription-Userinfo headers, single-node Profile.Node (pool list pending R4/R5) | 715cc5c |
| R8 UI | `web/src/design-tokens.ts` + stealth login (hover/focus reveal, reduced-motion, no layout-shift tell, sr-labels + autocomplete) + pure reveal state machine tests; vitest 69/69 + build green | b1a7935 |
| R6 coverd | migration 0029 (cover_content partitions + SECURITY DEFINER + RLS), 6 original persona packs, accurate jalaali-js port, polite RSS/Atom fetcher (stale-keeps-snapshot), hygiene server (human 404 probe sweep); 15 tests; Caddy routing deliberately OFF until live CAMO gates | 75415be |
| R4 fleet | `internal/fleet/manifest.go`: Ed25519 signed manifests, fail-closed verify + expiry; 6 tests | ab0b1c3 |

Open lanes (unclaimed or worker-assigned): R1 forwardproxy TCP_INFO sampling + ingest (#109 worker split),
R5 pool manager UI, R7 panel access backend (base path/port flows), R8 live-charts streaming backend,
R6 Caddy routing flip + coverd scheduler wiring, R2/R3 wiring to live telemetry.
