# Task13 pre-publication verification — 2026-09-01

Base main: `937865979c97325d3b7407fc4199b2e8ff933db2` (Merge PR #48).

Status: **GREEN locally on exact latest main, NOT YET PUBLISHED/MERGED/DEPLOYED.**

Fresh independent verification was run in an isolated worktree after applying the reconciled Task13 candidate to exact `origin/main`.

Verified gates:

- `git diff --check` — PASS
- `go vet ./...` — PASS
- `go test ./... -count=1` — PASS
- `go test -race ./internal/sessionkill/... ./internal/sessioncontrol/... ./internal/httpapi/... ./tests/rehearsal/... -count=1` — PASS
- `tests/accounting/pinned_forwardproxy_boundary_test.sh` — `PINNED_FORWARDPROXY_BOUNDARY_PROOF=PASSED`
- `tests/stages/WS1_reproducible_caddy_build_contract_test.sh` — `WS1_REPRODUCIBLE_CADDY_BUILD_CONTRACT=PASSED`
- Web: 18 test files / 63 tests PASS
- Web production build — PASS

The candidate preserves BUG-002 response buffering/commit-before-success semantics and schema19/Task14 behavior. It adds an exact one-live-session kill boundary using the full runtime credential/node/boot/session tuple, a local Unix control socket, API RBAC/IDOR checks, UI action, and focused race/rehearsal tests. It does not revoke an entire credential and does not require Caddy reload/restart for an individual kill.

Additional prior real-protocol rehearsal evidence on this worker confirmed both HTTP/1.1 and HTTP/2 CONNECT tunnels, exact kill of one HTTP/1.1 session while the sibling HTTP/2 session stayed alive, exactly one final accounting close for the killed session, unchanged Caddy PID, and no reload/restart.

Publication blocker: the accessible worker can fetch GitHub but local HTTPS `git push` has no credential (`could not read Username for https://github.com`). The GitHub connector has write access; code publication still needs the exact verified 18-file tree to be created through connector Git-data operations or another authenticated Git transport. Do not mark Task13 DONE until exact tree is published, CI is green, and Production rollout/postflight is proven.

Production blocker: `pv-primary` is connected in SentinelX but unavailable on the current Free plan because only one of three connected hosts may be active. Therefore no fresh Production audit or mutation was performed in this run.

Parallel lanes at this checkpoint:

- Task15 schema20 unique-IP limit: active redesign on trusted Caddy `RemoteAddr`/session-peer boundary; previous candidate remains rejected until deterministic PostgreSQL18 race proof, trusted-boundary negative tests, rollback/RLS and full propagation are green.
- Task16 schema21 bounded session/IP history: active local candidate; must remain PENDING until schema20 lands contiguously, with 30-day bounded retention, exact finalization sync, tenant/RLS isolation, maintenance-only purge and no fabricated legacy history.

Safety rule: do not merge/deploy from this evidence file alone. Source code must be published as the exact fresh-verified tree and pass exact-head CI first.
