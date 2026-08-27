# S03 Source ZIP Verification — 2026-08-27 20:19 UTC

Target host: `testAmir5-3`

Source ZIP: `/root/PV-NativePanel-95a2689a0770e71db16dbd11bb436e9a3e6d92ab.zip`

Source commit: `95a2689a0770e71db16dbd11bb436e9a3e6d92ab`

## Result

- ZIP integrity: `PASSED`
- Expected S03 files: `13`
- Verified S03 files: `13`
- Failed files: `0`
- Final gate: `S03_SOURCE_MANIFEST=PASSED`

All 13 S03-sensitive files were read directly from the ZIP and their Git blob SHA-1 values were compared against the blobs from commit `95a2689a0770e71db16dbd11bb436e9a3e6d92ab`. Every file matched byte-for-byte.

Verified paths:

- `db/migrations/0001_initial.down.sql` → `4ff634e1986720f9f283b57a93540fc2f643f6ea`
- `db/migrations/0001_initial.up.sql` → `99bcccb65b180855860e971a937756792684c776`
- `db/migrations/SHA256SUMS` → `e7cff7d3257ed95cd3207f50f8faea2779507164`
- `scripts/db/backup.sh` → `2fd5fa12eb2591a86e42fab9e406cf8204e875f1`
- `scripts/db/health.sh` → `fd6533101e4457878168591d354186957931bae9`
- `scripts/db/lib.sh` → `17c3a19092e85da873192ab8ea880757a085b805`
- `scripts/db/migrate.sh` → `dbfed23b748474e65eeb296dbb77a0d9bb656eb0`
- `scripts/db/restore.sh` → `e4ef114551ba11d7c48a16de9a4590a336495c41`
- `scripts/db/rollback.sh` → `5603aa46fb30510fa66f7ef52fba962050f3a11c`
- `scripts/stages/S03-database.sh` → `f14cc7ecab7aa07c56b0b4f90ccd582c7e039248`
- `scripts/stages/lib.sh` → `615d1ed4bc8fbaaa6ef9db2e3478e2d0ed6ff0fc`
- `ops/systemd/pvnaive-db-health.service` → `78acd260310c75c19588ba4b8c9374e139ecb403`
- `ops/systemd/pvnaive-db-health.timer` → `4c3067faf887b48cf84cde27cd69bb74a4444185`

No PostgreSQL package installation, S03 execution, Caddy change, SSH change, firewall change, or extraction to the live deployment tree occurred during this verification.

Next gate: create a fresh reproducible S03 tar.gz from these verified 13 files, validate migration checksums and shell syntax, calculate its SHA-256, and record that SHA before running S03.
