# PVNaive — Feature Matrix and Gap Analysis

Last updated: 2026-08-28

Legend: `✅` implemented/verified, `⚠️` partial or implemented in code/rehearsal but not yet live-production-gated, `❌` absent, `❓` not proven in the audited source/snapshot.

This matrix compares **PVNaive current implementation**, not aspirational route declarations. A registered route or schema column alone does not count as implemented.

## Reference snapshots

- 3x-ui: `MHSanaei/3x-ui` @ `f727d04f6522bb94a8fb52e8352fdcafb51c11e1` (v3.7.0, audited 2026-08-28). GPL-3.0.
- PasarGuard Panel: `PasarGuard/panel` @ `e81877c0df64e5f5235f4355b0490b6bb38e3adc` (v5.2.1). AGPL-3.0.
- Marzban: `Gozargah/Marzban` @ `7f396db3e703d71a28060bc9ce4a532ec64cb1f4` (v0.8.4 snapshot used by the existing deep audit). AGPL-3.0.
- Remnawave: existing repository audit snapshot `545e9a484bad9bc8d538aa79a364a651c1ae4b5f`; use as architecture/product reference, not blind-copy source. AGPL-3.0 at audit time.
- PVNetwork OpenVPN: `DashSaman/OV-PvNetwork`, public RC + explicitly labeled production-derived feature matrix. Upstream PrimeZ lineage is MIT; preserve attribution.
- Existing detailed source-oriented comparison: `docs/PANEL_DEEP_AUDIT_FA.md`.

**License rule:** GPL/AGPL projects above are pattern/behavior references unless a later explicit license-compatibility decision says otherwise. Track PVNaive's own license policy in `PVN-072`.

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
| Audit log | ⚠️ auth/runtime foundations | ✅ | ⚠️ | ⚠️ | ⚠️ | ✅ | ⚠️ | PVN-037-063 |
| User CRUD | ❌ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | PVN-037 |
| Suspend / enable / disable / revoke business user | ❌ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | PVN-037 |
| Expiry | ❌ runtime behavior | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | PVN-037/038 |
| Traffic quota | ❌ enforcement | ✅ only after exact accounting | ✅ | ✅ | ✅ | ✅ | ✅ | PVN-045-049 |
| Reset periods | ❌ enforcement | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | PVN-038/049 |
| Renewal / next-plan | ❌ | ✅ | ⚠️ | ✅ | ✅ | ⚠️ | ⚠️ | PVN-039 |
| Bulk user operations | ❌ | ✅ | ✅ | ✅ | ✅/some | ✅ | ⚠️ | PVN-041 |
| Advanced search/filter/sort | ❌ | ✅ | ✅ | ✅ | ✅ | ✅ | ⚠️ | PVN-042 |
| Computed status dimensions | ❌ | ✅ | ⚠️ | ✅ | ✅ | ✅ | ⚠️ | PVN-042 |
| Naive runtime credential import | ⚠️ implemented + full rehearsal; live preflight pending | ✅ | N/A | N/A | N/A | N/A | N/A | PVN-028/029 |
| Multiple Naive credentials | ⚠️ exact pinned Caddy proof + full rehearsal; live rollout pending | ✅ | N/A | N/A | N/A | N/A | N/A | PVN-028/029 |
| Runtime credential create/rename/rotate/enable/disable/revoke | ⚠️ implemented + rehearsal green | ✅ | ✅ analogous | ✅ analogous | ✅ analogous | ✅ analogous | ✅ analogous | PVN-029 |
| One-time generated secret | ✅ code/rehearsal | ✅ | varied | varied | varied | varied | varied | PVN-029 live smoke |
| Copy-ready `naive+https://` customer link | ✅ URI builder/UI; live Karing smoke pending | ✅ | N/A | N/A | N/A | N/A | N/A | PVN-029/054 |
| Atomic validate/apply/rollback | ⚠️ implemented + failure rehearsal; production execution pending | ✅ | ⚠️ | ⚠️ | ⚠️ | ✅ patterns | ✅ patterns | PVN-029 |
| Expected-SHA stale-write protection | ✅ code/rehearsal | ✅ | ❓ | ❓ | ❓ | ⚠️ | ✅ patterns | PVN-029 live gate |
| Desired vs applied revision | ✅ runtime state + saga | ✅ | ⚠️ | ⚠️ | ⚠️ | ✅ patterns | ✅ | PVN-029/051 |
| Runtime reconciliation-required state | ✅ explicit failure contract | ✅ | ❓ | ❓ | ❓ | ⚠️ | ✅ patterns | PVN-029 live evidence |
| Last-known-good runtime | ⚠️ exact operator backup/rollback exists; broader adapter state pending | ✅ | ❓ | ❓ | ❓ | ⚠️ | ✅ | PVN-051 |
| Exact per-credential accounting | ❌ unproven blocker | must be proven | Xray-based ✅ | Xray-based ✅ | Xray-based ✅ | ✅ runtime metrics | OpenVPN-specific ✅ | PVN-045-048 |
| Restart/double-count safety | ❌ | ✅ | ❓ | ❓ | ❓ | ⚠️ | ✅ production pattern | PVN-046 |
| Live session visibility | ❌ | capability-gated | ✅ | ✅ | ✅ | ✅ | integration hook | PVN-047/050 |
| Concurrency limit | ❌ enforcement | capability-gated | ✅ | ✅ | ✅ | ✅ | session-control hook | PVN-050 |
| Device/HWID limit | ❌ enforcement | capability-gated | ✅ | ✅ | ⚠️ | ✅ | ❓ | PVN-050 |
| Speed limit | ❌ | capability-gated | ✅/routing/runtime | ✅ | ✅ | ✅ | bandwidth hook | PVN-050 |
| Subscription token URL | ❌ implementation | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | PVN-052 |
| Subscription info/usage page | ❌ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | PVN-053 |
| QR/share links | ⚠️ one Naive URI copy only; no subscription/QR lifecycle | useful subset | ✅ | ✅ | ✅ | ✅ | ✅ | PVN-052/055 |
| Client compatibility lab/matrix | ❌ live matrix; URI format only unit-tested | ✅ | broad clients | broad clients | broad clients | broad clients | OpenVPN clients | PVN-054/055 |
| Multi-node / fleet | ❌ | **not R1 blocker** | ✅ | ✅ | ✅ nodes | ✅ | ✅ | future after PVN-067 |
| Node health/stats | ❌ | future | ✅ | ✅ | ✅ | ✅ strong | ✅ | future |
| Failover/load balancing | ❌ | future | ✅ | ✅ | ⚠️ | ✅ patterns | ✅ patterns | future |
| Realtime dashboard | ⚠️ authenticated runtime status only | ✅ real metrics only | ✅ | ✅ | ✅ | ✅ | ✅ | PVN-044/048/063 |
| Server resource monitoring | ❌ | ✅ | ✅ | ✅ | ⚠️ | ✅ | ✅ | PVN-063 |
| Notifications | ❌ | ✅ | Telegram ✅ | Telegram ✅ | Telegram ✅ | ✅ | Telegram/monitoring derived | PVN-056/057 |
| REST API | ⚠️ health/auth/runtime implemented; business API future | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | PVN-037+/071 |
| API docs/OpenAPI | ❌ current implementation truth | P2 | ✅ | ✅ | ✅ | ✅ | ❓ | after stable endpoints |
| API rate limit | ❌ HTTP layer | ✅ | ✅ patterns | ⚠️ | ⚠️ | ⚠️ | ❓ | PVN-034/071 |
| Webhook | ❌ | P2 after notification core | ⚠️ | ⚠️ | ✅ | ✅ | ❓ | post-PVN-056 |
| Database backup | ✅ encrypted/manual foundation | ✅ scheduled | ✅ | ✅ | ✅ | ✅ | ✅ | PVN-062 |
| Restore drill | ✅ S03 | ✅ recurring | ⚠️ | ⚠️ | ⚠️ | ⚠️ | ✅ lifecycle | PVN-062 |
| Config/Caddy backup | ✅ runtime operator exact backup + rollback | ✅ automated/scheduled | ⚠️ | ⚠️ | ⚠️ | ✅ patterns | ✅ | PVN-062 |
| Fresh installer | ❌ final | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | PVN-058 |
| Upgrade | ⚠️ guarded S04R Pilot upgrade for existing install | ✅ generic versioned | ✅ | ✅ | ✅ | ✅ | ✅ | PVN-059 |
| Rollback/uninstall | ⚠️ S04R stage rollback + historical stage rollback; no generic uninstall | ✅ | ⚠️ | ⚠️ | ⚠️ | ⚠️ | ✅ | PVN-060 |
| Health checks | ✅ DB/API/runtime foundations | ✅ full system | ✅ | ✅ | ✅ | ✅ | ✅ | PVN-032/063 |
| Logs/rotation/diagnostics | ❌ product layer | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ doctor pattern | PVN-063 |
| SBOM/SAST/secret scan/release signing | ❌ | ✅ | mixed | mixed | mixed | mixed | partial lifecycle | PVN-065 |
| Production load/capacity evidence | ❌ | ✅ | project-specific | project-specific | project-specific | project-specific | project-specific | PVN-066 |

