# S04 Caddy read-only inspection — PASSED

Timestamp: `2026-08-28T00:59:12Z`
Host: `testAmir5-3`

## Result

Read-only inspection completed with no configuration/service mutation:

```text
CADDY_INSPECTION=PASSED
CONFIG_CHANGED=false
SERVICE_RELOADED=false
SERVICE_RESTARTED=false
NEXT=DESIGN_AND_TEST_S04_CADDY_FINALIZER
```

## Exact live Caddy baseline

- `/etc/caddy/Caddyfile` SHA-256: `101884de2dd11cb9d276df8e72cd068bed50e4ec6eb4ebb477184dda7a86e8b1`.
- owner/group/mode: `root:caddy 0640`.
- active service: `caddy-naive.service`.
- service user/group: `caddy:caddy`.
- `ExecStart`: `/usr/local/bin/caddy run --environ --config /etc/caddy/Caddyfile --adapter caddyfile`.
- `ExecReload`: `/usr/local/bin/caddy reload --config /etc/caddy/Caddyfile --adapter caddyfile --force`.
- restart policy: `on-failure`.
- systemd hardening includes `NoNewPrivileges=true`, `ProtectSystem=full`, `PrivateTmp=true`, `ProtectHome=true`.
- Caddy listens on TCP 80/443; SSH on 22.
- PVNaive API remains loopback-only on `127.0.0.1:8080`.

## Current redacted Caddyfile structure

```caddy
{
    order forward_proxy before file_server
    servers {
        protocols h1 h2
    }
    log {
        exclude http.log.error
    }
}

:443, namir.softarg.ir {
    encode zstd gzip

    forward_proxy {
        basic_auth <redacted>
        hide_ip
        hide_via
        probe_resistance <redacted>
    }

    root * /var/www/naive
    file_server
}
```

Credentials/probe-resistance values are intentionally not recorded.

`caddy validate` returned valid configuration. Adapted order on the live config is effectively `vars(root) -> encode -> forward_proxy -> file_server`.

Relevant installed modules include `http.handlers.forward_proxy`, `http.handlers.reverse_proxy`, `http.handlers.file_server`, `http.handlers.headers`, `http.handlers.rewrite`, and `http.handlers.subroute`.

## PVNaive web/API state

- `/opt/pvnaive/web/current` -> `/opt/pvnaive/web/releases/20260828T001418Z`.
- web release owner/group/mode: `root:pvnaive 0750`.
- files: `index.html`, `pvnaive-mark.svg`, one JS bundle and one CSS bundle under `assets/`.
- Caddy runs as group `caddy`, so it must not serve directly from this `root:pvnaive 0750` release without an explicit publication/readability step.
- API live: `{"service":"pvnaive-api","status":"ok"}`.
- API ready: `{"ready":true,"status":"ready"}`.

## Exposure design constraints now known

1. Preserve the existing `forward_proxy` block byte-for-byte; never reconstruct secrets from redacted output.
2. Add a path-specific PVNaive route ahead of `forward_proxy`, then verify adapted handler order shows the new subroute before the forward proxy.
3. Expose API as `/api/*` to `127.0.0.1:8080` without stripping `/api` because the backend routes are already `/api/v1/...`.
4. Publish a Caddy-readable copy of the static web release under `/var/www/` rather than changing the immutable `/opt/pvnaive/web/releases/...` ownership in place.
5. Current S04 UI uses absolute `/api/...` requests and absolute `/pvnaive-mark.svg`; the build also uses root `/assets/...`. A `/panel/` preview therefore needs controlled routing for the exact static asset paths or a later build adjusted for subpath hosting.
6. Use exact pre-change Caddy backup/SHA, candidate validation, controlled reload only (never restart), and exact rollback+reload on any failure.
7. Preserve the existing camouflage/fallback root response and NaiveProxy behavior for all non-PVNaive paths.

Official `S04-AUTH` remains **IN PROGRESS** until public panel/API exposure and independent external postflight pass.