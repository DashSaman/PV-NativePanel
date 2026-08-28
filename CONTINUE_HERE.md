# CONTINUE HERE — PVNaive

If a Chat/Agent session was interrupted, start here.

## Read in this order

1. `ops/S04_LIVE_STATE.md` — authoritative latest live state and exact next action.
2. `ops/evidence/S04-20260828T002041Z-postflight-core-pass-health-release-blocked.md` — newest independent postflight evidence.
3. `AGENT_HANDOFF.md` — broader project history and non-negotiable constraints.
4. `ops/DEPLOYMENT_PROGRESS.md` — official stage ledger.
5. Active implementation branch: `s04-auth`; open draft PR: `#2`.

## Current one-line state

`S00-S03=PASSED`; `S04-AUTH=IN PROGRESS`. The S04 API/auth/web Stage is installed and healthy on `testAmir5-3`; schema is 2, API is active only on `127.0.0.1:8080`, encrypted rollback backup is valid, Caddy/SSH/firewall are unchanged, and the independent postflight core passed. **The only remaining blocker is `DB_TIMER_S04_AWARE=false`: the periodic DB health timer still selects the immutable S03 DB tooling release.**

Do NOT mark S04 PASSED and do NOT bootstrap Owner until this is fixed and postflight passes.

## Live S04 artifact already installed

- source commit: `11c54dc1faae99a1491c750b30db9faa44a0c3ae`
- CI run: `33128780602`
- artifact ID: `9669443464`
- archive: `PVNaive-S04-11c54dc1faae.tar.gz`
- SHA-256: `52acdde2bff6777abeb31c081c86e018d1e7f5f0cdb39f8eb4151efbba2820fc`

## Current repository fix for the only blocker

- branch: `s04-auth`
- PR: `#2`
- helper: `scripts/db/promote-release.sh`
- regression test: `tests/stages/S04_db_release_promotion_test.sh`
- expected new immutable DB tooling release: `0002-84bb735877d5`
- Stage wiring commit: `708a4e7fd71011e5b21f136ae7305612f295a258`
- one-shot workflow removed in clean user commit: `20ed774d06969a3f4c301fd6072a4db83fcffcca`
- clean-head CI run to verify before live mutation: `33130012929`

## Exact next action

Verify the clean-head CI is fully green. Then atomically create/select DB tooling release `0002-84bb735877d5` on `testAmir5-3`, preserving the old S03 release. Require the systemd periodic DB health service to pass using the S04-aware script, then rerun independent S04 postflight. Only after `S04_POSTFLIGHT=PASSED` may the real Owner be bootstrapped.

Never reuse the older `b4803e27...` bundle.
