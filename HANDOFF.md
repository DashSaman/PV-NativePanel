# PVNaive — Handoff

Last updated: 2026-08-28

This is the canonical development handoff on branch `s04-auth`. Production truth must still be cross-checked against the newer production evidence on `main` before any live mutation.

## PROJECT

PVNaive is a standalone-first management plane for standard NaiveProxy. The immediate Pilot is intentionally practical: the Owner uses the panel to manage real Naive credentials and hands a copy-ready `naive+https://...` link to a customer. Customer self-service, quota/accounting, expiry, reseller and subscriptions remain later work.

## REPOSITORY

- Repo: `DashSaman/PV-NativePanel`
- Default: `main`
- Active dev branch: `s04-auth`
- Draft PR: #2 `S04-AUTH: production authentication foundation`
- Branches are deliberately diverged; `PVN-070` owns non-force reconciliation.
- Never reset/force-push or overwrite production evidence.

## CURRENT PRODUCTION STATE BEFORE S04R

Authoritative live references remain:

1. `main:CONTINUE_HERE.md`
2. `main:ops/S04_LIVE_STATE.md`
3. newest `main:ops/evidence/*`

Last canonical facts:

- S00-S03 passed.
- S04 Auth/public panel milestones were deployed and independently checked, but formal S04 closure remains open.
- PostgreSQL production schema is still v2 until S04R is actually rolled out.
- API is loopback-only on `127.0.0.1:8080`.
- one real Owner exists.
- panel is public at `https://namir.softarg.ir/panel/`.
- Caddy/Naive data plane and camouflage must remain preserved.

## COMPLETED DEVELOPMENT

`PVN-001`…`PVN-027` are complete except tasks outside the linear S04R chain; specifically the full S04R implementation chain `PVN-020`…`PVN-027` is now DONE.

S04R provides:

- runtime secret AES-GCM envelope and conservative credential policy;
- byte-preserving Naive Caddy parser/renderer;
- root Runtime Agent on fixed AF_UNIX socket;
- fixed-capability expected-SHA Caddy operator with exact backup, validate, reload-only, postflight and rollback;
- runtime PostgreSQL store + desired/apply/applied revision saga;
- compensation and explicit `runtime_reconciliation_required` double-failure contract;
- Owner-only runtime API with CSRF/idempotency/If-Match/secret redaction;
- `/panel/#/runtime/naive` UI;
- secure live credential import semantics;
- create/rename/rotate/enable/disable/soft-revoke;
- last-active protection;
- one-time generated password;
- copy-ready percent-encoded `naive+https://...` link for Karing/compatible clients;
- full disposable PG18/API/Unix-socket/Runtime rehearsal;
- exact pinned `v2.11.2-naive` Caddy proof for multiple `basic_auth` entries;
- checksum-gated S04R artifact with three binaries, migration 0003, web, systemd units, preflight and guarded upgrade.

## VERIFIED GREEN CHECKPOINT

Full implementation/bundle checkpoint:

- commit: `a41fd84c2f17076a3b190eafad3539c47b430503`
- CI: `33190295766`
- Go PASS
- Web PASS
- PostgreSQL 18/database gates PASS
- exact pinned Caddy proof PASS
- S04 auth rehearsal PASS
- full S04R rehearsal PASS
- bundle contract PASS
- archive checksum PASS
- artifact upload PASS

Artifact from that run was also downloaded and independently checked: outer `.tar.gz.sha256` passed and every internal `SHA256SUMS` entry passed.

Customer-link increment followed TDD:

- RED run `33190665847`: only new `buildNaiveURI` test failed (`not a function`), 17 prior tests passed.
- implementation commits: `f7a8091128e8993f162f0965789e3a90f107125c` and `69259a2e4e8001dd125ad29e7272110864b07fd3`.
- web job on `33190796760`: runtime tests and production build PASS.

A fresh full CI must still be checked on the final documentation HEAD before claiming this handoff release-ready.

## CURRENT TASK

### PVN-028 — Read-only live Naive import preflight

Status: `IN_PROGRESS` because the script is implemented/rehearsed but has not yet been run and evidenced against the live server.

**Do not start with a mutation.**

Use exactly:

`docs/PILOT_INSTALL_FA.md`

Live order:

