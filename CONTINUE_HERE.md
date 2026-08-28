# CONTINUE HERE — PVNaive

If a Chat/Agent session was interrupted, start here.

## Read in this order

1. `ops/evidence/S04-20260828T010546Z-caddy-panel-exposure-pass.md` — public panel/API exposure PASSED with reload-only Caddy integration and preserved Naive/camouflage behavior.
2. `docs/superpowers/specs/2026-08-28-naive-runtime-credentials-design.md` on branch `s04-auth` — approved-in-chat architecture for the next Owner-prioritized Naive credential-management extension; written spec is awaiting final user review.
3. `ops/evidence/S04-20260828T005912Z-caddy-inspection-pass.md` — exact pre-exposure Caddy inspection.
4. `ops/evidence/S04-20260828T005544Z-real-owner-localhost-auth-pass.md` — real Owner localhost auth/session/logout/revocation PASSED.
5. `ops/evidence/S04-20260828T005033Z-owner-bootstrap-pass.md` — real Owner bootstrap PASSED.
6. `ops/S04_LIVE_STATE.md` — authoritative current live state.
7. `AGENT_HANDOFF.md` — broader project history and constraints.
8. `ops/DEPLOYMENT_PROGRESS.md` — official stage ledger.
9. Active implementation branch: `s04-auth`; draft PR `#2`.

## Current one-line state

`S00-S03=PASSED`; `S04-AUTH=IN PROGRESS`. Localhost auth, Owner bootstrap, real Owner auth lifecycle, Caddy inspection and public panel/API exposure have all PASSED. The Owner visually confirmed that both the login screen and authenticated dashboard render at `https://namir.softarg.ir/panel/`.

The next owner-prioritized feature is real management of the currently deployed NaiveProxy `forward_proxy basic_auth` credentials from the panel: import the current credential, add multiple credentials, rename usernames, rotate passwords, enable/disable and revoke/delete safely.

This work is named `S04R-NAIVE-CREDENTIALS`. It is a deliberately narrow pre-S05 extension requested by the Owner and **does not** mark full `S06-RUNTIME` complete or fabricate user/quota/accounting capability.

## Current production Caddy state after panel exposure

- pre-exposure SHA: `101884de2dd11cb9d276df8e72cd068bed50e4ec6eb4ebb477184dda7a86e8b1`;
- current post-exposure SHA: `21db739ca3911fb9974eaabe9db07f22e62d84ddc1309ef6bcea3ec247c9ab23`;
- exact pre-change backup: `/var/backups/pvnaive/caddy/20260828T010546Z/Caddyfile`;
- Caddy MainPID remained `1045` and NRestarts remained `0` during exposure;
- change used `systemctl reload caddy-naive.service`, never restart;
- panel `/panel/` returned 200;
- `/panel` returned 308;
- current JS/CSS/logo assets returned 200;
- `/api/v1/health/ready` through HTTPS returned 200/ready;
- camouflage `/` remained HTTP 200 with exact same body SHA `49ed9babd86c42c4664fdf71f4715080575e2a3a47b1b9aa9cbbad3343b11132`;
- API still listens only on `127.0.0.1:8080`;
- existing Naive `forward_proxy` block remained present;
- SSH/firewall unchanged.

## Current product limitation

The public UI is still the S04 protected preview. It authenticates for real but does not yet provide Runtime credential controls. Dashboard user/quota/usage values remain unavailable and must not be fabricated.

## Approved architecture for `S04R-NAIVE-CREDENTIALS`

The chat design was explicitly approved by the Owner.

Use PostgreSQL desired state plus a narrow privileged local `pvnaive-runtime-agent` over a Unix socket. The unprivileged API must never gain arbitrary root shell/filesystem access.

Core safety rules:

- dedicated runtime encryption key; do not reuse auth/MFA or backup keys;
- import the existing live Naive credential without changing it;
- no old plaintext password display in UI;
- new/generated password may be shown once only;
- exact Caddyfile transform must touch only supported credential directives;
- expected-current-SHA optimistic concurrency;
- `caddy validate` before install;
- exact backup before switch;
- reload only;
- PID/NRestarts/panel/API/camouflage/SSH/data-plane invariants after apply;
- exact rollback + reload on failure;
- block disabling/deleting the last active credential;
- no fake accounting/session/quota claims.

## Exact next action

Do **not** start implementation until the Owner reviews the written design spec:

`docs/superpowers/specs/2026-08-28-naive-runtime-credentials-design.md`

After the Owner explicitly approves that written spec, invoke the `writing-plans` workflow and create:

`docs/superpowers/plans/2026-08-28-naive-runtime-credentials.md`

Then implement through TDD/CI. No live credential mutation is allowed before implementation, rehearsal and guarded import gates pass.

Official S04 should not be falsely advanced merely because the preview is visible; an independent post-exposure gate remains part of the formal stage completion record.