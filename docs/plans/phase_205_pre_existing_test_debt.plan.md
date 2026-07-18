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
    status: completed
  - id: ws3-guardian-cluster
    content: "WS3: shared root cause was missing gr33n_token in localStorage — guardianProposals/guardianReadiness stores no-op API calls without it; fixed 8 test files by seeding test-token in beforeEach + adding guardian/models/health stubs where chat panel send was blocked by groundedModelBlockReason"
    status: completed
  - id: ws4-stale-assertions
    content: "WS4: phase-143/144/46-ws3/56 closure tests updated to match current code (SmokeTopicDriftNote, AnswerContainsMetaCorrection in answer_leak.go, attachProposals contextRef param, compare_ids source wiring)"
    status: completed
  - id: ws5-directive-warning
    content: "WS5: ui/src/test-setup.js registers v-nav-hint globally via vitest.config.js setupFiles"
    status: completed
  - id: ws6-shrink-baseline
    content: "WS6: test-baseline-known-failures.json emptied — full suite 242 files / 1222 tests green"
    status: completed
isProject: false
---

# Phase 205 — Pre-existing UI test debt + regression safety net

**Status:** shipped · **Depends on:** none (bugfix + process)

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
- [x] `auth.test.js` passes
- [x] Guardian cluster (8 files) passes — shared root cause: missing `gr33n_token` in test localStorage
- [x] `phase-143/144/46-ws3/56` closure assertions updated to match current (verified-correct) code
- [x] `nav-hint` directive registered globally in `ui/src/test-setup.js`
- [x] `ui/test-baseline-known-failures.json` is empty
- [x] Full UI suite green: 242 files / 1222 tests (`npm --prefix ui test -- --run`)

## Out of scope

Phase 206 (`docs/plans` archive migration) — separate concern, separate
plan, coordinate only on not touching the same files in the same PR.
