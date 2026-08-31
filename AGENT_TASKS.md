# PVNaive — Agent / Workstream Task Board

Last updated: 2026-08-31

`OWNER_REQUIREMENTS.md` plus the Owner's Production/parity master prompt define product behavior and execution order. This board records current ownership and prevents duplicate work.

## Shared rules

- start from latest `main`, never from a chat snapshot;
- preserve already-working Runtime/accounting/customer/subscription behavior;
- use isolated branches/PRs; never force-push/reset main;
- TDD/fail-first where practical;
- route/schema presence is not implementation evidence;
- no fake usage/online/HWID/speed;
- no secret/password/token/key in Git/chat/CI/evidence;
- Production mutation requires DB/config/Caddy/web/binary backups + rollback;
- GPL/AGPL competitors are behavior/architecture references unless explicit source-license compatibility is approved;
- update a unique workstream report during parallel work; Lead owns canonical reconciliation.

## Current Lead lane

| Field | Value |
|---|---|
| Agent | Lead Engineer / PM / Integration |
| Branch | `lead/parity-truth-2026-08-30` |
| PR | `#27` |
| Start main | `a021aa4b62c35b775fb521d042b2f8e6dbde10b0` |
| Plan | `docs/superpowers/plans/2026-08-30-production-parity-reconciliation.md` |
| Goal | Current repo/Production truth → current competitor parity → canonical docs → exact-head CI → safe PR #16 integration plan |
| Production mutation | none in this lane |

## Work already integrated — do not recreate

### WS1 Runtime / Accounting

Integrated main + Production contains:

- pinned forwardproxy exact successful-write accounting;
- Runtime UUID identity;
- dedicated accounting socket + Telemetry Agent;
- append-only/idempotent direct accounting;
- boot/session/sequence/cumulative semantics;
- ServiceTerm usage isolation;
- trusted first-successful-CONNECT activation producer;
- session/presence projection;
- shared finite-quota reservation/settlement core.

Remaining WS1-style work is acceptance/completion, not a rewrite: legacy baseline truth, read-model wiring, reset semantics, hard-quota/first-CONNECT controlled Production proof, sessions/limits.

### WS2 Customer Product

Integrated main contains:

- customer create/adopt/edit/suspend/resume/revoke-safe-delete;
- plans, renewal/new ServiceTerm, Next Plan/On Hold foundations;
- quota/unlimited/add/set volume;
- validity/expiry/no-expiry/extend days;
- groups/tags/notes;
- search/filter/sort/pagination;
- bulk preview/idempotent execution for supported actions;
- tenant/RLS product foundations.

### WS3 Subscription / Client Delivery

Integrated main contains:

- `/sub/<token>` machine endpoint;
- `/s/<token>` human account page;
- local QR;
- read-only current Subscription;
- explicit Subscription reissue;
- explicit password rotation;
- token/password mutation separation.

Real Karing acceptance matrix remains.

## Current ordered task queue

Do not reorder unless a real technical dependency is documented.

