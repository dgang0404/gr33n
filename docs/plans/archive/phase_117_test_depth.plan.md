---
name: Phase 117 — Test depth (zero-coverage packages, handler gaps, timeout hygiene)
overview: >
  The July 2026 audit found ~17 internal packages with zero tests (including
  security-relevant farmauthz, authctx, auditlog, pushnotify), most handler packages
  untested (auth, cost, farm, guardian, alert…), and test-infra hygiene issues:
  a 666-second default LLM timeout, smoke helpers on http.DefaultClient with no
  timeout, and E2E budgets that fail on CPU-only Ollama boxes. No browser E2E exists.
todos:
  - id: ws1-security-pkgs
    content: "WS1: Security-relevant units first — farmauthz (role matrix), authctx, auditlog, pushnotify dispatch (fake FCM), httputil error paths"
    status: completed
  - id: ws2-handler-smokes
    content: "WS2: Handler coverage — smoke tests for auth (register/login/password modes), cost, farm settings, guardian proposals list/dismiss, alert ack, fileattach limits, organization RBAC"
    status: completed
  - id: ws3-timeout-hygiene
    content: "WS3: Timeout hygiene — llm.DefaultTimeout 666s → 120s (env-tunable, doc'd); smoke helpers use shared http.Client with sane timeout; CPU-Ollama guidance (-timeout, LLM_TIMEOUT_SECONDS, LLM_MAX_TOKENS) in INSTALL.md ollama-smoke section"
    status: completed
  - id: ws4-ui-components
    content: "WS4: UI component tests — Pi wizard steps, GuardianModelSelector, ActuatorCard (post-114 fix), workspace nav gating; target the untested-SFC list from the audit"
    status: completed
  - id: ws5-browser-e2e
    content: "WS5: Browser E2E seed — Playwright with 3 journeys: login → dashboard; create task offline → sync; Guardian chat → proposal Confirm; wired as manual CI lane like hardware/ollama smokes"
    status: completed
  - id: ws6-worker-pkgs
    content: "WS6: Remaining units — costing, farmbootstrap, catalognotify, insertcommonsreceiver, fertigation/programfit, rag/embed (fixture-based)"
    status: completed
isProject: false
---

# Phase 117 — Test depth (zero-coverage packages, handler gaps, timeout hygiene)

## Status

**Shipped** (2026-07-03). Security-relevant unit tests, handler gap smokes,
timeout hygiene, UI component tests, Playwright E2E scaffolding, and worker-package
units landed in Phase 117.

---

## Findings driving this phase

### Packages with zero `*_test.go`

`authctx`, `auditlog`, `farmauthz`, `farmbootstrap`, `catalognotify`, `pushnotify`,
`httputil`, `pgxutil`, `plantcatalog`, `fileattachutil`, `costing`,
`insertcommonsreceiver`, `rag/embed`, `platform/commontypes`,
`fertigation/programfit` (+ partial: `platform/bootstraptemplates`,
`platform/devicetaxonomy`, `fertigation/mixplan`).

`farmauthz` is the RBAC gate for every farm-scoped write — it having no tests is the
single biggest coverage risk in the repo.

### Handler layer

Tested: chat, rag, automation, actuator, sensor, device, lighting, zone.
Untested: **auth**, cost, fileattach, farm, profile, guardian, alert, audit,
organization, naturalfarming, aquaponics, animal.

### Infra hygiene

| Issue | Evidence |
|-------|----------|
| `llm.DefaultTimeout = 666 * time.Second` — a joke value in prod default | `internal/rag/llm/chat.go:19` |
| Smoke helpers use `http.DefaultClient` (no timeout) — a hung server hangs the whole suite until go-test kills it with a useless goroutine dump | `cmd/api/smoke_helpers_test.go` |
| First Phase 112 E2E run on this machine died at the 10-min go-test default because CPU-only tinyllama + full grounding prompt exceeds it; rerun with `-timeout 40m LLM_MAX_TOKENS=60` passed all 6 | terminal log, 2026-07-03 |
| No browser E2E at all (no Playwright/Cypress) | — |

### Already good

Costs smoke coverage, automation worker (15 smokes + unit), interlocks, UI logic/parity
tests (135 files), openapi parity test, hardware + ollama CI lanes.

---

## Design notes

### WS3 — Timeout hygiene

`DefaultTimeout` becomes 120 s; `LLM_TIMEOUT_SECONDS` still overrides. Document in the
INSTALL.md ollama-smoke section that CPU-only boxes should run E2E with
`-timeout 40m LLM_TIMEOUT_SECONDS=150 LLM_MAX_TOKENS=60` (verified working on this
hardware). Smoke helpers get one shared `http.Client{Timeout: 60s}` so a wedged
endpoint fails one test with a clear error instead of wedging the suite.

### WS5 — Browser E2E

Deliberately tiny: three journeys, Chromium only, seeded demo DB, manual
`workflow_dispatch` lane (same pattern as `hardware-smoke`/`ollama-smoke`). The goal is
scaffolding that future phases add journeys to, not a full regression net now.

### Out of scope

- Coverage percentage gates in CI (ratchet later once baseline exists)
- Load/perf testing
- Mutation testing

---

## Acceptance

- [x] `farmauthz` role×action matrix table-driven test; `authctx`, `auditlog`, `pushnotify`, `httputil` covered
- [x] Smoke tests exist for every currently-untested handler package (at minimum happy path + RBAC deny)
- [x] `DefaultTimeout` 120 s; suite-wide smoke client timeout; both documented
- [x] INSTALL.md documents CPU-Ollama E2E flags
- [x] UI component tests for Pi wizard steps + GuardianModelSelector + ActuatorCard
- [x] `make e2e-browser` runs 3 green Playwright journeys against a seeded dev stack; CI lane documented
- [x] `go test ./...` and `npm test` remain green

---

## Files expected to change

| Area | Files |
|------|-------|
| Units | `internal/{farmauthz,authctx,auditlog,pushnotify,httputil,costing,…}/*_test.go` (new) |
| Smokes | `cmd/api/smoke_phase117_*.go` (new) |
| Infra | `internal/rag/llm/chat.go`, `cmd/api/smoke_helpers_test.go`, `INSTALL.md` |
| UI | `ui/src/__tests__/…` (new component tests) |
| E2E | `e2e/` (new Playwright project), `.github/workflows/ci.yml`, `Makefile` |
