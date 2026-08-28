# PVNaive — Feature Matrix and Gap Analysis

Last updated: 2026-08-28

Legend: `✅` implemented/verified, `⚠️` partial/scaffold/not fully production-gated, `❌` absent, `❓` not proven in the audited source/snapshot.

This matrix compares **PVNaive current implementation**, not aspirational route declarations. A registered route or schema column alone does not count as implemented.

## Reference snapshots

- 3x-ui: `MHSanaei/3x-ui` @ `f727d04f6522bb94a8fb52e8352fdcafb51c11e1` (v3.7.0, audited 2026-08-28). GPL-3.0.
- PasarGuard Panel: `PasarGuard/panel` @ `e81877c0df64e5f5235f4355b0490b6bb38e3adc` (v5.2.1). AGPL-3.0.
- Marzban: `Gozargah/Marzban` @ `7f396db3e703d71a28060bc9ce4a532ec64cb1f4` (v0.8.4 snapshot used by the existing deep audit). AGPL-3.0.
- Remnawave: existing repository audit snapshot `545e9a484bad9bc8d538aa79a364a651c1ae4b5f`; use as architecture/product reference, not blind-copy source. AGPL-3.0 at audit time.
- PVNetwork OpenVPN: `DashSaman/OV-PvNetwork`, public RC + explicitly labeled production-derived feature matrix. Upstream PrimeZ lineage is MIT; preserve attribution.
- Existing detailed source-oriented comparison: `docs/PANEL_DEEP_AUDIT_FA.md`.

**License rule:** 3x-ui/PasarGuard/Marzban/Remnawave are pattern/behavior references only unless a later explicit license-compatibility decision says otherwise. No competitor code is copied by this matrix. Track PVNaive's own missing license policy in `PVN-072`.

## Core feature matrix