| Order | Workstream | Status | Outcome / gate |
|---:|---|---|---|
| 1 | Lead audit latest repo/Production | IN_PROGRESS | repo/PR/CI/prod truth recorded, no secret leak |
| 2 | Lead competitor parity | IN_PROGRESS | 120-feature current matrix at `docs/PANEL_PARITY_MASTER_2026-08-30.md` |
| 3 | Lead canonical docs truth | IN_PROGRESS | FEATURE_MATRIX/PROJECT_STATUS/HANDOFF/ROADMAP/KNOWN_ISSUES/AGENT_TASKS/WORKLOG/CONTINUE_HERE current |
| 4 | WS4 integration / PR #16 | WAITING_ON_1_3 | extract useful ops units onto fresh branch; no blind merge |
| 5 | Accounting legacy baseline | TODO | no fake zero, provable baseline or Unknown |
| 6 | `/s` accounting projection | TODO | Used/Remaining/Upload/Download/Last Online/Online/Expiry/Quota real or explicit unavailable |
| 7 | Manual Reset Usage | TODO | confirm + audit + accounting event + idempotency; no token/password rotation |
| 8 | Bulk Reset Usage | TODO | Preview → Execute with same idempotency key |
| 9 | Periodic reset execution | TODO | persisted cursor/scheduler/timezone/exactly-once/audit/history |
| 10 | Hard quota Production proof | TODO | concurrency/race/exhaustion/reload/restart/reconnect/no bypass |
| 11 | First successful CONNECT proof | TODO | only successful authenticated CONNECT activates |
| 12 | Customer sessions | DONE | trusted peer IP + active timestamps + exact per-session bytes; PR #40 merged, release blockers #41/#43 fixed, Production rollout completed at schema17 |
| 13 | Kill session | IN_PROGRESS | real disconnect + confirmation/audit; next ordered implementation lane |
| 14 | Concurrent session limit | TODO | Unlimited/N enforced under race |
| 15 | Unique IP limit | TODO | simultaneous unique-IP semantics/enforcement |
| 16 | IP/session history | TODO | bounded privacy-aware retention |
| 17 | HWID PoC | TODO | implement only if identity is trustworthy |
| 18 | Speed-limit PoC | TODO | implement only if data plane can enforce |
| 19 | Reseller CRUD | TODO | create/edit/disable/revoke/list/search |
| 20 | Tenant isolation full audit | TODO | cross-reseller read/edit/renew/delete/subscription IDOR negatives |
| 21 | Reseller wallet/credit | TODO | correct balance mutation semantics |
| 22 | Financial ledger | TODO | immutable credit/debit/create/renew/refund/adjustment history |
| 23 | Reseller plan/restriction quotas | TODO | allowed plans/max users/max active/credit |
| 24 | Customer history | TODO | all sensitive service actions projected |
| 25 | Audit Explorer | TODO | actor/user/action/date/IP/result filters; redacted |
| 26 | Notification engine/preferences/history | TODO | quota/expiry/runtime/node/backup/security/disk/cert events |
| 27 | Telegram + rules | TODO | secure secret handling + retry/outbox/rule builder |
| 28 | Dashboard/monitoring/history | TODO | CPU/RAM/disk/network/traffic/online/runtime real data |
| 29 | Logs/request diagnostics/support bundle | TODO | application/runtime/security, request ID, redaction |
| 30 | Doctor | TODO | command/page health checks |
| 31 | Scheduled encrypted backup | TODO | retention/DB/config/runtime state/important files |
| 32 | Restore/verification/drill/UI | TODO | strong safeguards + recurring restore proof |
| 33 | API/OpenAPI | TODO | actual ready contracts only |
| 34 | API rate limit/idempotency/webhooks | TODO | stable policy/event contracts |
| 35 | Security BUG-001/002/003 | TODO | refresh reuse, commit-before-success, DB readiness |
| 36 | Authorization/IDOR/CSRF/redaction/fuzz | TODO | complete route × role matrix |
| 37 | Supply-chain security | TODO | dependency scan, secret scan, SAST, SBOM, signing/provenance/license |
| 38 | Multi-node/fleet | TODO | node model/auth/health/metrics/capacity/assignment/deploy |
| 39 | Drain/maintenance/canary/upgrade/rollback/failover/smart node | TODO | desired/applied + reconciliation loops |
| 40 | Fresh installer | TODO | secure version-pinned clean Ubuntu install + Doctor |
| 41 | Upgrade | TODO | pre-backup + safe migrations |
| 42 | Rollback/uninstall | TODO | conservative data-preserving lifecycle |
| 43 | Client compatibility | TODO | Karing Windows/Android/iOS/macOS/Linux first |
| 44 | Load/capacity | TODO | 50/100/200/400+ + correctness under load |
| 45 | Bulk/search completion | TODO | reset/delete/unlimited/no-expiry/change-expiry/next-plan, advanced filters/sorts/columns/URL state |
| 46 | Final UI polish | TODO | professional customer/dashboard/account page + responsive/accessibility/theme |
| 47 | Final docs reconciliation | TODO | no stale TODO for implemented features |
| 48 | Clean-server installation | TODO | fresh VM proof |
| 49 | Final Production smoke | TODO | exact RC deploy smoke + rollback readiness |
| 50 | Release Candidate | TODO | no logical P0/P1 release blocker |

## Old PR classification

| PR | Classification | Action |
|---:|---|---|
| #4 | STILL USEFUL, small extract | re-evaluate only explicit Karing sing-box profile/export during client lane |
| #5 | SUPERSEDED / MERGED ELSEWHERE | do not merge; newer customer lifecycle exists on main |
| #6 | SUPERSEDED / MERGED ELSEWHERE | do not merge; newer customer/product/subscription flow exists on main |
| #8 | SUPERSEDED / ARCHIVE | do not merge; replaced by integrated WS1 accounting |
| #16 | STILL USEFUL, manual extraction required | fresh integration branch; inspect candidate units one-by-one |

## File ownership / conflict rules for next integration

During PR #16 extraction:

- preserve current `internal/httpapi/server.go` customer/subscription/accounting changes;
- preserve current schema 11; append migrations only if required;
- metrics/observability/ops should avoid modifying customer semantics;
- notification/fleet foundations are not automatically product-complete;
- systemd backup/restore units require safe path/permission/retention review before Production;
- web System Dashboard/error boundary must not regress current customer/product UI;
- no Production deployment until the extracted branch has full green CI and required backups/rollback plan.

## Mandatory work-unit report

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

Lead updates canonical documents only after evidence, then automatically proceeds to the next ordered task.
