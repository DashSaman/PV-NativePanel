# PVNaive — Canonical Roadmap

Last updated: 2026-08-28

This is the canonical task ledger. Task IDs never change; status changes in place. Equal-weight completion is used only for the numerical project snapshot. A task is `DONE` only after implementation, review, tests, integration/security checks, and required docs/handoff updates.

Status vocabulary: `DONE`, `IN_PROGRESS`, `TODO`, `BLOCKED`.
Priority: `P0` blocker/critical, `P1` production-important, `P2` important capability, `P3` improvement/future, `P4` optional.

## Phase A — Foundation / Database

| ID | P | Status | Title | Depends | Done gate |
|---|---|---|---|---|---|
| PVN-001 | P0 | DONE | Product naming and repository safety baseline | — | PVNaive naming fixed; repo not renamed destructively |
| PVN-002 | P0 | DONE | Production preflight | 001 | DNS/TLS/Caddy/listeners/capacity verified read-only |
| PVN-003 | P0 | DONE | Filesystem/service foundation | 002 | service account, directories, permissions, backup base, postflight |
| PVN-004 | P0 | DONE | PostgreSQL 18 schema v1 + RLS | 003 | migration, roles, RLS and privilege boundaries verified |
| PVN-005 | P0 | DONE | Encrypted DB backup/restore/health | 004 | age backup, restore drill, loopback DB health and timer pass |
| PVN-006 | P0 | DONE | S03 production deployment/postflight | 005 | real production + independent postflight pass |

## Phase B — Authentication / protected public preview

| ID | P | Status | Title | Depends | Done gate |
|---|---|---|---|---|---|
| PVN-007 | P0 | DONE | S04 authentication design | 006 | written reviewed design and implementation plan |
| PVN-008 | P0 | DONE | Auth migration v2 | 007 | PostgreSQL 18 migration/down/checksum/RLS tests pass |
| PVN-009 | P0 | DONE | Password/token/TOTP/envelope primitives | 008 | Argon2id, opaque tokens, TOTP, AES-GCM tests pass |
| PVN-010 | P0 | DONE | Auth store + request-context/RLS binding | 008,009 | pre-auth and authenticated transaction tests pass |
| PVN-011 | P0 | DONE | HTTP auth/session/CSRF/MFA surface | 010 | login/refresh/logout/me/sessions/TOTP endpoints exercised |
| PVN-012 | P0 | DONE | Root-only Owner bootstrap | 010 | no default credential; bootstrap tests pass |
| PVN-013 | P0 | DONE | S04 CI/rehearsal/bundle | 008-012 | Go/Web/PG18/E2E auth/bundle green |
| PVN-014 | P0 | DONE | S04 localhost production deployment/recovery | 013 | startup/recovery defects fixed; API localhost healthy |
| PVN-015 | P0 | DONE | S04 localhost independent postflight | 014 | service/listener/schema/backup/Caddy/SSH/firewall verified |
| PVN-016 | P0 | DONE | Real Owner auth lifecycle | 015 | bootstrap + real login/me/CSRF logout/revocation pass |
| PVN-017 | P0 | DONE | Public panel/API exposure | 016 | Caddy validate/reload-only/rollback plan, panel/API/camouflage preservation and visual Owner confirmation |

> Formal `S04=PASSED` remains a separate gate in PVN-036; Phase B records implemented and verified milestones, not the final stage ledger transition.

## Phase C — S04R Naive runtime credential management

| ID | P | Status | Title | Depends | Done gate |
|---|---|---|---|---|---|
| PVN-018 | P0 | DONE | S04R architecture + detailed TDD plan | 017 | approved spec + plan in `docs/superpowers` |
| PVN-019 | P0 | DONE | Migration 0003 runtime credential state | 018 | owner-only FORCE RLS, runtime revision idempotency, rollback to v2, full DB CI green |
| PVN-020 | P0 | DONE | Runtime secret envelope + credential policy | 019 | AES-GCM/tamper/wrong-key/nonce/password/username tests green; no secret-exposing DTO |
| PVN-021 | P0 | DONE | Byte-preserving Caddy `forward_proxy` parser/renderer | 020 | zero/multiple/ambiguous fail closed; only credential span changes; injection tests green |
| PVN-022 | P0 | DONE | Unix-socket Runtime Agent protocol | 021 | AF_UNIX only, strict typed JSON, no arbitrary path/service/command, tests green |
| PVN-023 | P0 | DONE | Privileged validate/apply/rollback operator | 022 | expected-SHA, exact backup, validate-before-write, reload-only, postflight + exact rollback tests |
| PVN-024 | P0 | DONE | Runtime credential store + revision saga/compensation | 020,023 | desired→apply→applied state machine and DB-commit-failure compensation tests |
| PVN-025 | P0 | DONE | Owner-only `/api/v1/runtime/naive` API | 024 | RBAC/CSRF/idempotency/revision/secret-redaction tests |
| PVN-026 | P1 | DONE | `/runtime/naive` UI | 025 | metadata/actions/one-time generated secret/destructive confirmation/mobile client build; copy-ready Naive URI |
| PVN-027 | P0 | DONE | Full S04R disposable rehearsal | 021-026 | real binaries + PG18 + Unix socket + fake/safe Caddy lifecycle; multiple `basic_auth` syntax proven against pinned Caddy |
| PVN-028 | P0 | IN_PROGRESS | Read-only live Naive import preflight | 027 | exact live Caddy/custom binary syntax/import equivalence verified; no mutation |
| PVN-029 | P0 | TODO | Guarded production import + browser mutations postflight | 028 | current credential preserved, apply/rollback gates, new/rotate/disable/revoke behavior verified safely |