| Area / Feature | PVNaive now | R1 target | 3x-ui | PasarGuard | Marzban | Remnawave | OV-PvNetwork | Gap task |
|---|---:|---:|---:|---:|---:|---:|---:|---|
| Secure Owner login | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅/upstream | — |
| Opaque session + secure cookies | ✅ | ✅ | ❓ | ✅/framework | ✅/framework | ✅/framework | ❓ | PVN-035 review |
| CSRF for cookie mutations | ✅ | ✅ | ❓ | ❓ | ❓ | ❓ | ❓ | PVN-035 |
| TOTP MFA | ✅ | ✅ | ✅ | ⚠️ | ✅ | ✅ | ❓ | PVN-033/035 |
| Recovery-code login | ❌ | decision | ❓ | ❓ | ❓ | ❓ | ❓ | PVN-033 |
| Session list/revoke | ✅ | ✅ | ⚠️ | ⚠️ | ⚠️ | ✅ | ❓ | — |
| Login brute-force controls | ⚠️ DB lockout | ✅ | ✅/Fail2ban patterns | ⚠️ | ⚠️ | ⚠️ | ❓ | PVN-034 |
| RBAC roles | ⚠️ auth roles exist; business APIs absent | ✅ | ⚠️ | ✅ | ⚠️ | ✅ | ⚠️ | PVN-043 |
| Tenant/reseller isolation | ⚠️ DB/RLS foundation | ✅ | ❌/different model | ✅ | ⚠️ | ⚠️ | ❌/different model | PVN-039/043 |
| Audit log | ⚠️ auth foundation | ✅ | ⚠️ | ⚠️ | ⚠️ | ✅ | ⚠️ | PVN-037-063 |
| User CRUD | ❌ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | PVN-037 |
| Suspend / enable / disable / revoke | ❌ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | PVN-037 |
| Expiry | ❌ runtime behavior | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | PVN-037/038 |
| Traffic quota | ❌ enforcement | ✅ only after exact accounting | ✅ | ✅ | ✅ | ✅ | ✅ | PVN-045-049 |
| Reset periods | ❌ enforcement | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | PVN-038/049 |
| Renewal / next-plan | ❌ | ✅ | ⚠️ | ✅ | ✅ | ⚠️ | ⚠️ | PVN-039 |
| Bulk user operations | ❌ | ✅ | ✅ | ✅ | ✅/some | ✅ | ⚠️ | PVN-041 |
| Advanced search/filter/sort | ❌ | ✅ | ✅ | ✅ | ✅ | ✅ | ⚠️ | PVN-042 |
| Computed status dimensions | ❌ | ✅ | ⚠️ | ✅ | ✅ | ✅ | ⚠️ | PVN-042 |
| Naive runtime credential import | ❌ production; schema only in branch | ✅ | N/A | N/A | N/A | N/A | N/A | PVN-020-029 |
| Multiple Naive credentials | ❌ | ✅ if custom Caddy proves syntax | N/A | N/A | N/A | N/A | N/A | PVN-021/027 |
| Runtime credential rename/rotate | ❌ | ✅ | ✅ analogous | ✅ analogous | ✅ analogous | ✅ analogous | ✅ analogous | PVN-024-029 |
| Atomic validate/apply/rollback | ⚠️ design + prior Caddy exposure procedure | ✅ | ⚠️ | ⚠️ | ⚠️ | ✅ patterns | ✅ patterns | PVN-023/027 |
| Desired vs applied revision | ⚠️ schema foundation | ✅ | ⚠️ | ⚠️ | ⚠️ | ✅ patterns | ✅ | PVN-024/051 |
| Last-known-good runtime | ⚠️ design | ✅ | ❓ | ❓ | ❓ | ⚠️ | ✅ | PVN-023/051 |
| Exact per-credential accounting | ❌ unproven blocker | must be proven | Xray-based ✅ | Xray-based ✅ | Xray-based ✅ | ✅ runtime metrics | OpenVPN-specific ✅ | PVN-045-048 |
| Restart/double-count safety | ❌ | ✅ | ❓ | ❓ | ❓ | ⚠️ | ✅ production pattern | PVN-046 |
| Live session visibility | ❌ | capability-gated | ✅ | ✅ | ✅ | ✅ | integration hook | PVN-047/050 |
| Concurrency limit | ❌ enforcement | capability-gated | ✅ | ✅ | ✅ | ✅ | session-control hook | PVN-050 |
| Device/HWID limit | ❌ enforcement | capability-gated | ✅ | ✅ | ⚠️ | ✅ | ❓ | PVN-050 |
| Speed limit | ❌ | capability-gated | ✅/routing/runtime | ✅ | ✅ | ✅ | bandwidth hook | PVN-050 |
| Subscription token URL | ❌ implementation | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | PVN-052 |
| Subscription info/usage page | ❌ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | PVN-053 |
| QR/share links | ❌ | useful subset | ✅ | ✅ | ✅ | ✅ | ✅ | PVN-055 |
| Client compatibility lab/matrix | ❌ | ✅ | broad clients | broad clients | broad clients | broad clients | OpenVPN clients | PVN-054/055 |
| Multi-node / fleet | ❌ | **not R1 blocker** | ✅ | ✅ | ✅ nodes | ✅ | ✅ | future after PVN-067 |
| Node health/stats | ❌ | future | ✅ | ✅ | ✅ | ✅ strong | ✅ | future |
| Failover/load balancing | ❌ | future | ✅ | ✅ | ⚠️ | ✅ patterns | ✅ patterns | future |
| Realtime dashboard | ⚠️ authenticated preview only | ✅ real metrics only | ✅ | ✅ | ✅ | ✅ | ✅ | PVN-044/048/063 |
| Server resource monitoring | ❌ | ✅ | ✅ | ✅ | ⚠️ | ✅ | ✅ | PVN-063 |
| Notifications | ❌ | ✅ | Telegram ✅ | Telegram ✅ | Telegram ✅ | ✅ | Telegram/monitoring derived | PVN-056/057 |
| REST API | ⚠️ health/auth only; future registry exists | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | PVN-025/037+/071 |
| API docs/OpenAPI | ❌ current implementation truth | P2 | ✅ | ✅ | ✅ | ✅ | ❓ | after stable endpoints |
| API rate limit | ❌ HTTP layer | ✅ | ✅ patterns | ⚠️ | ⚠️ | ⚠️ | ❓ | PVN-034/071 |
| Webhook | ❌ | P2 after notification core | ⚠️ | ⚠️ | ✅ | ✅ | ❓ | post-PVN-056 |
| Database backup | ✅ encrypted/manual foundation | ✅ scheduled | ✅ | ✅ | ✅ | ✅ | ✅ | PVN-062 |
| Restore drill | ✅ S03 | ✅ recurring | ⚠️ | ⚠️ | ⚠️ | ⚠️ | ✅ lifecycle | PVN-062 |
| Config/Caddy backup | ✅ for panel exposure procedure | ✅ automated | ⚠️ | ⚠️ | ⚠️ | ✅ patterns | ✅ | PVN-023/062 |
| Fresh installer | ❌ final | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | PVN-058 |
| Upgrade | ❌ final | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | PVN-059 |
| Rollback/uninstall | ⚠️ stage-specific only | ✅ | ⚠️ | ⚠️ | ⚠️ | ⚠️ | ✅ | PVN-060 |
| Health checks | ✅ DB/API foundation | ✅ full system | ✅ | ✅ | ✅ | ✅ | ✅ | PVN-032/063 |
| Logs/rotation/diagnostics | ❌ product layer | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ doctor pattern | PVN-063 |
| SBOM/SAST/secret scan/release signing | ❌ | ✅ | mixed | mixed | mixed | mixed | partial lifecycle | PVN-065 |
| Production load/capacity evidence | ❌ | ✅ | project-specific | project-specific | project-specific | project-specific | project-specific | PVN-066 |

