# PVNaive — Worklog

This log records significant verified work. Fine-grained historical live evidence remains under `ops/evidence/`; this file prevents future agents from repeating already-verified work.

| Date (UTC) | Agent / provenance | Task | Description | Files / evidence | Tests / verification | Commit / state | Result |
|---|---|---|---|---|---|---|---|
| 2026-08-27 | prior PVNaive workflow; exact model not recorded in repo | PVN-001 | Product naming fixed to PVNaive / NaiveProxy while preserving repository name | docs/handoff history | review | historical | DONE |
| 2026-08-27 | prior PVNaive workflow; exact model not recorded | PVN-002 | Read-only production preflight of target host/domain/Caddy/network | `ops/evidence/*S03*preflight*` | live invariants | evidence | DONE |
| 2026-08-27 | prior PVNaive workflow; exact model not recorded | PVN-003 | S02 filesystem/service foundation deployed | `ops/DEPLOYMENT_PROGRESS.md` history | stage + postflight | historical | DONE |
| 2026-08-27 | prior PVNaive workflow; exact model not recorded | PVN-004 | PostgreSQL 18 v1 schema, roles, RLS, security context | `db/migrations/0001_*`, db tests | PG18 CI + live S03 | `d0398cd1...` main baseline later included it | DONE |
| 2026-08-27 | prior PVNaive workflow; exact model not recorded | PVN-005 | encrypted age backup/restore/health gates | `scripts/db/*`, `tests/db/*` | backup SHA, restore drill, timer | S03 evidence | DONE |
| 2026-08-27 | prior PVNaive workflow; exact model not recorded | PVN-006 | S03 production and independent postflight | `ops/evidence/S03-20260827T215310Z-production-pass.md` and related | live postflight | S03 evidence | DONE |
| 2026-08-27/28 | prior PVNaive workflow; exact model not recorded | PVN-007 | S04 auth architecture/spec/plan | `docs/superpowers/specs/2026-08-28-s04-auth-design.md`, plan | review + later implementation proof | branch history | DONE |
| 2026-08-28 | prior PVNaive workflow; exact model not recorded | PVN-008..013 | Auth migration, crypto/session/TOTP/store/HTTP/bootstrap, CI rehearsal and bundle | `0002_auth_foundation`, `internal/auth`, `internal/httpapi`, S04 scripts/tests | Go/Web/PG18/rehearsal/bundle | verified S04 bundle lineage includes `b4803e27...` | DONE |
| 2026-08-28 | prior PVNaive workflow; exact model not recorded | PVN-014 | Live localhost deployment/recovery; fixed missing `file`, backup collision, DB/service issues | S04 live evidence/history | repeated safe recovery + rollback invariants | multiple commits/evidence | DONE |
| 2026-08-28 | prior PVNaive workflow; exact model not recorded | PVN-015 | final localhost independent postflight | main evidence | service/listener/schema/Caddy/SSH/firewall | `431fec63...`, `0a406fa9...` documentation | DONE |
| 2026-08-28 | prior PVNaive workflow; exact model not recorded | PVN-016 | real Owner bootstrap and auth lifecycle | main evidence | real login/me/CSRF logout/revocation | `acc8c4e0...`, `fbc7715d...`, `87a8c78f...` docs | DONE |
| 2026-08-28 | prior PVNaive workflow; exact model not recorded | PVN-017 | public `/panel/` and `/api/` Caddy exposure | `main:ops/evidence/S04-20260828T010546Z-caddy-panel-exposure-pass.md` | validate, reload-only, PID/NRestarts, panel/API/assets/camouflage/forward_proxy | `c0255867...`, `b98ffbfb...` | DONE |
| 2026-08-28 | Lead Engineer | PVN-018 | Owner-approved S04R Naive credential-management design and detailed TDD plan | `docs/superpowers/specs/2026-08-28-naive-runtime-credentials-design.md`, plan | spec review | spec/plan branch commits | DONE |
| 2026-08-28 | Lead Engineer | PVN-019 | TDD migration 0003; owner-only runtime credential table + runtime revision idempotency; repaired rollback/version-aware tests | migration, DB tests, CI workflow | RED observed, then full CI run `33133931739`: Go/Web/DB/rehearsal/bundle PASS | green checkpoint `6178ab97e3ac328189e4baddbf578ebe79469c3b` | DONE |
| 2026-08-28 | Lead Engineer | PVN-020 | Added failing runtime secret/policy tests before implementation | `internal/runtimecred/secret_test.go`, `policy_test.go` | CI `33134072689`: Go RED at `undefined: ValidateUsername`; DB/Web PASS | `2c457c39...`, `cf0fbf3f...` | IN_PROGRESS / expected RED |
| 2026-08-28 | Lead Engineer / PM | PVN-068 | Audited repo, branches, PR, issues, code/docs/CI and production-vs-development truth; found 123-ahead/37-behind divergence and stale docs | GitHub audit; `PROJECT_STATUS.md` | cross-check `main:CONTINUE_HERE.md`, `main:ops/S04_LIVE_STATE.md`, PR #2, compare refs | `c8906836622d12611851de9e4629f334d167b354` | IN_PROGRESS |
| 2026-08-28 | Agent-RESEARCH workstream, reviewed by Lead | PVN-068 / matrix | Reused existing pinned deep audit and refreshed reference heads for 3x-ui/PasarGuard/Marzban; audited OV-PvNetwork public RC vs production-derived distinction | `docs/PANEL_DEEP_AUDIT_FA.md`, competitor repos, `OV-PvNetwork/docs/FEATURES.md` | source/commit cross-check | documentation batch pending at entry time | REVIEWED |

## Current exact continuation

- Last fully green feature checkpoint: `6178ab97e3ac328189e4baddbf578ebe79469c3b`, CI `33133931739` all jobs PASS.
- Current feature task: `PVN-020` and current test state is intentionally RED.
- Current project-management task: `PVN-068`.
- No S04R migration/runtime mutation has been performed on production.
- Next code action after PM docs are committed: implement minimal `internal/runtimecred/{secret.go,policy.go,types.go}` to satisfy the already-observed RED tests, then run full CI.

## Logging rule going forward

Every verified task transition adds one row containing: date, agent/workstream, task ID, exact change, files, tests/CI run, commit, and result. Failed attempts that reveal a real defect also get logged; do not erase them from history.
