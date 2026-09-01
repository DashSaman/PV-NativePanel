# Task15 schema20 candidate verification — 2026-09-01

Base GitHub main: `fe6a9fea76fa48577fb8063bb246563f2696846b`.

Local candidate commit on the persistent worker: `2b175991f1d5628dc084f4ffddfea6b63d960bf8` (`lead/task15-unique-ip-schema20-20260901`), with parent exactly equal to the base main above.

Status: **locally verified for Go/Web/static gates; NOT YET PUBLISHED/MERGED/DEPLOYED and NOT YET PostgreSQL18-verified.**

Fresh verified gates on the exact-current-main candidate:

- `git diff --check` — PASS.
- `go vet ./...` with the repository Go 1.25 toolchain — PASS.
- `go test ./...` — PASS.
- `go test -race ./internal/customer ./internal/httpapi ./internal/telemetry` — PASS.
- Web `vitest`: 18 test files / 61 tests — PASS.
- Web production build — PASS.
- `bash -n tests/db/unique_ip_limit_migration_test.sh` — PASS.

The schema20 candidate sources peer identity only from trusted Caddy `RemoteAddr`, propagates it through the accounting event, and enforces the limit against canonical `direct_naive_accounting_session_peers` / active accounting state. It never uses `Forwarded` or `X-Forwarded-For`. The admission wrapper locks the ServiceTerm row before the uniqueness decision, delegates accepted accounting to schema19, and records a trusted peer only when the underlying new session is actually accepted.

During review, the PostgreSQL concurrency test was corrected so it does not assume a deterministic lock winner. It now requires exactly one accepted caller and one `unique_ip_limit` rejection, then proves that the persisted session is the same caller that actually won the race.

The persistent worker's local PostgreSQL is 14.24. It cannot execute the repository's modern migration baseline because `security_invoker` is unsupported there. This is an environment limitation and is **not** accepted as PostgreSQL18 evidence. The candidate's CI workflow includes `tests/db/unique_ip_limit_migration_test.sh`; authoritative schema20 migration/concurrency proof must run on the repository's PostgreSQL18 CI environment after publication.

Publication blocker: the worker can fetch GitHub but HTTPS `git push` has no credential (`could not read Username for https://github.com`). Therefore this local commit is not repository truth yet. It must be published through authenticated Git/GitHub Git-data transport and pass exact-head CI before merge consideration.

Production was not touched. `pv-primary` remains connected at the SentinelX control plane but command execution is blocked by the current one-active-host plan while three hosts are connected. No schema20 deployment is permitted without fresh Production read-only audit, backup/rollback state, exact-head green gates and postflight proof.
