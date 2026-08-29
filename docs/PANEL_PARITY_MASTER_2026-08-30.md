# PVNaive — Production / Panel Parity Master 2026-08-30

Status: evidence-backed reconciliation in progress

## Snapshot lock

| Product | Audited ref | License / note |
|---|---|---|
| PVNaive | `DashSaman/PV-NativePanel@a021aa4b62c35b775fb521d042b2f8e6dbde10b0` | project source; Production read-only audited 2026-08-30 |
| 3x-ui / Sanaei | `MHSanaei/3x-ui@f727d04f6522bb94a8fb52e8352fdcafb51c11e1` (`v3.7.0`) | GPL-3.0; behavior/UX reference only |
| PasarGuard | `PasarGuard/panel@aebf7256927710329d380d67ce96224f287ae5f6` (`v5.3.0`) | AGPL-3.0; behavior/UX reference only |
| Hiddify Manager | `hiddify/Hiddify-Manager@a99c811aa63fe908f1e06607b81f475b502ebf07` (`dev`); stable `v12.3.3` | GPL-3.0; behavior/UX reference only |
| OV-PvNetwork | `DashSaman/OV-PvNetwork@5b6a578bfe7733ebc67c08d9c431da6e32ac7ced` (`v1.0.0-rc1` public snapshot) | MIT; operational patterns are reusable where architecture fits |

The Hiddify repository's default branch is `dev`, not `main`. The current PasarGuard source was checked beyond README; its user model directly exposes traffic reset strategy, notes, on-hold state, groups, HWID limit, next plan, traffic/online fields, advanced list filters/sorts, HWID history and bulk dry-run structures. 3x-ui source was checked at its current signed v3.7.0 main. OV-PvNetwork is a sanitized public RC snapshot; production-derived capabilities documented there are not treated as byte-for-byte public implementation proof.

## Status legend

PVNaive column:

- **DONE** — implementation plus evidence exists; where production-facing, current Production behavior/data supports the claim.
- **PARTIAL** — useful implementation/foundation exists but a required handler/UI/authorization/proof/production acceptance step is missing.
- **BLOCKED** — not implemented as a usable product capability yet, even if a route/schema placeholder exists.
- **OPTIONAL** — useful only if product direction calls for it.
- **N/A** — not appropriate to standard NaiveProxy/PVNaive architecture.

Reference columns:

- **✓** — current source/release evidence supports the capability.
- **△** — partial, differently modeled, protocol-specific, or not a directly equivalent product capability.
- **—** — absent/not applicable/not independently proven in this audit.

A declared route with `Ready=false`, a DB table by itself, or a UI placeholder does **not** count as implemented.

## Master matrix

