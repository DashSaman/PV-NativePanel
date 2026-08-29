# PVNaive — Owner Requirements / Customer Panel Gap Backlog

Last updated: 2026-08-29

> **OWNER-AUTHORITATIVE PRODUCT REQUIREMENTS**
>
> This file records the Owner's observed gaps and required behavior for the PVNaive customer-management panel. Agents must read this file before changing `/panel/#/customers`, customer lifecycle APIs, subscription/QR behavior, usage/accounting, or reseller/operator workflows.
>
> A backend endpoint, DB column, or hidden capability does **not** count as complete if the Owner cannot perform the action clearly from the UI.

## Current Owner-reported UX defects

The current customer UI is not yet acceptable as a Sanaei/PasarGuard-style production customer panel.

1. **No clear Edit action in the customer list.**
   - The Owner must be able to open a customer and edit service settings from an obvious `Edit` action.
   - Existing backend update capability does not make the feature complete until the UI exposes it safely.

2. **Viewing QR / existing subscription must not require creating a new link.**
   - The Owner must be able to click `View QR` at any time for the customer's currently active subscription.
   - Reading/viewing/copying a QR code must be a read-only action.
   - It must not rotate the subscription token.
   - It must not rotate the Runtime credential.
   - It must not change the password.
   - It must not change quota, expiry, first-use state, or any other service state.

3. **Password rotation must never be an accidental side effect of viewing QR/subscription.**
   - `View QR`, `Copy Subscription`, `Copy Direct Link`, and `View Details` are read-only operations.
   - `Change Password / Rotate Password` must be a separate explicit destructive/sensitive action with confirmation.
   - A password may be generated/displayed once only when a deliberate rotation/create action succeeds.

4. **No Delete Account action is visible.**
   - Customer list/details must expose a clear delete/revoke action.
   - Safe default is soft-delete / revoke business access plus Runtime credential revocation according to the established safety policy.
   - Permanent database destruction must not be the normal UI path.
   - Destructive action requires confirmation and audit evidence.

5. **Subscription reissue is not the same as viewing the existing QR.**
   - `View QR` = read current active subscription representation.
   - `Reissue Subscription` = explicit security action that revokes the previous subscription token and creates a new one.
   - Reissue must have confirmation and a warning that the old subscription URL stops working.
   - Reissue must not rotate the Runtime password unless the Owner explicitly selected a separate password-rotation action.

## Required customer actions — production UI

Every managed customer row/details page should eventually provide the following clear actions, capability-gated where required:

### Core lifecycle — P0

- Edit customer/service
- Enable / Resume
- Disable / Suspend
- Revoke customer access
- Delete account (safe soft-delete/revoke flow)
- Change username where safely supported
- Rotate/change password as an explicit separate action
- Reissue subscription as an explicit separate action
- View existing subscription
- View QR without mutation
- Copy subscription URL
- Copy direct `naive+https://...` link when available
- View customer details

### Quota and validity — P0

- Set numeric traffic quota in GB
- Unlimited traffic
- Add volume without recreating the customer
- Replace/set total volume deliberately
- Reset usage after exact accounting is proven
- Validity from creation
- Validity from first successful connection
- Fixed/manual expiry date
- Extend time by N days
- Change expiry without changing password or Runtime credential
- Unlimited/no-expiry where policy permits
- Expired/depleted state must be computed, not faked by a simple enabled boolean

### Renewal/product lifecycle — P1

- Renew service
- Next Plan
- On Hold / pending first use
- Reset strategy: no reset / daily / weekly / monthly / yearly / custom where technically valid
- Auto-renew metadata if implemented later
- Service-term history / renewal history

## Exact usage/accounting is still a hard gate

Do not fake traffic usage for UI completeness.

Required before `used`, `remaining`, `depleted`, reset usage, or hard byte-quota enforcement can be called production-ready:

- exact per-Runtime-credential upload/download accounting;
- restart/reload-safe counters;
- no double counting;
- reconciliation after process/server restart;
- deterministic credential-to-business-user binding;
- auditable reset events;
- hard quota enforcement only after the exact accounting proof passes.

Until the proof gate passes, show an explicit unavailable/capability state rather than `0 used` or invented remaining traffic.

## First successful connection requirement

`Validity from first connection` is not complete until the live pinned Naive/Caddy data path produces a trusted authenticated successful-CONNECT signal for the stable Runtime credential UUID.

