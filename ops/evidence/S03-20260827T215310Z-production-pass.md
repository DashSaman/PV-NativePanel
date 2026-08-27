# S03 production execution evidence — 2026-08-27T21:53:10Z

- Host: `testAmir5-3`
- Source commit: `02b02e0c2a0c2dbada04420b7a853adc08b000ab`
- S03 exit code: `0`
- Result: `S03_RESULT=PASSED`
- Schema version: `1`
- PostgreSQL cluster: `18/main`
- PostgreSQL port: `5432`
- Listen addresses: `127.0.0.1,::1`
- Migration SHA256: `7f66adefd8f09853db40e3160d0d7793f1bd0bcaf422a94d63d9d95a3a43a059`
- Encrypted database backup: `/var/backups/pvnaive/database/20260827T215317Z/pvnaive.dump.age`
- Restore drill: passed
- Health check: success as `pvnaive_app`
- Secret direct SELECT: denied
- Prechange backup: `/var/backups/pvnaive/20260827T215310Z-S03-pre`
- Health timer installed and health oneshot exited `0/SUCCESS`
- Caddy service active and Caddyfile SHA256 unchanged: `101884de2dd11cb9d276df8e72cd068bed50e4ec6eb4ebb477184dda7a86e8b1`
- PostgreSQL listeners observed only on `127.0.0.1:5432` and `[::1]:5432`
- SSH listeners remained on port 22
- Caddy listeners remained on ports 80/443
- Caddy restarted: false
- SSH changed: false
- Firewall changed: false

The independent postflight gate is still required before the deployment ledger is advanced from S03 to S04.
