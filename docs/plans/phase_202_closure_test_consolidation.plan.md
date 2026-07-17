---
name: Phase 202 — Closure test consolidation (Guardian + Dashboard)
overview: >
  The repo has 235 UI test files; 104 are phase-*-closure.test.js. Redundancy
  audit found GuardianChatPanel.vue touched by 24 test files and Dashboard.vue
  by 15 — many re-read the same source strings. This phase consolidates
  overlapping assertions into focused test modules without reducing behavioral
  coverage.
todos:
  - id: ws1-inventory
    content: "WS1: script inventory — for GuardianChatPanel and Dashboard, list each phase-closure file + assertion topic (grep data-test, file read, string contains); mark duplicate vs unique"
    status: pending
  - id: ws2-guardian-home
    content: "WS2: create ui/src/__tests__/guardian-chat-panel.test.js (or extend guardian-panel.test.js) as canonical home for GuardianChatPanel wiring; migrate unique assertions from phase-* files"
    status: pending
  - id: ws3-dashboard-home
    content: "WS3: extend dashboard-workspace-links.test.js + today-excellence-arc.test.js as Dashboard canonical homes; migrate FarmCanvas / site strip assertions"
    status: pending
  - id: ws4-thin-closures
    content: "WS4: replace duplicated phase-closure blocks with one-liner `import './guardian-chat-panel.test.js'` or delete redundant files; keep phase-N-closure only when it tests phase-specific behavior nowhere else"
    status: pending
  - id: ws5-ci-docs
    content: "WS5: document test ownership in docs/testing-ui.md (or README test section); phase-202-closure.test.js counts files before/after"
    status: pending
isProject: false
---

# Phase 202 — Closure test consolidation

**Status:** planned · **Depends on:** none (janitorial; safe anytime)

## The problem

Phase closure tests were the right tool while shipping fast — each phase locks
its diff. At 200+ phases the pattern creates:

- **GuardianChatPanel.vue** — 24 test files (18 phase-closure)
- **Dashboard.vue** — 15 test files (12 phase-closure)
- **Settings.vue** — 10; **GuardianActionProposal.vue** — 9

Most phase-closure files use `readFileSync` + `expect(source).toContain(...)`.
Five phases asserting the same `data-test="guardian-chat-panel"` string adds
noise, not safety.

**Do not** delete tests blindly — merge assertions, then delete duplicates.

## What to ship

### WS1 — Inventory spreadsheet (in plan or `docs/testing-ui.md`)

For each hot component:

| File | Unique assertion? | Merge target |
|------|-------------------|--------------|
| phase-170-closure.test.js | farm counsel flag | guardian-chat-panel |
| … | … | … |

### WS2 — Guardian canonical module

Target: **one** file owns:

- Panel mount smoke (existing `guardian-panel.test.js`)
- Chat panel layout / data-test anchors
- accuracy_note banner reload (from phase-200)
- citation link rendering (guardian-citation-links.test.js may stay separate if integration-heavy)

Phase closures that only `readFileSync(GuardianChatPanel.vue)` → delete after migrate.

### WS3 — Dashboard canonical module

Merge into:

- `dashboard-workspace-links.test.js` (links, bell, workspace CTAs)
- `today-excellence-arc.test.js` (Today layout arc)
- `farm-hub.test.js` (site strip / canvas if applicable)

Keep phase-173+ tests that assert **numeric behavior** (not string contains).

### WS4 — Thin remaining phase closures

Acceptable end state per shipped phase:

```js
// phase-170-closure.test.js — only phase-170-specific behavior
describe('Phase 170 — farm counsel auto-send', () => { ... })
// NOT: re-assert entire GuardianChatPanel template
```

Optional: rename `phase-NNN-closure.test.js` → topic files over time (not required in this phase).

### WS5 — Guardrails

- `npm test` in ui/ must pass with **same or fewer** skipped tests
- phase-202-closure.test.js: assert GuardianChatPanel mentioned in ≤8 test files (target), Dashboard ≤6

## Acceptance criteria

- [ ] GuardianChatPanel string-read assertions consolidated (≥50% reduction in phase files touching it)
- [ ] Dashboard same (≥40% reduction)
- [ ] No behavioral test removed without equivalent in canonical file
- [ ] CI green
- [ ] Short "where to add UI tests" doc section

## Out of scope

- Deleting all phase-closure files globally (only hot components in this phase)
- Backend test consolidation
- E2e / Playwright

## Ponytail note

**Deletion over addition** — merge first, delete second. One canonical test beats five copy-paste readers.
