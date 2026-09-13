# AGENTS.md — PVNaive mandatory agent instructions

Last reconciled: 2026-09-13

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
8. `docs/DEVELOPMENT_WORKER_POOL.md` — canonical five-server development inventory, resource policy and continuity rules.
9. `WORKLOG.md` — significant completed/failed work; do not repeat it.
10. `FEATURE_MATRIX.md` — short-form actual-vs-target capability truth.
11. `CONTINUE_HERE.md` — interruption recovery pointer.
12. relevant design/spec/plan under `docs/superpowers/`.
13. before **any Production mutation**, independently re-read newest `main`, current `ops/evidence/*`, live service/database/Caddy state, current backups and rollback plan.

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

## Mandatory five-server development pool

The Owner has made five servers available for PVNaive development. Their canonical identities, SSH aliases, host-specific safety notes and resource limits are in `docs/DEVELOPMENT_WORKER_POOL.md`.

Agents must actively use safe free capacity across the pool rather than unnecessarily serializing independent work on Primary. The normal target ceiling is approximately 70% total host pressure; unrelated Production workload always has priority. Heavy worker jobs should be resource-bounded where practical.

Parallelism does not waive dependency or verification rules:

- one writer per worktree;
- separate branch/workspace per coding task;
- independent verification preferably on a different worker;
- no important result may remain only on a temporary worker;
- do not manufacture parallel work that conflicts with an earlier required task;
- when a worker becomes free and a non-conflicting roadmap task is ready, dispatch it rather than leaving capacity idle;
- failed/exited agents must leave their workspace/report intact for diagnosis and safe restart/reassignment.

Development workers are not automatically product Fleet nodes.

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

## Live environment & client facts (verified 2026-09-13)

Hard-won operational truths from debugging the live deployment and strict clients. Re-verify on latest main/live server before relying on them.

- **Subscription URI must carry the port**: `naive+https://user:pass@host:443`. `PVNAIVE_NAIVE_PUBLIC_HOST` (derived from `PVNAIVE_DOMAIN`) must include `:443`; repo contract tests enforce it. Bare-host URIs fail to import in strict clients such as Karing.
- **TLS and bare IPs do not mix**: a deployment addressed by bare IPv4 gets Caddy `tls internal` (self-signed). Chromium-based clients (Android Karing) then fail with `handshake failed ... net_error -202` (= `ERR_CERT_AUTHORITY_INVALID`). Serving the same stack behind a resolvable hostname gives a real Let's Encrypt certificate automatically (currently a `<ip>.nip.io` wildcard in the test deployment; prefer an Owner-owned domain for release).
- **Caddyfile lifecycle**: the entrypoint renders `/etc/caddy/Caddyfile` ONLY when the file is absent (it is bind-mounted under the data dir and therefore persists across restarts). Regenerating it is a deliberate, backed-up operation, never a casual restart.
- **`probe_resistance` masks auth failures**: unauthorized `CONNECT` attempts to `forward_proxy` return a disguised 404. A 404 from the proxy port does NOT mean the route is missing — retest with valid customer credentials before diagnosing.
- **Database schema layout**: all panel tables live in the `pvnaive` PostgreSQL schema, not `public`.
- **Auth cookies**: session cookie `__Host-pvnaive_session`; CSRF cookie ends with `pvnaive_csrf`. Mutating API calls require the `X-CSRF-Token` header; customer creation additionally requires an `Idempotency-Key` and `validity.mode: on_creation`.
- **Karing diagnostics**: "no server available" means the subscription downloaded but parsed 0 nodes (check URI format/User-Agent); TLS `net_error -202` means an untrusted certificate (self-signed/bare IP), not a malformed URI.
- **Readiness gate**: readiness compares the compose-provided `PVNAIVE_EXPECTED_SCHEMA_VERSION` with the live database schema version; a mismatch marks the stack not-ready. Align the env whenever the schema version is bumped.

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

See `KNOWN_ISSUES.md` for exact evidence and Done gates. Always verify latest main before assuming a listed defect remains open; canonical docs may lag a newly merged fix.

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

## Live session results (2026-09-14, agent Super-Z)

Facts other agents can rely on; re-verify anything you mutate:

- **GitHub push is unblocked.** The fine-grained PAT now has real Contents:write (Git Data blob probe returned 201 — always verify with the blob probe, never trust the `/repos` permissions JSON). Five commits pushed to `main` as `0c245b5..dbbcb98`; another bot's docs-refresh commits were rebased on cleanly.
- **Production domain is `namir.softarg.ir`** (A record → 45.141.148.59). Let's Encrypt cert issued and auto-renews. `.env` keys: `PVNAIVE_DOMAIN=namir.softarg.ir`, `PVNAIVE_NAIVE_PUBLIC_HOST=namir.softarg.ir:443`. All subscription URIs are domain-based now — existing Karing clients must re-import their subscription once.
- **BBR + fq are live on the host** via `/etc/sysctl.d/99-pvnaive-tuning.conf` (bbr, fq, 64MB rmem/wmem, mtu probing, fastopen, backlog/somaxconn) plus `tc qdisc replace dev eth0 root fq`. Verified cubic→bbr. This file is host-side, not in the repo.
- **Self-service account security is live** (deployed image `pvnaive:fix2`): `POST /api/v1/me/password` + `PATCH /api/v1/me/profile`, store layer `store_me.go`, SECURITY DEFINER functions via deployed migration 0027, UI `web/src/SettingsSecurity.tsx` (owner nav item "امنیت و حساب" → `#/settings/security`). Live 7-step E2E passed (change→login new→restore; 401 wrong-current; 400 short).
- **UI/UX overhaul "Amber Command Deck" merged to `main`** (not yet in the deployed image): self-hosted Vazirmatn (`@fontsource/vazirmatn`), SVG icon set in `web/src/ui.tsx`, full design-system rewrite of `web/src/styles.css` (glass surfaces, aurora background, luminous borders, entrance animations, focus rings, reduced-motion safe, light+dark), iconified sidebar/login/dashboard, honest cumulative expiry sparkline. All 61 web tests + typecheck + build green locally.
- **Competitor research moved into the repo**: `docs/competitor/3xui-v3-release-notes.md` (full v3.0→v3.7 release notes). Numbered remaining-work ledger lives at the end of `ROADMAP.md`.
- **Two open infra issues are documented in `KNOWN_ISSUES.md`**: DEPLOY-001 (boot config renderer drops active credentials — P0) and LINEAGE-001 (deployed migrations 0022..0027 vs repo 0021 — reconcile before next repo-built deploy).
- Live login (owner): `https://namir.softarg.ir/panel/` with `admin@pvnaive.local` / password in `/opt/pvnaive/.env` on the server. Server root access: see `scripts/ssh_run.py` in the agent workspace (password list maintained there).

- **Readiness 503 incident (2026-09-14, resolved):** the fix2 image bakes `PVNAIVE_EXPECTED_SCHEMA_VERSION=23` in Docker ENV while the live DB is schema 27 — the readiness gate fail-closed for ~2h (panel SPA kept serving, so it looked fine externally). Fixed live via compose override `PVNAIVE_EXPECTED_SCHEMA_VERSION: "27"` in `/opt/pvnaive/docker-compose.override.yml`; container healthy. The image must render this from its own migration count instead of hardcoding.
- **pvbootstrap is obsolete as a proxy account:** the custom forward_proxy module requires a runtime UUID mapping per basic_auth user; unmapped users are rejected at config load. The live Caddyfile holds customer credentials only (22 entries). Do not re-add pvbootstrap; do not reload unvalidated Caddyfiles (failed reload leaves the running config intact).
