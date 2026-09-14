# Continue Here — PVNaive

Verified checkpoint: 2026-09-14 04:11 Asia/Tehran

Current repaired pre-documentation `main`: `21687777e6dd51ccf50650d4360b3c31c05421ad`; push CI `34793340297` SUCCESS. PR #111 repaired schema29 cover migration safety/health gates; Go, PostgreSQL18 migration+health+backup/restore, web and full rehearsal passed. Renderer/build provenance and migration-lineage blockers remain closed by prior recorded evidence; schema29 is now canonical repository head.

Active work:
- **Task13 PR #108**: still DRAFT. Prior exact head `f7d8dd5aa8f33b1bc09e3f19bd26ffb219e650d9` had green old-head CI/accounting/pinned-forwardproxy gates but is stale against repaired runtime-bearing main. Worker 3 must reconstruct/reconcile onto exact latest main and rerun exact-head gates. Worker 2 must then supply the REAL pinned-Caddy HTTP/1.1 + HTTP/2 proof: target-only kill, sibling survival, forged tuple rejection, repeat idempotency, credential survival, unchanged Caddy lifecycle, exactly-once final accounting. Do not merge from stale-head evidence.
- **Karing PR #101**: still DRAFT/non-mergeable at `216d53670066033403fe95f61b0402bb710186a3`; requires disposable real Karing import → parse → CONNECT → cleanup/revoke evidence, then reconstruction on latest main and rerun of exact-head repository gates.
- **R1 / STEER-001 #109**: activated now that main repair is green. Next migration must be >=0030; never rewrite/reuse 0029. Worker 3 = RED-first trusted TCP_INFO sampling from authoritative socket/RemoteAddr state; Worker 2 = RED-first DB ingest/replay/idempotency with network telemetry separated from exact byte-accounting truth; Worker 1 = independent schema/security/CI review; Worker 4 = later E2E.
- **Production #100**: latest recorded external probe keeps `45.141.148.59.nip.io` live/ready/panel healthy. `namir.softarg.ir` remains within the documented Let's Encrypt duplicate-certificate retry window through 2026-09-15 03:17:36 UTC; do not restart/recreate to force issuance.
- **Primary audit**: still required because Production Primary is not connected; capture deployed image/SHA, migration ledger/schema, services/readiness/listeners, Caddy identity/restarts, session-control socket permissions, credential reconcile count without secrets, disk, encrypted backup freshness and independent rollback snapshot read-only before any next runtime deploy.

Current remote executor inventory has no online PVNaive worker. Assignments are persisted on PR #108, PR #101, issue #109 and Production issue #100 for the next connected workers.

No Production mutation unless real Task13 + Karing gates are complete and the safety sequence is ready: fresh Primary audit → encrypted backup → independent rollback snapshot → exact deploy SHA lock → staged promotion → health/postflight → retain rollback.
