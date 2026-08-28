# PVNaive — Canonical Project Status

Last updated: 2026-08-28

> This file is the canonical **development** status for the active branch. Production truth must also be cross-checked against `main:CONTINUE_HERE.md`, `main:ops/S04_LIVE_STATE.md`, and production evidence before any live mutation.

## Project

PVNaive is PVNETWORK's standalone-first management plane for standard NaiveProxy, designed to grow into user lifecycle, accounting, subscriptions, operations and later optional fleet management without coupling the data plane to the UI.

## Repository state

- Repository: `DashSaman/PV-NativePanel`
- Default branch: `main`
- Active implementation branch: `s04-auth`
- Draft PR: `#2` — `S04-AUTH: production authentication foundation`
- `main` head audited: `b98ffbfbe8b30ddcc1bca06531650739a3647d22`
- `s04-auth` head at audit start: `cf0fbf3fea5052e7fbbcc9971a394c2c06d28713`
- Merge base: `d0398cd1b8c1098a21560d4ddd1ff9cfff48b69b`
- Branch relationship at audit: **diverged; s04-auth 123 ahead / 37 behind main**
- PR state at audit: open, draft, mergeable; no submitted reviews/comments.
- GitHub issues at audit: none.
- TODO/FIXME/HACK code search at audit: no indexed matches returned.

## Production state

Evidence on `main` proves:

- S00-S03 passed.
- PostgreSQL 18 is loopback-only and S04 production schema is v2.
- API is healthy on `127.0.0.1:8080` only.
- one real Owner exists and real login/session/CSRF logout/revocation passed.
- public panel/API exposure passed at `https://namir.softarg.ir/panel/`.
- Caddy was changed by reload only; existing Naive forward proxy and camouflage root were preserved.
- current recorded post-exposure Caddyfile SHA-256: `21db739ca3911fb9974eaabe9db07f22e62d84ddc1309ef6bcea3ec247c9ab23`.
- S04 is still formally `IN PROGRESS`; do not promote it solely because the panel is visible.

## Active development state

Owner-approved pre-S05 extension: `S04R-NAIVE-CREDENTIALS`.

Completed in development:

- design spec and detailed TDD implementation plan;
- migration `0003_naive_runtime_credentials` plus PostgreSQL 18 regression coverage;
- compatibility fixes so legacy S04 tests remain pinned to schema v2.

Current TDD task:

- `PVN-020` runtime secret envelope + conservative credential input policy.
- tests exist in `internal/runtimecred/secret_test.go` and `internal/runtimecred/policy_test.go`.
- current CI is intentionally RED because production implementation has not yet been written.
- observed failure: `undefined: ValidateUsername` in the Go job.

## Last verified green development point

- Commit: `6178ab97e3ac328189e4baddbf578ebe79469c3b`
- CI run: `33133931739`
- Go: PASS
- Web: PASS
- PostgreSQL 18/database gates: PASS
- S04 auth rehearsal: PASS
- S04 bundle: PASS

The later test-only head `cf0fbf...` is intentionally RED and must not be called release-ready.

## Numerical progress

Progress is **not an estimate**. It is an equal-weight count over the 72 concrete tasks in `ROADMAP.md`.

| Metric | Count |
|---|---:|
| Total tracked tasks | 72 |
| DONE | 19 |
| IN PROGRESS | 2 |
| BLOCKED | 0 |
| NOT STARTED | 51 |
| Overall completed | **26.4%** |

Formula: `DONE / TOTAL = 19 / 72 = 26.4%`. In-progress tasks receive no partial credit.

### Phase progress by task count

| Phase | Done / Total | Progress |
|---|---:|---:|
| A — Foundation / Database | 6 / 6 | 100% |
| B — Authentication / Public preview | 11 / 11 | 100% of implemented milestones; formal stage closure is separately tracked in Phase D |
| C — S04R Naive credential management | 2 / 12 | 16.7% |
| D — Security hardening / S04 closure | 0 / 7 | 0% |
| E — User / Plan / Reseller lifecycle | 0 / 8 | 0% |
| F — Accounting / full Naive runtime | 0 / 7 | 0% |
| G — Subscription / Notification | 0 / 6 | 0% |
| H — Installer / Operations / Release | 0 / 10 | 0% |
| I — Canonical PM / quality documentation | 0 / 5 | 0% (`PVN-068` currently in progress) |

## Current blockers and risks

1. **P0 branch truth split:** active branch is 37 commits behind production-documentation main. Do not reset or force-push; integrate deliberately after the current feature slice is green.
2. **P0 security:** refresh-token reuse detection is present in SQL but the HTTP refresh path first calls `BeginAuthenticated`, which rejects an already-revoked token before `auth_rotate_session` can mark family reuse. Track as `BUG-001`.
3. **P1 transaction integrity:** authenticated middleware writes the HTTP response before checking transaction commit; commit errors are ignored. Track as `BUG-002`.
4. **P1 readiness:** `/health/ready` checks injected dependencies but does not prove an ongoing DB ping. Track as `BUG-003`.
5. **P1 MFA recovery:** recovery codes are generated/consumed for MFA removal, but the login contract only accepts a TOTP code. Track as `BUG-004` unless product decision explicitly excludes recovery-code login.
6. **P1 auth abuse controls:** DB lockout exists, but the documented HTTP/IP rate-limit/progressive-delay layer is not implemented yet.
7. **Documentation drift:** README, SECURITY, PRODUCT_GAPS, API/feature docs contain scaffold/target statements that no longer match actual code or production.
8. **Release security gap:** CI has tests/build/rehearsal but no SBOM, dependency audit, secret scan, SAST or signed release gate yet.

## Scope discipline

Production R1 remains standalone-first. Multi-node/controller features are valuable but are P3/future until the single-node data model, runtime safety, accounting, subscription, installer, restore and pilot gates are proven.

## Read first

1. `HANDOFF.md`
2. `ROADMAP.md`
3. `KNOWN_ISSUES.md`
4. `AGENT_TASKS.md`
5. `WORKLOG.md`
6. `docs/superpowers/specs/2026-08-28-naive-runtime-credentials-design.md`
7. `docs/superpowers/plans/2026-08-28-naive-runtime-credentials.md`
8. `main:CONTINUE_HERE.md`
9. `main:ops/S04_LIVE_STATE.md`
10. newest `main:ops/evidence/*` before production changes
