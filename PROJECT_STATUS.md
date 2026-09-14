# PVNaive — Canonical Project Status

Last updated: 2026-09-15 (Asia/Tehran)

## Current verified GitHub truth

- Exact `main`: `ed0bf0f` (`feat(r6): gated cover-site flip wiring (R6-FLIP-001)`), preceded by
  `d02c18a` (R5-PULL-001 mTLS pull listener) and `f71cdc6` (R5-UI-001 owner pool-manager UI).
- Exact-head CI on `f71cdc6` and `d02c18a` is SUCCESS (all five jobs); `ed0bf0f` CI observed
  in progress at docs-write time — re-verify before quoting as green.
- The full R-program backend chain (R1 telemetry, R2 scheduler, R3 decision-sink + renderer,
  R4 manifests, R5 registry+UI+mTLS pull, R6 coverd+flip wiring, R7 panel access, R8 SSE)
  now lives on main; the earlier push blocker (PAT scopes) is resolved owner-side.
- No promotion PR is currently eligible to merge. Two old draft branches remain materially
  behind current main and still require real acceptance before any reconstruction/merge decision:
  - Task13 PR #108: head `44db803d`, materially behind main; real pinned-Caddy HTTP/1.1 + HTTP/2
    target-only acceptance remains mandatory.
  - Karing PR #101: head `66913926`, materially behind main; real disposable Karing import →
    parse → CONNECT → cleanup/revoke acceptance remains mandatory.

## Production truth ceiling

- Latest persistent verified Production checkpoint records schema **33** with the Master Upgrade
  Pack / `repo-fin2` generation live, healthy panel/API/SSE/real-customer CONNECT/accounting
  postflight, forward-only migrations 0031/0032/0033, and retained rollback/backup evidence.
- That persistent checkpoint is the current truth ceiling for Production. New code on main
  (R5 UI/pull, R6 flip wiring) is **not yet deployed**; production enablement of the mTLS pull
  listener and the cover flip are separately gated operator actions (see ops/caddy/COVERD_FLIP.md).

## Active roadmap lanes

1. **R5 enablement:** deploy the mTLS pull listener env on the primary; run disposable multi-node
   E2E (enroll → publish → pull → heartbeat → drift) on real hosts.
2. **R6-FLIP:** production flip stays default-OFF until the gated sequence is green
   (trusted read-only audit → fresh encrypted backup + rollback snapshot → flip → postflight).
3. **R8 command center:** live-charts UI wiring on top of the live SSE stream.
4. **#100 — Production lane:** keep read-only until the expected Production Primary registration
   is re-established; first action after reconnect is the identity/SHA/schema/service/Caddy/
   backup/disk/rollback audit.
5. **#108/#101:** keep DRAFT; reconcile only when real acceptance hosts exist.

## Safety invariants

- Accounting/session/quota semantics are canonical truth and must not be weakened by steering/telemetry work.
- Never trust XFF/Forwarded/client headers for authoritative node/session identity — fleet pull
  identity is TLS-client-cert-only.
- Applied migrations are immutable; all future DB changes are forward-only and ledger-checked.
- Production promotion sequence remains: exact-head CI → disposable rehearsal → trusted read-only
  Production audit → fresh encrypted backup + independent rollback snapshot → staged promotion →
  postflight → retain rollback.
- No completion credit is given for assignment-only work; evidence = exact SHA + executable proof.

## Coordinator checkpoint

Concrete progress this cycle: full R-program chain pushed to main with exact-head CI green
(f71cdc6, d02c18a; ed0bf0f in progress); owner pool-manager UI (#/pool) with fail-closed model
tests; dedicated sibling mTLS pull listener (TLS-cert identity, preflight, fail-closed envelope
verification); gated cover flip wiring + runbook; KNOWN_ISSUES/FEATURE_MATRIX reconciled with
the resolved push blocker. Historical checkpoints remain available in Git history.