| # | Feature | PVNaive | 3x-ui | PasarGuard | Hiddify | OV-PvNetwork | Priority | Action |
|---:|---|---|---|---|---|---|---|---|
| 1 | User management | DONE | ✓ | ✓ | ✓ | ✓ | P0 | Preserve customer/runtime separation |
| 2 | Create / Edit / Delete | DONE | ✓ | ✓ | ✓ | ✓ | P0 | Keep safe lifecycle and audit |
| 3 | Enable / Disable | PARTIAL | ✓ | ✓ | ✓ | ✓ | P1 | Customer lifecycle has suspend/resume; make explicit disable semantics consistent |
| 4 | Suspend / Resume | DONE | ✓ | ✓ | ✓ | ✓ | P0 | Preserve |
| 5 | Revoke | DONE | ✓ | ✓ | ✓ | ✓ | P0 | Preserve safe runtime coordination |
| 6 | Soft-delete | DONE | △ | ✓ | △ | ✓ | P1 | Keep safe-delete default and history |
| 7 | Username / password management | PARTIAL | ✓ | ✓ | ✓ | ✓ | P1 | Password rotation is explicit; finish consistent customer username edit semantics |
| 8 | Quota | DONE | ✓ | ✓ | ✓ | ✓ | P0 | Commercial quota + exact accounting core present |
| 9 | Unlimited quota | DONE | ✓ | ✓ | ✓ | ✓ | P0 | Preserve |
| 10 | Add volume | DONE | △ | ✓ | ✓ | ✓ | P0 | Preserve |
| 11 | Set total volume | DONE | ✓ | ✓ | ✓ | ✓ | P0 | Preserve |
| 12 | Reset usage | BLOCKED | ✓ | ✓ | ✓ | ✓ | P0 | Implement manual reset after explicit accounting-baseline design |
| 13 | Periodic traffic reset | PARTIAL | ✓ | ✓ | ✓ | △ | P0 | Model exists; add restart-safe scheduler/cursor/exactly-once execution |
| 14 | Expiry | DONE | ✓ | ✓ | ✓ | ✓ | P0 | Preserve |
| 15 | No expiry | DONE | ✓ | ✓ | ✓ | ✓ | P0 | Preserve |
| 16 | Validity from creation | DONE | ✓ | ✓ | ✓ | ✓ | P0 | Preserve |
| 17 | Validity from first successful connection | PARTIAL | △ | ✓ | ✓ | — | P0 | Producer/core exists; record controlled Production transition proof |
| 18 | Extend days | DONE | △ | ✓ | ✓ | ✓ | P0 | Preserve |
| 19 | Manual expiry | DONE | ✓ | ✓ | ✓ | ✓ | P0 | Preserve |
| 20 | Plans | DONE | △ | ✓ | ✓ | △ | P1 | Plan presets implemented; expand admin UX only as needed |
| 21 | Groups | DONE | △ | ✓ | ✓ | △ | P1 | Preserve tenant-safe associations |
| 22 | Tags | DONE | △ | △ | △ | △ | P1 | Preserve |
| 23 | Notes | DONE | △ | ✓ | ✓ | △ | P1 | Preserve |
| 24 | Next Plan | PARTIAL | △ | ✓ | △ | — | P1 | Backend semantics exist; finish history/operator UX proof |
| 25 | On Hold | PARTIAL | △ | ✓ | ✓ | — | P1 | Model/status exists; finish explicit operator workflow |
| 26 | Renewal | DONE | △ | ✓ | ✓ | △ | P1 | Preserve immutable ServiceTerm semantics |
| 27 | Service history | PARTIAL | △ | ✓ | △ | △ | P1 | Build customer history UI over immutable terms/audit |
| 28 | Search | DONE | ✓ | ✓ | ✓ | ✓ | P1 | Preserve server-side search |
| 29 | Filters | DONE | ✓ | ✓ | ✓ | ✓ | P1 | Extend after accounting/presence fields are fully wired |
| 30 | Sorting | DONE | ✓ | ✓ | ✓ | ✓ | P1 | Preserve server-side sort |
| 31 | Pagination | DONE | ✓ | ✓ | ✓ | ✓ | P1 | Preserve |
| 32 | Selectable table columns | BLOCKED | △ | ✓ | △ | — | P2 | Add after backend correctness work |
| 33 | Bulk operations | DONE | ✓ | ✓ | △ | △ | P1 | Current action set includes lifecycle/volume/plan/group/tag/subscription actions |
| 34 | Bulk preview / dry-run | DONE | △ | ✓ | △ | △ | P1 | Preserve preview-before-execute and idempotency |
| 35 | Subscription URL | DONE | ✓ | ✓ | ✓ | △ | P0 | `/sub/<token>` machine endpoint |
| 36 | Subscription info page | DONE | ✓ | ✓ | ✓ | △ | P0 | `/s/<token>` human page |
| 37 | QR | DONE | ✓ | ✓ | ✓ | △ | P0 | Local generation only |
| 38 | Direct config | DONE | ✓ | ✓ | ✓ | ✓ | P0 | Direct Naive delivery; keep separate from subscription |
| 39 | Subscription refresh | PARTIAL | ✓ | ✓ | ✓ | △ | P1 | Stable read endpoint done; real Karing refresh acceptance still pending |
| 40 | Subscription reissue | DONE | ✓ | ✓ | ✓ | △ | P0 | Keep separate from password rotation |
| 41 | Password rotation | DONE | ✓ | ✓ | ✓ | ✓ | P0 | Preserve one-time secret delivery |
| 42 | Online status | PARTIAL | ✓ | ✓ | ✓ | ✓ | P0 | Trusted telemetry exists; wire evidence-backed projection to all customer views |
| 43 | Last online | PARTIAL | ✓ | ✓ | ✓ | ✓ | P0 | Same read-model wiring gap |
| 44 | Sessions | PARTIAL | ✓ | ✓ | ✓ | ✓ | P0 | Accounting sessions exist; operator-facing session API/UI not ready |
| 45 | Kill session | BLOCKED | △ | ✓ | △ | ✓ | P0 | Implement only with real disconnect primitive + audit |
| 46 | Concurrent connection limit | BLOCKED | ✓ | ✓ | △ | ✓ | P0 | Design/enforce against trusted active sessions |
| 47 | IP limit | BLOCKED | ✓ | ✓ | △ | ✓ | P0 | Define simultaneous unique-IP semantics and enforce reliably |
| 48 | IP/session history | BLOCKED | △ | ✓ | △ | ✓ | P1 | Add privacy-aware bounded retention |
| 49 | Device / HWID limit | BLOCKED | — | ✓ | △ | — | P1 | PoC first; no fake identity if Karing/Naive cannot provide trustworthy device ID |
| 50 | Speed limit | BLOCKED | △ | ✓ | △ | ✓ | P1 | PoC at real data plane; hide UI until enforceable |
| 51 | Realtime speed | BLOCKED | ✓ | ✓ | ✓ | ✓ | P1 | Derive only from trusted counters/sampling |
| 52 | Exact traffic accounting | DONE | ✓ | ✓ | ✓ | ✓ | P0 | Direct Naive exact-write telemetry is live on Production |
| 53 | Restart-safe accounting | DONE | △ | ✓ | △ | ✓ | P0 | boot/session/sequence/cumulative model present |
| 54 | Reconciliation | PARTIAL | △ | ✓ | △ | ✓ | P0 | Core completeness semantics exist; add operator reconciliation workflow/report |
| 55 | Hard quota enforcement | PARTIAL | ✓ | ✓ | ✓ | ✓ | P0 | Shared reservation/settlement core exists; complete controlled Production race/restart proof |
| 56 | Dashboard | PARTIAL | ✓ | ✓ | ✓ | ✓ | P1 | Customer dashboard exists; system/traffic KPIs incomplete |
| 57 | CPU | BLOCKED | ✓ | ✓ | △ | ✓ | P1 | Extract/rebuild useful PR #16 metrics safely |
| 58 | RAM | BLOCKED | ✓ | ✓ | △ | ✓ | P1 | Same |
| 59 | Disk | BLOCKED | ✓ | ✓ | △ | ✓ | P1 | Add live + historical metrics and warning thresholds |
| 60 | Network | BLOCKED | ✓ | ✓ | △ | ✓ | P1 | Add correct rate semantics, not cumulative-as-rate |
| 61 | Runtime health | PARTIAL | ✓ | ✓ | ✓ | ✓ | P1 | Service health exists; unify dashboard/read model |
| 62 | Traffic charts | BLOCKED | ✓ | ✓ | ✓ | ✓ | P1 | Add persisted samples after metric pipeline integration |
| 63 | Audit logs | PARTIAL | △ | ✓ | ✓ | ✓ | P1 | Backend audit exists; build explorer/filtering/history views |
| 64 | Application logs | BLOCKED | △ | ✓ | ✓ | ✓ | P1 | PR #16 candidate; redact secrets |
| 65 | Runtime logs | BLOCKED | ✓ | ✓ | ✓ | ✓ | P1 | PR #16 candidate; bounded/redacted |
| 66 | Security logs | BLOCKED | △ | ✓ | ✓ | ✓ | P1 | Add dedicated redacted security view |
| 67 | Diagnostics | BLOCKED | △ | ✓ | ✓ | ✓ | P1 | Extract safe PR #16 diagnostics/support bundle |
| 68 | Doctor | BLOCKED | △ | △ | ✓ | ✓ | P1 | Extract/rebuild command/page; verify dependencies and permissions |
| 69 | Backup | PARTIAL | ✓ | ✓ | ✓ | ✓ | P1 | Encrypted/manual DB backup scripts exist; product workflow/UI incomplete |
| 70 | Scheduled backup | BLOCKED | ✓ | ✓ | ✓ | ✓ | P1 | No PVNaive scheduled-backup timer observed on Production |
| 71 | Restore | PARTIAL | ✓ | ✓ | ✓ | ✓ | P1 | Script/test foundation exists; product-safe workflow incomplete |
| 72 | Restore verification | PARTIAL | △ | △ | ✓ | ✓ | P1 | CI drills exist historically; automate recurring verification/evidence |
| 73 | Notifications | PARTIAL | ✓ | ✓ | ✓ | ✓ | P1 | DB foundation exists; delivery engine/product UI incomplete |
| 74 | Telegram | BLOCKED | ✓ | ✓ | ✓ | △ | P1 | Implement secure bot/token handling and delivery |
| 75 | Notification rules | PARTIAL | ✓ | ✓ | ✓ | △ | P1 | Schema/routes foundation; handlers/worker/UI required |
| 76 | Reseller | PARTIAL | △ | ✓ | ✓ | △ | P1 | RLS/foundations exist; full CRUD/product workflow not ready |
| 77 | RBAC | PARTIAL | △ | ✓ | ✓ | △ | P1 | Complete route × role authorization matrix |
| 78 | Multi-admin | PARTIAL | ✓ | ✓ | ✓ | △ | P1 | Auth role foundations exist; admin management UI/API incomplete |
| 79 | Tenant isolation | DONE | — | ✓ | ✓ | △ | P0 | Preserve RLS + trigger guards; expand IDOR tests |
| 80 | Wallet / Credit | PARTIAL | — | △ | △ | — | P1 | Schema foundation only; implement reseller product flow |
| 81 | Financial ledger | PARTIAL | — | △ | △ | — | P1 | Schema foundation only; immutable audited operations required |
| 82 | Reseller plan restrictions | BLOCKED | — | ✓ | ✓ | — | P1 | Add allowed-plan/max-user/credit policy |
| 83 | REST API | PARTIAL | ✓ | ✓ | ✓ | ✓ | P1 | Real API exists but route manifest includes many not-ready contracts |
| 84 | OpenAPI / Swagger | BLOCKED | ✓ | ✓ | ✓ | △ | P1 | PR #16 candidate; generate from actual ready contracts |
| 85 | Webhook | BLOCKED | △ | ✓ | ✓ | △ | P2 | Add only after stable event contracts |
| 86 | API rate limit | PARTIAL | ✓ | ✓ | ✓ | ✓ | P1 | Security foundation exists; general policy incomplete; PR #16 candidate |
| 87 | MFA | PARTIAL | ✓ | ✓ | ✓ | — | P0 | Owner auth/TOTP foundation exists; reconcile ready routes and recovery policy |
| 88 | Session security | PARTIAL | ✓ | ✓ | ✓ | ✓ | P0 | Review refresh/reuse/session revocation correctness |
| 89 | Brute-force protection | PARTIAL | ✓ | ✓ | ✓ | ✓ | P0 | Finish trusted-proxy/IP-aware progressive policy |
| 90 | Recovery codes | BLOCKED | △ | ✓ | ✓ | — | P1 | Implement completely or declare unsupported |
| 91 | CSRF | PARTIAL | ✓ | ✓ | ✓ | ✓ | P0 | Perform explicit cookie/header/mutation audit |
| 92 | IDOR protection | PARTIAL | △ | ✓ | ✓ | △ | P0 | Tenant protections exist; complete full route × role negative tests |
| 93 | Multi-node | BLOCKED | ✓ | ✓ | △ | ✓ | P1 | Start only after standalone P0/P1 correctness |
| 94 | Node health | BLOCKED | ✓ | ✓ | △ | ✓ | P1 | Fleet phase |
| 95 | User → node assignment | BLOCKED | △ | ✓ | △ | ✓ | P1 | Fleet phase |
| 96 | Remote node deployment | BLOCKED | △ | ✓ | △ | ✓ | P1 | Fleet phase; outbound-safe auth |
| 97 | Drain | BLOCKED | △ | ✓ | △ | ✓ | P1 | Fleet phase |
| 98 | Maintenance | BLOCKED | △ | ✓ | △ | ✓ | P1 | Fleet phase |
| 99 | Canary | BLOCKED | — | △ | — | ✓ | P1 | Fleet rollout phase |
| 100 | Failover | BLOCKED | △ | ✓ | △ | ✓ | P1 | Fleet phase after deterministic desired/applied state |
| 101 | Load balancing | BLOCKED | △ | ✓ | △ | ✓ | P2 | Only if multi-node architecture requires it |
| 102 | Smart node selection | BLOCKED | — | △ | △ | ✓ | P2 | Health/load/bandwidth/latency/capacity policy |
| 103 | Fleet | BLOCKED | △ | ✓ | △ | ✓ | P1 | PR #16 foundation may be useful, but must be re-integrated safely |
| 104 | Fresh installer | PARTIAL | ✓ | ✓ | ✓ | ✓ | P0 | Current staged installers are upgrade-specific; build clean Ubuntu path |
| 105 | One-line install | BLOCKED | ✓ | ✓ | ✓ | ✓ | P0 | Secure version-pinned installer only |
| 106 | Upgrade | PARTIAL | ✓ | ✓ | ✓ | ✓ | P0 | Stage-specific upgrades exist; create generic versioned lifecycle |
| 107 | Rollback | PARTIAL | △ | ✓ | ✓ | ✓ | P0 | Runtime/stage rollback strong; generic product rollback incomplete |
| 108 | Uninstall | BLOCKED | ✓ | ✓ | ✓ | ✓ | P1 | Add safe removal preserving optional backups/data |
| 109 | Auto-update | BLOCKED | ✓ | ✓ | ✓ | △ | P2 | Optional; never silently mutate Production |
| 110 | Release signing | BLOCKED | ✓ | △ | ✓ | ✓ | P1 | Add provenance/signing policy |
| 111 | SBOM | BLOCKED | — | — | △ | △ | P1 | Generate release SBOM |
| 112 | SAST | BLOCKED | △ | △ | △ | △ | P1 | Add CI static security analysis |
| 113 | Secret scanning | BLOCKED | △ | △ | △ | △ | P1 | Add repository/CI secret scanning |
| 114 | Dependency vulnerability scanning | BLOCKED | △ | △ | △ | △ | P1 | Add Go/npm/container dependency scan |
| 115 | Load testing | PARTIAL | △ | △ | △ | ✓ | P0 | Bounded rehearsals exist; perform 50/100/200/400+ capacity campaign |
| 116 | Client compatibility | PARTIAL | ✓ | ✓ | ✓ | △ | P0 | Real Karing acceptance is still mandatory; document OS/version/results |
| 117 | Responsive UI | DONE | ✓ | ✓ | ✓ | ✓ | P1 | Baseline responsive UI exists |
| 118 | Mobile UI | PARTIAL | ✓ | ✓ | ✓ | △ | P1 | Real mobile/browser QA still required |
| 119 | Localization | PARTIAL | ✓ | ✓ | ✓ | △ | P2 | `/s` Persian/English exists; panel-wide i18n incomplete |
| 120 | Dark / light theme | PARTIAL | ✓ | ✓ | ✓ | △ | P2 | Account page supports system scheme; panel-wide user control/polish incomplete |

