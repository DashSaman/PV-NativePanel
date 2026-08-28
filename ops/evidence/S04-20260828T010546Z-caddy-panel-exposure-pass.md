# S04 public panel/API exposure — PASSED

Timestamp: `2026-08-28T01:05:46Z`
Host: `testAmir5-3`
Domain: `namir.softarg.ir`

## Result

The S04 management preview was exposed through the existing production Caddy instance without restarting Caddy and without changing the existing camouflage root response or API loopback binding.

Observed live output included:

```text
PVNAIVE_PANEL_EXPOSURE=PASSED
PANEL_URL=https://namir.softarg.ir/panel/
PUBLIC_API=/api/v1/
CADDY_RELOAD_ONLY=true
CADDY_RESTARTED=false
EXISTING_ROOT_PRESERVED=true
NAIVE_FORWARD_PROXY_BLOCK_PRESERVED=true
API_LISTENER=127.0.0.1:8080
SSH_CHANGED=false
FIREWALL_CHANGED=false
```

## Exact live facts

Before the change:

- Caddy SHA-256: `101884de2dd11cb9d276df8e72cd068bed50e4ec6eb4ebb477184dda7a86e8b1`;
- Caddy MainPID: `1045`;
- Caddy NRestarts: `0`;
- camouflage root returned HTTP `200` with body SHA-256 `49ed9babd86c42c4664fdf71f4715080575e2a3a47b1b9aa9cbbad3343b11132`.

A Caddy-readable immutable publication copy was created under:

`/var/www/pvnaive-preview/current`

The exact web assets served were:

- `/assets/index-CVNNekwm.js`;
- `/assets/index-DT5seDoF.css`;
- `/pvnaive-mark.svg`.

Candidate validation proved:

- the new PVNaive route is before `forward_proxy`;
- `/api/*` uses upstream `127.0.0.1:8080`;
- the existing `forward_proxy` remains present;
- the pre-existing Caddy content outside the inserted management route was preserved.

Exact pre-change Caddy backup:

`/var/backups/pvnaive/caddy/20260828T010546Z/Caddyfile`

After installation:

- Caddy was reloaded only;
- MainPID remained `1045`;
- NRestarts remained `0`;
- `/panel/` returned HTTP `200`;
- `/panel` returned HTTP `308`;
- JS/CSS/logo each returned HTTP `200`;
- `/api/v1/health/ready` through HTTPS returned HTTP `200` and ready=true;
- camouflage root remained HTTP `200` with the exact same body SHA;
- API still listened only on `127.0.0.1:8080`.

Current post-exposure Caddyfile SHA-256:

`21db739ca3911fb9974eaabe9db07f22e62d84ddc1309ef6bcea3ec247c9ab23`

The Owner then visually opened the panel in a browser and confirmed both the login screen and authenticated dashboard render successfully.

## Product status after exposure

This is still an engineering preview. Authentication is real, but the dashboard still does not manage the live NaiveProxy credentials.

The Owner explicitly prioritized the next architectural extension: show and manage the existing Naive `basic_auth` credentials from the panel, including add, rename, password rotation, enable/disable and revoke/delete with validated Caddy reload and rollback.

That work is being designed as `S04R-NAIVE-CREDENTIALS`, a narrow pre-S05 owner-prioritized extension. It does not mark the later full S06 Runtime stage complete.

## Next action

Review and approve:

`docs/superpowers/specs/2026-08-28-naive-runtime-credentials-design.md`

Then create the detailed TDD implementation plan. Do not perform live credential mutations until the implementation, CI/rehearsal and guarded import gates pass.