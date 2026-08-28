# S04-AUTH Live State — testAmir5-3

Last updated: 2026-08-28 after public panel/API exposure PASS and visual Owner confirmation

> Authoritative fast continuation file. Read `CONTINUE_HERE.md`, newest evidence, and the pending Naive runtime-credential design before acting on production.

## Current state

- `S00-S03=PASSED`.
- `S04-AUTH=IN PROGRESS`; do not mark it official PASSED solely because the preview is visible.
- localhost S04 deployment and independent localhost postflight PASSED.
- exactly one real active Owner exists.
- real Owner login/session/me/CSRF logout/revocation PASSED.
- exact Caddy inspection PASSED.
- public panel/API exposure PASSED at `2026-08-28T01:05:46Z`.
- Owner visually confirmed the login screen and authenticated dashboard render at `https://namir.softarg.ir/panel/`.
- next owner-prioritized work: `S04R-NAIVE-CREDENTIALS`, a narrow pre-S05 extension for managing existing NaiveProxy credentials from the panel.

## Public exposure result

Newest evidence:

`ops/evidence/S04-20260828T010546Z-caddy-panel-exposure-pass.md`

Verified production facts:

- Caddy pre-exposure SHA: `101884de2dd11cb9d276df8e72cd068bed50e4ec6eb4ebb477184dda7a86e8b1`;
- Caddy post-exposure/current SHA: `21db739ca3911fb9974eaabe9db07f22e62d84ddc1309ef6bcea3ec247c9ab23`;
- exact backup: `/var/backups/pvnaive/caddy/20260828T010546Z/Caddyfile`;
- Caddy MainPID stayed `1045`;
- NRestarts stayed `0`;
- reload only, no restart;
- `/panel/` HTTP 200;
- `/panel` HTTP 308;
- deployed JS/CSS/logo HTTP 200;
- `/api/v1/health/ready` through the public HTTPS route HTTP 200 and ready=true;
- existing camouflage root remained byte-identical with SHA `49ed9babd86c42c4664fdf71f4715080575e2a3a47b1b9aa9cbbad3343b11132`;
- existing `forward_proxy` remained present;
- API listener remained exactly `127.0.0.1:8080`;
- SSH/firewall unchanged.

Caddy-readable web publication:

`/var/www/pvnaive-preview/current`

Current preview assets:

- `/assets/index-CVNNekwm.js`;
- `/assets/index-DT5seDoF.css`;
- `/pvnaive-mark.svg`.

## Product limitation now visible to the Owner

The preview is real and authenticated, but it still intentionally lacks business/runtime management. Runtime card is not connected to actual Naive management yet; users, quota and usage are not implemented and must not be fabricated.

The Owner explicitly requested the following next capabilities:

- display the current NaiveProxy credential that exists in Caddy;
- change/rotate its password;
- rename its username;
- add multiple new username/password credentials;
- enable/disable credentials;
- revoke/delete credentials safely;
- apply those changes without breaking NaiveProxy, panel, API or camouflage site.

## Stage semantics for the next work

This requested work is named `S04R-NAIVE-CREDENTIALS`.

It is an owner-prioritized pre-S05 extension and does not mean the entire later S06 Runtime stage is complete. It explicitly excludes quota/accounting/session usage/subscription/fleet management.

Written design spec on `s04-auth`:

`docs/superpowers/specs/2026-08-28-naive-runtime-credentials-design.md`

Spec commit:

`085f98306528a79beaa22532209cc0af27565bdf`

The in-chat architecture was approved. The written spec is now awaiting the Owner's final review before implementation planning, per the repository development workflow.

## Approved architecture summary

- PostgreSQL desired state is the management source of truth.
- Existing global Naive credential is not forced into the user/subscription-bound `pvnaive.credentials` table before S05.
- a narrowly privileged `pvnaive-runtime-agent` communicates with the unprivileged API over `/run/pvnaive/runtime-agent.sock` only;
- no arbitrary shell, path, service or URL operations;
- dedicated `/etc/pvnaive/runtime.key`, not auth/MFA or age backup keys;
- initial import preserves the working live credential;
- no retrieval/display of old plaintext password in the browser;
- exact Caddy transform only for supported auth directives;
- optimistic expected-current-SHA check;
- `caddy validate` + exact backup + reload-only + smoke + exact rollback;
- block disabling/deleting the last active credential;
- full secret-safe audit;
- capability claims remain honest: no exact accounting/session/quota claims until proven.

## Exact next action

Do not mutate the live runtime yet.

1. Owner reviews the written spec.
2. After explicit written-spec approval, create the detailed `writing-plans` TDD implementation plan.
3. Implement migration/parser-agent/API/UI in CI.
4. Perform a read-only live import preflight.
5. Only after all rehearsal gates pass, perform guarded import and then enable browser mutations.

Earlier evidence:

- `ops/evidence/S04-20260828T005912Z-caddy-inspection-pass.md`
- `ops/evidence/S04-20260828T005544Z-real-owner-localhost-auth-pass.md`
- `ops/evidence/S04-20260828T005033Z-owner-bootstrap-pass.md`
- `ops/evidence/S04-20260828T004344Z-final-independent-postflight-pass.md`

Real Owner credentials and Naive secrets are intentionally omitted from repository evidence.