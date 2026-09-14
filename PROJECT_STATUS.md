# PVNaive — Canonical Project Status

Last updated: 2026-09-14 (Asia/Tehran)

## Current verified GitHub truth

- Exact `main`: `722cc905464e2f581520b6350db90eec77993578`.
- Main CI `34882541754`: SUCCESS.
- R5 owner pool-manager UI is on main at `f71cdc662b38a3090132af1302be5fc2b84833bc`.
- R5 sibling mTLS pull listener is on main at `d02c18a326c480661ce5ac96ca3b9c9cfd612041`.
- R6 gated cover-site flip wiring is on main at `ed0bf0f5ae57fcf43875dddd6a32041ec8e373ca`.
- R8 live-chart frontend work is active in issue #116 / draft PR #117, exact head
  `3671a5d785e72802523f2f5194498b2a3cdafc55`. Local web verification is green
  (22 test files / 105 tests, production build, `git diff --check`). Exact-head GitHub
  all three exact-head gates are green: CI `34885703421`, Exact Accounting `34885703511`, Pinned Forwardproxy `34885703413`. Independent review remains pending.
- Task13 PR #108 and Karing PR #101 remain DRAFT and require their real acceptance gates before merge.

## Production truth ceiling

- Latest persistent verified Production checkpoint records schema **33** with the Master Upgrade
  Pack / `repo-fin2` generation live, healthy panel/API/SSE/real-customer CONNECT/accounting
  postflight, forward-only migrations 0031/0032/0033, and retained rollback/backup evidence.
- Fresh remote inventory does not expose an identifiable `PVNaive-Production-Primary`.
  The only online remote registration is not the PVNaive Production host and must not be used
  for Production actions.
- Therefore current container/image identity, migration ledger, encrypted-backup freshness,
  disk headroom and rollback snapshot are not freshly asserted. No Production mutation is allowed
  until trusted Primary reconnect + read-only audit + backup/rollback gates.

## Active roadmap lanes

1. **R8 #116/#117:** finish exact-head CI + independent review, then integrate only if every gate is green.
2. **R5 enablement (#113/#114):** code is on main; real disposable multi-node mTLS pull E2E and Production enablement remain gated.
3. **R6-FLIP #115:** code is on main and default-OFF; promotion requires trusted audit, fresh encrypted backup, independent rollback snapshot and postflight.
4. **#100 Production lane:** read-only audit first when the real Primary reconnects.
5. **#108/#101:** keep DRAFT until real pinned-Caddy HTTP/1.1+HTTP/2 and real Karing acceptance respectively exist.

## Safety invariants

- Accounting/session/quota semantics are canonical truth and must not be weakened by telemetry/UI work.
- Never trust XFF/Forwarded/client headers for authoritative node/session identity; fleet pull identity is TLS-client-cert-only.
- Applied migrations are immutable; future DB changes are forward-only and ledger-checked.
- Unknown/unavailable telemetry stays Unknown; UI must not convert missing data into zero or fabricated health.
- Production sequence: exact-head CI → disposable rehearsal → trusted read-only audit → fresh encrypted backup + independent rollback snapshot → staged promotion → postflight → retain rollback.

## Coordinator checkpoint

Concrete progress this cycle: verified green main `722cc905`; reconciled R5 UI/pull and R6 flip as code-complete but not Production-enabled; opened #116 and implemented the first R8 SSE dashboard slice in draft PR #117 with RED-first tests, bounded drop-oldest history, truthful Unknown gaps and last-valid-sample retention. Production remained mutation-free because the trusted Primary is not connected.
