# Naive exact-usage accounting — proof-gate implementation plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:test-driven-development for retained code. This Phase A is a feasibility/proof gate; do not implement or patch a production Caddy module until the gate result has been recorded and the fallback design has explicit Owner approval.

**Goal:** Prove whether a separate PVNaive Caddy handler can observe exact successful CONNECT payload bytes and trusted per-user identity on the pinned Naive `forward_proxy` path, then either authorize the separate-module architecture or stop and revise the design before any fork/wrapper implementation.

**Architecture:** This phase is deliberately non-production. It pins the exact upstream `forward_proxy` release commit, checks the real streaming ownership boundary, and adds a deterministic semantic proof showing why bytes *read from the client* are not equivalent to bytes *successfully written to the remote peer*. If the separate-handler boundary is insufficient, the phase ends with a reviewed design addendum for a minimal pinned forward-proxy patch/wrapper and encrypted recoverable Subscription-token material; no accounting capability is enabled and production Caddy remains untouched.

**Tech Stack:** Go 1.25, Bash, Git/GitHub, pinned `klzgrad/forwardproxy` commit `d62c80d3dd2c706b6b87579844d2397bddd18317`, existing GitHub Actions CI.

**Spec:** `docs/superpowers/specs/2026-08-29-naive-usage-accounting-design.md`

**Roadmap task:** `PVN-045` — Exact per-credential accounting feasibility PoC.

## Global constraints

- Product name remains `PVNaive`; repository name is not renamed.
- Current production deployment at schema 6 and Caddy v2.11.2 is not mutated by this plan.
- No package/OS upgrades.
- No access-log/process-wide estimate may be accepted as exact accounting.
- Exact accounting counts only payload bytes successfully transferred through an authenticated CONNECT tunnel.
- Failed authentication and failed CONNECT must count zero.
- Historical traffic for existing credentials is never invented.
- Do not copy/modify upstream `forward_proxy` production code until this proof gate has completed and any required fallback design has explicit Owner approval.
- The current `usage_capability.available=false` behavior remains correct throughout Phase A.
- Persistent Subscription/QR must not require plaintext raw tokens at rest; if later implemented, recoverable token material must be authenticated-encrypted and Owner-only to retrieve.

---

### Task 1: Pin and prove the real `forward_proxy` streaming ownership boundary

**Files:**
- Create: `tests/accounting/pinned_forwardproxy_boundary_test.sh`
- Modify: `.github/workflows/ci.yml`

**Evidence contract:**
- Upstream ref MUST resolve to `d62c80d3dd2c706b6b87579844d2397bddd18317`.
- The pinned source MUST show that `forward_proxy` itself creates/owns `targetConn` after authentication.
- The pinned HTTP/1 CONNECT path MUST hand that private connection to `serveHijack`.
- The pinned HTTP/2/3 CONNECT path MUST hand that private connection plus `r.Body`/`ResponseWriter` to `dualStream`.
- A preceding Caddy handler MUST therefore have no reference to the remote `targetConn` and cannot directly observe successful client→remote writes.

- [ ] Create `tests/accounting/pinned_forwardproxy_boundary_test.sh` with fixed constants:

```bash
UPSTREAM_REPO='https://github.com/klzgrad/forwardproxy.git'
UPSTREAM_COMMIT='d62c80d3dd2c706b6b87579844d2397bddd18317'
```

The script creates a temporary Git repository, fetches exactly that commit with depth 1, rejects any SHA mismatch, writes `forwardproxy.go` from `FETCH_HEAD`, and proves the expected CONNECT ownership tokens are present. It must not use `master`, a moving tag, or `latest`.

- [ ] Required checks in the script:

```bash
grep -Fq 'targetConn, err := h.dialContextCheckACL' forwardproxy.go
grep -Fq 'return serveHijack(w, targetConn)' forwardproxy.go
grep -Fq 'return dualStream(targetConn, r.Body, w)' forwardproxy.go
```

Also prove authentication occurs before the CONNECT target dial by comparing source line numbers for `h.checkCredentials(r)` and `h.dialContextCheckACL(ctx, "tcp", hostPort)`.

