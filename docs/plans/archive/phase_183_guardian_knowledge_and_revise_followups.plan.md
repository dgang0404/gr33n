---
name: Phase 183 — Contextual knowledge links + Help hub + task revise matchers
overview: >
  Three medium-effort follow-ups from the 2026-07-13 sit-in, layered on top
  of Phase 180 (knowledge discoverability): contextual crop/zone/alert links
  into the symptom guide, collapsing Help's four tabs into one Library hub
  with sections, and extending the rule-based proposal-revise matcher so
  create_task title changes bump Revision like feed/schedule revises do.
todos:
  - id: ws1-contextual-links
    content: "WS1: 'Symptoms for this crop' link from Plants / zone detail / alert cards → /symptom-guide with crop pre-selected"
    status: completed
  - id: ws2-help-hub
    content: "WS2: Collapse Help's four equal tabs (Guide/Knowledge/Catalog/Symptoms) into one Library hub with sections, built on Phase 180's 'what lives where' map"
    status: completed
  - id: ws3-task-revise-matchers
    content: "WS3: Extend internal/farmguardian rule-based proposals_revise.go so create_task title/description corrections bump Revision (parity with volume/schedule revise)"
    status: completed
  - id: ws4-tests-docs
    content: "WS4: Vitest for contextual links + hub sections; Go tests for task-title revise matcher; phase-183-closure"
    status: completed
isProject: false
---

# Phase 183 — Contextual knowledge links + Help hub + task revise matchers

**Status:** shipped · **Follows:** [182](phase_182_guardian_quick_ux_wins.plan.md) · **Builds on:** [180](phase_180_knowledge_surfaces_discoverability.plan.md)

## The problem

Three "natural Phase 181-ish" ideas from the same feedback pass, each
medium-sized rather than quick:

| Idea | Why |
|------|-----|
| Contextual links from operational pages | From Plants / a zone / an alert: "Symptoms for this crop" → symptom guide with crop pre-selected. Today the symptom guide is only reachable via Guardian citations or Help (Phase 180). |
| Help workspace density | Even after Phase 180's discoverability map, four equal tabs + jump pills + search + catalog list is a lot of surface for one workspace. |
| `create_task` revise via matchers | Rule-based revise (`proposals_revise.go`) already handles feed volume and schedule pause/resume well; task title/description corrections don't reliably bump `Revision`, so Refine on a task proposal may answer in chat without actually revising the pending row. |

## Workstreams

### WS1 — Contextual crop → symptom links
- Plants list/detail, zone detail, and alert cards get a small "Symptoms
  for this crop" (or "for this zone") link.
- Link target: `/symptom-guide?crop_key=<slug>` (reuse Phase 180 WS2's
  dropdown-backed filters — this is just a new entry point, not a new
  filter mechanism).
- Only render the link when the page already knows a `crop_key` (skip for
  ungrounded/unknown-crop rows).

### WS2 — Help Library hub
- Depends on Phase 180 WS1 (the four-surface map) shipping first.
- Replace the four equal-weight tabs with one **Library** landing that
  groups by task ("browse", "search", "import", "diagnose a symptom")
  instead of by internal surface name, using the same underlying
  Guide/Knowledge/Catalog/Symptom components as sections rather than tabs.
- Preserve all existing deep-link query params (`?tab=knowledge`,
  `?crop_key=`, etc.) as anchors/section scrolls so citation links keep
  working unchanged.

### WS3 — Task revise matchers
- Audit `internal/farmguardian/eval` / proposal revise matcher package
  (`proposals_revise.go`) for how volume ("0.3 L instead of 0.5") and
  schedule ("pause instead of resume") corrections are parsed into a new
  `Revision`.
- Add matchers for `create_task` / `create_task_from_alert`: title swap
  ("call it X instead"), description edits, and due-date/zone corrections
  where the tool args support it.
- Bump `Revision` and `SupersedesProposalID` the same way existing revise
  paths do so the Pending tab shows the corrected row, not a silent no-op.
- The new `scenario-task-dialogue-pending` fixture from Phase 184 is a
  ready-made manual/automated test bed for this once matchers land (today
  it only asks a clarifying question; extend it to send a real correction
  once WS3 ships).

### WS4 — Tests + docs
- Vitest: contextual links render only with a known `crop_key`; navigate
  to symptom guide with the right query param.
- Vitest: Help hub renders sections + preserves deep links.
- Go: unit tests for new task-revise matchers (title/description/due-date),
  asserting `Revision` increments and args merge correctly.
- `phase-183-closure.test.js`.

## Non-goals

- No new endpoints for contextual links — client-side routing only.
- No change to embedding/search quality (Phase 180 territory).
- No UI for authoring revise rules — matcher patterns stay code, same as
  existing volume/schedule matchers.

## Acceptance

- [x] "Symptoms for this crop" link visible on Plants/zone/alert where a
      crop is known; lands on symptom guide pre-filtered.
- [x] Help workspace reads as one Library with sections, not four
      unexplained tabs; existing deep links still resolve.
- [x] `create_task` title/description corrections bump proposal `Revision`
      via rule-based revise (no LLM round-trip required).
- Zone assignment revise shipped in [Phase 185](phase_185_guardian_task_zone_revise.plan.md).
- Due-date revise + tool execution shipped in [Phase 186](phase_186_guardian_task_due_date_revise.plan.md).
- Relative due-date revise (`due tomorrow`, `due in N days`) shipped in [Phase 187](phase_187_guardian_relative_due_date_revise.plan.md).
- [x] Vitest + Go test closure green.
