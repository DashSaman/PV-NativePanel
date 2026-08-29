# AGENTS.md — PVNaive mandatory agent instructions

Last reconciled: 2026-08-30

## Mission

PVNaive is a **standalone-first** management plane for standard NaiveProxy. Standalone correctness and Production safety come before fleet/multi-node expansion.

Current main already contains customer/product management, deterministic Subscription/Account Page delivery and exact direct-Naive accounting. Agents must not rebuild those features from old S04/S05 snapshots.

Unsupported or not-yet-enforced session/device/speed/reset/reseller/fleet capabilities must never be presented as implemented.

## Single source of truth / read order

Before any change, read in this order:

1. `OWNER_REQUIREMENTS.md` — Owner product/UX invariants.
2. `PROJECT_STATUS.md` — current repository + Production truth.
3. `HANDOFF.md` — exact current continuation.
4. `ROADMAP.md` — Owner-ordered Production-Ready execution ledger and historical PVN crosswalk.
5. `docs/PANEL_PARITY_MASTER_2026-08-30.md` — current 120-feature competitor/gap matrix.
6. `KNOWN_ISSUES.md` — current bugs/security/debt/ops risks.
7. `AGENT_TASKS.md` — workstream ownership and conflict rules.
8. `WORKLOG.md` — significant completed/failed work; do not repeat it.
9. `FEATURE_MATRIX.md` — short-form actual-vs-target capability truth.
10. `CONTINUE_HERE.md` — interruption recovery pointer.
11. relevant design/spec/plan under `docs/superpowers/`.
12. before **any Production mutation**, independently re-read newest `main`, current `ops/evidence/*`, live service/database/Caddy state, current backups and rollback plan.

Historical stage files such as `docs/PILOT_INSTALL_FA.md` and `docs/PANEL_PARITY_MASTER_2026-08-29.md` are explicitly archived/superseded and must not drive current Production actions.

## Current baseline

At the 2026-08-30 reconciliation start:

- audited main: `a021aa4b62c35b775fb521d042b2f8e6dbde10b0`;
- Production schema: 11;
- API, Runtime Agent, Telemetry Agent and Caddy active;
- exact direct-Naive accounting/session data live;
- six audited ServiceTerms had complete accounting projections;
- `/sub/<token>` machine delivery and `/s/<token>` human Account Page exist;
- customer CRUD/product plans/groups/tags/renewal/search/bulk foundations exist;
- shared hard-quota reservation/settlement and trusted first-CONNECT producer core exist.

Always fetch latest main again. The SHA above is only an audit marker, not a permanent pin.

## Work that must not be duplicated

Do not recreate from old branches:

- Runtime credential import/create/update/rotate/disable/revoke;
- expected-SHA Caddy validate/backup/reload/rollback safety;
- customer create/adopt/edit/suspend/resume/revoke-safe-delete;
- quota/unlimited/add/set volume;
- expiry/no-expiry/creation/first-CONNECT/manual validity/extend days;
- plans/groups/tags/notes/renewal;
- search/filter/sort/pagination;
- supported bulk preview/idempotent execute;
- `/sub`, `/s`, local QR, Subscription reissue/password-rotation separation;
- exact direct accounting and restart-safe telemetry core.

Before implementing any requested feature, verify handler/store/schema/UI/tests on **latest main** rather than trusting route names or old docs.

## Current ordered execution chain

Do not reorder unless a real technical dependency is documented with evidence.

1. audit latest main/Production;
2. current competitor parity;
3. canonical documentation truth;
4. safe PR #16 extraction/integration;
5. legacy/adopted accounting baseline truth;
6. `/s` accounting/presence completion;
7. Manual Reset Usage;
8. Bulk Reset Usage;
9. periodic traffic reset execution;
10. hard-quota controlled Production proof;
11. first-successful-CONNECT controlled Production proof;
12. sessions/kill/concurrent/IP/history;
13. HWID/speed PoCs;
14. reseller/RBAC/wallet/ledger/restrictions;
15. history/audit;
16. notifications/Telegram;
17. dashboard/monitoring/logs/diagnostics/Doctor;
18. scheduled backup/restore;
19. API/OpenAPI/rate-limit;
20. security defects/authorization/IDOR/fuzz/supply-chain;
21. multi-node/fleet;
22. installer/upgrade/rollback;
23. client compatibility;
24. load/capacity;
25. final bulk/search/UI/docs/clean install/Production smoke/RC.

The detailed 50-task sequence and exact gates are in `ROADMAP.md` / `AGENT_TASKS.md`.

## Before starting a task

