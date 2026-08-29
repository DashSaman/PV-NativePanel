# Self-review — exact Naive usage accounting design

Reviewed: `docs/superpowers/specs/2026-08-29-naive-usage-accounting-design.md`

Result: PASS with one implementation-time proof gate.

Checks:

- Scope is explicit: exact per-user payload accounting, quota enforcement, persistent Subscription/QR UI.
- Existing credentials remain invariant; quota depletion is not credential revocation.
- Historical usage is explicitly not fabricated.
- Source of truth is server-side; browser never supplies usage.
- Idempotency and concurrent increments are covered.
- Fail-closed policy is stated for finite-quota users.
- Bounded active-tunnel overshoot is stated and testable.
- Caddy binary replacement is acknowledged as requiring a controlled restart and rollback, unlike previous S05 web/API-only upgrades.
- PostgreSQL schema 7 and rollback are included.
- Production rollout is canary-first and capability remains disabled until deterministic traffic proof passes.
- QR/subscription viewing is explicitly non-rotating.
- No package/OS upgrade is included.

Implementation-time proof gate:

The preferred separate Caddy handler must be proven capable of observing exact CONNECT tunnel bytes and trusted authenticated identity. If Caddy handler composition cannot expose the tunnel stream, implementation must stop before modifying upstream `forward_proxy`; the design must be revised to document the minimal pinned wrapper/fork approach and its maintenance/supply-chain implications.
