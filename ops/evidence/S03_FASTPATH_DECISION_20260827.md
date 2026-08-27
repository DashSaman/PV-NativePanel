# S03 pre-deployment migration fastpath decision

S03 has never reached `PASSED` on the target host and no persistent production `pvnaive` database exists. Therefore the temporary `0002_pgcrypto_schema_hardening` branch experiment is not the canonical deployment path.

For the first production deployment, the pgcrypto/search_path fix will be folded into `0001_initial` so the initial schema remains version 1, rollback semantics stay simple, and no artificial post-release migration is introduced before the product has ever shipped.

The experimental branch remains only as debugging evidence and MUST NOT be deployed.
