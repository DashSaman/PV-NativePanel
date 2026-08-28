# CONTINUE HERE — PVNaive

If a Chat/Agent session was interrupted, start here.

## Read in this order

1. `ops/S04_LIVE_STATE.md` — latest live server state, failures, fixes, green artifact, and exact next action.
2. `AGENT_HANDOFF.md` — broader project history and non-negotiable constraints.
3. `ops/DEPLOYMENT_PROGRESS.md` — official stage ledger.
4. Active implementation branch: `s04-auth`; open PR: `#2`.

## Current one-line state

`S00-S03=PASSED`; `S04-AUTH=IN PROGRESS`; target server DB is already schema 2 in a clean recovery state; all failed S04 service artifacts were rolled back; Caddy/SSH/firewall are unchanged.

## Pinned green S04 recovery artifact

- source commit: `11c54dc1faae99a1491c750b30db9faa44a0c3ae`
- CI run: `33128780602` — Go/Web/PostgreSQL18/end-to-end rehearsal/bundle all SUCCESS
- artifact ID: `9669443464`
- inner archive: `PVNaive-S04-11c54dc1faae.tar.gz`
- required SHA-256: `52acdde2bff6777abeb31c081c86e018d1e7f5f0cdb39f8eb4151efbba2820fc`

Do not reuse the older `b4803e27...` bundle. For the exact recovery/deploy sequence and root-cause evidence, read `ops/S04_LIVE_STATE.md`.
