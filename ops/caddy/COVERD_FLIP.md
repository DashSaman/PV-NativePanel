# R6-FLIP-001 — Cover Site Flip Runbook (gated)

Status: **capability merged, production flip default-OFF** (`PVNAIVE_COVERD_ENABLED`
unset ⇒ the root keeps its 404). Perform the flip only with all pre-gates green.

## What flips

Before: every non-panel path on the node domain returns the Caddy 404.
After: the same paths serve this node's persona cover site, rendered by the
loopback `coverd` handler inside `pvnaive` (127.0.0.1:9444) through Caddy.
Panel paths (`/panel`, `/api/v1/*`, SSE) are untouched and keep precedence.

## Pre-gates (all mandatory)

1. Exact-head CI green on the deployed SHA (go + web + database + rehearsal + bundle).
2. Trusted read-only production audit: schema 33, services healthy, disk headroom.
3. Fresh encrypted backup **and** independent rollback snapshot taken within the hour.
4. Cover rehearsal evidence green (CAMO-001/002): no cookies, no CORS, no
   identifying headers, no panel vocabulary, natural degrade on source failure.
5. `PVNAIVE_COVERD_NODE_ID` chosen (stable identity used for persona + content key).

## Flip sequence

1. Add to `/opt/pvnaive/.env`:
   ```ini
   PVNAIVE_COVERD_ENABLED=1
   PVNAIVE_COVERD_NODE_ID=<stable-node-id>
   # optional pins:
   # PVNAIVE_COVERD_LISTEN=127.0.0.1:9444
   # PVNAIVE_COVERD_PERSONA=news-portal|cultural|city-services|news-agency|research|charity
   ```
2. Recreate the container so the env takes effect; wait for `/api/v1/health/ready`.
3. Loopback postflight **before** touching Caddy:
   - `curl -s http://127.0.0.1:9444/ | head -5` → persona HTML, 200.
   - `curl -s -o /dev/null -w '%{http_code}' http://127.0.0.1:9444/wp-admin/setup.php` → 404.
   - `curl -sI http://127.0.0.1:9444/ | grep -iE 'set-cookie|access-control'` → empty.
4. Apply the Caddy snippet (see `coverd-flip.caddyfile`): insert the
   `@cover not panel_paths` handle **above** the panel handle in the site block,
   `caddy validate --config /etc/caddy/Caddyfile`, then `systemctl reload caddy`.
5. Public postflight:
   - `https://<domain>/` → 200 persona HTML; `Server: web`; no cookies/CORS;
     no panel strings in markup.
   - `https://<domain>/panel/` → panel loads unchanged; login postflight on a
     real account; SSE stream still flushes; a real customer CONNECT succeeds.
   - `https://<domain>/robots.txt` → allow `/`, disallow `/feed.xml`.
6. Record evidence (curl transcripts + SHA) in the ops evidence folder.

## Rollback (single move, seconds)

- `PVNAIVE_COVERD_ENABLED` unset (or `systemctl`-level container env revert)
  → recreate container; Caddy route then fails closed to its own 404 for
  cover paths; panel unaffected. Remove the Caddy `@cover` handle on the next
  maintenance window (not required for immediate rollback).

## Red lines (unchanged)

- Personas are original official-style designs; never clone a real organization.
- Syndicated content keeps source attribution; media is embedded/linked, never re-hosted.
- No secrets, node ids or panel references ever appear in cover markup or headers.
