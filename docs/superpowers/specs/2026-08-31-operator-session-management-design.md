# Operator Session Management Design

## Scope

Task 12 exposes a trustworthy active-session view for PVNaive customers. It does not implement kill, concurrent limits, unique-IP limits, or historical retention; those remain Tasks 13–16.

## Trust boundary

- Client IP is captured only from Caddy's accepted connection peer (`http.Request.RemoteAddr`).
- `Forwarded`, `X-Forwarded-For`, User-Agent and any client-supplied identity are never accepted as the session peer identity.
- The peer must normalize via `net.SplitHostPort` + `net.ParseIP`; invalid/missing peer identity fails the tracked CONNECT closed before payload forwarding.
- Existing exact byte accounting and quota semantics remain unchanged.

## Data model

Schema 17 adds `pvnaive.direct_naive_accounting_session_peers`, keyed by the existing session identity `(runtime_credential_id,node_id,boot_id,session_id)`. The row stores `service_term_id`, `client_ip inet`, `first_recorded_at`, and `updated_at`.

A SECURITY DEFINER function records a peer only after the accounting open event has created the trusted session row. Replaying the same peer is idempotent. A different IP for the same session identity is rejected. Existing pre-schema17 historical sessions simply have no peer row and are never fabricated with an IP.

A second tenant-scoped SECURITY DEFINER read function returns only active sessions for one customer visible to the signed request context. Active means:

- `final=false`;
- `accounting_complete=true`;
- `last_observed_at >= observed_at - stale_after`;
- a trusted peer row exists.

Returned fields: session identity, client IP, node, connected timestamp, last activity, upload/download cumulative bytes, duration seconds, and service term id.

## Data-plane protocol

Telemetry adds `POST /v1/accounting/session-peer` with runtime credential UUID, node id, boot UUID, session UUID, client IP, and observed timestamp. Forwardproxy calls it immediately after `accountingSession.open()` and before wrapping/forwarding payload. Failure is fail-closed.

This is one extra local Unix-socket call per successful CONNECT only; it is not on the per-chunk accounting path.

## Operator API/UI

Authenticated product route: `GET /api/v1/product/users/{user_id}/sessions` (route name `users.sessions.index`). The Customer service performs the read inside the existing signed RLS transaction and returns `sessions` plus `observed_at`.

The Product Customers row menu adds `نشست‌های فعال`. The modal shows IP, node, connected time, last activity, duration, upload, download. It displays an explicit empty state when no trusted active session exists. It does not expose historical sessions.

## Security and privacy

- Tenant/customer scope is enforced in PostgreSQL, not by filtering rows after query.
- No endpoint accepts an arbitrary tenant id.
- Session peer IP is operator-only and not exposed on `/s` or `/sub`.
- Task16 will add bounded retention; Task12 does not create an unbounded second history store beyond the existing accounting session lifetime.
- No session-kill claim is made until Task13 has a real disconnect primitive.

## Acceptance

1. PostgreSQL18 migration 16→17 and clean rollback before peer data exist.
2. Peer record same-IP replay is idempotent; changed IP is rejected.
3. RLS/read function cannot cross tenant/customer scope.
4. Forwardproxy uses `RemoteAddr`, ignores spoofed forwarding headers, and fails closed if peer registration fails.
5. Active API excludes final/stale/incomplete/missing-peer sessions and reports exact cumulative bytes.
6. Web modal renders trusted session fields and empty state.
7. Existing exact-accounting, quota, first-CONNECT, periodic reset, R1 build and pinned-forwardproxy gates remain green.
