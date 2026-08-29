# WS3 — Subscription / Account Page / Client Delivery

Status: code/CI verified; real-client acceptance still blocked

## Starting point

- Starting `main` SHA: `28639c6fc93e19f98faab623607c32ca8a1b8436`
- Branch: `parallel/ws3-subscription-client`
- Verified implementation head: `8ecd31f161a90df5dc07d00ab1355aab3395c045`
- PR: #11 — `WS3: deterministic subscription and account delivery`
- No pre-existing `parallel/ws3-subscription-client` or subscription-split branch was found.
- Valid S06 foundations were retained rather than rewritten: opaque 32-byte subscription tokens, SHA-256 token lookup, encrypted Owner token recovery, local QR generation, read-only current-subscription retrieval, explicit subscription reissue, and separate password rotation.

## Old behavior

The public legacy endpoint `/api/v1/subscriptions/<token>` selected either machine output or an HTML page from the request `Accept` header. A normal browser could therefore change the semantic response of the same URL. The account page and machine subscription were not separate contracts.

## Final endpoint contract

- `/sub/<opaque-token>`: machine-only Naive subscription response, independent of `Accept`.
- `/s/<opaque-token>`: human HTML account page, independent of `Accept`.
- `/api/v1/subscriptions/<opaque-token>`: legacy compatibility endpoint, machine-only.
- Owner create/adopt/current/reissue delivery material now returns `subscription_path=/sub/...` and `account_page_path=/s/...`.
- Owner current-subscription and reissue responses retain `direct_uri` when it can be resolved safely.
- Request `Host` is not trusted to construct canonical subscription links; the configured subscription proxy host is used.

## Security invariants

- public delivery responses are `Cache-Control: no-store` and `Pragma: no-cache`;
- `X-Robots-Tag` blocks indexing/archive/snippets;
- `X-Content-Type-Options: nosniff`, `X-Frame-Options: DENY`, `Referrer-Policy: no-referrer`;
- `Permissions-Policy` disables camera, microphone and geolocation;
- account-page CSP is nonce-based and denies all sources by default, frames, forms and network connections while permitting only local/data QR images and nonce-bound inline style/script;
- invalid or non-resolvable token lookups return a generic 404 without exposing decrypt/token-oracle/debug information;
- no external QR service receives a subscription URL, opaque token or Direct Naive URI.

## Mutation safety

Read/view/copy behavior stays on `ResolveProfile` or encrypted Owner-token recovery reads. Those paths do not rotate the subscription token, rotate the Runtime password, change quota/expiry, activate first-use, or mutate Runtime credentials.

Reissue remains a separate explicit Owner mutation. Existing S06 semantics revoke the old active subscription token and create a new opaque token without rotating the Runtime password. Password rotation remains a separate explicit action and does not silently reissue the subscription.

The Account Page itself performs only subscription resolution and presentation; it does not call first-use activation.

## QR behavior

The Account Page renders locally generated QR data URIs:

1. Subscription QR for the canonical `/sub/<token>` URL.
2. Direct Naive QR for the resolved `naive+https://...` URI when the service is available.

No QR network call is used. For inactive/suspended/expired service state the page can remain status-readable while Direct Naive delivery/QR is unavailable.

## Account Page

A dedicated `subscription_page.go` renderer is the presentation/template boundary instead of embedding human rendering decisions in machine content negotiation.

It exposes:

- Username
- Status
- Total quota
- Used
- Remaining
- Expiry
- Remaining days
- Start policy
- Online
- Last online
- Subscription link
- Direct Naive when available
- Subscription QR
- Direct Naive QR

When exact accounting is not available, it renders `Usage unavailable` / `در دسترس نیست`; it never synthesizes `0 GB used`. Online and Last Online are also rendered as unavailable until a proved capability is wired.

The layout includes responsive breakpoints for tablet/phone widths and `prefers-color-scheme` light/dark behavior. A full device-browser visual regression run was not available in this execution environment, so responsive CSS is implemented and unit-protected indirectly but no mobile-browser PASS is claimed.

## i18n / branding / templates

- English: LTR foundation.
- Persian: RTL foundation.
- Branding: PVNaive / PVNETWORK.
- Translation strings are separated from account-page data/rendering, so further languages can be added without changing machine subscription behavior.
- Account-page presentation is separated from `subscription.go`, providing the requested theme/template boundary.
- No dynamic public announcement source was added in this lane because there is not yet a proved public-vs-admin announcement configuration boundary; inventing one inside a secret-bearing public handler would weaken the requested data-separation invariant. This remains an optional future extension.

## Compatibility matrix

No row is marked PASS without a real application/client execution.

