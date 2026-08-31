# Task14 concurrent-session limit — Production PASS

Date: 2026-08-31

- PR: #45
- Merge/main/deployed commit: `0645b2e4758d3cc6c197dd9dba9127e8de983d6c`
- Exact PR head: `c0c93e336e82b17912317cb981f619b8baec7fd5`
- PR-head workflows: CI PASS; WS1 Exact Accounting PASS; WS1 Pinned Forwardproxy PASS.
- Main CI: run `33419350522`, Go/Web/PostgreSQL18/full S04R/bundle all PASS.
- Release artifact SHA256: `9e44abbca08931c928024e5c2e0a06758ecc79f05e35f4d28208177bebf701ad`.
- Fresh encrypted DB/config backup: PASS and checksum-verified before migration.
- Migration: schema18 → schema19, transactional, PASS.
- Guarded R1 deploy: `PVNAIVE_R1_DEPLOY_RESULT=PASSED`; `PVNAIVE_R1_CADDY_ACTION=NONE`.
- Production postflight: PostgreSQL/Caddy/API/Runtime/Telemetry active; readiness true; Runtime/Telemetry health OK.
- Caddy binary SHA, config SHA, MainPID and NRestarts remained unchanged; NRestarts remained 0.
- Live schema contains nullable integer `service_terms.concurrency_limit` and both public schema19 ingest wrapper + private schema17 primitive.

No credentials, tokens, private keys or customer-specific data are recorded in this evidence.
