# CONTINUE HERE — PVNaive

If a Chat/Agent session was interrupted, start here.

## Read in this order

1. `ops/S04_LIVE_STATE.md` — authoritative latest live state and safety invariants.
2. `ops/evidence/S04-20260828T002041Z-postflight-core-pass-health-release-blocked.md` — newest independent postflight plus green TDD fix evidence.
3. `AGENT_HANDOFF.md` — broader project history and non-negotiable constraints.
4. `ops/DEPLOYMENT_PROGRESS.md` — official stage ledger.
5. Active implementation branch: `s04-auth`; open draft PR: `#2`.

## Current one-line state

`S00-S03=PASSED`; `S04-AUTH=IN PROGRESS`. The S04 API/auth/web Stage is installed and healthy on `testAmir5-3`; schema is 2, API is active only on `127.0.0.1:8080`, encrypted rollback backup is valid, Caddy/SSH/firewall are unchanged, and the independent postflight core passed. **The only remaining blocker is `DB_TIMER_S04_AWARE=false`: the periodic DB health timer still selects the immutable S03 DB tooling release.**

Do NOT mark S04 PASSED and do NOT bootstrap Owner until the live DB tooling release promotion and follow-up postflight pass.

## Repository fix for the blocker is green

- branch clean head: `20ed774d06969a3f4c301fd6072a4db83fcffcca`
- final CI run: `33130012929` — SUCCESS
- all five jobs/gates passed: Go, Web, PostgreSQL18 regression gates, end-to-end rehearsal, production bundle
- helper: `scripts/db/promote-release.sh`
- regression test: `tests/stages/S04_db_release_promotion_test.sh`
- Stage wiring commit: `708a4e7fd71011e5b21f136ae7305612f295a258`
- expected live immutable DB tooling release: `0002-84bb735877d5`
- old S03 immutable release must be preserved for rollback: `0001-7f66adefd8f0`

TDD RED evidence is in the newest S04 evidence file. Compare `11c54dc1...` to `20ed774d...` confirms the only `scripts/db` addition is `promote-release.sh`; migrations and existing DB scripts are unchanged.

## Live S04 artifact already installed

- source commit: `11c54dc1faae99a1491c750b30db9faa44a0c3ae`
- CI run: `33128780602`
- artifact ID: `9669443464`
- archive: `PVNaive-S04-11c54dc1faae.tar.gz`
- SHA-256: `52acdde2bff6777abeb31c081c86e018d1e7f5f0cdb39f8eb4151efbba2820fc`

Never reuse the older `b4803e27...` bundle.

## Exact next action

Perform ONE atomic live DB tooling release promotion on `testAmir5-3`: verify schema/migration/Caddy/API state, preserve `/opt/pvnaive/db/releases/0001-7f66adefd8f0`, create exact schema-2 tooling release `0002-84bb735877d5`, atomically repoint `/opt/pvnaive/db/current`, run the periodic DB health service and require MFA-aware schema-2 health, and rollback only the symlink if any post-switch invariant fails. Then rerun independent S04 postflight. Only after `DB_TIMER_S04_AWARE=true` and `S04_POSTFLIGHT=PASSED` may the real Owner be bootstrapped.
