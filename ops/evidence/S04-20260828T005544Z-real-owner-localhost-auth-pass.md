# S04 real Owner localhost authentication — PASSED

Timestamp: `2026-08-28T00:55:44Z`
Host: `testAmir5-3`

## Result

The real production Owner account successfully completed the localhost authentication/session lifecycle against the live API on `127.0.0.1:8080`.

Final live output included:

```text
REAL_OWNER_LOCALHOST_AUTH=PASSED
LOGIN=PASSED
ME=PASSED
LOGIN_AUDIT=PASSED
SESSION_TOKEN_HASH_LENGTH=32
CSRF_LOGOUT=PASSED
OLD_SESSION_HTTP=401
SESSION_REVOKED=true
API_HEALTH=PASSED
CADDY_CHANGED=false
SSH_CHANGED=false
FIREWALL_CHANGED=false
NEXT=PREPARE_CADDY_PANEL_EXPOSURE
```

The real Owner email/password are intentionally not recorded here.

## Independently observed auth lifecycle

Before login:

- exactly one active Owner with a password hash existed;
- successful `auth.login` audit count for the Owner was `0`;
- Owner auth-session count was `0`;
- API listener was only `127.0.0.1:8080`.

Real login returned HTTP `200` and both expected cookies were present:

- `__Host-pvnaive_session`;
- `__Host-pvnaive_csrf`.

`GET /api/v1/me` with the returned session cookie returned HTTP `200` and Owner role.

After login:

- successful `auth.login` audit count became exactly `1`;
- Owner auth-session count became exactly `1`;
- selected test session persisted a 32-byte token hash and was initially unrevoked: `32|0`.

Logout with the CSRF cookie value supplied as `X-CSRF-Token` returned HTTP `200` and the expected logged-out response.

After logout:

- using the old session cookie against `/api/v1/me` returned HTTP `401`;
- the same session row retained a 32-byte token hash and had non-null `revoked_at`: `32|1`.

## Infrastructure invariants after real auth test

- API liveness: `{"service":"pvnaive-api","status":"ok"}`.
- API readiness: `{"ready":true,"status":"ready"}`.
- API listener remained only `127.0.0.1:8080`.
- Caddyfile SHA-256 remained exactly:
  `101884de2dd11cb9d276df8e72cd068bed50e4ec6eb4ebb477184dda7a86e8b1`.
- SSH remained active as `ssh.service`.
- Caddy/SSH/firewall were not changed by this test.

## Exact next action

Do **not** mutate Caddy yet. First inspect the exact live `/etc/caddy/Caddyfile`, its ownership/mode, active Caddy unit ExecStart/ExecReload, current route/adapted-config structure, active certificates/listeners, and the installed web root. Preserve the exact current Caddyfile SHA as the pre-change baseline.

Only after the existing NaiveProxy/forward-proxy route structure is understood should the S04 Caddy exposure finalizer be implemented and tested. The finalizer must:

1. back up the exact old Caddyfile and SHA;
2. expose the panel and API without breaking NaiveProxy;
3. validate the candidate before install;
4. use controlled `systemctl reload caddy-naive.service` — never restart;
5. restore the exact old file and reload it on any failure;
6. externally smoke-test panel/API and preserve NaiveProxy, SSH and firewall invariants.

Official `S04-AUTH` remains **IN PROGRESS** until controlled Caddy exposure and an independent external postflight pass. Only then may the ledger advance to `S04-AUTH=PASSED` and `S05-USERS=NEXT`.