## Hiddify-specific / protocol-specific comparison

| Feature | PVNaive decision | Reason |
|---|---|---|
| Domain management | OPTIONAL | Useful only for managing PVNaive public/control domains; do not import Hiddify's multi-protocol domain model wholesale |
| CDN integration | PROTOCOL-SPECIFIC / OPTIONAL | Naive data plane should remain direct/DNS-only by default; CDN may be relevant only to web/control surfaces |
| Cloudflare automation | OPTIONAL | Potentially useful for panel/subscription DNS/control surface; not a requirement for Naive transport |
| WARP integration | N/A by default | Not part of current NaiveProxy standalone architecture or routing policy |
| Proxy-mode integrations / multi-protocol orchestration | N/A | PVNaive is intentionally NaiveProxy-first; do not add unrelated protocols for feature-count parity |

## Current PVNaive evidence summary

### Production read-only evidence, 2026-08-30

- `pvnaive-api.service`, `pvnaive-runtime-agent.service`, `pvnaive-telemetry-agent.service`, and `caddy-naive.service` active.
- API listens on loopback `127.0.0.1:8080`; local and public readiness returned HTTP 200.
- PostgreSQL schema version 11.
- six active users, six active Runtime credentials, six ServiceTerms, six active direct Subscription tokens.
- six direct-accounting term projections; all six reported `accounting_complete=true` at audit time.
- append-only direct accounting contained tens of thousands of events and active/session history records; legacy `usage_ledger` remained unused by this direct path.
- existing backup files are present, but no PVNaive scheduled-backup systemd timer was observed.
- root filesystem was 79% used at the audit point; capacity must be checked before large backup/load-test runs.
- Production provenance markers are stale/inconsistent with newer binary/web mtimes; this is an operational truth/provenance gap, not evidence that current runtime behavior is broken.

