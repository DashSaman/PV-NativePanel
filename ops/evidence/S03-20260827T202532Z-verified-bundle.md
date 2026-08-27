# S03 verified deployment bundle — 2026-08-27 20:25:32 UTC

Server: `testAmir5-3`

Source commit: `95a2689a0770e71db16dbd11bb436e9a3e6d92ab`

Source ZIP: `/root/PV-NativePanel-95a2689a0770e71db16dbd11bb436e9a3e6d92ab.zip`

Before bundle construction, all 13 S03 files were verified byte-for-byte against the Git blob SHA values recorded for commit `95a2689a0770e71db16dbd11bb436e9a3e6d92ab`. ZIP integrity passed, all expected files matched exactly, and no file failed verification.

Bundle construction result:

- `SOURCE_EXTRACTION=PASSED`
- `VERIFIED_FILES=13`
- migration SHA256 manifest: PASSED for `0001_initial.down.sql` and `0001_initial.up.sql`
- shell syntax: PASSED
- file modes: PASSED
- inventory count: 13
- bundle path: `/root/pvnaive-s03-95a2689-verified.tar.gz`
- bundle size: `20525` bytes
- bundle SHA-256: `c3b160711ea74ec05ad213dbcf7540befc173aa06f075c8e1301ddb872dcf302`
- `BUNDLE_FILES=13`
- `S03_BUNDLE_BUILD=PASSED`

The previously documented archive `pvnaive-s03-95a2689.tar.gz` with SHA-256 `9decbd705f548160343bdc41894b66ece86d53634e7fbf3719bac02f09be2b47` was not available on the target server and is not the deployment artifact being used now. The new archive differs at the archive-container level because it was rebuilt with explicit deterministic tar metadata, while its 13 payload files were independently verified against the exact Git blobs from the same source commit.

No PostgreSQL package installation, service restart, Caddy change, SSH change, firewall change, database migration, or S03 execution occurred during this step. `S03-DATABASE` remains `NEXT` until a real Stage run returns `S03_RESULT=PASSED`.
