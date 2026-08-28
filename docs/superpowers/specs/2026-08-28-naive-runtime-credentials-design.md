# Naive Runtime Credential Management Design

Date: 2026-08-28
Stage placement: `S04R-NAIVE-CREDENTIALS` — owner-prioritized pre-S05 extension
Status: design approved in chat; self-reviewed written spec pending final user review

## Goal

Add the first real NaiveProxy runtime-management surface to PVNaive so an authenticated Owner can see the currently deployed Naive forward-proxy account, import it safely into PVNaive state, add more username/password credentials, rename usernames, rotate passwords, enable/disable/revoke credentials, and apply those changes to the live Caddy `forward_proxy` block with validation, reload-only deployment and automatic rollback.

This extension is intentionally narrow. It exists because the Owner explicitly prioritized seeing and editing the already-running NaiveProxy credentials before the full S05 user lifecycle. It does **not** mark S06-RUNTIME complete and must not fake user, quota, accounting, session or subscription features that remain later-stage work.

## Current production baseline

The live target is `testAmir5-3`, domain `namir.softarg.ir`.

Already verified before this design:

- PostgreSQL 18 schema version is `2`.
- one real active Owner exists;
- real Owner login, `/me`, CSRF logout and session revocation passed;
- API is healthy on `127.0.0.1:8080` only;
- the management preview is publicly reachable at `https://namir.softarg.ir/panel/`;
- public API proxying under `/api/` works;
- the existing camouflage root response remained byte-identical during panel exposure;
- Caddy was reloaded, not restarted; PID and restart count stayed unchanged;
- the existing Naive `forward_proxy` block remained present;
- the current post-exposure Caddyfile SHA-256 is `21db739ca3911fb9974eaabe9db07f22e62d84ddc1309ef6bcea3ec247c9ab23`;
- the exact pre-panel Caddyfile is preserved at `/var/backups/pvnaive/caddy/20260828T010546Z/Caddyfile`.

The live Caddy site contains a single `forward_proxy` block with `basic_auth`, `hide_ip`, `hide_via`, and `probe_resistance`, plus the already-added path-specific PVNaive panel/API route and the existing camouflage file server.

## Stage semantics

`S04R-NAIVE-CREDENTIALS` is a limited owner-requested bridge between S04 and the later Runtime stage.

It may implement:

- Naive runtime inspection;
- Naive credential import and management;
- credential-secret encryption;
- validated Caddy credential-set rendering;
- privileged Caddy apply/rollback;
- Runtime revision/audit records required for those operations;
- `/runtime/naive` UI.

It does not implement:

- reseller or user CRUD;
- quota or byte accounting;
- connection/session accounting;
- subscription renderer;
- plan/purchase lifecycle;
- multi-node/fleet management;
- arbitrary Caddy editing;
- arbitrary shell execution;
- arbitrary protocol plugins.

Therefore S05/S06 remain distinct gates. Completing this extension does not allow `S06-RUNTIME=PASSED` to be recorded.

## Architecture decision

Use **PostgreSQL desired state + a narrow privileged local Runtime Agent**.

Do not let the existing unprivileged API edit `/etc/caddy/Caddyfile` directly. Do not make the Caddy Admin API the persistent source of truth.

Flow:

```text
Browser Owner
  -> PVNaive API (pvnaive, unprivileged)
  -> PostgreSQL desired/runtime state
  -> Unix socket /run/pvnaive/runtime-agent.sock
  -> pvnaive-runtime-agent (narrow privileged service)
  -> exact Caddyfile transform
  -> caddy validate
  -> exact backup
  -> caddy reload
  -> smoke/verify
  -> commit applied revision OR exact rollback + reload
```

The API never receives a general-purpose root capability. The agent exposes only typed operations needed by the Naive adapter.

## Dedicated Runtime Agent

Create executable `pvnaive-runtime-agent` and systemd unit `pvnaive-runtime-agent.service`.

Security contract:

- local Unix socket only: `/run/pvnaive/runtime-agent.sock`;
- no TCP listener;
- socket group `pvnaive`, mode `0660`;
- fixed Caddyfile path: `/etc/caddy/Caddyfile`;
- fixed Caddy binary: `/usr/local/bin/caddy`;
- fixed service: `caddy-naive.service`;
- fixed backup root: `/var/backups/pvnaive/caddy`;
- no caller-supplied executable path;
- no caller-supplied service name;
- no arbitrary shell command;
- no arbitrary filesystem path;
- no arbitrary URL fetch;
- request size and credential-count limits;
- timeouts on every privileged operation;
- secrets never logged.

Agent operations:

1. `InspectNaive`
   - find exactly one supported `forward_proxy` block;
   - return safe metadata plus credential usernames;
   - for first import only, return current secret material through the protected local Unix channel to the API process, never to browser/log/audit.
