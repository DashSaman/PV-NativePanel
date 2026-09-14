# PVNaive — Canonical Agent Task Board

Last updated: 2026-09-14 (operator session Super-Z). **This is the single canonical board.**
The root-level `AGENT_TASKS.md` is a stale legacy queue from an earlier orchestration run —
it defers to this file. The root `ROADMAP.md` keeps the 50-row production-readiness ledger.

Owner program merged 2026-09-14 from the owner-provided `PVNaive_Agent_Prompt.md`
(Master Upgrade Pack R1→R8) and its Persian companion `PVNaive_Steering_FA.md`
(received inline by the operator session; placeholders closed).

**Rule for every agent:** before working read `AGENTS.md`, `PROJECT_STATUS.md`,
`FEATURE_MATRIX.md`, `KNOWN_ISSUES.md`, `docs/EXTENSIBILITY_FA.md`,
`docs/DATABASE_FA.md`, `SECURITY.md`, root `ROADMAP.md`, and this file. Follow the
`WORKLOG.md` protocol exactly. Never credit assignment-only / static-only / inferred
evidence as completion. **No "DONE" without linked evidence.**

## Standard pipeline (every feature/fix)

1. RED test first (Go `go test`, web `npm test`) proving the gap.
2. Implement minimal, boring, rollback-safe code.
3. CI-mirror verification on the dev server before pushing (golang:1.25 gofmt/vet/test +
   node:22 npm ci/test/build — identical to `.github/workflows/ci.yml`).
4. Live E2E on the deployed stack (real CONNECT / real API call), record output.
5. Commit (conventional message) → push `main` → confirm CI green.
6. Update `FEATURE_MATRIX.md` / `KNOWN_ISSUES.md` / `PROJECT_STATUS.md` / `WORKLOG.md` /
   root `ROADMAP.md` / this board.

## Section 1 — Owner requirements (six first-class, Master Upgrade Pack)

| ID | Requirement | Workstream(s) |
|---|---|---|
| A | Adaptive steering & rotation: per-user best-server selection from **server-side** latency/jitter/loss; time-windowed rotation; multi-format subscription; zero user interaction after one-time import | R1–R4 |
| B | Seamless UX contract "ONE STABLE INTERNET" (STEER-005): switch must be invisible — Instagram must not hang, video must not break; first-class gate equal to accounting correctness | R1–R4 |
| C | Pool Manager: 2..100+ nodes fully managed from the web UI, no hardcoded caps | R5 |
| D | Official-style cover site: root path of every node serves a distinct, natural, official-style Persian site (leader videos/messages + news syndication); per-server diversity defeats cross-node correlation | R6 |
| E | Panel access management: admin username, password, panel base path, panel port all changeable in-UI — safe, audited, zero lockout | R7 |
| F | Command-center UI: stealth animated login + live charts (per-user / per-node / fleet) à la internet-monitoring products; design methodology of ui-ux-pro-max-skill + 21st.dev inspiration (license-audited) | R8 |

## Section 2 — Master Upgrade Pack R1→R8 (workstreams, mechanisms, gates)

### R1 — Server-side network telemetry → GATE STEER-001
- Sample TCP_INFO on live accepted sockets every 5–10s (configurable) **inside the pinned
  forwardproxy** — the same trusted boundary that produces exact byte accounting:
  `tcpi_rtt`, `tcpi_rttvar`, `tcpi_segs_out`, `tcpi_segs_retrans` + bytes + duration.
- Keyed by trusted credential identity (Task12 boundary; authoritative peer = Caddy
  RemoteAddr — NEVER client headers/XFF).
- New table `session_network_samples` (monthly partitions, forward-only, checksummed,
  ledger discipline). Ingest via existing Telemetry Agent path: batched, compressed,
  idempotent, restart-safe (boot/session/sequence/cumulative semantics), no double count
  across reload/restart/reconnect.
- Aggregates per (user, node) with EWMA: `rtt_median_ms`, `rtt_jitter_ms`,
  `retrans_ratio`, `throughput_bps`, `session_success_rate`, `sample_count`.
- Gate: aggregates queryable per (user,node,window); restart/reload sims show zero
  double-count; fail-closed ingest; existing accounting untouched (WS1 Exact Accounting CI green).

### R2 — Scoring & steering engine → GATE STEER-002
- `score(user,node) = EWMA-weighted rtt rank − jitter penalty − retrans penalty − node
  load penalty` (+ optional ISP/country affinity only if trustworthy). Parameters in
  config, never hardcoded.