- [ ] Run locally/CI:

```bash
bash tests/accounting/pinned_forwardproxy_boundary_test.sh
```

Expected terminal marker:

```text
PINNED_FORWARDPROXY_BOUNDARY_PROOF=PASSED
```

- [ ] Wire the proof into the existing `rehearsal` job before any custom-Caddy build work. Do not alter the existing pinned-Caddy multiple-`basic_auth` proof.

- [ ] Commit checkpoint:

```text
test(accounting): pin forwardproxy CONNECT ownership boundary
```

---

### Task 2: Prove attempted/read bytes are not exact successful remote-write bytes

**Files:**
- Create: `internal/accountingboundary/boundary_test.go`
- No retained production package implementation is allowed in this task.

**Proof intent:** A separate outer handler may count bytes consumed from `r.Body`, but the inner `forward_proxy` can experience a short/partial write to `targetConn`. Therefore outer-body read counts can exceed successfully transferred upload bytes. Exact billing requires instrumentation at the stream write boundary owned by `forward_proxy`.

- [ ] Write a deterministic Go test with:
  - a client payload larger than one write;
  - an outer `io.Reader` wrapper that records every byte returned by `Read`;
  - an inner remote writer that intentionally accepts only part of the supplied buffer and returns a short write/error;
  - assertions that outer observed bytes are greater than the successful remote-write count.

The core assertion must be equivalent to:

```go
if outerReadBytes == successfulRemoteWriteBytes {
    t.Fatal("fixture failed to demonstrate the accounting boundary")
}
```

and must independently assert the simulated successful remote-write count exactly, so the test cannot pass because both counters are broken.

- [ ] Add a second case proving download symmetry: an outer response wrapper can only count bytes accepted by its own `Write`; it cannot recover client→remote success, so one outer handler cannot provide both exact directions unless it owns/wraps the forwarding primitive.

- [ ] Run:

```bash
go test ./internal/accountingboundary -run 'TestSeparateHandlerBoundary' -count=1
```

Expected: PASS. This is a feasibility proof, not evidence that accounting is implemented.

- [ ] Run the normal Go gate:

```bash
gofmt -w internal/accountingboundary/boundary_test.go
test -z "$(gofmt -l .)"
go vet ./...
go test ./...
```

- [ ] Commit checkpoint:

```text
test(accounting): prove separate handler cannot count exact upload writes
```

---

### Task 3: Record the proof result and revise the architecture before any fork/wrapper code

**Files:**
- Modify: `docs/superpowers/specs/2026-08-29-naive-usage-accounting-design.md`
- Modify: `docs/superpowers/specs/2026-08-29-naive-usage-accounting-design-review.md`
- Create: `docs/superpowers/specs/2026-08-29-naive-usage-accounting-forwardproxy-addendum.md`
- Modify: `WORKLOG.md`
- Modify: `AGENT_TASKS.md`

**Decision rule:**
- If Tasks 1–2 unexpectedly prove a separate module can observe both trusted identity and exact successful writes in both directions, document the exact mechanism and proceed only after review.
- If Tasks 1–2 prove the expected limitation, `PVN-045` remains `IN_PROGRESS`; do **not** add production accounting code yet. Produce the fallback addendum below and stop for explicit Owner approval.

- [ ] The fallback addendum must pin this minimal supply-chain model:
  - Caddy `v2.11.2`.
  - upstream `klzgrad/forwardproxy` commit `d62c80d3dd2c706b6b87579844d2397bddd18317`.
  - keep the existing `forward_proxy` wire behavior and Caddyfile directive.
  - store the PVNaive change as a small reviewable patch/fork against that exact commit, never a moving branch.
  - the modified forwarding primitive owns successful-write accounting and uses a maximum forwarding/accounting chunk of 32 KiB for the quota overshoot bound.
  - authenticated username is the trusted identity from the existing `basic_auth` match; Runtime Agent resolves globally unique Runtime username → stable Runtime credential UUID server-side. Do not trust a client-supplied UUID/header.
  - before opening the target connection, the patched handler asks Runtime Agent `/v1/accounting/authorize`.
  - successful upload/download deltas are sent over `/run/pvnaive/runtime-agent.sock` with connection UUID + monotonic sequence.
  - finite-quota policy failure is fail-closed.
  - quota depletion closes the tunnel and preserves username/password/Runtime UUID.