2. `ValidateNaiveDesiredState`
   - validate username/password policy;
   - render candidate from the exact current Caddyfile;
   - run `caddy validate`;
   - inspect adapted JSON and require the PVNaive route and `forward_proxy` to remain present.
3. `ApplyNaiveDesiredState`
   - require `expected_current_sha256` optimistic-concurrency match;
   - create exact backup + checksum;
   - install validated candidate;
   - call `systemctl reload caddy-naive.service` only;
   - verify service active, PID unchanged and restart count unchanged;
   - verify panel/API/camouflage invariants;
   - return old/new SHA and backup path.
4. `RollbackNaiveRevision`
   - restore an agent-created validated backup only;
   - validate restored Caddyfile;
   - reload only;
   - rerun invariants.
5. `Health`
   - safe status only; no secrets.

## Caddy transformation contract

PVNaive must not reconstruct the entire Caddyfile from a template.

The transformer must parse brace depth sufficiently to locate the single supported `forward_proxy { ... }` block and replace only credential directives. Everything else in the file must remain byte-for-byte identical whenever possible.

Preserve:

- global Caddy options;
- `encode`;
- current PVNaive panel/API subroute;
- `hide_ip`;
- `hide_via`;
- `probe_resistance` and its secret value;
- camouflage `root` and `file_server`;
- unrelated comments/whitespace outside the credential replacement span.

Fail closed if:

- zero or more than one `forward_proxy` block is found;
- credential syntax cannot be parsed unambiguously;
- candidate removes or reorders required management/data-plane handlers unexpectedly;
- current SHA differs from the revision the caller inspected;
- `caddy validate` fails;
- reload causes PID/restart-count change;
- post-apply invariants fail.

Multiple `basic_auth` credentials must not be assumed supported merely from documentation. Before production ownership, CI/rehearsal or a live no-install validation must prove that the exact installed custom Caddy binary accepts the rendered multi-credential syntax.

## Credential data model

Do not misuse existing `pvnaive.credentials` for the current global Naive account. That table is tied by foreign keys to reseller/user/subscription rows, and S05 has not created the commercial user lifecycle yet. Creating fake user/subscription rows would corrupt the domain model.

Migration `0003` should introduce a temporary-but-first-class global runtime table such as `pvnaive.naive_runtime_credentials`:

- `id uuid PRIMARY KEY DEFAULT gen_random_uuid()`;
- `username text NOT NULL` with a case-sensitive unique constraint matching runtime semantics;
- `secret_hash bytea NOT NULL CHECK octet_length(secret_hash)=32`;
- `secret_ciphertext bytea NOT NULL`;
- `secret_nonce bytea NOT NULL CHECK octet_length(secret_nonce)=12`;
- `encryption_key_id text NOT NULL`;
- `status text NOT NULL CHECK status IN ('active','disabled','revoked')`;
- `origin text NOT NULL CHECK origin IN ('imported','panel')`;
- `created_by_actor_id uuid REFERENCES pvnaive.actors(id)`;
- `updated_by_actor_id uuid REFERENCES pvnaive.actors(id)`;
- `created_at timestamptz`;
- `updated_at timestamptz`;
- `rotated_at timestamptz`;
- `revoked_at timestamptz`;
- monotonically increasing record revision or equivalent optimistic-concurrency field.

At least one active runtime credential must exist after initial ownership. Mutations that would disable/revoke the last active credential are rejected before runtime apply.

`DELETE /api/v1/runtime/naive/credentials/{id}` is a **soft revoke**, not a physical row deletion. The row and secret remain encrypted for the configured rollback/audit retention window. Physical purge is a later maintenance/security-retention feature and is not part of this extension.

Later S05/S06 work will explicitly migrate/associate appropriate runtime credentials with `pvnaive.credentials`; that later mapping is not faked in this extension.

## Runtime revision state

Reuse the existing runtime revision model only if its actual schema can represent the required state without weakening constraints. During implementation, inspect `pvnaive.runtime_revisions` precisely before choosing whether migration 0003 extends it or adds a narrowly scoped Naive revision table.

Each successful or failed apply must retain safe metadata:

- revision ID;
- actor ID;
- desired credential IDs/usernames/statuses only, never plaintext passwords;
- previous Caddy SHA;
- candidate Caddy SHA;
- applied Caddy SHA;
- exact backup path/reference;
- validation result;
- apply result;
- rollback result if any;
- timestamps.

## Secret encryption

Runtime credential encryption gets its own key:

`/etc/pvnaive/runtime.key`

Contract:

- exactly 32 random bytes;
- `root:pvnaive 0640`;
- AES-256-GCM;
- unique random 12-byte nonce per encryption;
- SHA-256 fingerprint/hash stored separately for integrity/change detection;
- never reuse `/etc/pvnaive/auth.key`;
- never reuse the age backup key;
- raw secrets never appear in database logs, API logs, audit records, GitHub evidence or HTTP GET responses.

Imported current Caddy secret is encrypted immediately after local inspection and zeroed/released as soon as practical.

## Username/password policy

The first implementation intentionally uses a conservative renderer-safe character set to prevent Caddyfile injection.

Username:

- 1–64 bytes;
- allowed: ASCII letters, digits, `.`, `_`, `@`, `+`, `-`;
- no whitespace, colon, quotes, backslash, braces, control characters, newline or NUL.

Password:

- 14–128 bytes;
- no newline, carriage return, NUL or control characters;
- initial accepted alphabet should be renderer-safe visible ASCII; expand only after parser/escaping tests prove correctness.

Server-generated password option:

- 24 random bytes from CSPRNG;
- base64url without padding;
- displayed exactly once only after both runtime apply **and** final database applied-state commit succeed.

Existing imported credential is not silently changed to satisfy new password-generation policy; import must preserve the live working value exactly.

## API contract

Add typed routes under `/api/v1/runtime/naive`.

Read:

- `GET /api/v1/runtime/naive`
- `GET /api/v1/runtime/naive/credentials`
- `GET /api/v1/runtime/naive/revisions`

Mutations:

- `POST /api/v1/runtime/naive/credentials`
- `PATCH /api/v1/runtime/naive/credentials/{id}` for rename/status changes
- `POST /api/v1/runtime/naive/credentials/{id}/rotate-password`
- `DELETE /api/v1/runtime/naive/credentials/{id}` (soft revoke)
- `POST /api/v1/runtime/naive/revisions/{id}/rollback`

Authorization:

- Owner may perform all operations.
- Admin may read safe runtime status/credential metadata if product RBAC requires it.
- Mutating runtime credentials is Owner-only in this first extension.
- Operator may continue to use safe general runtime status but never read secrets.

All mutations require:

- authenticated session;
- existing CSRF protection;
- idempotency key;
- expected revision/ETag or equivalent optimistic concurrency;
- request-body unknown-field rejection;
- stable error codes;
- audit entry.

Password response behavior:

- listing never returns plaintext or ciphertext;
- rename/status responses never return secrets;
- user-supplied password is never echoed;
- a generated password may be returned once in the successful mutation response only after the applied database state is durably committed, and the response must be marked no-store.

## Audit contract

Record at least:

- initial live credential import;
- credential create;
- username rename;
- password rotation;
- credential enable/disable;
- credential revoke;
- desired-state validation failure;
- runtime apply success/failure;
- runtime rollback success/failure;
- reconciliation/compensation failure if database finalization and runtime state diverge.

Never audit:

- plaintext password;
- encrypted secret bytes;
- current `probe_resistance` secret;
- raw Caddyfile containing secrets.

## Initial import / ownership handoff

The first production deployment is a guarded import, not a blind rewrite.

1. Generate runtime encryption key if absent.
2. Install/start Runtime Agent on Unix socket.
3. Agent inspects current live Caddyfile and requires exactly one supported forward-proxy credential set.
4. API/operator import path encrypts current live credential into PostgreSQL without exposing it to browser/log.
5. Renderer reconstructs a candidate from imported state.
6. Candidate must validate with the exact live Caddy binary.
7. Candidate's non-credential Caddy content must match live content exactly.
8. Before any production credential mutation, prove that rendering the imported state is semantically equivalent to the current live forward-proxy behavior.
9. If import or equivalence is ambiguous, stop without changing Caddy.

## UI design

Add `/runtime/naive` to the authenticated panel.

Runtime summary:

- status: Active/Degraded/Down/Unknown;
- domain/listener metadata;
- Caddy service state;
- current applied revision;
- current Caddy SHA (safe metadata);
- last apply status/time.

Credential table:

- username;
- status;
- origin (`imported` or `panel`);
- created/updated/rotated timestamps;
- actions appropriate to Owner permission.

Owner actions:

- Add credential;
- Rename username;
- Change password;
- Generate new password;
- Enable/Disable;
- Revoke/Delete with explicit confirmation (backend soft-revokes);
- Roll back to a prior known-good runtime revision.

UI safety:

- never show existing password;
- one-time generated secret display has copy button and clear warning;
- destructive operations use confirmation;
- disabling/revoking last active credential is blocked with an explanatory error;
- show apply/validation/rollback state;
- error messages show request ID, not stack traces;
- responsive desktop/mobile;
- dashboard Runtime card may use real status only after API is available;
- user/quota/accounting metrics remain placeholders or clearly unavailable, never fabricated.

