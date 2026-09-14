# Continue Here — PVNaive

Verified checkpoint: 2026-09-15 (Asia/Tehran)

## GitHub truth
- Exact `main`: `ed0bf0f` — R5-UI (`f71cdc6`), R5-PULL-001 (`d02c18a`), R6-FLIP wiring (`ed0bf0f`).
- Exact-head CI on `f71cdc6` and `d02c18a`: SUCCESS (all five jobs). `ed0bf0f` was in progress at
  checkpoint write time — re-verify, then quote as green only if SUCCESS.
- The complete R-program chain (R1..R8 slices) is on main. The PAT push blocker is resolved
  owner-side; the superseded token ending `...Z7UW` should be revoked.

## Production truth
- Persistent verified Production checkpoint: schema 33, Master Upgrade Pack / repo-fin2 generation
  live, healthy panel/API/SSE/real-customer CONNECT/accounting postflight, retained backup/rollback
  evidence.
- New on main is NOT deployed yet: the R5 mTLS pull listener and the R6 cover flip are wired but
  disabled by default (`PVNAIVE_FLEET_PULL_*` unset, `PVNAIVE_COVERD_ENABLED` unset). Enablement
  is a separate gated operator action.

## Execute next
1. Watch CI on `ed0bf0f`; if any job fails, fix forward on main.
2. R5 enablement: set `PVNAIVE_FLEET_PULL_LISTEN/_CERT_FILE/_KEY_FILE/_CLIENT_CA_FILE` on the
   primary (all-or-none), enroll a sibling, publish a revision, pull with a cert-identity agent,
   verify heartbeat + drift on `#/pool`.
3. R6-FLIP: follow ops/caddy/COVERD_FLIP.md exactly — pre-gates (trusted audit, fresh encrypted
   backup + rollback snapshot), loopback postflight, Caddy snippet, public postflight (CAMO checks),
   single-move rollback.
4. R8: live-charts UI wiring on the SSE stream.
5. Tomorrow: namir.softarg.ir flip after the LE window (~03:21–03:26 UTC): update
   `PVNAIVE_DOMAIN` + `PVNAIVE_NAIVE_PUBLIC_HOST`, recreate, E2E, remind customers to refresh
   subscriptions.
6. #108/#101: keep DRAFT; reconcile only with real acceptance hosts.

## Invariants
- Preserve exact accounting, session, quota and credential semantics.
- No XFF/Forwarded/client-header authority for trusted identity — fleet pull identity is
  TLS-client-cert-only.
- No fabricated health/telemetry and no secret-bearing evidence.
- Forward-only migrations; never rewrite applied history.
- Promotion order: exact-head CI → disposable rehearsal → trusted Production read-only audit →
  fresh encrypted backup + rollback snapshot → staged promotion → postflight → retain rollback.

Concrete progress in the latest cycle: three feature slices pushed (owner pool-manager UI, mTLS
pull listener, gated cover flip wiring + runbook), CI green on the first two, docs reconciled.
