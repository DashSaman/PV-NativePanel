# PVNaive — Worklog

This log records significant verified work. Fine-grained historical live evidence remains under `ops/evidence/`; this file prevents future agents from repeating already-verified work.

| Date (UTC) | Agent / provenance | Task | Description | Files / evidence | Tests / verification | Commit / state | Result |
|---|---|---|---|---|---|---|---|
| 2026-08-27 | prior PVNaive workflow | PVN-001 | Product naming fixed to PVNaive / NaiveProxy while preserving repository name | docs/handoff history | review | historical | DONE |
| 2026-08-27 | prior PVNaive workflow | PVN-002 | Read-only production preflight of target host/domain/Caddy/network | `ops/evidence/*S03*preflight*` | live invariants | evidence | DONE |
| 2026-08-27 | prior PVNaive workflow | PVN-003 | S02 filesystem/service foundation deployed | `ops/DEPLOYMENT_PROGRESS.md` history | stage + postflight | historical | DONE |
| 2026-08-27 | prior PVNaive workflow | PVN-004 | PostgreSQL 18 v1 schema, roles, RLS, security context | `db/migrations/0001_*`, db tests | PG18 CI + live S03 | main lineage | DONE |
| 2026-08-27 | prior PVNaive workflow | PVN-005 | encrypted age backup/restore/health gates | `scripts/db/*`, `tests/db/*` | backup SHA, restore drill, timer | S03 evidence | DONE |
| 2026-08-27 | prior PVNaive workflow | PVN-006 | S03 production and independent postflight | S03 production evidence | live postflight | evidence | DONE |
| 2026-08-27/28 | prior PVNaive workflow | PVN-007 | S04 auth architecture/spec/plan | `docs/superpowers/specs/2026-08-28-s04-auth-design.md`, plan | review + later implementation proof | branch history | DONE |
| 2026-08-28 | prior PVNaive workflow | PVN-008..013 | Auth migration, crypto/session/TOTP/store/HTTP/bootstrap, CI rehearsal and bundle | auth/db/http/web files | Go/Web/PG18/rehearsal/bundle | verified lineage | DONE |
| 2026-08-28 | prior PVNaive workflow | PVN-014 | Live localhost deployment/recovery; fixed dependency/backup/DB/service failures | S04 live evidence/history | safe recovery + rollback invariants | evidence | DONE |
| 2026-08-28 | prior PVNaive workflow | PVN-015 | localhost independent postflight | main evidence | service/listener/schema/Caddy/SSH/firewall | main docs | DONE |
| 2026-08-28 | prior PVNaive workflow | PVN-016 | real Owner bootstrap and auth lifecycle | main evidence | login/me/CSRF logout/revocation | main docs | DONE |
| 2026-08-28 | prior PVNaive workflow | PVN-017 | public `/panel/` and `/api/` exposure | main S04 Caddy evidence | validate, reload-only, PID/NRestarts, panel/API/assets/camouflage/forward_proxy | `b98ffbfb...` production-doc head | DONE |
| 2026-08-28 | Lead Engineer | PVN-018 | Owner-approved S04R Naive credential-management design and detailed TDD plan | S04R spec + plan | spec review | branch commits | DONE |
| 2026-08-28 | Lead Engineer | PVN-019 | migration 0003; owner-only runtime credential table + runtime revision idempotency; rollback/version-aware tests | migration + DB tests | full CI `33133931739` | `6178ab97...` | DONE |
| 2026-08-28 | Lead Engineer | PVN-020 | Runtime AES-GCM secret envelope, SHA fingerprint, CSPRNG generator and conservative username/password policy completed from prior RED tests | `internal/runtimecred/*` | Go tests + later full S04R CI | S04R branch | DONE |
| 2026-08-28 | Lead Engineer | PVN-021 | Byte-preserving Naive Caddy `forward_proxy` parser/renderer with fail-closed ambiguity/injection handling | runtime Caddy transform code/tests | parser tests; pinned exact-Caddy proof later green | S04R branch | DONE |
| 2026-08-28 | Lead Engineer | PVN-022/023 | Fixed-capability root Runtime Agent on AF_UNIX and expected-SHA Caddy operator with exact backup, validate-before-write, reload-only, postflight and rollback | `cmd/pvnaive-runtime-agent`, `internal/runtimeagent`, systemd/tests | unit/failure rehearsal | S04R branch | DONE |
| 2026-08-28 | Lead Engineer | PVN-024 | Runtime DB store + desired/apply/applied saga; added compensation when DB commit fails after Runtime apply | runtime store/service tests | failure injection + full rehearsal | S04R branch | DONE |
| 2026-08-28 | Lead Engineer | PVN-024/025 | Found missing double-failure distinction: DB finalization failure plus Runtime rollback failure initially returned generic consistency/runtime error | CI RED `33188928527`, HTTP RED `33189300005` | sentinel + HTTP mapping regression then green | follow-up commits | FIXED |
| 2026-08-28 | Lead Engineer | PVN-025 | Owner-only `/api/v1/runtime/naive` endpoints completed with CSRF, idempotency, optimistic revision and secret redaction | `internal/httpapi/*`, runtime service | Go/API tests + full rehearsal | S04R branch | DONE |
| 2026-08-28 | Lead Engineer | PVN-026 | Runtime UI completed for secure import/create/rename/rotate/enable/disable/revoke and one-time generated secret | `web/src/RuntimeNaive.tsx`, `runtime.ts`, CSS | Web tests/build; runtime client CSRF/If-Match/idempotency tests | S04R branch | DONE |
| 2026-08-28 | Lead Engineer | PVN-027 | Wired full S04R disposable rehearsal into CI. First run exposed missing rehearsal helper binary | `.github/workflows/ci.yml`, `tests/stages/S04R_full_rehearsal.sh` | RED `33188524298`; helper build added; later green | branch commits | DONE |
| 2026-08-28 | Lead Engineer | PVN-027 | Proved multiple `basic_auth` directives using exact pinned Naive Caddy `v2.11.2-naive` asset + fixed archive SHA; expanded full path with rename/stale-revision/envelope checks | `S04R_caddy_multi_auth_proof.sh`, full rehearsal, CI | exact `caddy validate/adapt` + full rehearsal | branch commits | DONE |
| 2026-08-28 | Lead Engineer | PVN-027/Pilot bundle | Added S04R production bundle contract. Old S04 artifact failed because Runtime Agent/layout were missing | release tests/workflow | RED `33189668543` (`bundle missing bin/pvnaive`) | `40b139cc...` test checkpoint | FIXED |
| 2026-08-28 | Lead Engineer | PVN-028/029 preparation | Added read-only live preflight contract and guarded existing-install upgrade path; first RED proved scripts were absent | `S04R-preflight.sh`, `S04R-upgrade.sh`, contract tests | RED `33189915241`; contract tests then green | branch commits | READY_FOR_LIVE_PREFLIGHT |
| 2026-08-28 | Lead Engineer | PVN-028/029 preparation | Upgrade safety: exact preflight Caddy SHA lock, encrypted DB backup before migration, runtime key preservation, schema v3, runtime/API units, no installer Caddy mutation, rollback | stage scripts/systemd/release builder | DB safety contracts | branch commits | READY_FOR_LIVE_PREFLIGHT |
| 2026-08-28 | Lead Engineer | QA | Web Runtime tests all passed but TypeScript production build caught test typing errors; DB contract also found a false-positive line-order check | web/db contract tests | RED `33190197374`; both test defects corrected | `068ab3f8...`, `a41fd84c...` lineage | FIXED |
| 2026-08-28 | Lead Engineer / REVIEW | PVN-020..027 | Fresh whole-system checkpoint: Go, Web, PostgreSQL 18, exact pinned Caddy proof, S04 auth rehearsal, full S04R rehearsal, bundle contract, checksum and artifact upload all passed | CI run + artifact | CI `33190295766` all jobs PASS | `a41fd84c2f17076a3b190eafad3539c47b430503` | VERIFIED GREEN |
| 2026-08-28 | Lead Engineer / REVIEW | Release QA | Downloaded CI artifact, verified outer `.tar.gz.sha256`, unpacked artifact and verified every entry in internal `SHA256SUMS`; release metadata reports stage `S04R-RUNTIME-CREDENTIALS`, schema 3, Caddy installer mutation false | artifact `PVNaive-S04-a41fd84...` | independent checksum/unpack verification | artifact id `9693505842` | VERIFIED |
| 2026-08-28 | Lead Engineer | PVN-026 | TDD customer handoff convenience: required URI builder before implementation | `web/src/runtime.test.ts` | RED `33190665847`: only new `buildNaiveURI` test failed; 17 prior tests passed | `e6cd67b3...` | EXPECTED RED |
| 2026-08-28 | Lead Engineer | PVN-026 | Added percent-encoded `naive+https://user:pass@host:443` builder and one-click Karing/Naive link in one-time secret dialog | `web/src/runtime.ts`, `RuntimeNaive.tsx` | Web job `33190796760`: tests + production build PASS | `69259a2e...` | DONE |
| 2026-08-28 | Lead Engineer / PM | PVN-028 | Added Persian guarded Pilot install/customer-handoff runbook; explicitly scopes current Pilot vs future quota/customer portal/release work | `docs/PILOT_INSTALL_FA.md` | documentation review; final HEAD CI still required | docs commits | IN_PROGRESS |
| 2026-08-28 | Lead Engineer / PM | PVN-068 | Canonical repository audit/PM documents updated from obsolete PVN-020 RED state toward S04R Pilot truth | ROADMAP/PROJECT_STATUS/AGENT_TASKS/FEATURE_MATRIX/KNOWN_ISSUES/WORKLOG/HANDOFF | cross-document review + final CI required | docs sync | IN_PROGRESS |

## Current exact continuation

- S04R code/rehearsal is complete through `PVN-027`.
- `PVN-028` is the current live edge: **read-only preflight only** until its output passes.
- Production schema remains v2 until the guarded S04R upgrade is actually executed and evidenced.
- Use `docs/PILOT_INSTALL_FA.md` for the exact existing-server Pilot sequence.
- Customers receive only generated Naive links/credentials; Owner panel credentials are never shared.
- After live `PVN-029` smoke, return immediately to P0 auth hardening (`PVN-030`, `PVN-031`, `PVN-034`) before formal S04 closure.

## Logging rule going forward

Every verified task transition adds one row containing: date, agent/workstream, task ID, exact change, files, tests/CI run, commit, and result. Failed attempts that reveal a real defect also remain recorded; do not erase them from history.