No secret, password, token or encryption key is recorded here.

## Highest-value gaps in required execution order

1. Finish CI/repository/Production truth reconciliation and merge refreshed status docs.
2. Extract only still-useful PR #16 units on a fresh branch.
3. Prove/adopt legacy accounting baselines without fake zeroes.
4. Wire exact accounting/presence consistently into `/s` and customer status views.
5. Manual Reset Usage, then Bulk Reset Usage.
6. Restart-safe periodic traffic reset scheduler.
7. Hard-quota race/restart/reconnect Production proof.
8. First-successful-CONNECT controlled Production proof.
9. Sessions, kill session, concurrency/IP limits and bounded history.
10. HWID/speed PoCs with capability gating.
11. Reseller CRUD/isolation/wallet/ledger/restrictions.
12. Customer history/audit explorer, notifications/Telegram, monitoring/logs/doctor, scheduled backup/restore, API/security, fleet, installer, client lab, load/capacity, final bulk/UI/docs.

## Licensing rule

3x-ui is GPL-3.0, PasarGuard is AGPL-3.0, and Hiddify Manager is GPL-3.0. Their source is used to verify behavior, architecture and UX patterns; code is not copied into PVNaive unless an explicit compatibility review approves it. OV-PvNetwork is MIT and may supply compatible operational patterns, but architecture differences still require review and tests.