| Client | Version | OS | Import Method | Endpoint | Result | Notes |
|---|---|---|---|---|---|---|
| Karing | `v1.2.23.2606` latest checked 2026-08-29 | Not executed | Subscription profile link / clipboard / QR documented | `/sub/<token>` and Direct Naive candidate | **NOT RUN — BLOCKER** | Official Karing docs document subscription links, clipboard/file and QR profile-link import. A real Karing app smoke against this WS3 endpoint was not possible here, so import/update/connect behavior is not called PASS. |
| NaiveProxy native client | Upstream stable workflow documented; binary not executed | Not executed | `config.json` / `--proxy` | Direct credentials | **DOCUMENTED, NOT RUN** | Upstream native client documents `https://user:pass@host` as the proxy value; PVNaive's `naive+https://...` delivery URI therefore needs client-side URI handling/translation rather than being assumed identical to native config syntax. |
| sing-box | Naive outbound available since `1.13.0` | Not executed | Naive outbound JSON | Direct credentials | **SCHEMA SUPPORTED, NOT RUN** | Official sing-box docs expose Naive `server`, `server_port`, `username`, `password` and TLS fields. This proves schema capability, not live connectivity. |
| NekoBox | Not verified | Not verified | Not verified | Not verified | **NOT VERIFIED** | No evidence strong enough for PASS was produced in this run. |
| V2Box | Not verified | Not verified | Not verified | Not verified | **NOT VERIFIED** | No evidence strong enough for PASS was produced in this run. |

### Compatibility evidence checked

- Karing Add Profiles: https://karing.app/en/app-manual/add-profiles
- Karing latest release API/tag checked on 2026-08-29: `v1.2.23.2606`, published 2026-08-05.
- sing-box Naive outbound: https://sing-box.sagernet.org/configuration/outbound/naive/
- NaiveProxy upstream README client setup: https://github.com/klzgrad/naiveproxy/blob/master/README.md

### Tested client versions

No real GUI/native client was executed in this workstream environment. The table deliberately distinguishes documentation/schema evidence from runtime evidence. This is the reason WS3 cannot satisfy the prompt's mandatory Karing acceptance gate yet.

## Files changed

- `internal/httpapi/routes.go`
- `internal/httpapi/subscription.go`
- `internal/httpapi/subscription_page.go`
- `internal/httpapi/subscription_contract_test.go`
- `internal/httpapi/subscription_page_test.go`
- `internal/httpapi/customer.go`
- `internal/httpapi/customer_management.go`
- `internal/httpapi/customer_adopt_update.go`
- `internal/httpapi/customer_create_test.go`
- `internal/httpapi/customer_adopt_update_test.go`
- `docs/agent-reports/WS3_SUBSCRIPTION_CLIENT.md`

No telemetry, forwardproxy accounting, quota internals, Plan/Bulk/RBAC, installer or release implementation was changed.

## Tests / CI

### TDD RED proof

Commit `c35e4046797e0aaad5ec307e004fec94f73dea61`:

- formatting: PASS
- `go vet ./...`: PASS
- `go test ./...`: FAIL as expected before implementation because explicit `/sub` and `/s` behavior was not implemented yet
- web and database jobs: PASS

### GREEN proof

Verified implementation head: `8ecd31f161a90df5dc07d00ab1355aab3395c045`

GitHub Actions CI run `33266320835` / run #757 completed **SUCCESS** on 2026-08-29.

Evidence from that run:

- Go formatting gate: PASS
- `go vet ./...`: PASS
- `go test ./...`: PASS
- Runtime agent safety rehearsal: PASS
- Web `npm test`: PASS
- Web `npm run build`: PASS
- PostgreSQL 18 migration/health/backup/restore gate: PASS
- pinned forwardproxy boundary rehearsal: PASS
- pinned Naive Caddy multi-basic-auth rehearsal: PASS
- end-to-end auth/runtime rehearsal: PASS
- final CI bundle job: PASS

The new contract tests directly cover `/sub` raw behavior under browser `Accept`, legacy machine-only behavior, `/s` HTML behavior under non-HTML `Accept`, security headers/CSP, Persian RTL pending-first-use presentation, suspended-account presentation, local QR presence, configured canonical host, and honest accounting-unavailable state. Existing S06 tests continue to protect token reissue/password-separation semantics.

Explicit live-client/mobile-browser acceptance requested by this workstream is not silently inferred from CI and remains recorded below.

## WS1 dependency

Exact accounting / online presence is owned by WS1. Until that capability is integrated, the page intentionally displays unavailable state rather than synthetic usage or presence. WS3 does not modify accounting or first-use internals.

## Remaining / blockers

1. Mandatory real Karing smoke remains unexecuted: Subscription URL, Direct Naive, Subscription QR, Direct QR, Update Subscription, and import/connect behavior must be exercised in a real Karing application against a safe non-production test endpoint.
2. A real mobile-browser visual pass remains desirable for the Account Page; responsive CSS exists, but no device/browser PASS is claimed.
3. If a public announcement feature is wanted for R1, define a public-safe announcement source/config contract before wiring it into this secret-bearing page.

No production server was used for raw experimentation.

## Exact next step

On a disposable/safe test deployment of PR #11, run Karing `v1.2.23.2606` on at least one supported desktop/mobile OS through all six required flows (subscription link import, Direct Naive import, both QR paths, subscription refresh/update, and actual connection). Record exact OS/version/result here. If those pass without mutation or secret leakage, change `READY_FOR_INTEGRATION` to YES and move PR #11 out of draft.

READY_FOR_INTEGRATION: NO
