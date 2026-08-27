# S03 health-boundary failure evidence

Observed UTC run: `2026-08-27T20:55:58Z` bootstrap / `20260827T205603Z` S03.

## Passed gates before mutation

- Public bootstrap launcher download integrity passed.
- Pinned source commit: `27899b63fa676029ec4f24a30f16933515f5fe21`.
- Git source verification passed.
- 13 production S03 blob checks passed.
- Migration SHA256 checks passed.
- S03 preflight regression passed.
- APT candidate pipefail regression passed.
- Ubuntu 26.04 contract regression passed.
- APT dependency simulation passed.
- PostgreSQL 18 candidate resolved to `18.6-0ubuntu0.26.04.1`.

## Actual mutation reached

The target installed PostgreSQL 18, PostgreSQL client/common packages, `age`, and dependencies. PostgreSQL cluster `18/main` was created by the package installation.

Migration `0001` then completed successfully:

- `MIGRATION 0001=APPLIED`
- `PVNAIVE_SCHEMA_VERSION=1`
- `PVNAIVE_MIGRATION_RESULT=PASSED`

## Failure

The first application health gate then failed with:

`ERROR: application role can read the RLS signing key`

S03 reported exit 1 at line 447 and executed rollback.

## Rollback result

- `ROLLBACK=COMPLETED`
- packages intentionally retained for inspection/retry
- Caddy not restarted
- SSH unchanged
- firewall unchanged
- S03 success marker not created
- failed-run prechange backup: `/var/backups/pvnaive/20260827T205603Z-S03-pre`
- log: `/root/pvnaive-s03-20260827T205558Z.log`
- failed source workdir preserved by bootstrap: `/root/pvnaive-s03-src.20260827T205558Z.bEqNYa`

## Diagnostic finding

The old `health.sh` evaluated the signing-key privilege assertion before asserting that `current_user` was in fact the expected application role. Therefore a wrong database identity (for example `postgres`) would surface as the misleading signing-key error first.

Follow-up commit `6528cabee187f3fdd1c91a392d31f118713646b1` hardens health verification by:

- requiring `PVNAIVE_DB_USER=pvnaive_app` by default,
- refusing `PVNAIVE_RUN_AS_OS_USER` in application health,
- checking `current_user` and `session_user` before privilege assertions,
- checking loopback and `row_security` before the secret boundary,
- retaining `has_table_privilege` as a catalog assertion,
- adding an actual `SELECT ... LIMIT 0` negative permission probe so the secret is never output.

S03 remains `NEXT` until a real target run passes all health, encrypted backup, restore drill, systemd health and final invariants.