- Rotation: window-based (default **4h**); primary = best score; candidates ≥ 85% of best.
- Per-user phase offset `phase = hash(userID) mod window` → users never rotate simultaneously.
- EWMA low alpha (≈0.2) — single bad sample never switches a user (stability > reactivity).
- Hysteresis: switch only if candidate ≥ **25% better for 2 consecutive windows**.
- Kill-switch: node health failure → immediate failover, bypasses hysteresis.
- `Unknown` policy: pairs without sufficient samples are explicitly Unknown, excluded from
  scoring, never fabricated (ACCOUNTING-001 discipline). All decisions audit-logged (append-only, redacted).
- Gate: deterministic synthetic-sample simulation — no flapping, correct kill-switch,
  correct Unknown handling, phase-offset distribution.

### R3 — Time-aware multi-format subscription renderer → GATE STEER-003
- Extend machine `/sub/<token>` (human `/s/<token>` behavior unchanged) with User-Agent
  negotiation: Clash/Mihomo → YAML with `load-balance` (round-robin) + `url-test`
  (gstatic 204, interval 300, tolerance 50); sing-box/Karing → JSON urltest group;
  Hiddify → native format + update interval; v2rayNG/generic → base64 list **ordered by
  current steering decision**; naive raw → `naive://` of current primary + alternates.
- Mandatory headers: `profile-update-interval: 4`,
  `subscription-userinfo: upload=…; download=…; total=…; expire=…` (aggregate across nodes).
- Switch semantics: window boundaries only + **15-minute grace** where previous node stays
  valid; NEVER revoke credentials of a node with live sessions mid-window (Task13 kill is
  operator-only, never scheduler-driven). Atomic config apply (validate → apply → rollback),
  fail-closed, no-cache headers.
- Seamless rendering: Mihomo pools as `proxy-provider` (interval 4h + health-check) → hot
  updates, never full profile reload; NEVER `interrupt-exist-connections`; per-connection
  rotation; rendered configs hardcode node IPs + SNI (no DNS stalls/hijack).
- Gate: real client matrix — Hiddify, Karing, v2rayNG, a Mihomo client, naive raw — each
  imports one subscription, receives correct format, header-driven auto-update verified.

### STEER-005 — Seamless UX acceptance harness (first-class gate)
- Six mechanisms: per-connection rotation (established flows drain naturally); overlap
  validity (15-min grace, credentials replicated pool-wide); hot updates only; anti-flap
  stack (client tolerance 50ms + urltest 300 + EWMA α≈0.2 + 25%/2-window hysteresis);
  no-DNS (hardcoded IP+SNI); TLS cross-node resumption (shared session ticket keys via
  Runtime Agent secret path, rotated, encrypted-at-rest, never in repo).
- Acceptance: (a) 10+ min continuous HTTPS video-like download crossing a switch → zero
  failed requests, zero client-side RST, max stall < 1s; (b) node crash (kill -9 Caddy) →
  urltest failover within one health-check interval, no auth errors; (c) two nodes within
  15ms → zero oscillation over 24h simulated.

### R4 — Fleet-lite → GATE STEER-004
- Signed Node Manifest (Ed25519): `{node_id, endpoints, pubkey, weight, valid_from,
  valid_until, pool_id}` — verified before trust, no central controller required.
- Credential replication via existing Runtime Agent lifecycle over mTLS, outbound-only
  (no inbound ports on siblings beyond the data plane).
- Aggregate accounting: quota/remaining/usage = idempotent sum across nodes, restart-safe,
  no double counting.
- Degradation: sibling unreachable → standalone behavior fully functional locally
  (fail-closed steering, last-known-good manifest with expiry).
- Gate: kill a node mid-window in controlled rehearsal → invisible switch; aggregate bytes
  reconcile exactly (sum of nodes == ledger); no orphaned credentials.

### R5 — Pool Manager UI (2..100+ nodes) → GATE STEER-006
- Registry model (pull): primary panel stores desired state as **signed versioned
  revisions**; sibling agents PULL over mTLS (outbound-only, 30–60s poll). Registry
  unreachable → nodes keep serving last-known-good (expiry respected). Data plane never
  depends on the registry.
