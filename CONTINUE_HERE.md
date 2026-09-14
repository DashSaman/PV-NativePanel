# Continue Here — PVNaive

Verified checkpoint: 2026-09-14 (Asia/Tehran)

## GitHub truth
- Exact `main`: `722cc905464e2f581520b6350db90eec77993578`; CI `34882541754` SUCCESS.
- R5 owner pool UI (`f71cdc6`), sibling mTLS pull (`d02c18a`) and R6 gated cover flip wiring (`ed0bf0f`) are on main.
- R8 live dashboard slice is draft PR #117 on exact head `3671a5d785e72802523f2f5194498b2a3cdafc55`.
- #117 local web gates and all three exact-head GitHub gates are green (CI `34885703421`, Exact Accounting `34885703511`, Pinned Forwardproxy `34885703413`); independent review remains before any merge decision.

## Production truth
- Persistent verified ceiling: schema 33 / repo-fin2 with recorded healthy panel/API/SSE/real-customer CONNECT/accounting postflight and retained rollback/backup evidence.
- No trusted Production Primary is currently connected, so do not infer fresh image/schema/backup/disk/rollback state and do not mutate Production.

## Execute next
1. Obtain a genuinely independent diff/contract review for #117. All exact-head workflows are already green; merge only with expected-head guard after that review is real and no blocking finding remains.
2. If #117 is integrated, refresh canonical docs and continue R8 toward per-node/per-user charts, reconciliation and performance/RBAC gates.
3. R5: run real disposable multi-node enroll → publish → mTLS pull → heartbeat/drift E2E before Production enablement.
4. R6-FLIP: remain default-OFF until trusted Production audit + fresh encrypted backup + independent rollback snapshot + runbook postflight are available.
5. #108/#101: keep DRAFT until their real acceptance hosts exist; historical CI is not a substitute.
6. Production Primary reconnect: first action is read-only identity/SHA/schema/service/Caddy/backup/disk/rollback audit.

## Invariants
- Preserve exact accounting, session, quota and credential semantics.
- TLS client certificate is authoritative for fleet-pull identity; never trust forwarded/client headers.
- Missing data remains Unknown; never smooth or fabricate source data.
- Forward-only migrations; never rewrite applied history.
- Production promotion: exact-head CI → disposable rehearsal → trusted audit → fresh encrypted backup + rollback snapshot → staged promotion → postflight → retain rollback.