1. obtain the newest green S04R artifact for the final HEAD;
2. verify outer and internal checksums;
3. run `scripts/stages/S04R-preflight.sh` as root;
4. require `PREFLIGHT_RESULT=PASS`;
5. capture exact `CADDYFILE_SHA256`;
6. only then run the guarded upgrade with that SHA;
7. require schema v3, healthy Runtime Agent/API and unchanged Caddy SHA/MainPID/NRestarts;
8. Owner logs into panel and securely imports current Runtime credential;
9. create one generated customer credential;
10. copy the one-time Karing/Naive link and test it in Karing;
11. verify the old existing credential still works;
12. record production evidence and then close PVN-028/029.

## PILOT CUSTOMER HANDOFF

The customer receives a link like:

```text
naive+https://USERNAME:PASSWORD@namir.softarg.ir:443
```

The actual UI percent-encodes username/password and builds the host from the panel hostname.

**Never give the customer Owner panel email/password/session/MFA information.**

## PILOT LIMITS — DO NOT CLAIM THESE YET

- customer portal/login;
- exact usage/accounting;
- traffic quota;
- commercial expiry/reset;
- device/HWID/session limits;
- speed limit;
- reseller/credit;
- subscription URL/page lifecycle;
- notifications;
- final generic fresh installer.

## AFTER THE PILOT

Do not expand random features first. The next critical safety chain is:

1. `PVN-030` refresh-token reuse-family detection;
2. `PVN-031` transaction commit-before-success;
3. `PVN-032` DB-backed readiness;
4. `PVN-033` recovery-code login decision;
5. `PVN-034` HTTP/IP auth rate limit + progressive delay;
6. `PVN-035` public security headers/cookies/CSP review;
7. `PVN-036` independent formal S04 closure.

Then continue user lifecycle (`PVN-037+`), exact accounting (`PVN-045+`), subscription (`PVN-052+`) and generic installer/release (`PVN-058+`).

## OPEN CRITICAL RISKS

Read `KNOWN_ISSUES.md`. Highest priority after Pilot:

- BUG-001 / PVN-030 refresh-token family reuse;
- BUG-002 / PVN-031 generic commit-before-success integrity;
- SECURITY-001 / PVN-034 HTTP/IP auth abuse controls;
- DEPLOY-002 / PVN-028/029 live S04R evidence;
- TECH-DEBT-001 / PVN-070 branch divergence;
- TEST-003 / PVN-045+ exact accounting proof.

Exact multiple-basic-auth Caddy syntax is no longer an open test gap; it was proven against the pinned Naive Caddy binary in CI.

## IMPORTANT DECISIONS

- standalone-first R1;
- PostgreSQL stores management desired state;
- API remains unprivileged;
- privileged runtime changes only through narrow Unix-socket agent;
- no arbitrary shell/path/service/URL control;
- dedicated `/etc/pvnaive/runtime.key` separate from auth/backup keys;
- Caddy changes touch only supported credential directives;
- expected SHA + exact backup + validate + reload-only + postflight + rollback;
- DB finalization failure after Runtime apply must compensate; if compensation itself fails, return reconciliation-required, never success;
- delete means soft revoke;
- last active credential cannot be disabled/revoked;
- no secret in Git/log/evidence/list API;
- no quota/accounting/session/device/speed promises until proven.

## READ FIRST

1. `PROJECT_STATUS.md`
2. `ROADMAP.md`
3. `docs/PILOT_INSTALL_FA.md`
4. `KNOWN_ISSUES.md`
5. `AGENT_TASKS.md`
6. `WORKLOG.md`
7. `FEATURE_MATRIX.md`
8. `docs/superpowers/specs/2026-08-28-naive-runtime-credentials-design.md`
9. `docs/superpowers/plans/2026-08-28-naive-runtime-credentials.md`
10. before live work: `main:CONTINUE_HERE.md`, `main:ops/S04_LIVE_STATE.md`, newest main evidence.

## CONTINUE FROM HERE

Do not ask the Owner to repeat context.

1. verify current `s04-auth` HEAD and final CI;
2. if any job is red, fix it before rollout;
3. use the artifact from that exact green HEAD, not an older bundle;
4. perform only the read-only `PVN-028` preflight first;
5. wait for and inspect its complete output before running `PVN-029` mutation;
6. after Karing live smoke, commit evidence before continuing development.