The following must **not** start the service timer:

- opening the panel;
- viewing QR;
- copying subscription URL;
- fetching subscription;
- Caddy reload;
- failed authentication;
- health checks.

Only a proven successful customer connection may activate first-use validity.

## Customer list parity requirements — Sanaei / PasarGuard direction

### P1 — table usability

- Search by username
- Search by user/customer ID
- Search by subscription identifier/prefix where safe
- Filter by status
- Filter active / disabled / suspended / expired / depleted / pending/on-hold
- Filter unlimited volume / unlimited expiry
- Filter by quota/usage ranges after accounting is proven
- Filter by expiry range
- Filter by last-online range after presence is proven
- Sort by username
- Sort by created/updated time
- Sort by expiry
- Sort by usage/remaining after accounting is proven
- Sort by last online after presence is proven
- Select visible columns
- Pagination/page-size controls
- Persist useful filters in the URL when practical

### P1 — bulk actions

Bulk operations must support a dry-run/preview for destructive or high-impact actions where practical:

- Enable
- Disable / Suspend
- Revoke
- Delete/soft-delete
- Extend time
- Add/set volume
- Reset usage only after exact accounting proof
- Reissue subscription
- Apply plan/template
- Assign owner/reseller/group when those product layers are enabled

## Status model — do not collapse everything into enabled=true/false

Keep separate dimensions:

1. **Account lifecycle:** active / disabled / suspended / revoked
2. **Commercial/service state:** pending-first-use / active / expired / depleted / on-hold
3. **Presence:** online / idle / offline / unknown
4. **Quota:** unlimited / healthy / warning / depleted / unavailable
5. **Runtime health:** healthy / degraded / down / unknown

UI may summarize these dimensions, but one dimension must not silently overwrite another.

## Presence/session controls — capability-gated P1/P2

Only expose when the Runtime/data plane proves them accurately:

- Online / Offline
- Last Online
- Current sessions
- Session disconnect
- Realtime upload/download speed
- Connection/concurrency limit
- IP/session history
- Device/HWID limit
- Speed/bandwidth limit

Do not imitate Xray/PasarGuard controls if standard NaiveProxy cannot enforce them reliably in the deployed architecture.

## Customer metadata / operator ergonomics — P1

- Note/comment
- Group/tag
- Owner/admin/reseller attribution
- Created at
- Updated at
- Last renewal
- Subscription status
- Expiry status
- Traffic status
- Audit history for sensitive mutations

## Reseller / RBAC — P1

The DB/RLS foundation is not enough. Product completion requires:

- clear roles and action-level permissions;
- reseller-scoped customer visibility;
- reseller-scoped create/edit/renew/delete permissions;
- plan/term restrictions;
- credit ledger behavior where enabled;
- no cross-tenant/customer leakage;
- Owner can see and manage all appropriately scoped resources.

## Monitoring / dashboard — P1

Bring over proven patterns from OV-PvNetwork where they fit Naive:

- real server CPU/RAM/disk/network health;
- real Runtime health;
- customer counts by state;
- active/expired/depleted counts;
- online count only after presence proof;
- real traffic graphs only after accounting proof;
- Telegram/notification rules for meaningful operational/customer events;
- audit/log/diagnostic views with secret redaction.

## Multi-node / fleet — later, not a blocker for standalone R1

After standalone customer lifecycle/accounting is stable, consider proven OV-PvNetwork patterns:

- multi-node controller/agent;
- node health;
- customer-node assignment;
- reconciliation;
- failover;
- node recommendation;
- maintenance/drain/canary;
- bandwidth/node policy;
- auto-node deployment;
- safe node edit/delete lifecycle.

Do not delay standalone customer-panel completeness for fleet work.

## Installation / operations gaps

Before calling PVNaive a complete production release:

- generic fresh-server installer;
- versioned upgrade lifecycle;
- safe rollback;
- backup + verified restore drill;
- doctor/diagnostic command or equivalent operational workflow;
- log rotation/diagnostic bundle;
- release checksum/signing/supply-chain gates;
- production load/capacity evidence.

## Mandatory UI behavior rules

These rules are Owner requirements and must survive future refactors:

