# PVNaive — Customer Product UI Activation

AGENT: Lead Engineer / Customer Product UI Integration
TASK-ID: PVN-044 (supporting PVN-041, PVN-042, WS1 accounting read-model integration)
GOAL: Expose already-merged customer-product and exact-accounting capabilities through the protected panel without fabricating unsupported controls.
FILES: `web/src/*` customer/product panel files; narrowly scoped customer product read-model/API files if required; canonical status/worklog docs after verification.
DEPENDENCIES: `main@01600d66bd159c6c9960574ff51120e4e373e01a`; merged WS1 exact accounting/hard quota (PR #17 and production follow-ups); merged WS2 customer product management (PR #22); `OWNER_REQUIREMENTS.md`.

## Constraints

- Keep raw Naive Runtime management Owner-only and separate from business/customer operations.
- Read-only QR/subscription/details actions must remain read-only.
- Password rotation and subscription reissue remain explicit separate actions.
- Use WS2 server-side search/filter/sort/pagination and bulk preview/execute rather than client-side mutation loops.
- Show exact usage/presence only where WS1 provides evidence-backed state; show unavailable/unknown for unsupported or incomplete state.
- Do not expose future route-manifest entries whose handlers are not actually ready.
- No production/server/Caddy mutation is part of this work unit.

## Verification workflow

1. Add focused web/API contract tests first and observe RED.
2. Implement the minimum clients/views/read-model wiring to make those tests GREEN.
3. Run full Web tests/build and relevant Go tests/vet through CI on the feature branch.
4. Review diff/security/authorization implications.
5. Update canonical status/worklog documents with evidence from the exact verified HEAD.
