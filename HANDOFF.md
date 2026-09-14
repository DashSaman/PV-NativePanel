# PVNaive Handoff

Checkpoint: 2026-09-14 22:xx Asia/Tehran

## Verified baseline
- Exact main: `722cc905464e2f581520b6350db90eec77993578`.
- Main CI: `34882541754` SUCCESS.
- R5-UI `f71cdc6`, R5-PULL `d02c18a`, R6-FLIP wiring `ed0bf0f` are integrated on main.
- Production truth ceiling remains the persistent schema-33 / repo-fin2 checkpoint; no fresh Primary shell audit exists in this cycle.

## Production blocker
- No identifiable `PVNaive-Production-Primary` is connected.
- The online remote registration is not the PVNaive Production host; do not deploy or audit Production through it.
- On trusted Primary reconnect: read-only verify host identity, deployed image/SHA, schema/migration ledger, services/listeners/Caddy, encrypted backups, disk headroom and rollback snapshots before any mutation.

## Active integration work
- #116 / draft PR #117: R8 dashboard SSE wiring, exact head `3671a5d785e72802523f2f5194498b2a3cdafc55`.
- Local exact-head web gates: 22 files / 105 tests PASS; `npm run build` PASS; `git diff --check` PASS.
- GitHub exact-head: WS1 Exact Accounting SUCCESS; CI and WS1 Pinned Forwardproxy still running at checkpoint time. Do not merge early.
- #108 Task13 remains DRAFT pending real pinned-Caddy HTTP/1.1 + HTTP/2 target-only kill/accounting proof.
- #101 Karing remains DRAFT pending real disposable Karing import → parse → CONNECT → cleanup/revoke proof.

## Worker queue
- Worker 2: #116/#117 model/UI implementation completed to current draft head; remediate only verified CI/review defects.
- Worker 1: independent #117 honesty/accessibility review; also review R5/R6 security boundaries.
- Worker 3: #117 SSE contract/security regression; R5-PULL integration review.
- Worker 4: browser/perf E2E for #117 when a suitable host exists; R5 multi-node and R6 flip rehearsals remain queued.
- Coordinator: integrate only exact-head validated work and preserve Production backup/rollback gates.

## Safety gates
- Missing telemetry is Unknown, not zero.
- Exact accounting/session/quota truth is untouched by R8 visualization.
- Applied migrations are immutable and future migrations are forward-only.
- Production promotion order remains CI → disposable rehearsal → trusted audit → fresh encrypted backup + rollback snapshot → staged promotion → postflight → retained rollback.
