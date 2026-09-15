# PVNaive — Canonical Project Status

Last updated: 2026-09-15 20:43 (Asia/Tehran)

## Current verified GitHub truth

- Current verified `main`: `afb332ab9f6ced1045131b01037605496ae9264c`; exact-main GitHub CI `34997820889` is terminal SUCCESS.
- New validated work: Karing/Hiddify UA and `?family=karing|hiddify` now receive the universal base64 `naive+https` link list rather than sing-box JSON. Main documents that those importers reject the proposed `type: naive` sing-box outbound. Account guidance no longer promises a Mihomo-only PV-AUTO group to Karing.
- Independent execution-worker validation of current main: Web 23/23 files, 117/117 tests PASS; TypeScript/Vite production build PASS.
- `npm ci` reports 4 dependency audit findings (1 low, 1 moderate, 1 high, 1 critical). Treat as a dependency-security follow-up; do not apply forced upgrades without compatibility review.
- #101 is still the only open PR, DRAFT on stale head `6691392639be9fae4656a861db4a6d16580f850d`, `mergeable=false`. Its sing-box Karing-profile approach is now contradicted/superseded by current main and MUST NOT merge as-is. Real Karing import → parse → CONNECT → cleanup/revoke remains the acceptance gate before closing/superseding the lane.
- #114 basic two-node real mTLS proof remains retained; certificate overlap/rotation + old-cert retirement or explicit revocation + replay/fail-closed lifecycle proof remains open.

## Production truth ceiling

- Persistent verified Production ceiling remains schema **33** / `repo-fin2` checkpoint.
- Fresh device inventory at 20:43 shows one online execution worker `Pak-Nasheeee-haaaaaaaaa` and one stale duplicate offline. No trusted `PVNaive-Production-Primary` is connected.
- No Production deploy/migration/restart/reload/Caddy/DB/credential/backup/rollback mutation is permitted until trusted Primary reconnect + read-only audit + fresh encrypted backup + independent rollback snapshot.

## Active roadmap and worker allocation

1. W4: validate current-main base64 subscription in a real Karing client with disposable credentials; if accepted, supersede/close stale #101 rather than merging its sing-box code; then disposable R5/R6 E2E.
2. W3: #114 RED-first certificate overlap/rotation/revocation lifecycle; then STEER-006 scale/integration.
3. W1: independent client-truth + PKI/RBAC review; include current Karing/Hiddify claims and dependency-audit triage without forced upgrades.
4. W2: registry/replay monotonicity assertions; then R8 ledger/per-node/per-user truthful projections.
5. Coordinator: integrate only exact-head validated work; keep migrations forward-only and Production backup/rollback gates mandatory.

## Safety invariants

- Accounting/session/quota semantics remain canonical truth; telemetry/UI/control work cannot weaken them.
- Task13 kills only the selected session and preserves credential/sibling sessions.
- Fleet identity is TLS-client-cert authoritative; forwarded/client headers never override it.
- Applied migrations are immutable; future DB changes are forward-only and ledger-checked.
- Missing telemetry remains Unknown, never fabricated zero/health.
- Client compatibility claims require real-client evidence.
- Production promotion order: exact-head CI → disposable rehearsal → trusted read-only audit → fresh encrypted backup + independent rollback snapshot → staged promotion → postflight → retain rollback.

## Coordinator checkpoint — 2026-09-15 20:43 Asia/Tehran

- Re-inspected current main, CI, the only open PR, Production device inventory and canonical handoff.
- Reconciled `afb332ab...` only after terminal exact-head CI; independently reran Web tests/build successfully on the connected execution worker.
- Refreshed #101 with the new current-main Karing truth; stale sing-box code is not eligible for merge.
- Production remained untouched because the trusted Primary is disconnected.
## Coordinator checkpoint — 2026-09-15 22:41 Asia/Tehran

- Current verified `main`: `9891fdc650e4e65c6bdc91c5b4bd3719c507eb2c`; exact-main GitHub CI `35008842524` is terminal SUCCESS.
- Reconciled new Karing field-result fix: Karing UA / `?family=karing` now receives the direct Clash-compatible Naive profile (`type: naive`) rather than the rejected base64 `naive+https` list. Guidance explicitly excludes FlClash / Clash Meta for Android / v2rayNG for NaiveProxy.
- Independent execution-worker rerun on exact main: Web 23/23 files, 117/117 tests PASS; production build PASS. Go is not installed on this worker, so Go truth remains exact-head GitHub CI.
- #101 remains stale DRAFT and MUST NOT merge; current main supersedes both its old sing-box approach and the intermediate base64 approach. Real Karing CONNECT + cleanup/revoke remains required before closing the acceptance lane.
- Production Primary is still absent from connected-device inventory. No Production mutation is permitted; verified Production ceiling remains schema 33 / repo-fin2.
- Worker queue: W4 real-Karing validation of current direct Clash Naive profile then disposable R5/R6 E2E; W3 cert rotation/overlap + explicit revocation; W2 replay/registry/accounting monotonicity; W1 PKI/RBAC/client-truth and dependency-audit review.

