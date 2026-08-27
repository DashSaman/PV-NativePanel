# S03 pre-deployment fastpath plan

The target host has never reached `S03_RESULT=PASSED` and no persistent production `pvnaive` database exists.

Therefore the canonical first deployment will keep schema version `1` and fold the pgcrypto/search_path correction directly into `0001_initial` rather than introducing an artificial `0002` before first production release.

Required gates before the next target-host mutation:

1. PostgreSQL 18 CI passes migration, RLS, health, encrypted backup, restore and rollback tests.
2. The S03 bundle is rebuilt from the new canonical commit and checksum-verified.
3. A disposable rehearsal on `testAmir5-3` passes without creating the real `pvnaive` database or changing Caddy, SSH or firewall state.
4. Only then may the real `S03-database.sh` stage run.

The experimental `s03-pgcrypto-schema-fix` branch is debugging evidence only and must not be deployed.
