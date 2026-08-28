# CONTINUE HERE — PVNaive

If a Chat/Agent session was interrupted, start here.

## Read in this order

1. `ops/evidence/S04_RECOVERY_20260828T001418Z.md` — newest live execution evidence and exact next action.
2. `ops/S04_LIVE_STATE.md` — S04 failure history, root causes, fixes, and pinned recovery design.
3. `AGENT_HANDOFF.md` — broader project history and non-negotiable constraints.
4. `ops/DEPLOYMENT_PROGRESS.md` — official stage ledger.
5. Active implementation branch: `s04-auth`; open PR: `#2`.

## Current one-line state

`S00-S03=PASSED`; the fixed S04 recovery stage has now executed successfully on `testAmir5-3`, created the S04 marker, started the API on loopback, and moved DB expected schema to 2. **S04 is still IN PROGRESS until an independent postflight passes.** Caddy/SSH/firewall remained unchanged in the stage output.

## Pinned green S04 artifact actually executed

- source commit: `11c54dc1faae99a1491c750b30db9faa44a0c3ae`
- CI run: `33128780602` — Go/Web/PostgreSQL18/end-to-end rehearsal/bundle all SUCCESS
- artifact ID: `9669443464`
- inner archive: `PVNaive-S04-11c54dc1faae.tar.gz`
- required SHA-256: `52acdde2bff6777abeb31c081c86e018d1e7f5f0cdb39f8eb4151efbba2820fc`

Observed live stage output at `2026-08-28 00:14 UTC` included:

- `RECOVERY_PREFLIGHT=PASSED`
- `RECOVERY_MODE=SCHEMA2_WITHOUT_MARKER`
- `PVNAIVE_BACKUP_RESULT=PASSED`
- `PVNAIVE_DB_EXPECTED_SCHEMA_VERSION=2`
- `S04_RESULT=PASSED`
- `S04_MODE=LOCALHOST_READY`
- marker found
- API service active
- API listener `127.0.0.1:8080`
- liveness and readiness healthy
- Caddy active with unchanged SHA

Do not reuse the older `b4803e27...` bundle.

## Exact next action

Run the independent S04 postflight described in `ops/evidence/S04_RECOVERY_20260828T001418Z.md`. Do **not** mark S04 PASSED and do **not** bootstrap the real Owner until that postflight passes.