- [ ] The addendum must also resolve the persistent Subscription/QR security conflict already present in schema 6:
  - current `direct_subscription_tokens` stores only SHA-256 digest + non-secret prefix; raw token is one-time and unrecoverable;
  - persistent Owner-only “Subscription” / “QR” therefore requires the raw 256-bit token to be stored only as AES-256-GCM ciphertext + 12-byte nonce + key-id using the existing runtime encryption key;
  - plaintext raw token must never be stored/logged;
  - public resolution continues to use the hash, so wire/public behavior is unchanged;
  - normal view/copy/QR retrieval decrypts for the authenticated Owner and does not rotate the token;
  - explicit Rotate remains the only operation that revokes the old token and creates a new one.

- [ ] The addendum must define the next implementation phase boundaries:
  1. schema 7 usage ledger + encrypted recoverable subscription token;
  2. Runtime Agent DB-backed accounting authorize/delta/health endpoints over AF_UNIX only;
  3. pinned forward-proxy accounting patch + deterministic HTTP/1/H2 traffic proof;
  4. Customer API used/remaining/percent + stable Subscription retrieval;
  5. Sanaei-style UI with total/used/remaining/progress/expiry/Sub/QR/edit;
  6. new **S06 accounting** bundle/preflight/upgrade scripts that may perform one controlled Caddy restart and exact binary rollback, while preserving the old S05 “never restart Caddy” contract;
  7. canary production proof before setting `usage_capability.available=true`.

- [ ] Update `WORKLOG.md` with the actual proof result and commit/run evidence. Do not mark `PVN-045` DONE unless exact measured traffic accuracy is later demonstrated.

- [ ] Update `AGENT_TASKS.md` to assign the next accounting work to `PVN-045..049` and note the exact file-conflict boundary.

- [ ] Run documentation/safety scans:

```bash
grep -RInE 'TODO|TBD|FIXME|PLACEHOLDER|<[^>]+>' \
  docs/superpowers/specs/2026-08-29-naive-usage-accounting-* \
  docs/superpowers/plans/2026-08-29-naive-usage-accounting-proof-gate.md || true

git diff --check
```

Any match must be reviewed and either removed or proven to be literal example syntax rather than an unresolved placeholder.

- [ ] Commit checkpoint:

```text
docs(accounting): record forwardproxy proof and fallback architecture
```

- [ ] **HARD STOP:** Present the proof evidence and fallback addendum to the Owner. Do not start schema 7, Runtime Agent accounting endpoints, custom Caddy code, or production deployment until the revised fallback architecture is explicitly approved.

---

## Phase A verification gate

Before declaring this proof phase complete, run:

```bash
bash tests/accounting/pinned_forwardproxy_boundary_test.sh
go test ./internal/accountingboundary -count=1
go vet ./...
go test ./...
cd web && npm test && npm run build
```

Then confirm the branch CI is green for all existing jobs plus the new boundary proof.

**Required final report:**

```text
AGENT: Lead/Agent-ARCH
TASK-ID: PVN-045
STATUS: REVIEW
CHANGES: pinned upstream boundary proof + semantic successful-write proof + fallback addendum
FILES MODIFIED: exact paths
TESTS: exact commands + CI run ID
RESULT: separate-handler sufficient OR insufficient, with evidence
NEXT STEP: Owner approval of fallback architecture OR separate-module implementation plan
BLOCKERS: none / explicit proof result
```

## What is intentionally not implemented in this plan

This Phase A plan does **not** create schema 7, does not enable accounting, does not alter Runtime Agent policy behavior, does not build/replace Caddy, does not change customer quota enforcement, and does not modify production. Those changes require the post-proof architecture to be approved and will receive a separate detailed TDD implementation plan.