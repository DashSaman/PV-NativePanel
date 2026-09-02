# PVNaive — Canonical Project Status

Last updated: 2026-09-02 (automation verification)

Exact GitHub refs and fresh worker/Production observations override historical reports. No feature is DONE from partial or stale evidence.

## Repository truth

- Repository: `DashSaman/PV-NativePanel`.
- Current `main`: `0b921abe9b2bd1d827023f494fda11a407fe34d3`.
- Task13: draft PR #64, head `3fc14825e1b164bad558decaef47f56b792e81af`; exact-head GitHub gates remain green, but the real HTTP/1.1 + HTTP/2 rehearsal is still pending.
- Task16: draft PR #81, head `096db51ac10ab25ea530af2c18ff0ef9d6a35a27`; Schema21 TDD and Exact Accounting are green; repository-wide CI is red in the PostgreSQL database job; Pinned Forwardproxy was re-run after a transient upstream failure.
- Current CI failure evidence: generic fixture still expects `schema version=20` after migration to schema21.
- Production remains on Task15/schema20; no Task13 or schema21 code is deployed.

## Safety invariants

Never fabricate usage, online, IP, or session history. Production mutation requires fresh encrypted backup, rollback state, exact artifact provenance, intended migrations only, and postflight verification. Do not use Production as a development/test lane.

## Worker / orchestration

- `pv-primary`: Production-only lane; fresh probe was not completed this cycle because SentinelX routing was unavailable.
- `TrPaqet`: Task13 exact-head live rehearsal lane; still blocked by missing `jq`/tooling when the rehearsal preflight is attempted.
- `pv-worker-main`: Task16 repository-wide PostgreSQL18 lane.
- SentinelX currently reports three connected hosts; active-host execution remains capacity-constrained.

## Immediate next actions

1. Keep PR #64 draft until the real HTTP/1.1 + HTTP/2 proof passes.
2. Finish the generic schema21 fixture correction on PR #81, then rerun all exact-head gates.
3. Do not merge or deploy until every required gate is green and Production backup/rollback prerequisites are freshly verified.
4. Keep historical `AGENT_HANDOFF.md` and `ops/DEPLOYMENT_PROGRESS.md` as evidence only; they remain S04-era.