Record:

```text
AGENT
TASK-ID / MASTER ORDER
GOAL
FILES
DEPENDENCIES / EVIDENCE
```

Then:

1. fetch latest main and inspect code, not only docs;
2. inspect open/merged PRs and active branches to avoid duplicate work;
3. verify no active lane owns the same files;
4. for production code/bugfixes follow TDD: failing test → observe correct RED → minimal implementation → GREEN;
5. preserve unrelated/newer behavior;
6. record architecture/safety changes in relevant spec/handoff;
7. require exact-head CI before completion or merge.

## Mandatory work report

After every numbered work unit record:

```text
TASK #:
STATUS:
WHAT CHANGED:
FILES:
TESTS:
CI:
PRODUCTION:
EVIDENCE:
REMAINING:
NEXT TASK:
```

No agent may say only “Done”. Final DONE transition belongs to Lead/Agent-REVIEW after evidence review.

## Red lines

- Do not make standalone release depend on Controller/fleet/Iran topology.
- Do not alter Naive wire protocol without ADR, benchmark and client-compatibility evidence.
- Do not build default random chaff/fake browsing traffic.
- Do not use estimated access-log traffic as exact billing.
- Do not fabricate usage/remaining/online/HWID/device/speed/session-limit state.
- Do not commit passwords, tokens, subscription secrets, runtime secrets, private keys, raw secret-bearing Caddyfiles or Production dumps.
- Do not place Web UI/API on the data-plane availability path.
- Do not close SSH or casually change firewall.
- Do not run destructive migration/uninstall without backup + validation + rollback plan.
- Do not use unpinned `latest` for Production dependencies/artifacts.
- Do not let the unprivileged API gain arbitrary root shell/path/service/URL access.
- Do not interpret a route declaration/schema table as implemented behavior.
- Do not blind-merge stale PR #16 or old S04/S05 branches.
- Do not reset/force-push main to reconcile branch history.
- Do not copy GPL/AGPL competitor code without explicit license-compatibility review.

## Customer / Subscription invariants

- existing users must not be deleted during migrations/reconciliation;
- customer edit quota/expiry must not rotate password or Subscription token;
- View QR, Copy Subscription and View Account Page are read-only;
- `/s` is human-facing;
- `/sub` is machine/client-facing;
- Reissue Subscription is a separate explicit mutation from Rotate Password;
- first-use validity starts only on successful authenticated Naive CONNECT, never view/copy/health/reload/failed auth;
- if an accounting baseline cannot be proven, report `Unknown`, never fake zero.

## Runtime / Production mutation rule

Production mutation is forbidden merely because code or CI passes.

Before **every** Production mutation:

1. current read-only preflight;
2. DB backup;
3. config backup;
4. Caddy backup;
5. web backup;
6. binary backup;
7. explicit rollback plan;
8. validate/stage;
9. one bounded apply;
10. postflight;
11. rollback on failure;
12. evidence capture with secret redaction.

Runtime/Caddy changes follow:

`expected SHA → exact backup → validate → install/apply → reload where appropriate → verify → exact rollback`

Never restart Caddy as routine mutation when reload is sufficient.

## Current known P0 defects

Agents must not silently erase these until regression tests prove closure:

- refresh-token reuse-family detection can be bypassed by the pre-rotation `revoked_at IS NULL` lookup;
- generic authenticated HTTP response can be written before durable DB commit and commit error is ignored;
- readiness is not yet bounded DB/schema-backed.

See `KNOWN_ISSUES.md` for exact evidence and Done gates.

## Stale PR handling

- PR #4: old branch obsolete overall; small Karing sing-box export idea may be re-evaluated during client compatibility.
- PR #5/#6: superseded/merged elsewhere; do not merge.
- PR #8: superseded by integrated WS1 exact accounting; archive only.
- PR #16: still contains useful operations/observability foundations but must be extracted manually onto latest main, unit-by-unit, with current tests/CI.

## Definition of Done

A feature is not DONE unless all applicable Owner DoD items are satisfied, including:

- real backend and correct schema;
- authorization/tenant boundary;
- usable UI where operator-facing;
- no secret leak;
- idempotency where needed;
- failure-path tests;
- unit/integration/web tests;
- `go vet ./...` and `go test ./...`;
- web tests/build;
- exact-head CI;
- rollback test for Runtime-affecting changes;
- live verification for Production-facing changes;
- docs/evidence updates;
- no regression of earlier features.

## Context recovery

If context may be lost, update repository state first. A new Chat/Agent must be able to continue from canonical files without the old conversation or historical stage runbooks.
