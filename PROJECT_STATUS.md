# PVNaive — Canonical Project Status

Last updated: 2026-08-28

> This file is the canonical **development** status for the active branch. Production truth must also be cross-checked against `main:CONTINUE_HERE.md`, `main:ops/S04_LIVE_STATE.md`, and production evidence before any live mutation.

## Project

PVNaive is PVNETWORK's standalone-first management plane for standard NaiveProxy. The immediate Pilot target is intentionally narrow: the Owner manages real Naive credentials from the panel and hands a `naive+https://...` link to customers. Customer portal, quota/accounting, expiry, reseller and subscription lifecycle remain later phases.

## Repository state

- Repository: `DashSaman/PV-NativePanel`
- Default branch: `main`
- Active implementation branch: `s04-auth`
- Draft PR: `#2` — `S04-AUTH: production authentication foundation`
- `main` production-documentation head last audited: `b98ffbfbe8b30ddcc1bca06531650739a3647d22`
- Branches remain deliberately diverged; `PVN-070` owns non-force reconciliation.
- Never reset/force-push away production evidence or active S04R work.

## Current production state before S04R Pilot

Evidence on `main` proves:

- S00-S03 passed.
- PostgreSQL 18 is loopback-only and production schema is currently v2.
- API is healthy on `127.0.0.1:8080` only.
- one real Owner exists and real login/session/CSRF logout/revocation passed.
- public panel/API exposure passed at `https://namir.softarg.ir/panel/`.
- existing Caddy/Naive and camouflage were preserved during panel exposure.
- S04 is still formally `IN PROGRESS`; formal closure is `PVN-036`.

No S04R migration/runtime mutation has yet been recorded as completed on production in this branch's canonical status.

## S04R implementation state

Development is complete through disposable rehearsal (`PVN-020`…`PVN-027`). Implemented and covered by tests/rehearsal:

- AES-GCM runtime secret envelope, SHA-256 fingerprinting and conservative username/password policy;
- byte-preserving `forward_proxy` parser/renderer with fail-closed ambiguity/injection handling;
- root Runtime Agent on a fixed Unix socket, with no arbitrary shell/path/service/URL API;
- expected-SHA Caddy operator with exact backup, validate-before-write, reload-only postflight and rollback;
- runtime credential PostgreSQL store + desired/apply/applied revision saga and compensation;
- dedicated reconciliation-required error when DB finalization and Runtime rollback cannot be reconciled;
- Owner-only Runtime API with CSRF, idempotency, optimistic revisions and secret redaction;
- `/panel/#/runtime/naive` UI for import/create/rename/rotate/enable/disable/revoke;
- generated password is returned only after successful commit and only once to the browser;
- one-click copy-ready `naive+https://...` URI for Karing/compatible clients;
- last-active credential protection;
- full PG18 + API + Unix-socket + safe Runtime rehearsal;
- pinned Naive Caddy `v2.11.2-naive` proof that multiple `basic_auth` directives validate/adapt;
- guarded S04R bundle with migration 0003, three binaries, web build, systemd units, preflight and upgrade scripts.

## Verification checkpoints

### Full S04R implementation/bundle checkpoint

Commit: `a41fd84c2f17076a3b190eafad3539c47b430503`

CI run `33190295766`:

- Go: PASS
- Web: PASS
- PostgreSQL 18 / database safety contracts: PASS
- exact pinned Caddy proof: PASS
- S04 auth rehearsal: PASS
- full S04R runtime rehearsal: PASS
- production S04R bundle contract: PASS
- archive checksum: PASS
- uploaded artifact: `PVNaive-S04-a41fd84c2f17076a3b190eafad3539c47b430503`

The downloaded artifact was independently unpacked and both the outer archive checksum and every file in internal `SHA256SUMS` verified successfully.

### Customer-handoff UI increment

The subsequent UI slice adds the tested URI builder and one-click Karing/Naive link. It followed RED→GREEN TDD: run `33190665847` failed only because `buildNaiveURI` did not exist; after implementation the web job on run `33190796760` passed all tests and the production web build.