## Phase D — Security hardening / formal S04 closure

| ID | P | Status | Title | Depends | Done gate |
|---|---|---|---|---|---|
| PVN-030 | P0 | TODO | Fix refresh-token reuse-family detection | 011 | regression proves reused revoked token reaches family revocation and returns fail-closed |
| PVN-031 | P0 | TODO | Transaction commit-before-success integrity | 011 | mutation cannot emit success before durable DB commit; injected commit failure test |
| PVN-032 | P1 | TODO | DB-backed readiness | 011 | readiness proves current DB dependency/schema, with timeout/failure tests |
| PVN-033 | P1 | TODO | Recovery-code login decision + implementation | 011 | explicit product decision; if enabled, one-time recovery login/replay/session tests |
| PVN-034 | P0 | TODO | HTTP/IP auth rate limit + progressive delay | 011 | bounded login abuse per identity/IP, reverse-proxy-aware trust boundary, tests |
| PVN-035 | P1 | TODO | Public CSP/security-header/cookie exposure review | 017,030-034 | browser/API headers, cookie scope, cache controls and route exposure reviewed/tested |
| PVN-036 | P0 | TODO | Formal independent external S04 postflight and stage closure | 029-035 | independent production evidence; only then S04 ledger may become PASSED |

## Phase E — User / Plan / Reseller lifecycle

| ID | P | Status | Title | Depends | Done gate |
|---|---|---|---|---|---|
| PVN-037 | P0 | TODO | User CRUD + lifecycle state machine | 036 | create/edit/suspend/resume/expire/revoke with tenant isolation and audit |
| PVN-038 | P1 | TODO | Plans + quota policy lifecycle | 037 | plan/quota/reset/duration/concurrency/device capability-aware CRUD |
| PVN-039 | P1 | TODO | Reseller + credit ledger + purchase/renewal | 037,038 | append-only credit, idempotent purchase/renewal and authorization tests |
| PVN-040 | P0 | TODO | User-bound Naive credential lifecycle | 037,045 | commercial credentials map to real runtime without fake rows or secret leaks |
| PVN-041 | P1 | TODO | Bulk operations with dry-run | 037-040 | affected-count/conflict/rollback preview then idempotent apply |
| PVN-042 | P1 | TODO | Search/filter/sort/pagination + computed status | 037 | active/disabled/expired/depleted/on-hold style dimensions remain explainable |
| PVN-043 | P0 | TODO | User/reseller API authorization matrix | 037-042 | owner/admin/reseller/operator/auditor tenant/IDOR tests pass |
| PVN-044 | P1 | TODO | Users/plans/resellers production UI | 037-043 | responsive CRUD/bulk/status/error UX with no fabricated capabilities |

## Phase F — Accounting / full Naive runtime

| ID | P | Status | Title | Depends | Done gate |
|---|---|---|---|---|---|
| PVN-045 | P0 | TODO | Exact per-credential accounting feasibility PoC | 029 | measured runtime source, accuracy/error bounds documented; no estimated billing claim |
| PVN-046 | P0 | TODO | Restart/reconnect/double-count protection PoC | 045 | boot/sequence/reset handling proves no double count across restart/reconnect |
| PVN-047 | P0 | TODO | H2 multiplex + concurrent credential behavior PoC | 045 | client/runtime behavior measured and accounting implications recorded |
| PVN-048 | P0 | TODO | Usage delta collector + append-only ledger/reconciliation | 045-047 | idempotent deltas and rebuildable aggregates with reconciliation tests |
| PVN-049 | P0 | TODO | Quota/reset enforcement | 048 | local enforcement, calendar/interval reset, race/failure tests |
| PVN-050 | P1 | TODO | Concurrency/device limit capability proof | 047 | implement only proven controls; otherwise capability flags remain false |
| PVN-051 | P1 | TODO | Full Naive Runtime adapter/status/revisions integration | 023,048-050 | adapter reports honest capabilities, health, last-good revision and usage state |

