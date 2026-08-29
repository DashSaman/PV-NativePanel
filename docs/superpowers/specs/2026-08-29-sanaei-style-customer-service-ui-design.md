# PVNaive Sanaei-style Customer Service UI Design

Date: 2026-08-29
Branch: `s05-user-quota-design`
Status: approved behavior captured from chat; pending final user review before implementation

## Goal

Make PVNaive customer/service creation as clean and practical as the 3x-ui/Sanaei client workflow while preserving PVNaive's safer business/runtime separation.

An Owner must be able to create a customer service in one organized form and define:

- username;
- generated or custom password;
- traffic quota by entering a numeric GB value, or unlimited;
- service validity by entering a numeric day count;
- whether the numeric validity starts immediately or from the first successful authenticated connection;
- an exact manual expiry date/time instead of duration-based expiry;
- resulting Naive share link;
- resulting subscription link;
- locally generated QR code.

The existing Runtime credential table is not the long-term customer UI. Business service fields live in the customer/service layer and bind to a stable Runtime credential UUID.

## UX model

### Create customer/service form

Use one compact, ordered card similar in density to Sanaei rather than exposing separate low-level Runtime concepts to the normal Owner workflow.

Fields in order:

1. `نام کاربری`
2. `رمز عبور`
   - default: secure server-generated password
   - optional custom password
3. `حجم`
   - numeric input in GB
   - `نامحدود` toggle
4. `نوع اعتبار`
   - `از زمان ساخت`
   - `از اولین اتصال`
   - `تاریخ انقضای دستی`
5. Depending on validity mode:
   - creation/first-connection mode: numeric `مدت اعتبار` in days
   - manual mode: exact date/time picker
6. submit: `ساخت و اعمال اکانت`

Do not show multiple contradictory expiry inputs at once. The selected validity mode determines which field is authoritative.

## Expiry semantics

Exactly three mutually exclusive validity modes are supported.

### Mode A — duration from creation

Example:

```text
30 days
start_policy = on_creation
```

On successful creation:

```text
starts_at = creation/term activation timestamp
expires_at = starts_at + 30 days
```

### Mode B — duration from first connection

This is the recommended default.

Example:

```text
30 days
start_policy = on_first_successful_connection
```

Before first successful authenticated connection:

```text
starts_at = NULL
first_connected_at = NULL
expires_at = NULL
state = pending
```

After the first proven successful authenticated connection:

```text
first_connected_at = observed timestamp
starts_at = first_connected_at
expires_at = starts_at + 30 days
state = active
```

A mere subscription fetch, panel view, config creation, Runtime reload, or failed authentication must not activate the service.

### Mode C — exact manual expiry

Example:

```text
expires_at = 2026-10-15 23:59:00 +03:30
start_policy = fixed_timestamp/manual-expiry semantics
```

The Owner chooses an exact date/time. This value is authoritative and is not recomputed from a day count.

The UI must present the date in the user's locale while the API stores and exchanges an unambiguous timezone-aware timestamp.

## Precedence and validation

- Only one validity mode is authoritative for a service term.
- Numeric duration must be a positive integer number of days.
- Manual expiry must be in the future at creation time.
- Unlimited traffic is represented by `quota_bytes = NULL`; zero is not silently treated as unlimited.
- Numeric GB is converted server-side to bytes using one documented unit policy; the UI must display that same policy consistently.
- Switching validity mode in the form clears incompatible hidden values before submission.
- Editing an active term requires explicit audited renewal/adjustment semantics; changing a plan template must not rewrite historical terms.

## Customer table

Replace the current Runtime-only management experience for normal customer operations with a Sanaei-style customer table.

Columns:

- username;
- status;
- traffic quota;
- used traffic;
- remaining traffic;
- expiry;
- remaining time;
- start mode;
- subscription/QR;
- actions.

Until exact accounting passes PVN-045..047:

- quota may be configured and displayed;
- usage/remaining traffic must show `در دسترس نیست — حسابداری دقیق هنوز تأیید نشده`;
- no fake zero usage, estimated usage, or quota enforcement may be presented as exact.

Once exact accounting is proven, the same columns become live without redesigning the UI.

## Customer info / delivery modal

Immediately after creating a service, show one organized delivery modal containing:

- username;
- one-time password if newly generated;
- Naive URI;
- subscription URL;
- QR code for the subscription URL and/or share URI as applicable;
- quota;
- expiry mode;
- numeric days or exact expiry date;
- copy buttons.

QR generation must be local. Secret-bearing URI/token data must never be sent to a third-party QR API.

For later views, do not reveal a stored plaintext password. Rotation generates a new one-time delivery result.

## Subscription behavior

Each service gets a revocable opaque subscription token separate from the Runtime password.

Expected surface:

```text
https://<subscription-host>/sub/<opaque-token>
```

The subscription response renders the active Naive configuration for that service. Token rotation/revoke must not require changing the business user identity.

The customer detail page may expose:

- copy subscription URL;
- show QR;
- rotate subscription token;
- copy Naive link only when the required secret is available through a safe one-time delivery flow.

## Runtime enforcement

Expiry and, after accounting proof, quota depletion are enforced through the existing desired-state -> Runtime Agent -> validate -> backup -> reload-only -> verify/rollback pipeline.

The browser is never in the data-plane availability path.

For `on_first_successful_connection`, the first-connection transition must come from a proven authenticated Runtime signal and must be idempotent so concurrent first streams cannot activate the term twice.

## API shape

Creation should accept an explicit validity object rather than ambiguous independent fields.

Examples:

```json
{
  "username": "customer1",
  "quota_gb": 50,
  "validity": {
    "mode": "on_first_successful_connection",
    "duration_days": 30
  }
}
```

```json
{
  "username": "customer2",
  "quota_gb": null,
  "validity": {
    "mode": "fixed_expiry",
    "expires_at": "2026-10-15T23:59:00+03:30"
  }
}
```

The API normalizes these into the existing `service_terms` model (`quota_bytes`, `duration_seconds`, `start_policy`, `starts_at`, `first_connected_at`, `expires_at`).

## Testing requirements

Automated tests must cover at minimum:

- numeric GB -> bytes conversion;
- unlimited quota;
- positive numeric duration validation;
- duration from creation;
- duration from first successful connection;
- first connection activation idempotency;
- failed authentication does not activate;
- subscription fetch does not activate;
- exact manual expiry with timezone;
- past manual expiry rejection;
- switching validity modes does not submit stale hidden values;
- QR is generated locally;
- response/log/audit never leaks plaintext Runtime password or subscription token;
- usage remains capability-unavailable until exact accounting proof passes.

## Relationship to existing S05 design

This document refines the UI and expiry-entry behavior of:

`docs/superpowers/specs/2026-08-28-customer-quota-accounting-design.md`

It does not replace the existing accounting, ledger, Runtime Agent, or customer lifecycle architecture. In particular:

- business user != Runtime credential;
- exact accounting is still required before billable usage/enforcement claims;
- `on_first_successful_connection` remains the recommended default;
- manual exact expiry is a first-class supported mode;
- subscription and QR remain part of PVN-052..PVN-055.
