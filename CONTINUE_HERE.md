# CONTINUE HERE — PVNaive

If a Chat/Agent session was interrupted, start here.

## Read in this order

1. `ops/evidence/S04-20260828T005912Z-caddy-inspection-pass.md` — exact live Caddy inspection PASSED; current route/service/web-root constraints are known.
2. `ops/evidence/S04-20260828T005544Z-real-owner-localhost-auth-pass.md` — real Owner localhost auth/session/logout/revocation PASSED.
3. `ops/evidence/S04-20260828T005033Z-owner-bootstrap-pass.md` — real Owner bootstrap PASSED.
4. `ops/evidence/S04-20260828T004344Z-final-independent-postflight-pass.md` — final independent localhost S04 postflight PASSED.
5. `ops/S04_LIVE_STATE.md` — authoritative current live state.
6. `AGENT_HANDOFF.md` — broader project history and constraints.
7. `ops/DEPLOYMENT_PROGRESS.md` — official stage ledger.
8. Active implementation branch: `s04-auth`; draft PR `#2`.

## Current one-line state

`S00-S03=PASSED`; `S04-AUTH=IN PROGRESS`. Localhost S04, Owner bootstrap, real Owner login/session/me/CSRF logout/revocation, and live Caddy inspection all PASSED. The next requested action is to expose the current S04 web preview on `https://namir.softarg.ir/panel/` while preserving NaiveProxy and the existing camouflage root.

The current S04 UI is a protected authentication/dashboard preview. It does **not yet** implement management of the existing NaiveProxy `basic_auth` username/password; those runtime/management controls belong to later implementation stages. Do not misrepresent this preview as the finished panel.

## Exact live Caddy baseline

- Caddyfile SHA: `101884de2dd11cb9d276df8e72cd068bed50e4ec6eb4ebb477184dda7a86e8b1`.
- `/etc/caddy/Caddyfile`: `root:caddy 0640`.
- active service: `caddy-naive.service`, user/group `caddy:caddy`.
- reload command is configured as `/usr/local/bin/caddy reload --config /etc/caddy/Caddyfile --adapter caddyfile --force`.
- existing site is `:443, namir.softarg.ir` with `encode`, `forward_proxy` (credentials/probe resistance), then `/var/www/naive` file server.
- existing adapted handler order: root vars -> encode -> forward_proxy -> file_server.
- Caddy validate PASSED.
- API remains only `127.0.0.1:8080`, live/ready healthy.
- `/opt/pvnaive/web/current` -> `/opt/pvnaive/web/releases/20260828T001418Z`, owned `root:pvnaive 0750`; Caddy cannot safely serve that immutable release directly without a readability/publication step.
- web bundle contains `index.html`, `pvnaive-mark.svg`, one JS asset and one CSS asset.

## Exposure design to execute

Use a bounded, rollback-safe publication:

1. Keep the existing `forward_proxy` block and its secrets byte-for-byte untouched.
2. Copy the installed static web release into a dedicated Caddy-readable publication directory under `/var/www/pvnaive-preview` with `root:caddy` ownership; do not mutate the immutable `/opt/pvnaive/web/releases/...` release.
3. Insert one path-specific `route` block before `forward_proxy` by transforming the exact current Caddyfile, not reconstructing it.
4. Route `/api/*` to `127.0.0.1:8080` without stripping `/api`.
5. Route `/panel` -> `/panel/`, serve `/panel/*` from the published static copy, and route the exact root-level assets used by the current build (`/assets/...` and `/pvnaive-mark.svg`) to that copy so the existing S04 build renders without rebuilding.
6. Validate the candidate before installation and inspect adapted JSON to require the new subroute to occur before `forward_proxy`.
7. Take an exact pre-change Caddyfile backup/SHA.
8. Install candidate and use **reload only, never restart**.
9. Smoke `https://namir.softarg.ir/panel/`, exact JS/CSS/logo assets and `/api/v1/health/ready` through Caddy; require existing root camouflage response to remain byte-identical to its pre-change response.
10. On any failure after switch, restore the exact old Caddyfile and reload it.

After exposure PASS, have the user open the panel and log in visually. Then run an independent external postflight before advancing the official ledger. Only after external postflight may `S04-AUTH=PASSED` and `S05-USERS=NEXT`.

Never reuse the old `b4803e27...` bundle.