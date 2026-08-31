# PVNaive — Verified Gaps Queue After Current Execution Sequence

Date: 2026-08-31

Purpose: preserve the Owner-requested, independently re-checked gaps so current teams can finish the mandated ROADMAP sequence without losing later work. This file does **not** reorder the canonical `ROADMAP.md`. Teams must continue the existing numbered execution order. When the current ordered tasks reach the corresponding areas, use this as an additional verification checklist.

Source of truth remains latest `main` code + canonical project docs + exact-head CI + Production evidence. Documentation alone never makes a feature DONE.

## Important freshness note

The repository has advanced beyond the older Task #4 documentation snapshot. Tasks #5-#8 have implementation commits on `main`; do not recreate them from stale reports. In particular, manual and bulk usage reset have real implementations. Reconcile canonical documentation as part of the existing documentation task rather than rewriting completed work.

## Verified open gaps to retain in the queue

### Security / consistency — P0

1. **Refresh-token reuse-family handling** — ensure reuse of a rotated/revoked refresh token reaches family-revocation/audit semantics; close only with RED→GREEN regression evidence.
2. **Commit-aware HTTP success** — generic authenticated mutations must not emit success before durable DB commit is known; injected commit failure must never return success.
3. **DB/schema-backed readiness** — `/health/ready` needs a bounded current DB/schema probe with timeout/failure tests and no secret leakage.
4. **Whole-product authorization / IDOR / CSRF / redaction / fuzz gates** — complete negative Route × role × tenant coverage, especially cross-reseller access.

### Accounting / enforcement — P0

5. **Periodic traffic reset** — restart-safe persisted scheduler/cursor, timezone semantics, exactly-once/idempotent execution, audit/history and failure recovery.
6. **Hard-quota controlled Production proof** — simultaneous connections, race/exact exhaustion, reload/restart/reconnect, no negative remaining and no bypass.
7. **First-successful-CONNECT controlled Production proof** — prove reads/QR/subscription/health/reload/failed-auth are inert; only successful authenticated CONNECT activates; duplicate/concurrent/reconnect/restart remain idempotent.

### Sessions / limits

8. **Operator session management** — trustworthy active session list with IP, connected/last activity, duration and reliable bytes.
9. **Kill/disconnect session** — real data-plane disconnect primitive, confirmation, authorization, audit and failure tests.
10. **Concurrent-session limit** — Unlimited/N semantics enforced under races and reconnects.
11. **Simultaneous unique-IP limit** — exact semantics and real enforcement with race/failure tests.
12. **IP/session history** — privacy-aware bounded retention.
13. **HWID/device identity** — only implement if a trustworthy standard Naive/Karing identity is proven; otherwise explicitly unavailable. Never fabricate HWID.
14. **Per-user speed limit** — only expose after real data-plane shaping/enforcement is proven. Never create a UI-only/fake speed limit.

### Reseller / RBAC / audit

15. **Reseller CRUD** — create/edit/disable/revoke/list/search with tenant ownership boundaries.
16. **Reseller wallet/credit** — audited balance semantics.
17. **Immutable financial ledger** — credit/debit/create/renew/refund/adjustment history.
18. **Reseller restrictions** — allowed plans, max users/active users, credit limits and Owner oversight.
19. **Customer history** — create/renew/volume/expiry/plan/group/tag/suspend/resume/revoke/rotate/reissue/reset events.
20. **Audit Explorer** — actor/user/action/date/IP/result filters with strict redaction.

### Notifications / observability / API

21. **Notification product engine** — persistence, preferences, event wiring, history, retry/dedupe and operator workflow beyond existing foundations.
22. **Telegram product workflow** — secure configuration, rules, history and product UX beyond the tested transport foundation.
23. **Historical monitoring/traffic charts** — do not invent data; use trustworthy persisted samples only.
24. **Application/runtime/security log explorer** — product UI with authorization and strict redaction.
25. **OpenAPI/Swagger completion** — current ready-route foundation is partial; stabilize/version the broader API contract.
26. **Webhooks** — only after stable event contracts; authentication, signing, retries, idempotency and delivery history required.

### Fleet / lifecycle / clients / capacity

27. **Multi-node/fleet controller** — standalone must remain fully functional; add real node auth/health/metrics/capacity/assignment/reconciliation without fake health.
28. **Fleet operations** — drain/maintenance/canary/node upgrade/rollback/failover/smart selection/dashboard.
29. **Fresh secure Ubuntu 26.04 installer** — clean version-pinned installation of PostgreSQL/Caddy/API/agents/systemd/firewall/TLS/migrations/web + Doctor; prove on a clean supported VM.
30. **Generic versioned upgrade** — migration-aware backup/deploy/health/rollback path beyond same-schema deployment.
31. **Generic rollback + conservative uninstall** — explicit data-retention policy and tested version rollback/uninstall.
32. **Karing compatibility campaign** — real Windows/Android/iOS/macOS/Linux evidence for supported current versions.
33. **50/100/200/400+ capacity campaign** — measure CPU/RAM/disk/PostgreSQL/Caddy/telemetry/accounting lag/network/API latency and correctness under accounting/quota/session races/restarts/reconnects. Existing bounded request rehearsal is not capacity proof.
34. **Supply-chain security** — SAST, secret scanning, dependency vulnerability scanning, final SBOM/provenance/signing/NOTICE/license policy.

### Product/UI/release closure

35. **Advanced bulk/search completion** — remaining actions/filters/sorts/selectable columns/URL-persisted state as defined by canonical roadmap; preserve idempotency and confirmation for destructive actions.
36. **Final UI/UX polish** — accessibility, desktop/mobile QA, dark/light behavior and consistent capability-unavailable states; no fake metrics/status.
37. **Documentation reconciliation** — remove stale contradictions between main/schema/Production evidence and canonical status files while preserving historical snapshots as historical.
38. **Production/main drift closure** — do not assume merged `main` changes are deployed. Every Production-facing task needs backed-up rollout and live postflight evidence before Production status is advanced.
39. **Final clean-install proof, Production smoke and Release Candidate gates** — complete only after all P0 release blockers are closed and exact-head required workflows are green.

## Rules for future teams

- Do not reorder the canonical ROADMAP merely because this checklist exists.
- Before implementing an item, search latest `main`, recent commits/PRs and tests for newer valid work.
- Never recreate Tasks #5-#8 from stale snapshots if their current implementation remains valid.
- Never mark a schema/table/route/foundation as a finished product capability without backend behavior, authorization, tests and operator UI where applicable.
- Never fabricate accounting, online/session state, HWID, speed, node health or capacity evidence.
- Keep Runtime credentials and Subscription tokens independent from quota/expiry/reset/read/QR operations unless an explicit credential-rotation action is requested.
- Production mutation requires fresh preflight, backups, rollback plan, bounded apply, postflight and redacted evidence.
- Close security/enforcement items only with failure-path tests and exact-head CI evidence.

## Completion reporting

When an item above is reached in the canonical sequence, the implementing team should record:

`TASK / STATUS / WHAT CHANGED / FILES / TESTS / CI / PRODUCTION / EVIDENCE / REMAINING`

Final product status remains `NOT PRODUCTION READY` while any release-blocking P0 item is incomplete.
