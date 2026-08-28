# PVNaive — Handoff

Last updated: 2026-08-28

This is the canonical development handoff on branch `s04-auth`. Production truth must be cross-checked against the newer files on `main` before any server mutation.

## PROJECT

PVNaive is a standalone-first PVNETWORK management plane for standard NaiveProxy. R1 targets one external server with secure management, real user lifecycle, exact/proven accounting, subscription delivery, safe runtime operations, backup/restore and installer/release lifecycle. Multi-node/controller is future scope and must not become an R1 dependency.

## REPOSITORY

- Repo: `DashSaman/PV-NativePanel`
- Default: `main`
- Active dev branch: `s04-auth`
- Draft PR: #2 `S04-AUTH: production authentication foundation`
- Audit relationship: `s04-auth` 123 commits ahead / 37 behind `main`; merge base `d0398cd1b8c1098a21560d4ddd1ff9cfff48b69b`.
- Do not reset, force-push or blindly overwrite either side. `PVN-070` owns reconciliation at a green checkpoint.

## CURRENT PRODUCTION STATE

Authoritative production docs on `main`:

1. `CONTINUE_HERE.md`
2. `ops/S04_LIVE_STATE.md`
3. newest `ops/evidence/*`

Verified live facts:

- `S00-S03=PASSED`.
- `S04-AUTH=IN PROGRESS` formally.
- PostgreSQL 18 production schema remains v2.
- API healthy on `127.0.0.1:8080` only.
- one real Owner exists.
- real Owner login/session/me/CSRF logout/revocation passed.
- public panel/API exposure passed and Owner visually confirmed authenticated dashboard.
- current recorded Caddy SHA: `21db739ca3911fb9974eaabe9db07f22e62d84ddc1309ef6bcea3ec247c9ab23`.
- public panel: `https://namir.softarg.ir/panel/`.
- Caddy exposure used reload only; MainPID/restarts, existing Naive `forward_proxy`, camouflage, SSH and firewall were preserved.

Do not use older `AGENT_HANDOFF.md` or `ops/DEPLOYMENT_PROGRESS.md` alone as production truth; they are historically valuable but currently lag the later main evidence.

## COMPLETED

- PVN-001..006: foundation, DB, encrypted backup/restore, S03 production/postflight.
- PVN-007..017: S04 auth architecture, schema, auth primitives/store/API/MFA, bootstrap, CI/bundle, localhost production, real Owner lifecycle and public panel/API exposure.
- PVN-018: Owner-approved S04R Naive credential-management spec + detailed plan.
- PVN-019: migration 0003 + owner-only runtime credential state and PG18 regression tests.

See `ROADMAP.md` for the permanent ledger and `WORKLOG.md` for evidence history.

## CURRENT TASK

### PVN-020 — Runtime secret envelope and credential policy

Status: `IN_PROGRESS`, intentionally RED under TDD.

Existing tests:

- `internal/runtimecred/secret_test.go`
- `internal/runtimecred/policy_test.go`

Observed CI RED:

- run `33134072689`
- Go fails at `internal/runtimecred/policy_test.go:15:13: undefined: ValidateUsername`
- DB and Web pass.

Implement **only the minimal production code required by the tests**:

- `internal/runtimecred/secret.go`
- `internal/runtimecred/policy.go`
- `internal/runtimecred/types.go`

Required interfaces:

```go
EncryptSecret(key, plaintext []byte) (ciphertext, nonce []byte, err error)
DecryptSecret(key, nonce, ciphertext []byte) ([]byte, error)
HashSecret([]byte) [32]byte
GeneratePassword() (string, error)
ValidateUsername(string) error
ValidatePassword(string, imported bool) error
```

Implementation rules:

- stdlib only: AES-256-GCM, crypto/rand, SHA-256, base64.RawURLEncoding;
- exact 32-byte key, random 12-byte nonce;
- generator = 24 CSPRNG bytes → base64url no padding (32 chars);
- username: 1–64 bytes, ASCII alnum + `._@+-` only;
- new password: 14–128 bytes and conservative renderer-safe visible ASCII;
- imported current credential may be shorter than 14 so handoff does not silently rotate live secret, but newline/CR/NUL/control chars remain rejected;
- no JSON/browser DTO may expose plaintext/ciphertext/nonce by accident.