A fresh full CI on the final documentation HEAD is still required before a completion/release-ready claim.

## Immediate live task

`PVN-028` — **read-only live preflight**.

Use `docs/PILOT_INSTALL_FA.md` and perform the live process one step at a time:

1. verify artifact checksums;
2. run `S04R-preflight.sh` read-only;
3. continue only if `PREFLIGHT_RESULT=PASS`;
4. pass the exact `CADDYFILE_SHA256` into `S04R-upgrade.sh`;
5. require encrypted DB backup + schema v3 + unchanged Caddy SHA/PID/NRestarts;
6. Owner securely imports the existing credential;
7. create one new generated credential;
8. test the generated Naive link in Karing before handing it to a customer;
9. record live evidence under `ops/evidence/` and only then close `PVN-028/029`.

## Pilot capability boundary

### Ready in code/rehearsal

- Owner authentication and protected panel
- real Naive credential import and management
- multiple credentials
- one-time generated password
- copy-ready Naive/Karing link
- safe Caddy lifecycle and rollback

### Not part of this Pilot yet

- customer self-service login/portal
- traffic quota or exact billable usage
- automatic commercial expiry/reset
- device/HWID/session/speed enforcement
- reseller/credit
- subscription page/token lifecycle
- notifications
- generic fresh-server installer

Customers receive only their Naive link/credential, never Owner panel credentials.

## Numerical progress

Progress is an equal-weight count over the 72 concrete tasks in `ROADMAP.md`; in-progress work receives no partial credit.

| Metric | Count |
|---|---:|
| Total tracked tasks | 72 |
| DONE | 27 |
| IN PROGRESS | 2 (`PVN-028`, `PVN-068`) |
| BLOCKED | 0 |
| TODO | 43 |
| Overall completed | **37.5%** |

### Phase progress by task count

| Phase | Done / Total | Progress |
|---|---:|---:|
| A — Foundation / Database | 6 / 6 | 100% |
| B — Authentication / Public preview | 11 / 11 | 100% of implemented milestones; formal closure remains Phase D |
| C — S04R Naive credential management | 10 / 12 | 83.3%; live preflight/rollout remain |
| D — Security hardening / S04 closure | 0 / 7 | 0% |
| E — User / Plan / Reseller lifecycle | 0 / 8 | 0% |
| F — Accounting / full Naive runtime | 0 / 7 | 0% |
| G — Subscription / Notification | 0 / 6 | 0% |
| H — Installer / Operations / Release | 0 / 10 | 0%; S04R has a stage-specific Pilot upgrade path, not the general installer |
| I — Governance / quality | 0 / 5 | PVN-068 still syncing canonical PM docs |

## Known release/security boundaries

The Pilot is not equivalent to final Production R1. Highest-priority open items remain:

1. `PVN-030` refresh-token reuse-family detection;
2. `PVN-031` transaction commit-before-success integrity;
3. `PVN-034` HTTP/IP-aware auth abuse controls;
4. `PVN-035/036` formal public security review and independent S04 closure;
5. `PVN-045+` exact accounting before quota/billing claims;
6. `PVN-058+` generic fresh installer/release lifecycle.

For Pilot use, keep panel access Owner-only, use strong Owner credentials/MFA where configured, and never give management login details to customers.

## Read first

1. `HANDOFF.md`
2. `ROADMAP.md`
3. `docs/PILOT_INSTALL_FA.md`
4. `KNOWN_ISSUES.md`
5. `AGENT_TASKS.md`
6. `WORKLOG.md`
7. `FEATURE_MATRIX.md`
8. `docs/superpowers/specs/2026-08-28-naive-runtime-credentials-design.md`
9. `docs/superpowers/plans/2026-08-28-naive-runtime-credentials.md`
10. before any live change: `main:CONTINUE_HERE.md`, `main:ops/S04_LIVE_STATE.md`, newest `main:ops/evidence/*`