## Feature gaps that should become product capabilities

### P0 — required before Production R1

1. **Safe Naive credential ownership and mutation** — PVN-020…029. This is uniquely central to PVNaive and must be safer than generic config editors.
2. **Auth hardening defects/abuse controls** — PVN-030,031,034.
3. **User lifecycle and tenant authorization** — PVN-037,040,043.
4. **Exact accounting proof + anti-double-count + quota enforcement** — PVN-045,046,047,048,049. Competitors can rely on Xray/OpenVPN counters; PVNaive must prove its own Naive source rather than copy claims.
5. **Subscription rendering + real client compatibility** — PVN-052,054.
6. **Fresh install, upgrade, disaster restore, supply-chain gates, pilot evidence** — PVN-058,059,062,065,066,067.

### P1 — production-quality differentiators

- Dry-run bulk actions with conflict/rollback preview — PVN-041.
- Explainable multi-dimensional user status rather than one `enabled` boolean — PVN-042.
- Capability-first UI: hide speed/device/session/quota controls unless adapter proves them — PVN-050/051.
- Support/diagnostic bundle with redaction and request IDs — PVN-063.
- Certificate/domain rotation with same validate→reload→rollback discipline — PVN-064.

### P2/P3 — useful but not R1 blockers

- QR/template compatibility extras — PVN-055.
- Telegram notification channel — PVN-057.
- OpenAPI/webhook once endpoint contracts stabilize.
- Multi-node/fleet, node recommendation, drain/maintenance/canary only after standalone release is proven. OV-PvNetwork and Remnawave are strong architecture references for that later phase.

## Deliberately rejected feature bloat for R1

- arbitrary Caddy editor or root shell from the panel;
- custom wire protocol/chaff generator;
- full financial/payment gateway;
- multi-node controller as an R1 dependency;
- fake device/session/speed enforcement;
- estimated access-log traffic as billable exact usage;
- storing browsing destinations by default.

The target remains **Stable + Useful + Maintainable**, not feature-count parity.