- UI (RBAC owner/admin, audit-logged, deny-by-default): add-node wizard (host/port/region/
  capacity/tags → one-time enrollment token → mTLS enrollment → preflight (version, reach,
  time skew) → health badge → join pool) — 2nd and 100th node are the SAME flow, zero CLI;
  node list with live RTT/loss/load/version/last-seen; pool CRUD (premium/bulk, per-pool
  window/grace/threshold); DRAIN workflow (stop steering → natural drain → revoke cert →
  manifest update; never yank a live node); bulk CSV/JSON import/export + bulk pool assignment.
- Per-user node subset rendering: mobile subs get TOP-K by score (K configurable, default
  10) — never 100 endpoints (battery); same token/URL/page, subset changes silently.
- Scale gate (100 nodes): manifest ≤256KB; 100-agent pull fan-out (60s) zero errors;
  credential propagation ≤90s; subscription render p99 <150ms; simultaneous drain of 5
  nodes without user-visible failure; registry outage → last-known-good serving.

### R6 — Official-style cover site (CAMO) → GATES CAMO-001..004
- Routing (strict order, per node): ① panel base path → panel (loopback) ② CONNECT/Naive
  data plane → pinned forwardproxy ③ everything else → `coverd` (new internal Go service,
  loopback-only, never exposed except through Caddy) or static persona file_server.
- Persona packs: ≥6 distinct ORIGINAL packs (state-adjacent news portal, cultural
  foundation, city services, regional news agency, research institute, charity). Each = own
  HTML structure, class vocabulary, CSS, palette, Persian RTL typography, JS bundle, logo,
  favicon, footer, sitemap shape, own RSS. Default `persona_id = stable_hash(node_id) mod
  packs`; per-node override + live preview in UI. Two nodes never structurally identical.
- Content pipeline: `coverd` aggregates PUBLIC content on schedule (default 6h ± jitter,
  per-source conditional GET, rate-limited, exponential backoff) into `cover_content`
  (partition/ledger discipline). Leader videos/messages: EMBED official players/iframes or
  link out — NEVER re-host videos on data-plane nodes; cache only thumbnails/titles/
  summaries/links. Source down → keep last snapshot, mark stale, keep serving (never
  empty/broken, never crash). Panel "Cover health" card (per-source status, last fetch age, staleness).
- Naturalness: Persian-first RTL, Jalali dates, local timezone, prayer-times widget, news
  slider, leader-messages section, about/contact/static pages, human 404/500, robots.txt,
  sitemap.xml, favicon, RECENT timestamps (frozen site = fingerprint). Response hygiene: no
  cookies, no CORS, no X-Powered-By, no identifying Server header, no panel-hinting CSP,
  zero panel path references, no directory listings, no default framework error pages.
- **Legal/opsec red line:** original "official-style" designs only — never clone a specific
  real organization's name/logo/seal or claim to BE it; syndicated items keep attribution.
- Gates: CAMO-001 non-panel paths all serve cover + zero panel artifacts (automated grep +
  header audit in CI); CAMO-002 probe-resistance sweep (random/sensitive/oversized requests
  indistinguishable from a normal Persian site); CAMO-003 persona diversity (DOM-hash
  distance above threshold) + content age <24h; CAMO-004 resilience (kill all sources →
  site stays up; politeness limits; zero crashes under fuzzed feed input).

### R7 — Panel access management (ACCESS) → GATES ACCESS-001..004
- Settings model: single-row `panel_settings` (admin_username, password_hash argon2id,
  base_path, listen_port, session_ttl, grace_minutes default 10, exposure_mode
  `reverse_proxy` default | `direct` discouraged). Every change append-only audited
  (actor, old→new, secrets redacted).
- Four flows (all: step-up auth = current password + CSRF + rate limit 5/min + audit):
  username (uniqueness); password (min 12 chars + zxcvbn ≥3, argon2id, rotate session
  secret, invalidate ALL other sessions keep actor's); base path (validate
  `^/[a-z0-9][a-z0-9\-_]{2,63}$`, deny reserved prefixes /api /assets /health cover routes,
  dual-accept old+new during grace then hard-deny old, ALL UI links prefix-relative); port
  (loopback-only behind Caddy; atomic rebind + Caddy admin-API upstream update + health
  verify + automatic rollback; `direct` mode exists with explicit warning + typed confirmation).
