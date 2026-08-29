# PVNaive — Agent / Workstream Task Board

Last updated: 2026-08-29

The current Chat environment does **not** expose an independent sub-agent execution tool. The roles below are durable workstreams/ownership labels, not a claim that Claude/Gemini/Qwen/DeepSeek are currently running.

## Mandatory Owner requirements — READ BEFORE CUSTOMER WORK

`OWNER_REQUIREMENTS.md` is now the authoritative Owner product/UX backlog for `/panel/#/customers`, subscription/QR behavior, customer lifecycle, quota/validity, accounting and later operator-parity work.

Every agent touching customer code must read it first. In particular:

- a backend endpoint does not count as complete if the Owner cannot perform the action clearly from the UI;
- add a clear customer `Edit` action and customer details/action surface;
- `View QR`, `Copy Subscription`, `Copy Direct Link` and `View Details` must be read-only;
- viewing QR must not require reissuing the Subscription;
- viewing/copying existing Subscription must not rotate token or password;
- `Reissue Subscription` and `Rotate Password` are two separate explicit actions;
- customer `Delete/Revoke` must be exposed safely with confirmation/audit;
- complete enable/disable/suspend/revoke/delete, add/set volume, extend/change expiry and unlimited controls;
- do not fabricate usage/remaining/online/device/speed state before Runtime proof;
- after P0 lifecycle/accounting work, continue Sanaei/PasarGuard-class search/filter/sort/bulk/renewal ergonomics and then OV-PvNetwork-derived monitoring/fleet patterns.

If any older handoff text conflicts with `OWNER_REQUIREMENTS.md` about customer UX, the Owner Requirements file wins unless the Owner explicitly changes it later.

## Mandatory report format

Before work:

```text
AGENT: <role>
TASK-ID: <PVN-ID>
GOAL: <one outcome>
FILES: <allowed scope>
DEPENDENCIES: <IDs/evidence>
```

After a work unit:

```text
STATUS: RED | GREEN | BLOCKED | REVIEW
CHANGES: <what changed>
FILES MODIFIED: <paths>
TESTS: <commands/CI run>
RESULT: <evidence, not just Done>
NEXT STEP: <exact next action>
BLOCKERS: <none or evidence>
```

No workstream may mark a task `DONE`; final verification/ledger update belongs to Agent-REVIEW / Lead Engineer.

## Current state

`PVN-020` through `PVN-027` are complete in development/rehearsal. The active code-to-production edge is now `PVN-028`.

Full S04R checkpoint:

- commit `a41fd84c2f17076a3b190eafad3539c47b430503`;
- CI `33190295766` — Go/Web/PG18/pinned-Caddy proof/full S04R rehearsal/bundle all PASS;
- customer-handoff UI then added a TDD-tested `naive+https://...` builder and one-click Karing/Naive copy action.

## Active assignments

| Workstream | Task | Status | Scope | Handoff condition |
|---|---|---|---|---|
| Agent-REVIEW/DEVOPS | PVN-028 | ACTIVE | live read-only preflight on existing `testAmir5-3` | `PREFLIGHT_RESULT=PASS`; exact Caddy SHA captured; zero mutation |
| Agent-REVIEW/DEVOPS | PVN-029 | WAITING_ON_028 | guarded S04R upgrade, secure live import, one customer credential + Karing smoke | current credential preserved; Caddy invariants; new credential works; evidence committed |
| Agent-PM / Lead | PVN-068 | IN_PROGRESS | sync canonical PM/handoff files to S04R Pilot truth | PM files internally consistent + final HEAD CI checked |
| Agent-SEC | PVN-030 | READY_AFTER_PILOT | refresh-token reuse-family regression/fix | RED→GREEN + auth rehearsal |
| Agent-SEC | PVN-031 | READY_AFTER_PILOT | commit-before-success HTTP integrity | injected commit failure cannot emit success |
| Agent-BACKEND/OPS | PVN-032 | READY_AFTER_PILOT | DB-backed readiness | bounded DB/schema probe + failure tests |
| Agent-PM/SEC | PVN-033 | READY_AFTER_PILOT | recovery-code login product decision | explicit decision + tests/docs |
| Agent-SEC | PVN-034 | READY_AFTER_PILOT | IP/identity auth abuse controls | trusted proxy boundary + tested limits/delay |
| Agent-SEC/REVIEW | PVN-035/036 | WAITING | public security review + formal S04 closure | independent evidence after auth hardening |
| Agent-DOCS | PVN-069 | READY | reconcile legacy README/SECURITY/API/product docs | actual vs planned clearly labeled |
| Agent-ARCH/REVIEW | PVN-070 | WAIT_GREEN_CHECKPOINT | branch divergence integration | no force/reset; reviewed non-destructive integration + CI |
| Agent-QA/SEC | PVN-071 | FUTURE | authorization/E2E/fuzz gates | CI enforced |
| Agent-PM/OWNER | PVN-072 | NEEDS_OWNER_BUSINESS_DECISION | license policy | explicit license + NOTICE strategy |

## Completed S04R workstreams

| Task | Verified outcome |
|---|---|
| PVN-020 | runtime secret envelope + input policy |
| PVN-021 | byte-preserving Caddy parser/renderer |
| PVN-022 | fixed Unix-socket Runtime Agent protocol |
| PVN-023 | expected-SHA validate/backup/reload-only/rollback operator |
| PVN-024 | runtime DB store + revision saga + compensation/reconciliation error |
| PVN-025 | Owner-only runtime API + CSRF/idempotency/revision/secret boundaries |
| PVN-026 | runtime UI + one-time generated secret + copy-ready customer Naive URI |
| PVN-027 | full disposable S04R rehearsal + exact pinned Caddy multiple-basic-auth proof |

## File conflict rules

- `OWNER_REQUIREMENTS.md`: Owner-authoritative customer product behavior. Agents may update status/evidence after verified work but must not weaken/remove requirements without explicit Owner direction.
- `internal/runtimecred/*`: next changes only for later runtime/accounting work; do not reopen completed S04R behavior casually.
- `internal/runtimeagent/*`: production safety boundary; any change requires failure-path tests and full rehearsal.
- `internal/httpapi/*`: auth-hardening tasks `PVN-030/031/034` must be coordinated; do not parallel-write the same middleware/handler files.
- `web/src/runtime*` / `RuntimeNaive.tsx`: Pilot behavior is frozen until live smoke unless a live-blocking defect is proven.
- `db/migrations/*`: append-only after release; checksum changes require review.
- canonical PM files may be updated after verified transitions, but sequential writes to the same file are required.

## Production interaction rule

On `testAmir5-3`, use **one server step at a time** and preserve full output. Read-only preflight comes first. Never ask for or paste auth/runtime/age keys or raw secret-bearing Caddy content.

Runbook: `docs/PILOT_INSTALL_FA.md`.

The live sequence is:

`artifact checksum → PVN-028 read-only preflight → exact Caddy SHA lock → PVN-029 guarded S04R upgrade → Owner secure import → one generated customer credential → Karing smoke → evidence commit`

Do not hand the customer Owner panel credentials. The Pilot handoff is the generated Naive link only.

## After the Pilot

Immediately return to the critical security/closure chain:

`PVN-030 + PVN-031 + PVN-032 + PVN-033 + PVN-034 → PVN-035 → PVN-036`

Then execute the customer/product backlog according to `OWNER_REQUIREMENTS.md`, prioritizing P0 customer UX/lifecycle, exact accounting/hard quota, trusted first-use, and production S05 rollout before P1/P2 feature expansion.