1. Viewing customer information is read-only.
2. Viewing QR is read-only.
3. Copying subscription is read-only.
4. Copying direct link is read-only.
5. Editing quota/time must not change password.
6. Editing quota/time must not rotate subscription unless explicitly requested.
7. Reissuing subscription must not change password.
8. Rotating password must be an explicit separate action.
9. Deleting/revoking must be explicit, confirmed, audited, and safe.
10. Customer secrets must never appear in list APIs, logs, Git, audit payloads, or diagnostics.
11. One-time secrets must only be returned after a successful deliberate create/rotate transaction.
12. UI must not claim usage, online state, device limit, speed limit, or enforcement capabilities that the Runtime has not proven.

## Recommended implementation order

### P0 — finish the real customer product first

1. Fix customer action UX: Edit, Details, View QR, Copy Subscription, Delete/Revoke.
2. Separate read-only QR/subscription actions from Reissue and Password Rotation.
3. Complete customer lifecycle actions: enable/disable/suspend/revoke/delete.
4. Complete quota/time editing: add/set volume, extend/change expiry, unlimited options.
5. Prove exact per-credential accounting.
6. Implement hard quota and reset only after accounting proof.
7. Prove trusted first-successful-connect producer.
8. Production rollout of S05 with evidence and regression checks.

### P1 — reach/exceed Sanaei/PasarGuard operator usability

9. Advanced search/filter/sort/pagination/columns.
10. Bulk actions with safe preview/confirmation.
11. Multi-dimensional customer status.
12. Renewal/next-plan/on-hold/reset-strategy UX.
13. Notes/groups/operator metadata.
14. Reseller/RBAC product UI and enforcement.
15. Dashboard/monitoring/notifications/audit/diagnostics.

### P2 — advanced runtime/product capabilities

16. Presence/session controls if technically proven.
17. Concurrency/HWID/speed controls only if enforceable.
18. Multi-node/fleet based on proven OV-PvNetwork patterns.
19. Public API/OpenAPI/webhooks after endpoint contracts stabilize.

## Definition of Done for a customer action

A customer-management feature is complete only when all applicable conditions are satisfied:

- visible and understandable in Owner UI;
- backend/API behavior implemented;
- authorization enforced;
- idempotency/concurrency handled for sensitive mutations;
- no unrelated credential/password/subscription mutation;
- audit event recorded where appropriate;
- unit/integration/web tests added;
- failure behavior tested;
- CI green on exact HEAD;
- if it affects the live Runtime, guarded production rollout/evidence exists;
- user-facing behavior matches this Owner Requirements file.

## Agent instruction / reusable execution prompt

Use the following instruction when continuing the project:

> Continue `DashSaman/PV-NativePanel` / product `PVNaive` from the current active development branch. Read `OWNER_REQUIREMENTS.md`, `PROJECT_STATUS.md`, `docs/S05_HANDOFF.md`, `AGENT_TASKS.md`, `FEATURE_MATRIX.md`, `KNOWN_ISSUES.md`, and the relevant specs/plans before changing code. Treat `OWNER_REQUIREMENTS.md` as authoritative for customer-panel UX.
>
> First audit the actual UI and API; do not assume an endpoint means the UI feature is complete. Fix the Owner-reported customer workflow gaps first: add a clear Edit action, Details, read-only View QR for the existing active subscription, Copy Subscription, safe Delete/Revoke, and separate Password Rotation and Subscription Reissue. Viewing QR or copying an existing link must never create a new link, rotate a token, or change a password. Subscription reissue must be explicit and must not rotate password. Password rotation must be a separate explicit action.
>
> Then complete lifecycle actions (enable/disable/suspend/revoke/delete), quota/time editing (add/set volume, unlimited, extend/change expiry), and only then move into exact accounting, hard quota/reset, and trusted first-successful-connect instrumentation. Never fabricate usage/remaining/online/device/speed capabilities. After the P0 customer flow is stable, implement Sanaei/PasarGuard-class search/filter/sort/bulk/status/renewal ergonomics, then reseller/RBAC, monitoring/Telegram/diagnostics, and later multi-node features using proven OV-PvNetwork patterns.
>
> Use TDD and existing safety boundaries. Preserve Runtime Agent restrictions, expected-SHA validation, exact backup, reload-only behavior, rollback, RLS/tenant isolation, secret redaction, and idempotency. Do not force-push/reset `main`; production truth must remain evidence-backed. Update `OWNER_REQUIREMENTS.md`, `AGENT_TASKS.md`, handoff/status docs, tests, and evidence as each requirement transitions from missing to verified.