- No-lockout guarantees (lockout = defect): transactional apply (write → rebind/verify →
  commit; ANY failure → automatic rollback; never half-applied) + on-host recovery CLI
  `pvnaive admin reset-access` (username/password/path/port without UI), documented,
  recovery drill proven.
- Fleet note: admin credentials replicate over R4/R5 mTLS; base path/port may deliberately
  diverge per node (diversity is a feature); UI shows per-node access matrix.
- Gates: ACCESS-001 four changes from UI, old path dies after grace, actor session
  preserved/others invalidated, full audit; ACCESS-002 kill listener mid-change → automatic
  rollback, no lockout, no half-applied state; ACCESS-003 brute-force protection
  (rate limit/fail2ban on login) + password policy enforced; ACCESS-004 CLI recovery from
  simulated lockout < 5 minutes, documented.

### R8 — Command-center UI → GATES UI-001..004
- Design system FIRST: adopt ui-ux-pro-max-skill methodology (typography scale, 8pt grid,
  color science with semantic roles, WCAG AA, motion rules) → design tokens as code + docs.
  21st.dev inspiration license-audited (MIT/ISC/Apache-2.0 only, else rebuild from
  scratch); build a small internal component library — no pasted snippets. Persian RTL
  first-class (Vazirmatn-class font, mirrored layout, Jalali dates); English secondary.
  Stack decision via ADR (repo frontend exists — React/TS; charts: uPlot or ECharts,
  streaming-friendly).
- Stealth animated login: full-page GPU-cheap animated background (canvas gradient
  mesh/particles, honors prefers-reduced-motion); username/password INVISIBLE at rest (no
  boxes/placeholders/labels) — materialize on hover-reveal (150–250ms fade/slide) or
  keyboard focus (Tab + password-manager autofill MUST keep working); no hover outlines
  leaking locations; failed login causes no layout shift. Stealth is a UX layer, NOT the
  security control (real controls: ACCESS-003 rate limit/fail2ban + RBAC) — document this.
- Live monitoring: streaming via WebSocket or SSE (document reconnection, backpressure,
  drop-oldest): fleet total in/out bps area/line; per-node cards (sparkline + current bps +
  active sessions + health badge, expandable); per-user live consumption (bps, sessions,
  rolling 60s sparkline). RBAC on the stream endpoint itself (owner all, tenant self only,
  deny-by-default, 403 fail-closed). ~1s refresh, netdata/Grafana-live feel; smooth
  RENDERING only (never data); visible last-updated stamp; honest gaps; zoomable ranges.
  Derived numbers reconcile with exact-accounting ledger; nightly reconciliation job;
  Unknown renders as Unknown.
- Performance gate: dashboard with 100 nodes + 10k active sessions stays interactive
  (60fps, <250MB tab); bounded WS fan-out (ring buffers, drop-oldest per client, payload
  budget documented); R1 sampling overhead unchanged.
