### v3.0.0 (2026-05-10)
<h3>New</h3>

- [feat: Vue 3 migration — full frontend rewrite (#4198)](https://github.com/MHSanaei/3x-ui/commit/bc00d37a)
  - Stack: Vue 2 + Ant Design Vue 1 + Go HTML templates → Vue 3 + Ant Design Vue 4 + Vite 8 (multi-page bundles embedded into the Go binary via `web/dist/`)
  - Single-File Components with `<script setup>` and composables replacing global `Vue.component()` registrations and Vue 2 mixins
  - WebSocket-driven live updates across the panel (dashboard, inbounds, nodes) replace the legacy polling loop
  - Translation files migrated from TOML to JSON; vue-i18n 11 with per-locale code-split bundles for all 13 locales
  - Pages ported end-to-end: login, dashboard (with live status / xray / cpu-history / logs / backup / panel-update / custom-geo modals), settings (General / Security + 2FA / Telegram / Subscription), inbounds (list, search/filter, add/edit/delete/clone/reset, client modal + bulk-add, info + QR, expand-row), xray (Basics / Routing / Outbounds / Balancers / DNS / WARP / NordVPN), nodes
  - Inbound forms switched from raw JSON textareas to structured per-protocol/per-transport editors (an "Advanced" tab still exposes raw JSON for unsupported transports)
  - Multi-node deployment landed alongside the migration: new `runtime/{local,manager,remote}` layer, `node_heartbeat_job` + `node_traffic_sync_job`, central panel can deploy inbounds to remote node panels
  - Frontend bundle is split by route (login / xray / inbounds / settings / nodes) for faster initial load (LCP) on the dashboard
- [feat(nodes): multi-node deployment with traffic-writer queue, full-mirror sync, and WebSocket event fixes](https://github.com/MHSanaei/3x-ui/commit/8e7d215b)
- [feat(security): CSRF protection and security hardening across the application](https://github.com/MHSanaei/3x-ui/commit/10ebc6cb) @farhadh
- [feat(xray): add loopback outbound protocol](https://github.com/MHSanaei/3x-ui/commit/60e2af08)
- [feat(inbounds): mobile card layout for inbounds and clients](https://github.com/MHSanaei/3x-ui/commit/5ac88271)
- [feat(logs): mobile-friendly log modals with theme-aware colors](https://github.com/MHSanaei/3x-ui/commit/113a2973)
- [feat(custom-geo): refresh index UI](https://github.com/MHSanaei/3x-ui/commit/12c10dbd)
- [feat(xray/dns): expanded DNS settings with hosts editor, presets gallery, and Delete-All](https://github.com/MHSanaei/3x-ui/commit/a96612f5)

<h3>Update & improvement</h3>

- [refactor(websocket): split controller into service + thin controller](https://github.com/MHSanaei/3x-ui/commit/c394938f)
- [refactor(fallbacks): share template, tighter UX, cleaner JSON](https://github.com/MHSanaei/3x-ui/commit/28a3dddb)
- [refactor(xhttp): split fields by direction, expand outbound coverage](https://github.com/MHSanaei/3x-ui/commit/42b2ebc0)
- [refactor(vless): drop selectedAuth, expose two explicit auth buttons](https://github.com/MHSanaei/3x-ui/commit/3b64a621)
- [refactor(inbounds): reorder Inbound's Data tabs (client first, sub inline)](https://github.com/MHSanaei/3x-ui/commit/267fb1c8)
- [perf(xray): bound Xray-version request and extend cache](https://github.com/MHSanaei/3x-ui/commit/9735d26b)
- [i18n: localize sidebar theme toggle, xray-status badge, and nodes menu](https://github.com/MHSanaei/3x-ui/commit/cf5767ac)
- [avoid reset in QueryStatsRequest](https://github.com/MHSanaei/3x-ui/commit/14165fc5) @samssh
- [exclude virtual interfaces from network stats](https://github.com/MHSanaei/3x-ui/commit/ad302987)
- [outbound: reverse Sniffing](https://github.com/MHSanaei/3x-ui/commit/a8dff126)
- [inbound: check transport in port conflict, allow tcp and udp on same port](https://github.com/MHSanaei/3x-ui/commit/6a483fa9) @pwnnex
- [Reality: remove tesla.com because of blocking](https://github.com/MHSanaei/3x-ui/commit/f2bc4938)
- [skip Xray 26.5.3 and bump version cutoff](https://github.com/MHSanaei/3x-ui/commit/47163c14)
- [fix(scripts): harden server-IP detection with multi-provider + manual fallback](https://github.com/MHSanaei/3x-ui/commit/7f703f92)
- [Bump Go to 1.26.3](https://github.com/MHSanaei/3x-ui/commit/f2c79b57)
- [Axios v1.16.0](https://github.com/MHSanaei/3x-ui/commit/37fb48ff)
- [build frontend for CodeQL; remove release analyze job](https://github.com/MHSanaei/3x-ui/commit/439f4cf1)

<h3>Bug fixed</h3>

- [fix(panel): make webBasePath work end-to-end in dev and prod](https://github.com/MHSanaei/3x-ui/commit/61c84e82)
- [fix(panel): silence update-check WARN spam when offline](https://github.com/MHSanaei/3x-ui/commit/2fd2cd0a)
- [fix(panel-update): poll for restart, fix dark-mode version label](https://github.com/MHSanaei/3x-ui/commit/59c55dfc)
- [fix(websocket): guard stale events and disconnect race in JS client](https://github.com/MHSanaei/3x-ui/commit/b84b58ef)
- [fix(nodes): bind form-encoded posts and skip node inbounds in central xray](https://github.com/MHSanaei/3x-ui/commit/f70e131d)
- [fix(xray): align DNS outbound to spec and repair item-list rules UI](https://github.com/MHSanaei/3x-ui/commit/f68a14a3)
- [fix(xray): clear outbound test state on delete to prevent result bleed](https://github.com/MHSanaei/3x-ui/commit/9cbba130) @iAliF
- [fix(xray): surface reverse tags in routing and balancer dropdowns](https://github.com/MHSanaei/3x-ui/commit/917f9b30)
- [fix(xray): silently ignored error when saving outbound test URL setting](https://github.com/MHSanaei/3x-ui/commit/81b4ae56) @hobostay
- [fix(inbounds): remove stale reverse outbound tags after client deletion](https://github.com/MHSanaei/3x-ui/commit/c718e7ca)
- [fix(tun): use single mtu number per Xray spec](https://github.com/MHSanaei/3x-ui/commit/39bf31bd)
- [fix(vless): scope testseed to xtls-rprx-vision flow](https://github.com/MHSanaei/3x-ui/commit/79a7e7a5)
- [fix(warp): harden API client and frontend, bump to v0a4005](https://github.com/MHSanaei/3x-ui/commit/d8198f54)
- [fix(fail2ban): banning regression and Docker zero-jail issue](https://github.com/MHSanaei/3x-ui/commit/3349dcbc)
- [fix(security): overly permissive file permissions (os.ModePerm)](https://github.com/MHSanaei/3x-ui/commit/24cd2714) @hobostay
- [fix(security): silently ignored errors in password migration seeder](https://github.com/MHSanaei/3x-ui/commit/dee2525d) @hobostay
- [fix(docker): include web/translation in frontend and final stages](https://github.com/MHSanaei/3x-ui/commit/3505430e)
- [fix(x-ui.sh): pass silent flag to stop/start during IP SSL setup](https://github.com/MHSanaei/3x-ui/commit/72d8ebd2)
- [fix(arch): correct x-ui service path](https://github.com/MHSanaei/3x-ui/commit/e2649f98) @odrawq
- [fix(ui): mobile dashboard layout](https://github.com/MHSanaei/3x-ui/commit/b885a1f8)
- [fix(ui): correct responsive breakpoints for inbound form and settings](https://github.com/MHSanaei/3x-ui/commit/14781247)
- [fix(ui): correct responsive breakpoints for add client form and bulk](https://github.com/MHSanaei/3x-ui/commit/b776b334)
- [fix(ui): mobile filter view](https://github.com/MHSanaei/3x-ui/commit/7117d19f)
- [fix(outbound): mobile style](https://github.com/MHSanaei/3x-ui/commit/c88627a8)
- [fix: swap left/right classes for client table cells](https://github.com/MHSanaei/3x-ui/commit/33130860)
- [chore: fix shadowrocketUrl client](https://github.com/MHSanaei/3x-ui/commit/a1b23828) @harryngne
- [revert: Xray Core v26.5.3 (buggy — vless reverse broken)](https://github.com/MHSanaei/3x-ui/commit/03d8ad4d)

<h3> Reports </h3>

![total](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.0.0/total?label=total&color=success)
![amd64](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.0.0/x-ui-linux-amd64.tar.gz?label=linux-amd64)
![arm64](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.0.0/x-ui-linux-arm64.tar.gz?label=linux-arm64)
![386](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.0.0/x-ui-linux-386.tar.gz?label=linux-386)
![armv7](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.0.0/x-ui-linux-armv7.tar.gz?label=linux-armv7)
![armv6](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.0.0/x-ui-linux-armv6.tar.gz?label=linux-armv6)
![armv5](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.0.0/x-ui-linux-armv5.tar.gz?label=linux-armv5)
![s390x](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.0.0/x-ui-linux-s390x.tar.gz?label=linux-s390x)
![windows](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.0.0/x-ui-windows-amd64.zip?label=windows-amd64)

**Full Changelog**: https://github.com/MHSanaei/3x-ui/compare/v2.9.4...v3.0.0


### v3.0.1 (2026-05-11)
<h3>New</h3>

- [feat(frontend): refresh dark theme + redesign login page](https://github.com/MHSanaei/3x-ui/commit/c1efc486)
- [feat(inbounds): add sub/client link endpoints; hide panel version on login](https://github.com/MHSanaei/3x-ui/commit/6a90f98412dc91cc3ac230768a4d4428071ab16f)
- [feat(panel): in-panel API documentation page](https://github.com/MHSanaei/3x-ui/commit/e642f732)
- [feat(sidebar): pin Logout above trigger, inline 3-state theme cycle](https://github.com/MHSanaei/3x-ui/commit/b5479f3f)
- [feat(inbounds): bulk-select clients + UX polish](https://github.com/MHSanaei/3x-ui/commit/6d732d8d)
- [feat(xray/outbounds): TCP probe mode + Test All + timing breakdown](https://github.com/MHSanaei/3x-ui/commit/8834e5fb)
- [feat(xray/nord): searchable server list + colored load tag, surface API errors](https://github.com/MHSanaei/3x-ui/commit/5f3e9ed0)
- [feat(xray/balancer): restore observatory editor + auto-sync selectors](https://github.com/MHSanaei/3x-ui/commit/f1760b0a)
- [feat(install): add skip-SSL option for reverse-proxy / SSH-tunnel setups](https://github.com/MHSanaei/3x-ui/commit/e4900f1b)
- [feat(frontend): swap QRious for ant-design-vue's a-qrcode](https://github.com/MHSanaei/3x-ui/commit/04828246)

<h3>Update & improvement</h3>

- [refactor(panel): rename injected globals + collapse QR modal entries](https://github.com/MHSanaei/3x-ui/commit/745e394c)
- [add loopback and dns servers tag to inbound lists in RuleFormModal](https://github.com/MHSanaei/3x-ui/commit/e20d73ba) @samssh
- [chore: fix remarks shadowrocket subscription](https://github.com/MHSanaei/3x-ui/commit/9f06bffb) @harryngne

<h3>Bug fixed</h3>

- [fix(xray): implement graceful shutdown for xray process and add tests](https://github.com/MHSanaei/3x-ui/commit/9318c210) @farhadh
- [fix(inbounds): scope port check to node and preserve caller tag](https://github.com/MHSanaei/3x-ui/commit/7214ffaf)
- [fix(theme): default to dark, polish theme cycle visibility and hover](https://github.com/MHSanaei/3x-ui/commit/88061bac)
- [fix(inbounds): bulk-delete keeps last client to satisfy backend constraint](https://github.com/MHSanaei/3x-ui/commit/d8aedcdd)
- [fix(inbounds): paginate expanded client list, restore ID column, hide empty Remark](https://github.com/MHSanaei/3x-ui/commit/3e8a0eb9)
- [fix(alpine): restart_xray uses rc-service; OpenRC reload reads pidfile contents](https://github.com/MHSanaei/3x-ui/commit/4c291558)
- [fix(outbound): default VLESS encryption to "none"](https://github.com/MHSanaei/3x-ui/commit/737300b1)
- [fix: backup path with webbasepath](https://github.com/MHSanaei/3x-ui/commit/30469fcd) @GRCR13
- [fix(fail2ban): escape % in 3x-ipl action date format](https://github.com/MHSanaei/3x-ui/commit/887fca86)
- [fix(traffic-writer): replace sync.Once with Start/Stop cycle so SIGHUP restart works](https://github.com/MHSanaei/3x-ui/commit/8f3202f431373baab81545d8970236929676f71b)

<h3> Reports </h3>

![total](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.0.1/total?label=total&color=success)
![amd64](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.0.1/x-ui-linux-amd64.tar.gz?label=linux-amd64)
![arm64](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.0.1/x-ui-linux-arm64.tar.gz?label=linux-arm64)
![386](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.0.1/x-ui-linux-386.tar.gz?label=linux-386)
![armv7](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.0.1/x-ui-linux-armv7.tar.gz?label=linux-armv7)
![armv6](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.0.1/x-ui-linux-armv6.tar.gz?label=linux-armv6)
![armv5](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.0.1/x-ui-linux-armv5.tar.gz?label=linux-armv5)
![s390x](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.0.1/x-ui-linux-s390x.tar.gz?label=linux-s390x)
![windows](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.0.1/x-ui-windows-amd64.zip?label=windows-amd64)

**Full Changelog**: https://github.com/MHSanaei/3x-ui/compare/v3.0.0...v3.0.1


### v3.0.2 (2026-05-14)
<h3>New</h3>

- [feat: add API token to install output](https://github.com/MHSanaei/3x-ui/commit/033c5993) (#4322) @abdalrahmanx9
- [feat(panel): add Edit button to tables and enhance layout](https://github.com/MHSanaei/3x-ui/commit/194de886) (#4355) @BlackRockSoul
- [feat(json): swap raw textareas for a CodeMirror 6 JsonEditor](https://github.com/MHSanaei/3x-ui/commit/ce4c42e0)
- [feat(tabs): collapse settings and xray tab bars to evenly-spread icons](https://github.com/MHSanaei/3x-ui/commit/18614bd6)
- [feat(nodes): mobile card list, info modal, and tighter summary layout](https://github.com/MHSanaei/3x-ui/commit/e564c928)
- [feat(inbounds): collapse mobile cards to id/email + info button](https://github.com/MHSanaei/3x-ui/commit/933567d4)
- [feat(inbounds): align tunnel, tun, and hysteria UI with Xray docs](https://github.com/MHSanaei/3x-ui/commit/771bc7c8)
- [feat(api-tokens): manage multiple named tokens; add tab/section anchor URLs](https://github.com/MHSanaei/3x-ui/commit/b97ff40a)
- [feat(routing): drag-reorder rules, split balancer column, mobile card layout](https://github.com/MHSanaei/3x-ui/commit/46b6f8c6)
- [feat(ui): use the host as the browser tab title prefix](https://github.com/MHSanaei/3x-ui/commit/4e1b5979)
- [feat(api-docs): enhance in-panel API documentation](https://github.com/MHSanaei/3x-ui/commit/6e12329d) (#4312) @abdalrahmanx9
- [feat(nodes): blur address column with eye-toggle, mirroring IndexPage IP card](https://github.com/MHSanaei/3x-ui/commit/07bc74a5)
- [feat(inbounds): restore copy-clients-between-inbounds modal](https://github.com/MHSanaei/3x-ui/commit/80031e67)
- [Feat: clarify VLESS encryption auth selection](https://github.com/MHSanaei/3x-ui/commit/fdaa65ad) (#4271) @farhadh
- [feat: sortable inbounds table columns](https://github.com/MHSanaei/3x-ui/commit/89a8f549) (#4300) @abdalrahmanx9
- [feat(panel): xray metrics dashboard with observatory probe history](https://github.com/MHSanaei/3x-ui/commit/355bb4c9)

<h3>Update & improvement</h3>

- [style(api-docs): redesign TOC, section icons, endpoint rows, and code blocks with ultra-dark support](https://github.com/MHSanaei/3x-ui/commit/102df7a2) (#4332) @abdalrahmanx9
- [add: log rotate to 3xui.log file to avoid disk space consumption](https://github.com/MHSanaei/3x-ui/commit/4399fe2a) (#4277) @samssh
- [Security hardening: sessions, SSRF, CSP nonce, CSRF logout, trusted proxies](https://github.com/MHSanaei/3x-ui/commit/428f1333) (#4275) @farhadh
- [ci(codeql): run on push to main](https://github.com/MHSanaei/3x-ui/commit/3569b1be)
- [ci: gate workflows on relevant source paths](https://github.com/MHSanaei/3x-ui/commit/9fc47b3d)
- [Bump Go module dependency versions](https://github.com/MHSanaei/3x-ui/commit/210c25cf)

<h3>Bug fixed</h3>

- [fix(sub): include xhttp mode in extra JSON for karing compatibility](https://github.com/MHSanaei/3x-ui/commit/01a7dc80) (#4365) @abdalrahmanx9
- [fix(docker): update port mapping for 3xui service in docker-compose](https://github.com/MHSanaei/3x-ui/commit/6bf4a2c4) (#4362) @farhadh
- [fix(routing): make rule drag-and-drop work on mobile cards](https://github.com/MHSanaei/3x-ui/commit/21058eb6)
- [fix(qr): lock QR code modules to black-on-white across all themes](https://github.com/MHSanaei/3x-ui/commit/26accfd8)
- [Adjust QR panel sizing and collapse JSON subscription by default](https://github.com/MHSanaei/3x-ui/commit/2204c823)
- [fix(inbounds): refresh client rows live over websocket](https://github.com/MHSanaei/3x-ui/commit/2551a673)
- [fix(iplog): parse xray access-log timestamps in local time](https://github.com/MHSanaei/3x-ui/commit/61ab6028)
- [fix(warp): set license against Cloudflare API and surface errors inline](https://github.com/MHSanaei/3x-ui/commit/adc262a2)
- [fix(forms): validate JSON tabs before applying or saving](https://github.com/MHSanaei/3x-ui/commit/5543466f)
- [fix(inbounds): hide node UI when no enabled node exists](https://github.com/MHSanaei/3x-ui/commit/b10a9f1d)
- [fix(outbound): accept JSON-only configs and sync JSON to basic form on tab switch](https://github.com/MHSanaei/3x-ui/commit/6c6b40e0)
- [fix: single inbound traffic reset resets all inbounds](https://github.com/MHSanaei/3x-ui/commit/f29c8a5e) (#4334, #4338) @abdalrahmanx9
- [fix: strip main-panel TLS cert file paths when sending inbound to remote node](https://github.com/MHSanaei/3x-ui/commit/ad81649c) (#4339) @abdalrahmanx9
- [fix: reality random target/sni buttons not working](https://github.com/MHSanaei/3x-ui/commit/b47f794e) (#4337, #4340) @abdalrahmanx9
- [fix(auth): invalidate sessions when 2FA is enabled, fix dev 401 loop](https://github.com/MHSanaei/3x-ui/commit/bbefe910)
- [fix(inbound): require email when adding or updating a client](https://github.com/MHSanaei/3x-ui/commit/e40554a7)
- [fix(security): SSRF-guard node and remote HTTP clients](https://github.com/MHSanaei/3x-ui/commit/38da210d)
- [fix(api-docs): resolve no-useless-escape lint errors](https://github.com/MHSanaei/3x-ui/commit/406cb6db)
- [fix(fail2ban): escape percent signs in 3x-ipl datepattern](https://github.com/MHSanaei/3x-ui/commit/5fb36d34) (#4328) @usk2223
- [fix(graphs): increase y-axis paddingLeft from 32 to 56 to prevent clipped labels](https://github.com/MHSanaei/3x-ui/commit/4884a297) (#4309) @abdalrahmanx9
- [fix: delete button missing after searching for a user](https://github.com/MHSanaei/3x-ui/commit/9f7e8178) (#4315) @abdalrahmanx9
- [fix(hysteria2): restore missing masquerade config in inbound form](https://github.com/MHSanaei/3x-ui/commit/60e6b12f) (#4316) @abdalrahmanx9
- [fix: auto-renew must re-enable client in inbound settings JSON](https://github.com/MHSanaei/3x-ui/commit/0dbadf82) (#4317) @abdalrahmanx9
- [fix: show UDP tag for Hysteria and fix client count spacing](https://github.com/MHSanaei/3x-ui/commit/48e90bba) (#4318) @abdalrahmanx9
- [fix: preserve space between date and time in log modal](https://github.com/MHSanaei/3x-ui/commit/6de9b242) (#4326) @abdalrahmanx9
- [fix(api-docs): copy API token button](https://github.com/MHSanaei/3x-ui/commit/f570b991)
- [Fix: traffic writer restart freeze](https://github.com/MHSanaei/3x-ui/commit/d86e87ed) (#4265) @farhadh
- [fix(node): normalize base path during probe so missing trailing slash doesn't break status checks](https://github.com/MHSanaei/3x-ui/commit/9feeccff)
- [Add possibility to remove client email from sub](https://github.com/MHSanaei/3x-ui/commit/67b098df) (#4297) @Kasp42
- [fix: sync advancedJson before tab switch in convertLink](https://github.com/MHSanaei/3x-ui/commit/e7035b56fe8426fb2ed7f052911835e07b39ca06)
- [fix: ignore duplicate column errors during AutoMigrate on upgraded DBs](https://github.com/MHSanaei/3x-ui/commit/bd8d33980fd38ce3c9ec296025d646400d1fb505)

<h3> Reports </h3>

![total](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.0.2/total?label=total&color=success)
![amd64](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.0.2/x-ui-linux-amd64.tar.gz?label=linux-amd64)
![arm64](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.0.2/x-ui-linux-arm64.tar.gz?label=linux-arm64)
![386](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.0.2/x-ui-linux-386.tar.gz?label=linux-386)
![armv7](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.0.2/x-ui-linux-armv7.tar.gz?label=linux-armv7)
![armv6](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.0.2/x-ui-linux-armv6.tar.gz?label=linux-armv6)
![armv5](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.0.2/x-ui-linux-armv5.tar.gz?label=linux-armv5)
![s390x](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.0.2/x-ui-linux-s390x.tar.gz?label=linux-s390x)
![windows](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.0.2/x-ui-windows-amd64.zip?label=windows-amd64)

**Full Changelog**: https://github.com/MHSanaei/3x-ui/compare/v3.0.1...v3.0.2

### v3.1.0 (2026-05-23)
<h3>New</h3>

- [Frontend rewrite: React + TypeScript with AntD v6](https://github.com/MHSanaei/3x-ui/commit/edf0f369) (#4498) 
- [Feat/multi inbound clients](https://github.com/MHSanaei/3x-ui/commit/85e2ded0) (#4469) 
- [feat(bash): prompt for PostgreSQL](https://github.com/MHSanaei/3x-ui/commit/b71ed1e3ee634334c91815d88bb7d503715bd16e)
- [Bulk extend client expiry / traffic + clients page polish](https://github.com/MHSanaei/3x-ui/commit/9c60ed7e) (#4499) 
- [Add SockOpt.Mark and SockOpt.Interface parameters for Outbound stream](https://github.com/MHSanaei/3x-ui/commit/5f318f3b) (#4480) @githacs2022
- [Make HSTS policy configurable if https is enabled](https://github.com/MHSanaei/3x-ui/commit/758e1ad0) (#4462) @kayukin
- [feat(panel): copy connection strings for `mixed` inbound](https://github.com/MHSanaei/3x-ui/commit/121b6e0b) (#4450) @BlackRockSoul
- [feat(tgbot): add Flow picker when creating a VLESS client](https://github.com/MHSanaei/3x-ui/commit/2928b52b)
- [feat: click QR to copy/save image instead of link text](https://github.com/MHSanaei/3x-ui/commit/e4218a10)
- [feat(clients): add inbound filter + mobile page-size control](https://github.com/MHSanaei/3x-ui/commit/867a145979be5724becaedfea39543c30cf67d27)


<h3>Update & improvement</h3>

- [perf(frontend): lazy-load modals + split heavy vendor chunks](https://github.com/MHSanaei/3x-ui/commit/09df07dd) (#4501) 
- [Reduce list-page payloads with slim/paged endpoints](https://github.com/MHSanaei/3x-ui/commit/c5b71041) (#4500) 
- [i18n: translate hardcoded inbound action + security warning strings](https://github.com/MHSanaei/3x-ui/commit/95aebf1d) (#4502) 
- [fix(translation): correct typos and improve phrasing in English localization](https://github.com/MHSanaei/3x-ui/commit/f9ae0347) (#4430) @g3ntrix
- [fix: add i18n translations for Allow private address node option across all locales](https://github.com/MHSanaei/3x-ui/commit/19d50bd1) (#4386) @abdalrahmanx9
- [docs(readme): add Community Tools section](https://github.com/MHSanaei/3x-ui/commit/7065d41b) (#4114) @batonogov
- [refactor(inbounds): tighten advanced JSON helpers and fix dark-mode subtitles](https://github.com/MHSanaei/3x-ui/commit/5a101953)
- [refactor: remove legacy advancedJson state](https://github.com/MHSanaei/3x-ui/commit/b79abc8b)
- [Bump frontend deps: vue and vite](https://github.com/MHSanaei/3x-ui/commit/237b7c89)


<h3>Bug fixed</h3>

- [fix(frontend): override browser default background color on autofilled login inputs](https://github.com/MHSanaei/3x-ui/commit/6e2816d0) (#4478) @eric15342335
- [ix(clients): drop tombstone gate that blocked re-import after delete](https://github.com/MHSanaei/3x-ui/commit/6185db586a8d61205f5aa9aba91684cf6d7e7070)
- [fix(frontend): resolve lazy chunk URLs against runtime base path](https://github.com/MHSanaei/3x-ui/commit/c6123f96284272f73c6fde22ebad59d1c58e4889)
- [fix: parse XHTTP extra fields from V2Ray links and v2rayN JSON imports](https://github.com/MHSanaei/3x-ui/commit/fd3770c8) (#4426) @abdalrahmanx9
- [fix: prevent online clients from randomly disappearing from panel UI](https://github.com/MHSanaei/3x-ui/commit/78f1719c) (#4387) @abdalrahmanx9
- [fix: correct Hysteria2 Obfs password label to Auth password](https://github.com/MHSanaei/3x-ui/commit/f3c7660f) (#4388) @abdalrahmanx9
- [fix: protocol filter placeholder not showing on initial load](https://github.com/MHSanaei/3x-ui/commit/eacb9f63) (#4372) @abdalrahmanx9
- [fix(xray): resolve relative log paths under panel log folder](https://github.com/MHSanaei/3x-ui/commit/73683599)
- [fix(frontend): stack form fields on mobile in client/inbound/node modals](https://github.com/MHSanaei/3x-ui/commit/f2f5d584)
- [fix(sub): use standard sub://BASE64#REMARK scheme for Shadowrocket](https://github.com/MHSanaei/3x-ui/commit/9f80cfed)
- [fix(clients): honor global pageSize and widen size-changer dropdown](https://github.com/MHSanaei/3x-ui/commit/1b436bb3)
- [fix(migrate): include hysteria, hysteria2, shadowsocks in client sync](https://github.com/MHSanaei/3x-ui/commit/5b5ac3f0)
- [fix(clients): seed all clients when settings.clients has string tgId](https://github.com/MHSanaei/3x-ui/commit/3827d7d0)
- [fix(xray): allow private-IP destinations via freedom finalRules](https://github.com/MHSanaei/3x-ui/commit/d7f47d8b)
- [fix(security): redact at source and cap marshal sizes for CodeQL](https://github.com/MHSanaei/3x-ui/commit/b36e5e08)
- [fix(client): guard against int overflow in ClientWithAttachments marshal](https://github.com/MHSanaei/3x-ui/commit/788c979a)
- [fix(db): redact credentials in client-merge conflict logs](https://github.com/MHSanaei/3x-ui/commit/66f946ee)
- [fix(websocket): order register/unregister via single ops channel](https://github.com/MHSanaei/3x-ui/commit/6000bc71)
- [fix(inbounds): don't delete remote inbound when toggling enable](https://github.com/MHSanaei/3x-ui/commit/07cdb820)
- [fix(outbound): probe UDP-based outbounds over UDP instead of TCP](https://github.com/MHSanaei/3x-ui/commit/f00f82b3)
- [fix: disable balancer fallbackTag for random / roundRobin strategies](https://github.com/MHSanaei/3x-ui/commit/5cf8a085)
- [fix: split locale chunks by removing eager i18n glob](https://github.com/MHSanaei/3x-ui/commit/79a9be7b)
- [fix: Add base-path meta tag for Cloudflare Rocket Loader compatibility](https://github.com/MHSanaei/3x-ui/commit/3af45c14)
- [Remove streamSettings for protocols that don't support it](https://github.com/MHSanaei/3x-ui/commit/6badd829)
- [fix: remove Auth password](https://github.com/MHSanaei/3x-ui/commit/05b68c3b)
- [fix: guard certificate and key against undefined before join](https://github.com/MHSanaei/3x-ui/commit/9b0fd047)
- [fix(outbound): restore TLS, QUIC params and TCP masks when importing share links](https://github.com/MHSanaei/3x-ui/commit/1284756f)
- [fix: preserve TLS cert file paths when deploying inbound to remote node](https://github.com/MHSanaei/3x-ui/commit/1f052c0e)
- [fix: also hide QR code for ML-KEM-768 links (too long for QR generation)](https://github.com/MHSanaei/3x-ui/commit/ae6f13b5)
- [fix(clients): match by email when client identifier is stale](https://github.com/MHSanaei/3x-ui/commit/4c71669815dc93bffc3837de7196dead88dc8ceb)
- [fix: hide QR code for mldsa65 links (too long for QR generation)](https://github.com/MHSanaei/3x-ui/commit/1cf2582e)


<h3> Reports </h3>

![total](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.1.0/total?label=total&color=success)
![amd64](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.1.0/x-ui-linux-amd64.tar.gz?label=linux-amd64)
![arm64](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.1.0/x-ui-linux-arm64.tar.gz?label=linux-arm64)
![386](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.1.0/x-ui-linux-386.tar.gz?label=linux-386)
![armv7](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.1.0/x-ui-linux-armv7.tar.gz?label=linux-armv7)
![armv6](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.1.0/x-ui-linux-armv6.tar.gz?label=linux-armv6)
![armv5](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.1.0/x-ui-linux-armv5.tar.gz?label=linux-armv5)
![s390x](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.1.0/x-ui-linux-s390x.tar.gz?label=linux-s390x)
![windows](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.1.0/x-ui-windows-amd64.zip?label=windows-amd64)



**Full Changelog**: https://github.com/MHSanaei/3x-ui/compare/v3.0.2...v3.1.0


### v3.2.0 (2026-05-28)
<h3>New</h3>

- [feat(frontend): TanStack Query + React Router migration & in-panel API docs](https://github.com/MHSanaei/3x-ui/commit/cfe1b25c) (#4541)
- [Migrate frontend models/api/utils to TypeScript and modernize AntD theming](https://github.com/MHSanaei/3x-ui/commit/dc37f9b7) (#4563)
- [feat: complete Zod migration of frontend + bulk client batching](https://github.com/MHSanaei/3x-ui/commit/3f787ae1) (#4599)
- [feat(inbound): Advanced XHTTP and external TLS proxy settings](https://github.com/MHSanaei/3x-ui/commit/1f90d2a6) (#4491) @beehunt9r
- [feat(clients,groups): client groups + sub-links export + dedicated groups page](https://github.com/MHSanaei/3x-ui/commit/93eda068)
- [feat(clients): advanced filter drawer with multi-select state/protocol/inbound + expiry/usage ranges + auto-renew/tg/comment](https://github.com/MHSanaei/3x-ui/commit/3675f88c)
- [feat(clients): selective bulk attach + new bulk detach](https://github.com/MHSanaei/3x-ui/commit/72b68cce)
- [feat(inbounds): bulk-attach & assign-group client actions + form defaults](https://github.com/MHSanaei/3x-ui/commit/1a096d72)
- [feat(inbounds): row action to delete all clients of an inbound](https://github.com/MHSanaei/3x-ui/commit/e23599cb)
- [feat(clients,inbound): Auto Renew in Bulk Add + cleaner inbound wire payload](https://github.com/MHSanaei/3x-ui/commit/f1e433e8)
- [feat(settings): panel network proxy for the panel's own outbound requests](https://github.com/MHSanaei/3x-ui/commit/9d9737f4)
- [feat(tls): surface pinnedPeerCertSha256 in panel, share links, and subs](https://github.com/MHSanaei/3x-ui/commit/3f0b7fbe)
- [Random PostgreSQL role + post-install credentials display](https://github.com/MHSanaei/3x-ui/commit/058c030e) (#4608)
- [feat(inbound-form): salamander auto-seed for Hysteria + modernize random buttons](https://github.com/MHSanaei/3x-ui/commit/9d2a4f21)
- [feat(inbounds): expose Vision testseed field with sensible default](https://github.com/MHSanaei/3x-ui/commit/486ac9c2)
- [feat(inbounds): restore "Set Cert from Panel" / Clear buttons in TLS certs](https://github.com/MHSanaei/3x-ui/commit/9e005ffc)
- [feat(settings): include email in default remarkModel pattern](https://github.com/MHSanaei/3x-ui/commit/8d6d8452)
- [feat(clients): tidier bulk action toolbar + toolbar sort selector](https://github.com/MHSanaei/3x-ui/commit/bf1b488a)
- [feat(clients): restore Auto Renew field in client form](https://github.com/MHSanaei/3x-ui/commit/313d041d)
- [feat(fallbacks): add per-rule dest override](https://github.com/MHSanaei/3x-ui/commit/798e18b6eee74ef2034705d7937db78842023f68)


<h3>Update & improvement</h3>

- [refactor(clients): coherent group management — rename, split, extract](https://github.com/MHSanaei/3x-ui/commit/530e338c)
- [refactor(inbounds): cleaner network tags and cover Mixed/Tunnel + client form select polish](https://github.com/MHSanaei/3x-ui/commit/96a5c73e)
- [refactor(inbound-tag): node-prefixed + transport-suffixed canonical shape](https://github.com/MHSanaei/3x-ui/commit/7ade9d9a)
- [refactor(outbound): probe via xray burstObservatory instead of SOCKS round-trip](https://github.com/MHSanaei/3x-ui/commit/31d7ed51)
- [Client/inbound resilience + Postgres pool tuning + schema fixes](https://github.com/MHSanaei/3x-ui/commit/272854df) (#4607)
- [feat(port-conflict): include offending inbound + L4 in the error, cover quic and tunnel.allowedNetwork](https://github.com/MHSanaei/3x-ui/commit/980511bc)
- [i18n(panel): migrate hardcoded panel strings to en-US and translate all locales](https://github.com/MHSanaei/3x-ui/commit/72b97efa)
- [refactor(forms): modernize random buttons in client + outbound modals](https://github.com/MHSanaei/3x-ui/commit/43288e66)
- [refactor(metrics-modal): mark min/max on chart + improve grid contrast](https://github.com/MHSanaei/3x-ui/commit/2bba1d21)
- [change tg message when send qrCode](https://github.com/MHSanaei/3x-ui/commit/0829f1ec) (#4623) @sb15551


<h3>Bug fixed</h3>

- [Fix REALITY share links missing SNI](https://github.com/MHSanaei/3x-ui/commit/c03ecfe6) (#4621) @pooyaww
- [fix(groups): fetch full client list for Add/Remove/SubLinks modals](https://github.com/MHSanaei/3x-ui/commit/ffe661d2)
- [fix(clients): backfill missing subId on startup and guard create/update](https://github.com/MHSanaei/3x-ui/commit/99df5d70)
- [fix(inbounds): heal legacy client data and TLS cert form hydration](https://github.com/MHSanaei/3x-ui/commit/b42a4d93)
- [fix(links): include TCP HTTP host header in share links](https://github.com/MHSanaei/3x-ui/commit/8046d151)
- [fix(clients): avoid duplicate ClientRecord when email is changed on edit](https://github.com/MHSanaei/3x-ui/commit/5eb80eca)
- [fix(sub): preserve userinfo encoding in trojan/shadowsocks/hysteria links](https://github.com/MHSanaei/3x-ui/commit/3c5e9fa7)
- [fix(remote-traffic): handle tag collisions + readable warning format](https://github.com/MHSanaei/3x-ui/commit/d3476052)
- [fix: address open bug reports (#4539, #4538, #4535, #4531, #4515)](https://github.com/MHSanaei/3x-ui/commit/19e88c46) (#4545)
- [fix(ui): polish across routing, groups, inbounds, mobile sidebar](https://github.com/MHSanaei/3x-ui/commit/2fea7138)
- [fix(sub): stop external-proxy dest from clobbering TLS SNI](https://github.com/MHSanaei/3x-ui/commit/cda7f2ac17ab7d46c7fee2168532f1e78e31c8eb)
- [fix(inbounds): restore xHTTP Headers editor in form](https://github.com/MHSanaei/3x-ui/commit/b395a1b951ec78d821caa43b711d7adc2b8b1705)


<h3> Reports </h3>

![total](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.2.0/total?label=total&color=success)
![amd64](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.2.0/x-ui-linux-amd64.tar.gz?label=linux-amd64)
![arm64](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.2.0/x-ui-linux-arm64.tar.gz?label=linux-arm64)
![386](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.2.0/x-ui-linux-386.tar.gz?label=linux-386)
![armv7](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.2.0/x-ui-linux-armv7.tar.gz?label=linux-armv7)
![armv6](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.2.0/x-ui-linux-armv6.tar.gz?label=linux-armv6)
![armv5](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.2.0/x-ui-linux-armv5.tar.gz?label=linux-armv5)
![s390x](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.2.0/x-ui-linux-s390x.tar.gz?label=linux-s390x)
![windows](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.2.0/x-ui-windows-amd64.zip?label=windows-amd64)



**Full Changelog**: https://github.com/MHSanaei/3x-ui/compare/v3.1.0...v3.2.0


### v3.2.5 (2026-06-01)
<h3>New</h3>

- [feat(postgres): in-panel backup/restore and consistent CLI backend](https://github.com/MHSanaei/3x-ui/commit/cc34dc38)
- [feat(nodes): bulk panel self-update with live online indicator](https://github.com/MHSanaei/3x-ui/commit/971843f6)
- [feat(inbounds): multi-select and bulk delete](https://github.com/MHSanaei/3x-ui/commit/cf509529)
- [feat(inbounds): attach existing clients to an inbound in one click](https://github.com/MHSanaei/3x-ui/commit/dd14e9b3)
- [feat(clients): live online dot + last-online tooltip on offline](https://github.com/MHSanaei/3x-ui/commit/c8df1b19)
- [feat(clients): enforce unique subId per client like email](https://github.com/MHSanaei/3x-ui/commit/88a36773)
- [feat(clients/inbounds): IP log popups, clearer titles, tag-based inbound labels](https://github.com/MHSanaei/3x-ui/commit/987a6dd1)
- [feat(finalmask): sync transport with upstream Xray core changes](https://github.com/MHSanaei/3x-ui/commit/32f96298)
- [feat(outbound): sync DNS outbound config with Xray core changes](https://github.com/MHSanaei/3x-ui/commit/2bb9ed1c)
- [feat(sub): add HEAD method support for subscription endpoints](https://github.com/MHSanaei/3x-ui/commit/84a689cf) (#4684) @spokyle
- [feat(inbounds): clearer client validation errors on save](https://github.com/MHSanaei/3x-ui/commit/76dbbfc1)

<h3>Update & improvement</h3>

- [refactor(frontend): reorganize source tree & break down oversized modals/tabs](https://github.com/MHSanaei/3x-ui/commit/d1882c7f) (#4698)
- [chore: bump bundled Xray-core to v26.6.1](https://github.com/MHSanaei/3x-ui/commit/51d383b1)
- [refactor(inbounds): remove column sorter from inbound list](https://github.com/MHSanaei/3x-ui/commit/61e8bed3)
- [i18n(nodes): translate basePath and apiToken labels](https://github.com/MHSanaei/3x-ui/commit/7028c15e)
- [Update Go module dependency versions](https://github.com/MHSanaei/3x-ui/commit/eee5e8f6)

<h3>Bug fixed</h3>

- [fix(outbounds): preserve TLS/Reality security on save](https://github.com/MHSanaei/3x-ui/commit/df777c12)
- [fix(outbounds): lock hysteria to its QUIC transport + TLS, add version/masquerade](https://github.com/MHSanaei/3x-ui/commit/eee26e47)
- [fix(outbounds): prevent freedom save crash, complete its fields](https://github.com/MHSanaei/3x-ui/commit/f02018cf) (#4686)
- [fix(outbounds): parse wireguard:// links and fix ss:// query-string port](https://github.com/MHSanaei/3x-ui/commit/12afb862)
- [fix(outbounds): support proxyProtocol on freedom outbound](https://github.com/MHSanaei/3x-ui/commit/62c293e0)
- [fix(xray): test UDP outbounds via xray probe + Vision testseed & Flow form fixes](https://github.com/MHSanaei/3x-ui/commit/cb7af04c) (#4657)
- [fix(inbounds): preserve client data on delete and show traffic in detail](https://github.com/MHSanaei/3x-ui/commit/6bb5a3b5)
- [fix(inbounds): auto-increment WireGuard peer IP](https://github.com/MHSanaei/3x-ui/commit/9d99428c)
- [fix(model): accept tun protocol in inbound validation](https://github.com/MHSanaei/3x-ui/commit/8db97299)
- [fix(clients): store flow per-inbound for shared clients](https://github.com/MHSanaei/3x-ui/commit/7ea88e3e)
- [fix(clients): preserve UUID when toggling enable from clients page](https://github.com/MHSanaei/3x-ui/commit/8e301dbc)
- [fix(clients): persist group for node-inbound clients](https://github.com/MHSanaei/3x-ui/commit/24d0e4ec)
- [fix: reject spaces, slashes and control chars in client email, subId and URI path](https://github.com/MHSanaei/3x-ui/commit/a0865a67)
- [fix(ssl): prompt before setting IP cert path for panel](https://github.com/MHSanaei/3x-ui/commit/90a64a1b)
- [fix(qr): hide QR for post-quantum links on client QR page](https://github.com/MHSanaei/3x-ui/commit/5d0081a3)
- [fix(sub): keep listen/bind IP out of subscription links and pages](https://github.com/MHSanaei/3x-ui/commit/fb311afa)
- [fix(ui): exit infinite spinner with a retry card on failed initial load](https://github.com/MHSanaei/3x-ui/commit/b9cbc0c1)
- [fix(postgres): resync id sequences so adding clients no longer collides](https://github.com/MHSanaei/3x-ui/commit/e8c6c309)
- [fix(postgres): stop FK constraint from blocking inbound delete](https://github.com/MHSanaei/3x-ui/commit/998fa0df)
- [fix(postgres): record client traffic when inbound_id is stale](https://github.com/MHSanaei/3x-ui/commit/3f5e37b0)

<h3> Reports </h3>

![total](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.2.5/total?label=total&color=success)
![amd64](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.2.5/x-ui-linux-amd64.tar.gz?label=linux-amd64)
![arm64](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.2.5/x-ui-linux-arm64.tar.gz?label=linux-arm64)
![386](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.2.5/x-ui-linux-386.tar.gz?label=linux-386)
![armv7](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.2.5/x-ui-linux-armv7.tar.gz?label=linux-armv7)
![armv6](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.2.5/x-ui-linux-armv6.tar.gz?label=linux-armv6)
![armv5](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.2.5/x-ui-linux-armv5.tar.gz?label=linux-armv5)
![s390x](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.2.5/x-ui-linux-s390x.tar.gz?label=linux-s390x)
![windows](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.2.5/x-ui-windows-amd64.zip?label=windows-amd64)

**Full Changelog**: https://github.com/MHSanaei/3x-ui/compare/v3.2.0...v3.2.5


### v3.2.6 (2026-06-02)
<h3>New</h3>

- [feat(outbounds): pick dialerProxy from other outbound tags for proxy chaining](https://github.com/MHSanaei/3x-ui/commit/7bc31dd1)
- [feat(nodes): add per-node TLS verification mode for self-signed certs](https://github.com/MHSanaei/3x-ui/commit/56ec3590) (#4757)
- [feat(inbounds): support Unix domain socket path in Listen field](https://github.com/MHSanaei/3x-ui/commit/b2e2120e) (#4429)
- [feat(x-ui.sh): support Cloudflare API Token for DNS SSL (menu 20)](https://github.com/MHSanaei/3x-ui/commit/cb17eb8c) (#4595)
- [feat(x-ui.sh): add PostgreSQL management menu](https://github.com/MHSanaei/3x-ui/commit/47d9b496)

<h3>Update & improvement</h3>

- [perf(clients): batch bulk attach/detach to cut per-item DB work](https://github.com/MHSanaei/3x-ui/commit/4f597a08)
- [docs(readme): revamp README and sync all translations](https://github.com/MHSanaei/3x-ui/commit/87f446fe)
- [chore(generated): sync node types/zod with TLS verification fields](https://github.com/MHSanaei/3x-ui/commit/01d2ec50) (#4757)
- [chore(ui): remove cards jump on hover](https://github.com/MHSanaei/3x-ui/commit/6b2243a4) (#4755) @fgsfds
- [Replace static label with translation for downlink stats](https://github.com/MHSanaei/3x-ui/commit/f9aa363a) (#4762) @ckun52880
- [Remove .svg extension from shields URLs in READMEs](https://github.com/MHSanaei/3x-ui/commit/327228d8)

<h3>Bug fixed</h3>

- [fix(migrate): copy composite-key tables without FindInBatches](https://github.com/MHSanaei/3x-ui/commit/c8ad4263) (#4787)
- [fix(node): suppress unavoidable InsecureSkipVerify alert for cert pinning](https://github.com/MHSanaei/3x-ui/commit/f0e459e5)
- [fix(node): capture node cert via VerifyConnection for fingerprint fetch](https://github.com/MHSanaei/3x-ui/commit/d2dc589f)
- [fix(clients): keep Add Client modal in viewport with internal scroll](https://github.com/MHSanaei/3x-ui/commit/49ef1449)
- [fix(xray): clear dirty state after saving unchanged config](https://github.com/MHSanaei/3x-ui/commit/b9612f13)
- [fix(job): skip fail2ban IP limit when disabled](https://github.com/MHSanaei/3x-ui/commit/8fa248c6) (#4581) @Mayurifag
- [fix(fallbacks): allow free-form dest entries for external servers](https://github.com/MHSanaei/3x-ui/commit/49bec1db) (#4748)
- [fix(raw): complete the HTTP header section for inbound and outbound](https://github.com/MHSanaei/3x-ui/commit/5b6e05a0)
- [fix(x-ui.sh): preserve 2FA on credential reset](https://github.com/MHSanaei/3x-ui/commit/bcb982ae) (#4758)
- [fix(inbounds): allow port 0 for UDS inbounds](https://github.com/MHSanaei/3x-ui/commit/ccd0853b) (#4783)
- [fix(warp): persist client_id so WARP outbound gets reserved bytes](https://github.com/MHSanaei/3x-ui/commit/3657ed55) (#4781)
- [fix(nodes): sum client traffic across nodes instead of overwriting](https://github.com/MHSanaei/3x-ui/commit/5b9ed340)
- [fix(hysteria): use pinSHA256 for pinned cert and emit ech in share links](https://github.com/MHSanaei/3x-ui/commit/588ea862)
- [fix(sub): source Userinfo total/expiry from client config in multi-node](https://github.com/MHSanaei/3x-ui/commit/7f8c7967) (#4645)
- [fix(db): make password-hash migration idempotent to prevent lock-out](https://github.com/MHSanaei/3x-ui/commit/80173b1b) (#4612)
- [fix(outbound): add None option to uTLS fingerprint in TLS form](https://github.com/MHSanaei/3x-ui/commit/6ae1b386) (#4760)
- [fix(outbound): carry ALPN, fingerprint and UDP mask when importing a Hysteria2 link](https://github.com/MHSanaei/3x-ui/commit/803e0109) (#4760)
- [fix(sockopt): rename interfaceName to interface so xray honors it](https://github.com/MHSanaei/3x-ui/commit/b6641439)
- [fix(sub): ensure unique Clash proxy names](https://github.com/MHSanaei/3x-ui/commit/d29a17d3) (#4641)
- [fix(settings): enforce trafficDiff max of 100 in UI](https://github.com/MHSanaei/3x-ui/commit/39b71640) (#4769)
- [fix(outbound): fill encryption and pqv when importing VLESS link](https://github.com/MHSanaei/3x-ui/commit/13c04bb9)
- [fix(docker): grant NET_ADMIN/NET_RAW so fail2ban IP-limit bans apply](https://github.com/MHSanaei/3x-ui/commit/28330e60)
- [Fix IP limit enforcement and clarify related comments](https://github.com/MHSanaei/3x-ui/commit/16edb037) (#4699) @ALOKY
- [fix(sub): Add Clash subscription profile filename header](https://github.com/MHSanaei/3x-ui/commit/2b7c1eeb) (#4743) @xiaoxiyao

<h3> Reports </h3>


![total](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.2.6/total?label=total&color=success)
![amd64](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.2.6/x-ui-linux-amd64.tar.gz?label=linux-amd64)
![arm64](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.2.6/x-ui-linux-arm64.tar.gz?label=linux-arm64)
![386](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.2.6/x-ui-linux-386.tar.gz?label=linux-386)
![armv7](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.2.6/x-ui-linux-armv7.tar.gz?label=linux-armv7)
![armv6](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.2.6/x-ui-linux-armv6.tar.gz?label=linux-armv6)
![armv5](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.2.6/x-ui-linux-armv5.tar.gz?label=linux-armv5)
![s390x](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.2.6/x-ui-linux-s390x.tar.gz?label=linux-s390x)
![windows](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.2.6/x-ui-windows-amd64.zip?label=windows-amd64)

**Full Changelog**: https://github.com/MHSanaei/3x-ui/compare/v3.2.5...v3.2.6


### v3.2.7 (2026-06-03)
<h3>New</h3>

- [feat(dashboard): richer System History & Xray Metrics charts](https://github.com/MHSanaei/3x-ui/commit/4b11c542)
- [feat(dashboard): more System History metrics, persistence & localized labels](https://github.com/MHSanaei/3x-ui/commit/d4c020f3)
- [feat(xray): merge basic routing into the routing rules section](https://github.com/MHSanaei/3x-ui/commit/a4dae566)
- [feat(xray): add connIdle and bufferSize policy controls](https://github.com/MHSanaei/3x-ui/commit/ceef413d)
- [feat(settings): sidebar submenu nav for settings and xray with icon tabs](https://github.com/MHSanaei/3x-ui/commit/ac89ec72)
- [feat(settings): move the remark model control to the subscription tab](https://github.com/MHSanaei/3x-ui/commit/e63cde8f)
- [feat(inbounds): per-proxy Pinned Peer Cert SHA-256 + labeled External Proxy form](https://github.com/MHSanaei/3x-ui/commit/e7c11c91)
- [feat(tls): add ocspStapling to certificate config](https://github.com/MHSanaei/3x-ui/commit/1a64d7e9)
- [feat(links): richer share-link labels across QR, client info and sub views](https://github.com/MHSanaei/3x-ui/commit/d0998c1d)
- [feat(hysteria2): emit UDP port hopping in subscriptions and share links](https://github.com/MHSanaei/3x-ui/commit/13d02f01) (#4789)
- [feat(clients): show filtered count in clients list](https://github.com/MHSanaei/3x-ui/commit/91f325ec) (#4808)
- [feat(clients,routing): label inbounds by remark with tag fallback](https://github.com/MHSanaei/3x-ui/commit/61105c2b)

<h3>Update & improvement</h3>

- [chore(deps): bump xray-core to v1.260327.1 and add pion/wireguard deps](https://github.com/MHSanaei/3x-ui/commit/72944daa)
- [chore(frontend): bump deps to 0.2.7 and hide node row selection for single node](https://github.com/MHSanaei/3x-ui/commit/dc57c1e9)
- [i18n: translate connection-limit strings for all languages](https://github.com/MHSanaei/3x-ui/commit/7a72aeda)
- [fix(sidebar): set fixed sider width to 220](https://github.com/MHSanaei/3x-ui/commit/c7828540)
- [fix(ci): bump Go to 1.26.4 and exempt /panel/groups SPA route from api-docs test](https://github.com/MHSanaei/3x-ui/commit/039d05a7)
- [ci(issue-bot): ground the assistant in repo source with an investigation step](https://github.com/MHSanaei/3x-ui/commit/f6d4358f)

<h3>Bug fixed</h3>

- [fix(api-token): hash tokens at rest and show plaintext only once](https://github.com/MHSanaei/3x-ui/commit/4813a2fe)
- [fix(online): scope per-inbound online to inbounds that carried traffic](https://github.com/MHSanaei/3x-ui/commit/ef8882a5) (#4859)
- [fix(nodes): Set Cert from Panel uses the node's own web cert for node inbounds](https://github.com/MHSanaei/3x-ui/commit/55d67299) (#4854)
- [fix(panel): register /groups SPA route so hard refresh returns index.html](https://github.com/MHSanaei/3x-ui/commit/ccfd0421) (#4837)
- [fix(clients): keep reverse tag clearable and preserve flow on attach](https://github.com/MHSanaei/3x-ui/commit/b08fc0c9) (#4834)
- [fix(settings): fall back to defaults for empty/NULL setting values](https://github.com/MHSanaei/3x-ui/commit/fcc6787a) (#4830)
- [fix(links): use configured domain for panel copy/QR links on loopback](https://github.com/MHSanaei/3x-ui/commit/6ee462ac) (#4829)
- [fix(docker): make x-ui CLI menu work inside containers](https://github.com/MHSanaei/3x-ui/commit/f901cd42) (#4817)
- [fix(hysteria2): emit pinSHA256 as hex in subscriptions, not base64](https://github.com/MHSanaei/3x-ui/commit/ac67c522) (#4818)
- [fix(online): scope online status per node instead of a global union](https://github.com/MHSanaei/3x-ui/commit/3af2da01) (#4809)
- [fix(iplimit): populate client IP log without an IP limit](https://github.com/MHSanaei/3x-ui/commit/66d4d047) (#4800)
- [fix(sub): advertise routable inbound Listen in subscription links](https://github.com/MHSanaei/3x-ui/commit/a40d85ce) (#4798)
- [fix(outbounds): preserve SNI/TLS settings on transport change](https://github.com/MHSanaei/3x-ui/commit/5fb18b88) (#4791)
- [fix(clients): derive edit-form flow from per-inbound override](https://github.com/MHSanaei/3x-ui/commit/1e3c186b) (#4792)
- [fix(tls): correct pinned cert SHA-256 hint to hex, not base64](https://github.com/MHSanaei/3x-ui/commit/c9abda7a) (#4793)
- [fix(node): fix "invalid input" on save and gate save on connectivity](https://github.com/MHSanaei/3x-ui/commit/02043a43) (#4794)
- [fix(xray): default freedom finalRules to allow-all so reverse egress works](https://github.com/MHSanaei/3x-ui/commit/8f5a7b94)
- [fix(migrate): relax legacy freedom finalRules so reverse egress works on existing installs](https://github.com/MHSanaei/3x-ui/commit/6f6c7fc1) (#4782)
- [fix(panel-proxy): route custom geo and http(s) Telegram through panelProxy](https://github.com/MHSanaei/3x-ui/commit/db5ce062)
- [fix(migrate-db): preserve false-valued columns in SQLite to Postgres copy](https://github.com/MHSanaei/3x-ui/commit/71cf22fa)
- [fix(clients): use client_inbounds link to resolve inbound, not stale id](https://github.com/MHSanaei/3x-ui/commit/df7ccd3a)
- [fix(settings): allow pagination size of 0 to disable pagination](https://github.com/MHSanaei/3x-ui/commit/2f12b346)
- [fix(sub): escape Clash subscription profile filename header](https://github.com/MHSanaei/3x-ui/commit/10c185a5) (#4799) @xiaoxiyao

<h3> Reports </h3>


![total](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.2.7/total?label=total&color=success)
![amd64](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.2.7/x-ui-linux-amd64.tar.gz?label=linux-amd64)
![arm64](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.2.7/x-ui-linux-arm64.tar.gz?label=linux-arm64)
![386](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.2.7/x-ui-linux-386.tar.gz?label=linux-386)
![armv7](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.2.7/x-ui-linux-armv7.tar.gz?label=linux-armv7)
![armv6](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.2.7/x-ui-linux-armv6.tar.gz?label=linux-armv6)
![armv5](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.2.7/x-ui-linux-armv5.tar.gz?label=linux-armv5)
![s390x](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.2.7/x-ui-linux-s390x.tar.gz?label=linux-s390x)
![windows](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.2.7/x-ui-windows-amd64.zip?label=windows-amd64)

**Full Changelog**: https://github.com/MHSanaei/3x-ui/compare/v3.2.6...v3.2.7


### v3.2.8 (2026-06-05)
## 🚀 Multi-Node Resilience, ECH & Scale

- 🌐 **Multi-node resilience** — client/inbound edits survive an offline node, remote updates are scoped to a single inbound, and stale node snapshots no longer re-enable disabled clients or miscount traffic.
- 🔐 **End-to-end ECH** — now carried in TLS share links, JSON subscriptions, outbound import, and per-entry external proxy.
- 🧩 **Modern Xray JSON subscriptions** — new format with a unified finalmask editor.
- 🧭 **Clash routing** — routing rules and an enable-routing option for Clash subscriptions.
- 💾 **DB migration** — SQLite ⇄ `.dump` conversion and Download Migration from the Overview page.

### ⚡ Performance — scales to ~200k clients

Benchmarked on **PostgreSQL 16** (gains are largest on Postgres, where every round-trip pays network latency):

| Operation | Scale | Before | After | Improvement |
|---|---|---|---|---|
| Toggle one client (`SyncInbound`) | 50k-client inbound | 8m 54s | 0.9s | ~600× (~99.8%) |
| Seed clients | 50k clients | 2m 48s | 1.6s | ~100× (~99%) |
| Bulk create | large inbound | 8m 35s | ~1–5s | ~99% |
| Bulk detach | large inbound | 52s | ~4s | ~92% |
| Bulk delete | large inbound | 16s | ~1–4s | ~85% |
| Bulk adjust | large inbound | 20s | ~7–10s | ~55% |
| Delete-all clients | 100k-client inbound | ❌ crashed (param limit) | ~7s | now works |
| Bulk group add/remove | 100k clients | — | ~6s | scaled |
| Full client list | 100k clients | — | ~1s | scaled |
| `GetClientTrafficByEmail` | flat in N | 439ms | ~1.5ms | ~290× (~99.7%) |


<h3>🆕 New</h3>

- [feat(sub): modern xray JSON format with unified finalmask editor](https://github.com/MHSanaei/3x-ui/commit/97f88fb1) (#4912) @biohazardous-man
- [feat(Clash): add routing rules and enable-routing option for Clash subscriptions](https://github.com/MHSanaei/3x-ui/commit/f947fbd6) (#4904) @Misfit-s
- [feat(migrate-db): SQLite ⇄ .dump conversion and Download Migration in Overview](https://github.com/MHSanaei/3x-ui/commit/a07c7b7f)

<h3>⚡ Update & improvement</h3>

- [perf(clients): make SyncInbound bulk to fix large-inbound timeouts](https://github.com/MHSanaei/3x-ui/commit/756746db) (#4885)
- [perf(clients): scale add/delete and bulk client operations](https://github.com/MHSanaei/3x-ui/commit/f185d331)
- [perf(clients): chunk IN queries and de-quadratic bulk delete/group/list](https://github.com/MHSanaei/3x-ui/commit/d1e733b9)
- [perf(clients): scale-audit remaining client/inbound endpoints to 200k](https://github.com/MHSanaei/3x-ui/commit/d3db828b)
- [chore(deps): bump i18next from 26.3.0 to 26.3.1 in /frontend](https://github.com/MHSanaei/3x-ui/commit/ba63fa85) (#4901) 
- [i18n: add 1-year expiration to language cookie](https://github.com/MHSanaei/3x-ui/commit/a4b3e999) (#4890) @lim-kim930
- [docs(contributing): refresh frontend guide and add Postgres launch profile](https://github.com/MHSanaei/3x-ui/commit/f470bc7c)

<h3>🐞 Bug fixed</h3>

- [fix(node): keep client/inbound edits working when a node is offline](https://github.com/MHSanaei/3x-ui/commit/b40f869f) (#4923, #4931)
- [fix(multi-node): scope remote client update/delete to one inbound](https://github.com/MHSanaei/3x-ui/commit/db86007a) (#4892)
- [fix(node-traffic): prevent stale node snapshot from re-enabling disabled client](https://github.com/MHSanaei/3x-ui/commit/12d84c2a) (#4917) @younesvatan78
- [fix: restart remote xray after disabling a client to kill active sessions](https://github.com/MHSanaei/3x-ui/commit/d6d2085d) (#4918) @younesvatan78
- [fix(traffic): count local traffic for clients whose shared row is node-owned](https://github.com/MHSanaei/3x-ui/commit/e0845626) (#4921)
- [fix(sub): include ECH config in TLS share links and JSON subscription](https://github.com/MHSanaei/3x-ui/commit/f8e902a7)
- [fix(outbound): import ech and pcs from TLS share links](https://github.com/MHSanaei/3x-ui/commit/e7ffae53)
- [fix(external-proxy): relabel "Host" as "Address", add per-entry ECH](https://github.com/MHSanaei/3x-ui/commit/a8d5d0df) (#4935)
- [fix(ssl): clean ECC state, guard cert reuse, register renew hook](https://github.com/MHSanaei/3x-ui/commit/44291de9) (#4875)
- [fix(fail2ban): exempt SSH and panel ports from IP-limit ban](https://github.com/MHSanaei/3x-ui/commit/b1d079fc) (#4896)
- [fix(migrate-db): drop legacy client_traffics FK before Postgres copy](https://github.com/MHSanaei/3x-ui/commit/14e2d495) (#4882)
- [fix(tgbot): ignore commands for other bots](https://github.com/MHSanaei/3x-ui/commit/73ce1150) (#4894) @kanghouchao

<h3> Reports </h3>


![total](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.2.8/total?label=total&color=success)
![amd64](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.2.8/x-ui-linux-amd64.tar.gz?label=linux-amd64)
![arm64](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.2.8/x-ui-linux-arm64.tar.gz?label=linux-arm64)
![386](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.2.8/x-ui-linux-386.tar.gz?label=linux-386)
![armv7](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.2.8/x-ui-linux-armv7.tar.gz?label=linux-armv7)
![armv6](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.2.8/x-ui-linux-armv6.tar.gz?label=linux-armv6)
![armv5](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.2.8/x-ui-linux-armv5.tar.gz?label=linux-armv5)
![s390x](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.2.8/x-ui-linux-s390x.tar.gz?label=linux-s390x)
![windows](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.2.8/x-ui-windows-amd64.zip?label=windows-amd64)

**Full Changelog**: https://github.com/MHSanaei/3x-ui/compare/v3.2.7...v3.2.8


### v3.3.0 (2026-06-08)
## 🚀 MTProto, WARP Rotation, Subscription Outbounds & a Typed API

- 🛡️ **MTProto (FakeTLS)** — new protocol served through a managed `mtg` sidecar, no external setup required.
- 🌐 **WARP IP rotation** — rotate WARP egress IPs manually or automatically on a schedule, with API requests routed through the panel proxy.
- 🔄 **Subscription-based outbounds** — import outbounds straight from a subscription URL, with automatic refresh.
- 🎨 **Customizable subscription pages** — bring-your-own templates for the subscription landing page.
- 📑 **Typed API & OpenAPI** — components, schemas, and response examples generated directly from the Go structs; `/panel/setting` and `/panel/xray` consolidated under `/panel/api`.
- 🕸️ **Multi-hop nodes** — correct traffic attribution across chained sub-nodes, synchronized `access.log` client IPs across nodes, and a distinct purple indicator when the panel is online but the Xray core has failed.
- 📊 **Per-group traffic** — used traffic now shown for each group in the groups table.

> ⚠️ **Breaking:** `/panel/setting` and `/panel/xray` moved under `/panel/api`. Update any integrations that call those paths.

<h3>🆕 New</h3>

- [feat(mtproto): add MTProto (FakeTLS) protocol via managed mtg sidecar](https://github.com/MHSanaei/3x-ui/commit/1ca5924a) (#5076)
- [feat: add manual and automatic WARP IP rotation](https://github.com/MHSanaei/3x-ui/commit/d9ccf157) (#5099) @rqzbeh
- [feat: synchronize access.log client IPs across nodes](https://github.com/MHSanaei/3x-ui/commit/9f31d7d0) (#5098) @rqzbeh
- [feat: add support for subscription-based outbounds with auto-update](https://github.com/MHSanaei/3x-ui/commit/0daedd3d) (#5037) @rqzbeh
- [feat: customizable subscription page templates](https://github.com/MHSanaei/3x-ui/commit/abf6b879) (#5079) @rqzbeh
- [feat(nodes): multi-hop node attribution for chained sub-nodes](https://github.com/MHSanaei/3x-ui/commit/e6c1ce9a) (#4983, #5005)
- [feat(nodes): distinct purple indicator when panel is online but Xray core failed](https://github.com/MHSanaei/3x-ui/commit/1c74b995) (#5040) @rqzbeh
- [feat(groups): show used traffic per group in groups table](https://github.com/MHSanaei/3x-ui/commit/1fa51cf0)
- [feat(api-docs): generate OpenAPI components/schemas from Go structs](https://github.com/MHSanaei/3x-ui/commit/a014c017)
- [feat(api-docs): generate response examples from Go structs; fix SS2022 PSK regen](https://github.com/MHSanaei/3x-ui/commit/83799d71) (#4996)
- [feat(x-ui.sh): add migrateDB command for SQLite .db ⇄ .dump](https://github.com/MHSanaei/3x-ui/commit/0706b0b3) (#4910)

<h3>⚡ Update & improvement</h3>

- [refactor(api)!: move /panel/setting and /panel/xray under /panel/api](https://github.com/MHSanaei/3x-ui/commit/c6f15cd5)
- [fix(subClashService): improve merging of clash rules in YAML](https://github.com/MHSanaei/3x-ui/commit/98ba8803) (#5054) @shazzreab
- [fix(update.sh): allow skipping ssl setup when updating](https://github.com/MHSanaei/3x-ui/commit/4e253588) (#5071)
- [chore: bump frontend version and deps](https://github.com/MHSanaei/3x-ui/commit/9acde8da)
- [i18n(tr): improve Turkish translation consistency and terminology](https://github.com/MHSanaei/3x-ui/commit/b0fe21c8) (#5066)
- [docs(i18n): add Turkish translation for README](https://github.com/MHSanaei/3x-ui/commit/f6558571) (#5067)
- [docs(i18n): refine Turkish translation and network terminology](https://github.com/MHSanaei/3x-ui/commit/0d7b6872) (#5092)
- [i18n: translate sockopt / REALITY-target / Freedom strings for all locales](https://github.com/MHSanaei/3x-ui/commit/1b2a17f7) (#4988)

<h3>🐞 Bug fixed</h3>

- [fix: propagate inbound traffic reset to nodes](https://github.com/MHSanaei/3x-ui/commit/be8bd4e2) (#5103) @rqzbeh
- [fix: route WARP API requests through panel proxy](https://github.com/MHSanaei/3x-ui/commit/a32c6803) (#5101) @rqzbeh
- [fix(db): additional cross-DB and node-traffic edge cases (migration scan + node reset time)](https://github.com/MHSanaei/3x-ui/commit/94b8196e) (#5045) @rqzbeh
- [fix(postgres): make node traffic sync robust after public API inbound updates](https://github.com/MHSanaei/3x-ui/commit/21e01cc1) (#5038) @rqzbeh
- [fix(node-sync): merge client enable with boolean AND for PostgreSQL](https://github.com/MHSanaei/3x-ui/commit/eeb19b72)
- [fix(xray): sync routing rules when outbound tag is renamed](https://github.com/MHSanaei/3x-ui/commit/e8171ab4) (#5006) @nima1024m
- [fix(panel): normalize XHTTP/sockopt/Reality wire output and validate REALITY target](https://github.com/MHSanaei/3x-ui/commit/6ed6f57b) (#4988) @nima1024m
- [fix(sub): emit VLESS encryption in Clash configs](https://github.com/MHSanaei/3x-ui/commit/46684dd1) (#5053) @fs438187
- [fix(sub): restore standard base64 for Shadowrocket sub link](https://github.com/MHSanaei/3x-ui/commit/668c0922) (#5001)
- [fix(sub): don't project public inbounds through a fallback master](https://github.com/MHSanaei/3x-ui/commit/2b4e199a)
- [fix(inbounds): drop unknown nodeId when importing an inbound](https://github.com/MHSanaei/3x-ui/commit/b24b8524)
- [fix(finalmask): validate fragment mask length so empty/zero-min can't crash xray](https://github.com/MHSanaei/3x-ui/commit/483952cf)
- [fix(finalmask): treat sudoku customTables as array of tables](https://github.com/MHSanaei/3x-ui/commit/5b9db13e)
- [fix(iplimit): skip stale access-log emails after client rename/delete](https://github.com/MHSanaei/3x-ui/commit/e409bc30)
- [fix(tgbot): apply bot settings on panel restart without full service restart](https://github.com/MHSanaei/3x-ui/commit/3d6ff2b6)
- [fix(script): revoke also removes cert files and acme.sh tracking](https://github.com/MHSanaei/3x-ui/commit/8ce61f3c) (#5009)
- [fix: default hysteria tls to no utls fingerprint](https://github.com/MHSanaei/3x-ui/commit/af3c8084)
- [fix: correct arm architecture xray binary file name](https://github.com/MHSanaei/3x-ui/commit/d739bcf7) (#5060)
- [fix(api-docs): target the panel base path in OpenAPI servers](https://github.com/MHSanaei/3x-ui/commit/e56f6c63)
- [fix(ui): correct inline style syntax in client counts column on inbounds page](https://github.com/MHSanaei/3x-ui/commit/7d908834) (#5097)
- [fix(ui): remove pointer cursor from non-interactive elements in cards](https://github.com/MHSanaei/3x-ui/commit/5a7de025) (#5102)
- [fix(inbound-form): wrap long labels and shorten RU pinned-cert label](https://github.com/MHSanaei/3x-ui/commit/75bc6e80)
- [fix(mtproto): reap orphaned mtg, fix SysLog viewer, mtg log visibility, export remark](https://github.com/MHSanaei/3x-ui/commit/f8e89cc848b908d8507f30e0e35a0a74d6fe983c)

<h3> Reports </h3>

![total](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.3.0/total?label=total&color=success)
![amd64](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.3.0/x-ui-linux-amd64.tar.gz?label=linux-amd64)
![arm64](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.3.0/x-ui-linux-arm64.tar.gz?label=linux-arm64)
![386](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.3.0/x-ui-linux-386.tar.gz?label=linux-386)
![armv7](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.3.0/x-ui-linux-armv7.tar.gz?label=linux-armv7)
![armv6](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.3.0/x-ui-linux-armv6.tar.gz?label=linux-armv6)
![armv5](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.3.0/x-ui-linux-armv5.tar.gz?label=linux-armv5)
![s390x](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.3.0/x-ui-linux-s390x.tar.gz?label=linux-s390x)
![windows](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.3.0/x-ui-windows-amd64.zip?label=windows-amd64)

**Full Changelog**: https://github.com/MHSanaei/3x-ui/compare/v3.2.8...v3.3.0


### v3.3.1 (2026-06-12)
## 🚀 Live Config Apply, Native Geodata, Smarter Nodes & a Big Internal Refactor

- ⚡ **Live config apply** — inbound / outbound / routing changes now apply over the Xray gRPC API without a full core restart, so existing connections survive edits.
- 🌍 **Native geodata auto-update** — the custom geo manager is gone; geo files now auto-update through Xray-core's built-in mechanism.
- 📡 **Access-log-free online tracking** — onlines and per-client IP limits now read from Xray's online-stats API instead of parsing `access.log`.
- 🕸️ **Smarter multi-node sync** — filter inbounds and clients by node, push global client usage to nodes for display + local enforcement, and a per-inbound share-address strategy that carries through to subscriptions.
- 🌉 **Outbound-based egress bridge** — the panel proxy URL is replaced by a proper outbound egress bridge; a balancer can now serve as the panel traffic outbound.
- 🔐 **MTProto upgrades** — domain-fronting and essential `mtg` options, plus Telegram egress routed through your Xray routing rules.
- 🧩 **WireGuard refresh** — latest Xray-core WireGuard features and per-peer comments to identify devices.
- 🛡️ **Security fix** — `log.access` / `log.error` paths are confined to the panel log folder (GHSA-jm48 arbitrary file write).
- 🛠️ **Internal refactor** — focused service files, leaf subpackages and a cleaner `internal/` layout (no API surface change).

> ℹ️ **Heads-up:** geo data now auto-updates via Xray-core and the old panel proxy URL is superseded by the outbound egress bridge. If you relied on either, review your settings after upgrading.

<h3>🆕 New</h3>

- [feat: apply inbound/outbound/routing changes live via Xray gRPC API](https://github.com/MHSanaei/3x-ui/commit/6b16d8c3)
- [feat: replace panel proxy URL with outbound-based egress bridge](https://github.com/MHSanaei/3x-ui/commit/ca4f32e3)
- [feat(online): use xray online-stats API for onlines and access-log-free IP limit](https://github.com/MHSanaei/3x-ui/commit/7bcc5830)
- [feat(node-sync): push global client usage to nodes for display and local enforcement](https://github.com/MHSanaei/3x-ui/commit/58905d81)
- [feat: filter inbounds and clients by node](https://github.com/MHSanaei/3x-ui/commit/253063b7) (#4997)
- [feat: add inbound share address strategy](https://github.com/MHSanaei/3x-ui/commit/2a7342ba) (#5162) @yuanzhidao
- [feat: allow selecting inbounds synchronized from nodes](https://github.com/MHSanaei/3x-ui/commit/554d85c2) (#5178) @animesha3
- [feat(mtproto): add domain-fronting and essential mtg options](https://github.com/MHSanaei/3x-ui/commit/6c159469)
- [feat(mtproto): route Telegram egress through Xray routing rules](https://github.com/MHSanaei/3x-ui/commit/5eec1784)
- [feat: support latest Wireguard features from Xray-core](https://github.com/MHSanaei/3x-ui/commit/4002be4a) (#5131) @rqzbeh
- [feat(wireguard): per-peer comments for identifying devices](https://github.com/MHSanaei/3x-ui/commit/d04cb109) (#5168)
- [feat: implement inbound XMUX form fields](https://github.com/MHSanaei/3x-ui/commit/0766e166) (#5211) @rqzbeh
- [feat(inbound): support abstract unix sockets (@ prefix) in Address field](https://github.com/MHSanaei/3x-ui/commit/7c698c4b)
- [feat(sub): per-inbound sort order for subscription links](https://github.com/MHSanaei/3x-ui/commit/f1a4286e)
- [feat(sub): add Copy All Configs button to subscription page](https://github.com/MHSanaei/3x-ui/commit/ffde2f7e) (#5163) @NikanZeyaei
- [feat(outbound): batched connection tester with direct timed HTTP probes](https://github.com/MHSanaei/3x-ui/commit/5716ae59)
- [feat(settings): allow a balancer as the panel traffic outbound](https://github.com/MHSanaei/3x-ui/commit/8578b229)
- [feat(settings): schedule picker, toggle placement, sub-theme docs link](https://github.com/MHSanaei/3x-ui/commit/c7a01887)
- [feat(groups): show upload/download breakdown in group traffic](https://github.com/MHSanaei/3x-ui/commit/1c5cb844)
- [feat(api): include consumed traffic in the client-get response](https://github.com/MHSanaei/3x-ui/commit/d1a13844) (#4973)
- [feat(env): allow setting the initial URI path for the web panel](https://github.com/MHSanaei/3x-ui/commit/89b1137b) (#5149) @Ponywka
- [feat(routing): show tag (remark) in routing rules list](https://github.com/MHSanaei/3x-ui/commit/8f408d2d) (#5151) @aleskxyz
- [feat(clients): restore traffic usage progress bars on Clients page](https://github.com/MHSanaei/3x-ui/commit/941eba54) (#5150) @nima1024m
- [feat(clients): restore reset traffic button in edit client form](https://github.com/MHSanaei/3x-ui/commit/4eab37b6)
- [feat(ui): add select all / clear all shortcuts for inbound multi-select](https://github.com/MHSanaei/3x-ui/commit/07e5e849) (#5175) @NikanZeyaei
- [feat(ui): use CodeMirror editor for Import Inbound and Inbound JSON](https://github.com/MHSanaei/3x-ui/commit/0cefadd1)
- [feat(ui): improve client form modal UX](https://github.com/MHSanaei/3x-ui/commit/7ae3ea66)
- [feat(ui): allow custom fragment packets ranges, not just presets](https://github.com/MHSanaei/3x-ui/commit/bade1fce) (#5075)

<h3>⚡ Update & improvement</h3>

- [refactor: focused service files, leaf subpackages, and an internal/ layout](https://github.com/MHSanaei/3x-ui/commit/41645255) (#5167)
- [refactor: replace custom geo manager with Xray-core native geodata auto-update](https://github.com/MHSanaei/3x-ui/commit/3092326d)
- [refactor(settings): reorganize subscription settings into clearer tabs](https://github.com/MHSanaei/3x-ui/commit/08bc481a)
- [refactor(groups): restyle traffic summary into upload/download + usage cards](https://github.com/MHSanaei/3x-ui/commit/85983eec)
- [style(inbounds): show total up/down with directional arrows](https://github.com/MHSanaei/3x-ui/commit/0f7da02a)
- [style(ui): enlarge row action icons and rebalance clients table widths](https://github.com/MHSanaei/3x-ui/commit/6e205882)
- [Update ExecReload command in x-ui.service.debian](https://github.com/MHSanaei/3x-ui/commit/63a6d404) (#5219) @ssrlive
- [i18n: point API token hint at the Authentication page in all locales](https://github.com/MHSanaei/3x-ui/commit/7e87b7dc)
- [docs: add Turkish language link to other README files](https://github.com/MHSanaei/3x-ui/commit/d047075f) (#5138) @tarihcituranx
- [chore: pin generated files to LF to avoid phantom CRLF diffs on Windows](https://github.com/MHSanaei/3x-ui/commit/0711d307)
- [chore: bump Go indirect deps; update frontend lock](https://github.com/MHSanaei/3x-ui/commit/cd46730b)
- [chore(deps): bump golang.org/x/net from 0.55.0 to 0.56.0](https://github.com/MHSanaei/3x-ui/commit/eee652c4) (#5199)
- [ci(bot): update issue-bot repo map and tighten reply style](https://github.com/MHSanaei/3x-ui/commit/9730561f)

<h3>🐞 Bug fixed</h3>

- [fix(xray): confine log.access/error to the panel log folder](https://github.com/MHSanaei/3x-ui/commit/80e16878)
- [fix(client): preserve UUID/password/auth on partial client update](https://github.com/MHSanaei/3x-ui/commit/2969f6e9) (#5111)
- [fix(client): apply per-field client edits to every inbound of the email](https://github.com/MHSanaei/3x-ui/commit/1a525b4c) (#5039)
- [fix(client): match clients by email for delete/update, not credentials](https://github.com/MHSanaei/3x-ui/commit/26c549a9)
- [fix(clients): invalidate Xray config cache after client mutations](https://github.com/MHSanaei/3x-ui/commit/0c73862b)
- [fix(node-sync): keep node baseline while a sibling inbound still reports the email](https://github.com/MHSanaei/3x-ui/commit/21143a6d) (#5202)
- [fix(node-sync): keep shared client traffic row when email still lives on other inbounds](https://github.com/MHSanaei/3x-ui/commit/8258a26f)
- [fix(nodes): "Invalid input" when saving a node with inbound sync mode "all"](https://github.com/MHSanaei/3x-ui/commit/5c29851b)
- [fix(sub): honor per-inbound share address strategy in subscription output](https://github.com/MHSanaei/3x-ui/commit/cc65f371) (#5208)
- [fix(sub): deduplicate settings.clients entries per inbound in subscription output](https://github.com/MHSanaei/3x-ui/commit/1b0dbf8e) (#5134)
- [fix(sub): tag node-hosted entries with the node name in remarks](https://github.com/MHSanaei/3x-ui/commit/b062cb5a) (#5035)
- [fix: derive JSON/Clash subscription URLs from configured subURI](https://github.com/MHSanaei/3x-ui/commit/ec45d349) (#5203) @w3struk
- [fix: enable XTLS vision flow for VLESS+XHTTP+vlessenc in UI and share links](https://github.com/MHSanaei/3x-ui/commit/c7a76e96) (#5157, #5185) @rqzbeh
- [fix: expose streamSettings for Tunnel inbounds to support TProxy](https://github.com/MHSanaei/3x-ui/commit/1ad483ed) (#5171) @rqzbeh
- [fix(inbound): preserve custom share strategy on edit](https://github.com/MHSanaei/3x-ui/commit/90e62177) (#5225) @yuanzhidao
- [fix(inbound): offer node share-address strategy only when a node exists](https://github.com/MHSanaei/3x-ui/commit/c47a905a)
- [fix(inbound): avoid UNIQUE email constraint when importing inbounds that share clients](https://github.com/MHSanaei/3x-ui/commit/3af1afc5)
- [fix(inbound): remove stale mkcp-legacy finalmask when switching away from mKCP](https://github.com/MHSanaei/3x-ui/commit/5af02265)
- [fix(inbound): explain how to unlock fallbacks on the inbound form](https://github.com/MHSanaei/3x-ui/commit/a5e56408) (#5014)
- [fix(inbounds): show remark first, else inbound tag, in client labels](https://github.com/MHSanaei/3x-ui/commit/41cb0b8a)
- [fix(xhttp): stop injecting scMaxEachPostBytes/scMinPostsIntervalMs defaults](https://github.com/MHSanaei/3x-ui/commit/60da6bed) (#5141)
- [fix(hysteria): clamp udpIdleTimeout to xray-core's accepted 2-600s range](https://github.com/MHSanaei/3x-ui/commit/10a0c913) (#5117)
- [fix(warp): prefer IPv4 with v6 fallback and userspace TUN in generated WireGuard outbounds](https://github.com/MHSanaei/3x-ui/commit/09a887f9) (#5205)
- [fix(outbound): widen probe timeout and surface failure reason in outbound test](https://github.com/MHSanaei/3x-ui/commit/82577814) (#5152)
- [fix(outbound): include tested outbound in HTTP probe config](https://github.com/MHSanaei/3x-ui/commit/0bed5522) (#5120)
- [fix(settings): normalize tgCpu on load so a bad value can't block saving](https://github.com/MHSanaei/3x-ui/commit/0e0e4119) (#5091)
- [fix: DNS server edit modal showing defaults instead of saved values](https://github.com/MHSanaei/3x-ui/commit/1508666e) (#5155)
- [fix: apply only the x-ui sysctl config when toggling BBR](https://github.com/MHSanaei/3x-ui/commit/2db48174) (#5160)
- [fix(update): restart panel after regenerating webBasePath to fix login desync](https://github.com/MHSanaei/3x-ui/commit/f88f53cd)
- [fix(script): SSL management fixes](https://github.com/MHSanaei/3x-ui/commit/dbee150b) (#4994, #5010, #5070)
- [fix: properly configure fail2ban backend and dependencies on Ubuntu 22.04+](https://github.com/MHSanaei/3x-ui/commit/57e96617) (#5159, #5184) @rqzbeh
- [fix: accurately retrieve and generate API tokens via CLI with hashed storage](https://github.com/MHSanaei/3x-ui/commit/65fa40b8) (#5145, #5183) @rqzbeh
- [fix: inbound edit validation failure and legacy copy to clipboard](https://github.com/MHSanaei/3x-ui/commit/fe62c39a) (#5132) @rqzbeh
- [fix(ui): keep dropdown action menus inside the viewport](https://github.com/MHSanaei/3x-ui/commit/a27d57b2) (#5133)
- [fix(ui): keep client IP log modal above edit modal](https://github.com/MHSanaei/3x-ui/commit/f9b275dd) (#5137) @JScarlet
- [fix(ui): blink the online dot in mobile client cards like desktop](https://github.com/MHSanaei/3x-ui/commit/dc52e725)
- [fix(ui): classify ended clients as depleted, not disabled, on inbounds page](https://github.com/MHSanaei/3x-ui/commit/aeb2217a)
- [fix(ui): correct inline style syntax between clients count and active clients count on inbounds page](https://github.com/MHSanaei/3x-ui/commit/dbb269cf) (#5114) @jimself218-ops

<h3> Reports </h3>

![total](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.3.1/total?label=total&color=success)
![amd64](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.3.1/x-ui-linux-amd64.tar.gz?label=linux-amd64)
![arm64](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.3.1/x-ui-linux-arm64.tar.gz?label=linux-arm64)
![386](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.3.1/x-ui-linux-386.tar.gz?label=linux-386)
![armv7](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.3.1/x-ui-linux-armv7.tar.gz?label=linux-armv7)
![armv6](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.3.1/x-ui-linux-armv6.tar.gz?label=linux-armv6)
![armv5](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.3.1/x-ui-linux-armv5.tar.gz?label=linux-armv5)
![s390x](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.3.1/x-ui-linux-s390x.tar.gz?label=linux-s390x)
![windows](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.3.1/x-ui-windows-amd64.zip?label=windows-amd64)

## New Contributors
* @JScarlet made their first contribution in https://github.com/MHSanaei/3x-ui/pull/5137
* @aleskxyz made their first contribution in https://github.com/MHSanaei/3x-ui/pull/5151
* @NikanZeyaei made their first contribution in https://github.com/MHSanaei/3x-ui/pull/5163
* @w3struk made their first contribution in https://github.com/MHSanaei/3x-ui/pull/5203
* @yuanzhidao made their first contribution in https://github.com/MHSanaei/3x-ui/pull/5162
* @animesha3 made their first contribution in https://github.com/MHSanaei/3x-ui/pull/5178
* @ssrlive made their first contribution in https://github.com/MHSanaei/3x-ui/pull/5219

**Full Changelog**: https://github.com/MHSanaei/3x-ui/compare/v3.3.0...v3.3.1

### v3.4.0 (2026-06-23)
## 🚀 Multi-Node Hardening, Notification Event Bus, Managed Hosts & Scale to 100k Clients

- 🛰️ **Per-node outbound routing & node hardening** — route each node through its own outbound, plus mTLS, hashed + zstd reconcile transport, and per-node network metrics.
- 🔔 **Notification event bus** — a pub/sub architecture with Telegram and SMTP subscribers, a card-based notification settings layout, and memory-threshold alerts.
- 🌐 **Managed Hosts** — per-host overrides for subscription links so each host can advertise its own address.
- 🧾 **Subscription engine upgrades** — dynamic remark variables (Jalali date, transport, status tokens), full XHTTP mapping for Clash/Mihomo, per-client external links + remote subscriptions, and an option to hide server settings (happ).
- 📈 **Scale & stability to 50k–100k clients** — faster traffic/auto-renew/node bulk ops, DB indexes on hot columns, atomic config writes, panic-recovering cron/jobs, and bounded gRPC deadlines & response sizes.
- 🛡️ **fail2ban-native IP limiting** — IP limit is now gated on fail2ban (auto-installed on install/update) and reads onlines without parsing `access.log`.
- 🔐 **Native TLS/REALITY pinning** — remote cert pinning via a native uTLS handshake (no xray subprocess), ported xray TLS/REALITY fields, and cert-hash helpers.
- 🎯 **Real client IP behind CDN/relay** — capture the visitor IP behind a CDN/relay and attribute IP-limit per node.
- 🧬 **Xray-core v26.6.22** — core upgrade with XHTTP `sessionID` table/length controls, WireGuard field cleanup, and `trustedXForwardedFor` honored on gRPC inbounds.
- 🛠️ **Deployment pipeline & test-quality audit** — release-driven golden-image & unattended-install pipeline, plus a test-quality pass adding mutation/fuzz/CI tooling.

> ℹ️ **Heads-up:** IP limiting now relies on **fail2ban**, which is auto-installed on install/update, and no longer parses `access.log`. Legacy `panelProxy` / `tgBotProxy` settings are cleared automatically on upgrade. If you previously tuned IP limiting or those proxy settings, review them after upgrading.

<h3>🆕 New</h3>

- [feat(node): per node outbound routing](https://github.com/MHSanaei/3x-ui/commit/05ad7f41) (#5275) @NikanZeyaei
- [feat(node): node hardening — mTLS, hashed+zstd reconcile transport, per-node net metrics](https://github.com/MHSanaei/3x-ui/commit/37c5e0bf) (#5382)
- [feat(notifications): event bus architecture with Telegram and SMTP subscribers](https://github.com/MHSanaei/3x-ui/commit/eec030f8) (#5326) @Ssentiago
- [feat: replace notification checkboxes with card-based layout](https://github.com/MHSanaei/3x-ui/commit/55d08d2a) (#5421) @Ssentiago
- [feat(memory): add memory threshold alerts](https://github.com/MHSanaei/3x-ui/commit/891d3a87) (#5366) @Ssentiago
- [feat(hosts): managed Hosts for per-host subscription link overrides](https://github.com/MHSanaei/3x-ui/commit/709b332d) (#5409)
- [feat(nodes): per-node client IP attribution for IP-limit](https://github.com/MHSanaei/3x-ui/commit/9385b6c6)
- [feat(inbounds): add Real client IP presets to capture visitor IP behind CDN/relay](https://github.com/MHSanaei/3x-ui/commit/d882d6aa)
- [feat(sub): per-client external links and remote subscriptions](https://github.com/MHSanaei/3x-ui/commit/dcb923b4)
- [feat(sub): add dynamic remark variables with Jalali date, transport, and status tokens](https://github.com/MHSanaei/3x-ui/commit/605e90db) (#5430) @wahh3b-lgtm
- [feat(sub): full XHTTP field mapping for Clash/Mihomo subscriptions](https://github.com/MHSanaei/3x-ui/commit/1a4aef33) (#5417) @w3struk
- [feat(sub): add option to hide server settings in subscription (happ)](https://github.com/MHSanaei/3x-ui/commit/ce1d348e) (#5433) @IgorKha
- [feat(tls,reality): port xray TLS/REALITY fields, cert-hash helpers, fallback UX](https://github.com/MHSanaei/3x-ui/commit/7c888946)
- [feat(finalmask): support Salamander packetSize (Gecko) and Realm tlsConfig for Hysteria2](https://github.com/MHSanaei/3x-ui/commit/dab0add1) (#5278) @rqzbeh
- [feat(xray): add loopback sniffing and per-segment fragment masks](https://github.com/MHSanaei/3x-ui/commit/852b53db)
- [feat(xray): preview export in a modal and switch rule enable toggle](https://github.com/MHSanaei/3x-ui/commit/97c02ef6)
- [feat(xhttp): support sessionID* rename + sessionIDTable/Length (xray v26.6.22)](https://github.com/MHSanaei/3x-ui/commit/fea3c94b) (#5506) @rqzbeh
- [Add Enable/Disable Toggle for Xray Routing Rules](https://github.com/MHSanaei/3x-ui/commit/53f6ed39) (#5296) @abdalrahmanx9
- [feat(iplimit): gate IP limit on fail2ban and reset stale limits](https://github.com/MHSanaei/3x-ui/commit/ce8b1bed)
- [feat(iplimit): auto-install fail2ban on install and update](https://github.com/MHSanaei/3x-ui/commit/0d764f1b)
- [feat(ui): show per-inbound live speed](https://github.com/MHSanaei/3x-ui/commit/f4bbaf40) (#5261) @NikanZeyaei
- [feat(metrics): extend history bucket options to include 12h, 24h, and 48h intervals](https://github.com/MHSanaei/3x-ui/commit/648fc69c) (#5467) @shazzreab
- [feat(sidebar): move Routing/Outbounds to top-level items with clean URLs](https://github.com/MHSanaei/3x-ui/commit/718b7e16)
- [feat(clients): orphan cleanup + export/import via CodeMirror modals](https://github.com/MHSanaei/3x-ui/commit/0b0b6250)
- [feat(backup): name DB backup files after the server address](https://github.com/MHSanaei/3x-ui/commit/a7e959ff)
- [feat(backup): prefer browser request host for backup filename](https://github.com/MHSanaei/3x-ui/commit/dabd3f5d)
- [feat(web): cap request body size on state-changing routes](https://github.com/MHSanaei/3x-ui/commit/71616b7c) (#5271) @n0ctal
- [feat(docker): support XUI_PORT runtime override](https://github.com/MHSanaei/3x-ui/commit/7f34c306) (#5240) @surfdps
- [feat: release-driven golden-image & unattended-install deployment pipeline](https://github.com/MHSanaei/3x-ui/commit/7c2598fa) (#5323)

<h3>⚡ Update & improvement</h3>

- [Update Xray to v26.6.22](https://github.com/MHSanaei/3x-ui/commit/a2961fd0)
- [perf(scale): speed up traffic, auto-renew, and node bulk ops at 50k-100k clients](https://github.com/MHSanaei/3x-ui/commit/6a032bcb)
- [perf: prevent cron job overlap, auto-set GOMEMLIMIT, fix tgbot userStates race](https://github.com/MHSanaei/3x-ui/commit/7d23a2c1)
- [perf(settings): save all settings in one transaction](https://github.com/MHSanaei/3x-ui/commit/20094c8d)
- [perf(db): index group_name and client_traffics hot columns](https://github.com/MHSanaei/3x-ui/commit/21888306) (#5268) @n0ctal
- [perf(db): add an index on settings.key](https://github.com/MHSanaei/3x-ui/commit/3cf3fddf) (#5359) @n0ctal
- [perf(xray): compile log/traffic regexps once at package scope](https://github.com/MHSanaei/3x-ui/commit/26cc4838) (#5362) @n0ctal
- [chore(db): use DELETE journal mode so sqlite stays a single file](https://github.com/MHSanaei/3x-ui/commit/e0794901)
- [refactor(web): centralize background job cadences](https://github.com/MHSanaei/3x-ui/commit/d14f341b) (#5269) @n0ctal
- [refactor(job): drop access log from IP limiting, wipe it daily instead](https://github.com/MHSanaei/3x-ui/commit/42cd351e)
- [refactor(frontend): stack client credential fields and use label hints on inbound form](https://github.com/MHSanaei/3x-ui/commit/bbab83db)
- [refactor(frontend): move form-item hints from extra to tooltip](https://github.com/MHSanaei/3x-ui/commit/4915d6b1)
- [refactor(wireguard): drop removed workers field (xray v26.6.22)](https://github.com/MHSanaei/3x-ui/commit/b07fad0e) (#5509) @rqzbeh
- [Use efficient APIs and simplify loops](https://github.com/MHSanaei/3x-ui/commit/1c0b76c2)
- [Frontend operation button size optimization](https://github.com/MHSanaei/3x-ui/commit/b5872af2) (#5343) @tonymoses10
- [Test-quality audit: fix 2 prod bugs, strengthen weak tests, add mutation/fuzz/CI tooling](https://github.com/MHSanaei/3x-ui/commit/76059023) (#5345)
- [chore(deps): bump frontend deps and override js-yaml to patch DoS advisory](https://github.com/MHSanaei/3x-ui/commit/5b8504c7)
- [Bump frontend package & deps to new patch versions](https://github.com/MHSanaei/3x-ui/commit/fd092444)
- [chore(deps): bump telego to v1.10.0](https://github.com/MHSanaei/3x-ui/commit/dc781b28)
- [chore: bump dompurify to 3.4.11 and expand VS Code tasks](https://github.com/MHSanaei/3x-ui/commit/0537cbfb)
- [chore(deps): bump aws-actions/configure-aws-credentials from 4 to 6](https://github.com/MHSanaei/3x-ui/commit/a1d71d42) (#5426)
- [chore(deps): bump react-router-dom from 7.17.0 to 7.18.0 in /frontend](https://github.com/MHSanaei/3x-ui/commit/a1aa8fcc) (#5428)
- [chore(deps): bump actions/upload-artifact from 4 to 7](https://github.com/MHSanaei/3x-ui/commit/4f99e48a) (#5427)
- [chore(deps): bump actions/checkout from 6 to 7](https://github.com/MHSanaei/3x-ui/commit/1eaa73e7) (#5454)
- [i18n: sync 12 locales with en-US — add missing Hosts/subscription keys](https://github.com/MHSanaei/3x-ui/commit/5038fa1c)
- [Update zh-CN.json](https://github.com/MHSanaei/3x-ui/commit/dfd77caf) (#5459) @qin9125
- [feat(ci): add PR review job and commit-capable mention bot](https://github.com/MHSanaei/3x-ui/commit/caf80009)
- [feat(ci): let mention bot push commits to fork PR branches](https://github.com/MHSanaei/3x-ui/commit/29b14dac)
- [fix(ci): check out PR branch for mention bot so commits land on the PR](https://github.com/MHSanaei/3x-ui/commit/4ab2dffa)
- [fix(ci): use pull_request_target so claude bot gets secrets on fork PRs](https://github.com/MHSanaei/3x-ui/commit/d20b549b)
- [ci(claude-bot): tune models, Copilot-style PR review, issue research mode](https://github.com/MHSanaei/3x-ui/commit/b11c51e7)
- [ci(smoke): set least-privilege GITHUB_TOKEN permissions](https://github.com/MHSanaei/3x-ui/commit/a133282f)
- [ci(smoke): retry transient GitHub download failures](https://github.com/MHSanaei/3x-ui/commit/1c750349)
- [ci: use .nvmrc for setup-node version in codeql/release workflows](https://github.com/MHSanaei/3x-ui/commit/f3eba04e)

<h3>🐞 Bug fixed</h3>

- [fix(security): confine GetCertHash to known cert files (CWE-22)](https://github.com/MHSanaei/3x-ui/commit/33b029e1)
- [fix(sub): deliver vision flow for VLESS+XHTTP+REALITY in share links and Clash](https://github.com/MHSanaei/3x-ui/commit/3c68b039) (#5232)
- [fix(sub): stop appending the node name to subscription remarks](https://github.com/MHSanaei/3x-ui/commit/b7702879) (#5231)
- [fix(sub): emit Shadowsocks http-header links as SIP002 obfs-local plugin](https://github.com/MHSanaei/3x-ui/commit/21e9b94b)
- [fix(sub): wrap JSON-subscription SS/Trojan outbound in servers[] array](https://github.com/MHSanaei/3x-ui/commit/340d0df9)
- [fix(sub): emit JSON-subscription pinnedPeerCertSha256 as comma-separated string](https://github.com/MHSanaei/3x-ui/commit/d6cddaff)
- [fix(sub): {{INBOUND}} = inbound remark, fix {{TRAFFIC_LEFT}} across inbounds](https://github.com/MHSanaei/3x-ui/commit/6d9fd4b4) (#5443)
- [fix(sub): re-add xhttp mode to extra JSON for Karing](https://github.com/MHSanaei/3x-ui/commit/0a40ec5f) (#5446)
- [fix(sub): add missing :// in Shadowrocket subscription deep link](https://github.com/MHSanaei/3x-ui/commit/c58db81d) (#3945)
- [fix(sub): SS2022 share links must not base64-encode userinfo](https://github.com/MHSanaei/3x-ui/commit/a5bc71a6) (#5432)
- [fix(sub): preserve non-default scMinPostsIntervalMs and use per-inbound xmux in JSON subscriptions](https://github.com/MHSanaei/3x-ui/commit/d01d9867) (#5393) @w3struk
- [fix(sub): set read/write/idle timeouts on the subscription server](https://github.com/MHSanaei/3x-ui/commit/118d1e43) (#5360) @n0ctal
- [fix(sub): error instead of silently truncating oversized subscription](https://github.com/MHSanaei/3x-ui/commit/67344cae) (#5495) @n0ctal
- [fix(subscription): bound outbound response body](https://github.com/MHSanaei/3x-ui/commit/ecb0b0a9) (#5493) @n0ctal
- [fix(subscriptions): avoid shared mutable state during generation](https://github.com/MHSanaei/3x-ui/commit/ac8cb505) (#5270) @n0ctal
- [fix(links): bracket ipv6 hosts in share links and qr codes](https://github.com/MHSanaei/3x-ui/commit/7c737820) (#5310) @NikanZeyaei
- [fix(nodes): honor TLS verify mode skip/pin for remote node operations](https://github.com/MHSanaei/3x-ui/commit/4c8d3cb6) (#5264)
- [fix(nodes): propagate single-client deletion to remote nodes](https://github.com/MHSanaei/3x-ui/commit/cbb21b75) (#5352)
- [fix(nodes): stop multi-attached client traffic inflating across node inbounds](https://github.com/MHSanaei/3x-ui/commit/7fe082a7)
- [fix(nodes): route 'load inbounds' through the connection outbound](https://github.com/MHSanaei/3x-ui/commit/c1fdcd98)
- [fix(ui): match node connection-outbound picker to panel-outbound selector](https://github.com/MHSanaei/3x-ui/commit/33547060)
- [fix(nodes): sync "start after first connect" expiry so un-activated nodes do not reset it](https://github.com/MHSanaei/3x-ui/commit/62840611) (#5319) @rqzbeh
- [fix(nodes): strip central n<id>- tag prefix when pushing inbounds to remote](https://github.com/MHSanaei/3x-ui/commit/da9ecf6f) (#5399) @aleskxyz
- [fix(nodes): block node delete while inbounds are still attached](https://github.com/MHSanaei/3x-ui/commit/f5e50038) (#5394) @n0ctal
- [fix(node): mark node dirty on Update so sync reconciles before snapshot sweep](https://github.com/MHSanaei/3x-ui/commit/6f05c0a4) (#5469) @NikanZeyaei
- [fix(nodes): cloned-node attribution, node-hosted client display (online/speed/counts), and sync robustness](https://github.com/MHSanaei/3x-ui/commit/adc64bb8) (#5488)
- [fix(node-sync): give client-IP sync its own deadline; fix log spacing](https://github.com/MHSanaei/3x-ui/commit/4854f9c1)
- [fix(traffic): prevent phantom quota consumption from stale node data](https://github.com/MHSanaei/3x-ui/commit/fb03b0e9) (#5412) @younesvatan78
- [fix(xray): guard process lifecycle fields against concurrent access](https://github.com/MHSanaei/3x-ui/commit/abffa8f6) (#5395) @n0ctal
- [fix(xray): verify the release archive checksum before installing](https://github.com/MHSanaei/3x-ui/commit/2bb851dd) (#5396) @n0ctal
- [fix(xray): guard log-writer race and bound handler gRPC deadlines](https://github.com/MHSanaei/3x-ui/commit/2bb29468) (#5442) @n0ctal
- [fix(xray): write generated config atomically](https://github.com/MHSanaei/3x-ui/commit/523a593c) (#5494) @n0ctal
- [fix(web): recover panicking cron jobs instead of crashing the panel](https://github.com/MHSanaei/3x-ui/commit/bedbe04b) (#5363) @n0ctal
- [fix(jobs): isolate per-node background goroutines from panics](https://github.com/MHSanaei/3x-ui/commit/f63ed9f5) (#5397) @n0ctal
- [fix(runtime): cap remote node response size to bound master memory](https://github.com/MHSanaei/3x-ui/commit/b0ef6067) (#5361) @n0ctal
- [fix(service): serialize client/inbound writes to prevent Postgres deadlock](https://github.com/MHSanaei/3x-ui/commit/c5d31de4)
- [fix(deps): bump xray-core past finalmask UDP buffer fix](https://github.com/MHSanaei/3x-ui/commit/3aa76ea0) (#5462)
- [fix(sockopt): honor trustedXForwardedFor on gRPC inbounds (xray v26.6.22)](https://github.com/MHSanaei/3x-ui/commit/a0f4c13d) (#5503) @rqzbeh
- [fix(tls): pin remote cert via native uTLS handshake instead of xray subprocess](https://github.com/MHSanaei/3x-ui/commit/04832738)
- [fix(tls): ping the inbound's own port for remote cert pinning](https://github.com/MHSanaei/3x-ui/commit/03e89683)
- [fix(tls): default OCSP stapling to off for new inbound certs](https://github.com/MHSanaei/3x-ui/commit/39774a6a)
- [fix(inbound): strip XHTTP client-only fields from xray config, keep for subscriptions](https://github.com/MHSanaei/3x-ui/commit/cdaf5f80) (#5349) @nima1024m
- [fix(inbounds): flag conflicts with the reserved Xray API port](https://github.com/MHSanaei/3x-ui/commit/0d87bb8b) (#5304)
- [fix(inbound): regenerate SS-2022 client PSKs on method key-size change](https://github.com/MHSanaei/3x-ui/commit/98259596)
- [fix(inbound): persist streamSettings for tunnel so sockopt saves](https://github.com/MHSanaei/3x-ui/commit/315ecc25)
- [fix(reality): load `dest` as `target` alias so existing inbounds aren't wiped](https://github.com/MHSanaei/3x-ui/commit/66a9a788) (#5295) @volov-de
- [fix(routing): sync xray rules when panel inbound tags change or are deleted](https://github.com/MHSanaei/3x-ui/commit/af3f4600) (#5367) @nima1024m
- [fix(outbounds): test subscriptions in Test All, skip direct/dns](https://github.com/MHSanaei/3x-ui/commit/1c0fdb45)
- [fix(outbound): parse xmux from imported share links](https://github.com/MHSanaei/3x-ui/commit/c1fbfd05) (#5353)
- [fix(outbound): preserve non-ASCII characters in imported subscription tags](https://github.com/MHSanaei/3x-ui/commit/f7ffe898) (#5354)
- [fix(clients): centre the online dot inside the Online tag](https://github.com/MHSanaei/3x-ui/commit/8f556fe2) (#5238)
- [fix(clients): keep the client list live with a background poll](https://github.com/MHSanaei/3x-ui/commit/355262e6) (#5262)
- [fix(client): clear group when removed in the single-client editor](https://github.com/MHSanaei/3x-ui/commit/3088e964)
- [fix(frontend): TProxy schema, VLESS+XHTTP flow links, clearable Jalali date picker](https://github.com/MHSanaei/3x-ui/commit/f00512d1) (#5339, #5322, #5313)
- [fix(frontend): guard IntlUtil.formatDate against out-of-range timestamps](https://github.com/MHSanaei/3x-ui/commit/5d88e688) (#5468) @NikanZeyaei
- [fix(api-docs): exclude /panel/outbound and /panel/routing from route guard](https://github.com/MHSanaei/3x-ui/commit/68365367)
- [fix(tgbot): clear legacy panelProxy/tgBotProxy settings on upgrade](https://github.com/MHSanaei/3x-ui/commit/9a8247fa)
- [fix(tgbot): dedupe exhausted-client report by email](https://github.com/MHSanaei/3x-ui/commit/1259c20e) (#5453)
- [fix(settings): rename remark model 'Other' to 'External Proxy'](https://github.com/MHSanaei/3x-ui/commit/2d6dea4b) (#5265)
- [fix(iplimit): ban UDP as well as TCP in fail2ban action](https://github.com/MHSanaei/3x-ui/commit/cf5f37e4) (#5350)
- [fix(script): report per-file geo update status and skip restart when nothing changed](https://github.com/MHSanaei/3x-ui/commit/c200e248)
- [fix(cli): apply -webCert/-webCertKey on the setting subcommand](https://github.com/MHSanaei/3x-ui/commit/2392f04e) (#5482) @Taov-Russo
- [fix(install): support IPv6-only hosts](https://github.com/MHSanaei/3x-ui/commit/1b102ff9) (#5487) @m4tinbeigi-official
- [fix: resolve a batch of open bug-tagged issues (traffic accounting, share strategy, sub address, CPU)](https://github.com/MHSanaei/3x-ui/commit/679d2e1c) (#5477)

<h3> Reports </h3>

![total](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.4.0/total?label=total&color=success)
![amd64](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.4.0/x-ui-linux-amd64.tar.gz?label=linux-amd64)
![arm64](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.4.0/x-ui-linux-arm64.tar.gz?label=linux-arm64)
![386](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.4.0/x-ui-linux-386.tar.gz?label=linux-386)
![armv7](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.4.0/x-ui-linux-armv7.tar.gz?label=linux-armv7)
![armv6](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.4.0/x-ui-linux-armv6.tar.gz?label=linux-armv6)
![armv5](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.4.0/x-ui-linux-armv5.tar.gz?label=linux-armv5)
![s390x](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.4.0/x-ui-linux-s390x.tar.gz?label=linux-s390x)
![windows](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.4.0/x-ui-windows-amd64.zip?label=windows-amd64)

## New Contributors
* @surfdps made their first contribution in https://github.com/MHSanaei/3x-ui/pull/5240
* @n0ctal made their first contribution in https://github.com/MHSanaei/3x-ui/pull/5269
* @volov-de made their first contribution in https://github.com/MHSanaei/3x-ui/pull/5295
* @tonymoses10 made their first contribution in https://github.com/MHSanaei/3x-ui/pull/5343
* @Ssentiago made their first contribution in https://github.com/MHSanaei/3x-ui/pull/5326
* @IgorKha made their first contribution in https://github.com/MHSanaei/3x-ui/pull/5433
* @wahh3b-lgtm made their first contribution in https://github.com/MHSanaei/3x-ui/pull/5430
* @qin9125 made their first contribution in https://github.com/MHSanaei/3x-ui/pull/5459
* @Taov-Russo made their first contribution in https://github.com/MHSanaei/3x-ui/pull/5482
* @m4tinbeigi-official made their first contribution in https://github.com/MHSanaei/3x-ui/pull/5487

**Full Changelog**: https://github.com/MHSanaei/3x-ui/compare/v3.3.1...v3.4.0

### v3.4.1 (2026-06-25)
## 🚀 Rolling Dev Channel, Logs Viewer Overhaul, Leaner Memory & Client Bulk Ops

- 🧪 **Rolling Dev update channel** — opt into per-commit builds from the panel, node updates, the `x-ui.sh` menu, and the installer (`dev-latest`); the dev build version is surfaced in the UI, bot, and CLI, and dev nodes report `dev+<commit>` so they aren't flagged stale.
- 🧾 **Logs & Access Logs viewer overhaul** — the Xray access-log viewer is now labeled **Access Logs** across all languages, with an auto-update toggle, a 1000-row option, and verbatim rendering of plain log notices.
- 📉 **Leaner memory & tiered metrics history** — real process RSS reporting plus a smaller footprint via GOGC + periodic release, and a tiered rollup that keeps 7 days of metrics history at ~1.5 MB.
- 👥 **Client bulk operations** — bulk enable/disable and bulk-set XTLS flow from the Adjust dialog, with selection actions tidied into a **More** menu.
- 🧬 **VLESS encryption modes & tunnel health** — new VLESS encryption modes plus an Xray tunnel health monitor.
- 🔗 **Subscription engine upgrades** — new `PROTOCOL`/`TRANSPORT`/`SECURITY` remark variables, Incy client integration + routing tab, template-driven display remarks, and recovery of `{{TRAFFIC_USED}}` for orphaned traffic rows.
- 🛰️ **Node traffic history** — import per-client traffic history on a node-hosted inbound's first sync so totals don't start from zero.
- 🧹 **Uninstall & deploy cleanup** — the uninstaller now offers to purge PostgreSQL, and the legacy AWS golden-image build stack was dropped.

> ℹ️ **Heads-up:** The new **Dev** update channel ships rolling per-commit builds for testers — stable deployments stay on the release channel unless you switch. The uninstaller now also offers to **purge PostgreSQL** when removing the panel; decline if you share that database with other apps.

<h3>🆕 New</h3>

- [feat(update): add rolling dev update channel for per-commit builds](https://github.com/MHSanaei/3x-ui/commit/aad2b3eb)
- [feat(nodes): add Dev channel option to node panel updates](https://github.com/MHSanaei/3x-ui/commit/e8878b71)
- [feat(install): add dev-latest install option and sync README translations](https://github.com/MHSanaei/3x-ui/commit/2adb59bd)
- [feat(x-ui.sh): add Dev channel update option to the management menu](https://github.com/MHSanaei/3x-ui/commit/2830f97f)
- [feat(panel): surface dev-build version in UI, bot, and CLI](https://github.com/MHSanaei/3x-ui/commit/e4b881e5)
- [feat(web): vless encryption new modes](https://github.com/MHSanaei/3x-ui/commit/3ba43bd8) (#5517) @FunLay123
- [feat(xray): add tunnel health monitor](https://github.com/MHSanaei/3x-ui/commit/fe025e8a) (#5480) @m4tinbeigi-official
- [feat(clients): add bulk enable/disable and move selection actions into More menu](https://github.com/MHSanaei/3x-ui/commit/e64e9981)
- [feat(clients): bulk-set XTLS flow from the Adjust dialog](https://github.com/MHSanaei/3x-ui/commit/14de0557) (#5524) @rqzbeh
- [feat(logs): add auto-update toggle to Access Logs and Logs viewers](https://github.com/MHSanaei/3x-ui/commit/9381fa28)
- [feat(logs): label the Xray access-log viewer 'Access Logs' across all languages](https://github.com/MHSanaei/3x-ui/commit/e27f2490)
- [feat(logs): add 1000 rows option and drop 10 from log row count selectors](https://github.com/MHSanaei/3x-ui/commit/1d695082)
- [feat(sub): add PROTOCOL, TRANSPORT, SECURITY remark template variables](https://github.com/MHSanaei/3x-ui/commit/11c5b53f)
- [feat(sub): add Incy client integration and routing tab](https://github.com/MHSanaei/3x-ui/commit/48c2fb27)
- [feat(uninstall): offer to purge PostgreSQL when removing the panel](https://github.com/MHSanaei/3x-ui/commit/9dec15bd)

<h3>⚡ Update & improvement</h3>

- [perf(memory): report real RSS and cut footprint via GOGC + periodic release](https://github.com/MHSanaei/3x-ui/commit/69ad8b76)
- [perf(metrics): tiered rollup history (7d at ~1.5MB) and cleaner ranges](https://github.com/MHSanaei/3x-ui/commit/293c1e44)
- [chore: bump deps and modernize test loops](https://github.com/MHSanaei/3x-ui/commit/dc6d13b5)
- [chore(deploy): drop the AWS golden-image build stack](https://github.com/MHSanaei/3x-ui/commit/30796dc2)

<h3>🐞 Bug fixed</h3>

- [fix(web): remove deleted multi-inbound client from runtime regardless of shared email](https://github.com/MHSanaei/3x-ui/commit/896016f7) (#5543)
- [fix(web): show subscription outbounds in dialer proxy dropdown](https://github.com/MHSanaei/3x-ui/commit/e2d25d0a) (#5540)
- [fix(web): serve panel SPA routes from NoRoute](https://github.com/MHSanaei/3x-ui/commit/ae9bbdf2) (#5536) @w3struk
- [fix(node): import per-client traffic history on first sync of a node-hosted inbound](https://github.com/MHSanaei/3x-ui/commit/b32837e5)
- [fix(nodes): report dev builds as dev+<commit> so updated nodes aren't flagged stale](https://github.com/MHSanaei/3x-ui/commit/bcd13580)
- [fix(flow): restore XTLS Vision when an inbound becomes flow-eligible](https://github.com/MHSanaei/3x-ui/commit/82600936) (#5520) @rqzbeh
- [fix(inbounds): accept null rewritePort in tunnel settings](https://github.com/MHSanaei/3x-ui/commit/c93beef2) (#5516, #5525) @rqzbeh
- [fix(sub): recover {{TRAFFIC_USED}} for clients with orphaned traffic rows](https://github.com/MHSanaei/3x-ui/commit/a4be5a0d)
- [fix(sub): drive display remarks from the template and split multi-host subpage links](https://github.com/MHSanaei/3x-ui/commit/b0c1156d)
- [fix(sub): restore client email in panel copy/QR link remark](https://github.com/MHSanaei/3x-ui/commit/5dbd5b1d) (#5532)
- [fix(outbound): preserve custom headers for HTTP outbounds](https://github.com/MHSanaei/3x-ui/commit/bd60e770) (#5519)
- [fix(clients): use new email after rename and de-duplicate save toast](https://github.com/MHSanaei/3x-ui/commit/23e73cd4)
- [fix(tgbot): reload bot on settings save so a new token takes effect without a panel restart](https://github.com/MHSanaei/3x-ui/commit/93ff60e5)
- [fix(backup): name Telegram backups after webDomain/IP instead of x-ui](https://github.com/MHSanaei/3x-ui/commit/a5e865c1)
- [fix(update): read setUpdateChannel body as form field, not JSON](https://github.com/MHSanaei/3x-ui/commit/1d1128cf)
- [fix(hosts): show proper page title instead of falling back to 3X-UI](https://github.com/MHSanaei/3x-ui/commit/8f65aa7e)
- [fix(logs): render plain log notices verbatim instead of mangling them as timestamps](https://github.com/MHSanaei/3x-ui/commit/df0e52cd)

<h3> Reports </h3>

![total](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.4.1/total?label=total&color=success)
![amd64](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.4.1/x-ui-linux-amd64.tar.gz?label=linux-amd64)
![arm64](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.4.1/x-ui-linux-arm64.tar.gz?label=linux-arm64)
![386](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.4.1/x-ui-linux-386.tar.gz?label=linux-386)
![armv7](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.4.1/x-ui-linux-armv7.tar.gz?label=linux-armv7)
![armv6](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.4.1/x-ui-linux-armv6.tar.gz?label=linux-armv6)
![armv5](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.4.1/x-ui-linux-armv5.tar.gz?label=linux-armv5)
![s390x](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.4.1/x-ui-linux-s390x.tar.gz?label=linux-s390x)
![windows](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.4.1/x-ui-windows-amd64.zip?label=windows-amd64)

## New Contributors
* @FunLay123 made their first contribution in https://github.com/MHSanaei/3x-ui/pull/5517

**Full Changelog**: https://github.com/MHSanaei/3x-ui/compare/v3.4.0...v3.4.1

### v3.4.2 (2026-06-29)
## 🚀 WireGuard Multi-Client, Panel Accessibility, Balancer Observatory & Hardened Settings

- 🔐 **WireGuard goes multi-client** — WireGuard inbounds are now first-class multi-client (native `users`), with a reworked client-config UX, a collapsible config card, configurable DNS, and client IPs allocated inside the existing peer subnet.
- ♿ **Panel accessibility** — screen-reader and keyboard accessibility were brought across the whole panel.
- ⚖️ **Balancer Observatory** — a tabbed Observatory / Burst Observatory form, validation that defers errors until a field is touched or saved, and a burst observer for `random`/`roundRobin` strategies with a `fallbackTag`.
- 🛡️ **Hardened settings & restore** — sensitive setting changes now require re-confirming 2FA, and the database-restore body-cap exemption was tightened.
- 🛰️ **REALITY target scanner** — a live REALITY target scanner with IP/CIDR discovery.
- 🧪 **Dev channel from stable** — opt into the rolling dev channel directly from a stable build.
- 🔗 **Subscription & inbound polish** — Host VLESS Route baked into subscription UUIDs, `{{EMAIL}}` shown on the first sub-body link only, a remark template applied to *Export all* inbound links, and legacy `externalProxy` converted to hosts on import.
- 🧰 **Dev toolchain & deps** — a canonical `Makefile`, golangci-lint v2, plus Ant Design 6.5 and xray-core v26.6.27.

> ℹ️ **Heads-up:** WireGuard inbounds are now natively **multi-client** — existing single-peer inbounds migrate automatically to the clients model. The **minimum eligible Xray version was raised** (bundled core bumped to xray-core v26.6.27); update any custom cores accordingly. With 2FA enabled, **sensitive setting changes now prompt for a fresh 2FA confirmation** before they apply.

<h3>🆕 New</h3>

- [feat(wireguard): multi-client support](https://github.com/MHSanaei/3x-ui/commit/9c8cd08f)
- [feat(wireguard): client config UX, collapsible config card, configurable DNS](https://github.com/MHSanaei/3x-ui/commit/a329882e)
- [feat(a11y): screen-reader & keyboard accessibility across the panel](https://github.com/MHSanaei/3x-ui/commit/71aca201) (#5486, #5652) @nima1024m
- [feat(reality): add live REALITY target scanner with IP/CIDR discovery](https://github.com/MHSanaei/3x-ui/commit/6964d847)
- [feat(balancers): tabbed Observatory/Burst Observatory form](https://github.com/MHSanaei/3x-ui/commit/25a86b9e) (#5627) @nima1024m
- [feat(update): allow opting into the dev channel from a stable build](https://github.com/MHSanaei/3x-ui/commit/8e4c3682)
- [feat(groups): reset group traffic without touching client counters](https://github.com/MHSanaei/3x-ui/commit/9b8a0c9b)
- [feat(inbounds): apply remark template to Export all inbound links](https://github.com/MHSanaei/3x-ui/commit/439245d4)
- [feat(xhttp): default xmux maxConnections to 6](https://github.com/MHSanaei/3x-ui/commit/33aada0c)
- [feat(backup): prefix backup filenames with date and time](https://github.com/MHSanaei/3x-ui/commit/1bad2fcb) (#5606) @NikanZeyaei
- [feat: ldap skip tls verify](https://github.com/MHSanaei/3x-ui/commit/60c54827) (#5637) @NikanZeyaei
- [feat(sidebar): add documentation link button](https://github.com/MHSanaei/3x-ui/commit/451263f1)

<h3>⚡ Update & improvement</h3>

- [Bump minimum eligible Xray version](https://github.com/MHSanaei/3x-ui/commit/bbfbd7eb)
- [chore(deps): bump xray-core to v26.6.27](https://github.com/MHSanaei/3x-ui/commit/e44075a6)
- [chore(deps): bump antd to 6.5 and migrate deprecated component props](https://github.com/MHSanaei/3x-ui/commit/8332ba67)
- [chore: add Makefile as canonical task runner](https://github.com/MHSanaei/3x-ui/commit/2e851978)
- [style: adopt golangci-lint v2 and resolve all findings](https://github.com/MHSanaei/3x-ui/commit/fa1a19c0)
- [chore(ci): bump golangci-lint action to v9](https://github.com/MHSanaei/3x-ui/commit/d1c0d770)
- [docs: add CLAUDE.md agent guides for root and frontend](https://github.com/MHSanaei/3x-ui/commit/7efa0d9d)
- [docs: correct false RTL claim and stale Vite version in CONTRIBUTING.md](https://github.com/MHSanaei/3x-ui/commit/63fca9ef)
- [test(sub): align identity-token test with first-link-only EMAIL](https://github.com/MHSanaei/3x-ui/commit/d12b186a)
- [fix(lint): use errors.Is for io.EOF comparison in sys_linux](https://github.com/MHSanaei/3x-ui/commit/56b0be0b)

<h3>🐞 Bug fixed</h3>

- [fix(node): stop the offline-sync toast firing on saves to online nodes](https://github.com/MHSanaei/3x-ui/commit/86813758)
- [fix(clients): re-enable depleted clients on API renewal](https://github.com/MHSanaei/3x-ui/commit/789e92cd) (#5619)
- [fix(xray): clean stale routing references when a balancer or outbound is deleted](https://github.com/MHSanaei/3x-ui/commit/7a5d6da2) (#5648) @nima1024m
- [fix(clients): hide WireGuard config after detaching the WG inbound](https://github.com/MHSanaei/3x-ui/commit/6c71b725)
- [fix(wireguard): allocate client IPs in the existing peer subnet](https://github.com/MHSanaei/3x-ui/commit/79069d2b)
- [fix(sync): mark node dirty inside the mutation transaction (atomic ConfigDirty)](https://github.com/MHSanaei/3x-ui/commit/aef35ee0) (#5611) @n0ctal
- [fix(runtime): refresh cached node remotes on identity change](https://github.com/MHSanaei/3x-ui/commit/5713c099) (#5614) @n0ctal
- [fix(settings): require re-2FA confirmation for sensitive setting changes](https://github.com/MHSanaei/3x-ui/commit/2b10808f) (#5610) @n0ctal
- [fix(web): tighten database restore body-cap exemption](https://github.com/MHSanaei/3x-ui/commit/7f8cbf4c) (#5609) @n0ctal
- [fix(balancers): defer validation errors until touched or save](https://github.com/MHSanaei/3x-ui/commit/51ffba59) (#5626) @nima1024m
- [fix(balancers): create burst observer for random/roundRobin with fallbackTag](https://github.com/MHSanaei/3x-ui/commit/797b08cd)
- [fix(sub): bake Host VLESS Route into subscription UUIDs](https://github.com/MHSanaei/3x-ui/commit/d8221a81)
- [fix(sub): show {{EMAIL}} on first sub-body link only](https://github.com/MHSanaei/3x-ui/commit/876d55f2)
- [fix(inbound): convert legacy externalProxy to hosts on import](https://github.com/MHSanaei/3x-ui/commit/39eb5baf)
- [fix(shadowsocks): send per-user Account for SS-2022 runtime AddUser](https://github.com/MHSanaei/3x-ui/commit/4c177f0c)
- [fix(routing): write lowercase L4 network to xray config, display uppercase in UI](https://github.com/MHSanaei/3x-ui/commit/535b89a3)
- [fix(settings): normalize API token timestamps](https://github.com/MHSanaei/3x-ui/commit/7a217953) (#5599) @Tomilla
- [fix(logger): prevent nil-deref panic in migrate/setting CLI paths](https://github.com/MHSanaei/3x-ui/commit/522b1b64)

<h3> Reports </h3>

![Total Download](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.4.2/total?label=Total-Download&color=success)


## New Contributors
* @Tomilla made their first contribution in https://github.com/MHSanaei/3x-ui/pull/5599

**Full Changelog**: https://github.com/MHSanaei/3x-ui/compare/v3.4.1...v3.4.2

### v3.5.0 (2026-07-12)
## 🚀 MTProto Multi-Client, SQLite→PostgreSQL Migration, 500k-Scale Performance & Node Sync Hardening

- 📬 **MTProto goes multi-client** — MTProto inbounds now run on the `mtg-multi` engine with one FakeTLS secret per client, per-client ad-tags, and per-client quota & expiry enforced in the sidecar; client edits hot-apply through a management API so live connections survive.
- 🐘 **SQLite → PostgreSQL, end to end** — the PostgreSQL panel restore now accepts SQLite `.db` files and migration dumps directly (uploads are sniffed automatically), cross-db migration became lossless, transactional and pre-checked, and a new `x-ui pgclient` command installs or upgrades the PostgreSQL client tools.
- 🚄 **Built for 500k clients** — batched ip-limit lookups, depleted-client disables by id, delta WebSocket stats above a snapshot threshold, and subscriptions resolved from normalized tables — pinned by a scale test suite at 500,000 clients.
- 🧠 **Frontend platform overhaul** — full React Hook Form migration, axios replaced with the native Fetch API, charts moved to uPlot, and Husky / lint-staged / MSW / Storybook dev tooling.
- 🛰️ **Node sync hardening** — host overrides adopted into the master, no more premature inbound sweeps or Postgres deadlocks in sync, client edits no longer tear down node inbounds, and node auto-renewals open a fresh quota window.
- 🌍 **Outbound insight** — egress metadata (IP + country) per outbound, a real-delay connection test measured on a warm connection, and a `targetStrategy` field in the outbound editor.
- ⚖️ **Routing & balancers** — balancer-to-balancer fallback, a default outbound in basic routing, encrypted DNS presets, and private-IP `dns.servers` allowed past the `geoip:private` block rule.
- 🔎 **Panel QoL** — text search on the inbound list and node selects, column sorting, per-client realtime speed, bulk-adding hosts to multiple inbounds, and WireGuard export split into config and links tabs.

> ℹ️ **Heads-up:** MTProto inbounds are now natively **multi-client** — legacy single-secret inbounds migrate automatically to the clients model, and `tg://` deep links no longer carry the remark fragment. The bundled core was bumped to **xray-core v26.7.11**, and the **Final Mask + REALITY combination is now rejected** (it crashes Xray-core). A database migration also **repairs overflowed traffic counters** and drops the legacy UNIQUE constraint on inbound ports.

<h3>🆕 New</h3>

- [feat(mtproto): adopt mtg-multi and make MTProto inbounds multi-client](https://github.com/MHSanaei/3x-ui/commit/d97bd864)
- [feat(mtproto): per-client ad-tags, management-API auth, and secret sync](https://github.com/MHSanaei/3x-ui/commit/43500a54)
- [feat(mtproto): enforce per-client quota & expiry via mtg-multi limits](https://github.com/MHSanaei/3x-ui/commit/328d920e)
- [feat(db): import SQLite migration dumps through the PostgreSQL panel restore](https://github.com/MHSanaei/3x-ui/commit/30b61161)
- [feat(server): sniff SQLite panel restore uploads and keep the fallback on failure](https://github.com/MHSanaei/3x-ui/commit/77dffe9a)
- [feat(db): add pgclient command to install or upgrade PostgreSQL client tools](https://github.com/MHSanaei/3x-ui/commit/f36f481e)
- [feat(outbound): add outbound egress metadata (IP + country)](https://github.com/MHSanaei/3x-ui/commit/30f6bc18) (#5886) @isultanov99
- [feat(outbound): add real-delay connection test mode](https://github.com/MHSanaei/3x-ui/commit/ed66209e)
- [feat(balancer): add balancer-to-balancer fallback support](https://github.com/MHSanaei/3x-ui/commit/142dab9e) (#5586) @Ssentiago
- [feat(xray): default outbound in basic routing](https://github.com/MHSanaei/3x-ui/commit/ea24ef0a) (#5815) @rqzbeh
- [feat(dns): add encrypted DNS presets](https://github.com/MHSanaei/3x-ui/commit/b8a65496) (#5837) @rqzbeh
- [feat(hosts): bulk-add multiple hosts to multiple inbounds](https://github.com/MHSanaei/3x-ui/commit/42690e1b) (#5677) @AmirRnz
- [feat(ui): per-client realtime speed](https://github.com/MHSanaei/3x-ui/commit/b177e307) (#5687) @NikanZeyaei
- [feat(frontend): add text search to the inbound list](https://github.com/MHSanaei/3x-ui/commit/05cb70d8)
- [feat(inbounds): add column sorting to the inbounds table](https://github.com/MHSanaei/3x-ui/commit/f9cd7ac9) (#5661) @isultanov99
- [feat(frontend): add text search to node select components](https://github.com/MHSanaei/3x-ui/commit/dd4f55f6)
- [feat(frontend): treat WireGuard inbounds as multi-user in client actions](https://github.com/MHSanaei/3x-ui/commit/c4a1139d)
- [feat(frontend): split WireGuard inbound export into config and links tabs](https://github.com/MHSanaei/3x-ui/commit/44f2f426)
- [feat(wireguard): make client allowedIPs editable with validation](https://github.com/MHSanaei/3x-ui/commit/64c30603)
- [feat(reality): derive a stable per-client spiderX for shared links](https://github.com/MHSanaei/3x-ui/commit/c8ef1b1f)
- [feat(web): broadcast delta client stats above a snapshot threshold](https://github.com/MHSanaei/3x-ui/commit/fc5be5b9)
- [feat(sub): show the announcement on the subscription info page](https://github.com/MHSanaei/3x-ui/commit/323cf09d)
- [feat(sub): serve the HTML info page for browser requests on JSON and Clash URLs](https://github.com/MHSanaei/3x-ui/commit/ff3bd636)
- [feat(settings): let users clear stored secrets from the UI](https://github.com/MHSanaei/3x-ui/commit/92303094)
- [feat(clients): hide disabled inbounds in the client form selector](https://github.com/MHSanaei/3x-ui/commit/052dd85a)
- [feat(frontend): show client group in the client info modal](https://github.com/MHSanaei/3x-ui/commit/1afab47f)
- [feat(frontend): add targetStrategy field to the outbound editor](https://github.com/MHSanaei/3x-ui/commit/258d8b73)
- [feat(clients): clarify which protocols use the password field](https://github.com/MHSanaei/3x-ui/commit/6e75938c) (#5809) @ecgang
- [feat(tgbot): register usage, inbound, restart and clearall in the bot command menu](https://github.com/MHSanaei/3x-ui/commit/1f04912b)
- [feat(tgbot): show inbound remark alongside email in the online clients list](https://github.com/MHSanaei/3x-ui/commit/220dcb15)
- [feat(tgbot): include hostname in backup and ban-log messages](https://github.com/MHSanaei/3x-ui/commit/b2ceb854)

<h3>⚡ Update & improvement</h3>

- [feat(xray): update xray-core to v26.7.11 and adapt panel](https://github.com/MHSanaei/3x-ui/commit/814cda3f)
- [chore(mtproto): bump the mtg-multi binary pin to v1.14.0](https://github.com/MHSanaei/3x-ui/commit/406ce54f)
- [refactor(mtproto): manage ad-tags per client only](https://github.com/MHSanaei/3x-ui/commit/ad7a0f81)
- [Frontend dev tooling (Husky, lint-staged, MSW, Storybook) + full React Hook Form migration](https://github.com/MHSanaei/3x-ui/commit/61e12e4c) (#5859)
- [refactor(frontend): replace axios with the native Fetch API](https://github.com/MHSanaei/3x-ui/commit/bc309ed9)
- [refactor(frontend): replace recharts with uPlot for charts](https://github.com/MHSanaei/3x-ui/commit/8ee79cf4)
- [test(scale): cover traffic poll, ws payloads, ip-limit job, sub and xray config at 500k](https://github.com/MHSanaei/3x-ui/commit/4fc30168)
- [refactor: modernize Go with strings.SplitSeq and maps.Copy](https://github.com/MHSanaei/3x-ui/commit/a067f817)
- [refactor: use the built-in max/min to simplify the code](https://github.com/MHSanaei/3x-ui/commit/07d66aa6) (#5751) @alaningtrump
- [chore: add golangci-lint tasks and force LF on Go files](https://github.com/MHSanaei/3x-ui/commit/5c5a5096)
- [docs: vendor the documentation site into the monorepo](https://github.com/MHSanaei/3x-ui/commit/9b91f0f4)
- [docs: move architecture map into docs/ and refresh it against the live tree](https://github.com/MHSanaei/3x-ui/commit/28f76902)
- [docs(settings): clarify Sub Port/Sub Domain double as subscription-link fallback](https://github.com/MHSanaei/3x-ui/commit/6e0067fc) (#5721) @volov-de
- [docs: update the env vars example file](https://github.com/MHSanaei/3x-ui/commit/dbdecda0) (#5678) @nebulosa2007
- [ci(claude-bot): auto-fix trusted PRs and easy issue bugs](https://github.com/MHSanaei/3x-ui/commit/7780ab0e)
- [ci(claude-bot): gate write capability to trusted actors](https://github.com/MHSanaei/3x-ui/commit/de5b1300)
- [chore(frontend): bump version and deps](https://github.com/MHSanaei/3x-ui/commit/f905c2dc)
- [chore(deps): bump google.golang.org/grpc to 1.82.0](https://github.com/MHSanaei/3x-ui/commit/6626bf4a) (#5729)
- [chore(deps): bump golang.org/x/text to 0.40.0](https://github.com/MHSanaei/3x-ui/commit/3d513b50) (#5872)

<h3>🐞 Bug fixed</h3>

- [fix(node): stop client edits from tearing down node inbounds and harden reconcile fingerprints](https://github.com/MHSanaei/3x-ui/commit/a0989e0f)
- [fix(node): stop Postgres deadlocks and deleted-client resurrection in node sync](https://github.com/MHSanaei/3x-ui/commit/9a3a12b2)
- [fix(node): never sweep a node's inbounds before their first adoption](https://github.com/MHSanaei/3x-ui/commit/200ea091)
- [fix(node): adopt a node inbound's host overrides into the master](https://github.com/MHSanaei/3x-ui/commit/cbd2940a)
- [fix(node): stop one rejected inbound from starving a node's traffic sync](https://github.com/MHSanaei/3x-ui/commit/d105b274)
- [fix(node): fully delete clients on nodes instead of only detaching them](https://github.com/MHSanaei/3x-ui/commit/b1fa76f9)
- [fix(node): start a fresh quota window when a node auto-renews a client](https://github.com/MHSanaei/3x-ui/commit/2c49dbf5)
- [fix(node): stop force-restarting a node's Xray when its clients auto-disable](https://github.com/MHSanaei/3x-ui/commit/4d6f2ddd)
- [fix(node): show the activated first-use deadline on the Clients page](https://github.com/MHSanaei/3x-ui/commit/8dd3b31e)
- [fix(inbound): scope port-conflict check to the stored node on update](https://github.com/MHSanaei/3x-ui/commit/2c28fa5f) (#5833) @yukh975
- [fix(panel): use the hosting node address for WireGuard client configs](https://github.com/MHSanaei/3x-ui/commit/f90e4a69) (#5679) @STRENCH0
- [fix(inbound): reject finalmask + REALITY combo (crashes Xray-core)](https://github.com/MHSanaei/3x-ui/commit/7db92d63) (#5861) @mrnickson-hue
- [fix(inbounds): apply runtime changes after the DB commit](https://github.com/MHSanaei/3x-ui/commit/f431e9cc) (#5768) @n0ctal
- [fix(xray): reconcile client auto-disable through the API instead of a forced restart](https://github.com/MHSanaei/3x-ui/commit/e5b56c94)
- [fix(xray): force full restart for inbounds with a VLESS reverse client](https://github.com/MHSanaei/3x-ui/commit/49773c18)
- [fix(inbounds): apply the legacy xhttp session-key migration when editing](https://github.com/MHSanaei/3x-ui/commit/539bcc89)
- [fix(xhttp): stop XMUX maxConcurrency from reverting on save](https://github.com/MHSanaei/3x-ui/commit/201d4731)
- [fix(ui): make the Happy Eyeballs toggle produce a config xray actually enables](https://github.com/MHSanaei/3x-ui/commit/579a9daa)
- [fix(ui): keep an explicit zero happy-eyeballs delay across the round trip](https://github.com/MHSanaei/3x-ui/commit/9d1a21b4)
- [fix(web): sync the VLESS generate-key dropdown with the encryption field](https://github.com/MHSanaei/3x-ui/commit/97e2c9e7)
- [fix(routing): allow dns.servers on private IPs past the geoip:private block rule](https://github.com/MHSanaei/3x-ui/commit/e424cc0f) (#5774) @volov-de
- [fix(balancers): keep mixed strategies on one observer](https://github.com/MHSanaei/3x-ui/commit/ade74eb3) (#5674) @nima1024m
- [fix(logs): limit Xray log growth](https://github.com/MHSanaei/3x-ui/commit/15faec62) (#5840) @rqzbeh
- [fix(client): stop duplicate client entries accumulating in inbound settings](https://github.com/MHSanaei/3x-ui/commit/5a7b3b73)
- [fix(clients): reuse stored credentials when re-adding an existing identity](https://github.com/MHSanaei/3x-ui/commit/3b731cd6)
- [fix(clients): rename client record atomically with inbound settings](https://github.com/MHSanaei/3x-ui/commit/c4448f4e)
- [fix(clients): finish deleting from every inbound when one fails](https://github.com/MHSanaei/3x-ui/commit/6aa87f4e)
- [fix(client): clean node_client_traffics rows when deleting a client](https://github.com/MHSanaei/3x-ui/commit/7cb2adf4)
- [fix(clients): parse only settings.clients across protocols](https://github.com/MHSanaei/3x-ui/commit/567a4ac4) (#5855) @n0ctal
- [fix(clients): surface bulk-reset auto-enable failures](https://github.com/MHSanaei/3x-ui/commit/7c183dbd) (#5763) @n0ctal
- [fix(clients): include Telegram ID in client list search](https://github.com/MHSanaei/3x-ui/commit/ed9686bf) (#5888) @mrnickson-hue
- [fix(frontend): show zero client count for mtproto and wireguard inbounds](https://github.com/MHSanaei/3x-ui/commit/476bec45)
- [fix(groups): keep group traffic totals stable across client resets and deletes](https://github.com/MHSanaei/3x-ui/commit/1153d5db)
- [fix(mtproto): stop dropping connections on client/inbound edits; add live updates + ad-tag](https://github.com/MHSanaei/3x-ui/commit/6214ff4e) (#5838)
- [fix(mtproto): stop persisting a vestigial inbound-level secret](https://github.com/MHSanaei/3x-ui/commit/84b64230)
- [fix(mtproto): drop the remark fragment from tg proxy deep links](https://github.com/MHSanaei/3x-ui/commit/27fd1989)
- [fix(sub): include native WireGuard clients in Clash and JSON subscriptions](https://github.com/MHSanaei/3x-ui/commit/d2efe9b0) (#5676) @STRENCH0
- [fix(wireguard): build peers in GenXrayInboundConfig so node reconcile keeps clients](https://github.com/MHSanaei/3x-ui/commit/cb5b3a80) (#5684) @STRENCH0
- [fix(sub): carry a host's Final Mask into raw share links](https://github.com/MHSanaei/3x-ui/commit/cc3303dd)
- [fix(sub): use configured spiderX instead of always randomizing](https://github.com/MHSanaei/3x-ui/commit/1f2e3e14)
- [fix(sub): default https:// for scheme-less support and profile URLs](https://github.com/MHSanaei/3x-ui/commit/fb3a1559)
- [fix(sub): resolve subscription clients and stats from normalized tables](https://github.com/MHSanaei/3x-ui/commit/7c12700c)
- [fix(sub): apply host Allow Insecure to Hysteria2 subscription links](https://github.com/MHSanaei/3x-ui/commit/f3e99058) (#5866)
- [fix(link): strip query and trailing slash when parsing ss:// port](https://github.com/MHSanaei/3x-ui/commit/affcf6c4) (#5895) @Djjanks
- [fix(link): sanitize numeric quicParams taken from a share link's fm= param](https://github.com/MHSanaei/3x-ui/commit/11e45e81)
- [fix(link): reject non-finite and clamp out-of-range quicParams from fm=](https://github.com/MHSanaei/3x-ui/commit/0753f5ee)
- [fix(database): make cross-db migration lossless, transactional, and pre-checked](https://github.com/MHSanaei/3x-ui/commit/54fc0fd4)
- [fix(db): clamp traffic counters below int64 max and repair overflowed rows](https://github.com/MHSanaei/3x-ui/commit/837cf5f2)
- [fix(database): drop the legacy UNIQUE constraint on inbounds.port](https://github.com/MHSanaei/3x-ui/commit/fc625d8f)
- [fix(db): probe dump readability before PostgreSQL import](https://github.com/MHSanaei/3x-ui/commit/de70ecb0)
- [fix(database): stop noisy per-startup errors in the Postgres server log](https://github.com/MHSanaei/3x-ui/commit/273f8872)
- [fix(job): batch ip-limit per-email lookups and persistence](https://github.com/MHSanaei/3x-ui/commit/c0d17e13)
- [fix(job): gate ip-limit scan on clients.limit_ip instead of parsing all settings](https://github.com/MHSanaei/3x-ui/commit/c3cc8b43)
- [fix(iplimit): ban a dead connection once instead of every scan](https://github.com/MHSanaei/3x-ui/commit/975b1f1a)
- [fix(traffic): disable depleted clients by id instead of a second full scan](https://github.com/MHSanaei/3x-ui/commit/97588dd0)
- [fix(traffic): persist delayed-start expiry only for converted clients](https://github.com/MHSanaei/3x-ui/commit/fb1d055b)
- [fix(ldap): convert default total GB to bytes when auto-creating clients](https://github.com/MHSanaei/3x-ui/commit/57300f44) (#5854) @mrnickson-hue
- [fix(ldap): attach auto-created clients to every configured inbound tag](https://github.com/MHSanaei/3x-ui/commit/52d4af71)
- [fix(tgbot): find clients by tgId regardless of settings JSON formatting](https://github.com/MHSanaei/3x-ui/commit/b6183271)
- [fix(outbound): measure HTTP test delay on a warm connection](https://github.com/MHSanaei/3x-ui/commit/b6873c7a)
- [fix(settings): repair legacy path settings that block every settings save](https://github.com/MHSanaei/3x-ui/commit/a335456c)
- [fix(ui): align the subUpdates limit with the backend and show the range](https://github.com/MHSanaei/3x-ui/commit/0add6398)
- [fix(frontend): stop group modals clearing selection on background refetch](https://github.com/MHSanaei/3x-ui/commit/9f760cf0)
- [fix(web): opt panel pages out of Cloudflare Rocket Loader](https://github.com/MHSanaei/3x-ui/commit/e6bef229)
- [fix(api): preserve 64-bit integer schema formats](https://github.com/MHSanaei/3x-ui/commit/2c95e292) (#5908) @sanmaxdev
- [fix: make all self-managed file downloads/installs atomic, with real completion status](https://github.com/MHSanaei/3x-ui/commit/9e13b32c) (#5711) @nima1024m
- [fix(docker): start crond and persist acme.sh state so cert renewal works](https://github.com/MHSanaei/3x-ui/commit/a13a79b2)
- [fix(script): confirm auto-detected public IPv4 before issuing IP certificate](https://github.com/MHSanaei/3x-ui/commit/1c789c3e)
- [fix(scripts): pass --force to acme.sh --installcert so it survives sudo](https://github.com/MHSanaei/3x-ui/commit/62f30390)
- [fix(script): stop logging an error when Enter accepts the default ACME port](https://github.com/MHSanaei/3x-ui/commit/5e9606aa)
- [fix(script): rename the Xray binary to xray-linux-arm32 on 32-bit ARM](https://github.com/MHSanaei/3x-ui/commit/b6d1caf9)
- [fix(script): make local PostgreSQL and fail2ban setup work on RHEL-family distros](https://github.com/MHSanaei/3x-ui/commit/1bf9e5d5)
- [fix(script): stop running full system upgrades via pacman -Syu on Arch](https://github.com/MHSanaei/3x-ui/commit/26e88c7b)
- [fix(update): avoid full dnf system upgrade](https://github.com/MHSanaei/3x-ui/commit/5361b56e) (#5717) @v-2841
- [fix(scripts): avoid rpm package upgrades before installs](https://github.com/MHSanaei/3x-ui/commit/ed95acdd) (#5750) @v-2841
- [fix(script): correct hardcoded menu option numbers in x-ui.sh](https://github.com/MHSanaei/3x-ui/commit/e11e587c) (#5787) @lxk955

<h3> Reports </h3>

![Total Download](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.5.0/total?label=Total-Download&color=success)


## New Contributors
* @isultanov99 made their first contribution in https://github.com/MHSanaei/3x-ui/pull/5661
* @AmirRnz made their first contribution in https://github.com/MHSanaei/3x-ui/pull/5677
* @STRENCH0 made their first contribution in https://github.com/MHSanaei/3x-ui/pull/5679
* @v-2841 made their first contribution in https://github.com/MHSanaei/3x-ui/pull/5717
* @alaningtrump made their first contribution in https://github.com/MHSanaei/3x-ui/pull/5751
* @lxk955 made their first contribution in https://github.com/MHSanaei/3x-ui/pull/5787
* @ecgang made their first contribution in https://github.com/MHSanaei/3x-ui/pull/5809
* @yukh975 made their first contribution in https://github.com/MHSanaei/3x-ui/pull/5833
* @mrnickson-hue made their first contribution in https://github.com/MHSanaei/3x-ui/pull/5854
* @Djjanks made their first contribution in https://github.com/MHSanaei/3x-ui/pull/5895
* @sanmaxdev made their first contribution in https://github.com/MHSanaei/3x-ui/pull/5908

**Full Changelog**: https://github.com/MHSanaei/3x-ui/compare/v3.4.2...v3.5.0


### v3.6.0 (2026-07-30)
## 🚀 Trend-First Overview, xray-core v26.7.28, Subscription Correctness & Panel Hardening

- 🧭 **Overview rebuilt as a command deck** — the ten small cards are gone: four vitals tiles with 72-sample sparklines, a two-series throughput chart, a TCP/UDP connections chart, a grouped system strip, and a sidebar that became an auto-collapsed icon rail expanding on hover.
- 🛡 **xray-core v26.7.28** — the XMC finalmask breaking change is absorbed end to end: incomplete masks are rejected at save time with the missing field named, and only the offending mask is dropped at config-generation time so one bad row can no longer take every inbound offline.
- 🔗 **Subscription output correctness** — a long run of link and format fixes (forwarded-URL trust, coalesced external refreshes, Clash scalars a YAML parser would misread, VLESS flow gating, external link names, WireGuard Allowed IPs) plus opt-in identity tokens on every link, raw download actions, live online status with a new `?format=info` endpoint, and User-Agent format auto-detection.
- 🔐 **Panel surface tightened** — `openapi.json` moved behind session auth (it was serving the whole admin API surface unauthenticated), node API tokens became write-only, production sourcemaps no longer ship inside the binary, and the default freedom rules block private-range egress.
- 🧹 **Two large verified audit sweeps** — 54 fixes from a repo-wide self-correcting audit and 16 more from a bug-label issue sweep, spanning email, node sync, subscriptions, xray config and the database.
- 🗃 **Data integrity** — SQLite backup snapshots are taken online, legacy string `tgId` values in inbound settings are repaired on upgrade, and `client_traffics` rows are no longer deleted for detached-but-alive clients or left stale when an email is reused.
- 🎛 **Settings UX** — settings sitting at their shipped default are tagged as such, clearing a port field keeps the stored port instead of writing zero, date-pickers commit on selection rather than on confirm, and the REALITY client version range is validated at save time.
- 🧰 **Frontend platform** — the component Storybook became a validated, fully covered workbench and is published on the docs site, every axe accessibility violation in the library is resolved, react-router 8 and Node 24 LTS landed, and 210 dead translation keys were deleted with a test that fails the build on new ones.

> ℹ️ **Heads-up:** The bundled core moved to **xray-core v26.7.28**, where the XMC finalmask `usernames` list was replaced by a **required `profiles` array** (username + UUID + both Mojang texture fields, no "default to Dream" fallback) — a mask saved by an older panel is now **rejected at save time**, and stripped from the generated config with a warning rather than failing the whole core, so anyone using that obfuscation must refill their profiles. A **database migration repairs legacy string `tgId`** values in inbound settings that previously broke every client operation on the affected inbound. Two surfaces changed behavior for existing setups: **`GET /panel/api/openapi.json` now requires an authenticated session** (it was reachable without one), and **node API tokens are write-only** — the API no longer returns them, so tooling that read a token back must store it at creation time. The default freedom `finalRules` also gain a **`geoip:private` block rule**, applied in place to installs still carrying the stock rules.

<h3>🆕 New</h3>

- [feat(ui): redesign the overview page as a trend-first command deck](https://github.com/MHSanaei/3x-ui/commit/c3fa73d5)
- [feat(ui): tag settings that sit at their shipped default value](https://github.com/MHSanaei/3x-ui/commit/87ebcc7a) (#6128) @PathGao
- [feat(ui): validate the REALITY client version range at save time](https://github.com/MHSanaei/3x-ui/commit/ca6955d8) (#6126) @PathGao
- [feat(inbounds): allow custom monthly traffic reset days](https://github.com/MHSanaei/3x-ui/commit/17e6b5a4) (#6071) @Ki-Seki
- [feat(sub): auto-detect subscription format by User-Agent](https://github.com/MHSanaei/3x-ui/commit/129f50d9) (#5826) @Tomilla
- [feat(sub): allow identity tokens on every subscription link](https://github.com/MHSanaei/3x-ui/commit/8f49327e) (#5935) @H-TTTTT
- [feat(sub): expose live online status and add ?format=info endpoint](https://github.com/MHSanaei/3x-ui/commit/cd674c8d)
- [feat(sub): add raw subscription download actions](https://github.com/MHSanaei/3x-ui/commit/123fac22) (#6017) @w3struk
- [feat(sub): add XHTTP session field compatibility in share links and subscriptions](https://github.com/MHSanaei/3x-ui/commit/041476a3) (#5929) @beehunt9r
- [feat(api): add GET endpoint to look up clients by Telegram ID](https://github.com/MHSanaei/3x-ui/commit/6af29959) (#5945) @kimfom01
- [feat(notifications): add a consecutive-failure threshold for outbound.down alerts](https://github.com/MHSanaei/3x-ui/commit/2b1308ca) (#5968) @yukh975
- [feat(frontend): show client comments on mobile cards](https://github.com/MHSanaei/3x-ui/commit/658e6ab3) (#5942) @sanmaxdev

<h3>⚡️ Update & improvement</h3>

- [feat(xray): update xray-core to v26.7.28 and adapt panel](https://github.com/MHSanaei/3x-ui/commit/7f7b7e16)
- [perf(clients): make the clients page scale to large panels](https://github.com/MHSanaei/3x-ui/commit/f52c3c48)
- [perf(clients): take one email snapshot per client fan-out, not one per inbound](https://github.com/MHSanaei/3x-ui/commit/c3967e57) (#6091)
- [feat(frontend): make Storybook a validated, fully covered component workbench](https://github.com/MHSanaei/3x-ui/commit/7078abc1)
- [feat(docs): publish the component Storybook on the docs site](https://github.com/MHSanaei/3x-ui/commit/df3ba568)
- [refactor(frontend): migrate off deprecated Ant Design 6 props](https://github.com/MHSanaei/3x-ui/commit/ee9a6067)
- [refactor(ui): share one onNumber handler for numeric setting inputs](https://github.com/MHSanaei/3x-ui/commit/411271b4) (#6127) @PathGao
- [chore(lint): forbid the Number-or-clamp idiom in direct-write settings pages](https://github.com/MHSanaei/3x-ui/commit/55f02816) (#6129) @PathGao
- [chore(build): stop shipping production sourcemaps inside the binary](https://github.com/MHSanaei/3x-ui/commit/bcd71c92) (#6131) @PathGao
- [chore(i18n): delete 230 dead translation keys and guard against new ones](https://github.com/MHSanaei/3x-ui/commit/ea358843) (#6132) @PathGao
- [style(i18n): normalize Chinese-English spacing](https://github.com/MHSanaei/3x-ui/commit/48675ff1) (#6076) @Yosyoo
- [chore(deps): migrate to react-router 8 and refresh frontend dependencies](https://github.com/MHSanaei/3x-ui/commit/edb487a0)
- [chore: standardize the toolchain on Node 24 LTS](https://github.com/MHSanaei/3x-ui/commit/bbc41637)
- [chore: refresh dependencies, fix Linux tool tasks, modernize Go idioms](https://github.com/MHSanaei/3x-ui/commit/f4e79e70)
- [fix(docs): force transitive sharp up to patched 0.35.3](https://github.com/MHSanaei/3x-ui/commit/0b601543)
- [chore(deps): bump google.golang.org/grpc from 1.82.0 to 1.82.1](https://github.com/MHSanaei/3x-ui/commit/28b360f2) (#5994)
- [chore(deps): bump github.com/go-ldap/ldap/v3 from 3.4.13 to 3.4.14](https://github.com/MHSanaei/3x-ui/commit/8dfb639d) (#5993)
- [chore(deps): bump react-i18next from 17.0.9 to 17.0.10](https://github.com/MHSanaei/3x-ui/commit/46007711) (#5996)
- [chore(deps-dev): bump vite from 8.1.4 to 8.1.5](https://github.com/MHSanaei/3x-ui/commit/b9750438) (#5997)
- [chore(deps): bump actions/setup-go from 6 to 7](https://github.com/MHSanaei/3x-ui/commit/444e1e59) (#5995)
- [chore(deps): bump actions/setup-node from 6 to 7](https://github.com/MHSanaei/3x-ui/commit/73b479e5) (#5992)
- [refactor(ci): make the bot read-only except for PR conflict resolution](https://github.com/MHSanaei/3x-ui/commit/acbb879f)
- [fix(ci): survive transient GitHub 5xx outages in the release workflow](https://github.com/MHSanaei/3x-ui/commit/ab1a9228)
- [fix(ci): resolve the mtg-multi tag from the release-page redirect](https://github.com/MHSanaei/3x-ui/commit/455d1cd0)
- [fix(ci): publish dev-latest edit-first instead of probing for existence](https://github.com/MHSanaei/3x-ui/commit/2f156c8e)

<h3>🐞 Bug fixed</h3>

- [Repo-wide self-correcting audit: 54 verified bug fixes](https://github.com/MHSanaei/3x-ui/commit/5e1cb769) (#5970)
- [Bug-label issue sweep: 16 fixes](https://github.com/MHSanaei/3x-ui/commit/892c06c8) (#6083)
- [fix(api): authenticate GET /panel/api/openapi.json + pin the route registry to the router](https://github.com/MHSanaei/3x-ui/commit/33f72f8f) (#6133) @PathGao
- [fix(nodes): make node API tokens write-only](https://github.com/MHSanaei/3x-ui/commit/c77608bc) (#5613) @n0ctal
- [fix(xray): reject configs xray-core refuses, and check the fixtures against it](https://github.com/MHSanaei/3x-ui/commit/dc6a1601)
- [fix(xray): stop the runtime user API from crashing xray-core](https://github.com/MHSanaei/3x-ui/commit/fea6a20f)
- [fix(xray): emit an empty client array instead of null in the generated config](https://github.com/MHSanaei/3x-ui/commit/6f4cc1e5) (#6117)
- [fix(xray): validate generated egress targets](https://github.com/MHSanaei/3x-ui/commit/79e65f63) (#5989) @Alishahrokhiii
- [fix(xray): gate embedded unencrypted-outbound rejection on the running core version](https://github.com/MHSanaei/3x-ui/commit/d38c912d) (#6028) @mvanhorn
- [fix(xray): synchronize lifecycle state](https://github.com/MHSanaei/3x-ui/commit/ad5f2a28) (#6138) @PathGao
- [fix(balancer): pin loopback routing rules ahead of general rules](https://github.com/MHSanaei/3x-ui/commit/a9d5d9af) (#6054) @H-TTTTT
- [fix(mtproto): synchronize child-process lifecycle](https://github.com/MHSanaei/3x-ui/commit/03cc80bb) (#6141) @PathGao
- [fix(sub): honor trustedProxyCIDRs before forwarded URLs](https://github.com/MHSanaei/3x-ui/commit/ad288a7e) (#6135) @PathGao
- [fix(sub): coalesce external subscription refreshes](https://github.com/MHSanaei/3x-ui/commit/e467b25f) (#6139) @PathGao
- [fix(sub): quote Clash scalars a YAML parser would read as numbers](https://github.com/MHSanaei/3x-ui/commit/7fe9932d) (#6104)
- [fix(sub): gate the VLESS flow in JSON subscriptions like raw and Clash links](https://github.com/MHSanaei/3x-ui/commit/29557e21)
- [fix(sub): omit hyphen for empty remark variables](https://github.com/MHSanaei/3x-ui/commit/e862d81c) (#6101) @Tosd0
- [fix(sub): drop duplicated fingerprint in external-proxy tlsSettings](https://github.com/MHSanaei/3x-ui/commit/8bbca76b) (#6096) @n0liu
- [fix(sub): omit non-standard fm param from Hysteria2 URI](https://github.com/MHSanaei/3x-ui/commit/fde66ba8) (#6048) @H-TTTTT
- [fix(sub): preserve external link names in Clash/JSON](https://github.com/MHSanaei/3x-ui/commit/11f602fe) (#6049) @H-TTTTT
- [fix(sub): send the routing-enable header only when the toggle is on](https://github.com/MHSanaei/3x-ui/commit/c87649d9) (#6008) @a-poluyan
- [fix(wireguard): widen the client address pool past a full /24](https://github.com/MHSanaei/3x-ui/commit/aa60d54e) (#6089)
- [fix(wireguard): preserve all Allowed IPs in share link, .conf, and subscription](https://github.com/MHSanaei/3x-ui/commit/c80e5e27) (#6051) @H-TTTTT
- [fix(warp): preserve outbound customization when rotating IP](https://github.com/MHSanaei/3x-ui/commit/22ff07b2) (#6052) @H-TTTTT
- [fix(clients): keep a client editable when its subId is already shared](https://github.com/MHSanaei/3x-ui/commit/a652cb8c) (#6065)
- [fix(clients): allow case-only email updates without duplicates](https://github.com/MHSanaei/3x-ui/commit/a0dec000) (#6050) @H-TTTTT
- [fix(clients): persist all editable fields for clients with no inbound](https://github.com/MHSanaei/3x-ui/commit/d623410c) (#6053) @H-TTTTT
- [fix(clients): keep VLESS xtls-rprx-vision flow when inbound options reload](https://github.com/MHSanaei/3x-ui/commit/9e117bbd) (#5971) @sleepingF0x
- [fix(clients): stop deleting client_traffics for detached-but-alive clients](https://github.com/MHSanaei/3x-ui/commit/ff954ec4) (#6110) @mrnickson-hue
- [fix(clients): refresh stale client_traffics row when an inbound-deleted client's email is reused](https://github.com/MHSanaei/3x-ui/commit/8cd71e07) (#6003) @mrnickson-hue
- [fix(nodes): keep the credential-presence flag on the node heartbeat push](https://github.com/MHSanaei/3x-ui/commit/4605f00a)
- [fix(node): stop a departed master's frozen traffic from disabling clients](https://github.com/MHSanaei/3x-ui/commit/f8e9f2f0) (#6113)
- [fix(database): create SQLite backup snapshots online](https://github.com/MHSanaei/3x-ui/commit/af5a8e5d) (#6137) @PathGao
- [fix(database): repair legacy string tgId in inbound settings on upgrade](https://github.com/MHSanaei/3x-ui/commit/16b2bcf9)
- [fix(hosts): assign group ids to imported hosts and repair empty ones](https://github.com/MHSanaei/3x-ui/commit/8ef2eec3)
- [fix(dns): stop forcing port 53 on DoH/DoQ DNS server entries](https://github.com/MHSanaei/3x-ui/commit/ae0da4c5) (#5950) @mvanhorn
- [fix(frontend): keep DNS hosts synchronized](https://github.com/MHSanaei/3x-ui/commit/2c943da3) (#6158) @PathGao
- [fix(settings): keep the stored port when a port field is cleared](https://github.com/MHSanaei/3x-ui/commit/579acbc6) (#6121) @PathGao
- [fix(ui): commit date-picker selections immediately instead of on confirm](https://github.com/MHSanaei/3x-ui/commit/60498659) (#6122) @PathGao
- [fix(ui): explain the REALITY client version gate and drop the impossible placeholder](https://github.com/MHSanaei/3x-ui/commit/a2774bf2) (#6125) @PathGao
- [fix(frontend): preserve cancellation and reject invalid query data](https://github.com/MHSanaei/3x-ui/commit/86347378) (#6143) @PathGao
- [fix(frontend): preserve edited server drafts](https://github.com/MHSanaei/3x-ui/commit/66740b7e) (#6156) @PathGao
- [fix(frontend): preserve theme body classes](https://github.com/MHSanaei/3x-ui/commit/8d02ae28) (#6157) @PathGao
- [fix(frontend): stabilize speed tags on inbound and client pages](https://github.com/MHSanaei/3x-ui/commit/f2b17397) (#5930) @H-TTTTT
- [fix(frontend): resolve every axe accessibility violation in the component library](https://github.com/MHSanaei/3x-ui/commit/60316c83)
- [fix(job): bound the traffic-notify POST so a stalled receiver can't wedge it](https://github.com/MHSanaei/3x-ui/commit/0e69f64e) (#6115)
- [fix(email): build an RFC 5322 message with a proper From address and name](https://github.com/MHSanaei/3x-ui/commit/1cfd7b49) (#5941) @yukh975
- [fix(install.sh): use realpath instead of script name](https://github.com/MHSanaei/3x-ui/commit/34d2591e) (#6075) @Intervence
- [fix(script): remove release download time limit](https://github.com/MHSanaei/3x-ui/commit/65b5074b) (#5952) @sanmaxdev
- [fix(script): remove old mtg binary](https://github.com/MHSanaei/3x-ui/commit/b18c87dc) (#5955) @cherts

<h3> Reports </h3>

![Total Download](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.6.0/total?label=Total-Download&color=success)


## New Contributors
* @H-TTTTT made their first contribution in https://github.com/MHSanaei/3x-ui/pull/5930
* @kimfom01 made their first contribution in https://github.com/MHSanaei/3x-ui/pull/5945
* @mvanhorn made their first contribution in https://github.com/MHSanaei/3x-ui/pull/5950
* @sleepingF0x made their first contribution in https://github.com/MHSanaei/3x-ui/pull/5971
* @Alishahrokhiii made their first contribution in https://github.com/MHSanaei/3x-ui/pull/5989
* @mrnickson-hue made their first contribution in https://github.com/MHSanaei/3x-ui/pull/6003
* @a-poluyan made their first contribution in https://github.com/MHSanaei/3x-ui/pull/6008
* @Ki-Seki made their first contribution in https://github.com/MHSanaei/3x-ui/pull/6071
* @Intervence made their first contribution in https://github.com/MHSanaei/3x-ui/pull/6075
* @Yosyoo made their first contribution in https://github.com/MHSanaei/3x-ui/pull/6076
* @n0liu made their first contribution in https://github.com/MHSanaei/3x-ui/pull/6096
* @Tosd0 made their first contribution in https://github.com/MHSanaei/3x-ui/pull/6101
* @PathGao made their first contribution in https://github.com/MHSanaei/3x-ui/pull/6121

**Full Changelog**: https://github.com/MHSanaei/3x-ui/compare/v3.5.0...v3.6.0


### v3.7.0 (2026-08-24)
## 🚀 Native AmneziaWG, Calendar-Day Renewals, Scoped API Tokens & Node-Sync Hardening

- 🛡️ **Native AmneziaWG 3.1** — AmneziaWG inbounds run in-process on an embedded userspace device with a reconcile manager, so DPI-resistant peers need no kernel module, no Docker and no second panel.
- 🌐 **More outbound and routing reach** — PIA WireGuard outbounds added by login, remote routing URLs, a client picker inside user rules, and geosite/geoip categories browsable straight from a routing rule.
- 📅 **Client lifecycle you can actually bill against** — renewals that step whole calendar months, a per-client traffic reset cycle, a cap on auto-renewals, single-device HWID removal, per-client subscription HWID limits and per-client external-link controls.
- 🔑 **Credentials that can be scoped and revoked** — API tokens now carry scopes and an optional expiry, node API tokens can be encrypted at rest, replacing the TOTP secret requires a 2FA code, and the CLI stopped quietly accumulating admin-equivalent tokens.
- 🔗 **Subscription output** — client-side balancers in the JSON format, template variables in subscription metadata, an exposed last-fetch time, and a neutral copy-only page when a subscription URL is opened in a browser.
- 🕸️ **Multi-node hardening** — the sync adopts a matching deployed inbound instead of recreating it, keeps disabled inbounds a snapshot cannot report, stops undoing client extensions, validates every certificate in the mTLS trust bundle, and applies a rotated master certificate without a panel restart.
- 🧱 **Transactional correctness** — bulk client pushes, traffic maintenance side effects and group moves now fire only after the commit lands, and a half-applied startup migration can no longer commit silently.
- 🧰 **Toolchain and supply chain** — Go 1.27, TypeScript 7 on the oxc toolchain, provenance plus SBOM attestations on published images, and release binaries stamped with their source revision.
- 🖥️ **Panel polish** — a pinnable sidebar, network-only PWA installability, virtualized tables, real loading spinners, and localized log levels, access events and calendar labels.

> ℹ️ **Heads-up:** First start runs **automatic schema migrations** (client reset-cycle and last-fetch columns, node sync orphan columns, a lowercase client-email index, API token scope/expiry, external-link normalization) — **take a database backup before upgrading**. Two defaults changed for existing setups: **importing a database now keeps this machine's own listen addresses, ports, base path, certificate paths and node identity** (clear the new checkbox to restore the old whole-file clone), and **opening a subscription URL in a browser returns a copy-only page** unless you append `html=1` or `view=html`. `x-ui setting -getApiToken` now **rotates a single `cli-fallback` token** instead of minting a new one, so previously printed CLI fallback tokens stop working; calendar renewal is opt-in per client, so clients left on day 0 keep the old rolling interval. Building from source now requires the **Go 1.27 toolchain**.

<h3>🆕 New</h3>

- [feat(amneziawg): add native AmneziaWG protocol support](https://github.com/MHSanaei/3x-ui/commit/effcccce) (#6105) @Kuzz007
- [feat(pia): add PIA login-and-add WireGuard outbounds](https://github.com/MHSanaei/3x-ui/commit/bd6a6aba) (#6272) @Masterain98
- [feat(inbound): DisableFlow — opt an inbound out of auto XTLS Vision](https://github.com/MHSanaei/3x-ui/commit/930a0ed5) (#5698) @FZ1010
- [feat(routing): add client picker to user rules](https://github.com/MHSanaei/3x-ui/commit/02002dc1) (#6271) @ZaneL1u
- [feat(xray): browse geosite/geoip categories from routing rules](https://github.com/MHSanaei/3x-ui/commit/d7698ec7) (#6165) @STRENCH0
- [feat(routing): add remote routing URL support](https://github.com/MHSanaei/3x-ui/commit/380aff4d) (#6168) @yelloduxx
- [feat(clients): renew on a calendar day instead of a rolling interval](https://github.com/MHSanaei/3x-ui/commit/b8903fad) (#6239) @n0ctal
- [feat(clients): give each client its own traffic reset cycle](https://github.com/MHSanaei/3x-ui/commit/5c7ca5b5) (#6240) @n0ctal
- [feat(clients): cap how many times a client may auto-renew](https://github.com/MHSanaei/3x-ui/commit/e940f30b) (#6238) @n0ctal
- [feat(clients): allow removing a single HWID device](https://github.com/MHSanaei/3x-ui/commit/1250fbb7) (#6265) @Kuzz007
- [feat(sub): add per-client subscription HWID limits](https://github.com/MHSanaei/3x-ui/commit/694ad6de) (#5802) @rqzbeh
- [feat(clients): add per-client external link controls](https://github.com/MHSanaei/3x-ui/commit/abd32099) (#5650) @fastnas2023
- [feat(limitip): let operators exempt trusted addresses from the IP limit](https://github.com/MHSanaei/3x-ui/commit/d6472740) (#6230) @n0ctal
- [feat(sub): client-side balancers for the JSON subscription](https://github.com/MHSanaei/3x-ui/commit/da01b763) (#6243) @DIMFLIX
- [feat(sub): add template variables to subscription metadata](https://github.com/MHSanaei/3x-ui/commit/2d669fa4) (#6163) @isultanov99
- [feat(sub): expose last subscription fetch time](https://github.com/MHSanaei/3x-ui/commit/2217213e) (#6217) @hunmar
- [feat(sub): warn when salamander settings cannot reach the client](https://github.com/MHSanaei/3x-ui/commit/dafd3c0e) (#6177) @n0ctal
- [feat(inbounds): add a narrow endpoint for subscription sort order](https://github.com/MHSanaei/3x-ui/commit/b4e44786) (#6179) @n0ctal
- [feat(api): scoped, optionally expiring API tokens](https://github.com/MHSanaei/3x-ui/commit/1230559e) (#6201) @n0ctal
- [feat(nodes): opt-in encryption at rest for the outbound node API token](https://github.com/MHSanaei/3x-ui/commit/1793a9b8) (#6186) @n0ctal
- [feat(frontend): multi-node cloning initial implementation](https://github.com/MHSanaei/3x-ui/commit/8c8556ab) (#6216) @rlex
- [feat(inbounds): improve multi-node online attribution](https://github.com/MHSanaei/3x-ui/commit/be70535b) (#6164) @isultanov99
- [feat(server): keep this machine's own settings when importing a database](https://github.com/MHSanaei/3x-ui/commit/6e80a468) (#6227) @n0ctal
- [feat(ui): let users pin the sidebar](https://github.com/MHSanaei/3x-ui/commit/5373786f) (#6161) @PathGao
- [feat(web): add network-only PWA installability](https://github.com/MHSanaei/3x-ui/commit/3a2f9b48) (#6190) @korsun009
- [feat(i18n): translate the log levels, access events and calendar labels](https://github.com/MHSanaei/3x-ui/commit/5c9268c4) (#6226) @n0ctal

<h3>⚡ Update & improvement</h3>

- [chore(build): bump Go toolchain to 1.27.0](https://github.com/MHSanaei/3x-ui/commit/81fcacab)
- [Move to TypeScript 7 and the oxc toolchain (oxlint + oxfmt)](https://github.com/MHSanaei/3x-ui/commit/92fb94d8) (#6262)
- [chore(lint): adapt to staticcheck v0.8.0 under golangci-lint v2.13.1](https://github.com/MHSanaei/3x-ui/commit/e4798a02)
- [chore: bump dependencies and clear deprecated frontend APIs](https://github.com/MHSanaei/3x-ui/commit/fcf60eb2)
- [chore(deps): bump google.golang.org/grpc from 1.82.1 to 1.83.0](https://github.com/MHSanaei/3x-ui/commit/138e1bd8) (#6162)
- [chore(deps): bump github.com/klauspost/compress from 1.19.1 to 1.19.2](https://github.com/MHSanaei/3x-ui/commit/e2f75aca) (#6212)
- [chore(deps): bump dompurify](https://github.com/MHSanaei/3x-ui/commit/75032fd4) (#6193)
- [chore(deps-dev): bump postcss](https://github.com/MHSanaei/3x-ui/commit/d1423073) (#6173)
- [chore(frontend): resolve the high-severity brace-expansion advisory](https://github.com/MHSanaei/3x-ui/commit/7eacce6a) (#6180) @n0ctal
- [perf(clients): write client_inbounds deltas and check identity from the clients table](https://github.com/MHSanaei/3x-ui/commit/f7db247b)
- [perf(frontend): replace blank Suspense fallbacks with Spin, switch to matchMedia hook, add virtual table scrolling](https://github.com/MHSanaei/3x-ui/commit/8e7fb144) (#6187) @PathGao
- [refactor(tgbot): share numeric keypad transitions](https://github.com/MHSanaei/3x-ui/commit/238e4bb3) (#6211) @n0ctal
- [test(tgbot): detect open-coded keypad transitions](https://github.com/MHSanaei/3x-ui/commit/aecbad3a) (#6214) @n0ctal
- [ci: attach provenance and SBOM attestations to the published images](https://github.com/MHSanaei/3x-ui/commit/7c8a9a69) (#6130) @kobihikri
- [ci(release): stamp released binaries with their source revision](https://github.com/MHSanaei/3x-ui/commit/dc1979a1) (#6223) @n0ctal
- [ci: actually run the PostgreSQL schema and migration tests](https://github.com/MHSanaei/3x-ui/commit/2b1fe1fd) (#6224) @n0ctal
- [refactor(ci): replace the in-house review lanes with the official code-review skill](https://github.com/MHSanaei/3x-ui/commit/58669f61)
- [docs(api): document WireGuard and mtproto secret generation on clients/add](https://github.com/MHSanaei/3x-ui/commit/af3e6c11) (#6282) @rokokol
- [chore(i18n): update tr-TR translations](https://github.com/MHSanaei/3x-ui/commit/f204997c) (#6288) @tarihcituranx
- [fix(i18n): localize Chinese Xray labels](https://github.com/MHSanaei/3x-ui/commit/69a82375) (#6202) @nrps9909

<h3>🐞 Bug fixed</h3>

- [fix(node): stop the node sync from deleting clients it never meant to](https://github.com/MHSanaei/3x-ui/commit/5bc81dfd)
- [fix(node): stop stale expiry sync from undoing client extensions](https://github.com/MHSanaei/3x-ui/commit/6f7a3052) (#6231) @mrchatam
- [fix(node): adopt a matching deployed inbound instead of recreating it](https://github.com/MHSanaei/3x-ui/commit/bb29b6af) (#6197) @n0ctal
- [fix(node): keep disabled inbounds the node snapshot cannot report](https://github.com/MHSanaei/3x-ui/commit/6a674c7f) (#6221) @n0ctal
- [fix(node): don't stamp InboundsAdoptedAt when the sync adopted nothing](https://github.com/MHSanaei/3x-ui/commit/a255ab7c) (#6284) @yzxcj797
- [fix(nodes): log the inbound the node snapshot removes centrally](https://github.com/MHSanaei/3x-ui/commit/4b0e9f9b) (#6219) @n0ctal
- [fix(nodes): validate every certificate in the node mTLS trust bundle](https://github.com/MHSanaei/3x-ui/commit/bab39393) (#6188) @n0ctal
- [fix(nodes): apply a rotated master mTLS certificate without restarting the panel](https://github.com/MHSanaei/3x-ui/commit/7ecd88b9) (#6194) @n0ctal
- [fix(nodes): persist the master mTLS credential atomically and stop silent reissue](https://github.com/MHSanaei/3x-ui/commit/60453bf5) (#6195) @n0ctal
- [fix(nodes): report probe heartbeat persistence failures](https://github.com/MHSanaei/3x-ui/commit/4a5f6771) (#6207) @n0ctal
- [fix(clients): push bulk client changes to nodes only after the commit lands](https://github.com/MHSanaei/3x-ui/commit/c5dec64d) (#6181) @n0ctal
- [fix(clients): stop a stale IP row from blocking a client edit](https://github.com/MHSanaei/3x-ui/commit/2a8c3bc0)
- [fix(clients): stop recomputing the summary badges from the client_stats snapshot](https://github.com/MHSanaei/3x-ui/commit/ecadfd0e) (#6169) @mrnickson-hue
- [fix(traffic): clear cross-panel rows only for clients actually renewed](https://github.com/MHSanaei/3x-ui/commit/326009e9) (#6263) @n0ctal
- [fix(traffic): apply maintenance side effects only after the commit lands](https://github.com/MHSanaei/3x-ui/commit/13960050) (#6200) @n0ctal
- [fix(job): expire stored client IPs of offline clients](https://github.com/MHSanaei/3x-ui/commit/103b0dfe)
- [fix(job): force-disconnect over-limit Hysteria2 clients](https://github.com/MHSanaei/3x-ui/commit/d175050f)
- [fix(database): keep IP limits when the fail2ban probe is inconclusive](https://github.com/MHSanaei/3x-ui/commit/17fea2f6) (#6176) @n0ctal
- [fix(ldap): stop auto-delete from wiping every client on an empty directory](https://github.com/MHSanaei/3x-ui/commit/f4b7b08e)
- [fix(inbounds): close the port check-and-claim race on the serial writer](https://github.com/MHSanaei/3x-ui/commit/81cfd857) (#6225) @n0ctal
- [fix(inbounds): reject Hysteria inbound updates with empty client auth](https://github.com/MHSanaei/3x-ui/commit/585f4ecd) (#6268) @mvanhorn
- [fix(inbounds): surface form validation errors](https://github.com/MHSanaei/3x-ui/commit/3fa88adb) (#6084) @narcotics0507
- [fix(reality): make the REALITY target check usable on a private network](https://github.com/MHSanaei/3x-ui/commit/708a69ac) (#6242) @shustovTE
- [fix(groups): report changed bulk moves without restarting xray](https://github.com/MHSanaei/3x-ui/commit/0496c23a) (#6199) @n0ctal
- [fix(panel): stop one poisoned DNS answer from blocking outbound tests](https://github.com/MHSanaei/3x-ui/commit/2d30ab3a)
- [fix(outbounds): propagate allocation query failures](https://github.com/MHSanaei/3x-ui/commit/b70c5abc) (#6208) @n0ctal
- [fix(outbound): import Hysteria2 salamander properly from standard obfs params](https://github.com/MHSanaei/3x-ui/commit/d05e44e4) (#6166) @xMasterX
- [fix(warp): preserve WARP Plus license key when changing IP](https://github.com/MHSanaei/3x-ui/commit/5d6d98d1) (#6218) @rqzbeh
- [fix(warp): surface update-clock persistence failures](https://github.com/MHSanaei/3x-ui/commit/64f4f074) (#6209) @n0ctal
- [fix(netsafe): classify IPv6 transition and CGNAT ranges as internal](https://github.com/MHSanaei/3x-ui/commit/b51f0976)
- [fix(sub): keep Hysteria2 mport on external-proxy links](https://github.com/MHSanaei/3x-ui/commit/7a595cb4)
- [fix(sub): forward tlsSettings.cipherSuites into the JSON subscription](https://github.com/MHSanaei/3x-ui/commit/d9b599b9)
- [fix(sub): serve a copy-only page when a subscription URL is opened in a browser](https://github.com/MHSanaei/3x-ui/commit/43bc9153) (#6183) @n0ctal
- [fix(sub): restore the subscription info page for browser visits](https://github.com/MHSanaei/3x-ui/commit/f22df49a)
- [fix(sub): keep copy page within mobile viewport](https://github.com/MHSanaei/3x-ui/commit/338822ab)
- [fix(sub): render the full remark once per subscription, not once per credential](https://github.com/MHSanaei/3x-ui/commit/3b190915) (#6198) @n0ctal
- [fix(sub): use a fullwidth percent in USAGE_PERCENTAGE](https://github.com/MHSanaei/3x-ui/commit/ad32144c) (#6174) @n0ctal
- [fix(migration): stop a half-applied startup migration from committing silently](https://github.com/MHSanaei/3x-ui/commit/b56b0872) (#6182) @n0ctal
- [fix(db): harden unrestricted freedom outbounds](https://github.com/MHSanaei/3x-ui/commit/3bb87e80) (#6184) @n0ctal
- [fix(security): require a 2FA code to replace the stored TOTP secret](https://github.com/MHSanaei/3x-ui/commit/c8a3a2d7)
- [fix(cli): stop -getApiToken accumulating admin tokens](https://github.com/MHSanaei/3x-ui/commit/34c248bb) (#6175) @n0ctal
- [fix(web): fallback to default secret when database setting is empty](https://github.com/MHSanaei/3x-ui/commit/0f14ce75) (#6189) @CaMeDoZa
- [fix(web): report unexpected HTTP serve failures](https://github.com/MHSanaei/3x-ui/commit/20b3f84f) (#6210) @n0ctal
- [fix(frontend): make the jalali expiry clear button actually clear](https://github.com/MHSanaei/3x-ui/commit/b53a5515)
- [fix(frontend): refresh subscription settings after save](https://github.com/MHSanaei/3x-ui/commit/b73ceae0) (#6287) @dawNotPoi
- [fix(frontend): restore responsive table height](https://github.com/MHSanaei/3x-ui/commit/acbf09e7)
- [fix(frontend): disable table virtualization](https://github.com/MHSanaei/3x-ui/commit/03950b12)
- [fix(frontend): isolate swagger deps from main vendor chunk](https://github.com/MHSanaei/3x-ui/commit/8a8da885)
- [fix(frontend): restore the two rolldown bindings npm dropped from the lockfile](https://github.com/MHSanaei/3x-ui/commit/ce63bf3e)
- [fix(frontend): keep the MSW worker in step with the lockfile](https://github.com/MHSanaei/3x-ui/commit/f75ea08a) (#6222) @n0ctal
- [fix(install): preserve custom bin/ files (e.g. hand-added geoip) across updates](https://github.com/MHSanaei/3x-ui/commit/9165ab67) (#6152) @Kuzz007
- [fix(panel): forward the panel's proxy to update.sh's own downloads](https://github.com/MHSanaei/3x-ui/commit/94084249) (#6259) @Kuzz007
- [fix(tgbot): split long messages at line boundaries](https://github.com/MHSanaei/3x-ui/commit/f13baa9a) (#6293) @sanmaxdev
- [fix: dead code, typo, and minor bugs in main.go, process.go and index.go](https://github.com/MHSanaei/3x-ui/commit/31c1eed5) (#6167) @isuru709
- [fix: follow-ups from the post-merge reviews of #6221, #6227, #6230 and #6239](https://github.com/MHSanaei/3x-ui/commit/3f1dd4bf) (#6250) @n0ctal

<h3> Reports </h3>

![Total Download](https://img.shields.io/github/downloads/mhsanaei/3x-ui/v3.7.0/total?label=Total-Download&color=success)


## New Contributors
* @fastnas2023 made their first contribution in https://github.com/MHSanaei/3x-ui/pull/5650
* @FZ1010 made their first contribution in https://github.com/MHSanaei/3x-ui/pull/5698
* @narcotics0507 made their first contribution in https://github.com/MHSanaei/3x-ui/pull/6084
* @kobihikri made their first contribution in https://github.com/MHSanaei/3x-ui/pull/6130
* @Kuzz007 made their first contribution in https://github.com/MHSanaei/3x-ui/pull/6152
* @xMasterX made their first contribution in https://github.com/MHSanaei/3x-ui/pull/6166
* @isuru709 made their first contribution in https://github.com/MHSanaei/3x-ui/pull/6167
* @yelloduxx made their first contribution in https://github.com/MHSanaei/3x-ui/pull/6168
* @CaMeDoZa made their first contribution in https://github.com/MHSanaei/3x-ui/pull/6189
* @korsun009 made their first contribution in https://github.com/MHSanaei/3x-ui/pull/6190
* @nrps9909 made their first contribution in https://github.com/MHSanaei/3x-ui/pull/6202
* @rlex made their first contribution in https://github.com/MHSanaei/3x-ui/pull/6216
* @hunmar made their first contribution in https://github.com/MHSanaei/3x-ui/pull/6217
* @mrchatam made their first contribution in https://github.com/MHSanaei/3x-ui/pull/6231
* @shustovTE made their first contribution in https://github.com/MHSanaei/3x-ui/pull/6242
* @DIMFLIX made their first contribution in https://github.com/MHSanaei/3x-ui/pull/6243
* @ZaneL1u made their first contribution in https://github.com/MHSanaei/3x-ui/pull/6271
* @Masterain98 made their first contribution in https://github.com/MHSanaei/3x-ui/pull/6272
* @rokokol made their first contribution in https://github.com/MHSanaei/3x-ui/pull/6282
* @yzxcj797 made their first contribution in https://github.com/MHSanaei/3x-ui/pull/6284
* @dawNotPoi made their first contribution in https://github.com/MHSanaei/3x-ui/pull/6287

**Full Changelog**: https://github.com/MHSanaei/3x-ui/compare/v3.6.0...v3.7.0

