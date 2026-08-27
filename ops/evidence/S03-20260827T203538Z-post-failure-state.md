# S03 post-failure retry state

Observed on target `testAmir5-3` at `2026-08-27T20:35:38Z` after the failed S03 attempt.

- Running kernel: `7.0.0-30-generic`
- `/opt/pvnaive/S03_DATABASE.json`: missing
- `/var/lib/pvnaive/S03_POSTGRES_CLUSTER_OWNER`: present, mode `0640`, owner `root:pvnaive`
- Marker contents:
  - `stage=S03-DATABASE`
  - `host=testAmir5-3`
  - `purpose=pvnaive-dedicated-postgresql`
- Packages `postgresql`, `postgresql-client`, `postgresql-contrib`, `age`: not installed
- PostgreSQL commands / cluster tools: not installed
- PostgreSQL clusters: none / tool not installed
- Database secrets, backup keys, health units, and `/opt/pvnaive/db/current`: absent
- No S03 prechange/database backup was created; only the prior S02 backup remains
- TCP listeners remain on 22/80/443 only; no 5432 listener
- `caddy-naive.service`: active
- Caddyfile SHA-256 unchanged: `101884de2dd11cb9d276df8e72cd068bed50e4ec6eb4ebb477184dda7a86e8b1`
- Previous verified bundle remains `/root/pvnaive-s03-95a2689-verified.tar.gz` with SHA-256 `c3b160711ea74ec05ad213dbcf7540befc173aa06f075c8e1301ddb872dcf302`

Conclusion: rollback left a clean retry state. The durable cluster ownership marker is intentionally retained for S03 retry provenance. S03 remains `NEXT`; S04 remains blocked until a real `S03_RESULT=PASSED`.