## Immediate Pilot capability

The current S04R branch is deliberately useful before the full R1 exists. After `PVN-028/029` pass live, the Owner can:

1. securely import the existing Naive credential without exposing its password to the browser;
2. create additional Naive credentials;
3. receive generated passwords only once after successful commit;
4. copy a ready `naive+https://...` customer link;
5. rename, rotate, disable, re-enable or soft-revoke credentials;
6. retain the last-active guard and revision/idempotency protections.

This **does not** turn raw runtime credentials into business users with quota/expiry/accounting. Those remain separate roadmap work.

## Feature gaps that should become product capabilities

### P0 — required before final Production R1

1. Live S04R Pilot proof — PVN-028/029.
2. Auth hardening defects/abuse controls — PVN-030,031,034.
3. User lifecycle and tenant authorization — PVN-037,040,043.
4. Exact accounting proof + anti-double-count + quota enforcement — PVN-045…049.
5. Subscription rendering + real client compatibility — PVN-052,054.
6. Fresh install, disaster restore, supply-chain gates, pilot/load evidence — PVN-058,062,065,066,067.

### P1 — production-quality differentiators

- Dry-run bulk actions with conflict/rollback preview — PVN-041.
- Explainable multi-dimensional user status rather than one `enabled` boolean — PVN-042.
- Capability-first UI: hide speed/device/session/quota controls unless adapter proves them — PVN-050/051.
- Support/diagnostic bundle with redaction and request IDs — PVN-063.
- Certificate/domain rotation with the same validate→reload→rollback discipline — PVN-064.

### P2/P3 — useful but not R1 blockers

- QR/template compatibility extras — PVN-055.
- Telegram notification channel — PVN-057.
- OpenAPI/webhook once endpoint contracts stabilize.
- Multi-node/fleet after standalone release is proven.

## Deliberately rejected feature bloat for R1

- arbitrary Caddy editor or root shell from the panel;
- custom wire protocol/chaff generator;
- full financial/payment gateway;
- multi-node controller as an R1 dependency;
- fake device/session/speed enforcement;
- estimated access-log traffic as billable exact usage;
- storing browsing destinations by default.

The target remains **Stable + Useful + Maintainable**, not feature-count parity.