- Gates: UI-001 stealth behavior + keyboard/password-manager usable + reduced-motion;
  UI-002 live numbers reconcile (sum per-node == fleet total in tolerance; daily == ledger);
  UI-003 perf with documented numbers (fps, memory, WS payload/s, server CPU delta <2%);
  UI-004 Persian/RTL quality + RBAC isolation on live streams (tenant cannot subscribe to
  another tenant's stream).

## Section 3 — Gate index (18 gates)

| Gate | Workstream | Proof required |
|---|---|---|
| STEER-001 | R1 | aggregates per (user,node,window); zero double-count restart/reload sims; fail-closed; WS1 CI green |
| STEER-002 | R2 | deterministic simulation: no flapping, kill-switch, Unknown, phase offsets |
| STEER-003 | R3 | real client matrix (Hiddify/Karing/v2rayNG/Mihomo/naive) + auto-update headers |
| STEER-004 | R4 | mid-window node kill rehearsal: invisible switch + exact byte reconciliation |
| STEER-005 | R1–R4 | sustained-flow (zero fail, stall<1s) + hard-death + anti-flap 24h |
| STEER-006 | R5 | 100-node scale: manifest/fan-out/propagation/render p99/drain/registry-outage |
| CAMO-001 | R6 | cover on all non-panel paths; zero panel artifacts (automated audit) |
| CAMO-002 | R6 | probe sweep indistinguishable from normal site |
| CAMO-003 | R6 | persona DOM-hash diversity + content age <24h |
| CAMO-004 | R6 | sources down → last snapshot serves; fuzzed feeds zero crash |
| ACCESS-001 | R7 | 4 changes from UI; grace; session semantics; audit |
| ACCESS-002 | R7 | chaos: listener killed mid-change → auto-rollback, no lockout |
| ACCESS-003 | R7 | brute-force protection + password policy tested |
| ACCESS-004 | R7 | CLI recovery <5 min, documented |
| UI-001 | R8 | stealth login spec + accessibility fallback + reduced-motion |
| UI-002 | R8 | live totals reconcile (live + nightly vs ledger); honest Unknown |
| UI-003 | R8 | 100-node perf numbers (fps/memory/WS/CPU) |
| UI-004 | R8 | RTL quality + stream RBAC isolation tests |

## Section 4 — Hard constraints (non-negotiable checklist)

- [ ] Evidence culture: update `FEATURE_MATRIX.md`, `KNOWN_ISSUES.md` (new IDs
      STEER-001..006, CAMO-001..004, ACCESS-001..004, UI-001..004 + any new bugs),
      `PROJECT_STATUS.md`, `WORKLOG.md` per protocol; no claim without proof.
- [ ] GPL/AGPL competitor code is reference-only — never copy source. 21st.dev/community
      UI snippets require LICENSE AUDIT (MIT/ISC/Apache-2.0 only, else rebuild).
- [ ] Go + PostgreSQL; forward-only checksummed migration ledger; commit-before-success
      pattern (BUG-002 lesson) for all new authenticated mutations; fail-closed endpoints;
      RBAC deny-by-default; RLS tenant isolation on ALL new tables (cover content, panel
      settings, live-stream authorization included).
- [ ] No secrets, no production IPs/domains in code/docs/tests/logs — placeholder hosts
      only. OPSEC violation = change rejected. Keep hunting past-style leaks.
- [ ] Backwards compatibility: single-node deployments unchanged with steering and cover
      site disabled (both default OFF until their gates pass).
- [ ] UI honesty: no fabricated/interpolated values displayed as data; Unknown = Unknown.
- [ ] Performance: pprof before/after sampling; <3% CPU at 400 concurrent connections;
      UI-003 budget independent.
- [ ] Tests: race tests for new concurrent paths; restart/reload/reconnect double-count
      tests; R2 synthetic simulation; R3 real-client matrix; R4 controlled rehearsal with
      evidence; R6 probe-sweep + fuzzed-feed; R7 chaos/rollback drill; R8 perf + RBAC-stream.
- [ ] Documentation deliverables: `docs/STEERING_SPEC_FA.md` + `docs/CAMO_ACCESS_UI_SPEC_FA.md`
      (both Persian: DDL drafts, manifest JSON schema, EWMA/hysteresis parameter table,
      renderer contracts, coverd architecture, persona spec, panel-settings schema,
      stealth-login spec, live-dashboard contracts) + design tokens doc; update
      `docs/ROADMAP_FA.md` (insert program between phases 3 and 4).
- [ ] Out of scope (do NOT touch here): multi-runtime Xray/Hysteria2 adapter, HWID
      identity, per-user speed limit, Iran-specific scenario files.

## Section 5 — Deliverable order & Definition of Done

Order: spec docs (STEERING_SPEC_FA + CAMO_ACCESS_UI_SPEC_FA + design tokens) → R1 PR →
R2 PR → R3 PR → R4 PR → R5 PR → R6 PR → R7 PR → R8 PR (each: gates green → merge) →
PROJECT_STATUS.md final update + WORKLOG.md sections per task ID.

**Overall DoD:** multi-node test deployment managed entirely via web UI (2..100-node
simulated pool); one imported subscription yields automatic per-user best-node selection
from server-side measurements, silent 4h rotation with grace, invisible failure failover,
exact aggregate accounting, zero-touch node add/drain; every node's main domain serves a
distinct fresh official-style Persian cover site with zero panel leakage; admin
username/password/panel path/panel port changeable from UI with transactional apply,
auto-rollback, CLI recovery; command-center UI with stealth login + live fleet/per-node/
per-user charts reconciling with the ledger; STEER-005 harness and STEER-006 scale
rehearsal pass. User performs zero actions after import and experiences ONE stable internet.

## Section 6 — Operator live-ops batch status (server 45.141.148.59)

| # | Task | Status | Gate / next step |
|---:|---|---|---|
| 1 | GitHub push via fine-grained PAT + CI green | DONE | main `8792199`, all 5 checks success |
| 2 | Domain `namir.softarg.ir` + Let's Encrypt | BLOCKED (external) | LE rate-limit 429; Caddy auto-retries; LE retry-after **2026-09-15 03:26 UTC**. Then flip `/opt/pvnaive/.env` (`PVNAIVE_DOMAIN` + `PVNAIVE_NAIVE_PUBLIC_HOST=namir.softarg.ir:443`), recreate, E2E (dual-name block keeps clients working) |
| 3 | BBR + fq + sysctl tuning | DONE | `bbr` + `fq` verified live |
| 4 | Login link + credentials handed to owner | DONE | `https://45.141.148.59.nip.io/panel/` — `admin@pvnaive.local` (password in `/opt/pvnaive/.env`) |
| 5 | In-panel admin username/password change | DONE | `POST /api/v1/me/password`, `PATCH /api/v1/me/profile`, migration 0027/0028, `SettingsSecurity.tsx`, 7-step live E2E — R7 extends this to base path + port |
| 6 | Traffic accounting truth | NEXT | probe live DB counters vs real traffic; feeds R1 boundary |
| 7 | Competitor research (3x-ui v3.0→3.7 / PasarGuard / Hiddify) | DONE | `docs/competitor/3xui-v3-release-notes.md` |
| 8 | UI/UX overhaul "Amber Command Deck" | DONE (live) | R8 supersedes/extends with stealth login + live charts |
| 9 | Repo-documented progress | DONE (ongoing) | this board + AGENTS.md + ROADMAP.md + KNOWN_ISSUES.md |

## Section 7 — Competitor-informed feature backlog (maps into R-workstreams)

| # | Feature | Feeds |
|---:|---|---|
| 3 | Traffic accounting truth | R1 boundary |
| 4 | Per-client online status + realtime speed | R8 |
| 5 | DB backup/restore from panel + schedule/retention | R7 settings UX |
| 6 | Multi-format subscriptions (Clash/sing-box/Hiddify/UA auto) | R3 |
| 7 | Fail2ban-native IP limiting + trusted-IP exemptions | R7 / ACCESS-003 |
| 8 | Notification event bus (TG + SMTP + threshold alerts, per-event subscribe) | R5/R8 ops |
| 9 | Scoped expiring API tokens + in-panel OpenAPI | R4/R5 registry auth |
| 10 | Calendar-day renewals + per-client reset cycle | roadmap row 24 |
| 11 | Settings UX (shipped-default tags, 2FA re-confirm) | R7 |
| 12 | Accessibility pass | R8 |
| 13 | Dashboard throughput + TCP/UDP charts | R8 |
| 14 | Multi-node resilience | R4/R5 |
| 15 | Client lifecycle billing | roadmap row 24 |

## Section 8 — Master ledger cross-reference

P0 open rows in root `ROADMAP.md`: hard-quota production proof (10), first-CONNECT proof
(11), authorization/IDOR/CSRF/fuzz matrix (36), fresh secure Ubuntu installer (40),
clean-server install proof (48), production smoke (49), Release Candidate (50), Karing
client compatibility campaign (43), load/capacity campaign (44). Cite row numbers in PRs.

## Section 9 — Current execution pointer (pick up here)

DONE (2026-09-14 Super-Z session, all CI-mirror verified): spec docs (STEERING_SPEC_FA +
CAMO_ACCESS_UI_SPEC_FA), R2 engine (`internal/steering`), R3 renderer
(`internal/subscription/render.go`), R4 manifest (`internal/fleet/manifest.go`),
R6 coverd core + migration 0029 (`internal/coverd`, routing OFF), R8 design tokens +
stealth login (`web/src`). See WORKLOG.md 2026-09-14 entry for main hashes.

NEXT (in order):
1. R1 PR: forwardproxy TCP_INFO sampling → `session_network_samples` → aggregates
   (STEER-001; worker split in issue #109 — claim via issue before starting).
2. R7 PR: panel_settings model + base-path/port transactional apply + recovery CLI (ACCESS-001..004).
3. R8 PR: live-charts streaming backend (WS/SSE ring buffers, RBAC stream auth) + per-user/per-node cards.
4. R5 PR: pool manager UI on top of R4 manifest + R2 TopK (add-node wizard, drain, STEER-006 rehearsal).
5. R6 integration: coverd scheduler + Caddy routing flip behind a default-OFF flag → live CAMO gates.
6. Wire R2+R3 to live R1 aggregates (TopK into renderer, phase scheduler into /sub).
Live-ops lane: after LE window (2026-09-15 03:26 UTC) flip domain back to `namir.softarg.ir`
(Section 6 row 2); then traffic-accounting truth probe (row 6).

## 2026-09-14 01:24 UTC — production finalized on nip.io (owner instruction: domain flip moved to tomorrow)
- Owner instructed to finalize the panel on `https://45.141.148.59.nip.io/panel/` and postpone the
  `namir.softarg.ir` flip to the next day. LE 429 window unchanged (retry-after 2026-09-15 03:21 UTC).
- Built `pvnaive:repo-live2` (image `76c10697a03b`) on the server from canonical main `a4edea6`
  (clones at build time; content assertions: reconcile/R2/R3/R4/R6/R7/R8 present, migrations=30,
  SHA256SUMS verified). First deploy attempt exposed a real defect: the migration guard refused
  0029 (`DELETE FROM` inside a `$$` SECURITY DEFINER body) → container restart-loop → production
  was immediately rolled back to `pvnaive:repo-live` (schema stayed 28, data untouched, downtime
  ~4 minutes). Two fixes pushed and verified:
  - `6b361ff` (superseded by parallel fix `2168777` from the other coordinator lane, both compatible):
    migration destructive guard now lexes like PostgreSQL — dollar-quoted function bodies excluded
    from the migration-time scan, string literals preserved, unclosed dollar quotes fail closed;
    regression tests in `tests/db/migration_test.sh`; `tests/db/migration_test.sh` PASSED inside
    pvnaive-postgres18 against fresh main + patch.
  - `a4edea6` `internal/customer/validity.go`: `ValidityInput` JSON tags (mode/duration_days/expires_at).
    Live E2E caught `POST /api/v1/users` rejecting every payload with `validity.duration_days`
    (400 invalid_request) because `decodeRuntimeJSON` runs with `DisallowUnknownFields` and the
    struct was untagged — broke the UI timed-validity creation path. golang:1.25 CI-mirror:
    gofmt clean, vet ok, customer+httpapi tests ok; decode-contract regression tests added.
- Final deploy `repo-live2` healthy: schema 28→30 (0029 cover_site + 0030 panel_access applied,
  forward-only ledger), owner intact, credentials reconciled, readiness ready:true.
- Full external E2E **ALL_GREEN** (scripts/e2e_final_nip.py): strict-TLS LE cert for
  `45.141.148.59.nip.io`; panel 200; owner login 200; create customer 201 via `/api/v1/users`;
  subscription 200 with `Profile-Update-Interval: 4` + `Subscription-Userinfo` (R3 headers live);
  UA negotiation clash/sing-box/v2rayNG all 200; strict-TLS CONNECT through the proxy
  gstatic 204 + cloudflare 204.
- Cleanup: 7 e2e test customers revoked (200, history preserved). Note: customer lifecycle
  endpoints REQUIRE `Idempotency-Key` (first cleanup attempt without it returned 400).
- Production mutations this checkpoint: migrations 0029/0030 applied; test customers created and
  revoked. Schema at 30; previous `repo-live` image retained for instant rollback.

## 2026-09-14 02:53 coordinator checkpoint
- Verified canonical main `3b49e0b9dd10cd720dbbf33be50361f3ec003dce`; push CI `34789203279` SUCCESS.
- Task13 PR #108 refreshed by fast-forwarding its exact implementation history with current docs/spec main; new head `d42f1db4112fe43e71f4cd1b7feff941d78094af`, GitHub mergeable. Independent `git diff --check`, Docker Go 1.25 gofmt/vet/test and web 19/64 + build PASS. Fresh exact-head CI/accounting/forwardproxy are running; real pinned-Caddy HTTP/1.1 + HTTP/2 acceptance remains mandatory.
- Production remains mutation-free. Latest fresh external read-only probe: nip.io live/ready/panel healthy; `namir.softarg.ir` still inside documented Let's Encrypt retry window. Primary shell-level deployed identity/backups/rollback remain unverified.
- Opened #109 as the independent next roadmap lane: R1 / STEER-001 trusted-boundary network telemetry. Worker 3=forwardproxy sampling, Worker 2=DB/ingest/replay semantics, Worker 1=independent schema/CI/test-harness review, Worker 4=E2E rehearsal, Primary=read-only Production.