## Data-plane availability principle

The Web UI/API are management-plane components. Existing Naive data-plane sessions must not depend on the panel remaining online.

A failed API/UI deploy must leave the current Caddy/Naive configuration untouched. A failed runtime apply must restore the exact last-known-good Caddyfile through the agent.

## Runtime/DB consistency and compensation

A runtime mutation crosses PostgreSQL and an external side effect (Caddy reload), so it cannot pretend to be one ACID transaction.

Use an explicit state machine such as `desired -> validating -> applying -> applied` with failure/rollback states. The API must never report success merely because Caddy reload succeeded.

Critical compensation rule:

- if Caddy apply and smoke checks succeed but the final PostgreSQL transition to `applied` cannot be durably committed, immediately invoke the agent to restore the exact pre-apply Caddy backup and reload it;
- then persist/emit a reconciliation failure as soon as the database is reachable;
- if that compensating rollback also fails, fail closed, surface `runtime_reconciliation_required`, keep all evidence/backup references, and do not return a generated secret as successful output.

This rule prevents silent DB/runtime split-brain.

## Production apply gates

Every credential mutation follows:

1. validate HTTP/RBAC/CSRF/idempotency/revision;
2. persist desired mutation/revision in a non-applied state;
3. durably commit that desired state before privileged runtime side effects;
4. render complete active credential set;
5. ask agent to validate against expected current Caddy SHA;
6. exact Caddy backup + checksum;
7. install candidate;
8. `systemctl reload caddy-naive.service` only;
9. require Caddy active;
10. require MainPID unchanged;
11. require NRestarts unchanged;
12. require ports 22/80/443 still listening;
13. require API listener still exactly `127.0.0.1:8080`;
14. require `https://namir.softarg.ir/panel/` healthy;
15. require `/api/v1/health/ready` healthy;
16. require camouflage root response unchanged;
17. where safely testable, verify one valid Naive credential succeeds and disabled/revoked credential fails;
18. finalize the runtime revision and credential mutation to `applied` in PostgreSQL and durably commit;
19. only after step 18 may the API return success or a one-time generated password.

On any failure before step 18 after Caddy switched, restore exact backup, validate it, reload, verify invariants, and mark the attempted revision failed/rolled back. On failure of step 18 itself, apply the explicit compensation rule above.

## Tests / acceptance criteria

No implementation is accepted without RED->GREEN tests for:

1. migration 0003 apply/reapply/down/checksum on disposable PostgreSQL 18;
2. runtime secret AES-GCM roundtrip and wrong-key/tamper failure;
3. no raw secret in DB-safe outputs/log/audit fixtures;
4. username/password injection rejection;
5. Caddy parser finds one forward_proxy and rejects zero/multiple/ambiguous blocks;
6. transformer changes only credential directives and preserves all other bytes;
7. exact installed custom Caddy validates the chosen multiple-credential syntax before production ownership;
8. agent uses Unix socket only and has no TCP listener;
9. agent rejects arbitrary paths/services/commands;
10. expected-Caddy-SHA concurrency conflict fails before mutation;
11. last-active-credential guard;
12. DELETE is soft revoke and retains audit/rollback history;
13. API RBAC matrix;
14. API CSRF/idempotency/revision behavior;
15. secret never returned from GET/list endpoints;
16. generated password is one-time response only after durable applied-state commit;
17. apply success leaves Caddy PID/restart count unchanged;
18. injected validation/reload/smoke failure restores exact previous Caddyfile;
19. injected final DB-commit failure after successful Caddy apply triggers compensating Caddy rollback;
20. compensation failure produces reconciliation-required state and never reports success;
21. panel/API route and camouflage root remain intact;
22. initial live import preserves the existing working credential;
23. new credential works after apply;
24. rotated credential behavior is correct;
25. disabled/revoked credential no longer authenticates after successful apply;
26. rollback restores prior working credential set;
27. web `/runtime/naive` renders status and credential metadata without secret leakage.

## Non-goals / honesty rules

Until separate PoCs are complete, the UI must not claim:

- exact per-credential traffic accounting;
- live session visibility;
- speed limiting;
- concurrency enforcement;
- device counting;
- quota enforcement;
- production readiness for hundreds of users.

Capability flags must reflect what the adapter has actually proven.

## Rollout order

1. written spec approval;
2. implementation plan;
3. TDD migration + secret storage;
4. TDD parser/renderer;
5. TDD privileged agent;
6. TDD API/service layer;
7. TDD UI;
8. full CI/rehearsal;
9. read-only live import preflight;
10. guarded production import;
11. external panel/runtime postflight;
12. only then allow real credential mutations from the browser.

No server mutation is permitted merely because the code compiles or CI passes.