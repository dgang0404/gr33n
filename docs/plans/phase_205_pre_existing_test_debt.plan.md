---
name: Phase 205 — Pre-existing UI test debt + regression safety net
overview: >
  Full `npm --prefix ui test -- --run` finds 14 files / 24 tests failing on
  `main` today, unrelated to Phases 202-204. Confirmed via `git stash` that
  they predate this session. Root cause: reviewers/agents have only been
  running the new phase's own closure test before shipping, not the full
  suite, so collateral drift in shared components/helpers accumulates
  silently. This phase (a) triages and fixes the 24 tests, split by root
  cause, and (b) ships a baseline-diff check so future phases can't add to
  the pile without noticing.
todos:
  - id: ws1-safety-net
    content: "WS1: ui/test-baseline-known-failures.json + scripts/check-ui-test-baseline.mjs + make check-ui-test-baseline — fails only on failures NOT already in the baseline; verified it catches a deliberately-reintroduced known failure"
    status: completed
  - id: ws2-mock-drift
    content: "WS2: auth.test.js — add resetUnauthorizedGate to the ../api mock (store calls it on login(), test mock predates that call)"
    status: pending
  - id: ws3-guardian-cluster
    content: "WS3: investigate shared root cause across guardian-inbox, guardian-chat-background, guardian-chat-grounded-gate, guardian-chat-proposals, guardian-settings-awakening, guardian-settings-corpus, phase-182-closure, phase-197-closure (8 files, ~17 tests) — v-if gates staying closed / elements never rendering suggests one shared mock-shape or setup drift, not 8 unrelated bugs"
    status: pending
  - id: ws4-stale-assertions
    content: "WS4: phase-143-closure, phase-144-closure, phase-46-ws3-handler, phase-56-closure — each asserts an exact literal source string/call-site that a later, legitimate phase changed (e.g. attachProposals gained a context_ref param after Phase 46 shipped). Confirm current code is correct, then update the assertion — do not revert working code to match a stale test"
    status: pending
  - id: ws5-directive-warning
    content: "WS5: `[Vue warn]: Failed to resolve directive: nav-hint` noise in chat panel mounts — register the directive (or a no-op stub) in ui/src/test-setup or the shared mount helper so future chat-panel tests don't inherit the warning"
    status: pending
  - id: ws6-shrink-baseline
    content: "WS6: as each test is fixed, remove its entry from test-baseline-known-failures.json; phase closes when the file is empty (or documents which entries are intentionally deferred and why)"
    status: pending
isProject: false
---

# Phase 205 — Pre-existing UI test debt + regression safety net

**Status:** in progress · **Depends on:** none (bugfix + process)

## The problem

`npm --prefix ui test -- --run` on a clean `main` checkout (verified via
`git stash`, not something Phase 202-204 introduced):

```
Test Files  14 failed | 226 passed (240)
     Tests  24 failed | 1190 passed (1214)
```

Nobody caught this because the standard workflow runs the *new* phase's own
`phase-N-closure.test.js`, not the full suite — CONTRIBUTING.md said to run
`npm --prefix ui test -- --run` but nothing enforced it, and a wall of
pre-existing red makes it easy to miss one more red line.

## Root-cause triage (WS1 already ships the mechanism; WS2-5 are the fixes)

| # | Files | Root cause | Category |
|---|-------|------------|----------|
| 1 | `auth.test.js` (2 tests) | `auth.js` store calls `resetUnauthorizedGate()` on login; the test's `vi.mock('../api')` never added that export, so the mock throws | Stale mock — code is fine, test infra didn't track a new call |
| 2 | `guardian-inbox.test.js` (4), `guardian-chat-background.test.js` (1), `guardian-chat-grounded-gate.test.js` (1), `guardian-chat-proposals.test.js` (1), `guardian-settings-awakening.test.js` (6), `guardian-settings-corpus.test.js` (2), `phase-182-closure.test.js` (1), `phase-197-closure.test.js` (1) | Elements the tests look for never render — `v-if` branches stay closed (`<!--v-if-->` in the mounted HTML) or `api.get`/`api.post` are never called. All eight files touch Guardian chat/inbox/settings surfaces, suggesting **one** shared drift (a store initialization step, a changed API response shape the mocks don't match, or a new required prop) rather than 8 independent bugs | Needs investigation — likely 1-2 real fixes, not 17 |
| 3 | `phase-143-closure.test.js`, `phase-144-closure.test.js`, `phase-46-ws3-handler.test.js`, `phase-56-closure.test.js` (4 tests) | Test asserts an **exact literal string** at a specific call-site/file that a later phase legitimately changed. Confirmed example: Phase 46's closure test expects `attachProposals(r.Context(), farmID, hasUser, userID, sessionID, question, answer, liveSnap, &resp)` — current code is `attachProposals(r.Context(), farmID, hasUser, userID, sessionID, question, answer, liveSnap, pb.ContextRef, &resp)`, i.e. `context_ref` support was added correctly, the literal-string assertion just wasn't updated | Stale assertion — code is correct, update the test |
| 4 | Console noise: `[Vue warn]: Failed to resolve directive: nav-hint` on every chat panel mount in tests | `v-nav-hint` directive isn't registered in the test mount path | Cosmetic — not causing failures directly, but worth silencing so real warnings aren't lost in the noise |

## WS3 investigation approach (the 17-test cluster)

Before touching 8 files individually:

1. Pick the smallest failing file (`guardian-chat-grounded-gate.test.js`,
   1 test) and trace why `[data-test="chat-use-farm-context"]` never
   renders — check what condition gates it and what the test's mocked
   `api` / store state is missing.
2. If the same gate condition explains 2+ of the other files, fix the
   shared setup helper or mock fixture once, re-run all 8 files, and only
   then debug whatever's left individually.
3. For each fix, confirm it's a **test-side** fix (mock/setup didn't keep
   up) rather than a **component-side regression** (the UI genuinely
   broke) before changing anything — grep `git log -p` on the component
   for the phase that last touched the relevant `data-test` attribute.

## WS4 rule

Don't "fix" these by reverting code to match the old assertion. Verify the
current behavior is intentional (there's almost always a later phase plan
documenting the change — e.g. context_ref shipped well after Phase 46),
then update the assertion to match reality. If a WS4 file turns out to be
a real regression instead, move it to WS3's process (root-cause the
component, not the test).

## Acceptance criteria

- [x] `make check-ui-test-baseline` ships and is documented in CONTRIBUTING.md
- [ ] `auth.test.js` passes
- [ ] Guardian cluster (8 files) passes or has a documented shared root cause with a tracked follow-up if any part is deferred
- [ ] `phase-143/144/46-ws3/56` closure assertions updated to match current (verified-correct) code
- [ ] `nav-hint` directive warning silenced in test mounts
- [ ] `ui/test-baseline-known-failures.json` is empty, or every remaining entry has an inline reason it's deferred
- [ ] `make test-unit` / `make test` still green (no Go-side regressions from any fix)

## Out of scope

Phase 206 (`docs/plans` archive migration) — separate concern, separate
plan, coordinate only on not touching the same files in the same PR.
