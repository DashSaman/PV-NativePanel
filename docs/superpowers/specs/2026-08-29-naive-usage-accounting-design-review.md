# Self-review — exact Naive usage accounting design

Reviewed: `docs/superpowers/specs/2026-08-29-naive-usage-accounting-design.md`

Latest proof addendum: `docs/superpowers/specs/2026-08-29-naive-usage-accounting-forwardproxy-addendum.md`

Result: **PHASE-A PROOF PASS; separate-handler architecture rejected; fallback addendum requires explicit Owner approval before implementation.**

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

## Implementation-time proof gate — completed

Pinned upstream commit:

`d62c80d3dd2c706b6b87579844d2397bddd18317`

CI run `33227417609` produced:

```text
FORWARDPROXY_AUTH_LINE=256
FORWARDPROXY_TARGET_DIAL_LINE=333
FORWARDPROXY_HTTP1_STREAM_LINE=347
FORWARDPROXY_HTTP23_STREAM_LINE=352
PINNED_FORWARDPROXY_BOUNDARY_PROOF=PASSED
```

The pinned code owns the remote `targetConn` inside `forward_proxy`. A preceding handler can observe request-body reads but cannot know how many upload bytes the remote connection actually accepted; a handler after `forward_proxy` does not own the authenticated CONNECT stream. `internal/accountingboundary/boundary_test.go` independently proves the partial-write mismatch and passes in the same full CI checkpoint.

Therefore a separate PVNaive handler cannot satisfy the design rule “count only bytes successfully transferred” for both directions.

## Fallback design review

The addendum narrows the fallback to a patch against the exact pinned forwardproxy commit, preserves Naive wire behavior, uses trusted post-auth identity plus operator-rendered Runtime UUID metadata, instruments the actual forwarding write primitive, excludes H2 padding from customer usage, and keeps the quota-depleted credential intact.

A self-review correction tightens the original overshoot wording: use at most **16 KiB payload per directional tracked-stream iteration**, giving at most 32 KiB aggregate unacknowledged payload per active connection when one upload and one download iteration are simultaneously outstanding.

The addendum also resolves persistent QR/Subscription retrieval without weakening token secrecy: raw Subscription token material is AES-256-GCM encrypted at rest while the public resolver continues to use only SHA-256 token hashes. Pre-schema-7 tokens require one explicit rotation because their raw token cannot be recovered.

## Remaining hard gate

No schema-7 code, Runtime Agent accounting endpoint, forwardproxy patch, custom Caddy build, or production change may begin until the Owner explicitly approves `2026-08-29-naive-usage-accounting-forwardproxy-addendum.md`.