After implementation run fresh CI. Do not call PVN-020 DONE until full relevant CI is green and PM docs are updated.

## NEXT TASKS

1. PVN-021: failing tests first for byte-preserving Caddy parser/renderer.
2. PVN-022: Unix-socket Runtime Agent typed protocol.
3. PVN-023: privileged Caddy validate/backup/reload-only/postflight/rollback operator.
4. PVN-024: runtime store + desired/apply/applied saga and compensation.
5. PVN-025: Owner-only runtime Naive API.
6. PVN-026: runtime UI.
7. PVN-027: disposable end-to-end S04R rehearsal, including exact multiple-basic-auth syntax proof.
8. PVN-028: read-only live import preflight.
9. PVN-029: guarded import/mutation rollout.
10. PVN-030/031/034: critical auth hardening before formal S04 closure.

## BLOCKERS / KNOWN RISKS

Read `KNOWN_ISSUES.md`. Highest priority:

- BUG-001 refresh reuse detection flow;
- BUG-002 response-before-transaction-commit;
- SECURITY-001 HTTP/IP auth rate limit;
- TECH-DEBT-001 branch divergence;
- TEST-002 exact custom-Caddy multiple credential proof;
- TEST-003 exact Naive accounting proof.

## IMPORTANT DECISIONS

- Standalone-first R1; no controller/fleet dependency.
- PostgreSQL is source of management desired state.
- API is unprivileged.
- Runtime mutations go through a narrow privileged Unix-socket agent.
- no arbitrary shell/path/service/URL operations.
- dedicated `/etc/pvnaive/runtime.key`, separate from auth and age backup keys.
- Caddy transform touches supported credential directives only.
- `caddy validate` → exact backup → install → **reload only** → smoke/postflight → applied DB commit; exact rollback on failure.
- if Caddy succeeds but DB finalization fails, compensate by restoring exact pre-apply Caddy backup; never report success/generated secret in split-brain state.
- deleting a runtime credential means soft revoke.
- last active runtime credential cannot be disabled/revoked.
- no exact quota/accounting/session/device/speed claims until separately proven.
- no secret in Git, audit, logs, evidence or GET/list API responses.

## FILES TO READ FIRST

1. `PROJECT_STATUS.md`
2. `ROADMAP.md`
3. `KNOWN_ISSUES.md`
4. `AGENT_TASKS.md`
5. `WORKLOG.md`
6. `FEATURE_MATRIX.md`
7. `docs/superpowers/specs/2026-08-28-naive-runtime-credentials-design.md`
8. `docs/superpowers/plans/2026-08-28-naive-runtime-credentials.md`
9. `AGENTS.md`
10. for live work: `main:CONTINUE_HERE.md`, `main:ops/S04_LIVE_STATE.md`, newest main evidence.

## LAST VERIFIED COMMIT

Feature-code checkpoint known fully green:

`6178ab97e3ac328189e4baddbf578ebe79469c3b`

CI run `33133931739`: Go/Web/Database/S04 rehearsal/S04 bundle all PASS.

Later runtimecred test commits intentionally made the branch RED to follow TDD. Documentation-only commits after that do not change the feature-code verification status.

## CONTINUE FROM HERE

Do not ask the Owner to repeat project context. Do not touch production yet.

1. verify the current branch head and current `PVN-020` tests;
2. implement minimal runtimecred secret/policy/types code;
3. obtain fresh GREEN CI;
4. review diff/security; update ROADMAP/WORKLOG/PROJECT_STATUS/HANDOFF;
5. immediately begin PVN-021 with a failing parser/renderer test;
6. continue dependency chain autonomously;
7. only when PVN-027 rehearsal passes, prepare one read-only server command for PVN-028 and wait for its full output before the next server action.
