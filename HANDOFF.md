# PVNaive Handoff

Checkpoint: 2026-09-13 03:41 Asia/Tehran

- Verified pre-docs `main`: `7ab4d8a8bbb5af7ecef1135743bda40ef7bfa472`; combined status pending with zero published statuses and no commit-specific workflow runs. Do not call the docs lineage CI-green without observed runs.
- Last validated merged runtime integration remains Task16/schema21 PR #81 → `7efa359ccc5745c548cda9590bc5c516e9d5aa9e`.
- Task13 PR #64 remains OPEN/DRAFT/non-mergeable at `3fc14825e1b164bad558decaef47f56b792e81af`; current-main reconstruction and fresh real HTTP/1.1 + HTTP/2 pinned-Caddy kill/accounting rehearsal remain required.
- Karing legacy PR #4 remains OPEN/DRAFT/non-mergeable at `2501e39dc39e14063b6a501bc96b77bbfcae7384`.
- Karing current-main reconstruction is now draft PR #101. RED exact commit `95768324b5f5dd43ddaf153c0b115678fa24a49e` failed the web `npm test` gate as intended after adding only the missing-builder regression test. GREEN exact commit `6168c8445ce5b9358c9cbae12be98951d3153845` passes web tests and web build with the minimal `buildKaringSingBoxProfile` implementation. Do not merge: full exact-head workflows, current-main UI wiring, and independent real Karing smoke are still required.
- Security/accounting issue #99 has no new independent exact-main receipt.
- Production issue #100 has no new connected read-only audit receipt; deployed SHA/schema, services/listeners, Caddy state, backup freshness and rollback readiness remain unverified this cycle.
- Historical `AGENT_HANDOFF.md` and `ops/DEPLOYMENT_PROGRESS.md` are 2026-08-27 S04-era evidence only where they conflict with current canonical files/fresh GitHub state.

Execution order: finish PR #101 safely and independently; advance Task13 and #99 in parallel; keep Production read-only. Promotion remains read-only audit → encrypted backup → independent rollback snapshot → exact SHA lock → staged promotion → postflight → retained rollback.
