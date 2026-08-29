# WS3 — Subscription / Account Page / Client Delivery

Status: implementation under verification

## Starting point

- Starting `main` SHA: `28639c6fc93e19f98faab623607c32ca8a1b8436`
- Branch: `parallel/ws3-subscription-client`
- PR: #11
- No pre-existing `parallel/ws3-subscription-client` or subscription-split branch was found.
- Valid S06 foundations were retained: opaque 32-byte subscription tokens, hashed lookup, encrypted Owner token recovery, local QR generator, read-only current-subscription retrieval, explicit subscription reissue, and separate password rotation.

## Old behavior

The public legacy endpoint `/api/v1/subscriptions/<token>` selected either machine output or an HTML page from the request `Accept` header. A normal browser could therefore change the semantic response of the same URL. The account page and machine subscription were not separate contracts.

## Target/final contract being verified

- `/sub/<opaque-token>`: machine-only Naive subscription response, independent of `Accept`.
- `/s/<opaque-token>`: human HTML account page, independent of `Accept`.
- `/api/v1/subscriptions/<opaque-token>`: legacy compatibility endpoint, machine-only.
- Owner delivery material uses `subscription_path=/sub/...` and `account_page_path=/s/...`.
- Owner current-subscription and reissue responses retain `direct_uri` when it can be resolved safely.

## Security invariants

- public delivery responses are `Cache-Control: no-store` and `Pragma: no-cache`;
- `X-Robots-Tag` blocks indexing/archive/snippets;
- `X-Content-Type-Options: nosniff`, `X-Frame-Options: DENY`, `Referrer-Policy: no-referrer`;
- account-page CSP is nonce-based and denies all sources by default, frames, forms and network connections while permitting only local/data QR images and the nonce-bound inline style/script;
- request `Host` never determines the canonical subscription host;
- invalid/revoked token resolution returns a generic 404 without decrypt/token oracle details;
- no external QR service receives a subscription URL, token or Direct Naive URI.

## Mutation safety

Read/view/copy behavior stays on `ResolveProfile`/Owner token recovery reads. These paths contain no token rotation, password rotation, quota/expiry write, first-use activation or Runtime mutation. Reissue continues to revoke old subscription tokens and create a new opaque token without rotating the Runtime password. Password rotation remains an explicit separate action and does not silently reissue the subscription.

## QR behavior

The Account Page renders two locally generated QR codes when Direct Naive is available:

1. Subscription QR for canonical `/sub/<token>` URL.
2. Direct Naive QR for the resolved `naive+https://...` URI.

For inactive/suspended/expired accounts the account page remains status-readable, while Direct Naive delivery/QR is unavailable.

## Account Page / i18n / template boundary

A dedicated `subscription_page.go` renderer provides the public presentation boundary instead of embedding human rendering decisions into machine content negotiation. It includes responsive mobile layouts, `prefers-color-scheme` light/dark behavior, PVNaive/PVNETWORK branding, Persian RTL and English LTR foundations. When accounting is not available, it explicitly renders `Usage unavailable` / `در دسترس نیست` and does not fabricate `0 GB used`. Online and Last Online are also shown as unavailable until a proved capability is wired.

Opening the page resolves state only; it does not start first-use.

## Compatibility evidence

Real client PASS is deliberately not inferred from protocol/schema support. Karing official documentation supports adding profile URLs/content, clipboard/file import and QR scanning of profile links; current Karing release evidence is tracked separately during final verification. The repository also contains an earlier draft Karing-profile PR whose own exit criterion says a real Karing client smoke was still required.

No client will be marked PASS here without a real client run.

## Tests / CI

TDD RED evidence:

- commit `c35e4046797e0aaad5ec307e004fec94f73dea61`
- formatting: PASS
- `go vet ./...`: PASS
- `go test ./...`: FAIL as expected before implementation because the explicit `/sub` and `/s` contract was not implemented yet.
- web and database jobs: PASS on the RED run.

GREEN verification is pending on the implementation head.

## WS1 dependency

Exact accounting / online presence is owned by WS1. Until that capability is integrated, the page intentionally displays unavailable state rather than synthetic usage or presence.

## Remaining / blockers

- complete GREEN CI on the implementation head;
- run/record real Karing client compatibility. This environment currently has protocol/documentation evidence but no completed real Karing application smoke, so Karing must not be labeled PASS yet;
- record final compatibility matrix and exact tested versions;
- update this report with final head/files/CI and integration decision.

## Exact next step

Run CI against the implementation head, fix any code/test failure, then perform the mandatory real Karing smoke (Subscription URL, Direct Naive, QR, update/import) before declaring client compatibility PASS.

READY_FOR_INTEGRATION: NO
