# S04-AUTH Live State — testAmir5-3

Last updated: 2026-08-28 after real Owner auth PASS and exact live Caddy inspection PASS

> Authoritative fast continuation file. Read `CONTINUE_HERE.md` and newest S04 evidence before acting on production.

## Current state

- `S00-S03=PASSED`.
- `S04-AUTH=IN PROGRESS`; public Caddy exposure/external postflight are the remaining gates.
- S04 API/auth/web localhost deployment is healthy.
- final independent localhost postflight PASSED.
- exactly one real active Owner exists.
- real Owner login/session/me/CSRF logout/revocation PASSED.
- live Caddy read-only inspection PASSED.
- next requested action: expose the S04 web preview so it can be viewed in a browser before continuing later stages.

## Important product limitation at this exact stage

The current web UI is the S04 authentication/dashboard preview. It shows secure login and a protected dashboard, but user/runtime management controls are intentionally disabled. It does **not yet** expose editing of the existing NaiveProxy `forward_proxy basic_auth` username/password. That functionality must be implemented in later runtime/management stages; do not claim it exists now.

## Caddy baseline captured live at 2026-08-28T00:59:12Z

- `/etc/caddy/Caddyfile` SHA-256: `101884de2dd11cb9d276df8e72cd068bed50e4ec6eb4ebb477184dda7a86e8b1`.
- file metadata: `root:caddy 0640`, size 379 bytes.
- service: `caddy-naive.service`, active/enabled, PID was 1045 during inspection.
- service user/group: `caddy:caddy`.
- `ExecStart`: `/usr/local/bin/caddy run --environ --config /etc/caddy/Caddyfile --adapter caddyfile`.
- `ExecReload`: `/usr/local/bin/caddy reload --config /etc/caddy/Caddyfile --adapter caddyfile --force`.
- restart policy: `on-failure`; future controlled changes must call reload, never restart.
- existing Caddy site: `:443, namir.softarg.ir`.
- site behavior: `encode zstd gzip` -> `forward_proxy` with existing secret-bearing auth/probe-resistance block -> root `/var/www/naive` -> `file_server`.
- adapted live order: `vars(root)`, `encode`, `forward_proxy`, `file_server`.
- `caddy validate` PASSED; only existing formatting warning remains.
- modules required for finalizer are present: forward proxy, reverse proxy, file server, headers, rewrite, subroute.
- ports: SSH 22, Caddy 80/443, PVNaive API only `127.0.0.1:8080`.
- Caddy storage account path: `/var/lib/caddy`; account is `caddy`.

## Web publication constraint

`/opt/pvnaive/web/current` resolves to `/opt/pvnaive/web/releases/20260828T001418Z`, metadata `root:pvnaive 0750`. The active Caddy process runs as `caddy:caddy`, so the immutable release should not be re-owned or opened broadly merely to serve it.

Current build files:

- `index.html`
- `pvnaive-mark.svg`
- `assets/index-CVNNekwm.js`
- `assets/index-DT5seDoF.css`

The S04 source uses absolute `/api/v1/...` fetches and absolute `/pvnaive-mark.svg`; Vite output also uses root `/assets/...`. For immediate `/panel/` preview, publish a dedicated Caddy-readable copy and explicitly route the exact root-level build assets. A later web build can be normalized for subpath hosting.

## Exposure design / safety gates

1. Snapshot existing root camouflage response before change.
2. Publish static copy under `/var/www/pvnaive-preview` as `root:caddy`, leaving immutable `/opt/pvnaive/web/releases/...` unchanged.
3. Generate candidate by inserting one new PVNaive route into the exact original Caddyfile; existing `forward_proxy` block must remain byte-for-byte intact.
4. `/api/*` -> `127.0.0.1:8080` with URI preserved.
5. `/panel` redirects to `/panel/`; `/panel/*` serves SPA; exact JS/CSS/logo root paths serve from the preview publication.
6. Candidate must pass `caddy validate` and adapted-order verification requiring the new subroute before `forward_proxy`.
7. Back up exact old Caddyfile/SHA.
8. Install candidate and run `systemctl reload caddy-naive.service`; never restart.
9. Smoke panel, assets and API via HTTPS/domain and require pre-change root camouflage response unchanged.
10. Automatic exact Caddyfile restore + reload on any post-switch failure.

After visual panel login succeeds, run independent external postflight. Only then advance `S04-AUTH=PASSED` and `S05-USERS=NEXT`.

## Evidence

Newest: `ops/evidence/S04-20260828T005912Z-caddy-inspection-pass.md`.
Earlier:
- `ops/evidence/S04-20260828T005544Z-real-owner-localhost-auth-pass.md`
- `ops/evidence/S04-20260828T005033Z-owner-bootstrap-pass.md`
- `ops/evidence/S04-20260828T004344Z-final-independent-postflight-pass.md`

Real Owner credentials are intentionally omitted from the repository.