## Phase G — Subscription / Notification

| ID | P | Status | Title | Depends | Done gate |
|---|---|---|---|---|---|
| PVN-052 | P0 | TODO | Subscription token lifecycle + renderer | 040,051 | hashed token, revoke/expire, correct Naive config rendering |
| PVN-053 | P1 | TODO | Subscription info/usage page/API | 048,052 | purchase/expiry/remaining/usage data comes from verified state |
| PVN-054 | P0 | TODO | Client compatibility matrix | 052 | Windows/Android Karing/iOS/macOS representative clients tested and recorded |
| PVN-055 | P2 | TODO | QR/templates/update compatibility | 052-054 | useful formats/templates validated; no unsupported format claims |
| PVN-056 | P1 | TODO | Notification rules/outbox/delivery | 048,053 | volume/expiry events, idempotency/retry/audit tests |
| PVN-057 | P2 | TODO | Telegram/operational alert channel | 056 | optional channel cannot block data plane/bootstrap; retry and secret hygiene verified |

## Phase H — Installer / Operations / Release

| ID | P | Status | Title | Depends | Done gate |
|---|---|---|---|---|---|
| PVN-058 | P0 | TODO | Fresh standalone installer | 036,051-056 | non-destructive preflight, pinned artifacts/checksums, one-command fresh install + smoke |
| PVN-059 | P0 | TODO | Upgrade/migration workflow | 058 | pre-backup, idempotent versioned upgrade, migration validation and rollback path |
| PVN-060 | P1 | TODO | Conservative rollback/uninstall | 058,059 | preserve user data by default, explicit destructive confirmation, restore path |
| PVN-061 | P1 | TODO | OS/dependency/systemd support matrix | 058 | supported Ubuntu versions, unit hardening, dependency/version gates tested |
| PVN-062 | P0 | TODO | Scheduled DB/config backup + disaster restore | 051,058 | encrypted scheduled backups, config backup, isolated restore drill and runbook |
| PVN-063 | P1 | TODO | Logs/rotation/metrics/diagnostics/support bundle | 051 | redacted logs/metrics, request IDs, diagnostics bundle with secret preview/expiry |
| PVN-064 | P1 | TODO | Certificate/domain rotation | 051,058 | validate/stage/reload/rollback with data-plane and panel preservation |
| PVN-065 | P0 | TODO | Supply-chain/release security gates | 058 | dependency audit, SAST, secret scan, SBOM, pinned actions/artifacts, release signing/provenance |
| PVN-066 | P0 | TODO | Pilot load/benchmark and capacity gate | 045-065 | staged tests through target concurrency/throughput with CPU/RAM/reconnect/accounting error evidence |
| PVN-067 | P0 | TODO | Release Candidate → final Production release | 066 | clean install/upgrade/restore/security/integration/pilot gates all green; no P0/P1 known release blockers |

## Phase I — Project governance / quality

| ID | P | Status | Title | Depends | Done gate |
|---|---|---|---|---|---|
| PVN-068 | P0 | IN_PROGRESS | Canonical repository audit + PM documents | — | PROJECT_STATUS/ROADMAP/FEATURE_MATRIX/AGENT_TASKS/WORKLOG/HANDOFF/KNOWN_ISSUES + AGENTS links committed and internally consistent |
| PVN-069 | P1 | TODO | Documentation truth reconciliation | 068 | README/SECURITY/API/product gaps/old roadmap clearly match actual vs planned behavior |
| PVN-070 | P0 | TODO | Reconcile `s04-auth` ↔ `main` divergence | 020,068 | deliberate non-force integration; production evidence retained; CI green; PR diff reviewed |
| PVN-071 | P1 | TODO | Full API/web E2E + authorization/fuzz quality gates | 036,043,052 | route readiness, IDOR/RBAC, strict JSON, parser/fuzz/failure-path coverage in CI |
| PVN-072 | P1 | TODO | Project license + third-party attribution policy | 068 | explicit license decision; NOTICE/attribution; no incompatible copied code |

## Progress snapshot

- Total: 72
- DONE: 27
- IN_PROGRESS: 2 (`PVN-028`, `PVN-068`)
- BLOCKED: 0
- TODO: 43
- Completion: **37.5%** (`27 / 72`)

## Dependency-critical execution order from current state

The S04R implementation and disposable rehearsal chain is complete through `PVN-027`. The next production-safe sequence is:

`PVN-028 live read-only preflight → PVN-029 guarded pilot rollout/postflight → PVN-030/031/032/033/034/035 → PVN-036 formal S04 closure → PVN-037...`

For the immediate Pilot, use `docs/PILOT_INSTALL_FA.md`. Customers receive only their Naive credential/link; they do not receive Owner panel credentials.

`PVN-070` branch reconciliation must happen only at a controlled green point; never reset/force-push away production evidence or active feature work.
